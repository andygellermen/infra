package notification

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	EmailStatusQueued     = "queued"
	EmailStatusProcessing = "processing"
	EmailStatusSent       = "sent"
	EmailStatusFailed     = "failed"
)

type EmailJob struct {
	ID           string
	TenantID     string
	TemplateKey  string
	Recipient    string
	Subject      string
	BodyText     string
	BodyHTML     string
	Status       string
	ScheduledFor sql.NullString
	SentAt       sql.NullString
	ErrorMessage sql.NullString
	MetadataJSON string
	CreatedAt    string
	UpdatedAt    string
}

type QueueInput struct {
	TenantID     string
	TemplateKey  string
	Recipient    string
	Subject      string
	BodyText     string
	BodyHTML     string
	ScheduledFor *time.Time
	MetadataJSON string
}

type Repository struct {
	db    *sql.DB
	nowFn func() time.Time
	idFn  func(prefix string) string
}

func NewRepository(sqlDB *sql.DB) *Repository {
	return &Repository{
		db:    sqlDB,
		nowFn: func() time.Time { return time.Now().UTC() },
		idFn:  defaultID,
	}
}

func (r *Repository) Queue(ctx context.Context, input QueueInput) (EmailJob, error) {
	if r.db == nil {
		return EmailJob{}, fmt.Errorf("notification repository database is nil")
	}

	tenantID := strings.TrimSpace(input.TenantID)
	templateKey := strings.TrimSpace(input.TemplateKey)
	recipient := strings.TrimSpace(input.Recipient)
	subject := strings.TrimSpace(input.Subject)
	bodyText := strings.TrimSpace(input.BodyText)
	bodyHTML := strings.TrimSpace(input.BodyHTML)
	metadata := strings.TrimSpace(input.MetadataJSON)

	if tenantID == "" {
		return EmailJob{}, fmt.Errorf("tenant id must not be empty")
	}
	if templateKey == "" {
		return EmailJob{}, fmt.Errorf("template key must not be empty")
	}
	if recipient == "" {
		return EmailJob{}, fmt.Errorf("recipient must not be empty")
	}
	if subject == "" {
		return EmailJob{}, fmt.Errorf("subject must not be empty")
	}
	if bodyText == "" && bodyHTML == "" {
		return EmailJob{}, fmt.Errorf("body text or body html must be provided")
	}

	now := r.nowFn().UTC().Format(time.RFC3339)
	jobID := r.idFn("emj")
	var scheduledFor any
	if input.ScheduledFor != nil {
		scheduledFor = input.ScheduledFor.UTC().Format(time.RFC3339)
	}
	if _, err := r.db.ExecContext(
		ctx,
		`INSERT INTO email_jobs (
      id, tenant_id, template_key, recipient_email, subject, body_text, body_html,
      status, scheduled_for, metadata_json, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		jobID,
		tenantID,
		templateKey,
		recipient,
		subject,
		bodyText,
		nullable(bodyHTML),
		EmailStatusQueued,
		scheduledFor,
		nullable(metadata),
		now,
		now,
	); err != nil {
		return EmailJob{}, fmt.Errorf("insert email job: %w", err)
	}

	return r.GetByID(ctx, jobID)
}

func (r *Repository) GetByID(ctx context.Context, jobID string) (EmailJob, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, template_key, recipient_email, subject, body_text, COALESCE(body_html, ''),
            status, scheduled_for, sent_at, error_message, COALESCE(metadata_json, ''), created_at, updated_at
     FROM email_jobs
     WHERE id = ?
     LIMIT 1`,
		strings.TrimSpace(jobID),
	)

	var job EmailJob
	if err := row.Scan(
		&job.ID,
		&job.TenantID,
		&job.TemplateKey,
		&job.Recipient,
		&job.Subject,
		&job.BodyText,
		&job.BodyHTML,
		&job.Status,
		&job.ScheduledFor,
		&job.SentAt,
		&job.ErrorMessage,
		&job.MetadataJSON,
		&job.CreatedAt,
		&job.UpdatedAt,
	); err != nil {
		return EmailJob{}, err
	}
	return job, nil
}

func defaultID(prefix string) string {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, random)
}

func nullable(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
