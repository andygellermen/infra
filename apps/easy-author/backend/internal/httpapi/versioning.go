package httpapi

import (
	"net/http"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

func (a *App) handleListRevisions(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListRevisionsByChapter(r.Context(), r.PathValue("chapterId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"revisions": items})
}

func (a *App) handleListAutosaveDrafts(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListAutosaveDrafts(r.Context(), r.PathValue("chapterId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"autosaves": items})
}

func (a *App) handleCreateRevision(w http.ResponseWriter, r *http.Request) {
	var input store.CreateRevisionInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateRevision(r.Context(), r.PathValue("chapterId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleGetRevision(w http.ResponseWriter, r *http.Request) {
	item, err := a.store.GetRevision(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleRestoreRevision(w http.ResponseWriter, r *http.Request) {
	var input store.RestoreRevisionInput
	if r.ContentLength != 0 {
		if err := decodeJSON(r, &input); err != nil {
			writeClientError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	item, err := a.store.RestoreRevision(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleListMilestones(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListMilestonesByBook(r.Context(), r.PathValue("bookId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"milestones": items})
}

func (a *App) handleCreateMilestone(w http.ResponseWriter, r *http.Request) {
	var input store.CreateMilestoneInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateMilestone(r.Context(), r.PathValue("bookId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleUpdateMilestone(w http.ResponseWriter, r *http.Request) {
	var input store.UpdateMilestoneInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.UpdateMilestone(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleDeleteMilestone(w http.ResponseWriter, r *http.Request) {
	if err := a.store.DeleteMilestone(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleListRevisionEvents(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListRevisionEvents(r.Context(), r.PathValue("revisionId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": items})
}
