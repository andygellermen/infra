package registration

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/auth"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func setupRegistrationService(t *testing.T) (*Service, *sql.DB, tenant.Tenant) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "registration-test.sqlite"))
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
	tenantItem, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "demo",
		Name:          "Demo Tenant",
		PublicBaseURL: "https://events.example.com/demo",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	service := NewService(sqlDB, Config{
		BaseURL:         "https://events.example.com",
		TokenPepper:     "test-pepper",
		RegistrationTTL: 30 * time.Minute,
	})

	return service, sqlDB, tenantItem
}

func createPublishedEventForRegistration(t *testing.T, dbHandle *sql.DB, tenantID string, params event.CreateEventParams) event.Event {
	t.Helper()

	repo := event.NewRepository(dbHandle)
	created, err := repo.CreateEvent(context.Background(), tenantID, params)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	published, err := repo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}
	return published
}

func extractVerifyTokenFromLatestEmailJob(t *testing.T, dbHandle *sql.DB, tenantID, registrationID string) string {
	t.Helper()

	rows, err := dbHandle.QueryContext(
		context.Background(),
		`SELECT body_text, COALESCE(metadata_json, '')
     FROM email_jobs
     WHERE tenant_id = ? AND template_key = ?
     ORDER BY created_at DESC`,
		tenantID,
		DefaultVerificationTemplate,
	)
	if err != nil {
		t.Fatalf("query latest email job body text: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var bodyText string
		var metadataJSON string
		if err := rows.Scan(&bodyText, &metadataJSON); err != nil {
			t.Fatalf("scan email job row: %v", err)
		}
		if strings.TrimSpace(registrationID) != "" && !strings.Contains(metadataJSON, registrationID) {
			continue
		}
		for _, field := range strings.Fields(bodyText) {
			if !strings.Contains(field, "/registrations/verify?token=") {
				continue
			}
			parsed, err := url.Parse(strings.TrimSpace(field))
			if err != nil {
				t.Fatalf("parse verify url %q: %v", field, err)
			}
			token := parsed.Query().Get("token")
			if token != "" {
				return token
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate email jobs: %v", err)
	}

	t.Fatalf("no verification token found for registration %q", registrationID)
	return ""
}

func TestStartAndVerifyRegistrationFlow(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:     "open-workshop",
		Title:    "Open Workshop",
		StartsAt: time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
	})

	start, err := service.Start(context.Background(), StartInput{
		TenantID:          tenantItem.ID,
		TenantSlug:        tenantItem.Slug,
		EventID:           eventItem.ID,
		Name:              "Max Mustermann",
		Email:             "max@example.com",
		ParticipationType: event.ParticipationModeOnsite,
		PrivacyAccepted:   true,
		RequestIP:         "127.0.0.1",
		UserAgent:         "unit-test",
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if start.Status != StatusVerificationPending {
		t.Fatalf("expected verification_pending status, got %q", start.Status)
	}

	token := extractVerifyTokenFromLatestEmailJob(t, dbHandle, tenantItem.ID, start.RegistrationID)
	verifyResult, err := service.Verify(context.Background(), VerifyInput{
		TenantID:  tenantItem.ID,
		RawToken:  token,
		RequestIP: "127.0.0.1",
		UserAgent: "unit-test",
	})
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if verifyResult.Status != StatusConfirmed {
		t.Fatalf("expected confirmed status, got %q", verifyResult.Status)
	}
	if verifyResult.ConfirmedAt == nil {
		t.Fatalf("expected confirmed_at to be set")
	}

	var status string
	if err := dbHandle.QueryRowContext(
		context.Background(),
		`SELECT status FROM registrations WHERE id = ?`,
		start.RegistrationID,
	).Scan(&status); err != nil {
		t.Fatalf("query registration status: %v", err)
	}
	if status != StatusConfirmed {
		t.Fatalf("expected registration status confirmed, got %q", status)
	}

	var emailVerifiedAt sql.NullString
	if err := dbHandle.QueryRowContext(
		context.Background(),
		`SELECT email_verified_at FROM participants WHERE id = ?`,
		start.ParticipantID,
	).Scan(&emailVerifiedAt); err != nil {
		t.Fatalf("query participant email_verified_at: %v", err)
	}
	if !emailVerifiedAt.Valid {
		t.Fatalf("expected participant email_verified_at to be set")
	}
}

func TestVerifyReturnsEventFullWhenWaitlistDisabled(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	waitlistDisabled := false
	maxOne := 1
	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:            "tiny-room",
		Title:           "Tiny Room",
		StartsAt:        time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		MaxParticipants: &maxOne,
		WaitlistEnabled: &waitlistDisabled,
	})

	_, err := service.Start(context.Background(), StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Alice",
		Email:           "alice@example.com",
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("first Start returned error: %v", err)
	}
	firstToken := extractVerifyTokenFromLatestEmailJob(t, dbHandle, tenantItem.ID, "")
	if _, err := service.Verify(context.Background(), VerifyInput{TenantID: tenantItem.ID, RawToken: firstToken}); err != nil {
		t.Fatalf("first Verify returned error: %v", err)
	}

	secondStart, err := service.Start(context.Background(), StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Bob",
		Email:           "bob@example.com",
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("second Start returned error: %v", err)
	}
	secondToken := extractVerifyTokenFromLatestEmailJob(t, dbHandle, tenantItem.ID, secondStart.RegistrationID)
	_, err = service.Verify(context.Background(), VerifyInput{TenantID: tenantItem.ID, RawToken: secondToken})
	if !errors.Is(err, ErrEventFull) {
		t.Fatalf("expected ErrEventFull, got %v", err)
	}

	var status string
	if err := dbHandle.QueryRowContext(
		context.Background(),
		`SELECT status FROM registrations WHERE id = ?`,
		secondStart.RegistrationID,
	).Scan(&status); err != nil {
		t.Fatalf("query second registration status: %v", err)
	}
	if status != StatusExpired {
		t.Fatalf("expected expired status for full event verification, got %q", status)
	}
}

func TestVerifyMovesToWaitlistWhenEnabled(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	waitlistEnabled := true
	maxOne := 1
	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:            "with-waitlist",
		Title:           "With Waitlist",
		StartsAt:        time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		MaxParticipants: &maxOne,
		WaitlistEnabled: &waitlistEnabled,
	})

	firstStart, err := service.Start(context.Background(), StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Alice",
		Email:           "alice@example.com",
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("first Start returned error: %v", err)
	}
	firstToken := extractVerifyTokenFromLatestEmailJob(t, dbHandle, tenantItem.ID, firstStart.RegistrationID)
	if _, err := service.Verify(context.Background(), VerifyInput{TenantID: tenantItem.ID, RawToken: firstToken}); err != nil {
		t.Fatalf("first Verify returned error: %v", err)
	}
	if firstStart.RegistrationID == "" {
		t.Fatalf("expected first registration id")
	}

	secondStart, err := service.Start(context.Background(), StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Bob",
		Email:           "bob@example.com",
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("second Start returned error: %v", err)
	}
	secondToken := extractVerifyTokenFromLatestEmailJob(t, dbHandle, tenantItem.ID, secondStart.RegistrationID)
	verifyResult, err := service.Verify(context.Background(), VerifyInput{TenantID: tenantItem.ID, RawToken: secondToken})
	if err != nil {
		t.Fatalf("second Verify returned error: %v", err)
	}
	if verifyResult.Status != StatusWaitlist {
		t.Fatalf("expected waitlist status, got %q", verifyResult.Status)
	}
	if verifyResult.WaitlistPos != 1 {
		t.Fatalf("expected waitlist position 1, got %d", verifyResult.WaitlistPos)
	}
	if strings.TrimSpace(verifyResult.WaitlistID) == "" {
		t.Fatalf("expected waitlist id to be set")
	}
	if secondStart.RegistrationID == "" {
		t.Fatalf("expected second registration id")
	}

	var waitlistCount int
	if err := dbHandle.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*) FROM waitlist_entries WHERE registration_id = ?`,
		secondStart.RegistrationID,
	).Scan(&waitlistCount); err != nil {
		t.Fatalf("query waitlist entries: %v", err)
	}
	if waitlistCount != 1 {
		t.Fatalf("expected one waitlist entry, got %d", waitlistCount)
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	service, dbHandle, tenantItem := setupRegistrationService(t)
	eventItem := createPublishedEventForRegistration(t, dbHandle, tenantItem.ID, event.CreateEventParams{
		Slug:     "expiry-test",
		Title:    "Expiry Test",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	start, err := service.Start(context.Background(), StartInput{
		TenantID:        tenantItem.ID,
		TenantSlug:      tenantItem.Slug,
		EventID:         eventItem.ID,
		Name:            "Max",
		Email:           "max@example.com",
		PrivacyAccepted: true,
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if start.RegistrationID == "" {
		t.Fatalf("expected registration id")
	}

	token := extractVerifyTokenFromLatestEmailJob(t, dbHandle, tenantItem.ID, start.RegistrationID)
	if _, err := dbHandle.ExecContext(
		context.Background(),
		`UPDATE magic_links
     SET expires_at = ?
     WHERE tenant_id = ? AND purpose = ?`,
		time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
		tenantItem.ID,
		auth.PurposeRegistrationVerify,
	); err != nil {
		t.Fatalf("expire token: %v", err)
	}

	_, err = service.Verify(context.Background(), VerifyInput{
		TenantID: tenantItem.ID,
		RawToken: token,
	})
	if !errors.Is(err, ErrExpiredVerificationToken) {
		t.Fatalf("expected ErrExpiredVerificationToken, got %v", err)
	}
}
