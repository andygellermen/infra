package db_test

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/app"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db/migrations"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/seed"
)

func TestPostgresFoundation(t *testing.T) {
	databaseURL := os.Getenv("SAL_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SAL_TEST_DATABASE_URL is not configured")
	}

	ctx := context.Background()
	database, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	migrator := db.NewMigrator(database, migrations.Files, ".")
	first, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("first migration run: %v", err)
	}
	if first.Total != 3 {
		t.Fatalf("migration total = %d; want 3", first.Total)
	}
	second, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("second migration run: %v", err)
	}
	if len(second.Applied) != 0 {
		t.Fatalf("second migration run applied %v; want none", second.Applied)
	}

	foundation, err := os.Open("../../data/seed/sprach-a-lyzer_foundation_v0.3.json")
	if err != nil {
		t.Fatalf("open foundation: %v", err)
	}
	defer foundation.Close()
	golden, err := os.Open("../../data/golden/sprach-a-lyzer_vertical-slice_v0.2.json")
	if err != nil {
		t.Fatalf("open golden suite: %v", err)
	}
	defer golden.Close()
	result, err := seed.Apply(ctx, database, foundation, golden)
	if err != nil {
		t.Fatalf("seed foundation: %v", err)
	}
	if result.Dimensions != 6 || result.GoldenCases != 6 || result.PresentationBundles != 2 {
		t.Fatalf("unexpected seed result: %+v", result)
	}
	legacyFoundation, err := os.ReadFile("../../data/seed/sprach-a-lyzer_foundation_v0.3.json")
	if err != nil {
		t.Fatalf("read legacy foundation source: %v", err)
	}
	legacyGolden, err := os.ReadFile("../../data/golden/sprach-a-lyzer_vertical-slice_v0.2.json")
	if err != nil {
		t.Fatalf("read legacy golden source: %v", err)
	}
	legacyFoundation = []byte(strings.ReplaceAll(string(legacyFoundation), "VOLITION", "FREE_WILL"))
	legacyGolden = []byte(strings.ReplaceAll(string(legacyGolden), "VOLITION", "FREE_WILL"))
	legacyResult, err := seed.Apply(ctx, database, bytes.NewReader(legacyFoundation), bytes.NewReader(legacyGolden))
	if err != nil {
		t.Fatalf("seed legacy foundation: %v", err)
	}
	if legacyResult.LegacyMappings == 0 {
		t.Fatal("legacy seed applied without compatibility mappings")
	}

	assertCount(t, database, "dimensions", 6)
	assertCount(t, database, "golden_test_cases", 6)
	assertCount(t, database, "presentation_bundles", 2)
	assertScalar(t, database, `SELECT COUNT(*) FROM golden_test_cases WHERE suite_version = '0.2'`, 6)
	assertScalar(t, database, `SELECT COUNT(*) FROM golden_test_cases WHERE expected_payload ? 'contribution_trace'`, 6)
	assertScalar(t, database, `SELECT COUNT(*) FROM dimensions WHERE dimension_id = 'VOLITION'`, 1)
	assertScalar(t, database, `SELECT COUNT(*) FROM dimensions WHERE dimension_id = 'FREE_WILL'`, 0)
	assertScalar(t, database, `SELECT COUNT(*) FROM rules WHERE actions::text LIKE '%FREE_WILL%'`, 0)
	assertScalar(t, database, `SELECT COUNT(*) FROM rules WHERE contract_version = '0.4'`, 9)
	assertScalar(t, database, `SELECT COUNT(*) FROM rules WHERE jsonb_array_length(source_keys) > 0`, 9)
	assertScalar(t, database, `SELECT COUNT(*) FROM presentation_entries WHERE canonical_key = 'FREE_WILL'`, 0)
	assertScalar(t, database, `SELECT COUNT(*) FROM audit_events WHERE event_type = 'LEGACY_DIMENSION_MAPPED'`, 1)

	var rawAnalysisTableAbsent bool
	if err := database.QueryRow(`SELECT to_regclass('public.analyses') IS NULL`).Scan(&rawAnalysisTableAbsent); err != nil {
		t.Fatalf("check analysis table: %v", err)
	}
	if !rawAnalysisTableAbsent {
		t.Fatal("privacy-default foundation unexpectedly contains analyses table")
	}
	application := app.New(database)
	if err := application.Readiness.PingContext(ctx); err != nil {
		t.Fatalf("schema readiness: %v", err)
	}
	knowledgeSnapshot, err := application.Knowledge.Snapshot(ctx)
	if err != nil || knowledgeSnapshot.Dimensions != 6 {
		t.Fatalf("knowledge module snapshot = %+v, %v", knowledgeSnapshot, err)
	}
	ruleCatalogue, err := application.Rules.Active(ctx)
	if err != nil || ruleCatalogue.Version != "0.3" || len(ruleCatalogue.Rules) != 9 {
		t.Fatalf("rules module catalogue = %+v, %v", ruleCatalogue, err)
	}
	runtimeResult, err := application.Analysis.Analyze(analysis.Request{
		Text: "Ich muss das heute unbedingt noch schaffen.", Context: analysis.ContextSelfTalk,
	})
	if err != nil || !contains(runtimeResult.Patterns, "INTERNAL_PRESSURE") {
		t.Fatalf("database-backed runtime catalogue result = %+v, %v", runtimeResult, err)
	}
	if _, err := database.Exec(`UPDATE rules SET enabled = false WHERE rule_key = 'R-INTERNAL-PRESSURE' AND version = 3`); err != nil {
		t.Fatalf("disable runtime rule: %v", err)
	}
	t.Cleanup(func() {
		_, _ = database.Exec(`UPDATE rules SET enabled = true WHERE rule_key = 'R-INTERNAL-PRESSURE' AND version = 3`)
	})
	disabledResult, err := application.Analysis.Analyze(analysis.Request{
		Text: "Ich muss das heute unbedingt noch schaffen.", Context: analysis.ContextSelfTalk,
	})
	if err != nil || contains(disabledResult.Patterns, "INTERNAL_PRESSURE") {
		t.Fatalf("disabled database rule still active: %+v, %v", disabledResult, err)
	}
	corporateBundle, err := application.Presentation.Bundle(ctx, "CORPORATE", "de-DE")
	if err != nil {
		t.Fatalf("presentation module bundle: %v", err)
	}
	if got := corporateBundle.Resolve("METRIC_WING_SCORE"); got != "Wirkungsprofil" {
		t.Fatalf("corporate metric label = %q; want Wirkungsprofil", got)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func assertScalar(t *testing.T, database queryRower, query string, want int) {
	t.Helper()
	var got int
	if err := database.QueryRow(query).Scan(&got); err != nil {
		t.Fatalf("query scalar: %v", err)
	}
	if got != want {
		t.Fatalf("query scalar = %d; want %d (%s)", got, want, query)
	}
}

func assertCount(t *testing.T, database queryRower, table string, want int) {
	t.Helper()
	var got int
	if err := database.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s count = %d; want %d", table, got, want)
	}
}

type queryRower interface {
	QueryRow(query string, args ...any) *sql.Row
}
