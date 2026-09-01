package questions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

func TestQuestionAnswerRuntimeGoldenV02(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/golden/sprach-a-lyzer_question-answer-runtime_v0.2.json")
	if err != nil {
		t.Fatalf("read Q/A golden: %v", err)
	}
	var suite struct {
		Version             string `json:"version"`
		CatalogueVersion    string `json:"catalogue_version"`
		ObservationContract string `json:"observation_contract"`
		Cases               []struct {
			CaseID                string               `json:"case_id"`
			QuestionID            string               `json:"question_id"`
			Answer                string               `json:"answer"`
			ExpectedRelevance     float64              `json:"expected_relevance"`
			ExpectedPattern       string               `json:"expected_pattern"`
			ExpectedDirections    []DimensionDirection `json:"expected_directions"`
			ExpectedAssessability string               `json:"expected_assessability"`
			ExpectedInference     string               `json:"expected_inference"`
		} `json:"cases"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&suite); err != nil {
		t.Fatalf("decode Q/A golden: %v", err)
	}
	if suite.Version != "0.2" || suite.CatalogueVersion != CatalogueVersion ||
		suite.ObservationContract != ObservationContractVersion || len(suite.Cases) != 25 {
		t.Fatalf("unexpected Q/A suite envelope: version=%s catalogue=%s observation=%s cases=%d", suite.Version, suite.CatalogueVersion, suite.ObservationContract, len(suite.Cases))
	}
	service := NewDefault()
	for _, testCase := range suite.Cases {
		t.Run(testCase.CaseID, func(t *testing.T) {
			observation, err := service.Analyze(AnswerRequest{QuestionID: testCase.QuestionID, Answer: testCase.Answer})
			if err != nil {
				t.Fatalf("Analyze() error: %v", err)
			}
			if observation.ContractVersion != suite.ObservationContract || observation.AnswerRelevance != testCase.ExpectedRelevance ||
				observation.Assessability != testCase.ExpectedAssessability || observation.InferenceLevel != testCase.ExpectedInference ||
				observation.QuestionScoreBias != 0 || observation.ContextGain < 0 || observation.ContextGain > .15 {
				t.Fatalf("observation envelope = %+v", observation)
			}
			wantPatterns := []string{}
			if testCase.ExpectedPattern != "" {
				wantPatterns = []string{testCase.ExpectedPattern}
			}
			if !slices.Equal(observation.QAPatterns, wantPatterns) {
				t.Errorf("patterns = %v; want %v", observation.QAPatterns, wantPatterns)
			}
			gotDirections := make([]DimensionDirection, len(observation.DimensionEvidence))
			for index, evidence := range observation.DimensionEvidence {
				if evidence.Scoring {
					t.Errorf("Q/A evidence must remain non-scoring: %+v", evidence)
				}
				gotDirections[index] = DimensionDirection{Dimension: evidence.Dimension, Direction: evidence.Direction}
			}
			if !reflect.DeepEqual(gotDirections, testCase.ExpectedDirections) {
				t.Errorf("directions = %+v; want %+v", gotDirections, testCase.ExpectedDirections)
			}
			direct, err := analysis.NewDefault().Analyze(analysis.Request{
				Text: testCase.Answer, Locale: analysis.LocaleGerman, Context: "COACHING", InputMode: analysis.InputModeText,
				PresentationProfile: analysis.ProfilePrivate, AnalysisMode: analysis.AnalysisModeStandard,
			})
			if err != nil || !reflect.DeepEqual(observation.AnalysisResult, direct) {
				t.Fatalf("question context changed core analysis: direct=%+v observation=%+v err=%v", direct, observation.AnalysisResult, err)
			}
		})
	}
}

func TestAdaptiveSelectionOffersFiveToEightQuestions(t *testing.T) {
	t.Parallel()
	selection, err := NewDefault().Select(SelectionRequest{
		Profile: "CORPORATE", PreferredPhase: "P2_FOLLOWUP_1_10", ConstructGaps: []string{"OPTIONS"}, Limit: 5,
	})
	if err != nil {
		t.Fatalf("Select() error: %v", err)
	}
	if selection.ContractVersion != SelectionContractVersion || len(selection.Candidates) != 5 || selection.Candidates[0].QuestionID != "CQ034" {
		t.Fatalf("adaptive selection = %+v", selection)
	}
	seen := map[string]bool{}
	for _, candidate := range selection.Candidates {
		if !candidate.Offered || candidate.Text == "" || seen[candidate.QuestionID] {
			t.Errorf("invalid offered candidate: %+v", candidate)
		}
		seen[candidate.QuestionID] = true
	}
}

func TestSessionSupportsHedgedC2AndTemporalC3WithoutC4(t *testing.T) {
	t.Parallel()
	session, err := NewDefault().ComposeSession(SessionRequest{
		SessionID: "S-QA-001",
		Pairs: []AnswerRequest{
			{QuestionID: "CQ007", Baseline: "Mein Chef ist diese Woche im Urlaub.", Answer: "Eigentlich gar nichts. Ich kann nur abwarten."},
			{QuestionID: "CQ013", Answer: "Meist sage ich mir: Ich kann Hilfe holen und trotzdem verantwortlich bleiben."},
		},
	})
	if err != nil {
		t.Fatalf("ComposeSession() error: %v", err)
	}
	if session.ContractVersion != SessionContractVersion || session.InferenceLevel != "C3" || len(session.Trajectories) == 0 {
		t.Fatalf("session inference = %+v", session)
	}
	levels := []string{}
	for _, claim := range session.InferenceClaims {
		levels = append(levels, claim.Level)
	}
	if !slices.Contains(levels, "C2") || !slices.Contains(levels, "C3") || slices.Contains(levels, "C4") {
		t.Fatalf("session inference claims = %v", levels)
	}
	if session.Trajectories[0].InferenceLevel != "C3" || session.Trajectories[0].LanguagePatternChange != "NEGATIVE_TO_POSITIVE" {
		t.Fatalf("session trajectory = %+v", session.Trajectories)
	}
}

func TestQuestionCatalogueActivationAndDeactivationSmoke(t *testing.T) {
	t.Parallel()
	catalogue, err := DecodeCatalogue(bytes.NewReader(assets.QuestionCatalogueV01))
	if err != nil {
		t.Fatalf("decode catalogue: %v", err)
	}
	request := AnswerRequest{QuestionID: "CQ034", Answer: "Es gibt keine andere Option."}
	active, err := NewWithCatalogue(analysis.NewDefault(), StaticCatalogueProvider{Catalogue: catalogue}).Analyze(request)
	if err != nil || !slices.Contains(active.QAPatterns, "NO_OPTION_SPACE") {
		t.Fatalf("active observation = %+v, %v", active, err)
	}
	for index := range catalogue.Compositions {
		if catalogue.Compositions[index].OutputPattern == "NO_OPTION_SPACE" {
			catalogue.Compositions[index].PhrasesAll = []string{"catalogue rule intentionally inactive"}
		}
	}
	inactive, err := NewWithCatalogue(analysis.NewDefault(), StaticCatalogueProvider{Catalogue: catalogue}).Analyze(request)
	if err != nil || len(inactive.QAPatterns) != 0 || inactive.Assessability != "NOT_ASSESSABLE" {
		t.Fatalf("inactive observation = %+v, %v", inactive, err)
	}
	restored, err := NewDefault().Analyze(request)
	if err != nil || !reflect.DeepEqual(restored.QAPatterns, active.QAPatterns) {
		t.Fatalf("restored observation lacks parity: %+v, %v", restored, err)
	}
}

func TestQuestionCatalogueFailsClosed(t *testing.T) {
	t.Parallel()
	want := errors.New("catalogue unavailable")
	service := NewWithCatalogue(analysis.NewDefault(), catalogueProviderFunc(func(context.Context) (Catalogue, error) {
		return Catalogue{}, want
	}))
	_, err := service.Analyze(AnswerRequest{QuestionID: "CQ007", Answer: "Ja."})
	if !errors.Is(err, want) {
		t.Fatalf("Analyze() error = %v; want wrapped provider failure", err)
	}
}

func TestHistoricalQuestionCorpusMismatchIsExplicitlyQuarantined(t *testing.T) {
	t.Parallel()
	coreData, err := os.ReadFile("../../data/seed/sprachkompass_mvp-question-core-set_v0.1.json")
	if err != nil {
		t.Fatal(err)
	}
	goldenData, err := os.ReadFile("../../data/golden/sprachkompass_question-golden-corpus_v0.1.json")
	if err != nil {
		t.Fatal(err)
	}
	var core struct {
		Questions []struct {
			ID        string `json:"question_id"`
			Construct string `json:"primary_construct"`
		} `json:"questions"`
	}
	var historical struct {
		Cases []struct {
			ID        string `json:"question_id"`
			Construct string `json:"expected_primary_construct"`
		} `json:"cases"`
	}
	if json.Unmarshal(coreData, &core) != nil || json.Unmarshal(goldenData, &historical) != nil {
		t.Fatal("decode historical question artefacts")
	}
	constructs := map[string]string{}
	for _, question := range core.Questions {
		constructs[question.ID] = question.Construct
	}
	mismatches := 0
	for _, testCase := range historical.Cases {
		if constructs[testCase.ID] != testCase.Construct {
			mismatches++
		}
	}
	if mismatches != 21 {
		t.Fatalf("historical mismatch count = %d; want documented 21", mismatches)
	}
	if bytes.Contains(assets.QuestionCatalogueV01, []byte("FREE_WILL")) {
		t.Fatal("approved question catalogue contains legacy dimension ID")
	}
}

type catalogueProviderFunc func(context.Context) (Catalogue, error)

func (f catalogueProviderFunc) Active(ctx context.Context) (Catalogue, error) { return f(ctx) }

func TestUnknownQuestionAndEmptyAnswerAreRejected(t *testing.T) {
	t.Parallel()
	service := NewDefault()
	if _, err := service.Analyze(AnswerRequest{QuestionID: "CQ999", Answer: "Antwort"}); !errors.Is(err, ErrUnknownQuestion) {
		t.Fatalf("unknown question error = %v", err)
	}
	if _, err := service.Analyze(AnswerRequest{QuestionID: "CQ007", Answer: "  "}); !errors.Is(err, ErrEmptyAnswer) {
		t.Fatalf("empty answer error = %v", err)
	}
}

func TestQuestionRuntimeRejectsInvalidProfileAndOversizedAnswer(t *testing.T) {
	t.Parallel()
	service := NewDefault()
	if _, err := service.Select(SelectionRequest{Profile: "UNKNOWN", Limit: 5}); !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("Select() error = %v; want invalid profile", err)
	}
	if _, err := service.Analyze(AnswerRequest{QuestionID: "CQ007", Profile: "UNKNOWN", Answer: "Ja."}); !errors.Is(err, ErrInvalidProfile) {
		t.Fatalf("Analyze() profile error = %v; want invalid profile", err)
	}
	if _, err := service.Analyze(AnswerRequest{QuestionID: "CQ007", Answer: strings.Repeat("a", 10_001)}); !errors.Is(err, ErrAnswerTooLong) {
		t.Fatalf("Analyze() length error = %v; want oversized answer", err)
	}
}

func TestNoQuestionOutputContainsTraitOrCausalEffectClaims(t *testing.T) {
	t.Parallel()
	observation, err := NewDefault().Analyze(AnswerRequest{QuestionID: "CQ013", Answer: "Ich muss das allein schaffen."})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(observation)
	if err != nil {
		t.Fatal(err)
	}
	for _, prohibited := range []string{"C4", "DIAGNOSIS", "TRAIT", "PERSÖNLICHKEIT"} {
		if strings.Contains(strings.ToUpper(string(encoded)), prohibited) {
			t.Fatalf("observation contains prohibited claim %q: %s", prohibited, encoded)
		}
	}
}
