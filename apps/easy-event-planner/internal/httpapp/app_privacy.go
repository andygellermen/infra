package httpapp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/privacy"
)

func (a *App) handleAdminRetentionPoliciesCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.privacyService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Privacy-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	items, err := a.privacyService.ListPolicies(r.Context(), principal.TenantID)
	if err != nil {
		a.writePrivacyError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, retentionPolicyPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handleAdminRetentionPolicyItem(w http.ResponseWriter, r *http.Request) {
	policyID, ok := parseAdminRetentionPolicyPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "RETENTION_POLICY_NOT_FOUND", "Retention-Policy nicht gefunden.")
		return
	}

	if r.Method != http.MethodPatch {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.privacyService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Privacy-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		Action        *string `json:"action"`
		RetentionDays *int    `json:"retention_days"`
		Enabled       *bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	updated, err := a.privacyService.UpdatePolicy(r.Context(), principal.TenantID, policyID, privacy.UpdatePolicyParams{
		Action:        request.Action,
		RetentionDays: request.RetentionDays,
		Enabled:       request.Enabled,
	})
	if err != nil {
		a.writePrivacyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": retentionPolicyPayload(updated),
	})
}

func (a *App) handleAdminRetentionJobDryRun(w http.ResponseWriter, r *http.Request) {
	a.handleAdminRetentionJobExecute(w, r, true)
}

func (a *App) handleAdminRetentionJobRun(w http.ResponseWriter, r *http.Request) {
	a.handleAdminRetentionJobExecute(w, r, false)
}

func (a *App) handleAdminRetentionJobExecute(w http.ResponseWriter, r *http.Request, dryRun bool) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.privacyService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Privacy-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	result, err := a.privacyService.Execute(r.Context(), privacy.ExecuteInput{
		TenantID:    principal.TenantID,
		ActorUserID: principal.UserID,
		RequestIP:   clientIP(r),
		UserAgent:   strings.TrimSpace(r.UserAgent()),
		DryRun:      dryRun,
	})
	if err != nil {
		a.writePrivacyError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": retentionJobPayload(result),
	})
}

func (a *App) handleAdminRetentionJobsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.privacyService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Privacy-Service ist nicht verfuegbar.")
		return
	}

	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	limit := 25
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "limit muss eine Zahl sein.")
			return
		}
		limit = parsed
	}

	items, err := a.privacyService.ListJobs(r.Context(), principal.TenantID, limit)
	if err != nil {
		a.writePrivacyError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, retentionJobRecordPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func parseAdminRetentionPolicyPath(path string) (policyID string, ok bool) {
	const prefix = "/api/v1/admin/privacy/retention-policies/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	remainder := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if remainder == "" || strings.Contains(remainder, "/") {
		return "", false
	}
	return remainder, true
}

func retentionPolicyPayload(item privacy.Policy) map[string]any {
	return map[string]any{
		"id":             item.ID,
		"tenant_id":      item.TenantID,
		"data_category":  item.DataCategory,
		"action":         item.Action,
		"retention_days": item.RetentionDays,
		"enabled":        item.Enabled,
		"created_at":     item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":     item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func retentionJobPayload(item privacy.ExecuteResult) map[string]any {
	categories := make([]map[string]any, 0, len(item.Items))
	for _, category := range item.Items {
		categories = append(categories, retentionCategoryResultPayload(category))
	}
	return map[string]any{
		"dry_run":        item.DryRun,
		"started_at":     item.StartedAt.UTC().Format(time.RFC3339),
		"finished_at":    item.FinishedAt.UTC().Format(time.RFC3339),
		"total_affected": item.TotalAffected,
		"total_executed": item.TotalExecuted,
		"audit_id":       item.AuditID,
		"categories":     categories,
	}
}

func retentionJobRecordPayload(item privacy.JobRecord) map[string]any {
	categories := make([]map[string]any, 0, len(item.Items))
	for _, category := range item.Items {
		categories = append(categories, retentionCategoryResultPayload(category))
	}
	return map[string]any{
		"id":             item.ID,
		"tenant_id":      item.TenantID,
		"action":         item.Action,
		"actor_user_id":  nullableJSON(item.ActorUserID),
		"dry_run":        item.DryRun,
		"total_affected": item.TotalAffected,
		"total_executed": item.TotalExecuted,
		"created_at":     item.CreatedAt.UTC().Format(time.RFC3339),
		"categories":     categories,
	}
}

func retentionCategoryResultPayload(item privacy.CategoryResult) map[string]any {
	payload := map[string]any{
		"policy_id":      item.PolicyID,
		"data_category":  item.DataCategory,
		"action":         item.Action,
		"retention_days": item.RetentionDays,
		"enabled":        item.Enabled,
		"cutoff_at":      item.CutoffAt.UTC().Format(time.RFC3339),
		"affected":       item.Affected,
		"executed":       item.Executed,
	}
	if strings.TrimSpace(item.Note) != "" {
		payload["note"] = item.Note
	}
	return payload
}

func (a *App) writePrivacyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, privacy.ErrPolicyNotFound):
		writeAPIError(w, http.StatusNotFound, "RETENTION_POLICY_NOT_FOUND", "Retention-Policy nicht gefunden.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}

func nullableJSON(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
