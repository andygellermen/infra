package httpapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func TestAdminCustomDomainsFeatureGate(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := authenticateAdminSession(t, app, sender, "localhost:8080")
	tenantID := tenantIDBySlug(t, app, tenantSlug)

	settings := disableTenantFeaturesForTest(t, app, tenantID, tenant.FeatureCalendar, tenant.FeaturePayments, tenant.FeatureSeries, tenant.FeatureSnippets)

	if tenant.FeatureEnabledInSettings(settings.SettingsJSON, tenant.FeatureCustomDomains) {
		t.Fatalf("expected custom_domains feature disabled")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenant/domains", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden status 403, got %d", rec.Code)
	}
}

func TestPublicSnippetsFeatureGate(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	disableTenantFeaturesForTest(t, app, tenantID, tenant.FeatureCalendar, tenant.FeatureCustomDomains, tenant.FeaturePayments, tenant.FeatureSeries)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/snippet/events", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden status 403, got %d", rec.Code)
	}
}

func disableTenantFeaturesForTest(t *testing.T, app *App, tenantID string, enabledFeatures ...string) tenant.TenantSettings {
	t.Helper()

	current, err := app.tenantRepo.GetSettings(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	appSettings, existing, err := parseAdminTenantAppSettings(current.SettingsJSON)
	if err != nil {
		t.Fatalf("parse app settings: %v", err)
	}
	appSettings.EnabledFeatures = enabledFeatures
	settingsJSON, err := marshalAdminTenantAppSettings(existing, appSettings)
	if err != nil {
		t.Fatalf("marshal app settings: %v", err)
	}

	updated, err := app.tenantRepo.UpsertSettings(context.Background(), tenant.UpsertTenantSettingsParams{
		TenantID: tenantID,
		Settings: tenant.TenantSettingsInput{
			SenderEmail:          current.SenderEmail,
			SenderName:           current.SenderName,
			PayPalMode:           current.PayPalMode,
			PayPalClientID:       current.PayPalClientID,
			PayPalMerchantID:     current.PayPalMerchantID,
			DefaultRetentionDays: current.DefaultRetentionDays,
			SettingsJSON:         settingsJSON,
		},
	})
	if err != nil {
		t.Fatalf("upsert settings: %v", err)
	}
	return updated
}
