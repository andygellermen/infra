package httpapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/snippet"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

const (
	defaultSnippetViewType = "cards"
)

func (a *App) handleAdminSnippetsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleAdminSnippetList(w, r)
	case http.MethodPost:
		a.handleAdminSnippetCreate(w, r)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminSnippetsItem(w http.ResponseWriter, r *http.Request) {
	snippetID, action, ok := parseAdminSnippetPath(r.URL.Path)
	if !ok {
		writeAPIError(w, http.StatusNotFound, "SNIPPET_NOT_FOUND", "Snippet-Konfiguration nicht gefunden.")
		return
	}

	if action == "embed-code" {
		if r.Method != http.MethodGet {
			writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
			return
		}
		a.handleAdminSnippetEmbedCode(w, r, snippetID)
		return
	}
	if action != "" {
		writeAPIError(w, http.StatusNotFound, "SNIPPET_NOT_FOUND", "Snippet-Konfiguration nicht gefunden.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.handleAdminSnippetGet(w, r, snippetID)
	case http.MethodPatch:
		a.handleAdminSnippetPatch(w, r, snippetID)
	case http.MethodDelete:
		a.handleAdminSnippetDelete(w, r, snippetID)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
	}
}

func (a *App) handleAdminSnippetList(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.snippetRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Snippet-Service ist nicht verfuegbar.")
		return
	}

	items, err := a.snippetRepo.ListConfigs(r.Context(), principal.TenantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Snippet-Konfigurationen konnten nicht geladen werden.")
		return
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, snippetConfigPayload(item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": result,
		"total": len(result),
	})
}

func (a *App) handleAdminSnippetCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.snippetRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Snippet-Service ist nicht verfuegbar.")
		return
	}

	var request struct {
		Name           string           `json:"name"`
		Slug           string           `json:"slug"`
		ViewType       string           `json:"view_type"`
		EventFilter    *json.RawMessage `json:"event_filter"`
		DisplayOptions *json.RawMessage `json:"display_options"`
		IsActive       *bool            `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	eventFilterJSON, err := rawJSONObjectToString(request.EventFilter, "event_filter")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	displayOptionsJSON, err := rawJSONObjectToString(request.DisplayOptions, "display_options")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	created, err := a.snippetRepo.CreateConfig(r.Context(), principal.TenantID, snippet.CreateConfigParams{
		Name:               request.Name,
		Slug:               request.Slug,
		ViewType:           request.ViewType,
		EventFilterJSON:    eventFilterJSON,
		DisplayOptionsJSON: displayOptionsJSON,
		IsActive:           request.IsActive,
	})
	if err != nil {
		a.writeSnippetError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"item": snippetConfigPayload(created),
	})
}

func (a *App) handleAdminSnippetGet(w http.ResponseWriter, r *http.Request, snippetID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.snippetRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Snippet-Service ist nicht verfuegbar.")
		return
	}

	item, err := a.snippetRepo.GetConfigByID(r.Context(), principal.TenantID, snippetID)
	if err != nil {
		a.writeSnippetError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": snippetConfigPayload(item),
	})
}

func (a *App) handleAdminSnippetPatch(w http.ResponseWriter, r *http.Request, snippetID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.snippetRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Snippet-Service ist nicht verfuegbar.")
		return
	}

	var request struct {
		Name           *string          `json:"name"`
		Slug           *string          `json:"slug"`
		ViewType       *string          `json:"view_type"`
		EventFilter    *json.RawMessage `json:"event_filter"`
		DisplayOptions *json.RawMessage `json:"display_options"`
		IsActive       *bool            `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	eventFilterJSON, err := rawJSONObjectToOptionalStringPointer(request.EventFilter, "event_filter")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	displayOptionsJSON, err := rawJSONObjectToOptionalStringPointer(request.DisplayOptions, "display_options")
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	updated, err := a.snippetRepo.UpdateConfig(r.Context(), principal.TenantID, snippetID, snippet.UpdateConfigParams{
		Name:               request.Name,
		Slug:               request.Slug,
		ViewType:           request.ViewType,
		EventFilterJSON:    eventFilterJSON,
		DisplayOptionsJSON: displayOptionsJSON,
		IsActive:           request.IsActive,
	})
	if err != nil {
		a.writeSnippetError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item": snippetConfigPayload(updated),
	})
}

func (a *App) handleAdminSnippetDelete(w http.ResponseWriter, r *http.Request, snippetID string) {
	principal, ok := a.requireAdminPrincipal(w, r, true)
	if !ok {
		return
	}
	if a.snippetRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Snippet-Service ist nicht verfuegbar.")
		return
	}

	deleted, err := a.snippetRepo.DeleteConfig(r.Context(), principal.TenantID, snippetID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Snippet-Konfiguration konnte nicht geloescht werden.")
		return
	}
	if !deleted {
		writeAPIError(w, http.StatusNotFound, "SNIPPET_NOT_FOUND", "Snippet-Konfiguration nicht gefunden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleAdminSnippetEmbedCode(w http.ResponseWriter, r *http.Request, snippetID string) {
	principal, ok := a.requireAdminPrincipal(w, r, false)
	if !ok {
		return
	}
	if a.snippetRepo == nil || a.tenantRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Snippet-Service ist nicht verfuegbar.")
		return
	}

	item, err := a.snippetRepo.GetConfigByID(r.Context(), principal.TenantID, snippetID)
	if err != nil {
		a.writeSnippetError(w, err)
		return
	}
	tenantItem, err := a.tenantRepo.GetByID(r.Context(), principal.TenantID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Mandant konnte nicht geladen werden.")
		return
	}

	scriptSrc := buildSnippetScriptSrc(tenantItem.PublicBaseURL, tenantItem.Slug, item.Slug)
	embedCode := fmt.Sprintf(`<script src="%s" defer></script>`, scriptSrc)

	writeJSON(w, http.StatusOK, map[string]any{
		"item":       snippetConfigPayload(item),
		"script_src": scriptSrc,
		"embed_code": embedCode,
	})
}

func (a *App) handlePublicSnippetEvents(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if a.snippetRepo == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Snippet-Service ist nicht verfuegbar.")
		return
	}

	viewType := defaultSnippetViewType
	filter := event.PublicEventFilter{}
	eventSlug := ""
	query := r.URL.Query()
	configSlug := strings.TrimSpace(query.Get("config"))
	var configPayload any
	if configSlug != "" {
		if err := validateSnippetConfigOnlyQuery(query); err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}

		configItem, err := a.snippetRepo.GetConfigBySlug(r.Context(), tenantItem.ID, configSlug)
		if err != nil {
			if errors.Is(err, snippet.ErrSnippetNotFound) {
				writeAPIError(w, http.StatusNotFound, "SNIPPET_NOT_FOUND", "Snippet-Konfiguration nicht gefunden.")
				return
			}
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if !configItem.IsActive {
			writeAPIError(w, http.StatusNotFound, "SNIPPET_NOT_FOUND", "Snippet-Konfiguration nicht gefunden.")
			return
		}
		viewType = configItem.ViewType

		if err := applySnippetOptionsFromJSONString(configItem.EventFilterJSON, &filter, &eventSlug, nil); err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		displayOptions, err := parseSnippetJSONMap(configItem.DisplayOptionsJSON)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		configPayload = map[string]any{
			"id":              configItem.ID,
			"slug":            configItem.Slug,
			"name":            configItem.Name,
			"view_type":       configItem.ViewType,
			"display_options": displayOptions,
		}
	} else {
		queryOptionsView := ""
		if err := applySnippetOptionsFromValues(query, &filter, &eventSlug, &queryOptionsView); err != nil {
			writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if queryOptionsView != "" {
			viewType = queryOptionsView
		}
	}
	if strings.TrimSpace(viewType) == "" {
		viewType = defaultSnippetViewType
	}

	items, err := a.listSnippetPublicEvents(r, tenantItem, filter, eventSlug)
	if err != nil {
		if errors.Is(err, event.ErrEventNotFound) {
			writeAPIError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Veranstaltung nicht gefunden.")
			return
		}
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	eventDetailBaseURL := a.snippetEventDetailBaseURL(r, tenantItem)
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, snippetPublicEventPayload(tenantItem, eventDetailBaseURL, item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant": publicTenantPayload(tenantItem),
		"config": configPayload,
		"view":   viewType,
		"items":  result,
		"total":  len(result),
	})
}

func (a *App) snippetEventDetailBaseURL(r *http.Request, tenantItem tenant.Tenant) string {
	if a.tenantRepo == nil {
		return ""
	}
	settings, err := a.tenantRepo.GetSettings(r.Context(), tenantItem.ID)
	if err != nil {
		return ""
	}
	appSettings, _, err := parseAdminTenantAppSettings(settings.SettingsJSON)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(appSettings.EventDetailBaseURL)
}

func (a *App) listSnippetPublicEvents(r *http.Request, tenantItem tenant.Tenant, filter event.PublicEventFilter, eventSlug string) ([]event.PublicEvent, error) {
	if strings.TrimSpace(eventSlug) == "" {
		return a.eventRepo.ListPublicEvents(r.Context(), tenantItem.ID, filter)
	}

	item, err := a.eventRepo.GetPublicEventBySlug(r.Context(), tenantItem.ID, eventSlug)
	if err != nil {
		return nil, err
	}
	if !matchesSnippetEventFilter(item, filter, a.eventRepo) {
		return []event.PublicEvent{}, nil
	}
	return []event.PublicEvent{item}, nil
}

func matchesSnippetEventFilter(item event.PublicEvent, filter event.PublicEventFilter, repo *event.Repository) bool {
	now := repoNow(repo)
	if !filter.IncludePast && item.StartsAt.Before(now) {
		return false
	}
	if filter.SeriesSlug != "" && item.SeriesSlug != filter.SeriesSlug {
		return false
	}
	if filter.Mode != "" && item.ParticipationMode != filter.Mode {
		return false
	}
	if filter.From != nil && item.StartsAt.Before(filter.From.UTC()) {
		return false
	}
	if filter.To != nil && item.StartsAt.After(filter.To.UTC()) {
		return false
	}
	return true
}

func repoNow(repo *event.Repository) time.Time {
	if repo == nil {
		return time.Now().UTC()
	}
	return time.Now().UTC()
}

func (a *App) handleTenantAssetRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.tenantRepo == nil {
		if r.URL.Path == "/" {
			a.handleRoot(w, r)
			return
		}
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	resolved, ok := a.resolveTenantPublicRoute(r)
	if r.URL.Path == "/" && !ok {
		a.handleRoot(w, r)
		return
	}
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch resolved.routeType {
	case "events_overview":
		a.handleTenantPublicEventsOverviewPage(w, r, resolved.tenant)
		return
	case "event_page":
		a.handleTenantPublicEventPage(w, r, resolved.tenant, resolved.eventSlug)
		return
	case "include_js":
		writeSnippetIncludeJS(w, resolved.tenant)
	case "register_js":
		writeRegistrationEmbedJS(w, resolved.tenant)
	case "snippet_css":
		writeSnippetCSS(w)
	case "organizer_calendar":
		a.handleTenantOrganizerCalendarICS(w, r, resolved.tenant)
	default:
		http.NotFound(w, r)
	}
}

type resolvedTenantPublicRoute struct {
	tenant    tenant.Tenant
	routeType string
	eventSlug string
}

func (a *App) resolveTenantPublicRoute(r *http.Request) (resolvedTenantPublicRoute, bool) {
	requestPath := normalizePublicPath(r.URL.Path)
	if requestPath == "/" {
		if resolved, ok := a.resolveTenantPublicRouteByBaseURL(r, requestPath); ok {
			return resolved, true
		}
		return resolvedTenantPublicRoute{}, false
	}

	if resolved, ok := a.resolveTenantPublicRouteByBaseURL(r, requestPath); ok {
		return resolved, true
	}
	if resolved, ok := a.resolveTenantPublicRouteLegacy(r, requestPath); ok {
		return resolved, true
	}
	return resolvedTenantPublicRoute{}, false
}

func (a *App) resolveTenantPublicRouteByBaseURL(r *http.Request, requestPath string) (resolvedTenantPublicRoute, bool) {
	match, err := a.tenantRepo.LookupPublicRoute(r.Context(), buildPublicLookupURL(r, requestPath))
	if err != nil {
		return resolvedTenantPublicRoute{}, false
	}
	relativePath, ok := trimPublicBasePathPrefix(requestPath, match.BasePath)
	if !ok {
		return resolvedTenantPublicRoute{}, false
	}
	return classifyTenantPublicRelativePath(match.Tenant, relativePath)
}

func (a *App) resolveTenantPublicRouteLegacy(r *http.Request, requestPath string) (resolvedTenantPublicRoute, bool) {
	parts := strings.Split(strings.Trim(requestPath, "/"), "/")
	if len(parts) == 0 {
		return resolvedTenantPublicRoute{}, false
	}
	tenantSlug := strings.TrimSpace(parts[0])
	if tenantSlug == "" {
		return resolvedTenantPublicRoute{}, false
	}
	tenantItem, err := a.tenantRepo.LookupBySlug(r.Context(), tenantSlug)
	if err != nil {
		return resolvedTenantPublicRoute{}, false
	}
	relativePath := "/"
	if len(parts) > 1 {
		relativePath = "/" + strings.Join(parts[1:], "/")
	}
	return classifyTenantPublicRelativePath(tenantItem, relativePath)
}

func classifyTenantPublicRelativePath(tenantItem tenant.Tenant, relativePath string) (resolvedTenantPublicRoute, bool) {
	normalized := normalizePublicPath(relativePath)
	if normalized == "/" {
		return resolvedTenantPublicRoute{tenant: tenantItem, routeType: "events_overview"}, true
	}
	parts := strings.Split(strings.Trim(normalized, "/"), "/")
	if len(parts) == 0 {
		return resolvedTenantPublicRoute{}, false
	}

	switch len(parts) {
	case 1:
		switch strings.TrimSpace(parts[0]) {
		case "include.js":
			return resolvedTenantPublicRoute{tenant: tenantItem, routeType: "include_js"}, true
		case "register.js":
			return resolvedTenantPublicRoute{tenant: tenantItem, routeType: "register_js"}, true
		case "snippet.css":
			return resolvedTenantPublicRoute{tenant: tenantItem, routeType: "snippet_css"}, true
		default:
			return resolvedTenantPublicRoute{}, false
		}
	case 2:
		if strings.TrimSpace(parts[0]) == "events" && strings.TrimSpace(parts[1]) != "" {
			return resolvedTenantPublicRoute{tenant: tenantItem, routeType: "event_page", eventSlug: strings.TrimSpace(parts[1])}, true
		}
		if strings.TrimSpace(parts[0]) == "calendar" && strings.TrimSpace(parts[1]) == "admin.ics" {
			return resolvedTenantPublicRoute{tenant: tenantItem, routeType: "organizer_calendar"}, true
		}
		return resolvedTenantPublicRoute{}, false
	default:
		return resolvedTenantPublicRoute{}, false
	}
}

func normalizePublicPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	return "/" + strings.Trim(trimmed, "/")
}

func buildPublicLookupURL(r *http.Request, requestPath string) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = strings.ToLower(strings.TrimSpace(strings.Split(forwarded, ",")[0]))
	}
	return scheme + "://" + requestHost(r) + normalizePublicPath(requestPath)
}

func publicBasePathFromURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "/"
	}
	return normalizePublicPath(parsed.EscapedPath())
}

func trimPublicBasePathPrefix(requestPath, basePath string) (string, bool) {
	request := normalizePublicPath(requestPath)
	base := normalizePublicPath(basePath)
	if base == "/" {
		return request, true
	}
	if request == base {
		return "/", true
	}
	prefix := base + "/"
	if !strings.HasPrefix(request, prefix) {
		return "", false
	}
	return "/" + strings.TrimPrefix(request, prefix), true
}

func writeSnippetIncludeJS(w http.ResponseWriter, tenantItem tenant.Tenant) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(buildSnippetIncludeJS(tenantItem.Slug)))
}

func writeRegistrationEmbedJS(w http.ResponseWriter, tenantItem tenant.Tenant) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(buildRegistrationEmbedJS(tenantItem.Slug)))
}

func writeSnippetCSS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(snippetCSS))
}

func buildSnippetIncludeJS(tenantSlug string) string {
	return fmt.Sprintf(`(function () {
  const script = document.currentScript;
  if (!script) return;

  const source = new URL(script.src);
  const assetBase = source.origin + source.pathname.replace(/\/[^/]*$/, "");
  const rawParams = new URLSearchParams(source.search);
  const params = new URLSearchParams();
  const config = rawParams.get("config");
  if (config) {
    params.set("config", config);
  } else {
    rawParams.forEach(function (value, key) {
      params.append(key, value);
    });
  }
  const dataTarget = script.getAttribute("data-target");
  const targetSelector = dataTarget || (config ? "" : rawParams.get("target"));
  const tenantSlug = %q;
  const apiURL = source.origin + "/api/v1/public/" + encodeURIComponent(tenantSlug) + "/snippet/events?" + params.toString();

  if (!config && shouldLoadCSS(null, rawParams, script)) {
    ensureCSS(assetBase + "/snippet.css", tenantSlug);
  }
  const container = resolveContainer(script, targetSelector);
  container.classList.add("eep-widget");
  container.classList.add("eep-loading");

  fetch(apiURL, { credentials: "omit" })
    .then(function (res) {
      if (!res.ok) throw new Error("HTTP " + res.status);
      return res.json();
    })
    .then(function (payload) {
      const items = Array.isArray(payload.items) ? payload.items : [];
      const view = String(payload.view || "cards").toLowerCase();
      const displayOptions = payload && payload.config && payload.config.display_options ? payload.config.display_options : {};
      if (shouldLoadCSS(displayOptions, rawParams, script)) {
        ensureCSS(assetBase + "/snippet.css", tenantSlug);
      }
      applyTheme(container, displayOptions);
      render(container, payload, items, view, displayOptions);
      container.classList.remove("eep-loading");
    })
    .catch(function () {
      container.classList.remove("eep-loading");
      container.innerHTML = "<div class=\"eep-error\">Termine konnten gerade nicht geladen werden.</div>";
    });

  function resolveContainer(scriptNode, selector) {
    if (selector) {
      const match = document.querySelector(selector);
      if (match) return match;
    }
    const fallback = document.createElement("div");
    scriptNode.parentNode.insertBefore(fallback, scriptNode.nextSibling);
    return fallback;
  }

  function ensureCSS(href, slug) {
    const marker = "eep-css-" + slug;
    if (document.getElementById(marker)) return;
    const link = document.createElement("link");
    link.id = marker;
    link.rel = "stylesheet";
    link.href = href;
    document.head.appendChild(link);
  }

  function shouldLoadCSS(displayOptions, params, scriptNode) {
    const attr = normalizeCSSMode(scriptNode && scriptNode.getAttribute ? scriptNode.getAttribute("data-css") : "");
    if (attr !== "") return attr === "on";
    const query = normalizeCSSMode(params && params.get ? params.get("css") : "");
    if (query !== "") return query === "on";
    if (displayOptions && displayOptions.load_css === false) {
      return false;
    }
    return true;
  }

  function normalizeCSSMode(value) {
    switch (String(value || "").trim().toLowerCase()) {
      case "on":
      case "true":
      case "1":
      case "yes":
        return "on";
      case "off":
      case "false":
      case "0":
      case "no":
      case "inherit":
        return "off";
      default:
        return "";
    }
  }

  function applyTheme(root, displayOptions) {
    root.classList.remove("eep-theme-light", "eep-theme-sand", "eep-theme-dark");
    const theme = String(displayOptions && displayOptions.theme || "light").toLowerCase();
    if (theme === "sand" || theme === "dark") {
      root.classList.add("eep-theme-" + theme);
      return;
    }
    root.classList.add("eep-theme-light");
  }

  function render(root, payload, items, view, displayOptions) {
    if (!items.length) {
      root.innerHTML = "<div class=\"eep-empty\">Aktuell keine Termine.</div>";
      return;
    }
    if (view === "table") {
      root.innerHTML = renderTable(payload, items, displayOptions);
      return;
    }
    if (view === "list") {
      root.innerHTML = renderList(payload, items, displayOptions);
      return;
    }
    if (view === "minimal") {
      root.innerHTML = renderMinimal(items, displayOptions);
      return;
    }
    if (view === "button") {
      root.innerHTML = renderButtons(items, displayOptions);
      return;
    }
    root.innerHTML = renderCards(payload, items, displayOptions);
  }

  function renderCards(payload, items, displayOptions) {
    const intro = renderIntro(payload, items);
    const cards = items.map(function (item) {
      const title = escapeHTML(item.title || "Event");
      const subtitle = item.subtitle ? "<p class=\"eep-subtitle\">" + escapeHTML(item.subtitle) + "</p>" : "";
      const startsAt = formatDate(item.starts_at);
      const endsAt = item.ends_at ? " – " + escapeHTML(formatDate(item.ends_at)) : "";
      const dateLine = "<p class=\"eep-date\">" + startsAt + endsAt + "</p>";
      const description = item.description ? "<p class=\"eep-description\">" + escapeHTML(truncateText(item.description, 160)) + "</p>" : "";
      const meta = renderMeta(item);
      const series = item.series && item.series.title ? "<span class=\"eep-chip eep-chip-soft\">" + escapeHTML(item.series.title) + "</span>" : "";
      const mode = item.participation_mode ? "<span class=\"eep-chip\">" + escapeHTML(formatMode(item.participation_mode)) + "</span>" : "";
      const href = escapeHTML(item.event_url || "#");
      const cta = escapeHTML(callToActionLabel(displayOptions, item));
      return "<article class=\"eep-card\">" +
        "<div class=\"eep-card-top\"><div class=\"eep-card-chips\">" + series + mode + "</div></div>" +
        "<h3 class=\"eep-title\">" + title + "</h3>" +
        subtitle +
        dateLine +
        meta +
        description +
        "<a class=\"eep-link eep-primary-link\" href=\"" + href + "\">" + cta + "</a>" +
      "</article>";
    }).join("");
    return intro + "<div class=\"eep-cards\">" + cards + "</div>";
  }

  function renderList(payload, items, displayOptions) {
    const intro = renderIntro(payload, items);
    const rows = items.map(function (item) {
      const title = escapeHTML(item.title || "Event");
      const startsAt = formatDate(item.starts_at);
      const meta = renderMeta(item);
      const href = escapeHTML(item.event_url || "#");
      const cta = escapeHTML(callToActionLabel(displayOptions, item));
      return "<li class=\"eep-list-item\">" +
        "<div class=\"eep-list-main\"><span class=\"eep-list-date\">" + startsAt + "</span><a class=\"eep-list-link\" href=\"" + href + "\">" + title + "</a>" + meta + "</div>" +
        "<a class=\"eep-inline-cta\" href=\"" + href + "\">" + cta + "</a>" +
      "</li>";
    }).join("");
    return intro + "<ul class=\"eep-list\">" + rows + "</ul>";
  }

  function renderTable(payload, items, displayOptions) {
    const intro = renderIntro(payload, items);
    const rows = items.map(function (item) {
      const title = escapeHTML(item.title || "Event");
      const startsAt = formatDate(item.starts_at);
      const location = escapeHTML(item.location_name || formatMode(item.participation_mode) || "—");
      const href = escapeHTML(item.event_url || "#");
      const cta = escapeHTML(callToActionLabel(displayOptions, item));
      return "<tr>" +
        "<td>" + startsAt + "</td>" +
        "<td><a class=\"eep-list-link\" href=\"" + href + "\">" + title + "</a></td>" +
        "<td>" + location + "</td>" +
        "<td><a class=\"eep-inline-cta\" href=\"" + href + "\">" + cta + "</a></td>" +
      "</tr>";
    }).join("");
    return intro + "<div class=\"eep-table-wrap\"><table class=\"eep-table\"><thead><tr><th>Termin</th><th>Event</th><th>Ort / Format</th><th>Link</th></tr></thead><tbody>" + rows + "</tbody></table></div>";
  }

  function renderMinimal(items, displayOptions) {
    const links = items.map(function (item) {
      const title = escapeHTML(item.title || "Event");
      const startsAt = formatDate(item.starts_at);
      const href = escapeHTML(item.event_url || "#");
      const cta = escapeHTML(callToActionLabel(displayOptions, item));
      return "<a class=\"eep-minimal-link\" href=\"" + href + "\"><strong>" + title + "</strong><span>" + startsAt + "</span><em>" + cta + "</em></a>";
    }).join("");
    return "<div class=\"eep-minimal\">" + links + "</div>";
  }

  function renderButtons(items, displayOptions) {
    const buttons = items.map(function (item) {
      const href = escapeHTML(item.event_url || "#");
      const label = escapeHTML(item.title || callToActionLabel(displayOptions, item));
      return "<a class=\"eep-button-link\" href=\"" + href + "\">" + label + "</a>";
    }).join("");
    return "<div class=\"eep-buttons\">" + buttons + "</div>";
  }

  function callToActionLabel(displayOptions, item) {
    if (displayOptions && displayOptions.register && item && item.registration_enabled !== false) {
      return "Platz sichern";
    }
    if (item && item.registration_opens_at) {
      const opensAt = new Date(item.registration_opens_at);
      if (!Number.isNaN(opensAt.getTime()) && opensAt.getTime() > Date.now()) {
        return "Anmeldung ab " + formatDate(item.registration_opens_at);
      }
    }
    return "Mehr erfahren";
  }

  function renderIntro(payload, items) {
    const config = payload && payload.config ? payload.config : {};
    const tenant = payload && payload.tenant ? payload.tenant : {};
    const title = config.name || tenant.name || "";
    const count = items.length === 1 ? "1 Termin" : String(items.length) + " Termine";
    const subtitle = title ? count : "";
    if (!title && !subtitle) {
      return "";
    }
    return "<div class=\"eep-intro\">" +
      (title ? "<div class=\"eep-intro-title\">" + escapeHTML(title) + "</div>" : "") +
      (subtitle ? "<div class=\"eep-intro-subtitle\">" + escapeHTML(subtitle) + "</div>" : "") +
    "</div>";
  }

  function renderMeta(item) {
    const lines = [];
    const location = item.location_name ? escapeHTML(item.location_name) : "";
    const online = item.online_url ? "Online" : "";
    const mode = item.participation_mode ? formatMode(item.participation_mode) : "";
    const primary = [location, online].filter(Boolean).join(" · ");
    if (primary) {
      lines.push("<p class=\"eep-meta\">" + primary + "</p>");
    }
    if (!primary && mode) {
      lines.push("<p class=\"eep-meta\">" + escapeHTML(mode) + "</p>");
    }
    return lines.join("");
  }

  function formatMode(value) {
    switch (String(value || "").toLowerCase()) {
      case "online":
        return "Online";
      case "hybrid":
        return "Hybrid";
      case "onsite":
        return "Vor Ort";
      default:
        return "";
    }
  }

  function truncateText(value, limit) {
    const normalized = String(value || "").trim().replace(/\s+/g, " ");
    if (!normalized || normalized.length <= limit) {
      return normalized;
    }
    return normalized.slice(0, Math.max(0, limit - 1)).trim() + "…";
  }

  function formatDate(raw) {
    const value = String(raw || "").trim();
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    try {
      return date.toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
    } catch (err) {
      return value;
    }
  }

  function escapeHTML(value) {
    return String(value || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/\"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }
})();`, tenantSlug)
}

func buildRegistrationEmbedJS(tenantSlug string) string {
	return fmt.Sprintf(`(function () {
  const script = document.currentScript;
  if (!script) return;

  const source = new URL(script.src);
  const assetBase = source.origin + source.pathname.replace(/\/[^/]*$/, "");
  const rawParams = new URLSearchParams(source.search);
  const tenantSlug = %q;
  const eventSlug = String(script.getAttribute("data-event") || rawParams.get("event") || "").trim();
  const targetSelector = String(script.getAttribute("data-target") || rawParams.get("target") || "").trim();
  const apiBase = source.origin + "/api/v1/public/" + encodeURIComponent(tenantSlug);
  const detailURL = apiBase + "/events/" + encodeURIComponent(eventSlug);
  const registerURL = apiBase + "/registrations/start";
  const container = resolveContainer(script, targetSelector);

  if (shouldLoadCSS(rawParams, script)) {
    ensureCSS(assetBase + "/snippet.css", tenantSlug);
  }
  container.classList.add("eep-widget", "eep-registration-widget", "eep-loading");

  if (!eventSlug) {
    container.classList.remove("eep-loading");
    container.innerHTML = "<div class=\"eep-error\">Bitte einen Event-Slug fuer die Einbettung angeben.</div>";
    return;
  }

  fetch(detailURL, { credentials: "omit" })
    .then(function (res) {
      if (!res.ok) throw new Error("HTTP " + res.status);
      return res.json();
    })
    .then(function (payload) {
      const eventItem = payload && payload.item ? payload.item : null;
      const tenant = payload && payload.tenant ? payload.tenant : null;
      if (!eventItem || !eventItem.id) {
        throw new Error("EVENT_NOT_FOUND");
      }
      container.classList.remove("eep-loading");
      renderForm(container, tenant, eventItem);
      bindForm(container, eventItem);
    })
    .catch(function () {
      container.classList.remove("eep-loading");
      container.innerHTML = "<div class=\"eep-error\">Anmeldeformular konnte gerade nicht geladen werden.</div>";
    });

  function resolveContainer(scriptNode, selector) {
    if (selector) {
      const match = document.querySelector(selector);
      if (match) return match;
    }
    const fallback = document.createElement("div");
    scriptNode.parentNode.insertBefore(fallback, scriptNode.nextSibling);
    return fallback;
  }

  function ensureCSS(href, slug) {
    const marker = "eep-css-" + slug;
    if (document.getElementById(marker)) return;
    const link = document.createElement("link");
    link.id = marker;
    link.rel = "stylesheet";
    link.href = href;
    document.head.appendChild(link);
  }

  function shouldLoadCSS(params, scriptNode) {
    const attr = normalizeCSSMode(scriptNode && scriptNode.getAttribute ? scriptNode.getAttribute("data-css") : "");
    if (attr !== "") return attr === "on";
    const query = normalizeCSSMode(params && params.get ? params.get("css") : "");
    if (query !== "") return query === "on";
    return true;
  }

  function normalizeCSSMode(value) {
    switch (String(value || "").trim().toLowerCase()) {
      case "on":
      case "true":
      case "1":
      case "yes":
        return "on";
      case "off":
      case "false":
      case "0":
      case "no":
      case "inherit":
        return "off";
      default:
        return "";
    }
  }

  function renderForm(root, tenant, eventItem) {
    const disabled = eventItem.registration_enabled === false;
    const summary = [
      eventItem.starts_at ? formatDate(eventItem.starts_at) : "",
      eventItem.location_name || formatMode(eventItem.participation_mode)
    ].filter(Boolean).join(" · ");
    const description = eventItem.description
      ? "<p class=\"eep-description\">" + escapeHTML(truncateText(eventItem.description, 240)) + "</p>"
      : "";
    const title = escapeHTML(eventItem.title || "Event");
    const subtitle = eventItem.subtitle ? "<p class=\"eep-subtitle\">" + escapeHTML(eventItem.subtitle) + "</p>" : "";
    const intro = tenant && tenant.name ? "<div class=\"eep-intro-subtitle\">" + escapeHTML(tenant.name) + "</div>" : "";
    const modeField = renderParticipationField(eventItem);
    const disabledNote = disabled
      ? "<div class=\"eep-empty\">" + escapeHTML(registrationClosedMessage(eventItem)) + "</div>"
      : "";
    const form = disabled ? "" : "<form class=\"eep-registration-form\" novalidate>" +
      "<input type=\"hidden\" name=\"event_id\" value=\"" + escapeAttr(eventItem.id) + "\">" +
      modeField +
      "<label class=\"eep-field\"><span>Name</span><input name=\"name\" type=\"text\" autocomplete=\"name\" required></label>" +
      "<label class=\"eep-field\"><span>E-Mail</span><input name=\"email\" type=\"email\" autocomplete=\"email\" required></label>" +
      "<label class=\"eep-field\"><span>Telefon (optional)</span><input name=\"phone\" type=\"tel\" autocomplete=\"tel\"></label>" +
      "<label class=\"eep-check\"><input name=\"privacy_accepted\" type=\"checkbox\" required> <span>Ich stimme der Verarbeitung meiner Daten fuer diese Anmeldung zu.</span></label>" +
      "<button class=\"eep-submit\" type=\"submit\">Magic Link anfordern</button>" +
      "<div class=\"eep-form-feedback\" aria-live=\"polite\"></div>" +
    "</form>";

    root.innerHTML = "<section class=\"eep-registration-shell\">" +
      "<div class=\"eep-intro\">" +
        "<div class=\"eep-intro-title\">" + title + "</div>" +
        intro +
      "</div>" +
      subtitle +
      (summary ? "<p class=\"eep-date\">" + escapeHTML(summary) + "</p>" : "") +
      description +
      disabledNote +
      form +
    "</section>";
  }

  function bindForm(root, eventItem) {
    const form = root.querySelector("form");
    if (!form) return;
    const feedback = root.querySelector(".eep-form-feedback");
    const submitButton = form.querySelector("button[type=submit]");
    form.addEventListener("submit", function (ev) {
      ev.preventDefault();
      const body = buildPayload(form);
      if (!body) {
        setFeedback(feedback, "Bitte alle Pflichtfelder ausfuellen.", "error");
        return;
      }
      if (submitButton) {
        submitButton.disabled = true;
        submitButton.textContent = "Wird gesendet...";
      }
      setFeedback(feedback, "", "");

      fetch(registerURL, {
        method: "POST",
        credentials: "omit",
        headers: { "Content-Type": "application/json", "Accept": "application/json" },
        body: JSON.stringify(body)
      })
        .then(function (res) {
          return res.json().catch(function () { return null; }).then(function (payload) {
            if (!res.ok) {
              const message = payload && payload.error && payload.error.message ? payload.error.message : "Die Anmeldung konnte nicht gestartet werden.";
              throw new Error(message);
            }
            return payload;
          });
        })
        .then(function (payload) {
          form.reset();
          setDefaultParticipation(form, eventItem);
          const message = payload && payload.message ? payload.message : "Fast geschafft. Bitte den Link in deiner E-Mail bestaetigen.";
          setFeedback(feedback, message, "success");
        })
        .catch(function (err) {
          setFeedback(feedback, err && err.message ? err.message : "Die Anmeldung konnte nicht gestartet werden.", "error");
        })
        .finally(function () {
          if (submitButton) {
            submitButton.disabled = false;
            submitButton.textContent = "Magic Link anfordern";
          }
        });
    });

    setDefaultParticipation(form, eventItem);
  }

  function renderParticipationField(eventItem) {
    const mode = String(eventItem && eventItem.participation_mode || "").toLowerCase();
    if (mode === "hybrid") {
      return "<label class=\"eep-field\"><span>Teilnahme</span><select name=\"participation_type\"><option value=\"onsite\">Vor Ort</option><option value=\"online\">Online</option></select></label>";
    }
    const value = mode === "online" ? "online" : "onsite";
    const label = mode === "online" ? "Online" : "Vor Ort";
    return "<input type=\"hidden\" name=\"participation_type\" value=\"" + value + "\"><div class=\"eep-field eep-field-static\"><span>Teilnahme</span><strong>" + label + "</strong></div>";
  }

  function setDefaultParticipation(form, eventItem) {
    const field = form.querySelector("[name=participation_type]");
    if (!field) return;
    const mode = String(eventItem && eventItem.participation_mode || "").toLowerCase();
    if (field.tagName === "SELECT") {
      field.value = mode === "online" ? "online" : "onsite";
      return;
    }
    field.value = mode === "online" ? "online" : "onsite";
  }

  function buildPayload(form) {
    const payload = {
      event_id: readField(form, "event_id"),
      name: readField(form, "name"),
      email: readField(form, "email"),
      phone: readField(form, "phone"),
      participation_type: readField(form, "participation_type"),
      privacy_accepted: !!(form.querySelector("[name=privacy_accepted]") && form.querySelector("[name=privacy_accepted]").checked)
    };
    if (!payload.event_id || !payload.name || !payload.email || !payload.privacy_accepted) {
      return null;
    }
    if (!payload.participation_type) {
      payload.participation_type = "onsite";
    }
    return payload;
  }

  function readField(form, name) {
    const field = form.querySelector("[name=\"" + name + "\"]");
    return String(field && field.value ? field.value : "").trim();
  }

  function setFeedback(node, message, tone) {
    if (!node) return;
    node.className = "eep-form-feedback" + (tone ? " is-" + tone : "");
    node.textContent = String(message || "").trim();
  }

  function formatDate(raw) {
    const value = String(raw || "").trim();
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    try {
      return date.toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" });
    } catch (err) {
      return value;
    }
  }

  function formatMode(value) {
    switch (String(value || "").toLowerCase()) {
      case "online":
        return "Online";
      case "hybrid":
        return "Hybrid";
      case "onsite":
        return "Vor Ort";
      default:
        return "";
    }
  }

  function truncateText(value, limit) {
    const normalized = String(value || "").trim().replace(/\s+/g, " ");
    if (!normalized || normalized.length <= limit) {
      return normalized;
    }
    return normalized.slice(0, Math.max(0, limit - 1)).trim() + "…";
  }

  function registrationClosedMessage(eventItem) {
    if (!eventItem || eventItem.registration_configured === false) {
      return "Die Anmeldung ist fuer dieses Event aktuell nicht aktiv.";
    }
    if (eventItem.registration_opens_at) {
      const opensAt = new Date(eventItem.registration_opens_at);
      if (!Number.isNaN(opensAt.getTime()) && opensAt.getTime() > Date.now()) {
        return "Die Anmeldung startet am " + formatDate(eventItem.registration_opens_at) + ".";
      }
    }
    if (eventItem.registration_closes_at) {
      const closesAt = new Date(eventItem.registration_closes_at);
      if (!Number.isNaN(closesAt.getTime()) && closesAt.getTime() < Date.now()) {
        return "Die Anmeldung ist seit " + formatDate(eventItem.registration_closes_at) + " geschlossen.";
      }
    }
    return "Die Anmeldung ist fuer dieses Event aktuell nicht aktiv.";
  }

  function escapeHTML(value) {
    return String(value || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/\"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function escapeAttr(value) {
    return escapeHTML(value);
  }
})();`, tenantSlug)
}

func validateSnippetConfigOnlyQuery(values url.Values) error {
	if strings.TrimSpace(values.Get("config")) == "" {
		return fmt.Errorf("config query parameter must not be empty")
	}
	for key, items := range values {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if normalizedKey != "config" {
			return fmt.Errorf("config-based snippets only allow the config query parameter")
		}
		if len(items) != 1 {
			return fmt.Errorf("config query parameter must be provided exactly once")
		}
	}
	return nil
}

const snippetCSS = `.eep-widget {
  font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
  color: #17212b;
  line-height: 1.5;
}

.eep-widget.eep-theme-light {
  color: #17212b;
}

.eep-widget.eep-theme-sand {
  color: #4f3523;
}

.eep-widget.eep-theme-dark {
  color: #f4f7fb;
}

.eep-loading,
.eep-empty,
.eep-error {
  padding: 14px 16px;
  border-radius: 14px;
  background: #f4f7fa;
  border: 1px solid #e2e8ef;
}

.eep-error {
  background: #fff0f0;
  color: #8f1f1f;
}

.eep-widget.eep-theme-dark .eep-loading,
.eep-widget.eep-theme-dark .eep-empty {
  background: #213244;
  color: #f4f7fb;
}

.eep-widget.eep-theme-sand .eep-loading,
.eep-widget.eep-theme-sand .eep-empty {
  background: #f4eadc;
  color: #5c4332;
}

.eep-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 14px;
}

.eep-card {
  border: 1px solid #dbe3eb;
  border-radius: 18px;
  padding: 16px;
  background: linear-gradient(180deg, #ffffff, #fbfdff);
  box-shadow: 0 12px 30px rgba(18, 33, 43, 0.06);
}

.eep-widget.eep-theme-dark .eep-card {
  background: #17212b;
  border-color: #314355;
}

.eep-widget.eep-theme-sand .eep-card {
  background: #fffaf2;
  border-color: #e6d2b7;
}

.eep-intro {
  margin-bottom: 12px;
}

.eep-intro-title {
  font-size: 1.05rem;
  font-weight: 700;
  margin-bottom: 2px;
}

.eep-intro-subtitle {
  color: #4b5b6c;
  font-size: 0.92rem;
}

.eep-card-top {
  margin-bottom: 8px;
}

.eep-card-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.eep-chip {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 9px;
  border-radius: 999px;
  background: #edf5ff;
  color: #27537a;
  font-size: 0.75rem;
  font-weight: 700;
}

.eep-chip-soft {
  background: #f5efe4;
  color: #6f4d22;
}

.eep-title {
  margin: 0 0 6px 0;
  font-size: 1.04rem;
  line-height: 1.28;
}

.eep-subtitle,
.eep-date,
.eep-location,
.eep-meta,
.eep-description {
  margin: 0 0 6px 0;
  color: #4b5b6c;
  font-size: 0.92rem;
}

.eep-description {
  line-height: 1.45;
}

.eep-link {
  color: #0f5ca8;
  text-decoration: none;
  font-weight: 600;
}

.eep-primary-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  padding: 0 14px;
  border-radius: 999px;
  background: #0f5ca8;
  color: #fff;
  margin-top: 6px;
}

.eep-inline-cta,
.eep-button-link {
  color: #0f5ca8;
  text-decoration: none;
  font-weight: 600;
}

.eep-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.eep-list-item {
  padding: 12px 0;
  border-bottom: 1px solid #e4e9ef;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
}

.eep-list-main {
  display: grid;
  gap: 3px;
}

.eep-list-date {
  color: #4b5b6c;
  font-size: 0.9rem;
}

.eep-list-link {
  color: #0f5ca8;
  text-decoration: none;
  font-weight: 600;
}

.eep-table-wrap {
  overflow-x: auto;
}

.eep-table {
  width: 100%;
  border-collapse: collapse;
  background: #fff;
  border-radius: 14px;
  overflow: hidden;
}

.eep-table th,
.eep-table td {
  padding: 10px 8px;
  border-bottom: 1px solid #e4e9ef;
  text-align: left;
  font-size: 0.92rem;
}

.eep-minimal {
  display: grid;
  gap: 10px;
}

.eep-minimal-link {
  display: grid;
  gap: 4px;
  padding: 12px 0;
  border-bottom: 1px solid #e4e9ef;
  color: inherit;
  text-decoration: none;
}

.eep-minimal-link span,
.eep-minimal-link em {
  color: #4b5b6c;
  font-size: 0.9rem;
  font-style: normal;
}

.eep-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.eep-button-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 42px;
  padding: 0 14px;
  border-radius: 999px;
  background: #0f5ca8;
  color: #fff;
  box-shadow: 0 10px 24px rgba(15, 92, 168, 0.18);
}

.eep-widget a:hover {
  opacity: 0.92;
}

.eep-registration-shell {
  border: 1px solid #dbe3eb;
  border-radius: 18px;
  padding: 18px;
  background: linear-gradient(180deg, #ffffff, #fbfdff);
  box-shadow: 0 12px 30px rgba(18, 33, 43, 0.06);
}

.eep-widget.eep-theme-dark .eep-registration-shell {
  background: #17212b;
  border-color: #314355;
}

.eep-widget.eep-theme-sand .eep-registration-shell {
  background: #fffaf2;
  border-color: #e6d2b7;
}

.eep-registration-form {
  display: grid;
  gap: 12px;
  margin-top: 14px;
}

.eep-field {
  display: grid;
  gap: 6px;
  font-size: 0.92rem;
}

.eep-field span {
  font-weight: 600;
}

.eep-field input,
.eep-field select {
  min-height: 42px;
  padding: 0 12px;
  border-radius: 12px;
  border: 1px solid #ccd7e2;
  background: #fff;
  color: inherit;
  font: inherit;
}

.eep-field-static {
  padding: 10px 12px;
  border: 1px solid #dbe3eb;
  border-radius: 12px;
  background: #f7fafc;
}

.eep-widget.eep-theme-dark .eep-field input,
.eep-widget.eep-theme-dark .eep-field select,
.eep-widget.eep-theme-dark .eep-field-static {
  background: #213244;
  border-color: #314355;
  color: #f4f7fb;
}

.eep-widget.eep-theme-sand .eep-field input,
.eep-widget.eep-theme-sand .eep-field select,
.eep-widget.eep-theme-sand .eep-field-static {
  background: #fffdf8;
  border-color: #e6d2b7;
}

.eep-check {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 0.9rem;
}

.eep-submit {
  min-height: 44px;
  border: 0;
  border-radius: 999px;
  background: #0f5ca8;
  color: #fff;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
  box-shadow: 0 10px 24px rgba(15, 92, 168, 0.18);
}

.eep-submit:disabled {
  opacity: 0.7;
  cursor: wait;
}

.eep-form-feedback {
  min-height: 20px;
  font-size: 0.9rem;
  color: #4b5b6c;
}

.eep-form-feedback.is-success {
  color: #0f6a48;
}

.eep-form-feedback.is-error {
  color: #8f1f1f;
}

@media (max-width: 640px) {
  .eep-cards {
    grid-template-columns: 1fr;
  }

  .eep-list-item {
    align-items: flex-start;
  }
}`

func snippetConfigPayload(item snippet.Config) map[string]any {
	eventFilter, _ := parseSnippetJSONMap(item.EventFilterJSON)
	displayOptions, _ := parseSnippetJSONMap(item.DisplayOptionsJSON)
	return map[string]any{
		"id":              item.ID,
		"tenant_id":       item.TenantID,
		"name":            item.Name,
		"slug":            item.Slug,
		"view_type":       item.ViewType,
		"event_filter":    eventFilter,
		"display_options": displayOptions,
		"is_active":       item.IsActive,
		"created_at":      item.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":      item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func parseAdminSnippetPath(path string) (snippetID, action string, ok bool) {
	const prefix = "/api/v1/admin/snippets/"
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
	if len(parts) == 2 {
		id := strings.TrimSpace(parts[0])
		action := strings.TrimSpace(parts[1])
		if id == "" || action == "" {
			return "", "", false
		}
		return id, action, true
	}
	return "", "", false
}

func rawJSONObjectToString(raw *json.RawMessage, fieldName string) (string, error) {
	if raw == nil {
		return "", nil
	}
	return normalizeRawJSONObject(*raw, fieldName)
}

func rawJSONObjectToOptionalStringPointer(raw *json.RawMessage, fieldName string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	normalized, err := normalizeRawJSONObject(*raw, fieldName)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func normalizeRawJSONObject(raw json.RawMessage, fieldName string) (string, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return "", nil
	}
	if !json.Valid([]byte(trimmed)) {
		return "", fmt.Errorf("%s must be valid JSON", fieldName)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return "", fmt.Errorf("%s must be a JSON object", fieldName)
	}
	if payload == nil {
		return "", fmt.Errorf("%s must be a JSON object", fieldName)
	}
	normalized, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", fieldName, err)
	}
	return string(normalized), nil
}

func parseSnippetJSONMap(raw string) (map[string]any, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return map[string]any{}, nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil, err
	}
	if payload == nil {
		return map[string]any{}, nil
	}
	return payload, nil
}

func applySnippetOptionsFromJSONString(raw string, filter *event.PublicEventFilter, eventSlug *string, viewType *string) error {
	options, err := parseSnippetJSONMap(raw)
	if err != nil {
		return fmt.Errorf("event filter configuration is invalid: %w", err)
	}
	values := url.Values{}
	for key, value := range options {
		values.Set(key, fmt.Sprintf("%v", value))
	}
	return applySnippetOptionsFromValues(values, filter, eventSlug, viewType)
}

func applySnippetOptionsFromValues(values url.Values, filter *event.PublicEventFilter, eventSlug *string, viewType *string) error {
	for key, vals := range values {
		if len(vals) == 0 {
			continue
		}
		value := strings.TrimSpace(vals[0])
		if value == "" {
			continue
		}

		switch key {
		case "view", "layout":
			if viewType == nil {
				continue
			}
			normalizedView := strings.ToLower(value)
			switch normalizedView {
			case "cards", "list", "table", "minimal", "button":
				*viewType = normalizedView
			default:
				return fmt.Errorf("view must be one of cards|list|table|minimal|button")
			}
		case "limit":
			limit, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("limit must be a number")
			}
			filter.Limit = limit
		case "include_past":
			includePast, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("include_past must be true or false")
			}
			filter.IncludePast = includePast
		case "events":
			switch strings.ToLower(value) {
			case "all":
				filter.IncludePast = true
			case "upcoming":
				filter.IncludePast = false
			default:
				return fmt.Errorf("events must be all or upcoming")
			}
		case "series":
			filter.SeriesSlug = value
		case "mode":
			filter.Mode = value
		case "event":
			*eventSlug = value
		case "from":
			from, err := parsePublicDateFilter(value, false)
			if err != nil {
				return err
			}
			filter.From = from
		case "to":
			to, err := parsePublicDateFilter(value, true)
			if err != nil {
				return err
			}
			filter.To = to
		}
	}
	return nil
}

func snippetPublicEventPayload(tenantItem tenant.Tenant, eventDetailBaseURL string, item event.PublicEvent) map[string]any {
	now := time.Now().UTC()
	var endsAt any
	if item.EndsAt != nil {
		endsAt = item.EndsAt.UTC().Format(time.RFC3339)
	}
	var publicVisibleFrom any
	if item.PublicVisibleFrom != nil {
		publicVisibleFrom = item.PublicVisibleFrom.UTC().Format(time.RFC3339)
	}
	var registrationOpensAt any
	if item.RegistrationOpensAt != nil {
		registrationOpensAt = item.RegistrationOpensAt.UTC().Format(time.RFC3339)
	}
	var registrationClosesAt any
	if item.RegistrationClosesAt != nil {
		registrationClosesAt = item.RegistrationClosesAt.UTC().Format(time.RFC3339)
	}
	var series any
	if strings.TrimSpace(item.SeriesSlug) != "" {
		series = map[string]any{
			"id":    item.SeriesID,
			"slug":  item.SeriesSlug,
			"title": item.SeriesTitle,
		}
	}

	return map[string]any{
		"id":                      item.ID,
		"slug":                    item.Slug,
		"title":                   item.Title,
		"subtitle":                item.Subtitle,
		"description":             item.Description,
		"starts_at":               item.StartsAt.UTC().Format(time.RFC3339),
		"ends_at":                 endsAt,
		"timezone":                item.Timezone,
		"location_name":           item.LocationName,
		"address":                 item.Address,
		"online_url":              item.OnlineURL,
		"participation_mode":      item.ParticipationMode,
		"status":                  item.Status,
		"series":                  series,
		"event_url":               buildPublicEventURL(tenantItem.PublicBaseURL, eventDetailBaseURL, tenantItem.Slug, item.SeriesSlug, item.Slug),
		"public_visible_from":     publicVisibleFrom,
		"registration_enabled":    item.IsRegistrationOpenAt(now),
		"registration_configured": item.RegistrationEnabled,
		"registration_opens_at":   registrationOpensAt,
		"registration_closes_at":  registrationClosesAt,
		"is_registration_open":    item.IsRegistrationOpenAt(now),
	}
}

func buildPublicEventURL(publicBaseURL, eventDetailBaseURL, tenantSlug, seriesSlug, eventSlug string) string {
	slug := url.PathEscape(strings.TrimSpace(eventSlug))
	if slug == "" {
		return ""
	}

	detailBase := strings.TrimRight(strings.TrimSpace(eventDetailBaseURL), "/")
	if detailBase != "" {
		replaced := detailBase
		replaced = strings.ReplaceAll(replaced, "{slug}", slug)
		replaced = strings.ReplaceAll(replaced, "{event_slug}", slug)
		replaced = strings.ReplaceAll(replaced, "{tenant_slug}", url.PathEscape(strings.TrimSpace(tenantSlug)))
		replaced = strings.ReplaceAll(replaced, "{series_slug}", url.PathEscape(strings.TrimSpace(seriesSlug)))
		replaced = strings.ReplaceAll(replaced, "%7Bslug%7D", slug)
		replaced = strings.ReplaceAll(replaced, "%7Bevent_slug%7D", slug)
		replaced = strings.ReplaceAll(replaced, "%7Btenant_slug%7D", url.PathEscape(strings.TrimSpace(tenantSlug)))
		replaced = strings.ReplaceAll(replaced, "%7Bseries_slug%7D", url.PathEscape(strings.TrimSpace(seriesSlug)))
		if replaced != detailBase {
			return replaced
		}
		return fmt.Sprintf("%s/%s", detailBase, slug)
	}

	publicBase := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if publicBase == "" {
		return ""
	}
	return fmt.Sprintf("%s/events/%s", publicBase, slug)
}

func buildSnippetScriptSrc(baseURL, tenantSlug, configSlug string) string {
	assetBase := buildTenantPublicAssetBaseURL(baseURL, tenantSlug)
	if assetBase == "" {
		return ""
	}
	return fmt.Sprintf("%s/include.js?config=%s", assetBase, url.QueryEscape(strings.TrimSpace(configSlug)))
}

func buildRegistrationEmbedScriptSrc(baseURL, tenantSlug, eventSlug string) string {
	assetBase := buildTenantPublicAssetBaseURL(baseURL, tenantSlug)
	if assetBase == "" {
		return ""
	}
	return fmt.Sprintf("%s/register.js?event=%s", assetBase, url.QueryEscape(strings.TrimSpace(eventSlug)))
}

func buildTenantPublicAssetBaseURL(baseURL, tenantSlug string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	if parsed, err := url.Parse(base); err == nil {
		path := normalizePublicPath(parsed.EscapedPath())
		if path != "/" {
			parsed.Path = strings.TrimRight(path, "/")
			parsed.RawPath = parsed.Path
			return strings.TrimRight(parsed.String(), "/")
		}
	}
	return fmt.Sprintf("%s/%s", base, url.PathEscape(strings.TrimSpace(tenantSlug)))
}

func buildPublicEventDetailAPIURL(baseURL, tenantSlug, eventSlug string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/api/v1/public/%s/events/%s", base, url.PathEscape(strings.TrimSpace(tenantSlug)), url.PathEscape(strings.TrimSpace(eventSlug)))
}

func buildPublicRegistrationStartAPIURL(baseURL, tenantSlug string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/api/v1/public/%s/registrations/start", base, url.PathEscape(strings.TrimSpace(tenantSlug)))
}

func (a *App) writeSnippetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, snippet.ErrSnippetNotFound):
		writeAPIError(w, http.StatusNotFound, "SNIPPET_NOT_FOUND", "Snippet-Konfiguration nicht gefunden.")
	case errors.Is(err, snippet.ErrSnippetSlugExists):
		writeAPIError(w, http.StatusConflict, "SNIPPET_SLUG_EXISTS", "Eine Snippet-Konfiguration mit diesem Slug existiert bereits.")
	default:
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
}
