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
		http.NotFound(w, r)
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
		if err := a.mailer.SendMagicLink(strings.TrimSpace(req.Email), verifyURL); err != nil {
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

	req, path, _, source, regionHTML, ok := a.loadEditableRequest(w, r, tenant)
	if !ok {
		return
	}

	updatedHTML, err := editor.ApplyRegionHTML(source, tenant.MainSelector, tenant.AllowedBlockTags, tenant.AllowedInlineTags, regionHTML)
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

	req, path, fullPath, source, regionHTML, ok := a.loadEditableRequest(w, r, tenant)
	if !ok {
		return
	}

	updatedHTML, err := editor.ApplyRegionHTML(source, tenant.MainSelector, tenant.AllowedBlockTags, tenant.AllowedInlineTags, regionHTML)
	if err != nil {
		http.Error(w, "could not save document", http.StatusBadRequest)
		return
	}

	backupPath, err := backupFile(tenant.BackupRoot, path, []byte(source))
	if err != nil {
		http.Error(w, "could not create backup", http.StatusInternalServerError)
		return
	}
	if err := writeFileAtomically(fullPath, []byte(updatedHTML), 0o644); err != nil {
		http.Error(w, "could not write updated file", http.StatusInternalServerError)
		return
	}
	commitHash, err := gitops.CommitFile(tenant.RepoRoot, fullPath, a.cfg.GitAuthorName, a.gitAuthorEmail(session.Email), gitCommitMessage(tenant.Domain, path, session.Email))
	if err != nil {
		http.Error(w, "file saved but git commit failed", http.StatusInternalServerError)
		return
	}
	pushed := false
	pushTarget := ""
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
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(model.SaveResponse{
		OK:         true,
		Message:    "Datei gespeichert",
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
	return fmt.Sprintf(`<!doctype html>
<html lang="de">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Edit %s</title>
  <link rel="stylesheet" href="%s">
  <style>
    :root { --ink:#2d241d; --paper:#f7f0e6; --accent:#8a3c1a; --line:#d7c8b7; }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: Georgia, serif; color: var(--ink); background: linear-gradient(180deg, #fcfaf6 0%%, #f1e8da 100%%); }
    .bar { position: sticky; top: 0; z-index: 9; display: flex; justify-content: space-between; gap: 1rem; align-items: center; padding: 1rem 1.25rem; border-bottom: 1px solid var(--line); background: rgba(255,252,246,0.96); backdrop-filter: blur(8px); }
    .meta { font-size: 0.95rem; }
    .meta strong { color: var(--accent); }
    .actions { display: flex; gap: 0.75rem; align-items: center; }
    .actions a, .actions button { font: inherit; padding: 0.7rem 0.95rem; border-radius: 999px; border: 1px solid var(--accent); background: #fff7ef; color: var(--accent); text-decoration: none; cursor: pointer; }
    .hint { max-width: 1100px; margin: 1rem auto 0; padding: 0 1rem; color: #64584c; }
    .canvas { max-width: 1100px; margin: 1rem auto 3rem; padding: 0 1rem 2rem; }
    .frame { background: white; border: 1px solid var(--line); border-radius: 18px; overflow: hidden; box-shadow: 0 18px 40px rgba(84,56,28,0.08); }
    .preview { position: fixed; right: 1rem; bottom: 1rem; width: min(46rem, calc(100vw - 2rem)); max-height: 70vh; display: none; flex-direction: column; border: 1px solid var(--line); border-radius: 18px; overflow: hidden; background: #fffdf8; box-shadow: 0 22px 48px rgba(46,36,29,0.18); z-index: 30; }
    .preview.is-open { display: flex; }
    .preview-head { display: flex; justify-content: space-between; align-items: center; gap: 1rem; padding: 0.85rem 1rem; border-bottom: 1px solid var(--line); background: #fbf3e8; }
    .preview-frame { width: 100%%; min-height: 26rem; border: 0; background: #fff; }
    [data-editor-id] { outline: 2px dashed rgba(138,60,26,0.24); outline-offset: 0.16rem; }
    [data-editable] { position: relative; }
    [data-editable]::before { content: "Editable region"; position: absolute; top: 0.35rem; right: 0.5rem; font: 600 0.72rem/1 system-ui, sans-serif; letter-spacing: 0.04em; text-transform: uppercase; color: #8a3c1a; background: rgba(255,247,239,0.92); border: 1px solid rgba(138,60,26,0.24); border-radius: 999px; padding: 0.28rem 0.5rem; }
    .ct-app .ct-widget.ct-ignition { top: 5.6rem; left: 1rem; }
  </style>
</head>
<body>
  <div class="bar">
    <div class="meta">
      <strong>Edit-Modus</strong> fuer %s
      <div>Datei: <code>%s</code> | Nutzer: <code>%s</code> | Markierte Knoten: <code>%d</code></div>
    </div>
    <div class="actions">
      <a href="/">Start</a>
      <button type="button" id="preview-close" hidden>Vorschau schliessen</button>
      <button type="button" id="save-button" hidden>Speichern</button>
      <form method="post" action="/auth/logout" style="margin:0">
        <button type="submit">Abmelden</button>
      </form>
    </div>
  </div>
  <div class="hint">
    Der erste Edit-Call markiert bereits erlaubte Textcontainer mit <code>data-editor-id</code>. Als naechstes haengen wir den eigentlichen Inline-Editor, Vorschau und Speichern daran.
  </div>
  <div class="canvas">
    <div class="frame">%s</div>
  </div>
  <div class="preview" id="preview-panel" aria-live="polite">
    <div class="preview-head">
      <strong>Vorschau</strong>
      <span id="preview-status">Noch keine Vorschau erzeugt.</span>
    </div>
    <iframe id="preview-frame" class="preview-frame" title="Preview"></iframe>
  </div>
  <script src="%s"></script>
  <script>
    window.addEventListener('load', function () {
      if (!window.ContentTools) {
        console.warn('ContentTools konnte nicht geladen werden');
        return;
      }

      var editPath = %q;
      var previewPanel = document.getElementById('preview-panel');
      var previewFrame = document.getElementById('preview-frame');
      var previewStatus = document.getElementById('preview-status');
      var closeButton = document.getElementById('preview-close');
      var saveButton = document.getElementById('save-button');
      var latestRegions = null;
      var editor = ContentTools.EditorApp.get();
      editor.init('[data-editable]', 'data-name');
      editor.addEventListener('saved', function (ev) {
        var regions = ev.detail().regions || {};
        if (Object.keys(regions).length === 0) {
          return;
        }
        latestRegions = regions;
        fetch('/preview', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'same-origin',
          body: JSON.stringify({ path: editPath, regions: regions })
        })
          .then(function (response) {
            if (!response.ok) {
              throw new Error('Preview fehlgeschlagen');
            }
            return response.json();
          })
          .then(function (payload) {
            previewPanel.classList.add('is-open');
            closeButton.hidden = false;
            saveButton.hidden = false;
            previewStatus.textContent = payload.message || 'Preview erstellt';
            previewFrame.srcdoc = payload.preview_html || '';
            new ContentTools.FlashUI('ok');
          })
          .catch(function (error) {
            console.error(error);
            previewStatus.textContent = 'Preview fehlgeschlagen';
            new ContentTools.FlashUI('no');
          });
      });

      closeButton.addEventListener('click', function () {
        previewPanel.classList.remove('is-open');
        closeButton.hidden = true;
        saveButton.hidden = true;
      });

      saveButton.addEventListener('click', function () {
        if (!latestRegions) {
          window.alert('Bitte zuerst ueber den ContentTools-Save-Button eine Vorschau erzeugen.');
          return;
        }
        fetch('/save', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'same-origin',
          body: JSON.stringify({ path: editPath, regions: latestRegions })
        })
          .then(function (response) {
            if (!response.ok) {
              throw new Error('Speichern fehlgeschlagen');
            }
            return response.json();
          })
          .then(function (payload) {
            previewStatus.textContent = payload.message || 'Datei gespeichert';
            new ContentTools.FlashUI('ok');
            window.alert('Gespeichert. Backup: ' + (payload.backup_path || 'angelegt'));
          })
          .catch(function (error) {
            console.error(error);
            previewStatus.textContent = 'Speichern fehlgeschlagen';
            new ContentTools.FlashUI('no');
          });
      });
    });
  </script>
</body>
</html>`, htmlEscape(targetPath), htmlEscape(contentToolsCSSURL), htmlEscape(tenant.Domain), htmlEscape(targetPath), htmlEscape(session.Email), len(prepared.EditableIDs), prepared.HTML, htmlEscape(contentToolsJSURL), targetPath)
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

func (a *App) loadEditableRequest(w http.ResponseWriter, r *http.Request, tenant model.Tenant) (model.PreviewRequest, string, string, string, string, bool) {
	var req model.PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return model.PreviewRequest{}, "", "", "", "", false
	}

	targetPath := strings.TrimSpace(req.Path)
	if targetPath == "" {
		targetPath = tenant.StartPath
	}
	fullPath, err := resolveStaticPath(tenant.StaticRoot, targetPath)
	if err != nil {
		http.Error(w, "invalid edit path", http.StatusBadRequest)
		return model.PreviewRequest{}, "", "", "", "", false
	}

	sourceBytes, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "could not read html file", http.StatusNotFound)
		return model.PreviewRequest{}, "", "", "", "", false
	}

	regionHTML, ok := req.Regions["main-content"]
	if !ok {
		http.Error(w, "missing editable region", http.StatusBadRequest)
		return model.PreviewRequest{}, "", "", "", "", false
	}

	return req, targetPath, fullPath, string(sourceBytes), regionHTML, true
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
