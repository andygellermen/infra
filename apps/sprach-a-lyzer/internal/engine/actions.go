package engine

import (
	"fmt"
	"slices"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/rules"
)

func executeRule(result *domain.AnalysisResult, evidenceItems *[]evidence, definition rules.Definition) error {
	for _, action := range definition.Actions {
		switch action.Type {
		case policy.AddPattern:
			appendPattern(result, action.Key)
		case policy.SelectSense:
			sense, ok := resolvedSense(action.Lexeme, action.Sense)
			if !ok {
				return fmt.Errorf("rule %s selects unknown runtime sense %s", definition.Key, action.Sense)
			}
			result.ResolvedSenses = append(result.ResolvedSenses, sense)
		case policy.AddResonanceHint:
			if action.SemanticScore == nil || *action.SemanticScore {
				return fmt.Errorf("rule %s attempted scoring resonance", definition.Key)
			}
			message, ok := resonanceMessage(action.MessageKey)
			if !ok {
				return fmt.Errorf("rule %s uses unknown resonance message %s", definition.Key, action.MessageKey)
			}
			result.ResonanceHints = append(result.ResonanceHints, domain.ResonanceHint{
				Kind: "HOMOPHONE", Tokens: slices.Clone(action.Tokens),
				SemanticScore: false, Message: message,
			})
		case policy.AddContribution:
			if action.Value == nil {
				return fmt.Errorf("rule %s contribution has no value", definition.Key)
			}
			metadata, ok := contributionMetadataFor(action.ReasonKey)
			if !ok {
				return fmt.Errorf("rule %s uses unknown contribution reason %s", definition.Key, action.ReasonKey)
			}
			*evidenceItems = append(*evidenceItems, item(
				definition.Key, metadata.evidence, action.Dimension,
				*action.Value, metadata.strength, metadata.reason,
			))
		default:
			return fmt.Errorf("rule %s uses action %s not enabled in Foundation runtime", definition.Key, action.Type)
		}
	}
	return nil
}

func appendPattern(result *domain.AnalysisResult, pattern string) {
	for _, existing := range result.Patterns {
		if existing == pattern {
			return
		}
	}
	result.Patterns = append(result.Patterns, pattern)
}

func resolvedSense(lexeme, sense string) (domain.ResolvedSense, bool) {
	switch sense {
	case "SAFETY_NECESSITY":
		return domain.ResolvedSense{
			Lexeme: lexeme, Sense: sense, Confidence: 0.79,
			Reason: "Der explizite Sicherheitskontext schlägt die isolierte Modalverbdeutung.",
		}, true
	case "FREE_OF_CHARGE":
		return domain.ResolvedSense{
			Lexeme: lexeme, Sense: sense, Confidence: 0.80,
			Reason: "Die Kollokation „Eintritt ist frei“ bezeichnet Kostenfreiheit.",
		}, true
	default:
		return domain.ResolvedSense{}, false
	}
}

func resonanceMessage(key string) (string, bool) {
	if key != "HOMOPHONE_HAST_HASST" {
		return "", false
	}
	return "Die lautliche Nähe wird nur als Resonanzhinweis geführt und verändert keine semantische Bewertung.", true
}

type contributionMetadata struct {
	evidence string
	strength float64
	reason   string
}

func contributionMetadataFor(key string) (contributionMetadata, bool) {
	metadata := map[string]contributionMetadata{
		"INTERNAL_PRESSURE_VOLITION": {
			evidence: "ich muss", strength: 0.62,
			reason: "Notwendigkeitssprache reduziert den sichtbaren eigenen Wahlraum.",
		},
		"INTERNAL_PRESSURE_OPENNESS": {
			evidence: "ich muss", strength: 0.53,
			reason: "Die Formulierung stellt zunächst nur einen zwingenden Weg dar.",
		},
	}
	value, ok := metadata[key]
	return value, ok
}
