package httpapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

const (
	defaultEventTimeStart       = "08:00"
	defaultEventTimeEnd         = "22:00"
	defaultEventTimeStepMinutes = 15
	defaultEventSlugMode        = "optional"
)

type adminTenantAppSettings struct {
	EventTimeStart      string   `json:"event_time_start"`
	EventTimeEnd        string   `json:"event_time_end"`
	EventTimeStepMinute int      `json:"event_time_step_minutes"`
	EventSlugMode       string   `json:"event_slug_mode"`
	AllowedEmbedOrigins []string `json:"allowed_embed_origins"`
	EventDetailBaseURL  string   `json:"event_detail_base_url"`
}

func (a *App) handleAdminTenantItem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleAdminTenantGet(w, r)
	case http.MethodPatch:
		a.handleAdminTenantPatch(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminTenantSettingsItem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleAdminTenantSettingsGet(w, r)
	case http.MethodPatch:
		a.handleAdminTenantSettingsPatch(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminTenantGet(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	item, err := a.tenantRepo.GetByID(r.Context(), principal.TenantID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": tenantPayload(item),
	})
}

func (a *App) handleAdminTenantPatch(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	var request struct {
		Name            *string `json:"name"`
		PublicBaseURL   *string `json:"public_base_url"`
		DefaultTimezone *string `json:"default_timezone"`
		DefaultLocale   *string `json:"default_locale"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	item, err := a.tenantRepo.UpdateTenant(r.Context(), principal.TenantID, tenant.UpdateTenantParams{
		Name:            request.Name,
		PublicBaseURL:   request.PublicBaseURL,
		DefaultTimezone: request.DefaultTimezone,
		DefaultLocale:   request.DefaultLocale,
	})
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": tenantPayload(item),
	})
}

func (a *App) handleAdminTenantSettingsGet(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	settings, err := a.tenantRepo.GetSettings(r.Context(), principal.TenantID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	payload, err := tenantSettingsPayload(settings)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Tenant-Settings konnten nicht gelesen werden.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": payload,
	})
}

func (a *App) handleAdminTenantSettingsPatch(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Tenant-Service ist nicht verfuegbar.")
		return
	}

	current, err := a.tenantRepo.GetSettings(r.Context(), principal.TenantID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	var request struct {
		SenderEmail          *string                 `json:"sender_email"`
		SenderName           *string                 `json:"sender_name"`
		PayPalMode           *string                 `json:"paypal_mode"`
		PayPalClientID       *string                 `json:"paypal_client_id"`
		PayPalMerchantID     *string                 `json:"paypal_merchant_id"`
		DefaultRetentionDays *int                    `json:"default_retention_days"`
		AppSettings          *adminTenantAppSettings `json:"app_settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	appSettings, existingMap, err := parseAdminTenantAppSettings(current.SettingsJSON)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if request.AppSettings != nil {
		appSettings = mergeAdminTenantAppSettings(appSettings, *request.AppSettings)
	}
	settingsJSON, err := marshalAdminTenantAppSettings(existingMap, appSettings)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	next := tenant.TenantSettingsInput{
		SenderEmail:          current.SenderEmail,
		SenderName:           current.SenderName,
		PayPalMode:           current.PayPalMode,
		PayPalClientID:       current.PayPalClientID,
		PayPalMerchantID:     current.PayPalMerchantID,
		DefaultRetentionDays: current.DefaultRetentionDays,
		SettingsJSON:         settingsJSON,
	}
	if request.SenderEmail != nil {
		next.SenderEmail = strings.TrimSpace(*request.SenderEmail)
	}
	if request.SenderName != nil {
		next.SenderName = strings.TrimSpace(*request.SenderName)
	}
	if request.PayPalMode != nil {
		next.PayPalMode = strings.TrimSpace(*request.PayPalMode)
	}
	if request.PayPalClientID != nil {
		next.PayPalClientID = strings.TrimSpace(*request.PayPalClientID)
	}
	if request.PayPalMerchantID != nil {
		next.PayPalMerchantID = strings.TrimSpace(*request.PayPalMerchantID)
	}
	if request.DefaultRetentionDays != nil {
		next.DefaultRetentionDays = *request.DefaultRetentionDays
	}

	updated, err := a.tenantRepo.UpsertSettings(r.Context(), tenant.UpsertTenantSettingsParams{
		TenantID: principal.TenantID,
		Settings: next,
	})
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	payload, err := tenantSettingsPayload(updated)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Tenant-Settings konnten nicht gelesen werden.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": payload,
	})
}

func tenantPayload(item tenant.Tenant) map[string]any {
	return map[string]any{
		"id":               item.ID,
		"slug":             item.Slug,
		"name":             item.Name,
		"public_base_url":  item.PublicBaseURL,
		"default_timezone": item.DefaultTimezone,
		"default_locale":   item.DefaultLocale,
		"status":           item.Status,
		"created_at":       item.CreatedAt.Format(time.RFC3339),
		"updated_at":       item.UpdatedAt.Format(time.RFC3339),
	}
}

func tenantSettingsPayload(item tenant.TenantSettings) (map[string]any, error) {
	appSettings, _, err := parseAdminTenantAppSettings(item.SettingsJSON)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"tenant_id":              item.TenantID,
		"sender_email":           item.SenderEmail,
		"sender_name":            item.SenderName,
		"paypal_mode":            item.PayPalMode,
		"paypal_client_id":       item.PayPalClientID,
		"paypal_merchant_id":     item.PayPalMerchantID,
		"default_retention_days": item.DefaultRetentionDays,
		"settings_json":          item.SettingsJSON,
		"app_settings":           appSettings,
		"created_at":             item.CreatedAt.Format(time.RFC3339),
		"updated_at":             item.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (a *App) writeTenantError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tenant.ErrTenantNotFound):
		writeAPIError(w, http.StatusNotFound, "TENANT_NOT_FOUND", "Mandant nicht gefunden.")
	case errors.Is(err, tenant.ErrTenantSettingsNotFound):
		writeAPIError(w, http.StatusNotFound, "TENANT_SETTINGS_NOT_FOUND", "Mandanten-Einstellungen nicht gefunden.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}

func parseAdminTenantAppSettings(raw string) (adminTenantAppSettings, map[string]any, error) {
	data := map[string]any{}
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			return adminTenantAppSettings{}, nil, fmt.Errorf("settings_json ist ungueltig")
		}
	}

	settings := adminTenantAppSettings{
		EventTimeStart:      defaultEventTimeStart,
		EventTimeEnd:        defaultEventTimeEnd,
		EventTimeStepMinute: defaultEventTimeStepMinutes,
		EventSlugMode:       defaultEventSlugMode,
		AllowedEmbedOrigins: []string{},
		EventDetailBaseURL:  "",
	}
	if value, ok := data["event_time_start"].(string); ok && strings.TrimSpace(value) != "" {
		settings.EventTimeStart = strings.TrimSpace(value)
	}
	if value, ok := data["event_time_end"].(string); ok && strings.TrimSpace(value) != "" {
		settings.EventTimeEnd = strings.TrimSpace(value)
	}
	if value, ok := data["event_time_step_minutes"].(float64); ok && value > 0 {
		settings.EventTimeStepMinute = int(value)
	}
	if value, ok := data["event_slug_mode"].(string); ok && strings.TrimSpace(value) != "" {
		settings.EventSlugMode = strings.TrimSpace(value)
	}
	if value, ok := data["event_detail_base_url"].(string); ok && strings.TrimSpace(value) != "" {
		settings.EventDetailBaseURL = strings.TrimSpace(value)
	}
	if rawOrigins, ok := data["allowed_embed_origins"].([]any); ok {
		settings.AllowedEmbedOrigins = make([]string, 0, len(rawOrigins))
		for _, entry := range rawOrigins {
			if value, ok := entry.(string); ok && strings.TrimSpace(value) != "" {
				settings.AllowedEmbedOrigins = append(settings.AllowedEmbedOrigins, strings.TrimSpace(value))
			}
		}
	}

	normalized, err := normalizeAdminTenantAppSettings(settings)
	if err != nil {
		return adminTenantAppSettings{}, nil, err
	}
	return normalized, data, nil
}

func mergeAdminTenantAppSettings(current, update adminTenantAppSettings) adminTenantAppSettings {
	next := current
	if strings.TrimSpace(update.EventTimeStart) != "" {
		next.EventTimeStart = strings.TrimSpace(update.EventTimeStart)
	}
	if strings.TrimSpace(update.EventTimeEnd) != "" {
		next.EventTimeEnd = strings.TrimSpace(update.EventTimeEnd)
	}
	if update.EventTimeStepMinute > 0 {
		next.EventTimeStepMinute = update.EventTimeStepMinute
	}
	if strings.TrimSpace(update.EventSlugMode) != "" {
		next.EventSlugMode = strings.TrimSpace(update.EventSlugMode)
	}
	if update.AllowedEmbedOrigins != nil {
		next.AllowedEmbedOrigins = update.AllowedEmbedOrigins
	}
	if update.EventDetailBaseURL != "" || current.EventDetailBaseURL != "" {
		next.EventDetailBaseURL = strings.TrimSpace(update.EventDetailBaseURL)
	}
	return next
}

func marshalAdminTenantAppSettings(existing map[string]any, settings adminTenantAppSettings) (string, error) {
	normalized, err := normalizeAdminTenantAppSettings(settings)
	if err != nil {
		return "", err
	}

	payload := map[string]any{}
	for key, value := range existing {
		payload[key] = value
	}
	payload["event_time_start"] = normalized.EventTimeStart
	payload["event_time_end"] = normalized.EventTimeEnd
	payload["event_time_step_minutes"] = normalized.EventTimeStepMinute
	payload["event_slug_mode"] = normalized.EventSlugMode
	payload["allowed_embed_origins"] = normalized.AllowedEmbedOrigins
	payload["event_detail_base_url"] = normalized.EventDetailBaseURL

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("settings_json konnte nicht serialisiert werden")
	}
	return string(encoded), nil
}

func normalizeAdminTenantAppSettings(settings adminTenantAppSettings) (adminTenantAppSettings, error) {
	step := settings.EventTimeStepMinute
	if step <= 0 {
		step = defaultEventTimeStepMinutes
	}
	if step > 60 || 60%step != 0 {
		return adminTenantAppSettings{}, fmt.Errorf("event_time_step_minutes muss ein Teiler von 60 sein")
	}

	start, err := normalizeScheduleSettingTime(settings.EventTimeStart, step)
	if err != nil {
		return adminTenantAppSettings{}, fmt.Errorf("event_time_start %w", err)
	}
	end, err := normalizeScheduleSettingTime(settings.EventTimeEnd, step)
	if err != nil {
		return adminTenantAppSettings{}, fmt.Errorf("event_time_end %w", err)
	}
	if scheduleTimeToMinutes(end) <= scheduleTimeToMinutes(start) {
		return adminTenantAppSettings{}, fmt.Errorf("event_time_end muss nach event_time_start liegen")
	}

	slugMode := strings.ToLower(strings.TrimSpace(settings.EventSlugMode))
	switch slugMode {
	case "", "optional":
		slugMode = defaultEventSlugMode
	case "required", "auto":
	default:
		return adminTenantAppSettings{}, fmt.Errorf("event_slug_mode ist ungueltig")
	}

	origins, err := normalizeAllowedEmbedOrigins(settings.AllowedEmbedOrigins)
	if err != nil {
		return adminTenantAppSettings{}, err
	}

	detailBaseURL := strings.TrimSpace(settings.EventDetailBaseURL)
	if detailBaseURL != "" {
		parsed, err := url.Parse(detailBaseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return adminTenantAppSettings{}, fmt.Errorf("event_detail_base_url muss eine absolute URL sein")
		}
		detailBaseURL = strings.TrimRight(parsed.String(), "/")
	}

	return adminTenantAppSettings{
		EventTimeStart:      start,
		EventTimeEnd:        end,
		EventTimeStepMinute: step,
		EventSlugMode:       slugMode,
		AllowedEmbedOrigins: origins,
		EventDetailBaseURL:  detailBaseURL,
	}, nil
}

func normalizeScheduleSettingTime(raw string, step int) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("darf nicht leer sein")
	}
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("muss im Format HH:MM vorliegen")
	}

	hours := 0
	minutes := 0
	if _, err := fmt.Sscanf(value, "%02d:%02d", &hours, &minutes); err != nil {
		return "", fmt.Errorf("muss im Format HH:MM vorliegen")
	}
	if hours < 0 || hours > 23 || minutes < 0 || minutes > 59 {
		return "", fmt.Errorf("ist ungueltig")
	}
	if minutes%step != 0 {
		return "", fmt.Errorf("muss auf die konfigurierte Schrittweite %d Minuten passen", step)
	}
	return fmt.Sprintf("%02d:%02d", hours, minutes), nil
}

func scheduleTimeToMinutes(value string) int {
	var hours, minutes int
	_, _ = fmt.Sscanf(value, "%02d:%02d", &hours, &minutes)
	return (hours * 60) + minutes
}

func normalizeAllowedEmbedOrigins(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}

	seen := map[string]struct{}{}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		if value == "*" {
			return []string{"*"}, nil
		}
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return nil, fmt.Errorf("allowed_embed_origins enthaelt eine ungueltige Origin")
		}
		origin := fmt.Sprintf("%s://%s", strings.ToLower(parsed.Scheme), strings.ToLower(parsed.Host))
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		result = append(result, origin)
	}
	sort.Strings(result)
	return result, nil
}
