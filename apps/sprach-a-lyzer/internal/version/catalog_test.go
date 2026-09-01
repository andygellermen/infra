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
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_release-manifest_v0.3.0.json")
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
			QAObservation              string `json:"qa_observation"`
			QuestionSelection          string `json:"question_selection"`
			QuestionSession            string `json:"question_session"`
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
	if manifest.ManifestVersion != "0.3" || manifest.Product != "sprach-a-lyzer-core" ||
		manifest.ReleaseVersion != Core || manifest.GitTag != "v"+Core || manifest.Status != "RELEASED" ||
		manifest.Capability != "QUESTION_ANSWER_MVP" || manifest.ReleaseDate != "2026-09-01" {
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
		QAObservation              string `json:"qa_observation"`
		QuestionSelection          string `json:"question_selection"`
		QuestionSession            string `json:"question_session"`
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
		DatabaseSchema             int    `json:"database_schema"`
	}{
		HTTPAPI: "3", AnalysisRequest: "0.1", AnalysisResult: "0.1", AnalysisTrace: "0.2",
		ResolverResult: "0.2", ResolverCatalogue: "0.1", ConstructOntology: "0.2",
		CanonicalQuestion: "0.1", QuestionRendering: "0.1", QuestionCatalogue: "0.1",
		QAObservation: "0.1", QuestionSelection: "0.1", QuestionSession: "0.1",
		RuleContract: "0.5", RuleSet: "0.4", PolicyRegistry: "0.7",
		ParameterContract: "0.1", ParameterSet: "0.1", PresentationBundle: "0.2",
		CoreGoldenSuite: "0.2", ResolverGoldenSuite: "0.3", PropositionTraceGolden: "0.1",
		ConstructCompositionGolden: "0.1", QARuntimeGolden: "0.2", DatabaseSchema: 4,
	}
	if !reflect.DeepEqual(manifest.VersionVector, wantVector) {
		t.Fatalf("release version vector = %+v; want %+v", manifest.VersionVector, wantVector)
	}
	wantEvidence := []string{
		"CORE_V0_2_BACKWARD_PARITY",
		"QA_RUNTIME_GOLDEN_V0_2_PARITY",
		"ADAPTIVE_QUESTION_SELECTION_5_TO_8",
		"QUESTION_SESSION_C0_TO_C3",
		"NO_CAUSAL_C4_CLAIMS",
		"QUESTION_SCORE_BIAS_ZERO",
		"QA_EVIDENCE_NON_SCORING",
		"QUESTION_CATALOGUE_ACTIVATION_REACTIVATION_SMOKE",
		"PUBLIC_QUESTION_API_V0_3",
		"DATABASE_SCHEMA_READINESS",
	}
	if !reflect.DeepEqual(manifest.ClosureEvidence, wantEvidence) {
		t.Fatalf("closure evidence = %v; want %v", manifest.ClosureEvidence, wantEvidence)
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
