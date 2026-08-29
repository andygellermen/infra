package rules

import (
	"os"
	"strings"
	"testing"
)

func TestDecodeCanonicalRuleFixture(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_rule-contract-fixture_v0.5.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	definition, err := DecodeDefinition(data)
	if err != nil {
		t.Fatalf("DecodeDefinition() error: %v", err)
	}
	if definition.ContractVersion != ContractVersion || definition.Key != "R-RESPECTFUL-BOUNDARY-V2" {
		t.Fatalf("unexpected definition: %+v", definition)
	}
}

func TestDecodeDefinitionRejectsUnknownField(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_rule-contract-fixture_v0.5.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	invalid := strings.Replace(string(data), `"priority": 100`, `"priority": 100, "calibration_magic": true`, 1)
	_, err = DecodeDefinition([]byte(invalid))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("DecodeDefinition() error = %v; want unknown field rejection", err)
	}
}

func TestRuleV04CannotUseV05ConstructFacts(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_rule-contract-fixture_v0.5.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	legacy := strings.Replace(string(data), `"contract_version": "0.5"`, `"contract_version": "0.4"`, 1)
	if _, err := DecodeDefinition([]byte(legacy)); err == nil || !strings.Contains(err.Error(), "unsupported predicate") {
		t.Fatalf("Rule v0.4 composition predicate error = %v", err)
	}
}
