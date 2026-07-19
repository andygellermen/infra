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
	if !strings.Contains(shellBody, "Externe Detailseiten-Basis-URL") {
		t.Fatalf("expected clarified external detail base url label in admin shell")
	}
	if !strings.Contains(shellBody, "Public-Domains") {
		t.Fatalf("expected public domains card in admin shell")
	}
	if !strings.Contains(shellBody, "tenantDomainForm") {
		t.Fatalf("expected tenant domain form in admin shell")
	}
	if !strings.Contains(shellBody, "DNS vorbereiten") || !strings.Contains(shellBody, "SSL ausstehend") {
		t.Fatalf("expected clarified domain lifecycle options in admin shell")
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
	if !strings.Contains(shellBody, "public_visible_from") || !strings.Contains(shellBody, "registration_opens_at") || !strings.Contains(shellBody, "registration_closes_at") {
		t.Fatalf("expected publication and registration window fields in admin shell")
	}
	if !strings.Contains(shellBody, "form-subtab-close") || !strings.Contains(shellBody, "title=\"Bearbeitung abbrechen\"") {
		t.Fatalf("expected compact event edit cancel action in event form tabs")
	}
	if !strings.Contains(shellBody, "eventResetChangesBtn") || !strings.Contains(shellBody, "Aenderungen zuruecksetzen") {
		t.Fatalf("expected reset action in event form tabs")
	}
	if !strings.Contains(shellBody, "form id=\"eventForm\" class=\"form-grid\" novalidate") {
		t.Fatalf("expected event form to disable native browser validation")
	}
	if !strings.Contains(shellBody, "Preis in EUR") || !strings.Contains(shellBody, "placeholder=\"49,00\"") {
		t.Fatalf("expected event payment inputs to use euro amounts")
	}
	if !strings.Contains(shellBody, "eventTitleInput") || !strings.Contains(shellBody, "Eventtitel") {
		t.Fatalf("expected prominent event title field in editor header")
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
	if !strings.Contains(jsBody, "Aktuelle externe Detailseiten-Vorschau") {
		t.Fatalf("expected clarified external detail page preview copy in admin js")
	}
	if !strings.Contains(jsBody, "saveTenantDomain") {
		t.Fatalf("expected tenant domain save support in admin js")
	}
	if !strings.Contains(jsBody, "findPrimaryTenantDomain") {
		t.Fatalf("expected primary tenant domain helper in admin js")
	}
	if !strings.Contains(jsBody, "rotate-verification-token") || !strings.Contains(jsBody, "refresh-check") {
		t.Fatalf("expected domain verification and refresh actions in admin js")
	}
	if !strings.Contains(jsBody, "updateEventPublicationHint") {
		t.Fatalf("expected publication hint support in admin js")
	}
	if !strings.Contains(jsBody, "getPublicationMeta") {
		t.Fatalf("expected publication meta helper in admin js")
	}
	if !strings.Contains(jsBody, "applySteppedDateTimeInputConfig") {
		t.Fatalf("expected stepped datetime configuration in admin js")
	}
	if !strings.Contains(jsBody, "applyRegistrationActivationShortcut") {
		t.Fatalf("expected registration activation shortcut in admin js")
	}
	if !strings.Contains(jsBody, "currentSteppedLocalDateTimeValue") {
		t.Fatalf("expected stepped current datetime helper in admin js")
	}
	if !strings.Contains(jsBody, "confirmRegistrationActivationShortcut") {
		t.Fatalf("expected registration activation confirmation in admin js")
	}
	if !strings.Contains(jsBody, "Moechtest du fortfahren?") {
		t.Fatalf("expected confirmation copy in admin js")
	}
	if !strings.Contains(jsBody, "parseOptionalEuroAmountToCents") || !strings.Contains(jsBody, "Bitte keinen Punkt verwenden") {
		t.Fatalf("expected euro amount parsing and validation helpers in admin js")
	}
	if !strings.Contains(jsBody, "updateEventEditorChrome") || !strings.Contains(jsBody, "resetEventEditorChanges") {
		t.Fatalf("expected persistent editor chrome helpers in admin js")
	}
}
