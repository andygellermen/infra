package registration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
)

type AdminRegistration struct {
	ID                string
	TenantID          string
	EventID           string
	EventSlug         string
	EventTitle        string
	ParticipantID     string
	ParticipantName   string
	ParticipantEmail  string
	ParticipantPhone  string
	Status            string
	ParticipationType string
	Quantity          int
	PaymentStatus     string
	Source            string
	CancellationNote  string
	ReservedUntil     *time.Time
	ConfirmedAt       *time.Time
	CancelledAt       *time.Time
	AttendedAt        *time.Time
	PrivacyAcceptedAt *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Dashboard struct {
	Stats      DashboardStats
	Today      []DashboardEvent
	NextEvents []DashboardEvent
}

type DashboardStats struct {
	TodayEvents           int
	UpcomingEvents        int
	ConfirmedParticipants int
	WaitlistEntries       int
	FreeSeats             int
	OpenEmailJobs         int
	LastRetentionRunAt    *time.Time
}

type DashboardEvent struct {
	ID                    string
	Slug                  string
	Title                 string
	Status                string
	StartsAt              time.Time
	LocationName          string
	MaxParticipants       *int
	ConfirmedParticipants int
	WaitlistEntries       int
	OccupiedSeats         int
	FreeSeats             *int
}

func (s *Service) ListEventRegistrations(ctx context.Context, tenantID, eventID string) ([]AdminRegistration, error) {
	if s.db == nil {
		return nil, fmt.Errorf("registration service database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return nil, fmt.Errorf("event id must not be empty")
	}
	if _, err := s.lookupEvent(ctx, tenant, eventID); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(
		ctx,
		adminRegistrationsSelect+` WHERE r.tenant_id = ? AND r.event_id = ? ORDER BY r.created_at DESC`,
		tenant,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("list event registrations: %w", err)
	}
	defer rows.Close()

	items := make([]AdminRegistration, 0)
	for rows.Next() {
		item, scanErr := scanAdminRegistration(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate event registrations: %w", err)
	}
	return items, nil
}

func (s *Service) GetRegistration(ctx context.Context, tenantID, registrationID string) (AdminRegistration, error) {
	if s.db == nil {
		return AdminRegistration{}, fmt.Errorf("registration service database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return AdminRegistration{}, fmt.Errorf("tenant id must not be empty")
	}
	registrationID = strings.TrimSpace(registrationID)
	if registrationID == "" {
		return AdminRegistration{}, fmt.Errorf("registration id must not be empty")
	}

	row := s.db.QueryRowContext(
		ctx,
		adminRegistrationsSelect+` WHERE r.tenant_id = ? AND r.id = ? LIMIT 1`,
		tenant,
		registrationID,
	)
	return scanAdminRegistration(row)
}

func (s *Service) GetDashboard(ctx context.Context, tenantID string) (Dashboard, error) {
	if s.db == nil {
		return Dashboard{}, fmt.Errorf("registration service database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Dashboard{}, fmt.Errorf("tenant id must not be empty")
	}

	now := s.nowFn().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfTomorrow := startOfToday.Add(24 * time.Hour)

	todayCount, err := s.countDashboardMetric(
		ctx,
		`SELECT COUNT(*)
FROM events
WHERE tenant_id = ?
  AND starts_at >= ?
  AND starts_at < ?
  AND status IN (?, ?, ?)`,
		tenant,
		startOfToday.Format(time.RFC3339),
		startOfTomorrow.Format(time.RFC3339),
		event.EventStatusScheduled,
		event.EventStatusChanged,
		event.EventStatusPostponed,
	)
	if err != nil {
		return Dashboard{}, err
	}

	upcomingCount, err := s.countDashboardMetric(
		ctx,
		`SELECT COUNT(*)
FROM events
WHERE tenant_id = ?
  AND starts_at >= ?
  AND status IN (?, ?, ?)`,
		tenant,
		now.Format(time.RFC3339),
		event.EventStatusScheduled,
		event.EventStatusChanged,
		event.EventStatusPostponed,
	)
	if err != nil {
		return Dashboard{}, err
	}

	confirmedCount, err := s.countDashboardMetric(
		ctx,
		`SELECT COUNT(*)
FROM registrations
WHERE tenant_id = ? AND status = ?`,
		tenant,
		StatusConfirmed,
	)
	if err != nil {
		return Dashboard{}, err
	}

	waitlistCount, err := s.countDashboardMetric(
		ctx,
		`SELECT COUNT(*)
FROM waitlist_entries
WHERE tenant_id = ? AND status IN (?, ?)`,
		tenant,
		WaitlistStatusWaiting,
		WaitlistStatusOffered,
	)
	if err != nil {
		return Dashboard{}, err
	}

	freeSeats, err := s.countDashboardMetric(
		ctx,
		`SELECT COALESCE(SUM(
  CASE
    WHEN e.max_participants IS NULL THEN 0
    ELSE MAX(
      e.max_participants - (
        SELECT COUNT(*)
        FROM registrations r
        WHERE r.tenant_id = e.tenant_id
          AND r.event_id = e.id
          AND r.status IN (?, ?, ?)
          AND (r.reserved_until IS NULL OR r.reserved_until > ?)
      ),
      0
    )
  END
), 0)
FROM events e
WHERE e.tenant_id = ?
  AND e.starts_at >= ?
  AND e.status IN (?, ?, ?)`,
		StatusConfirmed,
		StatusReserved,
		StatusPaymentPending,
		now.Format(time.RFC3339),
		tenant,
		now.Format(time.RFC3339),
		event.EventStatusScheduled,
		event.EventStatusChanged,
		event.EventStatusPostponed,
	)
	if err != nil {
		return Dashboard{}, err
	}

	openEmailJobs, err := s.countDashboardMetric(
		ctx,
		`SELECT COUNT(*)
FROM email_jobs
WHERE tenant_id = ? AND status <> ?`,
		tenant,
		"sent",
	)
	if err != nil {
		return Dashboard{}, err
	}

	lastRetentionRunAt, err := s.lookupLastRetentionRun(ctx, tenant)
	if err != nil {
		return Dashboard{}, err
	}

	todayEvents, err := s.listDashboardEvents(
		ctx,
		tenant,
		startOfToday,
		&startOfTomorrow,
		now,
		5,
	)
	if err != nil {
		return Dashboard{}, err
	}
	nextEvents, err := s.listDashboardEvents(
		ctx,
		tenant,
		now,
		nil,
		now,
		8,
	)
	if err != nil {
		return Dashboard{}, err
	}

	return Dashboard{
		Stats: DashboardStats{
			TodayEvents:           todayCount,
			UpcomingEvents:        upcomingCount,
			ConfirmedParticipants: confirmedCount,
			WaitlistEntries:       waitlistCount,
			FreeSeats:             freeSeats,
			OpenEmailJobs:         openEmailJobs,
			LastRetentionRunAt:    lastRetentionRunAt,
		},
		Today:      todayEvents,
		NextEvents: nextEvents,
	}, nil
}

func (s *Service) countDashboardMetric(ctx context.Context, query string, args ...any) (int, error) {
	row := s.db.QueryRowContext(ctx, query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("query dashboard metric: %w", err)
	}
	return count, nil
}

func (s *Service) lookupLastRetentionRun(ctx context.Context, tenantID string) (*time.Time, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(MAX(created_at), '')
FROM audit_log
WHERE tenant_id = ?
  AND action IN (?, ?, ?)`,
		tenantID,
		"retention_job_run",
		"retention_run",
		"retention_job_dry_run",
	)
	var raw string
	if err := row.Scan(&raw); err != nil {
		return nil, fmt.Errorf("query last retention run: %w", err)
	}
	return parseOptionalRFC3339(raw, "last retention run")
}

func (s *Service) listDashboardEvents(
	ctx context.Context,
	tenantID string,
	start time.Time,
	end *time.Time,
	now time.Time,
	limit int,
) ([]DashboardEvent, error) {
	query := dashboardEventSelect +
		` WHERE e.tenant_id = ?
  AND e.starts_at >= ?`
	args := []any{
		StatusConfirmed,
		WaitlistStatusWaiting,
		WaitlistStatusOffered,
		StatusConfirmed,
		StatusReserved,
		StatusPaymentPending,
		now.Format(time.RFC3339),
		tenantID,
		start.Format(time.RFC3339),
	}
	if end != nil {
		query += "\n  AND e.starts_at < ?"
		args = append(args, end.UTC().Format(time.RFC3339))
	}
	query += `
  AND e.status IN (?, ?, ?)
ORDER BY e.starts_at ASC
LIMIT ?`
	args = append(
		args,
		event.EventStatusScheduled,
		event.EventStatusChanged,
		event.EventStatusPostponed,
		limit,
	)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dashboard events: %w", err)
	}
	defer rows.Close()

	items := make([]DashboardEvent, 0)
	for rows.Next() {
		item, scanErr := scanDashboardEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dashboard events: %w", err)
	}
	return items, nil
}

type adminRegistrationRow struct {
	ID                string
	TenantID          string
	EventID           string
	EventSlug         string
	EventTitle        string
	ParticipantID     string
	ParticipantName   string
	ParticipantEmail  string
	ParticipantPhone  string
	Status            string
	ParticipationType string
	Quantity          int
	Source            string
	CancellationNote  string
	ReservedUntilRaw  string
	ConfirmedAtRaw    string
	CancelledAtRaw    string
	AttendedAtRaw     string
	PrivacyAtRaw      string
	CreatedAtRaw      string
	UpdatedAtRaw      string
	PaymentStatus     string
}

const adminRegistrationsSelect = `SELECT r.id, r.tenant_id, r.event_id, COALESCE(e.slug, ''), COALESCE(e.title, ''),
       r.participant_id, COALESCE(p.name, ''), COALESCE(p.email, ''), COALESCE(p.phone, ''),
       r.status, r.participation_type, r.quantity, COALESCE(r.source, ''), COALESCE(r.cancellation_reason, ''),
       COALESCE(r.reserved_until, ''), COALESCE(r.confirmed_at, ''), COALESCE(r.cancelled_at, ''), COALESCE(r.attended_at, ''), COALESCE(r.privacy_accepted_at, ''),
       r.created_at, r.updated_at,
       COALESCE((SELECT py.status
                 FROM payments py
                 WHERE py.tenant_id = r.tenant_id
                   AND py.registration_id = r.id
                 ORDER BY py.created_at DESC
                 LIMIT 1), '')
FROM registrations r
LEFT JOIN participants p ON p.id = r.participant_id
LEFT JOIN events e ON e.id = r.event_id`

func scanAdminRegistration(row interface{ Scan(dest ...any) error }) (AdminRegistration, error) {
	raw, err := scanAdminRegistrationRow(row)
	if err != nil {
		return AdminRegistration{}, err
	}
	return mapAdminRegistration(raw)
}

func scanAdminRegistrationRow(row interface{ Scan(dest ...any) error }) (adminRegistrationRow, error) {
	var raw adminRegistrationRow
	if err := row.Scan(
		&raw.ID,
		&raw.TenantID,
		&raw.EventID,
		&raw.EventSlug,
		&raw.EventTitle,
		&raw.ParticipantID,
		&raw.ParticipantName,
		&raw.ParticipantEmail,
		&raw.ParticipantPhone,
		&raw.Status,
		&raw.ParticipationType,
		&raw.Quantity,
		&raw.Source,
		&raw.CancellationNote,
		&raw.ReservedUntilRaw,
		&raw.ConfirmedAtRaw,
		&raw.CancelledAtRaw,
		&raw.AttendedAtRaw,
		&raw.PrivacyAtRaw,
		&raw.CreatedAtRaw,
		&raw.UpdatedAtRaw,
		&raw.PaymentStatus,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return adminRegistrationRow{}, ErrRegistrationNotFound
		}
		return adminRegistrationRow{}, fmt.Errorf("scan admin registration: %w", err)
	}
	return raw, nil
}

func mapAdminRegistration(raw adminRegistrationRow) (AdminRegistration, error) {
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAtRaw)
	if err != nil {
		return AdminRegistration{}, fmt.Errorf("parse registration created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, raw.UpdatedAtRaw)
	if err != nil {
		return AdminRegistration{}, fmt.Errorf("parse registration updated_at: %w", err)
	}

	reservedUntil, err := parseOptionalRFC3339(raw.ReservedUntilRaw, "registration reserved_until")
	if err != nil {
		return AdminRegistration{}, err
	}
	confirmedAt, err := parseOptionalRFC3339(raw.ConfirmedAtRaw, "registration confirmed_at")
	if err != nil {
		return AdminRegistration{}, err
	}
	cancelledAt, err := parseOptionalRFC3339(raw.CancelledAtRaw, "registration cancelled_at")
	if err != nil {
		return AdminRegistration{}, err
	}
	attendedAt, err := parseOptionalRFC3339(raw.AttendedAtRaw, "registration attended_at")
	if err != nil {
		return AdminRegistration{}, err
	}
	privacyAt, err := parseOptionalRFC3339(raw.PrivacyAtRaw, "registration privacy_accepted_at")
	if err != nil {
		return AdminRegistration{}, err
	}

	return AdminRegistration{
		ID:                raw.ID,
		TenantID:          raw.TenantID,
		EventID:           raw.EventID,
		EventSlug:         raw.EventSlug,
		EventTitle:        raw.EventTitle,
		ParticipantID:     raw.ParticipantID,
		ParticipantName:   raw.ParticipantName,
		ParticipantEmail:  raw.ParticipantEmail,
		ParticipantPhone:  raw.ParticipantPhone,
		Status:            raw.Status,
		ParticipationType: raw.ParticipationType,
		Quantity:          raw.Quantity,
		PaymentStatus:     raw.PaymentStatus,
		Source:            raw.Source,
		CancellationNote:  raw.CancellationNote,
		ReservedUntil:     reservedUntil,
		ConfirmedAt:       confirmedAt,
		CancelledAt:       cancelledAt,
		AttendedAt:        attendedAt,
		PrivacyAcceptedAt: privacyAt,
		CreatedAt:         createdAt.UTC(),
		UpdatedAt:         updatedAt.UTC(),
	}, nil
}

const dashboardEventSelect = `SELECT e.id, e.slug, e.title, e.status, e.starts_at, COALESCE(e.location_name, ''), e.max_participants,
       (SELECT COUNT(*)
        FROM registrations r
        WHERE r.tenant_id = e.tenant_id
          AND r.event_id = e.id
          AND r.status = ?) AS confirmed_count,
       (SELECT COUNT(*)
        FROM waitlist_entries w
        WHERE w.tenant_id = e.tenant_id
          AND w.event_id = e.id
          AND w.status IN (?, ?)) AS waitlist_count,
       (SELECT COUNT(*)
        FROM registrations r2
        WHERE r2.tenant_id = e.tenant_id
          AND r2.event_id = e.id
          AND r2.status IN (?, ?, ?)
          AND (r2.reserved_until IS NULL OR r2.reserved_until > ?)) AS occupied_count
FROM events e`

func scanDashboardEvent(row interface{ Scan(dest ...any) error }) (DashboardEvent, error) {
	var (
		item            DashboardEvent
		startsAtRaw     string
		maxParticipants sql.NullInt64
	)
	if err := row.Scan(
		&item.ID,
		&item.Slug,
		&item.Title,
		&item.Status,
		&startsAtRaw,
		&item.LocationName,
		&maxParticipants,
		&item.ConfirmedParticipants,
		&item.WaitlistEntries,
		&item.OccupiedSeats,
	); err != nil {
		return DashboardEvent{}, fmt.Errorf("scan dashboard event: %w", err)
	}

	startsAt, err := time.Parse(time.RFC3339, startsAtRaw)
	if err != nil {
		return DashboardEvent{}, fmt.Errorf("parse dashboard event starts_at: %w", err)
	}
	item.StartsAt = startsAt.UTC()

	if maxParticipants.Valid {
		maxVal := int(maxParticipants.Int64)
		item.MaxParticipants = &maxVal
		free := maxVal - item.OccupiedSeats
		if free < 0 {
			free = 0
		}
		item.FreeSeats = &free
	}

	return item, nil
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
	parsed = parsed.UTC()
	return &parsed, nil
}
