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

	scriptSrc := buildSnippetScriptSrc(a.cfg.BaseURL, tenantItem.Slug, item.Slug)
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
	configSlug := strings.TrimSpace(r.URL.Query().Get("config"))
	var configPayload any
	if configSlug != "" {
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
	}

	queryOptionsView := ""
	if err := applySnippetOptionsFromValues(r.URL.Query(), &filter, &eventSlug, &queryOptionsView); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if queryOptionsView != "" {
		viewType = queryOptionsView
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

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, snippetPublicEventPayload(tenantItem, item))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant": publicTenantPayload(tenantItem),
		"config": configPayload,
		"view":   viewType,
		"items":  result,
		"total":  len(result),
	})
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
	tenantSlug, assetType, ok := parseTenantAssetPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.tenantRepo == nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	tenantItem, err := a.tenantRepo.LookupBySlug(r.Context(), tenantSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch assetType {
	case "include_js":
		writeSnippetIncludeJS(w, tenantItem)
	case "snippet_css":
		writeSnippetCSS(w)
	case "organizer_calendar":
		a.handleTenantOrganizerCalendarICS(w, r, tenantItem)
	default:
		http.NotFound(w, r)
	}
}

func parseTenantAssetPath(path string) (tenantSlug, assetType string, ok bool) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "/" {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	switch len(parts) {
	case 2:
		tenantSlug = strings.TrimSpace(parts[0])
		asset := strings.TrimSpace(parts[1])
		if tenantSlug == "" {
			return "", "", false
		}
		switch asset {
		case "include.js":
			return tenantSlug, "include_js", true
		case "snippet.css":
			return tenantSlug, "snippet_css", true
		default:
			return "", "", false
		}
	case 3:
		tenantSlug = strings.TrimSpace(parts[0])
		group := strings.TrimSpace(parts[1])
		asset := strings.TrimSpace(parts[2])
		if tenantSlug == "" {
			return "", "", false
		}
		if group == "calendar" && asset == "admin.ics" {
			return tenantSlug, "organizer_calendar", true
		}
		return "", "", false
	default:
		return "", "", false
	}
}

func writeSnippetIncludeJS(w http.ResponseWriter, tenantItem tenant.Tenant) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(buildSnippetIncludeJS(tenantItem.Slug)))
}

func writeSnippetCSS(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(snippetCSS))
}

func buildSnippetIncludeJS(tenantSlug string) string {
	return fmt.Sprintf(`(function () {
  const script = document.currentScript;
  if (!script) return;

  const source = new URL(script.src);
  const params = new URLSearchParams(source.search);
  const tenantSlug = %q;
  const apiURL = source.origin + "/api/v1/public/" + encodeURIComponent(tenantSlug) + "/snippet/events?" + params.toString();

  ensureCSS(source.origin + "/" + encodeURIComponent(tenantSlug) + "/snippet.css", tenantSlug);
  const container = resolveContainer(script, params.get("target"));
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
      render(container, items, view);
      container.classList.remove("eep-loading");
    })
    .catch(function () {
      container.classList.remove("eep-loading");
      container.innerHTML = "<div class=\"eep-error\">Events konnten nicht geladen werden.</div>";
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

  function render(root, items, view) {
    if (!items.length) {
      root.innerHTML = "<div class=\"eep-empty\">Derzeit keine Termine.</div>";
      return;
    }
    if (view === "list" || view === "table") {
      root.innerHTML = renderList(items);
      return;
    }
    root.innerHTML = renderCards(items);
  }

  function renderCards(items) {
    const cards = items.map(function (item) {
      const title = escapeHTML(item.title || "Event");
      const subtitle = item.subtitle ? "<p class=\"eep-subtitle\">" + escapeHTML(item.subtitle) + "</p>" : "";
      const startsAt = formatDate(item.starts_at);
      const location = item.location_name ? "<p class=\"eep-location\">" + escapeHTML(item.location_name) + "</p>" : "";
      const href = escapeHTML(item.event_url || "#");
      return "<article class=\"eep-card\">" +
        "<h3 class=\"eep-title\">" + title + "</h3>" +
        subtitle +
        "<p class=\"eep-date\">" + startsAt + "</p>" +
        location +
        "<a class=\"eep-link\" href=\"" + href + "\">Details</a>" +
      "</article>";
    }).join("");
    return "<div class=\"eep-cards\">" + cards + "</div>";
  }

  function renderList(items) {
    const rows = items.map(function (item) {
      const title = escapeHTML(item.title || "Event");
      const startsAt = formatDate(item.starts_at);
      const href = escapeHTML(item.event_url || "#");
      return "<li class=\"eep-list-item\">" +
        "<span class=\"eep-list-date\">" + startsAt + "</span>" +
        "<a class=\"eep-list-link\" href=\"" + href + "\">" + title + "</a>" +
      "</li>";
    }).join("");
    return "<ul class=\"eep-list\">" + rows + "</ul>";
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

const snippetCSS = `.eep-widget {
  font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
  color: #17212b;
}

.eep-loading,
.eep-empty,
.eep-error {
  padding: 12px 14px;
  border-radius: 10px;
  background: #f4f7fa;
}

.eep-error {
  background: #fff0f0;
  color: #8f1f1f;
}

.eep-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.eep-card {
  border: 1px solid #dbe3eb;
  border-radius: 12px;
  padding: 14px;
  background: #ffffff;
}

.eep-title {
  margin: 0 0 6px 0;
  font-size: 1rem;
}

.eep-subtitle,
.eep-date,
.eep-location {
  margin: 0 0 6px 0;
  color: #4b5b6c;
  font-size: 0.92rem;
}

.eep-link {
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
  padding: 10px 0;
  border-bottom: 1px solid #e4e9ef;
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.eep-list-date {
  color: #4b5b6c;
  font-size: 0.9rem;
  min-width: 140px;
}

.eep-list-link {
  color: #0f5ca8;
  text-decoration: none;
  font-weight: 600;
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

func snippetPublicEventPayload(tenantItem tenant.Tenant, item event.PublicEvent) map[string]any {
	var endsAt any
	if item.EndsAt != nil {
		endsAt = item.EndsAt.UTC().Format(time.RFC3339)
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
		"id":                   item.ID,
		"slug":                 item.Slug,
		"title":                item.Title,
		"subtitle":             item.Subtitle,
		"starts_at":            item.StartsAt.UTC().Format(time.RFC3339),
		"ends_at":              endsAt,
		"timezone":             item.Timezone,
		"location_name":        item.LocationName,
		"participation_mode":   item.ParticipationMode,
		"status":               item.Status,
		"series":               series,
		"event_url":            buildPublicEventURL(tenantItem.PublicBaseURL, item.Slug),
		"registration_enabled": item.RegistrationEnabled,
	}
}

func buildPublicEventURL(publicBaseURL, eventSlug string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		return ""
	}
	return fmt.Sprintf("%s/events/%s", base, url.PathEscape(strings.TrimSpace(eventSlug)))
}

func buildSnippetScriptSrc(baseURL, tenantSlug, configSlug string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return fmt.Sprintf("%s/%s/include.js?config=%s", base, url.PathEscape(strings.TrimSpace(tenantSlug)), url.QueryEscape(strings.TrimSpace(configSlug)))
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
