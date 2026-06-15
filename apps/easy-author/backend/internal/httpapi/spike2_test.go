package httpapi_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/config"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/db"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/httpapi"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

func TestSpikeTwoEndpoints(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	database, err := db.OpenSQLite(filepath.Join(tempDir, "easy-author.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	appStore := store.New(database, filepath.Join(tempDir, "library"))
	if err := appStore.Init(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	app := httpapi.New(config.Config{
		AllowedOrigin: "http://127.0.0.1:5173",
	}, appStore)

	projectID := createJSON(t, app.Handler(), http.MethodPost, "/api/projects", map[string]any{
		"title":          "UX-Projekt",
		"description":    "Dashboard und Struktur",
		"status":         "active",
		"last_opened_at": "2026-06-12T09:00:00Z",
	})["id"].(string)

	updatedProject := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/projects/"+projectID, map[string]any{
		"title":          "UX-Projekt",
		"description":    "Atticus-inspirierter Spike",
		"status":         "paused",
		"last_opened_at": "2026-06-12T10:00:00Z",
	}, http.StatusOK)
	if updatedProject["status"].(string) != "paused" {
		t.Fatalf("expected paused project, got %#v", updatedProject["status"])
	}

	theme := createJSON(t, app.Handler(), http.MethodPost, "/api/themes", map[string]any{
		"name":          "Calm PDF",
		"description":   "Ruhiges Grundtheme",
		"target":        "pdf",
		"settings_json": `{"page_format":"A5"}`,
	})
	themeID := theme["id"].(string)

	updatedTheme := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/themes/"+themeID, map[string]any{
		"name":          "Calm PDF",
		"description":   "Ruhiges Grundtheme fuer Druck",
		"target":        "pdf",
		"settings_json": `{"page_format":"A5","margins":"normal"}`,
	}, http.StatusOK)
	if updatedTheme["description"].(string) != "Ruhiges Grundtheme fuer Druck" {
		t.Fatalf("expected updated theme description, got %#v", updatedTheme["description"])
	}

	bookID := createJSON(t, app.Handler(), http.MethodPost, "/api/projects/"+projectID+"/books", map[string]any{
		"title":          "Das stille Manuskript",
		"subtitle":       "MVP 2",
		"author":         "",
		"visibility":     "private",
		"work_type":      "freebie",
		"theme_id":       themeID,
		"cover_asset_id": "",
	})["id"].(string)

	updatedBook := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/books/"+bookID, map[string]any{
		"title":          "Das stille Manuskript",
		"subtitle":       "MVP 2",
		"author":         "Cody",
		"visibility":     "private",
		"work_type":      "book",
		"theme_id":       themeID,
		"cover_asset_id": "cover-asset-1",
	}, http.StatusOK)
	if updatedBook["work_type"].(string) != "book" {
		t.Fatalf("expected updated work type, got %#v", updatedBook["work_type"])
	}
	if updatedBook["cover_asset_id"].(string) != "cover-asset-1" {
		t.Fatalf("expected cover asset id, got %#v", updatedBook["cover_asset_id"])
	}

	frontMatterChapterID := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/chapters", map[string]any{
		"title":            "Vorwort",
		"markdown_content": "TODO: Vorwort ausformulieren",
		"section_type":     "front_matter",
		"status":           "idea",
	})["id"].(string)
	bodyChapterID := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/chapters", map[string]any{
		"title":            "Kapitel 1",
		"markdown_content": "# Kapitel 1\n\nEin kurzer Kapiteltext.",
		"section_type":     "body",
		"status":           "draft",
	})["id"].(string)
	fragmentChapterID := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/chapters", map[string]any{
		"title":            "Szene-Splitter",
		"markdown_content": "",
		"section_type":     "fragment",
		"status":           "idea",
	})["id"].(string)

	updatedChapter := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/chapters/"+bodyChapterID+"/options", map[string]any{
		"section_type":          "body",
		"status":                "review",
		"is_included_in_export": false,
		"is_visible_in_toc":     true,
		"is_sample_content":     true,
	}, http.StatusOK)
	if updatedChapter["status"].(string) != "review" {
		t.Fatalf("expected review chapter status, got %#v", updatedChapter["status"])
	}
	if updatedChapter["is_included_in_export"] != false {
		t.Fatalf("expected excluded export chapter, got %#v", updatedChapter["is_included_in_export"])
	}

	reusablePage := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/reusable-pages", map[string]any{
		"title":            "Impressum",
		"page_type":        "imprint",
		"content_markdown": "Impressumstext",
		"content_json":     "",
		"is_global":        false,
	})
	reusablePageID := reusablePage["id"].(string)

	updatedReusablePage := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/reusable-pages/"+reusablePageID, map[string]any{
		"title":            "Impressum aktualisiert",
		"page_type":        "imprint",
		"content_markdown": "Aktualisierter Impressumstext",
		"content_json":     "",
		"is_global":        true,
	}, http.StatusOK)
	if updatedReusablePage["is_global"] != true {
		t.Fatalf("expected global reusable page, got %#v", updatedReusablePage["is_global"])
	}

	reusablePages := requestJSON(t, app.Handler(), http.MethodGet, "/api/books/"+bookID+"/reusable-pages", nil)["reusable_pages"].([]any)
	if len(reusablePages) != 1 {
		t.Fatalf("expected one reusable page, got %d", len(reusablePages))
	}

	structure := requestJSON(t, app.Handler(), http.MethodGet, "/api/books/"+bookID+"/structure", nil)
	if len(structure["front_matter"].([]any)) != 1 {
		t.Fatalf("expected one front matter chapter")
	}
	if len(structure["body"].([]any)) != 1 {
		t.Fatalf("expected one body chapter")
	}
	if len(structure["fragments"].([]any)) != 1 {
		t.Fatalf("expected one fragment chapter")
	}

	themes := requestJSON(t, app.Handler(), http.MethodGet, "/api/themes", nil)["themes"].([]any)
	if len(themes) == 0 {
		t.Fatalf("expected at least one theme")
	}

	dashboard := requestJSON(t, app.Handler(), http.MethodGet, "/api/dashboard", nil)["projects"].([]any)
	if len(dashboard) == 0 {
		t.Fatalf("expected dashboard projects")
	}
	booksSummary := dashboard[0].(map[string]any)["books"].([]any)
	if len(booksSummary) == 0 {
		t.Fatalf("expected dashboard books summary")
	}

	proofingChecks := requestJSON(t, app.Handler(), http.MethodGet, "/api/books/"+bookID+"/proofing-checks", nil)["checks"].([]any)
	if len(proofingChecks) == 0 {
		t.Fatalf("expected proofing checks")
	}

	requestJSONWithStatus(t, app.Handler(), http.MethodDelete, "/api/reusable-pages/"+reusablePageID, nil, http.StatusNoContent)
	reusablePagesAfterDelete := requestJSON(t, app.Handler(), http.MethodGet, "/api/books/"+bookID+"/reusable-pages", nil)["reusable_pages"].([]any)
	if len(reusablePagesAfterDelete) != 0 {
		t.Fatalf("expected reusable page to be deleted, got %d", len(reusablePagesAfterDelete))
	}

	_ = frontMatterChapterID
	_ = fragmentChapterID
}
