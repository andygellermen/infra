package event

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestListPublicEventsFilters(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)
	repo.nowFn = func() time.Time { return time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC) }

	publicSeries, err := repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:  "open-series",
		Title: "Open Series",
	})
	if err != nil {
		t.Fatalf("create public series: %v", err)
	}
	privateFlag := false
	privateSeries, err := repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:     "private-series",
		Title:    "Private Series",
		IsPublic: &privateFlag,
	})
	if err != nil {
		t.Fatalf("create private series: %v", err)
	}

	upcomingHybrid := createPublishedEvent(t, repo, tenantID, CreateEventParams{
		SeriesID:          publicSeries.ID,
		Slug:              "upcoming-hybrid",
		Title:             "Upcoming Hybrid",
		StartsAt:          "2026-09-12T10:00:00Z",
		ParticipationMode: ParticipationModeHybrid,
	})
	if upcomingHybrid.ID == "" {
		t.Fatalf("expected upcoming hybrid event id")
	}

	createPublishedEvent(t, repo, tenantID, CreateEventParams{
		SeriesID:          publicSeries.ID,
		Slug:              "past-onsite",
		Title:             "Past Onsite",
		StartsAt:          "2026-09-05T10:00:00Z",
		ParticipationMode: ParticipationModeOnsite,
	})

	createPublishedEvent(t, repo, tenantID, CreateEventParams{
		SeriesID:          privateSeries.ID,
		Slug:              "upcoming-private-series",
		Title:             "Upcoming Private Series",
		StartsAt:          "2026-09-13T10:00:00Z",
		ParticipationMode: ParticipationModeOnline,
	})

	_, err = repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "still-draft",
		Title:    "Still Draft",
		StartsAt: "2026-09-14T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("create draft event: %v", err)
	}

	defaultList, err := repo.ListPublicEvents(context.Background(), tenantID, PublicEventFilter{})
	if err != nil {
		t.Fatalf("list public events: %v", err)
	}
	if len(defaultList) != 2 {
		t.Fatalf("expected 2 upcoming public events, got %d", len(defaultList))
	}

	hybridOnly, err := repo.ListPublicEvents(context.Background(), tenantID, PublicEventFilter{
		Mode: ParticipationModeHybrid,
	})
	if err != nil {
		t.Fatalf("list public events with mode filter: %v", err)
	}
	if len(hybridOnly) != 1 || hybridOnly[0].Slug != "upcoming-hybrid" {
		t.Fatalf("expected only upcoming-hybrid event, got %+v", hybridOnly)
	}

	seriesOnly, err := repo.ListPublicEvents(context.Background(), tenantID, PublicEventFilter{
		SeriesSlug: "open-series",
	})
	if err != nil {
		t.Fatalf("list public events with series filter: %v", err)
	}
	if len(seriesOnly) != 1 || seriesOnly[0].Slug != "upcoming-hybrid" {
		t.Fatalf("expected only open-series upcoming event, got %+v", seriesOnly)
	}

	includePast, err := repo.ListPublicEvents(context.Background(), tenantID, PublicEventFilter{
		IncludePast: true,
	})
	if err != nil {
		t.Fatalf("list public events include_past=true: %v", err)
	}
	if len(includePast) != 3 {
		t.Fatalf("expected 3 public events with include_past=true, got %d", len(includePast))
	}

	from := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 12, 23, 59, 59, 0, time.UTC)
	windowed, err := repo.ListPublicEvents(context.Background(), tenantID, PublicEventFilter{
		IncludePast: true,
		From:        &from,
		To:          &to,
	})
	if err != nil {
		t.Fatalf("list public events by date window: %v", err)
	}
	if len(windowed) != 1 || windowed[0].Slug != "upcoming-hybrid" {
		t.Fatalf("expected one event in date window, got %+v", windowed)
	}

	limited, err := repo.ListPublicEvents(context.Background(), tenantID, PublicEventFilter{
		IncludePast: true,
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("list public events with limit: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("expected one event with limit=1, got %d", len(limited))
	}
}

func TestGetPublicEventBySlugVisibility(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	created, err := repo.CreateEvent(context.Background(), tenantID, CreateEventParams{
		Slug:     "public-lookup",
		Title:    "Public Lookup",
		StartsAt: "2026-09-12T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}

	_, err = repo.GetPublicEventBySlug(context.Background(), tenantID, created.Slug)
	if !errors.Is(err, ErrEventNotFound) {
		t.Fatalf("expected ErrEventNotFound for draft event, got %v", err)
	}

	_, err = repo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}

	loaded, err := repo.GetPublicEventBySlug(context.Background(), tenantID, created.Slug)
	if err != nil {
		t.Fatalf("get public event: %v", err)
	}
	if loaded.ID != created.ID {
		t.Fatalf("expected event id %s, got %s", created.ID, loaded.ID)
	}
}

func TestPublicSeriesVisibility(t *testing.T) {
	repo, tenantID, _ := setupEventRepository(t)

	_, err := repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:  "visible-series",
		Title: "Visible Series",
	})
	if err != nil {
		t.Fatalf("create visible series: %v", err)
	}
	privateFlag := false
	_, err = repo.CreateSeries(context.Background(), tenantID, CreateSeriesParams{
		Slug:     "hidden-series",
		Title:    "Hidden Series",
		IsPublic: &privateFlag,
	})
	if err != nil {
		t.Fatalf("create hidden series: %v", err)
	}

	items, err := repo.ListPublicSeries(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("list public series: %v", err)
	}
	if len(items) != 1 || items[0].Slug != "visible-series" {
		t.Fatalf("expected only visible series, got %+v", items)
	}

	_, err = repo.GetPublicSeriesBySlug(context.Background(), tenantID, "hidden-series")
	if !errors.Is(err, ErrSeriesNotFound) {
		t.Fatalf("expected ErrSeriesNotFound for hidden series, got %v", err)
	}
}

func createPublishedEvent(t *testing.T, repo *Repository, tenantID string, params CreateEventParams) Event {
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
