package httpapp

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handleAdminTenantDomainsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleAdminTenantDomainsList(w, r)
	case http.MethodPost:
		a.handleAdminTenantDomainsCreate(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminTenantDomainsItem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPatch:
		a.handleAdminTenantDomainsUpdate(w, r)
	case http.MethodDelete:
		a.handleAdminTenantDomainsDelete(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminTenantDomainsList(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	items, err := a.tenantRepo.ListDomainBindings(r.Context(), principal.TenantID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	targetHost := tenantDomainDNSTargetHost(a.cfg.BaseURL)
	payloadItems := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payloadItems = append(payloadItems, tenantDomainBindingPayload(item, targetHost))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":           payloadItems,
		"dns_target_host": targetHost,
		"dns_record_type": "CNAME",
		"dns_hint":        tenantDomainDNSHint(targetHost),
	})
}

func (a *App) handleAdminTenantDomainsCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	var request struct {
		Domain    string `json:"domain"`
		BasePath  string `json:"base_path"`
		Status    string `json:"status"`
		IsPrimary bool   `json:"is_primary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	item, err := a.tenantRepo.CreateDomainBinding(r.Context(), tenant.CreateTenantDomainBindingParams{
		TenantID:  principal.TenantID,
		Domain:    request.Domain,
		BasePath:  request.BasePath,
		Status:    request.Status,
		IsPrimary: request.IsPrimary,
	})
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	targetHost := tenantDomainDNSTargetHost(a.cfg.BaseURL)
	writeJSON(w, http.StatusCreated, map[string]any{
		"item": tenantDomainBindingPayload(item, targetHost),
	})
}

func (a *App) handleAdminTenantDomainsUpdate(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	bindingID := tenantDomainBindingIDFromPath(r.URL.Path)
	if bindingID == "" {
		writeAPIError(w, http.StatusNotFound, "TENANT_DOMAIN_BINDING_NOT_FOUND", "Domain-Binding nicht gefunden.")
		return
	}

	var request struct {
		Domain    *string `json:"domain"`
		BasePath  *string `json:"base_path"`
		Status    *string `json:"status"`
		IsPrimary *bool   `json:"is_primary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	item, err := a.tenantRepo.UpdateDomainBinding(r.Context(), principal.TenantID, bindingID, tenant.UpdateTenantDomainBindingParams{
		Domain:    request.Domain,
		BasePath:  request.BasePath,
		Status:    request.Status,
		IsPrimary: request.IsPrimary,
	})
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	targetHost := tenantDomainDNSTargetHost(a.cfg.BaseURL)
	writeJSON(w, http.StatusOK, map[string]any{
		"item": tenantDomainBindingPayload(item, targetHost),
	})
}

func (a *App) handleAdminTenantDomainsDelete(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	bindingID := tenantDomainBindingIDFromPath(r.URL.Path)
	if bindingID == "" {
		writeAPIError(w, http.StatusNotFound, "TENANT_DOMAIN_BINDING_NOT_FOUND", "Domain-Binding nicht gefunden.")
		return
	}
	if err := a.tenantRepo.DeleteDomainBinding(r.Context(), principal.TenantID, bindingID); err != nil {
		a.writeTenantError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func tenantDomainBindingIDFromPath(path string) string {
	const prefix = "/api/v1/admin/tenant/domains/"
	if !strings.HasPrefix(strings.TrimSpace(path), prefix) {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(strings.TrimSpace(path), prefix), "/")
}

func tenantDomainBindingPayload(item tenant.TenantDomainBinding, dnsTargetHost string) map[string]any {
	return map[string]any{
		"id":              item.ID,
		"tenant_id":       item.TenantID,
		"domain":          item.Domain,
		"base_path":       item.BasePath,
		"status":          item.Status,
		"is_primary":      item.IsPrimary,
		"public_base_url": buildTenantDomainPublicBaseURL(item.Domain, item.BasePath),
		"dns_target_host": dnsTargetHost,
		"dns_record_type": "CNAME",
		"created_at":      item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func buildTenantDomainPublicBaseURL(domain, basePath string) string {
	domain = strings.TrimSpace(domain)
	basePath = normalizePublicPath(basePath)
	if domain == "" {
		return ""
	}
	if basePath == "/" {
		return "https://" + domain
	}
	return "https://" + domain + basePath
}

func tenantDomainDNSTargetHost(rawBaseURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawBaseURL))
	if err != nil {
		return "events.example.com"
	}
	host := strings.TrimSpace(parsed.Host)
	if host == "" {
		return "events.example.com"
	}
	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		host = splitHost
	}
	host = strings.Trim(host, "[]")
	if host == "" {
		return "events.example.com"
	}
	return host
}

func tenantDomainDNSHint(targetHost string) string {
	return fmt.Sprintf("Empfohlen ist eine eigene Subdomain wie events.deinedomain.tld. Lege dafuer beim DNS-Anbieter einen CNAME auf %s an. Falls dein Provider auf Root-Domains keinen CNAME erlaubt, nutze ALIAS oder ANAME auf dasselbe Ziel.", targetHost)
}
