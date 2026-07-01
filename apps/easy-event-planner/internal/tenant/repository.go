package tenant

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrTenantNotFound         = errors.New("tenant not found")
	ErrTenantSettingsNotFound = errors.New("tenant settings not found")
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
