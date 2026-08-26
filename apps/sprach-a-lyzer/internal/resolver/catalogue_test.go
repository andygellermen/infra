package resolver

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

const catalogueFixturePath = "../../data/seed/sprach-a-lyzer_resolver-catalogue_v0.1.json"

func TestDecodeCanonicalResolverCatalogue(t *testing.T) {
	t.Parallel()
	file, err := os.Open(catalogueFixturePath)
	if err != nil {
		t.Fatalf("open resolver catalogue: %v", err)
	}
	defer file.Close()
	catalogue, err := DecodeCatalogue(file)
	if err != nil {
		t.Fatalf("DecodeCatalogue() error: %v", err)
	}
	if catalogue.Version != CatalogueVersion || len(catalogue.Lexemes) != 8 || len(catalogue.Connectors) != 8 || len(catalogue.ScopeRules) != 3 {
		t.Fatalf("unexpected resolver catalogue: version=%s lexemes=%d connectors=%d scopes=%d", catalogue.Version, len(catalogue.Lexemes), len(catalogue.Connectors), len(catalogue.ScopeRules))
	}
	if catalogue.SenseThresholds.High.MinimumConfidence != 0.75 || catalogue.SenseThresholds.High.MinimumGap != 0.20 ||
		catalogue.SenseThresholds.Medium.MinimumConfidence != 0.60 || catalogue.SenseThresholds.Medium.MinimumGap != 0.10 ||
		catalogue.SenseThresholds.Fallback != policy.SenseAmbiguous {
		t.Fatalf("canonical sense thresholds drifted: %+v", catalogue.SenseThresholds)
	}
	for _, guardrail := range catalogue.HardGuardrails {
		if !slices.Contains(policy.HardGuardrails(), guardrail) {
			t.Errorf("resolver catalogue uses unregistered guardrail %q", guardrail)
		}
	}
}

func TestDecodeResolverCatalogueRejectsUnknownField(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(catalogueFixturePath)
	if err != nil {
		t.Fatalf("read resolver catalogue: %v", err)
	}
	invalid := strings.Replace(string(data), `"locale": "de-DE"`, `"locale": "de-DE", "calibration_magic": true`, 1)
	_, err = DecodeCatalogue(strings.NewReader(invalid))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("DecodeCatalogue() error = %v; want unknown field", err)
	}
}

func TestResolverCatalogueRequiresLockedGuardrails(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(catalogueFixturePath)
	if err != nil {
		t.Fatalf("read resolver catalogue: %v", err)
	}
	invalid := strings.Replace(string(data), `    "RESOLVER_CANDIDATE_CANNOT_BYPASS_RULES"`, `    "NO_TRAIT_CLAIMS"`, 1)
	_, err = DecodeCatalogue(strings.NewReader(invalid))
	if err == nil || !strings.Contains(err.Error(), "hard guardrails") {
		t.Fatalf("DecodeCatalogue() error = %v; want locked guardrail rejection", err)
	}
}
