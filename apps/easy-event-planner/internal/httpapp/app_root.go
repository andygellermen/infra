package httpapp

import (
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/registration"
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
      .actions {
        display: flex;
        flex-wrap: wrap;
        gap: 14px;
        margin-top: 20px;
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
          <div class="actions">
            {{if .OverviewURL}}<a class="cta" href="{{.OverviewURL}}">Zur Veranstaltungs-Übersicht</a>{{end}}
          </div>
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

var publicEventsOverviewTemplate = template.Must(template.New("eep-public-events-overview").Parse(`<!DOCTYPE html>
<html lang="de">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{.TenantName}} · Veranstaltungs-Übersicht</title>
    <style>
      :root {
        --page-bg: #f4efe6;
        --ink: #17303b;
        --muted: #5d7582;
        --panel: #fffdf8;
        --panel-border: #d8e0e5;
        --accent: #0f5ca8;
        --accent-soft: #e8f1fa;
        --ok: #0f766e;
        --ok-soft: #dff5f0;
        --warn: #9a3412;
        --warn-soft: #fde7db;
      }
      * {
        box-sizing: border-box;
      }
      body {
        margin: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
        background:
          radial-gradient(circle at top left, rgba(17, 92, 168, 0.10), transparent 32%),
          linear-gradient(180deg, #f9f5ee 0%, var(--page-bg) 100%);
        color: var(--ink);
      }
      main {
        max-width: 1080px;
        margin: 0 auto;
        padding: 28px 18px 64px;
      }
      .shell {
        display: grid;
        gap: 18px;
      }
      .hero,
      .panel {
        background: var(--panel);
        border: 1px solid var(--panel-border);
        border-radius: 24px;
        box-shadow: 0 18px 40px rgba(23, 48, 59, 0.08);
      }
      .hero {
        padding: 28px;
      }
      .eyebrow {
        margin: 0 0 12px;
        color: var(--muted);
        font-size: 0.92rem;
      }
      h1 {
        margin: 0 0 12px;
        font-size: clamp(2.1rem, 5vw, 3.8rem);
        line-height: 0.98;
        letter-spacing: -0.04em;
      }
      .hero-copy {
        max-width: 760px;
        margin: 0;
        line-height: 1.65;
        color: #35505e;
      }
      .summary {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
        margin-top: 18px;
      }
      .summary span,
      .filter-link {
        display: inline-flex;
        align-items: center;
        min-height: 38px;
        padding: 0 14px;
        border-radius: 999px;
        font-size: 0.92rem;
        font-weight: 700;
        text-decoration: none;
      }
      .summary span {
        background: #edf4f7;
        color: #284858;
      }
      .panel {
        padding: 22px;
      }
      .filters {
        display: grid;
        gap: 14px;
      }
      .filter-group {
        display: grid;
        gap: 8px;
      }
      .filter-label {
        font-size: 0.88rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: var(--muted);
      }
      .filter-links {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
      }
      .filter-link {
        background: #f3f6f8;
        color: #35505e;
        border: 1px solid transparent;
      }
      .filter-link.is-active {
        background: var(--accent-soft);
        color: var(--accent);
        border-color: rgba(15, 92, 168, 0.18);
      }
      .events-grid {
        display: grid;
        gap: 16px;
      }
      .event-card {
        display: grid;
        gap: 14px;
        padding: 22px;
        border-radius: 22px;
        border: 1px solid var(--panel-border);
        background: #ffffff;
      }
      .event-card-head {
        display: flex;
        flex-wrap: wrap;
        justify-content: space-between;
        gap: 12px;
      }
      .event-title {
        margin: 0;
        font-size: 1.45rem;
        line-height: 1.08;
      }
      .event-title a {
        color: inherit;
        text-decoration: none;
      }
      .event-title a:hover {
        color: var(--accent);
      }
      .event-subtitle {
        margin: 8px 0 0;
        color: #466271;
      }
      .event-meta,
      .event-tags {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;
      }
      .event-meta span,
      .event-tags span {
        display: inline-flex;
        align-items: center;
        min-height: 34px;
        padding: 0 12px;
        border-radius: 999px;
        background: #eef4f7;
        color: #274655;
        font-size: 0.92rem;
        font-weight: 700;
      }
      .event-tags .status-warning {
        background: var(--warn-soft);
        color: var(--warn);
      }
      .event-tags .status-ok {
        background: var(--ok-soft);
        color: var(--ok);
      }
      .event-description {
        margin: 0;
        line-height: 1.65;
        color: #35505e;
        white-space: pre-wrap;
      }
      .event-actions {
        display: flex;
        flex-wrap: wrap;
        gap: 14px;
      }
      .event-link {
        color: var(--accent);
        font-weight: 700;
        text-decoration: none;
      }
      .empty-state {
        padding: 28px;
        border-radius: 20px;
        background: #fbfcfd;
        border: 1px dashed #cad5dc;
      }
      .empty-state h2 {
        margin: 0 0 8px;
        font-size: 1.35rem;
      }
      .empty-state p {
        margin: 0;
        line-height: 1.6;
        color: #4f6774;
      }
      @media (max-width: 720px) {
        .hero,
        .panel,
        .event-card {
          padding: 18px;
        }
      }
    </style>
  </head>
  <body>
    <main>
      <div class="shell">
        <section class="hero">
          <p class="eyebrow">{{.TenantName}}</p>
          <h1>Veranstaltungs-Übersicht</h1>
          <p class="hero-copy">Hier findest du alle aktuell freigegebenen Veranstaltungen. Nutze die Kategorien und Formate als schnelle Orientierung, wenn du einfach erst einmal stöbern möchtest.</p>
          <div class="summary">
            <span>{{.SummaryLabel}}</span>
            {{if .IncludePast}}<span>Vergangene Termine werden mit angezeigt</span>{{end}}
          </div>
        </section>

        <section class="panel">
          <div class="filters">
            <div class="filter-group">
              <div class="filter-label">Ansicht</div>
              <div class="filter-links">
                {{range .OverviewLinks}}
                  <a class="filter-link{{if .Active}} is-active{{end}}" href="{{.URL}}">{{.Label}}</a>
                {{end}}
              </div>
            </div>
            {{if .CategoryLinks}}
            <div class="filter-group">
              <div class="filter-label">Kategorien</div>
              <div class="filter-links">
                {{range .CategoryLinks}}
                  <a class="filter-link{{if .Active}} is-active{{end}}" href="{{.URL}}">{{.Label}}</a>
                {{end}}
              </div>
            </div>
            {{end}}
            <div class="filter-group">
              <div class="filter-label">Format</div>
              <div class="filter-links">
                {{range .ModeLinks}}
                  <a class="filter-link{{if .Active}} is-active{{end}}" href="{{.URL}}">{{.Label}}</a>
                {{end}}
              </div>
            </div>
          </div>
        </section>

        <section class="events-grid">
          {{if .Events}}
            {{range .Events}}
            <article class="event-card">
              <div class="event-card-head">
                <div>
                  <h2 class="event-title"><a href="{{.DetailURL}}">{{.Title}}</a></h2>
                  {{if .Subtitle}}<p class="event-subtitle">{{.Subtitle}}</p>{{end}}
                </div>
                <div class="event-tags">
                  {{if .Category}}<span>{{.Category}}</span>{{end}}
                  {{if .StatusLabel}}<span class="{{.StatusClass}}">{{.StatusLabel}}</span>{{end}}
                  {{if .RegistrationLabel}}<span class="{{.RegistrationClass}}">{{.RegistrationLabel}}</span>{{end}}
                </div>
              </div>
              <div class="event-meta">
                {{if .StartsAt}}<span>{{.StartsAt}}</span>{{end}}
                {{if .Location}}<span>{{.Location}}</span>{{end}}
                {{if .Mode}}<span>{{.Mode}}</span>{{end}}
              </div>
              {{if .Description}}<p class="event-description">{{.Description}}</p>{{end}}
              <div class="event-actions">
                <a class="event-link" href="{{.DetailURL}}">Details & Anmeldung</a>
              </div>
            </article>
            {{end}}
          {{else}}
            <div class="empty-state">
              <h2>{{.EmptyTitle}}</h2>
              <p>{{.EmptyMessage}}</p>
            </div>
          {{end}}
        </section>
      </div>
    </main>
  </body>
</html>
`))

var publicRegistrationVerifyTemplate = template.Must(template.New("eep-public-registration-verify").Parse(`<!DOCTYPE html>
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
        max-width: 760px;
        margin: 0 auto;
        padding: 40px 18px 64px;
      }
      .panel {
        background: #fff;
        border: 1px solid #d9e1e8;
        border-radius: 22px;
        padding: 28px;
        box-shadow: 0 16px 34px rgba(24, 49, 61, 0.08);
      }
      .eyebrow {
        font-size: 0.92rem;
        color: #587180;
        margin-bottom: 10px;
      }
      h1 {
        margin: 0 0 12px;
        font-size: clamp(2rem, 4vw, 2.7rem);
        line-height: 1.08;
      }
      p {
        margin: 0 0 14px;
        line-height: 1.65;
      }
      .meta {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
        margin: 18px 0 22px;
      }
      .meta span {
        padding: 8px 12px;
        border-radius: 999px;
        background: #eef5f8;
        color: #244657;
        font-size: 0.92rem;
        font-weight: 600;
      }
      .actions {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        margin-top: 22px;
      }
      .button {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-height: 48px;
        padding: 0 18px;
        border-radius: 999px;
        text-decoration: none;
        font-weight: 700;
      }
      .button-primary {
        background: #159a9c;
        color: #fff;
      }
      .button-secondary {
        background: #eef5f8;
        color: #244657;
      }
    </style>
  </head>
  <body>
    <main>
      <section class="panel">
        <div class="eyebrow">{{.TenantName}}</div>
        <h1>{{.Title}}</h1>
        <p>{{.Message}}</p>
        {{if .EventTitle}}<p><strong>{{.EventTitle}}</strong></p>{{end}}
        {{if .StatusLabel}}
          <div class="meta">
            <span>{{.StatusLabel}}</span>
            {{if .ConfirmedAt}}<span>{{.ConfirmedAt}}</span>{{end}}
          </div>
        {{end}}
        <div class="actions">
          {{if .CalendarURL}}<a class="button button-primary" href="{{.CalendarURL}}">Kalendereintrag laden</a>{{end}}
          {{if .EventURL}}<a class="button button-secondary" href="{{.EventURL}}">Zur Veranstaltungsseite</a>{{end}}
        </div>
      </section>
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

type publicOverviewFilterLink struct {
	Label  string
	URL    string
	Active bool
}

type publicOverviewEventCard struct {
	Title             string
	Subtitle          string
	Category          string
	StartsAt          string
	Location          string
	Mode              string
	Description       string
	DetailURL         string
	StatusLabel       string
	StatusClass       string
	RegistrationLabel string
	RegistrationClass string
}

func (a *App) handleTenantPublicEventsOverviewPage(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if a.eventRepo == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	filter, includePast, err := publicOverviewFilterFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	items, err := a.eventRepo.ListPublicEvents(r.Context(), tenantItem.ID, filter)
	if err != nil {
		http.Error(w, "events unavailable", http.StatusBadRequest)
		return
	}

	seriesItems, err := a.eventRepo.ListPublicSeries(r.Context(), tenantItem.ID)
	if err != nil {
		http.Error(w, "series unavailable", http.StatusInternalServerError)
		return
	}

	overviewURL := buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, "", "", includePast)
	if overviewURL == "" {
		overviewURL = "/"
	}

	pageEvents := make([]publicOverviewEventCard, 0, len(items))
	for _, item := range items {
		pageEvents = append(pageEvents, publicOverviewEventPayload(tenantItem, item))
	}

	categoryLinks := []publicOverviewFilterLink{
		{
			Label:  "Alle Kategorien",
			URL:    buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, "", filter.Mode, includePast),
			Active: strings.TrimSpace(filter.SeriesSlug) == "",
		},
	}
	for _, seriesItem := range seriesItems {
		categoryLinks = append(categoryLinks, publicOverviewFilterLink{
			Label:  strings.TrimSpace(seriesItem.Title),
			URL:    buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, seriesItem.Slug, filter.Mode, includePast),
			Active: strings.TrimSpace(filter.SeriesSlug) == strings.TrimSpace(seriesItem.Slug),
		})
	}

	modeLinks := []publicOverviewFilterLink{
		{
			Label:  "Alle Formate",
			URL:    buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, filter.SeriesSlug, "", includePast),
			Active: strings.TrimSpace(filter.Mode) == "",
		},
		{
			Label:  "Vor Ort",
			URL:    buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, filter.SeriesSlug, event.ParticipationModeOnsite, includePast),
			Active: strings.TrimSpace(filter.Mode) == event.ParticipationModeOnsite,
		},
		{
			Label:  "Online",
			URL:    buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, filter.SeriesSlug, event.ParticipationModeOnline, includePast),
			Active: strings.TrimSpace(filter.Mode) == event.ParticipationModeOnline,
		},
		{
			Label:  "Hybrid",
			URL:    buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, filter.SeriesSlug, event.ParticipationModeHybrid, includePast),
			Active: strings.TrimSpace(filter.Mode) == event.ParticipationModeHybrid,
		},
	}

	overviewLinks := []publicOverviewFilterLink{
		{
			Label:  "Bevorstehende Termine",
			URL:    buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, filter.SeriesSlug, filter.Mode, false),
			Active: !includePast,
		},
		{
			Label:  "Mit vergangenen Terminen",
			URL:    buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, filter.SeriesSlug, filter.Mode, true),
			Active: includePast,
		},
	}

	emptyTitle := "Aktuell sind keine freigegebenen Veranstaltungen sichtbar."
	emptyMessage := "Sobald neue Termine freigegeben sind, erscheinen sie hier automatisch."
	if strings.TrimSpace(filter.SeriesSlug) != "" || strings.TrimSpace(filter.Mode) != "" {
		emptyTitle = "Fuer diese Auswahl wurden keine passenden Veranstaltungen gefunden."
		emptyMessage = "Probiere eine andere Kategorie oder ein anderes Format, um weitere freigegebene Termine zu sehen."
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = publicEventsOverviewTemplate.Execute(w, struct {
		TenantName    string
		SummaryLabel  string
		IncludePast   bool
		OverviewLinks []publicOverviewFilterLink
		CategoryLinks []publicOverviewFilterLink
		ModeLinks     []publicOverviewFilterLink
		Events        []publicOverviewEventCard
		EmptyTitle    string
		EmptyMessage  string
	}{
		TenantName:    strings.TrimSpace(tenantItem.Name),
		SummaryLabel:  publicOverviewSummaryLabel(len(pageEvents), includePast),
		IncludePast:   includePast,
		OverviewLinks: overviewLinks,
		CategoryLinks: categoryLinks,
		ModeLinks:     modeLinks,
		Events:        pageEvents,
		EmptyTitle:    emptyTitle,
		EmptyMessage:  emptyMessage,
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
		OverviewURL       string
		RegisterScriptSrc string
	}{
		TenantName:        strings.TrimSpace(tenantItem.Name),
		Title:             strings.TrimSpace(item.Title),
		Subtitle:          strings.TrimSpace(item.Subtitle),
		Description:       strings.TrimSpace(item.Description),
		StartsAt:          formatPublicEventPageDate(item.StartsAt, item.Timezone),
		Location:          publicEventPageLocation(item),
		Mode:              publicEventPageMode(item.ParticipationMode),
		OverviewURL:       buildTenantPublicOverviewURL(tenantItem.PublicBaseURL, tenantItem.Slug, "", "", false),
		RegisterScriptSrc: buildRegistrationEmbedScriptSrc(tenantItem.PublicBaseURL, tenantItem.Slug, item.Slug),
	})
}

func publicOverviewFilterFromRequest(r *http.Request) (event.PublicEventFilter, bool, error) {
	query := r.URL.Query()
	includePast := parseBoolQueryValue(query.Get("include_past"))
	filter := event.PublicEventFilter{
		IncludePast: includePast,
		SeriesSlug:  strings.TrimSpace(query.Get("series")),
		Mode:        strings.TrimSpace(query.Get("mode")),
		Limit:       event.MaxPublicEventLimit,
	}
	return filter, includePast, nil
}

func publicOverviewEventPayload(tenantItem tenant.Tenant, item event.PublicEvent) publicOverviewEventCard {
	statusLabel, statusClass := publicOverviewStatus(item.Status)
	registrationLabel, registrationClass := publicOverviewRegistration(item.RegistrationEnabled)

	return publicOverviewEventCard{
		Title:             strings.TrimSpace(item.Title),
		Subtitle:          strings.TrimSpace(item.Subtitle),
		Category:          strings.TrimSpace(item.SeriesTitle),
		StartsAt:          formatPublicEventPageDate(item.StartsAt, item.Timezone),
		Location:          publicEventPageLocation(item),
		Mode:              publicEventPageMode(item.ParticipationMode),
		Description:       strings.TrimSpace(item.Description),
		DetailURL:         buildPublicEventPageURL(tenantItem.PublicBaseURL, tenantItem.Slug, item.Slug),
		StatusLabel:       statusLabel,
		StatusClass:       statusClass,
		RegistrationLabel: registrationLabel,
		RegistrationClass: registrationClass,
	}
}

func publicOverviewStatus(value string) (label, className string) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case event.EventStatusChanged:
		return "Aktualisiert", "status-warning"
	case event.EventStatusPostponed:
		return "Verschoben", "status-warning"
	case event.EventStatusCancelled:
		return "Abgesagt", "status-warning"
	case event.EventStatusCompleted:
		return "Abgeschlossen", "status-ok"
	default:
		return "", ""
	}
}

func publicOverviewRegistration(enabled bool) (label, className string) {
	if enabled {
		return "Anmeldung offen", "status-ok"
	}
	return "Anmeldung geschlossen", "status-warning"
}

func publicOverviewSummaryLabel(total int, includePast bool) string {
	if includePast {
		if total == 1 {
			return "1 freigegebene Veranstaltung"
		}
		return strconv.Itoa(total) + " freigegebene Veranstaltungen"
	}
	if total == 1 {
		return "1 bevorstehende freigegebene Veranstaltung"
	}
	return strconv.Itoa(total) + " bevorstehende freigegebene Veranstaltungen"
}

func buildTenantPublicOverviewURL(baseURL, tenantSlug, seriesSlug, mode string, includePast bool) string {
	base := buildTenantPublicAssetBaseURL(baseURL, tenantSlug)
	if base == "" {
		return ""
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}
	query := url.Values{}
	if strings.TrimSpace(seriesSlug) != "" {
		query.Set("series", strings.TrimSpace(seriesSlug))
	}
	if strings.TrimSpace(mode) != "" {
		query.Set("mode", strings.TrimSpace(mode))
	}
	if includePast {
		query.Set("include_past", "true")
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func parseBoolQueryValue(raw string) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "1", "true", "yes", "ja", "on":
		return true
	default:
		return false
	}
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

func (a *App) renderPublicRegistrationVerifyPage(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant, result registration.VerifyResult, verifyErr error) {
	data := struct {
		TenantName  string
		Title       string
		Message     string
		EventTitle  string
		StatusLabel string
		ConfirmedAt string
		CalendarURL string
		EventURL    string
	}{
		TenantName: strings.TrimSpace(tenantItem.Name),
	}

	statusCode := http.StatusOK
	if verifyErr != nil {
		statusCode = http.StatusBadRequest
		data.Title, data.Message = publicRegistrationVerifyErrorCopy(verifyErr)
	} else {
		data.Title = "Anmeldung bestaetigt"
		data.Message = "Deine Anmeldung ist jetzt aktiv. Wenn du magst, kannst du den Termin direkt in deinen Kalender uebernehmen."
		switch strings.TrimSpace(result.Status) {
		case registration.StatusWaitlist:
			data.Title = "Warteliste erfolgreich"
			data.Message = "Du stehst jetzt auf der Warteliste. Sobald ein Platz frei wird, geht es von dort fuer dich weiter."
			data.StatusLabel = "Warteliste"
		case registration.StatusConfirmed:
			data.StatusLabel = "Teilnahme bestaetigt"
		default:
			data.StatusLabel = strings.TrimSpace(result.Status)
		}
		if result.ConfirmedAt != nil {
			data.ConfirmedAt = result.ConfirmedAt.UTC().Format("02.01.2006 · 15:04 UTC")
		}
		if a.eventRepo != nil && strings.TrimSpace(result.EventID) != "" {
			if item, err := a.eventRepo.GetEventByID(r.Context(), tenantItem.ID, result.EventID); err == nil {
				data.EventTitle = strings.TrimSpace(item.Title)
				data.EventURL = buildPublicEventPageURL(tenantItem.PublicBaseURL, tenantItem.Slug, item.Slug)
			}
		}
		if a.calendarService != nil && strings.TrimSpace(result.RegistrationID) != "" && strings.TrimSpace(result.ParticipantID) != "" {
			data.CalendarURL = a.calendarService.ParticipantCalendarURL(tenantItem.Slug, tenantItem.ID, result.RegistrationID, result.ParticipantID)
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = publicRegistrationVerifyTemplate.Execute(w, data)
}

func publicRegistrationVerifyErrorCopy(err error) (title, message string) {
	switch {
	case errors.Is(err, registration.ErrInvalidVerificationToken), errors.Is(err, registration.ErrRegistrationVerificationNil):
		return "Magic-Link ungueltig", "Der Link konnte nicht erkannt werden. Bitte fordere bei Bedarf einen neuen Link an."
	case errors.Is(err, registration.ErrExpiredVerificationToken):
		return "Magic-Link abgelaufen", "Der Link ist leider abgelaufen. Bitte starte die Anmeldung erneut, damit wir dir einen frischen Link senden koennen."
	case errors.Is(err, registration.ErrRegistrationState):
		return "Link bereits verwendet", "Die Anmeldung wurde bereits verarbeitet. Wenn du unsicher bist, pruefe bitte deine bestaetigte E-Mail oder fordere bei Bedarf einen neuen Link an."
	case errors.Is(err, registration.ErrEventFull):
		return "Veranstaltung ausgebucht", "Die Veranstaltung ist aktuell ausgebucht. Falls eine Warteliste aktiv ist, fuehren wir dich dort weiter."
	default:
		return "Anmeldung aktuell nicht moeglich", "Der Link konnte gerade nicht verarbeitet werden. Bitte versuche es spaeter erneut oder fordere einen neuen Link an."
	}
}

func buildPublicEventPageURL(baseURL, tenantSlug, eventSlug string) string {
	return buildPublicEventURL(baseURL, "", tenantSlug, "", eventSlug)
}
