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
	if published.PublishedAt == nil {
		t.Fatalf("expected published_at to be set after publish")
	}

	unpublished, err := repo.UnpublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("UnpublishEvent returned error: %v", err)
	}
	if unpublished.Status != EventStatusScheduled {
		t.Fatalf("expected scheduled status after unpublish, got %q", unpublished.Status)
	}
	if !unpublished.IsPublic {
		t.Fatalf("expected public intent to remain after unpublish")
	}
	if unpublished.PublishedAt != nil {
		t.Fatalf("expected published_at to be cleared after unpublish")
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

func TestUpdateEventClearsPublishedAtWhenPublicVisibilityIsRemoved(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	isPublic := true
	created, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "private-again",
		Title:    "Private Again",
		StartsAt: "2026-09-01T18:00:00Z",
		IsPublic: &isPublic,
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	published, err := repo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}
	if published.PublishedAt == nil {
		t.Fatalf("expected published_at after publish")
	}

	disablePublic := false
	updated, err := repo.UpdateEvent(context.Background(), tenantID, created.ID, UpdateEventParams{
		IsPublic: &disablePublic,
	})
	if err != nil {
		t.Fatalf("update event: %v", err)
	}
	if updated.IsPublic {
		t.Fatalf("expected is_public=false after update")
	}
	if updated.PublishedAt != nil {
		t.Fatalf("expected published_at to be cleared when public visibility is removed")
	}
	if updated.IsPublished() {
		t.Fatalf("expected event to no longer be published")
	}
}

func TestEventVisibilityAndRegistrationWindows(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	isPublic := true
	created, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:                 "visibility-window",
		Title:                "Visibility Window",
		StartsAt:             "2026-09-10T18:00:00Z",
		IsPublic:             &isPublic,
		PublicVisibleFrom:    "2026-08-01T09:00:00Z",
		RegistrationOpensAt:  "2026-08-02T10:00:00Z",
		RegistrationClosesAt: "2026-08-09T18:00:00Z",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	published, err := repo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}

	if published.IsVisibleAt(time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected event to stay hidden before public_visible_from")
	}
	if !published.IsVisibleAt(time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected event to become visible at public_visible_from")
	}
	if published.IsRegistrationOpenAt(time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected registration to stay closed before registration_opens_at")
	}
	if !published.IsRegistrationOpenAt(time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected registration to be open inside configured window")
	}
	if published.IsRegistrationOpenAt(time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected registration to be closed after registration_closes_at")
	}
}

func TestEventMaintenanceActions(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	created, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "maintenance-event",
		Title:    "Maintenance Event",
		StartsAt: "2026-09-01T18:00:00Z",
		EndsAt:   "2026-09-01T20:00:00Z",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	published, err := repo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}
	if published.Status != EventStatusScheduled {
		t.Fatalf("expected scheduled status, got %q", published.Status)
	}

	newTitle := "Maintenance Event Updated"
	changed, err := repo.UpdateEvent(context.Background(), tenantID, created.ID, UpdateEventParams{
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("update event after publish: %v", err)
	}
	if changed.Status != EventStatusChanged {
		t.Fatalf("expected changed status after update, got %q", changed.Status)
	}

	postponed, err := repo.PostponeEvent(
		context.Background(),
		tenantID,
		created.ID,
		"2026-10-01T18:30:00Z",
		"2026-10-01T20:30:00Z",
		"Termin wurde organisatorisch verschoben.",
	)
	if err != nil {
		t.Fatalf("postpone event: %v", err)
	}
	if postponed.Status != EventStatusPostponed {
		t.Fatalf("expected postponed status, got %q", postponed.Status)
	}
	if postponed.ChangeNote == "" {
		t.Fatalf("expected postpone change note")
	}
	if postponed.StartsAt.Format(time.RFC3339) != "2026-10-01T18:30:00Z" {
		t.Fatalf("expected updated starts_at, got %s", postponed.StartsAt.Format(time.RFC3339))
	}

	completed, err := repo.MarkEventCompleted(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("mark event completed: %v", err)
	}
	if completed.Status != EventStatusCompleted {
		t.Fatalf("expected completed status, got %q", completed.Status)
	}
	if completed.RegistrationEnabled {
		t.Fatalf("expected registration disabled on completed event")
	}
	if completed.WaitlistEnabled {
		t.Fatalf("expected waitlist disabled on completed event")
	}

	_, err = repo.CancelEvent(context.Background(), tenantID, created.ID, "Abgesagt", "Nachtraegliche Absage")
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition when canceling completed event, got %v", err)
	}
}

func TestCancelEvent(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	created, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "cancel-event",
		Title:    "Cancel Event",
		StartsAt: "2026-09-01T18:00:00Z",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if _, err := repo.PublishEvent(context.Background(), tenantID, created.ID); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	cancelled, err := repo.CancelEvent(
		context.Background(),
		tenantID,
		created.ID,
		"Der Veranstaltungsort ist nicht verfuegbar.",
		"Wir informieren ueber einen Ersatztermin.",
	)
	if err != nil {
		t.Fatalf("cancel event: %v", err)
	}
	if cancelled.Status != EventStatusCancelled {
		t.Fatalf("expected cancelled status, got %q", cancelled.Status)
	}
	if cancelled.CancelledReason == "" {
		t.Fatalf("expected cancelled reason to be set")
	}
	if cancelled.RegistrationEnabled {
		t.Fatalf("expected registration disabled on cancelled event")
	}
	if cancelled.WaitlistEnabled {
		t.Fatalf("expected waitlist disabled on cancelled event")
	}

	_, err = repo.PostponeEvent(context.Background(), tenantID, created.ID, "2026-10-01T18:00:00Z", "", "Nachholtermin")
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition when postponing cancelled event, got %v", err)
	}
}
