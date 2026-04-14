package httpapp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/static-inline-editor/internal/auth"
	"github.com/andygellermann/infra/apps/static-inline-editor/internal/config"
	"github.com/andygellermann/infra/apps/static-inline-editor/internal/editor"
	"github.com/andygellermann/infra/apps/static-inline-editor/internal/gitops"
	"github.com/andygellermann/infra/apps/static-inline-editor/internal/model"
)

const (
	defaultContentToolsCSSURL = "https://cdn.jsdelivr.net/npm/ContentTools@1.6.1/build/content-tools.min.css"
	defaultContentToolsJSURL  = "https://cdn.jsdelivr.net/npm/ContentTools@1.6.1/build/content-tools.min.js"
)

type App struct {
	cfg    config.Config
	mux    *http.ServeMux
	store  *auth.Store
	mailer auth.Mailer
}

func (a *App) mailerForTenant(tenant model.Tenant) auth.Mailer {
	fromEmail := strings.TrimSpace(tenant.FromEmail)
	if fromEmail == "" {
		return a.mailer
	}
	return auth.NewMailer(a.cfg.SMTPHost, a.cfg.SMTPPort, a.cfg.SMTPUsername, a.cfg.SMTPPassword, fromEmail, a.cfg.SMTPFromName)
}

func New(cfg config.Config) *App {
	if strings.TrimSpace(cfg.ContentToolsCSSURL) == "" {
		cfg.ContentToolsCSSURL = defaultContentToolsCSSURL
	}
	if strings.TrimSpace(cfg.ContentToolsJSURL) == "" {
		cfg.ContentToolsJSURL = defaultContentToolsJSURL
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Printf("static-inline-editor: create data dir %s: %v", cfg.DataDir, err)
	}
	app := &App{
		cfg:    cfg,
		mux:    http.NewServeMux(),
		store:  auth.NewStore(filepath.Join(cfg.DataDir, "auth-state.json")),
		mailer: auth.NewMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFromEmail, cfg.SMTPFromName),
	}
	app.routes()
	return app
}

func (a *App) Handler() http.Handler {
	return a.mux
}

func (a *App) routes() {
	a.mux.HandleFunc("/", a.handleHome)
	a.mux.HandleFunc("/login", a.handleLogin)
	a.mux.HandleFunc("/healthz", a.handleHealth)
	a.mux.HandleFunc("/debug/tenants", a.handleTenants)
	a.mux.HandleFunc("/auth/request-link", a.handleRequestLink)
	a.mux.HandleFunc("/auth/verify", a.handleVerify)
	a.mux.HandleFunc("/auth/logout", a.handleLogout)
	a.mux.HandleFunc("/edit", a.handleEdit)
	a.mux.HandleFunc("/preview", a.handlePreview)
	a.mux.HandleFunc("/save", a.handleSave)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.handleStaticAsset(w, r)
		return
	}

	tenant := a.tenantForHost(r.Host)
	if tenant.Domain == "" {
		http.Error(w, "unknown tenant host", http.StatusNotFound)
		return
	}

	session, ok := a.currentSession(r)
	if !ok || session.Tenant != tenant.Domain {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	writeHTML(w, http.StatusOK, fmt.Sprintf(`<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Static Inline Editor</title>
  <style>
    body { font-family: Georgia, serif; max-width: 720px; margin: 3rem auto; padding: 0 1rem; line-height: 1.5; }
    .card { border: 1px solid #d9d1c4; border-radius: 14px; padding: 1.25rem 1.5rem; background: #fffdf8; }
    a, button { color: #7a2f16; }
    button { font: inherit; padding: 0.7rem 1rem; border-radius: 10px; border: 1px solid #7a2f16; background: #f4e5d3; cursor: pointer; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Bearbeitung freigeschaltet</h1>
    <p>Angemeldet fuer <strong>%s</strong> auf <strong>%s</strong>.</p>
    <p>Der Edit-Modus ist bereit. Ueber den folgenden Link oeffnest du direkt die konfigurierte Startseite im Editor.</p>
    <p><a href="%s">Startseite im Editor oeffnen</a></p>
    <form method="post" action="/auth/logout">
      <button type="submit">Abmelden</button>
    </form>
  </div>
</body>
</html>`, htmlEscape(session.Email), htmlEscape(tenant.Domain), htmlEscape(a.startEditURL(tenant.StartPath))))
}

func (a *App) handleStaticAsset(w http.ResponseWriter, r *http.Request) {
	tenant := a.tenantForHost(r.Host)
	if tenant.Domain == "" {
		http.Error(w, "unknown tenant host", http.StatusNotFound)
		return
	}

	resolvedPath, err := resolveStaticPath(tenant.StaticRoot, r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if info.IsDir() {
		resolvedPath = filepath.Join(resolvedPath, "index.html")
		if _, err := os.Stat(resolvedPath); err != nil {
			http.NotFound(w, r)
			return
		}
	}

	http.ServeFile(w, r, resolvedPath)
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenant := a.tenantForHost(r.Host)
	if tenant.Domain == "" {
		http.Error(w, "unknown tenant host", http.StatusNotFound)
		return
	}

	if session, ok := a.currentSession(r); ok && session.Tenant == tenant.Domain {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	writeHTML(w, http.StatusOK, fmt.Sprintf(`<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Magic Link Login</title>
  <style>
    body { font-family: Georgia, serif; max-width: 720px; margin: 3rem auto; padding: 0 1rem; line-height: 1.5; color: #2e241e; }
    .card { border: 1px solid #d9d1c4; border-radius: 14px; padding: 1.5rem; background: #fffdf8; }
    input, button { font: inherit; padding: 0.8rem 0.9rem; width: 100%%; box-sizing: border-box; border-radius: 10px; }
    input { border: 1px solid #c9bfaf; margin: 0.5rem 0 1rem; }
    button { border: 1px solid #7a2f16; background: #f4e5d3; cursor: pointer; }
    small { color: #65584e; }
  </style>
</head>
<body>
  <div class="card">
    <h1>Bearbeitung anmelden</h1>
    <p>Wir senden dir einen Magic Link an eine freigegebene E-Mail-Adresse fuer <strong>%s</strong>.</p>
    <form method="post" action="/auth/request-link">
      <label for="email">E-Mail-Adresse</label>
      <input id="email" name="email" type="email" autocomplete="email" required>
      <button type="submit">Magic Link anfordern</button>
    </form>
    <small>Der Link ist zeitlich begrenzt und nur fuer diese Bearbeitungsdomain gueltig.</small>
  </div>
</body>
</html>`, htmlEscape(tenant.Domain)))
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

func (a *App) handleRequestLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.MagicLinkRequest
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		req.Email = r.FormValue("email")
	}

	tenant := a.tenantForHost(r.Host)
	if tenant.Domain == "" {
		http.Error(w, "unknown tenant host", http.StatusNotFound)
		return
	}
	result := model.MagicLinkRequestResult{
		OK:      true,
		Message: "Wenn die E-Mail-Adresse freigegeben ist, wurde ein Magic-Link versendet.",
	}

	if auth.EmailAllowed(tenant.AllowedEmails, req.Email) {
		token, err := a.store.CreateMagicLink(tenant.Domain, strings.TrimSpace(req.Email), a.cfg.MagicLinkTTL)
		if err != nil {
			http.Error(w, "could not create magic link", http.StatusInternalServerError)
			return
		}
		verifyURL := a.verifyURL(r, tenant, token)
		if err := a.mailerForTenant(tenant).SendMagicLink(strings.TrimSpace(req.Email), verifyURL); err != nil {
			log.Printf("static-inline-editor: request-link failed tenant=%s email=%s: %v", tenant.Domain, strings.TrimSpace(req.Email), err)
			http.Error(w, "could not send magic link", http.StatusInternalServerError)
			return
		}
	}

	if acceptsHTML(r) {
		writeHTML(w, http.StatusOK, `<!doctype html>
<html lang="de">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Magic Link versendet</title></head>
<body style="font-family: Georgia, serif; max-width: 720px; margin: 3rem auto; padding: 0 1rem;">
  <h1>Fast geschafft</h1>
  <p>Wenn die E-Mail-Adresse freigegeben ist, wurde ein Magic Link versendet.</p>
  <p><a href="/login">Zurueck zum Login</a></p>
</body>
</html>`)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(result)
}

func (a *App) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	magicLink, err := a.store.ConsumeMagicLink(token)
	if err != nil {
		http.Error(w, "magic link invalid or expired", http.StatusBadRequest)
		return
	}

	sessionToken, err := a.store.CreateSession(magicLink.Tenant, magicLink.Email, a.cfg.SessionTTL)
	if err != nil {
		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "static_editor_session",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   a.cfg.SecureCookies,
	})

	startPath := "/index.html"
	if tenant, ok := a.cfg.Tenants[magicLink.Tenant]; ok {
		startPath = tenant.StartPath
	}
	http.Redirect(w, r, a.startEditURL(startPath), http.StatusSeeOther)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if cookie, err := r.Cookie("static_editor_session"); err == nil {
		_ = a.store.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "static_editor_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   a.cfg.SecureCookies,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) handleEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenant := a.tenantForHost(r.Host)
	if tenant.Domain == "" {
		http.Error(w, "unknown tenant host", http.StatusNotFound)
		return
	}

	session, ok := a.currentSession(r)
	if !ok || session.Tenant != tenant.Domain {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	targetPath := strings.TrimSpace(r.URL.Query().Get("path"))
	if targetPath == "" {
		targetPath = tenant.StartPath
	}

	resolvedPath, err := resolveStaticPath(tenant.StaticRoot, targetPath)
	if err != nil {
		http.Error(w, "invalid edit path", http.StatusBadRequest)
		return
	}

	source, err := os.ReadFile(resolvedPath)
	if err != nil {
		http.Error(w, "could not read html file", http.StatusNotFound)
		return
	}

	prepared, err := editor.PrepareDocument(string(source), tenant.MainSelector, tenant.AllowedBlockTags)
	if err != nil {
		http.Error(w, "could not prepare document for editing", http.StatusBadRequest)
		return
	}

	writeHTML(w, http.StatusOK, renderEditPage(tenant, session, targetPath, prepared, a.cfg.ContentToolsCSSURL, a.cfg.ContentToolsJSURL))
}

func (a *App) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenant, session, ok := a.requireTenantSession(w, r)
	if !ok {
		return
	}
	_ = session

	req, path, _, source, regions, ok := a.loadEditableRequest(w, r, tenant)
	if !ok {
		return
	}

	updatedHTML, err := editor.ApplyRegionsHTML(source, tenant.MainSelector, tenant.AllowedBlockTags, tenant.AllowedInlineTags, regions)
	if err != nil {
		http.Error(w, "could not build preview", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(model.PreviewResponse{
		OK:          true,
		Message:     "Preview erstellt",
		PreviewHTML: updatedHTML,
	})
	_ = req
	_ = path
}

func (a *App) handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenant, session, ok := a.requireTenantSession(w, r)
	if !ok {
		return
	}
	_ = session

	req, path, fullPath, source, regions, ok := a.loadEditableRequest(w, r, tenant)
	if !ok {
		return
	}

	updatedHTML, err := editor.ApplyRegionsHTML(source, tenant.MainSelector, tenant.AllowedBlockTags, tenant.AllowedInlineTags, regions)
	if err != nil {
		http.Error(w, "could not save document", http.StatusBadRequest)
		return
	}

	backupPath, err := backupFile(tenant.UndoBackupsRoot, path, []byte(source))
	if err != nil {
		http.Error(w, "could not create backup", http.StatusInternalServerError)
		return
	}
	if err := writeFileAtomically(fullPath, []byte(updatedHTML), 0o644); err != nil {
		http.Error(w, "could not write updated file", http.StatusInternalServerError)
		return
	}

	commitHash := ""
	pushed := false
	pushTarget := ""
	message := "Datei gespeichert"
	if a.cfg.GitCommitOnSave {
		commitHash, err = gitops.CommitFile(tenant.RepoRoot, fullPath, a.cfg.GitAuthorName, a.gitAuthorEmail(session.Email), gitCommitMessage(tenant.Domain, path, session.Email))
		if err != nil {
			http.Error(w, "file saved but git commit failed", http.StatusInternalServerError)
			return
		}

		message = "Datei gespeichert und versioniert"
		if a.cfg.GitPushOnSave {
			pushTarget, err = gitops.Push(tenant.RepoRoot, a.cfg.GitRemoteName, a.cfg.GitBranch, gitops.PushAuth{
				HTTPUsername: a.cfg.GitHTTPUsername,
				HTTPPassword: a.cfg.GitHTTPPassword,
			})
			if err != nil {
				http.Error(w, "file saved and committed, but git push failed", http.StatusInternalServerError)
				return
			}
			pushed = true
			message = "Datei gespeichert, versioniert und gepusht"
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(model.SaveResponse{
		OK:         true,
		Message:    message,
		BackupPath: backupPath,
		CommitHash: commitHash,
		Pushed:     pushed,
		PushTarget: pushTarget,
	})
	_ = req
}

func (a *App) tenantForHost(host string) model.Tenant {
	host = stripPort(host)
	for _, domain := range a.cfg.SortedTenantDomains() {
		tenant := a.cfg.Tenants[domain]
		if tenant.LoginDomain == host || tenant.Domain == host {
			return tenant
		}
		for _, alias := range tenant.Aliases {
			if alias == host {
				return tenant
			}
		}
	}
	if (host == "localhost" || host == "127.0.0.1") && len(a.cfg.Tenants) == 1 {
		for _, domain := range a.cfg.SortedTenantDomains() {
			return a.cfg.Tenants[domain]
		}
	}
	return model.Tenant{}
}

func stripPort(host string) string {
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return parsedHost
	}
	return host
}

func (a *App) currentSession(r *http.Request) (auth.Session, bool) {
	cookie, err := r.Cookie("static_editor_session")
	if err != nil {
		return auth.Session{}, false
	}
	session, err := a.store.GetSession(cookie.Value)
	if err != nil {
		return auth.Session{}, false
	}
	return session, true
}

func (a *App) verifyURL(r *http.Request, tenant model.Tenant, token string) string {
	scheme := "https"
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	} else if r.TLS == nil && stripPort(r.Host) == "localhost" {
		scheme = "http"
	}

	host := tenant.LoginDomain
	currentHost := stripPort(r.Host)
	if currentHost == "localhost" || currentHost == "127.0.0.1" {
		host = r.Host
	}

	return fmt.Sprintf("%s://%s/auth/verify?token=%s", scheme, host, token)
}

func acceptsHTML(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

func writeHTML(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func htmlEscape(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(value)
}

func (a *App) gitAuthorEmail(sessionEmail string) string {
	if configured := strings.TrimSpace(a.cfg.GitAuthorEmail); configured != "" {
		return configured
	}
	return sessionEmail
}

func (a *App) startEditURL(targetPath string) string {
	cleanTarget := strings.TrimSpace(targetPath)
	if cleanTarget == "" {
		cleanTarget = "/index.html"
	}
	return "/edit?path=" + url.QueryEscape(cleanTarget)
}

func resolveStaticPath(root, target string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", fmt.Errorf("empty static root")
	}
	if !strings.HasPrefix(target, "/") {
		return "", fmt.Errorf("path must be absolute")
	}

	cleanTarget := filepath.Clean(target)
	if cleanTarget == "." || cleanTarget == "/" {
		cleanTarget = "/index.html"
	}

	fullPath := filepath.Join(root, strings.TrimPrefix(cleanTarget, "/"))
	relative, err := filepath.Rel(root, fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve relative path: %w", err)
	}
	if strings.HasPrefix(relative, "..") {
		return "", fmt.Errorf("path traversal rejected")
	}
	return fullPath, nil
}

func renderEditPage(tenant model.Tenant, session auth.Session, targetPath string, prepared editor.PreparedDocument, contentToolsCSSURL, contentToolsJSURL string) string {
	_ = contentToolsCSSURL
	_ = contentToolsJSURL
	headInjection := fmt.Sprintf(`
  <style>
    body.static-inline-editor-active { padding-top: 5.5rem !important; }
    .static-inline-editor-bar { position: fixed; inset: 0 0 auto 0; z-index: 9999; display: flex; justify-content: space-between; gap: 1rem; align-items: center; padding: 0.9rem 1rem; border-bottom: 1px solid rgba(121,91,61,0.24); background: rgba(255,252,246,0.96); backdrop-filter: blur(8px); font: 14px/1.35 Georgia, serif; color: #2d241d; box-shadow: 0 10px 24px rgba(46,36,29,0.1); }
    .static-inline-editor-meta strong { color: #8a3c1a; }
    .static-inline-editor-actions { display: flex; gap: 0.75rem; align-items: center; }
    .static-inline-editor-actions a, .static-inline-editor-actions button { font: inherit; padding: 0.65rem 0.9rem; border-radius: 999px; border: 1px solid #8a3c1a; background: #fff7ef; color: #8a3c1a; text-decoration: none; cursor: pointer; }
    .static-inline-editor-preview { position: fixed; right: 1rem; bottom: 1rem; width: min(46rem, calc(100vw - 2rem)); max-height: 70vh; display: none; flex-direction: column; border: 1px solid rgba(121,91,61,0.24); border-radius: 18px; overflow: hidden; background: #fffdf8; box-shadow: 0 22px 48px rgba(46,36,29,0.18); z-index: 9999; }
    .static-inline-editor-preview.is-open { display: flex; }
    .static-inline-editor-preview-head { display: flex; justify-content: space-between; align-items: center; gap: 1rem; padding: 0.85rem 1rem; border-bottom: 1px solid rgba(121,91,61,0.24); background: #fbf3e8; font: 14px/1.35 Georgia, serif; }
    .static-inline-editor-preview-frame { width: 100%%; min-height: 26rem; border: 0; background: #fff; }
    [data-editable] { min-height: 1.25rem; outline: 2px dashed transparent; outline-offset: 0.2rem; transition: outline-color 120ms ease, background-color 120ms ease, box-shadow 120ms ease; color: inherit; caret-color: #ffd200; }
    [data-editable][contenteditable="true"] { cursor: text; }
    [data-editable][contenteditable="true"]:hover { outline-color: rgba(255,210,0,0.45); background: rgba(255,210,0,0.08); box-shadow: inset 0 0 0 1px rgba(255,210,0,0.08); }
    [data-editable][contenteditable="true"]:focus { outline-color: rgba(255,210,0,0.95); background: rgba(255,210,0,0.14); box-shadow: inset 0 0 0 1px rgba(255,210,0,0.16); }
  </style>`)

	bodyPrefix := fmt.Sprintf(`
  <div class="static-inline-editor-bar">
    <div class="static-inline-editor-meta">
      <strong>Edit-Modus</strong> fuer %s
      <div>Datei: <code>%s</code> | Nutzer: <code>%s</code> | Markierte Knoten: <code>%d</code></div>
    </div>
    <div class="static-inline-editor-actions">
      <a href="/">Start</a>
      <button type="button" id="preview-button">Vorschau</button>
      <button type="button" id="save-button">Speichern</button>
      <button type="button" id="preview-close" hidden>Vorschau schliessen</button>
      <form method="post" action="/auth/logout" style="margin:0">
        <button type="submit">Abmelden</button>
      </form>
    </div>
  </div>`, htmlEscape(tenant.Domain), htmlEscape(targetPath), htmlEscape(session.Email), len(prepared.EditableIDs))

	bodySuffix := fmt.Sprintf(`
  <div class="static-inline-editor-preview" id="preview-panel" aria-live="polite">
    <div class="static-inline-editor-preview-head">
      <strong>Vorschau</strong>
      <span id="preview-status">Noch keine Vorschau erzeugt.</span>
    </div>
    <iframe id="preview-frame" class="static-inline-editor-preview-frame" title="Preview"></iframe>
  </div>
  <script>
    window.addEventListener('load', function () {
      document.body.classList.add('static-inline-editor-active');

      var editPath = %q;
      var previewPanel = document.getElementById('preview-panel');
      var previewFrame = document.getElementById('preview-frame');
      var previewStatus = document.getElementById('preview-status');
      var previewButton = document.getElementById('preview-button');
      var closeButton = document.getElementById('preview-close');
      var saveButton = document.getElementById('save-button');
      var editableNodes = Array.prototype.slice.call(document.querySelectorAll('[data-editable]'));
      var latestRegions = {};

      editableNodes.forEach(function (node) {
        node.setAttribute('contenteditable', 'true');
        node.setAttribute('spellcheck', 'true');
      });

      function readError(response, fallbackMessage) {
        return response.text().then(function (text) {
          var message = (text || '').trim();
          throw new Error(message || fallbackMessage);
        });
      }

      function flash(kind) {
        previewStatus.dataset.state = kind;
      }

      function collectRegions() {
        var regions = {};
        editableNodes.forEach(function (node) {
          var name = node.getAttribute('data-name');
          if (!name) {
            return;
          }
          regions[name] = node.innerHTML;
        });
        return regions;
      }

      function requestPreview() {
        var regions = collectRegions();
        if (Object.keys(regions).length === 0) {
          window.alert('Keine bearbeitbaren Bereiche gefunden.');
          return Promise.resolve(null);
        }
        latestRegions = regions;
        return fetch('/preview', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'same-origin',
          body: JSON.stringify({ path: editPath, regions: regions })
        })
          .then(function (response) {
            if (!response.ok) {
              return readError(response, 'Preview fehlgeschlagen');
            }
            return response.json();
          })
          .then(function (payload) {
            previewPanel.classList.add('is-open');
            closeButton.hidden = false;
            previewStatus.textContent = payload.message || 'Preview erstellt';
            previewFrame.srcdoc = payload.preview_html || '';
            flash('ok');
            return payload;
          })
          .catch(function (error) {
            console.error(error);
            previewStatus.textContent = error && error.message ? error.message : 'Preview fehlgeschlagen';
            flash('no');
            throw error;
          });
      }

      previewButton.addEventListener('click', function () {
        requestPreview().catch(function () {});
      });

      closeButton.addEventListener('click', function () {
        previewPanel.classList.remove('is-open');
        closeButton.hidden = true;
      });

      saveButton.addEventListener('click', function () {
        latestRegions = collectRegions();
        fetch('/save', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'same-origin',
            body: JSON.stringify({ path: editPath, regions: latestRegions })
          })
            .then(function (response) {
              if (!response.ok) {
                return readError(response, 'Speichern fehlgeschlagen');
              }
              return response.json();
            })
            .then(function (payload) {
              previewStatus.textContent = payload.message || 'Datei gespeichert';
              flash('ok');
              window.alert('Gespeichert. Backup: ' + (payload.backup_path || 'angelegt'));
            })
            .catch(function (error) {
              console.error(error);
              previewStatus.textContent = error && error.message ? error.message : 'Speichern fehlgeschlagen';
              window.alert(error && error.message ? error.message : 'Speichern fehlgeschlagen');
              flash('no');
            });
      });
    });
  </script>
</body>`, targetPath)

	page := prepared.HTML
	page = injectIntoHead(page, headInjection)
	page = injectAfterBodyStart(page, bodyPrefix)
	page = injectBeforeBodyEnd(page, bodySuffix)
	return page
}

func injectIntoHead(source, snippet string) string {
	lower := strings.ToLower(source)
	if idx := strings.Index(lower, "</head>"); idx >= 0 {
		return source[:idx] + snippet + source[idx:]
	}
	if idx := strings.Index(lower, "<body"); idx >= 0 {
		return source[:idx] + "<head>" + snippet + "\n</head>\n" + source[idx:]
	}
	return "<head>" + snippet + "\n</head>\n" + source
}

func injectAfterBodyStart(source, snippet string) string {
	lower := strings.ToLower(source)
	idx := strings.Index(lower, "<body")
	if idx < 0 {
		return "<body>" + snippet + source + "</body>"
	}
	end := strings.Index(lower[idx:], ">")
	if end < 0 {
		return source + snippet
	}
	insertPos := idx + end + 1
	return source[:insertPos] + snippet + source[insertPos:]
}

func injectBeforeBodyEnd(source, snippet string) string {
	lower := strings.ToLower(source)
	if idx := strings.LastIndex(lower, "</body>"); idx >= 0 {
		return source[:idx] + snippet + source[idx:]
	}
	return source + snippet
}

func (a *App) requireTenantSession(w http.ResponseWriter, r *http.Request) (model.Tenant, auth.Session, bool) {
	tenant := a.tenantForHost(r.Host)
	if tenant.Domain == "" {
		http.Error(w, "unknown tenant host", http.StatusNotFound)
		return model.Tenant{}, auth.Session{}, false
	}

	session, ok := a.currentSession(r)
	if !ok || session.Tenant != tenant.Domain {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return model.Tenant{}, auth.Session{}, false
	}

	return tenant, session, true
}

func (a *App) loadEditableRequest(w http.ResponseWriter, r *http.Request, tenant model.Tenant) (model.PreviewRequest, string, string, string, map[string]string, bool) {
	var req model.PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return model.PreviewRequest{}, "", "", "", nil, false
	}

	targetPath := strings.TrimSpace(req.Path)
	if targetPath == "" {
		targetPath = tenant.StartPath
	}
	fullPath, err := resolveStaticPath(tenant.StaticRoot, targetPath)
	if err != nil {
		http.Error(w, "invalid edit path", http.StatusBadRequest)
		return model.PreviewRequest{}, "", "", "", nil, false
	}

	sourceBytes, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "could not read html file", http.StatusNotFound)
		return model.PreviewRequest{}, "", "", "", nil, false
	}

	if len(req.Regions) == 0 {
		http.Error(w, "missing editable region", http.StatusBadRequest)
		return model.PreviewRequest{}, "", "", "", nil, false
	}

	return req, targetPath, fullPath, string(sourceBytes), req.Regions, true
}

func backupFile(backupRoot, targetPath string, content []byte) (string, error) {
	if strings.TrimSpace(backupRoot) == "" {
		return "", fmt.Errorf("empty backup root")
	}
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return "", fmt.Errorf("mkdir backup root: %w", err)
	}

	base := strings.Trim(strings.ReplaceAll(targetPath, "/", "_"), "_")
	if base == "" {
		base = "index.html"
	}
	filename := fmt.Sprintf("%s_%s", time.Now().UTC().Format("2006-01-02T15-04-05"), base)
	fullPath := filepath.Join(backupRoot, filename)
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return "", fmt.Errorf("write backup file: %w", err)
	}
	return fullPath, nil
}

func writeFileAtomically(path string, content []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir target dir: %w", err)
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, content, mode); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

func gitCommitMessage(domain, targetPath, email string) string {
	return fmt.Sprintf("edit(%s): %s by %s", domain, targetPath, email)
}
