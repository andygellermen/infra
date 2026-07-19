package httpapp

import (
	"context"
	"net/http"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) tenantFeatureEnabled(ctx context.Context, tenantID, feature string) (bool, error) {
	if a.tenantRepo == nil {
		return false, nil
	}
	settings, err := a.tenantRepo.GetSettings(ctx, tenantID)
	if err != nil {
		return false, err
	}
	return tenant.FeatureEnabledInSettings(settings.SettingsJSON, feature), nil
}

func (a *App) requireTenantFeatureByTenantID(w http.ResponseWriter, r *http.Request, tenantID, feature, code, message string) bool {
	enabled, err := a.tenantFeatureEnabled(r.Context(), tenantID, feature)
	if err != nil {
		a.writeTenantError(w, err)
		return false
	}
	if !enabled {
		writeAPIError(w, http.StatusForbidden, code, message)
		return false
	}
	return true
}

func (a *App) requireTenantFeatureForPrincipal(w http.ResponseWriter, r *http.Request, principalTenantID, feature, code, message string) bool {
	return a.requireTenantFeatureByTenantID(w, r, principalTenantID, feature, code, message)
}
