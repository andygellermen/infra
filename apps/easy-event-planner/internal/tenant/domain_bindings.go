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
		`SELECT id, tenant_id, domain_host, base_path, status, is_primary, created_at, updated_at
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
		`SELECT id, tenant_id, domain_host, base_path, status, is_primary, created_at, updated_at
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
		`SELECT id, tenant_id, domain_host, base_path, status, is_primary, created_at, updated_at
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
	now := r.nowFn().UTC().Format(time.RFC3339)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return TenantDomainBinding{}, fmt.Errorf("begin tenant domain binding transaction: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO tenant_domain_bindings (
      id, tenant_id, domain_host, base_path, status, is_primary, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		itemID,
		tenantID,
		domain,
		basePath,
		status,
		boolToInt(params.IsPrimary),
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
	if params.Domain != nil {
		domain, err = normalizeDomainBindingDomain(*params.Domain)
		if err != nil {
			return TenantDomainBinding{}, err
		}
	}
	basePath := current.BasePath
	if params.BasePath != nil {
		basePath = normalizePublicBasePath(*params.BasePath)
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
     SET domain_host = ?, base_path = ?, status = ?, is_primary = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		domain,
		basePath,
		status,
		boolToInt(isPrimary),
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
	case DomainBindingStatusActive, DomainBindingStatusDisabled:
		return status, nil
	default:
		return "", fmt.Errorf("tenant domain binding status %q ist ungueltig", raw)
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
		if ignoreTenantID != "" && strings.TrimSpace(tenantID) == ignoreTenantID {
			return ErrTenantDomainBindingConflict
		}
		return ErrTenantDomainBindingConflict
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate tenant domain routes: %w", err)
	}
	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func scanTenantDomainBinding(row rowScanner) (TenantDomainBinding, error) {
	var (
		item         TenantDomainBinding
		isPrimaryRaw int
		createdAtRaw string
		updatedAtRaw string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.Domain,
		&item.BasePath,
		&item.Status,
		&isPrimaryRaw,
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

	item.BasePath = normalizePublicBasePath(item.BasePath)
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
