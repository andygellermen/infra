package db

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db/migrations"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/seed"
)

func TestPostgresFoundation(t *testing.T) {
	databaseURL := os.Getenv("SAL_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SAL_TEST_DATABASE_URL is not configured")
	}

	ctx := context.Background()
	database, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	migrator := NewMigrator(database, migrations.Files, ".")
	first, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("first migration run: %v", err)
	}
	if first.Total != 1 {
		t.Fatalf("migration total = %d; want 1", first.Total)
	}
	second, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("second migration run: %v", err)
	}
	if len(second.Applied) != 0 {
		t.Fatalf("second migration run applied %v; want none", second.Applied)
	}

	foundation, err := os.Open("../../data/seed/sprach-a-lyzer_foundation_v0.1.json")
	if err != nil {
		t.Fatalf("open foundation: %v", err)
	}
	defer foundation.Close()
	golden, err := os.Open("../../data/golden/sprach-a-lyzer_vertical-slice_v0.1.json")
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

	assertCount(t, database, "dimensions", 6)
	assertCount(t, database, "golden_test_cases", 6)
	assertCount(t, database, "presentation_bundles", 2)

	var rawAnalysisTableAbsent bool
	if err := database.QueryRow(`SELECT to_regclass('public.analyses') IS NULL`).Scan(&rawAnalysisTableAbsent); err != nil {
		t.Fatalf("check analysis table: %v", err)
	}
	if !rawAnalysisTableAbsent {
		t.Fatal("privacy-default foundation unexpectedly contains analyses table")
	}
	if err := NewSchemaPinger(database, RequiredSchemaVersion).PingContext(ctx); err != nil {
		t.Fatalf("schema readiness: %v", err)
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
