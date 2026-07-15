package event

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultPublicEventLimit = 20
	MaxPublicEventLimit     = 100
)

type PublicEventFilter struct {
	Limit       int
	IncludePast bool
	SeriesSlug  string
	Mode        string
	From        *time.Time
	To          *time.Time
}

type PublicEvent struct {
	Event
	SeriesSlug  string
	SeriesTitle string
}

func (r *Repository) ListPublicEvents(ctx context.Context, tenantID string, filter PublicEventFilter) ([]PublicEvent, error) {
	if r.db == nil {
		return nil, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}

	normalized, err := normalizePublicEventFilter(filter)
	if err != nil {
		return nil, err
	}

	query := strings.Builder{}
	query.WriteString(`SELECT e.id, e.tenant_id, COALESCE(e.series_id, ''), e.slug, e.title, COALESCE(e.subtitle, ''), COALESCE(e.description, ''),
      e.starts_at, COALESCE(e.ends_at, ''), e.timezone, COALESCE(e.location_name, ''), COALESCE(e.address, ''), COALESCE(e.online_url, ''),
      e.participation_mode, e.status, e.is_public, COALESCE(e.published_at, ''), e.registration_enabled, e.waitlist_enabled, e.max_participants,
      COALESCE(e.change_note, ''), COALESCE(e.cancelled_reason, ''), e.created_at, e.updated_at,
      COALESCE(s.slug, ''), COALESCE(s.title, '')
    FROM events e
    LEFT JOIN event_series s ON s.tenant_id = e.tenant_id AND s.id = e.series_id AND s.is_public = 1
    WHERE e.tenant_id = ? AND e.is_public = 1 AND e.published_at IS NOT NULL AND e.status <> ? AND e.status <> ?`)
	args := []any{tenant, EventStatusDraft, EventStatusArchived}

	if !normalized.IncludePast {
		query.WriteString(` AND e.starts_at >= ?`)
		args = append(args, r.nowFn().UTC().Format(time.RFC3339))
	}
	if normalized.SeriesSlug != "" {
		query.WriteString(` AND EXISTS (
      SELECT 1
      FROM event_series fs
      WHERE fs.tenant_id = e.tenant_id
        AND fs.id = e.series_id
        AND fs.slug = ?
        AND fs.is_public = 1
    )`)
		args = append(args, normalized.SeriesSlug)
	}
	if normalized.Mode != "" {
		query.WriteString(` AND e.participation_mode = ?`)
		args = append(args, normalized.Mode)
	}
	if normalized.From != nil {
		query.WriteString(` AND e.starts_at >= ?`)
		args = append(args, normalized.From.UTC().Format(time.RFC3339))
	}
	if normalized.To != nil {
		query.WriteString(` AND e.starts_at <= ?`)
		args = append(args, normalized.To.UTC().Format(time.RFC3339))
	}

	query.WriteString(` ORDER BY e.starts_at ASC, e.created_at DESC LIMIT ?`)
	args = append(args, normalized.Limit)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list public events: %w", err)
	}
	defer rows.Close()

	items := make([]PublicEvent, 0)
	for rows.Next() {
		item, scanErr := scanPublicEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public events rows: %w", err)
	}
	return items, nil
}

func (r *Repository) GetPublicEventBySlug(ctx context.Context, tenantID, eventSlug string) (PublicEvent, error) {
	if r.db == nil {
		return PublicEvent{}, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return PublicEvent{}, fmt.Errorf("tenant id must not be empty")
	}
	slug, err := normalizeEventSlug(eventSlug)
	if err != nil {
		return PublicEvent{}, err
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT e.id, e.tenant_id, COALESCE(e.series_id, ''), e.slug, e.title, COALESCE(e.subtitle, ''), COALESCE(e.description, ''),
          e.starts_at, COALESCE(e.ends_at, ''), e.timezone, COALESCE(e.location_name, ''), COALESCE(e.address, ''), COALESCE(e.online_url, ''),
          e.participation_mode, e.status, e.is_public, COALESCE(e.published_at, ''), e.registration_enabled, e.waitlist_enabled, e.max_participants,
          COALESCE(e.change_note, ''), COALESCE(e.cancelled_reason, ''), e.created_at, e.updated_at,
          COALESCE(s.slug, ''), COALESCE(s.title, '')
     FROM events e
     LEFT JOIN event_series s ON s.tenant_id = e.tenant_id AND s.id = e.series_id AND s.is_public = 1
     WHERE e.tenant_id = ? AND e.slug = ? AND e.is_public = 1 AND e.published_at IS NOT NULL AND e.status <> ? AND e.status <> ?
     LIMIT 1`,
		tenant,
		slug,
		EventStatusDraft,
		EventStatusArchived,
	)
	item, err := scanPublicEvent(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PublicEvent{}, ErrEventNotFound
		}
		return PublicEvent{}, err
	}
	return item, nil
}

func (r *Repository) ListPublicSeries(ctx context.Context, tenantID string) ([]EventSeries, error) {
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
     WHERE tenant_id = ? AND is_public = 1
     ORDER BY title ASC, created_at DESC`,
		tenant,
	)
	if err != nil {
		return nil, fmt.Errorf("list public event series: %w", err)
	}
	defer rows.Close()

	items := make([]EventSeries, 0)
	for rows.Next() {
		item, scanErr := scanSeries(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate public event series rows: %w", err)
	}
	return items, nil
}

func (r *Repository) GetPublicSeriesBySlug(ctx context.Context, tenantID, seriesSlug string) (EventSeries, error) {
	if r.db == nil {
		return EventSeries{}, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return EventSeries{}, fmt.Errorf("tenant id must not be empty")
	}
	slug, err := normalizeSeriesSlug(seriesSlug)
	if err != nil {
		return EventSeries{}, err
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, slug, title, COALESCE(description, ''),
          COALESCE(default_location_name, ''), COALESCE(default_address, ''), COALESCE(default_online_url, ''),
          is_public, created_at, updated_at
     FROM event_series
     WHERE tenant_id = ? AND slug = ? AND is_public = 1
     LIMIT 1`,
		tenant,
		slug,
	)
	series, err := scanSeries(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EventSeries{}, ErrSeriesNotFound
		}
		return EventSeries{}, err
	}
	return series, nil
}

func normalizePublicEventFilter(filter PublicEventFilter) (PublicEventFilter, error) {
	normalized := filter

	limit := normalized.Limit
	if limit <= 0 {
		limit = DefaultPublicEventLimit
	}
	if limit > MaxPublicEventLimit {
		return PublicEventFilter{}, fmt.Errorf("public events limit must be <= %d", MaxPublicEventLimit)
	}
	normalized.Limit = limit

	if strings.TrimSpace(normalized.SeriesSlug) != "" {
		slug, err := normalizeSeriesSlug(normalized.SeriesSlug)
		if err != nil {
			return PublicEventFilter{}, err
		}
		normalized.SeriesSlug = slug
	}

	if strings.TrimSpace(normalized.Mode) != "" {
		mode, err := normalizeParticipationMode(normalized.Mode)
		if err != nil {
			return PublicEventFilter{}, err
		}
		normalized.Mode = mode
	}

	if normalized.From != nil {
		from := normalized.From.UTC()
		normalized.From = &from
	}
	if normalized.To != nil {
		to := normalized.To.UTC()
		normalized.To = &to
	}
	if normalized.From != nil && normalized.To != nil && normalized.From.After(*normalized.To) {
		return PublicEventFilter{}, fmt.Errorf("public events from must be <= to")
	}

	return normalized, nil
}

func scanPublicEvent(row rowScanner) (PublicEvent, error) {
	var (
		item                PublicEvent
		startsAtRaw         string
		endsAtRaw           string
		isPublicInt         int
		publishedAtRaw      string
		registrationEnabled int
		waitlistEnabled     int
		maxParticipantsRaw  sql.NullInt64
		createdAtRaw        string
		updatedAtRaw        string
	)

	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.SeriesID,
		&item.Slug,
		&item.Title,
		&item.Subtitle,
		&item.Description,
		&startsAtRaw,
		&endsAtRaw,
		&item.Timezone,
		&item.LocationName,
		&item.Address,
		&item.OnlineURL,
		&item.ParticipationMode,
		&item.Status,
		&isPublicInt,
		&publishedAtRaw,
		&registrationEnabled,
		&waitlistEnabled,
		&maxParticipantsRaw,
		&item.ChangeNote,
		&item.CancelledReason,
		&createdAtRaw,
		&updatedAtRaw,
		&item.SeriesSlug,
		&item.SeriesTitle,
	); err != nil {
		return PublicEvent{}, err
	}

	startsAt, err := time.Parse(time.RFC3339, startsAtRaw)
	if err != nil {
		return PublicEvent{}, fmt.Errorf("parse public event starts_at: %w", err)
	}
	var endsAt *time.Time
	if strings.TrimSpace(endsAtRaw) != "" {
		parsedEndsAt, parseErr := time.Parse(time.RFC3339, endsAtRaw)
		if parseErr != nil {
			return PublicEvent{}, fmt.Errorf("parse public event ends_at: %w", parseErr)
		}
		parsedEndsAt = parsedEndsAt.UTC()
		endsAt = &parsedEndsAt
	}
	var publishedAt *time.Time
	if strings.TrimSpace(publishedAtRaw) != "" {
		parsedPublishedAt, parseErr := time.Parse(time.RFC3339, publishedAtRaw)
		if parseErr != nil {
			return PublicEvent{}, fmt.Errorf("parse public event published_at: %w", parseErr)
		}
		parsedPublishedAt = parsedPublishedAt.UTC()
		publishedAt = &parsedPublishedAt
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return PublicEvent{}, fmt.Errorf("parse public event created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return PublicEvent{}, fmt.Errorf("parse public event updated_at: %w", err)
	}

	item.StartsAt = startsAt.UTC()
	item.EndsAt = endsAt
	item.IsPublic = isPublicInt == 1
	item.PublishedAt = publishedAt
	item.RegistrationEnabled = registrationEnabled == 1
	item.WaitlistEnabled = waitlistEnabled == 1
	if maxParticipantsRaw.Valid {
		value := int(maxParticipantsRaw.Int64)
		item.MaxParticipants = &value
	}
	item.CreatedAt = createdAt.UTC()
	item.UpdatedAt = updatedAt.UTC()
	return item, nil
}
