package experience

import (
	"errors"
	"reflect"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/questions"
)

func TestMVPExperienceComposesNoAIPrivateResultWithoutPersistence(t *testing.T) {
	t.Parallel()
	analyzer := analysis.NewDefault()
	service := New(analyzer, questions.New(analyzer))
	request := Request{Text: "Ich muss das heute unbedingt noch schaffen.", Context: "SELF_TALK", Profile: "PRIVATE"}
	result, err := service.Analyze(request)
	if err != nil {
		t.Fatal(err)
	}
	direct, err := analyzer.Analyze(analysis.Request{Text: request.Text, Locale: analysis.LocaleGerman, Context: analysis.ContextSelfTalk, InputMode: analysis.InputModeText, PresentationProfile: analysis.ProfilePrivate, AnalysisMode: analysis.AnalysisModeStandard})
	if err != nil || !reflect.DeepEqual(result.Core, direct) {
		t.Fatalf("experience changed canonical core: %v", err)
	}
	if result.ExperienceMode != "CORE_NO_AI" || result.Privacy.RawTextStored || result.Privacy.AnalysisStored || result.Privacy.ExternalTransfer || result.Privacy.AIUsed {
		t.Fatalf("privacy receipt = %+v", result.Privacy)
	}
	if len(result.Dimensions) != 6 || len(result.SuggestedQuestions) != 5 || result.ReflectionQuestion == nil || len(result.Alternatives) != 2 {
		t.Fatalf("experience result = %+v", result)
	}
}

func TestMVPExperienceKeepsProfilesIsolatedAndSupportsEasyQuestions(t *testing.T) {
	t.Parallel()
	analyzer := analysis.NewDefault()
	service := New(analyzer, questions.New(analyzer))
	private, err := service.Analyze(Request{Text: "Hast du Geld?", Profile: "PRIVATE"})
	if err != nil {
		t.Fatal(err)
	}
	corporate, err := service.Analyze(Request{Text: "Hast du Geld?", Profile: "CORPORATE", LanguageLevel: "EASY"})
	if err != nil {
		t.Fatal(err)
	}
	if private.Dimensions[4].Label != "Freier Wille" || corporate.Dimensions[4].Label != "Handlungsspielraum" || private.SuggestedQuestions[0].Text == corporate.SuggestedQuestions[0].Text {
		t.Fatalf("profile views leaked: private=%+v corporate=%+v", private.Dimensions, corporate.Dimensions)
	}
	if !reflect.DeepEqual(private.Core.Dimensions, corporate.Core.Dimensions) {
		t.Fatal("profile changed core dimension result")
	}
}

func TestMVPExperienceFailsClosed(t *testing.T) {
	t.Parallel()
	service := New(analysis.NewDefault(), questions.NewDefault())
	for _, test := range []struct {
		request Request
		want    error
	}{
		{Request{}, ErrEmptyText},
		{Request{Text: "Text", Profile: "TEAM_RANKING"}, ErrInvalidProfile},
		{Request{Text: "Text", LanguageLevel: "GENERATIVE"}, ErrInvalidLanguage},
		{Request{Text: "Text", Context: "PERSONALITY_DIAGNOSIS"}, ErrInvalidContext},
	} {
		if _, err := service.Analyze(test.request); !errors.Is(err, test.want) {
			t.Fatalf("Analyze(%+v) error=%v; want %v", test.request, err, test.want)
		}
	}
}
