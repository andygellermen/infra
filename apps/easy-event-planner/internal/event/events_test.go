package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func TestEventCRUDAndPublishFlow(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	created, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:      "erste-veranstaltung",
		Title:     "Erste Veranstaltung",
		StartsAt:  "2026-09-01T18:00:00Z",
		EndsAt:    "2026-09-01T20:00:00Z",
		Timezone:  "Europe/Berlin",
		OnlineURL: "https://meet.example.com/erste-veranstaltung",
	})
	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}
	if created.Status != EventStatusDraft {
		t.Fatalf("expected draft status, got %q", created.Status)
	}
	if created.IsPublic {
		t.Fatalf("expected is_public false on create")
	}

	listed, err := repo.ListEvents(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListEvents returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected one event, got %d", len(listed))
	}

	loaded, err := repo.GetEventByID(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("GetEventByID returned error: %v", err)
	}
	if loaded.Title != "Erste Veranstaltung" {
		t.Fatalf("expected title Erste Veranstaltung, got %q", loaded.Title)
	}

	newTitle := "Erste Veranstaltung Plus"
	newMode := ParticipationModeHybrid
	maxParticipants := 120
	updated, err := repo.UpdateEvent(context.Background(), tenantID, created.ID, UpdateEventParams{
		Title:             &newTitle,
		ParticipationMode: &newMode,
		MaxParticipants:   &maxParticipants,
	})
	if err != nil {
		t.Fatalf("UpdateEvent returned error: %v", err)
	}
	if updated.Title != "Erste Veranstaltung Plus" {
		t.Fatalf("expected updated title, got %q", updated.Title)
	}
	if updated.ParticipationMode != ParticipationModeHybrid {
		t.Fatalf("expected hybrid mode, got %q", updated.ParticipationMode)
	}
	if updated.MaxParticipants == nil || *updated.MaxParticipants != 120 {
		t.Fatalf("expected max participants 120, got %v", updated.MaxParticipants)
	}

	published, err := repo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("PublishEvent returned error: %v", err)
	}
	if published.Status != EventStatusScheduled {
		t.Fatalf("expected scheduled status after publish, got %q", published.Status)
	}
	if !published.IsPublic {
		t.Fatalf("expected is_public true after publish")
	}

	unpublished, err := repo.UnpublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("UnpublishEvent returned error: %v", err)
	}
	if unpublished.Status != EventStatusDraft {
		t.Fatalf("expected draft status after unpublish, got %q", unpublished.Status)
	}
	if unpublished.IsPublic {
		t.Fatalf("expected is_public false after unpublish")
	}

	deleted, err := repo.DeleteEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("DeleteEvent returned error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted=true")
	}
}

func TestCreateEventDuplicateSlug(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	_, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "gleicher-slug",
		Title:    "Event A",
		StartsAt: "2026-09-01T18:00:00Z",
	})
	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}

	_, err = repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "gleicher-slug",
		Title:    "Event B",
		StartsAt: "2026-09-02T18:00:00Z",
	})
	if !errors.Is(err, ErrEventSlugExists) {
		t.Fatalf("expected ErrEventSlugExists, got %v", err)
	}
}

func TestCreateEventRejectsSeriesFromOtherTenant(t *testing.T) {
	repo, tenantID, sqlDB := setupEventRepository(t)

	eventRepo := NewRepository(sqlDB)
	ownSeries, err := eventRepo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:  "shared-series",
		Title: "Shared",
	})
	if err != nil {
		t.Fatalf("create own series: %v", err)
	}
	if ownSeries.ID == "" {
		t.Fatalf("expected series id")
	}

	tenantRepo := tenant.NewRepository(sqlDB)
	secondTenant, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "second",
		Name:          "Second Tenant",
		PublicBaseURL: "https://events.example.com/second",
	})
	if err != nil {
		t.Fatalf("create second tenant: %v", err)
	}

	_, err = repo.CreateEvent(context.Background(), secondTenant.ID, CreateEventParams{
		SeriesID: ownSeries.ID,
		Slug:     "tenant-mismatch",
		Title:    "Mismatch Event",
		StartsAt: "2026-09-01T18:00:00Z",
	})
	if !errors.Is(err, ErrEventSeriesScopeMismatch) {
		t.Fatalf("expected ErrEventSeriesScopeMismatch, got %v", err)
	}
}

func TestPublishRejectsTerminalStatuses(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	created, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "terminal-status",
		Title:    "Terminal Status",
		StartsAt: "2026-09-01T18:00:00Z",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	// Direct status mutation for test setup, status changes beyond publish/unpublish are planned later.
	if _, execErr := repo.db.ExecContext(
		context.Background(),
		`UPDATE events SET status = ?, updated_at = ? WHERE id = ?`,
		EventStatusCancelled,
		time.Now().UTC().Format(time.RFC3339),
		created.ID,
	); execErr != nil {
		t.Fatalf("set cancelled status: %v", execErr)
	}

	_, err = repo.PublishEvent(context.Background(), tenantID, created.ID)
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
	_, err = repo.UnpublishEvent(context.Background(), tenantID, created.ID)
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

func TestUpdateEventRejectsEmptyPatch(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	created, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "no-patch-event",
		Title:    "No Patch Event",
		StartsAt: "2026-09-01T18:00:00Z",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	_, err = repo.UpdateEvent(context.Background(), tenantID, created.ID, UpdateEventParams{})
	if err == nil {
		t.Fatal("expected error for empty patch")
	}
}
