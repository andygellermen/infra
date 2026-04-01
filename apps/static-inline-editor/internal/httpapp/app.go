package httpapp

import (
	"encoding/json"
	"net/http"

	"github.com/andygellermann/infra/apps/static-inline-editor/internal/config"
)

type App struct {
	cfg config.Config
	mux *http.ServeMux
}

func New(cfg config.Config) *App {
	app := &App{
		cfg: cfg,
		mux: http.NewServeMux(),
	}
	app.routes()
	return app
}

func (a *App) Handler() http.Handler {
	return a.mux
}

func (a *App) routes() {
	a.mux.HandleFunc("/healthz", a.handleHealth)
	a.mux.HandleFunc("/debug/tenants", a.handleTenants)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

func (a *App) handleTenants(w http.ResponseWriter, _ *http.Request) {
	type tenantInfo struct {
		Domain      string   `json:"domain"`
		LoginDomain string   `json:"login_domain"`
		Aliases     []string `json:"aliases,omitempty"`
		StartPath   string   `json:"start_path"`
	}

	out := make([]tenantInfo, 0, len(a.cfg.Tenants))
	for _, domain := range a.cfg.SortedTenantDomains() {
		tenant := a.cfg.Tenants[domain]
		out = append(out, tenantInfo{
			Domain:      tenant.Domain,
			LoginDomain: tenant.LoginDomain,
			Aliases:     tenant.Aliases,
			StartPath:   tenant.StartPath,
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"addr":    a.cfg.Addr,
		"tenants": out,
	})
}
