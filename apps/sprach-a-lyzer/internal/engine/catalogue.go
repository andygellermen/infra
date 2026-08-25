package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

type staticCatalogue struct {
	catalogue rules.Catalogue
}

func (s staticCatalogue) Active(context.Context) (rules.Catalogue, error) {
	return s.catalogue, nil
}

func NewDefault() *Engine {
	foundation, err := seed.DecodeFoundation(bytes.NewReader(assets.FoundationV02))
	if err != nil {
		panic(fmt.Sprintf("decode embedded Foundation catalogue: %v", err))
	}
	return New(staticCatalogue{catalogue: rules.Catalogue{
		Version: foundation.RuleSet.Version,
		Rules:   foundation.Rules,
	}})
}

type catalogueFacts struct {
	phrase    string
	tokens    []string
	context   string
	inputMode string
	patterns  []string
	senses    []string
}

func (e *Engine) matchingRules(request domain.AnalysisRequest, normalizedText string) (map[string]rules.Definition, error) {
	if e.catalogue == nil {
		return nil, fmt.Errorf("rule catalogue provider is nil")
	}
	catalogue, err := e.catalogue.Active(context.Background())
	if err != nil {
		return nil, fmt.Errorf("load active rule catalogue: %w", err)
	}
	if catalogue.Version == "" || len(catalogue.Rules) == 0 {
		return nil, fmt.Errorf("active rule catalogue is empty")
	}
	definitions := slices.Clone(catalogue.Rules)
	sort.SliceStable(definitions, func(i, j int) bool {
		if definitions[i].Priority == definitions[j].Priority {
			return definitions[i].Key < definitions[j].Key
		}
		return definitions[i].Priority > definitions[j].Priority
	})
	facts := catalogueFacts{
		phrase:    normalizePhrase(normalizedText),
		tokens:    catalogueTokens(normalizedText),
		context:   strings.ToUpper(strings.TrimSpace(string(request.Context))),
		inputMode: strings.ToUpper(strings.TrimSpace(string(request.InputMode))),
	}
	if facts.context == "" {
		facts.context = string(domain.ContextUnspecified)
	}
	if facts.inputMode == "" {
		facts.inputMode = string(domain.InputModeText)
	}
	matched := make(map[string]rules.Definition)
	for _, definition := range definitions {
		if !definition.Enabled {
			continue
		}
		if err := definition.Validate(); err != nil {
			return nil, fmt.Errorf("validate active catalogue: %w", err)
		}
		matches, err := evaluateCondition(definition.Condition, facts)
		if err != nil {
			return nil, fmt.Errorf("evaluate rule %s: %w", definition.Key, err)
		}
		if matches {
			matched[definition.Key] = definition
		}
	}
	return matched, nil
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
	default:
		return false, fmt.Errorf("runtime field %q is not supported", condition.Field)
	}
	return evaluatePredicate(values, condition.Operator, condition.Value, condition.CaseSensitive, condition.Field == "phrase")
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
		actual = lowerStrings(actual)
		expected = lowerStrings(expected)
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
		return false, fmt.Errorf("MATCHES is not enabled in the deterministic Foundation runtime")
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
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
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
	return strings.TrimFunc(text, func(r rune) bool {
		return unicode.IsPunct(r) || unicode.IsSpace(r)
	})
}

func lowerStrings(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = strings.ToLower(value)
	}
	return result
}
