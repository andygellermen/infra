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
)

func TestAdminDashboardAndRegistrationEndpoints(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")
	_, eventID, _, firstRegistrationID := createWaitlistScenario(t, app, tenantSlug)

	dashboardReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)
	dashboardReq.AddCookie(sessionCookie)
	dashboardRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(dashboardRec, dashboardReq)
	if dashboardRec.Code != http.StatusOK {
		t.Fatalf("expected dashboard status 200, got %d", dashboardRec.Code)
	}
	dashboardPayload := decodeBody[map[string]any](t, dashboardRec)
	stats, ok := dashboardPayload["stats"].(map[string]any)
	if !ok {
		t.Fatalf("expected stats payload")
	}
	if _, ok := stats["confirmed_participants"]; !ok {
		t.Fatalf("expected confirmed_participants in stats")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID+"/registrations", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected registration list status 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	if listPayload["total"] != float64(2) {
		t.Fatalf("expected registration total=2, got %v", listPayload["total"])
	}
	items, ok := listPayload["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected two registration items")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/registrations/"+firstRegistrationID, nil)
	getReq.AddCookie(sessionCookie)
	getRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected registration get status 200, got %d", getRec.Code)
	}
	getPayload := decodeBody[map[string]any](t, getRec)
	item, ok := getPayload["item"].(map[string]any)
	if !ok {
		t.Fatalf("expected item payload in registration get")
	}
	if item["id"] != firstRegistrationID {
		t.Fatalf("expected registration id %q, got %v", firstRegistrationID, item["id"])
	}
	if item["status"] != "confirmed" {
		t.Fatalf("expected confirmed registration status, got %v", item["status"])
	}
}

func TestAdminDashboardAndRegistrationsRequireAuth(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	_, eventID, _, firstRegistrationID := createWaitlistScenario(t, app, tenantSlug)

	dashboardReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)
	dashboardRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(dashboardRec, dashboardReq)
	if dashboardRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected dashboard unauthorized status 401, got %d", dashboardRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID+"/registrations", nil)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected list unauthorized status 401, got %d", listRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/registrations/"+firstRegistrationID, nil)
	getRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected get unauthorized status 401, got %d", getRec.Code)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventID+"/registrations/manual", nil)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected manual create unauthorized status 401, got %d", createRec.Code)
	}
}

func TestAdminDashboardAndRegistrationsReadonlyCanRead(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE tenant_users SET role = 'readonly' WHERE email = ?`,
		"owner@example.com",
	); err != nil {
		t.Fatalf("set readonly role: %v", err)
	}
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")
	_, eventID, _, firstRegistrationID := createWaitlistScenario(t, app, tenantSlug)

	dashboardReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)
	dashboardReq.AddCookie(sessionCookie)
	dashboardRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(dashboardRec, dashboardReq)
	if dashboardRec.Code != http.StatusOK {
		t.Fatalf("expected readonly dashboard status 200, got %d", dashboardRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID+"/registrations", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected readonly list status 200, got %d", listRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/registrations/"+firstRegistrationID, nil)
	getReq.AddCookie(sessionCookie)
	getRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected readonly get status 200, got %d", getRec.Code)
	}
}

func TestAdminManualRegistrationCreateFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "admin-manual-create-flow",
		Title:    "Admin Manual Create Flow",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	requestBody, _ := json.Marshal(map[string]any{
		"name":               "Max Mustermann",
		"email":              "max@example.com",
		"phone":              "+4912345",
		"participation_type": "onsite",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventItem.ID+"/registrations/manual", bytes.NewReader(requestBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected manual create status 201, got %d", createRec.Code)
	}
	createPayload := decodeBody[map[string]any](t, createRec)
	item, ok := createPayload["item"].(map[string]any)
	if !ok {
		t.Fatalf("expected item payload")
	}
	if item["status"] != "confirmed" {
		t.Fatalf("expected confirmed status, got %v", item["status"])
	}
	if item["source"] != "admin_manual" {
		t.Fatalf("expected source admin_manual, got %v", item["source"])
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventItem.ID+"/registrations", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	if listPayload["total"] != float64(1) {
		t.Fatalf("expected registration total=1, got %v", listPayload["total"])
	}
}

func TestAdminManualRegistrationReadonlyForbidden(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "admin-manual-readonly",
		Title:    "Admin Manual Readonly",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE tenant_users SET role = 'readonly' WHERE email = ?`,
		"owner@example.com",
	); err != nil {
		t.Fatalf("set readonly role: %v", err)
	}
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	requestBody, _ := json.Marshal(map[string]any{
		"name":               "Readonly User",
		"email":              "readonly@example.com",
		"participation_type": "onsite",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/events/"+eventItem.ID+"/registrations/manual", bytes.NewReader(requestBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusForbidden {
		t.Fatalf("expected readonly manual create status 403, got %d", createRec.Code)
	}
}
