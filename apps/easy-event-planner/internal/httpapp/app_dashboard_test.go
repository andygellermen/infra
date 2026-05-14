package httpapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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
