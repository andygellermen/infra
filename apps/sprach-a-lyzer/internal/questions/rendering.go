package questions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"slices"
	"strings"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

const (
	RenderingCatalogueVersion = "0.1"
	RenderingV02Contract      = "0.2"
	RenderingResultContract   = "0.1"
)

var (
	ErrInvalidRenderAction = errors.New("render action must be DEFAULT, SIMPLIFY or REPHRASE")
	renderingIDPattern     = regexp.MustCompile(`^CQ[0-9]{3}:(PRIVATE|CORPORATE):(STANDARD|EASY|DEEP_REFLECTIVE):(CANONICAL|SIMPLIFY|ALTERNATIVE_PERSPECTIVE):v[1-9][0-9]*$`)
)

type RenderingCatalogue struct {
	Schema                    string                 `json:"$schema"`
	Version                   string                 `json:"catalogue_version"`
	Status                    string                 `json:"status"`
	Locale                    string                 `json:"locale"`
	CanonicalQuestionContract string                 `json:"canonical_question_contract"`
	RenderingContract         string                 `json:"rendering_contract"`
	CanonicalIntents          map[string]string      `json:"canonical_intents"`
	QualityLimits             RenderingQualityLimits `json:"quality_limits"`
	Renderings                []QuestionRenderingV02 `json:"renderings"`
	HardGuardrails            []policy.GuardrailID   `json:"hard_guardrails"`
}

type RenderingQualityLimits struct {
	LeadingnessDelta           float64 `json:"leadingness_delta"`
	IntimacyDelta              float64 `json:"intimacy_delta"`
	SpiritualExplicitnessDelta float64 `json:"spiritual_explicitness_delta"`
}

type QuestionRenderingV02 struct {
	ContractVersion       string  `json:"contract_version"`
	ID                    string  `json:"rendering_id"`
	QuestionID            string  `json:"question_id"`
	ConstructIntent       string  `json:"construct_intent"`
	Profile               string  `json:"profile"`
	PresentationMode      string  `json:"presentation_mode"`
	Variant               string  `json:"variant"`
	LanguageLevel         string  `json:"language_level"`
	Locale                string  `json:"locale"`
	Text                  string  `json:"text"`
	Leadingness           float64 `json:"leadingness"`
	Specificity           float64 `json:"specificity"`
	IntimacyLevel         float64 `json:"intimacy_level"`
	SpiritualExplicitness float64 `json:"spiritual_explicitness"`
	RelationalWarmth      float64 `json:"relational_warmth"`
	RequiresOptIn         bool    `json:"requires_opt_in"`
	SemanticEquivalence   bool    `json:"semantic_equivalence"`
	Scoring               bool    `json:"scoring"`
	Status                string  `json:"status"`
	Version               int     `json:"version"`
}

type RenderRequest struct {
	QuestionID          string `json:"question_id"`
	Profile             string `json:"profile,omitempty"`
	Action              string `json:"action,omitempty"`
	DeepReflectionOptIn bool   `json:"deep_reflection_opt_in,omitempty"`
}

type RenderingQuality struct {
	IntentPreserved            bool    `json:"intent_preserved"`
	LeadingnessDelta           float64 `json:"leadingness_delta"`
	IntimacyDelta              float64 `json:"intimacy_delta"`
	SpiritualExplicitnessDelta float64 `json:"spiritual_explicitness_delta"`
	WithinLimits               bool    `json:"within_limits"`
	Scoring                    bool    `json:"scoring"`
}

type RenderedQuestion struct {
	ContractVersion          string               `json:"contract_version"`
	CanonicalQuestionVersion string               `json:"canonical_question_version"`
	Rendering                QuestionRenderingV02 `json:"rendering"`
	Quality                  RenderingQuality     `json:"quality"`
	FallbackApplied          bool                 `json:"fallback_applied"`
	FallbackReason           string               `json:"fallback_reason,omitempty"`
}

type RenderingCatalogueProvider interface {
	Active(context.Context) (RenderingCatalogue, error)
}

type StaticRenderingCatalogueProvider struct{ Catalogue RenderingCatalogue }

func (s StaticRenderingCatalogueProvider) Active(context.Context) (RenderingCatalogue, error) {
	return s.Catalogue, nil
}

type embeddedRenderingProvider struct{}

func (embeddedRenderingProvider) Active(context.Context) (RenderingCatalogue, error) {
	return DecodeRenderingCatalogue(strings.NewReader(string(assets.QuestionRenderingCatalogueV01)))
}

func DecodeRenderingCatalogue(reader io.Reader) (RenderingCatalogue, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var catalogue RenderingCatalogue
	if err := decoder.Decode(&catalogue); err != nil {
		return RenderingCatalogue{}, fmt.Errorf("decode question rendering catalogue: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return RenderingCatalogue{}, fmt.Errorf("decode question rendering catalogue: trailing JSON value")
	}
	if err := catalogue.Validate(); err != nil {
		return RenderingCatalogue{}, err
	}
	return catalogue, nil
}

func (c RenderingCatalogue) Validate() error {
	if c.Version != RenderingCatalogueVersion || c.Status != "APPROVED" || c.Locale != "de-DE" ||
		c.CanonicalQuestionContract != CanonicalContractVersion || c.RenderingContract != RenderingV02Contract ||
		len(c.CanonicalIntents) != maximumSessionQuestionCount || len(c.Renderings) < 40 || !c.QualityLimits.valid() {
		return fmt.Errorf("invalid question rendering catalogue envelope")
	}
	wantGuardrails := []policy.GuardrailID{
		policy.QuestionScoreBiasIsZero, policy.CorporateHasNoCanonicalFallback,
		policy.NoPersonDiagnosis, policy.NoTraitClaims, policy.NoEmployeeRanking,
	}
	if !slices.Equal(c.HardGuardrails, wantGuardrails) {
		return fmt.Errorf("question rendering guardrails = %v; want %v", c.HardGuardrails, wantGuardrails)
	}
	seen := map[string]bool{}
	coverage := map[string]bool{}
	for _, rendering := range c.Renderings {
		if rendering.ContractVersion != RenderingV02Contract || !renderingIDPattern.MatchString(rendering.ID) ||
			seen[rendering.ID] || c.CanonicalIntents[rendering.QuestionID] != rendering.ConstructIntent ||
			!slices.Contains([]string{"PRIVATE", "CORPORATE"}, rendering.Profile) ||
			!slices.Contains([]string{"STANDARD", "EASY", "DEEP_REFLECTIVE"}, rendering.PresentationMode) ||
			!slices.Contains([]string{"CANONICAL", "SIMPLIFY", "ALTERNATIVE_PERSPECTIVE"}, rendering.Variant) ||
			!slices.Contains([]string{"STANDARD", "EASY"}, rendering.LanguageLevel) || rendering.Locale != "de-DE" ||
			strings.TrimSpace(rendering.Text) == "" || len([]rune(rendering.Text)) > 280 ||
			!bounded(rendering.Leadingness) || !bounded(rendering.Specificity) || !bounded(rendering.IntimacyLevel) ||
			!bounded(rendering.SpiritualExplicitness) || !bounded(rendering.RelationalWarmth) ||
			!rendering.SemanticEquivalence || rendering.Scoring || !slices.Contains([]string{"APPROVED", "ARCHIVED"}, rendering.Status) || rendering.Version < 1 {
			return fmt.Errorf("invalid question rendering %q", rendering.ID)
		}
		if rendering.Profile == "CORPORATE" && (rendering.SpiritualExplicitness != 0 || rendering.IntimacyLevel > .4 || rendering.PresentationMode == "DEEP_REFLECTIVE") {
			return fmt.Errorf("corporate rendering %q violates profile isolation", rendering.ID)
		}
		if rendering.PresentationMode == "DEEP_REFLECTIVE" && (rendering.Profile != "PRIVATE" || !rendering.RequiresOptIn) {
			return fmt.Errorf("deep rendering %q requires private opt-in", rendering.ID)
		}
		if containsProhibitedRenderingClaim(rendering.Text) {
			return fmt.Errorf("rendering %q contains prohibited claim language", rendering.ID)
		}
		seen[rendering.ID] = true
		if rendering.Status == "APPROVED" {
			coverage[rendering.QuestionID+":"+rendering.Profile+":"+rendering.PresentationMode+":"+rendering.Variant] = true
		}
	}
	for questionID := range c.CanonicalIntents {
		for _, key := range []string{
			":PRIVATE:STANDARD:CANONICAL", ":PRIVATE:EASY:SIMPLIFY",
			":CORPORATE:STANDARD:CANONICAL", ":CORPORATE:EASY:SIMPLIFY",
		} {
			if !coverage[questionID+key] {
				return fmt.Errorf("question %s lacks required profile rendering %s", questionID, key)
			}
		}
	}
	return nil
}

func (limits RenderingQualityLimits) valid() bool {
	return bounded(limits.LeadingnessDelta) && bounded(limits.IntimacyDelta) && bounded(limits.SpiritualExplicitnessDelta)
}

func containsProhibitedRenderingClaim(text string) bool {
	normalized := strings.ToLower(text)
	for _, phrase := range []string{"diagnose", "persönlichkeitstyp", "ungeeignet", "leistungsbewertung"} {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func (s *Service) Render(request RenderRequest) (RenderedQuestion, error) {
	canonical, err := s.activeCatalogue()
	if err != nil {
		return RenderedQuestion{}, err
	}
	renderings, err := s.activeRenderingCatalogue()
	if err != nil {
		return RenderedQuestion{}, err
	}
	question, _, err := resolveQuestion(canonical, request.QuestionID, request.Profile)
	if err != nil {
		return RenderedQuestion{}, err
	}
	if renderings.CanonicalIntents[question.ID] != question.PrimaryConstruct {
		return RenderedQuestion{}, fmt.Errorf("rendering intent drift for %s", question.ID)
	}
	profile, err := resolveProfile(request.Profile)
	if err != nil {
		return RenderedQuestion{}, err
	}
	profile = strings.TrimSuffix(profile, "_STANDARD")
	action := strings.ToUpper(strings.TrimSpace(request.Action))
	if action == "" {
		action = "DEFAULT"
	}
	mode, variant := "STANDARD", "CANONICAL"
	switch action {
	case "DEFAULT":
	case "SIMPLIFY":
		mode, variant = "EASY", "SIMPLIFY"
	case "REPHRASE":
		mode, variant = "STANDARD", "ALTERNATIVE_PERSPECTIVE"
		if profile == "PRIVATE" && request.DeepReflectionOptIn {
			mode, variant = "DEEP_REFLECTIVE", "ALTERNATIVE_PERSPECTIVE"
		}
	default:
		return RenderedQuestion{}, ErrInvalidRenderAction
	}
	selected, ok := findApprovedRendering(renderings, question.ID, profile, mode, variant)
	fallback, fallbackReason := false, ""
	if !ok {
		fallback = true
		if action == "REPHRASE" && !request.DeepReflectionOptIn {
			fallbackReason = "DEEP_REFLECTION_OPT_IN_REQUIRED"
		} else {
			fallbackReason = "NO_APPROVED_SAFE_VARIANT"
		}
		selected, ok = findApprovedRendering(renderings, question.ID, profile, "STANDARD", "CANONICAL")
	}
	if !ok {
		return RenderedQuestion{}, fmt.Errorf("no approved safe rendering for %s/%s", question.ID, profile)
	}
	baseline, ok := findApprovedRendering(renderings, question.ID, profile, "STANDARD", "CANONICAL")
	if !ok {
		return RenderedQuestion{}, fmt.Errorf("no canonical profile baseline for %s/%s", question.ID, profile)
	}
	quality := RenderingQuality{
		IntentPreserved:            selected.ConstructIntent == question.PrimaryConstruct,
		LeadingnessDelta:           absDelta(selected.Leadingness, baseline.Leadingness),
		IntimacyDelta:              absDelta(selected.IntimacyLevel, baseline.IntimacyLevel),
		SpiritualExplicitnessDelta: absDelta(selected.SpiritualExplicitness, baseline.SpiritualExplicitness),
		Scoring:                    false,
	}
	quality.WithinLimits = quality.IntentPreserved && quality.LeadingnessDelta <= renderings.QualityLimits.LeadingnessDelta &&
		quality.IntimacyDelta <= renderings.QualityLimits.IntimacyDelta &&
		quality.SpiritualExplicitnessDelta <= renderings.QualityLimits.SpiritualExplicitnessDelta
	if !quality.WithinLimits {
		return RenderedQuestion{}, fmt.Errorf("rendering %s exceeds approved quality delta", selected.ID)
	}
	return RenderedQuestion{
		ContractVersion: RenderingResultContract, CanonicalQuestionVersion: CanonicalContractVersion,
		Rendering: selected, Quality: quality, FallbackApplied: fallback, FallbackReason: fallbackReason,
	}, nil
}

func (s *Service) activeRenderingCatalogue() (RenderingCatalogue, error) {
	if s == nil || s.renderingProvider == nil {
		return RenderingCatalogue{}, fmt.Errorf("question rendering provider is unavailable")
	}
	catalogue, err := s.renderingProvider.Active(context.Background())
	if err != nil {
		return RenderingCatalogue{}, fmt.Errorf("load active question rendering catalogue: %w", err)
	}
	if err := catalogue.Validate(); err != nil {
		return RenderingCatalogue{}, fmt.Errorf("validate active question rendering catalogue: %w", err)
	}
	return catalogue, nil
}

func findApprovedRendering(c RenderingCatalogue, questionID, profile, mode, variant string) (QuestionRenderingV02, bool) {
	for _, rendering := range c.Renderings {
		if rendering.QuestionID == questionID && rendering.Profile == profile && rendering.PresentationMode == mode && rendering.Variant == variant && rendering.Status == "APPROVED" {
			return rendering, true
		}
	}
	return QuestionRenderingV02{}, false
}

func absDelta(left, right float64) float64 { return math.Abs(left - right) }
