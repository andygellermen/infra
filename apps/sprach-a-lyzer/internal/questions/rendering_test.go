package questions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

func TestQuestionRenderingGoldenV01(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/golden/sprach-a-lyzer_question-rendering_v0.1.json")
	if err != nil {
		t.Fatal(err)
	}
	var suite struct {
		Version            string `json:"version"`
		RenderingCatalogue string `json:"rendering_catalogue"`
		RenderingContract  string `json:"rendering_contract"`
		ResultContract     string `json:"result_contract"`
		Cases              []struct {
			CaseID              string        `json:"case_id"`
			Request             RenderRequest `json:"request"`
			ExpectedRenderingID string        `json:"expected_rendering_id"`
			ExpectedFallback    bool          `json:"expected_fallback"`
			ExpectedReason      string        `json:"expected_reason"`
		} `json:"cases"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&suite); err != nil {
		t.Fatal(err)
	}
	if suite.Version != "0.1" || suite.RenderingCatalogue != RenderingCatalogueVersion ||
		suite.RenderingContract != RenderingV02Contract || suite.ResultContract != RenderingResultContract || len(suite.Cases) != 8 {
		t.Fatalf("unexpected rendering golden envelope: %+v", suite)
	}
	service := NewDefault()
	for _, testCase := range suite.Cases {
		t.Run(testCase.CaseID, func(t *testing.T) {
			result, err := service.Render(testCase.Request)
			if err != nil {
				t.Fatal(err)
			}
			if result.ContractVersion != RenderingResultContract || result.Rendering.ID != testCase.ExpectedRenderingID ||
				result.FallbackApplied != testCase.ExpectedFallback || result.FallbackReason != testCase.ExpectedReason ||
				!result.Quality.IntentPreserved || !result.Quality.WithinLimits || result.Quality.Scoring || result.Rendering.Scoring {
				t.Fatalf("render result = %+v", result)
			}
		})
	}
}

func TestRenderingCatalogueHasCompleteProfileIsolation(t *testing.T) {
	t.Parallel()
	catalogue, err := DecodeRenderingCatalogue(bytes.NewReader(assets.QuestionRenderingCatalogueV01))
	if err != nil {
		t.Fatal(err)
	}
	if len(catalogue.Renderings) != 40 {
		t.Fatalf("renderings = %d; want 40", len(catalogue.Renderings))
	}
	for _, rendering := range catalogue.Renderings {
		if rendering.Profile == "CORPORATE" && (rendering.SpiritualExplicitness != 0 || rendering.IntimacyLevel > .4 || rendering.RequiresOptIn) {
			t.Errorf("corporate leakage: %+v", rendering)
		}
		if rendering.PresentationMode == "DEEP_REFLECTIVE" && (rendering.Profile != "PRIVATE" || !rendering.RequiresOptIn) {
			t.Errorf("unsafe deep rendering: %+v", rendering)
		}
	}
}

func TestRenderingPreservesCanonicalCoreAndQuestionSelection(t *testing.T) {
	t.Parallel()
	service := NewDefault()
	private, err := service.Render(RenderRequest{QuestionID: "CQ009", Profile: "PRIVATE"})
	if err != nil {
		t.Fatal(err)
	}
	corporate, err := service.Render(RenderRequest{QuestionID: "CQ009", Profile: "CORPORATE"})
	if err != nil {
		t.Fatal(err)
	}
	deep, err := service.Render(RenderRequest{QuestionID: "CQ009", Profile: "PRIVATE", Action: "REPHRASE", DeepReflectionOptIn: true})
	if err != nil {
		t.Fatal(err)
	}
	if private.Rendering.Text == corporate.Rendering.Text || private.Rendering.Text == deep.Rendering.Text ||
		private.Rendering.ConstructIntent != corporate.Rendering.ConstructIntent || corporate.Rendering.ConstructIntent != deep.Rendering.ConstructIntent {
		t.Fatalf("profile renderings do not preserve isolated common intent: private=%+v corporate=%+v deep=%+v", private, corporate, deep)
	}
	request := AnswerRequest{QuestionID: "CQ009", Answer: "Mir ist wichtig, dass wir fair entscheiden."}
	before, err := service.Analyze(request)
	if err != nil {
		t.Fatal(err)
	}
	after, err := service.Analyze(request)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("rendering changed deterministic Q/A core: before=%+v after=%+v err=%v", before, after, err)
	}
}

func TestCorporateAndPrivateProfilesPreserveCoreScores(t *testing.T) {
	t.Parallel()
	analyzer := analysis.NewDefault()
	base := analysis.Request{
		Text: "Ich muss das heute unbedingt noch schaffen.", Context: analysis.ContextSelfTalk,
		Locale: analysis.LocaleGerman, InputMode: analysis.InputModeText, AnalysisMode: analysis.AnalysisModeStandard,
	}
	base.PresentationProfile = analysis.ProfilePrivate
	private, err := analyzer.Analyze(base)
	if err != nil {
		t.Fatal(err)
	}
	base.PresentationProfile = analysis.ProfileCorporate
	corporate, err := analyzer.Analyze(base)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(private.Propositions, corporate.Propositions) ||
		!reflect.DeepEqual(private.Patterns, corporate.Patterns) ||
		!reflect.DeepEqual(private.Dimensions, corporate.Dimensions) {
		t.Fatalf("profile leaked into canonical core: private=%+v corporate=%+v", private, corporate)
	}
	if len(private.ContributionTrace) != len(corporate.ContributionTrace) {
		t.Fatalf("profile changed contribution count: private=%d corporate=%d", len(private.ContributionTrace), len(corporate.ContributionTrace))
	}
	for index := range private.ContributionTrace {
		left, right := private.ContributionTrace[index], corporate.ContributionTrace[index]
		if left.RuleID != right.RuleID || left.Dimension != right.Dimension || left.Delta != right.Delta {
			t.Fatalf("profile changed contribution %d: private=%+v corporate=%+v", index, left, right)
		}
	}
}

func TestRenderingActivationDeactivationAndFallbackSmoke(t *testing.T) {
	t.Parallel()
	catalogue, err := DecodeRenderingCatalogue(bytes.NewReader(assets.QuestionRenderingCatalogueV01))
	if err != nil {
		t.Fatal(err)
	}
	request := RenderRequest{QuestionID: "CQ034", Profile: "PRIVATE", Action: "REPHRASE", DeepReflectionOptIn: true}
	active, err := NewDefault().Render(request)
	if err != nil || active.FallbackApplied {
		t.Fatalf("active rendering = %+v, %v", active, err)
	}
	for index := range catalogue.Renderings {
		if catalogue.Renderings[index].ID == active.Rendering.ID {
			catalogue.Renderings[index].Status = "ARCHIVED"
		}
	}
	semantic, err := DecodeCatalogue(bytes.NewReader(assets.QuestionCatalogueV01))
	if err != nil {
		t.Fatal(err)
	}
	inactive := NewWithCatalogues(analysis.NewDefault(), StaticCatalogueProvider{Catalogue: semantic}, StaticRenderingCatalogueProvider{Catalogue: catalogue})
	fallback, err := inactive.Render(request)
	if err != nil || !fallback.FallbackApplied || fallback.FallbackReason != "NO_APPROVED_SAFE_VARIANT" || fallback.Rendering.Variant != "CANONICAL" {
		t.Fatalf("inactive fallback = %+v, %v", fallback, err)
	}
	restored, err := NewDefault().Render(request)
	if err != nil || restored.Rendering.ID != active.Rendering.ID {
		t.Fatalf("restored rendering = %+v, %v", restored, err)
	}
}

func TestRenderingFailsClosedOnProviderOrIntentDrift(t *testing.T) {
	t.Parallel()
	semantic, err := DecodeCatalogue(bytes.NewReader(assets.QuestionCatalogueV01))
	if err != nil {
		t.Fatal(err)
	}
	want := errors.New("renderings unavailable")
	service := NewWithCatalogues(analysis.NewDefault(), StaticCatalogueProvider{Catalogue: semantic}, renderingProviderFunc(func(context.Context) (RenderingCatalogue, error) {
		return RenderingCatalogue{}, want
	}))
	if _, err := service.Render(RenderRequest{QuestionID: "CQ007"}); !errors.Is(err, want) {
		t.Fatalf("provider error = %v; want wrapped failure", err)
	}
	catalogue, err := DecodeRenderingCatalogue(bytes.NewReader(assets.QuestionRenderingCatalogueV01))
	if err != nil {
		t.Fatal(err)
	}
	catalogue.CanonicalIntents["CQ007"] = "VALUES"
	for index := range catalogue.Renderings {
		if catalogue.Renderings[index].QuestionID == "CQ007" {
			catalogue.Renderings[index].ConstructIntent = "VALUES"
		}
	}
	drifted := NewWithCatalogues(analysis.NewDefault(), StaticCatalogueProvider{Catalogue: semantic}, StaticRenderingCatalogueProvider{Catalogue: catalogue})
	if _, err := drifted.Render(RenderRequest{QuestionID: "CQ007"}); err == nil || !strings.Contains(err.Error(), "intent drift") {
		t.Fatalf("intent drift error = %v", err)
	}
}

type renderingProviderFunc func(context.Context) (RenderingCatalogue, error)

func (function renderingProviderFunc) Active(ctx context.Context) (RenderingCatalogue, error) {
	return function(ctx)
}
