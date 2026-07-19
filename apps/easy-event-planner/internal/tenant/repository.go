package tenant

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrTenantNotFound                            = errors.New("tenant not found")
	ErrTenantSettingsNotFound                    = errors.New("tenant settings not found")
	ErrTenantHostAmbiguous                       = errors.New("tenant host ambiguous")
	ErrTenantPathAmbiguous                       = errors.New("tenant public path ambiguous")
	ErrTenantPublicBaseURLConflict               = errors.New("tenant public base url conflicts with another tenant")
	ErrTenantDomainBindingNotFound               = errors.New("tenant domain binding not found")
	ErrTenantDomainBindingConflict               = errors.New("tenant domain binding conflicts with another public route")
	ErrTenantPrimaryDomainBindingLocked          = errors.New("primary tenant domain binding must remain active until another primary domain is selected")
	ErrTenantPublicBaseURLManagedByDomainBinding = errors.New("tenant public base url is managed by the primary domain binding")
	ErrTenantUserNotFound                        = errors.New("tenant user not found")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Repository struct {
	db    *sql.DB
	nowFn func() time.Time
	idFn  func(prefix string) string
}

func NewRepository(sqlDB *sql.DB) *Repository {
	return &Repository{
		db:    sqlDB,
		nowFn: func() time.Time { return time.Now().UTC() },
		idFn:  defaultID,
	}
}

func (r *Repository) CreateTenant(ctx context.Context, params CreateTenantParams) (Tenant, error) {
	if r.db == nil {
		return Tenant{}, fmt.Errorf("tenant repository database is nil")
	}

	slug, err := normalizeSlug(params.Slug)
	if err != nil {
		return Tenant{}, err
	}
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return Tenant{}, fmt.Errorf("tenant name must not be empty")
	}
	publicBaseURL, err := normalizePublicBaseURL(params.PublicBaseURL)
	if err != nil {
		return Tenant{}, err
	}
	if err := r.ensurePublicBaseURLAvailable(ctx, "", publicBaseURL); err != nil {
		return Tenant{}, err
	}

	settings, err := normalizeSettingsInput(params.Settings)
	if err != nil {
		return Tenant{}, err
	}

	tenantID := strings.TrimSpace(params.ID)
	if tenantID == "" {
		tenantID = r.idFn("tnt")
	}

	tenant := Tenant{
		ID:              tenantID,
		Slug:            slug,
		Name:            name,
		PublicBaseURL:   publicBaseURL,
		DefaultTimezone: firstNonEmpty(params.DefaultTimezone, DefaultTimezone),
		DefaultLocale:   firstNonEmpty(params.DefaultLocale, DefaultLocale),
		Status:          firstNonEmpty(params.Status, DefaultStatus),
	}

	now := r.nowFn().UTC().Format(time.RFC3339)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Tenant{}, fmt.Errorf("begin tenant create transaction: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO tenants (
      id, slug, name, public_base_url, default_timezone, default_locale, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tenant.ID,
		tenant.Slug,
		tenant.Name,
		tenant.PublicBaseURL,
		tenant.DefaultTimezone,
		tenant.DefaultLocale,
		tenant.Status,
		now,
		now,
	); err != nil {
		_ = tx.Rollback()
		return Tenant{}, fmt.Errorf("insert tenant: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO tenant_settings (
      tenant_id, sender_email, sender_name, paypal_mode, paypal_client_id, paypal_merchant_id,
      default_retention_days, settings_json, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tenant.ID,
		settings.SenderEmail,
		settings.SenderName,
		settings.PayPalMode,
		settings.PayPalClientID,
		settings.PayPalMerchantID,
		settings.DefaultRetentionDays,
		settings.SettingsJSON,
		now,
		now,
	); err != nil {
		_ = tx.Rollback()
		return Tenant{}, fmt.Errorf("insert tenant settings: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return Tenant{}, fmt.Errorf("commit tenant create transaction: %w", err)
	}

	return r.GetByID(ctx, tenant.ID)
}

func (r *Repository) ListTenants(ctx context.Context) ([]Tenant, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, slug, name, public_base_url, default_timezone, default_locale, status, created_at, updated_at
     FROM tenants
     ORDER BY name ASC, slug ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query tenants: %w", err)
	}
	defer rows.Close()

	items := make([]Tenant, 0)
	for rows.Next() {
		item, scanErr := scanTenant(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan tenant list: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenant list: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, tenantID string) (Tenant, error) {
	id := strings.TrimSpace(tenantID)
	if id == "" {
		return Tenant{}, fmt.Errorf("tenant id must not be empty")
	}
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, slug, name, public_base_url, default_timezone, default_locale, status, created_at, updated_at
     FROM tenants
     WHERE id = ?`,
		id,
	)
	tenant, err := scanTenant(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tenant{}, ErrTenantNotFound
		}
		return Tenant{}, fmt.Errorf("query tenant by id: %w", err)
	}
	return tenant, nil
}

func (r *Repository) LookupBySlug(ctx context.Context, slug string) (Tenant, error) {
	normalizedSlug, err := normalizeSlug(slug)
	if err != nil {
		return Tenant{}, err
	}
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, slug, name, public_base_url, default_timezone, default_locale, status, created_at, updated_at
     FROM tenants
     WHERE slug = ?`,
		normalizedSlug,
	)
	tenant, err := scanTenant(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tenant{}, ErrTenantNotFound
		}
		return Tenant{}, fmt.Errorf("query tenant by slug: %w", err)
	}
	return tenant, nil
}

func (r *Repository) LookupByPublicHost(ctx context.Context, host string) (Tenant, error) {
	normalizedHost, err := normalizeLookupHost(host)
	if err != nil {
		return Tenant{}, err
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT t.id, t.slug, t.name, t.public_base_url, t.default_timezone, t.default_locale, t.status, t.created_at, t.updated_at
     FROM tenant_domain_bindings b
     INNER JOIN tenants t ON t.id = b.tenant_id
     WHERE b.status = ? AND b.domain_host = ?`,
		DomainBindingStatusActive,
		normalizedHost,
	)
	if err != nil {
		return Tenant{}, fmt.Errorf("query tenants by active domain binding host: %w", err)
	}
	defer rows.Close()

	bindingMatches := make([]Tenant, 0, 1)
	seenTenantIDs := map[string]struct{}{}
	for rows.Next() {
		item, scanErr := scanTenant(rows)
		if scanErr != nil {
			return Tenant{}, fmt.Errorf("scan tenant by active domain binding host: %w", scanErr)
		}
		if _, ok := seenTenantIDs[item.ID]; ok {
			continue
		}
		seenTenantIDs[item.ID] = struct{}{}
		bindingMatches = append(bindingMatches, item)
	}
	if err := rows.Err(); err != nil {
		return Tenant{}, fmt.Errorf("iterate tenants by active domain binding host: %w", err)
	}
	switch len(bindingMatches) {
	case 1:
		return bindingMatches[0], nil
	case 0:
	default:
		return Tenant{}, ErrTenantHostAmbiguous
	}

	rows, err = r.db.QueryContext(
		ctx,
		`SELECT id, slug, name, public_base_url, default_timezone, default_locale, status, created_at, updated_at
     FROM tenants`,
	)
	if err != nil {
		return Tenant{}, fmt.Errorf("query tenants by public host: %w", err)
	}
	defer rows.Close()

	matches := make([]Tenant, 0, 1)
	for rows.Next() {
		item, scanErr := scanTenant(rows)
		if scanErr != nil {
			return Tenant{}, fmt.Errorf("scan tenant by public host: %w", scanErr)
		}

		candidateHost, hostErr := normalizeLookupHost(item.PublicBaseURL)
		if hostErr != nil {
			continue
		}
		if candidateHost == normalizedHost {
			matches = append(matches, item)
		}
	}
	if err := rows.Err(); err != nil {
		return Tenant{}, fmt.Errorf("iterate tenants by public host: %w", err)
	}

	switch len(matches) {
	case 0:
		return Tenant{}, ErrTenantNotFound
	case 1:
		return matches[0], nil
	default:
		return Tenant{}, ErrTenantHostAmbiguous
	}
}

func (r *Repository) LookupByPublicBaseURL(ctx context.Context, rawURL string) (Tenant, error) {
	match, err := r.LookupPublicRoute(ctx, rawURL)
	if err != nil {
		return Tenant{}, err
	}
	return match.Tenant, nil
}

func (r *Repository) UpdateTenant(ctx context.Context, tenantID string, params UpdateTenantParams) (Tenant, error) {
	current, err := r.GetByID(ctx, tenantID)
	if err != nil {
		return Tenant{}, err
	}

	name := current.Name
	if params.Name != nil {
		name = strings.TrimSpace(*params.Name)
		if name == "" {
			return Tenant{}, fmt.Errorf("tenant name must not be empty")
		}
	}

	publicBaseURL := current.PublicBaseURL
	if params.PublicBaseURL != nil {
		publicBaseURL, err = normalizePublicBaseURL(*params.PublicBaseURL)
		if err != nil {
			return Tenant{}, err
		}
		if primaryBinding, bindingErr := r.GetPrimaryDomainBinding(ctx, current.ID); bindingErr == nil {
			managedPublicBaseURL := buildDomainBindingPublicBaseURL(primaryBinding.Domain, primaryBinding.BasePath)
			if managedPublicBaseURL != publicBaseURL {
				return Tenant{}, ErrTenantPublicBaseURLManagedByDomainBinding
			}
		} else if bindingErr != nil && !errors.Is(bindingErr, ErrTenantDomainBindingNotFound) {
			return Tenant{}, bindingErr
		}
		if err := r.ensurePublicBaseURLAvailable(ctx, current.ID, publicBaseURL); err != nil {
			return Tenant{}, err
		}
	}

	defaultTimezone := current.DefaultTimezone
	if params.DefaultTimezone != nil {
		defaultTimezone = firstNonEmpty(*params.DefaultTimezone, DefaultTimezone)
	}

	defaultLocale := current.DefaultLocale
	if params.DefaultLocale != nil {
		defaultLocale = firstNonEmpty(*params.DefaultLocale, DefaultLocale)
	}

	if _, err := r.db.ExecContext(
		ctx,
		`UPDATE tenants
     SET name = ?, public_base_url = ?, default_timezone = ?, default_locale = ?, updated_at = ?
     WHERE id = ?`,
		name,
		publicBaseURL,
		defaultTimezone,
		defaultLocale,
		r.nowFn().UTC().Format(time.RFC3339),
		strings.TrimSpace(tenantID),
	); err != nil {
		return Tenant{}, fmt.Errorf("update tenant: %w", err)
	}

	return r.GetByID(ctx, tenantID)
}

func (r *Repository) GetSettings(ctx context.Context, tenantID string) (TenantSettings, error) {
	id := strings.TrimSpace(tenantID)
	if id == "" {
		return TenantSettings{}, fmt.Errorf("tenant id must not be empty")
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT tenant_id, sender_email, sender_name, paypal_mode, paypal_client_id, paypal_merchant_id,
            default_retention_days, settings_json, created_at, updated_at
     FROM tenant_settings
     WHERE tenant_id = ?`,
		id,
	)

	settings, err := scanTenantSettings(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TenantSettings{}, ErrTenantSettingsNotFound
		}
		return TenantSettings{}, fmt.Errorf("query tenant settings: %w", err)
	}
	return settings, nil
}

func (r *Repository) UpsertSettings(ctx context.Context, params UpsertTenantSettingsParams) (TenantSettings, error) {
	tenantID := strings.TrimSpace(params.TenantID)
	if tenantID == "" {
		return TenantSettings{}, fmt.Errorf("tenant id must not be empty")
	}

	settings, err := normalizeSettingsInput(params.Settings)
	if err != nil {
		return TenantSettings{}, err
	}
	now := r.nowFn().UTC().Format(time.RFC3339)

	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO tenant_settings (
      tenant_id, sender_email, sender_name, paypal_mode, paypal_client_id, paypal_merchant_id,
      default_retention_days, settings_json, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(tenant_id) DO UPDATE SET
      sender_email = excluded.sender_email,
      sender_name = excluded.sender_name,
      paypal_mode = excluded.paypal_mode,
      paypal_client_id = excluded.paypal_client_id,
      paypal_merchant_id = excluded.paypal_merchant_id,
      default_retention_days = excluded.default_retention_days,
      settings_json = excluded.settings_json,
      updated_at = excluded.updated_at`,
		tenantID,
		settings.SenderEmail,
		settings.SenderName,
		settings.PayPalMode,
		settings.PayPalClientID,
		settings.PayPalMerchantID,
		settings.DefaultRetentionDays,
		settings.SettingsJSON,
		now,
		now,
	); err != nil {
		return TenantSettings{}, fmt.Errorf("upsert tenant settings: %w", err)
	}

	return r.GetSettings(ctx, tenantID)
}

func (r *Repository) SeedTenant(ctx context.Context, input SeedInput) (SeedResult, error) {
	lookup, err := r.LookupBySlug(ctx, input.Slug)
	if err != nil && !errors.Is(err, ErrTenantNotFound) {
		return SeedResult{}, err
	}

	result := SeedResult{}
	if errors.Is(err, ErrTenantNotFound) {
		created, createErr := r.CreateTenant(ctx, CreateTenantParams{
			Slug:            input.Slug,
			Name:            input.Name,
			PublicBaseURL:   input.PublicBaseURL,
			DefaultTimezone: input.DefaultTimezone,
			DefaultLocale:   input.DefaultLocale,
			Status:          input.Status,
			Settings:        input.Settings,
		})
		if createErr != nil {
			return SeedResult{}, createErr
		}
		result.Tenant = created
		result.Created = true
	} else {
		result.Tenant = lookup
	}

	settings, err := r.UpsertSettings(ctx, UpsertTenantSettingsParams{
		TenantID: result.Tenant.ID,
		Settings: input.Settings,
	})
	if err != nil {
		return SeedResult{}, err
	}
	result.Settings = settings

	if err := r.upsertSeedAdminUser(ctx, result.Tenant.ID, input.AdminUser); err != nil {
		return SeedResult{}, err
	}

	return result, nil
}

func (r *Repository) upsertSeedAdminUser(ctx context.Context, tenantID string, input SeedAdminUserInput) error {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return nil
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("parse seed admin email: %w", err)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = "Owner"
	}

	role := strings.ToLower(strings.TrimSpace(input.Role))
	if role == "" {
		role = "owner"
	}

	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status == "" {
		status = "active"
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO tenant_users (
      id, tenant_id, email, name, role, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT(tenant_id, email) DO UPDATE SET
      name = excluded.name,
      role = excluded.role,
      status = excluded.status,
      updated_at = excluded.updated_at`,
		r.idFn("usr"),
		strings.TrimSpace(tenantID),
		email,
		name,
		role,
		status,
		now,
		now,
	); err != nil {
		return fmt.Errorf("upsert seed admin user: %w", err)
	}

	return nil
}

func (r *Repository) ListUsers(ctx context.Context, tenantID string) ([]TenantUser, error) {
	id := strings.TrimSpace(tenantID)
	if id == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, email, name, role, status, created_at, updated_at
     FROM tenant_users
     WHERE tenant_id = ?
     ORDER BY email ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("query tenant users: %w", err)
	}
	defer rows.Close()

	items := make([]TenantUser, 0)
	for rows.Next() {
		item, scanErr := scanTenantUser(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan tenant user: %w", scanErr)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenant users: %w", err)
	}
	return items, nil
}

func (r *Repository) GetUserByID(ctx context.Context, tenantID, userID string) (TenantUser, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedTenantID == "" {
		return TenantUser{}, fmt.Errorf("tenant id must not be empty")
	}
	if normalizedUserID == "" {
		return TenantUser{}, fmt.Errorf("user id must not be empty")
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, email, name, role, status, created_at, updated_at
     FROM tenant_users
     WHERE tenant_id = ? AND id = ?`,
		normalizedTenantID,
		normalizedUserID,
	)
	item, err := scanTenantUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TenantUser{}, ErrTenantUserNotFound
		}
		return TenantUser{}, fmt.Errorf("query tenant user: %w", err)
	}
	return item, nil
}

func (r *Repository) CreateUser(ctx context.Context, tenantID string, params CreateTenantUserParams) (TenantUser, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return TenantUser{}, fmt.Errorf("tenant id must not be empty")
	}

	email, name, role, status, err := normalizeTenantUserInput(params.Email, params.Name, params.Role, params.Status)
	if err != nil {
		return TenantUser{}, err
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	userID := r.idFn("usr")
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO tenant_users (
      id, tenant_id, email, name, role, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID,
		normalizedTenantID,
		email,
		name,
		role,
		status,
		now,
		now,
	); err != nil {
		return TenantUser{}, fmt.Errorf("insert tenant user: %w", err)
	}

	return r.GetUserByID(ctx, normalizedTenantID, userID)
}

func (r *Repository) UpdateUser(ctx context.Context, tenantID, userID string, params UpdateTenantUserParams) (TenantUser, error) {
	current, err := r.GetUserByID(ctx, tenantID, userID)
	if err != nil {
		return TenantUser{}, err
	}

	email := current.Email
	name := current.Name
	role := current.Role
	status := current.Status
	if params.Email != nil {
		email = *params.Email
	}
	if params.Name != nil {
		name = *params.Name
	}
	if params.Role != nil {
		role = *params.Role
	}
	if params.Status != nil {
		status = *params.Status
	}

	email, name, role, status, err = normalizeTenantUserInput(email, name, role, status)
	if err != nil {
		return TenantUser{}, err
	}

	if _, err := r.db.ExecContext(
		ctx,
		`UPDATE tenant_users
     SET email = ?, name = ?, role = ?, status = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		email,
		name,
		role,
		status,
		r.nowFn().UTC().Format(time.RFC3339),
		strings.TrimSpace(tenantID),
		strings.TrimSpace(userID),
	); err != nil {
		return TenantUser{}, fmt.Errorf("update tenant user: %w", err)
	}

	return r.GetUserByID(ctx, tenantID, userID)
}

func (r *Repository) DeleteUser(ctx context.Context, tenantID, userID string) (bool, error) {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM tenant_users
     WHERE tenant_id = ? AND id = ?`,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(userID),
	)
	if err != nil {
		return false, fmt.Errorf("delete tenant user: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("tenant user delete rows affected: %w", err)
	}
	return rowsAffected > 0, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTenant(row rowScanner) (Tenant, error) {
	var (
		tenant          Tenant
		createdAtRaw    string
		updatedAtRaw    string
		defaultTimezone string
		defaultLocale   string
		status          string
		publicBaseURL   string
	)

	if err := row.Scan(
		&tenant.ID,
		&tenant.Slug,
		&tenant.Name,
		&publicBaseURL,
		&defaultTimezone,
		&defaultLocale,
		&status,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		return Tenant{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return Tenant{}, fmt.Errorf("parse tenant created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return Tenant{}, fmt.Errorf("parse tenant updated_at: %w", err)
	}

	tenant.PublicBaseURL = publicBaseURL
	tenant.DefaultTimezone = defaultTimezone
	tenant.DefaultLocale = defaultLocale
	tenant.Status = status
	tenant.CreatedAt = createdAt
	tenant.UpdatedAt = updatedAt
	return tenant, nil
}

func scanTenantSettings(row rowScanner) (TenantSettings, error) {
	var (
		settings     TenantSettings
		createdAtRaw string
		updatedAtRaw string
	)

	if err := row.Scan(
		&settings.TenantID,
		&settings.SenderEmail,
		&settings.SenderName,
		&settings.PayPalMode,
		&settings.PayPalClientID,
		&settings.PayPalMerchantID,
		&settings.DefaultRetentionDays,
		&settings.SettingsJSON,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		return TenantSettings{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return TenantSettings{}, fmt.Errorf("parse tenant settings created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return TenantSettings{}, fmt.Errorf("parse tenant settings updated_at: %w", err)
	}

	settings.CreatedAt = createdAt
	settings.UpdatedAt = updatedAt
	return settings, nil
}

func scanTenantUser(row rowScanner) (TenantUser, error) {
	var (
		item         TenantUser
		createdAtRaw string
		updatedAtRaw string
	)

	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.Email,
		&item.Name,
		&item.Role,
		&item.Status,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		return TenantUser{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return TenantUser{}, fmt.Errorf("parse tenant user created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return TenantUser{}, fmt.Errorf("parse tenant user updated_at: %w", err)
	}

	item.Email = strings.ToLower(strings.TrimSpace(item.Email))
	item.Name = strings.TrimSpace(item.Name)
	item.Role = normalizeTenantUserRole(item.Role)
	item.Status = normalizeTenantUserStatus(item.Status)
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return item, nil
}

func normalizeTenantUserInput(rawEmail, rawName, rawRole, rawStatus string) (string, string, string, string, error) {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if email == "" {
		return "", "", "", "", fmt.Errorf("user email must not be empty")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", "", "", "", fmt.Errorf("parse user email: %w", err)
	}

	name := strings.TrimSpace(rawName)
	if name == "" {
		name = email
	}

	role := normalizeTenantUserRole(rawRole)
	status := normalizeTenantUserStatus(rawStatus)
	return email, name, role, status, nil
}

func normalizeTenantUserRole(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "owner", "admin", "readonly":
		return strings.ToLower(strings.TrimSpace(raw))
	case "", "event_manager":
		return "event_manager"
	default:
		return "event_manager"
	}
}

func normalizeTenantUserStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "active":
		return "active"
	case "invited", "disabled":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "active"
	}
}

func normalizeSlug(raw string) (string, error) {
	slug := strings.ToLower(strings.Trim(strings.TrimSpace(raw), "/"))
	if slug == "" {
		return "", fmt.Errorf("tenant slug must not be empty")
	}
	if len(slug) > 64 {
		return "", fmt.Errorf("tenant slug must be <= 64 characters")
	}
	if !slugPattern.MatchString(slug) {
		return "", fmt.Errorf("tenant slug %q is invalid", raw)
	}
	return slug, nil
}

func normalizePublicBaseURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("tenant public base url must not be empty")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse tenant public base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("tenant public base url must include scheme and host")
	}

	return strings.TrimRight(value, "/"), nil
}

func normalizeLookupHost(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return "", fmt.Errorf("tenant host must not be empty")
	}

	if strings.Contains(value, "://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("parse tenant host: %w", err)
		}
		value = parsed.Host
	}

	value = strings.TrimSuffix(value, ".")
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	value = strings.Trim(value, "[]")
	if value == "" {
		return "", fmt.Errorf("tenant host must not be empty")
	}
	return value, nil
}

func (r *Repository) ensurePublicBaseURLAvailable(ctx context.Context, ignoreTenantID, rawURL string) error {
	if err := r.ensurePublicRouteAvailable(ctx, ignoreTenantID, "", rawURL); err != nil {
		if errors.Is(err, ErrTenantDomainBindingConflict) {
			return ErrTenantPublicBaseURLConflict
		}
		return err
	}
	return nil
}

type publicBaseLookup struct {
	host string
	path string
}

func normalizePublicBaseLookup(raw string) (publicBaseLookup, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return publicBaseLookup{}, fmt.Errorf("tenant public base url must not be empty")
	}

	if !strings.Contains(value, "://") {
		value = "https://" + strings.TrimLeft(value, "/")
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return publicBaseLookup{}, fmt.Errorf("parse tenant public base lookup: %w", err)
	}
	if parsed.Host == "" {
		return publicBaseLookup{}, fmt.Errorf("tenant public base url must include host")
	}

	host, err := normalizeLookupHost(parsed.Host)
	if err != nil {
		return publicBaseLookup{}, err
	}
	path := normalizePublicBasePath(parsed.EscapedPath())
	return publicBaseLookup{host: host, path: path}, nil
}

func normalizePublicBasePath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	return "/" + strings.Trim(strings.TrimSpace(trimmed), "/")
}

func publicBasePathMatches(basePath, requestPath string) bool {
	base := normalizePublicBasePath(basePath)
	request := normalizePublicBasePath(requestPath)
	if base == "/" {
		return true
	}
	if request == base {
		return true
	}
	return strings.HasPrefix(request, base+"/")
}

func normalizeSettingsInput(input TenantSettingsInput) (TenantSettingsInput, error) {
	settings := TenantSettingsInput{
		SenderEmail:          strings.TrimSpace(input.SenderEmail),
		SenderName:           strings.TrimSpace(input.SenderName),
		PayPalMode:           firstNonEmpty(input.PayPalMode, DefaultPayPalMode),
		PayPalClientID:       strings.TrimSpace(input.PayPalClientID),
		PayPalMerchantID:     strings.TrimSpace(input.PayPalMerchantID),
		DefaultRetentionDays: input.DefaultRetentionDays,
		SettingsJSON:         strings.TrimSpace(input.SettingsJSON),
	}

	if settings.DefaultRetentionDays <= 0 {
		settings.DefaultRetentionDays = DefaultRetentionDays
	}
	if settings.DefaultRetentionDays <= 0 {
		return TenantSettingsInput{}, fmt.Errorf("default retention days must be > 0")
	}

	if err := validatePayPalMode(settings.PayPalMode); err != nil {
		return TenantSettingsInput{}, err
	}
	return settings, nil
}

func validatePayPalMode(raw string) error {
	mode := strings.ToLower(strings.TrimSpace(raw))
	switch mode {
	case "disabled", "sandbox", "live":
		return nil
	default:
		return fmt.Errorf("unsupported paypal mode %q", raw)
	}
}

func firstNonEmpty(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func defaultID(prefix string) string {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, random)
}
