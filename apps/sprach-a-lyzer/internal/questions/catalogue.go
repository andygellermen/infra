// Package questions owns the deterministic, non-scoring Question/Answer MVP.
package questions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

const (
	CatalogueVersion            = "0.1"
	CanonicalContractVersion    = "0.1"
	RenderingContractVersion    = "0.1"
	ObservationContractVersion  = "0.1"
	SelectionContractVersion    = "0.1"
	SessionContractVersion      = "0.1"
	CataloguePolicyRegistry     = "0.7"
	minimumSessionQuestionCount = 5
	maximumSessionQuestionCount = 8
)

var (
	questionIDPattern = regexp.MustCompile(`^CQ[0-9]{3}$`)
	canonicalKey      = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

type Catalogue struct {
	Schema          string               `json:"$schema"`
	Version         string               `json:"catalogue_version"`
	Status          string               `json:"status"`
	Locale          string               `json:"locale"`
	VersionContract VersionContract      `json:"version_contract"`
	Questions       []CanonicalQuestion  `json:"questions"`
	Renderings      []Rendering          `json:"renderings"`
	Compositions    []CompositionRule    `json:"composition_rules"`
	Priorities      []QuestionPriority   `json:"question_priorities"`
	Selection       SelectionWeights     `json:"selection_weights"`
	HardGuardrails  []policy.GuardrailID `json:"hard_guardrails"`
}

type VersionContract struct {
	CanonicalQuestion string `json:"canonical_question"`
	QuestionRendering string `json:"question_rendering"`
	QAObservation     string `json:"qa_observation"`
	QuestionSelection string `json:"question_selection"`
	QuestionSession   string `json:"question_session"`
	PolicyRegistry    string `json:"policy_registry"`
}

type CanonicalQuestion struct {
	ContractVersion     string                 `json:"contract_version"`
	ID                  string                 `json:"question_id"`
	PrimaryConstruct    string                 `json:"construct_intent"`
	SecondaryConstructs []string               `json:"secondary_constructs"`
	Phase               string                 `json:"phase"`
	Audience            string                 `json:"audience_scope"`
	RiskLevel           string                 `json:"risk_level"`
	CanonicalSemantics  string                 `json:"canonical_semantics"`
	ProhibitedSemantics []string               `json:"prohibited_semantics"`
	ExpectedDimensions  []analysis.DimensionID `json:"expected_dimensions"`
	QuestionScoreBias   float64                `json:"question_score_bias"`
	Status              string                 `json:"status"`
	Version             int                    `json:"version"`
}

type Rendering struct {
	ContractVersion       string  `json:"contract_version"`
	ID                    string  `json:"rendering_id"`
	QuestionID            string  `json:"question_id"`
	Profile               string  `json:"rendering_profile"`
	LanguageLevel         string  `json:"language_level"`
	Locale                string  `json:"locale"`
	Text                  string  `json:"text"`
	RephraseType          string  `json:"rephrase_type"`
	Leadingness           float64 `json:"leadingness"`
	Specificity           float64 `json:"specificity"`
	IntimacyLevel         float64 `json:"intimacy_level"`
	SpiritualExplicitness float64 `json:"spiritual_explicitness"`
	RelationalWarmth      float64 `json:"relational_warmth"`
	RequiresOptIn         bool    `json:"requires_opt_in"`
	Status                string  `json:"status"`
	Version               int     `json:"version"`
}

type CompositionRule struct {
	QuestionID      string               `json:"question_id"`
	OutputPattern   string               `json:"output_pattern"`
	PhrasesAll      []string             `json:"phrases_all,omitempty"`
	PhrasesAny      []string             `json:"phrases_any,omitempty"`
	AnswerRelevance float64              `json:"answer_relevance"`
	ConstructFit    float64              `json:"construct_fit"`
	BaseConfidence  float64              `json:"base_confidence"`
	Directions      []DimensionDirection `json:"dimension_directions"`
}

type DimensionDirection struct {
	Dimension analysis.DimensionID `json:"dimension"`
	Direction string               `json:"direction"`
}

type QuestionPriority struct {
	QuestionID      string  `json:"question_id"`
	InformationGain float64 `json:"information_gain"`
}

type SelectionWeights struct {
	InformationGain        float64 `json:"information_gain"`
	ConstructGap           float64 `json:"construct_gap"`
	AnswerRelevanceHistory float64 `json:"answer_relevance_history"`
	PhaseFit               float64 `json:"phase_fit"`
	UserPreferenceFit      float64 `json:"user_preference_fit"`
	RedundancyPenalty      float64 `json:"redundancy_penalty"`
	LeadingnessPenalty     float64 `json:"leadingness_penalty"`
	RiskWithoutOptIn       float64 `json:"risk_without_opt_in_penalty"`
}

type CatalogueProvider interface {
	Active(context.Context) (Catalogue, error)
}

type StaticCatalogueProvider struct{ Catalogue Catalogue }

func (s StaticCatalogueProvider) Active(context.Context) (Catalogue, error) { return s.Catalogue, nil }

type embeddedProvider struct{}

func (embeddedProvider) Active(context.Context) (Catalogue, error) {
	return DecodeCatalogue(strings.NewReader(string(assets.QuestionCatalogueV01)))
}

func DecodeCatalogue(reader io.Reader) (Catalogue, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var catalogue Catalogue
	if err := decoder.Decode(&catalogue); err != nil {
		return Catalogue{}, fmt.Errorf("decode question catalogue: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Catalogue{}, fmt.Errorf("decode question catalogue: trailing JSON value")
	}
	if err := catalogue.Validate(); err != nil {
		return Catalogue{}, err
	}
	return catalogue, nil
}

func (c Catalogue) Validate() error {
	if c.Version != CatalogueVersion || c.Status != "APPROVED" || c.Locale != "de-DE" {
		return fmt.Errorf("question catalogue requires version %s, APPROVED status and de-DE locale", CatalogueVersion)
	}
	wantContract := VersionContract{
		CanonicalQuestion: CanonicalContractVersion, QuestionRendering: RenderingContractVersion,
		QAObservation: ObservationContractVersion, QuestionSelection: SelectionContractVersion,
		QuestionSession: SessionContractVersion, PolicyRegistry: CataloguePolicyRegistry,
	}
	if c.VersionContract != wantContract {
		return fmt.Errorf("question catalogue version contract = %+v; want %+v", c.VersionContract, wantContract)
	}
	if len(c.Questions) != maximumSessionQuestionCount {
		return fmt.Errorf("question catalogue contains %d questions; want %d", len(c.Questions), maximumSessionQuestionCount)
	}
	knownDimensions := analysis.Dimensions()
	questions := make(map[string]CanonicalQuestion, len(c.Questions))
	for _, question := range c.Questions {
		if question.ContractVersion != CanonicalContractVersion || !questionIDPattern.MatchString(question.ID) ||
			!canonicalKey.MatchString(question.PrimaryConstruct) || questions[question.ID].ID != "" ||
			!slices.Contains([]string{"P1_ENTRY", "P2_FOLLOWUP_1_10", "P3_ADVANCED_10_20", "P4_INTEGRATION"}, question.Phase) ||
			!slices.Contains([]string{"PRIVATE", "CORPORATE", "BOTH"}, question.Audience) ||
			!slices.Contains([]string{"LOW", "MEDIUM", "HIGH"}, question.RiskLevel) ||
			question.CanonicalSemantics == "" || len(question.ProhibitedSemantics) == 0 || question.QuestionScoreBias != 0 ||
			question.Status != "APPROVED" || question.Version < 1 ||
			!uniqueNonEmpty(question.SecondaryConstructs) || !uniqueNonEmpty(question.ProhibitedSemantics) ||
			!uniqueDimensions(question.ExpectedDimensions, knownDimensions) {
			return fmt.Errorf("invalid canonical question %q", question.ID)
		}
		questions[question.ID] = question
	}
	renderingCounts := make(map[string]int, len(questions))
	renderingKeys := map[string]bool{}
	for _, rendering := range c.Renderings {
		question, exists := questions[rendering.QuestionID]
		key := rendering.QuestionID + ":" + rendering.Profile
		if !exists || rendering.ContractVersion != RenderingContractVersion || rendering.ID == "" || renderingKeys[key] ||
			!slices.Contains([]string{"PRIVATE_STANDARD", "CORPORATE_STANDARD"}, rendering.Profile) || rendering.LanguageLevel != "STANDARD" || rendering.Locale != "de-DE" ||
			rendering.Text == "" || !bounded(rendering.Leadingness) || !bounded(rendering.Specificity) || !bounded(rendering.IntimacyLevel) ||
			rendering.RephraseType != "CANONICAL" || !bounded(rendering.SpiritualExplicitness) || !bounded(rendering.RelationalWarmth) ||
			rendering.Status != "APPROVED" || rendering.Version < 1 || (rendering.Profile == "CORPORATE_STANDARD" && question.Audience == "PRIVATE") ||
			(rendering.Profile == "CORPORATE_STANDARD" && (rendering.IntimacyLevel > .4 || rendering.SpiritualExplicitness != 0)) ||
			(question.RiskLevel == "HIGH" && !rendering.RequiresOptIn) {
			return fmt.Errorf("invalid question rendering %q", rendering.ID)
		}
		renderingKeys[key] = true
		renderingCounts[rendering.QuestionID]++
	}
	for id := range questions {
		if renderingCounts[id] != 2 {
			return fmt.Errorf("question %s requires private and corporate rendering", id)
		}
	}
	priorityCounts := map[string]bool{}
	for _, priority := range c.Priorities {
		if questions[priority.QuestionID].ID == "" || priorityCounts[priority.QuestionID] || !bounded(priority.InformationGain) {
			return fmt.Errorf("invalid question priority %q", priority.QuestionID)
		}
		priorityCounts[priority.QuestionID] = true
	}
	if len(priorityCounts) != len(questions) {
		return fmt.Errorf("question priorities do not cover the catalogue")
	}
	patterns := map[string]bool{}
	compositionCounts := map[string]int{}
	for _, rule := range c.Compositions {
		if questions[rule.QuestionID].ID == "" || !canonicalKey.MatchString(rule.OutputPattern) || patterns[rule.OutputPattern] ||
			len(rule.PhrasesAll)+len(rule.PhrasesAny) == 0 || !uniqueNonEmpty(rule.PhrasesAll) || !uniqueNonEmpty(rule.PhrasesAny) ||
			!bounded(rule.AnswerRelevance) || rule.AnswerRelevance < .55 || !bounded(rule.ConstructFit) ||
			!bounded(rule.BaseConfidence) || !validDirections(rule.Directions, knownDimensions) {
			return fmt.Errorf("invalid question composition %q", rule.OutputPattern)
		}
		patterns[rule.OutputPattern] = true
		compositionCounts[rule.QuestionID]++
	}
	for id := range questions {
		if compositionCounts[id] != 2 {
			return fmt.Errorf("question %s requires confirming and disconfirming compositions", id)
		}
	}
	if !c.Selection.valid() {
		return fmt.Errorf("invalid adaptive selection weights")
	}
	wantGuardrails := []policy.GuardrailID{
		policy.QuestionScoreBiasIsZero, policy.QuestionAloneIsNotAssessable,
		policy.NoPersonDiagnosis, policy.NoTraitClaims, policy.NoEmployeeRanking,
	}
	if !slices.Equal(c.HardGuardrails, wantGuardrails) {
		return fmt.Errorf("question catalogue hard guardrails = %v; want %v", c.HardGuardrails, wantGuardrails)
	}
	return nil
}

func (weights SelectionWeights) valid() bool {
	values := []float64{
		weights.InformationGain, weights.ConstructGap, weights.AnswerRelevanceHistory,
		weights.PhaseFit, weights.UserPreferenceFit, weights.RedundancyPenalty,
		weights.LeadingnessPenalty, weights.RiskWithoutOptIn,
	}
	for _, value := range values {
		if value < 0 || value > 1 {
			return false
		}
	}
	return weights.InformationGain+weights.ConstructGap+weights.AnswerRelevanceHistory+weights.PhaseFit+weights.UserPreferenceFit > 0
}

func validDirections(values []DimensionDirection, known []analysis.DimensionID) bool {
	seen := map[analysis.DimensionID]bool{}
	for _, value := range values {
		if seen[value.Dimension] || !slices.Contains(known, value.Dimension) || !slices.Contains([]string{"POSITIVE", "NEGATIVE"}, value.Direction) {
			return false
		}
		seen[value.Dimension] = true
	}
	return true
}

func uniqueDimensions(values, known []analysis.DimensionID) bool {
	if len(values) == 0 {
		return false
	}
	seen := map[analysis.DimensionID]bool{}
	for _, value := range values {
		if seen[value] || !slices.Contains(known, value) {
			return false
		}
		seen[value] = true
	}
	return true
}

func uniqueNonEmpty(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func bounded(value float64) bool { return value >= 0 && value <= 1 }
