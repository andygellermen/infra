package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAdminEventCRUDAndPublishFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"slug":               "sommer-retreat",
		"title":              "Sommer Retreat",
		"starts_at":          "2026-08-10T08:00:00Z",
		"ends_at":            "2026-08-10T16:00:00Z",
		"timezone":           "Europe/Berlin",
		"participation_mode": "hybrid",
		"location_name":      "Haus am See",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	createResult := decodeBody[map[string]any](t, createRec)
	item, ok := createResult["item"].(map[string]any)
	if !ok {
		t.Fatalf("expected item payload in create response")
	}
	eventID, ok := item["id"].(string)
	if !ok || eventID == "" {
		t.Fatalf("expected created event id")
	}
	if item["status"] != "draft" {
		t.Fatalf("expected draft status on create, got %v", item["status"])
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}
	listResult := decodeBody[map[string]any](t, listRec)
	if listResult["total"] != float64(1) {
		t.Fatalf("expected total=1, got %v", listResult["total"])
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID, nil)
	getReq.AddCookie(sessionCookie)
	getRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected get status 200, got %d", getRec.Code)
	}

	patchPayload := map[string]any{
		"title":            "Sommer Retreat Plus",
		"max_participants": 80,
	}
	patchBody, _ := json.Marshal(patchPayload)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/events/"+eventID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(sessionCookie)
	patchRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected patch status 200, got %d", patchRec.Code)
	}
	patchResult := decodeBody[map[string]any](t, patchRec)
	updatedItem, ok := patchResult["item"].(map[string]any)
	if !ok {
		t.Fatalf("expected patch response item")
	}
	if updatedItem["title"] != "Sommer Retreat Plus" {
		t.Fatalf("expected updated title, got %v", updatedItem["title"])
	}

	publishReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/publish", nil)
	publishReq.AddCookie(sessionCookie)
	publishRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(publishRec, publishReq)
	if publishRec.Code != http.StatusOK {
		t.Fatalf("expected publish status 200, got %d", publishRec.Code)
	}
	publishResult := decodeBody[map[string]any](t, publishRec)
	publishedItem := publishResult["item"].(map[string]any)
	if publishedItem["status"] != "scheduled" {
		t.Fatalf("expected scheduled status after publish, got %v", publishedItem["status"])
	}
	if publishedItem["is_public"] != true {
		t.Fatalf("expected is_public=true after publish, got %v", publishedItem["is_public"])
	}

	unpublishReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/unpublish", nil)
	unpublishReq.AddCookie(sessionCookie)
	unpublishRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(unpublishRec, unpublishReq)
	if unpublishRec.Code != http.StatusOK {
		t.Fatalf("expected unpublish status 200, got %d", unpublishRec.Code)
	}
	unpublishResult := decodeBody[map[string]any](t, unpublishRec)
	unpublishedItem := unpublishResult["item"].(map[string]any)
	if unpublishedItem["status"] != "draft" {
		t.Fatalf("expected draft status after unpublish, got %v", unpublishedItem["status"])
	}
	if unpublishedItem["is_public"] != false {
		t.Fatalf("expected is_public=false after unpublish, got %v", unpublishedItem["is_public"])
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/events/"+eventID, nil)
	deleteReq.AddCookie(sessionCookie)
	deleteRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d", deleteRec.Code)
	}

	getAfterDeleteReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID, nil)
	getAfterDeleteReq.AddCookie(sessionCookie)
	getAfterDeleteRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getAfterDeleteRec, getAfterDeleteReq)
	if getAfterDeleteRec.Code != http.StatusNotFound {
		t.Fatalf("expected get-after-delete status 404, got %d", getAfterDeleteRec.Code)
	}
}

func TestAdminEventListCountsAndArchiveFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"slug":               "wochenend-workshop",
		"title":              "Wochenend Workshop",
		"starts_at":          "2026-08-14T16:00:00Z",
		"ends_at":            "2026-08-16T14:00:00Z",
		"timezone":           "Europe/Berlin",
		"participation_mode": "onsite",
		"location_name":      "Seminarhaus",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	eventID := decodeBody[map[string]any](t, createRec)["item"].(map[string]any)["id"].(string)

	publishReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/publish", nil)
	publishReq.AddCookie(sessionCookie)
	publishRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(publishRec, publishReq)
	if publishRec.Code != http.StatusOK {
		t.Fatalf("expected publish status 200, got %d", publishRec.Code)
	}

	registrationPayload := map[string]any{
		"name":               "Max Mustermann",
		"email":              "max@example.com",
		"participation_type": "onsite",
	}
	registrationBody, _ := json.Marshal(registrationPayload)
	registrationReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/registrations/manual", bytes.NewReader(registrationBody))
	registrationReq.Header.Set("Content-Type", "application/json")
	registrationReq.AddCookie(sessionCookie)
	registrationRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(registrationRec, registrationReq)
	if registrationRec.Code != http.StatusCreated {
		t.Fatalf("expected manual registration status 201, got %d", registrationRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}
	items := decodeBody[map[string]any](t, listRec)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected exactly one event in list, got %d", len(items))
	}
	listed := items[0].(map[string]any)
	if listed["confirmed_participants"] != float64(1) {
		t.Fatalf("expected confirmed_participants=1, got %v", listed["confirmed_participants"])
	}
	if listed["waitlist_entries"] != float64(0) {
		t.Fatalf("expected waitlist_entries=0, got %v", listed["waitlist_entries"])
	}

	archiveReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/archive", nil)
	archiveReq.AddCookie(sessionCookie)
	archiveRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(archiveRec, archiveReq)
	if archiveRec.Code != http.StatusOK {
		t.Fatalf("expected archive status 200, got %d", archiveRec.Code)
	}
	archived := decodeBody[map[string]any](t, archiveRec)["item"].(map[string]any)
	if archived["status"] != "archived" {
		t.Fatalf("expected archived status, got %v", archived["status"])
	}
	if archived["is_public"] != false {
		t.Fatalf("expected archived event to be hidden, got %v", archived["is_public"])
	}
	if archived["registration_enabled"] != false {
		t.Fatalf("expected archived event registration disabled, got %v", archived["registration_enabled"])
	}

	rePublishReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/publish", nil)
	rePublishReq.AddCookie(sessionCookie)
	rePublishRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rePublishRec, rePublishReq)
	if rePublishRec.Code != http.StatusConflict {
		t.Fatalf("expected publish-after-archive status 409, got %d", rePublishRec.Code)
	}
}

func TestAdminEventRegistrationEmbedCode(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"slug":                 "embed-event",
		"title":                "Embed Event",
		"starts_at":            "2026-08-12T09:00:00Z",
		"participation_mode":   "hybrid",
		"registration_enabled": true,
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	eventID := decodeBody[map[string]any](t, createRec)["item"].(map[string]any)["id"].(string)

	publishReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/publish", nil)
	publishReq.AddCookie(sessionCookie)
	publishRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(publishRec, publishReq)
	if publishRec.Code != http.StatusOK {
		t.Fatalf("expected publish status 200, got %d", publishRec.Code)
	}

	embedReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID+"/embed-code", nil)
	embedReq.AddCookie(sessionCookie)
	embedRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(embedRec, embedReq)
	if embedRec.Code != http.StatusOK {
		t.Fatalf("expected embed status 200, got %d", embedRec.Code)
	}
	payload := decodeBody[map[string]any](t, embedRec)
	if payload["kind"] != "registration_form" {
		t.Fatalf("expected kind registration_form, got %v", payload["kind"])
	}
	embedCode, _ := payload["embed_code"].(string)
	if !strings.Contains(embedCode, "/"+tenantSlug+"/register.js?event=embed-event") {
		t.Fatalf("expected embed code to contain register.js URL, got %q", embedCode)
	}
	detailURL, _ := payload["event_detail_api_url"].(string)
	if !strings.Contains(detailURL, "/api/v1/public/"+tenantSlug+"/events/embed-event") {
		t.Fatalf("expected event_detail_api_url for public event detail, got %q", detailURL)
	}
	warnings, ok := payload["warnings"].([]any)
	if !ok {
		t.Fatalf("expected warnings array")
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for published public event, got %v", warnings)
	}
}

func TestAdminEventRequiresAuth(t *testing.T) {
	app, _, _ := setupAuthApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAdminEventDuplicateSlug(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	create := func(t *testing.T, slug string, startsAt string) *httptest.ResponseRecorder {
		t.Helper()
		payload := map[string]any{
			"slug":      slug,
			"title":     slug,
			"starts_at": startsAt,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)
		rec := httptest.NewRecorder()
		app.Handler().ServeHTTP(rec, req)
		return rec
	}

	first := create(t, "gleicher-event", "2026-09-01T10:00:00Z")
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first create 201, got %d", first.Code)
	}
	second := create(t, "gleicher-event", "2026-09-02T10:00:00Z")
	if second.Code != http.StatusConflict {
		t.Fatalf("expected second create 409, got %d", second.Code)
	}
}

func TestAdminEventReadonlyCannotWrite(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)

	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE tenant_users SET role = 'readonly' WHERE email = ?`,
		"owner@example.com",
	); err != nil {
		t.Fatalf("set readonly role: %v", err)
	}

	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"slug":      "readonly-event",
		"title":     "Readonly Event",
		"starts_at": "2026-09-01T10:00:00Z",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusForbidden {
		t.Fatalf("expected readonly create status 403, got %d", createRec.Code)
	}
}

func TestAdminEventPublishRejectsCancelled(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"slug":      "cancelled-event",
		"title":     "Cancelled Event",
		"starts_at": "2026-09-01T10:00:00Z",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	createResult := decodeBody[map[string]any](t, createRec)
	item := createResult["item"].(map[string]any)
	eventID := item["id"].(string)

	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE events SET status = ?, updated_at = ? WHERE id = ?`,
		"cancelled",
		time.Now().UTC().Format(time.RFC3339),
		eventID,
	); err != nil {
		t.Fatalf("set cancelled status: %v", err)
	}

	publishReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/publish", nil)
	publishReq.AddCookie(sessionCookie)
	publishRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(publishRec, publishReq)
	if publishRec.Code != http.StatusConflict {
		t.Fatalf("expected publish conflict status 409, got %d", publishRec.Code)
	}
}

func TestAdminEventMaintenanceActions(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"slug":      "maintenance-http-event",
		"title":     "Maintenance HTTP Event",
		"starts_at": "2026-09-01T10:00:00Z",
		"ends_at":   "2026-09-01T12:00:00Z",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	createResult := decodeBody[map[string]any](t, createRec)
	eventID := createResult["item"].(map[string]any)["id"].(string)

	publishReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/publish", nil)
	publishReq.AddCookie(sessionCookie)
	publishRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(publishRec, publishReq)
	if publishRec.Code != http.StatusOK {
		t.Fatalf("expected publish status 200, got %d", publishRec.Code)
	}

	newTitle := map[string]any{"title": "Maintenance HTTP Event Updated"}
	newTitleBody, _ := json.Marshal(newTitle)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/events/"+eventID, bytes.NewReader(newTitleBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(sessionCookie)
	patchRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected patch status 200, got %d", patchRec.Code)
	}
	patchItem := decodeBody[map[string]any](t, patchRec)["item"].(map[string]any)
	if patchItem["status"] != "changed" {
		t.Fatalf("expected status changed after patch, got %v", patchItem["status"])
	}

	postponePayload := map[string]any{
		"starts_at":   "2026-10-01T10:30:00Z",
		"ends_at":     "2026-10-01T12:30:00Z",
		"change_note": "Termin wurde verschoben.",
	}
	postponeBody, _ := json.Marshal(postponePayload)
	postponeReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/postpone", bytes.NewReader(postponeBody))
	postponeReq.Header.Set("Content-Type", "application/json")
	postponeReq.AddCookie(sessionCookie)
	postponeRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(postponeRec, postponeReq)
	if postponeRec.Code != http.StatusOK {
		t.Fatalf("expected postpone status 200, got %d", postponeRec.Code)
	}
	postponeItem := decodeBody[map[string]any](t, postponeRec)["item"].(map[string]any)
	if postponeItem["status"] != "postponed" {
		t.Fatalf("expected status postponed, got %v", postponeItem["status"])
	}
	if postponeItem["change_note"] == "" {
		t.Fatalf("expected postpone change_note")
	}

	completedReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/mark-completed", nil)
	completedReq.AddCookie(sessionCookie)
	completedRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(completedRec, completedReq)
	if completedRec.Code != http.StatusOK {
		t.Fatalf("expected mark-completed status 200, got %d", completedRec.Code)
	}
	completedItem := decodeBody[map[string]any](t, completedRec)["item"].(map[string]any)
	if completedItem["status"] != "completed" {
		t.Fatalf("expected status completed, got %v", completedItem["status"])
	}
	if completedItem["registration_enabled"] != false {
		t.Fatalf("expected registration_enabled=false, got %v", completedItem["registration_enabled"])
	}
	if completedItem["waitlist_enabled"] != false {
		t.Fatalf("expected waitlist_enabled=false, got %v", completedItem["waitlist_enabled"])
	}

	cancelPayload := map[string]any{
		"cancelled_reason": "Nachtraegliche Absage.",
		"change_note":      "Bitte Kalender aktualisieren.",
	}
	cancelBody, _ := json.Marshal(cancelPayload)
	cancelReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/cancel", bytes.NewReader(cancelBody))
	cancelReq.Header.Set("Content-Type", "application/json")
	cancelReq.AddCookie(sessionCookie)
	cancelRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusConflict {
		t.Fatalf("expected cancel-after-complete status 409, got %d", cancelRec.Code)
	}
}

func TestAdminEventSupportsSeriesDetailsAndClearingOptionalFields(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	seriesPayload := map[string]any{
		"slug":                  "retreat-reihe",
		"title":                 "Retreat Reihe",
		"description":           "Mehrtaegige Workshop-Reihe.",
		"default_location_name": "Seminarhaus",
		"default_address":       "Hauptstrasse 1",
		"default_online_url":    "https://meet.example.com/retreat",
		"is_public":             true,
	}
	seriesBody, _ := json.Marshal(seriesPayload)
	seriesReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/event-series", bytes.NewReader(seriesBody))
	seriesReq.Header.Set("Content-Type", "application/json")
	seriesReq.AddCookie(sessionCookie)
	seriesRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(seriesRec, seriesReq)
	if seriesRec.Code != http.StatusCreated {
		t.Fatalf("expected series create status 201, got %d", seriesRec.Code)
	}
	seriesID := decodeBody[map[string]any](t, seriesRec)["item"].(map[string]any)["id"].(string)

	createPayload := map[string]any{
		"series_id":            seriesID,
		"slug":                 "retreat-juli",
		"title":                "Retreat Juli",
		"subtitle":             "Block A",
		"description":          "Mit Uebernachtung und Tagesprogramm.",
		"starts_at":            "2026-07-20T08:00:00Z",
		"ends_at":              "2026-07-24T16:00:00Z",
		"timezone":             "Europe/Berlin",
		"location_name":        "Seminarhaus",
		"address":              "Hauptstrasse 1",
		"online_url":           "https://meet.example.com/retreat-juli",
		"participation_mode":   "hybrid",
		"max_participants":     24,
		"is_public":            true,
		"registration_enabled": true,
		"waitlist_enabled":     true,
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	created := decodeBody[map[string]any](t, createRec)["item"].(map[string]any)
	eventID := created["id"].(string)
	if created["series_id"] != seriesID {
		t.Fatalf("expected series assignment %q, got %v", seriesID, created["series_id"])
	}
	if created["description"] != "Mit Uebernachtung und Tagesprogramm." {
		t.Fatalf("expected description in create response, got %v", created["description"])
	}

	patchPayload := map[string]any{
		"series_id":              "",
		"subtitle":               "Block A – aktualisiert",
		"description":            "Aktualisierte Beschreibung",
		"ends_at":                "",
		"max_participants":       nil,
		"clear_max_participants": true,
		"change_note":            "Termin wurde inhaltlich aktualisiert.",
	}
	patchBody, _ := json.Marshal(patchPayload)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/events/"+eventID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(sessionCookie)
	patchRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected patch status 200, got %d", patchRec.Code)
	}
	updated := decodeBody[map[string]any](t, patchRec)["item"].(map[string]any)
	if updated["series_id"] != nil {
		t.Fatalf("expected cleared series_id, got %v", updated["series_id"])
	}
	if updated["subtitle"] != "Block A – aktualisiert" {
		t.Fatalf("expected updated subtitle, got %v", updated["subtitle"])
	}
	if updated["description"] != "Aktualisierte Beschreibung" {
		t.Fatalf("expected updated description, got %v", updated["description"])
	}
	if updated["ends_at"] != nil {
		t.Fatalf("expected cleared ends_at, got %v", updated["ends_at"])
	}
	if updated["max_participants"] != nil {
		t.Fatalf("expected cleared max_participants, got %v", updated["max_participants"])
	}
}

func TestAdminEventCancelEndpoint(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"slug":      "cancel-http-event",
		"title":     "Cancel HTTP Event",
		"starts_at": "2026-09-01T10:00:00Z",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	createResult := decodeBody[map[string]any](t, createRec)
	eventID := createResult["item"].(map[string]any)["id"].(string)

	publishReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/publish", nil)
	publishReq.AddCookie(sessionCookie)
	publishRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(publishRec, publishReq)
	if publishRec.Code != http.StatusOK {
		t.Fatalf("expected publish status 200, got %d", publishRec.Code)
	}

	cancelPayload := map[string]any{
		"cancelled_reason": "Location ist nicht verfuegbar.",
		"change_note":      "Wir informieren ueber einen Ersatztermin.",
	}
	cancelBody, _ := json.Marshal(cancelPayload)
	cancelReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/cancel", bytes.NewReader(cancelBody))
	cancelReq.Header.Set("Content-Type", "application/json")
	cancelReq.AddCookie(sessionCookie)
	cancelRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusOK {
		t.Fatalf("expected cancel status 200, got %d", cancelRec.Code)
	}
	cancelledItem := decodeBody[map[string]any](t, cancelRec)["item"].(map[string]any)
	if cancelledItem["status"] != "cancelled" {
		t.Fatalf("expected status cancelled, got %v", cancelledItem["status"])
	}
	if cancelledItem["cancelled_reason"] == "" {
		t.Fatalf("expected cancelled_reason to be set")
	}
	if cancelledItem["registration_enabled"] != false {
		t.Fatalf("expected registration_enabled=false, got %v", cancelledItem["registration_enabled"])
	}
	if cancelledItem["waitlist_enabled"] != false {
		t.Fatalf("expected waitlist_enabled=false, got %v", cancelledItem["waitlist_enabled"])
	}

	postponePayload := map[string]any{
		"starts_at":   "2026-10-01T10:00:00Z",
		"change_note": "Nachholtermin",
	}
	postponeBody, _ := json.Marshal(postponePayload)
	postponeReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/postpone", bytes.NewReader(postponeBody))
	postponeReq.Header.Set("Content-Type", "application/json")
	postponeReq.AddCookie(sessionCookie)
	postponeRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(postponeRec, postponeReq)
	if postponeRec.Code != http.StatusConflict {
		t.Fatalf("expected postpone-after-cancel status 409, got %d", postponeRec.Code)
	}
}
