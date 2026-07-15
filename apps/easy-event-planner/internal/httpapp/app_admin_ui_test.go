package httpapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func TestAdminUIRoutesServeShellAndAssets(t *testing.T) {
	app := New(testConfig(), nil)

	tests := []struct {
		name         string
		method       string
		path         string
		wantStatus   int
		wantType     string
		wantContains string
	}{
		{
			name:         "admin shell",
			method:       http.MethodGet,
			path:         "/admin",
			wantStatus:   http.StatusOK,
			wantType:     "text/html",
			wantContains: "Admin Cockpit",
		},
		{
			name:         "admin css",
			method:       http.MethodGet,
			path:         "/admin-ui.css",
			wantStatus:   http.StatusOK,
			wantType:     "text/css",
			wantContains: "--brand:",
		},
		{
			name:         "admin js",
			method:       http.MethodGet,
			path:         "/admin-ui.js",
			wantStatus:   http.StatusOK,
			wantType:     "application/javascript",
			wantContains: "apiRequest",
		},
		{
			name:         "smoke footer include page",
			method:       http.MethodGet,
			path:         "/smoke/footer-include.html",
			wantStatus:   http.StatusOK,
			wantType:     "text/html",
			wantContains: "/smoke/footer-include.html?tenant=demo&config=dein-snippet-slug",
		},
		{
			name:         "admin slash redirects",
			method:       http.MethodGet,
			path:         "/admin/",
			wantStatus:   http.StatusTemporaryRedirect,
			wantType:     "text/html",
			wantContains: "<a href=\"/admin\">",
		},
		{
			name:         "admin method not allowed",
			method:       http.MethodPost,
			path:         "/admin",
			wantStatus:   http.StatusMethodNotAllowed,
			wantType:     "text/plain",
			wantContains: "method not allowed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			app.Handler().ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}
			if got := rec.Header().Get("Content-Type"); !strings.Contains(got, tc.wantType) {
				t.Fatalf("expected content-type containing %q, got %q", tc.wantType, got)
			}
			if body := rec.Body.String(); !strings.Contains(body, tc.wantContains) {
				t.Fatalf("expected body to contain %q, got %q", tc.wantContains, body)
			}
		})
	}
}

func TestAdminUIRouteDelegatesAdminTenantAssets(t *testing.T) {
	app, _, _ := setupAuthApp(t)

	_, err := app.tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "admin",
		Name:          "Admin Tenant",
		PublicBaseURL: "http://localhost:8080/admin",
	})
	if err != nil {
		t.Fatalf("create admin tenant: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/include.js", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected include.js status 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/javascript") {
		t.Fatalf("expected js content type, got %q", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "const tenantSlug = \"admin\";") {
		t.Fatalf("expected include.js to include tenant slug admin, got %q", body)
	}
	if !strings.Contains(body, "/api/v1/public/\" + encodeURIComponent(tenantSlug) + \"/snippet/events?") {
		t.Fatalf("expected include.js to target snippet events endpoint, got %q", body)
	}
}

func TestAdminUIContainsSnippetEditAndDailyRecurrence(t *testing.T) {
	app := New(testConfig(), nil)

	shellReq := httptest.NewRequest(http.MethodGet, "/admin", nil)
	shellRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(shellRec, shellReq)
	if shellRec.Code != http.StatusOK {
		t.Fatalf("expected admin shell status 200, got %d", shellRec.Code)
	}
	shellBody := shellRec.Body.String()
	if !strings.Contains(shellBody, "EEP-CSS automatisch laden") {
		t.Fatalf("expected snippet css toggle in admin shell")
	}
	if !strings.Contains(shellBody, "value=\"daily\"") {
		t.Fatalf("expected daily recurrence option in admin shell")
	}
	if !strings.Contains(shellBody, "eventDetailBaseUrlHint") {
		t.Fatalf("expected detail base url hint in admin shell")
	}
	if !strings.Contains(shellBody, ".../events/{event_slug}") {
		t.Fatalf("expected detail base url route hint in admin shell")
	}
	if !strings.Contains(shellBody, "eventPublicationHint") {
		t.Fatalf("expected publication hint in admin shell")
	}
	if !strings.Contains(shellBody, "Nach Freigabe in oeffentlicher Uebersicht anzeigen") {
		t.Fatalf("expected clarified public visibility wording in admin shell")
	}

	jsReq := httptest.NewRequest(http.MethodGet, "/admin-ui.js", nil)
	jsRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(jsRec, jsReq)
	if jsRec.Code != http.StatusOK {
		t.Fatalf("expected admin js status 200, got %d", jsRec.Code)
	}
	jsBody := jsRec.Body.String()
	if !strings.Contains(jsBody, "populateSnippetFormForEdit") {
		t.Fatalf("expected snippet edit support in admin js")
	}
	if !strings.Contains(jsBody, "mode === \"daily\"") {
		t.Fatalf("expected daily recurrence handling in admin js")
	}
	if !strings.Contains(jsBody, "updateEventDetailBaseURLHint") {
		t.Fatalf("expected detail base url preview support in admin js")
	}
	if !strings.Contains(jsBody, "updateEventPublicationHint") {
		t.Fatalf("expected publication hint support in admin js")
	}
	if !strings.Contains(jsBody, "getPublicationMeta") {
		t.Fatalf("expected publication meta helper in admin js")
	}
}
