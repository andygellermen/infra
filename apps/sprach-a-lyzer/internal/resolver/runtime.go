package resolver

import (
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
)

type catalogueRuntime struct {
	catalogue  Catalogue
	lexemes    map[string]LexemeDefinition
	connectors []runtimeConnector
	scopes     []ScopeDefinition
}

type runtimeConnector struct {
	marker     string
	relation   domain.DiscourseRelationID
	placement  string
	confidence float64
}

func newCatalogueRuntime(catalogue Catalogue) catalogueRuntime {
	runtime := catalogueRuntime{
		catalogue: catalogue,
		lexemes:   make(map[string]LexemeDefinition, len(catalogue.Lexemes)),
		scopes:    slices.Clone(catalogue.ScopeRules),
	}
	for _, lexeme := range catalogue.Lexemes {
		runtime.lexemes[lexeme.Key] = lexeme
	}
	for _, connector := range catalogue.Connectors {
		for _, marker := range connector.Markers {
			runtime.connectors = append(runtime.connectors, runtimeConnector{
				marker: normalize(marker), relation: domain.DiscourseRelationID(connector.Relation),
				placement: connector.Placement, confidence: connector.Confidence,
			})
		}
	}
	sort.SliceStable(runtime.connectors, func(i, j int) bool {
		return len(runtime.connectors[i].marker) > len(runtime.connectors[j].marker)
	})
	sort.SliceStable(runtime.scopes, func(i, j int) bool {
		return runtime.scopes[i].Priority > runtime.scopes[j].Priority
	})
	return runtime
}

func (r catalogueRuntime) matchesLexeme(key, text string) bool {
	lexeme, ok := r.lexemes[key]
	if !ok {
		return false
	}
	for _, form := range lexeme.Forms {
		if hasWord(text, normalize(form)) {
			return true
		}
	}
	return false
}

func (r catalogueRuntime) hasSense(key, senseID string) bool {
	lexeme, ok := r.lexemes[key]
	if !ok {
		return false
	}
	for _, sense := range lexeme.Senses {
		if sense.ID == senseID {
			return true
		}
	}
	return false
}

func (r catalogueRuntime) lexemePosition(key, text string) int {
	lexeme, ok := r.lexemes[key]
	if !ok {
		return len(text)
	}
	position := len(text)
	for _, form := range lexeme.Forms {
		if index := strings.Index(text, normalize(form)); index >= 0 && index < position {
			position = index
		}
	}
	return position
}

func (r catalogueRuntime) senseState(confidence, gap float64) domain.SenseState {
	thresholds := r.catalogue.SenseThresholds
	switch {
	case confidence >= thresholds.High.MinimumConfidence && gap >= thresholds.High.MinimumGap:
		return domain.SenseHigh
	case confidence >= thresholds.Medium.MinimumConfidence && gap >= thresholds.Medium.MinimumGap:
		return domain.SenseMedium
	default:
		return domain.SenseState(thresholds.Fallback)
	}
}

func (r catalogueRuntime) connectorAt(text string, infixOnly bool) (int, runtimeConnector, bool) {
	lower := strings.ToLower(text)
	bestPosition := len(lower)
	var best runtimeConnector
	found := false
	for _, connector := range r.connectors {
		if infixOnly && connector.placement != "INFIX" {
			continue
		}
		searchFrom := 0
		for searchFrom < len(lower) {
			relative := strings.Index(lower[searchFrom:], connector.marker)
			if relative < 0 {
				break
			}
			position := searchFrom + relative
			if markerBounded(lower, position, position+len(connector.marker)) && (!infixOnly || position > 0) {
				if position < bestPosition || position == bestPosition && len(connector.marker) > len(best.marker) {
					bestPosition, best, found = position, connector, true
				}
				break
			}
			searchFrom = position + 1
		}
	}
	return bestPosition, best, found
}

func (r catalogueRuntime) relationIn(text string) (runtimeConnector, bool) {
	_, connector, ok := r.connectorAt(text, false)
	return connector, ok
}

func (r catalogueRuntime) prefixClauseAt(text string) (int, runtimeConnector, bool) {
	lower := strings.ToLower(text)
	for _, connector := range r.connectors {
		if connector.placement != "PREFIX" || connector.relation != domain.RelationCondition || !strings.HasPrefix(lower, connector.marker) ||
			!markerBounded(lower, 0, len(connector.marker)) {
			continue
		}
		comma := strings.Index(lower[len(connector.marker):], ",")
		if comma < 0 {
			return 0, runtimeConnector{}, false
		}
		return len(connector.marker) + comma, connector, true
	}
	return 0, runtimeConnector{}, false
}

func (r catalogueRuntime) negationScope(text string) domain.NegationScopeID {
	for _, scope := range r.scopes {
		for _, cue := range scope.Cues {
			if wordsAppearInOrder(text, normalize(cue)) {
				return domain.NegationScopeID(scope.Output)
			}
		}
	}
	return domain.NegationAmbiguous
}

func markerBounded(text string, start, end int) bool {
	left, _ := utf8.DecodeLastRuneInString(text[:start])
	right, _ := utf8.DecodeRuneInString(text[end:])
	return (start == 0 || !isWordRune(left)) && (end == len(text) || !isWordRune(right))
}

func isWordRune(value rune) bool { return unicode.IsLetter(value) || unicode.IsDigit(value) }

func wordsAppearInOrder(text, cue string) bool {
	textWords := lexicalWords(text)
	cueWords := lexicalWords(cue)
	if len(cueWords) == 0 {
		return false
	}
	position := 0
	for _, word := range textWords {
		if word != cueWords[position] {
			continue
		}
		position++
		if position == len(cueWords) {
			return true
		}
	}
	return false
}

func lexicalWords(text string) []string {
	return strings.FieldsFunc(text, func(value rune) bool {
		return unicode.IsSpace(value) || unicode.IsPunct(value)
	})
}
