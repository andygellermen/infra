package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	OrganizerSummaryTemplateKey = "organizer_morning_summary"
)

type organizerSummaryEvent struct {
	ID              string
	TenantID        string
	Slug            string
	Title           string
	StartsAt        time.Time
	Timezone        string
	LocationName    string
	MaxParticipants *int
}

type organizerSummaryRecipient struct {
	Name  string
	Email string
}

type organizerSummaryParticipant struct {
	Name              string
	Email             string
	Phone             string
	Status            string
	ParticipationType string
}

type organizerSummaryStats struct {
	Confirmed           int
	ReservedActive      int
	PaymentPending      int
	Waitlist            int
	VerificationPending int
}

type organizerSummaryMetadata struct {
	Kind          string `json:"kind"`
	EventID       string `json:"event_id"`
	EventSlug     string `json:"event_slug"`
	SummaryDate   string `json:"summary_date"`
	RecipientMail string `json:"recipient_email"`
	GeneratedAt   string `json:"generated_at"`
}

func (w *Worker) enqueueOrganizerSummaries(ctx context.Context) (int, error) {
	now := w.nowFn().UTC()
	events, err := w.listOrganizerSummaryEvents(ctx, now)
	if err != nil {
		return 0, err
	}
	if len(events) == 0 {
		return 0, nil
	}

	repo := NewRepository(w.db)
	repo.nowFn = w.nowFn

	enqueued := 0
	for _, item := range events {
		location := loadLocationOrUTC(item.Timezone)
		nowLocal := now.In(location)
		startLocal := item.StartsAt.In(location)
		if !sameDate(startLocal, nowLocal) {
			continue
		}

		summaryDate := nowLocal.Format("2006-01-02")
		stats, err := w.loadOrganizerSummaryStats(ctx, item.TenantID, item.ID, now)
		if err != nil {
			return enqueued, err
		}
		participants, err := w.listOrganizerSummaryParticipants(ctx, item.TenantID, item.ID, 120)
		if err != nil {
			return enqueued, err
		}
		recipients, err := w.listOrganizerSummaryRecipients(ctx, item.TenantID)
		if err != nil {
			return enqueued, err
		}
		if len(recipients) == 0 {
			continue
		}

		subject := buildOrganizerSummarySubject(item, nowLocal)
		for _, recipient := range recipients {
			exists, err := w.organizerSummaryAlreadyQueued(ctx, item.TenantID, item.ID, summaryDate, recipient.Email)
			if err != nil {
				return enqueued, err
			}
			if exists {
				continue
			}

			body := buildOrganizerSummaryBody(item, recipient, nowLocal, stats, participants)
			metadataRaw, err := json.Marshal(organizerSummaryMetadata{
				Kind:          "organizer_summary",
				EventID:       item.ID,
				EventSlug:     item.Slug,
				SummaryDate:   summaryDate,
				RecipientMail: recipient.Email,
				GeneratedAt:   now.Format(time.RFC3339),
			})
			if err != nil {
				return enqueued, fmt.Errorf("marshal organizer summary metadata: %w", err)
			}

			if _, err := repo.Queue(ctx, QueueInput{
				TenantID:     item.TenantID,
				TemplateKey:  OrganizerSummaryTemplateKey,
				Recipient:    recipient.Email,
				Subject:      subject,
				BodyText:     body,
				MetadataJSON: string(metadataRaw),
			}); err != nil {
				return enqueued, fmt.Errorf("queue organizer summary email: %w", err)
			}
			enqueued++
		}
	}

	return enqueued, nil
}

func (w *Worker) listOrganizerSummaryEvents(ctx context.Context, now time.Time) ([]organizerSummaryEvent, error) {
	// UTC window broad enough to include same-day events across timezones (-12h to +14h).
	windowStart := now.Add(-14 * time.Hour).Format(time.RFC3339)
	windowEnd := now.Add(38 * time.Hour).Format(time.RFC3339)

	rows, err := w.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, slug, title, starts_at, COALESCE(timezone, 'UTC'), COALESCE(location_name, ''), max_participants
     FROM events
     WHERE status IN ('scheduled', 'changed', 'postponed')
       AND starts_at >= ?
       AND starts_at < ?
     ORDER BY starts_at ASC`,
		windowStart,
		windowEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("list organizer summary events: %w", err)
	}
	defer rows.Close()

	result := make([]organizerSummaryEvent, 0)
	for rows.Next() {
		var (
			item        organizerSummaryEvent
			startsAtRaw string
			maxValue    sql.NullInt64
		)
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.Slug,
			&item.Title,
			&startsAtRaw,
			&item.Timezone,
			&item.LocationName,
			&maxValue,
		); err != nil {
			return nil, fmt.Errorf("scan organizer summary event: %w", err)
		}
		startsAt, err := time.Parse(time.RFC3339, startsAtRaw)
		if err != nil {
			return nil, fmt.Errorf("parse organizer summary event starts_at: %w", err)
		}
		item.StartsAt = startsAt.UTC()
		if maxValue.Valid {
			value := int(maxValue.Int64)
			item.MaxParticipants = &value
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organizer summary events: %w", err)
	}

	return result, nil
}

func (w *Worker) listOrganizerSummaryRecipients(ctx context.Context, tenantID string) ([]organizerSummaryRecipient, error) {
	rows, err := w.db.QueryContext(
		ctx,
		`SELECT COALESCE(name, ''), email
     FROM tenant_users
     WHERE tenant_id = ?
       AND status = 'active'
       AND role IN ('owner', 'admin', 'event_manager')
     ORDER BY role ASC, created_at ASC`,
		strings.TrimSpace(tenantID),
	)
	if err != nil {
		return nil, fmt.Errorf("list organizer summary recipients: %w", err)
	}
	defer rows.Close()

	seen := map[string]struct{}{}
	result := make([]organizerSummaryRecipient, 0)
	for rows.Next() {
		var item organizerSummaryRecipient
		if err := rows.Scan(&item.Name, &item.Email); err != nil {
			return nil, fmt.Errorf("scan organizer summary recipient: %w", err)
		}
		item.Email = strings.ToLower(strings.TrimSpace(item.Email))
		if item.Email == "" {
			continue
		}
		if _, ok := seen[item.Email]; ok {
			continue
		}
		seen[item.Email] = struct{}{}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organizer summary recipients: %w", err)
	}
	return result, nil
}

func (w *Worker) loadOrganizerSummaryStats(ctx context.Context, tenantID, eventID string, now time.Time) (organizerSummaryStats, error) {
	row := w.db.QueryRowContext(
		ctx,
		`SELECT
       COALESCE(SUM(CASE WHEN status = 'confirmed' THEN 1 ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN status = 'reserved' AND (reserved_until IS NULL OR reserved_until > ?) THEN 1 ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN status = 'payment_pending' AND (reserved_until IS NULL OR reserved_until > ?) THEN 1 ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN status = 'waitlist' THEN 1 ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN status = 'verification_pending' THEN 1 ELSE 0 END), 0)
     FROM registrations
     WHERE tenant_id = ? AND event_id = ?`,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
		strings.TrimSpace(tenantID),
		strings.TrimSpace(eventID),
	)

	var stats organizerSummaryStats
	if err := row.Scan(
		&stats.Confirmed,
		&stats.ReservedActive,
		&stats.PaymentPending,
		&stats.Waitlist,
		&stats.VerificationPending,
	); err != nil {
		return organizerSummaryStats{}, fmt.Errorf("load organizer summary stats: %w", err)
	}
	return stats, nil
}

func (w *Worker) listOrganizerSummaryParticipants(ctx context.Context, tenantID, eventID string, limit int) ([]organizerSummaryParticipant, error) {
	if limit <= 0 {
		limit = 120
	}
	rows, err := w.db.QueryContext(
		ctx,
		`SELECT COALESCE(p.name, ''), COALESCE(p.email, ''), COALESCE(p.phone, ''), r.status, COALESCE(r.participation_type, '')
     FROM registrations r
     LEFT JOIN participants p ON p.id = r.participant_id
     WHERE r.tenant_id = ? AND r.event_id = ?
       AND r.status IN ('confirmed', 'reserved', 'payment_pending', 'waitlist', 'verification_pending')
     ORDER BY CASE r.status
       WHEN 'confirmed' THEN 1
       WHEN 'reserved' THEN 2
       WHEN 'payment_pending' THEN 3
       WHEN 'waitlist' THEN 4
       WHEN 'verification_pending' THEN 5
       ELSE 9 END,
       r.created_at ASC
     LIMIT ?`,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(eventID),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list organizer summary participants: %w", err)
	}
	defer rows.Close()

	result := make([]organizerSummaryParticipant, 0)
	for rows.Next() {
		var item organizerSummaryParticipant
		if err := rows.Scan(&item.Name, &item.Email, &item.Phone, &item.Status, &item.ParticipationType); err != nil {
			return nil, fmt.Errorf("scan organizer summary participant: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organizer summary participants: %w", err)
	}
	return result, nil
}

func (w *Worker) organizerSummaryAlreadyQueued(ctx context.Context, tenantID, eventID, summaryDate, recipientEmail string) (bool, error) {
	row := w.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
     FROM email_jobs
     WHERE tenant_id = ?
       AND template_key = ?
       AND recipient_email = ?
       AND status IN ('queued', 'processing', 'sent')
       AND COALESCE(metadata_json, '') LIKE ?
       AND COALESCE(metadata_json, '') LIKE ?`,
		strings.TrimSpace(tenantID),
		OrganizerSummaryTemplateKey,
		strings.ToLower(strings.TrimSpace(recipientEmail)),
		`%"event_id":"`+strings.TrimSpace(eventID)+`"%`,
		`%"summary_date":"`+strings.TrimSpace(summaryDate)+`"%`,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check organizer summary dedupe: %w", err)
	}
	return count > 0, nil
}

func buildOrganizerSummarySubject(eventItem organizerSummaryEvent, nowLocal time.Time) string {
	dateLabel := nowLocal.Format("02.01.2006")
	title := strings.TrimSpace(eventItem.Title)
	if title == "" {
		title = "Veranstaltung"
	}
	return fmt.Sprintf("Tagesuebersicht %s - %s", title, dateLabel)
}

func buildOrganizerSummaryBody(
	eventItem organizerSummaryEvent,
	recipient organizerSummaryRecipient,
	nowLocal time.Time,
	stats organizerSummaryStats,
	participants []organizerSummaryParticipant,
) string {
	var b strings.Builder
	name := strings.TrimSpace(recipient.Name)
	if name == "" {
		name = "Veranstalter-Team"
	}
	startLocal := eventItem.StartsAt.In(loadLocationOrUTC(eventItem.Timezone))

	activeSeats := stats.Confirmed + stats.ReservedActive + stats.PaymentPending
	freeSeatsLabel := "unbegrenzt"
	if eventItem.MaxParticipants != nil {
		free := *eventItem.MaxParticipants - activeSeats
		if free < 0 {
			free = 0
		}
		freeSeatsLabel = fmt.Sprintf("%d von %d frei", free, *eventItem.MaxParticipants)
	}

	b.WriteString("Hallo ")
	b.WriteString(name)
	b.WriteString(",\n\n")
	b.WriteString("hier ist die Tagesuebersicht fuer eure Veranstaltung.\n\n")
	b.WriteString("Event: ")
	b.WriteString(strings.TrimSpace(eventItem.Title))
	b.WriteString("\n")
	b.WriteString("Datum: ")
	b.WriteString(startLocal.Format("02.01.2006"))
	b.WriteString("\n")
	b.WriteString("Start: ")
	b.WriteString(startLocal.Format("15:04"))
	b.WriteString(" (")
	b.WriteString(strings.TrimSpace(eventItem.Timezone))
	b.WriteString(")\n")
	if strings.TrimSpace(eventItem.LocationName) != "" {
		b.WriteString("Ort: ")
		b.WriteString(strings.TrimSpace(eventItem.LocationName))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString("Anmeldestatus:\n")
	b.WriteString(fmt.Sprintf("- bestaetigt: %d\n", stats.Confirmed))
	b.WriteString(fmt.Sprintf("- reserviert (aktiv): %d\n", stats.ReservedActive))
	b.WriteString(fmt.Sprintf("- zahlung ausstehend: %d\n", stats.PaymentPending))
	b.WriteString(fmt.Sprintf("- warteliste: %d\n", stats.Waitlist))
	b.WriteString(fmt.Sprintf("- verifizierung ausstehend: %d\n", stats.VerificationPending))
	b.WriteString(fmt.Sprintf("- freie plaetze: %s\n", freeSeatsLabel))
	b.WriteString("\n")

	if len(participants) == 0 {
		b.WriteString("Derzeit liegen keine aktiven Anmeldungen vor.\n")
		return b.String()
	}

	b.WriteString("Teilnehmerliste:\n")
	for i, p := range participants {
		displayName := strings.TrimSpace(p.Name)
		if displayName == "" {
			displayName = "(ohne Name)"
		}
		displayEmail := strings.TrimSpace(p.Email)
		if displayEmail == "" {
			displayEmail = "(ohne E-Mail)"
		}
		status := strings.TrimSpace(p.Status)
		mode := strings.TrimSpace(p.ParticipationType)
		if mode == "" {
			mode = "-"
		}
		b.WriteString(fmt.Sprintf("%d. %s <%s> - %s / %s\n", i+1, displayName, displayEmail, status, mode))
	}

	return b.String()
}

func loadLocationOrUTC(name string) *time.Location {
	value := strings.TrimSpace(name)
	if value == "" {
		return time.UTC
	}
	location, err := time.LoadLocation(value)
	if err != nil {
		return time.UTC
	}
	return location
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
