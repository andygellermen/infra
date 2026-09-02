package experience_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/experience"
)

func TestMVPExperienceSchemasAreStrictAndMatchRuntimeShapes(t *testing.T) {
	t.Parallel()
	request := readSchema(t, "sprach-a-lyzer_mvp-experience-request_v0.1.json")
	result := readSchema(t, "sprach-a-lyzer_mvp-experience-result_v0.1.json")
	if request["additionalProperties"] != false || result["additionalProperties"] != false {
		t.Fatal("MVP experience top-level schemas must reject unknown fields")
	}
	assertProperties(t, request, reflect.TypeOf(experience.Request{}))
	assertProperties(t, result, reflect.TypeOf(experience.Result{}))
	encoded, _ := json.Marshal(result)
	text := string(encoded)
	for _, guardrail := range []string{`"raw_text_stored":{"const":false}`, `"external_transfer":{"const":false}`, `"ai_used":{"const":false}`, `"experience_mode":{"const":"CORE_NO_AI"}`} {
		if !strings.Contains(text, guardrail) {
			t.Fatalf("result schema lacks guardrail %s", guardrail)
		}
	}
}

func readSchema(t *testing.T, name string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func assertProperties(t *testing.T, schema map[string]any, runtime reflect.Type) {
	t.Helper()
	properties := schema["properties"].(map[string]any)
	want := map[string]bool{}
	for index := 0; index < runtime.NumField(); index++ {
		name := strings.Split(runtime.Field(index).Tag.Get("json"), ",")[0]
		if name != "" && name != "-" {
			want[name] = true
		}
	}
	if len(properties) != len(want) {
		t.Fatalf("schema properties=%v runtime=%v", reflect.ValueOf(properties).MapKeys(), want)
	}
	for name := range want {
		if _, ok := properties[name]; !ok {
			t.Fatalf("schema lacks runtime property %q", name)
		}
	}
}
