package event

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrSeriesNotFound   = errors.New("event series not found")
	ErrSeriesSlugExists = errors.New("event series slug already exists")
)

var seriesSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type EventSeries struct {
	ID                  string
	TenantID            string
	Slug                string
	Title               string
	Description         string
	DefaultLocationName string
	DefaultAddress      string
	DefaultOnlineURL    string
	IsPublic            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type CreateSeriesParams struct {
	Slug                string
	Title               string
	Description         string
	DefaultLocationName string
	DefaultAddress      string
	DefaultOnlineURL    string
	IsPublic            *bool
}

type UpdateSeriesParams struct {
	Slug                *string
	Title               *string
	Description         *string
	DefaultLocationName *string
	DefaultAddress      *string
	DefaultOnlineURL    *string
	IsPublic            *bool
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

func (r *Repository) ListSeries(ctx context.Context, tenantID string) ([]EventSeries, error) {
	if r.db == nil {
		return nil, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, slug, title, COALESCE(description, ''),
            COALESCE(default_location_name, ''), COALESCE(default_address, ''), COALESCE(default_online_url, ''),
            is_public, created_at, updated_at
     FROM event_series
     WHERE tenant_id = ?
     ORDER BY created_at DESC`,
		tenant,
	)
	if err != nil {
		return nil, fmt.Errorf("list event series: %w", err)
	}
	defer rows.Close()

	series := make([]EventSeries, 0)
	for rows.Next() {
		item, scanErr := scanSeries(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		series = append(series, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate event series rows: %w", err)
	}
	return series, nil
}

func (r *Repository) CreateSeries(ctx context.Context, tenantID string, params CreateSeriesParams) (EventSeries, error) {
	if r.db == nil {
		return EventSeries{}, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return EventSeries{}, fmt.Errorf("tenant id must not be empty")
	}

	input, err := normalizeCreateParams(params)
	if err != nil {
		return EventSeries{}, err
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	seriesID := r.idFn("srs")
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO event_series (
      id, tenant_id, slug, title, description, default_location_name, default_address, default_online_url,
      is_public, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		seriesID,
		tenant,
		input.Slug,
		input.Title,
		nullable(input.Description),
		nullable(input.DefaultLocationName),
		nullable(input.DefaultAddress),
		nullable(input.DefaultOnlineURL),
		boolToInt(input.IsPublic),
		now,
		now,
	)
	if err != nil {
		if isSlugConstraintError(err) {
			return EventSeries{}, ErrSeriesSlugExists
		}
		return EventSeries{}, fmt.Errorf("insert event series: %w", err)
	}

	return r.GetSeriesByID(ctx, tenant, seriesID)
}

func (r *Repository) GetSeriesByID(ctx context.Context, tenantID, seriesID string) (EventSeries, error) {
	if r.db == nil {
		return EventSeries{}, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return EventSeries{}, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(seriesID)
	if id == "" {
		return EventSeries{}, fmt.Errorf("event series id must not be empty")
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, slug, title, COALESCE(description, ''),
            COALESCE(default_location_name, ''), COALESCE(default_address, ''), COALESCE(default_online_url, ''),
            is_public, created_at, updated_at
     FROM event_series
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenant,
		id,
	)
	item, err := scanSeries(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EventSeries{}, ErrSeriesNotFound
		}
		return EventSeries{}, err
	}
	return item, nil
}

func (r *Repository) UpdateSeries(ctx context.Context, tenantID, seriesID string, params UpdateSeriesParams) (EventSeries, error) {
	if r.db == nil {
		return EventSeries{}, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return EventSeries{}, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(seriesID)
	if id == "" {
		return EventSeries{}, fmt.Errorf("event series id must not be empty")
	}

	current, err := r.GetSeriesByID(ctx, tenant, id)
	if err != nil {
		return EventSeries{}, err
	}

	updated, hasChange, err := applyUpdate(current, params)
	if err != nil {
		return EventSeries{}, err
	}
	if !hasChange {
		return EventSeries{}, fmt.Errorf("at least one field must be set for update")
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE event_series
     SET slug = ?, title = ?, description = ?, default_location_name = ?, default_address = ?,
         default_online_url = ?, is_public = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		updated.Slug,
		updated.Title,
		nullable(updated.Description),
		nullable(updated.DefaultLocationName),
		nullable(updated.DefaultAddress),
		nullable(updated.DefaultOnlineURL),
		boolToInt(updated.IsPublic),
		now,
		tenant,
		id,
	)
	if err != nil {
		if isSlugConstraintError(err) {
			return EventSeries{}, ErrSeriesSlugExists
		}
		return EventSeries{}, fmt.Errorf("update event series: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return EventSeries{}, fmt.Errorf("event series update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return EventSeries{}, ErrSeriesNotFound
	}

	return r.GetSeriesByID(ctx, tenant, id)
}

func (r *Repository) DeleteSeries(ctx context.Context, tenantID, seriesID string) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return false, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(seriesID)
	if id == "" {
		return false, fmt.Errorf("event series id must not be empty")
	}

	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM event_series WHERE tenant_id = ? AND id = ?`,
		tenant,
		id,
	)
	if err != nil {
		return false, fmt.Errorf("delete event series: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("event series delete rows affected: %w", err)
	}
	return rowsAffected == 1, nil
}

type normalizedCreate struct {
	Slug                string
	Title               string
	Description         string
	DefaultLocationName string
	DefaultAddress      string
	DefaultOnlineURL    string
	IsPublic            bool
}

func normalizeCreateParams(params CreateSeriesParams) (normalizedCreate, error) {
	slug, err := normalizeSeriesSlug(params.Slug)
	if err != nil {
		return normalizedCreate{}, err
	}
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return normalizedCreate{}, fmt.Errorf("event series title must not be empty")
	}
	if len(title) > 180 {
		return normalizedCreate{}, fmt.Errorf("event series title must be <= 180 characters")
	}

	defaultOnlineURL, err := normalizeOptionalURL(params.DefaultOnlineURL)
	if err != nil {
		return normalizedCreate{}, err
	}

	isPublic := true
	if params.IsPublic != nil {
		isPublic = *params.IsPublic
	}

	return normalizedCreate{
		Slug:                slug,
		Title:               title,
		Description:         strings.TrimSpace(params.Description),
		DefaultLocationName: strings.TrimSpace(params.DefaultLocationName),
		DefaultAddress:      strings.TrimSpace(params.DefaultAddress),
		DefaultOnlineURL:    defaultOnlineURL,
		IsPublic:            isPublic,
	}, nil
}

func applyUpdate(current EventSeries, params UpdateSeriesParams) (EventSeries, bool, error) {
	updated := current
	hasChange := false

	if params.Slug != nil {
		slug, err := normalizeSeriesSlug(*params.Slug)
		if err != nil {
			return EventSeries{}, false, err
		}
		updated.Slug = slug
		hasChange = true
	}
	if params.Title != nil {
		title := strings.TrimSpace(*params.Title)
		if title == "" {
			return EventSeries{}, false, fmt.Errorf("event series title must not be empty")
		}
		if len(title) > 180 {
			return EventSeries{}, false, fmt.Errorf("event series title must be <= 180 characters")
		}
		updated.Title = title
		hasChange = true
	}
	if params.Description != nil {
		updated.Description = strings.TrimSpace(*params.Description)
		hasChange = true
	}
	if params.DefaultLocationName != nil {
		updated.DefaultLocationName = strings.TrimSpace(*params.DefaultLocationName)
		hasChange = true
	}
	if params.DefaultAddress != nil {
		updated.DefaultAddress = strings.TrimSpace(*params.DefaultAddress)
		hasChange = true
	}
	if params.DefaultOnlineURL != nil {
		normalizedURL, err := normalizeOptionalURL(*params.DefaultOnlineURL)
		if err != nil {
			return EventSeries{}, false, err
		}
		updated.DefaultOnlineURL = normalizedURL
		hasChange = true
	}
	if params.IsPublic != nil {
		updated.IsPublic = *params.IsPublic
		hasChange = true
	}

	return updated, hasChange, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSeries(row rowScanner) (EventSeries, error) {
	var (
		item         EventSeries
		isPublicInt  int
		createdAtRaw string
		updatedAtRaw string
	)

	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.Slug,
		&item.Title,
		&item.Description,
		&item.DefaultLocationName,
		&item.DefaultAddress,
		&item.DefaultOnlineURL,
		&isPublicInt,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		return EventSeries{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return EventSeries{}, fmt.Errorf("parse event series created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return EventSeries{}, fmt.Errorf("parse event series updated_at: %w", err)
	}
	item.IsPublic = isPublicInt == 1
	item.CreatedAt = createdAt
	item.UpdatedAt = updatedAt
	return item, nil
}

func normalizeSeriesSlug(raw string) (string, error) {
	slug := strings.ToLower(strings.Trim(strings.TrimSpace(raw), "/"))
	if slug == "" {
		return "", fmt.Errorf("event series slug must not be empty")
	}
	if len(slug) > 120 {
		return "", fmt.Errorf("event series slug must be <= 120 characters")
	}
	if !seriesSlugPattern.MatchString(slug) {
		return "", fmt.Errorf("event series slug %q is invalid", raw)
	}
	return slug, nil
}

func normalizeOptionalURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse event series online url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("event series online url must include scheme and host")
	}
	return strings.TrimSpace(value), nil
}

func isSlugConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: event_series.tenant_id, event_series.slug")
}

func nullable(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func boolToInt(flag bool) int {
	if flag {
		return 1
	}
	return 0
}

func defaultID(prefix string) string {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, random)
}
