package golden_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

type constructCompositionSuite struct {
	Version          string `json:"version"`
	OntologyContract string `json:"ontology_contract"`
	RuleContract     string `json:"rule_contract"`
	Cases            []struct {
		ID                 string           `json:"id"`
		Request            analysis.Request `json:"request"`
		ExpectedPattern    string           `json:"expected_pattern"`
		CompositionRuleID  string           `json:"composition_rule_id"`
		ExpectedConstructs []struct {
			ID             policy.ConstructID `json:"id"`
			PropositionIDs []string           `json:"proposition_ids"`
		} `json:"expected_constructs"`
		ExpectedCompositionPropositionIDs []string `json:"expected_composition_proposition_ids"`
		ExpectedCompositionContributions  int      `json:"expected_composition_contributions"`
	} `json:"cases"`
}

func TestConstructCompositionRuntimeGolden(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/golden/sprach-a-lyzer_construct-composition-runtime_v0.1.json")
	if err != nil {
		t.Fatalf("read construct composition golden: %v", err)
	}
	var suite constructCompositionSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("decode construct composition golden: %v", err)
	}
	if suite.Version != "0.1" || suite.OntologyContract != "0.2" || suite.RuleContract != "0.5" || len(suite.Cases) != 3 {
		t.Fatalf("unexpected suite envelope: %+v", suite)
	}
	analyzer := analysis.NewDefault()
	for _, testCase := range suite.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			result, err := analyzer.Analyze(testCase.Request)
			if err != nil {
				t.Fatalf("Analyze() error: %v", err)
			}
			if !slices.Contains(result.Patterns, testCase.ExpectedPattern) {
				t.Fatalf("patterns %v lack %s", result.Patterns, testCase.ExpectedPattern)
			}
			for _, want := range testCase.ExpectedConstructs {
				found := false
				for _, got := range result.TraceProvenance.ConstructEvidence {
					if got.ConstructID == want.ID && slices.Equal(got.PropositionIDs, want.PropositionIDs) {
						found = true
					}
				}
				if !found {
					t.Errorf("construct evidence %+v lacks %s/%v", result.TraceProvenance.ConstructEvidence, want.ID, want.PropositionIDs)
				}
			}
			count := 0
			for _, contribution := range result.TraceV02().Contributions {
				if contribution.RuleID == testCase.CompositionRuleID {
					count++
					if !slices.Equal(contribution.PropositionIDs, testCase.ExpectedCompositionPropositionIDs) {
						t.Errorf("composition contribution IDs = %v; want %v", contribution.PropositionIDs, testCase.ExpectedCompositionPropositionIDs)
					}
				}
			}
			if count != testCase.ExpectedCompositionContributions {
				t.Errorf("composition contributions = %d; want %d", count, testCase.ExpectedCompositionContributions)
			}
		})
	}
}
