package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
