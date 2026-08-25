package ruleschema_test

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/dimension"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

func TestRuleV03UsesCanonicalDimensionsAndActionIDs(t *testing.T) {
	t.Parallel()

	schema := readObject(t, "sprach-a-lyzer_rule_v0.3.json")
	definitions := object(t, schema["$defs"])
	dimensionDefinition := object(t, definitions["dimensionID"])
	gotDimensions := stringArray(t, dimensionDefinition["enum"])
	wantDimensions := make([]string, 0, len(dimension.All()))
	for _, id := range dimension.All() {
		wantDimensions = append(wantDimensions, string(id))
	}
	if !slices.Equal(gotDimensions, wantDimensions) {
		t.Errorf("rule schema dimensions = %v; want %v", gotDimensions, wantDimensions)
	}

	actionTypes := []string{}
	for _, definitionName := range []string{"addContribution", "factorAction", "valueAction", "dimensionOnlyAction", "keyAction", "selectSenseAction", "resonanceHintAction", "terminalAction"} {
		definition := object(t, definitions[definitionName])
		properties := object(t, definition["properties"])
		typeProperty := object(t, properties["type"])
		if value, exists := typeProperty["const"]; exists {
			actionTypes = append(actionTypes, value.(string))
		}
		if values, exists := typeProperty["enum"]; exists {
			actionTypes = append(actionTypes, stringArray(t, values)...)
		}
	}
	wantActions := make([]string, 0, len(policy.RuleActionTypes()))
	for _, id := range policy.RuleActionTypes() {
		wantActions = append(wantActions, string(id))
	}
	slices.Sort(actionTypes)
	slices.Sort(wantActions)
	if !slices.Equal(actionTypes, wantActions) {
		t.Errorf("rule schema actions = %v; policy = %v", actionTypes, wantActions)
	}
	data, _ := json.Marshal(schema)
	if bytes.Contains(data, []byte("FREE_WILL")) {
		t.Error("rule v0.3 schema contains legacy dimension ID")
	}
}

func TestCanonicalRuleFixtureUsesOnlyRegisteredActions(t *testing.T) {
	t.Parallel()

	fixture := readObject(t, "../../data/seed/sprach-a-lyzer_rule-contract-fixture_v0.3.json")
	if fixture["contract_version"] != "0.3" {
		t.Fatalf("fixture contract version = %#v; want 0.3", fixture["contract_version"])
	}
	allowed := map[string]bool{}
	for _, action := range policy.RuleActionTypes() {
		allowed[string(action)] = true
	}
	for _, raw := range array(t, fixture["actions"]) {
		action := object(t, raw)
		if !allowed[action["type"].(string)] {
			t.Errorf("fixture uses unknown action %q", action["type"])
		}
		if action["dimension"] == "FREE_WILL" {
			t.Error("fixture uses legacy dimension ID")
		}
	}
}

func TestPolicyRegistrySchemaAndSeedAreJSON(t *testing.T) {
	t.Parallel()
	readObject(t, "sprach-a-lyzer_policy-registry_v0.2.json")
	readObject(t, "sprach-a-lyzer_parameter_v0.1.json")
	readObject(t, "../../data/seed/sprach-a-lyzer_policy-registry_v0.2.json")
	fixture := readObject(t, "../../data/seed/sprach-a-lyzer_parameter-contract-fixture_v0.1.json")
	for _, raw := range array(t, fixture["parameters"]) {
		parameter := object(t, raw)
		if parameter["editable"] == true && parameter["requires_approval"] != true {
			t.Errorf("editable parameter %q does not require approval", parameter["key"])
		}
		if minimum, ok := parameter["minimum"].(float64); ok {
			maximum := parameter["maximum"].(float64)
			value := parameter["default_value"].(float64)
			if minimum > maximum || value < minimum || value > maximum {
				t.Errorf("parameter %q has invalid bounds/default", parameter["key"])
			}
		}
	}
}

func readObject(t *testing.T, filename string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("decode %s: %v", filename, err)
	}
	return result
}

func object(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is not object: %#v", value)
	}
	return result
}

func array(t *testing.T, value any) []any {
	t.Helper()
	result, ok := value.([]any)
	if !ok {
		t.Fatalf("value is not array: %#v", value)
	}
	return result
}

func stringArray(t *testing.T, value any) []string {
	t.Helper()
	values := array(t, value)
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = value.(string)
	}
	return result
}
