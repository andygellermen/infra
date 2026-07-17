package httpapp

import (
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handleInfraDomainBindingsExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if !a.requireInfraSyncToken(w, r) {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	items, err := a.tenantRepo.ListAllDomainBindings(r.Context())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Domain-Bindings konnten nicht geladen werden.")
		return
	}

	targetHost := tenantDomainDNSTargetHost(a.cfg.BaseURL)
	payloadItems := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payload := tenantDomainBindingPayload(item, targetHost)
		payload["edge_enabled"] = tenantDomainBindingEdgeEnabled(item.Status)
		payloadItems = append(payloadItems, payload)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"generated_at":    time.Now().UTC().Format(time.RFC3339),
		"dns_target_host": targetHost,
		"items":           payloadItems,
	})
}

func (a *App) handleInfraDomainBindingsRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if !a.requireInfraSyncToken(w, r) {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	items, err := a.tenantRepo.ListAllDomainBindings(r.Context())
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Domain-Bindings konnten nicht geladen werden.")
		return
	}

	targetHost := tenantDomainDNSTargetHost(a.cfg.BaseURL)
	updatedItems := make([]map[string]any, 0, len(items))
	errorItems := make([]map[string]any, 0)
	checkedCount := 0
	for _, item := range items {
		if strings.TrimSpace(item.Status) == tenant.DomainBindingStatusDisabled {
			continue
		}
		checkedCount++

		result, refreshErr := refreshTenantDomainChecks(r.Context(), item, targetHost)
		if refreshErr != nil {
			errorItems = append(errorItems, map[string]any{
				"id":     item.ID,
				"domain": item.Domain,
				"error":  refreshErr.Error(),
			})
			continue
		}

		updated, applyErr := a.tenantRepo.ApplyDomainCheckResult(r.Context(), item.TenantID, item.ID, result)
		if applyErr != nil {
			errorItems = append(errorItems, map[string]any{
				"id":     item.ID,
				"domain": item.Domain,
				"error":  applyErr.Error(),
			})
			continue
		}

		payload := tenantDomainBindingPayload(updated, targetHost)
		payload["edge_enabled"] = tenantDomainBindingEdgeEnabled(updated.Status)
		updatedItems = append(updatedItems, payload)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"generated_at":  time.Now().UTC().Format(time.RFC3339),
		"checked_count": checkedCount,
		"updated_count": len(updatedItems),
		"items":         updatedItems,
		"errors":        errorItems,
	})
}

func (a *App) requireInfraSyncToken(w http.ResponseWriter, r *http.Request) bool {
	expected := strings.TrimSpace(a.cfg.InfraSyncToken)
	if expected == "" {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Route nicht gefunden.")
		return false
	}

	provided := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(r.Header.Get("Authorization")), "Bearer "))
	if provided == "" || provided != expected {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Zugriff verweigert.")
		return false
	}
	return true
}

func tenantDomainBindingEdgeEnabled(status string) bool {
	switch strings.TrimSpace(status) {
	case tenant.DomainBindingStatusDNSVerified, tenant.DomainBindingStatusSSLPending, tenant.DomainBindingStatusActive:
		return true
	default:
		return false
	}
}
