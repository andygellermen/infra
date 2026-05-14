package snippet

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func setupSnippetRepository(t *testing.T) (*Repository, string, *sql.DB) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "snippet-test.sqlite"))
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

func TestSnippetCRUDAndEmbedData(t *testing.T) {
	repo, tenantID, _ := setupSnippetRepository(t)

	created, err := repo.CreateConfig(context.Background(), tenantID, CreateConfigParams{
		Name:               "Footer Upcoming",
		Slug:               "footer-upcoming",
		ViewType:           "cards",
		EventFilterJSON:    `{"events":"upcoming","limit":6}`,
		DisplayOptionsJSON: `{"theme":"light","register":true}`,
	})
	if err != nil {
		t.Fatalf("CreateConfig returned error: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected snippet id")
	}
	if created.Slug != "footer-upcoming" {
		t.Fatalf("expected slug footer-upcoming, got %q", created.Slug)
	}
	if !created.IsActive {
		t.Fatalf("expected new snippet to be active by default")
	}

	list, err := repo.ListConfigs(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListConfigs returned error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one snippet config, got %d", len(list))
	}

	byID, err := repo.GetConfigByID(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("GetConfigByID returned error: %v", err)
	}
	if byID.Name != "Footer Upcoming" {
		t.Fatalf("expected name Footer Upcoming, got %q", byID.Name)
	}

	bySlug, err := repo.GetConfigBySlug(context.Background(), tenantID, "FOOTER-UPCOMING")
	if err != nil {
		t.Fatalf("GetConfigBySlug returned error: %v", err)
	}
	if bySlug.ID != created.ID {
		t.Fatalf("expected same snippet id, got %q and %q", bySlug.ID, created.ID)
	}

	newView := "list"
	newActive := false
	updated, err := repo.UpdateConfig(context.Background(), tenantID, created.ID, UpdateConfigParams{
		ViewType: &newView,
		IsActive: &newActive,
	})
	if err != nil {
		t.Fatalf("UpdateConfig returned error: %v", err)
	}
	if updated.ViewType != "list" {
		t.Fatalf("expected updated view_type list, got %q", updated.ViewType)
	}
	if updated.IsActive {
		t.Fatalf("expected updated snippet to be inactive")
	}

	deleted, err := repo.DeleteConfig(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("DeleteConfig returned error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted=true")
	}

	_, err = repo.GetConfigByID(context.Background(), tenantID, created.ID)
	if !errors.Is(err, ErrSnippetNotFound) {
		t.Fatalf("expected ErrSnippetNotFound, got %v", err)
	}
}

func TestCreateConfigDuplicateSlug(t *testing.T) {
	repo, tenantID, _ := setupSnippetRepository(t)

	_, err := repo.CreateConfig(context.Background(), tenantID, CreateConfigParams{
		Name:     "Snippet A",
		Slug:     "same-snippet",
		ViewType: "cards",
	})
	if err != nil {
		t.Fatalf("CreateConfig returned error: %v", err)
	}

	_, err = repo.CreateConfig(context.Background(), tenantID, CreateConfigParams{
		Name:     "Snippet B",
		Slug:     "same-snippet",
		ViewType: "list",
	})
	if !errors.Is(err, ErrSnippetSlugExists) {
		t.Fatalf("expected ErrSnippetSlugExists, got %v", err)
	}
}

func TestCreateConfigRejectsInvalidJSON(t *testing.T) {
	repo, tenantID, _ := setupSnippetRepository(t)

	_, err := repo.CreateConfig(context.Background(), tenantID, CreateConfigParams{
		Name:            "Invalid JSON",
		Slug:            "invalid-json",
		ViewType:        "cards",
		EventFilterJSON: `["wrong"]`,
	})
	if err == nil {
		t.Fatal("expected validation error for invalid event_filter_json object")
	}
}
