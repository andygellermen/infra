package tenant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func (r *Repository) ListDomainBindings(ctx context.Context, tenantID string) ([]TenantDomainBinding, error) {
	id := strings.TrimSpace(tenantID)
	if id == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, domain_host, base_path, status, is_primary, verification_token,
            COALESCE(dns_verified_at, ''), COALESCE(routing_verified_at, ''), COALESCE(last_dns_check_at, ''), COALESCE(last_dns_error, ''),
            COALESCE(last_routing_check_at, ''), COALESCE(last_routing_error, ''), ssl_status, COALESCE(ssl_certificate_issuer, ''),
            COALESCE(ssl_certificate_expires_at, ''), COALESCE(last_ssl_check_at, ''), COALESCE(last_ssl_error, ''), created_at, updated_at
     FROM tenant_domain_bindings
     WHERE tenant_id = ?
     ORDER BY is_primary DESC, domain_host ASC, base_path ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("query tenant domain bindings: %w", err)
	}
	defer rows.Close()

	items := make([]TenantDomainBinding, 0)
	for rows.Next() {
		item, scanErr := scanTenantDomainBinding(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan tenant domain binding: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenant domain bindings: %w", err)
	}
	return items, nil
}

func (r *Repository) ListAllDomainBindings(ctx context.Context) ([]TenantDomainBinding, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, domain_host, base_path, status, is_primary, verification_token,
            COALESCE(dns_verified_at, ''), COALESCE(routing_verified_at, ''), COALESCE(last_dns_check_at, ''), COALESCE(last_dns_error, ''),
            COALESCE(last_routing_check_at, ''), COALESCE(last_routing_error, ''), ssl_status, COALESCE(ssl_certificate_issuer, ''),
            COALESCE(ssl_certificate_expires_at, ''), COALESCE(last_ssl_check_at, ''), COALESCE(last_ssl_error, ''), created_at, updated_at
     FROM tenant_domain_bindings
     ORDER BY domain_host ASC, base_path ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query all tenant domain bindings: %w", err)
	}
	defer rows.Close()

	items := make([]TenantDomainBinding, 0)
	for rows.Next() {
		item, scanErr := scanTenantDomainBinding(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan all tenant domain bindings: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate all tenant domain bindings: %w", err)
	}
	return items, nil
}

func (r *Repository) GetDomainBindingByID(ctx context.Context, tenantID, bindingID string) (TenantDomainBinding, error) {
	tenantID = strings.TrimSpace(tenantID)
	bindingID = strings.TrimSpace(bindingID)
	if tenantID == "" {
		return TenantDomainBinding{}, fmt.Errorf("tenant id must not be empty")
	}
	if bindingID == "" {
		return TenantDomainBinding{}, fmt.Errorf("binding id must not be empty")
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, domain_host, base_path, status, is_primary, verification_token,
            COALESCE(dns_verified_at, ''), COALESCE(routing_verified_at, ''), COALESCE(last_dns_check_at, ''), COALESCE(last_dns_error, ''),
            COALESCE(last_routing_check_at, ''), COALESCE(last_routing_error, ''), ssl_status, COALESCE(ssl_certificate_issuer, ''),
            COALESCE(ssl_certificate_expires_at, ''), COALESCE(last_ssl_check_at, ''), COALESCE(last_ssl_error, ''), created_at, updated_at
     FROM tenant_domain_bindings
     WHERE tenant_id = ? AND id = ?`,
		tenantID,
		bindingID,
	)
	item, err := scanTenantDomainBinding(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TenantDomainBinding{}, ErrTenantDomainBindingNotFound
		}
		return TenantDomainBinding{}, fmt.Errorf("query tenant domain binding: %w", err)
	}
	return item, nil
}

func (r *Repository) GetPrimaryDomainBinding(ctx context.Context, tenantID string) (TenantDomainBinding, error) {
	id := strings.TrimSpace(tenantID)
	if id == "" {
		return TenantDomainBinding{}, fmt.Errorf("tenant id must not be empty")
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, domain_host, base_path, status, is_primary, verification_token,
            COALESCE(dns_verified_at, ''), COALESCE(routing_verified_at, ''), COALESCE(last_dns_check_at, ''), COALESCE(last_dns_error, ''),
            COALESCE(last_routing_check_at, ''), COALESCE(last_routing_error, ''), ssl_status, COALESCE(ssl_certificate_issuer, ''),
            COALESCE(ssl_certificate_expires_at, ''), COALESCE(last_ssl_check_at, ''), COALESCE(last_ssl_error, ''), created_at, updated_at
     FROM tenant_domain_bindings
     WHERE tenant_id = ? AND is_primary = 1
     LIMIT 1`,
		id,
	)
	item, err := scanTenantDomainBinding(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TenantDomainBinding{}, ErrTenantDomainBindingNotFound
		}
		return TenantDomainBinding{}, fmt.Errorf("query primary tenant domain binding: %w", err)
	}
	return item, nil
}

func (r *Repository) CreateDomainBinding(ctx context.Context, params CreateTenantDomainBindingParams) (TenantDomainBinding, error) {
	tenantID := strings.TrimSpace(params.TenantID)
	if tenantID == "" {
		return TenantDomainBinding{}, fmt.Errorf("tenant id must not be empty")
	}
	if _, err := r.GetByID(ctx, tenantID); err != nil {
		return TenantDomainBinding{}, err
	}

	domain, err := normalizeDomainBindingDomain(params.Domain)
	if err != nil {
		return TenantDomainBinding{}, err
	}
	basePath := normalizePublicBasePath(params.BasePath)
	status, err := normalizeDomainBindingStatus(params.Status)
	if err != nil {
		return TenantDomainBinding{}, err
	}
	if params.IsPrimary && status != DomainBindingStatusActive {
		return TenantDomainBinding{}, fmt.Errorf("primaere Domain-Bindings muessen aktiv sein")
	}

	publicBaseURL := buildDomainBindingPublicBaseURL(domain, basePath)
	if err := r.ensureDomainBindingAvailable(ctx, tenantID, "", publicBaseURL); err != nil {
		return TenantDomainBinding{}, err
	}

	itemID := r.idFn("tdb")
	verificationToken := newDomainBindingVerificationToken(r)
	now := r.nowFn().UTC().Format(time.RFC3339)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TenantDomainBinding{}, fmt.Errorf("begin tenant domain binding transaction: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO tenant_domain_bindings (
      id, tenant_id, domain_host, base_path, status, is_primary, verification_token,
      ssl_status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		itemID,
		tenantID,
		domain,
		basePath,
		status,
		boolToInt(params.IsPrimary),
		verificationToken,
		initialSSLStatusForDomainBinding(status),
		now,
		now,
	); err != nil {
		_ = tx.Rollback()
		return TenantDomainBinding{}, fmt.Errorf("insert tenant domain binding: %w", err)
	}

	if params.IsPrimary {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE tenant_domain_bindings
       SET is_primary = 0, updated_at = ?
       WHERE tenant_id = ? AND id <> ?`,
			now,
			tenantID,
			itemID,
		); err != nil {
			_ = tx.Rollback()
			return TenantDomainBinding{}, fmt.Errorf("clear primary tenant domain bindings: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE tenants
       SET public_base_url = ?, updated_at = ?
       WHERE id = ?`,
			publicBaseURL,
			now,
			tenantID,
		); err != nil {
			_ = tx.Rollback()
			return TenantDomainBinding{}, fmt.Errorf("sync tenant public base url from primary domain binding: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return TenantDomainBinding{}, fmt.Errorf("commit tenant domain binding transaction: %w", err)
	}
	return r.GetDomainBindingByID(ctx, tenantID, itemID)
}

func (r *Repository) UpdateDomainBinding(ctx context.Context, tenantID, bindingID string, params UpdateTenantDomainBindingParams) (TenantDomainBinding, error) {
	current, err := r.GetDomainBindingByID(ctx, tenantID, bindingID)
	if err != nil {
		return TenantDomainBinding{}, err
	}

	domain := current.Domain
	basePath := current.BasePath
	domainChanged := false
	if params.Domain != nil {
		domain, err = normalizeDomainBindingDomain(*params.Domain)
		if err != nil {
			return TenantDomainBinding{}, err
		}
		domainChanged = domain != current.Domain
	}
	if params.BasePath != nil {
		basePath = normalizePublicBasePath(*params.BasePath)
		domainChanged = domainChanged || basePath != current.BasePath
	}

	if current.IsPrimary && domainChanged {
		return TenantDomainBinding{}, ErrTenantPrimaryDomainBindingLocked
	}

	status := current.Status
	if params.Status != nil {
		status, err = normalizeDomainBindingStatus(*params.Status)
		if err != nil {
			return TenantDomainBinding{}, err
		}
	}
	isPrimary := current.IsPrimary
	if params.IsPrimary != nil {
		isPrimary = *params.IsPrimary
	}
	if isPrimary && status != DomainBindingStatusActive {
		return TenantDomainBinding{}, fmt.Errorf("primaere Domain-Bindings muessen aktiv sein")
	}
	if current.IsPrimary && !isPrimary {
		return TenantDomainBinding{}, ErrTenantPrimaryDomainBindingLocked
	}
	if current.IsPrimary && status != DomainBindingStatusActive {
		return TenantDomainBinding{}, ErrTenantPrimaryDomainBindingLocked
	}

	verificationToken := current.VerificationToken
	dnsVerifiedAt := current.DNSVerifiedAt
	routingVerifiedAt := current.RoutingVerifiedAt
	lastDNSCheckAt := current.LastDNSCheckAt
	lastDNSError := current.LastDNSError
	lastRoutingCheckAt := current.LastRoutingCheckAt
	lastRoutingError := current.LastRoutingError
	sslStatus := current.SSLStatus
	sslCertificateIssuer := current.SSLCertificateIssuer
	sslCertificateExpiresAt := current.SSLCertificateExpiresAt
	lastSSLCheckAt := current.LastSSLCheckAt
	lastSSLError := current.LastSSLError
	if domainChanged {
		verificationToken = newDomainBindingVerificationToken(r)
		dnsVerifiedAt = nil
		routingVerifiedAt = nil
		lastDNSCheckAt = nil
		lastDNSError = ""
		lastRoutingCheckAt = nil
		lastRoutingError = ""
		sslStatus = DomainBindingSSLStatusPending
		sslCertificateIssuer = ""
		sslCertificateExpiresAt = nil
		lastSSLCheckAt = nil
		lastSSLError = ""
		if status != DomainBindingStatusDisabled {
			status = DomainBindingStatusPendingDNS
		}
	}

	publicBaseURL := buildDomainBindingPublicBaseURL(domain, basePath)
	if err := r.ensureDomainBindingAvailable(ctx, current.TenantID, current.ID, publicBaseURL); err != nil {
		return TenantDomainBinding{}, err
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TenantDomainBinding{}, fmt.Errorf("begin tenant domain binding update transaction: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE tenant_domain_bindings
     SET domain_host = ?, base_path = ?, status = ?, is_primary = ?, verification_token = ?,
         dns_verified_at = ?, routing_verified_at = ?, last_dns_check_at = ?, last_dns_error = ?,
         last_routing_check_at = ?, last_routing_error = ?, ssl_status = ?, ssl_certificate_issuer = ?,
         ssl_certificate_expires_at = ?, last_ssl_check_at = ?, last_ssl_error = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		domain,
		basePath,
		status,
		boolToInt(isPrimary),
		verificationToken,
		formatOptionalRFC3339(dnsVerifiedAt),
		formatOptionalRFC3339(routingVerifiedAt),
		formatOptionalRFC3339(lastDNSCheckAt),
		lastDNSError,
		formatOptionalRFC3339(lastRoutingCheckAt),
		lastRoutingError,
		normalizeSSLStatus(sslStatus),
		sslCertificateIssuer,
		formatOptionalRFC3339(sslCertificateExpiresAt),
		formatOptionalRFC3339(lastSSLCheckAt),
		lastSSLError,
		now,
		current.TenantID,
		current.ID,
	); err != nil {
		_ = tx.Rollback()
		return TenantDomainBinding{}, fmt.Errorf("update tenant domain binding: %w", err)
	}

	if isPrimary {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE tenant_domain_bindings
       SET is_primary = 0, updated_at = ?
       WHERE tenant_id = ? AND id <> ?`,
			now,
			current.TenantID,
			current.ID,
		); err != nil {
			_ = tx.Rollback()
			return TenantDomainBinding{}, fmt.Errorf("clear tenant primary domain bindings: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE tenants
       SET public_base_url = ?, updated_at = ?
       WHERE id = ?`,
			publicBaseURL,
			now,
			current.TenantID,
		); err != nil {
			_ = tx.Rollback()
			return TenantDomainBinding{}, fmt.Errorf("sync tenant public base url from updated primary domain binding: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return TenantDomainBinding{}, fmt.Errorf("commit tenant domain binding update transaction: %w", err)
	}
	return r.GetDomainBindingByID(ctx, current.TenantID, current.ID)
}

func (r *Repository) RotateDomainBindingVerificationToken(ctx context.Context, tenantID, bindingID string) (TenantDomainBinding, error) {
	current, err := r.GetDomainBindingByID(ctx, tenantID, bindingID)
	if err != nil {
		return TenantDomainBinding{}, err
	}

	nextStatus := current.Status
	if nextStatus != DomainBindingStatusDisabled {
		nextStatus = DomainBindingStatusPendingDNS
	}

	if _, err := r.db.ExecContext(
		ctx,
		`UPDATE tenant_domain_bindings
     SET verification_token = ?, status = ?, dns_verified_at = NULL, routing_verified_at = NULL,
         last_dns_check_at = NULL, last_dns_error = '', last_routing_check_at = NULL, last_routing_error = '',
         ssl_status = ?, ssl_certificate_issuer = '', ssl_certificate_expires_at = NULL,
         last_ssl_check_at = NULL, last_ssl_error = '', updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		newDomainBindingVerificationToken(r),
		nextStatus,
		initialSSLStatusForDomainBinding(nextStatus),
		r.nowFn().UTC().Format(time.RFC3339),
		current.TenantID,
		current.ID,
	); err != nil {
		return TenantDomainBinding{}, fmt.Errorf("rotate tenant domain binding verification token: %w", err)
	}
	return r.GetDomainBindingByID(ctx, current.TenantID, current.ID)
}

func (r *Repository) ApplyDomainCheckResult(ctx context.Context, tenantID, bindingID string, result DomainCheckResult) (TenantDomainBinding, error) {
	current, err := r.GetDomainBindingByID(ctx, tenantID, bindingID)
	if err != nil {
		return TenantDomainBinding{}, err
	}

	status := normalizeSystemManagedDomainBindingStatus(current.Status, result)
	if current.Status == DomainBindingStatusDisabled {
		status = DomainBindingStatusDisabled
	}

	if _, err := r.db.ExecContext(
		ctx,
		`UPDATE tenant_domain_bindings
     SET status = ?, dns_verified_at = ?, routing_verified_at = ?, last_dns_check_at = ?, last_dns_error = ?,
         last_routing_check_at = ?, last_routing_error = ?, ssl_status = ?, ssl_certificate_issuer = ?,
         ssl_certificate_expires_at = ?, last_ssl_check_at = ?, last_ssl_error = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		status,
		formatOptionalRFC3339(optionalTimeIf(result.DNSVerified, result.LastDNSCheckAt)),
		formatOptionalRFC3339(optionalTimeIf(result.RoutingVerified, result.LastRoutingCheckAt)),
		formatOptionalRFC3339(result.LastDNSCheckAt),
		strings.TrimSpace(result.LastDNSError),
		formatOptionalRFC3339(result.LastRoutingCheckAt),
		strings.TrimSpace(result.LastRoutingError),
		normalizeSSLStatus(result.SSLStatus),
		strings.TrimSpace(result.SSLCertificateIssuer),
		formatOptionalRFC3339(result.SSLCertificateExpiresAt),
		formatOptionalRFC3339(result.LastSSLCheckAt),
		strings.TrimSpace(result.LastSSLError),
		r.nowFn().UTC().Format(time.RFC3339),
		current.TenantID,
		current.ID,
	); err != nil {
		return TenantDomainBinding{}, fmt.Errorf("apply tenant domain binding check result: %w", err)
	}
	return r.GetDomainBindingByID(ctx, current.TenantID, current.ID)
}

func (r *Repository) DeleteDomainBinding(ctx context.Context, tenantID, bindingID string) error {
	current, err := r.GetDomainBindingByID(ctx, tenantID, bindingID)
	if err != nil {
		return err
	}
	if current.IsPrimary {
		return ErrTenantPrimaryDomainBindingLocked
	}
	if _, err := r.db.ExecContext(
		ctx,
		`DELETE FROM tenant_domain_bindings
     WHERE tenant_id = ? AND id = ?`,
		current.TenantID,
		current.ID,
	); err != nil {
		return fmt.Errorf("delete tenant domain binding: %w", err)
	}
	return nil
}

func (r *Repository) LookupPublicRoute(ctx context.Context, rawURL string) (PublicRouteMatch, error) {
	lookup, err := normalizePublicBaseLookup(rawURL)
	if err != nil {
		return PublicRouteMatch{}, err
	}

	bestPathLen := -1
	candidates := make([]PublicRouteMatch, 0, 1)
	indexByKey := map[string]int{}
	consider := func(match PublicRouteMatch) {
		pathLen := len(match.BasePath)
		if pathLen > bestPathLen {
			bestPathLen = pathLen
			candidates = []PublicRouteMatch{match}
			indexByKey = map[string]int{match.Tenant.ID + "|" + match.BasePath: 0}
			return
		}
		if pathLen < bestPathLen {
			return
		}

		key := match.Tenant.ID + "|" + match.BasePath
		if idx, ok := indexByKey[key]; ok {
			if candidates[idx].Source != "domain_binding" && match.Source == "domain_binding" {
				candidates[idx] = match
			}
			return
		}
		indexByKey[key] = len(candidates)
		candidates = append(candidates, match)
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT b.domain_host, b.base_path, t.id, t.slug, t.name, t.public_base_url, t.default_timezone, t.default_locale, t.status, t.created_at, t.updated_at
     FROM tenant_domain_bindings b
     INNER JOIN tenants t ON t.id = b.tenant_id
     WHERE b.status = ?`,
		DomainBindingStatusActive,
	)
	if err != nil {
		return PublicRouteMatch{}, fmt.Errorf("query tenant domain bindings for public route lookup: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			domainHost string
			basePath   string
		)
		item, scanErr := scanTenant(scanTenantRowAdapter{
			row: rows,
			prefixDestinations: []any{
				&domainHost,
				&basePath,
			},
		})
		if scanErr != nil {
			return PublicRouteMatch{}, fmt.Errorf("scan tenant domain binding public route: %w", scanErr)
		}
		if domainHost != lookup.host {
			continue
		}
		normalizedBasePath := normalizePublicBasePath(basePath)
		if !publicBasePathMatches(normalizedBasePath, lookup.path) {
			continue
		}
		consider(PublicRouteMatch{
			Tenant:   item,
			BaseURL:  buildDomainBindingPublicBaseURL(domainHost, normalizedBasePath),
			BasePath: normalizedBasePath,
			Source:   "domain_binding",
		})
	}
	if err := rows.Err(); err != nil {
		return PublicRouteMatch{}, fmt.Errorf("iterate tenant domain binding public routes: %w", err)
	}

	rows, err = r.db.QueryContext(
		ctx,
		`SELECT id, slug, name, public_base_url, default_timezone, default_locale, status, created_at, updated_at
     FROM tenants`,
	)
	if err != nil {
		return PublicRouteMatch{}, fmt.Errorf("query tenants by public base url: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		item, scanErr := scanTenant(rows)
		if scanErr != nil {
			return PublicRouteMatch{}, fmt.Errorf("scan tenant public base route: %w", scanErr)
		}
		candidate, candidateErr := normalizePublicBaseLookup(item.PublicBaseURL)
		if candidateErr != nil {
			continue
		}
		if candidate.host != lookup.host {
			continue
		}
		if !publicBasePathMatches(candidate.path, lookup.path) {
			continue
		}
		consider(PublicRouteMatch{
			Tenant:   item,
			BaseURL:  item.PublicBaseURL,
			BasePath: candidate.path,
			Source:   "tenant_public_base_url",
		})
	}
	if err := rows.Err(); err != nil {
		return PublicRouteMatch{}, fmt.Errorf("iterate tenants by public base url: %w", err)
	}

	switch len(candidates) {
	case 0:
		return PublicRouteMatch{}, ErrTenantNotFound
	case 1:
		return candidates[0], nil
	default:
		return PublicRouteMatch{}, ErrTenantPathAmbiguous
	}
}

func normalizeDomainBindingDomain(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("tenant domain binding host must not be empty")
	}

	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("parse tenant domain binding host: %w", err)
		}
		if strings.TrimSpace(parsed.Path) != "" && strings.TrimSpace(parsed.Path) != "/" {
			return "", fmt.Errorf("tenant domain binding host darf keinen Pfad enthalten")
		}
		if parsed.RawQuery != "" || parsed.Fragment != "" {
			return "", fmt.Errorf("tenant domain binding host darf keine Query oder Fragment enthalten")
		}
		value = parsed.Host
	}

	host, err := normalizeLookupHost(value)
	if err != nil {
		return "", err
	}
	return host, nil
}

func normalizeDomainBindingStatus(raw string) (string, error) {
	status := strings.ToLower(strings.TrimSpace(raw))
	switch status {
	case "", DomainBindingStatusPendingDNS:
		return DomainBindingStatusPendingDNS, nil
	case DomainBindingStatusDNSVerified, DomainBindingStatusSSLPending, DomainBindingStatusActive, DomainBindingStatusDisabled:
		return status, nil
	default:
		return "", fmt.Errorf("tenant domain binding status %q ist ungueltig", raw)
	}
}

func normalizeSSLStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case DomainBindingSSLStatusValid:
		return DomainBindingSSLStatusValid
	case DomainBindingSSLStatusInvalid:
		return DomainBindingSSLStatusInvalid
	case DomainBindingSSLStatusExpired:
		return DomainBindingSSLStatusExpired
	default:
		return DomainBindingSSLStatusPending
	}
}

func buildDomainBindingPublicBaseURL(domain, basePath string) string {
	host := strings.TrimSpace(domain)
	path := normalizePublicBasePath(basePath)
	if host == "" {
		return ""
	}
	if path == "/" {
		return "https://" + host
	}
	return "https://" + host + path
}

func buildDomainBindingVerificationRecordName(domain string) string {
	normalized := strings.Trim(strings.TrimSpace(domain), ".")
	if normalized == "" {
		return ""
	}
	return "_eep-domain-verification." + normalized
}

func buildDomainBindingVerificationRecordValue(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	return "eep-domain-verification=" + trimmed
}

func initialSSLStatusForDomainBinding(status string) string {
	if strings.TrimSpace(status) == DomainBindingStatusActive {
		return DomainBindingSSLStatusValid
	}
	return DomainBindingSSLStatusPending
}

func normalizeSystemManagedDomainBindingStatus(currentStatus string, result DomainCheckResult) string {
	if currentStatus == DomainBindingStatusDisabled {
		return DomainBindingStatusDisabled
	}
	if !result.DNSVerified {
		return DomainBindingStatusPendingDNS
	}
	if !result.RoutingVerified {
		return DomainBindingStatusDNSVerified
	}
	if normalizeSSLStatus(result.SSLStatus) != DomainBindingSSLStatusValid {
		return DomainBindingStatusSSLPending
	}
	return DomainBindingStatusActive
}

func (r *Repository) ensureDomainBindingAvailable(ctx context.Context, tenantID, ignoreBindingID, rawURL string) error {
	if err := r.ensurePublicRouteAvailable(ctx, tenantID, ignoreBindingID, rawURL); err != nil {
		if errors.Is(err, ErrTenantPublicBaseURLConflict) {
			return ErrTenantDomainBindingConflict
		}
		return err
	}
	return nil
}

func (r *Repository) ensurePublicRouteAvailable(ctx context.Context, ignoreTenantID, ignoreBindingID, rawURL string) error {
	lookup, err := normalizePublicBaseLookup(rawURL)
	if err != nil {
		return err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, public_base_url
     FROM tenants`,
	)
	if err != nil {
		return fmt.Errorf("query tenant public routes: %w", err)
	}
	defer rows.Close()

	ignoreTenantID = strings.TrimSpace(ignoreTenantID)
	for rows.Next() {
		var tenantID string
		var publicBaseURL string
		if err := rows.Scan(&tenantID, &publicBaseURL); err != nil {
			return fmt.Errorf("scan tenant public route: %w", err)
		}
		if ignoreTenantID != "" && strings.TrimSpace(tenantID) == ignoreTenantID {
			continue
		}
		candidate, err := normalizePublicBaseLookup(publicBaseURL)
		if err != nil {
			continue
		}
		if candidate.host == lookup.host && candidate.path == lookup.path {
			return ErrTenantPublicBaseURLConflict
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tenant public routes: %w", err)
	}

	rows, err = r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, domain_host, base_path
     FROM tenant_domain_bindings`,
	)
	if err != nil {
		return fmt.Errorf("query tenant domain routes: %w", err)
	}
	defer rows.Close()

	ignoreBindingID = strings.TrimSpace(ignoreBindingID)
	for rows.Next() {
		var (
			bindingID string
			tenantID  string
			domain    string
			basePath  string
		)
		if err := rows.Scan(&bindingID, &tenantID, &domain, &basePath); err != nil {
			return fmt.Errorf("scan tenant domain route: %w", err)
		}
		if ignoreBindingID != "" && strings.TrimSpace(bindingID) == ignoreBindingID {
			continue
		}
		candidate := publicBaseLookup{
			host: strings.TrimSpace(domain),
			path: normalizePublicBasePath(basePath),
		}
		if candidate.host != lookup.host || candidate.path != lookup.path {
			continue
		}
		return ErrTenantDomainBindingConflict
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tenant domain routes: %w", err)
	}
	return nil
}

func newDomainBindingVerificationToken(r *Repository) string {
	token := strings.TrimSpace(strings.TrimPrefix(r.idFn("dnsv"), "dnsv_"))
	token = strings.ReplaceAll(token, "_", "")
	if token == "" {
		token = strings.ReplaceAll(strings.TrimSpace(r.idFn("tdb")), "_", "")
	}
	return token
}

func formatOptionalRFC3339(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func parseOptionalRFC3339(raw, label string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", label, err)
	}
	utc := parsed.UTC()
	return &utc, nil
}

func optionalTimeIf(condition bool, value *time.Time) *time.Time {
	if !condition {
		return nil
	}
	return value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func scanTenantDomainBinding(row rowScanner) (TenantDomainBinding, error) {
	var (
		item                    TenantDomainBinding
		isPrimaryRaw            int
		dnsVerifiedAtRaw        string
		routingVerifiedAtRaw    string
		lastDNSCheckAtRaw       string
		lastRoutingCheckAtRaw   string
		sslCertificateExpiryRaw string
		lastSSLCheckAtRaw       string
		createdAtRaw            string
		updatedAtRaw            string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.Domain,
		&item.BasePath,
		&item.Status,
		&isPrimaryRaw,
		&item.VerificationToken,
		&dnsVerifiedAtRaw,
		&routingVerifiedAtRaw,
		&lastDNSCheckAtRaw,
		&item.LastDNSError,
		&lastRoutingCheckAtRaw,
		&item.LastRoutingError,
		&item.SSLStatus,
		&item.SSLCertificateIssuer,
		&sslCertificateExpiryRaw,
		&lastSSLCheckAtRaw,
		&item.LastSSLError,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		return TenantDomainBinding{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return TenantDomainBinding{}, fmt.Errorf("parse tenant domain binding created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return TenantDomainBinding{}, fmt.Errorf("parse tenant domain binding updated_at: %w", err)
	}

	item.DNSVerifiedAt, err = parseOptionalRFC3339(dnsVerifiedAtRaw, "tenant domain binding dns_verified_at")
	if err != nil {
		return TenantDomainBinding{}, err
	}
	item.RoutingVerifiedAt, err = parseOptionalRFC3339(routingVerifiedAtRaw, "tenant domain binding routing_verified_at")
	if err != nil {
		return TenantDomainBinding{}, err
	}
	item.LastDNSCheckAt, err = parseOptionalRFC3339(lastDNSCheckAtRaw, "tenant domain binding last_dns_check_at")
	if err != nil {
		return TenantDomainBinding{}, err
	}
	item.LastRoutingCheckAt, err = parseOptionalRFC3339(lastRoutingCheckAtRaw, "tenant domain binding last_routing_check_at")
	if err != nil {
		return TenantDomainBinding{}, err
	}
	item.SSLCertificateExpiresAt, err = parseOptionalRFC3339(sslCertificateExpiryRaw, "tenant domain binding ssl_certificate_expires_at")
	if err != nil {
		return TenantDomainBinding{}, err
	}
	item.LastSSLCheckAt, err = parseOptionalRFC3339(lastSSLCheckAtRaw, "tenant domain binding last_ssl_check_at")
	if err != nil {
		return TenantDomainBinding{}, err
	}

	item.BasePath = normalizePublicBasePath(item.BasePath)
	item.Status = firstNonEmpty(item.Status, DomainBindingStatusPendingDNS)
	item.SSLStatus = normalizeSSLStatus(item.SSLStatus)
	item.IsPrimary = isPrimaryRaw == 1
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return item, nil
}

type scanTenantRowAdapter struct {
	row                rowScanner
	prefixDestinations []any
}

func (a scanTenantRowAdapter) Scan(dest ...any) error {
	values := append([]any{}, a.prefixDestinations...)
	values = append(values, dest...)
	return a.row.Scan(values...)
}
