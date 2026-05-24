package httpapp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/certificate"
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

func (a *App) handleAdminEventRegistrationManualCreate(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		Name              string `json:"name"`
		Email             string `json:"email"`
		Phone             string `json:"phone"`
		ParticipationType string `json:"participation_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	item, err := a.regService.CreateManualRegistration(r.Context(), registration.ManualRegistrationInput{
		TenantID:          principal.TenantID,
		EventID:           eventID,
		Name:              request.Name,
		Email:             request.Email,
		Phone:             request.Phone,
		ParticipationType: request.ParticipationType,
	})
	if err != nil {
		a.writeRegistrationError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"item": adminRegistrationPayload(item),
	})
}

func (a *App) handleAdminRegistrationItem(w http.ResponseWriter, r *http.Request) {
	registrationID, action, ok := parseAdminRegistrationPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Teilnahme nicht gefunden.")
		return
	}
	if action == "" {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handleAdminRegistrationGet(w, r, registrationID)
		return
	}

	switch action {
	case "mark-attended":
		if r.Method != http.MethodPost {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handleAdminRegistrationMarkAttended(w, r, registrationID)
	case "issue-certificate":
		if r.Method != http.MethodPost {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handleAdminRegistrationIssueCertificate(w, r, registrationID)
	case "certificate":
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handleAdminRegistrationCertificateGet(w, r, registrationID)
	default:
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Teilnahme nicht gefunden.")
	}
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

func (a *App) handleAdminRegistrationMarkAttended(w http.ResponseWriter, r *http.Request, registrationID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	item, err := a.regService.MarkRegistrationAttended(r.Context(), principal.TenantID, registrationID)
	if err != nil {
		a.writeRegistrationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": adminRegistrationPayload(item),
	})
}

func (a *App) handleAdminRegistrationIssueCertificate(w http.ResponseWriter, r *http.Request, registrationID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.certificateService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Zertifikats-Service ist nicht verfuegbar.")
		return
	}

	result, err := a.certificateService.IssueForRegistration(r.Context(), principal.TenantID, registrationID)
	if err != nil {
		a.writeCertificateError(w, err)
		return
	}

	payload := certificatePayload(result.Certificate)
	if strings.TrimSpace(result.VerificationCode) != "" {
		payload["verification_url"] = a.certificateService.VerificationURL(
			result.Certificate.TenantSlug,
			result.Certificate.CertificateNumber,
			result.VerificationCode,
		)
	}

	statusCode := http.StatusCreated
	if result.AlreadyIssued {
		statusCode = http.StatusOK
	}
	writeJSON(w, statusCode, map[string]any{
		"item":           payload,
		"already_issued": result.AlreadyIssued,
	})
}

func (a *App) handleAdminRegistrationCertificateGet(w http.ResponseWriter, r *http.Request, registrationID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.certificateService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Zertifikats-Service ist nicht verfuegbar.")
		return
	}

	item, err := a.certificateService.GetByRegistration(r.Context(), principal.TenantID, registrationID)
	if err != nil {
		a.writeCertificateError(w, err)
		return
	}

	payload := certificatePayload(item)
	payload["verify_path"] = "/api/v1/public/" + url.PathEscape(item.TenantSlug) + "/certificates/verify"
	writeJSON(w, http.StatusOK, map[string]any{
		"item": payload,
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

func certificatePayload(item certificate.Certificate) map[string]any {
	var attendedAt any
	if item.AttendedAt != nil {
		attendedAt = item.AttendedAt.UTC().Format(time.RFC3339)
	}
	var revokedAt any
	if item.RevokedAt != nil {
		revokedAt = item.RevokedAt.UTC().Format(time.RFC3339)
	}
	return map[string]any{
		"id":                     item.ID,
		"tenant_id":              item.TenantID,
		"tenant_slug":            item.TenantSlug,
		"registration_id":        item.RegistrationID,
		"participant_id":         item.ParticipantID,
		"participant_name":       item.ParticipantName,
		"participant_email":      item.ParticipantEmail,
		"event_id":               item.EventID,
		"event_slug":             item.EventSlug,
		"event_title":            item.EventTitle,
		"event_starts_at":        item.EventStartsAt.UTC().Format(time.RFC3339),
		"event_timezone":         item.EventTimezone,
		"certificate_number":     item.CertificateNumber,
		"status":                 item.Status,
		"verification_code_hint": emptyToNil(item.VerificationCodeHint),
		"issued_at":              item.IssuedAt.UTC().Format(time.RFC3339),
		"attended_at":            attendedAt,
		"revoked_at":             revokedAt,
		"file_sha256":            item.FileSHA256,
	}
}

func (a *App) writeRegistrationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, registration.ErrEventNotFound):
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
	case errors.Is(err, registration.ErrRegistrationDisabled), errors.Is(err, registration.ErrRegistrationClosed):
		writeAPIError(w, http.StatusConflict, "REGISTRATION_CLOSED", "Anmeldung ist fuer diese Veranstaltung derzeit nicht moeglich.")
	case errors.Is(err, registration.ErrAlreadyRegistered):
		writeAPIError(w, http.StatusConflict, "ALREADY_REGISTERED", "Teilnahme ist bereits bestaetigt.")
	case errors.Is(err, registration.ErrAlreadyWaitlisted):
		writeAPIError(w, http.StatusConflict, "ALREADY_WAITLISTED", "Teilnahme steht bereits auf der Warteliste.")
	case errors.Is(err, registration.ErrEventFull):
		writeAPIError(w, http.StatusConflict, "EVENT_FULL", "Die Veranstaltung ist ausgebucht. Eine Warteliste ist verfuegbar.")
	case errors.Is(err, registration.ErrRegistrationNotFound):
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Teilnahme nicht gefunden.")
	case errors.Is(err, registration.ErrRegistrationAttendNotAllowed):
		writeAPIError(w, http.StatusConflict, "REGISTRATION_STATE_INVALID", "Teilnahme kann nicht als anwesend markiert werden.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}

func (a *App) writeCertificateError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, certificate.ErrRegistrationNotFound):
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Teilnahme nicht gefunden.")
	case errors.Is(err, certificate.ErrCertificateNotFound):
		writeAPIError(w, http.StatusNotFound, "CERTIFICATE_NOT_FOUND", "Zertifikat nicht gefunden.")
	case errors.Is(err, certificate.ErrCertificateEligibility):
		writeAPIError(w, http.StatusConflict, "CERTIFICATE_NOT_ELIGIBLE", "Teilnahme ist nicht fuer ein Zertifikat freigegeben.")
	case errors.Is(err, certificate.ErrCertificateAccessDenied):
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Zugriff verweigert.")
	case errors.Is(err, certificate.ErrInvalidVerificationCode):
		writeAPIError(w, http.StatusBadRequest, "INVALID_CERTIFICATE_CODE", "Zertifikat konnte nicht verifiziert werden.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}
