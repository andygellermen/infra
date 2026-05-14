package httpapp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/invitation"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handleAdminInvitationsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleAdminInvitationsList(w, r)
	case http.MethodPost:
		a.handleAdminInvitationsCreate(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminInvitationsItem(w http.ResponseWriter, r *http.Request) {
	invitationID, ok := parseAdminInvitationPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "Einladung nicht gefunden.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.handleAdminInvitationsGet(w, r, invitationID)
	case http.MethodPatch:
		a.handleAdminInvitationsPatch(w, r, invitationID)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminInvitationsList(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.invitationService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Invitation-Service ist nicht verfuegbar.")
		return
	}

	items, err := a.invitationService.ListLinks(r.Context(), principal.TenantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Einladungen konnten nicht geladen werden.")
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, invitationPayload(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handleAdminInvitationsCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.invitationService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Invitation-Service ist nicht verfuegbar.")
		return
	}

	var request struct {
		EventID         string `json:"event_id"`
		SeriesID        string `json:"series_id"`
		Code            string `json:"code"`
		Label           string `json:"label"`
		InviteType      string `json:"invite_type"`
		DiscountType    string `json:"discount_type"`
		DiscountValue   *int   `json:"discount_value"`
		MaxUses         *int   `json:"max_uses"`
		MaxUsesPerEmail *int   `json:"max_uses_per_email"`
		StartsAt        string `json:"starts_at"`
		ExpiresAt       string `json:"expires_at"`
		IsShareable     *bool  `json:"is_shareable"`
		Status          string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	startsAt, err := parseOptionalTimeRFC3339(request.StartsAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "starts_at muss RFC3339 sein.")
		return
	}
	expiresAt, err := parseOptionalTimeRFC3339(request.ExpiresAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "expires_at muss RFC3339 sein.")
		return
	}

	created, err := a.invitationService.CreateLink(r.Context(), principal.TenantID, invitation.CreateLinkInput{
		EventID:         request.EventID,
		SeriesID:        request.SeriesID,
		Code:            request.Code,
		Label:           request.Label,
		InviteType:      request.InviteType,
		DiscountType:    request.DiscountType,
		DiscountValue:   request.DiscountValue,
		MaxUses:         request.MaxUses,
		MaxUsesPerEmail: request.MaxUsesPerEmail,
		StartsAt:        startsAt,
		ExpiresAt:       expiresAt,
		IsShareable:     request.IsShareable,
		Status:          request.Status,
	})
	if err != nil {
		a.writeInvitationError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"item": invitationPayload(created),
	})
}

func (a *App) handleAdminInvitationsGet(w http.ResponseWriter, r *http.Request, invitationID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.invitationService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Invitation-Service ist nicht verfuegbar.")
		return
	}

	item, err := a.invitationService.GetLinkByID(r.Context(), principal.TenantID, invitationID)
	if err != nil {
		a.writeInvitationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": invitationPayload(item)})
}

func (a *App) handleAdminInvitationsPatch(w http.ResponseWriter, r *http.Request, invitationID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.invitationService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Invitation-Service ist nicht verfuegbar.")
		return
	}

	var request struct {
		EventID         *string `json:"event_id"`
		SeriesID        *string `json:"series_id"`
		Code            *string `json:"code"`
		Label           *string `json:"label"`
		InviteType      *string `json:"invite_type"`
		DiscountType    *string `json:"discount_type"`
		DiscountValue   *int    `json:"discount_value"`
		MaxUses         *int    `json:"max_uses"`
		MaxUsesPerEmail *int    `json:"max_uses_per_email"`
		StartsAt        *string `json:"starts_at"`
		ExpiresAt       *string `json:"expires_at"`
		IsShareable     *bool   `json:"is_shareable"`
		Status          *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	startsAt, err := parseOptionalTimeRFC3339Pointer(request.StartsAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "starts_at muss RFC3339 sein.")
		return
	}
	expiresAt, err := parseOptionalTimeRFC3339Pointer(request.ExpiresAt)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "expires_at muss RFC3339 sein.")
		return
	}

	updated, err := a.invitationService.UpdateLink(r.Context(), principal.TenantID, invitationID, invitation.UpdateLinkInput{
		EventID:         request.EventID,
		SeriesID:        request.SeriesID,
		Code:            request.Code,
		Label:           request.Label,
		InviteType:      request.InviteType,
		DiscountType:    request.DiscountType,
		DiscountValue:   request.DiscountValue,
		MaxUses:         request.MaxUses,
		MaxUsesPerEmail: request.MaxUsesPerEmail,
		StartsAt:        startsAt,
		ExpiresAt:       expiresAt,
		IsShareable:     request.IsShareable,
		Status:          request.Status,
	})
	if err != nil {
		a.writeInvitationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": invitationPayload(updated)})
}

func (a *App) handlePublicInvitationResolve(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.invitationService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Invitation-Service ist nicht verfuegbar.")
		return
	}

	var request struct {
		EventID          string `json:"event_id"`
		InviteCode       string `json:"invite_code"`
		ParticipantEmail string `json:"participant_email"`
		BaseAmountCents  int    `json:"base_amount_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	result, err := a.invitationService.ResolveCode(r.Context(), invitation.ResolveInput{
		TenantID:         tenantItem.ID,
		EventID:          request.EventID,
		ParticipantEmail: request.ParticipantEmail,
		Code:             request.InviteCode,
		BaseAmountCents:  request.BaseAmountCents,
	})
	if err != nil {
		a.writeInvitationError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
		"invite": map[string]any{
			"item":                  invitationPayload(result.Link),
			"base_amount_cents":     result.BaseAmountCents,
			"discount_amount_cents": result.DiscountAmountCents,
			"credit_amount_cents":   result.CreditAmountCents,
			"final_amount_cents":    result.FinalAmountCents,
			"sponsored":             result.Sponsored,
		},
	})
}

func parseAdminInvitationPath(path string) (invitationID string, ok bool) {
	const prefix = "/api/v1/admin/invitations/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	invitationID = strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if invitationID == "" || strings.Contains(invitationID, "/") {
		return "", false
	}
	return invitationID, true
}

func invitationPayload(item invitation.Link) map[string]any {
	var startsAt any
	if item.StartsAt != nil {
		startsAt = item.StartsAt.UTC().Format(time.RFC3339)
	}
	var expiresAt any
	if item.ExpiresAt != nil {
		expiresAt = item.ExpiresAt.UTC().Format(time.RFC3339)
	}
	var discountValue any
	if item.DiscountValue != nil {
		discountValue = *item.DiscountValue
	}
	var maxUses any
	if item.MaxUses != nil {
		maxUses = *item.MaxUses
	}
	var maxUsesPerEmail any
	if item.MaxUsesPerEmail != nil {
		maxUsesPerEmail = *item.MaxUsesPerEmail
	}

	return map[string]any{
		"id":                      item.ID,
		"tenant_id":               item.TenantID,
		"event_id":                emptyToNil(item.EventID),
		"series_id":               emptyToNil(item.SeriesID),
		"code":                    item.Code,
		"label":                   emptyToNil(item.Label),
		"invite_type":             item.InviteType,
		"discount_type":           emptyToNil(item.DiscountType),
		"discount_type_effective": item.ComputedDiscount,
		"discount_value":          discountValue,
		"max_uses":                maxUses,
		"used_count":              item.UsedCount,
		"max_uses_per_email":      maxUsesPerEmail,
		"starts_at":               startsAt,
		"expires_at":              expiresAt,
		"is_shareable":            item.IsShareable,
		"status":                  item.Status,
		"created_at":              item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":              item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (a *App) writeInvitationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, invitation.ErrInvitationNotFound):
		writeAPIError(w, http.StatusNotFound, "INVITATION_NOT_FOUND", "Einladung nicht gefunden.")
	case errors.Is(err, invitation.ErrEventNotFound):
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
	case errors.Is(err, invitation.ErrInvitationStatusInvalid):
		writeAPIError(w, http.StatusConflict, "INVITATION_INACTIVE", "Einladung ist nicht aktiv.")
	case errors.Is(err, invitation.ErrInvitationNotStarted):
		writeAPIError(w, http.StatusConflict, "INVITATION_NOT_STARTED", "Einladung ist noch nicht aktiv.")
	case errors.Is(err, invitation.ErrInvitationExpired):
		writeAPIError(w, http.StatusConflict, "INVITATION_EXPIRED", "Einladung ist abgelaufen.")
	case errors.Is(err, invitation.ErrInvitationScopeMismatch):
		writeAPIError(w, http.StatusConflict, "INVITATION_SCOPE_MISMATCH", "Einladung passt nicht zu dieser Veranstaltung.")
	case errors.Is(err, invitation.ErrInvitationUsageExceeded), errors.Is(err, invitation.ErrInvitationEmailExceeded):
		writeAPIError(w, http.StatusConflict, "INVITATION_LIMIT_REACHED", "Einladung kann nicht mehr eingelost werden.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}

func parseOptionalTimeRFC3339(raw string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func parseOptionalTimeRFC3339Pointer(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	return parseOptionalTimeRFC3339(*raw)
}
