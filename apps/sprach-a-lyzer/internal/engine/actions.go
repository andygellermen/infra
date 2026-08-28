package engine

import (
	"fmt"
	"slices"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/rules"
)

type executionState struct {
	result            *domain.AnalysisResult
	evidence          []evidence
	nonAssessable     map[domain.DimensionID]bool
	texts             map[string]string
	patternReferences map[string][]string
	senseReferences   map[string][]string
	stop              bool
}

func (s *executionState) execute(definition rules.Definition, propositionIDs []string) error {
	if s.patternReferences == nil {
		s.patternReferences = make(map[string][]string)
	}
	if s.senseReferences == nil {
		s.senseReferences = make(map[string][]string)
	}
	resolve := func(key string) (string, error) {
		value := s.texts[key]
		if value == "" {
			return "", fmt.Errorf("rule %s references unpublished presentation key %s", definition.Key, key)
		}
		return value, nil
	}
	for _, action := range definition.Actions {
		switch action.Type {
		case policy.AddPattern:
			appendPattern(s.result, action.Key)
			s.patternReferences[action.Key] = appendUniqueFact(s.patternReferences[action.Key], propositionIDs...)
		case policy.SelectSense:
			reason, err := resolve(action.ReasonKey)
			if err != nil {
				return err
			}
			s.result.ResolvedSenses = append(s.result.ResolvedSenses, domain.ResolvedSense{
				Lexeme: action.Lexeme, Sense: action.Sense,
				Confidence: roundThree(*action.Confidence * definition.ConfidenceModifier), Reason: reason,
			})
			s.senseReferences[action.Sense] = appendUniqueFact(s.senseReferences[action.Sense], propositionIDs...)
		case policy.AddResonanceHint:
			if action.SemanticScore == nil || *action.SemanticScore {
				return fmt.Errorf("rule %s attempted scoring resonance", definition.Key)
			}
			message, err := resolve(action.MessageKey)
			if err != nil {
				return err
			}
			s.result.ResonanceHints = append(s.result.ResonanceHints, domain.ResonanceHint{
				Kind: "HOMOPHONE", Tokens: slices.Clone(action.Tokens), SemanticScore: false, Message: message,
			})
		case policy.AddContribution:
			if s.nonAssessable[action.Dimension] {
				continue
			}
			reason, err := resolve(action.ReasonKey)
			if err != nil {
				return err
			}
			evidenceText, err := resolve(action.EvidenceKey)
			if err != nil {
				return err
			}
			s.evidence = append(s.evidence, item(definition.Key, evidenceText, action.Dimension,
				*action.Value, *action.Confidence*definition.ConfidenceModifier, reason, propositionIDs))
		case policy.MultiplyContribution:
			if _, err := resolve(action.ReasonKey); err != nil {
				return err
			}
			s.transform(action.Dimension, func(value float64) float64 { return value * *action.Factor })
		case policy.Invert:
			if _, err := resolve(action.ReasonKey); err != nil {
				return err
			}
			s.transform(action.Dimension, func(value float64) float64 { return -value })
		case policy.Suppress:
			if _, err := resolve(action.ReasonKey); err != nil {
				return err
			}
			s.removeDimension(action.Dimension)
		case policy.MarkNonAssessable:
			if _, err := resolve(action.ReasonKey); err != nil {
				return err
			}
			s.removeDimension(action.Dimension)
			s.nonAssessable[action.Dimension] = true
		case policy.CapMin, policy.CapMax, policy.SetValue:
			if err := s.applyAbsoluteModifier(definition.Key, action, propositionIDs, resolve); err != nil {
				return err
			}
		case policy.AddExplanation:
			value, err := resolve(action.Key)
			if err != nil {
				return err
			}
			s.result.Notes = append(s.result.Notes, value)
		case policy.AddReflectionPrompt:
			value, err := resolve(action.Key)
			if err != nil {
				return err
			}
			s.result.ReflectionQuestion = &value
		case policy.AddAlternative:
			value, err := resolve(action.Key)
			if err != nil {
				return err
			}
			s.result.Alternatives = append(s.result.Alternatives, value)
		case policy.StopRuleChain:
			s.stop = true
		default:
			return fmt.Errorf("rule %s uses unsupported runtime action %s", definition.Key, action.Type)
		}
	}
	if definition.StopProcessing {
		s.stop = true
	}
	return nil
}

func (s *executionState) transform(dimension domain.DimensionID, operation func(float64) float64) {
	for index := range s.evidence {
		if s.evidence[index].contribution.Dimension == dimension {
			s.evidence[index].contribution.Delta = roundOne(operation(s.evidence[index].contribution.Delta))
		}
	}
}

func (s *executionState) removeDimension(dimension domain.DimensionID) {
	filtered := s.evidence[:0]
	for _, entry := range s.evidence {
		if entry.contribution.Dimension != dimension {
			filtered = append(filtered, entry)
		}
	}
	s.evidence = filtered
}

func (s *executionState) applyAbsoluteModifier(ruleID string, action rules.Action, propositionIDs []string, resolve func(string) (string, error)) error {
	if s.nonAssessable[action.Dimension] {
		return nil
	}
	current := 0.0
	strength := 0.0
	for _, entry := range s.evidence {
		if entry.contribution.Dimension == action.Dimension {
			current += entry.contribution.Delta
			if entry.strength > strength {
				strength = entry.strength
			}
		}
	}
	target := current
	switch action.Type {
	case policy.CapMin:
		if target < *action.Value {
			target = *action.Value
		}
	case policy.CapMax:
		if target > *action.Value {
			target = *action.Value
		}
	case policy.SetValue:
		target = *action.Value
	}
	if target == current {
		return nil
	}
	reason, err := resolve(action.ReasonKey)
	if err != nil {
		return err
	}
	s.evidence = append(s.evidence, item(ruleID, reason, action.Dimension, roundOne(target-current), strength, reason, propositionIDs))
	return nil
}

func appendPattern(result *domain.AnalysisResult, pattern string) {
	if !slices.Contains(result.Patterns, pattern) {
		result.Patterns = append(result.Patterns, pattern)
	}
}
