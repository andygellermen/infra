package notification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
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

func TestWorkerGeneratesOrganizerSummaryOnEventDay(t *testing.T) {
	sqlDB, tenantID := setupNotificationDB(t)
	now := time.Date(2026, 9, 10, 6, 0, 0, 0, time.UTC)
	eventRepo := event.NewRepository(sqlDB)

	eventItem := createPublishedEventForNotificationTest(t, eventRepo, tenantID, event.CreateEventParams{
		Slug:            "event-day-summary",
		Title:           "Event Day Summary",
		StartsAt:        now.Add(8 * time.Hour).Format(time.RFC3339),
		Timezone:        "UTC",
		MaxParticipants: intPointer(8),
	})

	addOrganizerUser(t, sqlDB, tenantID, "owner@example.com", "Owner", "owner")

	insertRegistrationWithParticipant(t, sqlDB, tenantID, eventItem.ID, "Alice", "alice@example.com", "confirmed", "onsite", nil)
	insertRegistrationWithParticipant(t, sqlDB, tenantID, eventItem.ID, "Bob", "bob@example.com", "waitlist", "online", nil)
	insertRegistrationWithParticipant(t, sqlDB, tenantID, eventItem.ID, "Carla", "carla@example.com", "verification_pending", "onsite", nil)
	insertRegistrationWithParticipant(t, sqlDB, tenantID, eventItem.ID, "Dave", "dave@example.com", "reserved", "onsite", timePointer(now.Add(1*time.Hour)))
	insertRegistrationWithParticipant(t, sqlDB, tenantID, eventItem.ID, "Eve", "eve@example.com", "payment_pending", "hybrid", timePointer(now.Add(2*time.Hour)))

	mailer := &captureMailer{}
	worker := NewWorker(sqlDB, mailer, WorkerConfig{BatchSize: 10, PollInterval: time.Second})
	worker.nowFn = func() time.Time { return now }

	stats, err := worker.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("ProcessOnce returned error: %v", err)
	}
	if stats.Queued != 1 {
		t.Fatalf("expected one queued organizer summary, got %d", stats.Queued)
	}
	if stats.Claimed != 1 || stats.Sent != 1 || stats.Failed != 0 {
		t.Fatalf("unexpected process stats: %+v", stats)
	}
	if len(mailer.messages) != 1 {
		t.Fatalf("expected one delivered message, got %d", len(mailer.messages))
	}

	message := mailer.messages[0]
	if message.TemplateKey != OrganizerSummaryTemplateKey {
		t.Fatalf("expected template key %q, got %q", OrganizerSummaryTemplateKey, message.TemplateKey)
	}
	if !strings.Contains(message.Subject, "Event Day Summary") {
		t.Fatalf("expected subject to include event title, got %q", message.Subject)
	}
	if !strings.Contains(message.BodyText, "- bestaetigt: 1") {
		t.Fatalf("expected confirmed count in body, got %q", message.BodyText)
	}
	if !strings.Contains(message.BodyText, "- warteliste: 1") {
		t.Fatalf("expected waitlist count in body, got %q", message.BodyText)
	}
	if !strings.Contains(message.BodyText, "alice@example.com") {
		t.Fatalf("expected participant listing in body, got %q", message.BodyText)
	}

	statsSecondRun, err := worker.ProcessOnce(context.Background())
	if err != nil {
		t.Fatalf("second ProcessOnce returned error: %v", err)
	}
	if statsSecondRun.Queued != 0 || statsSecondRun.Claimed != 0 {
		t.Fatalf("expected no duplicate queue/send in second run, got %+v", statsSecondRun)
	}
	if len(mailer.messages) != 1 {
		t.Fatalf("expected no additional messages after second run, got %d", len(mailer.messages))
	}
}

func addOrganizerUser(t *testing.T, dbHandle *sql.DB, tenantID, email, name, role string) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := dbHandle.ExecContext(
		context.Background(),
		`INSERT INTO tenant_users (id, tenant_id, email, name, role, status, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, 'active', ?, ?)`,
		fmt.Sprintf("usr_%d", time.Now().UTC().UnixNano()),
		tenantID,
		email,
		name,
		role,
		now,
		now,
	); err != nil {
		t.Fatalf("insert organizer user: %v", err)
	}
}

func createPublishedEventForNotificationTest(t *testing.T, repo *event.Repository, tenantID string, params event.CreateEventParams) event.Event {
	t.Helper()
	created, err := repo.CreateEvent(context.Background(), tenantID, params)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	published, err := repo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}
	return published
}

func insertRegistrationWithParticipant(
	t *testing.T,
	dbHandle *sql.DB,
	tenantID,
	eventID,
	name,
	email,
	status,
	participationType string,
	reservedUntil *time.Time,
) {
	t.Helper()

	now := time.Now().UTC()
	safeMail := strings.NewReplacer("@", "_", ".", "_", "+", "_", "-", "_").Replace(strings.ToLower(strings.TrimSpace(email)))
	participantID := fmt.Sprintf("par_%s_%d", safeMail, now.UnixNano())
	registrationID := fmt.Sprintf("reg_%s_%d", safeMail, now.UnixNano())

	if _, err := dbHandle.ExecContext(
		context.Background(),
		`INSERT INTO participants (id, tenant_id, email, phone, name, created_at, updated_at)
     VALUES (?, ?, ?, NULL, ?, ?, ?)`,
		participantID,
		tenantID,
		strings.ToLower(strings.TrimSpace(email)),
		name,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	); err != nil {
		t.Fatalf("insert participant: %v", err)
	}

	var reservedUntilRaw any
	if reservedUntil != nil {
		reservedUntilRaw = reservedUntil.UTC().Format(time.RFC3339)
	}
	if _, err := dbHandle.ExecContext(
		context.Background(),
		`INSERT INTO registrations (
      id, tenant_id, event_id, participant_id, status, participation_type, quantity, reserved_until, source, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, 1, ?, 'public_page', ?, ?)`,
		registrationID,
		tenantID,
		eventID,
		participantID,
		status,
		participationType,
		reservedUntilRaw,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	); err != nil {
		t.Fatalf("insert registration: %v", err)
	}
}

func intPointer(value int) *int {
	return &value
}

func timePointer(value time.Time) *time.Time {
	return &value
}
