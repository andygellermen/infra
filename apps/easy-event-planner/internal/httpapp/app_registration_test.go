package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func tenantIDBySlug(t *testing.T, app *App, tenantSlug string) string {
	t.Helper()
	tenantItem, err := app.tenantRepo.LookupBySlug(context.Background(), tenantSlug)
	if err != nil {
		t.Fatalf("lookup tenant by slug: %v", err)
	}
	return tenantItem.ID
}

func createPublishedEventForRegistrationHTTP(t *testing.T, app *App, tenantID string, params event.CreateEventParams) event.Event {
	t.Helper()
	created, err := app.eventRepo.CreateEvent(context.Background(), tenantID, params)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	published, err := app.eventRepo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}
	return published
}

func extractVerifyTokenFromJobInHTTPTest(t *testing.T, app *App, tenantID, registrationID string) string {
	t.Helper()
	rows, err := app.db.QueryContext(
		context.Background(),
		`SELECT body_text, COALESCE(metadata_json, '')
     FROM email_jobs
     WHERE tenant_id = ? AND template_key = ?
     ORDER BY created_at DESC`,
		tenantID,
		"registration_verify",
	)
	if err != nil {
		t.Fatalf("query verify email jobs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var bodyText string
		var metadataJSON string
		if err := rows.Scan(&bodyText, &metadataJSON); err != nil {
			t.Fatalf("scan verify email job: %v", err)
		}
		if strings.TrimSpace(registrationID) != "" && !strings.Contains(metadataJSON, registrationID) {
			continue
		}
		for _, field := range strings.Fields(bodyText) {
			if !strings.Contains(field, "/registrations/verify?token=") {
				continue
			}
			parsed, err := url.Parse(field)
			if err != nil {
				t.Fatalf("parse verify URL %q: %v", field, err)
			}
			token := parsed.Query().Get("token")
			if token != "" {
				return token
			}
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate verify email jobs: %v", err)
	}

	t.Fatalf("no token found for registration %q", registrationID)
	return ""
}

func TestPublicRegistrationStartAndVerifyFlow(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "public-registration",
		Title:    "Public Registration",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	startPayload := map[string]any{
		"event_id":           eventItem.ID,
		"name":               "Max Mustermann",
		"email":              "max@example.com",
		"participation_type": "onsite",
		"privacy_accepted":   true,
	}
	startBody, _ := json.Marshal(startPayload)
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/start", bytes.NewReader(startBody))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusAccepted {
		t.Fatalf("expected start status 202, got %d", startRec.Code)
	}
	startResult := decodeBody[map[string]any](t, startRec)
	if startResult["status"] != "verification_pending" {
		t.Fatalf("expected status verification_pending, got %v", startResult["status"])
	}
	registrationID, _ := startResult["registration_id"].(string)
	if registrationID == "" {
		t.Fatalf("expected registration id in response")
	}

	token := extractVerifyTokenFromJobInHTTPTest(t, app, tenantID, registrationID)
	verifyPayload := map[string]any{
		"token": token,
	}
	verifyBody, _ := json.Marshal(verifyPayload)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected verify status 200, got %d", verifyRec.Code)
	}
	verifyResult := decodeBody[map[string]any](t, verifyRec)
	if verifyResult["status"] != "confirmed" {
		t.Fatalf("expected status confirmed, got %v", verifyResult["status"])
	}
	if verifyResult["registration_id"] != registrationID {
		t.Fatalf("expected registration id %q, got %v", registrationID, verifyResult["registration_id"])
	}
}

func TestPublicRegistrationVerifyReturnsEventFull(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	waitlistEnabled := false
	maxParticipants := 1
	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:            "public-registration-full",
		Title:           "Public Registration Full",
		StartsAt:        time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		WaitlistEnabled: &waitlistEnabled,
		MaxParticipants: &maxParticipants,
	})

	registerAndVerify := func(name, email string) *httptest.ResponseRecorder {
		t.Helper()
		startPayload := map[string]any{
			"event_id":           eventItem.ID,
			"name":               name,
			"email":              email,
			"participation_type": "onsite",
			"privacy_accepted":   true,
		}
		startBody, _ := json.Marshal(startPayload)
		startReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/start", bytes.NewReader(startBody))
		startReq.Header.Set("Content-Type", "application/json")
		startRec := httptest.NewRecorder()
		app.Handler().ServeHTTP(startRec, startReq)
		if startRec.Code != http.StatusAccepted {
			t.Fatalf("expected start status 202, got %d", startRec.Code)
		}
		startResult := decodeBody[map[string]any](t, startRec)
		registrationID, _ := startResult["registration_id"].(string)

		token := extractVerifyTokenFromJobInHTTPTest(t, app, tenantID, registrationID)
		verifyPayload := map[string]any{"token": token}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/verify", bytes.NewReader(verifyBody))
		verifyReq.Header.Set("Content-Type", "application/json")
		verifyRec := httptest.NewRecorder()
		app.Handler().ServeHTTP(verifyRec, verifyReq)
		return verifyRec
	}

	firstVerify := registerAndVerify("Alice", "alice@example.com")
	if firstVerify.Code != http.StatusOK {
		t.Fatalf("expected first verify 200, got %d", firstVerify.Code)
	}

	secondVerify := registerAndVerify("Bob", "bob@example.com")
	if secondVerify.Code != http.StatusConflict {
		t.Fatalf("expected second verify 409, got %d", secondVerify.Code)
	}
	errorPayload := decodeBody[map[string]any](t, secondVerify)["error"].(map[string]any)
	if errorPayload["code"] != "EVENT_FULL" {
		t.Fatalf("expected EVENT_FULL, got %v", errorPayload["code"])
	}
}

func TestPublicRegistrationVerifyMovesToWaitlist(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	waitlistEnabled := true
	maxParticipants := 1
	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:            "public-registration-waitlist",
		Title:           "Public Registration Waitlist",
		StartsAt:        time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		WaitlistEnabled: &waitlistEnabled,
		MaxParticipants: &maxParticipants,
	})

	registerAndVerify := func(name, email string) map[string]any {
		t.Helper()
		startPayload := map[string]any{
			"event_id":           eventItem.ID,
			"name":               name,
			"email":              email,
			"participation_type": "onsite",
			"privacy_accepted":   true,
		}
		startBody, _ := json.Marshal(startPayload)
		startReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/start", bytes.NewReader(startBody))
		startReq.Header.Set("Content-Type", "application/json")
		startRec := httptest.NewRecorder()
		app.Handler().ServeHTTP(startRec, startReq)
		if startRec.Code != http.StatusAccepted {
			t.Fatalf("expected start status 202, got %d", startRec.Code)
		}
		startResult := decodeBody[map[string]any](t, startRec)
		registrationID, _ := startResult["registration_id"].(string)

		token := extractVerifyTokenFromJobInHTTPTest(t, app, tenantID, registrationID)
		verifyPayload := map[string]any{"token": token}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/verify", bytes.NewReader(verifyBody))
		verifyReq.Header.Set("Content-Type", "application/json")
		verifyRec := httptest.NewRecorder()
		app.Handler().ServeHTTP(verifyRec, verifyReq)
		if verifyRec.Code != http.StatusOK {
			t.Fatalf("expected verify status 200, got %d", verifyRec.Code)
		}
		return decodeBody[map[string]any](t, verifyRec)
	}

	first := registerAndVerify("Alice", "alice@example.com")
	if first["status"] != "confirmed" {
		t.Fatalf("expected first status confirmed, got %v", first["status"])
	}

	second := registerAndVerify("Bob", "bob@example.com")
	if second["status"] != "waitlist" {
		t.Fatalf("expected second status waitlist, got %v", second["status"])
	}
	waitlist, ok := second["waitlist"].(map[string]any)
	if !ok {
		t.Fatalf("expected waitlist payload")
	}
	if waitlist["position"] != float64(1) {
		t.Fatalf("expected waitlist position 1, got %v", waitlist["position"])
	}
}

func TestPublicRegistrationCORSAllowedOrigin(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	if _, err := app.tenantRepo.UpsertSettings(context.Background(), tenant.UpsertTenantSettingsParams{
		TenantID: tenantID,
		Settings: tenant.TenantSettingsInput{
			SettingsJSON: `{"allowed_embed_origins":["https://ghost.geller.men"]}`,
		},
	}); err != nil {
		t.Fatalf("upsert tenant settings: %v", err)
	}

	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "cors-registration",
		Title:    "CORS Registration",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	preflightReq := httptest.NewRequest(http.MethodOptions, "/api/v1/public/"+tenantSlug+"/registrations/start", nil)
	preflightReq.Header.Set("Origin", "https://ghost.geller.men")
	preflightReq.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflightRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(preflightRec, preflightReq)
	if preflightRec.Code != http.StatusNoContent {
		t.Fatalf("expected preflight status 204, got %d", preflightRec.Code)
	}
	if preflightRec.Header().Get("Access-Control-Allow-Origin") != "https://ghost.geller.men" {
		t.Fatalf("expected allow origin header, got %q", preflightRec.Header().Get("Access-Control-Allow-Origin"))
	}

	startPayload := map[string]any{
		"event_id":           eventItem.ID,
		"name":               "CORS User",
		"email":              "cors@example.com",
		"participation_type": "onsite",
		"privacy_accepted":   true,
	}
	startBody, _ := json.Marshal(startPayload)
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/start", bytes.NewReader(startBody))
	startReq.Header.Set("Origin", "https://ghost.geller.men")
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusAccepted {
		t.Fatalf("expected start status 202, got %d", startRec.Code)
	}
	if startRec.Header().Get("Access-Control-Allow-Origin") != "https://ghost.geller.men" {
		t.Fatalf("expected allow origin header on POST, got %q", startRec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestPublicRegistrationCORSRejectsUnknownOrigin(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/public/"+tenantSlug+"/registrations/start", nil)
	req.Header.Set("Origin", "https://unknown.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
	payload := decodeBody[map[string]any](t, rec)
	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != "CORS_ORIGIN_NOT_ALLOWED" {
		t.Fatalf("expected CORS_ORIGIN_NOT_ALLOWED, got %v", errorPayload["code"])
	}
}

func TestPublicRegistrationStartRequiresPrivacyAcceptance(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "privacy-required",
		Title:    "Privacy Required",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	startPayload := map[string]any{
		"event_id":           eventItem.ID,
		"name":               "Max Mustermann",
		"email":              "max@example.com",
		"participation_type": "onsite",
		"privacy_accepted":   false,
	}
	startBody, _ := json.Marshal(startPayload)
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/start", bytes.NewReader(startBody))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusBadRequest {
		t.Fatalf("expected start status 400, got %d", startRec.Code)
	}
	errorPayload := decodeBody[map[string]any](t, startRec)["error"].(map[string]any)
	if errorPayload["code"] != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %v", errorPayload["code"])
	}
}
