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

func TestRuleV05UsesCanonicalDimensionsAndActionIDs(t *testing.T) {
	t.Parallel()

	schema := readObject(t, "sprach-a-lyzer_rule_v0.5.json")
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
		t.Error("rule v0.5 schema contains legacy dimension ID")
	}
	predicate := object(t, definitions["predicate"])
	fields := stringArray(t, object(t, object(t, predicate["properties"])["field"])["enum"])
	if !slices.Contains(fields, "construct") || !slices.Contains(fields, "composition") {
		t.Fatalf("Rule v0.5 fields = %v; want construct and composition", fields)
	}
}

func TestCanonicalRuleFixtureUsesOnlyRegisteredActions(t *testing.T) {
	t.Parallel()

	fixture := readObject(t, "../../data/seed/sprach-a-lyzer_rule-contract-fixture_v0.5.json")
	if fixture["contract_version"] != "0.5" {
		t.Fatalf("fixture contract version = %#v; want 0.5", fixture["contract_version"])
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

func TestPolicyRegistryV07SchemaAndSeedAreLocked(t *testing.T) {
	t.Parallel()
	schema := readObject(t, "sprach-a-lyzer_policy-registry_v0.7.json")
	readObject(t, "sprach-a-lyzer_parameter_v0.1.json")
	seed := readObject(t, "../../data/seed/sprach-a-lyzer_policy-registry_v0.7.json")
	if seed["registry_version"] != policy.RegistryVersion {
		t.Fatalf("registry version = %#v; want %s", seed["registry_version"], policy.RegistryVersion)
	}
	schemaProperties := object(t, schema["properties"])
	seedCanonicalIDs := object(t, seed["canonical_ids"])
	schemaCanonicalIDs := object(t, object(t, schemaProperties["canonical_ids"])["properties"])
	for key, seedValue := range seedCanonicalIDs {
		if !jsonEqual(seedValue, object(t, schemaCanonicalIDs[key])["const"]) {
			t.Errorf("canonical IDs %q differ between registry seed and schema", key)
		}
	}
	definitions := object(t, schema["$defs"])
	guardrailDefinition := object(t, definitions["hardGuardrail"])
	guardrailProperties := object(t, guardrailDefinition["properties"])
	wantGuardrails := stringArray(t, object(t, guardrailProperties["id"])["enum"])
	gotGuardrails := make([]string, 0, len(wantGuardrails))
	for _, raw := range array(t, seed["hard_guardrails"]) {
		guardrail := object(t, raw)
		gotGuardrails = append(gotGuardrails, guardrail["id"].(string))
		if guardrail["editable"] != false {
			t.Errorf("hard guardrail %q must be immutable", guardrail["id"])
		}
	}
	if !slices.Equal(gotGuardrails, wantGuardrails) {
		t.Errorf("hard guardrails differ between registry seed and schema")
	}

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

func jsonEqual(left, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return bytes.Equal(leftJSON, rightJSON)
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
