package sync

import "testing"

func TestParseRoutes(t *testing.T) {
	rows := [][]string{
		{"Domain", "Path", "Type", "Passphrase", "Target", "Title", "Description", "ListSheet", "Enabled"},
		{"geller.men", "/", "link", "", "https://example.org", "Start", "Beschreibung", "", "true"},
		{"geller.men", "agileebooks/", "list", "scrum", "Downloads", "Agile", "Liste", "list_agileebooks", "true"},
	}

	routes, err := parseRoutes(rows)
	if err != nil {
		t.Fatalf("parseRoutes returned error: %v", err)
	}
	if got := len(routes); got != 2 {
		t.Fatalf("expected 2 routes, got %d", got)
	}
	if routes[1].Path != "/agileebooks" {
		t.Fatalf("expected trailing slash to be removed from path, got %q", routes[1].Path)
	}
	if routes[1].ListSheet != "list_agileebooks" {
		t.Fatalf("expected list sheet to survive parsing, got %q", routes[1].ListSheet)
	}
}

func TestCollectListSheets(t *testing.T) {
	routes, err := parseRoutes([][]string{
		{"Domain", "Path", "Type", "ListSheet"},
		{"geller.men", "/a", "list", "list_agileebooks"},
		{"geller.men", "/b", "list", "list_agileebooks"},
	})
	if err != nil {
		t.Fatalf("parseRoutes returned error: %v", err)
	}

	sheets := collectListSheets(routes, "list_")
	if got := len(sheets); got != 1 {
		t.Fatalf("expected 1 collected list sheet, got %d", got)
	}
	if sheets[0] != "list_agileebooks" {
		t.Fatalf("unexpected list sheet %q", sheets[0])
	}
}

func TestNormalizeRouteListSheets(t *testing.T) {
	routes, err := parseRoutes([][]string{
		{"Domain", "Path", "Type", "ListSheet"},
		{"geller.men", "/a", "list", "agileebooks"},
	})
	if err != nil {
		t.Fatalf("parseRoutes returned error: %v", err)
	}

	normalized := normalizeRouteListSheets(routes, "list_")
	if normalized[0].ListSheet != "list_agileebooks" {
		t.Fatalf("expected normalized list sheet name, got %q", normalized[0].ListSheet)
	}
}

func TestParsePublishedSheetRefs(t *testing.T) {
	html := `items.push({name: "routes", pageUrl: "https://docs.google.com/...gid=0", gid: "0",initialSheet: true});items.push({name: "text_entries", pageUrl: "https://docs.google.com/...gid=1782944325", gid: "1782944325",initialSheet: false});`

	refs, err := parsePublishedSheetRefs(html)
	if err != nil {
		t.Fatalf("parsePublishedSheetRefs returned error: %v", err)
	}
	if refs["routes"].GID != "0" {
		t.Fatalf("expected routes gid 0, got %q", refs["routes"].GID)
	}
	if refs["text_entries"].GID != "1782944325" {
		t.Fatalf("expected text_entries gid 1782944325, got %q", refs["text_entries"].GID)
	}
}

func TestPublicPublishedCSVURL(t *testing.T) {
	got := publicPublishedCSVURL("https://docs.google.com/spreadsheets/d/e/example/pubhtml", "1302663852")
	want := "https://docs.google.com/spreadsheets/d/e/example/pub?gid=1302663852&output=csv&single=true"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
