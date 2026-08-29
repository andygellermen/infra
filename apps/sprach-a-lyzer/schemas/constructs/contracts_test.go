package constructschema_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

func TestOntologySchemaAndSeedLockCanonicalIDsAndGuardrails(t *testing.T) {
	t.Parallel()
	schema := readObject(t, "sprach-a-lyzer_construct-ontology_v0.2.json")
	seed := readObject(t, "../../data/seed/sprach-a-lyzer_construct-ontology_v0.2.json")
	if seed["ontology_version"] != "0.2" || seed["policy_registry"] != "0.6" {
		t.Fatalf("ontology version vector = %v/%v", seed["ontology_version"], seed["policy_registry"])
	}
	definitions := object(t, schema["$defs"])
	constructID := object(t, definitions["constructID"])
	wantIDs := make([]any, 0, len(policy.Constructs()))
	for _, id := range policy.Constructs() {
		wantIDs = append(wantIDs, string(id))
	}
	if !reflect.DeepEqual(constructID["enum"], wantIDs) {
		t.Fatalf("schema construct IDs = %#v; policy = %#v", constructID["enum"], wantIDs)
	}
	entries := array(t, seed["constructs"])
	if len(entries) != len(wantIDs) {
		t.Fatalf("seed constructs = %d; want %d", len(entries), len(wantIDs))
	}
	seen := map[string]bool{}
	for _, raw := range entries {
		entry := object(t, raw)
		id := entry["id"].(string)
		if seen[id] || entry["core_scoring"] != false {
			t.Errorf("duplicate or scoring construct %s", id)
		}
		seen[id] = true
	}
	wantGuardrails := object(t, object(t, schema["properties"])["hard_guardrails"])["const"]
	if !reflect.DeepEqual(seed["hard_guardrails"], wantGuardrails) {
		t.Fatalf("ontology guardrails = %#v; schema = %#v", seed["hard_guardrails"], wantGuardrails)
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
