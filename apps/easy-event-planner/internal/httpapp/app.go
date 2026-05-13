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
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

type App struct {
	cfg         config.Config
	mux         *http.ServeMux
	db          *sql.DB
	authService *auth.Service
	startedAt   time.Time
}

func New(cfg config.Config, sqlDB *sql.DB) *App {
	app := &App{
		cfg:       cfg,
		mux:       http.NewServeMux(),
		db:        sqlDB,
		startedAt: time.Now().UTC(),
	}
	if sqlDB != nil {
		app.authService = auth.NewService(
			sqlDB,
			tenant.NewRepository(sqlDB),
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
			nil,
		)
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
