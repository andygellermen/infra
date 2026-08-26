package analysisschema_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

func TestGoContractsMatchJSONSchemaObjectShapes(t *testing.T) {
	t.Parallel()

	request := readSchema(t, "sprach-a-lyzer_analysis-request_v0.1.json")
	result := readSchema(t, "sprach-a-lyzer_analysis-result_v0.1.json")
	trace := readSchema(t, "sprach-a-lyzer_analysis-trace_v0.1.json")
	resolverResult := readSchema(t, "sprach-a-lyzer_resolver-result_v0.2.json")

	assertObjectShape(t, reflect.TypeOf(analysis.Request{}), request)
	assertObjectShape(t, reflect.TypeOf(analysis.Result{}), result)
	assertObjectShape(t, reflect.TypeOf(analysis.Trace{}), trace)
	assertObjectShape(t, reflect.TypeOf(analysis.Proposition{}), definition(t, result, "proposition"))
	assertObjectShape(t, reflect.TypeOf(analysis.ResolvedSense{}), definition(t, result, "resolvedSense"))
	assertObjectShape(t, reflect.TypeOf(analysis.DimensionResult{}), definition(t, result, "dimensionResult"))
	assertObjectShape(t, reflect.TypeOf(analysis.ContributionTraceEntry{}), definition(t, result, "contribution"))
	assertObjectShape(t, reflect.TypeOf(analysis.ResonanceHint{}), definition(t, result, "resonanceHint"))
	assertObjectShape(t, reflect.TypeOf(analysis.ContributionTraceEntry{}), definition(t, trace, "contributionTraceEntry"))
	assertObjectShape(t, reflect.TypeOf(analysis.AssessabilityTraceEntry{}), definition(t, trace, "assessabilityTraceEntry"))
	assertObjectShape(t, reflect.TypeOf(analysis.ResolverResult{}), resolverResult)
	assertObjectShape(t, reflect.TypeOf(analysis.PropositionGraph{}), definition(t, resolverResult, "propositionGraph"))
	assertObjectShape(t, reflect.TypeOf(analysis.PropositionNode{}), definition(t, resolverResult, "propositionNode"))
	assertObjectShape(t, reflect.TypeOf(analysis.PropositionEdge{}), definition(t, resolverResult, "propositionEdge"))
	assertObjectShape(t, reflect.TypeOf(analysis.ResolverSense{}), definition(t, resolverResult, "resolverSense"))
	assertObjectShape(t, reflect.TypeOf(analysis.Ambiguity{}), definition(t, resolverResult, "ambiguity"))
}

func TestSchemasUseExactlyTheCanonicalDimensions(t *testing.T) {
	t.Parallel()

	want := make([]string, 0, len(analysis.Dimensions()))
	for _, id := range analysis.Dimensions() {
		want = append(want, string(id))
	}
	for _, filename := range []string{
		"sprach-a-lyzer_analysis-result_v0.1.json",
		"sprach-a-lyzer_analysis-trace_v0.1.json",
	} {
		schema := readSchema(t, filename)
		got := stringsFrom(t, definition(t, schema, "dimensionID")["enum"])
		if !slices.Equal(got, want) {
			t.Errorf("%s dimension enum = %v; want %v", filename, got, want)
		}
		containerName := "dimensions"
		if strings.Contains(filename, "trace") {
			containerName = "assessability"
		}
		container := property(t, schema, containerName)
		propertyNames := keysOf(t, container["properties"])
		requiredNames := stringsFrom(t, container["required"])
		slices.Sort(propertyNames)
		slices.Sort(requiredNames)
		sortedWant := slices.Clone(want)
		slices.Sort(sortedWant)
		if !slices.Equal(propertyNames, sortedWant) || !slices.Equal(requiredNames, sortedWant) {
			t.Errorf("%s %s keys/required = %v/%v; want %v", filename, containerName, propertyNames, requiredNames, sortedWant)
		}
		encoded, err := json.Marshal(schema)
		if err != nil {
			t.Fatalf("marshal %s: %v", filename, err)
		}
		if bytes.Contains(encoded, []byte("FREE_WILL")) {
			t.Errorf("%s contains legacy dimension ID", filename)
		}
	}
}

func TestRequestEnumConstantsMatchSchema(t *testing.T) {
	t.Parallel()

	schema := readSchema(t, "sprach-a-lyzer_analysis-request_v0.1.json")
	assertConst(t, property(t, schema, "locale"), string(analysis.LocaleGerman))
	assertConst(t, property(t, schema, "input_mode"), string(analysis.InputModeText))
	assertConst(t, property(t, schema, "analysis_mode"), string(analysis.AnalysisModeStandard))
	profiles := stringsFrom(t, property(t, schema, "presentation_profile")["enum"])
	wantProfiles := []string{string(analysis.ProfilePrivate), string(analysis.ProfileCorporate)}
	if !slices.Equal(profiles, wantProfiles) {
		t.Fatalf("presentation profiles = %v; want %v", profiles, wantProfiles)
	}
}

func TestEngineOutputAndDerivedTraceRespectContractShapes(t *testing.T) {
	t.Parallel()

	result, err := analysis.NewDefault().Analyze(analysis.Request{
		Text: "Ich muss das heute unbedingt noch schaffen.", Context: analysis.ContextSelfTalk,
		Locale: analysis.LocaleGerman, InputMode: analysis.InputModeText,
		PresentationProfile: analysis.ProfilePrivate, AnalysisMode: analysis.AnalysisModeStandard,
	})
	if err != nil {
		t.Fatalf("analyze contract fixture: %v", err)
	}
	assertEncodedTopLevelShape(t, result, readSchema(t, "sprach-a-lyzer_analysis-result_v0.1.json"))
	trace := result.Trace()
	assertEncodedTopLevelShape(t, trace, readSchema(t, "sprach-a-lyzer_analysis-trace_v0.1.json"))

	for id, entry := range trace.Assessability {
		for _, index := range entry.ContributionIndexes {
			if index < 0 || index >= len(trace.Contributions) {
				t.Fatalf("%s contribution index %d outside trace", id, index)
			}
			if trace.Contributions[index].Dimension != id {
				t.Fatalf("%s links contribution %d for %s", id, index, trace.Contributions[index].Dimension)
			}
		}
	}
}

func TestResolverCatalogueSchemaAndSeedAreVersionLocked(t *testing.T) {
	t.Parallel()
	schema := readSchema(t, "sprach-a-lyzer_resolver-catalogue_v0.1.json")
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_resolver-catalogue_v0.1.json")
	if err != nil {
		t.Fatalf("read resolver catalogue: %v", err)
	}
	var seed map[string]any
	if err := json.Unmarshal(data, &seed); err != nil {
		t.Fatalf("decode resolver catalogue: %v", err)
	}
	assertConst(t, property(t, schema, "catalogue_version"), "0.1")
	if seed["catalogue_version"] != "0.1" || seed["status"] != "APPROVED" || seed["locale"] != "de-DE" {
		t.Fatalf("unexpected resolver catalogue envelope: %#v", seed)
	}
	versionContract := seed["version_contract"].(map[string]any)
	if versionContract["resolver_result"] != "0.2" || versionContract["policy_registry"] != "0.4" {
		t.Fatalf("resolver version contract = %#v", versionContract)
	}
	wantGuardrails := property(t, schema, "hard_guardrails")["const"]
	if !reflect.DeepEqual(seed["hard_guardrails"], wantGuardrails) {
		t.Fatalf("resolver guardrails = %#v; schema = %#v", seed["hard_guardrails"], wantGuardrails)
	}
}

func TestResolverResultEnumsMatchCanonicalPolicyIDs(t *testing.T) {
	t.Parallel()
	schema := readSchema(t, "sprach-a-lyzer_resolver-result_v0.2.json")
	assertSameStringSet(t, "actor", definition(t, schema, "actor")["enum"], policy.Actors())
	assertSameStringSet(t, "target type", definition(t, schema, "targetType")["enum"], policy.TargetTypes())
	assertSameStringSet(t, "expectation source", definition(t, schema, "expectationSource")["enum"], policy.ExpectationSources())
	assertSameStringSet(t, "relation", definition(t, schema, "relation")["enum"], policy.DiscourseRelations())
	assertSameStringSet(t, "modality", definition(t, schema, "modality")["enum"], policy.Modalities())
	assertSameStringSet(t, "negation scope", definition(t, schema, "negationScope")["enum"], policy.NegationScopes())
	assertSameStringSet(t, "sense state", definition(t, schema, "senseState")["enum"], policy.SenseStates())
	ambiguityType := property(t, definition(t, schema, "ambiguity"), "type")
	assertSameStringSet(t, "ambiguity type", ambiguityType["enum"], policy.AmbiguityTypes())
}

func assertSameStringSet[T ~string](t *testing.T, name string, schemaValues any, policyValues []T) {
	t.Helper()
	got := stringsFrom(t, schemaValues)
	want := make([]string, len(policyValues))
	for index, value := range policyValues {
		want[index] = string(value)
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("resolver %s IDs = %v; policy = %v", name, got, want)
	}
}

func readSchema(t *testing.T, filename string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("decode %s: %v", filename, err)
	}
	return schema
}

func definition(t *testing.T, schema map[string]any, name string) map[string]any {
	t.Helper()
	definitions, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("schema lacks $defs")
	}
	definition, ok := definitions[name].(map[string]any)
	if !ok {
		t.Fatalf("schema lacks definition %q", name)
	}
	return definition
}

func property(t *testing.T, schema map[string]any, name string) map[string]any {
	t.Helper()
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema lacks properties")
	}
	property, ok := properties[name].(map[string]any)
	if !ok {
		t.Fatalf("schema lacks property %q", name)
	}
	return property
}

func keysOf(t *testing.T, value any) []string {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is not a JSON object: %#v", value)
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	return keys
}

func assertConst(t *testing.T, schema map[string]any, want string) {
	t.Helper()
	if got, ok := schema["const"].(string); !ok || got != want {
		t.Errorf("schema const = %#v; want %q", schema["const"], want)
	}
}

func assertObjectShape(t *testing.T, goType reflect.Type, schema map[string]any) {
	t.Helper()
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema for %s lacks properties", goType)
	}
	wantProperties := make([]string, 0, len(properties))
	for name := range properties {
		wantProperties = append(wantProperties, name)
	}
	slices.Sort(wantProperties)

	gotProperties := make([]string, 0, goType.NumField())
	gotRequired := []string{}
	for index := 0; index < goType.NumField(); index++ {
		tag := goType.Field(index).Tag.Get("json")
		parts := strings.Split(tag, ",")
		if parts[0] == "" || parts[0] == "-" {
			continue
		}
		gotProperties = append(gotProperties, parts[0])
		if !slices.Contains(parts[1:], "omitempty") {
			gotRequired = append(gotRequired, parts[0])
		}
	}
	slices.Sort(gotProperties)
	slices.Sort(gotRequired)
	wantRequired := stringsFrom(t, schema["required"])
	slices.Sort(wantRequired)
	if !slices.Equal(gotProperties, wantProperties) {
		t.Errorf("%s JSON properties = %v; schema = %v", goType, gotProperties, wantProperties)
	}
	if !slices.Equal(gotRequired, wantRequired) {
		t.Errorf("%s required fields = %v; schema = %v", goType, gotRequired, wantRequired)
	}
}

func assertEncodedTopLevelShape(t *testing.T, value any, schema map[string]any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal contract value: %v", err)
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatalf("decode contract value: %v", err)
	}
	properties := schema["properties"].(map[string]any)
	for key := range object {
		if _, exists := properties[key]; !exists {
			t.Errorf("encoded contract contains undocumented property %q", key)
		}
	}
	for _, required := range stringsFrom(t, schema["required"]) {
		if _, exists := object[required]; !exists {
			t.Errorf("encoded contract lacks required property %q", required)
		}
	}
}

func stringsFrom(t *testing.T, value any) []string {
	t.Helper()
	values, ok := value.([]any)
	if !ok {
		t.Fatalf("value is not a JSON string array: %#v", value)
	}
	result := make([]string, len(values))
	for index, value := range values {
		stringValue, ok := value.(string)
		if !ok {
			t.Fatalf("array value is not a string: %#v", value)
		}
		result[index] = stringValue
	}
	return result
}
