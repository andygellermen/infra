package invitation

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

type invitationFixture struct {
	service  *Service
	db       *sql.DB
	tenant   tenant.Tenant
	eventID  string
	seriesID string
	eventID2 string
	now      time.Time
}

func setupInvitationFixture(t *testing.T) invitationFixture {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "invitation-service.sqlite"))
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
	tenantItem, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "inv-tenant",
		Name:          "Invitation Tenant",
		PublicBaseURL: "http://localhost:8080/inv-tenant",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	eventRepo := event.NewRepository(sqlDB)
	series, err := eventRepo.CreateSeries(context.Background(), tenantItem.ID, event.CreateSeriesParams{
		Slug:  "spring-series",
		Title: "Spring Series",
	})
	if err != nil {
		t.Fatalf("create event series: %v", err)
	}

	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	evt, err := eventRepo.CreateEvent(context.Background(), tenantItem.ID, event.CreateEventParams{
		SeriesID: series.ID,
		Slug:     "spring-workshop",
		Title:    "Spring Workshop",
		StartsAt: now.Add(48 * time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	evt2, err := eventRepo.CreateEvent(context.Background(), tenantItem.ID, event.CreateEventParams{
		Slug:     "other-workshop",
		Title:    "Other Workshop",
		StartsAt: now.Add(72 * time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create second event: %v", err)
	}

	service := NewService(sqlDB)
	service.nowFn = func() time.Time { return now }

	return invitationFixture{
		service:  service,
		db:       sqlDB,
		tenant:   tenantItem,
		eventID:  evt.ID,
		seriesID: series.ID,
		eventID2: evt2.ID,
		now:      now,
	}
}

func createRegistrationForInvitationTest(t *testing.T, fx invitationFixture, email string) string {
	t.Helper()
	participantID := fx.service.idFn("par")
	registrationID := fx.service.idFn("reg")
	now := fx.now.Format(time.RFC3339)

	if _, err := fx.db.ExecContext(
		context.Background(),
		`INSERT INTO participants (id, tenant_id, email, name, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?)`,
		participantID,
		fx.tenant.ID,
		email,
		"Invitation Tester",
		now,
		now,
	); err != nil {
		t.Fatalf("insert participant: %v", err)
	}

	if _, err := fx.db.ExecContext(
		context.Background(),
		`INSERT INTO registrations (
      id, tenant_id, event_id, participant_id, status, participation_type, quantity,
      source, privacy_accepted_at, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?)`,
		registrationID,
		fx.tenant.ID,
		fx.eventID,
		participantID,
		"confirmed",
		"onsite",
		"public_page",
		now,
		now,
		now,
	); err != nil {
		t.Fatalf("insert registration: %v", err)
	}
	return registrationID
}

func TestCreateResolveAndApplyInvitationCode(t *testing.T) {
	fx := setupInvitationFixture(t)

	maxUses := 2
	maxUsesPerEmail := 1
	discountValue := 20
	created, err := fx.service.CreateLink(context.Background(), fx.tenant.ID, CreateLinkInput{
		EventID:         fx.eventID,
		Code:            "freunde20",
		InviteType:      InviteTypeDiscountPercent,
		DiscountValue:   &discountValue,
		MaxUses:         &maxUses,
		MaxUsesPerEmail: &maxUsesPerEmail,
	})
	if err != nil {
		t.Fatalf("create invitation link: %v", err)
	}
	if created.Code != "FREUNDE20" {
		t.Fatalf("expected normalized code FREUNDE20, got %q", created.Code)
	}

	resolved, err := fx.service.ResolveCode(context.Background(), ResolveInput{
		TenantID:         fx.tenant.ID,
		EventID:          fx.eventID,
		ParticipantEmail: "max@example.com",
		Code:             "freunde20",
		BaseAmountCents:  5000,
	})
	if err != nil {
		t.Fatalf("resolve invitation code: %v", err)
	}
	if resolved.DiscountAmountCents != 1000 {
		t.Fatalf("expected discount 1000, got %d", resolved.DiscountAmountCents)
	}
	if resolved.FinalAmountCents != 4000 {
		t.Fatalf("expected final amount 4000, got %d", resolved.FinalAmountCents)
	}

	registrationID := createRegistrationForInvitationTest(t, fx, "max@example.com")
	tx, err := fx.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	_, err = fx.service.ApplyCodeToRegistrationTx(context.Background(), tx, ApplyCodeToRegistrationInput{
		TenantID:         fx.tenant.ID,
		EventID:          fx.eventID,
		RegistrationID:   registrationID,
		ParticipantEmail: "max@example.com",
		Code:             "FREUNDE20",
		BaseAmountCents:  5000,
	})
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("apply invitation code: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit apply tx: %v", err)
	}

	var inviteID string
	if err := fx.db.QueryRowContext(
		context.Background(),
		`SELECT COALESCE(invite_id, '') FROM registrations WHERE id = ?`,
		registrationID,
	).Scan(&inviteID); err != nil {
		t.Fatalf("query registration invite_id: %v", err)
	}
	if inviteID != created.ID {
		t.Fatalf("expected invite_id %q, got %q", created.ID, inviteID)
	}

	_, err = fx.service.ResolveCode(context.Background(), ResolveInput{
		TenantID:         fx.tenant.ID,
		EventID:          fx.eventID,
		ParticipantEmail: "max@example.com",
		Code:             "FREUNDE20",
		BaseAmountCents:  5000,
	})
	if !errors.Is(err, ErrInvitationEmailExceeded) {
		t.Fatalf("expected ErrInvitationEmailExceeded, got %v", err)
	}
}

func TestResolveInvitationGuards(t *testing.T) {
	fx := setupInvitationFixture(t)

	discountValue := 5000
	expiresAt := fx.now.Add(-1 * time.Hour)
	_, err := fx.service.CreateLink(context.Background(), fx.tenant.ID, CreateLinkInput{
		SeriesID:      fx.seriesID,
		Code:          "fullvoucher",
		InviteType:    InviteTypeVoucherFixed,
		DiscountValue: &discountValue,
		Status:        StatusActive,
		ExpiresAt:     &expiresAt,
	})
	if err != nil {
		t.Fatalf("create expired invitation: %v", err)
	}

	_, err = fx.service.ResolveCode(context.Background(), ResolveInput{
		TenantID:         fx.tenant.ID,
		EventID:          fx.eventID,
		ParticipantEmail: "any@example.com",
		Code:             "FULLVOUCHER",
		BaseAmountCents:  2000,
	})
	if !errors.Is(err, ErrInvitationExpired) {
		t.Fatalf("expected ErrInvitationExpired, got %v", err)
	}

	status := StatusPaused
	updated, err := fx.service.CreateLink(context.Background(), fx.tenant.ID, CreateLinkInput{
		Code:       "scope-1",
		InviteType: InviteTypePlainInvitation,
		Status:     status,
	})
	if err != nil {
		t.Fatalf("create paused invitation: %v", err)
	}

	_, err = fx.service.ResolveCode(context.Background(), ResolveInput{
		TenantID:         fx.tenant.ID,
		EventID:          fx.eventID,
		ParticipantEmail: "any@example.com",
		Code:             updated.Code,
		BaseAmountCents:  1000,
	})
	if !errors.Is(err, ErrInvitationStatusInvalid) {
		t.Fatalf("expected ErrInvitationStatusInvalid, got %v", err)
	}

	active := StatusActive
	scoped, err := fx.service.UpdateLink(context.Background(), fx.tenant.ID, updated.ID, UpdateLinkInput{
		Status:   &active,
		SeriesID: &fx.seriesID,
	})
	if err != nil {
		t.Fatalf("update invitation scope: %v", err)
	}
	_, err = fx.service.ResolveCode(context.Background(), ResolveInput{
		TenantID:         fx.tenant.ID,
		EventID:          fx.eventID2,
		ParticipantEmail: "any@example.com",
		Code:             scoped.Code,
		BaseAmountCents:  1000,
	})
	if !errors.Is(err, ErrInvitationScopeMismatch) {
		t.Fatalf("expected ErrInvitationScopeMismatch, got %v", err)
	}
}
