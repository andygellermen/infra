package httpapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/invitation"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/registration"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handlePublicRoutes(w http.ResponseWriter, r *http.Request) {
	if a.eventRepo == nil || a.tenantRepo == nil || a.regService == nil || a.snippetRepo == nil {
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

	if handled := a.handlePublicCORS(w, r, tenantItem, routeType); handled {
		return
	}

	switch routeType {
	case "events_list":
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		filter, err := parsePublicEventFilter(r)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		a.handlePublicEventsList(w, r, tenantItem, filter)
	case "event_detail":
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handlePublicEventDetail(w, r, tenantItem, routeSlug)
	case "series_list":
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handlePublicSeriesList(w, r, tenantItem)
	case "series_events":
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		filter, err := parsePublicEventFilter(r)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		filter.SeriesSlug = routeSlug
		a.handlePublicSeriesEvents(w, r, tenantItem, routeSlug, filter)
	case "snippet_events":
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handlePublicSnippetEvents(w, r, tenantItem)
	case "registrations_start":
		if r.Method != http.MethodPost {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handlePublicRegistrationStart(w, r, tenantItem)
	case "registrations_verify":
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handlePublicRegistrationVerify(w, r, tenantItem)
	case "registrations_calendar":
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handlePublicRegistrationCalendar(w, r, tenantItem, routeSlug)
	case "invitations_resolve":
		a.handlePublicInvitationResolve(w, r, tenantItem)
	case "participants_portal_request":
		a.handlePublicParticipantPortalRequest(w, r, tenantItem)
	case "participants_portal_verify":
		a.handlePublicParticipantPortalVerify(w, r, tenantItem)
	case "participants_portal_me":
		a.handlePublicParticipantPortalMe(w, r, tenantItem)
	case "participants_portal_logout":
		a.handlePublicParticipantPortalLogout(w, r, tenantItem)
	case "participants_portal_registrations":
		a.handlePublicParticipantPortalRegistrations(w, r, tenantItem)
	case "participants_portal_registration_cancel":
		a.handlePublicParticipantPortalRegistrationCancel(w, r, tenantItem, routeSlug)
	case "participants_portal_certificates":
		a.handlePublicParticipantPortalCertificates(w, r, tenantItem)
	case "participants_portal_certificate":
		a.handlePublicParticipantPortalCertificateGet(w, r, tenantItem, routeSlug)
	case "participants_portal_certificate_download":
		a.handlePublicParticipantPortalCertificateDownload(w, r, tenantItem, routeSlug)
	case "certificates_verify_public":
		a.handlePublicCertificateVerify(w, r, tenantItem)
	case "payments_paypal_create_order":
		if r.Method != http.MethodPost {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handlePublicPayPalCreateOrder(w, r, tenantItem)
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
		switch parts[1] {
		case "events":
			return tenantSlug, "event_detail", strings.TrimSpace(parts[2]), true
		case "snippet":
			if parts[2] != "events" {
				return "", "", "", false
			}
			return tenantSlug, "snippet_events", "", true
		case "registrations":
			switch parts[2] {
			case "start":
				return tenantSlug, "registrations_start", "", true
			case "verify":
				return tenantSlug, "registrations_verify", "", true
			default:
				return "", "", "", false
			}
		case "invitations":
			if parts[2] != "resolve" {
				return "", "", "", false
			}
			return tenantSlug, "invitations_resolve", "", true
		case "certificates":
			if parts[2] != "verify" {
				return "", "", "", false
			}
			return tenantSlug, "certificates_verify_public", "", true
		default:
			return "", "", "", false
		}
	case 4:
		switch parts[1] {
		case "series":
			if parts[3] != "events" {
				return "", "", "", false
			}
			return tenantSlug, "series_events", strings.TrimSpace(parts[2]), true
		case "registrations":
			if parts[3] != "calendar.ics" {
				return "", "", "", false
			}
			return tenantSlug, "registrations_calendar", strings.TrimSpace(parts[2]), true
		case "payments":
			if parts[2] != "paypal" || parts[3] != "create-order" {
				return "", "", "", false
			}
			return tenantSlug, "payments_paypal_create_order", "", true
		case "participants":
			if parts[2] != "portal" {
				return "", "", "", false
			}
			switch parts[3] {
			case "request":
				return tenantSlug, "participants_portal_request", "", true
			case "verify":
				return tenantSlug, "participants_portal_verify", "", true
			case "me":
				return tenantSlug, "participants_portal_me", "", true
			case "logout":
				return tenantSlug, "participants_portal_logout", "", true
			case "registrations":
				return tenantSlug, "participants_portal_registrations", "", true
			case "certificates":
				return tenantSlug, "participants_portal_certificates", "", true
			default:
				return "", "", "", false
			}
		default:
			return "", "", "", false
		}
	case 5:
		if parts[1] != "participants" || parts[2] != "portal" || parts[3] != "certificates" {
			return "", "", "", false
		}
		return tenantSlug, "participants_portal_certificate", strings.TrimSpace(parts[4]), true
	case 6:
		if parts[1] == "participants" && parts[2] == "portal" && parts[3] == "registrations" && parts[5] == "cancel" {
			return tenantSlug, "participants_portal_registration_cancel", strings.TrimSpace(parts[4]), true
		}
		if parts[1] == "participants" && parts[2] == "portal" && parts[3] == "certificates" && parts[5] == "download" {
			return tenantSlug, "participants_portal_certificate_download", strings.TrimSpace(parts[4]), true
		}
		return "", "", "", false
	default:
		return "", "", "", false
	}
}

func (a *App) handlePublicRegistrationStart(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	var request struct {
		EventID           string `json:"event_id"`
		Name              string `json:"name"`
		Email             string `json:"email"`
		Phone             string `json:"phone"`
		ParticipationType string `json:"participation_type"`
		InviteCode        string `json:"invite_code"`
		InviteAmountCents int    `json:"invite_amount_cents"`
		PrivacyAccepted   bool   `json:"privacy_accepted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	result, err := a.regService.Start(r.Context(), registration.StartInput{
		TenantID:          tenantItem.ID,
		TenantSlug:        tenantItem.Slug,
		EventID:           request.EventID,
		Name:              request.Name,
		Email:             request.Email,
		Phone:             request.Phone,
		ParticipationType: request.ParticipationType,
		InviteCode:        request.InviteCode,
		InviteAmountCents: request.InviteAmountCents,
		PrivacyAccepted:   request.PrivacyAccepted,
		RequestIP:         clientIP(r),
		UserAgent:         strings.TrimSpace(r.UserAgent()),
	})
	if err != nil {
		a.writePublicRegistrationError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"ok":                true,
		"registration_id":   result.RegistrationID,
		"participant_id":    result.ParticipantID,
		"event_id":          result.EventID,
		"status":            result.Status,
		"verify_expires_at": result.VerifyExpires.UTC().Format(time.RFC3339),
		"message":           "Bitte bestaetige die Anmeldung ueber den Link in der E-Mail.",
		"invite": map[string]any{
			"id":                    emptyToNil(result.InviteID),
			"code":                  emptyToNil(result.InviteCode),
			"discount_amount_cents": result.DiscountAmountCents,
			"credit_amount_cents":   result.CreditAmountCents,
			"final_amount_cents":    result.FinalAmountCents,
			"sponsored":             result.Sponsored,
		},
	})
}

func (a *App) handlePublicRegistrationVerify(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if r.Method == http.MethodPost {
		var request struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
			return
		}
		token = strings.TrimSpace(request.Token)
	}

	result, err := a.regService.Verify(r.Context(), registration.VerifyInput{
		TenantID:  tenantItem.ID,
		RawToken:  token,
		RequestIP: clientIP(r),
		UserAgent: strings.TrimSpace(r.UserAgent()),
	})
	if err != nil {
		if r.Method == http.MethodGet {
			a.renderPublicRegistrationVerifyPage(w, r, tenantItem, registration.VerifyResult{}, err)
			return
		}
		a.writePublicRegistrationError(w, err)
		return
	}

	payload := a.publicRegistrationVerifyPayload(tenantItem, result)
	if r.Method == http.MethodGet {
		a.renderPublicRegistrationVerifyPage(w, r, tenantItem, result, nil)
		return
	}
	writeJSON(w, http.StatusOK, payload)
}

func (a *App) publicRegistrationVerifyPayload(tenantItem tenant.Tenant, result registration.VerifyResult) map[string]any {
	payload := map[string]any{
		"ok":              true,
		"registration_id": result.RegistrationID,
		"participant_id":  result.ParticipantID,
		"event_id":        result.EventID,
		"status":          result.Status,
	}
	if result.ConfirmedAt != nil {
		payload["confirmed_at"] = result.ConfirmedAt.UTC().Format(time.RFC3339)
	}
	if strings.TrimSpace(result.WaitlistID) != "" {
		payload["waitlist"] = map[string]any{
			"id":       result.WaitlistID,
			"position": result.WaitlistPos,
		}
	}
	if a.calendarService != nil {
		calendarURL := a.calendarService.ParticipantCalendarURL(
			tenantItem.Slug,
			tenantItem.ID,
			result.RegistrationID,
			result.ParticipantID,
		)
		if calendarURL != "" {
			payload["calendar_url"] = calendarURL
		}
	}
	return payload
}

func (a *App) writePublicRegistrationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, registration.ErrEventNotFound):
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
	case errors.Is(err, registration.ErrRegistrationDisabled), errors.Is(err, registration.ErrRegistrationClosed):
		writeAPIError(w, http.StatusConflict, "REGISTRATION_CLOSED", "Anmeldung ist fuer diese Veranstaltung derzeit nicht moeglich.")
	case errors.Is(err, registration.ErrPrivacyAcceptanceRequired):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Datenschutzerklaerung muss akzeptiert werden.")
	case errors.Is(err, registration.ErrAlreadyRegistered):
		writeAPIError(w, http.StatusConflict, "ALREADY_REGISTERED", "Teilnahme ist bereits bestaetigt.")
	case errors.Is(err, registration.ErrAlreadyWaitlisted):
		writeAPIError(w, http.StatusConflict, "ALREADY_WAITLISTED", "Teilnahme steht bereits auf der Warteliste.")
	case errors.Is(err, registration.ErrInvalidVerificationToken), errors.Is(err, registration.ErrRegistrationVerificationNil):
		writeAPIError(w, http.StatusBadRequest, "INVALID_MAGIC_LINK", "Magic-Link ist ungueltig.")
	case errors.Is(err, registration.ErrExpiredVerificationToken):
		writeAPIError(w, http.StatusBadRequest, "EXPIRED_MAGIC_LINK", "Magic-Link ist abgelaufen.")
	case errors.Is(err, registration.ErrRegistrationNotFound):
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Anmeldung nicht gefunden.")
	case errors.Is(err, registration.ErrRegistrationState):
		writeAPIError(w, http.StatusConflict, "REGISTRATION_STATE_INVALID", "Anmeldung kann nicht bestaetigt werden.")
	case errors.Is(err, registration.ErrEventFull):
		writeAPIError(w, http.StatusConflict, "EVENT_FULL", "Die Veranstaltung ist ausgebucht. Eine Warteliste ist verfuegbar.")
	case errors.Is(err, invitation.ErrInvitationNotFound):
		writeAPIError(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "Einladungscode wurde nicht gefunden.")
	case errors.Is(err, invitation.ErrInvitationStatusInvalid):
		writeAPIError(w, http.StatusConflict, "INVITATION_INACTIVE", "Einladungscode ist derzeit nicht aktiv.")
	case errors.Is(err, invitation.ErrInvitationNotStarted):
		writeAPIError(w, http.StatusConflict, "INVITATION_NOT_STARTED", "Einladungscode ist noch nicht aktiv.")
	case errors.Is(err, invitation.ErrInvitationExpired):
		writeAPIError(w, http.StatusConflict, "INVITATION_EXPIRED", "Einladungscode ist abgelaufen.")
	case errors.Is(err, invitation.ErrInvitationScopeMismatch):
		writeAPIError(w, http.StatusConflict, "INVITATION_SCOPE_MISMATCH", "Einladungscode passt nicht zu dieser Veranstaltung.")
	case errors.Is(err, invitation.ErrInvitationUsageExceeded), errors.Is(err, invitation.ErrInvitationEmailExceeded):
		writeAPIError(w, http.StatusConflict, "INVITATION_LIMIT_REACHED", "Einladungscode kann nicht mehr eingelost werden.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
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

func emptyToNil(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
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

func (a *App) handlePublicCORS(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, routeType string) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}

	allowedOrigin, ok := a.resolveAllowedPublicOrigin(r, tenantItem, origin, routeType)
	if !ok {
		writeAPIError(w, http.StatusForbidden, "CORS_ORIGIN_NOT_ALLOWED", "Diese Origin ist fuer die Einbettung nicht freigegeben.")
		return true
	}

	methods := publicRouteMethods(routeType)
	headers := w.Header()
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
	headers.Set("Access-Control-Allow-Origin", allowedOrigin)
	headers.Set("Access-Control-Allow-Methods", strings.Join(methods, ", "))
	headers.Set("Access-Control-Allow-Headers", "Accept, Content-Type")
	headers.Set("Access-Control-Max-Age", "600")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func (a *App) resolveAllowedPublicOrigin(r *http.Request, tenantItem tenant.Tenant, rawOrigin string, routeType string) (string, bool) {
	normalizedOrigin, ok := normalizeOrigin(rawOrigin)
	if !ok {
		return "", false
	}

	if publicRouteSupportsUniversalCORS(routeType) {
		return "*", true
	}

	allowedOrigins := map[string]struct{}{}
	if tenantOrigin, ok := normalizeOrigin(tenantItem.PublicBaseURL); ok {
		allowedOrigins[tenantOrigin] = struct{}{}
	}

	hasExplicitEmbedOrigins := false
	if a.tenantRepo != nil {
		settings, err := a.tenantRepo.GetSettings(r.Context(), tenantItem.ID)
		if err == nil {
			appSettings, _, err := parseAdminTenantAppSettings(settings.SettingsJSON)
			if err == nil {
				hasExplicitEmbedOrigins = len(appSettings.AllowedEmbedOrigins) > 0
				for _, origin := range appSettings.AllowedEmbedOrigins {
					if strings.TrimSpace(origin) == "*" {
						return "*", true
					}
					if normalized, ok := normalizeOrigin(origin); ok {
						allowedOrigins[normalized] = struct{}{}
					}
				}
			}
		}
	}

	if !hasExplicitEmbedOrigins {
		_, exists := allowedOrigins[normalizedOrigin]
		return normalizedOrigin, exists
	}

	_, exists := allowedOrigins[normalizedOrigin]
	return normalizedOrigin, exists
}

func publicRouteSupportsUniversalCORS(routeType string) bool {
	switch routeType {
	case "events_list", "event_detail", "series_list", "series_events", "snippet_events", "registrations_start", "registrations_verify", "registrations_calendar":
		return true
	default:
		return false
	}
}

func publicRouteMethods(routeType string) []string {
	switch routeType {
	case "events_list", "event_detail", "series_list", "series_events", "snippet_events", "registrations_calendar", "certificates_verify_public":
		return []string{http.MethodGet, http.MethodOptions}
	case "registrations_start", "payments_paypal_create_order":
		return []string{http.MethodPost, http.MethodOptions}
	case "registrations_verify":
		return []string{http.MethodGet, http.MethodPost, http.MethodOptions}
	case "invitations_resolve", "participants_portal_me", "participants_portal_registrations", "participants_portal_certificates", "participants_portal_certificate", "participants_portal_certificate_download":
		return []string{http.MethodGet, http.MethodOptions}
	case "participants_portal_request", "participants_portal_verify", "participants_portal_logout", "participants_portal_registration_cancel":
		return []string{http.MethodPost, http.MethodOptions}
	default:
		return []string{http.MethodGet, http.MethodPost, http.MethodOptions}
	}
}

func normalizeOrigin(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", false
	}
	return fmt.Sprintf("%s://%s", strings.ToLower(parsed.Scheme), strings.ToLower(parsed.Host)), true
}

func publicEventPayload(item event.PublicEvent) map[string]any {
	var endsAt any
	if item.EndsAt != nil {
		endsAt = item.EndsAt.UTC().Format(time.RFC3339)
	}
	var publishedAt any
	if item.PublishedAt != nil {
		publishedAt = item.PublishedAt.UTC().Format(time.RFC3339)
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
		"is_published":         item.IsPublished(),
		"publication_state":    item.PublicationState(),
		"published_at":         publishedAt,
		"registration_enabled": item.RegistrationEnabled,
		"waitlist_enabled":     item.WaitlistEnabled,
		"max_participants":     maxParticipants,
		"change_note":          item.ChangeNote,
		"cancelled_reason":     item.CancelledReason,
	}
}
