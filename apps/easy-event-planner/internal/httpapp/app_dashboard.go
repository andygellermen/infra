package httpapp

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/registration"
)

func (a *App) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}

	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	dashboard, err := a.regService.GetDashboard(r.Context(), principal.TenantID)
	if err != nil {
		a.writeRegistrationError(w, err)
		return
	}

	today := make([]map[string]any, 0, len(dashboard.Today))
	for _, item := range dashboard.Today {
		today = append(today, dashboardEventPayload(item))
	}
	nextEvents := make([]map[string]any, 0, len(dashboard.NextEvents))
	for _, item := range dashboard.NextEvents {
		nextEvents = append(nextEvents, dashboardEventPayload(item))
	}

	var lastRetentionRunAt any
	if dashboard.Stats.LastRetentionRunAt != nil {
		lastRetentionRunAt = dashboard.Stats.LastRetentionRunAt.UTC().Format(time.RFC3339)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"stats": map[string]any{
			"today_events":            dashboard.Stats.TodayEvents,
			"upcoming_events":         dashboard.Stats.UpcomingEvents,
			"confirmed_participants":  dashboard.Stats.ConfirmedParticipants,
			"waitlist_entries":        dashboard.Stats.WaitlistEntries,
			"free_seats":              dashboard.Stats.FreeSeats,
			"open_email_jobs":         dashboard.Stats.OpenEmailJobs,
			"last_retention_run_at":   lastRetentionRunAt,
			"today_events_total":      len(today),
			"next_events_total":       len(nextEvents),
			"dashboard_generated_utc": time.Now().UTC().Format(time.RFC3339),
		},
		"today_events": today,
		"next_events":  nextEvents,
	})
}

func (a *App) handleAdminEventRegistrationList(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	items, err := a.regService.ListEventRegistrations(r.Context(), principal.TenantID, eventID)
	if err != nil {
		a.writeRegistrationError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, adminRegistrationPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handleAdminRegistrationItem(w http.ResponseWriter, r *http.Request) {
	registrationID, action, ok := parseAdminRegistrationPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Teilnahme nicht gefunden.")
		return
	}
	if action != "" {
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Teilnahme nicht gefunden.")
		return
	}
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	a.handleAdminRegistrationGet(w, r, registrationID)
}

func (a *App) handleAdminRegistrationGet(w http.ResponseWriter, r *http.Request, registrationID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	item, err := a.regService.GetRegistration(r.Context(), principal.TenantID, registrationID)
	if err != nil {
		a.writeRegistrationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": adminRegistrationPayload(item),
	})
}

func parseAdminRegistrationPath(path string) (registrationID, action string, ok bool) {
	const prefix = "/api/v1/admin/registrations/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	remainder := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if remainder == "" {
		return "", "", false
	}

	parts := strings.Split(remainder, "/")
	if len(parts) == 1 {
		id := strings.TrimSpace(parts[0])
		if id == "" {
			return "", "", false
		}
		return id, "", true
	}
	if len(parts) == 2 {
		id := strings.TrimSpace(parts[0])
		action := strings.TrimSpace(parts[1])
		if id == "" || action == "" {
			return "", "", false
		}
		return id, action, true
	}
	return "", "", false
}

func adminRegistrationPayload(item registration.AdminRegistration) map[string]any {
	var reservedUntil any
	if item.ReservedUntil != nil {
		reservedUntil = item.ReservedUntil.UTC().Format(time.RFC3339)
	}
	var confirmedAt any
	if item.ConfirmedAt != nil {
		confirmedAt = item.ConfirmedAt.UTC().Format(time.RFC3339)
	}
	var cancelledAt any
	if item.CancelledAt != nil {
		cancelledAt = item.CancelledAt.UTC().Format(time.RFC3339)
	}
	var attendedAt any
	if item.AttendedAt != nil {
		attendedAt = item.AttendedAt.UTC().Format(time.RFC3339)
	}
	var privacyAcceptedAt any
	if item.PrivacyAcceptedAt != nil {
		privacyAcceptedAt = item.PrivacyAcceptedAt.UTC().Format(time.RFC3339)
	}
	var paymentStatus any
	if strings.TrimSpace(item.PaymentStatus) != "" {
		paymentStatus = item.PaymentStatus
	}

	return map[string]any{
		"id":                  item.ID,
		"tenant_id":           item.TenantID,
		"event_id":            item.EventID,
		"event_slug":          item.EventSlug,
		"event_title":         item.EventTitle,
		"participant_id":      item.ParticipantID,
		"participant_name":    item.ParticipantName,
		"participant_email":   item.ParticipantEmail,
		"participant_phone":   item.ParticipantPhone,
		"status":              item.Status,
		"participation_type":  item.ParticipationType,
		"quantity":            item.Quantity,
		"payment_status":      paymentStatus,
		"source":              item.Source,
		"cancellation_reason": item.CancellationNote,
		"reserved_until":      reservedUntil,
		"confirmed_at":        confirmedAt,
		"cancelled_at":        cancelledAt,
		"attended_at":         attendedAt,
		"privacy_accepted_at": privacyAcceptedAt,
		"created_at":          item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":          item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func dashboardEventPayload(item registration.DashboardEvent) map[string]any {
	var maxParticipants any
	if item.MaxParticipants != nil {
		maxParticipants = *item.MaxParticipants
	}
	var freeSeats any
	if item.FreeSeats != nil {
		freeSeats = *item.FreeSeats
	}

	return map[string]any{
		"id":                     item.ID,
		"slug":                   item.Slug,
		"title":                  item.Title,
		"status":                 item.Status,
		"starts_at":              item.StartsAt.UTC().Format(time.RFC3339),
		"location_name":          item.LocationName,
		"max_participants":       maxParticipants,
		"confirmed_participants": item.ConfirmedParticipants,
		"waitlist_entries":       item.WaitlistEntries,
		"occupied_seats":         item.OccupiedSeats,
		"free_seats":             freeSeats,
	}
}

func (a *App) writeRegistrationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, registration.ErrEventNotFound):
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
	case errors.Is(err, registration.ErrRegistrationNotFound):
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Teilnahme nicht gefunden.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}
