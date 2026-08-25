package seed

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/rules"
)

func TestDecodeFoundation(t *testing.T) {
	t.Parallel()

	input := `{"version":"0.1","dimensions":[
{"id":"AGENCY"},{"id":"CONNECTION"},{"id":"APPRECIATION"},
{"id":"CLARITY"},{"id":"VOLITION"},{"id":"OPENNESS"}]}`
	foundation, err := DecodeFoundation(strings.NewReader(input))
	if err != nil {
		t.Fatalf("DecodeFoundation() error: %v", err)
	}
	if len(foundation.Dimensions) != 6 || foundation.Dimensions[4].ID != "VOLITION" {
		t.Fatalf("unexpected foundation: %+v", foundation)
	}
}

func TestCanonicalFoundationKeepsPresentationBundlesSeparate(t *testing.T) {
	t.Parallel()

	file, err := os.Open("../../data/seed/sprach-a-lyzer_foundation_v0.2.json")
	if err != nil {
		t.Fatalf("open canonical foundation: %v", err)
	}
	defer file.Close()
	foundation, err := DecodeFoundation(file)
	if err != nil {
		t.Fatalf("DecodeFoundation() error: %v", err)
	}
	if len(foundation.PresentationBundles) != 2 {
		t.Fatalf("presentation bundles = %d; want 2", len(foundation.PresentationBundles))
	}
	if foundation.Version != "0.2" || foundation.RuleSet.Version != "0.2" || len(foundation.Rules) != 6 {
		t.Fatalf("foundation/rule-set migration incomplete: version=%s rule_set=%s rules=%d", foundation.Version, foundation.RuleSet.Version, len(foundation.Rules))
	}
	for _, rule := range foundation.Rules {
		if rule.ContractVersion != rules.ContractVersion {
			t.Errorf("rule %s contract = %q; want %q", rule.Key, rule.ContractVersion, rules.ContractVersion)
		}
		for _, action := range rule.Actions {
			if !slices.Contains(policy.RuleActionTypes(), action.Type) {
				t.Errorf("rule %s contains unregistered action %q", rule.Key, action.Type)
			}
			encoded, _ := json.Marshal(action)
			if strings.Contains(string(encoded), "FREE_WILL") || strings.Contains(string(encoded), "semantic_dimension_contribution") {
				t.Errorf("rule %s contains legacy action data: %s", rule.Key, encoded)
			}
		}
	}
}

func TestDecodeFoundationRejectsScoringResonanceHint(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_foundation_v0.2.json")
	if err != nil {
		t.Fatalf("read canonical foundation: %v", err)
	}
	invalid := strings.Replace(string(data), `"semantic_score": false`, `"semantic_score": true`, 1)
	_, err = DecodeFoundation(strings.NewReader(invalid))
	if err == nil || !strings.Contains(err.Error(), "semantic_score=false") {
		t.Fatalf("DecodeFoundation() error = %v; want resonance guardrail rejection", err)
	}
}

func TestDecodeFoundationRejectsLegacyRuleShape(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_foundation_v0.2.json")
	if err != nil {
		t.Fatalf("read canonical foundation: %v", err)
	}
	invalid := strings.Replace(string(data), `"type": "ADD_PATTERN", "key": "INTERNAL_PRESSURE"`, `"pattern": "INTERNAL_PRESSURE"`, 1)
	_, err = DecodeFoundation(strings.NewReader(invalid))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("DecodeFoundation() error = %v; want legacy shape rejection", err)
	}
}

func TestDecodeFoundationMapsLegacyDimensionSet(t *testing.T) {
	t.Parallel()

	input := `{"version":"0.1","dimensions":[
{"id":"AGENCY"},{"id":"CONNECTION"},{"id":"APPRECIATION"},
{"id":"CLARITY"},{"id":"FREE_WILL"},{"id":"OPENNESS"}]}`
	foundation, report, err := DecodeFoundationWithReport(strings.NewReader(input))
	if err != nil {
		t.Fatalf("legacy foundation rejected: %v", err)
	}
	if foundation.Dimensions[4].ID != "VOLITION" {
		t.Fatalf("legacy dimension decoded as %q; want VOLITION", foundation.Dimensions[4].ID)
	}
	if report.LegacyCount() != 1 {
		t.Fatalf("legacy mappings = %d; want 1", report.LegacyCount())
	}
}
