package golden_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

const propositionTraceGoldenPath = "../../data/golden/sprach-a-lyzer_proposition-trace_v0.1.json"

type propositionTraceSuite struct {
	Version       string `json:"version"`
	TraceContract string `json:"trace_contract"`
	Cases         []struct {
		ID                   string           `json:"id"`
		Request              analysis.Request `json:"request"`
		ExpectedPropositions []struct {
			ID                string                       `json:"id"`
			TargetType        analysis.TargetTypeID        `json:"target_type"`
			ExpectationSource analysis.ExpectationSourceID `json:"expectation_source"`
		} `json:"expected_propositions"`
		ExpectedContributionRefs []struct {
			RuleID         string   `json:"rule_id"`
			PropositionIDs []string `json:"proposition_ids"`
		} `json:"expected_contribution_refs"`
	} `json:"cases"`
}

func TestPropositionTraceGolden(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(propositionTraceGoldenPath)
	if err != nil {
		t.Fatalf("read proposition trace golden: %v", err)
	}
	var suite propositionTraceSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("decode proposition trace golden: %v", err)
	}
	if suite.Version != "0.1" || suite.TraceContract != "0.2" || len(suite.Cases) != 2 {
		t.Fatalf("unexpected proposition trace suite: version=%q trace=%q cases=%d", suite.Version, suite.TraceContract, len(suite.Cases))
	}
	analyzer := analysis.NewDefault()
	for _, testCase := range suite.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			result, err := analyzer.Analyze(testCase.Request)
			if err != nil {
				t.Fatalf("Analyze() error: %v", err)
			}
			trace := result.TraceV02()
			if trace.ContractVersion != suite.TraceContract || len(trace.Propositions) != len(testCase.ExpectedPropositions) {
				t.Fatalf("trace envelope/propositions = %+v", trace)
			}
			for index, want := range testCase.ExpectedPropositions {
				got := trace.Propositions[index]
				if got.ID != want.ID || got.TargetType != want.TargetType || got.ExpectationSource != want.ExpectationSource {
					t.Errorf("proposition %d = %+v; want %+v", index, got, want)
				}
			}
			if len(trace.Contributions) != len(testCase.ExpectedContributionRefs) {
				t.Fatalf("contributions = %d; want %d", len(trace.Contributions), len(testCase.ExpectedContributionRefs))
			}
			for index, want := range testCase.ExpectedContributionRefs {
				got := trace.Contributions[index]
				if got.RuleID != want.RuleID || !slices.Equal(got.PropositionIDs, want.PropositionIDs) {
					t.Errorf("contribution %d = %s/%v; want %s/%v", index, got.RuleID, got.PropositionIDs, want.RuleID, want.PropositionIDs)
				}
			}
		})
	}
}
