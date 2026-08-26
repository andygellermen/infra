package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

const CatalogueVersion = "0.1"

var catalogueKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

type Catalogue struct {
	Schema          string                   `json:"$schema"`
	Version         string                   `json:"catalogue_version"`
	Status          string                   `json:"status"`
	Locale          string                   `json:"locale"`
	VersionContract CatalogueVersionContract `json:"version_contract"`
	SenseThresholds SenseThresholds          `json:"sense_thresholds"`
	Lexemes         []LexemeDefinition       `json:"lexemes"`
	Connectors      []ConnectorDefinition    `json:"connectors"`
	ScopeRules      []ScopeDefinition        `json:"scope_rules"`
	HardGuardrails  []policy.GuardrailID     `json:"hard_guardrails"`
}

type CatalogueVersionContract struct {
	ResolverResult string `json:"resolver_result"`
	PolicyRegistry string `json:"policy_registry"`
}

type SenseThresholds struct {
	High     SenseThreshold      `json:"high"`
	Medium   SenseThreshold      `json:"medium"`
	Fallback policy.SenseStateID `json:"fallback"`
}

type SenseThreshold struct {
	MinimumConfidence float64 `json:"minimum_confidence"`
	MinimumGap        float64 `json:"minimum_gap"`
}

type LexemeDefinition struct {
	Key    string            `json:"key"`
	Forms  []string          `json:"forms"`
	Senses []SenseDefinition `json:"senses"`
	Status string            `json:"status"`
}

type SenseDefinition struct {
	ID             string                     `json:"id"`
	PhraseSignals  []string                   `json:"phrase_signals"`
	TokenSignals   []string                   `json:"token_signals"`
	ContextSignals []policy.AnalysisContextID `json:"context_signals"`
}

type ConnectorDefinition struct {
	Key        string                     `json:"key"`
	Markers    []string                   `json:"markers"`
	Relation   policy.DiscourseRelationID `json:"relation"`
	Placement  string                     `json:"placement"`
	Confidence float64                    `json:"confidence"`
	Status     string                     `json:"status"`
}

type ScopeDefinition struct {
	Key        string                 `json:"key"`
	Kind       string                 `json:"kind"`
	Priority   int                    `json:"priority"`
	Cues       []string               `json:"cues"`
	Output     policy.NegationScopeID `json:"output"`
	Confidence float64                `json:"confidence"`
	Status     string                 `json:"status"`
}

// CatalogueProvider supplies the approved resolver catalogue for one
// resolution. Providers must fail closed when no valid catalogue is active.
type CatalogueProvider interface {
	Active(context.Context) (Catalogue, error)
}

type StaticCatalogueProvider struct {
	Catalogue Catalogue
}

func (s StaticCatalogueProvider) Active(context.Context) (Catalogue, error) {
	return s.Catalogue, nil
}

func DecodeCatalogue(reader io.Reader) (Catalogue, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var catalogue Catalogue
	if err := decoder.Decode(&catalogue); err != nil {
		return Catalogue{}, fmt.Errorf("decode resolver catalogue: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Catalogue{}, fmt.Errorf("decode resolver catalogue: trailing JSON value")
	}
	if err := catalogue.Validate(); err != nil {
		return Catalogue{}, err
	}
	return catalogue, nil
}

func (c Catalogue) Validate() error {
	if c.Version != CatalogueVersion || c.Status != "APPROVED" || c.Locale != "de-DE" {
		return fmt.Errorf("resolver catalogue requires version %s, APPROVED status and de-DE locale", CatalogueVersion)
	}
	if c.VersionContract.ResolverResult != ContractVersion || c.VersionContract.PolicyRegistry != policy.RegistryVersion {
		return fmt.Errorf("resolver catalogue version contract must reference resolver %s and policy %s", ContractVersion, policy.RegistryVersion)
	}
	if !validThreshold(c.SenseThresholds.High) || !validThreshold(c.SenseThresholds.Medium) ||
		c.SenseThresholds.High.MinimumConfidence <= c.SenseThresholds.Medium.MinimumConfidence ||
		c.SenseThresholds.High.MinimumGap <= c.SenseThresholds.Medium.MinimumGap ||
		c.SenseThresholds.Fallback != policy.SenseAmbiguous {
		return fmt.Errorf("resolver catalogue contains invalid sense thresholds")
	}
	if len(c.Lexemes) < 8 || len(c.Connectors) < 8 || len(c.ScopeRules) < 3 {
		return fmt.Errorf("resolver catalogue requires at least 8 lexemes, 8 connectors and 3 scope rules")
	}
	seen := make(map[string]bool)
	for _, lexeme := range c.Lexemes {
		if err := validateLexeme(lexeme, seen); err != nil {
			return err
		}
	}
	for _, connector := range c.Connectors {
		if seen[connector.Key] || !catalogueKeyPattern.MatchString(connector.Key) || len(connector.Markers) == 0 ||
			!uniqueNonEmpty(connector.Markers) || !slices.Contains(policy.DiscourseRelations(), connector.Relation) ||
			!slices.Contains([]string{"PREFIX", "INFIX"}, connector.Placement) || !bounded(connector.Confidence) || connector.Status != "APPROVED" {
			return fmt.Errorf("invalid resolver connector %q", connector.Key)
		}
		seen[connector.Key] = true
	}
	priorities := make(map[int]bool, len(c.ScopeRules))
	for _, scope := range c.ScopeRules {
		if seen[scope.Key] || !catalogueKeyPattern.MatchString(scope.Key) || scope.Kind != "NEGATION" || scope.Priority < 0 ||
			priorities[scope.Priority] || len(scope.Cues) == 0 || !uniqueNonEmpty(scope.Cues) || !slices.Contains(policy.NegationScopes(), scope.Output) ||
			!bounded(scope.Confidence) || scope.Status != "APPROVED" {
			return fmt.Errorf("invalid resolver scope rule %q", scope.Key)
		}
		seen[scope.Key] = true
		priorities[scope.Priority] = true
	}
	wantGuardrails := []policy.GuardrailID{
		policy.AmbiguousFeatureCannotHardScore,
		policy.PropositionSpanMustMatchSource,
		policy.ResolverCandidateCannotBypassRules,
	}
	if !slices.Equal(c.HardGuardrails, wantGuardrails) {
		return fmt.Errorf("resolver catalogue hard guardrails = %v; want %v", c.HardGuardrails, wantGuardrails)
	}
	return nil
}

func validateLexeme(lexeme LexemeDefinition, seen map[string]bool) error {
	if seen[lexeme.Key] || !catalogueKeyPattern.MatchString(lexeme.Key) || len(lexeme.Forms) == 0 ||
		!uniqueNonEmpty(lexeme.Forms) || len(lexeme.Senses) < 2 || lexeme.Status != "APPROVED" {
		return fmt.Errorf("invalid resolver lexeme %q", lexeme.Key)
	}
	seen[lexeme.Key] = true
	senses := make(map[string]bool, len(lexeme.Senses))
	for _, sense := range lexeme.Senses {
		if senses[sense.ID] || !catalogueKeyPattern.MatchString(sense.ID) || !uniqueNonEmpty(sense.PhraseSignals) ||
			!uniqueNonEmpty(sense.TokenSignals) || !uniqueContexts(sense.ContextSignals) {
			return fmt.Errorf("invalid resolver sense %q/%q", lexeme.Key, sense.ID)
		}
		senses[sense.ID] = true
	}
	return nil
}

func validThreshold(value SenseThreshold) bool {
	return bounded(value.MinimumConfidence) && bounded(value.MinimumGap)
}

func bounded(value float64) bool { return value >= 0 && value <= 1 }

func uniqueNonEmpty(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func uniqueContexts(values []policy.AnalysisContextID) bool {
	seen := make(map[policy.AnalysisContextID]bool, len(values))
	for _, value := range values {
		if seen[value] || !slices.Contains(policy.AnalysisContexts(), value) {
			return false
		}
		seen[value] = true
	}
	return true
}
