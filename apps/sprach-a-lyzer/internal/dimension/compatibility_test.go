package dimension

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeJSONCoversLegacyShapes(t *testing.T) {
	t.Parallel()

	input := []byte(`{
  "dimension":"FREE_WILL",
  "expected_dimensions":"AGENCY,FREE_WILL,CLARITY",
  "expected_dimension_direction":"AGENCY:+;FREE_WILL:-",
  "dimensions":{"FREE_WILL":{"score":42}},
  "constructs":{"FREE_WILL":{"expected_dimensions":["FREE_WILL","AGENCY"]}},
  "description":"FREE_WILL remains documentation here"
}`)
	normalized, report, err := NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON() error: %v", err)
	}
	if report.LegacyCount() != 6 {
		t.Fatalf("legacy mappings = %d; want 6 (%+v)", report.LegacyCount(), report)
	}
	var result map[string]any
	if err := json.Unmarshal(normalized, &result); err != nil {
		t.Fatalf("decode normalized result: %v", err)
	}
	if result["dimension"] != "VOLITION" || result["expected_dimension_direction"] != "AGENCY:+;VOLITION:-" {
		t.Fatalf("semantic fields were not canonicalized: %s", normalized)
	}
	if result["description"] != "FREE_WILL remains documentation here" {
		t.Fatalf("non-semantic documentation string was changed: %s", normalized)
	}
}

func TestNormalizeVersionedLegacyArtifacts(t *testing.T) {
	t.Parallel()

	legacyArtifacts := 0
	err := filepath.WalkDir("../../data", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Contains(data, []byte(LegacyFreeWill)) {
			return nil
		}
		legacyArtifacts++

		var normalized []byte
		var report CompatibilityReport
		switch filepath.Ext(path) {
		case ".json":
			normalized, report, err = NormalizeJSON(data)
		case ".csv":
			normalized, report, err = NormalizeCSV(bytes.NewReader(data))
		default:
			return nil
		}
		if err != nil {
			t.Errorf("normalize %s: %v", path, err)
			return nil
		}
		if report.LegacyCount() == 0 || bytes.Contains(normalized, []byte(LegacyFreeWill)) {
			t.Errorf("%s not fully canonicalized; mappings=%d", path, report.LegacyCount())
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk versioned artifacts: %v", err)
	}
	if legacyArtifacts == 0 {
		t.Fatal("no versioned legacy artifact exercised")
	}
}

func TestNormalizeJSONRejectsDimensionKeyCollision(t *testing.T) {
	t.Parallel()

	_, _, err := NormalizeJSON([]byte(`{"dimensions":{"FREE_WILL":1,"VOLITION":2}}`))
	if err == nil {
		t.Fatal("legacy/canonical collision unexpectedly accepted")
	}
}

func TestNormalizeExpression(t *testing.T) {
	t.Parallel()

	got, count := NormalizeExpression("AGENCY:+;FREE_WILL:-")
	if got != "AGENCY:+;VOLITION:-" || count != 1 {
		t.Fatalf("NormalizeExpression() = %q, %d", got, count)
	}
	got, count = NormalizeExpression("FREE_WILLING")
	if got != "FREE_WILLING" || count != 0 {
		t.Fatalf("NormalizeExpression() rewrote unrelated identifier: %q, %d", got, count)
	}
}

func TestNormalizeJSONProducesValidJSON(t *testing.T) {
	t.Parallel()

	got, _, err := NormalizeJSON([]byte(`{"expected_dimensions":["FREE_WILL"]}`))
	if err != nil {
		t.Fatalf("NormalizeJSON() error: %v", err)
	}
	if !json.Valid(got) {
		t.Fatalf("NormalizeJSON() returned invalid JSON: %s", got)
	}
}

func TestNormalizeJSONMapsDimensionReferenceKeys(t *testing.T) {
	t.Parallel()

	input := []byte(`{
  "rules":[{"actions":[{
    "dimension":"FREE_WILL",
    "reason_key":"REASON_INTERNAL_PRESSURE_FREE_WILL"
  }]}],
  "presentation_bundles":[{"entries":{
    "REASON_INTERNAL_PRESSURE_FREE_WILL":"Documentation keeps FREE_WILL wording"
  }}]
}`)
	normalized, report, err := NormalizeJSON(input)
	if err != nil {
		t.Fatalf("NormalizeJSON() error: %v", err)
	}
	if report.LegacyCount() != 3 {
		t.Fatalf("legacy mappings = %d; want 3 (%+v)", report.LegacyCount(), report)
	}
	if bytes.Contains(normalized, []byte(`"reason_key":"REASON_INTERNAL_PRESSURE_FREE_WILL"`)) ||
		bytes.Contains(normalized, []byte(`"REASON_INTERNAL_PRESSURE_FREE_WILL":`)) {
		t.Fatalf("dimension reference key was not canonicalized: %s", normalized)
	}
	if !bytes.Contains(normalized, []byte(`Documentation keeps FREE_WILL wording`)) {
		t.Fatalf("presentation text was unexpectedly rewritten: %s", normalized)
	}
}
