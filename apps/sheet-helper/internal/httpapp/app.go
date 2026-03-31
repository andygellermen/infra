package httpapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/andygellermann/infra/apps/sheet-helper/internal/config"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/model"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/storage"
)

type Syncer interface {
	Sync(ctx context.Context) error
}

type App struct {
	cfg     config.Config
	store   *storage.Store
	syncers map[string]Syncer
	mux     *http.ServeMux
}

func New(cfg config.Config, store *storage.Store, syncers map[string]Syncer) *App {
	app := &App{
		cfg:     cfg,
		store:   store,
		syncers: syncers,
		mux:     http.NewServeMux(),
	}
	app.routes()
	return app
}

func (a *App) Handler() http.Handler {
	return a.mux
}

func (a *App) routes() {
	a.mux.HandleFunc("/healthz", a.handleHealth)
	a.mux.HandleFunc("/internal/sync/", a.handleSync)
	a.mux.HandleFunc("/", a.handleRoot)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

func (a *App) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenant := strings.TrimPrefix(r.URL.Path, "/internal/sync/")
	tenantCfg, ok := a.cfg.Tenants[tenant]
	if tenant == "" || !ok {
		http.Error(w, "unknown tenant", http.StatusNotFound)
		return
	}
	syncer, ok := a.syncers[tenant]
	if !ok {
		http.NotFound(w, r)
		return
	}
	if tenantCfg.SyncToken == "" {
		http.Error(w, "sync token not configured", http.StatusForbidden)
		return
	}
	if r.Header.Get("X-Sheet-Helper-Token") != tenantCfg.SyncToken {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := syncer.Sync(r.Context()); err != nil {
		http.Error(w, "sync failed: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	domain := stripPort(r.Host)
	path := normalizedPath(r.URL.Path)

	route, found, err := a.store.LookupRoute(ctx, domain, path)
	if err != nil {
		http.Error(w, "route lookup failed", http.StatusInternalServerError)
		return
	}
	if !found {
		http.NotFound(w, r)
		return
	}

	if route.Passphrase != "" && !a.hasAccess(r, route) {
		if r.Method == http.MethodPost {
			a.handleUnlock(w, r, route)
			return
		}
		a.renderPasswordPrompt(w, route, "")
		return
	}

	switch route.Type {
	case model.RouteTypeLink:
		a.recordClickAsync(ctx, r, route)
		http.Redirect(w, r, route.Target, http.StatusFound)
	case model.RouteTypeText:
		a.handleText(w, r, route)
	case model.RouteTypeVCard:
		a.handleVCard(w, r, route)
	case model.RouteTypeList:
		a.handleList(w, r, route)
	default:
		http.Error(w, "unsupported route type", http.StatusBadRequest)
	}
}

func (a *App) handleUnlock(w http.ResponseWriter, r *http.Request, route model.Route) {
	if err := r.ParseForm(); err != nil {
		a.renderPasswordPrompt(w, route, "Formular konnte nicht gelesen werden.")
		return
	}

	if r.FormValue("passphrase") != route.Passphrase {
		a.renderPasswordPrompt(w, route, "Passphrase ist nicht korrekt.")
		return
	}

	cookie := &http.Cookie{
		Name:     accessCookieName(route),
		Value:    a.signRoute(route),
		Path:     route.Path,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, route.Path, http.StatusSeeOther)
}

func (a *App) handleText(w http.ResponseWriter, r *http.Request, route model.Route) {
	entry, found, err := a.store.GetText(r.Context(), route.Domain, route.Path)
	if err != nil {
		http.Error(w, "text lookup failed", http.StatusInternalServerError)
		return
	}

	body := route.Target
	copyHint := "Inhalt kopieren"
	if found {
		body = entry.Content
		if entry.CopyHint != "" {
			copyHint = entry.CopyHint
		}
	}

	data := map[string]any{
		"Title":       fallback(route.Title, route.Path),
		"Description": route.Description,
		"Content":     body,
		"CopyHint":    copyHint,
	}
	renderHTML(w, textPageTemplate, data)
}

func (a *App) handleVCard(w http.ResponseWriter, r *http.Request, route model.Route) {
	entry, found, err := a.store.GetVCard(r.Context(), route.Domain, route.Path)
	if err != nil {
		http.Error(w, "vcard lookup failed", http.StatusInternalServerError)
		return
	}
	if !found {
		http.NotFound(w, r)
		return
	}

	if r.URL.Query().Get("download") == "vcf" {
		filename := strings.ReplaceAll(strings.ToLower(entry.FullName), " ", "-")
		w.Header().Set("Content-Type", "text/vcard; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.vcf"`, filename))
		_, _ = w.Write([]byte(buildVCF(entry)))
		return
	}

	data := map[string]any{
		"Title":       fallback(route.Title, entry.FullName),
		"Description": route.Description,
		"Entry":       entry,
		"DownloadURL": route.Path + "?download=vcf",
	}
	renderHTML(w, vcardPageTemplate, data)
}

func (a *App) handleList(w http.ResponseWriter, r *http.Request, route model.Route) {
	items, err := a.store.ListItems(r.Context(), route.ListSheet)
	if err != nil {
		http.Error(w, "list lookup failed", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title":       fallback(route.Title, route.Path),
		"Description": route.Description,
		"Items":       items,
	}
	renderHTML(w, listPageTemplate, data)
}

func (a *App) renderPasswordPrompt(w http.ResponseWriter, route model.Route, errorMessage string) {
	data := map[string]any{
		"Title":       fallback(route.Title, route.Path),
		"Description": route.Description,
		"Error":       errorMessage,
	}
	renderHTML(w, unlockPageTemplate, data)
}

func (a *App) hasAccess(r *http.Request, route model.Route) bool {
	cookie, err := r.Cookie(accessCookieName(route))
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(cookie.Value), []byte(a.signRoute(route)))
}

func (a *App) signRoute(route model.Route) string {
	tenant, ok := a.cfg.Tenants[route.Domain]
	secret := "dev-only-change-me"
	if ok && tenant.CookieSecret != "" {
		secret = tenant.CookieSecret
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(route.Domain))
	_, _ = mac.Write([]byte("|"))
	_, _ = mac.Write([]byte(route.Path))
	_, _ = mac.Write([]byte("|"))
	_, _ = mac.Write([]byte(route.Passphrase))
	return hex.EncodeToString(mac.Sum(nil))
}

func (a *App) recordClickAsync(ctx context.Context, r *http.Request, route model.Route) {
	go func() {
		if err := a.store.RecordClick(ctx, model.ClickEvent{
			Domain:    route.Domain,
			Path:      route.Path,
			Type:      route.Type,
			Target:    route.Target,
			Referrer:  r.Referer(),
			UserAgent: r.UserAgent(),
		}); err != nil {
			log.Printf("record click: %v", err)
		}
	}()
}

func stripPort(host string) string {
	if value, _, err := net.SplitHostPort(host); err == nil {
		return value
	}
	return host
}

func normalizedPath(path string) string {
	if path == "" {
		return "/"
	}
	return path
}

func accessCookieName(route model.Route) string {
	hash := sha256.Sum256([]byte(route.Domain + "|" + route.Path))
	return "sh_auth_" + hex.EncodeToString(hash[:8])
}

func buildVCF(entry model.VCardEntry) string {
	lines := []string{
		"BEGIN:VCARD",
		"VERSION:3.0",
		"N:" + entry.FullName,
		"FN:" + entry.FullName,
		"ORG:" + entry.Organization,
		"TITLE:" + entry.JobTitle,
		"EMAIL:" + entry.Email,
		"TEL;TYPE=CELL:" + entry.PhoneMobile,
		"ADR;TYPE=WORK:;;" + entry.Address,
		"URL:" + entry.Website,
		"NOTE:" + fallback(entry.Note, entry.JobTitle),
		"END:VCARD",
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}

func fallback(value, alt string) string {
	if value != "" {
		return value
	}
	return alt
}

func renderHTML(w http.ResponseWriter, page string, data any) {
	tpl := template.Must(template.New("page").Parse(page))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}
