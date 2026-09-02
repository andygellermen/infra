package questions

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

var (
	ErrUnknownQuestion = errors.New("unknown question")
	ErrEmptyAnswer     = errors.New("answer must not be empty")
	ErrAnswerTooLong   = errors.New("answer must not exceed 10,000 characters")
	ErrInvalidProfile  = errors.New("profile must be PRIVATE or CORPORATE")
)

type LanguageAnalyzer interface {
	Analyze(analysis.Request) (analysis.Result, error)
}

type Service struct {
	provider          CatalogueProvider
	renderingProvider RenderingCatalogueProvider
	analyzer          LanguageAnalyzer
}

func New(analyzer LanguageAnalyzer) *Service {
	return NewWithCatalogue(analyzer, embeddedProvider{})
}

func NewDefault() *Service { return New(analysis.NewDefault()) }

func NewWithCatalogue(analyzer LanguageAnalyzer, provider CatalogueProvider) *Service {
	return NewWithCatalogues(analyzer, provider, embeddedRenderingProvider{})
}

func NewWithCatalogues(analyzer LanguageAnalyzer, provider CatalogueProvider, renderingProvider RenderingCatalogueProvider) *Service {
	return &Service{provider: provider, renderingProvider: renderingProvider, analyzer: analyzer}
}

type AnswerRequest struct {
	QuestionID string `json:"question_id"`
	Answer     string `json:"answer"`
	Profile    string `json:"profile,omitempty"`
	Baseline   string `json:"baseline_answer,omitempty"`
}

type Observation struct {
	ContractVersion      string              `json:"contract_version"`
	QuestionID           string              `json:"question_id"`
	Answer               string              `json:"answer"`
	PrimaryConstruct     string              `json:"primary_construct"`
	AnswerRelevance      float64             `json:"answer_relevance"`
	RelevanceBand        string              `json:"relevance_band"`
	ConstructFit         float64             `json:"construct_fit"`
	ResponseIndependence float64             `json:"response_independence"`
	ContextGain          float64             `json:"context_gain"`
	QuestionScoreBias    float64             `json:"question_score_bias"`
	QAPatterns           []string            `json:"qa_patterns"`
	DimensionEvidence    []DimensionEvidence `json:"dimension_evidence"`
	Assessability        string              `json:"assessability"`
	InferenceLevel       string              `json:"inference_level"`
	ExplanationKey       string              `json:"explanation_key"`
	AnalysisResult       analysis.Result     `json:"analysis_result"`
}

type DimensionEvidence struct {
	Dimension  analysis.DimensionID `json:"dimension"`
	Direction  string               `json:"direction"`
	Confidence float64              `json:"confidence"`
	Scoring    bool                 `json:"scoring"`
}

type SelectionRequest struct {
	Profile                    string             `json:"profile,omitempty"`
	PreferredPhase             string             `json:"preferred_phase,omitempty"`
	ConstructGaps              []string           `json:"construct_gaps,omitempty"`
	PreferredConstructs        []string           `json:"preferred_constructs,omitempty"`
	CompletedQuestionIDs       []string           `json:"completed_question_ids,omitempty"`
	CompletedConstructs        []string           `json:"completed_constructs,omitempty"`
	AnswerRelevanceByConstruct map[string]float64 `json:"answer_relevance_by_construct,omitempty"`
	DeepReflectionOptIn        bool               `json:"deep_reflection_opt_in,omitempty"`
	Limit                      int                `json:"limit,omitempty"`
}

type Selection struct {
	ContractVersion string              `json:"contract_version"`
	Candidates      []QuestionCandidate `json:"candidates"`
}

type QuestionCandidate struct {
	QuestionID         string                 `json:"question_id"`
	Text               string                 `json:"text"`
	PrimaryConstruct   string                 `json:"primary_construct"`
	Phase              string                 `json:"phase"`
	RiskLevel          string                 `json:"risk_level"`
	ExpectedDimensions []analysis.DimensionID `json:"expected_dimensions"`
	SelectionScore     float64                `json:"selection_score"`
	Offered            bool                   `json:"offered"`
}

type SessionRequest struct {
	SessionID string          `json:"session_id"`
	Profile   string          `json:"profile,omitempty"`
	Pairs     []AnswerRequest `json:"pairs"`
}

type Session struct {
	ContractVersion string              `json:"contract_version"`
	SessionID       string              `json:"session_id"`
	Observations    []Observation       `json:"observations"`
	Trajectories    []Trajectory        `json:"trajectories"`
	InferenceLevel  string              `json:"inference_level"`
	InferenceClaims []InferenceClaim    `json:"inference_claims"`
	NextQuestions   []QuestionCandidate `json:"next_questions"`
}

type Trajectory struct {
	Dimension             analysis.DimensionID `json:"dimension"`
	ObservationIndexes    []int                `json:"observation_indexes"`
	LanguagePatternChange string               `json:"language_pattern_change"`
	Confidence            float64              `json:"confidence"`
	InferenceLevel        string               `json:"inference_level"`
}

type InferenceClaim struct {
	Level          string `json:"level"`
	ExplanationKey string `json:"explanation_key"`
}

func (s *Service) Analyze(request AnswerRequest) (Observation, error) {
	catalogue, err := s.activeCatalogue()
	if err != nil {
		return Observation{}, err
	}
	return s.analyzeWithCatalogue(catalogue, request)
}

func (s *Service) analyzeWithCatalogue(catalogue Catalogue, request AnswerRequest) (Observation, error) {
	question, rendering, err := resolveQuestion(catalogue, request.QuestionID, request.Profile)
	if err != nil {
		return Observation{}, err
	}
	answer := strings.TrimSpace(request.Answer)
	if answer == "" {
		return Observation{}, ErrEmptyAnswer
	}
	if !utf8.ValidString(answer) || utf8.RuneCountInString(answer) > 10_000 {
		return Observation{}, ErrAnswerTooLong
	}
	profile := analysis.ProfilePrivate
	if rendering.Profile == "CORPORATE_STANDARD" {
		profile = analysis.ProfileCorporate
	}
	result, err := s.analyzer.Analyze(analysis.Request{
		Text: answer, Locale: analysis.LocaleGerman, Context: "COACHING",
		InputMode: analysis.InputModeText, PresentationProfile: profile, AnalysisMode: analysis.AnalysisModeStandard,
	})
	if err != nil {
		return Observation{}, fmt.Errorf("analyze answer language: %w", err)
	}
	rule, matched, err := matchComposition(catalogue, question.ID, answer)
	if err != nil {
		return Observation{}, err
	}
	independence := responseIndependence(answer)
	relevance, fit := .18, .1
	assessability, inference, explanation := "NOT_ASSESSABLE", "C0", "QA_OBSERVATION_ONLY"
	patterns := []string{}
	evidence := []DimensionEvidence{}
	if matched {
		relevance, fit = rule.AnswerRelevance, rule.ConstructFit
		assessability, inference, explanation = "ASSESSABLE", "C1", "QA_QUESTION_CONDITIONED_RELEVANCE"
		patterns = []string{rule.OutputPattern}
	} else if wordCount(answer) <= 2 {
		relevance, fit = .52, .35
		assessability, inference, explanation = "WEAK", "C1", "QA_SHORT_ANSWER_WEAK"
	}
	contextGain := round(math.Min(.15, .15*relevance*fit*rendering.Specificity*(1-rendering.Leadingness)*independence), 6)
	if matched {
		for _, direction := range rule.Directions {
			evidence = append(evidence, DimensionEvidence{
				Dimension: direction.Dimension, Direction: direction.Direction,
				Confidence: round(math.Min(1, rule.BaseConfidence*(1+contextGain)), 6), Scoring: false,
			})
		}
	}
	return Observation{
		ContractVersion: ObservationContractVersion, QuestionID: question.ID, Answer: answer,
		PrimaryConstruct: question.PrimaryConstruct, AnswerRelevance: relevance, RelevanceBand: relevanceBand(relevance),
		ConstructFit: fit, ResponseIndependence: independence, ContextGain: contextGain, QuestionScoreBias: 0,
		QAPatterns: patterns, DimensionEvidence: evidence, Assessability: assessability,
		InferenceLevel: inference, ExplanationKey: explanation, AnalysisResult: result,
	}, nil
}

func (s *Service) Select(request SelectionRequest) (Selection, error) {
	catalogue, err := s.activeCatalogue()
	if err != nil {
		return Selection{}, err
	}
	return selectQuestions(catalogue, request)
}

func (s *Service) ComposeSession(request SessionRequest) (Session, error) {
	if strings.TrimSpace(request.SessionID) == "" || len(request.Pairs) == 0 || len(request.Pairs) > maximumSessionQuestionCount {
		return Session{}, fmt.Errorf("session requires an ID and one to %d progressive answers", maximumSessionQuestionCount)
	}
	catalogue, err := s.activeCatalogue()
	if err != nil {
		return Session{}, err
	}
	result := Session{
		ContractVersion: SessionContractVersion, SessionID: request.SessionID,
		Observations: []Observation{}, Trajectories: []Trajectory{}, InferenceLevel: "C0",
		InferenceClaims: []InferenceClaim{{Level: "C0", ExplanationKey: "QA_OBSERVATIONAL_LANGUAGE_ONLY"}},
		NextQuestions:   []QuestionCandidate{},
	}
	completed := make([]string, 0, len(request.Pairs))
	hasC1 := false
	hasC2 := false
	for _, pair := range request.Pairs {
		if pair.Profile == "" {
			pair.Profile = request.Profile
		}
		observation, err := s.analyzeWithCatalogue(catalogue, pair)
		if err != nil {
			return Session{}, err
		}
		result.Observations = append(result.Observations, observation)
		completed = append(completed, pair.QuestionID)
		if observation.InferenceLevel == "C1" {
			result.InferenceLevel = "C1"
			hasC1 = true
		}
		if strings.TrimSpace(pair.Baseline) != "" && len(observation.QAPatterns) > 0 {
			_, baselineMatched, matchErr := matchComposition(catalogue, pair.QuestionID, pair.Baseline)
			if matchErr != nil {
				return Session{}, matchErr
			}
			if !baselineMatched {
				hasC2 = true
			}
		}
	}
	if hasC1 {
		result.InferenceClaims = append(result.InferenceClaims, InferenceClaim{Level: "C1", ExplanationKey: "QA_QUESTION_CONDITIONED_ASSOCIATION"})
	}
	if hasC2 {
		result.InferenceLevel = "C2"
		result.InferenceClaims = append(result.InferenceClaims, InferenceClaim{Level: "C2", ExplanationKey: "QA_ELICITATION_ASSOCIATION_HEDGED"})
	}
	result.Trajectories = buildTrajectories(result.Observations)
	if len(result.Trajectories) > 0 {
		result.InferenceLevel = "C3"
		result.InferenceClaims = append(result.InferenceClaims, InferenceClaim{Level: "C3", ExplanationKey: "QA_WITHIN_SESSION_TEMPORAL_ASSOCIATION"})
	}
	selection, err := selectQuestions(catalogue, SelectionRequest{
		Profile: request.Profile, CompletedQuestionIDs: completed, Limit: minimumSessionQuestionCount,
	})
	if err != nil && !errors.Is(err, errInsufficientCandidates) {
		return Session{}, err
	}
	result.NextQuestions = selection.Candidates
	return result, nil
}

var errInsufficientCandidates = errors.New("insufficient question candidates")

func selectQuestions(catalogue Catalogue, request SelectionRequest) (Selection, error) {
	profile, err := resolveProfile(request.Profile)
	if err != nil {
		return Selection{}, err
	}
	limit := request.Limit
	if limit == 0 {
		limit = maximumSessionQuestionCount
	}
	if limit < minimumSessionQuestionCount || limit > maximumSessionQuestionCount {
		return Selection{}, fmt.Errorf("question selection limit must be between %d and %d", minimumSessionQuestionCount, maximumSessionQuestionCount)
	}
	completedIDs, completedConstructs := stringSet(request.CompletedQuestionIDs), stringSet(request.CompletedConstructs)
	gaps, preferences := stringSet(request.ConstructGaps), stringSet(request.PreferredConstructs)
	informationGain := make(map[string]float64, len(catalogue.Priorities))
	for _, priority := range catalogue.Priorities {
		informationGain[priority.QuestionID] = priority.InformationGain
	}
	candidates := make([]QuestionCandidate, 0, len(catalogue.Questions))
	for _, question := range catalogue.Questions {
		if completedIDs[question.ID] || (profile == "CORPORATE_STANDARD" && question.Audience == "PRIVATE") ||
			(question.RiskLevel == "HIGH" && !request.DeepReflectionOptIn) {
			continue
		}
		rendering, ok := findRendering(catalogue, question.ID, profile)
		if !ok || (rendering.RequiresOptIn && !request.DeepReflectionOptIn) {
			continue
		}
		phaseFit := .5
		if request.PreferredPhase == "" || request.PreferredPhase == question.Phase {
			phaseFit = 1
		}
		gapFit := .5
		if gaps[question.PrimaryConstruct] {
			gapFit = 1
		}
		preferenceFit := .5
		if preferences[question.PrimaryConstruct] {
			preferenceFit = 1
		}
		history := request.AnswerRelevanceByConstruct[question.PrimaryConstruct]
		if history == 0 {
			history = .5
		}
		redundancy := 0.0
		if completedConstructs[question.PrimaryConstruct] {
			redundancy = 1
		}
		risk := 0.0
		if question.RiskLevel == "HIGH" && !request.DeepReflectionOptIn {
			risk = 1
		}
		weights := catalogue.Selection
		score := weights.InformationGain*informationGain[question.ID] + weights.ConstructGap*gapFit +
			weights.AnswerRelevanceHistory*history + weights.PhaseFit*phaseFit + weights.UserPreferenceFit*preferenceFit -
			weights.RedundancyPenalty*redundancy - weights.LeadingnessPenalty*rendering.Leadingness - weights.RiskWithoutOptIn*risk
		candidates = append(candidates, QuestionCandidate{
			QuestionID: question.ID, Text: rendering.Text, PrimaryConstruct: question.PrimaryConstruct,
			Phase: question.Phase, RiskLevel: question.RiskLevel,
			ExpectedDimensions: append([]analysis.DimensionID(nil), question.ExpectedDimensions...),
			SelectionScore:     round(score, 6), Offered: true,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].SelectionScore == candidates[j].SelectionScore {
			return candidates[i].QuestionID < candidates[j].QuestionID
		}
		return candidates[i].SelectionScore > candidates[j].SelectionScore
	})
	if len(candidates) < minimumSessionQuestionCount {
		return Selection{ContractVersion: SelectionContractVersion, Candidates: candidates}, errInsufficientCandidates
	}
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return Selection{ContractVersion: SelectionContractVersion, Candidates: candidates}, nil
}

func buildTrajectories(observations []Observation) []Trajectory {
	type point struct {
		index      int
		direction  string
		confidence float64
	}
	points := map[analysis.DimensionID][]point{}
	for index, observation := range observations {
		for _, evidence := range observation.DimensionEvidence {
			points[evidence.Dimension] = append(points[evidence.Dimension], point{index: index, direction: evidence.Direction, confidence: evidence.Confidence})
		}
	}
	trajectories := []Trajectory{}
	for _, dimension := range analysis.Dimensions() {
		values := points[dimension]
		if len(values) < 2 || values[0].direction == values[len(values)-1].direction {
			continue
		}
		indexes := make([]int, len(values))
		confidence := 1.0
		for index, value := range values {
			indexes[index] = value.index
			confidence = math.Min(confidence, value.confidence)
		}
		trajectories = append(trajectories, Trajectory{
			Dimension: dimension, ObservationIndexes: indexes,
			LanguagePatternChange: values[0].direction + "_TO_" + values[len(values)-1].direction,
			Confidence:            round(confidence, 6), InferenceLevel: "C3",
		})
	}
	return trajectories
}

func (s *Service) activeCatalogue() (Catalogue, error) {
	if s == nil || s.provider == nil || s.analyzer == nil {
		return Catalogue{}, fmt.Errorf("question service dependencies are unavailable")
	}
	catalogue, err := s.provider.Active(context.Background())
	if err != nil {
		return Catalogue{}, fmt.Errorf("load active question catalogue: %w", err)
	}
	if err := catalogue.Validate(); err != nil {
		return Catalogue{}, fmt.Errorf("validate active question catalogue: %w", err)
	}
	return catalogue, nil
}

func resolveQuestion(catalogue Catalogue, id, profile string) (CanonicalQuestion, Rendering, error) {
	resolvedProfile, err := resolveProfile(profile)
	if err != nil {
		return CanonicalQuestion{}, Rendering{}, err
	}
	for _, question := range catalogue.Questions {
		if question.ID != id {
			continue
		}
		rendering, ok := findRendering(catalogue, id, resolvedProfile)
		if !ok {
			return CanonicalQuestion{}, Rendering{}, fmt.Errorf("%w: rendering for %s", ErrUnknownQuestion, id)
		}
		return question, rendering, nil
	}
	return CanonicalQuestion{}, Rendering{}, fmt.Errorf("%w: %s", ErrUnknownQuestion, id)
}

func findRendering(catalogue Catalogue, id, profile string) (Rendering, bool) {
	for _, rendering := range catalogue.Renderings {
		if rendering.QuestionID == id && rendering.Profile == profile {
			return rendering, true
		}
	}
	return Rendering{}, false
}

func matchComposition(catalogue Catalogue, questionID, answer string) (CompositionRule, bool, error) {
	normalized := normalizeText(answer)
	var matched *CompositionRule
	for index := range catalogue.Compositions {
		rule := &catalogue.Compositions[index]
		if rule.QuestionID != questionID || !matchesPhrases(normalized, rule.PhrasesAll, rule.PhrasesAny) {
			continue
		}
		if matched != nil {
			return CompositionRule{}, false, fmt.Errorf("answer ambiguously matches %s and %s", matched.OutputPattern, rule.OutputPattern)
		}
		matched = rule
	}
	if matched == nil {
		return CompositionRule{}, false, nil
	}
	return *matched, true, nil
}

func matchesPhrases(text string, all, any []string) bool {
	for _, phrase := range all {
		if !strings.Contains(text, normalizeText(phrase)) {
			return false
		}
	}
	if len(any) == 0 {
		return true
	}
	for _, phrase := range any {
		if strings.Contains(text, normalizeText(phrase)) {
			return true
		}
	}
	return false
}

func normalizeText(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func wordCount(value string) int {
	return len(strings.FieldsFunc(value, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) }))
}

func responseIndependence(answer string) float64 {
	words := wordCount(answer)
	if words <= 2 {
		return .25
	}
	return round(math.Min(1, .45+.05*float64(min(words, 11))), 6)
}

func relevanceBand(value float64) string {
	switch {
	case value < .35:
		return "OFF_TOPIC"
	case value < .55:
		return "WEAK"
	case value < .75:
		return "RELEVANT"
	default:
		return "STRONGLY_RELEVANT"
	}
}

func resolveProfile(value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "CORPORATE" || value == "CORPORATE_STANDARD" {
		return "CORPORATE_STANDARD", nil
	}
	if value == "" || value == "PRIVATE" || value == "PRIVATE_STANDARD" {
		return "PRIVATE_STANDARD", nil
	}
	return "", ErrInvalidProfile
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func round(value float64, places int) float64 {
	factor := math.Pow10(places)
	return math.Round(value*factor) / factor
}
