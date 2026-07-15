package registration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/auth"
)

const (
	WaitlistStatusOffered  = "offered"
	WaitlistStatusPromoted = "promoted"
	WaitlistStatusRemoved  = "removed"
)

var (
	ErrWaitlistEntryNotFound = errors.New("waitlist entry not found")
	ErrWaitlistStateInvalid  = errors.New("waitlist state is invalid")
)

type WaitlistEntry struct {
	ID                 string
	TenantID           string
	EventID            string
	RegistrationID     string
	ParticipantID      string
	ParticipantName    string
	ParticipantEmail   string
	Position           int
	Status             string
	RegistrationStatus string
	OfferedAt          *time.Time
	OfferExpiresAt     *time.Time
	AcceptedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (s *Service) ListWaitlistEntries(ctx context.Context, tenantID, eventID string) ([]WaitlistEntry, error) {
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
		waitlistSelectQuery+` WHERE w.tenant_id = ? AND w.event_id = ? ORDER BY w.position ASC, w.created_at ASC`,
		tenant,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("list waitlist entries: %w", err)
	}
	defer rows.Close()

	result := make([]WaitlistEntry, 0)
	for rows.Next() {
		item, scanErr := scanWaitlistEntryWithMeta(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate waitlist entries: %w", err)
	}
	return result, nil
}

func (s *Service) OfferWaitlistEntry(ctx context.Context, tenantID, waitlistEntryID, requestIP, userAgent string) (WaitlistEntry, error) {
	if s.db == nil {
		return WaitlistEntry{}, fmt.Errorf("registration service database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return WaitlistEntry{}, fmt.Errorf("tenant id must not be empty")
	}
	waitlistEntryID = strings.TrimSpace(waitlistEntryID)
	if waitlistEntryID == "" {
		return WaitlistEntry{}, fmt.Errorf("waitlist entry id must not be empty")
	}

	now := s.nowFn().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WaitlistEntry{}, fmt.Errorf("begin waitlist offer transaction: %w", err)
	}

	entry, err := s.getWaitlistEntryTx(ctx, tx, tenant, waitlistEntryID)
	if err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, err
	}
	if !canOfferWaitlistStatus(entry.Status) {
		_ = tx.Rollback()
		return WaitlistEntry{}, ErrWaitlistStateInvalid
	}
	if !canPromoteRegistrationStatus(entry.RegistrationStatus) {
		_ = tx.Rollback()
		return WaitlistEntry{}, ErrWaitlistStateInvalid
	}

	expiresAt := now.Add(s.cfg.WaitlistOfferTTL)
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE waitlist_entries
     SET status = ?, offered_at = ?, offer_expires_at = ?, updated_at = ?
     WHERE id = ? AND tenant_id = ?`,
		WaitlistStatusOffered,
		now.Format(time.RFC3339),
		expiresAt.Format(time.RFC3339),
		now.Format(time.RFC3339),
		waitlistEntryID,
		tenant,
	); err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, fmt.Errorf("update waitlist entry offered status: %w", err)
	}

	offerToken, tokenErr := s.tokFn()
	if tokenErr != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, fmt.Errorf("generate waitlist offer token: %w", tokenErr)
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO magic_links (
      id, tenant_id, user_id, participant_id, purpose, token_hash, redirect_path, expires_at,
      request_ip, user_agent, created_at
    ) VALUES (?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.idFn("mlk"),
		tenant,
		entry.ParticipantID,
		auth.PurposeWaitlistOffer,
		s.hash(offerToken),
		waitlistEntryID,
		expiresAt.Format(time.RFC3339),
		nullable(strings.TrimSpace(requestIP)),
		nullable(strings.TrimSpace(userAgent)),
		now.Format(time.RFC3339),
	); err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, fmt.Errorf("insert waitlist offer magic link: %w", err)
	}

	offerURL := s.buildWaitlistOfferURL(entry.TenantSlug, offerToken)
	subject := "Wartelistenplatz verfuegbar"
	bodyText := fmt.Sprintf(
		"Hallo %s,\n\nfuer \"%s\" ist ein Platz frei geworden.\n\nBitte nutze diesen Link bis %s:\n%s\n",
		entry.ParticipantName,
		entry.EventTitle,
		expiresAt.Format(time.RFC3339),
		offerURL,
	)
	if queueErr := s.queueEmailJobTx(ctx, tx, tenant, "waitlist_offer", entry.ParticipantEmail, subject, bodyText, map[string]any{
		"waitlist_entry_id": waitlistEntryID,
		"registration_id":   entry.RegistrationID,
		"event_id":          entry.EventID,
		"offer_expires_at":  expiresAt.Format(time.RFC3339),
	}); queueErr != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, queueErr
	}

	if err := tx.Commit(); err != nil {
		return WaitlistEntry{}, fmt.Errorf("commit waitlist offer transaction: %w", err)
	}

	return s.getWaitlistEntry(ctx, tenant, waitlistEntryID)
}

func (s *Service) PromoteWaitlistEntry(ctx context.Context, tenantID, waitlistEntryID string) (WaitlistEntry, error) {
	if s.db == nil {
		return WaitlistEntry{}, fmt.Errorf("registration service database is nil")
	}

	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return WaitlistEntry{}, fmt.Errorf("tenant id must not be empty")
	}
	waitlistEntryID = strings.TrimSpace(waitlistEntryID)
	if waitlistEntryID == "" {
		return WaitlistEntry{}, fmt.Errorf("waitlist entry id must not be empty")
	}

	now := s.nowFn().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WaitlistEntry{}, fmt.Errorf("begin waitlist promote transaction: %w", err)
	}

	entry, err := s.getWaitlistEntryTx(ctx, tx, tenant, waitlistEntryID)
	if err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, err
	}
	if !canPromoteWaitlistStatus(entry.Status) {
		_ = tx.Rollback()
		return WaitlistEntry{}, ErrWaitlistStateInvalid
	}
	if !canPromoteRegistrationStatus(entry.RegistrationStatus) {
		_ = tx.Rollback()
		return WaitlistEntry{}, ErrWaitlistStateInvalid
	}

	eventItem, err := s.lookupEventTx(ctx, tx, tenant, entry.EventID)
	if err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, err
	}
	occupied, err := s.countOccupiedSeatsTx(ctx, tx, tenant, entry.EventID, now)
	if err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, err
	}
	if eventItem.MaxParticipants.Valid && occupied >= int(eventItem.MaxParticipants.Int64) {
		_ = tx.Rollback()
		return WaitlistEntry{}, ErrEventFull
	}

	confirmedAt := now.Format(time.RFC3339)
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE registrations
     SET status = ?, confirmed_at = ?, reserved_until = NULL, updated_at = ?
     WHERE id = ? AND tenant_id = ?`,
		StatusConfirmed,
		confirmedAt,
		confirmedAt,
		entry.RegistrationID,
		tenant,
	); err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, fmt.Errorf("promote waitlist registration: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE participants
     SET email_verified_at = CASE WHEN email_verified_at IS NULL THEN ? ELSE email_verified_at END,
         updated_at = ?
     WHERE id = ?`,
		confirmedAt,
		confirmedAt,
		entry.ParticipantID,
	); err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, fmt.Errorf("update participant verification for waitlist promote: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE waitlist_entries
     SET status = ?, accepted_at = ?, updated_at = ?
     WHERE id = ? AND tenant_id = ?`,
		WaitlistStatusPromoted,
		confirmedAt,
		confirmedAt,
		waitlistEntryID,
		tenant,
	); err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, fmt.Errorf("mark waitlist entry promoted: %w", err)
	}

	tenantSlug, err := s.lookupTenantSlugTx(ctx, tx, tenant)
	if err != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, err
	}
	cancelDeadlineHours := s.lookupParticipantCancelDeadlineHoursTx(ctx, tx, tenant)
	eventURL := s.buildPublicEventPageURL(tenantSlug, eventItem.Slug)
	calendarURL := ""
	if s.participantCalendarURLFn != nil {
		calendarURL = strings.TrimSpace(s.participantCalendarURLFn(tenantSlug, tenant, entry.RegistrationID, entry.ParticipantID))
	}
	subject, bodyText := BuildWaitlistPromotedEmailContent(ConfirmationEmailContentInput{
		RecipientName:                  entry.ParticipantName,
		EventTitle:                     eventItem.Title,
		EventStartsAt:                  eventItem.StartsAt,
		EventTimezone:                  eventItem.Timezone,
		EventLocationName:              eventItem.LocationName,
		EventOnlineURL:                 eventItem.OnlineURL,
		EventURL:                       eventURL,
		CalendarURL:                    calendarURL,
		ParticipantCancelDeadlineHours: cancelDeadlineHours,
	})
	if queueErr := s.queueEmailJobTx(ctx, tx, tenant, "waitlist_promoted", entry.ParticipantEmail, subject, bodyText, map[string]any{
		"waitlist_entry_id":                 waitlistEntryID,
		"registration_id":                   entry.RegistrationID,
		"event_id":                          entry.EventID,
		"event_slug":                        eventItem.Slug,
		"event_url":                         eventURL,
		"calendar_url":                      calendarURL,
		"participant_cancel_deadline_hours": cancelDeadlineHours,
		"participant_cancel_deadline_at":    participantCancelDeadlineAt(eventItem.StartsAt, cancelDeadlineHours).UTC().Format(time.RFC3339),
	}); queueErr != nil {
		_ = tx.Rollback()
		return WaitlistEntry{}, queueErr
	}

	if err := tx.Commit(); err != nil {
		return WaitlistEntry{}, fmt.Errorf("commit waitlist promote transaction: %w", err)
	}

	return s.getWaitlistEntry(ctx, tenant, waitlistEntryID)
}

type waitlistRow struct {
	ID                 string
	TenantID           string
	EventID            string
	RegistrationID     string
	ParticipantID      string
	ParticipantName    string
	ParticipantEmail   string
	Position           int
	Status             string
	RegistrationStatus string
	OfferedAtRaw       string
	OfferExpiresAtRaw  string
	AcceptedAtRaw      string
	CreatedAtRaw       string
	UpdatedAtRaw       string
	EventTitle         string
	EventSlug          string
	TenantSlug         string
}

func (s *Service) getWaitlistEntry(ctx context.Context, tenantID, waitlistEntryID string) (WaitlistEntry, error) {
	row := s.db.QueryRowContext(
		ctx,
		waitlistSelectQuery+` WHERE w.tenant_id = ? AND w.id = ? LIMIT 1`,
		tenantID,
		waitlistEntryID,
	)
	return scanWaitlistEntryWithMeta(row)
}

func (s *Service) getWaitlistEntryTx(ctx context.Context, tx *sql.Tx, tenantID, waitlistEntryID string) (waitlistRow, error) {
	row := tx.QueryRowContext(
		ctx,
		waitlistSelectQuery+` WHERE w.tenant_id = ? AND w.id = ? LIMIT 1`,
		tenantID,
		waitlistEntryID,
	)
	return scanWaitlistRow(row)
}

const waitlistSelectQuery = `SELECT w.id, w.tenant_id, w.event_id, w.registration_id,
      COALESCE(r.participant_id, ''), COALESCE(p.name, ''), COALESCE(p.email, ''),
      w.position, w.status, COALESCE(r.status, ''),
      COALESCE(w.offered_at, ''), COALESCE(w.offer_expires_at, ''), COALESCE(w.accepted_at, ''),
      w.created_at, w.updated_at, COALESCE(e.title, ''), COALESCE(e.slug, ''), COALESCE(t.slug, '')
FROM waitlist_entries w
LEFT JOIN registrations r ON r.id = w.registration_id
LEFT JOIN participants p ON p.id = r.participant_id
LEFT JOIN events e ON e.id = w.event_id
LEFT JOIN tenants t ON t.id = w.tenant_id`

func scanWaitlistEntry(row interface{ Scan(dest ...any) error }) (WaitlistEntry, error) {
	raw, err := scanWaitlistRow(row)
	if err != nil {
		return WaitlistEntry{}, err
	}
	return mapWaitlistRow(raw)
}

func scanWaitlistEntryWithMeta(row interface{ Scan(dest ...any) error }) (WaitlistEntry, error) {
	raw, err := scanWaitlistRow(row)
	if err != nil {
		return WaitlistEntry{}, err
	}
	return mapWaitlistRow(raw)
}

func scanWaitlistRow(row interface{ Scan(dest ...any) error }) (waitlistRow, error) {
	var raw waitlistRow
	if err := row.Scan(
		&raw.ID,
		&raw.TenantID,
		&raw.EventID,
		&raw.RegistrationID,
		&raw.ParticipantID,
		&raw.ParticipantName,
		&raw.ParticipantEmail,
		&raw.Position,
		&raw.Status,
		&raw.RegistrationStatus,
		&raw.OfferedAtRaw,
		&raw.OfferExpiresAtRaw,
		&raw.AcceptedAtRaw,
		&raw.CreatedAtRaw,
		&raw.UpdatedAtRaw,
		&raw.EventTitle,
		&raw.EventSlug,
		&raw.TenantSlug,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return waitlistRow{}, ErrWaitlistEntryNotFound
		}
		return waitlistRow{}, fmt.Errorf("scan waitlist row: %w", err)
	}
	return raw, nil
}

func mapWaitlistRow(raw waitlistRow) (WaitlistEntry, error) {
	createdAt, err := time.Parse(time.RFC3339, raw.CreatedAtRaw)
	if err != nil {
		return WaitlistEntry{}, fmt.Errorf("parse waitlist created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, raw.UpdatedAtRaw)
	if err != nil {
		return WaitlistEntry{}, fmt.Errorf("parse waitlist updated_at: %w", err)
	}

	item := WaitlistEntry{
		ID:                 raw.ID,
		TenantID:           raw.TenantID,
		EventID:            raw.EventID,
		RegistrationID:     raw.RegistrationID,
		ParticipantID:      raw.ParticipantID,
		ParticipantName:    raw.ParticipantName,
		ParticipantEmail:   raw.ParticipantEmail,
		Position:           raw.Position,
		Status:             raw.Status,
		RegistrationStatus: raw.RegistrationStatus,
		CreatedAt:          createdAt.UTC(),
		UpdatedAt:          updatedAt.UTC(),
	}

	if strings.TrimSpace(raw.OfferedAtRaw) != "" {
		offeredAt, err := time.Parse(time.RFC3339, raw.OfferedAtRaw)
		if err != nil {
			return WaitlistEntry{}, fmt.Errorf("parse waitlist offered_at: %w", err)
		}
		offeredAt = offeredAt.UTC()
		item.OfferedAt = &offeredAt
	}
	if strings.TrimSpace(raw.OfferExpiresAtRaw) != "" {
		expiresAt, err := time.Parse(time.RFC3339, raw.OfferExpiresAtRaw)
		if err != nil {
			return WaitlistEntry{}, fmt.Errorf("parse waitlist offer_expires_at: %w", err)
		}
		expiresAt = expiresAt.UTC()
		item.OfferExpiresAt = &expiresAt
	}
	if strings.TrimSpace(raw.AcceptedAtRaw) != "" {
		acceptedAt, err := time.Parse(time.RFC3339, raw.AcceptedAtRaw)
		if err != nil {
			return WaitlistEntry{}, fmt.Errorf("parse waitlist accepted_at: %w", err)
		}
		acceptedAt = acceptedAt.UTC()
		item.AcceptedAt = &acceptedAt
	}
	return item, nil
}

func canOfferWaitlistStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case WaitlistStatusWaiting, WaitlistStatusOffered:
		return true
	default:
		return false
	}
}

func canPromoteWaitlistStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case WaitlistStatusWaiting, WaitlistStatusOffered:
		return true
	default:
		return false
	}
}

func canPromoteRegistrationStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case StatusWaitlist:
		return true
	default:
		return false
	}
}

func (s *Service) buildWaitlistOfferURL(tenantSlug, token string) string {
	base := strings.TrimRight(strings.TrimSpace(s.cfg.BaseURL), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/api/v1/public/%s/registrations/waitlist-offer?token=%s", base, tenantSlug, url.QueryEscape(strings.TrimSpace(token)))
}
