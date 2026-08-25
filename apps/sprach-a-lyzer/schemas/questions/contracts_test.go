package questionschema_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

type questionFixture struct {
	CanonicalQuestions []map[string]any `json:"canonical_questions"`
	Renderings         []map[string]any `json:"renderings"`
}

func TestCanonicalQuestionAndRenderingStaySeparated(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_question-contract-fixture_v0.1.json")
	if err != nil {
		t.Fatalf("read question fixture: %v", err)
	}
	var fixture questionFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode question fixture: %v", err)
	}
	if len(fixture.CanonicalQuestions) == 0 || len(fixture.Renderings) == 0 {
		t.Fatal("question fixture must contain canonical questions and renderings")
	}

	questionIDs := map[string]bool{}
	for _, question := range fixture.CanonicalQuestions {
		if _, exists := question["text"]; exists {
			t.Error("canonical question must not contain visible text")
		}
		if bias, ok := question["question_score_bias"].(float64); !ok || bias != 0 {
			t.Errorf("canonical question bias = %#v; want 0", question["question_score_bias"])
		}
		questionIDs[question["question_id"].(string)] = true
	}
	for _, rendering := range fixture.Renderings {
		if _, exists := rendering["construct_intent"]; exists {
			t.Error("rendering must not redefine canonical construct intent")
		}
		if !questionIDs[rendering["question_id"].(string)] {
			t.Errorf("rendering references unknown question %q", rendering["question_id"])
		}
		if rendering["text"] == "" {
			t.Error("rendering must contain visible text")
		}
	}
}

func TestQuestionSchemasAreStrictAndDoNotContainLegacyDimension(t *testing.T) {
	t.Parallel()

	for _, filename := range []string{
		"sprach-a-lyzer_question-canonical_v0.1.json",
		"sprach-a-lyzer_question-rendering_v0.1.json",
	} {
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("read %s: %v", filename, err)
		}
		var schema map[string]any
		if err := json.Unmarshal(data, &schema); err != nil {
			t.Fatalf("decode %s: %v", filename, err)
		}
		if strict, ok := schema["additionalProperties"].(bool); !ok || strict {
			t.Errorf("%s must reject unknown top-level properties", filename)
		}
		if bytes.Contains(data, []byte("FREE_WILL")) {
			t.Errorf("%s contains legacy dimension ID", filename)
		}
	}
}
