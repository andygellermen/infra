package httpapp

import (
	"html/template"
	"net/http"
	"strings"
)

var rootLandingTemplate = template.Must(template.New("eep-root").Parse(`<!DOCTYPE html>
<html lang="de">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Easy Event Planner</title>
    <style>
      body {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
        background: #f6f7fb;
        color: #1f2937;
      }
      main {
        max-width: 720px;
        margin: 0 auto;
        padding: 48px 20px 64px;
      }
      .panel {
        background: #ffffff;
        border: 1px solid #dbe1ea;
        border-radius: 16px;
        padding: 28px;
        box-shadow: 0 12px 30px rgba(15, 23, 42, 0.08);
      }
      h1 {
        margin: 0 0 12px;
        font-size: 2rem;
      }
      p {
        line-height: 1.6;
      }
      ul {
        padding-left: 20px;
      }
      a {
        color: #0f62fe;
      }
      code {
        background: #eef2ff;
        border-radius: 6px;
        padding: 2px 6px;
      }
    </style>
  </head>
  <body>
    <main>
      <div class="panel">
        <h1>Easy Event Planner</h1>
        <p>Der Dienst ist erreichbar. Der öffentliche Event-Bereich ist mandantenbezogen und wird über tenant-spezifische Pfade oder Snippets ausgeliefert.</p>
        <ul>
          <li><a href="/admin">Admin-Oberfläche</a></li>
          <li><a href="/version">Version</a></li>
          <li><a href="/healthz">Healthcheck</a></li>
        </ul>
        <p>Öffentliche API-Routen folgen dem Muster <code>/api/v1/public/{tenantSlug}/...</code>.</p>
        {{if .BaseURL}}<p>Konfigurierte Basis-URL: <code>{{.BaseURL}}</code></p>{{end}}
      </div>
    </main>
  </body>
</html>
`))

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = rootLandingTemplate.Execute(w, struct {
		BaseURL string
	}{
		BaseURL: strings.TrimSpace(a.cfg.BaseURL),
	})
}
