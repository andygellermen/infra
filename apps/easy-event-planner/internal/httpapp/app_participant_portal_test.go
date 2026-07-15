package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func registrationAndParticipantForPortalHTTPTest(t *testing.T, app *App, tenantSlug string, eventID string, email string) (string, string) {
	t.Helper()
	registrationID := createConfirmedRegistrationForPaymentHTTPTest(t, app, tenantSlug, eventID, email)

	var participantID string
	if err := app.db.QueryRowContext(
		context.Background(),
		`SELECT participant_id FROM registrations WHERE id = ? LIMIT 1`,
		registrationID,
	).Scan(&participantID); err != nil {
		t.Fatalf("query registration participant id: %v", err)
	}
	if participantID == "" {
		t.Fatalf("expected participant id for registration %s", registrationID)
	}
	return registrationID, participantID
}

func participantSessionCookieFromResponse(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == participantSessionCookieName {
			return cookie
		}
	}
	t.Fatalf("expected participant session cookie %q", participantSessionCookieName)
	return nil
}

func TestParticipantPortalHTTPFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)

	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "participant-portal-http",
		Title:    "Participant Portal HTTP",
		StartsAt: time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
	})
	registrationID, participantID := registrationAndParticipantForPortalHTTPTest(
		t,
		app,
		tenantSlug,
		eventItem.ID,
		"portal-http@example.com",
	)

	requestPayload := map[string]any{
		"email": "portal-http@example.com",
	}
	requestBody, _ := json.Marshal(requestPayload)
	requestReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/request", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(requestRec, requestReq)
	if requestRec.Code != http.StatusOK {
		t.Fatalf("expected portal request status 200, got %d", requestRec.Code)
	}
	if sender.lastMessage.VerifyURL == "" {
		t.Fatalf("expected participant portal verify url in sender")
	}
	if sender.lastMessage.Purpose != "participant_login" {
		t.Fatalf("expected participant_login purpose, got %q", sender.lastMessage.Purpose)
	}

	verifyToken := extractTokenFromVerifyURL(t, sender.lastMessage.VerifyURL)
	verifyPayload := map[string]any{"token": verifyToken}
	verifyBody, _ := json.Marshal(verifyPayload)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected portal verify status 200, got %d", verifyRec.Code)
	}
	verifyResult := decodeBody[map[string]any](t, verifyRec)
	if verifyResult["participant_id"] != participantID {
		t.Fatalf("expected participant_id %q, got %v", participantID, verifyResult["participant_id"])
	}
	participantCookie := participantSessionCookieFromResponse(t, verifyRec)

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/participants/portal/me", nil)
	meReq.AddCookie(participantCookie)
	meRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("expected portal me status 200, got %d", meRec.Code)
	}
	mePayload := decodeBody[map[string]any](t, meRec)
	participantPayload := mePayload["participant"].(map[string]any)
	if participantPayload["id"] != participantID {
		t.Fatalf("expected me participant id %q, got %v", participantID, participantPayload["id"])
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/participants/portal/registrations", nil)
	listReq.AddCookie(participantCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected portal registrations status 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	items, ok := listPayload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected at least one registration item in portal list")
	}
	firstItem := items[0].(map[string]any)
	if firstItem["self_cancel_allowed"] != true {
		t.Fatalf("expected self_cancel_allowed true, got %v", firstItem["self_cancel_allowed"])
	}
	if firstItem["self_cancel_deadline_hours"] != float64(24) {
		t.Fatalf("expected self_cancel_deadline_hours 24, got %v", firstItem["self_cancel_deadline_hours"])
	}

	cancelPayload := map[string]any{"reason": "Teilnahme storniert im Portal"}
	cancelBody, _ := json.Marshal(cancelPayload)
	cancelReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/registrations/"+registrationID+"/cancel", bytes.NewReader(cancelBody))
	cancelReq.Header.Set("Content-Type", "application/json")
	cancelReq.AddCookie(participantCookie)
	cancelRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusOK {
		t.Fatalf("expected portal cancel status 200, got %d", cancelRec.Code)
	}
	cancelResult := decodeBody[map[string]any](t, cancelRec)
	cancelItem := cancelResult["item"].(map[string]any)
	if cancelItem["status"] != "cancelled" {
		t.Fatalf("expected cancelled status, got %v", cancelItem["status"])
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/logout", nil)
	logoutReq.AddCookie(participantCookie)
	logoutRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("expected portal logout status 204, got %d", logoutRec.Code)
	}

	meAfterLogoutReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/participants/portal/me", nil)
	meAfterLogoutReq.AddCookie(participantCookie)
	meAfterLogoutRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(meAfterLogoutRec, meAfterLogoutReq)
	if meAfterLogoutRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized me after logout, got %d", meAfterLogoutRec.Code)
	}
}

func TestParticipantPortalCancelDeadlineExceededHTTP(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	if _, err := app.tenantRepo.UpsertSettings(context.Background(), tenant.UpsertTenantSettingsParams{
		TenantID: tenantID,
		Settings: tenant.TenantSettingsInput{
			SettingsJSON: `{"participant_cancel_deadline_hours":24}`,
		},
	}); err != nil {
		t.Fatalf("upsert tenant settings: %v", err)
	}

	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "participant-portal-too-late",
		Title:    "Participant Portal Too Late",
		StartsAt: time.Now().UTC().Add(12 * time.Hour).Format(time.RFC3339),
	})
	registrationID, _ := registrationAndParticipantForPortalHTTPTest(
		t,
		app,
		tenantSlug,
		eventItem.ID,
		"portal-too-late@example.com",
	)

	requestPayload := map[string]any{
		"email": "portal-too-late@example.com",
	}
	requestBody, _ := json.Marshal(requestPayload)
	requestReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/request", bytes.NewReader(requestBody))
	requestReq.Header.Set("Content-Type", "application/json")
	requestRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(requestRec, requestReq)
	if requestRec.Code != http.StatusOK {
		t.Fatalf("expected portal request status 200, got %d", requestRec.Code)
	}

	verifyToken := extractTokenFromVerifyURL(t, sender.lastMessage.VerifyURL)
	verifyPayload := map[string]any{"token": verifyToken}
	verifyBody, _ := json.Marshal(verifyPayload)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected portal verify status 200, got %d", verifyRec.Code)
	}
	participantCookie := participantSessionCookieFromResponse(t, verifyRec)

	cancelPayload := map[string]any{"reason": "Zu spaet"}
	cancelBody, _ := json.Marshal(cancelPayload)
	cancelReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/participants/portal/registrations/"+registrationID+"/cancel", bytes.NewReader(cancelBody))
	cancelReq.Header.Set("Content-Type", "application/json")
	cancelReq.AddCookie(participantCookie)
	cancelRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusConflict {
		t.Fatalf("expected portal cancel status 409, got %d", cancelRec.Code)
	}
	payload := decodeBody[map[string]any](t, cancelRec)
	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != "REGISTRATION_CANCEL_DEADLINE_EXCEEDED" {
		t.Fatalf("expected REGISTRATION_CANCEL_DEADLINE_EXCEEDED, got %v", errorPayload["code"])
	}
}
