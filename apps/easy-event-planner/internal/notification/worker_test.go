package notification

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

type captureMailer struct {
	messages []Message
	sendErr  error
}

func (m *captureMailer) Send(_ context.Context, message Message) error {
	m.messages = append(m.messages, message)
	if m.sendErr != nil {
		return m.sendErr
	}
	return nil
}

func setupNotificationDB(t *testing.T) (*sql.DB, string) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "notification-test.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	migrator := db.NewMigrator(sqlDB, migrations.Files, ".")
	if _, err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	tenantRepo := tenant.NewRepository(sqlDB)
	createdTenant, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "demo",
		Name:          "Demo Tenant",
		PublicBaseURL: "https://events.example.com/demo",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	if _, err := tenantRepo.UpsertSettings(context.Background(), tenant.UpsertTenantSettingsParams{
		TenantID: createdTenant.ID,
		Settings: tenant.TenantSettingsInput{
			SenderEmail: "sender@example.com",
			SenderName:  "Demo Sender",
		},
	}); err != nil {
		t.Fatalf("upsert tenant settings: %v", err)
	}

	return sqlDB, createdTenant.ID
}

func TestWorkerProcessOnceSendsQueuedEmail(t *testing.T) {
	sqlDB, tenantID := setupNotificationDB(t)
	repo := NewRepository(sqlDB)

	job, err := repo.Queue(context.Background(), QueueInput{
		TenantID:    tenantID,
		TemplateKey: "organizer_summary",
		Recipient:   "owner@example.com",
		Subject:     "Morgenuebersicht",
		BodyText:    "Alles bereit.",
	})
	if err != nil {
		t.Fatalf("queue email job: %v", err)
	}

	mailer := &captureMailer{}
	worker := NewWorker(sqlDB, mailer, WorkerConfig{BatchSize: 5, PollInterval: time.Second})

	stats, err := worker.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("ProcessOnce returned error: %v", err)
	}
	if stats.Claimed != 1 || stats.Sent != 1 || stats.Failed != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if len(mailer.messages) != 1 {
		t.Fatalf("expected one sent message, got %d", len(mailer.messages))
	}
	if mailer.messages[0].FromEmail != "sender@example.com" {
		t.Fatalf("expected sender@example.com as from email, got %q", mailer.messages[0].FromEmail)
	}

	stored, err := repo.GetByID(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("get email job: %v", err)
	}
	if stored.Status != EmailStatusSent {
		t.Fatalf("expected status sent, got %q", stored.Status)
	}
	if !stored.SentAt.Valid {
		t.Fatalf("expected sent_at to be set")
	}
}

func TestWorkerLeavesFutureScheduledJobQueued(t *testing.T) {
	sqlDB, tenantID := setupNotificationDB(t)
	repo := NewRepository(sqlDB)

	scheduled := time.Now().UTC().Add(2 * time.Hour)
	job, err := repo.Queue(context.Background(), QueueInput{
		TenantID:     tenantID,
		TemplateKey:  "future_notice",
		Recipient:    "owner@example.com",
		Subject:      "Future",
		BodyText:     "not due yet",
		ScheduledFor: &scheduled,
	})
	if err != nil {
		t.Fatalf("queue email job: %v", err)
	}

	mailer := &captureMailer{}
	worker := NewWorker(sqlDB, mailer, WorkerConfig{BatchSize: 2, PollInterval: time.Second})

	stats, err := worker.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("ProcessOnce returned error: %v", err)
	}
	if stats.Claimed != 0 || stats.Sent != 0 || stats.Failed != 0 {
		t.Fatalf("expected no processed jobs for future schedule, got %+v", stats)
	}
	if len(mailer.messages) != 0 {
		t.Fatalf("expected no messages to be sent")
	}

	stored, err := repo.GetByID(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("get email job: %v", err)
	}
	if stored.Status != EmailStatusQueued {
		t.Fatalf("expected status queued, got %q", stored.Status)
	}
}

func TestWorkerMarksFailedOnMailerError(t *testing.T) {
	sqlDB, tenantID := setupNotificationDB(t)
	repo := NewRepository(sqlDB)

	job, err := repo.Queue(context.Background(), QueueInput{
		TenantID:    tenantID,
		TemplateKey: "broken_delivery",
		Recipient:   "owner@example.com",
		Subject:     "Fail",
		BodyText:    "delivery should fail",
	})
	if err != nil {
		t.Fatalf("queue email job: %v", err)
	}

	mailer := &captureMailer{sendErr: errors.New("simulated delivery failure")}
	worker := NewWorker(sqlDB, mailer, WorkerConfig{
		BatchSize:         1,
		PollInterval:      time.Second,
		ErrorMessageLimit: 12,
	})

	stats, err := worker.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("ProcessOnce returned error: %v", err)
	}
	if stats.Claimed != 1 || stats.Sent != 0 || stats.Failed != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	stored, err := repo.GetByID(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("get email job: %v", err)
	}
	if stored.Status != EmailStatusFailed {
		t.Fatalf("expected status failed, got %q", stored.Status)
	}
	if !stored.ErrorMessage.Valid {
		t.Fatalf("expected error_message to be set")
	}
	if len(stored.ErrorMessage.String) != 12 {
		t.Fatalf("expected truncated error length 12, got %d", len(stored.ErrorMessage.String))
	}
	if !strings.Contains(stored.ErrorMessage.String, "simulated") {
		t.Fatalf("expected error_message to include source error, got %q", stored.ErrorMessage.String)
	}
}
