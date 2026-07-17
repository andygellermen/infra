package httpapp

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

var (
	tenantDomainLookupTXT = func(ctx context.Context, name string) ([]string, error) {
		return net.DefaultResolver.LookupTXT(ctx, name)
	}
	tenantDomainLookupCNAME = func(ctx context.Context, host string) (string, error) {
		return net.DefaultResolver.LookupCNAME(ctx, host)
	}
	tenantDomainLookupHost = func(ctx context.Context, host string) ([]string, error) {
		return net.DefaultResolver.LookupHost(ctx, host)
	}
	tenantDomainTLSProbe = func(ctx context.Context, host string) (domainTLSProbeResult, error) {
		return probeTenantDomainTLS(ctx, host)
	}
)

type domainTLSProbeResult struct {
	Status    string
	Issuer    string
	ExpiresAt *time.Time
	ErrorText string
}

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
	bindingID, action := tenantDomainBindingRouteParts(r.URL.Path)
	if bindingID == "" {
		writeAPIError(w, http.StatusNotFound, "TENANT_DOMAIN_BINDING_NOT_FOUND", "Domain-Binding nicht gefunden.")
		return
	}

	switch {
	case action == "" && r.Method == http.MethodPatch:
		a.handleAdminTenantDomainsUpdate(w, r, bindingID)
	case action == "" && r.Method == http.MethodDelete:
		a.handleAdminTenantDomainsDelete(w, r, bindingID)
	case action == "refresh-check" && r.Method == http.MethodPost:
		a.handleAdminTenantDomainsRefreshCheck(w, r, bindingID)
	case action == "rotate-verification-token" && r.Method == http.MethodPost:
		a.handleAdminTenantDomainsRotateVerificationToken(w, r, bindingID)
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

func (a *App) handleAdminTenantDomainsUpdate(w http.ResponseWriter, r *http.Request, bindingID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
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

func (a *App) handleAdminTenantDomainsDelete(w http.ResponseWriter, r *http.Request, bindingID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	if err := a.tenantRepo.DeleteDomainBinding(r.Context(), principal.TenantID, bindingID); err != nil {
		a.writeTenantError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleAdminTenantDomainsRotateVerificationToken(w http.ResponseWriter, r *http.Request, bindingID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	item, err := a.tenantRepo.RotateDomainBindingVerificationToken(r.Context(), principal.TenantID, bindingID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	targetHost := tenantDomainDNSTargetHost(a.cfg.BaseURL)
	writeJSON(w, http.StatusOK, map[string]any{
		"item": tenantDomainBindingPayload(item, targetHost),
	})
}

func (a *App) handleAdminTenantDomainsRefreshCheck(w http.ResponseWriter, r *http.Request, bindingID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	item, err := a.tenantRepo.GetDomainBindingByID(r.Context(), principal.TenantID, bindingID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	targetHost := tenantDomainDNSTargetHost(a.cfg.BaseURL)

	result, err := refreshTenantDomainChecks(r.Context(), item, targetHost)
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, "TENANT_DOMAIN_CHECK_FAILED", err.Error())
		return
	}

	updated, err := a.tenantRepo.ApplyDomainCheckResult(r.Context(), principal.TenantID, item.ID, result)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": tenantDomainBindingPayload(updated, targetHost),
	})
}

func tenantDomainBindingRouteParts(path string) (bindingID, action string) {
	const prefix = "/api/v1/admin/tenant/domains/"
	trimmed := strings.TrimSpace(path)
	if !strings.HasPrefix(trimmed, prefix) {
		return "", ""
	}
	rest := strings.Trim(strings.TrimPrefix(trimmed, prefix), "/")
	if rest == "" {
		return "", ""
	}
	parts := strings.Split(rest, "/")
	bindingID = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		action = strings.TrimSpace(parts[1])
	}
	return bindingID, action
}

func refreshTenantDomainChecks(ctx context.Context, item tenant.TenantDomainBinding, targetHost string) (tenant.DomainCheckResult, error) {
	now := time.Now().UTC()
	result := tenant.DomainCheckResult{
		VerificationRecordName:  tenantVerificationRecordName(item.Domain),
		VerificationRecordValue: tenantVerificationRecordValue(item.VerificationToken),
		LastDNSCheckAt:          &now,
		LastRoutingCheckAt:      &now,
		LastSSLCheckAt:          &now,
		SSLStatus:               tenant.DomainBindingSSLStatusPending,
	}

	dnsVerified, dnsErr := checkTenantDomainTXT(ctx, result.VerificationRecordName, result.VerificationRecordValue)
	result.DNSVerified = dnsVerified
	result.LastDNSError = dnsErr

	routingVerified, routingErr := checkTenantDomainRouting(ctx, item.Domain, targetHost)
	result.RoutingVerified = routingVerified
	result.LastRoutingError = routingErr

	tlsResult, tlsErr := tenantDomainTLSProbe(ctx, item.Domain)
	if tlsErr != nil {
		result.SSLStatus = tenant.DomainBindingSSLStatusPending
		result.LastSSLError = tlsErr.Error()
	} else {
		result.SSLStatus = tlsResult.Status
		result.SSLCertificateIssuer = tlsResult.Issuer
		result.SSLCertificateExpiresAt = tlsResult.ExpiresAt
		result.LastSSLError = tlsResult.ErrorText
	}

	result.Status = deriveTenantDomainStatus(result)
	return result, nil
}

func tenantDomainBindingPayload(item tenant.TenantDomainBinding, dnsTargetHost string) map[string]any {
	var dnsVerifiedAt any
	if item.DNSVerifiedAt != nil {
		dnsVerifiedAt = item.DNSVerifiedAt.UTC().Format(time.RFC3339)
	}
	var routingVerifiedAt any
	if item.RoutingVerifiedAt != nil {
		routingVerifiedAt = item.RoutingVerifiedAt.UTC().Format(time.RFC3339)
	}
	var lastDNSCheckAt any
	if item.LastDNSCheckAt != nil {
		lastDNSCheckAt = item.LastDNSCheckAt.UTC().Format(time.RFC3339)
	}
	var lastRoutingCheckAt any
	if item.LastRoutingCheckAt != nil {
		lastRoutingCheckAt = item.LastRoutingCheckAt.UTC().Format(time.RFC3339)
	}
	var sslCertificateExpiresAt any
	if item.SSLCertificateExpiresAt != nil {
		sslCertificateExpiresAt = item.SSLCertificateExpiresAt.UTC().Format(time.RFC3339)
	}
	var lastSSLCheckAt any
	if item.LastSSLCheckAt != nil {
		lastSSLCheckAt = item.LastSSLCheckAt.UTC().Format(time.RFC3339)
	}

	return map[string]any{
		"id":                         item.ID,
		"tenant_id":                  item.TenantID,
		"domain":                     item.Domain,
		"base_path":                  item.BasePath,
		"status":                     item.Status,
		"is_primary":                 item.IsPrimary,
		"public_base_url":            buildTenantDomainPublicBaseURL(item.Domain, item.BasePath),
		"verification_record_name":   tenantVerificationRecordName(item.Domain),
		"verification_record_value":  tenantVerificationRecordValue(item.VerificationToken),
		"dns_verified_at":            dnsVerifiedAt,
		"routing_verified_at":        routingVerifiedAt,
		"last_dns_check_at":          lastDNSCheckAt,
		"last_dns_error":             item.LastDNSError,
		"last_routing_check_at":      lastRoutingCheckAt,
		"last_routing_error":         item.LastRoutingError,
		"ssl_status":                 item.SSLStatus,
		"ssl_certificate_issuer":     item.SSLCertificateIssuer,
		"ssl_certificate_expires_at": sslCertificateExpiresAt,
		"last_ssl_check_at":          lastSSLCheckAt,
		"last_ssl_error":             item.LastSSLError,
		"dns_target_host":            dnsTargetHost,
		"dns_record_type":            "CNAME",
		"created_at":                 item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":                 item.UpdatedAt.UTC().Format(time.RFC3339),
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

func tenantVerificationRecordName(domain string) string {
	trimmed := strings.Trim(strings.TrimSpace(domain), ".")
	if trimmed == "" {
		return ""
	}
	return "_eep-domain-verification." + trimmed
}

func tenantVerificationRecordValue(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	return "eep-domain-verification=" + trimmed
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
	return fmt.Sprintf("Empfohlen ist eine eigene Subdomain wie events.deinedomain.tld. Lege dafuer beim DNS-Anbieter einen CNAME auf %s an. Falls dein Provider auf Root-Domains keinen CNAME erlaubt, nutze ALIAS oder ANAME auf dasselbe Ziel. Zusaetzlich braucht EEP fuer die Eigentumspruefung einen TXT-Record auf _eep-domain-verification.<deine-domain>.", targetHost)
}

func checkTenantDomainTXT(ctx context.Context, recordName, expectedValue string) (bool, string) {
	values, err := tenantDomainLookupTXT(ctx, recordName)
	if err != nil {
		return false, fmt.Sprintf("TXT-Record %s konnte nicht gelesen werden: %v", recordName, err)
	}
	for _, value := range values {
		if strings.TrimSpace(value) == strings.TrimSpace(expectedValue) {
			return true, ""
		}
	}
	return false, fmt.Sprintf("TXT-Record %s enthaelt noch nicht den erwarteten Wert.", recordName)
}

func checkTenantDomainRouting(ctx context.Context, domain, targetHost string) (bool, string) {
	normalizedDomain := normalizeTenantDomainLookupHost(domain)
	normalizedTarget := normalizeTenantDomainLookupHost(targetHost)
	if normalizedDomain == "" || normalizedTarget == "" {
		return false, "Domain oder Zielhost ist ungueltig."
	}

	if cname, err := tenantDomainLookupCNAME(ctx, normalizedDomain); err == nil {
		if normalizeTenantDomainLookupHost(cname) == normalizedTarget {
			return true, ""
		}
	}

	targetIPs, targetErr := tenantDomainLookupHost(ctx, normalizedTarget)
	domainIPs, domainErr := tenantDomainLookupHost(ctx, normalizedDomain)
	if targetErr != nil || domainErr != nil {
		return false, fmt.Sprintf("Routing-Ziel konnte noch nicht bestaetigt werden. Domain zeigt noch nicht sichtbar auf %s.", normalizedTarget)
	}

	targetSet := map[string]struct{}{}
	for _, item := range targetIPs {
		targetSet[strings.TrimSpace(item)] = struct{}{}
	}
	for _, item := range domainIPs {
		if _, ok := targetSet[strings.TrimSpace(item)]; ok {
			return true, ""
		}
	}
	return false, fmt.Sprintf("Domain zeigt aktuell noch nicht auf %s.", normalizedTarget)
}

func deriveTenantDomainStatus(result tenant.DomainCheckResult) string {
	if !result.DNSVerified {
		return tenant.DomainBindingStatusPendingDNS
	}
	if !result.RoutingVerified {
		return tenant.DomainBindingStatusDNSVerified
	}
	if result.SSLStatus != tenant.DomainBindingSSLStatusValid {
		return tenant.DomainBindingStatusSSLPending
	}
	return tenant.DomainBindingStatusActive
}

func normalizeTenantDomainLookupHost(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return ""
	}
	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return ""
		}
		value = parsed.Host
	}
	value = strings.TrimSuffix(value, ".")
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	return strings.Trim(value, "[]")
}

func probeTenantDomainTLS(ctx context.Context, host string) (domainTLSProbeResult, error) {
	normalizedHost := normalizeTenantDomainLookupHost(host)
	if normalizedHost == "" {
		return domainTLSProbeResult{}, fmt.Errorf("ungueltiger Domain-Host")
	}

	timeout := 5 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(normalizedHost, "443"), &tls.Config{
		ServerName: normalizedHost,
		MinVersion: tls.VersionTLS12,
	})
	if err != nil {
		return domainTLSProbeResult{}, fmt.Errorf("TLS-Verbindung zu %s:443 fehlgeschlagen: %w", normalizedHost, err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return domainTLSProbeResult{
			Status:    tenant.DomainBindingSSLStatusPending,
			ErrorText: "Kein Zertifikat vom Zielhost erhalten.",
		}, nil
	}
	cert := state.PeerCertificates[0]
	result := domainTLSProbeResult{
		Status:    tenant.DomainBindingSSLStatusValid,
		ExpiresAt: &cert.NotAfter,
	}
	if cert.Issuer.CommonName != "" {
		result.Issuer = cert.Issuer.CommonName
	} else {
		result.Issuer = cert.Issuer.String()
	}
	if err := cert.VerifyHostname(normalizedHost); err != nil {
		result.Status = tenant.DomainBindingSSLStatusInvalid
		result.ErrorText = "Zertifikat passt noch nicht zur Domain."
		return result, nil
	}
	if cert.NotAfter.UTC().Before(time.Now().UTC()) {
		result.Status = tenant.DomainBindingSSLStatusExpired
		result.ErrorText = "Zertifikat ist abgelaufen."
		return result, nil
	}
	return result, nil
}
