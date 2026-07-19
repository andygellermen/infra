package httpapp

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handleAdminUsersCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleAdminUsersList(w, r)
	case http.MethodPost:
		a.handleAdminUsersCreate(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminUsersItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := parseAdminUserPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "TENANT_USER_NOT_FOUND", "Benutzer nicht gefunden.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.handleAdminUserGet(w, r, userID)
	case http.MethodPatch:
		a.handleAdminUserPatch(w, r, userID)
	case http.MethodDelete:
		a.handleAdminUserDelete(w, r, userID)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminUsersList(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	items, err := a.tenantRepo.ListUsers(r.Context(), principal.TenantID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, tenantUserPayload(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handleAdminUsersCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		Email  string `json:"email"`
		Name   string `json:"name"`
		Role   string `json:"role"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	created, err := a.tenantRepo.CreateUser(r.Context(), principal.TenantID, tenant.CreateTenantUserParams{
		Email:  request.Email,
		Name:   request.Name,
		Role:   request.Role,
		Status: request.Status,
	})
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"item": tenantUserPayload(created),
	})
}

func (a *App) handleAdminUserGet(w http.ResponseWriter, r *http.Request, userID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	item, err := a.tenantRepo.GetUserByID(r.Context(), principal.TenantID, userID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": tenantUserPayload(item),
	})
}

func (a *App) handleAdminUserPatch(w http.ResponseWriter, r *http.Request, userID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		Email  *string `json:"email"`
		Name   *string `json:"name"`
		Role   *string `json:"role"`
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	updated, err := a.tenantRepo.UpdateUser(r.Context(), principal.TenantID, userID, tenant.UpdateTenantUserParams{
		Email:  request.Email,
		Name:   request.Name,
		Role:   request.Role,
		Status: request.Status,
	})
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": tenantUserPayload(updated),
	})
}

func (a *App) handleAdminUserDelete(w http.ResponseWriter, r *http.Request, userID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	deleted, err := a.tenantRepo.DeleteUser(r.Context(), principal.TenantID, userID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	if !deleted {
		writeAPIError(w, http.StatusNotFound, "TENANT_USER_NOT_FOUND", "Benutzer nicht gefunden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseAdminUserPath(path string) (string, bool) {
	const prefix = "/api/v1/admin/users/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	value := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if value == "" || strings.Contains(value, "/") {
		return "", false
	}
	return value, true
}

func tenantUserPayload(item tenant.TenantUser) map[string]any {
	return map[string]any{
		"id":         item.ID,
		"tenant_id":  item.TenantID,
		"email":      item.Email,
		"name":       item.Name,
		"role":       item.Role,
		"status":     item.Status,
		"created_at": item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at": item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
