package calendar

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func setupCalendarService(t *testing.T) (*Service, *tenant.Repository, *event.Repository, tenant.Tenant) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "calendar-service.sqlite"))
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
		Slug:          "calendar-tenant",
		Name:          "Calendar Tenant",
		PublicBaseURL: "http://localhost:8080/calendar-tenant",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	calendarSvc := NewService(sqlDB, Config{
		BaseURL:     "http://localhost:8080",
		TokenPepper: "calendar-test-pepper",
	})
	return calendarSvc, tenantRepo, event.NewRepository(sqlDB), tenantItem
}

func TestOrganizerFeedTokenStableAcrossFetchAndRotation(t *testing.T) {
	service, _, eventRepo, tenantItem := setupCalendarService(t)

	first, err := service.GetOrCreateOrganizerFeed(context.Background(), tenantItem.ID)
	if err != nil {
		t.Fatalf("create organizer feed: %v", err)
	}
	second, err := service.GetOrCreateOrganizerFeed(context.Background(), tenantItem.ID)
	if err != nil {
		t.Fatalf("fetch organizer feed: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected same feed id, got %q vs %q", second.ID, first.ID)
	}
	if second.Token != first.Token {
		t.Fatalf("expected stable token across fetch, got %q vs %q", second.Token, first.Token)
	}

	createdEvent, err := eventRepo.CreateEvent(context.Background(), tenantItem.ID, event.CreateEventParams{
		Slug:     "calendar-service-event",
		Title:    "Calendar Service Event",
		StartsAt: time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if _, err := eventRepo.PublishEvent(context.Background(), tenantItem.ID, createdEvent.ID); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	ics, err := service.RenderOrganizerICS(context.Background(), tenantItem.ID, tenantItem.Slug, second.Token)
	if err != nil {
		t.Fatalf("render organizer ics: %v", err)
	}
	if !strings.Contains(ics, "SUMMARY:Calendar Service Event") {
		t.Fatalf("expected event summary in organizer ICS")
	}

	rotated, err := service.RotateOrganizerFeed(context.Background(), tenantItem.ID)
	if err != nil {
		t.Fatalf("rotate organizer feed: %v", err)
	}
	if rotated.Token == first.Token {
		t.Fatalf("expected rotated token to differ")
	}

	if _, err := service.RenderOrganizerICS(context.Background(), tenantItem.ID, tenantItem.Slug, first.Token); !errors.Is(err, ErrInvalidFeedToken) {
		t.Fatalf("expected invalid old token error, got %v", err)
	}
	if _, err := service.RenderOrganizerICS(context.Background(), tenantItem.ID, tenantItem.Slug, rotated.Token); err != nil {
		t.Fatalf("render organizer ics with rotated token: %v", err)
	}
}

func TestParticipantCalendarURLBuildsDeterministicPath(t *testing.T) {
	service, _, _, tenantItem := setupCalendarService(t)

	link := service.ParticipantCalendarURL(tenantItem.Slug, tenantItem.ID, "reg_123", "par_123")
	if !strings.Contains(link, "/api/v1/public/"+tenantItem.Slug+"/registrations/reg_123/calendar.ics?token=") {
		t.Fatalf("expected participant calendar path, got %q", link)
	}
	if got := service.ParticipantCalendarURL(tenantItem.Slug, tenantItem.ID, "reg_123", ""); got != "" {
		t.Fatalf("expected empty URL with missing participant id, got %q", got)
	}
}
