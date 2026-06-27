package httpapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		AppName:               "easy-event-planner",
		Env:                   "test",
		Addr:                  ":0",
		BaseURL:               "http://localhost:8080",
		Version:               "test-version",
		DBDriver:              "sqlite",
		DBPath:                "./data/test.sqlite",
		CertificateStorageDir: "./certificates",
		TokenPepper:           "test-pepper",
		SessionCookieName:     "eep_session",
		SecureCookies:         false,
		SessionTTL:            12 * time.Hour,
		MagicLinkTTL:          15 * time.Minute,
		RegistrationTTL:       30 * time.Minute,
		WaitlistOfferTTL:      24 * time.Hour,
		CertificateTTL:        30 * time.Minute,
		AuthRateLimit:         5,
		AuthRateWindow:        15 * time.Minute,
		ReadHeaderTimeout:     5 * time.Second,
		ShutdownTimeout:       10 * time.Second,
	}
}

func TestSystemRoutes(t *testing.T) {
	app := New(testConfig(), nil)

	tests := []struct {
		name         string
		method       string
		path         string
		wantStatus   int
		wantContains string
		wantType     string
	}{
		{name: "root", method: http.MethodGet, path: "/", wantStatus: http.StatusOK, wantContains: "Easy Event Planner", wantType: "text/html"},
		{name: "healthz", method: http.MethodGet, path: "/healthz", wantStatus: http.StatusOK, wantContains: "ok", wantType: "text/plain"},
		{name: "readyz", method: http.MethodGet, path: "/readyz", wantStatus: http.StatusOK, wantContains: "ready", wantType: "text/plain"},
		{name: "version", method: http.MethodGet, path: "/version", wantStatus: http.StatusOK, wantContains: "test-version", wantType: "application/json"},
		{name: "healthz method", method: http.MethodPost, path: "/healthz", wantStatus: http.StatusMethodNotAllowed, wantContains: "method not allowed", wantType: "text/plain"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			app.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}

			if body := rec.Body.String(); !strings.Contains(body, tc.wantContains) {
				t.Fatalf("expected body to contain %q, got %q", tc.wantContains, body)
			}

			if gotType := rec.Header().Get("Content-Type"); !strings.Contains(gotType, tc.wantType) {
				t.Fatalf("expected content type containing %q, got %q", tc.wantType, gotType)
			}
		})
	}
}

func TestVersionPayloadShape(t *testing.T) {
	app := New(testConfig(), nil)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if payload["service"] != "easy-event-planner" {
		t.Fatalf("expected service easy-event-planner, got %q", payload["service"])
	}
	if payload["version"] != "test-version" {
		t.Fatalf("expected version test-version, got %q", payload["version"])
	}
	if payload["env"] != "test" {
		t.Fatalf("expected env test, got %q", payload["env"])
	}
	if _, err := time.Parse(time.RFC3339, payload["started_at"]); err != nil {
		t.Fatalf("expected started_at in RFC3339, got %q", payload["started_at"])
	}
}
