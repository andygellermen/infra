package httpapp

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
)

type App struct {
	cfg       config.Config
	mux       *http.ServeMux
	startedAt time.Time
}

func New(cfg config.Config) *App {
	app := &App{
		cfg:       cfg,
		mux:       http.NewServeMux(),
		startedAt: time.Now().UTC(),
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
