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
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/ontology"
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
	foundation, err := seed.DecodeFoundation(bytes.NewReader(assets.FoundationV04))
	if err != nil {
		panic(fmt.Sprintf("decode embedded Foundation catalogue: %v", err))
	}
	return foundation
}

func embeddedOntologyProvider() ontology.CatalogueProvider {
	catalogue, err := ontology.Decode(bytes.NewReader(assets.ConstructOntologyV02))
	if err != nil {
		panic(fmt.Sprintf("decode embedded construct ontology: %v", err))
	}
	return ontology.StaticProvider{Catalogue: catalogue}
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
	phrase, context, inputMode             string
	tokens, patterns, senses, targetTypes  []string
	expectationSources, discourseRelations []string
	propositionFeatures, constructs        []string
	compositions                           []string
	references                             map[string][]factReference
}

type factReference struct {
	value          string
	propositionIDs []string
}

type conditionMatch struct {
	matched        bool
	propositionIDs []string
}

func (e *Engine) activeDefinitions(request domain.AnalysisRequest, normalizedText string, resolution domain.ResolverResult, constructResult ontology.Result) ([]rules.Definition, catalogueFacts, error) {
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
		references: make(map[string][]factReference),
	}
	for _, node := range resolution.PropositionGraph.Nodes {
		facts.addReference("phrase", normalizePhrase(normalize(node.Text)), node.ID)
		for _, token := range catalogueTokens(normalize(node.Text)) {
			facts.addReference("tokens", token, node.ID)
		}
		facts.targetTypes = appendUniqueFact(facts.targetTypes, string(node.TargetType))
		facts.expectationSources = appendUniqueFact(facts.expectationSources, string(node.ExpectationSource))
		facts.addReference("target_type", string(node.TargetType), node.ID)
		facts.addReference("expectation_source", string(node.ExpectationSource), node.ID)
	}
	for _, evidence := range constructResult.Evidence {
		value := string(evidence.ConstructID)
		facts.constructs = appendUniqueFact(facts.constructs, value)
		facts.addReference("construct", value, evidence.PropositionIDs...)
	}
	for _, composition := range constructResult.Compositions {
		facts.compositions = appendUniqueFact(facts.compositions, composition.Pattern)
		facts.addReference("composition", composition.Pattern, composition.PropositionIDs...)
	}
	// Resolver candidates are not rule patterns. Only a non-ambiguous selected
	// sense becomes an addressable fact; this enforces both resolver scoring
	// guardrails at the engine boundary.
	for _, sense := range resolution.SelectedSenses {
		if sense.State == domain.SenseAmbiguous {
			continue
		}
		facts.senses = appendUniqueFact(facts.senses, sense.Sense)
		facts.addReference("selected_sense", sense.Sense, sense.PropositionID)
	}
	for _, edge := range resolution.PropositionGraph.Edges {
		facts.discourseRelations = append(facts.discourseRelations, string(edge.Relation))
		facts.addReference("discourse_relation", string(edge.Relation), edge.Source, edge.Target)
	}
	for _, node := range resolution.PropositionGraph.Nodes {
		if node.Predicate {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "PREDICATE")
			facts.addReference("proposition_feature", "PREDICATE", node.ID)
		}
		if node.Target {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "TARGET")
			facts.addReference("proposition_feature", "TARGET", node.ID)
		}
		if node.Time {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "TIME")
			facts.addReference("proposition_feature", "TIME", node.ID)
		}
		if node.Boundary {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "BOUNDARY")
			facts.addReference("proposition_feature", "BOUNDARY", node.ID)
		}
		if node.Decision {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "DECISION")
			facts.addReference("proposition_feature", "DECISION", node.ID)
		}
		if node.Negation {
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, "NEGATION")
			facts.addReference("proposition_feature", "NEGATION", node.ID)
		}
		if node.Modality != domain.ModalityNone {
			value := "MODALITY_" + string(node.Modality)
			facts.propositionFeatures = appendUniqueFact(facts.propositionFeatures, value)
			facts.addReference("proposition_feature", value, node.ID)
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

func (f *catalogueFacts) addReference(field, value string, propositionIDs ...string) {
	if value == "" {
		return
	}
	for index := range f.references[field] {
		if f.references[field][index].value != value {
			continue
		}
		f.references[field][index].propositionIDs = appendUniqueFact(f.references[field][index].propositionIDs, propositionIDs...)
		return
	}
	ids := []string{}
	ids = appendUniqueFact(ids, propositionIDs...)
	f.references[field] = append(f.references[field], factReference{value: value, propositionIDs: ids})
}

func trustedResolverSenses(senses []domain.ResolverSense) []string {
	result := make([]string, 0, len(senses))
	for _, sense := range senses {
		if sense.State == domain.SenseAmbiguous {
			continue
		}
		result = appendUniqueFact(result, sense.Sense)
	}
	return result
}

func evaluateCondition(condition rules.Condition, facts catalogueFacts) (conditionMatch, error) {
	switch condition.Op {
	case "AND":
		result := conditionMatch{matched: true, propositionIDs: []string{}}
		for _, child := range condition.Children {
			childMatch, err := evaluateCondition(child, facts)
			if err != nil || !childMatch.matched {
				return conditionMatch{}, err
			}
			result.propositionIDs = appendUniqueFact(result.propositionIDs, childMatch.propositionIDs...)
		}
		return result, nil
	case "OR":
		result := conditionMatch{propositionIDs: []string{}}
		for _, child := range condition.Children {
			childMatch, err := evaluateCondition(child, facts)
			if err != nil {
				return conditionMatch{}, err
			}
			if childMatch.matched {
				result.matched = true
				result.propositionIDs = appendUniqueFact(result.propositionIDs, childMatch.propositionIDs...)
			}
		}
		return result, nil
	case "NOT":
		childMatch, err := evaluateCondition(*condition.Child, facts)
		return conditionMatch{matched: !childMatch.matched, propositionIDs: []string{}}, err
	}
	values, supported := facts.values(condition.Field)
	if !supported {
		return conditionMatch{}, fmt.Errorf("runtime field %q is not supported", condition.Field)
	}
	matched, err := evaluatePredicate(values, condition.Operator, condition.Value, condition.CaseSensitive, condition.Field == "phrase")
	if err != nil || !matched {
		return conditionMatch{matched: matched}, err
	}
	propositionIDs, err := facts.provenance(condition)
	if err != nil {
		return conditionMatch{}, err
	}
	return conditionMatch{matched: true, propositionIDs: propositionIDs}, nil
}

func (f catalogueFacts) values(field string) ([]string, bool) {
	switch field {
	case "tokens":
		return f.tokens, true
	case "phrase":
		return []string{f.phrase}, true
	case "context":
		return []string{f.context}, true
	case "input_mode":
		return []string{f.inputMode}, true
	case "pattern":
		return f.patterns, true
	case "selected_sense":
		return f.senses, true
	case "target_type":
		return f.targetTypes, true
	case "expectation_source":
		return f.expectationSources, true
	case "discourse_relation":
		return f.discourseRelations, true
	case "proposition_feature":
		return f.propositionFeatures, true
	case "construct":
		return f.constructs, true
	case "composition":
		return f.compositions, true
	default:
		return nil, false
	}
}

func (f catalogueFacts) provenance(condition rules.Condition) ([]string, error) {
	if condition.Operator == "NOT_EQUALS" || condition.Operator == "NOT_CONTAINS" || condition.Operator == "NOT_EXISTS" {
		return []string{}, nil
	}
	references := f.references[condition.Field]
	if condition.Operator == "EXISTS" {
		result := []string{}
		for _, reference := range references {
			result = appendUniqueFact(result, reference.propositionIDs...)
		}
		return result, nil
	}
	expected, err := decodeExpected(condition.Value)
	if err != nil {
		return nil, err
	}
	result := []string{}
	if condition.Operator == "CONTAINS_ALL" {
		for _, want := range expected {
			for _, reference := range references {
				matched, err := evaluatePredicate([]string{reference.value}, "CONTAINS", mustJSON(want), condition.CaseSensitive, condition.Field == "phrase")
				if err != nil {
					return nil, err
				}
				if matched {
					result = appendUniqueFact(result, reference.propositionIDs...)
				}
			}
		}
		return result, nil
	}
	for _, reference := range references {
		matched, err := evaluatePredicate([]string{reference.value}, condition.Operator, condition.Value, condition.CaseSensitive, condition.Field == "phrase")
		if err != nil {
			return nil, err
		}
		if matched {
			result = appendUniqueFact(result, reference.propositionIDs...)
		}
	}
	return result, nil
}

func mustJSON(value string) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}

func appendUniqueFact(values []string, additions ...string) []string {
	for _, value := range additions {
		if value != "" && !slices.Contains(values, value) {
			values = append(values, value)
		}
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
