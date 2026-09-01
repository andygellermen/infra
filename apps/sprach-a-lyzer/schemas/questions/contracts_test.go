package questionschema_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/questions"
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
		"sprach-a-lyzer_question-catalogue_v0.1.json",
		"sprach-a-lyzer_question-answer-observation_v0.1.json",
		"sprach-a-lyzer_question-selection_v0.1.json",
		"sprach-a-lyzer_question-session_v0.1.json",
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

func TestQuestionRuntimeTopLevelShapesMatchSchemas(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filename string
		value    any
	}{
		{"sprach-a-lyzer_question-catalogue_v0.1.json", questions.Catalogue{}},
		{"sprach-a-lyzer_question-answer-observation_v0.1.json", questions.Observation{}},
		{"sprach-a-lyzer_question-selection_v0.1.json", questions.Selection{}},
		{"sprach-a-lyzer_question-session_v0.1.json", questions.Session{}},
	}
	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			schemaKeys := schemaPropertyKeys(t, test.filename)
			runtimeKeys := jsonFieldKeys(reflect.TypeOf(test.value))
			if !reflect.DeepEqual(runtimeKeys, schemaKeys) {
				t.Fatalf("runtime fields = %v; schema fields = %v", runtimeKeys, schemaKeys)
			}
		})
	}
}

func TestApprovedCatalogueUsesCanonicalAndRenderingShapes(t *testing.T) {
	t.Parallel()
	data, err := os.Open("../../data/seed/sprach-a-lyzer_question-catalogue_v0.1.json")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	catalogue, err := questions.DecodeCatalogue(data)
	if err != nil {
		t.Fatal(err)
	}
	canonicalKeys := schemaPropertyKeys(t, "sprach-a-lyzer_question-canonical_v0.1.json")
	renderingKeys := schemaPropertyKeys(t, "sprach-a-lyzer_question-rendering_v0.1.json")
	if got := jsonFieldKeys(reflect.TypeOf(catalogue.Questions[0])); !reflect.DeepEqual(got, canonicalKeys) {
		t.Fatalf("canonical runtime fields = %v; schema fields = %v", got, canonicalKeys)
	}
	if got := jsonFieldKeys(reflect.TypeOf(catalogue.Renderings[0])); !reflect.DeepEqual(got, renderingKeys) {
		t.Fatalf("rendering runtime fields = %v; schema fields = %v", got, renderingKeys)
	}
}

func schemaPropertyKeys(t *testing.T, filename string) []string {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(schema.Properties))
	for key := range schema.Properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func jsonFieldKeys(value reflect.Type) []string {
	keys := make([]string, 0, value.NumField())
	for index := 0; index < value.NumField(); index++ {
		name := value.Field(index).Tag.Get("json")
		if comma := bytes.IndexByte([]byte(name), ','); comma >= 0 {
			name = name[:comma]
		}
		if name != "" && name != "-" {
			keys = append(keys, name)
		}
	}
	sort.Strings(keys)
	return keys
}
