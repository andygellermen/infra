// Package ontology owns the immutable Construct Ontology contract and its
// deterministic, non-scoring runtime definitions.
package ontology

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/dimension"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

const (
	ContractVersion               = "0.2"
	ContractPolicyRegistryVersion = "0.6"
)

var patternID = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

type Catalogue struct {
	Schema         string               `json:"$schema"`
	Version        string               `json:"ontology_version"`
	Status         string               `json:"status"`
	Locale         string               `json:"locale"`
	PolicyRegistry string               `json:"policy_registry"`
	Constructs     []Definition         `json:"constructs"`
	Compositions   []Composition        `json:"compositions"`
	HardGuardrails []policy.GuardrailID `json:"hard_guardrails"`
}

type Definition struct {
	ID               policy.ConstructID `json:"id"`
	Layer            string             `json:"layer"`
	InferenceClass   string             `json:"inference_class"`
	Definition       string             `json:"definition"`
	Evidence         []string           `json:"evidence"`
	NonEvidence      []string           `json:"non_evidence"`
	AllowedClaim     string             `json:"allowed_claim"`
	ProhibitedClaims []string           `json:"prohibited_claims"`
	ClaimMode        string             `json:"claim_mode"`
	Assessability    Assessability      `json:"assessability"`
	DimensionLinks   []dimension.ID     `json:"dimension_links"`
	CoreScoring      bool               `json:"core_scoring"`
	RuntimeSignals   []RuntimeSignal    `json:"runtime_signals"`
}

type Assessability struct {
	MinimumDistinctEvidence int  `json:"minimum_distinct_evidence"`
	QuestionContextRequired bool `json:"question_context_required"`
	LongitudinalRequired    bool `json:"longitudinal_required"`
}

type RuntimeSignal struct {
	PhrasesAll          []string                     `json:"phrases_all,omitempty"`
	PhrasesAny          []string                     `json:"phrases_any,omitempty"`
	Actors              []policy.ActorID             `json:"actors,omitempty"`
	Modalities          []policy.ModalityID          `json:"modalities,omitempty"`
	TargetTypes         []policy.TargetTypeID        `json:"target_types,omitempty"`
	ExpectationSources  []policy.ExpectationSourceID `json:"expectation_sources,omitempty"`
	PropositionFeatures []string                     `json:"proposition_features_all,omitempty"`
	SelectedSenses      []string                     `json:"selected_senses_any,omitempty"`
}

type Composition struct {
	OutputPattern         string                       `json:"output_pattern"`
	RequiredConstructs    []policy.ConstructID         `json:"required_constructs"`
	Ordered               bool                         `json:"ordered"`
	RelationsAny          []policy.DiscourseRelationID `json:"relations_any"`
	MaximumPropositionGap int                          `json:"maximum_proposition_gap"`
}

func Decode(reader io.Reader) (Catalogue, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var catalogue Catalogue
	if err := decoder.Decode(&catalogue); err != nil {
		return Catalogue{}, fmt.Errorf("decode construct ontology: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Catalogue{}, fmt.Errorf("decode construct ontology: trailing JSON value")
	}
	if err := catalogue.Validate(); err != nil {
		return Catalogue{}, err
	}
	return catalogue, nil
}

func (c Catalogue) Validate() error {
	if c.Version != ContractVersion || c.Status != "APPROVED" || c.Locale != "de-DE" || c.PolicyRegistry != ContractPolicyRegistryVersion {
		return fmt.Errorf("construct ontology requires version %s, APPROVED status, de-DE locale and policy %s", ContractVersion, ContractPolicyRegistryVersion)
	}
	want := policy.Constructs()
	if len(c.Constructs) != len(want) {
		return fmt.Errorf("construct ontology contains %d constructs; want %d", len(c.Constructs), len(want))
	}
	seen := make(map[policy.ConstructID]bool, len(c.Constructs))
	for _, definition := range c.Constructs {
		if seen[definition.ID] || !slices.Contains(want, definition.ID) {
			return fmt.Errorf("construct ontology contains duplicate or unknown ID %q", definition.ID)
		}
		seen[definition.ID] = true
		if err := definition.Validate(); err != nil {
			return err
		}
	}
	for _, id := range want {
		if !seen[id] {
			return fmt.Errorf("construct ontology lacks canonical ID %q", id)
		}
	}
	patterns := map[string]bool{}
	for _, composition := range c.Compositions {
		if err := composition.Validate(seen); err != nil {
			return err
		}
		if patterns[composition.OutputPattern] {
			return fmt.Errorf("duplicate composition pattern %q", composition.OutputPattern)
		}
		patterns[composition.OutputPattern] = true
	}
	wantGuardrails := []policy.GuardrailID{policy.ConstructRequiresExplicitEvidence, policy.WorkingHypothesisRequiresHedging, policy.ReflectiveConstructCannotScore}
	if !slices.Equal(c.HardGuardrails, wantGuardrails) {
		return fmt.Errorf("construct ontology hard guardrails = %v; want %v", c.HardGuardrails, wantGuardrails)
	}
	return nil
}

func (d Definition) Validate() error {
	if d.Definition == "" || len(d.Evidence) == 0 || len(d.NonEvidence) == 0 || d.AllowedClaim == "" || len(d.ProhibitedClaims) == 0 {
		return fmt.Errorf("construct %s lacks definition, evidence or claim boundaries", d.ID)
	}
	if !uniqueNonEmpty(d.Evidence) || !uniqueNonEmpty(d.NonEvidence) || !uniqueNonEmpty(d.ProhibitedClaims) || !uniqueValues(d.DimensionLinks) {
		return fmt.Errorf("construct %s contains duplicate or empty contract values", d.ID)
	}
	allowed := map[string][]string{
		"LANGUAGE_FEATURE":     {"OBSERVABLE", "DIRECT_OBSERVATION"},
		"CONTEXTUAL_CONSTRUCT": {"INFERABLE", "QUALIFIED_INFERENCE"},
		"WORKING_HYPOTHESIS":   {"WORKING_HYPOTHESIS", "HYPOTHESIS_ONLY"},
		"REFLECTIVE":           {"REFLECTIVE_ONLY", "REFLECTION_ONLY"},
	}
	want, ok := allowed[d.Layer]
	if !ok || d.InferenceClass != want[0] || d.ClaimMode != want[1] {
		return fmt.Errorf("construct %s has invalid layer/inference/claim mode", d.ID)
	}
	if d.Assessability.MinimumDistinctEvidence < 1 || d.CoreScoring {
		return fmt.Errorf("construct %s must require explicit evidence and cannot score directly", d.ID)
	}
	if d.Layer == "REFLECTIVE" && !d.Assessability.QuestionContextRequired {
		return fmt.Errorf("reflective construct %s requires question context", d.ID)
	}
	for _, id := range d.DimensionLinks {
		if !slices.Contains(dimension.All(), id) {
			return fmt.Errorf("construct %s links unknown dimension %s", d.ID, id)
		}
	}
	for _, signal := range d.RuntimeSignals {
		if err := signal.Validate(); err != nil {
			return fmt.Errorf("construct %s runtime signal: %w", d.ID, err)
		}
	}
	return nil
}

func (s RuntimeSignal) empty() bool {
	return len(s.PhrasesAll)+len(s.PhrasesAny)+len(s.Actors)+len(s.Modalities)+len(s.TargetTypes)+len(s.ExpectationSources)+len(s.PropositionFeatures)+len(s.SelectedSenses) == 0
}

func (s RuntimeSignal) Validate() error {
	if s.empty() || !uniqueNonEmpty(s.PhrasesAll) || !uniqueNonEmpty(s.PhrasesAny) || !uniqueValues(s.Actors) ||
		!uniqueValues(s.Modalities) || !uniqueValues(s.TargetTypes) || !uniqueValues(s.ExpectationSources) ||
		!uniqueNonEmpty(s.PropositionFeatures) || !uniqueNonEmpty(s.SelectedSenses) {
		return fmt.Errorf("runtime signal contains empty or duplicate values")
	}
	for _, value := range s.Actors {
		if !slices.Contains(policy.Actors(), value) {
			return fmt.Errorf("runtime signal references unknown actor %s", value)
		}
	}
	for _, value := range s.Modalities {
		if !slices.Contains(policy.Modalities(), value) {
			return fmt.Errorf("runtime signal references unknown modality %s", value)
		}
	}
	for _, value := range s.TargetTypes {
		if !slices.Contains(policy.TargetTypes(), value) {
			return fmt.Errorf("runtime signal references unknown target %s", value)
		}
	}
	for _, value := range s.ExpectationSources {
		if !slices.Contains(policy.ExpectationSources(), value) {
			return fmt.Errorf("runtime signal references unknown expectation source %s", value)
		}
	}
	for _, value := range s.PropositionFeatures {
		if !slices.Contains([]string{"PREDICATE", "TARGET", "TIME", "BOUNDARY", "DECISION", "NEGATION", "MODALITY_NONE", "MODALITY_NECESSITY", "MODALITY_POSSIBILITY", "MODALITY_PERMISSION", "MODALITY_EXPECTATION", "MODALITY_INTENTION", "MODALITY_PROBABILITY"}, value) {
			return fmt.Errorf("runtime signal references unknown proposition feature %s", value)
		}
	}
	return nil
}

func (c Composition) Validate(known map[policy.ConstructID]bool) error {
	if !patternID.MatchString(c.OutputPattern) || len(c.RequiredConstructs) < 2 || !uniqueValues(c.RequiredConstructs) || !uniqueValues(c.RelationsAny) || c.MaximumPropositionGap < 0 || c.MaximumPropositionGap > 8 {
		return fmt.Errorf("composition %q has invalid pattern, requirements or gap", c.OutputPattern)
	}
	for _, id := range c.RequiredConstructs {
		if !known[id] {
			return fmt.Errorf("composition %s references unknown construct %s", c.OutputPattern, id)
		}
	}
	for _, relation := range c.RelationsAny {
		if !slices.Contains(policy.DiscourseRelations(), relation) {
			return fmt.Errorf("composition %s references unknown relation %s", c.OutputPattern, relation)
		}
	}
	return nil
}

func uniqueNonEmpty(values []string) bool {
	for _, value := range values {
		if value == "" {
			return false
		}
	}
	return uniqueValues(values)
}

func uniqueValues[T comparable](values []T) bool {
	seen := make(map[T]bool, len(values))
	for _, value := range values {
		if seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}
