package httpapp

import (
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
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

var publicEventPageTemplate = template.Must(template.New("eep-public-event").Parse(`<!DOCTYPE html>
<html lang="de">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{.Title}}</title>
    <style>
      body {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
        background: #f6f1e7;
        color: #18313d;
      }
      main {
        max-width: 920px;
        margin: 0 auto;
        padding: 32px 18px 64px;
      }
      .shell {
        display: grid;
        gap: 20px;
      }
      .panel {
        background: #fff;
        border: 1px solid #d9e1e8;
        border-radius: 20px;
        padding: 24px;
        box-shadow: 0 16px 34px rgba(24, 49, 61, 0.08);
      }
      .eyebrow {
        font-size: 0.92rem;
        color: #587180;
        margin-bottom: 10px;
      }
      h1 {
        margin: 0 0 10px;
        font-size: clamp(2rem, 4vw, 3rem);
        line-height: 1.05;
      }
      .subtitle {
        margin: 0 0 12px;
        font-size: 1.05rem;
        color: #43606f;
      }
      .meta {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
        margin: 0 0 16px;
        padding: 0;
        list-style: none;
      }
      .meta li {
        padding: 8px 12px;
        border-radius: 999px;
        background: #eef5f8;
        color: #244657;
        font-size: 0.92rem;
        font-weight: 600;
      }
      .description {
        white-space: pre-wrap;
        line-height: 1.7;
      }
      .cta {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        color: #0f5ca8;
        text-decoration: none;
        font-weight: 700;
      }
      #eep-registration {
        min-height: 80px;
      }
    </style>
  </head>
  <body>
    <main>
      <div class="shell">
        <section class="panel">
          <div class="eyebrow">{{.TenantName}}</div>
          <h1>{{.Title}}</h1>
          {{if .Subtitle}}<p class="subtitle">{{.Subtitle}}</p>{{end}}
          <ul class="meta">
            {{if .StartsAt}}<li>{{.StartsAt}}</li>{{end}}
            {{if .Location}}<li>{{.Location}}</li>{{end}}
            {{if .Mode}}<li>{{.Mode}}</li>{{end}}
          </ul>
          {{if .Description}}<div class="description">{{.Description}}</div>{{end}}
        </section>
        <section class="panel">
          <div class="eyebrow">Anmeldung</div>
          <div id="eep-registration"></div>
          <script src="{{.RegisterScriptSrc}}" data-target="#eep-registration" defer></script>
        </section>
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

func (a *App) handleTenantPublicEventPage(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, eventSlug string) {
	if a.eventRepo == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	item, err := a.eventRepo.GetPublicEventBySlug(r.Context(), tenantItem.ID, eventSlug)
	if err != nil {
		if errors.Is(err, event.ErrEventNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "event unavailable", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = publicEventPageTemplate.Execute(w, struct {
		TenantName        string
		Title             string
		Subtitle          string
		Description       string
		StartsAt          string
		Location          string
		Mode              string
		RegisterScriptSrc string
	}{
		TenantName:        strings.TrimSpace(tenantItem.Name),
		Title:             strings.TrimSpace(item.Title),
		Subtitle:          strings.TrimSpace(item.Subtitle),
		Description:       strings.TrimSpace(item.Description),
		StartsAt:          formatPublicEventPageDate(item.StartsAt, item.Timezone),
		Location:          publicEventPageLocation(item),
		Mode:              publicEventPageMode(item.ParticipationMode),
		RegisterScriptSrc: buildRegistrationEmbedScriptSrc(a.cfg.BaseURL, tenantItem.Slug, item.Slug),
	})
}

func formatPublicEventPageDate(startsAt time.Time, timezone string) string {
	location := time.UTC
	if strings.TrimSpace(timezone) != "" {
		if loaded, err := time.LoadLocation(strings.TrimSpace(timezone)); err == nil {
			location = loaded
		}
	}
	return startsAt.In(location).Format("02.01.2006 · 15:04")
}

func publicEventPageLocation(item event.PublicEvent) string {
	if value := strings.TrimSpace(item.LocationName); value != "" {
		return value
	}
	switch strings.TrimSpace(strings.ToLower(item.ParticipationMode)) {
	case "online":
		return "Online"
	case "hybrid":
		return "Hybrid"
	case "onsite":
		return "Vor Ort"
	default:
		return ""
	}
}

func publicEventPageMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "online":
		return "Online"
	case "hybrid":
		return "Hybrid"
	case "onsite":
		return "Vor Ort"
	default:
		return ""
	}
}
