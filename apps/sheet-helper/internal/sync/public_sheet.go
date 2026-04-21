package sync

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/sheet-helper/internal/config"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/model"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/pathutil"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/storage"
)

type PublicSheetSyncer struct {
	client *http.Client
	cfg    config.TenantConfig
	store  *storage.Store
}

type publishedSheetRef struct {
	Name string
	GID  string
}

func NewPublicSheetSyncer(cfg config.TenantConfig, store *storage.Store) *PublicSheetSyncer {
	return &PublicSheetSyncer{
		client: &http.Client{Timeout: 30 * time.Second},
		cfg:    cfg,
		store:  store,
	}
}

func (s *PublicSheetSyncer) Sync(ctx context.Context) error {
	refs, err := s.resolvePublishedSheets(ctx)
	if err != nil {
		return fmt.Errorf("resolve published sheets: %w", err)
	}

	routesRows, err := s.fetchSheet(ctx, refs, s.cfg.RoutesSheet)
	if err != nil {
		return fmt.Errorf("fetch routes sheet: %w", err)
	}
	routes, err := parseRoutes(routesRows)
	if err != nil {
		return fmt.Errorf("parse routes sheet: %w", err)
	}
	routes = normalizeRouteListSheets(routes, s.cfg.DefaultListPref)

	vcardRows, err := s.fetchSheet(ctx, refs, s.cfg.VCardsSheet)
	if err != nil {
		return fmt.Errorf("fetch vcard sheet: %w", err)
	}
	vcards, err := parseVCards(vcardRows)
	if err != nil {
		return fmt.Errorf("parse vcard sheet: %w", err)
	}

	textRows, err := s.fetchSheet(ctx, refs, s.cfg.TextsSheet)
	if err != nil {
		return fmt.Errorf("fetch text sheet: %w", err)
	}
	texts, err := parseTexts(textRows)
	if err != nil {
		return fmt.Errorf("parse text sheet: %w", err)
	}

	var listItems []model.ListItem
	for _, sheetName := range collectListSheets(routes, s.cfg.DefaultListPref) {
		rows, err := s.fetchSheet(ctx, refs, sheetName)
		if err != nil {
			return fmt.Errorf("fetch list sheet %s: %w", sheetName, err)
		}
		items, err := parseListItems(sheetName, rows)
		if err != nil {
			return fmt.Errorf("parse list sheet %s: %w", sheetName, err)
		}
		listItems = append(listItems, items...)
	}

	if err := s.store.ReplaceAll(ctx, routes, vcards, texts, listItems); err != nil {
		return fmt.Errorf("replace synced data: %w", err)
	}
	return nil
}

func (s *PublicSheetSyncer) resolvePublishedSheets(ctx context.Context) (map[string]publishedSheetRef, error) {
	if strings.TrimSpace(s.cfg.PublishedURL) == "" {
		return nil, fmt.Errorf("SHEET_HELPER_PUBLISHED_URL is required for public sheet sync")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.PublishedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build pubhtml request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform pubhtml request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected pubhtml status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read pubhtml body: %w", err)
	}

	refs, err := parsePublishedSheetRefs(string(body))
	if err != nil {
		return nil, err
	}
	return refs, nil
}

func (s *PublicSheetSyncer) fetchSheet(ctx context.Context, refs map[string]publishedSheetRef, sheetName string) ([][]string, error) {
	ref, ok := refs[sheetName]
	if !ok {
		return nil, fmt.Errorf("sheet %q not found in published html", sheetName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, publicPublishedCSVURL(s.cfg.PublishedURL, ref.GID), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	reader := csv.NewReader(resp.Body)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	return rows, nil
}

func publicPublishedCSVURL(publishedURL, gid string) string {
	base := strings.TrimSpace(publishedURL)
	base = strings.TrimSuffix(base, "/")
	base = strings.Replace(base, "/pubhtml", "/pub", 1)
	base = strings.Replace(base, "/pubhtml?", "/pub?", 1)

	u, err := url.Parse(base)
	if err != nil {
		return base
	}

	query := u.Query()
	query.Set("gid", gid)
	query.Set("single", "true")
	query.Set("output", "csv")
	u.RawQuery = query.Encode()
	return u.String()
}

func parsePublishedSheetRefs(html string) (map[string]publishedSheetRef, error) {
	re := regexp.MustCompile(`items\.push\(\{name:\s*"([^"]+)",\s*pageUrl:\s*"[^"]*gid=([0-9-]+)"`)
	matches := re.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no published sheet references found")
	}

	refs := make(map[string]publishedSheetRef, len(matches))
	for _, match := range matches {
		refs[match[1]] = publishedSheetRef{
			Name: match[1],
			GID:  match[2],
		}
	}
	return refs, nil
}

func collectListSheets(routes []model.Route, prefix string) []string {
	seen := make(map[string]struct{})
	var names []string
	for _, route := range routes {
		if route.Type != model.RouteTypeList {
			continue
		}
		name := strings.TrimSpace(route.ListSheet)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	return names
}

func normalizeRouteListSheets(routes []model.Route, prefix string) []model.Route {
	if prefix == "" {
		return routes
	}

	out := make([]model.Route, 0, len(routes))
	for _, route := range routes {
		if route.Type == model.RouteTypeList && route.ListSheet != "" && !strings.HasPrefix(route.ListSheet, prefix) {
			route.ListSheet = prefix + route.ListSheet
		}
		out = append(out, route)
	}
	return out
}

func parseRoutes(rows [][]string) ([]model.Route, error) {
	records, err := toRecordMaps(rows)
	if err != nil {
		return nil, err
	}
	out := make([]model.Route, 0, len(records))
	for _, record := range records {
		out = append(out, model.Route{
			Domain:      record["domain"],
			Path:        normalizePath(record["path"]),
			Type:        model.RouteType(strings.ToLower(record["type"])),
			Passphrase:  record["passphrase"],
			Target:      record["target"],
			Title:       record["title"],
			Description: record["description"],
			ListSheet:   record["listsheet"],
			Enabled:     parseBool(record["enabled"], true),
		})
	}
	return out, nil
}

func parseVCards(rows [][]string) ([]model.VCardEntry, error) {
	records, err := toRecordMaps(rows)
	if err != nil {
		return nil, err
	}
	out := make([]model.VCardEntry, 0, len(records))
	for _, record := range records {
		out = append(out, model.VCardEntry{
			Domain:       record["domain"],
			Path:         normalizePath(record["path"]),
			FullName:     record["fullname"],
			Organization: record["organization"],
			JobTitle:     record["jobtitle"],
			Email:        record["email"],
			PhoneMobile:  record["phonemobile"],
			Address:      record["address"],
			Website:      record["website"],
			ImageURL:     record["imageurl"],
			Note:         record["note"],
			Enabled:      parseBool(record["enabled"], true),
		})
	}
	return out, nil
}

func parseTexts(rows [][]string) ([]model.TextEntry, error) {
	records, err := toRecordMaps(rows)
	if err != nil {
		return nil, err
	}
	out := make([]model.TextEntry, 0, len(records))
	for _, record := range records {
		out = append(out, model.TextEntry{
			Domain:      record["domain"],
			Path:        normalizePath(record["path"]),
			ContentType: firstNonEmpty(record["contenttype"], "text/plain"),
			Content:     record["content"],
			CopyHint:    record["copyhint"],
			ExpiresAt:   record["expiresat"],
			Enabled:     parseBool(record["enabled"], true),
		})
	}
	return out, nil
}

func parseListItems(sheetName string, rows [][]string) ([]model.ListItem, error) {
	records, err := toRecordMaps(rows)
	if err != nil {
		return nil, err
	}
	out := make([]model.ListItem, 0, len(records))
	for _, record := range records {
		out = append(out, model.ListItem{
			SheetName:   sheetName,
			Sort:        parseInt(record["sort"], 0),
			Label:       record["label"],
			URL:         record["url"],
			Description: record["description"],
			Category:    record["category"],
			Password:    record["password"],
			Enabled:     parseBool(record["enabled"], true),
		})
	}
	return out, nil
}

func toRecordMaps(rows [][]string) ([]map[string]string, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	headers := make([]string, 0, len(rows[0]))
	for _, header := range rows[0] {
		headers = append(headers, normalizeHeader(header))
	}

	var records []map[string]string
	for _, row := range rows[1:] {
		if rowIsEmpty(row) {
			continue
		}
		record := make(map[string]string, len(headers))
		for idx, header := range headers {
			if header == "" {
				continue
			}
			if idx < len(row) {
				record[header] = strings.TrimSpace(row[idx])
			}
		}
		records = append(records, record)
	}
	return records, nil
}

func normalizeHeader(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	replacer := strings.NewReplacer(" ", "", "_", "", "-", "", "(", "", ")", "", "/", "")
	return replacer.Replace(value)
}

func normalizePath(path string) string {
	return pathutil.Normalize(path)
}

func parseBool(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}

func parseInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func firstNonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func rowIsEmpty(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}
