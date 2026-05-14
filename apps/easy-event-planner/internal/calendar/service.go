package calendar

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	FeedTypeOrganizer = "organizer"
	FeedStatusActive  = "active"

	participantTokenPurpose = "participant_calendar"
)

var (
	ErrFeedNotFound             = errors.New("calendar feed not found")
	ErrInvalidFeedToken         = errors.New("invalid calendar feed token")
	ErrRegistrationNotFound     = errors.New("registration not found")
	ErrInvalidRegistrationToken = errors.New("invalid registration calendar token")
)

type Config struct {
	BaseURL     string
	TokenPepper string
}

type Service struct {
	db    *sql.DB
	cfg   Config
	nowFn func() time.Time
	idFn  func(prefix string) string
}

type OrganizerFeed struct {
	ID           string
	TenantID     string
	FeedType     string
	Status       string
	Token        string
	URL          string
	LastAccessed *time.Time
	CreatedAt    time.Time
	RotatedAt    *time.Time
}

type calendarEvent struct {
	ID                string
	TenantID          string
	Slug              string
	Title             string
	Description       string
	StartsAt          time.Time
	EndsAt            *time.Time
	Timezone          string
	LocationName      string
	Address           string
	OnlineURL         string
	ParticipationMode string
	Status            string
	ChangeNote        string
	CancelledReason   string
	UpdatedAt         time.Time
}

type participantCalendarItem struct {
	RegistrationID     string
	RegistrationStatus string
	ParticipantID      string
	ParticipantName    string
	Event              calendarEvent
}

type feedRow struct {
	ID              string
	TenantID        string
	FeedType        string
	TokenHash       string
	Status          string
	LastAccessedRaw string
	CreatedAtRaw    string
	RotatedAtRaw    string
}

func NewService(sqlDB *sql.DB, cfg Config) *Service {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	pepper := strings.TrimSpace(cfg.TokenPepper)
	if pepper == "" {
		pepper = "dev-only-change-me"
	}
	return &Service{
		db: sqlDB,
		cfg: Config{
			BaseURL:     strings.TrimRight(baseURL, "/"),
			TokenPepper: pepper,
		},
		nowFn: func() time.Time { return time.Now().UTC() },
		idFn:  defaultID,
	}
}

func (s *Service) GetOrCreateOrganizerFeed(ctx context.Context, tenantID string) (OrganizerFeed, error) {
	if s.db == nil {
		return OrganizerFeed{}, fmt.Errorf("calendar service database is nil")
	}
	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return OrganizerFeed{}, fmt.Errorf("tenant id must not be empty")
	}

	row, err := s.fetchFeedRow(ctx, tenant, FeedTypeOrganizer)
	if err != nil {
		if !errors.Is(err, ErrFeedNotFound) {
			return OrganizerFeed{}, err
		}
		now := s.nowFn().UTC()
		seed := now.Format(time.RFC3339Nano)
		token := s.deriveOrganizerToken(tenant, seed)
		tokenHash := s.hashToken(token)
		feedID := s.idFn("cal")
		if _, err := s.db.ExecContext(
			ctx,
			`INSERT INTO calendar_feeds (
        id, tenant_id, user_id, feed_type, token_hash, status, last_accessed_at, created_at, rotated_at
      ) VALUES (?, ?, NULL, ?, ?, ?, NULL, ?, ?)`,
			feedID,
			tenant,
			FeedTypeOrganizer,
			tokenHash,
			FeedStatusActive,
			seed,
			seed,
		); err != nil {
			return OrganizerFeed{}, fmt.Errorf("insert organizer calendar feed: %w", err)
		}
		row, err = s.fetchFeedRow(ctx, tenant, FeedTypeOrganizer)
		if err != nil {
			return OrganizerFeed{}, err
		}
		return s.mapFeedRow(row, token), nil
	}

	token := s.deriveOrganizerToken(tenant, organizerSeed(row))
	if strings.TrimSpace(row.TokenHash) == "" {
		if _, err := s.db.ExecContext(
			ctx,
			`UPDATE calendar_feeds SET token_hash = ?, status = ? WHERE id = ?`,
			s.hashToken(token),
			FeedStatusActive,
			row.ID,
		); err != nil {
			// fallback gracefully; feed still usable via derived token
		}
	}
	return s.mapFeedRow(row, token), nil
}

func (s *Service) RotateOrganizerFeed(ctx context.Context, tenantID string) (OrganizerFeed, error) {
	if s.db == nil {
		return OrganizerFeed{}, fmt.Errorf("calendar service database is nil")
	}
	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return OrganizerFeed{}, fmt.Errorf("tenant id must not be empty")
	}

	feed, err := s.GetOrCreateOrganizerFeed(ctx, tenant)
	if err != nil {
		return OrganizerFeed{}, err
	}
	now := s.nowFn().UTC()
	seed := now.Format(time.RFC3339Nano)
	token := s.deriveOrganizerToken(tenant, seed)
	if _, err := s.db.ExecContext(
		ctx,
		`UPDATE calendar_feeds
     SET token_hash = ?, status = ?, rotated_at = ?
     WHERE id = ?`,
		s.hashToken(token),
		FeedStatusActive,
		seed,
		feed.ID,
	); err != nil {
		return OrganizerFeed{}, fmt.Errorf("rotate organizer calendar feed: %w", err)
	}

	row, err := s.fetchFeedRow(ctx, tenant, FeedTypeOrganizer)
	if err != nil {
		return OrganizerFeed{}, err
	}
	return s.mapFeedRow(row, token), nil
}

func (s *Service) OrganizerFeedEmbedURL(feed OrganizerFeed, tenantSlug string) string {
	trimmedSlug := strings.TrimSpace(tenantSlug)
	if trimmedSlug == "" || strings.TrimSpace(feed.Token) == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/calendar/admin.ics?token=%s", s.cfg.BaseURL, url.PathEscape(trimmedSlug), url.QueryEscape(feed.Token))
}

func (s *Service) ParticipantCalendarURL(tenantSlug, tenantID, registrationID, participantID string) string {
	slug := strings.TrimSpace(tenantSlug)
	tenant := strings.TrimSpace(tenantID)
	regID := strings.TrimSpace(registrationID)
	participant := strings.TrimSpace(participantID)
	if slug == "" || tenant == "" || regID == "" || participant == "" {
		return ""
	}
	token := s.participantToken(tenant, regID, participant)
	return fmt.Sprintf("%s/api/v1/public/%s/registrations/%s/calendar.ics?token=%s", s.cfg.BaseURL, url.PathEscape(slug), url.PathEscape(regID), url.QueryEscape(token))
}

func (s *Service) RenderOrganizerICS(ctx context.Context, tenantID, tenantSlug, token string) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("calendar service database is nil")
	}
	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return "", fmt.Errorf("tenant id must not be empty")
	}
	feedRow, err := s.fetchFeedRow(ctx, tenant, FeedTypeOrganizer)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(feedRow.Status) != "" && strings.ToLower(strings.TrimSpace(feedRow.Status)) != FeedStatusActive {
		return "", ErrInvalidFeedToken
	}
	providedHash := s.hashToken(strings.TrimSpace(token))
	storedHash := strings.TrimSpace(feedRow.TokenHash)
	if storedHash == "" || subtle.ConstantTimeCompare([]byte(storedHash), []byte(providedHash)) != 1 {
		return "", ErrInvalidFeedToken
	}

	events, err := s.listOrganizerEvents(ctx, tenant)
	if err != nil {
		return "", err
	}

	if _, err := s.db.ExecContext(
		ctx,
		`UPDATE calendar_feeds SET last_accessed_at = ? WHERE id = ?`,
		s.nowFn().UTC().Format(time.RFC3339Nano),
		feedRow.ID,
	); err != nil {
		return "", fmt.Errorf("update organizer calendar last_accessed_at: %w", err)
	}

	calendarName := fmt.Sprintf("%s Organizer Kalender", strings.TrimSpace(tenantSlug))
	if strings.TrimSpace(tenantSlug) == "" {
		calendarName = "Organizer Kalender"
	}
	return s.buildOrganizerICS(calendarName, tenantSlug, events), nil
}

func (s *Service) RenderParticipantICS(ctx context.Context, tenantID, tenantSlug, registrationID, token string) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("calendar service database is nil")
	}
	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return "", fmt.Errorf("tenant id must not be empty")
	}
	regID := strings.TrimSpace(registrationID)
	if regID == "" {
		return "", fmt.Errorf("registration id must not be empty")
	}

	item, err := s.lookupParticipantCalendarItem(ctx, tenant, regID)
	if err != nil {
		return "", err
	}
	expected := s.participantToken(tenant, item.RegistrationID, item.ParticipantID)
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(token)), []byte(expected)) != 1 {
		return "", ErrInvalidRegistrationToken
	}

	calendarName := fmt.Sprintf("%s Teilnahme", item.Event.Title)
	if strings.TrimSpace(item.Event.Title) == "" {
		calendarName = "Teilnahme"
	}
	return s.buildParticipantICS(calendarName, tenantSlug, item), nil
}

func (s *Service) fetchFeedRow(ctx context.Context, tenantID, feedType string) (feedRow, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, feed_type, COALESCE(token_hash, ''), COALESCE(status, ''),
            COALESCE(last_accessed_at, ''), created_at, COALESCE(rotated_at, '')
     FROM calendar_feeds
     WHERE tenant_id = ? AND feed_type = ?
     ORDER BY created_at DESC
     LIMIT 1`,
		tenantID,
		feedType,
	)
	var item feedRow
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.FeedType,
		&item.TokenHash,
		&item.Status,
		&item.LastAccessedRaw,
		&item.CreatedAtRaw,
		&item.RotatedAtRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return feedRow{}, ErrFeedNotFound
		}
		return feedRow{}, fmt.Errorf("query calendar feed: %w", err)
	}
	return item, nil
}

func (s *Service) mapFeedRow(row feedRow, token string) OrganizerFeed {
	createdAt := mustParseTime(row.CreatedAtRaw)
	rotatedAt := parseOptionalTime(row.RotatedAtRaw)
	lastAccessed := parseOptionalTime(row.LastAccessedRaw)
	if strings.TrimSpace(token) == "" {
		token = s.deriveOrganizerToken(row.TenantID, organizerSeed(row))
	}
	feed := OrganizerFeed{
		ID:           row.ID,
		TenantID:     row.TenantID,
		FeedType:     row.FeedType,
		Status:       firstNonEmpty(strings.TrimSpace(row.Status), FeedStatusActive),
		Token:        token,
		LastAccessed: lastAccessed,
		CreatedAt:    createdAt,
		RotatedAt:    rotatedAt,
	}
	return feed
}

func organizerSeed(row feedRow) string {
	if strings.TrimSpace(row.RotatedAtRaw) != "" {
		return strings.TrimSpace(row.RotatedAtRaw)
	}
	return strings.TrimSpace(row.CreatedAtRaw)
}

func (s *Service) deriveOrganizerToken(tenantID, seed string) string {
	h := hmac.New(sha256.New, []byte(s.cfg.TokenPepper))
	_, _ = h.Write([]byte(strings.TrimSpace(tenantID)))
	_, _ = h.Write([]byte("|" + FeedTypeOrganizer + "|" + strings.TrimSpace(seed)))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func (s *Service) participantToken(tenantID, registrationID, participantID string) string {
	h := hmac.New(sha256.New, []byte(s.cfg.TokenPepper))
	_, _ = h.Write([]byte(strings.TrimSpace(tenantID)))
	_, _ = h.Write([]byte("|" + strings.TrimSpace(registrationID)))
	_, _ = h.Write([]byte("|" + strings.TrimSpace(participantID)))
	_, _ = h.Write([]byte("|" + participantTokenPurpose))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func (s *Service) hashToken(raw string) string {
	sum := sha256.Sum256([]byte(s.cfg.TokenPepper + ":" + strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func (s *Service) listOrganizerEvents(ctx context.Context, tenantID string) ([]calendarEvent, error) {
	windowStart := s.nowFn().UTC().AddDate(0, 0, -30).Format(time.RFC3339)
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, slug, title, COALESCE(description, ''), starts_at, COALESCE(ends_at, ''),
            COALESCE(timezone, 'UTC'), COALESCE(location_name, ''), COALESCE(address, ''), COALESCE(online_url, ''),
            COALESCE(participation_mode, ''), status, COALESCE(change_note, ''), COALESCE(cancelled_reason, ''), updated_at
     FROM events
     WHERE tenant_id = ?
       AND status IN ('scheduled', 'changed', 'postponed', 'cancelled')
       AND starts_at >= ?
     ORDER BY starts_at ASC, created_at DESC`,
		tenantID,
		windowStart,
	)
	if err != nil {
		return nil, fmt.Errorf("list organizer calendar events: %w", err)
	}
	defer rows.Close()

	items := make([]calendarEvent, 0)
	for rows.Next() {
		item, err := scanCalendarEvent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organizer calendar events: %w", err)
	}
	return items, nil
}

func (s *Service) lookupParticipantCalendarItem(ctx context.Context, tenantID, registrationID string) (participantCalendarItem, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT r.id, r.status, r.participant_id, COALESCE(p.name, ''),
            e.id, e.tenant_id, e.slug, e.title, COALESCE(e.description, ''), e.starts_at, COALESCE(e.ends_at, ''),
            COALESCE(e.timezone, 'UTC'), COALESCE(e.location_name, ''), COALESCE(e.address, ''), COALESCE(e.online_url, ''),
            COALESCE(e.participation_mode, ''), e.status, COALESCE(e.change_note, ''), COALESCE(e.cancelled_reason, ''), e.updated_at
     FROM registrations r
     JOIN events e ON e.id = r.event_id AND e.tenant_id = r.tenant_id
     LEFT JOIN participants p ON p.id = r.participant_id
     WHERE r.tenant_id = ? AND r.id = ?
       AND r.status IN ('confirmed', 'waitlist', 'reserved', 'payment_pending', 'attended', 'cancelled')
     LIMIT 1`,
		tenantID,
		registrationID,
	)

	var item participantCalendarItem
	var startsAtRaw string
	var endsAtRaw string
	var updatedRaw string
	if err := row.Scan(
		&item.RegistrationID,
		&item.RegistrationStatus,
		&item.ParticipantID,
		&item.ParticipantName,
		&item.Event.ID,
		&item.Event.TenantID,
		&item.Event.Slug,
		&item.Event.Title,
		&item.Event.Description,
		&startsAtRaw,
		&endsAtRaw,
		&item.Event.Timezone,
		&item.Event.LocationName,
		&item.Event.Address,
		&item.Event.OnlineURL,
		&item.Event.ParticipationMode,
		&item.Event.Status,
		&item.Event.ChangeNote,
		&item.Event.CancelledReason,
		&updatedRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return participantCalendarItem{}, ErrRegistrationNotFound
		}
		return participantCalendarItem{}, fmt.Errorf("query participant calendar registration: %w", err)
	}

	startsAt, err := parseRFC3339(startsAtRaw)
	if err != nil {
		return participantCalendarItem{}, fmt.Errorf("parse participant calendar starts_at: %w", err)
	}
	item.Event.StartsAt = startsAt.UTC()
	if strings.TrimSpace(endsAtRaw) != "" {
		endsAt, err := parseRFC3339(endsAtRaw)
		if err != nil {
			return participantCalendarItem{}, fmt.Errorf("parse participant calendar ends_at: %w", err)
		}
		endsAt = endsAt.UTC()
		item.Event.EndsAt = &endsAt
	}
	updatedAt, err := parseRFC3339(updatedRaw)
	if err != nil {
		return participantCalendarItem{}, fmt.Errorf("parse participant calendar updated_at: %w", err)
	}
	item.Event.UpdatedAt = updatedAt.UTC()
	return item, nil
}

func scanCalendarEvent(row interface{ Scan(dest ...any) error }) (calendarEvent, error) {
	var (
		item        calendarEvent
		startsAtRaw string
		endsAtRaw   string
		updatedRaw  string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.Slug,
		&item.Title,
		&item.Description,
		&startsAtRaw,
		&endsAtRaw,
		&item.Timezone,
		&item.LocationName,
		&item.Address,
		&item.OnlineURL,
		&item.ParticipationMode,
		&item.Status,
		&item.ChangeNote,
		&item.CancelledReason,
		&updatedRaw,
	); err != nil {
		return calendarEvent{}, err
	}
	startsAt, err := parseRFC3339(startsAtRaw)
	if err != nil {
		return calendarEvent{}, fmt.Errorf("parse calendar event starts_at: %w", err)
	}
	item.StartsAt = startsAt.UTC()
	if strings.TrimSpace(endsAtRaw) != "" {
		endsAt, err := parseRFC3339(endsAtRaw)
		if err != nil {
			return calendarEvent{}, fmt.Errorf("parse calendar event ends_at: %w", err)
		}
		endsAt = endsAt.UTC()
		item.EndsAt = &endsAt
	}
	updatedAt, err := parseRFC3339(updatedRaw)
	if err != nil {
		return calendarEvent{}, fmt.Errorf("parse calendar event updated_at: %w", err)
	}
	item.UpdatedAt = updatedAt.UTC()
	return item, nil
}

func (s *Service) buildOrganizerICS(calendarName, tenantSlug string, events []calendarEvent) string {
	now := s.nowFn().UTC()
	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//easy-event-planner//Organizer Feed//DE",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"X-WR-CALNAME:" + escapeICSText(calendarName),
	}
	for _, item := range events {
		lines = append(lines, s.buildEventLines(item, now, tenantSlug, true)...)
	}
	lines = append(lines, "END:VCALENDAR")
	return strings.Join(lines, "\r\n") + "\r\n"
}

func (s *Service) buildParticipantICS(calendarName, tenantSlug string, item participantCalendarItem) string {
	now := s.nowFn().UTC()
	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//easy-event-planner//Participant Calendar//DE",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"X-WR-CALNAME:" + escapeICSText(calendarName),
	}
	lines = append(lines, s.buildEventLines(item.Event, now, tenantSlug, false)...)
	lines = append(lines, "END:VCALENDAR")
	return strings.Join(lines, "\r\n") + "\r\n"
}

func (s *Service) buildEventLines(item calendarEvent, now time.Time, tenantSlug string, organizerView bool) []string {
	lines := []string{"BEGIN:VEVENT"}
	uid := fmt.Sprintf("%s@easy-event-planner", strings.TrimSpace(item.ID))
	lines = append(lines, "UID:"+escapeICSText(uid))
	lines = append(lines, "DTSTAMP:"+formatICSDateTime(now))
	lines = append(lines, "DTSTART:"+formatICSDateTime(item.StartsAt))
	if item.EndsAt != nil {
		lines = append(lines, "DTEND:"+formatICSDateTime(item.EndsAt.UTC()))
	}
	lines = append(lines, "SUMMARY:"+escapeICSText(item.Title))

	status := strings.ToLower(strings.TrimSpace(item.Status))
	if status == "cancelled" {
		lines = append(lines, "STATUS:CANCELLED")
	} else {
		lines = append(lines, "STATUS:CONFIRMED")
	}

	if strings.TrimSpace(item.LocationName) != "" {
		lines = append(lines, "LOCATION:"+escapeICSText(item.LocationName))
	}

	descriptionParts := make([]string, 0)
	if strings.TrimSpace(item.Description) != "" {
		descriptionParts = append(descriptionParts, item.Description)
	}
	if strings.TrimSpace(item.Address) != "" {
		descriptionParts = append(descriptionParts, "Adresse: "+item.Address)
	}
	if strings.TrimSpace(item.OnlineURL) != "" {
		descriptionParts = append(descriptionParts, "Online: "+item.OnlineURL)
	}
	if strings.TrimSpace(item.ChangeNote) != "" {
		descriptionParts = append(descriptionParts, "Hinweis: "+item.ChangeNote)
	}
	if strings.TrimSpace(item.CancelledReason) != "" {
		descriptionParts = append(descriptionParts, "Absagegrund: "+item.CancelledReason)
	}

	if organizerView {
		adminURL := fmt.Sprintf("%s/admin/events/%s", s.cfg.BaseURL, url.PathEscape(strings.TrimSpace(item.ID)))
		descriptionParts = append(descriptionParts, "Admin: "+adminURL)
		lines = append(lines, "URL:"+escapeICSText(adminURL))
	} else {
		publicURL := fmt.Sprintf("%s/api/v1/public/%s/events/%s", s.cfg.BaseURL, url.PathEscape(strings.TrimSpace(tenantSlug)), url.PathEscape(strings.TrimSpace(item.Slug)))
		lines = append(lines, "URL:"+escapeICSText(publicURL))
	}

	if len(descriptionParts) > 0 {
		lines = append(lines, "DESCRIPTION:"+escapeICSText(strings.Join(descriptionParts, "\n")))
	}
	lines = append(lines, "LAST-MODIFIED:"+formatICSDateTime(item.UpdatedAt.UTC()))
	lines = append(lines, "END:VEVENT")
	return lines
}

func formatICSDateTime(value time.Time) string {
	return value.UTC().Format("20060102T150405Z")
}

func escapeICSText(raw string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		"\n", `\\n`,
		"\r", "",
		";", `\\;`,
		",", `\\,`,
	)
	return replacer.Replace(strings.TrimSpace(raw))
}

func parseOptionalTime(raw string) *time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	parsed, err := parseRFC3339(value)
	if err != nil {
		return nil
	}
	parsed = parsed.UTC()
	return &parsed
}

func mustParseTime(raw string) time.Time {
	parsed, err := parseRFC3339(strings.TrimSpace(raw))
	if err != nil {
		return time.Unix(0, 0).UTC()
	}
	return parsed.UTC()
}

func parseRFC3339(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("time value is empty")
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func defaultID(prefix string) string {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, random)
}
