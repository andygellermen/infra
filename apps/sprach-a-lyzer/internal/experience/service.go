// Package experience composes the deterministic modules into the transient MVP product experience.
package experience

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/questions"
)

const ContractVersion = "0.1"

var (
	ErrEmptyText       = errors.New("experience text must not be empty")
	ErrTextTooLong     = errors.New("experience text must not exceed 10,000 characters")
	ErrInvalidProfile  = errors.New("experience profile must be PRIVATE or CORPORATE")
	ErrInvalidLanguage = errors.New("language level must be STANDARD or EASY")
	ErrInvalidContext  = errors.New("experience context is not registered")
)

type Analyzer interface {
	Analyze(analysis.Request) (analysis.Result, error)
}

type Questioner interface {
	Select(questions.SelectionRequest) (questions.Selection, error)
	Render(questions.RenderRequest) (questions.RenderedQuestion, error)
}

type Request struct {
	Text          string `json:"text"`
	Context       string `json:"context,omitempty"`
	Profile       string `json:"profile,omitempty"`
	LanguageLevel string `json:"language_level,omitempty"`
}

type Result struct {
	ContractVersion    string                        `json:"contract_version"`
	ProductProfile     string                        `json:"product_profile"`
	ExperienceMode     string                        `json:"experience_mode"`
	Core               analysis.Result               `json:"core_result"`
	Headline           string                        `json:"headline"`
	Summary            string                        `json:"summary"`
	Dimensions         []DimensionView               `json:"dimensions"`
	Trace              []TraceView                   `json:"explanation_trace"`
	ReflectionQuestion *string                       `json:"reflection_question"`
	Alternatives       []string                      `json:"alternatives"`
	SuggestedQuestions []questions.QuestionCandidate `json:"suggested_questions"`
	Notices            []string                      `json:"notices"`
	Privacy            PrivacyReceipt                `json:"privacy"`
}

type DimensionView struct {
	ID            analysis.DimensionID `json:"id"`
	Label         string               `json:"label"`
	State         string               `json:"state"`
	Score         *float64             `json:"score"`
	Confidence    float64              `json:"confidence"`
	Assessability float64              `json:"assessability"`
}

type TraceView struct {
	RuleID    string               `json:"rule_id"`
	Dimension analysis.DimensionID `json:"dimension"`
	Label     string               `json:"label"`
	Evidence  string               `json:"evidence"`
	Reason    string               `json:"reason"`
	Delta     float64              `json:"delta"`
}

type PrivacyReceipt struct {
	Mode             string `json:"mode"`
	RawTextStored    bool   `json:"raw_text_stored"`
	AnalysisStored   bool   `json:"analysis_stored"`
	ExternalTransfer bool   `json:"external_transfer"`
	AIUsed           bool   `json:"ai_used"`
}

type Service struct {
	analyzer  Analyzer
	questions Questioner
}

func New(analyzer Analyzer, questioner Questioner) *Service {
	return &Service{analyzer: analyzer, questions: questioner}
}

func (s *Service) Analyze(request Request) (Result, error) {
	if s == nil || s.analyzer == nil || s.questions == nil {
		return Result{}, fmt.Errorf("experience dependencies unavailable")
	}
	text := strings.TrimSpace(request.Text)
	if text == "" {
		return Result{}, ErrEmptyText
	}
	if !utf8.ValidString(text) || utf8.RuneCountInString(text) > 10_000 {
		return Result{}, ErrTextTooLong
	}
	profile := strings.ToUpper(strings.TrimSpace(request.Profile))
	if profile == "" {
		profile = "PRIVATE"
	}
	if profile != "PRIVATE" && profile != "CORPORATE" {
		return Result{}, ErrInvalidProfile
	}
	languageLevel := strings.ToUpper(strings.TrimSpace(request.LanguageLevel))
	if languageLevel == "" {
		languageLevel = "STANDARD"
	}
	if languageLevel != "STANDARD" && languageLevel != "EASY" {
		return Result{}, ErrInvalidLanguage
	}
	contextValue := strings.ToUpper(strings.TrimSpace(request.Context))
	if contextValue == "" {
		contextValue = "UNSPECIFIED"
	}
	if !slices.Contains(policy.AnalysisContexts(), policy.AnalysisContextID(contextValue)) {
		return Result{}, ErrInvalidContext
	}
	core, err := s.analyzer.Analyze(analysis.Request{
		Text: text, Locale: analysis.LocaleGerman, Context: analysis.Context(contextValue),
		InputMode: analysis.InputModeText, PresentationProfile: analysis.PresentationProfile(profile),
		AnalysisMode: analysis.AnalysisModeStandard,
	})
	if err != nil {
		return Result{}, fmt.Errorf("analyze experience: %w", err)
	}
	selection, err := s.questions.Select(questions.SelectionRequest{Profile: profile, Limit: 5})
	if err != nil {
		return Result{}, fmt.Errorf("select experience questions: %w", err)
	}
	if languageLevel == "EASY" {
		for index := range selection.Candidates {
			rendered, renderErr := s.questions.Render(questions.RenderRequest{QuestionID: selection.Candidates[index].QuestionID, Profile: profile, Action: "SIMPLIFY"})
			if renderErr != nil {
				return Result{}, fmt.Errorf("render easy experience question: %w", renderErr)
			}
			selection.Candidates[index].Text = rendered.Rendering.Text
		}
	}
	labels := dimensionLabels(profile)
	dimensions := make([]DimensionView, 0, len(analysis.Dimensions()))
	for _, id := range analysis.Dimensions() {
		value := core.Dimensions[id]
		dimensions = append(dimensions, DimensionView{ID: id, Label: labels[id], State: string(value.State), Score: value.Score, Confidence: value.Confidence, Assessability: value.Assessability})
	}
	trace := make([]TraceView, 0, len(core.ContributionTrace))
	for _, item := range core.ContributionTrace {
		trace = append(trace, TraceView{RuleID: item.RuleID, Dimension: item.Dimension, Label: labels[item.Dimension], Evidence: item.Evidence, Reason: item.Reason, Delta: item.Delta})
	}
	headline, summary := explain(profile, core)
	privacyMode := "PRIVATE_TRANSIENT"
	if profile == "CORPORATE" {
		privacyMode = "CORPORATE_TRANSIENT"
	}
	return Result{
		ContractVersion: ContractVersion, ProductProfile: profile, ExperienceMode: "CORE_NO_AI",
		Core: core, Headline: headline, Summary: summary, Dimensions: dimensions, Trace: trace,
		ReflectionQuestion: core.ReflectionQuestion, Alternatives: append([]string(nil), core.Alternatives...),
		SuggestedQuestions: selection.Candidates,
		Notices: []string{
			"Die Auswertung beschreibt ausschließlich diesen Text, nicht dich als Person.",
			"Nicht ausreichend belegte Dimensionen bleiben bewusst ohne Punktwert.",
		},
		Privacy: PrivacyReceipt{Mode: privacyMode, RawTextStored: false, AnalysisStored: false, ExternalTransfer: false, AIUsed: false},
	}, nil
}

func dimensionLabels(profile string) map[analysis.DimensionID]string {
	if profile == "CORPORATE" {
		return map[analysis.DimensionID]string{"AGENCY": "Wirksamkeit", "CONNECTION": "Verbindung", "APPRECIATION": "Wertschätzung", "CLARITY": "Klarheit", "VOLITION": "Handlungsspielraum", "OPENNESS": "Offenheit"}
	}
	return map[analysis.DimensionID]string{"AGENCY": "Selbstwirksamkeit", "CONNECTION": "Verbindung", "APPRECIATION": "Wertschätzung", "CLARITY": "Klarheit", "VOLITION": "Freier Wille", "OPENNESS": "Offenheit"}
}

func explain(profile string, result analysis.Result) (string, string) {
	headline := "Dein aktuelles Sprachbild"
	if profile == "CORPORATE" {
		headline = "Ihr aktueller Sprachkompass"
	}
	if len(result.Patterns) == 0 {
		return headline, "Der Core erkennt in diesem Text noch kein ausreichend gesichertes Muster. Das ist ein transparentes Ergebnis, kein fehlender Mittelwert."
	}
	if len(result.Patterns) == 1 {
		return headline, "Ein sprachliches Muster ist nachvollziehbar belegt. Öffne die Erklärung, um Evidenz und Wirkung auf Textebene zu prüfen."
	}
	return headline, fmt.Sprintf("%d sprachliche Muster sind nachvollziehbar belegt. Sie werden getrennt erklärt und niemals als Eigenschaft einer Person gedeutet.", len(result.Patterns))
}
