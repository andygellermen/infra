package event

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	EventStatusDraft     = "draft"
	EventStatusScheduled = "scheduled"
	EventStatusChanged   = "changed"
	EventStatusPostponed = "postponed"
	EventStatusCancelled = "cancelled"
	EventStatusCompleted = "completed"
	EventStatusArchived  = "archived"

	ParticipationModeOnsite = "onsite"
	ParticipationModeOnline = "online"
	ParticipationModeHybrid = "hybrid"
)

var (
	ErrEventNotFound            = errors.New("event not found")
	ErrEventSlugExists          = errors.New("event slug already exists")
	ErrInvalidStatusTransition  = errors.New("invalid event status transition")
	ErrEventSeriesScopeMismatch = errors.New("event series does not belong to tenant")
)

var eventSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Event struct {
	ID                     string
	TenantID               string
	SeriesID               string
	Slug                   string
	Title                  string
	Subtitle               string
	Description            string
	StartsAt               time.Time
	EndsAt                 *time.Time
	Timezone               string
	LocationName           string
	Address                string
	OnlineURL              string
	ParticipationMode      string
	Status                 string
	IsPublic               bool
	PublishedAt            *time.Time
	PublicVisibleFrom      *time.Time
	RegistrationOpensAt    *time.Time
	RegistrationClosesAt   *time.Time
	RegistrationEnabled    bool
	WaitlistEnabled        bool
	MaxParticipants        *int
	ConfirmedParticipants  int
	WaitlistEntries        int
	TicketName             string
	PriceCents             int
	Currency               string
	DonationEnabled        bool
	DonationMinCents       *int
	DonationSuggestedCents *int
	ChangeNote             string
	CancelledReason        string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type CreateEventParams struct {
	SeriesID               string
	Slug                   string
	Title                  string
	Subtitle               string
	Description            string
	StartsAt               string
	EndsAt                 string
	Timezone               string
	LocationName           string
	Address                string
	OnlineURL              string
	ParticipationMode      string
	IsPublic               *bool
	PublicVisibleFrom      string
	RegistrationOpensAt    string
	RegistrationClosesAt   string
	RegistrationEnabled    *bool
	WaitlistEnabled        *bool
	MaxParticipants        *int
	TicketName             string
	PriceCents             *int
	Currency               string
	DonationEnabled        *bool
	DonationMinCents       *int
	DonationSuggestedCents *int
}

type UpdateEventParams struct {
	SeriesID                    *string
	Slug                        *string
	Title                       *string
	Subtitle                    *string
	Description                 *string
	StartsAt                    *string
	EndsAt                      *string
	Timezone                    *string
	LocationName                *string
	Address                     *string
	OnlineURL                   *string
	ParticipationMode           *string
	IsPublic                    *bool
	PublicVisibleFrom           *string
	RegistrationOpensAt         *string
	RegistrationClosesAt        *string
	RegistrationEnabled         *bool
	WaitlistEnabled             *bool
	MaxParticipants             *int
	ClearMaxParticipants        bool
	TicketName                  *string
	PriceCents                  *int
	Currency                    *string
	DonationEnabled             *bool
	DonationMinCents            *int
	ClearDonationMinCents       bool
	DonationSuggestedCents      *int
	ClearDonationSuggestedCents bool
	ChangeNote                  *string
	CancelledReason             *string
}

func (r *Repository) ListEvents(ctx context.Context, tenantID string) ([]Event, error) {
	if r.db == nil {
		return nil, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, COALESCE(series_id, ''), slug, title, COALESCE(subtitle, ''), COALESCE(description, ''),
            starts_at, COALESCE(ends_at, ''), timezone, COALESCE(location_name, ''), COALESCE(address, ''), COALESCE(online_url, ''),
            participation_mode, status, is_public, COALESCE(published_at, ''), COALESCE(public_visible_from, ''),
            COALESCE(registration_opens_at, ''), COALESCE(registration_closes_at, ''), registration_enabled, waitlist_enabled, max_participants,
            (SELECT COUNT(*)
             FROM registrations r
             WHERE r.tenant_id = events.tenant_id
               AND r.event_id = events.id
               AND r.status = 'confirmed') AS confirmed_count,
            (SELECT COUNT(*)
             FROM waitlist_entries w
             WHERE w.tenant_id = events.tenant_id
               AND w.event_id = events.id
               AND w.status IN ('waiting', 'offered')) AS waitlist_count,
            COALESCE((SELECT name FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1), ''),
            COALESCE((SELECT price_cents FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1), 0),
            COALESCE((SELECT currency FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1), 'EUR'),
            COALESCE((SELECT donation_enabled FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1), 0),
            (SELECT donation_min_cents FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1),
            (SELECT donation_suggested_cents FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1),
            COALESCE(change_note, ''), COALESCE(cancelled_reason, ''), created_at, updated_at
     FROM events
     WHERE tenant_id = ?
     ORDER BY starts_at ASC, created_at DESC`,
		tenant,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]Event, 0)
	for rows.Next() {
		item, scanErr := scanEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		events = append(events, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events rows: %w", err)
	}
	return events, nil
}

func (r *Repository) CreateEvent(ctx context.Context, tenantID string, params CreateEventParams) (Event, error) {
	if r.db == nil {
		return Event{}, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Event{}, fmt.Errorf("tenant id must not be empty")
	}

	normalized, err := normalizeCreateEventParams(params)
	if err != nil {
		return Event{}, err
	}

	seriesID, err := r.normalizeSeriesForTenant(ctx, tenant, normalized.SeriesID)
	if err != nil {
		return Event{}, err
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	eventID := r.idFn("evt")
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Event{}, fmt.Errorf("begin event create transaction: %w", err)
	}
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO events (
      id, tenant_id, series_id, slug, title, subtitle, description, starts_at, ends_at, timezone,
      location_name, address, online_url, participation_mode, status, is_public, published_at, public_visible_from,
      registration_opens_at, registration_closes_at, registration_enabled, waitlist_enabled, max_participants, change_note,
      cancelled_reason, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		eventID,
		tenant,
		nullable(seriesID),
		normalized.Slug,
		normalized.Title,
		nullable(normalized.Subtitle),
		nullable(normalized.Description),
		normalized.StartsAt.Format(time.RFC3339),
		nullableTime(normalized.EndsAt),
		normalized.Timezone,
		nullable(normalized.LocationName),
		nullable(normalized.Address),
		nullable(normalized.OnlineURL),
		normalized.ParticipationMode,
		EventStatusDraft,
		boolToInt(normalized.IsPublic),
		nil,
		nullableTime(normalized.PublicVisibleFrom),
		nullableTime(normalized.RegistrationOpensAt),
		nullableTime(normalized.RegistrationClosesAt),
		boolToInt(normalized.RegistrationEnabled),
		boolToInt(normalized.WaitlistEnabled),
		nullableInt(normalized.MaxParticipants),
		nil,
		nil,
		now,
		now,
	)
	if err != nil {
		_ = tx.Rollback()
		if isEventSlugConstraintError(err) {
			return Event{}, ErrEventSlugExists
		}
		return Event{}, fmt.Errorf("insert event: %w", err)
	}
	if err := r.syncDefaultEventTicketTx(ctx, tx, tenant, eventID, normalized); err != nil {
		_ = tx.Rollback()
		return Event{}, err
	}
	if err := tx.Commit(); err != nil {
		return Event{}, fmt.Errorf("commit event create transaction: %w", err)
	}

	return r.GetEventByID(ctx, tenant, eventID)
}

func (r *Repository) GetEventByID(ctx context.Context, tenantID, eventID string) (Event, error) {
	if r.db == nil {
		return Event{}, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Event{}, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(eventID)
	if id == "" {
		return Event{}, fmt.Errorf("event id must not be empty")
	}

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, COALESCE(series_id, ''), slug, title, COALESCE(subtitle, ''), COALESCE(description, ''),
            starts_at, COALESCE(ends_at, ''), timezone, COALESCE(location_name, ''), COALESCE(address, ''), COALESCE(online_url, ''),
            participation_mode, status, is_public, COALESCE(published_at, ''), COALESCE(public_visible_from, ''),
            COALESCE(registration_opens_at, ''), COALESCE(registration_closes_at, ''), registration_enabled, waitlist_enabled, max_participants,
            (SELECT COUNT(*)
             FROM registrations r
             WHERE r.tenant_id = events.tenant_id
               AND r.event_id = events.id
               AND r.status = 'confirmed') AS confirmed_count,
            (SELECT COUNT(*)
             FROM waitlist_entries w
             WHERE w.tenant_id = events.tenant_id
               AND w.event_id = events.id
               AND w.status IN ('waiting', 'offered')) AS waitlist_count,
            COALESCE((SELECT name FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1), ''),
            COALESCE((SELECT price_cents FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1), 0),
            COALESCE((SELECT currency FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1), 'EUR'),
            COALESCE((SELECT donation_enabled FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1), 0),
            (SELECT donation_min_cents FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1),
            (SELECT donation_suggested_cents FROM event_tickets t WHERE t.tenant_id = events.tenant_id AND t.event_id = events.id ORDER BY t.created_at ASC LIMIT 1),
            COALESCE(change_note, ''), COALESCE(cancelled_reason, ''), created_at, updated_at
     FROM events
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenant,
		id,
	)
	item, err := scanEvent(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Event{}, ErrEventNotFound
		}
		return Event{}, err
	}
	return item, nil
}

func (r *Repository) UpdateEvent(ctx context.Context, tenantID, eventID string, params UpdateEventParams) (Event, error) {
	if r.db == nil {
		return Event{}, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Event{}, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(eventID)
	if id == "" {
		return Event{}, fmt.Errorf("event id must not be empty")
	}

	current, err := r.GetEventByID(ctx, tenant, id)
	if err != nil {
		return Event{}, err
	}

	updated, hasChange, err := r.applyEventUpdate(ctx, tenant, current, params)
	if err != nil {
		return Event{}, err
	}
	if !hasChange {
		return Event{}, fmt.Errorf("at least one field must be set for update")
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Event{}, fmt.Errorf("begin event update transaction: %w", err)
	}
	result, err := tx.ExecContext(
		ctx,
		`UPDATE events
     SET series_id = ?, slug = ?, title = ?, subtitle = ?, description = ?, starts_at = ?, ends_at = ?, timezone = ?,
         location_name = ?, address = ?, online_url = ?, participation_mode = ?, is_public = ?, published_at = ?, public_visible_from = ?,
         registration_opens_at = ?, registration_closes_at = ?, registration_enabled = ?, waitlist_enabled = ?, max_participants = ?,
         change_note = ?, cancelled_reason = ?, status = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		nullable(updated.SeriesID),
		updated.Slug,
		updated.Title,
		nullable(updated.Subtitle),
		nullable(updated.Description),
		updated.StartsAt.Format(time.RFC3339),
		nullableTime(updated.EndsAt),
		updated.Timezone,
		nullable(updated.LocationName),
		nullable(updated.Address),
		nullable(updated.OnlineURL),
		updated.ParticipationMode,
		boolToInt(updated.IsPublic),
		nullableTime(updated.PublishedAt),
		nullableTime(updated.PublicVisibleFrom),
		nullableTime(updated.RegistrationOpensAt),
		nullableTime(updated.RegistrationClosesAt),
		boolToInt(updated.RegistrationEnabled),
		boolToInt(updated.WaitlistEnabled),
		nullableInt(updated.MaxParticipants),
		nullable(updated.ChangeNote),
		nullable(updated.CancelledReason),
		updated.Status,
		now,
		tenant,
		id,
	)
	if err != nil {
		_ = tx.Rollback()
		if isEventSlugConstraintError(err) {
			return Event{}, ErrEventSlugExists
		}
		return Event{}, fmt.Errorf("update event: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return Event{}, fmt.Errorf("event update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		_ = tx.Rollback()
		return Event{}, ErrEventNotFound
	}
	if err := r.syncDefaultEventTicketFromEventTx(ctx, tx, updated); err != nil {
		_ = tx.Rollback()
		return Event{}, err
	}
	if err := tx.Commit(); err != nil {
		return Event{}, fmt.Errorf("commit event update transaction: %w", err)
	}

	return r.GetEventByID(ctx, tenant, id)
}

func (r *Repository) DeleteEvent(ctx context.Context, tenantID, eventID string) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("event repository database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return false, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(eventID)
	if id == "" {
		return false, fmt.Errorf("event id must not be empty")
	}

	result, err := r.db.ExecContext(ctx, `DELETE FROM events WHERE tenant_id = ? AND id = ?`, tenant, id)
	if err != nil {
		return false, fmt.Errorf("delete event: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("event delete rows affected: %w", err)
	}
	return rowsAffected == 1, nil
}

func (r *Repository) PublishEvent(ctx context.Context, tenantID, eventID string) (Event, error) {
	return r.transitionPublishState(ctx, tenantID, eventID, true)
}

func (r *Repository) UnpublishEvent(ctx context.Context, tenantID, eventID string) (Event, error) {
	return r.transitionPublishState(ctx, tenantID, eventID, false)
}

func (r *Repository) CancelEvent(ctx context.Context, tenantID, eventID, cancelledReason, changeNote string) (Event, error) {
	current, err := r.GetEventByID(ctx, tenantID, eventID)
	if err != nil {
		return Event{}, err
	}
	if !canCancel(current.Status) {
		return Event{}, ErrInvalidStatusTransition
	}

	reason := strings.TrimSpace(cancelledReason)
	if reason == "" {
		return Event{}, fmt.Errorf("cancelled_reason must not be empty")
	}
	if len(reason) > 1000 {
		return Event{}, fmt.Errorf("cancelled_reason must be <= 1000 characters")
	}
	note := strings.TrimSpace(changeNote)
	if len(note) > 2000 {
		return Event{}, fmt.Errorf("change_note must be <= 2000 characters")
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE events
     SET status = ?, cancelled_reason = ?, change_note = ?, registration_enabled = 0, waitlist_enabled = 0, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		EventStatusCancelled,
		reason,
		nullable(note),
		now,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(eventID),
	)
	if err != nil {
		return Event{}, fmt.Errorf("cancel event: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Event{}, fmt.Errorf("cancel event rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Event{}, ErrEventNotFound
	}
	return r.GetEventByID(ctx, tenantID, eventID)
}

func (r *Repository) PostponeEvent(ctx context.Context, tenantID, eventID, startsAtRaw, endsAtRaw, changeNote string) (Event, error) {
	current, err := r.GetEventByID(ctx, tenantID, eventID)
	if err != nil {
		return Event{}, err
	}
	if !canPostpone(current.Status) {
		return Event{}, ErrInvalidStatusTransition
	}

	startsAt, err := parseRequiredRFC3339(startsAtRaw, "event starts_at")
	if err != nil {
		return Event{}, err
	}
	endsAt, err := parseOptionalRFC3339(endsAtRaw, "event ends_at")
	if err != nil {
		return Event{}, err
	}
	if endsAt != nil && endsAt.Before(startsAt) {
		return Event{}, fmt.Errorf("event ends_at must be >= starts_at")
	}
	note := strings.TrimSpace(changeNote)
	if note == "" {
		return Event{}, fmt.Errorf("change_note must not be empty")
	}
	if len(note) > 2000 {
		return Event{}, fmt.Errorf("change_note must be <= 2000 characters")
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE events
     SET status = ?, starts_at = ?, ends_at = ?, change_note = ?, cancelled_reason = NULL, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		EventStatusPostponed,
		startsAt.UTC().Format(time.RFC3339),
		nullableTime(endsAt),
		note,
		now,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(eventID),
	)
	if err != nil {
		return Event{}, fmt.Errorf("postpone event: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Event{}, fmt.Errorf("postpone event rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Event{}, ErrEventNotFound
	}
	return r.GetEventByID(ctx, tenantID, eventID)
}

func (r *Repository) MarkEventCompleted(ctx context.Context, tenantID, eventID string) (Event, error) {
	current, err := r.GetEventByID(ctx, tenantID, eventID)
	if err != nil {
		return Event{}, err
	}
	if !canMarkCompleted(current.Status) {
		return Event{}, ErrInvalidStatusTransition
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE events
     SET status = ?, registration_enabled = 0, waitlist_enabled = 0, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		EventStatusCompleted,
		now,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(eventID),
	)
	if err != nil {
		return Event{}, fmt.Errorf("mark event completed: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Event{}, fmt.Errorf("mark event completed rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Event{}, ErrEventNotFound
	}
	return r.GetEventByID(ctx, tenantID, eventID)
}

func (r *Repository) ArchiveEvent(ctx context.Context, tenantID, eventID string) (Event, error) {
	current, err := r.GetEventByID(ctx, tenantID, eventID)
	if err != nil {
		return Event{}, err
	}
	if !canArchive(current.Status) {
		return Event{}, ErrInvalidStatusTransition
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE events
     SET status = ?, is_public = 0, published_at = NULL, registration_enabled = 0, waitlist_enabled = 0, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		EventStatusArchived,
		now,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(eventID),
	)
	if err != nil {
		return Event{}, fmt.Errorf("archive event: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Event{}, fmt.Errorf("archive event rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Event{}, ErrEventNotFound
	}
	return r.GetEventByID(ctx, tenantID, eventID)
}

func (r *Repository) transitionPublishState(ctx context.Context, tenantID, eventID string, published bool) (Event, error) {
	current, err := r.GetEventByID(ctx, tenantID, eventID)
	if err != nil {
		return Event{}, err
	}

	if !canTogglePublish(current.Status) {
		return Event{}, ErrInvalidStatusTransition
	}

	targetStatus := current.Status
	if published && strings.EqualFold(strings.TrimSpace(current.Status), EventStatusDraft) {
		targetStatus = EventStatusScheduled
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	var publishedAt any
	targetIsPublic := current.IsPublic
	if published {
		targetIsPublic = true
		publishedAt = now
	} else {
		publishedAt = nil
	}
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE events
     SET status = ?, is_public = ?, published_at = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		targetStatus,
		boolToInt(targetIsPublic),
		publishedAt,
		now,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(eventID),
	)
	if err != nil {
		return Event{}, fmt.Errorf("publish state transition: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Event{}, fmt.Errorf("publish transition rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Event{}, ErrEventNotFound
	}

	return r.GetEventByID(ctx, tenantID, eventID)
}

func canTogglePublish(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case EventStatusCancelled, EventStatusCompleted, EventStatusArchived:
		return false
	default:
		return true
	}
}

func canCancel(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case EventStatusCancelled, EventStatusCompleted, EventStatusArchived:
		return false
	default:
		return true
	}
}

func canPostpone(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case EventStatusScheduled, EventStatusChanged, EventStatusPostponed:
		return true
	default:
		return false
	}
}

func canMarkCompleted(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case EventStatusScheduled, EventStatusChanged, EventStatusPostponed:
		return true
	default:
		return false
	}
}

func canArchive(status string) bool {
	return strings.ToLower(strings.TrimSpace(status)) != EventStatusArchived
}

func canEventAcceptRegistrations(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case EventStatusScheduled, EventStatusChanged, EventStatusPostponed:
		return true
	default:
		return false
	}
}

func (e Event) IsPublished() bool {
	return e.IsPublic && e.PublishedAt != nil && canEventBePublic(e.Status)
}

func (e Event) IsVisible() bool {
	return e.IsVisibleAt(time.Now().UTC())
}

func (e Event) IsVisibleAt(now time.Time) bool {
	if !e.IsPublished() {
		return false
	}
	if e.PublicVisibleFrom == nil {
		return true
	}
	return !now.UTC().Before(e.PublicVisibleFrom.UTC())
}

func (e Event) IsRegistrationOpen() bool {
	return e.IsRegistrationOpenAt(time.Now().UTC())
}

func (e Event) IsRegistrationOpenAt(now time.Time) bool {
	if !e.IsVisibleAt(now) || !e.RegistrationEnabled || !canEventAcceptRegistrations(e.Status) {
		return false
	}
	if e.RegistrationOpensAt != nil && now.UTC().Before(e.RegistrationOpensAt.UTC()) {
		return false
	}
	if e.RegistrationClosesAt != nil && now.UTC().After(e.RegistrationClosesAt.UTC()) {
		return false
	}
	return true
}

func (e Event) PublicationState() string {
	return e.PublicationStateAt(time.Now().UTC())
}

func (e Event) PublicationStateAt(now time.Time) string {
	switch {
	case strings.EqualFold(strings.TrimSpace(e.Status), EventStatusArchived):
		return "archived"
	case e.IsPublished() && !e.IsVisibleAt(now):
		return "scheduled_publication"
	case e.IsPublished():
		return "published"
	case e.IsPublic:
		return "prepared"
	default:
		return "internal"
	}
}

func (r *Repository) applyEventUpdate(ctx context.Context, tenantID string, current Event, params UpdateEventParams) (Event, bool, error) {
	updated := current
	hasChange := false
	contentChanged := false

	if params.SeriesID != nil {
		seriesID, err := r.normalizeSeriesForTenant(ctx, tenantID, *params.SeriesID)
		if err != nil {
			return Event{}, false, err
		}
		updated.SeriesID = seriesID
		hasChange = true
		contentChanged = true
	}
	if params.Slug != nil {
		slug, err := normalizeEventSlug(*params.Slug)
		if err != nil {
			return Event{}, false, err
		}
		updated.Slug = slug
		hasChange = true
		contentChanged = true
	}
	if params.Title != nil {
		title := strings.TrimSpace(*params.Title)
		if title == "" {
			return Event{}, false, fmt.Errorf("event title must not be empty")
		}
		if len(title) > 180 {
			return Event{}, false, fmt.Errorf("event title must be <= 180 characters")
		}
		updated.Title = title
		hasChange = true
		contentChanged = true
	}
	if params.Subtitle != nil {
		updated.Subtitle = strings.TrimSpace(*params.Subtitle)
		hasChange = true
		contentChanged = true
	}
	if params.Description != nil {
		updated.Description = strings.TrimSpace(*params.Description)
		hasChange = true
		contentChanged = true
	}
	if params.StartsAt != nil {
		startsAt, err := parseRequiredRFC3339(*params.StartsAt, "event starts_at")
		if err != nil {
			return Event{}, false, err
		}
		updated.StartsAt = startsAt
		hasChange = true
		contentChanged = true
	}
	if params.EndsAt != nil {
		endsAt, err := parseOptionalRFC3339(*params.EndsAt, "event ends_at")
		if err != nil {
			return Event{}, false, err
		}
		updated.EndsAt = endsAt
		hasChange = true
		contentChanged = true
	}
	if updated.EndsAt != nil && updated.EndsAt.Before(updated.StartsAt) {
		return Event{}, false, fmt.Errorf("event ends_at must be >= starts_at")
	}
	if params.Timezone != nil {
		timezone := strings.TrimSpace(*params.Timezone)
		if timezone == "" {
			return Event{}, false, fmt.Errorf("event timezone must not be empty")
		}
		updated.Timezone = timezone
		hasChange = true
		contentChanged = true
	}
	if params.LocationName != nil {
		updated.LocationName = strings.TrimSpace(*params.LocationName)
		hasChange = true
		contentChanged = true
	}
	if params.Address != nil {
		updated.Address = strings.TrimSpace(*params.Address)
		hasChange = true
		contentChanged = true
	}
	if params.OnlineURL != nil {
		onlineURL, err := normalizeOptionalEventURL(*params.OnlineURL)
		if err != nil {
			return Event{}, false, err
		}
		updated.OnlineURL = onlineURL
		hasChange = true
		contentChanged = true
	}
	if params.ParticipationMode != nil {
		mode, err := normalizeParticipationMode(*params.ParticipationMode)
		if err != nil {
			return Event{}, false, err
		}
		updated.ParticipationMode = mode
		hasChange = true
		contentChanged = true
	}
	if params.IsPublic != nil {
		updated.IsPublic = *params.IsPublic
		if !updated.IsPublic {
			updated.PublishedAt = nil
		}
		hasChange = true
	}
	if params.PublicVisibleFrom != nil {
		visibleFrom, err := parseOptionalRFC3339(*params.PublicVisibleFrom, "event public_visible_from")
		if err != nil {
			return Event{}, false, err
		}
		updated.PublicVisibleFrom = visibleFrom
		hasChange = true
	}
	if params.RegistrationOpensAt != nil {
		opensAt, err := parseOptionalRFC3339(*params.RegistrationOpensAt, "event registration_opens_at")
		if err != nil {
			return Event{}, false, err
		}
		updated.RegistrationOpensAt = opensAt
		hasChange = true
	}
	if params.RegistrationClosesAt != nil {
		closesAt, err := parseOptionalRFC3339(*params.RegistrationClosesAt, "event registration_closes_at")
		if err != nil {
			return Event{}, false, err
		}
		updated.RegistrationClosesAt = closesAt
		hasChange = true
	}
	if params.RegistrationEnabled != nil {
		updated.RegistrationEnabled = *params.RegistrationEnabled
		hasChange = true
	}
	if params.WaitlistEnabled != nil {
		updated.WaitlistEnabled = *params.WaitlistEnabled
		hasChange = true
	}
	if params.ClearMaxParticipants {
		updated.MaxParticipants = nil
		hasChange = true
		contentChanged = true
	}
	if params.MaxParticipants != nil {
		value := *params.MaxParticipants
		if value <= 0 {
			return Event{}, false, fmt.Errorf("event max_participants must be > 0")
		}
		updated.MaxParticipants = &value
		hasChange = true
		contentChanged = true
	}
	if params.TicketName != nil {
		updated.TicketName = strings.TrimSpace(*params.TicketName)
		hasChange = true
		contentChanged = true
	}
	if params.PriceCents != nil {
		if *params.PriceCents < 0 {
			return Event{}, false, fmt.Errorf("event price_cents must be >= 0")
		}
		updated.PriceCents = *params.PriceCents
		hasChange = true
		contentChanged = true
	}
	if params.Currency != nil {
		currency := strings.ToUpper(strings.TrimSpace(*params.Currency))
		if currency == "" {
			currency = "EUR"
		}
		if len(currency) != 3 {
			return Event{}, false, fmt.Errorf("event currency must be a 3-letter code")
		}
		updated.Currency = currency
		hasChange = true
		contentChanged = true
	}
	if params.DonationEnabled != nil {
		updated.DonationEnabled = *params.DonationEnabled
		hasChange = true
		contentChanged = true
	}
	if params.ClearDonationMinCents {
		updated.DonationMinCents = nil
		hasChange = true
		contentChanged = true
	}
	if params.DonationMinCents != nil {
		if *params.DonationMinCents < 0 {
			return Event{}, false, fmt.Errorf("event donation_min_cents must be >= 0")
		}
		value := *params.DonationMinCents
		updated.DonationMinCents = &value
		hasChange = true
		contentChanged = true
	}
	if params.ClearDonationSuggestedCents {
		updated.DonationSuggestedCents = nil
		hasChange = true
		contentChanged = true
	}
	if params.DonationSuggestedCents != nil {
		if *params.DonationSuggestedCents < 0 {
			return Event{}, false, fmt.Errorf("event donation_suggested_cents must be >= 0")
		}
		value := *params.DonationSuggestedCents
		updated.DonationSuggestedCents = &value
		hasChange = true
		contentChanged = true
	}
	if params.ChangeNote != nil {
		updated.ChangeNote = strings.TrimSpace(*params.ChangeNote)
		hasChange = true
	}
	if params.CancelledReason != nil {
		updated.CancelledReason = strings.TrimSpace(*params.CancelledReason)
		hasChange = true
	}
	if err := validateEventReleaseSchedule(updated.PublicVisibleFrom, updated.RegistrationOpensAt, updated.RegistrationClosesAt); err != nil {
		return Event{}, false, err
	}
	if err := validateEventTicketConfig(updated.TicketName, updated.PriceCents, updated.Currency, updated.DonationEnabled, updated.DonationMinCents, updated.DonationSuggestedCents); err != nil {
		return Event{}, false, err
	}

	if contentChanged {
		switch strings.ToLower(strings.TrimSpace(current.Status)) {
		case EventStatusScheduled, EventStatusChanged:
			updated.Status = EventStatusChanged
		}
	}

	return updated, hasChange, nil
}

type normalizedCreateEvent struct {
	SeriesID               string
	Slug                   string
	Title                  string
	Subtitle               string
	Description            string
	StartsAt               time.Time
	EndsAt                 *time.Time
	Timezone               string
	LocationName           string
	Address                string
	OnlineURL              string
	ParticipationMode      string
	IsPublic               bool
	PublicVisibleFrom      *time.Time
	RegistrationOpensAt    *time.Time
	RegistrationClosesAt   *time.Time
	RegistrationEnabled    bool
	WaitlistEnabled        bool
	MaxParticipants        *int
	TicketName             string
	PriceCents             int
	Currency               string
	DonationEnabled        bool
	DonationMinCents       *int
	DonationSuggestedCents *int
}

func normalizeCreateEventParams(params CreateEventParams) (normalizedCreateEvent, error) {
	slug, err := normalizeEventSlug(params.Slug)
	if err != nil {
		return normalizedCreateEvent{}, err
	}
	title := strings.TrimSpace(params.Title)
	if title == "" {
		return normalizedCreateEvent{}, fmt.Errorf("event title must not be empty")
	}
	if len(title) > 180 {
		return normalizedCreateEvent{}, fmt.Errorf("event title must be <= 180 characters")
	}
	startsAt, err := parseRequiredRFC3339(params.StartsAt, "event starts_at")
	if err != nil {
		return normalizedCreateEvent{}, err
	}
	endsAt, err := parseOptionalRFC3339(params.EndsAt, "event ends_at")
	if err != nil {
		return normalizedCreateEvent{}, err
	}
	if endsAt != nil && endsAt.Before(startsAt) {
		return normalizedCreateEvent{}, fmt.Errorf("event ends_at must be >= starts_at")
	}

	mode, err := normalizeParticipationMode(params.ParticipationMode)
	if err != nil {
		return normalizedCreateEvent{}, err
	}
	onlineURL, err := normalizeOptionalEventURL(params.OnlineURL)
	if err != nil {
		return normalizedCreateEvent{}, err
	}

	timezone := strings.TrimSpace(params.Timezone)
	if timezone == "" {
		timezone = "Europe/Berlin"
	}

	isPublic := false
	if params.IsPublic != nil {
		isPublic = *params.IsPublic
	}
	publicVisibleFrom, err := parseOptionalRFC3339(params.PublicVisibleFrom, "event public_visible_from")
	if err != nil {
		return normalizedCreateEvent{}, err
	}
	registrationOpensAt, err := parseOptionalRFC3339(params.RegistrationOpensAt, "event registration_opens_at")
	if err != nil {
		return normalizedCreateEvent{}, err
	}
	registrationClosesAt, err := parseOptionalRFC3339(params.RegistrationClosesAt, "event registration_closes_at")
	if err != nil {
		return normalizedCreateEvent{}, err
	}
	registrationEnabled := true
	if params.RegistrationEnabled != nil {
		registrationEnabled = *params.RegistrationEnabled
	}
	waitlistEnabled := true
	if params.WaitlistEnabled != nil {
		waitlistEnabled = *params.WaitlistEnabled
	}

	var maxParticipants *int
	if params.MaxParticipants != nil {
		value := *params.MaxParticipants
		if value <= 0 {
			return normalizedCreateEvent{}, fmt.Errorf("event max_participants must be > 0")
		}
		maxParticipants = &value
	}
	ticketName := strings.TrimSpace(params.TicketName)
	priceCents := 0
	if params.PriceCents != nil {
		priceCents = *params.PriceCents
	}
	currency := strings.ToUpper(strings.TrimSpace(params.Currency))
	if currency == "" {
		currency = "EUR"
	}
	donationEnabled := false
	if params.DonationEnabled != nil {
		donationEnabled = *params.DonationEnabled
	}
	var donationMinCents *int
	if params.DonationMinCents != nil {
		value := *params.DonationMinCents
		donationMinCents = &value
	}
	var donationSuggestedCents *int
	if params.DonationSuggestedCents != nil {
		value := *params.DonationSuggestedCents
		donationSuggestedCents = &value
	}
	if err := validateEventReleaseSchedule(publicVisibleFrom, registrationOpensAt, registrationClosesAt); err != nil {
		return normalizedCreateEvent{}, err
	}
	if err := validateEventTicketConfig(ticketName, priceCents, currency, donationEnabled, donationMinCents, donationSuggestedCents); err != nil {
		return normalizedCreateEvent{}, err
	}

	return normalizedCreateEvent{
		SeriesID:               strings.TrimSpace(params.SeriesID),
		Slug:                   slug,
		Title:                  title,
		Subtitle:               strings.TrimSpace(params.Subtitle),
		Description:            strings.TrimSpace(params.Description),
		StartsAt:               startsAt,
		EndsAt:                 endsAt,
		Timezone:               timezone,
		LocationName:           strings.TrimSpace(params.LocationName),
		Address:                strings.TrimSpace(params.Address),
		OnlineURL:              onlineURL,
		ParticipationMode:      mode,
		IsPublic:               isPublic,
		PublicVisibleFrom:      publicVisibleFrom,
		RegistrationOpensAt:    registrationOpensAt,
		RegistrationClosesAt:   registrationClosesAt,
		RegistrationEnabled:    registrationEnabled,
		WaitlistEnabled:        waitlistEnabled,
		MaxParticipants:        maxParticipants,
		TicketName:             ticketName,
		PriceCents:             priceCents,
		Currency:               currency,
		DonationEnabled:        donationEnabled,
		DonationMinCents:       donationMinCents,
		DonationSuggestedCents: donationSuggestedCents,
	}, nil
}

func (r *Repository) normalizeSeriesForTenant(ctx context.Context, tenantID, seriesID string) (string, error) {
	id := strings.TrimSpace(seriesID)
	if id == "" {
		return "", nil
	}
	if _, err := r.GetSeriesByID(ctx, tenantID, id); err != nil {
		if errors.Is(err, ErrSeriesNotFound) {
			return "", ErrEventSeriesScopeMismatch
		}
		return "", err
	}
	return id, nil
}

func normalizeEventSlug(raw string) (string, error) {
	slug := strings.ToLower(strings.Trim(strings.TrimSpace(raw), "/"))
	if slug == "" {
		return "", fmt.Errorf("event slug must not be empty")
	}
	if len(slug) > 120 {
		return "", fmt.Errorf("event slug must be <= 120 characters")
	}
	if !eventSlugPattern.MatchString(slug) {
		return "", fmt.Errorf("event slug %q is invalid", raw)
	}
	return slug, nil
}

func parseRequiredRFC3339(raw, field string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("%s must not be empty", field)
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %s: %w", field, err)
	}
	return parsed.UTC(), nil
}

func parseOptionalRFC3339(raw, field string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", field, err)
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func normalizeOptionalEventURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse event online_url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("event online_url must include scheme and host")
	}
	return value, nil
}

func normalizeParticipationMode(raw string) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(raw))
	if mode == "" {
		mode = ParticipationModeOnsite
	}
	switch mode {
	case ParticipationModeOnsite, ParticipationModeOnline, ParticipationModeHybrid:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported event participation_mode %q", raw)
	}
}

func isEventSlugConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: events.tenant_id, events.slug")
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func validateEventReleaseSchedule(publicVisibleFrom, registrationOpensAt, registrationClosesAt *time.Time) error {
	if registrationOpensAt != nil && registrationClosesAt != nil && registrationClosesAt.Before(*registrationOpensAt) {
		return fmt.Errorf("event registration_closes_at must be >= registration_opens_at")
	}
	if publicVisibleFrom != nil && registrationClosesAt != nil && registrationClosesAt.Before(*publicVisibleFrom) {
		return fmt.Errorf("event registration_closes_at must be >= public_visible_from")
	}
	return nil
}

func scanEvent(row rowScanner) (Event, error) {
	var (
		item                    Event
		startsAtRaw             string
		endsAtRaw               string
		isPublicInt             int
		publishedAtRaw          string
		publicVisibleFromRaw    string
		registrationOpensAtRaw  string
		registrationClosesAtRaw string
		registrationEnabled     int
		waitlistEnabled         int
		maxParticipantsRaw      sql.NullInt64
		donationEnabledRaw      int
		donationMinCentsRaw     sql.NullInt64
		donationSuggestedRaw    sql.NullInt64
		createdAtRaw            string
		updatedAtRaw            string
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
		&publicVisibleFromRaw,
		&registrationOpensAtRaw,
		&registrationClosesAtRaw,
		&registrationEnabled,
		&waitlistEnabled,
		&maxParticipantsRaw,
		&item.ConfirmedParticipants,
		&item.WaitlistEntries,
		&item.TicketName,
		&item.PriceCents,
		&item.Currency,
		&donationEnabledRaw,
		&donationMinCentsRaw,
		&donationSuggestedRaw,
		&item.ChangeNote,
		&item.CancelledReason,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		return Event{}, err
	}

	startsAt, err := time.Parse(time.RFC3339, startsAtRaw)
	if err != nil {
		return Event{}, fmt.Errorf("parse event starts_at: %w", err)
	}
	var endsAt *time.Time
	if strings.TrimSpace(endsAtRaw) != "" {
		parsedEndsAt, parseErr := time.Parse(time.RFC3339, endsAtRaw)
		if parseErr != nil {
			return Event{}, fmt.Errorf("parse event ends_at: %w", parseErr)
		}
		parsedEndsAt = parsedEndsAt.UTC()
		endsAt = &parsedEndsAt
	}
	var publishedAt *time.Time
	if strings.TrimSpace(publishedAtRaw) != "" {
		parsedPublishedAt, parseErr := time.Parse(time.RFC3339, publishedAtRaw)
		if parseErr != nil {
			return Event{}, fmt.Errorf("parse event published_at: %w", parseErr)
		}
		parsedPublishedAt = parsedPublishedAt.UTC()
		publishedAt = &parsedPublishedAt
	}
	publicVisibleFrom, err := parseOptionalRFC3339(publicVisibleFromRaw, "event public_visible_from")
	if err != nil {
		return Event{}, err
	}
	registrationOpensAt, err := parseOptionalRFC3339(registrationOpensAtRaw, "event registration_opens_at")
	if err != nil {
		return Event{}, err
	}
	registrationClosesAt, err := parseOptionalRFC3339(registrationClosesAtRaw, "event registration_closes_at")
	if err != nil {
		return Event{}, err
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return Event{}, fmt.Errorf("parse event created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return Event{}, fmt.Errorf("parse event updated_at: %w", err)
	}

	item.StartsAt = startsAt.UTC()
	item.EndsAt = endsAt
	item.IsPublic = isPublicInt == 1
	item.PublishedAt = publishedAt
	item.PublicVisibleFrom = publicVisibleFrom
	item.RegistrationOpensAt = registrationOpensAt
	item.RegistrationClosesAt = registrationClosesAt
	item.RegistrationEnabled = registrationEnabled == 1
	item.WaitlistEnabled = waitlistEnabled == 1
	item.DonationEnabled = donationEnabledRaw == 1
	if maxParticipantsRaw.Valid {
		value := int(maxParticipantsRaw.Int64)
		item.MaxParticipants = &value
	}
	if donationMinCentsRaw.Valid {
		value := int(donationMinCentsRaw.Int64)
		item.DonationMinCents = &value
	}
	if donationSuggestedRaw.Valid {
		value := int(donationSuggestedRaw.Int64)
		item.DonationSuggestedCents = &value
	}
	item.CreatedAt = createdAt.UTC()
	item.UpdatedAt = updatedAt.UTC()
	return item, nil
}

func (e Event) RequiresPayment() bool {
	return e.PriceCents > 0
}

func validateEventTicketConfig(ticketName string, priceCents int, currency string, donationEnabled bool, donationMinCents, donationSuggestedCents *int) error {
	normalizedCurrency := strings.ToUpper(strings.TrimSpace(currency))
	if normalizedCurrency == "" {
		normalizedCurrency = "EUR"
	}
	if priceCents < 0 {
		return fmt.Errorf("event price_cents must be >= 0")
	}
	if len(normalizedCurrency) != 3 {
		return fmt.Errorf("event currency must be a 3-letter code")
	}
	if donationEnabled && priceCents <= 0 {
		return fmt.Errorf("event donation_enabled currently requires price_cents > 0")
	}
	if donationMinCents != nil && *donationMinCents < 0 {
		return fmt.Errorf("event donation_min_cents must be >= 0")
	}
	if donationSuggestedCents != nil && *donationSuggestedCents < 0 {
		return fmt.Errorf("event donation_suggested_cents must be >= 0")
	}
	if donationMinCents != nil && donationSuggestedCents != nil && *donationSuggestedCents < *donationMinCents {
		return fmt.Errorf("event donation_suggested_cents must be >= donation_min_cents")
	}
	if priceCents > 0 && strings.TrimSpace(ticketName) == "" {
		return fmt.Errorf("event ticket_name must not be empty when price_cents > 0")
	}
	return nil
}

func (r *Repository) syncDefaultEventTicketTx(ctx context.Context, tx *sql.Tx, tenantID, eventID string, config normalizedCreateEvent) error {
	return r.syncDefaultEventTicketRecordTx(ctx, tx, tenantID, eventID, config.TicketName, config.PriceCents, config.Currency, config.DonationEnabled, config.DonationMinCents, config.DonationSuggestedCents)
}

func (r *Repository) syncDefaultEventTicketFromEventTx(ctx context.Context, tx *sql.Tx, item Event) error {
	return r.syncDefaultEventTicketRecordTx(ctx, tx, item.TenantID, item.ID, item.TicketName, item.PriceCents, item.Currency, item.DonationEnabled, item.DonationMinCents, item.DonationSuggestedCents)
}

func (r *Repository) syncDefaultEventTicketRecordTx(ctx context.Context, tx *sql.Tx, tenantID, eventID, ticketName string, priceCents int, currency string, donationEnabled bool, donationMinCents, donationSuggestedCents *int) error {
	if tx == nil {
		return fmt.Errorf("transaction must not be nil")
	}
	if priceCents <= 0 && !donationEnabled {
		if _, err := tx.ExecContext(ctx, `DELETE FROM event_tickets WHERE tenant_id = ? AND event_id = ?`, tenantID, eventID); err != nil {
			return fmt.Errorf("delete event ticket: %w", err)
		}
		return nil
	}

	row := tx.QueryRowContext(ctx, `SELECT id FROM event_tickets WHERE tenant_id = ? AND event_id = ? ORDER BY created_at ASC LIMIT 1`, tenantID, eventID)
	var ticketID string
	err := row.Scan(&ticketID)
	now := r.nowFn().UTC().Format(time.RFC3339)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("lookup event ticket: %w", err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO event_tickets (
        id, tenant_id, event_id, name, ticket_type, price_cents, currency, max_quantity,
        donation_enabled, donation_min_cents, donation_suggested_cents, created_at, updated_at
      ) VALUES (?, ?, ?, ?, 'standard', ?, ?, NULL, ?, ?, ?, ?, ?)`,
			r.idFn("tkt"),
			tenantID,
			eventID,
			ticketName,
			priceCents,
			strings.ToUpper(strings.TrimSpace(currency)),
			boolToInt(donationEnabled),
			nullableInt(donationMinCents),
			nullableInt(donationSuggestedCents),
			now,
			now,
		); err != nil {
			return fmt.Errorf("insert event ticket: %w", err)
		}
		return nil
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE event_tickets
     SET name = ?, price_cents = ?, currency = ?, donation_enabled = ?, donation_min_cents = ?, donation_suggested_cents = ?, updated_at = ?
     WHERE id = ? AND tenant_id = ?`,
		ticketName,
		priceCents,
		strings.ToUpper(strings.TrimSpace(currency)),
		boolToInt(donationEnabled),
		nullableInt(donationMinCents),
		nullableInt(donationSuggestedCents),
		now,
		ticketID,
		tenantID,
	); err != nil {
		return fmt.Errorf("update event ticket: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM event_tickets WHERE tenant_id = ? AND event_id = ? AND id <> ?`, tenantID, eventID, ticketID); err != nil {
		return fmt.Errorf("cleanup event tickets: %w", err)
	}
	return nil
}

func canEventBePublic(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case EventStatusDraft, EventStatusArchived:
		return false
	default:
		return true
	}
}
