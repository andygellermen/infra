package httpapp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handleSuperadminMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if !a.requireSuperadminToken(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": true,
	})
}

func (a *App) handleSuperadminTenantsCollection(w http.ResponseWriter, r *http.Request) {
	if !a.requireSuperadminToken(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.handleSuperadminTenantsList(w, r)
	case http.MethodPost:
		a.handleSuperadminTenantsCreate(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleSuperadminTenantsItem(w http.ResponseWriter, r *http.Request) {
	if !a.requireSuperadminToken(w, r) {
		return
	}

	tenantID, nested, nestedID, ok := parseSuperadminTenantPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "TENANT_NOT_FOUND", "Mandant nicht gefunden.")
		return
	}

	switch {
	case nested == "" && r.Method == http.MethodGet:
		a.handleSuperadminTenantGet(w, r, tenantID)
	case nested == "" && r.Method == http.MethodPatch:
		a.handleSuperadminTenantPatch(w, r, tenantID)
	case nested == "users" && nestedID == "" && r.Method == http.MethodGet:
		a.handleSuperadminTenantUsersList(w, r, tenantID)
	case nested == "users" && nestedID == "" && r.Method == http.MethodPost:
		a.handleSuperadminTenantUsersCreate(w, r, tenantID)
	case nested == "users" && nestedID != "" && r.Method == http.MethodPatch:
		a.handleSuperadminTenantUserPatch(w, r, tenantID, nestedID)
	case nested == "users" && nestedID != "" && r.Method == http.MethodDelete:
		a.handleSuperadminTenantUserDelete(w, r, tenantID, nestedID)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleSuperadminTenantsList(w http.ResponseWriter, r *http.Request) {
	items, err := a.tenantRepo.ListTenants(r.Context())
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		payload, buildErr := a.superadminTenantPayload(r, item)
		if buildErr != nil {
			a.writeTenantError(w, buildErr)
			return
		}
		result = append(result, payload)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handleSuperadminTenantsCreate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Slug            string                         `json:"slug"`
		Name            string                         `json:"name"`
		PublicBaseURL   string                         `json:"public_base_url"`
		DefaultTimezone string                         `json:"default_timezone"`
		DefaultLocale   string                         `json:"default_locale"`
		Status          string                         `json:"status"`
		Owner           *tenant.CreateTenantUserParams `json:"owner"`
		Settings        *superadminSettingsBody        `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	settingsInput, err := a.superadminTenantSettingsInput(request.Settings)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	created, err := a.tenantRepo.CreateTenant(r.Context(), tenant.CreateTenantParams{
		Slug:            request.Slug,
		Name:            request.Name,
		PublicBaseURL:   request.PublicBaseURL,
		DefaultTimezone: request.DefaultTimezone,
		DefaultLocale:   request.DefaultLocale,
		Status:          request.Status,
		Settings:        settingsInput,
	})
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	if request.Owner != nil {
		if _, err := a.tenantRepo.CreateUser(r.Context(), created.ID, *request.Owner); err != nil {
			a.writeTenantError(w, err)
			return
		}
	}

	payload, err := a.superadminTenantPayload(r, created)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"item": payload,
	})
}

func (a *App) handleSuperadminTenantGet(w http.ResponseWriter, r *http.Request, tenantID string) {
	item, err := a.tenantRepo.GetByID(r.Context(), tenantID)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	payload, err := a.superadminTenantPayload(r, item)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": payload})
}

func (a *App) handleSuperadminTenantPatch(w http.ResponseWriter, r *http.Request, tenantID string) {
	var request struct {
		Name            *string                 `json:"name"`
		PublicBaseURL   *string                 `json:"public_base_url"`
		DefaultTimezone *string                 `json:"default_timezone"`
		DefaultLocale   *string                 `json:"default_locale"`
		Settings        *superadminSettingsBody `json:"settings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	updated, err := a.tenantRepo.UpdateTenant(r.Context(), tenantID, tenant.UpdateTenantParams{
		Name:            request.Name,
		PublicBaseURL:   request.PublicBaseURL,
		DefaultTimezone: request.DefaultTimezone,
		DefaultLocale:   request.DefaultLocale,
	})
	if err != nil {
		a.writeTenantError(w, err)
		return
	}

	if request.Settings != nil {
		settingsInput, settingsErr := a.superadminTenantSettingsInput(request.Settings)
		if settingsErr != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", settingsErr.Error())
			return
		}
		if _, err := a.tenantRepo.UpsertSettings(r.Context(), tenant.UpsertTenantSettingsParams{
			TenantID: tenantID,
			Settings: settingsInput,
		}); err != nil {
			a.writeTenantError(w, err)
			return
		}
	}

	payload, err := a.superadminTenantPayload(r, updated)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": payload})
}

func (a *App) handleSuperadminTenantUsersList(w http.ResponseWriter, r *http.Request, tenantID string) {
	items, err := a.tenantRepo.ListUsers(r.Context(), tenantID)
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

func (a *App) handleSuperadminTenantUsersCreate(w http.ResponseWriter, r *http.Request, tenantID string) {
	var request tenant.CreateTenantUserParams
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}
	created, err := a.tenantRepo.CreateUser(r.Context(), tenantID, request)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"item": tenantUserPayload(created)})
}

func (a *App) handleSuperadminTenantUserPatch(w http.ResponseWriter, r *http.Request, tenantID, userID string) {
	var request tenant.UpdateTenantUserParams
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}
	updated, err := a.tenantRepo.UpdateUser(r.Context(), tenantID, userID, request)
	if err != nil {
		a.writeTenantError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": tenantUserPayload(updated)})
}

func (a *App) handleSuperadminTenantUserDelete(w http.ResponseWriter, r *http.Request, tenantID, userID string) {
	deleted, err := a.tenantRepo.DeleteUser(r.Context(), tenantID, userID)
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

type superadminSettingsBody struct {
	SenderEmail          string                  `json:"sender_email"`
	SenderName           string                  `json:"sender_name"`
	PayPalMode           string                  `json:"paypal_mode"`
	PayPalClientID       string                  `json:"paypal_client_id"`
	PayPalMerchantID     string                  `json:"paypal_merchant_id"`
	DefaultRetentionDays int                     `json:"default_retention_days"`
	AppSettings          *adminTenantAppSettings `json:"app_settings"`
}

func (a *App) superadminTenantPayload(r *http.Request, item tenant.Tenant) (map[string]any, error) {
	settings, err := a.tenantRepo.GetSettings(r.Context(), item.ID)
	if err != nil {
		return nil, err
	}
	settingsPayload, err := tenantSettingsPayload(settings)
	if err != nil {
		return nil, err
	}
	users, err := a.tenantRepo.ListUsers(r.Context(), item.ID)
	if err != nil {
		return nil, err
	}
	domains, err := a.tenantRepo.ListDomainBindings(r.Context(), item.ID)
	if err != nil {
		return nil, err
	}

	payload := tenantPayload(item)
	payload["settings"] = settingsPayload
	payload["user_count"] = len(users)
	payload["domain_count"] = len(domains)
	return payload, nil
}

func (a *App) superadminTenantSettingsInput(body *superadminSettingsBody) (tenant.TenantSettingsInput, error) {
	current := tenant.TenantSettingsInput{}
	if body == nil {
		return current, nil
	}

	appSettings := adminTenantAppSettings{
		CustomerStatus:                 tenant.CustomerStatusActive,
		EnabledFeatures:                tenant.EnabledFeaturesFromSettingsJSON(""),
		EventSlugMode:                  defaultEventSlugMode,
		EventTimeStart:                 defaultEventTimeStart,
		EventTimeEnd:                   defaultEventTimeEnd,
		EventTimeStepMinute:            defaultEventTimeStepMinutes,
		ParticipantCancelDeadlineHours: tenant.DefaultParticipantCancelDeadlineHours,
	}
	if body.AppSettings != nil {
		appSettings = *body.AppSettings
	}
	settingsJSON, err := marshalAdminTenantAppSettings(map[string]any{}, appSettings)
	if err != nil {
		return tenant.TenantSettingsInput{}, err
	}

	return tenant.TenantSettingsInput{
		SenderEmail:          body.SenderEmail,
		SenderName:           body.SenderName,
		PayPalMode:           body.PayPalMode,
		PayPalClientID:       body.PayPalClientID,
		PayPalMerchantID:     body.PayPalMerchantID,
		DefaultRetentionDays: body.DefaultRetentionDays,
		SettingsJSON:         settingsJSON,
	}, nil
}

func (a *App) requireSuperadminToken(w http.ResponseWriter, r *http.Request) bool {
	expected := strings.TrimSpace(a.cfg.SuperadminToken)
	if expected == "" {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "Route nicht gefunden.")
		return false
	}
	provided := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(r.Header.Get("Authorization")), "Bearer "))
	if provided == "" || provided != expected {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Zugriff verweigert.")
		return false
	}
	return true
}

func parseSuperadminTenantPath(path string) (tenantID, nested, nestedID string, ok bool) {
	const prefix = "/api/v1/internal/superadmin/tenants/"
	if !strings.HasPrefix(path, prefix) {
		return "", "", "", false
	}
	remainder := strings.Trim(strings.TrimSpace(strings.TrimPrefix(path, prefix)), "/")
	if remainder == "" {
		return "", "", "", false
	}
	parts := strings.Split(remainder, "/")
	switch len(parts) {
	case 1:
		return strings.TrimSpace(parts[0]), "", "", strings.TrimSpace(parts[0]) != ""
	case 2:
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), "", strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != ""
	case 3:
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" && strings.TrimSpace(parts[2]) != ""
	default:
		return "", "", "", false
	}
}
