package notification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

const defaultErrorMessageLimit = 512

type WorkerConfig struct {
	PollInterval      time.Duration
	BatchSize         int
	ErrorMessageLimit int
}

type ProcessStats struct {
	Claimed int
	Sent    int
	Failed  int
}

type Worker struct {
	db    *sql.DB
	cfg   WorkerConfig
	mail  Mailer
	nowFn func() time.Time
}

type claimedJob struct {
	ID           string
	TenantID     string
	TemplateKey  string
	Recipient    string
	Subject      string
	BodyText     string
	BodyHTML     string
	MetadataJSON string
	FromEmail    string
	FromName     string
}

func NewWorker(sqlDB *sql.DB, mailer Mailer, cfg WorkerConfig) *Worker {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 3 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.ErrorMessageLimit <= 0 {
		cfg.ErrorMessageLimit = defaultErrorMessageLimit
	}

	if mailer == nil {
		mailer = &LogMailer{}
	}

	return &Worker{
		db:    sqlDB,
		cfg:   cfg,
		mail:  mailer,
		nowFn: func() time.Time { return time.Now().UTC() },
	}
}

func (w *Worker) Run(ctx context.Context) error {
	if w.db == nil {
		return fmt.Errorf("worker database is nil")
	}

	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	for {
		stats, err := w.ProcessOnce(ctx)
		if err != nil {
			log.Printf("email worker process cycle failed: %v", err)
		} else if stats.Claimed > 0 {
			log.Printf("email worker processed claimed=%d sent=%d failed=%d", stats.Claimed, stats.Sent, stats.Failed)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (w *Worker) ProcessOnce(ctx context.Context) (ProcessStats, error) {
	if w.db == nil {
		return ProcessStats{}, fmt.Errorf("worker database is nil")
	}

	var stats ProcessStats
	for i := 0; i < w.cfg.BatchSize; i++ {
		job, ok, err := w.claimNextJob(ctx)
		if err != nil {
			return stats, err
		}
		if !ok {
			return stats, nil
		}
		stats.Claimed++

		sendErr := w.mail.Send(ctx, Message{
			TenantID:     job.TenantID,
			TemplateKey:  job.TemplateKey,
			Recipient:    job.Recipient,
			Subject:      job.Subject,
			BodyText:     job.BodyText,
			BodyHTML:     job.BodyHTML,
			MetadataJSON: job.MetadataJSON,
			FromEmail:    job.FromEmail,
			FromName:     job.FromName,
		})
		if sendErr != nil {
			stats.Failed++
			if err := w.markFailed(ctx, job.ID, sendErr); err != nil {
				return stats, err
			}
			continue
		}

		stats.Sent++
		if err := w.markSent(ctx, job.ID); err != nil {
			return stats, err
		}
	}

	return stats, nil
}

func (w *Worker) claimNextJob(ctx context.Context) (claimedJob, bool, error) {
	now := w.nowFn().UTC().Format(time.RFC3339)

	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return claimedJob{}, false, fmt.Errorf("begin email job claim: %w", err)
	}

	row := tx.QueryRowContext(
		ctx,
		`SELECT ej.id, ej.tenant_id, ej.template_key, ej.recipient_email, ej.subject, ej.body_text, COALESCE(ej.body_html, ''),
            COALESCE(ej.metadata_json, ''), COALESCE(ts.sender_email, ''), COALESCE(ts.sender_name, '')
     FROM email_jobs ej
     LEFT JOIN tenant_settings ts ON ts.tenant_id = ej.tenant_id
     WHERE ej.status = ? AND (ej.scheduled_for IS NULL OR datetime(ej.scheduled_for) <= datetime(?))
     ORDER BY CASE WHEN ej.scheduled_for IS NULL THEN 0 ELSE 1 END, datetime(ej.scheduled_for), ej.created_at
     LIMIT 1`,
		EmailStatusQueued,
		now,
	)

	var job claimedJob
	if err := row.Scan(
		&job.ID,
		&job.TenantID,
		&job.TemplateKey,
		&job.Recipient,
		&job.Subject,
		&job.BodyText,
		&job.BodyHTML,
		&job.MetadataJSON,
		&job.FromEmail,
		&job.FromName,
	); err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return claimedJob{}, false, nil
		}
		return claimedJob{}, false, fmt.Errorf("select next email job: %w", err)
	}

	result, err := tx.ExecContext(
		ctx,
		`UPDATE email_jobs
     SET status = ?, error_message = NULL, updated_at = ?
     WHERE id = ? AND status = ?`,
		EmailStatusProcessing,
		now,
		job.ID,
		EmailStatusQueued,
	)
	if err != nil {
		_ = tx.Rollback()
		return claimedJob{}, false, fmt.Errorf("mark email job processing: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return claimedJob{}, false, fmt.Errorf("email job claim rows affected: %w", err)
	}
	if rowsAffected != 1 {
		_ = tx.Rollback()
		return claimedJob{}, false, nil
	}

	if err := tx.Commit(); err != nil {
		return claimedJob{}, false, fmt.Errorf("commit email job claim: %w", err)
	}
	return job, true, nil
}

func (w *Worker) markSent(ctx context.Context, jobID string) error {
	now := w.nowFn().UTC().Format(time.RFC3339)
	if _, err := w.db.ExecContext(
		ctx,
		`UPDATE email_jobs
     SET status = ?, sent_at = ?, error_message = NULL, updated_at = ?
     WHERE id = ?`,
		EmailStatusSent,
		now,
		now,
		strings.TrimSpace(jobID),
	); err != nil {
		return fmt.Errorf("mark email job sent: %w", err)
	}
	return nil
}

func (w *Worker) markFailed(ctx context.Context, jobID string, cause error) error {
	now := w.nowFn().UTC().Format(time.RFC3339)
	message := truncate(strings.TrimSpace(cause.Error()), w.cfg.ErrorMessageLimit)
	if _, err := w.db.ExecContext(
		ctx,
		`UPDATE email_jobs
     SET status = ?, error_message = ?, updated_at = ?
     WHERE id = ?`,
		EmailStatusFailed,
		nullable(message),
		now,
		strings.TrimSpace(jobID),
	); err != nil {
		return fmt.Errorf("mark email job failed: %w", err)
	}
	return nil
}

func truncate(value string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen]
}
