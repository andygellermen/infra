package snippet

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	DefaultViewType = "cards"
)

var (
	ErrSnippetNotFound   = errors.New("snippet config not found")
	ErrSnippetSlugExists = errors.New("snippet slug already exists")
)

var (
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	viewTypes   = map[string]struct{}{
		"cards":   {},
		"list":    {},
		"table":   {},
		"minimal": {},
		"button":  {},
	}
)

type Config struct {
	ID                 string
	TenantID           string
	Name               string
	Slug               string
	ViewType           string
	EventFilterJSON    string
	DisplayOptionsJSON string
	IsActive           bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateConfigParams struct {
	Name               string
	Slug               string
	ViewType           string
	EventFilterJSON    string
	DisplayOptionsJSON string
	IsActive           *bool
}

type UpdateConfigParams struct {
	Name               *string
	Slug               *string
	ViewType           *string
	EventFilterJSON    *string
	DisplayOptionsJSON *string
	IsActive           *bool
}

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

func (r *Repository) ListConfigs(ctx context.Context, tenantID string) ([]Config, error) {
	if r.db == nil {
		return nil, fmt.Errorf("snippet repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, name, slug, view_type, COALESCE(event_filter_json, ''), COALESCE(display_options_json, ''),
            is_active, created_at, updated_at
     FROM snippet_configs
     WHERE tenant_id = ?
     ORDER BY updated_at DESC, created_at DESC`,
		tenant,
	)
	if err != nil {
		return nil, fmt.Errorf("list snippet configs: %w", err)
	}
	defer rows.Close()

	items := make([]Config, 0)
	for rows.Next() {
		item, scanErr := scanConfig(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate snippet configs: %w", err)
	}
	return items, nil
}

func (r *Repository) CreateConfig(ctx context.Context, tenantID string, params CreateConfigParams) (Config, error) {
	if r.db == nil {
		return Config{}, fmt.Errorf("snippet repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Config{}, fmt.Errorf("tenant id must not be empty")
	}

	input, err := normalizeCreateParams(params)
	if err != nil {
		return Config{}, err
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	configID := r.idFn("snp")
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO snippet_configs (
      id, tenant_id, name, slug, view_type, event_filter_json, display_options_json, is_active, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		configID,
		tenant,
		input.Name,
		input.Slug,
		input.ViewType,
		nullable(input.EventFilterJSON),
		nullable(input.DisplayOptionsJSON),
		boolToInt(input.IsActive),
		now,
		now,
	)
	if err != nil {
		if isSlugConstraintError(err) {
			return Config{}, ErrSnippetSlugExists
		}
		return Config{}, fmt.Errorf("insert snippet config: %w", err)
	}

	return r.GetConfigByID(ctx, tenant, configID)
}

func (r *Repository) GetConfigByID(ctx context.Context, tenantID, configID string) (Config, error) {
	if r.db == nil {
		return Config{}, fmt.Errorf("snippet repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Config{}, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(configID)
	if id == "" {
		return Config{}, fmt.Errorf("snippet config id must not be empty")
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, name, slug, view_type, COALESCE(event_filter_json, ''), COALESCE(display_options_json, ''),
            is_active, created_at, updated_at
     FROM snippet_configs
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenant,
		id,
	)
	return scanConfig(row)
}

func (r *Repository) GetConfigBySlug(ctx context.Context, tenantID, configSlug string) (Config, error) {
	if r.db == nil {
		return Config{}, fmt.Errorf("snippet repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Config{}, fmt.Errorf("tenant id must not be empty")
	}
	slug, err := normalizeSlug(configSlug)
	if err != nil {
		return Config{}, err
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, name, slug, view_type, COALESCE(event_filter_json, ''), COALESCE(display_options_json, ''),
            is_active, created_at, updated_at
     FROM snippet_configs
     WHERE tenant_id = ? AND slug = ?
     LIMIT 1`,
		tenant,
		slug,
	)
	return scanConfig(row)
}

func (r *Repository) UpdateConfig(ctx context.Context, tenantID, configID string, params UpdateConfigParams) (Config, error) {
	if r.db == nil {
		return Config{}, fmt.Errorf("snippet repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Config{}, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(configID)
	if id == "" {
		return Config{}, fmt.Errorf("snippet config id must not be empty")
	}

	current, err := r.GetConfigByID(ctx, tenant, id)
	if err != nil {
		return Config{}, err
	}

	updated, hasChange, err := applyUpdate(current, params)
	if err != nil {
		return Config{}, err
	}
	if !hasChange {
		return Config{}, fmt.Errorf("at least one field must be set for update")
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE snippet_configs
     SET name = ?, slug = ?, view_type = ?, event_filter_json = ?, display_options_json = ?, is_active = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		updated.Name,
		updated.Slug,
		updated.ViewType,
		nullable(updated.EventFilterJSON),
		nullable(updated.DisplayOptionsJSON),
		boolToInt(updated.IsActive),
		now,
		tenant,
		id,
	)
	if err != nil {
		if isSlugConstraintError(err) {
			return Config{}, ErrSnippetSlugExists
		}
		return Config{}, fmt.Errorf("update snippet config: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return Config{}, fmt.Errorf("read snippet update rows affected: %w", err)
	}
	if affected == 0 {
		return Config{}, ErrSnippetNotFound
	}

	return r.GetConfigByID(ctx, tenant, id)
}

func (r *Repository) DeleteConfig(ctx context.Context, tenantID, configID string) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("snippet repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return false, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(configID)
	if id == "" {
		return false, fmt.Errorf("snippet config id must not be empty")
	}

	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM snippet_configs WHERE tenant_id = ? AND id = ?`,
		tenant,
		id,
	)
	if err != nil {
		return false, fmt.Errorf("delete snippet config: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read snippet delete rows affected: %w", err)
	}
	return affected > 0, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanConfig(row rowScanner) (Config, error) {
	var (
		item        Config
		isActiveInt int
		createdRaw  string
		updatedRaw  string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.Name,
		&item.Slug,
		&item.ViewType,
		&item.EventFilterJSON,
		&item.DisplayOptionsJSON,
		&isActiveInt,
		&createdRaw,
		&updatedRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Config{}, ErrSnippetNotFound
		}
		return Config{}, fmt.Errorf("scan snippet config: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, createdRaw)
	if err != nil {
		return Config{}, fmt.Errorf("parse snippet created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedRaw)
	if err != nil {
		return Config{}, fmt.Errorf("parse snippet updated_at: %w", err)
	}

	item.IsActive = isActiveInt == 1
	item.CreatedAt = createdAt.UTC()
	item.UpdatedAt = updatedAt.UTC()
	return item, nil
}

func normalizeCreateParams(params CreateConfigParams) (Config, error) {
	name, err := normalizeName(params.Name)
	if err != nil {
		return Config{}, err
	}
	slug, err := normalizeSlug(params.Slug)
	if err != nil {
		return Config{}, err
	}
	viewType, err := normalizeViewType(params.ViewType)
	if err != nil {
		return Config{}, err
	}
	eventFilterJSON, err := normalizeOptionalJSONObject(params.EventFilterJSON, "event_filter_json")
	if err != nil {
		return Config{}, err
	}
	displayOptionsJSON, err := normalizeOptionalJSONObject(params.DisplayOptionsJSON, "display_options_json")
	if err != nil {
		return Config{}, err
	}
	isActive := true
	if params.IsActive != nil {
		isActive = *params.IsActive
	}

	return Config{
		Name:               name,
		Slug:               slug,
		ViewType:           viewType,
		EventFilterJSON:    eventFilterJSON,
		DisplayOptionsJSON: displayOptionsJSON,
		IsActive:           isActive,
	}, nil
}

func applyUpdate(current Config, params UpdateConfigParams) (Config, bool, error) {
	updated := current
	hasChange := false

	if params.Name != nil {
		name, err := normalizeName(*params.Name)
		if err != nil {
			return Config{}, false, err
		}
		updated.Name = name
		hasChange = true
	}
	if params.Slug != nil {
		slug, err := normalizeSlug(*params.Slug)
		if err != nil {
			return Config{}, false, err
		}
		updated.Slug = slug
		hasChange = true
	}
	if params.ViewType != nil {
		viewType, err := normalizeViewType(*params.ViewType)
		if err != nil {
			return Config{}, false, err
		}
		updated.ViewType = viewType
		hasChange = true
	}
	if params.EventFilterJSON != nil {
		normalized, err := normalizeOptionalJSONObject(*params.EventFilterJSON, "event_filter_json")
		if err != nil {
			return Config{}, false, err
		}
		updated.EventFilterJSON = normalized
		hasChange = true
	}
	if params.DisplayOptionsJSON != nil {
		normalized, err := normalizeOptionalJSONObject(*params.DisplayOptionsJSON, "display_options_json")
		if err != nil {
			return Config{}, false, err
		}
		updated.DisplayOptionsJSON = normalized
		hasChange = true
	}
	if params.IsActive != nil {
		updated.IsActive = *params.IsActive
		hasChange = true
	}

	return updated, hasChange, nil
}

func normalizeName(value string) (string, error) {
	name := strings.TrimSpace(value)
	if name == "" {
		return "", fmt.Errorf("snippet name must not be empty")
	}
	if len(name) > 160 {
		return "", fmt.Errorf("snippet name must be <= 160 characters")
	}
	return name, nil
}

func normalizeSlug(value string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(value))
	if slug == "" {
		return "", fmt.Errorf("snippet slug must not be empty")
	}
	if len(slug) > 120 {
		return "", fmt.Errorf("snippet slug must be <= 120 characters")
	}
	if !slugPattern.MatchString(slug) {
		return "", fmt.Errorf("snippet slug %q must match %s", value, slugPattern.String())
	}
	return slug, nil
}

func normalizeViewType(value string) (string, error) {
	viewType := strings.ToLower(strings.TrimSpace(value))
	if viewType == "" {
		viewType = DefaultViewType
	}
	if _, ok := viewTypes[viewType]; !ok {
		return "", fmt.Errorf("snippet view_type %q is not supported", value)
	}
	return viewType, nil
}

func normalizeOptionalJSONObject(raw, fieldName string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	if !json.Valid([]byte(value)) {
		return "", fmt.Errorf("%s must be valid JSON", fieldName)
	}

	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.UseNumber()
	var payload map[string]any
	if err := decoder.Decode(&payload); err != nil {
		return "", fmt.Errorf("%s must be a JSON object", fieldName)
	}
	if payload == nil {
		return "", fmt.Errorf("%s must be a JSON object", fieldName)
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", fieldName, err)
	}
	return string(normalized), nil
}

func boolToInt(flag bool) int {
	if flag {
		return 1
	}
	return 0
}

func nullable(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func isSlugConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: snippet_configs.tenant_id, snippet_configs.slug")
}

func defaultID(prefix string) string {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, random)
}
