package seed

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/dimension"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/rules"
)

type Foundation struct {
	Version             string               `json:"version"`
	Dimensions          []Dimension          `json:"dimensions"`
	RuleSet             VersionedSet         `json:"rule_set"`
	Rules               []rules.Definition   `json:"rules"`
	ParameterSet        VersionedSet         `json:"parameter_set"`
	Parameters          []Parameter          `json:"parameters"`
	PresentationBundles []PresentationBundle `json:"presentation_bundles"`
}

type Dimension struct {
	ID            dimension.ID `json:"id"`
	Slug          string       `json:"slug"`
	PositiveLabel string       `json:"positive_label"`
	NegativeLabel string       `json:"negative_label"`
	Description   string       `json:"description"`
}

type VersionedSet struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	Status    string `json:"status"`
	Changelog string `json:"changelog"`
}

type Parameter struct {
	Key         string          `json:"key"`
	Category    string          `json:"category"`
	Value       json.RawMessage `json:"value"`
	Min         json.RawMessage `json:"min"`
	Max         json.RawMessage `json:"max"`
	Description string          `json:"description"`
	Editable    bool            `json:"editable"`
}

type PresentationBundle struct {
	ID        string            `json:"id"`
	Profile   string            `json:"profile"`
	Locale    string            `json:"locale"`
	Version   string            `json:"version"`
	Status    string            `json:"status"`
	Fallbacks json.RawMessage   `json:"fallbacks"`
	Entries   map[string]string `json:"entries"`
}

type goldenSuite struct {
	Version               string           `json:"version"`
	AnalysisContract      string           `json:"analysis_contract"`
	TraceContract         string           `json:"trace_contract"`
	CanonicalDimensionIDs []dimension.ID   `json:"canonical_dimension_ids"`
	Cases                 []goldenTestCase `json:"cases"`
}

type goldenTestCase struct {
	ID             string          `json:"id"`
	Request        json.RawMessage `json:"request"`
	Expected       json.RawMessage `json:"expected"`
	ExpectedResult json.RawMessage `json:"expected_result"`
}

type Result struct {
	Dimensions          int `json:"dimensions"`
	Rules               int `json:"rules"`
	Parameters          int `json:"parameters"`
	PresentationBundles int `json:"presentation_bundles"`
	GoldenCases         int `json:"golden_cases"`
	LegacyMappings      int `json:"legacy_mappings"`
}

func DecodeFoundation(reader io.Reader) (Foundation, error) {
	foundation, _, err := DecodeFoundationWithReport(reader)
	return foundation, err
}

func DecodeFoundationWithReport(reader io.Reader) (Foundation, dimension.CompatibilityReport, error) {
	normalized, report, err := dimension.NormalizeReader(reader)
	if err != nil {
		return Foundation{}, dimension.CompatibilityReport{}, err
	}
	var foundation Foundation
	decoder := json.NewDecoder(bytes.NewReader(normalized))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&foundation); err != nil {
		return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("decode foundation seed: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("decode foundation seed: trailing JSON value")
	}
	if foundation.Version == "" || len(foundation.Dimensions) != 6 {
		return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("foundation seed must contain a version and six dimensions")
	}
	wantDimensions := map[dimension.ID]bool{
		dimension.Agency: true, dimension.Connection: true, dimension.Appreciation: true,
		dimension.Clarity: true, dimension.Volition: true, dimension.Openness: true,
	}
	for _, definition := range foundation.Dimensions {
		if !wantDimensions[definition.ID] {
			return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("foundation seed contains non-canonical dimension %q", definition.ID)
		}
		delete(wantDimensions, definition.ID)
	}
	if len(wantDimensions) != 0 {
		return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("foundation seed has duplicate or missing canonical dimensions")
	}
	if len(foundation.Rules) > 0 && len(foundation.Rules) != 6 {
		return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("foundation must contain exactly six rules")
	}
	seenRuleKeys := make(map[string]bool, len(foundation.Rules))
	for _, rule := range foundation.Rules {
		if err := rule.Validate(); err != nil {
			return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("invalid foundation rule: %w", err)
		}
		if seenRuleKeys[rule.Key] {
			return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("duplicate foundation rule key %q", rule.Key)
		}
		seenRuleKeys[rule.Key] = true
	}
	profiles := make(map[string]PresentationBundle)
	for _, bundle := range foundation.PresentationBundles {
		if bundle.Profile != "PRIVATE" && bundle.Profile != "CORPORATE" {
			return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("unsupported presentation profile %q", bundle.Profile)
		}
		if len(bundle.Fallbacks) == 0 || len(bundle.Entries) == 0 {
			return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("presentation profile %s must contain entries and fallbacks", bundle.Profile)
		}
		profiles[bundle.Profile] = bundle
	}
	if len(foundation.PresentationBundles) > 0 && (profiles["PRIVATE"].ID == "" || profiles["CORPORATE"].ID == "") {
		return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("foundation must contain isolated PRIVATE and CORPORATE presentation bundles")
	}
	if corporate, ok := profiles["CORPORATE"]; ok && corporate.Entries["METRIC_WING_SCORE"] == "WingScore" {
		return Foundation{}, dimension.CompatibilityReport{}, fmt.Errorf("corporate presentation bundle leaks private canonical label")
	}
	return foundation, report, nil
}

func Apply(ctx context.Context, database *sql.DB, foundationReader, goldenReader io.Reader) (Result, error) {
	if database == nil {
		return Result{}, fmt.Errorf("seed database is nil")
	}
	foundation, foundationReport, err := DecodeFoundationWithReport(foundationReader)
	if err != nil {
		return Result{}, err
	}
	normalizedGolden, goldenReport, err := dimension.NormalizeReader(goldenReader)
	if err != nil {
		return Result{}, err
	}
	var golden goldenSuite
	decoder := json.NewDecoder(bytes.NewReader(normalizedGolden))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&golden); err != nil {
		return Result{}, fmt.Errorf("decode golden seed: %w", err)
	}
	if golden.Version == "" || !slices.Equal(golden.CanonicalDimensionIDs, dimension.All()) {
		return Result{}, fmt.Errorf("golden seed must contain a version and the canonical dimension catalogue")
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("begin seed transaction: %w", err)
	}
	defer tx.Rollback()

	if err := upsertFoundation(ctx, tx, foundation); err != nil {
		return Result{}, err
	}
	for _, testCase := range golden.Cases {
		if len(testCase.Expected) > 0 && len(testCase.ExpectedResult) > 0 {
			return Result{}, fmt.Errorf("golden case %s contains two expected payloads", testCase.ID)
		}
		expectedPayload := testCase.ExpectedResult
		if len(expectedPayload) == 0 {
			expectedPayload = testCase.Expected
		}
		if testCase.ID == "" || len(testCase.Request) == 0 || len(expectedPayload) == 0 {
			return Result{}, fmt.Errorf("golden case must contain id, request and expected payload")
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO golden_test_cases(case_id, suite_version, input_payload, expected_payload)
VALUES($1, $2, $3::jsonb, $4::jsonb)
ON CONFLICT (case_id) DO UPDATE SET
  suite_version = EXCLUDED.suite_version,
  input_payload = EXCLUDED.input_payload,
  expected_payload = EXCLUDED.expected_payload,
  updated_at = now()`, testCase.ID, golden.Version, testCase.Request, expectedPayload); err != nil {
			return Result{}, fmt.Errorf("upsert golden case %s: %w", testCase.ID, err)
		}
	}
	legacyMappings := append(foundationReport.Mappings, goldenReport.Mappings...)
	if len(legacyMappings) > 0 {
		detail, err := json.Marshal(map[string]any{"mappings": legacyMappings})
		if err != nil {
			return Result{}, fmt.Errorf("encode dimension compatibility audit: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO audit_events(event_type, entity_type, entity_id, detail)
VALUES('LEGACY_DIMENSION_MAPPED', 'FOUNDATION_SEED', $1, $2::jsonb)`, foundation.Version, detail); err != nil {
			return Result{}, fmt.Errorf("record dimension compatibility audit: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit seed transaction: %w", err)
	}
	return Result{
		Dimensions: len(foundation.Dimensions), Rules: len(foundation.Rules),
		Parameters: len(foundation.Parameters), PresentationBundles: len(foundation.PresentationBundles),
		GoldenCases: len(golden.Cases), LegacyMappings: len(legacyMappings),
	}, nil
}

func upsertFoundation(ctx context.Context, tx *sql.Tx, foundation Foundation) error {
	for _, dimension := range foundation.Dimensions {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO dimensions(dimension_id, slug, positive_label, negative_label, short_description, version)
VALUES($1, $2, $3, $4, $5, 1)
ON CONFLICT (dimension_id) DO UPDATE SET
  slug = EXCLUDED.slug, positive_label = EXCLUDED.positive_label,
  negative_label = EXCLUDED.negative_label, short_description = EXCLUDED.short_description,
  updated_at = now()`, dimension.ID, dimension.Slug, dimension.PositiveLabel, dimension.NegativeLabel, dimension.Description); err != nil {
			return fmt.Errorf("upsert dimension %s: %w", dimension.ID, err)
		}
	}

	if err := upsertSet(ctx, tx, "rule_sets", foundation.RuleSet); err != nil {
		return err
	}
	for _, rule := range foundation.Rules {
		condition, err := json.Marshal(rule.Condition)
		if err != nil {
			return fmt.Errorf("encode rule %s condition: %w", rule.Key, err)
		}
		actions, err := json.Marshal(rule.Actions)
		if err != nil {
			return fmt.Errorf("encode rule %s actions: %w", rule.Key, err)
		}
		sourceKeys, err := json.Marshal(rule.SourceKeys)
		if err != nil {
			return fmt.Errorf("encode rule %s sources: %w", rule.Key, err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO rules(
  id, contract_version, rule_key, name, description, priority, enabled, scope,
  condition_tree, actions, confidence_modifier, stop_processing, status, version,
  evidence_class, source_keys)
VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb, $11, $12, $13, $14, $15, $16::jsonb)
ON CONFLICT (rule_key, version) DO UPDATE SET
  contract_version = EXCLUDED.contract_version, name = EXCLUDED.name,
  description = EXCLUDED.description, priority = EXCLUDED.priority,
  enabled = EXCLUDED.enabled, scope = EXCLUDED.scope,
  condition_tree = EXCLUDED.condition_tree, actions = EXCLUDED.actions,
  confidence_modifier = EXCLUDED.confidence_modifier,
  stop_processing = EXCLUDED.stop_processing, status = EXCLUDED.status,
  version = EXCLUDED.version, evidence_class = EXCLUDED.evidence_class,
  source_keys = EXCLUDED.source_keys, updated_at = now()`,
			rule.ID, rule.ContractVersion, rule.Key, rule.Name, rule.Description,
			rule.Priority, rule.Enabled, rule.Scope, condition, actions,
			rule.ConfidenceModifier, rule.StopProcessing, rule.Status, rule.Version,
			rule.EvidenceClass, sourceKeys); err != nil {
			return fmt.Errorf("upsert rule %s: %w", rule.Key, err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO rule_set_rules(rule_set_id, rule_id) VALUES($1, $2)
ON CONFLICT DO NOTHING`, foundation.RuleSet.ID, rule.ID); err != nil {
			return fmt.Errorf("link rule %s: %w", rule.Key, err)
		}
	}

	if err := upsertSet(ctx, tx, "parameter_sets", foundation.ParameterSet); err != nil {
		return err
	}
	for _, parameter := range foundation.Parameters {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO parameters(parameter_key, category, default_value, min_value, max_value, description, editable)
VALUES($1, $2, $3::jsonb, $4::jsonb, $5::jsonb, $6, $7)
ON CONFLICT (parameter_key) DO UPDATE SET
  category = EXCLUDED.category, default_value = EXCLUDED.default_value,
  min_value = EXCLUDED.min_value, max_value = EXCLUDED.max_value,
  description = EXCLUDED.description, editable = EXCLUDED.editable, updated_at = now()`,
			parameter.Key, parameter.Category, parameter.Value, nullableJSON(parameter.Min), nullableJSON(parameter.Max), parameter.Description, parameter.Editable); err != nil {
			return fmt.Errorf("upsert parameter %s: %w", parameter.Key, err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO parameter_set_values(parameter_set_id, parameter_key, value)
VALUES($1, $2, $3::jsonb)
ON CONFLICT (parameter_set_id, parameter_key) DO UPDATE SET value = EXCLUDED.value`,
			foundation.ParameterSet.ID, parameter.Key, parameter.Value); err != nil {
			return fmt.Errorf("link parameter %s: %w", parameter.Key, err)
		}
	}

	for _, bundle := range foundation.PresentationBundles {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO presentation_bundles(id, profile, locale, version, status, fallbacks)
VALUES($1, $2, $3, $4, $5, $6::jsonb)
ON CONFLICT (profile, locale, version) DO UPDATE SET
  status = EXCLUDED.status, fallbacks = EXCLUDED.fallbacks, updated_at = now()`,
			bundle.ID, bundle.Profile, bundle.Locale, bundle.Version, bundle.Status, bundle.Fallbacks); err != nil {
			return fmt.Errorf("upsert presentation bundle %s: %w", bundle.Profile, err)
		}
		for key, value := range bundle.Entries {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO presentation_entries(bundle_id, canonical_key, display_value)
VALUES($1, $2, $3)
ON CONFLICT (bundle_id, canonical_key) DO UPDATE SET display_value = EXCLUDED.display_value`,
				bundle.ID, key, value); err != nil {
				return fmt.Errorf("upsert presentation entry %s/%s: %w", bundle.Profile, key, err)
			}
		}
	}
	return nil
}

func upsertSet(ctx context.Context, tx *sql.Tx, table string, set VersionedSet) error {
	if set.Status == "PRODUCTION" {
		query := fmt.Sprintf(`
UPDATE %s SET status = 'ARCHIVED', updated_at = now()
WHERE status = 'PRODUCTION' AND version <> $1`, table)
		if _, err := tx.ExecContext(ctx, query, set.Version); err != nil {
			return fmt.Errorf("archive previous production %s: %w", table, err)
		}
	}
	query := fmt.Sprintf(`
INSERT INTO %s(id, version, status, changelog, published_at)
VALUES($1, $2, $3, $4, CASE WHEN $3 = 'PRODUCTION' THEN now() ELSE NULL END)
ON CONFLICT (version) DO UPDATE SET
  status = EXCLUDED.status, changelog = EXCLUDED.changelog,
  published_at = COALESCE(%s.published_at, EXCLUDED.published_at), updated_at = now()`, table, table)
	if _, err := tx.ExecContext(ctx, query, set.ID, set.Version, set.Status, set.Changelog); err != nil {
		return fmt.Errorf("upsert %s %s: %w", table, set.Version, err)
	}
	return nil
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return value
}
