package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/config"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

type App struct {
	cfg   config.Config
	store *store.Store
	mux   *http.ServeMux
}

func New(cfg config.Config, appStore *store.Store) *App {
	app := &App{
		cfg:   cfg,
		store: appStore,
		mux:   http.NewServeMux(),
	}
	app.routes()
	return app
}

func (a *App) Handler() http.Handler {
	return a.withCORS(a.logRequests(a.mux))
}

func (a *App) routes() {
	a.mux.HandleFunc("GET /api/health", a.handleHealth)
	a.mux.HandleFunc("GET /api/projects", a.handleListProjects)
	a.mux.HandleFunc("POST /api/projects", a.handleCreateProject)
	a.mux.HandleFunc("GET /api/projects/{id}", a.handleGetProject)
	a.mux.HandleFunc("GET /api/projects/{projectId}/knowledge-items", a.handleListKnowledgeItems)
	a.mux.HandleFunc("POST /api/projects/{projectId}/knowledge-items", a.handleCreateKnowledgeItem)
	a.mux.HandleFunc("GET /api/books/{id}", a.handleGetBook)
	a.mux.HandleFunc("POST /api/projects/{projectId}/books", a.handleCreateBook)
	a.mux.HandleFunc("GET /api/books/{bookId}/chapters", a.handleListChapters)
	a.mux.HandleFunc("POST /api/books/{bookId}/chapters", a.handleCreateChapter)
	a.mux.HandleFunc("GET /api/chapters/{id}", a.handleGetChapter)
	a.mux.HandleFunc("PUT /api/chapters/{id}", a.handleUpdateChapter)
	a.mux.HandleFunc("GET /api/books/{bookId}/workflow-boxes", a.handleListWorkflowBoxes)
	a.mux.HandleFunc("POST /api/books/{bookId}/workflow-boxes", a.handleCreateWorkflowBox)
	a.mux.HandleFunc("PUT /api/workflow-boxes/{id}", a.handleUpdateWorkflowBox)
	a.mux.HandleFunc("GET /api/chapters/{chapterId}/anchors", a.handleListAnchors)
	a.mux.HandleFunc("POST /api/chapters/{chapterId}/anchors", a.handleCreateAnchor)
	a.mux.HandleFunc("DELETE /api/anchors/{id}", a.handleDeleteAnchor)
	a.mux.HandleFunc("GET /api/books/{bookId}/clipboard", a.handleListClipboard)
	a.mux.HandleFunc("POST /api/books/{bookId}/clipboard", a.handleCreateClipboard)
	a.mux.HandleFunc("PUT /api/clipboard/{id}", a.handleUpdateClipboard)
	a.mux.HandleFunc("DELETE /api/clipboard/{id}", a.handleDeleteClipboard)
	a.mux.HandleFunc("PUT /api/knowledge-items/{id}", a.handleUpdateKnowledgeItem)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *App) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := a.store.ListProjects(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (a *App) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var input store.CreateProjectInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	project, err := a.store.CreateProject(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (a *App) handleGetProject(w http.ResponseWriter, r *http.Request) {
	project, books, err := a.store.GetProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"project": project,
		"books":   books,
	})
}

func (a *App) handleListKnowledgeItems(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListKnowledgeItems(r.Context(), r.PathValue("projectId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"knowledge_items": items})
}

func (a *App) handleCreateKnowledgeItem(w http.ResponseWriter, r *http.Request) {
	var input store.CreateKnowledgeItemInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateKnowledgeItem(r.Context(), r.PathValue("projectId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleGetBook(w http.ResponseWriter, r *http.Request) {
	bundle, err := a.store.GetBook(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bundle)
}

func (a *App) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var input store.CreateBookInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	book, err := a.store.CreateBook(r.Context(), r.PathValue("projectId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, book)
}

func (a *App) handleListChapters(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListChapters(r.Context(), r.PathValue("bookId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"chapters": items})
}

func (a *App) handleCreateChapter(w http.ResponseWriter, r *http.Request) {
	var input store.CreateChapterInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateChapter(r.Context(), r.PathValue("bookId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleGetChapter(w http.ResponseWriter, r *http.Request) {
	item, err := a.store.GetChapter(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleUpdateChapter(w http.ResponseWriter, r *http.Request) {
	var input store.UpdateChapterInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.UpdateChapter(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleListWorkflowBoxes(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListWorkflowBoxes(r.Context(), r.PathValue("bookId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workflow_boxes": items})
}

func (a *App) handleCreateWorkflowBox(w http.ResponseWriter, r *http.Request) {
	var input store.CreateWorkflowBoxInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateWorkflowBox(r.Context(), r.PathValue("bookId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleUpdateWorkflowBox(w http.ResponseWriter, r *http.Request) {
	var input store.UpdateWorkflowBoxInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.UpdateWorkflowBox(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleListAnchors(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListAnchors(r.Context(), r.PathValue("chapterId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"anchors": items})
}

func (a *App) handleCreateAnchor(w http.ResponseWriter, r *http.Request) {
	var input store.CreateAnchorInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateAnchor(r.Context(), r.PathValue("chapterId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleDeleteAnchor(w http.ResponseWriter, r *http.Request) {
	if err := a.store.DeleteAnchor(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleListClipboard(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListClipboardItems(r.Context(), r.PathValue("bookId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"clipboard": items})
}

func (a *App) handleCreateClipboard(w http.ResponseWriter, r *http.Request) {
	var input store.CreateClipboardItemInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateClipboardItem(r.Context(), r.PathValue("bookId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleUpdateClipboard(w http.ResponseWriter, r *http.Request) {
	var input store.UpdateClipboardItemInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.UpdateClipboardItem(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleDeleteClipboard(w http.ResponseWriter, r *http.Request) {
	if err := a.store.DeleteClipboardItem(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleUpdateKnowledgeItem(w http.ResponseWriter, r *http.Request) {
	var input store.UpdateKnowledgeItemInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.UpdateKnowledgeItem(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := strings.TrimSpace(a.cfg.AllowedOrigin); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started).Round(time.Millisecond))
	})
}

func decodeJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeClientError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeClientError(w, http.StatusNotFound, "resource not found")
	default:
		writeClientError(w, http.StatusInternalServerError, err.Error())
	}
}

func PingContext(ctx context.Context, handler http.Handler) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/health", nil)
	if err != nil {
		return err
	}
	recorder := dummyResponseWriter{}
	handler.ServeHTTP(recorder, request)
	return nil
}

type dummyResponseWriter struct{}

func (dummyResponseWriter) Header() http.Header       { return http.Header{} }
func (dummyResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (dummyResponseWriter) WriteHeader(int)           {}
