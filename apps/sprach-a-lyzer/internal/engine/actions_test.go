package engine

import (
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/rules"
)

func TestGenericContributionModifiers(t *testing.T) {
	t.Parallel()

	value, confidence, factor, cap := 10.0, .5, 2.0, 15.0
	result := domain.AnalysisResult{Patterns: []string{}, Alternatives: []string{}, Notes: []string{}}
	state := executionState{
		result: &result, nonAssessable: map[domain.DimensionID]bool{},
		texts: map[string]string{"EVIDENCE": "evidence", "REASON": "reason"},
	}
	definition := rules.Definition{Key: "R-MODIFIER", ConfidenceModifier: 1, Actions: []rules.Action{
		{Type: policy.AddContribution, Dimension: domain.DimensionVolition, Value: &value, Confidence: &confidence, EvidenceKey: "EVIDENCE", ReasonKey: "REASON"},
		{Type: policy.MultiplyContribution, Dimension: domain.DimensionVolition, Factor: &factor, ReasonKey: "REASON"},
		{Type: policy.CapMax, Dimension: domain.DimensionVolition, Value: &cap, ReasonKey: "REASON"},
	}}
	if err := state.execute(definition, []string{"P0"}); err != nil {
		t.Fatalf("execute() error: %v", err)
	}
	total := 0.0
	for _, entry := range state.evidence {
		total += entry.contribution.Delta
	}
	if total != 15 {
		t.Fatalf("effective contribution = %v; want 15", total)
	}

	state.evidence = append(state.evidence, item("R-X", "e", domain.DimensionClarity, 7, .5, "r", []string{"P0"}))
	if err := state.execute(rules.Definition{Key: "R-GUARD", Actions: []rules.Action{
		{Type: policy.Invert, Dimension: domain.DimensionVolition, ReasonKey: "REASON"},
		{Type: policy.Suppress, Dimension: domain.DimensionVolition, ReasonKey: "REASON"},
		{Type: policy.MarkNonAssessable, Dimension: domain.DimensionClarity, ReasonKey: "REASON"},
	}}, nil); err != nil {
		t.Fatalf("execute guard actions: %v", err)
	}
	if len(state.evidence) != 0 || !state.nonAssessable[domain.DimensionClarity] {
		t.Fatalf("suppress/non-assessable state = %+v/%+v", state.evidence, state.nonAssessable)
	}
}
