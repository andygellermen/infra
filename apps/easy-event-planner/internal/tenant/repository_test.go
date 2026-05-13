package tenant

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
)

func newMigratedDB(t *testing.T) *sql.DB {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "tenant-test.sqlite"))
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

	return sqlDB
}

func TestCreateTenantAndLookupBySlug(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	created, err := repo.CreateTenant(context.Background(), CreateTenantParams{
		Slug:          "Customer-XYZ",
		Name:          "Customer XYZ",
		PublicBaseURL: "https://events.example.com/customer-xyz",
	})
	if err != nil {
		t.Fatalf("CreateTenant returned error: %v", err)
	}

	if created.Slug != "customer-xyz" {
		t.Fatalf("expected normalized slug customer-xyz, got %q", created.Slug)
	}
	if created.DefaultTimezone != DefaultTimezone {
		t.Fatalf("expected default timezone %q, got %q", DefaultTimezone, created.DefaultTimezone)
	}

	lookup, err := repo.LookupBySlug(context.Background(), "CUSTOMER-XYZ")
	if err != nil {
		t.Fatalf("LookupBySlug returned error: %v", err)
	}
	if lookup.ID != created.ID {
		t.Fatalf("expected same tenant id, got %q and %q", created.ID, lookup.ID)
	}

	settings, err := repo.GetSettings(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if settings.PayPalMode != DefaultPayPalMode {
		t.Fatalf("expected default paypal mode %q, got %q", DefaultPayPalMode, settings.PayPalMode)
	}
	if settings.DefaultRetentionDays != DefaultRetentionDays {
		t.Fatalf("expected default retention %d, got %d", DefaultRetentionDays, settings.DefaultRetentionDays)
	}
}

func TestUpsertSettings(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	created, err := repo.CreateTenant(context.Background(), CreateTenantParams{
		Slug:          "demo",
		Name:          "Demo Tenant",
		PublicBaseURL: "https://events.example.com/demo",
	})
	if err != nil {
		t.Fatalf("CreateTenant returned error: %v", err)
	}

	updated, err := repo.UpsertSettings(context.Background(), UpsertTenantSettingsParams{
		TenantID: created.ID,
		Settings: TenantSettingsInput{
			SenderEmail:          "noreply@events.example.com",
			SenderName:           "Demo Events",
			PayPalMode:           "sandbox",
			DefaultRetentionDays: 45,
			SettingsJSON:         `{"theme":"amber"}`,
		},
	})
	if err != nil {
		t.Fatalf("UpsertSettings returned error: %v", err)
	}

	if updated.SenderEmail != "noreply@events.example.com" {
		t.Fatalf("unexpected sender email %q", updated.SenderEmail)
	}
	if updated.PayPalMode != "sandbox" {
		t.Fatalf("unexpected paypal mode %q", updated.PayPalMode)
	}
	if updated.DefaultRetentionDays != 45 {
		t.Fatalf("unexpected retention days %d", updated.DefaultRetentionDays)
	}
}

func TestSeedTenantIsIdempotent(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	first, err := repo.SeedTenant(context.Background(), SeedInput{
		Slug:          "community",
		Name:          "Community Tenant",
		PublicBaseURL: "https://events.example.com/community",
		Settings: TenantSettingsInput{
			SenderEmail:          "first@example.com",
			DefaultRetentionDays: 30,
		},
	})
	if err != nil {
		t.Fatalf("first SeedTenant returned error: %v", err)
	}
	if !first.Created {
		t.Fatalf("expected first seed run to create tenant")
	}

	second, err := repo.SeedTenant(context.Background(), SeedInput{
		Slug:          "community",
		Name:          "Community Tenant",
		PublicBaseURL: "https://events.example.com/community",
		Settings: TenantSettingsInput{
			SenderEmail:          "second@example.com",
			DefaultRetentionDays: 60,
		},
	})
	if err != nil {
		t.Fatalf("second SeedTenant returned error: %v", err)
	}
	if second.Created {
		t.Fatalf("expected second seed run to update existing tenant")
	}
	if second.Tenant.ID != first.Tenant.ID {
		t.Fatalf("expected same tenant id for repeated seed")
	}
	if second.Settings.SenderEmail != "second@example.com" {
		t.Fatalf("expected updated sender email, got %q", second.Settings.SenderEmail)
	}
	if second.Settings.DefaultRetentionDays != 60 {
		t.Fatalf("expected updated retention days, got %d", second.Settings.DefaultRetentionDays)
	}
}

func TestLookupBySlugNotFound(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	_, err := repo.LookupBySlug(context.Background(), "missing")
	if !errors.Is(err, ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}
}
