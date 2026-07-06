package httpapp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

	updateSettingsBody, _ := json.Marshal(map[string]any{
		"sender_email":           "events@example.com",
		"sender_name":            "Event Team",
		"default_retention_days": 45,
		"app_settings": map[string]any{
			"event_time_start":        "09:00",
			"event_time_end":          "21:00",
			"event_time_step_minutes": 30,
			"event_slug_mode":         "required",
			"allowed_embed_origins":   []string{"https://www.geller.men", "https://ghost.geller.men"},
			"event_detail_base_url":   "https://www.geller.men/events",
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
}
