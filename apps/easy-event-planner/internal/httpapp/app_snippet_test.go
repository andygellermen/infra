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

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/snippet"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func TestAdminSnippetCRUDAndEmbedFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	createPayload := map[string]any{
		"name":      "Footer Upcoming",
		"slug":      "footer-upcoming",
		"view_type": "cards",
		"event_filter": map[string]any{
			"events": "upcoming",
			"limit":  6,
		},
		"display_options": map[string]any{
			"theme":    "light",
			"register": true,
		},
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/snippets", bytes.NewReader(createBody))
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
		t.Fatalf("expected create item payload")
	}
	snippetID, _ := item["id"].(string)
	if snippetID == "" {
		t.Fatalf("expected snippet id")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/snippets", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}
	listResult := decodeBody[map[string]any](t, listRec)
	if listResult["total"] != float64(1) {
		t.Fatalf("expected list total=1, got %v", listResult["total"])
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/snippets/"+snippetID, nil)
	getReq.AddCookie(sessionCookie)
	getRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected get status 200, got %d", getRec.Code)
	}

	patchPayload := map[string]any{
		"view_type": "list",
		"is_active": false,
	}
	patchBody, _ := json.Marshal(patchPayload)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/snippets/"+snippetID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(sessionCookie)
	patchRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected patch status 200, got %d", patchRec.Code)
	}
	patchResult := decodeBody[map[string]any](t, patchRec)
	patchedItem := patchResult["item"].(map[string]any)
	if patchedItem["view_type"] != "list" {
		t.Fatalf("expected patched view_type list, got %v", patchedItem["view_type"])
	}
	if patchedItem["is_active"] != false {
		t.Fatalf("expected patched is_active false, got %v", patchedItem["is_active"])
	}

	embedReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/snippets/"+snippetID+"/embed-code", nil)
	embedReq.AddCookie(sessionCookie)
	embedRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(embedRec, embedReq)
	if embedRec.Code != http.StatusOK {
		t.Fatalf("expected embed status 200, got %d", embedRec.Code)
	}
	embedResult := decodeBody[map[string]any](t, embedRec)
	embedCode, _ := embedResult["embed_code"].(string)
	if !strings.Contains(embedCode, "/"+tenantSlug+"/include.js?config=footer-upcoming") {
		t.Fatalf("expected embed code to contain tenant include.js URL, got %q", embedCode)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/snippets/"+snippetID, nil)
	deleteReq.AddCookie(sessionCookie)
	deleteRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d", deleteRec.Code)
	}
}

func TestAdminSnippetReadonlyCannotWrite(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE tenant_users SET role = 'readonly' WHERE email = ?`,
		"owner@example.com",
	); err != nil {
		t.Fatalf("set readonly role: %v", err)
	}
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/snippets", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected readonly list status 200, got %d", listRec.Code)
	}

	createPayload := map[string]any{
		"name": "Readonly Snippet",
		"slug": "readonly-snippet",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/snippets", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusForbidden {
		t.Fatalf("expected readonly create status 403, got %d", createRec.Code)
	}
}

func TestPublicSnippetEndpoints(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	if _, err := app.tenantRepo.UpsertSettings(context.Background(), tenant.UpsertTenantSettingsParams{
		TenantID: tenantID,
		Settings: tenant.TenantSettingsInput{
			SettingsJSON: `{"event_detail_base_url":"https://www.example.com/events"}`,
		},
	}); err != nil {
		t.Fatalf("upsert tenant settings: %v", err)
	}

	createPublishedEventForPublicTest(t, app, tenantID, event.CreateEventParams{
		Slug:     "snippet-future",
		Title:    "Snippet Future",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})
	createPublishedEventForPublicTest(t, app, tenantID, event.CreateEventParams{
		Slug:     "snippet-past",
		Title:    "Snippet Past",
		StartsAt: time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339),
	})

	createdConfig, err := app.snippetRepo.CreateConfig(context.Background(), tenantID, snippet.CreateConfigParams{
		Name:            "Public Snippet",
		Slug:            "public-snippet",
		ViewType:        "list",
		EventFilterJSON: `{"events":"upcoming","limit":1}`,
	})
	if err != nil {
		t.Fatalf("create snippet config: %v", err)
	}
	if createdConfig.ID == "" {
		t.Fatalf("expected snippet config id")
	}

	includeReq := httptest.NewRequest(http.MethodGet, "/"+tenantSlug+"/include.js?config=public-snippet", nil)
	includeRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(includeRec, includeReq)
	if includeRec.Code != http.StatusOK {
		t.Fatalf("expected include.js status 200, got %d", includeRec.Code)
	}
	if !strings.Contains(includeRec.Header().Get("Content-Type"), "application/javascript") {
		t.Fatalf("expected javascript content type, got %q", includeRec.Header().Get("Content-Type"))
	}
	includeBody := includeRec.Body.String()
	if !strings.Contains(includeBody, "/api/v1/public/") || !strings.Contains(includeBody, "/snippet/events") {
		t.Fatalf("expected include.js payload to reference snippet events endpoint")
	}
	if !strings.Contains(includeBody, "\""+tenantSlug+"\"") {
		t.Fatalf("expected include.js payload to include tenant slug %q", tenantSlug)
	}

	registerReq := httptest.NewRequest(http.MethodGet, "/"+tenantSlug+"/register.js?event=snippet-future", nil)
	registerRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(registerRec, registerReq)
	if registerRec.Code != http.StatusOK {
		t.Fatalf("expected register.js status 200, got %d", registerRec.Code)
	}
	if !strings.Contains(registerRec.Header().Get("Content-Type"), "application/javascript") {
		t.Fatalf("expected register.js javascript content type, got %q", registerRec.Header().Get("Content-Type"))
	}
	registerBody := registerRec.Body.String()
	if !strings.Contains(registerBody, "/api/v1/public/") || !strings.Contains(registerBody, "/registrations/start") {
		t.Fatalf("expected register.js payload to reference public registration endpoints")
	}
	if !strings.Contains(registerBody, "Magic Link anfordern") {
		t.Fatalf("expected register.js payload to contain form submit label")
	}

	cssReq := httptest.NewRequest(http.MethodGet, "/"+tenantSlug+"/snippet.css", nil)
	cssRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(cssRec, cssReq)
	if cssRec.Code != http.StatusOK {
		t.Fatalf("expected snippet.css status 200, got %d", cssRec.Code)
	}
	if !strings.Contains(cssRec.Header().Get("Content-Type"), "text/css") {
		t.Fatalf("expected css content type, got %q", cssRec.Header().Get("Content-Type"))
	}
	if !strings.Contains(cssRec.Body.String(), ".eep-cards") {
		t.Fatalf("expected snippet.css to contain eep card styles")
	}

	eventsReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/snippet/events?config=public-snippet", nil)
	eventsReq.Header.Set("Origin", "https://erweckedeinekraft.de")
	eventsRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(eventsRec, eventsReq)
	if eventsRec.Code != http.StatusOK {
		t.Fatalf("expected snippet events status 200, got %d", eventsRec.Code)
	}
	if eventsRec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected wildcard allow origin, got %q", eventsRec.Header().Get("Access-Control-Allow-Origin"))
	}
	eventsPayload := decodeBody[map[string]any](t, eventsRec)
	if eventsPayload["total"] != float64(1) {
		t.Fatalf("expected snippet total=1 from config limit, got %v", eventsPayload["total"])
	}
	if eventsPayload["view"] != "list" {
		t.Fatalf("expected snippet view list, got %v", eventsPayload["view"])
	}
	items, ok := eventsPayload["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected one public snippet item")
	}
	firstItem, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("expected public snippet item payload")
	}
	if firstItem["event_url"] != "https://www.example.com/events/snippet-future" {
		t.Fatalf("expected event_url to use external detail base url, got %v", firstItem["event_url"])
	}

	tamperedReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/snippet/events?config=public-snippet&limit=50", nil)
	tamperedRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(tamperedRec, tamperedReq)
	if tamperedRec.Code != http.StatusBadRequest {
		t.Fatalf("expected tampered config request status 400, got %d", tamperedRec.Code)
	}
	tamperedPayload := decodeBody[map[string]any](t, tamperedRec)
	tamperedErr := tamperedPayload["error"].(map[string]any)
	if tamperedErr["code"] != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR for tampered config request, got %v", tamperedErr["code"])
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/snippet/events?config=missing", nil)
	missingRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("expected missing config status 404, got %d", missingRec.Code)
	}
	missingPayload := decodeBody[map[string]any](t, missingRec)
	errorPayload := missingPayload["error"].(map[string]any)
	if errorPayload["code"] != "SNIPPET_NOT_FOUND" {
		t.Fatalf("expected SNIPPET_NOT_FOUND, got %v", errorPayload["code"])
	}
}
