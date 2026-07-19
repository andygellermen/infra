package httpapp

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/calendar"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handleAdminCalendarFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.calendarService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Kalender-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if !a.requireTenantFeatureForPrincipal(w, r, principal.TenantID, tenant.FeatureCalendar, "FEATURE_DISABLED", "Kalender sind fuer diesen Mandanten nicht freigeschaltet.") {
		return
	}

	feed, err := a.calendarService.GetOrCreateOrganizerFeed(r.Context(), principal.TenantID)
	if err != nil {
		a.writeCalendarError(w, err)
		return
	}
	feed.URL = a.calendarService.OrganizerFeedEmbedURL(feed, principal.TenantSlug)

	writeJSON(w, http.StatusOK, map[string]any{
		"item": calendarFeedPayload(feed),
	})
}

func (a *App) handleAdminCalendarFeedRotateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.calendarService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Kalender-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if !a.requireTenantFeatureForPrincipal(w, r, principal.TenantID, tenant.FeatureCalendar, "FEATURE_DISABLED", "Kalender sind fuer diesen Mandanten nicht freigeschaltet.") {
		return
	}

	feed, err := a.calendarService.RotateOrganizerFeed(r.Context(), principal.TenantID)
	if err != nil {
		a.writeCalendarError(w, err)
		return
	}
	feed.URL = a.calendarService.OrganizerFeedEmbedURL(feed, principal.TenantSlug)

	writeJSON(w, http.StatusOK, map[string]any{
		"item": calendarFeedPayload(feed),
	})
}

func (a *App) handleAdminCalendarFeedEmbedURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.calendarService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Kalender-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if !a.requireTenantFeatureForPrincipal(w, r, principal.TenantID, tenant.FeatureCalendar, "FEATURE_DISABLED", "Kalender sind fuer diesen Mandanten nicht freigeschaltet.") {
		return
	}

	feed, err := a.calendarService.GetOrCreateOrganizerFeed(r.Context(), principal.TenantID)
	if err != nil {
		a.writeCalendarError(w, err)
		return
	}
	feed.URL = a.calendarService.OrganizerFeedEmbedURL(feed, principal.TenantSlug)

	writeJSON(w, http.StatusOK, map[string]any{
		"url":   feed.URL,
		"token": feed.Token,
	})
}

func (a *App) handleTenantOrganizerCalendarICS(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if a.calendarService == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	enabled, err := a.tenantFeatureEnabled(r.Context(), tenantItem.ID, tenant.FeatureCalendar)
	if err != nil || !enabled {
		http.NotFound(w, r)
		return
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	icsContent, err := a.calendarService.RenderOrganizerICS(r.Context(), tenantItem.ID, tenantItem.Slug, token)
	if err != nil {
		switch {
		case errors.Is(err, calendar.ErrInvalidFeedToken):
			http.Error(w, "invalid token", http.StatusUnauthorized)
		case errors.Is(err, calendar.ErrFeedNotFound):
			http.NotFound(w, r)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	filename := tenantItem.Slug + "-organizer.ics"
	writeICS(w, filename, icsContent, false)
}

func (a *App) handlePublicRegistrationCalendar(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, registrationID string) {
	if a.calendarService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Kalender-Service ist nicht verfuegbar.")
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Token fehlt.")
		return
	}

	icsContent, err := a.calendarService.RenderParticipantICS(r.Context(), tenantItem.ID, tenantItem.Slug, registrationID, token)
	if err != nil {
		switch {
		case errors.Is(err, calendar.ErrRegistrationNotFound):
			writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Anmeldung nicht gefunden.")
		case errors.Is(err, calendar.ErrInvalidRegistrationToken):
			writeAPIError(w, http.StatusUnauthorized, "INVALID_MAGIC_LINK", "Kalender-Link ist ungueltig.")
		default:
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		return
	}

	filename := strings.TrimSpace(registrationID) + ".ics"
	if filename == ".ics" {
		filename = "registration.ics"
	}
	writeICS(w, filename, icsContent, true)
}

func calendarFeedPayload(feed calendar.OrganizerFeed) map[string]any {
	var lastAccessedAt any
	if feed.LastAccessed != nil {
		lastAccessedAt = feed.LastAccessed.UTC().Format(time.RFC3339)
	}
	var rotatedAt any
	if feed.RotatedAt != nil {
		rotatedAt = feed.RotatedAt.UTC().Format(time.RFC3339)
	}

	return map[string]any{
		"id":               feed.ID,
		"tenant_id":        feed.TenantID,
		"feed_type":        feed.FeedType,
		"status":           feed.Status,
		"token":            feed.Token,
		"url":              feed.URL,
		"last_accessed_at": lastAccessedAt,
		"created_at":       feed.CreatedAt.UTC().Format(time.RFC3339),
		"rotated_at":       rotatedAt,
	}
}

func writeICS(w http.ResponseWriter, filename, content string, attachment bool) {
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	disposition := "inline"
	if attachment {
		disposition = "attachment"
	}
	w.Header().Set("Content-Disposition", disposition+`; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(content))
}

func (a *App) writeCalendarError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, calendar.ErrFeedNotFound):
		writeAPIError(w, http.StatusNotFound, "CALENDAR_FEED_NOT_FOUND", "Kalender-Feed nicht gefunden.")
	case errors.Is(err, calendar.ErrInvalidFeedToken):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_MAGIC_LINK", "Kalender-Link ist ungueltig.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}
