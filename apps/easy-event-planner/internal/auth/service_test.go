package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

type captureSender struct {
	messages []MagicLinkMessage
}

func (s *captureSender) SendMagicLink(_ context.Context, message MagicLinkMessage) error {
	s.messages = append(s.messages, message)
	return nil
}

func setupAuthService(t *testing.T, cfg Config) (*Service, *captureSender, string, string, *sql.DB) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "auth-test.sqlite"))
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
		Slug:          "customerxyz",
		Name:          "Customer XYZ",
		PublicBaseURL: "https://events.example.com/customerxyz",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	userID := "usr_test_001"
	if _, err := sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO tenant_users (id, tenant_id, email, name, role, status, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID,
		createdTenant.ID,
		"owner@example.com",
		"Owner",
		"owner",
		"active",
		time.Now().UTC().Format(time.RFC3339),
		time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		t.Fatalf("insert tenant user: %v", err)
	}

	sender := &captureSender{}
	service := NewService(sqlDB, tenantRepo, cfg, sender)
	return service, sender, createdTenant.Slug, userID, sqlDB
}

func testAuthConfig() Config {
	return Config{
		BaseURL:           "https://events.example.com",
		TokenPepper:       "pepper-test",
		SessionTTL:        12 * time.Hour,
		MagicLinkTTL:      15 * time.Minute,
		RegistrationTTL:   30 * time.Minute,
		WaitlistOfferTTL:  24 * time.Hour,
		CertificateTTL:    30 * time.Minute,
		RateLimitRequests: 5,
		RateLimitWindow:   15 * time.Minute,
	}
}

func extractTokenFromURL(t *testing.T, rawURL string) string {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse verify url: %v", err)
	}
	token := parsed.Query().Get("token")
	if token == "" {
		t.Fatalf("missing token in verify url %q", rawURL)
	}
	return token
}

func TestRequestAndVerifyMagicLinkFlow(t *testing.T) {
	service, sender, tenantSlug, userID, sqlDB := setupAuthService(t, testAuthConfig())

	requestResult, err := service.RequestMagicLink(context.Background(), RequestMagicLinkInput{
		TenantSlug:   tenantSlug,
		Email:        "owner@example.com",
		Purpose:      PurposeOrganizerLogin,
		RedirectPath: "/admin/dashboard",
		RequestIP:    "127.0.0.1",
		UserAgent:    "unit-test",
	})
	if err != nil {
		t.Fatalf("RequestMagicLink returned error: %v", err)
	}
	if !requestResult.Accepted || !requestResult.Sent {
		t.Fatalf("expected accepted+sent request result, got %+v", requestResult)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("expected one sent message, got %d", len(sender.messages))
	}

	token := extractTokenFromURL(t, sender.messages[0].VerifyURL)

	var tokenHash string
	if err := sqlDB.QueryRow(`SELECT token_hash FROM magic_links LIMIT 1`).Scan(&tokenHash); err != nil {
		t.Fatalf("query token hash: %v", err)
	}
	if tokenHash == "" {
		t.Fatalf("expected token hash to be stored")
	}
	if strings.Contains(tokenHash, token) {
		t.Fatalf("token hash must not contain raw token")
	}

	verifyResult, err := service.VerifyMagicLink(context.Background(), VerifyMagicLinkInput{
		RawToken:  token,
		RequestIP: "127.0.0.1",
		UserAgent: "unit-test",
	})
	if err != nil {
		t.Fatalf("VerifyMagicLink returned error: %v", err)
	}
	if verifyResult.SessionToken == "" {
		t.Fatalf("expected session token to be created")
	}
	if verifyResult.RedirectPath != "/admin/dashboard" {
		t.Fatalf("expected redirect path /admin/dashboard, got %q", verifyResult.RedirectPath)
	}
	if verifyResult.UserID != userID {
		t.Fatalf("expected user id %q, got %q", userID, verifyResult.UserID)
	}

	if _, err := service.VerifyMagicLink(context.Background(), VerifyMagicLinkInput{RawToken: token}); !errors.Is(err, ErrInvalidMagicLink) {
		t.Fatalf("expected second verify to fail with ErrInvalidMagicLink, got %v", err)
	}
}

func TestAuthenticateAndRevokeSession(t *testing.T) {
	service, sender, tenantSlug, _, _ := setupAuthService(t, testAuthConfig())

	if _, err := service.RequestMagicLink(context.Background(), RequestMagicLinkInput{
		TenantSlug: tenantSlug,
		Email:      "owner@example.com",
		Purpose:    PurposeOrganizerLogin,
	}); err != nil {
		t.Fatalf("RequestMagicLink returned error: %v", err)
	}
	token := extractTokenFromURL(t, sender.messages[0].VerifyURL)

	verifyResult, err := service.VerifyMagicLink(context.Background(), VerifyMagicLinkInput{RawToken: token})
	if err != nil {
		t.Fatalf("VerifyMagicLink returned error: %v", err)
	}

	principal, err := service.AuthenticateSession(context.Background(), verifyResult.SessionToken)
	if err != nil {
		t.Fatalf("AuthenticateSession returned error: %v", err)
	}
	if principal.Email != "owner@example.com" {
		t.Fatalf("expected email owner@example.com, got %q", principal.Email)
	}

	revoked, err := service.RevokeSession(context.Background(), verifyResult.SessionToken, "127.0.0.1", "unit-test")
	if err != nil {
		t.Fatalf("RevokeSession returned error: %v", err)
	}
	if !revoked {
		t.Fatalf("expected revoke to return true")
	}

	if _, err := service.AuthenticateSession(context.Background(), verifyResult.SessionToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected revoked session to be invalid, got %v", err)
	}
}

func TestRequestMagicLinkNeutralForUnknownEmail(t *testing.T) {
	service, sender, tenantSlug, _, _ := setupAuthService(t, testAuthConfig())

	result, err := service.RequestMagicLink(context.Background(), RequestMagicLinkInput{
		TenantSlug: tenantSlug,
		Email:      "unknown@example.com",
		Purpose:    PurposeOrganizerLogin,
	})
	if err != nil {
		t.Fatalf("RequestMagicLink returned error: %v", err)
	}
	if !result.Accepted || result.Sent {
		t.Fatalf("expected accepted=true sent=false for unknown email, got %+v", result)
	}
	if len(sender.messages) != 0 {
		t.Fatalf("expected no mail for unknown email")
	}
}

func TestRequestMagicLinkResolvesTenantByRequestHost(t *testing.T) {
	service, sender, _, _, _ := setupAuthService(t, testAuthConfig())

	result, err := service.RequestMagicLink(context.Background(), RequestMagicLinkInput{
		Email:       "owner@example.com",
		Purpose:     PurposeOrganizerLogin,
		RequestHost: "events.example.com",
	})
	if err != nil {
		t.Fatalf("RequestMagicLink returned error: %v", err)
	}
	if !result.Accepted || !result.Sent {
		t.Fatalf("expected accepted=true sent=true, got %+v", result)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("expected one sent message, got %d", len(sender.messages))
	}
}

func TestRequestMagicLinkRateLimit(t *testing.T) {
	cfg := testAuthConfig()
	cfg.RateLimitRequests = 1
	cfg.RateLimitWindow = time.Hour

	service, _, tenantSlug, _, _ := setupAuthService(t, cfg)

	if _, err := service.RequestMagicLink(context.Background(), RequestMagicLinkInput{
		TenantSlug: tenantSlug,
		Email:      "owner@example.com",
		Purpose:    PurposeOrganizerLogin,
		RequestIP:  "127.0.0.1",
	}); err != nil {
		t.Fatalf("first request should succeed, got %v", err)
	}

	if _, err := service.RequestMagicLink(context.Background(), RequestMagicLinkInput{
		TenantSlug: tenantSlug,
		Email:      "owner@example.com",
		Purpose:    PurposeOrganizerLogin,
		RequestIP:  "127.0.0.1",
	}); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestRequestAndVerifyParticipantPortalMagicLinkFlow(t *testing.T) {
	service, sender, tenantSlug, _, sqlDB := setupAuthService(t, testAuthConfig())

	tenantID := ""
	if err := sqlDB.QueryRowContext(
		context.Background(),
		`SELECT id FROM tenants WHERE slug = ? LIMIT 1`,
		tenantSlug,
	).Scan(&tenantID); err != nil {
		t.Fatalf("query tenant id: %v", err)
	}

	eventRepo := event.NewRepository(sqlDB)
	createdEvent, err := eventRepo.CreateEvent(context.Background(), tenantID, event.CreateEventParams{
		Slug:     "participant-portal-event",
		Title:    "Participant Portal Event",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	participantID := "par_portal_001"
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO participants (id, tenant_id, email, name, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?)`,
		participantID,
		tenantID,
		"participant@example.com",
		"Participant One",
		now,
		now,
	); err != nil {
		t.Fatalf("insert participant: %v", err)
	}
	if _, err := sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO registrations (
      id, tenant_id, event_id, participant_id, status, participation_type, quantity,
      source, privacy_accepted_at, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?)`,
		"reg_portal_001",
		tenantID,
		createdEvent.ID,
		participantID,
		"confirmed",
		"onsite",
		"public_page",
		now,
		now,
		now,
	); err != nil {
		t.Fatalf("insert registration: %v", err)
	}

	requestResult, err := service.RequestMagicLink(context.Background(), RequestMagicLinkInput{
		TenantSlug: tenantSlug,
		Email:      "participant@example.com",
		Purpose:    PurposeParticipantLogin,
	})
	if err != nil {
		t.Fatalf("request participant magic link: %v", err)
	}
	if !requestResult.Accepted || !requestResult.Sent {
		t.Fatalf("expected accepted+sent participant link request, got %+v", requestResult)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("expected one sent participant message, got %d", len(sender.messages))
	}

	token := extractTokenFromURL(t, sender.messages[0].VerifyURL)
	verifyResult, err := service.VerifyMagicLink(context.Background(), VerifyMagicLinkInput{
		RawToken: token,
	})
	if err != nil {
		t.Fatalf("verify participant magic link: %v", err)
	}
	if verifyResult.UserID != "" {
		t.Fatalf("expected empty user id for participant session, got %q", verifyResult.UserID)
	}
	if verifyResult.ParticipantID != participantID {
		t.Fatalf("expected participant id %q, got %q", participantID, verifyResult.ParticipantID)
	}
	if verifyResult.SessionToken == "" {
		t.Fatalf("expected participant session token")
	}

	participantPrincipal, err := service.AuthenticateParticipantSession(context.Background(), verifyResult.SessionToken)
	if err != nil {
		t.Fatalf("authenticate participant session: %v", err)
	}
	if participantPrincipal.ParticipantID != participantID {
		t.Fatalf("expected participant principal id %q, got %q", participantID, participantPrincipal.ParticipantID)
	}
	if participantPrincipal.Email != "participant@example.com" {
		t.Fatalf("expected participant email participant@example.com, got %q", participantPrincipal.Email)
	}

	if _, err := service.AuthenticateSession(context.Background(), verifyResult.SessionToken); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected participant session not to authenticate as admin session, got %v", err)
	}
}
