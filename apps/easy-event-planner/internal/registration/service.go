package registration

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/auth"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/invitation"
)

const (
	StatusVerificationPending = "verification_pending"
	StatusReserved            = "reserved"
	StatusPaymentPending      = "payment_pending"
	StatusConfirmed           = "confirmed"
	StatusWaitlist            = "waitlist"
	StatusCancelled           = "cancelled"
	StatusExpired             = "expired"
	StatusAttended            = "attended"
	StatusNoShow              = "no_show"

	WaitlistStatusWaiting = "waiting"

	SourcePublicPage = "public_page"
	SourceAdminManual = "admin_manual"

	DefaultVerificationTemplate = "registration_verify"
	DefaultConfirmedTemplate    = "registration_confirmed"
	DefaultWaitlistTemplate     = "registration_waitlist"
)

var (
	ErrEventNotFound                = errors.New("event not found")
	ErrRegistrationDisabled         = errors.New("registration is disabled")
	ErrRegistrationClosed           = errors.New("registration is closed for this event status")
	ErrPrivacyAcceptanceRequired    = errors.New("privacy acceptance is required")
	ErrAlreadyRegistered            = errors.New("participant is already registered")
	ErrAlreadyWaitlisted            = errors.New("participant is already on waitlist")
	ErrInvalidVerificationToken     = errors.New("invalid verification token")
	ErrExpiredVerificationToken     = errors.New("expired verification token")
	ErrRegistrationNotFound         = errors.New("registration not found")
	ErrRegistrationState            = errors.New("registration state does not allow verification")
	ErrRegistrationAttendNotAllowed = errors.New("registration state does not allow attendance mark")
	ErrEventFull                    = errors.New("event is full")
	ErrUnsupportedParticipation     = errors.New("unsupported participation type")
	ErrInvalidStartInput            = errors.New("invalid registration start input")
	ErrRegistrationVerificationNil  = errors.New("registration verification token is empty")
	ErrParticipantAccessDenied      = errors.New("participant access denied")
	ErrRegistrationCancelNotAllowed = errors.New("registration cannot be cancelled")
)

type Config struct {
	BaseURL          string
	TokenPepper      string
	RegistrationTTL  time.Duration
	WaitlistOfferTTL time.Duration
}

type Service struct {
	db                *sql.DB
	cfg               Config
	invitationService *invitation.Service
	nowFn             func() time.Time
	idFn              func(prefix string) string
	tokFn             func() (string, error)
}

type StartInput struct {
	TenantID          string
	TenantSlug        string
	EventID           string
	Name              string
	Email             string
	Phone             string
	ParticipationType string
	InviteCode        string
	InviteAmountCents int
	PrivacyAccepted   bool
	RequestIP         string
	UserAgent         string
}

type StartResult struct {
	RegistrationID      string
	ParticipantID       string
	EventID             string
	Status              string
	VerifyExpires       time.Time
	InviteID            string
	InviteCode          string
	DiscountAmountCents int
	CreditAmountCents   int
	FinalAmountCents    int
	Sponsored           bool
}

type VerifyInput struct {
	TenantID  string
	RawToken  string
	RequestIP string
	UserAgent string
}

type VerifyResult struct {
	RegistrationID string
	ParticipantID  string
	EventID        string
	Status         string
	ConfirmedAt    *time.Time
	WaitlistID     string
	WaitlistPos    int
}

type eventRecord struct {
	ID                  string
	TenantID            string
	Title               string
	Slug                string
	Status              string
	IsPublic            bool
	RegistrationEnabled bool
	WaitlistEnabled     bool
	MaxParticipants     sql.NullInt64
}

type registrationRecord struct {
	ID            string
	TenantID      string
	EventID       string
	ParticipantID string
	Status        string
}

type magicLinkRecord struct {
	ID             string
	TenantID       string
	ParticipantID  string
	RegistrationID string
	ExpiresAt      time.Time
	UsedAt         sql.NullString
}

func NewService(sqlDB *sql.DB, cfg Config) *Service {
	registrationTTL := cfg.RegistrationTTL
	if registrationTTL <= 0 {
		registrationTTL = 30 * time.Minute
	}
	waitlistOfferTTL := cfg.WaitlistOfferTTL
	if waitlistOfferTTL <= 0 {
		waitlistOfferTTL = 24 * time.Hour
	}

	return &Service{
		db: sqlDB,
		cfg: Config{
			BaseURL:          strings.TrimSpace(cfg.BaseURL),
			TokenPepper:      strings.TrimSpace(cfg.TokenPepper),
			RegistrationTTL:  registrationTTL,
			WaitlistOfferTTL: waitlistOfferTTL,
		},
		invitationService: invitation.NewService(sqlDB),
		nowFn:             func() time.Time { return time.Now().UTC() },
		idFn:              defaultID,
		tokFn:             randomToken,
	}
}

func (s *Service) Start(ctx context.Context, input StartInput) (StartResult, error) {
	if s.db == nil {
		return StartResult{}, fmt.Errorf("registration service database is nil")
	}
	tenantID := strings.TrimSpace(input.TenantID)
	if tenantID == "" {
		return StartResult{}, fmt.Errorf("%w: tenant id must not be empty", ErrInvalidStartInput)
	}
	tenantSlug := strings.TrimSpace(input.TenantSlug)
	if tenantSlug == "" {
		return StartResult{}, fmt.Errorf("%w: tenant slug must not be empty", ErrInvalidStartInput)
	}
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" {
		return StartResult{}, fmt.Errorf("%w: event id must not be empty", ErrInvalidStartInput)
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return StartResult{}, fmt.Errorf("%w: name must not be empty", ErrInvalidStartInput)
	}
	if len(name) > 180 {
		return StartResult{}, fmt.Errorf("%w: name must be <= 180 characters", ErrInvalidStartInput)
	}
	normalizedEmail, err := normalizeEmail(input.Email)
	if err != nil {
		return StartResult{}, err
	}
	phone := strings.TrimSpace(input.Phone)
	if len(phone) > 64 {
		return StartResult{}, fmt.Errorf("%w: phone must be <= 64 characters", ErrInvalidStartInput)
	}
	if !input.PrivacyAccepted {
		return StartResult{}, ErrPrivacyAcceptanceRequired
	}
	participationType, err := normalizeParticipationType(input.ParticipationType)
	if err != nil {
		return StartResult{}, err
	}
	inviteCode := strings.TrimSpace(input.InviteCode)
	if len(inviteCode) > 64 {
		return StartResult{}, fmt.Errorf("%w: invite_code must be <= 64 characters", ErrInvalidStartInput)
	}
	if input.InviteAmountCents < 0 {
		return StartResult{}, fmt.Errorf("%w: invite_amount_cents must be >= 0", ErrInvalidStartInput)
	}

	eventItem, err := s.lookupEvent(ctx, tenantID, eventID)
	if err != nil {
		return StartResult{}, err
	}
	if !eventItem.IsPublic {
		return StartResult{}, ErrEventNotFound
	}
	if !eventItem.RegistrationEnabled {
		return StartResult{}, ErrRegistrationDisabled
	}
	if !canRegisterForEventStatus(eventItem.Status) {
		return StartResult{}, ErrRegistrationClosed
	}

	now := s.nowFn().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return StartResult{}, fmt.Errorf("begin registration transaction: %w", err)
	}

	participantID, err := s.upsertParticipantTx(ctx, tx, tenantID, normalizedEmail, name, phone, now)
	if err != nil {
		_ = tx.Rollback()
		return StartResult{}, err
	}

	activeReg, found, err := s.lookupActiveRegistrationTx(ctx, tx, tenantID, eventID, participantID)
	if err != nil {
		_ = tx.Rollback()
		return StartResult{}, err
	}
	if found {
		switch activeReg.Status {
		case StatusVerificationPending:
			rawToken, expiresAt, resendErr := s.createVerificationMagicLinkTx(ctx, tx, tenantID, participantID, activeReg.ID, input.RequestIP, input.UserAgent, now)
			if resendErr != nil {
				_ = tx.Rollback()
				return StartResult{}, resendErr
			}
			verifyURL := s.buildVerifyURL(tenantSlug, rawToken)
			if queueErr := s.queueRegistrationMailTx(ctx, tx, tenantID, normalizedEmail, activeReg.ID, eventItem, verifyURL, expiresAt, DefaultVerificationTemplate, name, "verification_pending"); queueErr != nil {
				_ = tx.Rollback()
				return StartResult{}, queueErr
			}
			if err := tx.Commit(); err != nil {
				return StartResult{}, fmt.Errorf("commit registration start transaction: %w", err)
			}
			return StartResult{
				RegistrationID: activeReg.ID,
				ParticipantID:  participantID,
				EventID:        eventID,
				Status:         StatusVerificationPending,
				VerifyExpires:  expiresAt,
			}, nil
		case StatusConfirmed, StatusReserved, StatusPaymentPending:
			_ = tx.Rollback()
			return StartResult{}, ErrAlreadyRegistered
		case StatusWaitlist:
			_ = tx.Rollback()
			return StartResult{}, ErrAlreadyWaitlisted
		}
	}

	registrationID := s.idFn("reg")
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO registrations (
      id, tenant_id, event_id, participant_id, status, participation_type, quantity,
      source, privacy_accepted_at, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?)`,
		registrationID,
		tenantID,
		eventID,
		participantID,
		StatusVerificationPending,
		participationType,
		SourcePublicPage,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	); err != nil {
		_ = tx.Rollback()
		return StartResult{}, fmt.Errorf("insert registration: %w", err)
	}

	inviteResult := invitation.ResolveResult{}
	if inviteCode != "" && s.invitationService != nil {
		inviteResult, err = s.invitationService.ApplyCodeToRegistrationTx(ctx, tx, invitation.ApplyCodeToRegistrationInput{
			TenantID:         tenantID,
			EventID:          eventID,
			RegistrationID:   registrationID,
			ParticipantEmail: normalizedEmail,
			Code:             inviteCode,
			BaseAmountCents:  input.InviteAmountCents,
		})
		if err != nil {
			_ = tx.Rollback()
			return StartResult{}, err
		}
	}

	rawToken, expiresAt, err := s.createVerificationMagicLinkTx(ctx, tx, tenantID, participantID, registrationID, input.RequestIP, input.UserAgent, now)
	if err != nil {
		_ = tx.Rollback()
		return StartResult{}, err
	}
	verifyURL := s.buildVerifyURL(tenantSlug, rawToken)
	if queueErr := s.queueRegistrationMailTx(ctx, tx, tenantID, normalizedEmail, registrationID, eventItem, verifyURL, expiresAt, DefaultVerificationTemplate, name, "verification_pending"); queueErr != nil {
		_ = tx.Rollback()
		return StartResult{}, queueErr
	}

	if err := tx.Commit(); err != nil {
		return StartResult{}, fmt.Errorf("commit registration start transaction: %w", err)
	}

	return StartResult{
		RegistrationID:      registrationID,
		ParticipantID:       participantID,
		EventID:             eventID,
		Status:              StatusVerificationPending,
		VerifyExpires:       expiresAt,
		InviteID:            inviteResult.Link.ID,
		InviteCode:          inviteResult.Link.Code,
		DiscountAmountCents: inviteResult.DiscountAmountCents,
		CreditAmountCents:   inviteResult.CreditAmountCents,
		FinalAmountCents:    inviteResult.FinalAmountCents,
		Sponsored:           inviteResult.Sponsored,
	}, nil
}

func (s *Service) Verify(ctx context.Context, input VerifyInput) (VerifyResult, error) {
	if s.db == nil {
		return VerifyResult{}, fmt.Errorf("registration service database is nil")
	}
	tenantID := strings.TrimSpace(input.TenantID)
	if tenantID == "" {
		return VerifyResult{}, fmt.Errorf("tenant id must not be empty")
	}
	token := strings.TrimSpace(input.RawToken)
	if token == "" {
		return VerifyResult{}, ErrRegistrationVerificationNil
	}

	link, err := s.lookupVerificationMagicLink(ctx, tenantID, token)
	if err != nil {
		return VerifyResult{}, err
	}

	now := s.nowFn().UTC()
	if link.UsedAt.Valid {
		return VerifyResult{}, ErrInvalidVerificationToken
	}
	if now.After(link.ExpiresAt) {
		return VerifyResult{}, ErrExpiredVerificationToken
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("begin verification transaction: %w", err)
	}

	res, err := tx.ExecContext(
		ctx,
		`UPDATE magic_links SET used_at = ? WHERE id = ? AND used_at IS NULL`,
		now.Format(time.RFC3339),
		link.ID,
	)
	if err != nil {
		_ = tx.Rollback()
		return VerifyResult{}, fmt.Errorf("mark verification token used: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return VerifyResult{}, fmt.Errorf("fetch verification token rows affected: %w", err)
	}
	if rowsAffected != 1 {
		_ = tx.Rollback()
		return VerifyResult{}, ErrInvalidVerificationToken
	}

	registrationItem, err := s.lookupRegistrationByIDTx(ctx, tx, tenantID, link.RegistrationID)
	if err != nil {
		_ = tx.Rollback()
		return VerifyResult{}, err
	}
	if registrationItem.Status != StatusVerificationPending {
		_ = tx.Rollback()
		return VerifyResult{}, ErrRegistrationState
	}

	eventItem, err := s.lookupEventTx(ctx, tx, tenantID, registrationItem.EventID)
	if err != nil {
		_ = tx.Rollback()
		return VerifyResult{}, err
	}
	if !eventItem.RegistrationEnabled || !canRegisterForEventStatus(eventItem.Status) || !eventItem.IsPublic {
		if _, updErr := tx.ExecContext(
			ctx,
			`UPDATE registrations SET status = ?, updated_at = ? WHERE id = ?`,
			StatusExpired,
			now.Format(time.RFC3339),
			registrationItem.ID,
		); updErr != nil {
			_ = tx.Rollback()
			return VerifyResult{}, fmt.Errorf("expire registration for closed event: %w", updErr)
		}
		if err := tx.Commit(); err != nil {
			return VerifyResult{}, fmt.Errorf("commit closed registration verification: %w", err)
		}
		return VerifyResult{}, ErrRegistrationClosed
	}

	occupied, err := s.countOccupiedSeatsTx(ctx, tx, tenantID, registrationItem.EventID, now)
	if err != nil {
		_ = tx.Rollback()
		return VerifyResult{}, err
	}

	if eventItem.MaxParticipants.Valid && occupied >= int(eventItem.MaxParticipants.Int64) {
		if eventItem.WaitlistEnabled {
			waitlistID, position, waitErr := s.moveRegistrationToWaitlistTx(ctx, tx, registrationItem, now)
			if waitErr != nil {
				_ = tx.Rollback()
				return VerifyResult{}, waitErr
			}
			if queueErr := s.queueOutcomeMailTx(ctx, tx, tenantID, registrationItem.ID, eventItem, registrationItem.ParticipantID, DefaultWaitlistTemplate, now); queueErr != nil {
				_ = tx.Rollback()
				return VerifyResult{}, queueErr
			}
			if err := tx.Commit(); err != nil {
				return VerifyResult{}, fmt.Errorf("commit waitlist verification: %w", err)
			}
			return VerifyResult{
				RegistrationID: registrationItem.ID,
				ParticipantID:  registrationItem.ParticipantID,
				EventID:        registrationItem.EventID,
				Status:         StatusWaitlist,
				WaitlistID:     waitlistID,
				WaitlistPos:    position,
			}, nil
		}

		if _, updErr := tx.ExecContext(
			ctx,
			`UPDATE registrations SET status = ?, updated_at = ? WHERE id = ?`,
			StatusExpired,
			now.Format(time.RFC3339),
			registrationItem.ID,
		); updErr != nil {
			_ = tx.Rollback()
			return VerifyResult{}, fmt.Errorf("expire registration for full event: %w", updErr)
		}
		if err := tx.Commit(); err != nil {
			return VerifyResult{}, fmt.Errorf("commit full registration verification: %w", err)
		}
		return VerifyResult{}, ErrEventFull
	}

	confirmedAtRaw := now.Format(time.RFC3339)
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE registrations
     SET status = ?, confirmed_at = ?, reserved_until = NULL, updated_at = ?
     WHERE id = ?`,
		StatusConfirmed,
		confirmedAtRaw,
		confirmedAtRaw,
		registrationItem.ID,
	); err != nil {
		_ = tx.Rollback()
		return VerifyResult{}, fmt.Errorf("confirm registration: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE participants
     SET email_verified_at = CASE WHEN email_verified_at IS NULL THEN ? ELSE email_verified_at END,
         updated_at = ?
     WHERE id = ?`,
		confirmedAtRaw,
		confirmedAtRaw,
		registrationItem.ParticipantID,
	); err != nil {
		_ = tx.Rollback()
		return VerifyResult{}, fmt.Errorf("mark participant verified: %w", err)
	}

	if queueErr := s.queueOutcomeMailTx(ctx, tx, tenantID, registrationItem.ID, eventItem, registrationItem.ParticipantID, DefaultConfirmedTemplate, now); queueErr != nil {
		_ = tx.Rollback()
		return VerifyResult{}, queueErr
	}

	if err := tx.Commit(); err != nil {
		return VerifyResult{}, fmt.Errorf("commit registration verification: %w", err)
	}

	confirmedAt := now
	return VerifyResult{
		RegistrationID: registrationItem.ID,
		ParticipantID:  registrationItem.ParticipantID,
		EventID:        registrationItem.EventID,
		Status:         StatusConfirmed,
		ConfirmedAt:    &confirmedAt,
	}, nil
}

func (s *Service) lookupEvent(ctx context.Context, tenantID, eventID string) (eventRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, title, slug, status, is_public, registration_enabled, waitlist_enabled, max_participants
     FROM events
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenantID,
		eventID,
	)
	return scanEvent(row)
}

func (s *Service) lookupEventTx(ctx context.Context, tx *sql.Tx, tenantID, eventID string) (eventRecord, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, title, slug, status, is_public, registration_enabled, waitlist_enabled, max_participants
     FROM events
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenantID,
		eventID,
	)
	return scanEvent(row)
}

func scanEvent(row interface{ Scan(dest ...any) error }) (eventRecord, error) {
	var (
		item               eventRecord
		isPublicInt        int
		registrationInt    int
		waitlistEnabledInt int
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.Title,
		&item.Slug,
		&item.Status,
		&isPublicInt,
		&registrationInt,
		&waitlistEnabledInt,
		&item.MaxParticipants,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return eventRecord{}, ErrEventNotFound
		}
		return eventRecord{}, fmt.Errorf("query event: %w", err)
	}
	item.IsPublic = isPublicInt == 1
	item.RegistrationEnabled = registrationInt == 1
	item.WaitlistEnabled = waitlistEnabledInt == 1
	return item, nil
}

func (s *Service) upsertParticipantTx(ctx context.Context, tx *sql.Tx, tenantID, email, name, phone string, now time.Time) (string, error) {
	var participantID string
	row := tx.QueryRowContext(
		ctx,
		`SELECT id
     FROM participants
     WHERE tenant_id = ? AND lower(email) = ?
     LIMIT 1`,
		tenantID,
		strings.ToLower(email),
	)
	if err := row.Scan(&participantID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("lookup participant by email: %w", err)
	}
	if participantID != "" {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE participants
       SET name = ?, phone = ?, updated_at = ?
       WHERE id = ?`,
			name,
			nullable(phone),
			now.Format(time.RFC3339),
			participantID,
		); err != nil {
			return "", fmt.Errorf("update participant: %w", err)
		}
		return participantID, nil
	}

	participantID = s.idFn("par")
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO participants (
      id, tenant_id, email, phone, name, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		participantID,
		tenantID,
		strings.ToLower(email),
		nullable(phone),
		name,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	); err != nil {
		return "", fmt.Errorf("insert participant: %w", err)
	}
	return participantID, nil
}

func (s *Service) lookupActiveRegistrationTx(ctx context.Context, tx *sql.Tx, tenantID, eventID, participantID string) (registrationRecord, bool, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, event_id, participant_id, status
     FROM registrations
     WHERE tenant_id = ? AND event_id = ? AND participant_id = ?
       AND status IN (?, ?, ?, ?, ?)
     ORDER BY created_at DESC
     LIMIT 1`,
		tenantID,
		eventID,
		participantID,
		StatusVerificationPending,
		StatusConfirmed,
		StatusReserved,
		StatusPaymentPending,
		StatusWaitlist,
	)

	var item registrationRecord
	if err := row.Scan(&item.ID, &item.TenantID, &item.EventID, &item.ParticipantID, &item.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return registrationRecord{}, false, nil
		}
		return registrationRecord{}, false, fmt.Errorf("lookup active registration: %w", err)
	}
	return item, true, nil
}

func (s *Service) createVerificationMagicLinkTx(ctx context.Context, tx *sql.Tx, tenantID, participantID, registrationID, requestIP, userAgent string, now time.Time) (string, time.Time, error) {
	rawToken, err := s.tokFn()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate registration verify token: %w", err)
	}
	expiresAt := now.Add(s.cfg.RegistrationTTL)
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO magic_links (
      id, tenant_id, user_id, participant_id, purpose, token_hash, redirect_path, expires_at, request_ip, user_agent, created_at
    ) VALUES (?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.idFn("mlk"),
		tenantID,
		participantID,
		auth.PurposeRegistrationVerify,
		s.hash(rawToken),
		registrationID,
		expiresAt.Format(time.RFC3339),
		nullable(strings.TrimSpace(requestIP)),
		nullable(strings.TrimSpace(userAgent)),
		now.Format(time.RFC3339),
	); err != nil {
		return "", time.Time{}, fmt.Errorf("insert registration verify magic link: %w", err)
	}
	return rawToken, expiresAt, nil
}

func (s *Service) queueRegistrationMailTx(ctx context.Context, tx *sql.Tx, tenantID, recipientEmail, registrationID string, eventItem eventRecord, verifyURL string, expiresAt time.Time, templateKey, recipientName, targetStatus string) error {
	subject := "Bitte bestaetige deine Anmeldung"
	bodyText := fmt.Sprintf(
		"Hallo %s,\n\nbitte bestaetige deine Anmeldung fuer \"%s\".\n\nLink: %s\n\nDer Link ist gueltig bis %s.\n",
		strings.TrimSpace(recipientName),
		eventItem.Title,
		verifyURL,
		expiresAt.UTC().Format(time.RFC3339),
	)
	return s.queueEmailJobTx(ctx, tx, tenantID, templateKey, recipientEmail, subject, bodyText, map[string]any{
		"registration_id": registrationID,
		"event_id":        eventItem.ID,
		"event_slug":      eventItem.Slug,
		"target_status":   targetStatus,
	})
}

func (s *Service) queueOutcomeMailTx(ctx context.Context, tx *sql.Tx, tenantID, registrationID string, eventItem eventRecord, participantID, templateKey string, now time.Time) error {
	row := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(email, ''), COALESCE(name, '')
     FROM participants
     WHERE id = ?
     LIMIT 1`,
		participantID,
	)
	var recipientEmail string
	var recipientName string
	if err := row.Scan(&recipientEmail, &recipientName); err != nil {
		return fmt.Errorf("lookup participant for outcome mail: %w", err)
	}
	if strings.TrimSpace(recipientEmail) == "" {
		return nil
	}

	subject := "Deine Anmeldung wurde bestaetigt"
	bodyText := fmt.Sprintf(
		"Hallo %s,\n\ndeine Anmeldung fuer \"%s\" wurde bestaetigt.\n",
		strings.TrimSpace(recipientName),
		eventItem.Title,
	)
	targetStatus := StatusConfirmed
	if templateKey == DefaultWaitlistTemplate {
		subject = "Du stehst auf der Warteliste"
		bodyText = fmt.Sprintf(
			"Hallo %s,\n\ndie Veranstaltung \"%s\" ist derzeit voll. Du stehst jetzt auf der Warteliste.\n",
			strings.TrimSpace(recipientName),
			eventItem.Title,
		)
		targetStatus = StatusWaitlist
	}

	return s.queueEmailJobTx(ctx, tx, tenantID, templateKey, recipientEmail, subject, bodyText, map[string]any{
		"registration_id": registrationID,
		"event_id":        eventItem.ID,
		"event_slug":      eventItem.Slug,
		"target_status":   targetStatus,
		"processed_at":    now.Format(time.RFC3339),
	})
}

func (s *Service) queueEmailJobTx(ctx context.Context, tx *sql.Tx, tenantID, templateKey, recipient, subject, bodyText string, metadata map[string]any) error {
	metadataJSON := ""
	if len(metadata) > 0 {
		payload, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("marshal email metadata: %w", err)
		}
		metadataJSON = string(payload)
	}
	now := s.nowFn().UTC().Format(time.RFC3339)
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO email_jobs (
      id, tenant_id, template_key, recipient_email, subject, body_text, body_html, status,
      scheduled_for, metadata_json, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, NULL, 'queued', NULL, ?, ?, ?)`,
		s.idFn("emj"),
		tenantID,
		templateKey,
		recipient,
		subject,
		bodyText,
		nullable(metadataJSON),
		now,
		now,
	); err != nil {
		return fmt.Errorf("insert email job: %w", err)
	}
	return nil
}

func (s *Service) lookupVerificationMagicLink(ctx context.Context, tenantID, rawToken string) (magicLinkRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, COALESCE(participant_id, ''), COALESCE(redirect_path, ''), expires_at, used_at
     FROM magic_links
     WHERE tenant_id = ? AND purpose = ? AND token_hash = ?
     LIMIT 1`,
		tenantID,
		auth.PurposeRegistrationVerify,
		s.hash(rawToken),
	)
	var (
		item         magicLinkRecord
		expiresAtRaw string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.ParticipantID,
		&item.RegistrationID,
		&expiresAtRaw,
		&item.UsedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return magicLinkRecord{}, ErrInvalidVerificationToken
		}
		return magicLinkRecord{}, fmt.Errorf("query registration magic link: %w", err)
	}
	if strings.TrimSpace(item.RegistrationID) == "" {
		return magicLinkRecord{}, ErrInvalidVerificationToken
	}
	expiresAt, err := time.Parse(time.RFC3339, expiresAtRaw)
	if err != nil {
		return magicLinkRecord{}, fmt.Errorf("parse registration magic link expires_at: %w", err)
	}
	item.ExpiresAt = expiresAt.UTC()
	return item, nil
}

func (s *Service) lookupRegistrationByIDTx(ctx context.Context, tx *sql.Tx, tenantID, registrationID string) (registrationRecord, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, event_id, participant_id, status
     FROM registrations
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenantID,
		registrationID,
	)
	var item registrationRecord
	if err := row.Scan(&item.ID, &item.TenantID, &item.EventID, &item.ParticipantID, &item.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return registrationRecord{}, ErrRegistrationNotFound
		}
		return registrationRecord{}, fmt.Errorf("query registration: %w", err)
	}
	return item, nil
}

func (s *Service) countOccupiedSeatsTx(ctx context.Context, tx *sql.Tx, tenantID, eventID string, now time.Time) (int, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
     FROM registrations
     WHERE tenant_id = ?
       AND event_id = ?
       AND status IN (?, ?, ?)
       AND (reserved_until IS NULL OR reserved_until > ?)`,
		tenantID,
		eventID,
		StatusConfirmed,
		StatusReserved,
		StatusPaymentPending,
		now.Format(time.RFC3339),
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count occupied seats: %w", err)
	}
	return count, nil
}

func (s *Service) moveRegistrationToWaitlistTx(ctx context.Context, tx *sql.Tx, registrationItem registrationRecord, now time.Time) (string, int, error) {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE registrations
     SET status = ?, updated_at = ?
     WHERE id = ?`,
		StatusWaitlist,
		now.Format(time.RFC3339),
		registrationItem.ID,
	); err != nil {
		return "", 0, fmt.Errorf("set registration waitlist status: %w", err)
	}

	row := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(MAX(position), 0) + 1
     FROM waitlist_entries
     WHERE tenant_id = ? AND event_id = ?`,
		registrationItem.TenantID,
		registrationItem.EventID,
	)
	var position int
	if err := row.Scan(&position); err != nil {
		return "", 0, fmt.Errorf("calculate waitlist position: %w", err)
	}
	waitlistID := s.idFn("wle")
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO waitlist_entries (
      id, tenant_id, event_id, registration_id, position, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		waitlistID,
		registrationItem.TenantID,
		registrationItem.EventID,
		registrationItem.ID,
		position,
		WaitlistStatusWaiting,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	); err != nil {
		return "", 0, fmt.Errorf("insert waitlist entry: %w", err)
	}
	return waitlistID, position, nil
}

func canRegisterForEventStatus(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case event.EventStatusScheduled, event.EventStatusChanged, event.EventStatusPostponed:
		return true
	default:
		return false
	}
}

func normalizeEmail(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return "", fmt.Errorf("%w: email must not be empty", ErrInvalidStartInput)
	}
	if len(value) > 254 {
		return "", fmt.Errorf("%w: email must be <= 254 characters", ErrInvalidStartInput)
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return "", fmt.Errorf("%w: invalid email: %v", ErrInvalidStartInput, err)
	}
	return value, nil
}

func normalizeParticipationType(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return event.ParticipationModeOnsite, nil
	}
	switch value {
	case event.ParticipationModeOnsite, event.ParticipationModeOnline, event.ParticipationModeHybrid:
		return value, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedParticipation, raw)
	}
}

func (s *Service) buildVerifyURL(tenantSlug, token string) string {
	base := strings.TrimRight(strings.TrimSpace(s.cfg.BaseURL), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/api/v1/public/%s/registrations/verify?token=%s", base, tenantSlug, url.QueryEscape(strings.TrimSpace(token)))
}

func (s *Service) hash(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(s.cfg.TokenPepper) + ":" + strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func nullable(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
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

func randomToken() (string, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(random[:]), nil
}
