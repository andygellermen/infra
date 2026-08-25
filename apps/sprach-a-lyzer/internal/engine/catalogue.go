package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/rules"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/seed"
)

type CatalogueProvider interface {
	Active(context.Context) (rules.Catalogue, error)
}

// TextProvider supplies one presentation-safe key catalogue per request.
type TextProvider interface {
	Texts(context.Context, string, string) (map[string]string, error)
}

type staticCatalogue struct{ catalogue rules.Catalogue }

func (s staticCatalogue) Active(context.Context) (rules.Catalogue, error) { return s.catalogue, nil }

type staticTexts struct{ bundles map[string]map[string]string }

func (s staticTexts) Texts(_ context.Context, profile, locale string) (map[string]string, error) {
	entries, ok := s.bundles[profile+"/"+locale]
	if !ok {
		return nil, fmt.Errorf("presentation bundle %s/%s is unavailable", profile, locale)
	}
	return mapsClone(entries), nil
}

func NewDefault() *Engine {
	foundation := embeddedFoundation()
	return NewWithProviders(staticCatalogue{catalogue: rules.Catalogue{
		Version: foundation.RuleSet.Version, Rules: foundation.Rules,
	}}, staticTextProvider(foundation))
}

func embeddedFoundation() seed.Foundation {
	foundation, err := seed.DecodeFoundation(bytes.NewReader(assets.FoundationV03))
	if err != nil {
		panic(fmt.Sprintf("decode embedded Foundation catalogue: %v", err))
	}
	return foundation
}

func staticTextProvider(foundation seed.Foundation) TextProvider {
	bundles := make(map[string]map[string]string, len(foundation.PresentationBundles))
	for _, bundle := range foundation.PresentationBundles {
		bundles[bundle.Profile+"/"+bundle.Locale] = mapsClone(bundle.Entries)
	}
	return staticTexts{bundles: bundles}
}

func mapsClone(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

type catalogueFacts struct {
	phrase, context, inputMode, targetType, expectationSource         string
	tokens, patterns, senses, discourseRelations, propositionFeatures []string
}

func (e *Engine) activeDefinitions(request domain.AnalysisRequest, normalizedText string, resolution domain.ResolverResult) ([]rules.Definition, catalogueFacts, error) {
	if e.catalogue == nil {
		return nil, catalogueFacts{}, fmt.Errorf("rule catalogue provider is nil")
	}
	catalogue, err := e.catalogue.Active(context.Background())
	if err != nil {
		return nil, catalogueFacts{}, fmt.Errorf("load active rule catalogue: %w", err)
	}
	if catalogue.Version == "" || len(catalogue.Rules) == 0 {
		return nil, catalogueFacts{}, fmt.Errorf("active rule catalogue is empty")
	}
	definitions := slices.Clone(catalogue.Rules)
	sort.SliceStable(definitions, func(i, j int) bool {
		if definitions[i].Priority == definitions[j].Priority {
			return definitions[i].Key < definitions[j].Key
		}
		return definitions[i].Priority > definitions[j].Priority
	})
	for _, definition := range definitions {
		if definition.Enabled {
			if err := definition.Validate(); err != nil {
				return nil, catalogueFacts{}, fmt.Errorf("validate active catalogue: %w", err)
			}
		}
	}
	facts := catalogueFacts{
		phrase: normalizePhrase(normalizedText), tokens: catalogueTokens(normalizedText),
		context:    strings.ToUpper(strings.TrimSpace(string(request.Context))),
		inputMode:  strings.ToUpper(strings.TrimSpace(string(request.InputMode))),
		targetType: string(resolution.TargetType), expectationSource: string(resolution.ExpectationSource),
	}
	for _, sense := range resolution.SelectedSenses {
		facts.senses = append(facts.senses, sense.Sense)
	}
	for _, edge := range resolution.PropositionGraph.Edges {
		facts.discourseRelations = append(facts.discourseRelations, string(edge.Relation))
	}
	for _, node := range resolution.PropositionGraph.Nodes {
		if node.Predicate {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "PREDICATE")
		}
		if node.Target {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "TARGET")
		}
		if node.Time {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "TIME")
		}
		if node.Boundary {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "BOUNDARY")
		}
		if node.Decision {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "DECISION")
		}
		if node.Negation {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "NEGATION")
		}
		if node.Modality != domain.ModalityNone {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "MODALITY_"+string(node.Modality))
		}
	}
	if facts.context == "" {
		facts.context = string(domain.ContextUnspecified)
	}
	if facts.inputMode == "" {
		facts.inputMode = string(domain.InputModeText)
	}
	return definitions, facts, nil
}

func evaluateCondition(condition rules.Condition, facts catalogueFacts) (bool, error) {
	switch condition.Op {
	case "AND":
		for _, child := range condition.Children {
			matched, err := evaluateCondition(child, facts)
			if err != nil || !matched {
				return false, err
			}
		}
		return true, nil
	case "OR":
		for _, child := range condition.Children {
			matched, err := evaluateCondition(child, facts)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
		return false, nil
	case "NOT":
		matched, err := evaluateCondition(*condition.Child, facts)
		return !matched, err
	}
	var values []string
	switch condition.Field {
	case "tokens":
		values = facts.tokens
	case "phrase":
		values = []string{facts.phrase}
	case "context":
		values = []string{facts.context}
	case "input_mode":
		values = []string{facts.inputMode}
	case "pattern":
		values = facts.patterns
	case "selected_sense":
		values = facts.senses
	case "target_type":
		values = []string{facts.targetType}
	case "expectation_source":
		values = []string{facts.expectationSource}
	case "discourse_relation":
		values = facts.discourseRelations
	case "proposition_feature":
		values = facts.propositionFeatures
	default:
		return false, fmt.Errorf("runtime field %q is not supported", condition.Field)
	}
	return evaluatePredicate(values, condition.Operator, condition.Value, condition.CaseSensitive, condition.Field == "phrase")
}

func appendUniqueFact(values []string, value string) []string {
	if !slices.Contains(values, value) {
		return append(values, value)
	}
	return values
}

func evaluatePredicate(actual []string, operator string, rawExpected json.RawMessage, caseSensitive, substringMatch bool) (bool, error) {
	if operator == "EXISTS" {
		return len(actual) > 0, nil
	}
	if operator == "NOT_EXISTS" {
		return len(actual) == 0, nil
	}
	expected, err := decodeExpected(rawExpected)
	if err != nil {
		return false, err
	}
	if !caseSensitive {
		actual, expected = lowerStrings(actual), lowerStrings(expected)
	}
	equalsAny := func(want string) bool { return slices.Contains(actual, want) }
	contains := func(want string) bool {
		for _, value := range actual {
			if value == want || (substringMatch && strings.Contains(value, want)) {
				return true
			}
		}
		return false
	}
	switch operator {
	case "EQUALS":
		return len(expected) == 1 && equalsAny(expected[0]), nil
	case "NOT_EQUALS":
		return len(expected) == 1 && !equalsAny(expected[0]), nil
	case "CONTAINS":
		return len(expected) == 1 && contains(expected[0]), nil
	case "CONTAINS_ANY":
		for _, want := range expected {
			if contains(want) {
				return true, nil
			}
		}
		return false, nil
	case "CONTAINS_ALL":
		for _, want := range expected {
			if !contains(want) {
				return false, nil
			}
		}
		return true, nil
	case "MATCHES":
		if len(expected) != 1 || len(expected[0]) > 512 {
			return false, fmt.Errorf("MATCHES requires one bounded expression")
		}
		expression, err := regexp.Compile(expected[0])
		if err != nil {
			return false, fmt.Errorf("compile MATCHES expression: %w", err)
		}
		for _, value := range actual {
			if expression.MatchString(value) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("operator %q is not supported", operator)
	}
}

func decodeExpected(value json.RawMessage) ([]string, error) {
	var single string
	if err := json.Unmarshal(value, &single); err == nil {
		return []string{single}, nil
	}
	var multiple []string
	if err := json.Unmarshal(value, &multiple); err != nil {
		return nil, fmt.Errorf("predicate value must be a string or string array")
	}
	return multiple, nil
}

func catalogueTokens(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) })
	result := make([]string, 0, len(fields)+1)
	for _, field := range fields {
		field = strings.ToLower(field)
		result = append(result, field)
		if field == "muss" || field == "musst" || field == "müssen" {
			result = append(result, "müssen")
		}
	}
	return result
}

func normalizePhrase(text string) string {
	return strings.TrimFunc(text, func(r rune) bool { return unicode.IsPunct(r) || unicode.IsSpace(r) })
}

func lowerStrings(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = strings.ToLower(value)
	}
	return result
}
