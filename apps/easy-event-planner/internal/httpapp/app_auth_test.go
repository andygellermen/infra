package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/auth"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

type fakeMagicLinkSender struct {
	lastMessage auth.MagicLinkMessage
}

func (s *fakeMagicLinkSender) SendMagicLink(_ context.Context, message auth.MagicLinkMessage) error {
	s.lastMessage = message
	return nil
}

func setupAuthApp(t *testing.T) (*App, *fakeMagicLinkSender, string) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "http-auth.sqlite"))
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

	repo := tenant.NewRepository(sqlDB)
	createdTenant, err := repo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "customerxyz",
		Name:          "Customer XYZ",
		PublicBaseURL: "http://localhost:8080/customerxyz",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	if _, err := sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO tenant_users (id, tenant_id, email, name, role, status, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"usr_http_001",
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

	cfg := testConfig()
	cfg.TokenPepper = "test-pepper"
	cfg.SessionCookieName = "eep_session"
	cfg.CertificateStorageDir = filepath.Join(t.TempDir(), "certificates")
	cfg.SecureCookies = false
	cfg.SessionTTL = 12 * time.Hour
	cfg.MagicLinkTTL = 15 * time.Minute
	cfg.RegistrationTTL = 30 * time.Minute
	cfg.WaitlistOfferTTL = 24 * time.Hour
	cfg.CertificateTTL = 30 * time.Minute
	cfg.AuthRateLimit = 5
	cfg.AuthRateWindow = 15 * time.Minute

	app := New(cfg, sqlDB)
	fakeSender := &fakeMagicLinkSender{}
	app.authService.SetSender(fakeSender)

	return app, fakeSender, createdTenant.Slug
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var payload T
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return payload
}

func TestAuthHTTPFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)

	requestPayload := map[string]any{
		"tenant_slug": tenantSlug,
		"email":       "owner@example.com",
		"purpose":     "organizer_login",
	}
	requestBody, _ := json.Marshal(requestPayload)

	requestReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-link/request", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestReq.RemoteAddr = "127.0.0.1:12345"
	requestRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(requestRec, requestReq)

	if requestRec.Code != http.StatusOK {
		t.Fatalf("expected request status 200, got %d", requestRec.Code)
	}
	if sender.lastMessage.VerifyURL == "" {
		t.Fatalf("expected magic link sender to receive verify url")
	}

	verifyPayload := map[string]any{
		"token": extractTokenFromVerifyURL(t, sender.lastMessage.VerifyURL),
	}
	verifyBody, _ := json.Marshal(verifyPayload)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-link/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyReq.RemoteAddr = "127.0.0.1:12345"
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)

	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected verify status 200, got %d", verifyRec.Code)
	}
	cookies := verifyRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie to be set")
	}
	sessionCookie := cookies[0]

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	meRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusOK {
		t.Fatalf("expected me status 200, got %d", meRec.Code)
	}
	mePayload := decodeBody[map[string]any](t, meRec)
	if mePayload["authenticated"] != true {
		t.Fatalf("expected authenticated=true, got %v", mePayload["authenticated"])
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(logoutRec, logoutReq)

	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("expected logout status 204, got %d", logoutRec.Code)
	}

	meAfterLogoutReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meAfterLogoutReq.AddCookie(sessionCookie)
	meAfterLogoutRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(meAfterLogoutRec, meAfterLogoutReq)

	if meAfterLogoutRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected me-after-logout status 401, got %d", meAfterLogoutRec.Code)
	}
}

func TestMagicLinkRequestNeutralForUnknownEmail(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)

	requestPayload := map[string]any{
		"tenant_slug": tenantSlug,
		"email":       "missing@example.com",
		"purpose":     "organizer_login",
	}
	requestBody, _ := json.Marshal(requestPayload)

	requestReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-link/request", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(requestRec, requestReq)

	if requestRec.Code != http.StatusOK {
		t.Fatalf("expected request status 200, got %d", requestRec.Code)
	}

	payload := decodeBody[map[string]any](t, requestRec)
	if payload["ok"] != true {
		t.Fatalf("expected neutral ok response, got %v", payload["ok"])
	}
}

func extractTokenFromVerifyURL(t *testing.T, verifyURL string) string {
	t.Helper()
	parsed, err := url.Parse(verifyURL)
	if err != nil {
		t.Fatalf("parse verify url: %v", err)
	}
	token := parsed.Query().Get("token")
	if token == "" {
		t.Fatalf("missing token in verify url")
	}
	return token
}
