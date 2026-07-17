package httpapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func TestInfraDomainBindingsExportRequiresToken(t *testing.T) {
	app, _, _ := setupAuthApp(t)
	app.cfg.InfraSyncToken = "infra-secret"

	req := httptest.NewRequest(http.MethodGet, "/api/v1/internal/infra/domain-bindings/export", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status 401, got %d", rec.Code)
	}
}

func TestInfraDomainBindingsExportAndRefresh(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	app.cfg.InfraSyncToken = "infra-secret"

	tenantItem, err := app.tenantRepo.LookupBySlug(context.Background(), tenantSlug)
	if err != nil {
		t.Fatalf("lookup tenant: %v", err)
	}

	binding, err := app.tenantRepo.CreateDomainBinding(context.Background(), tenant.CreateTenantDomainBindingParams{
		TenantID:  tenantItem.ID,
		Domain:    "events.customer-domain.example",
		BasePath:  "/",
		Status:    tenant.DomainBindingStatusDNSVerified,
		IsPrimary: false,
	})
	if err != nil {
		t.Fatalf("create domain binding: %v", err)
	}

	exportReq := httptest.NewRequest(http.MethodGet, "/api/v1/internal/infra/domain-bindings/export", nil)
	exportReq.Header.Set("Authorization", "Bearer infra-secret")
	exportRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(exportRec, exportReq)

	if exportRec.Code != http.StatusOK {
		t.Fatalf("expected export status 200, got %d", exportRec.Code)
	}
	exportPayload := decodeBody[map[string]any](t, exportRec)
	items := exportPayload["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 exported item, got %d", len(items))
	}
	exported := items[0].(map[string]any)
	if exported["domain"] != binding.Domain {
		t.Fatalf("expected exported domain %q, got %v", binding.Domain, exported["domain"])
	}
	if exported["edge_enabled"] != true {
		t.Fatalf("expected edge_enabled true, got %v", exported["edge_enabled"])
	}

	oldTXT := tenantDomainLookupTXT
	oldCNAME := tenantDomainLookupCNAME
	oldHost := tenantDomainLookupHost
	oldTLS := tenantDomainTLSProbe
	defer func() {
		tenantDomainLookupTXT = oldTXT
		tenantDomainLookupCNAME = oldCNAME
		tenantDomainLookupHost = oldHost
		tenantDomainTLSProbe = oldTLS
	}()

	tenantDomainLookupTXT = func(_ context.Context, name string) ([]string, error) {
		if name != tenantVerificationRecordName(binding.Domain) {
			t.Fatalf("unexpected txt lookup name %q", name)
		}
		return []string{tenantVerificationRecordValue(binding.VerificationToken)}, nil
	}
	tenantDomainLookupCNAME = func(_ context.Context, host string) (string, error) {
		return "events.geller.men.", nil
	}
	tenantDomainLookupHost = func(_ context.Context, host string) ([]string, error) {
		if host == binding.Domain {
			return []string{"127.0.0.1"}, nil
		}
		return []string{"127.0.0.1"}, nil
	}
	tenantDomainTLSProbe = func(_ context.Context, host string) (domainTLSProbeResult, error) {
		expiresAt := time.Now().UTC().Add(14 * 24 * time.Hour)
		return domainTLSProbeResult{
			Status:    tenant.DomainBindingSSLStatusValid,
			Issuer:    "Traefik Test LE",
			ExpiresAt: &expiresAt,
		}, nil
	}

	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/internal/infra/domain-bindings/refresh", nil)
	refreshReq.Header.Set("Authorization", "Bearer infra-secret")
	refreshRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(refreshRec, refreshReq)

	if refreshRec.Code != http.StatusOK {
		t.Fatalf("expected refresh status 200, got %d", refreshRec.Code)
	}
	refreshPayload := decodeBody[map[string]any](t, refreshRec)
	if refreshPayload["checked_count"] != float64(1) {
		t.Fatalf("expected checked_count 1, got %v", refreshPayload["checked_count"])
	}
	refreshedItems := refreshPayload["items"].([]any)
	if len(refreshedItems) != 1 {
		t.Fatalf("expected 1 refreshed item, got %d", len(refreshedItems))
	}
	refreshed := refreshedItems[0].(map[string]any)
	if refreshed["status"] != tenant.DomainBindingStatusActive {
		t.Fatalf("expected active status, got %v", refreshed["status"])
	}
	if refreshed["ssl_status"] != tenant.DomainBindingSSLStatusValid {
		t.Fatalf("expected valid ssl status, got %v", refreshed["ssl_status"])
	}
	errorsPayload := refreshPayload["errors"].([]any)
	if len(errorsPayload) != 0 {
		t.Fatalf("expected no refresh errors, got %v", errorsPayload)
	}
}
