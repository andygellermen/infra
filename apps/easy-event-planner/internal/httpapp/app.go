package httpapp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/auth"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/calendar"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/certificate"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/invitation"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/payment"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/privacy"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/registration"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/snippet"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

type App struct {
	cfg                config.Config
	mux                *http.ServeMux
	db                 *sql.DB
	authService        *auth.Service
	eventRepo          *event.Repository
	tenantRepo         *tenant.Repository
	regService         *registration.Service
	invitationService  *invitation.Service
	calendarService    *calendar.Service
	certificateService *certificate.Service
	paymentService     *payment.Service
	privacyService     *privacy.Service
	snippetRepo        *snippet.Repository
	startedAt          time.Time
}

func New(cfg config.Config, sqlDB *sql.DB) *App {
	app := &App{
		cfg:       cfg,
		mux:       http.NewServeMux(),
		db:        sqlDB,
		startedAt: time.Now().UTC(),
	}
	if sqlDB != nil {
		app.tenantRepo = tenant.NewRepository(sqlDB)
		app.authService = auth.NewService(
			sqlDB,
			app.tenantRepo,
			auth.Config{
				BaseURL:           cfg.BaseURL,
				TokenPepper:       cfg.TokenPepper,
				SessionTTL:        cfg.SessionTTL,
				MagicLinkTTL:      cfg.MagicLinkTTL,
				RegistrationTTL:   cfg.RegistrationTTL,
				WaitlistOfferTTL:  cfg.WaitlistOfferTTL,
				CertificateTTL:    cfg.CertificateTTL,
				RateLimitRequests: cfg.AuthRateLimit,
				RateLimitWindow:   cfg.AuthRateWindow,
			},
			buildMagicLinkSender(cfg),
		)
		app.eventRepo = event.NewRepository(sqlDB)
		app.regService = registration.NewService(sqlDB, registration.Config{
			BaseURL:          cfg.BaseURL,
			TokenPepper:      cfg.TokenPepper,
			RegistrationTTL:  cfg.RegistrationTTL,
			WaitlistOfferTTL: cfg.WaitlistOfferTTL,
		})
		app.invitationService = invitation.NewService(sqlDB)
		app.calendarService = calendar.NewService(sqlDB, calendar.Config{
			BaseURL:     cfg.BaseURL,
			TokenPepper: cfg.TokenPepper,
		})
		app.certificateService = certificate.NewService(sqlDB, certificate.Config{
			BaseURL:     cfg.BaseURL,
			TokenPepper: cfg.TokenPepper,
			StorageDir:  cfg.CertificateStorageDir,
		})
		app.paymentService = payment.NewService(sqlDB, payment.Config{
			FallbackClientID:     cfg.PayPalClientID,
			FallbackClientSecret: cfg.PayPalClientSecret,
			FallbackWebhookID:    cfg.PayPalWebhookID,
			SandboxAPIBaseURL:    cfg.PayPalSandboxAPIBaseURL,
			LiveAPIBaseURL:       cfg.PayPalLiveAPIBaseURL,
			HTTPTimeout:          cfg.PayPalHTTPTimeout,
			UseRealPayPalAPI:     cfg.PayPalUseRealAPI,
		})
		app.privacyService = privacy.NewService(sqlDB)
		app.snippetRepo = snippet.NewRepository(sqlDB)
	}
	app.routes()
	return app
}

func (a *App) Handler() http.Handler {
	return a.mux
}

func (a *App) routes() {
	a.mux.HandleFunc("/healthz", a.handleHealth)
	a.mux.HandleFunc("/readyz", a.handleReady)
	a.mux.HandleFunc("/version", a.handleVersion)

	a.mux.HandleFunc("/api/v1/auth/magic-link/request", a.handleMagicLinkRequest)
	a.mux.HandleFunc("/api/v1/auth/magic-link/verify", a.handleMagicLinkVerify)
	a.mux.HandleFunc("/api/v1/auth/me", a.handleAuthMe)
	a.mux.HandleFunc("/api/v1/auth/logout", a.handleAuthLogout)

	a.mux.HandleFunc("/api/v1/admin/event-series", a.handleAdminEventSeriesCollection)
	a.mux.HandleFunc("/api/v1/admin/event-series/", a.handleAdminEventSeriesItem)

	a.mux.HandleFunc("/api/v1/admin/dashboard", a.handleAdminDashboard)
	a.mux.HandleFunc("/api/v1/admin/events", a.handleAdminEventsCollection)
	a.mux.HandleFunc("/api/v1/admin/events/", a.handleAdminEventsItem)
	a.mux.HandleFunc("/api/v1/admin/calendar/feed", a.handleAdminCalendarFeed)
	a.mux.HandleFunc("/api/v1/admin/calendar/feed/rotate-token", a.handleAdminCalendarFeedRotateToken)
	a.mux.HandleFunc("/api/v1/admin/calendar/feed/embed-url", a.handleAdminCalendarFeedEmbedURL)
	a.mux.HandleFunc("/api/v1/admin/privacy/retention-policies", a.handleAdminRetentionPoliciesCollection)
	a.mux.HandleFunc("/api/v1/admin/privacy/retention-policies/", a.handleAdminRetentionPolicyItem)
	a.mux.HandleFunc("/api/v1/admin/privacy/retention-jobs/dry-run", a.handleAdminRetentionJobDryRun)
	a.mux.HandleFunc("/api/v1/admin/privacy/retention-jobs/run", a.handleAdminRetentionJobRun)
	a.mux.HandleFunc("/api/v1/admin/privacy/retention-jobs", a.handleAdminRetentionJobsList)
	a.mux.HandleFunc("/api/v1/admin/snippets", a.handleAdminSnippetsCollection)
	a.mux.HandleFunc("/api/v1/admin/snippets/", a.handleAdminSnippetsItem)
	a.mux.HandleFunc("/api/v1/admin/invitations", a.handleAdminInvitationsCollection)
	a.mux.HandleFunc("/api/v1/admin/invitations/", a.handleAdminInvitationsItem)
	a.mux.HandleFunc("/api/v1/admin/registrations/", a.handleAdminRegistrationItem)
	a.mux.HandleFunc("/api/v1/admin/waitlist/", a.handleAdminWaitlistItem)

	a.mux.HandleFunc("/api/v1/public/", a.handlePublicRoutes)
	a.mux.HandleFunc("/api/v1/webhooks/paypal", a.handlePayPalWebhook)
	a.mux.HandleFunc("/admin", a.handleAdminUIRoutes)
	a.mux.HandleFunc("/admin/", a.handleAdminUIRoutes)
	a.mux.HandleFunc("/admin-ui.css", a.handleAdminUIRoutes)
	a.mux.HandleFunc("/admin-ui.js", a.handleAdminUIRoutes)
	a.mux.HandleFunc("/smoke/footer-include.html", a.handleAdminUIRoutes)
	a.mux.HandleFunc("/", a.handleTenantAssetRoutes)
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeText(w, http.StatusOK, "ok\n")
}

func (a *App) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.db.PingContext(ctx); err != nil {
			http.Error(w, "database not ready", http.StatusServiceUnavailable)
			return
		}
	}
	writeText(w, http.StatusOK, "ready\n")
}

func (a *App) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"service":    a.cfg.AppName,
		"version":    a.cfg.Version,
		"env":        a.cfg.Env,
		"started_at": a.startedAt.Format(time.RFC3339),
	})
}

func (a *App) handleMagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Authentifizierung ist nicht verfuegbar.")
		return
	}

	var request struct {
		TenantSlug   string `json:"tenant_slug"`
		Email        string `json:"email"`
		Purpose      string `json:"purpose"`
		RedirectPath string `json:"redirect_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	_, err := a.authService.RequestMagicLink(r.Context(), auth.RequestMagicLinkInput{
		TenantSlug:   request.TenantSlug,
		Email:        request.Email,
		Purpose:      request.Purpose,
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
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Magic-Link konnte nicht angefordert werden.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "Wenn die E-Mail-Adresse bekannt ist, wurde ein Login-Link versendet.",
	})
}

func (a *App) handleMagicLinkVerify(w http.ResponseWriter, r *http.Request) {
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

	if result.SessionToken != "" {
		a.setSessionCookie(w, result.SessionToken, result.SessionExpiresAt)
	}

	if r.Method == http.MethodGet {
		http.Redirect(w, r, result.RedirectPath, http.StatusSeeOther)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"purpose":         result.Purpose,
		"redirect_path":   result.RedirectPath,
		"session_expires": result.SessionExpiresAt.Format(time.RFC3339),
	})
}

func (a *App) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Authentifizierung ist nicht verfuegbar.")
		return
	}

	rawSessionToken := a.sessionTokenFromCookie(r)
	principal, err := a.authService.AuthenticateSession(r.Context(), rawSessionToken)
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			a.clearSessionCookie(w)
			writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Nicht angemeldet.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Session konnte nicht gelesen werden.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"tenant": map[string]any{
			"id":   principal.TenantID,
			"slug": principal.TenantSlug,
		},
		"user": map[string]any{
			"id":    principal.UserID,
			"email": principal.Email,
			"name":  principal.Name,
			"role":  principal.Role,
		},
		"session_expires_at": principal.SessionExpiresAt.Format(time.RFC3339),
	})
}

func (a *App) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.authService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Authentifizierung ist nicht verfuegbar.")
		return
	}

	rawSessionToken := a.sessionTokenFromCookie(r)
	if rawSessionToken != "" {
		_, _ = a.authService.RevokeSession(r.Context(), rawSessionToken, clientIP(r), strings.TrimSpace(r.UserAgent()))
	}
	a.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleAdminEventSeriesCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleAdminEventSeriesList(w, r)
	case http.MethodPost:
		a.handleAdminEventSeriesCreate(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminEventSeriesItem(w http.ResponseWriter, r *http.Request) {
	seriesID, ok := parseEventSeriesIDFromPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "EVENT_SERIES_NOT_FOUND", "Event-Serie nicht gefunden.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.handleAdminEventSeriesGet(w, r, seriesID)
	case http.MethodPatch:
		a.handleAdminEventSeriesPatch(w, r, seriesID)
	case http.MethodDelete:
		a.handleAdminEventSeriesDelete(w, r, seriesID)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminEventSeriesList(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	series, err := a.eventRepo.ListSeries(r.Context(), principal.TenantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Event-Serien konnten nicht geladen werden.")
		return
	}

	items := make([]map[string]any, 0, len(series))
	for _, item := range series {
		items = append(items, eventSeriesPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": len(items),
	})
}

func (a *App) handleAdminEventSeriesCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		Slug                string `json:"slug"`
		Title               string `json:"title"`
		Description         string `json:"description"`
		DefaultLocationName string `json:"default_location_name"`
		DefaultAddress      string `json:"default_address"`
		DefaultOnlineURL    string `json:"default_online_url"`
		IsPublic            *bool  `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	created, err := a.eventRepo.CreateSeries(r.Context(), principal.TenantID, event.CreateSeriesParams{
		Slug:                request.Slug,
		Title:               request.Title,
		Description:         request.Description,
		DefaultLocationName: request.DefaultLocationName,
		DefaultAddress:      request.DefaultAddress,
		DefaultOnlineURL:    request.DefaultOnlineURL,
		IsPublic:            request.IsPublic,
	})
	if err != nil {
		if errors.Is(err, event.ErrSeriesSlugExists) {
			writeAPIError(w, http.StatusConflict, "EVENT_SERIES_SLUG_EXISTS", "Eine Event-Serie mit diesem Slug existiert bereits.")
			return
		}
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"item": eventSeriesPayload(created),
	})
}

func (a *App) handleAdminEventSeriesGet(w http.ResponseWriter, r *http.Request, seriesID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	item, err := a.eventRepo.GetSeriesByID(r.Context(), principal.TenantID, seriesID)
	if err != nil {
		if errors.Is(err, event.ErrSeriesNotFound) {
			writeAPIError(w, http.StatusNotFound, "EVENT_SERIES_NOT_FOUND", "Event-Serie nicht gefunden.")
			return
		}
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Event-Serie konnte nicht geladen werden.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventSeriesPayload(item),
	})
}

func (a *App) handleAdminEventSeriesPatch(w http.ResponseWriter, r *http.Request, seriesID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		Slug                *string `json:"slug"`
		Title               *string `json:"title"`
		Description         *string `json:"description"`
		DefaultLocationName *string `json:"default_location_name"`
		DefaultAddress      *string `json:"default_address"`
		DefaultOnlineURL    *string `json:"default_online_url"`
		IsPublic            *bool   `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	updated, err := a.eventRepo.UpdateSeries(r.Context(), principal.TenantID, seriesID, event.UpdateSeriesParams{
		Slug:                request.Slug,
		Title:               request.Title,
		Description:         request.Description,
		DefaultLocationName: request.DefaultLocationName,
		DefaultAddress:      request.DefaultAddress,
		DefaultOnlineURL:    request.DefaultOnlineURL,
		IsPublic:            request.IsPublic,
	})
	if err != nil {
		if errors.Is(err, event.ErrSeriesNotFound) {
			writeAPIError(w, http.StatusNotFound, "EVENT_SERIES_NOT_FOUND", "Event-Serie nicht gefunden.")
			return
		}
		if errors.Is(err, event.ErrSeriesSlugExists) {
			writeAPIError(w, http.StatusConflict, "EVENT_SERIES_SLUG_EXISTS", "Eine Event-Serie mit diesem Slug existiert bereits.")
			return
		}
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventSeriesPayload(updated),
	})
}

func (a *App) handleAdminEventSeriesDelete(w http.ResponseWriter, r *http.Request, seriesID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	deleted, err := a.eventRepo.DeleteSeries(r.Context(), principal.TenantID, seriesID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Event-Serie konnte nicht geloescht werden.")
		return
	}
	if !deleted {
		writeAPIError(w, http.StatusNotFound, "EVENT_SERIES_NOT_FOUND", "Event-Serie nicht gefunden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) requireAdminPrincipal(w http.ResponseWriter, r *http.Request, writeAccess bool) (auth.SessionPrincipal, bool) {
	if a.authService == nil || a.eventRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service ist nicht verfuegbar.")
		return auth.SessionPrincipal{}, false
	}

	rawSessionToken := a.sessionTokenFromCookie(r)
	principal, err := a.authService.AuthenticateSession(r.Context(), rawSessionToken)
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			a.clearSessionCookie(w)
			writeAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Nicht angemeldet.")
			return auth.SessionPrincipal{}, false
		}
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Session konnte nicht gelesen werden.")
		return auth.SessionPrincipal{}, false
	}

	if !isAdminRoleAllowed(principal.Role, writeAccess) {
		writeAPIError(w, http.StatusForbidden, "FORBIDDEN", "Zugriff verweigert.")
		return auth.SessionPrincipal{}, false
	}

	return principal, true
}

func isAdminRoleAllowed(role string, writeAccess bool) bool {
	normalized := strings.ToLower(strings.TrimSpace(role))
	switch normalized {
	case "owner", "admin":
		return true
	case "event_manager":
		return true
	case "readonly":
		return !writeAccess
	default:
		return false
	}
}

func parseEventSeriesIDFromPath(path string) (string, bool) {
	const prefix = "/api/v1/admin/event-series/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}
	raw := strings.TrimSpace(strings.TrimPrefix(path, prefix))
	if raw == "" || strings.Contains(raw, "/") {
		return "", false
	}
	return raw, true
}

func eventSeriesPayload(series event.EventSeries) map[string]any {
	return map[string]any{
		"id":                    series.ID,
		"tenant_id":             series.TenantID,
		"slug":                  series.Slug,
		"title":                 series.Title,
		"description":           series.Description,
		"default_location_name": series.DefaultLocationName,
		"default_address":       series.DefaultAddress,
		"default_online_url":    series.DefaultOnlineURL,
		"is_public":             series.IsPublic,
		"created_at":            series.CreatedAt.Format(time.RFC3339),
		"updated_at":            series.UpdatedAt.Format(time.RFC3339),
	}
}

func (a *App) handleAdminEventsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleAdminEventsList(w, r)
	case http.MethodPost:
		a.handleAdminEventsCreate(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminEventsItem(w http.ResponseWriter, r *http.Request) {
	eventID, action, ok := parseAdminEventPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
		return
	}

	if action == "" {
		switch r.Method {
		case http.MethodGet:
			a.handleAdminEventGet(w, r, eventID)
		case http.MethodPatch:
			a.handleAdminEventPatch(w, r, eventID)
		case http.MethodDelete:
			a.handleAdminEventDelete(w, r, eventID)
		default:
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		}
		return
	}

	if action == "waitlist" {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handleAdminEventWaitlistList(w, r, eventID)
		return
	}
	if action == "registrations" {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handleAdminEventRegistrationList(w, r, eventID)
		return
	}
	if action == "registrations/manual" {
		if r.Method != http.MethodPost {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handleAdminEventRegistrationManualCreate(w, r, eventID)
		return
	}

	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	switch action {
	case "publish":
		a.handleAdminEventPublish(w, r, eventID)
	case "unpublish":
		a.handleAdminEventUnpublish(w, r, eventID)
	case "cancel":
		a.handleAdminEventCancel(w, r, eventID)
	case "postpone":
		a.handleAdminEventPostpone(w, r, eventID)
	case "mark-completed":
		a.handleAdminEventMarkCompleted(w, r, eventID)
	default:
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
	}
}

func (a *App) handleAdminEventsList(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	items, err := a.eventRepo.ListEvents(r.Context(), principal.TenantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Veranstaltungen konnten nicht geladen werden.")
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, eventPayload(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handleAdminEventsCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		SeriesID            string `json:"series_id"`
		Slug                string `json:"slug"`
		Title               string `json:"title"`
		Subtitle            string `json:"subtitle"`
		Description         string `json:"description"`
		StartsAt            string `json:"starts_at"`
		EndsAt              string `json:"ends_at"`
		Timezone            string `json:"timezone"`
		LocationName        string `json:"location_name"`
		Address             string `json:"address"`
		OnlineURL           string `json:"online_url"`
		ParticipationMode   string `json:"participation_mode"`
		IsPublic            *bool  `json:"is_public"`
		RegistrationEnabled *bool  `json:"registration_enabled"`
		WaitlistEnabled     *bool  `json:"waitlist_enabled"`
		MaxParticipants     *int   `json:"max_participants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	created, err := a.eventRepo.CreateEvent(r.Context(), principal.TenantID, event.CreateEventParams{
		SeriesID:            request.SeriesID,
		Slug:                request.Slug,
		Title:               request.Title,
		Subtitle:            request.Subtitle,
		Description:         request.Description,
		StartsAt:            request.StartsAt,
		EndsAt:              request.EndsAt,
		Timezone:            request.Timezone,
		LocationName:        request.LocationName,
		Address:             request.Address,
		OnlineURL:           request.OnlineURL,
		ParticipationMode:   request.ParticipationMode,
		IsPublic:            request.IsPublic,
		RegistrationEnabled: request.RegistrationEnabled,
		WaitlistEnabled:     request.WaitlistEnabled,
		MaxParticipants:     request.MaxParticipants,
	})
	if err != nil {
		a.writeEventError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"item": eventPayload(created),
	})
}

func (a *App) handleAdminEventGet(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}

	item, err := a.eventRepo.GetEventByID(r.Context(), principal.TenantID, eventID)
	if err != nil {
		a.writeEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventPayload(item),
	})
}

func (a *App) handleAdminEventPatch(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		SeriesID             *string `json:"series_id"`
		Slug                 *string `json:"slug"`
		Title                *string `json:"title"`
		Subtitle             *string `json:"subtitle"`
		Description          *string `json:"description"`
		StartsAt             *string `json:"starts_at"`
		EndsAt               *string `json:"ends_at"`
		Timezone             *string `json:"timezone"`
		LocationName         *string `json:"location_name"`
		Address              *string `json:"address"`
		OnlineURL            *string `json:"online_url"`
		ParticipationMode    *string `json:"participation_mode"`
		IsPublic             *bool   `json:"is_public"`
		RegistrationEnabled  *bool   `json:"registration_enabled"`
		WaitlistEnabled      *bool   `json:"waitlist_enabled"`
		MaxParticipants      *int    `json:"max_participants"`
		ClearMaxParticipants bool    `json:"clear_max_participants"`
		ChangeNote           *string `json:"change_note"`
		CancelledReason      *string `json:"cancelled_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	updated, err := a.eventRepo.UpdateEvent(r.Context(), principal.TenantID, eventID, event.UpdateEventParams{
		SeriesID:             request.SeriesID,
		Slug:                 request.Slug,
		Title:                request.Title,
		Subtitle:             request.Subtitle,
		Description:          request.Description,
		StartsAt:             request.StartsAt,
		EndsAt:               request.EndsAt,
		Timezone:             request.Timezone,
		LocationName:         request.LocationName,
		Address:              request.Address,
		OnlineURL:            request.OnlineURL,
		ParticipationMode:    request.ParticipationMode,
		IsPublic:             request.IsPublic,
		RegistrationEnabled:  request.RegistrationEnabled,
		WaitlistEnabled:      request.WaitlistEnabled,
		MaxParticipants:      request.MaxParticipants,
		ClearMaxParticipants: request.ClearMaxParticipants,
		ChangeNote:           request.ChangeNote,
		CancelledReason:      request.CancelledReason,
	})
	if err != nil {
		a.writeEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventPayload(updated),
	})
}

func (a *App) handleAdminEventDelete(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	deleted, err := a.eventRepo.DeleteEvent(r.Context(), principal.TenantID, eventID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Veranstaltung konnte nicht geloescht werden.")
		return
	}
	if !deleted {
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleAdminEventPublish(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	item, err := a.eventRepo.PublishEvent(r.Context(), principal.TenantID, eventID)
	if err != nil {
		a.writeEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventPayload(item),
	})
}

func (a *App) handleAdminEventUnpublish(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	item, err := a.eventRepo.UnpublishEvent(r.Context(), principal.TenantID, eventID)
	if err != nil {
		a.writeEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventPayload(item),
	})
}

func (a *App) handleAdminEventCancel(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		CancelledReason string `json:"cancelled_reason"`
		ChangeNote      string `json:"change_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	item, err := a.eventRepo.CancelEvent(
		r.Context(),
		principal.TenantID,
		eventID,
		request.CancelledReason,
		request.ChangeNote,
	)
	if err != nil {
		a.writeEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventPayload(item),
	})
}

func (a *App) handleAdminEventPostpone(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	var request struct {
		StartsAt   string `json:"starts_at"`
		EndsAt     string `json:"ends_at"`
		ChangeNote string `json:"change_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	item, err := a.eventRepo.PostponeEvent(
		r.Context(),
		principal.TenantID,
		eventID,
		request.StartsAt,
		request.EndsAt,
		request.ChangeNote,
	)
	if err != nil {
		a.writeEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventPayload(item),
	})
}

func (a *App) handleAdminEventMarkCompleted(w http.ResponseWriter, r *http.Request, eventID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}

	item, err := a.eventRepo.MarkEventCompleted(r.Context(), principal.TenantID, eventID)
	if err != nil {
		a.writeEventError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": eventPayload(item),
	})
}

func (a *App) writeEventError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, event.ErrEventNotFound):
		writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
	case errors.Is(err, event.ErrEventSlugExists):
		writeAPIError(w, http.StatusConflict, "EVENT_SLUG_EXISTS", "Eine Veranstaltung mit diesem Slug existiert bereits.")
	case errors.Is(err, event.ErrInvalidStatusTransition):
		writeAPIError(w, http.StatusConflict, "EVENT_STATUS_INVALID", "Statuswechsel ist fuer diese Veranstaltung nicht erlaubt.")
	case errors.Is(err, event.ErrEventSeriesScopeMismatch):
		writeAPIError(w, http.StatusBadRequest, "EVENT_SERIES_INVALID", "Die Event-Serie gehoert nicht zu diesem Tenant.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}

func parseAdminEventPath(path string) (eventID, action string, ok bool) {
	const prefix = "/api/v1/admin/events/"
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

	id := strings.TrimSpace(parts[0])
	if id == "" {
		return "", "", false
	}
	actionParts := make([]string, 0, len(parts)-1)
	for _, part := range parts[1:] {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return "", "", false
		}
		actionParts = append(actionParts, trimmed)
	}
	if len(actionParts) == 0 {
		return "", "", false
	}
	return id, strings.Join(actionParts, "/"), true
}

func eventPayload(item event.Event) map[string]any {
	var endsAt any
	if item.EndsAt != nil {
		endsAt = item.EndsAt.UTC().Format(time.RFC3339)
	}
	var seriesID any
	if strings.TrimSpace(item.SeriesID) != "" {
		seriesID = item.SeriesID
	}
	var maxParticipants any
	if item.MaxParticipants != nil {
		maxParticipants = *item.MaxParticipants
	}

	return map[string]any{
		"id":                   item.ID,
		"tenant_id":            item.TenantID,
		"series_id":            seriesID,
		"slug":                 item.Slug,
		"title":                item.Title,
		"subtitle":             item.Subtitle,
		"description":          item.Description,
		"starts_at":            item.StartsAt.UTC().Format(time.RFC3339),
		"ends_at":              endsAt,
		"timezone":             item.Timezone,
		"location_name":        item.LocationName,
		"address":              item.Address,
		"online_url":           item.OnlineURL,
		"participation_mode":   item.ParticipationMode,
		"status":               item.Status,
		"is_public":            item.IsPublic,
		"registration_enabled": item.RegistrationEnabled,
		"waitlist_enabled":     item.WaitlistEnabled,
		"max_participants":     maxParticipants,
		"change_note":          item.ChangeNote,
		"cancelled_reason":     item.CancelledReason,
		"created_at":           item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":           item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (a *App) sessionTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(a.cfg.SessionCookieName)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func (a *App) setSessionCookie(w http.ResponseWriter, rawSessionToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.SessionCookieName,
		Value:    rawSessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt.UTC(),
	})
}

func (a *App) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cfg.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}

func clientIP(r *http.Request) string {
	forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func writeText(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
			"details": map[string]any{},
		},
	})
}
