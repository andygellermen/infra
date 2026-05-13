package event

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func setupEventRepository(t *testing.T) (*Repository, string, *sql.DB) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "event-series-test.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	migrator := db.NewMigrator(sqlDB, migrations.Files, ".")
	if _, err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	tenantRepo := tenant.NewRepository(sqlDB)
	createdTenant, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "demo",
		Name:          "Demo Tenant",
		PublicBaseURL: "https://events.example.com/demo",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	return NewRepository(sqlDB), createdTenant.ID, sqlDB
}

func TestEventSeriesCRUD(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	created, err := repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:                "first-series",
		Title:               "Erste Serie",
		Description:         "Beschreibung",
		DefaultLocationName: "Seminarhaus",
		DefaultAddress:      "Musterstrasse 1",
		DefaultOnlineURL:    "https://meet.example.com/raum-1",
	})
	if err != nil {
		t.Fatalf("CreateSeries returned error: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected created series id")
	}
	if created.Slug != "first-series" {
		t.Fatalf("expected slug first-series, got %q", created.Slug)
	}
	if !created.IsPublic {
		t.Fatalf("expected created series to be public by default")
	}

	listed, err := repo.ListSeries(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListSeries returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected exactly one series, got %d", len(listed))
	}

	loaded, err := repo.GetSeriesByID(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("GetSeriesByID returned error: %v", err)
	}
	if loaded.Title != "Erste Serie" {
		t.Fatalf("expected title Erste Serie, got %q", loaded.Title)
	}

	newTitle := "Aktualisierte Serie"
	isPublic := false
	updated, err := repo.UpdateSeries(context.Background(), tenantID, created.ID, UpdateSeriesParams{
		Title:    &newTitle,
		IsPublic: &isPublic,
	})
	if err != nil {
		t.Fatalf("UpdateSeries returned error: %v", err)
	}
	if updated.Title != "Aktualisierte Serie" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}
	if updated.IsPublic {
		t.Fatalf("expected updated is_public false")
	}

	deleted, err := repo.DeleteSeries(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("DeleteSeries returned error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted=true")
	}

	_, err = repo.GetSeriesByID(context.Background(), tenantID, created.ID)
	if !errors.Is(err, ErrSeriesNotFound) {
		t.Fatalf("expected ErrSeriesNotFound, got %v", err)
	}
}

func TestCreateSeriesDuplicateSlug(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	_, err := repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:  "duplicate",
		Title: "Serie A",
	})
	if err != nil {
		t.Fatalf("CreateSeries returned error: %v", err)
	}

	_, err = repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:  "duplicate",
		Title: "Serie B",
	})
	if !errors.Is(err, ErrSeriesSlugExists) {
		t.Fatalf("expected ErrSeriesSlugExists, got %v", err)
	}
}

func TestListSeriesTenantIsolation(t *testing.T) {
	repo, tenantID, sqlDB := setupEventRepository(t)

	tenantRepo := tenant.NewRepository(sqlDB)
	secondTenant, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "second",
		Name:          "Second Tenant",
		PublicBaseURL: "https://events.example.com/second",
	})
	if err != nil {
		t.Fatalf("create second tenant: %v", err)
	}

	_, err = repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:  "series-a",
		Title: "Series A",
	})
	if err != nil {
		t.Fatalf("create first tenant series: %v", err)
	}
	_, err = repo.CreateSeries(context.Background(), secondTenant.ID, CreateSeriesParams{
		Slug:  "series-b",
		Title: "Series B",
	})
	if err != nil {
		t.Fatalf("create second tenant series: %v", err)
	}

	firstList, err := repo.ListSeries(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("list first tenant series: %v", err)
	}
	if len(firstList) != 1 {
		t.Fatalf("expected first tenant list length 1, got %d", len(firstList))
	}
	if firstList[0].Slug != "series-a" {
		t.Fatalf("expected first tenant slug series-a, got %q", firstList[0].Slug)
	}

	secondList, err := repo.ListSeries(context.Background(), secondTenant.ID)
	if err != nil {
		t.Fatalf("list second tenant series: %v", err)
	}
	if len(secondList) != 1 {
		t.Fatalf("expected second tenant list length 1, got %d", len(secondList))
	}
	if secondList[0].Slug != "series-b" {
		t.Fatalf("expected second tenant slug series-b, got %q", secondList[0].Slug)
	}
}

func TestUpdateSeriesRejectsEmptyPatch(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	created, err := repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:  "no-patch",
		Title: "No Patch",
	})
	if err != nil {
		t.Fatalf("create series: %v", err)
	}

	_, err = repo.UpdateSeries(context.Background(), tenantID, created.ID, UpdateSeriesParams{})
	if err == nil {
		t.Fatal("expected error for empty patch")
	}
}

func TestCreateSeriesWithPrivateFlag(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	private := false
	created, err := repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:     "private-series",
		Title:    "Private",
		IsPublic: &private,
	})
	if err != nil {
		t.Fatalf("create series with private flag: %v", err)
	}
	if created.IsPublic {
		t.Fatalf("expected private series")
	}
	if time.Since(created.CreatedAt) > 10*time.Second {
		t.Fatalf("unexpected created_at timestamp %s", created.CreatedAt.Format(time.RFC3339))
	}
}
