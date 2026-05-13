package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func loginSessionCookie(t *testing.T, app *App, sender *fakeMagicLinkSender, tenantSlug, email string) *http.Cookie {
	t.Helper()

	requestPayload := map[string]any{
		"tenant_slug": tenantSlug,
		"email":       email,
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
	if sender.lastMessage.VerifyURL == "" {
		t.Fatalf("expected magic link verify url")
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

func TestAdminEventSeriesCRUDFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"slug":                  "angst-workshop",
		"title":                 "Angst Workshop",
		"description":           "Ein Workshop fuer innere Stabilitaet.",
		"default_location_name": "Community Space",
		"default_address":       "Musterweg 7",
		"default_online_url":    "https://meet.example.com/angst-workshop",
		"is_public":             true,
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/event-series", bytes.NewReader(createBody))
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
		t.Fatalf("expected item payload, got %v", createResult["item"])
	}
	seriesID, ok := item["id"].(string)
	if !ok || seriesID == "" {
		t.Fatalf("expected created series id, got %v", item["id"])
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/event-series", nil)
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

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/event-series/"+seriesID, nil)
	getReq.AddCookie(sessionCookie)
	getRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected get status 200, got %d", getRec.Code)
	}

	patchPayload := map[string]any{
		"title":     "Angst Workshop Plus",
		"is_public": false,
	}
	patchBody, _ := json.Marshal(patchPayload)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/event-series/"+seriesID, bytes.NewReader(patchBody))
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
		t.Fatalf("expected item in patch response")
	}
	if updatedItem["title"] != "Angst Workshop Plus" {
		t.Fatalf("expected updated title, got %v", updatedItem["title"])
	}
	if updatedItem["is_public"] != false {
		t.Fatalf("expected is_public=false, got %v", updatedItem["is_public"])
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/event-series/"+seriesID, nil)
	deleteReq.AddCookie(sessionCookie)
	deleteRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d", deleteRec.Code)
	}

	getAfterDeleteReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/event-series/"+seriesID, nil)
	getAfterDeleteReq.AddCookie(sessionCookie)
	getAfterDeleteRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getAfterDeleteRec, getAfterDeleteReq)
	if getAfterDeleteRec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 after delete, got %d", getAfterDeleteRec.Code)
	}
}

func TestAdminEventSeriesRequiresAuth(t *testing.T) {
	app, _, _ := setupAuthApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/event-series", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAdminEventSeriesReadonlyForbiddenForWrite(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)

	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE tenant_users SET role = 'readonly' WHERE email = ?`,
		"owner@example.com",
	); err != nil {
		t.Fatalf("set role readonly: %v", err)
	}

	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	readReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/event-series", nil)
	readReq.AddCookie(sessionCookie)
	readRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(readRec, readReq)
	if readRec.Code != http.StatusOK {
		t.Fatalf("expected readonly list status 200, got %d", readRec.Code)
	}

	createPayload := map[string]any{
		"slug":  "readonly-attempt",
		"title": "Readonly Attempt",
	}
	createBody, _ := json.Marshal(createPayload)
	writeReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/event-series", bytes.NewReader(createBody))
	writeReq.Header.Set("Content-Type", "application/json")
	writeReq.AddCookie(sessionCookie)
	writeRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(writeRec, writeReq)
	if writeRec.Code != http.StatusForbidden {
		t.Fatalf("expected readonly write status 403, got %d", writeRec.Code)
	}
}

func TestAdminEventSeriesDuplicateSlug(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	create := func(t *testing.T, slug string) *httptest.ResponseRecorder {
		t.Helper()
		payload := map[string]any{
			"slug":  slug,
			"title": slug,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/event-series", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(sessionCookie)
		rec := httptest.NewRecorder()
		app.Handler().ServeHTTP(rec, req)
		return rec
	}

	first := create(t, "same-slug")
	if first.Code != http.StatusCreated {
		t.Fatalf("expected first create 201, got %d", first.Code)
	}
	second := create(t, "same-slug")
	if second.Code != http.StatusConflict {
		t.Fatalf("expected second create 409, got %d", second.Code)
	}
}
