package importschema_test

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/managedimport"
)

func TestManagedImportSchemasAreStrictAndMatchRuntimeShapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		file  string
		value any
	}{
		{"sprach-a-lyzer_managed-import-request_v0.1.json", managedimport.PrepareRequest{}},
		{"sprach-a-lyzer_managed-import-plan_v0.1.json", managedimport.Plan{}},
		{"sprach-a-lyzer_managed-import-operation_v0.1.json", managedimport.OperationResult{}},
	}
	for _, test := range tests {
		data, err := os.ReadFile(test.file)
		if err != nil {
			t.Fatal(err)
		}
		var schema struct {
			AdditionalProperties bool           `json:"additionalProperties"`
			Properties           map[string]any `json:"properties"`
		}
		if err := json.Unmarshal(data, &schema); err != nil {
			t.Fatal(err)
		}
		if schema.AdditionalProperties {
			t.Errorf("%s is not strict", test.file)
		}
		want := keys(schema.Properties)
		got := fieldKeys(reflect.TypeOf(test.value))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s runtime=%v schema=%v", test.file, got, want)
		}
	}
}
func keys(values map[string]any) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
func fieldKeys(value reflect.Type) []string {
	result := []string{}
	for index := 0; index < value.NumField(); index++ {
		name := strings.Split(value.Field(index).Tag.Get("json"), ",")[0]
		if name != "" && name != "-" {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}
