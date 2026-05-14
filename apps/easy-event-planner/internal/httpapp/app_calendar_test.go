package httpapp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
)

func TestAdminCalendarFeedAndTenantICSFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	published := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "calendar-admin-event",
		Title:    "Calendar Admin Event",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	feedReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/calendar/feed", nil)
	feedReq.AddCookie(sessionCookie)
	feedRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(feedRec, feedReq)
	if feedRec.Code != http.StatusOK {
		t.Fatalf("expected feed status 200, got %d", feedRec.Code)
	}
	feedPayload := decodeBody[map[string]any](t, feedRec)
	item, ok := feedPayload["item"].(map[string]any)
	if !ok {
		t.Fatalf("expected feed item payload")
	}
	tokenOne, _ := item["token"].(string)
	feedURL, _ := item["url"].(string)
	if tokenOne == "" || feedURL == "" {
		t.Fatalf("expected feed token/url to be set")
	}
	parsedFeedURL, err := url.Parse(feedURL)
	if err != nil {
		t.Fatalf("parse feed url: %v", err)
	}

	tenantFeedReq := httptest.NewRequest(http.MethodGet, requestPathWithQuery(parsedFeedURL), nil)
	tenantFeedRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(tenantFeedRec, tenantFeedReq)
	if tenantFeedRec.Code != http.StatusOK {
		t.Fatalf("expected tenant feed status 200, got %d", tenantFeedRec.Code)
	}
	if !strings.Contains(tenantFeedRec.Header().Get("Content-Type"), "text/calendar") {
		t.Fatalf("expected text/calendar content type, got %q", tenantFeedRec.Header().Get("Content-Type"))
	}
	if !strings.HasPrefix(tenantFeedRec.Header().Get("Content-Disposition"), "inline;") {
		t.Fatalf("expected inline content disposition, got %q", tenantFeedRec.Header().Get("Content-Disposition"))
	}
	icsBody := tenantFeedRec.Body.String()
	if !strings.Contains(icsBody, "BEGIN:VCALENDAR") {
		t.Fatalf("expected calendar body")
	}
	if !strings.Contains(icsBody, "SUMMARY:"+published.Title) {
		t.Fatalf("expected organizer ICS to include event summary")
	}
	if !strings.Contains(icsBody, "/admin/events/"+published.ID) {
		t.Fatalf("expected organizer ICS to include admin event URL")
	}

	rotateReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/calendar/feed/rotate-token", nil)
	rotateReq.AddCookie(sessionCookie)
	rotateRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rotateRec, rotateReq)
	if rotateRec.Code != http.StatusOK {
		t.Fatalf("expected rotate status 200, got %d", rotateRec.Code)
	}
	rotatePayload := decodeBody[map[string]any](t, rotateRec)
	rotateItem, ok := rotatePayload["item"].(map[string]any)
	if !ok {
		t.Fatalf("expected rotate item payload")
	}
	tokenTwo, _ := rotateItem["token"].(string)
	if tokenTwo == "" || tokenTwo == tokenOne {
		t.Fatalf("expected rotated token to change")
	}

	oldTokenReq := httptest.NewRequest(http.MethodGet, "/"+tenantSlug+"/calendar/admin.ics?token="+url.QueryEscape(tokenOne), nil)
	oldTokenRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(oldTokenRec, oldTokenReq)
	if oldTokenRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected old token status 401, got %d", oldTokenRec.Code)
	}

	newTokenReq := httptest.NewRequest(http.MethodGet, "/"+tenantSlug+"/calendar/admin.ics?token="+url.QueryEscape(tokenTwo), nil)
	newTokenRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(newTokenRec, newTokenReq)
	if newTokenRec.Code != http.StatusOK {
		t.Fatalf("expected new token status 200, got %d", newTokenRec.Code)
	}

	embedReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/calendar/feed/embed-url", nil)
	embedReq.AddCookie(sessionCookie)
	embedRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(embedRec, embedReq)
	if embedRec.Code != http.StatusOK {
		t.Fatalf("expected embed-url status 200, got %d", embedRec.Code)
	}
	embedPayload := decodeBody[map[string]any](t, embedRec)
	if embedPayload["token"] != tokenTwo {
		t.Fatalf("expected embed token to match latest token, got %v", embedPayload["token"])
	}
}

func TestPublicRegistrationCalendarRouteAndVerifyPayload(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	published := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "calendar-public-event",
		Title:    "Calendar Public Event",
		StartsAt: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	startPayload := map[string]any{
		"event_id":           published.ID,
		"name":               "Mila Muster",
		"email":              "mila@example.com",
		"participation_type": "onsite",
		"privacy_accepted":   true,
	}
	startBody, _ := json.Marshal(startPayload)
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/start", bytes.NewReader(startBody))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusAccepted {
		t.Fatalf("expected start status 202, got %d", startRec.Code)
	}
	startResult := decodeBody[map[string]any](t, startRec)
	registrationID, _ := startResult["registration_id"].(string)
	if registrationID == "" {
		t.Fatalf("expected registration_id in start payload")
	}

	verifyToken := extractVerifyTokenFromJobInHTTPTest(t, app, tenantID, registrationID)
	verifyPayload := map[string]any{"token": verifyToken}
	verifyBody, _ := json.Marshal(verifyPayload)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected verify status 200, got %d", verifyRec.Code)
	}
	verifyResult := decodeBody[map[string]any](t, verifyRec)
	calendarURL, _ := verifyResult["calendar_url"].(string)
	if calendarURL == "" {
		t.Fatalf("expected calendar_url in verify payload")
	}
	parsedCalendarURL, err := url.Parse(calendarURL)
	if err != nil {
		t.Fatalf("parse participant calendar url: %v", err)
	}

	calendarReq := httptest.NewRequest(http.MethodGet, requestPathWithQuery(parsedCalendarURL), nil)
	calendarRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(calendarRec, calendarReq)
	if calendarRec.Code != http.StatusOK {
		t.Fatalf("expected participant calendar status 200, got %d", calendarRec.Code)
	}
	if !strings.Contains(calendarRec.Header().Get("Content-Type"), "text/calendar") {
		t.Fatalf("expected text/calendar content type, got %q", calendarRec.Header().Get("Content-Type"))
	}
	if !strings.HasPrefix(calendarRec.Header().Get("Content-Disposition"), "attachment;") {
		t.Fatalf("expected attachment content disposition, got %q", calendarRec.Header().Get("Content-Disposition"))
	}
	participantICS := calendarRec.Body.String()
	if !strings.Contains(participantICS, "SUMMARY:"+published.Title) {
		t.Fatalf("expected participant ICS to include event summary")
	}
	if strings.Contains(participantICS, "/admin/events/") {
		t.Fatalf("participant ICS must not contain admin links")
	}
	if !strings.Contains(participantICS, "/api/v1/public/"+tenantSlug+"/events/"+published.Slug) {
		t.Fatalf("participant ICS must contain public event URL")
	}

	invalidTokenReq := httptest.NewRequest(http.MethodGet, parsedCalendarURL.Path+"?token=invalid", nil)
	invalidTokenRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(invalidTokenRec, invalidTokenReq)
	if invalidTokenRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected invalid token status 401, got %d", invalidTokenRec.Code)
	}
	invalidPayload := decodeBody[map[string]any](t, invalidTokenRec)
	invalidError := invalidPayload["error"].(map[string]any)
	if invalidError["code"] != "INVALID_MAGIC_LINK" {
		t.Fatalf("expected INVALID_MAGIC_LINK, got %v", invalidError["code"])
	}
}

func requestPathWithQuery(parsed *url.URL) string {
	path := parsed.Path
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	return path
}
