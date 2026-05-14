package auth

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
	"net/url"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

const (
	PurposeOrganizerLogin     = "organizer_login"
	PurposeRegistrationVerify = "registration_verify"
	PurposeParticipantLogin   = "participant_login"
	PurposeWaitlistOffer      = "waitlist_offer"
	PurposeRegistrationCancel = "registration_cancel"
	PurposeCertificateAccess  = "certificate_access"
)

var (
	ErrRateLimited        = errors.New("rate limit exceeded")
	ErrInvalidMagicLink   = errors.New("invalid magic link")
	ErrSessionNotFound    = errors.New("session not found")
	ErrUnsupportedPurpose = errors.New("unsupported purpose")
)

type Config struct {
	BaseURL           string
	TokenPepper       string
	SessionTTL        time.Duration
	MagicLinkTTL      time.Duration
	RegistrationTTL   time.Duration
	WaitlistOfferTTL  time.Duration
	CertificateTTL    time.Duration
	RateLimitRequests int
	RateLimitWindow   time.Duration
}

type Service struct {
	db      *sql.DB
	tenants *tenant.Repository
	cfg     Config
	sender  Sender
	limiter *RateLimiter
	nowFn   func() time.Time
	idFn    func(prefix string) string
	tokenFn func() (string, error)
}

type RequestMagicLinkInput struct {
	TenantSlug   string
	Email        string
	Purpose      string
	RedirectPath string
	RequestIP    string
	UserAgent    string
}

type RequestMagicLinkResult struct {
	Accepted bool
	Sent     bool
}

type VerifyMagicLinkInput struct {
	RawToken  string
	RequestIP string
	UserAgent string
}

type VerifyMagicLinkResult struct {
	Purpose          string
	TenantID         string
	UserID           string
	ParticipantID    string
	RedirectPath     string
	SessionToken     string
	SessionExpiresAt time.Time
}

type SessionPrincipal struct {
	SessionID        string
	TenantID         string
	TenantSlug       string
	UserID           string
	Email            string
	Name             string
	Role             string
	SessionExpiresAt time.Time
}

type ParticipantSessionPrincipal struct {
	SessionID        string
	TenantID         string
	TenantSlug       string
	ParticipantID    string
	Email            string
	Name             string
	SessionExpiresAt time.Time
}

type auditEvent struct {
	TenantID      string
	UserID        string
	ParticipantID string
	Action        string
	Details       map[string]any
	RequestIP     string
	UserAgent     string
}

func NewService(sqlDB *sql.DB, tenantRepo *tenant.Repository, cfg Config, sender Sender) *Service {
	if sender == nil {
		sender = &LogSender{}
	}
	return &Service{
		db:      sqlDB,
		tenants: tenantRepo,
		cfg:     cfg,
		sender:  sender,
		limiter: NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow),
		nowFn:   func() time.Time { return time.Now().UTC() },
		idFn:    defaultID,
		tokenFn: randomToken,
	}
}

func (s *Service) SetSender(sender Sender) {
	if sender == nil {
		return
	}
	s.sender = sender
}

func (s *Service) RequestMagicLink(ctx context.Context, input RequestMagicLinkInput) (RequestMagicLinkResult, error) {
	if s.db == nil {
		return RequestMagicLinkResult{}, fmt.Errorf("auth service database is nil")
	}
	purpose, err := normalizePurpose(input.Purpose)
	if err != nil {
		return RequestMagicLinkResult{}, err
	}

	tenantSlug := strings.TrimSpace(input.TenantSlug)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	now := s.nowFn().UTC()

	rateKeys := []string{
		fmt.Sprintf("ip:%s|purpose:%s", strings.TrimSpace(input.RequestIP), purpose),
		fmt.Sprintf("tenant:%s|email:%s|purpose:%s", tenantSlug, email, purpose),
		fmt.Sprintf("tenant:%s|purpose:%s", tenantSlug, purpose),
	}
	for _, key := range rateKeys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if !s.limiter.Allow(key, now) {
			_ = s.writeAudit(ctx, auditEvent{
				Action:    "magic_link_rejected",
				Details:   map[string]any{"reason": "rate_limited", "purpose": purpose, "tenant_slug": tenantSlug},
				RequestIP: input.RequestIP,
				UserAgent: input.UserAgent,
			})
			return RequestMagicLinkResult{}, ErrRateLimited
		}
	}

	tenantRecord, err := s.tenants.LookupBySlug(ctx, tenantSlug)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) {
			return RequestMagicLinkResult{Accepted: true, Sent: false}, nil
		}
		return RequestMagicLinkResult{}, err
	}

	userID := ""
	participantID := ""
	switch purpose {
	case PurposeParticipantLogin:
		participantID, err = s.lookupPortalParticipantByEmail(ctx, tenantRecord.ID, email)
		if err != nil {
			return RequestMagicLinkResult{}, err
		}
	default:
		userID, err = s.lookupActiveTenantUser(ctx, tenantRecord.ID, email)
		if err != nil {
			return RequestMagicLinkResult{}, err
		}
	}

	_ = s.writeAudit(ctx, auditEvent{
		TenantID:      tenantRecord.ID,
		UserID:        userID,
		ParticipantID: participantID,
		Action:        "magic_link_requested",
		Details:       map[string]any{"purpose": purpose},
		RequestIP:     input.RequestIP,
		UserAgent:     input.UserAgent,
	})

	if userID == "" && participantID == "" {
		return RequestMagicLinkResult{Accepted: true, Sent: false}, nil
	}

	rawToken, err := s.tokenFn()
	if err != nil {
		return RequestMagicLinkResult{}, fmt.Errorf("generate magic link token: %w", err)
	}

	tokenHash := s.hash(rawToken)
	expiresAt := now.Add(s.ttlForPurpose(purpose))
	redirectPath := sanitizeRedirectPath(input.RedirectPath)
	if purpose == PurposeParticipantLogin && strings.TrimSpace(input.RedirectPath) == "" {
		redirectPath = "/portal"
	}
	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO magic_links (
      id, tenant_id, user_id, participant_id, purpose, token_hash, redirect_path, expires_at, request_ip, user_agent, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.idFn("mlk"),
		tenantRecord.ID,
		nullable(userID),
		nullable(participantID),
		purpose,
		tokenHash,
		redirectPath,
		expiresAt.Format(time.RFC3339),
		strings.TrimSpace(input.RequestIP),
		strings.TrimSpace(input.UserAgent),
		now.Format(time.RFC3339),
	); err != nil {
		return RequestMagicLinkResult{}, fmt.Errorf("insert magic link: %w", err)
	}

	verifyURL := fmt.Sprintf(
		"%s/api/v1/auth/magic-link/verify?token=%s",
		strings.TrimRight(s.cfg.BaseURL, "/"),
		url.QueryEscape(rawToken),
	)
	if purpose == PurposeParticipantLogin {
		verifyURL = fmt.Sprintf(
			"%s/api/v1/public/%s/participants/portal/verify?token=%s",
			strings.TrimRight(s.cfg.BaseURL, "/"),
			url.PathEscape(tenantRecord.Slug),
			url.QueryEscape(rawToken),
		)
	}
	if err := s.sender.SendMagicLink(ctx, MagicLinkMessage{
		ToEmail:    email,
		TenantSlug: tenantRecord.Slug,
		Purpose:    purpose,
		VerifyURL:  verifyURL,
		ExpiresAt:  expiresAt,
	}); err != nil {
		return RequestMagicLinkResult{}, fmt.Errorf("send magic link: %w", err)
	}

	_ = s.writeAudit(ctx, auditEvent{
		TenantID:      tenantRecord.ID,
		UserID:        userID,
		ParticipantID: participantID,
		Action:        "magic_link_sent",
		Details:       map[string]any{"purpose": purpose},
		RequestIP:     input.RequestIP,
		UserAgent:     input.UserAgent,
	})

	return RequestMagicLinkResult{Accepted: true, Sent: true}, nil
}

func (s *Service) VerifyMagicLink(ctx context.Context, input VerifyMagicLinkInput) (VerifyMagicLinkResult, error) {
	rawToken := strings.TrimSpace(input.RawToken)
	if rawToken == "" {
		return VerifyMagicLinkResult{}, ErrInvalidMagicLink
	}

	link, err := s.lookupMagicLinkByToken(ctx, rawToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = s.writeAudit(ctx, auditEvent{
				Action:    "magic_link_rejected",
				Details:   map[string]any{"reason": "unknown_token"},
				RequestIP: input.RequestIP,
				UserAgent: input.UserAgent,
			})
			return VerifyMagicLinkResult{}, ErrInvalidMagicLink
		}
		return VerifyMagicLinkResult{}, err
	}

	now := s.nowFn().UTC()
	if link.usedAt.Valid || now.After(link.expiresAt) {
		reason := "used"
		if now.After(link.expiresAt) {
			reason = "expired"
		}
		_ = s.writeAudit(ctx, auditEvent{
			TenantID:      link.tenantID,
			UserID:        link.userID,
			ParticipantID: link.participantID,
			Action:        "magic_link_rejected",
			Details:       map[string]any{"reason": reason, "purpose": link.purpose},
			RequestIP:     input.RequestIP,
			UserAgent:     input.UserAgent,
		})
		return VerifyMagicLinkResult{}, ErrInvalidMagicLink
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return VerifyMagicLinkResult{}, fmt.Errorf("begin verify transaction: %w", err)
	}

	updateResult, err := tx.ExecContext(
		ctx,
		`UPDATE magic_links SET used_at = ? WHERE id = ? AND used_at IS NULL`,
		now.Format(time.RFC3339),
		link.id,
	)
	if err != nil {
		_ = tx.Rollback()
		return VerifyMagicLinkResult{}, fmt.Errorf("mark magic link used: %w", err)
	}
	rowsAffected, err := updateResult.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return VerifyMagicLinkResult{}, fmt.Errorf("fetch verify rows affected: %w", err)
	}
	if rowsAffected != 1 {
		_ = tx.Rollback()
		return VerifyMagicLinkResult{}, ErrInvalidMagicLink
	}

	result := VerifyMagicLinkResult{
		Purpose:       link.purpose,
		TenantID:      link.tenantID,
		UserID:        link.userID,
		ParticipantID: link.participantID,
		RedirectPath:  sanitizeRedirectPath(link.redirectPath),
	}

	if link.userID != "" || link.participantID != "" {
		sessionToken, tokenErr := s.tokenFn()
		if tokenErr != nil {
			_ = tx.Rollback()
			return VerifyMagicLinkResult{}, fmt.Errorf("generate session token: %w", tokenErr)
		}

		sessionExpiresAt := now.Add(s.cfg.SessionTTL)
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO sessions (
        id, tenant_id, user_id, participant_id, session_hash, expires_at, created_at, last_seen_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			s.idFn("ses"),
			link.tenantID,
			nullable(link.userID),
			nullable(link.participantID),
			s.hash(sessionToken),
			sessionExpiresAt.Format(time.RFC3339),
			now.Format(time.RFC3339),
			now.Format(time.RFC3339),
		); err != nil {
			_ = tx.Rollback()
			return VerifyMagicLinkResult{}, fmt.Errorf("insert session: %w", err)
		}

		result.SessionToken = sessionToken
		result.SessionExpiresAt = sessionExpiresAt
	}

	if err := tx.Commit(); err != nil {
		return VerifyMagicLinkResult{}, fmt.Errorf("commit verify transaction: %w", err)
	}

	_ = s.writeAudit(ctx, auditEvent{
		TenantID:      link.tenantID,
		UserID:        link.userID,
		ParticipantID: link.participantID,
		Action:        "magic_link_verified",
		Details:       map[string]any{"purpose": link.purpose},
		RequestIP:     input.RequestIP,
		UserAgent:     input.UserAgent,
	})
	if result.SessionToken != "" {
		_ = s.writeAudit(ctx, auditEvent{
			TenantID:      link.tenantID,
			UserID:        link.userID,
			ParticipantID: link.participantID,
			Action:        "session_created",
			Details:       map[string]any{"purpose": link.purpose},
			RequestIP:     input.RequestIP,
			UserAgent:     input.UserAgent,
		})
	}

	return result, nil
}

func (s *Service) AuthenticateSession(ctx context.Context, rawSessionToken string) (SessionPrincipal, error) {
	token := strings.TrimSpace(rawSessionToken)
	if token == "" {
		return SessionPrincipal{}, ErrSessionNotFound
	}
	row := s.db.QueryRowContext(
		ctx,
		`SELECT s.id, s.tenant_id, COALESCE(s.user_id, ''), s.expires_at, s.revoked_at,
            COALESCE(t.slug, ''), COALESCE(u.email, ''), COALESCE(u.name, ''), COALESCE(u.role, '')
     FROM sessions s
     LEFT JOIN tenants t ON t.id = s.tenant_id
     LEFT JOIN tenant_users u ON u.id = s.user_id
     WHERE s.session_hash = ?
     LIMIT 1`,
		s.hash(token),
	)

	var (
		principal    SessionPrincipal
		expiresAtRaw string
		revokedAtRaw sql.NullString
	)
	if err := row.Scan(
		&principal.SessionID,
		&principal.TenantID,
		&principal.UserID,
		&expiresAtRaw,
		&revokedAtRaw,
		&principal.TenantSlug,
		&principal.Email,
		&principal.Name,
		&principal.Role,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionPrincipal{}, ErrSessionNotFound
		}
		return SessionPrincipal{}, fmt.Errorf("query session: %w", err)
	}
	if strings.TrimSpace(principal.UserID) == "" {
		return SessionPrincipal{}, ErrSessionNotFound
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtRaw)
	if err != nil {
		return SessionPrincipal{}, fmt.Errorf("parse session expires_at: %w", err)
	}
	if revokedAtRaw.Valid || s.nowFn().UTC().After(expiresAt) {
		return SessionPrincipal{}, ErrSessionNotFound
	}

	principal.SessionExpiresAt = expiresAt
	_, _ = s.db.ExecContext(
		ctx,
		`UPDATE sessions SET last_seen_at = ? WHERE id = ?`,
		s.nowFn().UTC().Format(time.RFC3339),
		principal.SessionID,
	)

	return principal, nil
}

func (s *Service) AuthenticateParticipantSession(ctx context.Context, rawSessionToken string) (ParticipantSessionPrincipal, error) {
	token := strings.TrimSpace(rawSessionToken)
	if token == "" {
		return ParticipantSessionPrincipal{}, ErrSessionNotFound
	}
	row := s.db.QueryRowContext(
		ctx,
		`SELECT s.id, s.tenant_id, s.participant_id, s.expires_at, s.revoked_at,
            COALESCE(t.slug, ''), COALESCE(p.email, ''), COALESCE(p.name, '')
     FROM sessions s
     LEFT JOIN tenants t ON t.id = s.tenant_id
     LEFT JOIN participants p ON p.id = s.participant_id
     WHERE s.session_hash = ?
     LIMIT 1`,
		s.hash(token),
	)

	var (
		principal    ParticipantSessionPrincipal
		expiresAtRaw string
		revokedAtRaw sql.NullString
	)
	if err := row.Scan(
		&principal.SessionID,
		&principal.TenantID,
		&principal.ParticipantID,
		&expiresAtRaw,
		&revokedAtRaw,
		&principal.TenantSlug,
		&principal.Email,
		&principal.Name,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ParticipantSessionPrincipal{}, ErrSessionNotFound
		}
		return ParticipantSessionPrincipal{}, fmt.Errorf("query participant session: %w", err)
	}
	if strings.TrimSpace(principal.ParticipantID) == "" {
		return ParticipantSessionPrincipal{}, ErrSessionNotFound
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtRaw)
	if err != nil {
		return ParticipantSessionPrincipal{}, fmt.Errorf("parse participant session expires_at: %w", err)
	}
	if revokedAtRaw.Valid || s.nowFn().UTC().After(expiresAt) {
		return ParticipantSessionPrincipal{}, ErrSessionNotFound
	}

	principal.SessionExpiresAt = expiresAt
	_, _ = s.db.ExecContext(
		ctx,
		`UPDATE sessions SET last_seen_at = ? WHERE id = ?`,
		s.nowFn().UTC().Format(time.RFC3339),
		principal.SessionID,
	)

	return principal, nil
}

func (s *Service) RevokeSession(ctx context.Context, rawSessionToken, requestIP, userAgent string) (bool, error) {
	token := strings.TrimSpace(rawSessionToken)
	if token == "" {
		return false, nil
	}

	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, COALESCE(user_id, ''), COALESCE(participant_id, ''), revoked_at
     FROM sessions
     WHERE session_hash = ?
     LIMIT 1`,
		s.hash(token),
	)

	var (
		sessionID     string
		tenantID      string
		userID        string
		participantID string
		revokedAt     sql.NullString
	)
	if err := row.Scan(&sessionID, &tenantID, &userID, &participantID, &revokedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query session for revoke: %w", err)
	}
	if revokedAt.Valid {
		return false, nil
	}

	updateResult, err := s.db.ExecContext(
		ctx,
		`UPDATE sessions SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`,
		s.nowFn().UTC().Format(time.RFC3339),
		sessionID,
	)
	if err != nil {
		return false, fmt.Errorf("revoke session: %w", err)
	}
	rowsAffected, err := updateResult.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("fetch revoke rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return false, nil
	}

	_ = s.writeAudit(ctx, auditEvent{
		TenantID:      tenantID,
		UserID:        userID,
		ParticipantID: participantID,
		Action:        "session_revoked",
		RequestIP:     requestIP,
		UserAgent:     userAgent,
	})
	return true, nil
}

type magicLinkRecord struct {
	id            string
	tenantID      string
	userID        string
	participantID string
	purpose       string
	redirectPath  string
	expiresAt     time.Time
	usedAt        sql.NullString
}

func (s *Service) lookupMagicLinkByToken(ctx context.Context, rawToken string) (magicLinkRecord, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, COALESCE(user_id, ''), COALESCE(participant_id, ''), purpose, redirect_path, expires_at, used_at
     FROM magic_links
     WHERE token_hash = ?
     LIMIT 1`,
		s.hash(rawToken),
	)

	var (
		link         magicLinkRecord
		expiresAtRaw string
	)
	if err := row.Scan(
		&link.id,
		&link.tenantID,
		&link.userID,
		&link.participantID,
		&link.purpose,
		&link.redirectPath,
		&expiresAtRaw,
		&link.usedAt,
	); err != nil {
		return magicLinkRecord{}, err
	}
	expiresAt, err := time.Parse(time.RFC3339, expiresAtRaw)
	if err != nil {
		return magicLinkRecord{}, fmt.Errorf("parse magic link expires_at: %w", err)
	}
	link.expiresAt = expiresAt
	return link, nil
}

func (s *Service) lookupActiveTenantUser(ctx context.Context, tenantID, email string) (string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return "", nil
	}

	row := s.db.QueryRowContext(
		ctx,
		`SELECT id
     FROM tenant_users
     WHERE tenant_id = ? AND lower(email) = ? AND status = 'active'
     LIMIT 1`,
		tenantID,
		normalizedEmail,
	)

	var userID string
	if err := row.Scan(&userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("query tenant user by email: %w", err)
	}
	return userID, nil
}

func (s *Service) lookupPortalParticipantByEmail(ctx context.Context, tenantID, email string) (string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return "", nil
	}

	row := s.db.QueryRowContext(
		ctx,
		`SELECT p.id
     FROM participants p
     WHERE p.tenant_id = ?
       AND lower(COALESCE(p.email, '')) = ?
       AND p.anonymized_at IS NULL
       AND EXISTS (
         SELECT 1
         FROM registrations r
         WHERE r.tenant_id = p.tenant_id
           AND r.participant_id = p.id
       )
     LIMIT 1`,
		tenantID,
		normalizedEmail,
	)

	var participantID string
	if err := row.Scan(&participantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("query participant by email: %w", err)
	}
	return participantID, nil
}

func (s *Service) ttlForPurpose(purpose string) time.Duration {
	switch purpose {
	case PurposeOrganizerLogin, PurposeParticipantLogin:
		return s.cfg.MagicLinkTTL
	case PurposeRegistrationVerify, PurposeRegistrationCancel:
		return s.cfg.RegistrationTTL
	case PurposeWaitlistOffer:
		return s.cfg.WaitlistOfferTTL
	case PurposeCertificateAccess:
		return s.cfg.CertificateTTL
	default:
		return s.cfg.MagicLinkTTL
	}
}

func (s *Service) hash(raw string) string {
	sum := sha256.Sum256([]byte(s.cfg.TokenPepper + ":" + strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func (s *Service) writeAudit(ctx context.Context, event auditEvent) error {
	detailsJSON := ""
	if len(event.Details) > 0 {
		payload, err := json.Marshal(event.Details)
		if err != nil {
			return fmt.Errorf("marshal audit details: %w", err)
		}
		detailsJSON = string(payload)
	}

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO audit_log (
      id, tenant_id, actor_user_id, actor_participant_id, action, entity_type, entity_id,
      details_json, request_ip, user_agent, created_at
    ) VALUES (?, ?, ?, ?, ?, 'auth', NULL, ?, ?, ?, ?)`,
		s.idFn("aud"),
		nullable(event.TenantID),
		nullable(event.UserID),
		nullable(event.ParticipantID),
		event.Action,
		nullable(detailsJSON),
		nullable(event.RequestIP),
		nullable(event.UserAgent),
		s.nowFn().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func normalizePurpose(raw string) (string, error) {
	purpose := strings.ToLower(strings.TrimSpace(raw))
	if purpose == "" {
		return PurposeOrganizerLogin, nil
	}
	switch purpose {
	case PurposeOrganizerLogin, PurposeRegistrationVerify, PurposeParticipantLogin, PurposeWaitlistOffer, PurposeRegistrationCancel, PurposeCertificateAccess:
		return purpose, nil
	default:
		return "", ErrUnsupportedPurpose
	}
}

func sanitizeRedirectPath(raw string) string {
	redirect := strings.TrimSpace(raw)
	if redirect == "" {
		return "/admin"
	}
	if !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") {
		return "/admin"
	}
	return redirect
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
