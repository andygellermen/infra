package tenant

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
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

func TestLookupByPublicHost(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	created, err := repo.CreateTenant(context.Background(), CreateTenantParams{
		Slug:          "customer-xyz",
		Name:          "Customer XYZ",
		PublicBaseURL: "https://events.example.com/customer-xyz",
	})
	if err != nil {
		t.Fatalf("CreateTenant returned error: %v", err)
	}

	lookup, err := repo.LookupByPublicHost(context.Background(), "events.example.com:443")
	if err != nil {
		t.Fatalf("LookupByPublicHost returned error: %v", err)
	}
	if lookup.ID != created.ID {
		t.Fatalf("expected tenant id %q, got %q", created.ID, lookup.ID)
	}
}

func TestLookupByPublicHostFailsWhenAmbiguous(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	for _, item := range []CreateTenantParams{
		{
			Slug:          "customer-xyz",
			Name:          "Customer XYZ",
			PublicBaseURL: "https://events.example.com/customer-xyz",
		},
		{
			Slug:          "customer-abc",
			Name:          "Customer ABC",
			PublicBaseURL: "https://events.example.com/customer-abc",
		},
	} {
		if _, err := repo.CreateTenant(context.Background(), item); err != nil {
			t.Fatalf("CreateTenant returned error: %v", err)
		}
	}

	_, err := repo.LookupByPublicHost(context.Background(), "events.example.com")
	if !errors.Is(err, ErrTenantHostAmbiguous) {
		t.Fatalf("expected ErrTenantHostAmbiguous, got %v", err)
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

func TestUpdateTenant(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	created, err := repo.CreateTenant(context.Background(), CreateTenantParams{
		Slug:          "demo",
		Name:          "Demo Tenant",
		PublicBaseURL: "https://events.example.com/demo",
	})
	if err != nil {
		t.Fatalf("CreateTenant returned error: %v", err)
	}

	name := "Demo Tenant Updated"
	baseURL := "https://events-updated.example.com"
	timezone := "UTC"
	locale := "en-GB"
	updated, err := repo.UpdateTenant(context.Background(), created.ID, UpdateTenantParams{
		Name:            &name,
		PublicBaseURL:   &baseURL,
		DefaultTimezone: &timezone,
		DefaultLocale:   &locale,
	})
	if err != nil {
		t.Fatalf("UpdateTenant returned error: %v", err)
	}

	if updated.Name != name {
		t.Fatalf("expected updated name %q, got %q", name, updated.Name)
	}
	if updated.PublicBaseURL != baseURL {
		t.Fatalf("expected updated public base url %q, got %q", baseURL, updated.PublicBaseURL)
	}
	if updated.DefaultTimezone != timezone {
		t.Fatalf("expected updated timezone %q, got %q", timezone, updated.DefaultTimezone)
	}
	if updated.DefaultLocale != locale {
		t.Fatalf("expected updated locale %q, got %q", locale, updated.DefaultLocale)
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

func TestSeedTenantCreatesOrUpdatesAdminUser(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	first, err := repo.SeedTenant(context.Background(), SeedInput{
		Slug:          "community",
		Name:          "Community Tenant",
		PublicBaseURL: "https://events.example.com/community",
		AdminUser: SeedAdminUserInput{
			Email: "Admin@Example.com",
			Name:  "First Admin",
			Role:  "owner",
		},
	})
	if err != nil {
		t.Fatalf("first SeedTenant returned error: %v", err)
	}

	var (
		email  string
		name   string
		role   string
		status string
	)
	if err := repo.db.QueryRowContext(
		context.Background(),
		`SELECT email, name, role, status
     FROM tenant_users
     WHERE tenant_id = ?`,
		first.Tenant.ID,
	).Scan(&email, &name, &role, &status); err != nil {
		t.Fatalf("query seeded admin user: %v", err)
	}

	if email != "admin@example.com" {
		t.Fatalf("expected normalized admin email, got %q", email)
	}
	if name != "First Admin" {
		t.Fatalf("expected admin name First Admin, got %q", name)
	}
	if role != "owner" {
		t.Fatalf("expected admin role owner, got %q", role)
	}
	if status != "active" {
		t.Fatalf("expected admin status active, got %q", status)
	}

	second, err := repo.SeedTenant(context.Background(), SeedInput{
		Slug:          "community",
		Name:          "Community Tenant",
		PublicBaseURL: "https://events.example.com/community",
		AdminUser: SeedAdminUserInput{
			Email: "admin@example.com",
			Name:  "Updated Admin",
			Role:  "event_manager",
		},
	})
	if err != nil {
		t.Fatalf("second SeedTenant returned error: %v", err)
	}

	var count int
	if err := repo.db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*) FROM tenant_users WHERE tenant_id = ? AND lower(email) = ?`,
		second.Tenant.ID,
		"admin@example.com",
	).Scan(&count); err != nil {
		t.Fatalf("count seeded admin users: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one admin user, got %d", count)
	}

	if err := repo.db.QueryRowContext(
		context.Background(),
		`SELECT name, role FROM tenant_users WHERE tenant_id = ? AND lower(email) = ?`,
		second.Tenant.ID,
		"admin@example.com",
	).Scan(&name, &role); err != nil {
		t.Fatalf("query updated seeded admin user: %v", err)
	}
	if name != "Updated Admin" {
		t.Fatalf("expected updated admin name, got %q", name)
	}
	if role != "event_manager" {
		t.Fatalf("expected updated admin role event_manager, got %q", role)
	}
}

func TestSeedTenantRejectsInvalidAdminEmail(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	_, err := repo.SeedTenant(context.Background(), SeedInput{
		Slug:          "community",
		Name:          "Community Tenant",
		PublicBaseURL: "https://events.example.com/community",
		AdminUser: SeedAdminUserInput{
			Email: "not-an-email",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "parse seed admin email") {
		t.Fatalf("expected invalid admin email error, got %v", err)
	}
}

func TestLookupBySlugNotFound(t *testing.T) {
	repo := NewRepository(newMigratedDB(t))

	_, err := repo.LookupBySlug(context.Background(), "missing")
	if !errors.Is(err, ErrTenantNotFound) {
		t.Fatalf("expected ErrTenantNotFound, got %v", err)
	}
}
