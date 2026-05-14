package invitation

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	StatusActive   = "active"
	StatusPaused   = "paused"
	StatusDisabled = "disabled"

	InviteTypePlainInvitation = "plain_invitation"
	InviteTypeDiscountPercent = "discount_percent"
	InviteTypeDiscountFixed   = "discount_fixed"
	InviteTypeVoucherFixed    = "voucher_fixed"
	InviteTypeVoucherFull     = "voucher_full"
	InviteTypeSponsorshipFull = "sponsorship_full"
	InviteTypeDonationEnabled = "donation_enabled"
	InviteTypeEarlyBird       = "early_bird"
	InviteTypeShareable       = "shareable_referral"

	DiscountTypeNone    = ""
	DiscountTypePercent = "percent"
	DiscountTypeFixed   = "fixed"
	DiscountTypeFull    = "full"
)

var (
	ErrInvitationNotFound       = errors.New("invitation not found")
	ErrInvalidInvitationInput   = errors.New("invalid invitation input")
	ErrInvitationStatusInvalid  = errors.New("invitation status is not active")
	ErrInvitationNotStarted     = errors.New("invitation is not active yet")
	ErrInvitationExpired        = errors.New("invitation is expired")
	ErrInvitationScopeMismatch  = errors.New("invitation does not apply to this event")
	ErrInvitationUsageExceeded  = errors.New("invitation usage limit exceeded")
	ErrInvitationEmailExceeded  = errors.New("invitation email usage limit exceeded")
	ErrInvitationCodeRequired   = errors.New("invitation code is required")
	ErrEventNotFound            = errors.New("event not found")
	ErrInvitationDiscountConfig = errors.New("invitation discount configuration is invalid")
)

var codePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{2,63}$`)

type Service struct {
	db    *sql.DB
	nowFn func() time.Time
	idFn  func(prefix string) string
}

type Link struct {
	ID               string
	TenantID         string
	EventID          string
	SeriesID         string
	Code             string
	Label            string
	InviteType       string
	DiscountType     string
	DiscountValue    *int
	MaxUses          *int
	UsedCount        int
	MaxUsesPerEmail  *int
	StartsAt         *time.Time
	ExpiresAt        *time.Time
	IsShareable      bool
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ComputedDiscount string
}

type CreateLinkInput struct {
	EventID         string
	SeriesID        string
	Code            string
	Label           string
	InviteType      string
	DiscountType    string
	DiscountValue   *int
	MaxUses         *int
	MaxUsesPerEmail *int
	StartsAt        *time.Time
	ExpiresAt       *time.Time
	IsShareable     *bool
	Status          string
}

type UpdateLinkInput struct {
	EventID         *string
	SeriesID        *string
	Code            *string
	Label           *string
	InviteType      *string
	DiscountType    *string
	DiscountValue   *int
	MaxUses         *int
	MaxUsesPerEmail *int
	StartsAt        *time.Time
	ExpiresAt       *time.Time
	IsShareable     *bool
	Status          *string
}

type ResolveInput struct {
	TenantID          string
	EventID           string
	ParticipantEmail  string
	Code              string
	BaseAmountCents   int
	ReferenceDateTime time.Time
}

type ResolveResult struct {
	Link                Link
	BaseAmountCents     int
	DiscountAmountCents int
	CreditAmountCents   int
	FinalAmountCents    int
	Sponsored           bool
}

type ApplyCodeToRegistrationInput struct {
	TenantID         string
	EventID          string
	RegistrationID   string
	ParticipantEmail string
	Code             string
	BaseAmountCents  int
}

type queryExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewService(sqlDB *sql.DB) *Service {
	return &Service{
		db:    sqlDB,
		nowFn: func() time.Time { return time.Now().UTC() },
		idFn:  defaultID,
	}
}

func (s *Service) ListLinks(ctx context.Context, tenantID string) ([]Link, error) {
	if s.db == nil {
		return nil, fmt.Errorf("invitation service database is nil")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, COALESCE(event_id, ''), COALESCE(series_id, ''), code,
            COALESCE(label, ''), invite_type, COALESCE(discount_type, ''), discount_value,
            max_uses, used_count, max_uses_per_email, COALESCE(starts_at, ''), COALESCE(expires_at, ''),
            is_shareable, status, created_at, updated_at
     FROM invitation_links
     WHERE tenant_id = ?
     ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("list invitation links: %w", err)
	}
	defer rows.Close()

	items := make([]Link, 0)
	for rows.Next() {
		item, err := scanLinkRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invitation links: %w", err)
	}
	return items, nil
}

func (s *Service) GetLinkByID(ctx context.Context, tenantID, linkID string) (Link, error) {
	if s.db == nil {
		return Link{}, fmt.Errorf("invitation service database is nil")
	}
	tenantID = strings.TrimSpace(tenantID)
	linkID = strings.TrimSpace(linkID)
	if tenantID == "" || linkID == "" {
		return Link{}, fmt.Errorf("tenant id and link id must not be empty")
	}

	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, COALESCE(event_id, ''), COALESCE(series_id, ''), code,
            COALESCE(label, ''), invite_type, COALESCE(discount_type, ''), discount_value,
            max_uses, used_count, max_uses_per_email, COALESCE(starts_at, ''), COALESCE(expires_at, ''),
            is_shareable, status, created_at, updated_at
     FROM invitation_links
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenantID,
		linkID,
	)
	return scanLinkRow(row)
}

func (s *Service) CreateLink(ctx context.Context, tenantID string, input CreateLinkInput) (Link, error) {
	if s.db == nil {
		return Link{}, fmt.Errorf("invitation service database is nil")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return Link{}, fmt.Errorf("tenant id must not be empty")
	}

	normalized, err := normalizeCreateInput(input)
	if err != nil {
		return Link{}, err
	}

	now := s.nowFn().UTC()
	linkID := s.idFn("inv")
	if _, err := s.db.ExecContext(
		ctx,
		`INSERT INTO invitation_links (
      id, tenant_id, event_id, series_id, code, label, invite_type, discount_type,
      discount_value, max_uses, used_count, max_uses_per_email, starts_at, expires_at,
      is_shareable, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?, ?, ?)`,
		linkID,
		tenantID,
		nullableString(normalized.EventID),
		nullableString(normalized.SeriesID),
		normalized.Code,
		nullableString(normalized.Label),
		normalized.InviteType,
		nullableString(normalized.DiscountType),
		nullableIntPtr(normalized.DiscountValue),
		nullableIntPtr(normalized.MaxUses),
		nullableIntPtr(normalized.MaxUsesPerEmail),
		nullableTime(normalized.StartsAt),
		nullableTime(normalized.ExpiresAt),
		boolToInt(*normalized.IsShareable),
		normalized.Status,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return Link{}, fmt.Errorf("%w: code already exists", ErrInvalidInvitationInput)
		}
		return Link{}, fmt.Errorf("insert invitation link: %w", err)
	}

	return s.GetLinkByID(ctx, tenantID, linkID)
}

func (s *Service) UpdateLink(ctx context.Context, tenantID, linkID string, input UpdateLinkInput) (Link, error) {
	if s.db == nil {
		return Link{}, fmt.Errorf("invitation service database is nil")
	}
	tenantID = strings.TrimSpace(tenantID)
	linkID = strings.TrimSpace(linkID)
	if tenantID == "" || linkID == "" {
		return Link{}, fmt.Errorf("tenant id and link id must not be empty")
	}

	current, err := s.GetLinkByID(ctx, tenantID, linkID)
	if err != nil {
		return Link{}, err
	}

	merged, err := mergeUpdateInput(current, input)
	if err != nil {
		return Link{}, err
	}

	now := s.nowFn().UTC()
	if _, err := s.db.ExecContext(
		ctx,
		`UPDATE invitation_links
     SET event_id = ?, series_id = ?, code = ?, label = ?, invite_type = ?, discount_type = ?,
         discount_value = ?, max_uses = ?, max_uses_per_email = ?, starts_at = ?, expires_at = ?,
         is_shareable = ?, status = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		nullableString(merged.EventID),
		nullableString(merged.SeriesID),
		merged.Code,
		nullableString(merged.Label),
		merged.InviteType,
		nullableString(merged.DiscountType),
		nullableIntPtr(merged.DiscountValue),
		nullableIntPtr(merged.MaxUses),
		nullableIntPtr(merged.MaxUsesPerEmail),
		nullableTime(merged.StartsAt),
		nullableTime(merged.ExpiresAt),
		boolToInt(*merged.IsShareable),
		merged.Status,
		now.Format(time.RFC3339),
		tenantID,
		linkID,
	); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return Link{}, fmt.Errorf("%w: code already exists", ErrInvalidInvitationInput)
		}
		return Link{}, fmt.Errorf("update invitation link: %w", err)
	}

	return s.GetLinkByID(ctx, tenantID, linkID)
}

func (s *Service) ResolveCode(ctx context.Context, input ResolveInput) (ResolveResult, error) {
	if s.db == nil {
		return ResolveResult{}, fmt.Errorf("invitation service database is nil")
	}
	resolved, err := s.resolveCode(ctx, s.db, input)
	if err != nil {
		return ResolveResult{}, err
	}
	return resolved.result, nil
}

func (s *Service) ApplyCodeToRegistrationTx(ctx context.Context, tx *sql.Tx, input ApplyCodeToRegistrationInput) (ResolveResult, error) {
	if tx == nil {
		return ResolveResult{}, fmt.Errorf("transaction must not be nil")
	}
	resolved, err := s.resolveCode(ctx, tx, ResolveInput{
		TenantID:         input.TenantID,
		EventID:          input.EventID,
		ParticipantEmail: input.ParticipantEmail,
		Code:             input.Code,
		BaseAmountCents:  input.BaseAmountCents,
	})
	if err != nil {
		return ResolveResult{}, err
	}

	now := s.nowFn().UTC()
	res, err := tx.ExecContext(
		ctx,
		`UPDATE invitation_links
     SET used_count = used_count + 1, updated_at = ?
     WHERE tenant_id = ? AND id = ?
       AND (max_uses IS NULL OR max_uses <= 0 OR used_count < max_uses)`,
		now.Format(time.RFC3339),
		resolved.link.TenantID,
		resolved.link.ID,
	)
	if err != nil {
		return ResolveResult{}, fmt.Errorf("increment invitation usage: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return ResolveResult{}, fmt.Errorf("fetch invitation usage update rows: %w", err)
	}
	if rowsAffected != 1 {
		return ResolveResult{}, ErrInvitationUsageExceeded
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE registrations
     SET invite_id = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		resolved.link.ID,
		now.Format(time.RFC3339),
		strings.TrimSpace(input.TenantID),
		strings.TrimSpace(input.RegistrationID),
	); err != nil {
		return ResolveResult{}, fmt.Errorf("set registration invite id: %w", err)
	}

	email := strings.ToLower(strings.TrimSpace(input.ParticipantEmail))
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO discount_redemptions (
      id, tenant_id, invitation_link_id, registration_id, participant_email,
      discount_amount_cents, redeemed_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		s.idFn("red"),
		strings.TrimSpace(input.TenantID),
		resolved.link.ID,
		strings.TrimSpace(input.RegistrationID),
		nullableString(email),
		resolved.result.DiscountAmountCents,
		now.Format(time.RFC3339),
	); err != nil {
		return ResolveResult{}, fmt.Errorf("insert discount redemption: %w", err)
	}

	return resolved.result, nil
}

type resolvedCode struct {
	link   Link
	result ResolveResult
}

func (s *Service) resolveCode(ctx context.Context, q queryExec, input ResolveInput) (resolvedCode, error) {
	tenantID := strings.TrimSpace(input.TenantID)
	if tenantID == "" {
		return resolvedCode{}, fmt.Errorf("%w: tenant id must not be empty", ErrInvalidInvitationInput)
	}
	code, err := normalizeCode(input.Code)
	if err != nil {
		return resolvedCode{}, err
	}
	if strings.TrimSpace(code) == "" {
		return resolvedCode{}, ErrInvitationCodeRequired
	}
	if input.BaseAmountCents < 0 {
		return resolvedCode{}, fmt.Errorf("%w: base_amount_cents must be >= 0", ErrInvalidInvitationInput)
	}

	row := q.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, COALESCE(event_id, ''), COALESCE(series_id, ''), code,
            COALESCE(label, ''), invite_type, COALESCE(discount_type, ''), discount_value,
            max_uses, used_count, max_uses_per_email, COALESCE(starts_at, ''), COALESCE(expires_at, ''),
            is_shareable, status, created_at, updated_at
     FROM invitation_links
     WHERE tenant_id = ? AND upper(code) = ?
     LIMIT 1`,
		tenantID,
		code,
	)
	link, err := scanLinkRow(row)
	if err != nil {
		if errors.Is(err, ErrInvitationNotFound) {
			return resolvedCode{}, err
		}
		return resolvedCode{}, fmt.Errorf("lookup invitation by code: %w", err)
	}

	now := input.ReferenceDateTime
	if now.IsZero() {
		now = s.nowFn().UTC()
	}
	if strings.ToLower(strings.TrimSpace(link.Status)) != StatusActive {
		return resolvedCode{}, ErrInvitationStatusInvalid
	}
	if link.StartsAt != nil && now.Before(link.StartsAt.UTC()) {
		return resolvedCode{}, ErrInvitationNotStarted
	}
	if link.ExpiresAt != nil && now.After(link.ExpiresAt.UTC()) {
		return resolvedCode{}, ErrInvitationExpired
	}

	eventID := strings.TrimSpace(input.EventID)
	if link.EventID != "" || link.SeriesID != "" {
		if eventID == "" {
			return resolvedCode{}, ErrInvitationScopeMismatch
		}
		eventSeriesID, err := lookupEventSeriesID(ctx, q, tenantID, eventID)
		if err != nil {
			return resolvedCode{}, err
		}
		if link.EventID != "" && !strings.EqualFold(link.EventID, eventID) {
			return resolvedCode{}, ErrInvitationScopeMismatch
		}
		if link.SeriesID != "" && !strings.EqualFold(link.SeriesID, eventSeriesID) {
			return resolvedCode{}, ErrInvitationScopeMismatch
		}
	}

	if link.MaxUses != nil && link.UsedCount >= *link.MaxUses {
		return resolvedCode{}, ErrInvitationUsageExceeded
	}

	email := strings.ToLower(strings.TrimSpace(input.ParticipantEmail))
	if link.MaxUsesPerEmail != nil && *link.MaxUsesPerEmail > 0 && email != "" {
		count, err := countEmailRedemptions(ctx, q, tenantID, link.ID, email)
		if err != nil {
			return resolvedCode{}, err
		}
		if count >= *link.MaxUsesPerEmail {
			return resolvedCode{}, ErrInvitationEmailExceeded
		}
	}

	discountAmount, creditAmount, finalAmount, sponsored, err := computePricing(link, input.BaseAmountCents)
	if err != nil {
		return resolvedCode{}, err
	}

	result := ResolveResult{
		Link:                link,
		BaseAmountCents:     input.BaseAmountCents,
		DiscountAmountCents: discountAmount,
		CreditAmountCents:   creditAmount,
		FinalAmountCents:    finalAmount,
		Sponsored:           sponsored,
	}
	return resolvedCode{link: link, result: result}, nil
}

func lookupEventSeriesID(ctx context.Context, q queryExec, tenantID, eventID string) (string, error) {
	row := q.QueryRowContext(
		ctx,
		`SELECT COALESCE(series_id, '')
     FROM events
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenantID,
		eventID,
	)
	var seriesID string
	if err := row.Scan(&seriesID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrEventNotFound
		}
		return "", fmt.Errorf("query event scope: %w", err)
	}
	return strings.TrimSpace(seriesID), nil
}

func countEmailRedemptions(ctx context.Context, q queryExec, tenantID, invitationID, email string) (int, error) {
	row := q.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
     FROM discount_redemptions
     WHERE tenant_id = ?
       AND invitation_link_id = ?
       AND lower(COALESCE(participant_email, '')) = ?`,
		tenantID,
		invitationID,
		strings.ToLower(strings.TrimSpace(email)),
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count email redemptions: %w", err)
	}
	return count, nil
}

func computePricing(link Link, baseAmountCents int) (discountAmount int, creditAmount int, finalAmount int, sponsored bool, err error) {
	effectiveDiscountType := strings.TrimSpace(link.DiscountType)
	if effectiveDiscountType == "" {
		effectiveDiscountType = defaultDiscountTypeForInviteType(link.InviteType)
	}

	value := 0
	if link.DiscountValue != nil {
		value = *link.DiscountValue
	}

	candidate := 0
	switch effectiveDiscountType {
	case DiscountTypeNone:
		candidate = 0
	case DiscountTypePercent:
		if value < 0 {
			return 0, 0, 0, false, fmt.Errorf("%w: percent discount must be >= 0", ErrInvitationDiscountConfig)
		}
		if value > 100 {
			value = 100
		}
		candidate = (baseAmountCents * value) / 100
	case DiscountTypeFixed:
		if value < 0 {
			return 0, 0, 0, false, fmt.Errorf("%w: fixed discount must be >= 0", ErrInvitationDiscountConfig)
		}
		candidate = value
	case DiscountTypeFull:
		candidate = baseAmountCents
	default:
		return 0, 0, 0, false, fmt.Errorf("%w: unsupported discount type %q", ErrInvitationDiscountConfig, effectiveDiscountType)
	}

	if candidate < 0 {
		candidate = 0
	}
	discountAmount = candidate
	if discountAmount > baseAmountCents {
		discountAmount = baseAmountCents
		creditAmount = candidate - baseAmountCents
	}
	if creditAmount < 0 {
		creditAmount = 0
	}
	finalAmount = baseAmountCents - discountAmount
	if finalAmount < 0 {
		finalAmount = 0
	}

	inviteType := strings.ToLower(strings.TrimSpace(link.InviteType))
	sponsored = (inviteType == InviteTypeSponsorshipFull || inviteType == InviteTypeVoucherFull || effectiveDiscountType == DiscountTypeFull) && finalAmount == 0
	return discountAmount, creditAmount, finalAmount, sponsored, nil
}

func normalizeCreateInput(input CreateLinkInput) (CreateLinkInput, error) {
	inviteType, err := normalizeInviteType(input.InviteType)
	if err != nil {
		return CreateLinkInput{}, err
	}
	discountType, err := normalizeDiscountType(input.DiscountType, inviteType)
	if err != nil {
		return CreateLinkInput{}, err
	}

	code, err := normalizeCode(input.Code)
	if err != nil {
		return CreateLinkInput{}, err
	}
	status, err := normalizeStatus(input.Status)
	if err != nil {
		return CreateLinkInput{}, err
	}
	isShareable := false
	if input.IsShareable != nil {
		isShareable = *input.IsShareable
	}

	normalized := CreateLinkInput{
		EventID:         strings.TrimSpace(input.EventID),
		SeriesID:        strings.TrimSpace(input.SeriesID),
		Code:            code,
		Label:           strings.TrimSpace(input.Label),
		InviteType:      inviteType,
		DiscountType:    discountType,
		DiscountValue:   copyIntPointer(input.DiscountValue),
		MaxUses:         sanitizeLimitPointer(input.MaxUses),
		MaxUsesPerEmail: sanitizeLimitPointer(input.MaxUsesPerEmail),
		StartsAt:        copyTimePointer(input.StartsAt),
		ExpiresAt:       copyTimePointer(input.ExpiresAt),
		IsShareable:     &isShareable,
		Status:          status,
	}
	if err := validateTimeline(normalized.StartsAt, normalized.ExpiresAt); err != nil {
		return CreateLinkInput{}, err
	}
	if err := validateDiscountValue(discountType, normalized.DiscountValue); err != nil {
		return CreateLinkInput{}, err
	}

	return normalized, nil
}

func mergeUpdateInput(current Link, input UpdateLinkInput) (CreateLinkInput, error) {
	merged := CreateLinkInput{
		EventID:         current.EventID,
		SeriesID:        current.SeriesID,
		Code:            current.Code,
		Label:           current.Label,
		InviteType:      current.InviteType,
		DiscountType:    current.DiscountType,
		DiscountValue:   copyIntPointer(current.DiscountValue),
		MaxUses:         copyIntPointer(current.MaxUses),
		MaxUsesPerEmail: copyIntPointer(current.MaxUsesPerEmail),
		StartsAt:        copyTimePointer(current.StartsAt),
		ExpiresAt:       copyTimePointer(current.ExpiresAt),
		IsShareable:     boolPointer(current.IsShareable),
		Status:          current.Status,
	}

	if input.EventID != nil {
		merged.EventID = strings.TrimSpace(*input.EventID)
	}
	if input.SeriesID != nil {
		merged.SeriesID = strings.TrimSpace(*input.SeriesID)
	}
	if input.Code != nil {
		merged.Code = strings.TrimSpace(*input.Code)
	}
	if input.Label != nil {
		merged.Label = strings.TrimSpace(*input.Label)
	}
	if input.InviteType != nil {
		merged.InviteType = strings.TrimSpace(*input.InviteType)
	}
	if input.DiscountType != nil {
		merged.DiscountType = strings.TrimSpace(*input.DiscountType)
	}
	if input.DiscountValue != nil {
		value := *input.DiscountValue
		merged.DiscountValue = &value
	}
	if input.MaxUses != nil {
		value := *input.MaxUses
		merged.MaxUses = &value
	}
	if input.MaxUsesPerEmail != nil {
		value := *input.MaxUsesPerEmail
		merged.MaxUsesPerEmail = &value
	}
	if input.StartsAt != nil {
		merged.StartsAt = copyTimePointer(input.StartsAt)
	}
	if input.ExpiresAt != nil {
		merged.ExpiresAt = copyTimePointer(input.ExpiresAt)
	}
	if input.IsShareable != nil {
		value := *input.IsShareable
		merged.IsShareable = &value
	}
	if input.Status != nil {
		merged.Status = strings.TrimSpace(*input.Status)
	}

	return normalizeCreateInput(merged)
}

func normalizeCode(raw string) (string, error) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if code == "" {
		return "", fmt.Errorf("%w: code must not be empty", ErrInvalidInvitationInput)
	}
	if !codePattern.MatchString(code) {
		return "", fmt.Errorf("%w: invitation code %q is invalid", ErrInvalidInvitationInput, raw)
	}
	return code, nil
}

func normalizeInviteType(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		value = InviteTypePlainInvitation
	}
	switch value {
	case InviteTypePlainInvitation,
		InviteTypeDiscountPercent,
		InviteTypeDiscountFixed,
		InviteTypeVoucherFixed,
		InviteTypeVoucherFull,
		InviteTypeSponsorshipFull,
		InviteTypeDonationEnabled,
		InviteTypeEarlyBird,
		InviteTypeShareable:
		return value, nil
	default:
		return "", fmt.Errorf("%w: unsupported invite_type %q", ErrInvalidInvitationInput, raw)
	}
}

func normalizeDiscountType(raw, inviteType string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return defaultDiscountTypeForInviteType(inviteType), nil
	}
	switch value {
	case DiscountTypePercent, DiscountTypeFixed, DiscountTypeFull:
		return value, nil
	default:
		return "", fmt.Errorf("%w: unsupported discount_type %q", ErrInvalidInvitationInput, raw)
	}
}

func defaultDiscountTypeForInviteType(inviteType string) string {
	switch strings.ToLower(strings.TrimSpace(inviteType)) {
	case InviteTypeDiscountPercent, InviteTypeEarlyBird, InviteTypeShareable:
		return DiscountTypePercent
	case InviteTypeDiscountFixed, InviteTypeVoucherFixed:
		return DiscountTypeFixed
	case InviteTypeVoucherFull, InviteTypeSponsorshipFull:
		return DiscountTypeFull
	default:
		return DiscountTypeNone
	}
}

func validateDiscountValue(discountType string, discountValue *int) error {
	value := 0
	if discountValue != nil {
		value = *discountValue
	}
	switch strings.ToLower(strings.TrimSpace(discountType)) {
	case DiscountTypeNone:
		return nil
	case DiscountTypePercent:
		if value < 0 || value > 100 {
			return fmt.Errorf("%w: percent discount_value must be between 0 and 100", ErrInvalidInvitationInput)
		}
		return nil
	case DiscountTypeFixed:
		if value < 0 {
			return fmt.Errorf("%w: fixed discount_value must be >= 0", ErrInvalidInvitationInput)
		}
		return nil
	case DiscountTypeFull:
		return nil
	default:
		return fmt.Errorf("%w: unsupported discount_type %q", ErrInvalidInvitationInput, discountType)
	}
}

func normalizeStatus(raw string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		value = StatusActive
	}
	switch value {
	case StatusActive, StatusPaused, StatusDisabled:
		return value, nil
	default:
		return "", fmt.Errorf("%w: unsupported status %q", ErrInvalidInvitationInput, raw)
	}
}

func validateTimeline(startsAt, expiresAt *time.Time) error {
	if startsAt == nil || expiresAt == nil {
		return nil
	}
	if expiresAt.Before(startsAt.UTC()) {
		return fmt.Errorf("%w: expires_at must be after starts_at", ErrInvalidInvitationInput)
	}
	return nil
}

func scanLinkRow(row interface{ Scan(dest ...any) error }) (Link, error) {
	var (
		item            Link
		eventID         string
		seriesID        string
		label           string
		discountType    string
		discountValue   sql.NullInt64
		maxUses         sql.NullInt64
		maxUsesPerEmail sql.NullInt64
		startsAtRaw     string
		expiresAtRaw    string
		isShareableInt  int
		createdAtRaw    string
		updatedAtRaw    string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&eventID,
		&seriesID,
		&item.Code,
		&label,
		&item.InviteType,
		&discountType,
		&discountValue,
		&maxUses,
		&item.UsedCount,
		&maxUsesPerEmail,
		&startsAtRaw,
		&expiresAtRaw,
		&isShareableInt,
		&item.Status,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Link{}, ErrInvitationNotFound
		}
		return Link{}, fmt.Errorf("scan invitation link: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return Link{}, fmt.Errorf("parse invitation created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return Link{}, fmt.Errorf("parse invitation updated_at: %w", err)
	}

	item.EventID = strings.TrimSpace(eventID)
	item.SeriesID = strings.TrimSpace(seriesID)
	item.Label = strings.TrimSpace(label)
	item.DiscountType = strings.TrimSpace(discountType)
	item.IsShareable = isShareableInt == 1
	item.CreatedAt = createdAt.UTC()
	item.UpdatedAt = updatedAt.UTC()
	item.ComputedDiscount = defaultDiscountTypeForInviteType(item.InviteType)

	if discountValue.Valid {
		value := int(discountValue.Int64)
		item.DiscountValue = &value
	}
	if maxUses.Valid && maxUses.Int64 > 0 {
		value := int(maxUses.Int64)
		item.MaxUses = &value
	}
	if maxUsesPerEmail.Valid && maxUsesPerEmail.Int64 > 0 {
		value := int(maxUsesPerEmail.Int64)
		item.MaxUsesPerEmail = &value
	}
	if strings.TrimSpace(startsAtRaw) != "" {
		parsed, err := time.Parse(time.RFC3339, startsAtRaw)
		if err != nil {
			return Link{}, fmt.Errorf("parse invitation starts_at: %w", err)
		}
		parsed = parsed.UTC()
		item.StartsAt = &parsed
	}
	if strings.TrimSpace(expiresAtRaw) != "" {
		parsed, err := time.Parse(time.RFC3339, expiresAtRaw)
		if err != nil {
			return Link{}, fmt.Errorf("parse invitation expires_at: %w", err)
		}
		parsed = parsed.UTC()
		item.ExpiresAt = &parsed
	}
	return item, nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableIntPtr(value *int) any {
	if value == nil {
		return nil
	}
	if *value <= 0 {
		return nil
	}
	return *value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return utc.Format(time.RFC3339)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func copyIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func sanitizeLimitPointer(value *int) *int {
	if value == nil {
		return nil
	}
	if *value <= 0 {
		return nil
	}
	copy := *value
	return &copy
}

func copyTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := value.UTC()
	return &copy
}

func boolPointer(value bool) *bool {
	copy := value
	return &copy
}

func defaultID(prefix string) string {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, random)
}
