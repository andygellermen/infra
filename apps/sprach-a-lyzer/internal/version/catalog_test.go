package version

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestCoreMatchesReleaseManifest(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_release-manifest_v0.5.0.json")
	if err != nil {
		t.Fatalf("read release manifest: %v", err)
	}
	var manifest struct {
		ManifestVersion string `json:"manifest_version"`
		Product         string `json:"product"`
		ReleaseVersion  string `json:"release_version"`
		ReleaseDate     string `json:"release_date"`
		Status          string `json:"status"`
		GitTag          string `json:"git_tag"`
		Capability      string `json:"capability"`
		VersionVector   struct {
			HTTPAPI                    string `json:"http_api"`
			AnalysisRequest            string `json:"analysis_request"`
			AnalysisResult             string `json:"analysis_result"`
			AnalysisTrace              string `json:"analysis_trace"`
			ResolverResult             string `json:"resolver_result"`
			ResolverCatalogue          string `json:"resolver_catalogue"`
			ConstructOntology          string `json:"construct_ontology"`
			CanonicalQuestion          string `json:"canonical_question"`
			QuestionRendering          string `json:"question_rendering"`
			QuestionCatalogue          string `json:"question_catalogue"`
			QuestionRenderingCatalogue string `json:"question_rendering_catalogue"`
			QuestionRenderingResult    string `json:"question_rendering_result"`
			QAObservation              string `json:"qa_observation"`
			QuestionSelection          string `json:"question_selection"`
			QuestionSession            string `json:"question_session"`
			ManagedImportRequest       string `json:"managed_import_request"`
			ManagedImportPlan          string `json:"managed_import_plan"`
			ManagedImportOperation     string `json:"managed_import_operation"`
			RuleContract               string `json:"rule_contract"`
			RuleSet                    string `json:"rule_set"`
			PolicyRegistry             string `json:"policy_registry"`
			ParameterContract          string `json:"parameter_contract"`
			ParameterSet               string `json:"parameter_set"`
			PresentationBundle         string `json:"presentation_bundle"`
			CoreGoldenSuite            string `json:"core_golden_suite"`
			ResolverGoldenSuite        string `json:"resolver_golden_suite"`
			PropositionTraceGolden     string `json:"proposition_trace_golden"`
			ConstructCompositionGolden string `json:"construct_composition_golden"`
			QARuntimeGolden            string `json:"qa_runtime_golden"`
			QuestionRenderingGolden    string `json:"question_rendering_golden"`
			DatabaseSchema             int    `json:"database_schema"`
		} `json:"version_vector"`
		ClosureEvidence []string `json:"closure_evidence"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode release manifest: %v", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("release manifest contains trailing JSON: %v", err)
	}
	if manifest.ManifestVersion != "0.5" || manifest.Product != "sprach-a-lyzer-core" ||
		manifest.ReleaseVersion != Core || manifest.GitTag != "v"+Core || manifest.Status != "RELEASED" ||
		manifest.Capability != "MANAGED_KNOWLEDGE_OPERATIONS" || manifest.ReleaseDate != "2026-09-02" {
		t.Fatalf("release manifest envelope = %+v; core = %s", manifest, Core)
	}
	wantVector := struct {
		HTTPAPI                    string `json:"http_api"`
		AnalysisRequest            string `json:"analysis_request"`
		AnalysisResult             string `json:"analysis_result"`
		AnalysisTrace              string `json:"analysis_trace"`
		ResolverResult             string `json:"resolver_result"`
		ResolverCatalogue          string `json:"resolver_catalogue"`
		ConstructOntology          string `json:"construct_ontology"`
		CanonicalQuestion          string `json:"canonical_question"`
		QuestionRendering          string `json:"question_rendering"`
		QuestionCatalogue          string `json:"question_catalogue"`
		QuestionRenderingCatalogue string `json:"question_rendering_catalogue"`
		QuestionRenderingResult    string `json:"question_rendering_result"`
		QAObservation              string `json:"qa_observation"`
		QuestionSelection          string `json:"question_selection"`
		QuestionSession            string `json:"question_session"`
		ManagedImportRequest       string `json:"managed_import_request"`
		ManagedImportPlan          string `json:"managed_import_plan"`
		ManagedImportOperation     string `json:"managed_import_operation"`
		RuleContract               string `json:"rule_contract"`
		RuleSet                    string `json:"rule_set"`
		PolicyRegistry             string `json:"policy_registry"`
		ParameterContract          string `json:"parameter_contract"`
		ParameterSet               string `json:"parameter_set"`
		PresentationBundle         string `json:"presentation_bundle"`
		CoreGoldenSuite            string `json:"core_golden_suite"`
		ResolverGoldenSuite        string `json:"resolver_golden_suite"`
		PropositionTraceGolden     string `json:"proposition_trace_golden"`
		ConstructCompositionGolden string `json:"construct_composition_golden"`
		QARuntimeGolden            string `json:"qa_runtime_golden"`
		QuestionRenderingGolden    string `json:"question_rendering_golden"`
		DatabaseSchema             int    `json:"database_schema"`
	}{
		HTTPAPI: "5", AnalysisRequest: "0.1", AnalysisResult: "0.1", AnalysisTrace: "0.2",
		ResolverResult: "0.2", ResolverCatalogue: "0.1", ConstructOntology: "0.2",
		CanonicalQuestion: "0.1", QuestionRendering: "0.2", QuestionCatalogue: "0.1",
		QuestionRenderingCatalogue: "0.1", QuestionRenderingResult: "0.1",
		QAObservation: "0.1", QuestionSelection: "0.1", QuestionSession: "0.1",
		ManagedImportRequest: "0.1", ManagedImportPlan: "0.1", ManagedImportOperation: "0.1",
		RuleContract: "0.5", RuleSet: "0.4", PolicyRegistry: "0.7",
		ParameterContract: "0.1", ParameterSet: "0.1", PresentationBundle: "0.2",
		CoreGoldenSuite: "0.2", ResolverGoldenSuite: "0.3", PropositionTraceGolden: "0.1",
		ConstructCompositionGolden: "0.1", QARuntimeGolden: "0.2", QuestionRenderingGolden: "0.1", DatabaseSchema: 5,
	}
	if !reflect.DeepEqual(manifest.VersionVector, wantVector) {
		t.Fatalf("release version vector = %+v; want %+v", manifest.VersionVector, wantVector)
	}
	wantEvidence := []string{
		"V0_4_BACKWARD_PARITY", "JSON_CSV_XLSX_PARSING", "AUTOMATIC_AND_MANUAL_FIELD_MAPPING", "NATURAL_AND_SECONDARY_MATCHING", "FIELD_LEVEL_DIFF_PREVIEW",
		"CONFLICT_AND_CRITICAL_FIELD_REVIEW", "REFERENCE_AND_DEPENDENCY_GRAPH_VALIDATION", "CORE_GOLDEN_DRY_RUN", "DUPLICATE_SOURCE_SHA256", "VALIDATE_ONLY_WRITE_PROTECTION",
		"ROLE_GUARDED_TRANSACTIONAL_COMMIT", "REVERSIBLE_CHANGE_LOG", "IMMUTABLE_AUDIT_LOG", "PUBLIC_MANAGED_IMPORT_API_V0_5",
		"DATABASE_SCHEMA_READINESS",
	}
	if !reflect.DeepEqual(manifest.ClosureEvidence, wantEvidence) {
		t.Fatalf("closure evidence = %v; want %v", manifest.ClosureEvidence, wantEvidence)
	}
}

func TestHistoricalV040ReleaseManifestRemainsImmutable(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_release-manifest_v0.4.0.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		ManifestVersion string `json:"manifest_version"`
		ReleaseVersion  string `json:"release_version"`
		GitTag          string `json:"git_tag"`
		Capability      string `json:"capability"`
		VersionVector   struct {
			HTTPAPI           string `json:"http_api"`
			QuestionRendering string `json:"question_rendering"`
			DatabaseSchema    int    `json:"database_schema"`
		} `json:"version_vector"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.ManifestVersion != "0.4" || manifest.ReleaseVersion != "0.4.0" || manifest.GitTag != "v0.4.0" || manifest.Capability != "PROFILE_ISOLATED_QUESTION_RENDERING" || manifest.VersionVector.HTTPAPI != "4" || manifest.VersionVector.QuestionRendering != "0.2" || manifest.VersionVector.DatabaseSchema != 4 {
		t.Fatalf("historical v0.4.0 manifest drifted: %+v", manifest)
	}
}

func TestHistoricalV030ReleaseManifestRemainsImmutable(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_release-manifest_v0.3.0.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		ManifestVersion string `json:"manifest_version"`
		ReleaseVersion  string `json:"release_version"`
		GitTag          string `json:"git_tag"`
		Capability      string `json:"capability"`
		VersionVector   struct {
			HTTPAPI           string `json:"http_api"`
			QuestionRendering string `json:"question_rendering"`
			QARuntimeGolden   string `json:"qa_runtime_golden"`
			DatabaseSchema    int    `json:"database_schema"`
		} `json:"version_vector"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.ManifestVersion != "0.3" || manifest.ReleaseVersion != "0.3.0" || manifest.GitTag != "v0.3.0" ||
		manifest.Capability != "QUESTION_ANSWER_MVP" || manifest.VersionVector.HTTPAPI != "3" ||
		manifest.VersionVector.QuestionRendering != "0.1" || manifest.VersionVector.QARuntimeGolden != "0.2" || manifest.VersionVector.DatabaseSchema != 4 {
		t.Fatalf("historical v0.3.0 manifest drifted: %+v", manifest)
	}
}

func TestHistoricalV020ReleaseManifestRemainsImmutable(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_release-manifest_v0.2.0.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		ManifestVersion string `json:"manifest_version"`
		ReleaseVersion  string `json:"release_version"`
		GitTag          string `json:"git_tag"`
		Capability      string `json:"capability"`
		VersionVector   struct {
			HTTPAPI        string `json:"http_api"`
			AnalysisTrace  string `json:"analysis_trace"`
			PolicyRegistry string `json:"policy_registry"`
			DatabaseSchema int    `json:"database_schema"`
		} `json:"version_vector"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.ManifestVersion != "0.2" || manifest.ReleaseVersion != "0.2.0" || manifest.GitTag != "v0.2.0" ||
		manifest.Capability != "CONTEXT_AND_PROPOSITION_CORE" || manifest.VersionVector.HTTPAPI != "2" ||
		manifest.VersionVector.AnalysisTrace != "0.2" || manifest.VersionVector.PolicyRegistry != "0.7" || manifest.VersionVector.DatabaseSchema != 4 {
		t.Fatalf("historical v0.2.0 manifest drifted: %+v", manifest)
	}
}

func TestHistoricalV010ReleaseManifestRemainsImmutable(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_release-manifest_v0.1.0.json")
	if err != nil {
		t.Fatalf("read historical release manifest: %v", err)
	}
	var manifest struct {
		ManifestVersion string `json:"manifest_version"`
		ReleaseVersion  string `json:"release_version"`
		GitTag          string `json:"git_tag"`
		Capability      string `json:"capability"`
		VersionVector   struct {
			AnalysisTrace  string `json:"analysis_trace"`
			RuleContract   string `json:"rule_contract"`
			RuleSet        string `json:"rule_set"`
			PolicyRegistry string `json:"policy_registry"`
			DatabaseSchema int    `json:"database_schema"`
		} `json:"version_vector"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode historical release manifest: %v", err)
	}
	if manifest.ManifestVersion != "0.1" || manifest.ReleaseVersion != "0.1.0" || manifest.GitTag != "v0.1.0" ||
		manifest.Capability != "DETERMINISTIC_LANGUAGE_CORE" || manifest.VersionVector.AnalysisTrace != "0.1" ||
		manifest.VersionVector.RuleContract != "0.4" || manifest.VersionVector.RuleSet != "0.3" ||
		manifest.VersionVector.PolicyRegistry != "0.3" || manifest.VersionVector.DatabaseSchema != 3 {
		t.Fatalf("historical v0.1.0 manifest drifted: %+v", manifest)
	}
}
