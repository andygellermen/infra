package httpapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
)

func TestPublicEventsAndDetailFlow(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)

	tenantItem, err := app.tenantRepo.LookupBySlug(context.Background(), tenantSlug)
	if err != nil {
		t.Fatalf("lookup tenant by slug: %v", err)
	}

	publicSeries, err := app.eventRepo.CreateSeries(context.Background(), tenantItem.ID, event.CreateSeriesParams{
		Slug:  "workshops",
		Title: "Workshops",
	})
	if err != nil {
		t.Fatalf("create public series: %v", err)
	}

	startsAt := time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339)
	published := createPublishedEventForPublicTest(t, app, tenantItem.ID, event.CreateEventParams{
		SeriesID:          publicSeries.ID,
		Slug:              "stress-management",
		Title:             "Stress Management",
		StartsAt:          startsAt,
		ParticipationMode: event.ParticipationModeHybrid,
	})

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/events", nil)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	if listPayload["total"] != float64(1) {
		t.Fatalf("expected total=1, got %v", listPayload["total"])
	}
	tenantPayload, ok := listPayload["tenant"].(map[string]any)
	if !ok || tenantPayload["slug"] != tenantSlug {
		t.Fatalf("expected tenant payload with slug %q, got %v", tenantSlug, listPayload["tenant"])
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/events/"+published.Slug, nil)
	detailRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusOK {
		t.Fatalf("expected detail status 200, got %d", detailRec.Code)
	}
	detailPayload := decodeBody[map[string]any](t, detailRec)
	item, ok := detailPayload["item"].(map[string]any)
	if !ok {
		t.Fatalf("expected detail item payload")
	}
	if item["slug"] != published.Slug {
		t.Fatalf("expected detail slug %q, got %v", published.Slug, item["slug"])
	}
}

func TestPublicEventsFiltersAndValidation(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantItem, err := app.tenantRepo.LookupBySlug(context.Background(), tenantSlug)
	if err != nil {
		t.Fatalf("lookup tenant by slug: %v", err)
	}

	now := time.Now().UTC()
	createPublishedEventForPublicTest(t, app, tenantItem.ID, event.CreateEventParams{
		Slug:              "future-hybrid",
		Title:             "Future Hybrid",
		StartsAt:          now.Add(72 * time.Hour).Format(time.RFC3339),
		ParticipationMode: event.ParticipationModeHybrid,
	})
	createPublishedEventForPublicTest(t, app, tenantItem.ID, event.CreateEventParams{
		Slug:              "past-online",
		Title:             "Past Online",
		StartsAt:          now.Add(-72 * time.Hour).Format(time.RFC3339),
		ParticipationMode: event.ParticipationModeOnline,
	})

	defaultReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/events", nil)
	defaultRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(defaultRec, defaultReq)
	if defaultRec.Code != http.StatusOK {
		t.Fatalf("expected default list status 200, got %d", defaultRec.Code)
	}
	defaultPayload := decodeBody[map[string]any](t, defaultRec)
	if defaultPayload["total"] != float64(1) {
		t.Fatalf("expected default total=1 (only future), got %v", defaultPayload["total"])
	}

	includePastReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/events?include_past=true", nil)
	includePastRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(includePastRec, includePastReq)
	if includePastRec.Code != http.StatusOK {
		t.Fatalf("expected include_past status 200, got %d", includePastRec.Code)
	}
	includePastPayload := decodeBody[map[string]any](t, includePastRec)
	if includePastPayload["total"] != float64(2) {
		t.Fatalf("expected include_past total=2, got %v", includePastPayload["total"])
	}

	modeReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/events?include_past=true&mode=online", nil)
	modeRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(modeRec, modeReq)
	if modeRec.Code != http.StatusOK {
		t.Fatalf("expected mode filter status 200, got %d", modeRec.Code)
	}
	modePayload := decodeBody[map[string]any](t, modeRec)
	if modePayload["total"] != float64(1) {
		t.Fatalf("expected mode filter total=1, got %v", modePayload["total"])
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/events?limit=abc", nil)
	invalidRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(invalidRec, invalidReq)
	if invalidRec.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid limit status 400, got %d", invalidRec.Code)
	}
	invalidPayload := decodeBody[map[string]any](t, invalidRec)
	errorPayload := invalidPayload["error"].(map[string]any)
	if errorPayload["code"] != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %v", errorPayload["code"])
	}
}

func TestPublicSeriesRoutes(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantItem, err := app.tenantRepo.LookupBySlug(context.Background(), tenantSlug)
	if err != nil {
		t.Fatalf("lookup tenant by slug: %v", err)
	}

	visibleSeries, err := app.eventRepo.CreateSeries(context.Background(), tenantItem.ID, event.CreateSeriesParams{
		Slug:  "visible-series",
		Title: "Visible Series",
	})
	if err != nil {
		t.Fatalf("create visible series: %v", err)
	}
	privateFlag := false
	_, err = app.eventRepo.CreateSeries(context.Background(), tenantItem.ID, event.CreateSeriesParams{
		Slug:     "hidden-series",
		Title:    "Hidden Series",
		IsPublic: &privateFlag,
	})
	if err != nil {
		t.Fatalf("create hidden series: %v", err)
	}

	createPublishedEventForPublicTest(t, app, tenantItem.ID, event.CreateEventParams{
		SeriesID: visibleSeries.ID,
		Slug:     "series-event",
		Title:    "Series Event",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	seriesReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/series", nil)
	seriesRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(seriesRec, seriesReq)
	if seriesRec.Code != http.StatusOK {
		t.Fatalf("expected series list status 200, got %d", seriesRec.Code)
	}
	seriesPayload := decodeBody[map[string]any](t, seriesRec)
	if seriesPayload["total"] != float64(1) {
		t.Fatalf("expected visible series total=1, got %v", seriesPayload["total"])
	}

	seriesEventsReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/series/visible-series/events", nil)
	seriesEventsRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(seriesEventsRec, seriesEventsReq)
	if seriesEventsRec.Code != http.StatusOK {
		t.Fatalf("expected series events status 200, got %d", seriesEventsRec.Code)
	}
	seriesEventsPayload := decodeBody[map[string]any](t, seriesEventsRec)
	if seriesEventsPayload["total"] != float64(1) {
		t.Fatalf("expected series events total=1, got %v", seriesEventsPayload["total"])
	}

	hiddenSeriesReq := httptest.NewRequest(http.MethodGet, "/api/v1/public/"+tenantSlug+"/series/hidden-series/events", nil)
	hiddenSeriesRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(hiddenSeriesRec, hiddenSeriesReq)
	if hiddenSeriesRec.Code != http.StatusNotFound {
		t.Fatalf("expected hidden series status 404, got %d", hiddenSeriesRec.Code)
	}
}

func TestPublicTenantNotFound(t *testing.T) {
	app, _, _ := setupAuthApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/not-existing/events", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	payload := decodeBody[map[string]any](t, rec)
	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != "TENANT_NOT_FOUND" {
		t.Fatalf("expected TENANT_NOT_FOUND, got %v", errorPayload["code"])
	}
}

func createPublishedEventForPublicTest(t *testing.T, app *App, tenantID string, params event.CreateEventParams) event.Event {
	t.Helper()

	created, err := app.eventRepo.CreateEvent(context.Background(), tenantID, params)
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	published, err := app.eventRepo.PublishEvent(context.Background(), tenantID, created.ID)
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}
	return published
}
