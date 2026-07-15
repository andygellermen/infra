package httpapp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/auth"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/registration"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

const participantSessionCookieName = "eep_participant_session"

func (a *App) handlePublicParticipantPortalRequest(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Authentifizierung ist nicht verfuegbar.")
		return
	}

	var request struct {
		Email        string `json:"email"`
		RedirectPath string `json:"redirect_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	_, err := a.authService.RequestMagicLink(r.Context(), auth.RequestMagicLinkInput{
		TenantSlug:   tenantItem.Slug,
		Email:        request.Email,
		Purpose:      auth.PurposeParticipantLogin,
		RedirectPath: request.RedirectPath,
		RequestIP:    clientIP(r),
		UserAgent:    strings.TrimSpace(r.UserAgent()),
	})
	if err != nil {
		if errors.Is(err, auth.ErrRateLimited) {
			writeAPIError(w, http.StatusTooManyRequests, "RATE_LIMITED", "Zu viele Anfragen. Bitte spaeter erneut versuchen.")
			return
		}
		if errors.Is(err, auth.ErrUnsupportedPurpose) {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Unbekannter Magic-Link-Zweck.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Teilnehmer-Login-Link konnte nicht angefordert werden.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Wenn die E-Mail-Adresse bekannt ist, wurde ein Login-Link versendet.",
	})
}

func (a *App) handlePublicParticipantPortalVerify(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Authentifizierung ist nicht verfuegbar.")
		return
	}

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

	result, err := a.authService.VerifyMagicLink(r.Context(), auth.VerifyMagicLinkInput{
		RawToken:  token,
		RequestIP: clientIP(r),
		UserAgent: strings.TrimSpace(r.UserAgent()),
	})
	if err != nil {
		if errors.Is(err, auth.ErrInvalidMagicLink) {
			writeAPIError(w, http.StatusBadRequest, "INVALID_MAGIC_LINK", "Magic-Link ist ungueltig oder abgelaufen.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Magic-Link konnte nicht verifiziert werden.")
		return
	}
	if result.Purpose != auth.PurposeParticipantLogin || result.TenantID != tenantItem.ID || strings.TrimSpace(result.ParticipantID) == "" || strings.TrimSpace(result.SessionToken) == "" {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Dieser Link ist fuer das Teilnehmer-Portal nicht gueltig.")
		return
	}

	a.setParticipantSessionCookie(w, result.SessionToken, result.SessionExpiresAt)
	if r.Method == http.MethodGet {
		http.Redirect(w, r, result.RedirectPath, http.StatusSeeOther)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"purpose":         result.Purpose,
		"participant_id":  result.ParticipantID,
		"redirect_path":   result.RedirectPath,
		"session_expires": result.SessionExpiresAt.Format(time.RFC3339),
	})
}

func (a *App) handlePublicParticipantPortalMe(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	principal, ok := a.requireParticipantPrincipal(w, r, tenantItem)
	if !ok {
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"tenant": map[string]any{
			"id":   tenantItem.ID,
			"slug": tenantItem.Slug,
		},
		"participant": map[string]any{
			"id":    principal.ParticipantID,
			"email": principal.Email,
			"name":  principal.Name,
		},
		"session_expires_at": principal.SessionExpiresAt.UTC().Format(time.RFC3339),
	})
}

func (a *App) handlePublicParticipantPortalRegistrations(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.regService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Registration-Service ist nicht verfuegbar.")
		return
	}
	principal, ok := a.requireParticipantPrincipal(w, r, tenantItem)
	if !ok {
		return
	}

	items, err := a.regService.ListParticipantRegistrations(r.Context(), tenantItem.ID, principal.ParticipantID)
	if err != nil {
		a.writeParticipantPortalRegistrationError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, participantPortalRegistrationPayload(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handlePublicParticipantPortalRegistrationCancel(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, registrationID string) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.regService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Registration-Service ist nicht verfuegbar.")
		return
	}
	principal, ok := a.requireParticipantPrincipal(w, r, tenantItem)
	if !ok {
		return
	}

	var request struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	updated, err := a.regService.CancelParticipantRegistration(
		r.Context(),
		tenantItem.ID,
		principal.ParticipantID,
		registrationID,
		request.Reason,
	)
	if err != nil {
		a.writeParticipantPortalRegistrationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": participantPortalRegistrationPayload(updated),
	})
}

func (a *App) handlePublicParticipantPortalLogout(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Authentifizierung ist nicht verfuegbar.")
		return
	}

	rawToken := a.participantSessionTokenFromCookie(r)
	if rawToken != "" {
		_, _ = a.authService.RevokeSession(r.Context(), rawToken, clientIP(r), strings.TrimSpace(r.UserAgent()))
	}
	a.clearParticipantSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) requireParticipantPrincipal(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) (auth.ParticipantSessionPrincipal, bool) {
	if a.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Authentifizierung ist nicht verfuegbar.")
		return auth.ParticipantSessionPrincipal{}, false
	}

	rawSessionToken := a.participantSessionTokenFromCookie(r)
	principal, err := a.authService.AuthenticateParticipantSession(r.Context(), rawSessionToken)
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			a.clearParticipantSessionCookie(w)
			writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Nicht angemeldet.")
			return auth.ParticipantSessionPrincipal{}, false
		}
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Session konnte nicht gelesen werden.")
		return auth.ParticipantSessionPrincipal{}, false
	}

	if principal.TenantID != tenantItem.ID || principal.TenantSlug != tenantItem.Slug {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Zugriff verweigert.")
		return auth.ParticipantSessionPrincipal{}, false
	}

	return principal, true
}

func (a *App) participantSessionTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(participantSessionCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func (a *App) setParticipantSessionCookie(w http.ResponseWriter, rawSessionToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     participantSessionCookieName,
		Value:    rawSessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt.UTC(),
	})
}

func (a *App) clearParticipantSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     participantSessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}

func participantPortalRegistrationPayload(item registration.ParticipantPortalRegistration) map[string]any {
	var endsAt any
	if item.EventEndsAt != nil {
		endsAt = item.EventEndsAt.UTC().Format(time.RFC3339)
	}
	var confirmedAt any
	if item.ConfirmedAt != nil {
		confirmedAt = item.ConfirmedAt.UTC().Format(time.RFC3339)
	}
	var cancelledAt any
	if item.CancelledAt != nil {
		cancelledAt = item.CancelledAt.UTC().Format(time.RFC3339)
	}
	var selfCancelDeadline any
	if item.SelfCancelDeadline != nil {
		selfCancelDeadline = item.SelfCancelDeadline.UTC().Format(time.RFC3339)
	}

	return map[string]any{
		"id":                         item.ID,
		"event_id":                   item.EventID,
		"event_slug":                 item.EventSlug,
		"event_title":                item.EventTitle,
		"event_starts_at":            item.EventStartsAt.UTC().Format(time.RFC3339),
		"event_ends_at":              endsAt,
		"event_timezone":             item.EventTimezone,
		"event_location_name":        item.EventLocationName,
		"event_online_url":           item.EventOnlineURL,
		"status":                     item.Status,
		"participation_type":         item.ParticipationType,
		"quantity":                   item.Quantity,
		"source":                     item.Source,
		"payment_status":             emptyToNil(item.PaymentStatus),
		"confirmed_at":               confirmedAt,
		"cancelled_at":               cancelledAt,
		"cancellation_reason":        emptyToNil(item.CancellationReason),
		"self_cancel_allowed":        item.SelfCancelAllowed,
		"self_cancel_deadline_at":    selfCancelDeadline,
		"self_cancel_deadline_hours": item.SelfCancelDeadlineHours,
		"created_at":                 item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":                 item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (a *App) writeParticipantPortalRegistrationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, registration.ErrParticipantAccessDenied):
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Zugriff verweigert.")
	case errors.Is(err, registration.ErrRegistrationCancelDeadlineExceeded):
		writeAPIError(w, http.StatusConflict, "REGISTRATION_CANCEL_DEADLINE_EXCEEDED", "Die Abmeldefrist fuer diese Anmeldung ist bereits abgelaufen.")
	case errors.Is(err, registration.ErrRegistrationCancelNotAllowed):
		writeAPIError(w, http.StatusConflict, "REGISTRATION_STATE_INVALID", "Anmeldung kann in diesem Status nicht storniert werden.")
	case errors.Is(err, registration.ErrRegistrationNotFound):
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Anmeldung nicht gefunden.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}
