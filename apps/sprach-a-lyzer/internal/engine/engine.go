package engine

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/resolver"
)

var ErrEmptyText = errors.New("analysis text must not be empty")

type Engine struct {
	catalogue CatalogueProvider
	texts     TextProvider
	resolver  contextResolver
}

type contextResolver interface {
	Resolve(domain.AnalysisRequest) (domain.ResolverResult, error)
}

// New retains the catalogue-only test seam and uses the embedded presentation
// bundle. Production composition supplies both providers explicitly.
func New(catalogue CatalogueProvider) *Engine {
	return NewWithProviders(catalogue, staticTextProvider(embeddedFoundation()))
}

func NewWithProviders(catalogue CatalogueProvider, texts TextProvider) *Engine {
	return &Engine{catalogue: catalogue, texts: texts, resolver: resolver.New()}
}

func (e *Engine) Resolve(request domain.AnalysisRequest) (domain.ResolverResult, error) {
	if e.resolver == nil {
		return domain.ResolverResult{}, fmt.Errorf("context resolver is nil")
	}
	return e.resolver.Resolve(request)
}

type evidence struct {
	contribution   domain.ContributionTraceEntry
	strength       float64
	propositionIDs []string
}

func (e *Engine) Analyze(request domain.AnalysisRequest) (domain.AnalysisResult, error) {
	text := strings.TrimSpace(request.Text)
	if text == "" {
		return domain.AnalysisResult{}, ErrEmptyText
	}
	contextValue := domain.AnalysisContext(strings.ToUpper(strings.TrimSpace(string(request.Context))))
	if contextValue == "" {
		contextValue = domain.ContextUnspecified
	}
	inputMode := domain.InputMode(strings.ToUpper(strings.TrimSpace(string(request.InputMode))))
	if inputMode == "" {
		inputMode = domain.InputModeText
	}
	profile := strings.ToUpper(strings.TrimSpace(string(request.PresentationProfile)))
	if profile == "" {
		profile = string(domain.ProfilePrivate)
	}
	locale := strings.TrimSpace(string(request.Locale))
	if locale == "" {
		locale = string(domain.LocaleGerman)
	}
	request.Context, request.InputMode = contextValue, inputMode

	result := domain.AnalysisResult{
		Text: text, Context: contextValue, InputMode: inputMode, Propositions: propositions(text),
		ResolvedSenses: []domain.ResolvedSense{}, Patterns: []string{}, Dimensions: emptyDimensions(),
		ContributionTrace: []domain.ContributionTraceEntry{}, Alternatives: []string{},
		ResonanceHints: []domain.ResonanceHint{}, Notes: []string{},
	}
	resolution, err := e.Resolve(request)
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("resolve context and propositions: %w", err)
	}
	for _, node := range resolution.PropositionGraph.Nodes {
		result.TraceProvenance.Propositions = append(result.TraceProvenance.Propositions, domain.TraceProposition{
			ID: node.ID, Text: node.Text, SourceStart: node.SourceStart, SourceEnd: node.SourceEnd,
			TargetType: node.TargetType, ExpectationSource: node.ExpectationSource,
		})
	}
	definitions, facts, err := e.activeDefinitions(request, normalize(text), resolution)
	if err != nil {
		return domain.AnalysisResult{}, err
	}
	if e.texts == nil {
		return domain.AnalysisResult{}, fmt.Errorf("presentation text provider is nil")
	}
	texts, err := e.texts.Texts(context.Background(), profile, locale)
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("load presentation texts: %w", err)
	}
	state := executionState{
		result: &result, nonAssessable: make(map[domain.DimensionID]bool), texts: texts,
		patternReferences: make(map[string][]string), senseReferences: make(map[string][]string),
	}
	executed := make(map[string]bool, len(definitions))
	for pass := 0; pass < len(definitions) && !state.stop; pass++ {
		progress := false
		for _, definition := range definitions {
			if !definition.Enabled || executed[definition.Key] {
				continue
			}
			match, err := evaluateCondition(definition.Condition, facts)
			if err != nil {
				return domain.AnalysisResult{}, fmt.Errorf("evaluate rule %s: %w", definition.Key, err)
			}
			if !match.matched {
				continue
			}
			if err := state.execute(definition, match.propositionIDs); err != nil {
				return domain.AnalysisResult{}, err
			}
			executed[definition.Key], progress = true, true
			facts.patterns = append([]string(nil), result.Patterns...)
			facts.references["pattern"] = nil
			for _, pattern := range result.Patterns {
				facts.addReference("pattern", pattern, state.patternReferences[pattern]...)
			}
			facts.senses = trustedResolverSenses(resolution.SelectedSenses)
			facts.references["selected_sense"] = nil
			for _, sense := range resolution.SelectedSenses {
				if sense.State != domain.SenseAmbiguous {
					facts.addReference("selected_sense", sense.Sense, sense.PropositionID)
				}
			}
			for _, sense := range result.ResolvedSenses {
				facts.senses = append(facts.senses, sense.Sense)
				facts.addReference("selected_sense", sense.Sense, state.senseReferences[sense.Sense]...)
			}
			if state.stop {
				break
			}
		}
		if !progress {
			break
		}
	}
	applyEvidence(&result, state.evidence, state.nonAssessable)
	return result, nil
}

func item(ruleID, matched string, dimension domain.DimensionID, delta, strength float64, reason string, propositionIDs []string) evidence {
	return evidence{
		contribution: domain.ContributionTraceEntry{RuleID: ruleID, Evidence: matched, Dimension: dimension, Delta: delta, Reason: reason},
		strength:     strength, propositionIDs: append([]string(nil), propositionIDs...),
	}
}

func applyEvidence(result *domain.AnalysisResult, items []evidence, nonAssessable map[domain.DimensionID]bool) {
	scores, strengths := make(map[domain.DimensionID]float64), make(map[domain.DimensionID]float64)
	for _, item := range items {
		if nonAssessable[item.contribution.Dimension] {
			continue
		}
		result.ContributionTrace = append(result.ContributionTrace, item.contribution)
		result.TraceProvenance.ContributionPropositionIDs = append(
			result.TraceProvenance.ContributionPropositionIDs,
			append([]string(nil), item.propositionIDs...),
		)
		dimension := item.contribution.Dimension
		scores[dimension] += item.contribution.Delta
		if item.strength > strengths[dimension] {
			strengths[dimension] = item.strength
		}
	}
	for dimension, delta := range scores {
		score := roundOne(math.Max(0, math.Min(100, 50+delta)))
		strength := strengths[dimension]
		result.Dimensions[dimension] = domain.DimensionResult{
			State: stateFor(strength), Score: &score, Confidence: roundTwo(math.Min(0.98, 0.50+strength*0.60)), Assessability: strength,
		}
	}
}

func emptyDimensions() map[domain.DimensionID]domain.DimensionResult {
	result := make(map[domain.DimensionID]domain.DimensionResult, len(domain.CanonicalDimensions()))
	for _, dimension := range domain.CanonicalDimensions() {
		result[dimension] = domain.DimensionResult{State: domain.NotAssessable, Score: nil}
	}
	return result
}

func stateFor(assessability float64) domain.AssessabilityState {
	switch {
	case assessability >= .80:
		return domain.Strong
	case assessability >= .51:
		return domain.Assessable
	case assessability >= .35:
		return domain.Weak
	default:
		return domain.NotAssessable
	}
}

func normalize(text string) string { return strings.ToLower(strings.Join(strings.Fields(text), " ")) }

func propositions(text string) []domain.Proposition {
	parts := strings.FieldsFunc(text, func(r rune) bool { return r == '.' || r == '!' || r == '?' })
	result := make([]domain.Proposition, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		proposition := domain.Proposition{ID: fmt.Sprintf("P%d", len(result)), Text: part}
		if strings.Contains(strings.ToLower(part), "trotzdem") {
			proposition.Relation = "CONCESSION"
		}
		result = append(result, proposition)
	}
	return result
}

func roundOne(value float64) float64   { return math.Round(value*10) / 10 }
func roundTwo(value float64) float64   { return math.Round(value*100) / 100 }
func roundThree(value float64) float64 { return math.Round(value*1000) / 1000 }
