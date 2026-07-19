package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func TestSuperadminTenantsAndUsers(t *testing.T) {
	app, _, _ := setupAuthApp(t)
	app.cfg.SuperadminToken = "super-secret"

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/internal/superadmin/tenants", nil)
	listReq.Header.Set("Authorization", "Bearer super-secret")
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	if listPayload["total"] != float64(1) {
		t.Fatalf("expected total 1, got %v", listPayload["total"])
	}

	createBody, _ := json.Marshal(map[string]any{
		"slug":            "kunden-neu",
		"name":            "Kunden Neu",
		"public_base_url": "https://events.example.com/kunden-neu",
		"settings": map[string]any{
			"sender_email":           "events@example.com",
			"sender_name":            "Events Team",
			"default_retention_days": 45,
			"app_settings": map[string]any{
				"customer_status":         "trial",
				"enabled_features":        []string{"custom_domains", "payments", "series"},
				"event_time_start":        "09:00",
				"event_time_end":          "21:00",
				"event_time_step_minutes": 30,
			},
		},
		"owner": map[string]any{
			"email":  "owner2@example.com",
			"name":   "Owner Two",
			"role":   "owner",
			"status": "active",
		},
	})
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/internal/superadmin/tenants", bytes.NewReader(createBody))
	createReq.Header.Set("Authorization", "Bearer super-secret")
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRec.Code)
	}
	created := decodeBody[map[string]any](t, createRec)["item"].(map[string]any)
	tenantID := created["id"].(string)
	if created["slug"] != "kunden-neu" {
		t.Fatalf("expected slug kunden-neu, got %v", created["slug"])
	}

	usersReq := httptest.NewRequest(http.MethodGet, "/api/v1/internal/superadmin/tenants/"+tenantID+"/users", nil)
	usersReq.Header.Set("Authorization", "Bearer super-secret")
	usersRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(usersRec, usersReq)
	if usersRec.Code != http.StatusOK {
		t.Fatalf("expected users status 200, got %d", usersRec.Code)
	}
	usersPayload := decodeBody[map[string]any](t, usersRec)
	if usersPayload["total"] != float64(1) {
		t.Fatalf("expected one owner user, got %v", usersPayload["total"])
	}

	settings, err := app.tenantRepo.GetSettings(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("get tenant settings: %v", err)
	}
	if !tenant.FeatureEnabledInSettings(settings.SettingsJSON, tenant.FeaturePayments) {
		t.Fatalf("expected payments feature enabled in settings")
	}
}

func TestSuperadminRequiresToken(t *testing.T) {
	app, _, _ := setupAuthApp(t)
	app.cfg.SuperadminToken = "super-secret"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal/superadmin/me", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status 401, got %d", rec.Code)
	}
}
