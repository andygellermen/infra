package httpapp

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handlePublicRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.eventRepo == nil || a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service ist nicht verfuegbar.")
		return
	}

	tenantSlug, routeType, routeSlug, ok := parsePublicPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Route nicht gefunden.")
		return
	}

	tenantItem, err := a.tenantRepo.LookupBySlug(r.Context(), tenantSlug)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) {
			writeAPIError(w, http.StatusNotFound, "TENANT_NOT_FOUND", "Mandant nicht gefunden.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Mandant konnte nicht geladen werden.")
		return
	}

	switch routeType {
	case "events_list":
		filter, err := parsePublicEventFilter(r)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		a.handlePublicEventsList(w, r, tenantItem, filter)
	case "event_detail":
		a.handlePublicEventDetail(w, r, tenantItem, routeSlug)
	case "series_list":
		a.handlePublicSeriesList(w, r, tenantItem)
	case "series_events":
		filter, err := parsePublicEventFilter(r)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		filter.SeriesSlug = routeSlug
		a.handlePublicSeriesEvents(w, r, tenantItem, routeSlug, filter)
	default:
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Route nicht gefunden.")
	}
}

func (a *App) handlePublicEventsList(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, filter event.PublicEventFilter) {
	items, err := a.eventRepo.ListPublicEvents(r.Context(), tenantItem.ID, filter)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicEventPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant": publicTenantPayload(tenantItem),
		"items":  result,
		"total":  len(result),
	})
}

func (a *App) handlePublicEventDetail(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, eventSlug string) {
	item, err := a.eventRepo.GetPublicEventBySlug(r.Context(), tenantItem.ID, eventSlug)
	if err != nil {
		if errors.Is(err, event.ErrEventNotFound) {
			writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
			return
		}
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant": publicTenantPayload(tenantItem),
		"item":   publicEventPayload(item),
	})
}

func (a *App) handlePublicSeriesList(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	items, err := a.eventRepo.ListPublicSeries(r.Context(), tenantItem.ID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Event-Serien konnten nicht geladen werden.")
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicSeriesPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant": publicTenantPayload(tenantItem),
		"items":  result,
		"total":  len(result),
	})
}

func (a *App) handlePublicSeriesEvents(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, seriesSlug string, filter event.PublicEventFilter) {
	seriesItem, err := a.eventRepo.GetPublicSeriesBySlug(r.Context(), tenantItem.ID, seriesSlug)
	if err != nil {
		if errors.Is(err, event.ErrSeriesNotFound) {
			writeAPIError(w, http.StatusNotFound, "EVENT_SERIES_NOT_FOUND", "Event-Serie nicht gefunden.")
			return
		}
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	items, err := a.eventRepo.ListPublicEvents(r.Context(), tenantItem.ID, filter)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, publicEventPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant": publicTenantPayload(tenantItem),
		"series": publicSeriesPayload(seriesItem),
		"items":  result,
		"total":  len(result),
	})
}

func parsePublicPath(path string) (tenantSlug, routeType, routeSlug string, ok bool) {
	const prefix = "/api/v1/public/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", "", false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if rest == "" {
		return "", "", "", false
	}

	parts := strings.Split(rest, "/")
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return "", "", "", false
		}
	}

	tenantSlug = strings.TrimSpace(parts[0])
	switch len(parts) {
	case 2:
		switch parts[1] {
		case "events":
			return tenantSlug, "events_list", "", true
		case "series":
			return tenantSlug, "series_list", "", true
		default:
			return "", "", "", false
		}
	case 3:
		if parts[1] != "events" {
			return "", "", "", false
		}
		return tenantSlug, "event_detail", strings.TrimSpace(parts[2]), true
	case 4:
		if parts[1] != "series" || parts[3] != "events" {
			return "", "", "", false
		}
		return tenantSlug, "series_events", strings.TrimSpace(parts[2]), true
	default:
		return "", "", "", false
	}
}

func parsePublicEventFilter(r *http.Request) (event.PublicEventFilter, error) {
	query := r.URL.Query()
	filter := event.PublicEventFilter{
		SeriesSlug: strings.TrimSpace(query.Get("series")),
		Mode:       strings.TrimSpace(query.Get("mode")),
	}

	limitRaw := strings.TrimSpace(query.Get("limit"))
	if limitRaw != "" {
		limit, err := strconv.Atoi(limitRaw)
		if err != nil {
			return event.PublicEventFilter{}, fmt.Errorf("limit must be a number")
		}
		filter.Limit = limit
	}

	includePastRaw := strings.TrimSpace(query.Get("include_past"))
	if includePastRaw != "" {
		includePast, err := strconv.ParseBool(includePastRaw)
		if err != nil {
			return event.PublicEventFilter{}, fmt.Errorf("include_past must be true or false")
		}
		filter.IncludePast = includePast
	}

	from, err := parsePublicDateFilter(query.Get("from"), false)
	if err != nil {
		return event.PublicEventFilter{}, err
	}
	to, err := parsePublicDateFilter(query.Get("to"), true)
	if err != nil {
		return event.PublicEventFilter{}, err
	}
	filter.From = from
	filter.To = to
	return filter, nil
}

func parsePublicDateFilter(raw string, dayEnd bool) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		utc := parsed.UTC()
		return &utc, nil
	}

	parsedDate, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, fmt.Errorf("date filter %q must be RFC3339 or YYYY-MM-DD", value)
	}
	parsedDate = parsedDate.UTC()
	if dayEnd {
		parsedDate = parsedDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}
	return &parsedDate, nil
}

func publicTenantPayload(item tenant.Tenant) map[string]any {
	return map[string]any{
		"slug":            item.Slug,
		"name":            item.Name,
		"public_base_url": item.PublicBaseURL,
		"timezone":        item.DefaultTimezone,
		"locale":          item.DefaultLocale,
	}
}

func publicSeriesPayload(item event.EventSeries) map[string]any {
	return map[string]any{
		"id":                    item.ID,
		"slug":                  item.Slug,
		"title":                 item.Title,
		"description":           item.Description,
		"default_location_name": item.DefaultLocationName,
		"default_address":       item.DefaultAddress,
		"default_online_url":    item.DefaultOnlineURL,
		"is_public":             item.IsPublic,
	}
}

func publicEventPayload(item event.PublicEvent) map[string]any {
	var endsAt any
	if item.EndsAt != nil {
		endsAt = item.EndsAt.UTC().Format(time.RFC3339)
	}
	var maxParticipants any
	if item.MaxParticipants != nil {
		maxParticipants = *item.MaxParticipants
	}
	var series any
	if strings.TrimSpace(item.SeriesSlug) != "" {
		series = map[string]any{
			"id":    item.SeriesID,
			"slug":  item.SeriesSlug,
			"title": item.SeriesTitle,
		}
	}

	return map[string]any{
		"id":                   item.ID,
		"series":               series,
		"slug":                 item.Slug,
		"title":                item.Title,
		"subtitle":             item.Subtitle,
		"description":          item.Description,
		"starts_at":            item.StartsAt.UTC().Format(time.RFC3339),
		"ends_at":              endsAt,
		"timezone":             item.Timezone,
		"location_name":        item.LocationName,
		"address":              item.Address,
		"online_url":           item.OnlineURL,
		"participation_mode":   item.ParticipationMode,
		"status":               item.Status,
		"is_public":            item.IsPublic,
		"registration_enabled": item.RegistrationEnabled,
		"waitlist_enabled":     item.WaitlistEnabled,
		"max_participants":     maxParticipants,
		"change_note":          item.ChangeNote,
		"cancelled_reason":     item.CancelledReason,
	}
}
