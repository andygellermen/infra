package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/config"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/db"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/httpapi"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

func TestAuthorFlowEndpoints(t *testing.T) {
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
		"title":       "Testprojekt",
		"description": "Konzepttest",
	})["id"].(string)

	updatedProject := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/projects/"+projectID, map[string]any{
		"title":       "Testprojekt Alpha",
		"description": "Geschaerfter Projektrahmen",
	}, http.StatusOK)
	if updatedProject["title"].(string) != "Testprojekt Alpha" {
		t.Fatalf("expected updated project title, got %#v", updatedProject["title"])
	}
	if updatedProject["description"].(string) != "Geschaerfter Projektrahmen" {
		t.Fatalf("expected updated project description, got %#v", updatedProject["description"])
	}

	bookID := createJSON(t, app.Handler(), http.MethodPost, "/api/projects/"+projectID+"/books", map[string]any{
		"title":      "Testbuch",
		"subtitle":   "Erste Beschreibung",
		"author":     "Codex",
		"visibility": "private",
	})["id"].(string)

	updatedBook := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/books/"+bookID, map[string]any{
		"title":      "Testbuch",
		"subtitle":   "Geschaerfte Positionierung",
		"author":     "Codex",
		"visibility": "shared",
	}, http.StatusOK)
	if updatedBook["subtitle"].(string) != "Geschaerfte Positionierung" {
		t.Fatalf("expected updated subtitle, got %#v", updatedBook["subtitle"])
	}
	if updatedBook["visibility"].(string) != "shared" {
		t.Fatalf("expected updated visibility, got %#v", updatedBook["visibility"])
	}

	knowledgeItemID := createJSON(t, app.Handler(), http.MethodPost, "/api/projects/"+projectID+"/knowledge-items", map[string]any{
		"type":    "person",
		"name":    "Mara",
		"summary": "Mutige Protagonistin",
		"tags":    []string{"figur", "perspektive"},
	})["id"].(string)

	chapterID := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/chapters", map[string]any{
		"title":            "Kapitel A",
		"markdown_content": "# Hallo\n\nDies ist ein Test.",
		"editor_json":      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Dies ist ein Test."}]}]}`,
	})["id"].(string)
	chapterTwoID := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/chapters", map[string]any{
		"title":            "Kapitel B",
		"markdown_content": "# Zweites Kapitel\n\nNoch ein Test.",
		"editor_json":      `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Noch ein Test."}]}]}`,
	})["id"].(string)

	reordered := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/books/"+bookID+"/chapters/reorder", map[string]any{
		"chapter_ids": []string{chapterTwoID, chapterID},
	}, http.StatusOK)
	reorderedChapters := reordered["chapters"].([]any)
	if reorderedChapters[0].(map[string]any)["id"].(string) != chapterTwoID {
		t.Fatalf("expected reordered chapter first, got %#v", reorderedChapters[0])
	}

	_ = createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/workflow-boxes", map[string]any{
		"title":        "Notizen",
		"type":         "notes",
		"is_collapsed": false,
	})

	boxes := requestJSON(t, app.Handler(), http.MethodGet, "/api/books/"+bookID+"/workflow-boxes", nil)
	boxID := boxes["workflow_boxes"].([]any)[0].(map[string]any)["id"].(string)

	anchor := createJSON(t, app.Handler(), http.MethodPost, "/api/chapters/"+chapterID+"/anchors", map[string]any{
		"workflow_box_id": boxID,
		"anchor_type":     "passage",
		"selected_text":   "Dies ist ein Test.",
		"start_offset":    0,
		"end_offset":      18,
		"note":            "Wichtige Passage",
	})
	anchorID := anchor["id"].(string)

	clipboard := createJSON(t, app.Handler(), http.MethodPost, "/api/books/"+bookID+"/clipboard", map[string]any{
		"chapter_id":       chapterID,
		"content":          "Dies ist ein Test.",
		"content_type":     "text/markdown",
		"source_anchor_id": anchorID,
		"is_pinned":        true,
		"slot":             1,
	})
	if clipboard["slot"].(float64) != 1 {
		t.Fatalf("expected slot 1, got %#v", clipboard["slot"])
	}

	updatedKnowledge := requestJSONWithStatus(t, app.Handler(), http.MethodPut, "/api/knowledge-items/"+knowledgeItemID, map[string]any{
		"type":    "person",
		"name":    "Mara",
		"summary": "Mutige Hauptfigur",
		"body":    "Kennt den stillen Garten.",
		"tags":    []string{"figur", "demo"},
	}, http.StatusOK)
	if updatedKnowledge["summary"].(string) != "Mutige Hauptfigur" {
		t.Fatalf("expected updated summary, got %#v", updatedKnowledge["summary"])
	}

	snapshotPath := filepath.Join(tempDir, "library")
	if _, err := os.Stat(snapshotPath); err != nil {
		t.Fatalf("expected markdown snapshot root to exist: %v", err)
	}

	bookBundle := requestJSON(t, app.Handler(), http.MethodGet, "/api/books/"+bookID, nil)
	if len(bookBundle["chapters"].([]any)) == 0 {
		t.Fatalf("expected chapters in book bundle")
	}

	knowledgeItems := requestJSON(t, app.Handler(), http.MethodGet, "/api/projects/"+projectID+"/knowledge-items", nil)
	if len(knowledgeItems["knowledge_items"].([]any)) == 0 {
		t.Fatalf("expected knowledge items in project response")
	}
}

func createJSON(t *testing.T, handler http.Handler, method, path string, payload any) map[string]any {
	t.Helper()
	response := requestJSONWithStatus(t, handler, method, path, payload, http.StatusCreated)
	return response
}

func requestJSON(t *testing.T, handler http.Handler, method, path string, payload any) map[string]any {
	t.Helper()
	return requestJSONWithStatus(t, handler, method, path, payload, http.StatusOK)
}

func requestJSONWithStatus(t *testing.T, handler http.Handler, method, path string, payload any, expectedStatus int) map[string]any {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode payload: %v", err)
		}
	}

	request := httptest.NewRequest(method, path, &body)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != expectedStatus {
		t.Fatalf("%s %s: expected status %d, got %d with body %s", method, path, expectedStatus, response.Code, response.Body.String())
	}

	var decoded map[string]any
	if response.Body.Len() == 0 {
		return decoded
	}
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return decoded
}
