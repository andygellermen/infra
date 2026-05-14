package registration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ParticipantPortalRegistration struct {
	ID                 string
	TenantID           string
	EventID            string
	EventSlug          string
	EventTitle         string
	EventStartsAt      time.Time
	EventEndsAt        *time.Time
	EventTimezone      string
	EventLocationName  string
	EventOnlineURL     string
	Status             string
	ParticipationType  string
	Quantity           int
	Source             string
	PaymentStatus      string
	ConfirmedAt        *time.Time
	CancelledAt        *time.Time
	CancellationReason string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (s *Service) ListParticipantRegistrations(ctx context.Context, tenantID, participantID string) ([]ParticipantPortalRegistration, error) {
	if s.db == nil {
		return nil, fmt.Errorf("registration service database is nil")
	}
	tenantID = strings.TrimSpace(tenantID)
	participantID = strings.TrimSpace(participantID)
	if tenantID == "" || participantID == "" {
		return nil, ErrParticipantAccessDenied
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT r.id, r.tenant_id, r.event_id,
            COALESCE(e.slug, ''), COALESCE(e.title, ''),
            COALESCE(e.starts_at, ''), COALESCE(e.ends_at, ''), COALESCE(e.timezone, ''),
            COALESCE(e.location_name, ''), COALESCE(e.online_url, ''),
            r.status, r.participation_type, r.quantity, COALESCE(r.source, ''),
            COALESCE(r.confirmed_at, ''), COALESCE(r.cancelled_at, ''), COALESCE(r.cancellation_reason, ''),
            r.created_at, r.updated_at,
            COALESCE((SELECT py.status
                     FROM payments py
                     WHERE py.tenant_id = r.tenant_id
                       AND py.registration_id = r.id
                     ORDER BY py.created_at DESC
                     LIMIT 1), '')
     FROM registrations r
     LEFT JOIN events e ON e.id = r.event_id
     WHERE r.tenant_id = ?
       AND r.participant_id = ?
     ORDER BY e.starts_at DESC, r.created_at DESC`,
		tenantID,
		participantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list participant registrations: %w", err)
	}
	defer rows.Close()

	items := make([]ParticipantPortalRegistration, 0)
	for rows.Next() {
		item, scanErr := scanParticipantPortalRegistration(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate participant registrations: %w", err)
	}
	return items, nil
}

func (s *Service) GetParticipantRegistration(ctx context.Context, tenantID, participantID, registrationID string) (ParticipantPortalRegistration, error) {
	if s.db == nil {
		return ParticipantPortalRegistration{}, fmt.Errorf("registration service database is nil")
	}
	tenantID = strings.TrimSpace(tenantID)
	participantID = strings.TrimSpace(participantID)
	registrationID = strings.TrimSpace(registrationID)
	if tenantID == "" || participantID == "" || registrationID == "" {
		return ParticipantPortalRegistration{}, ErrParticipantAccessDenied
	}

	row := s.db.QueryRowContext(
		ctx,
		`SELECT r.id, r.tenant_id, r.event_id,
            COALESCE(e.slug, ''), COALESCE(e.title, ''),
            COALESCE(e.starts_at, ''), COALESCE(e.ends_at, ''), COALESCE(e.timezone, ''),
            COALESCE(e.location_name, ''), COALESCE(e.online_url, ''),
            r.status, r.participation_type, r.quantity, COALESCE(r.source, ''),
            COALESCE(r.confirmed_at, ''), COALESCE(r.cancelled_at, ''), COALESCE(r.cancellation_reason, ''),
            r.created_at, r.updated_at,
            COALESCE((SELECT py.status
                     FROM payments py
                     WHERE py.tenant_id = r.tenant_id
                       AND py.registration_id = r.id
                     ORDER BY py.created_at DESC
                     LIMIT 1), '')
     FROM registrations r
     LEFT JOIN events e ON e.id = r.event_id
     WHERE r.tenant_id = ?
       AND r.participant_id = ?
       AND r.id = ?
     LIMIT 1`,
		tenantID,
		participantID,
		registrationID,
	)
	item, err := scanParticipantPortalRegistration(row)
	if err != nil {
		if errors.Is(err, ErrRegistrationNotFound) {
			return ParticipantPortalRegistration{}, ErrParticipantAccessDenied
		}
		return ParticipantPortalRegistration{}, err
	}
	return item, nil
}

func (s *Service) CancelParticipantRegistration(ctx context.Context, tenantID, participantID, registrationID, reason string) (ParticipantPortalRegistration, error) {
	if s.db == nil {
		return ParticipantPortalRegistration{}, fmt.Errorf("registration service database is nil")
	}
	tenantID = strings.TrimSpace(tenantID)
	participantID = strings.TrimSpace(participantID)
	registrationID = strings.TrimSpace(registrationID)
	if tenantID == "" || participantID == "" || registrationID == "" {
		return ParticipantPortalRegistration{}, ErrParticipantAccessDenied
	}

	current, err := s.GetParticipantRegistration(ctx, tenantID, participantID, registrationID)
	if err != nil {
		return ParticipantPortalRegistration{}, err
	}
	if !canCancelParticipantRegistrationStatus(current.Status) {
		return ParticipantPortalRegistration{}, ErrRegistrationCancelNotAllowed
	}

	now := s.nowFn().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ParticipantPortalRegistration{}, fmt.Errorf("begin participant cancellation transaction: %w", err)
	}

	cancelReason := strings.TrimSpace(reason)
	if cancelReason == "" {
		cancelReason = "Teilnehmer-Storno ueber Portal"
	}
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE registrations
     SET status = ?, cancelled_at = COALESCE(cancelled_at, ?), cancellation_reason = ?,
         reserved_until = NULL, updated_at = ?
     WHERE tenant_id = ? AND participant_id = ? AND id = ?`,
		StatusCancelled,
		now.Format(time.RFC3339),
		cancelReason,
		now.Format(time.RFC3339),
		tenantID,
		participantID,
		registrationID,
	); err != nil {
		_ = tx.Rollback()
		return ParticipantPortalRegistration{}, fmt.Errorf("cancel participant registration: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE waitlist_entries
     SET status = ?, updated_at = ?
     WHERE tenant_id = ? AND registration_id = ? AND status IN (?, ?)`,
		WaitlistStatusRemoved,
		now.Format(time.RFC3339),
		tenantID,
		registrationID,
		WaitlistStatusWaiting,
		WaitlistStatusOffered,
	); err != nil {
		_ = tx.Rollback()
		return ParticipantPortalRegistration{}, fmt.Errorf("remove waitlist entries on participant cancellation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ParticipantPortalRegistration{}, fmt.Errorf("commit participant cancellation transaction: %w", err)
	}

	return s.GetParticipantRegistration(ctx, tenantID, participantID, registrationID)
}

func canCancelParticipantRegistrationStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case StatusConfirmed, StatusReserved, StatusPaymentPending, StatusWaitlist:
		return true
	default:
		return false
	}
}

func scanParticipantPortalRegistration(row interface{ Scan(dest ...any) error }) (ParticipantPortalRegistration, error) {
	var (
		item             ParticipantPortalRegistration
		eventStartsAtRaw string
		eventEndsAtRaw   string
		confirmedAtRaw   string
		cancelledAtRaw   string
		createdAtRaw     string
		updatedAtRaw     string
	)

	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.EventID,
		&item.EventSlug,
		&item.EventTitle,
		&eventStartsAtRaw,
		&eventEndsAtRaw,
		&item.EventTimezone,
		&item.EventLocationName,
		&item.EventOnlineURL,
		&item.Status,
		&item.ParticipationType,
		&item.Quantity,
		&item.Source,
		&confirmedAtRaw,
		&cancelledAtRaw,
		&item.CancellationReason,
		&createdAtRaw,
		&updatedAtRaw,
		&item.PaymentStatus,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ParticipantPortalRegistration{}, ErrRegistrationNotFound
		}
		return ParticipantPortalRegistration{}, fmt.Errorf("scan participant registration: %w", err)
	}

	startsAt, err := time.Parse(time.RFC3339, eventStartsAtRaw)
	if err != nil {
		return ParticipantPortalRegistration{}, fmt.Errorf("parse participant registration event starts_at: %w", err)
	}
	item.EventStartsAt = startsAt.UTC()
	item.EventEndsAt, err = parseOptionalRFC3339(eventEndsAtRaw, "participant registration event ends_at")
	if err != nil {
		return ParticipantPortalRegistration{}, err
	}
	item.ConfirmedAt, err = parseOptionalRFC3339(confirmedAtRaw, "participant registration confirmed_at")
	if err != nil {
		return ParticipantPortalRegistration{}, err
	}
	item.CancelledAt, err = parseOptionalRFC3339(cancelledAtRaw, "participant registration cancelled_at")
	if err != nil {
		return ParticipantPortalRegistration{}, err
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return ParticipantPortalRegistration{}, fmt.Errorf("parse participant registration created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return ParticipantPortalRegistration{}, fmt.Errorf("parse participant registration updated_at: %w", err)
	}
	item.CreatedAt = createdAt.UTC()
	item.UpdatedAt = updatedAt.UTC()

	return item, nil
}
