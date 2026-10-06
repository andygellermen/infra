package httpapi

import (
	"net/http"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

func (a *App) handleListContexts(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListChapterContexts(r.Context(), r.PathValue("chapterId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"contexts": items})
}

func (a *App) handleCreateContext(w http.ResponseWriter, r *http.Request) {
	var input store.CreateContextInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateContext(r.Context(), r.PathValue("chapterId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleDeleteContext(w http.ResponseWriter, r *http.Request) {
	if err := a.store.DeleteContext(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleUpdateThread(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	thread, err := a.store.UpdateThreadStatus(r.Context(), r.PathValue("id"), input.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, thread)
}

func (a *App) handleCreateThreadMessage(w http.ResponseWriter, r *http.Request) {
	var input store.CreateThreadMessageInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	message, err := a.store.CreateThreadMessage(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, message)
}
