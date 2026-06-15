package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/config"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/db"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/httpapi"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

func TestVersioningEndpoints(t *testing.T) {
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
		"title":       "Versionierungsprojekt",
		"description": "Test fuer Revisionen",
	})["id"].(string)

	bookID := createJSON(t, app.Handler(), http.MethodPost, "/api/projects/"+projectID+"/books", map[string]any{
		"title":      "Versionsbuch",
		"subtitle":   "Backend-Paket A",
		"author":     "Codex",
		"visibility": "private",
	})["id"].(string)

	initialMarkdown := "# Kapitel 1\n\nErster Stand des Kapitels."
	chapterID := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/chapters", map[string]any{
		"title":            "Kapitel 1",
		"summary":          "Ausgangslage",
		"markdown_content": initialMarkdown,
		"editor_json":      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Erster Stand des Kapitels."}]}]}`,
	})["id"].(string)

	initialRevision := createJSON(t, app.Handler(), http.MethodPost, "/api/chapters/"+chapterID+"/revisions", map[string]any{
		"revision_type": "manual",
		"title":         "Vor Ueberarbeitung",
		"description":   "Erster bewusst gesetzter Stand.",
		"created_by":    "Tester",
	})
	initialRevisionID := initialRevision["id"].(string)

	fetchedRevision := requestJSON(t, app.Handler(), http.MethodGet, "/api/revisions/"+initialRevisionID, nil)
	if fetchedRevision["title"].(string) != "Vor Ueberarbeitung" {
		t.Fatalf("expected fetched revision title, got %#v", fetchedRevision["title"])
	}

	requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/chapters/"+chapterID, map[string]any{
		"title":            "Kapitel 1",
		"summary":          "Aktualisierte Fassung",
		"markdown_content": "# Kapitel 1\n\nZweiter Stand mit deutlicher Aenderung.",
		"editor_json":      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Zweiter Stand mit deutlicher Aenderung."}]}]}`,
	}, http.StatusOK)

	restoreResult := requestJSONWithStatus(t, app.Handler(), http.MethodPost, "/api/revisions/"+initialRevisionID+"/restore", map[string]any{
		"created_by": "Tester",
	}, http.StatusOK)
	chapter := restoreResult["chapter"].(map[string]any)
	if chapter["markdown_content"].(string) != initialMarkdown {
		t.Fatalf("expected restored markdown content, got %#v", chapter["markdown_content"])
	}
	if restoreResult["protection_revision"].(map[string]any)["id"].(string) == "" {
		t.Fatalf("expected protection revision id")
	}
	restoredRevisionID := restoreResult["restored_revision"].(map[string]any)["id"].(string)
	if restoredRevisionID == "" {
		t.Fatalf("expected restored revision id")
	}

	revisions := requestJSON(t, app.Handler(), http.MethodGet, "/api/chapters/"+chapterID+"/revisions", nil)["revisions"].([]any)
	if len(revisions) < 3 {
		t.Fatalf("expected at least three revisions after restore, got %d", len(revisions))
	}

	milestone := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/milestones", map[string]any{
		"revision_id":    initialRevisionID,
		"title":          "Rohfassung gesichert",
		"description":    "Startpunkt fuer den naechsten Review-Schritt.",
		"milestone_type": "rough_draft",
		"locked":         true,
		"created_by":     "Tester",
	})
	if milestone["locked"] != true {
		t.Fatalf("expected locked milestone, got %#v", milestone["locked"])
	}

	milestones := requestJSON(t, app.Handler(), http.MethodGet, "/api/books/"+bookID+"/milestones", nil)["milestones"].([]any)
	if len(milestones) != 1 {
		t.Fatalf("expected one milestone, got %d", len(milestones))
	}
	milestoneID := milestone["id"].(string)

	updatedMilestone := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/milestones/"+milestoneID, map[string]any{
		"title":          "Review vorbereitet",
		"description":    "Jetzt in der Vor-Review-Phase.",
		"milestone_type": "before_review",
		"locked":         false,
		"created_by":     "Tester",
	}, http.StatusOK)
	if updatedMilestone["milestone_type"].(string) != "before_review" {
		t.Fatalf("expected updated milestone type, got %#v", updatedMilestone["milestone_type"])
	}
	if updatedMilestone["locked"].(bool) != false {
		t.Fatalf("expected updated milestone lock false, got %#v", updatedMilestone["locked"])
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodDelete, "/api/milestones/"+milestoneID, nil)
	app.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected delete milestone status 204, got %d", response.Code)
	}

	milestonesResponse := requestJSON(t, app.Handler(), http.MethodGet, "/api/books/"+bookID+"/milestones", nil)
	milestones, _ = milestonesResponse["milestones"].([]any)
	if len(milestones) != 0 {
		t.Fatalf("expected milestone deletion, got %d items", len(milestones))
	}

	initialEventsResponse := requestJSON(t, app.Handler(), http.MethodGet, "/api/revisions/"+initialRevisionID+"/events", nil)
	initialEvents, _ := initialEventsResponse["events"].([]any)
	if len(initialEvents) < 4 {
		t.Fatalf("expected milestone lifecycle events on initial revision, got %d", len(initialEvents))
	}

	restoredEvents := requestJSON(t, app.Handler(), http.MethodGet, "/api/revisions/"+restoredRevisionID+"/events", nil)["events"].([]any)
	if len(restoredEvents) == 0 {
		t.Fatalf("expected restore events for restored revision")
	}
	if restoredEvents[0].(map[string]any)["event_type"].(string) != "restore_performed" {
		t.Fatalf("expected restore_performed event, got %#v", restoredEvents[0])
	}
}

func TestChapterSaveCreatesAutosaveDraftsAndManualRevisions(t *testing.T) {
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
		"title": "Save-Projekt",
	})["id"].(string)

	bookID := createJSON(t, app.Handler(), http.MethodPost, "/api/projects/"+projectID+"/books", map[string]any{
		"title": "Save-Buch",
	})["id"].(string)

	chapterID := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/chapters", map[string]any{
		"title":            "Kapitel Save",
		"markdown_content": "# Save\n\nStart.",
		"editor_json":      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Start."}]}]}`,
	})["id"].(string)

	requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/chapters/"+chapterID, map[string]any{
		"title":            "Kapitel Save",
		"markdown_content": "# Save\n\nAutosave Zwischenstand.",
		"editor_json":      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Autosave Zwischenstand."}]}]}`,
		"save_mode":        "autosave",
		"autosave_reason":  "idle_autosave",
		"session_id":       "session-1",
	}, http.StatusOK)

	autosaves := requestJSON(t, app.Handler(), http.MethodGet, "/api/chapters/"+chapterID+"/autosaves", nil)["autosaves"].([]any)
	if len(autosaves) != 1 {
		t.Fatalf("expected one autosave draft, got %d", len(autosaves))
	}
	firstDraft := autosaves[0].(map[string]any)
	if firstDraft["reason"].(string) != "idle_autosave" {
		t.Fatalf("expected idle autosave reason, got %#v", firstDraft["reason"])
	}
	if firstDraft["session_id"].(string) != "session-1" {
		t.Fatalf("expected autosave session_id, got %#v", firstDraft["session_id"])
	}

	revisionsAfterAutosave := requestJSON(t, app.Handler(), http.MethodGet, "/api/chapters/"+chapterID+"/revisions", nil)["revisions"].([]any)
	if len(revisionsAfterAutosave) != 0 {
		t.Fatalf("expected no revisions after autosave, got %d", len(revisionsAfterAutosave))
	}

	requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/chapters/"+chapterID, map[string]any{
		"title":                "Kapitel Save",
		"markdown_content":     "# Save\n\nManueller Sicherungsstand.",
		"editor_json":          `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Manueller Sicherungsstand."}]}]}`,
		"save_mode":            "manual",
		"autosave_reason":      "manual_save",
		"session_id":           "session-1",
		"create_revision":      true,
		"revision_type":        "manual",
		"revision_title":       "Bewusster Speicherpunkt",
		"revision_description": "Wird direkt aus dem normalen Speichern erzeugt.",
		"created_by":           "Tester",
	}, http.StatusOK)

	revisions := requestJSON(t, app.Handler(), http.MethodGet, "/api/chapters/"+chapterID+"/revisions", nil)["revisions"].([]any)
	if len(revisions) != 1 {
		t.Fatalf("expected one revision after manual save, got %d", len(revisions))
	}
	firstRevision := revisions[0].(map[string]any)
	if firstRevision["title"].(string) != "Bewusster Speicherpunkt" {
		t.Fatalf("expected revision title from save flow, got %#v", firstRevision["title"])
	}
	if firstRevision["created_by"].(string) != "Tester" {
		t.Fatalf("expected created_by on revision, got %#v", firstRevision["created_by"])
	}

	autosaves = requestJSON(t, app.Handler(), http.MethodGet, "/api/chapters/"+chapterID+"/autosaves", nil)["autosaves"].([]any)
	if len(autosaves) != 2 {
		t.Fatalf("expected two autosave drafts after manual save, got %d", len(autosaves))
	}
	latestDraft := autosaves[0].(map[string]any)
	if latestDraft["reason"].(string) != "manual_save" {
		t.Fatalf("expected manual_save draft reason, got %#v", latestDraft["reason"])
	}
}

func TestGetUnknownRevisionReturnsNotFound(t *testing.T) {
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

	app := httpapi.New(config.Config{}, appStore)
	request := httptest.NewRequest(http.MethodGet, "/api/revisions/does-not-exist", nil)
	response := httptest.NewRecorder()
	app.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d with body %s", response.Code, response.Body.String())
	}
}
