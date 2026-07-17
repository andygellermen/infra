package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func authenticateAdminSession(t *testing.T, app *App, sender *fakeMagicLinkSender, host string) *http.Cookie {
	t.Helper()

	requestPayload := map[string]any{
		"email":   "owner@example.com",
		"purpose": "organizer_login",
	}
	requestBody, _ := json.Marshal(requestPayload)
	requestReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-link/request", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestReq.Host = host
	requestRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(requestRec, requestReq)

	if requestRec.Code != http.StatusOK {
		t.Fatalf("expected request status 200, got %d", requestRec.Code)
	}

	verifyPayload := map[string]any{
		"token": extractTokenFromVerifyURL(t, sender.lastMessage.VerifyURL),
	}
	verifyBody, _ := json.Marshal(verifyPayload)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/magic-link/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)

	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected verify status 200, got %d", verifyRec.Code)
	}
	cookies := verifyRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
	return cookies[0]
}

func TestAdminTenantProfileAndSettings(t *testing.T) {
	app, sender, _ := setupAuthApp(t)
	sessionCookie := authenticateAdminSession(t, app, sender, "localhost:8080")

	getTenantReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenant", nil)
	getTenantReq.AddCookie(sessionCookie)
	getTenantRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getTenantRec, getTenantReq)

	if getTenantRec.Code != http.StatusOK {
		t.Fatalf("expected tenant get status 200, got %d", getTenantRec.Code)
	}
	tenantPayload := decodeBody[map[string]any](t, getTenantRec)
	item := tenantPayload["item"].(map[string]any)
	if item["slug"] != "customerxyz" {
		t.Fatalf("expected tenant slug customerxyz, got %v", item["slug"])
	}

	updateTenantBody, _ := json.Marshal(map[string]any{
		"name":             "Customer XYZ Updated",
		"public_base_url":  "https://events.example.com",
		"default_timezone": "UTC",
		"default_locale":   "en-GB",
	})
	updateTenantReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenant", bytes.NewReader(updateTenantBody))
	updateTenantReq.Header.Set("Content-Type", "application/json")
	updateTenantReq.AddCookie(sessionCookie)
	updateTenantRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(updateTenantRec, updateTenantReq)

	if updateTenantRec.Code != http.StatusOK {
		t.Fatalf("expected tenant patch status 200, got %d", updateTenantRec.Code)
	}

	getSettingsReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenant/settings", nil)
	getSettingsReq.AddCookie(sessionCookie)
	getSettingsRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getSettingsRec, getSettingsReq)

	if getSettingsRec.Code != http.StatusOK {
		t.Fatalf("expected settings get status 200, got %d", getSettingsRec.Code)
	}
	settingsPayload := decodeBody[map[string]any](t, getSettingsRec)
	settingsItem := settingsPayload["item"].(map[string]any)
	appSettings := settingsItem["app_settings"].(map[string]any)
	if appSettings["event_time_start"] != "08:00" {
		t.Fatalf("expected default event_time_start 08:00, got %v", appSettings["event_time_start"])
	}
	if appSettings["participant_cancel_deadline_hours"] != float64(24) {
		t.Fatalf("expected default participant_cancel_deadline_hours 24, got %v", appSettings["participant_cancel_deadline_hours"])
	}

	updateSettingsBody, _ := json.Marshal(map[string]any{
		"sender_email":           "events@example.com",
		"sender_name":            "Event Team",
		"default_retention_days": 45,
		"app_settings": map[string]any{
			"event_time_start":                  "09:00",
			"event_time_end":                    "21:00",
			"event_time_step_minutes":           30,
			"event_slug_mode":                   "required",
			"allowed_embed_origins":             []string{"https://www.geller.men", "https://ghost.geller.men"},
			"event_detail_base_url":             "https://www.geller.men/events",
			"participant_cancel_deadline_hours": 48,
		},
	})
	updateSettingsReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenant/settings", bytes.NewReader(updateSettingsBody))
	updateSettingsReq.Header.Set("Content-Type", "application/json")
	updateSettingsReq.AddCookie(sessionCookie)
	updateSettingsRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(updateSettingsRec, updateSettingsReq)

	if updateSettingsRec.Code != http.StatusOK {
		t.Fatalf("expected settings patch status 200, got %d", updateSettingsRec.Code)
	}

	updatedSettingsPayload := decodeBody[map[string]any](t, updateSettingsRec)
	updatedItem := updatedSettingsPayload["item"].(map[string]any)
	if updatedItem["sender_email"] != "events@example.com" {
		t.Fatalf("expected updated sender_email, got %v", updatedItem["sender_email"])
	}
	updatedAppSettings := updatedItem["app_settings"].(map[string]any)
	if updatedAppSettings["event_slug_mode"] != "required" {
		t.Fatalf("expected event_slug_mode required, got %v", updatedAppSettings["event_slug_mode"])
	}
	if updatedAppSettings["event_time_step_minutes"] != float64(30) {
		t.Fatalf("expected event_time_step_minutes 30, got %v", updatedAppSettings["event_time_step_minutes"])
	}
	if updatedAppSettings["participant_cancel_deadline_hours"] != float64(48) {
		t.Fatalf("expected participant_cancel_deadline_hours 48, got %v", updatedAppSettings["participant_cancel_deadline_hours"])
	}
}

func TestAdminTenantPatchRejectsConflictingPublicBaseURL(t *testing.T) {
	app, sender, _ := setupAuthApp(t)
	sessionCookie := authenticateAdminSession(t, app, sender, "localhost:8080")

	_, err := app.tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "second-tenant",
		Name:          "Second Tenant",
		PublicBaseURL: "https://events.example.com/second",
	})
	if err != nil {
		t.Fatalf("create second tenant: %v", err)
	}

	updateTenantBody, _ := json.Marshal(map[string]any{
		"name":             "Customer XYZ Updated",
		"public_base_url":  "https://events.example.com/second",
		"default_timezone": "UTC",
		"default_locale":   "en-GB",
	})
	updateTenantReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenant", bytes.NewReader(updateTenantBody))
	updateTenantReq.Header.Set("Content-Type", "application/json")
	updateTenantReq.AddCookie(sessionCookie)
	updateTenantRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(updateTenantRec, updateTenantReq)

	if updateTenantRec.Code != http.StatusConflict {
		t.Fatalf("expected tenant patch status 409, got %d", updateTenantRec.Code)
	}
	payload := decodeBody[map[string]any](t, updateTenantRec)
	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != "TENANT_PUBLIC_BASE_URL_CONFLICT" {
		t.Fatalf("expected TENANT_PUBLIC_BASE_URL_CONFLICT, got %v", errorPayload["code"])
	}
}

func TestAdminTenantDomainBindingsLifecycle(t *testing.T) {
	app, sender, _ := setupAuthApp(t)
	sessionCookie := authenticateAdminSession(t, app, sender, "localhost:8080")

	createBody, _ := json.Marshal(map[string]any{
		"domain":     "events.customer-domain.example",
		"base_path":  "/",
		"status":     "pending_dns",
		"is_primary": false,
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenant/domains", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	createPayload := decodeBody[map[string]any](t, createRec)
	createdItem := createPayload["item"].(map[string]any)
	if createdItem["domain"] != "events.customer-domain.example" {
		t.Fatalf("expected created domain, got %v", createdItem["domain"])
	}
	bindingID := createdItem["id"].(string)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenant/domains", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	if listPayload["dns_target_host"] == "" {
		t.Fatalf("expected dns_target_host in payload")
	}
	items := listPayload["items"].([]any)
	firstItem := items[0].(map[string]any)
	if firstItem["verification_record_name"] == "" || firstItem["verification_record_value"] == "" {
		t.Fatalf("expected verification record metadata in payload")
	}

	updateBody, _ := json.Marshal(map[string]any{
		"status":     "active",
		"is_primary": true,
	})
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenant/domains/"+bindingID, bytes.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(sessionCookie)
	updateRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d", updateRec.Code)
	}
	updatePayload := decodeBody[map[string]any](t, updateRec)
	updatedItem := updatePayload["item"].(map[string]any)
	if updatedItem["is_primary"] != true {
		t.Fatalf("expected binding to be primary, got %v", updatedItem["is_primary"])
	}

	getTenantReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenant", nil)
	getTenantReq.AddCookie(sessionCookie)
	getTenantRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getTenantRec, getTenantReq)
	if getTenantRec.Code != http.StatusOK {
		t.Fatalf("expected tenant get status 200, got %d", getTenantRec.Code)
	}
	tenantPayload := decodeBody[map[string]any](t, getTenantRec)
	tenantItem := tenantPayload["item"].(map[string]any)
	if tenantItem["public_base_url"] != "https://events.customer-domain.example" {
		t.Fatalf("expected synced public_base_url, got %v", tenantItem["public_base_url"])
	}

	lockedTenantBody, _ := json.Marshal(map[string]any{
		"name":             "Customer XYZ Updated",
		"public_base_url":  "https://another.example.com",
		"default_timezone": "UTC",
		"default_locale":   "en-GB",
	})
	lockedTenantReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/tenant", bytes.NewReader(lockedTenantBody))
	lockedTenantReq.Header.Set("Content-Type", "application/json")
	lockedTenantReq.AddCookie(sessionCookie)
	lockedTenantRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(lockedTenantRec, lockedTenantReq)
	if lockedTenantRec.Code != http.StatusConflict {
		t.Fatalf("expected tenant patch status 409, got %d", lockedTenantRec.Code)
	}
	lockedPayload := decodeBody[map[string]any](t, lockedTenantRec)
	lockedError := lockedPayload["error"].(map[string]any)
	if lockedError["code"] != "TENANT_PUBLIC_BASE_URL_MANAGED_BY_DOMAIN_BINDING" {
		t.Fatalf("expected TENANT_PUBLIC_BASE_URL_MANAGED_BY_DOMAIN_BINDING, got %v", lockedError["code"])
	}
}

func TestAdminTenantDomainRefreshCheckAndRotateToken(t *testing.T) {
	app, sender, _ := setupAuthApp(t)
	sessionCookie := authenticateAdminSession(t, app, sender, "localhost:8080")

	createBody, _ := json.Marshal(map[string]any{
		"domain":    "events.customer-domain.example",
		"base_path": "/",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenant/domains", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	bindingItem := decodeBody[map[string]any](t, createRec)["item"].(map[string]any)
	bindingID := bindingItem["id"].(string)
	oldToken := bindingItem["verification_record_value"].(string)

	oldTXT := tenantDomainLookupTXT
	oldCNAME := tenantDomainLookupCNAME
	oldHost := tenantDomainLookupHost
	oldTLS := tenantDomainTLSProbe
	t.Cleanup(func() {
		tenantDomainLookupTXT = oldTXT
		tenantDomainLookupCNAME = oldCNAME
		tenantDomainLookupHost = oldHost
		tenantDomainTLSProbe = oldTLS
	})

	tenantDomainLookupTXT = func(_ context.Context, name string) ([]string, error) {
		if name != "_eep-domain-verification.events.customer-domain.example" {
			t.Fatalf("unexpected TXT name %q", name)
		}
		return []string{oldToken}, nil
	}
	tenantDomainLookupCNAME = func(_ context.Context, host string) (string, error) {
		return "localhost", nil
	}
	tenantDomainLookupHost = func(_ context.Context, host string) ([]string, error) {
		switch host {
		case "events.customer-domain.example", "localhost":
			return []string{"127.0.0.1"}, nil
		default:
			return []string{"127.0.0.2"}, nil
		}
	}
	tenantDomainTLSProbe = func(_ context.Context, host string) (domainTLSProbeResult, error) {
		expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)
		return domainTLSProbeResult{
			Status:    tenant.DomainBindingSSLStatusValid,
			Issuer:    "Local Test CA",
			ExpiresAt: &expiresAt,
		}, nil
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenant/domains/"+bindingID+"/refresh-check", nil)
	refreshReq.AddCookie(sessionCookie)
	refreshRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(refreshRec, refreshReq)
	if refreshRec.Code != http.StatusOK {
		t.Fatalf("expected refresh status 200, got %d", refreshRec.Code)
	}
	refreshedItem := decodeBody[map[string]any](t, refreshRec)["item"].(map[string]any)
	if refreshedItem["status"] != tenant.DomainBindingStatusActive {
		t.Fatalf("expected active status after successful checks, got %v", refreshedItem["status"])
	}
	if refreshedItem["ssl_status"] != tenant.DomainBindingSSLStatusValid {
		t.Fatalf("expected ssl_status valid, got %v", refreshedItem["ssl_status"])
	}

	rotateReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/tenant/domains/"+bindingID+"/rotate-verification-token", nil)
	rotateReq.AddCookie(sessionCookie)
	rotateRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rotateRec, rotateReq)
	if rotateRec.Code != http.StatusOK {
		t.Fatalf("expected rotate status 200, got %d", rotateRec.Code)
	}
	rotatedItem := decodeBody[map[string]any](t, rotateRec)["item"].(map[string]any)
	if rotatedItem["verification_record_value"] == oldToken {
		t.Fatalf("expected rotated verification token")
	}
	if rotatedItem["status"] != tenant.DomainBindingStatusPendingDNS {
		t.Fatalf("expected pending_dns after token rotation, got %v", rotatedItem["status"])
	}
}
