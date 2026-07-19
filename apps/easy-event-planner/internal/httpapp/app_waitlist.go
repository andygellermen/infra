package httpapp

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/registration"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handleAdminEventWaitlistList(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if !a.requireTenantFeatureForPrincipal(w, r, principal.TenantID, tenant.FeatureWaitlist, "FEATURE_DISABLED", "Die Warteliste ist fuer diesen Mandanten nicht freigeschaltet.") {
		return
	}

	items, err := a.regService.ListWaitlistEntries(r.Context(), principal.TenantID, eventID)
	if err != nil {
		a.writeWaitlistError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, waitlistEntryPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handleAdminWaitlistItem(w http.ResponseWriter, r *http.Request) {
	waitlistEntryID, action, ok := parseAdminWaitlistPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "WAITLIST_ENTRY_NOT_FOUND", "Wartelisteneintrag nicht gefunden.")
		return
	}
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}

	switch action {
	case "offer":
		a.handleAdminWaitlistOffer(w, r, waitlistEntryID)
	case "promote":
		a.handleAdminWaitlistPromote(w, r, waitlistEntryID)
	default:
		writeAPIError(w, http.StatusNotFound, "WAITLIST_ENTRY_NOT_FOUND", "Wartelisteneintrag nicht gefunden.")
	}
}

func (a *App) handleAdminWaitlistOffer(w http.ResponseWriter, r *http.Request, waitlistEntryID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if !a.requireTenantFeatureForPrincipal(w, r, principal.TenantID, tenant.FeatureWaitlist, "FEATURE_DISABLED", "Die Warteliste ist fuer diesen Mandanten nicht freigeschaltet.") {
		return
	}

	item, err := a.regService.OfferWaitlistEntry(
		r.Context(),
		principal.TenantID,
		waitlistEntryID,
		clientIP(r),
		strings.TrimSpace(r.UserAgent()),
	)
	if err != nil {
		a.writeWaitlistError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": waitlistEntryPayload(item),
	})
}

func (a *App) handleAdminWaitlistPromote(w http.ResponseWriter, r *http.Request, waitlistEntryID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if !a.requireTenantFeatureForPrincipal(w, r, principal.TenantID, tenant.FeatureWaitlist, "FEATURE_DISABLED", "Die Warteliste ist fuer diesen Mandanten nicht freigeschaltet.") {
		return
	}

	item, err := a.regService.PromoteWaitlistEntry(r.Context(), principal.TenantID, waitlistEntryID)
	if err != nil {
		a.writeWaitlistError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": waitlistEntryPayload(item),
	})
}

func parseAdminWaitlistPath(path string) (waitlistEntryID, action string, ok bool) {
	const prefix = "/api/v1/admin/waitlist/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	remainder := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if remainder == "" {
		return "", "", false
	}
	parts := strings.Split(remainder, "/")
	if len(parts) != 2 {
		return "", "", false
	}
	waitlistEntryID = strings.TrimSpace(parts[0])
	action = strings.TrimSpace(parts[1])
	if waitlistEntryID == "" || action == "" {
		return "", "", false
	}
	return waitlistEntryID, action, true
}

func waitlistEntryPayload(item registration.WaitlistEntry) map[string]any {
	var offeredAt any
	if item.OfferedAt != nil {
		offeredAt = item.OfferedAt.UTC().Format(time.RFC3339)
	}
	var offerExpiresAt any
	if item.OfferExpiresAt != nil {
		offerExpiresAt = item.OfferExpiresAt.UTC().Format(time.RFC3339)
	}
	var acceptedAt any
	if item.AcceptedAt != nil {
		acceptedAt = item.AcceptedAt.UTC().Format(time.RFC3339)
	}

	return map[string]any{
		"id":                  item.ID,
		"event_id":            item.EventID,
		"registration_id":     item.RegistrationID,
		"participant_id":      item.ParticipantID,
		"participant_name":    item.ParticipantName,
		"participant_email":   item.ParticipantEmail,
		"position":            item.Position,
		"status":              item.Status,
		"registration_status": item.RegistrationStatus,
		"offered_at":          offeredAt,
		"offer_expires_at":    offerExpiresAt,
		"accepted_at":         acceptedAt,
		"created_at":          item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":          item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (a *App) writeWaitlistError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, registration.ErrEventNotFound):
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
	case errors.Is(err, registration.ErrWaitlistEntryNotFound):
		writeAPIError(w, http.StatusNotFound, "WAITLIST_ENTRY_NOT_FOUND", "Wartelisteneintrag nicht gefunden.")
	case errors.Is(err, registration.ErrWaitlistStateInvalid), errors.Is(err, registration.ErrRegistrationState):
		writeAPIError(w, http.StatusConflict, "WAITLIST_STATE_INVALID", "Wartelisteneintrag kann in diesem Status nicht verarbeitet werden.")
	case errors.Is(err, registration.ErrEventFull):
		writeAPIError(w, http.StatusConflict, "EVENT_FULL", "Die Veranstaltung ist ausgebucht. Eine Warteliste ist verfuegbar.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}
