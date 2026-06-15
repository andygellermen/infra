package httpapi

import (
	"net/http"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

func (a *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.GetDashboard(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": items})
}

func (a *App) handleGetBookStructure(w http.ResponseWriter, r *http.Request) {
	item, err := a.store.GetBookStructure(r.Context(), r.PathValue("bookId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleUpdateChapterOptions(w http.ResponseWriter, r *http.Request) {
	var input store.UpdateChapterOptionsInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.UpdateChapterOptions(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleListReusablePages(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListReusableBookPages(r.Context(), r.PathValue("bookId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reusable_pages": items})
}

func (a *App) handleCreateReusablePage(w http.ResponseWriter, r *http.Request) {
	var input store.CreateReusableBookPageInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateReusableBookPage(r.Context(), r.PathValue("bookId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleUpdateReusablePage(w http.ResponseWriter, r *http.Request) {
	var input store.UpdateReusableBookPageInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.UpdateReusableBookPage(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleDeleteReusablePage(w http.ResponseWriter, r *http.Request) {
	if err := a.store.DeleteReusableBookPage(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleListThemes(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.ListThemes(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"themes": items})
}

func (a *App) handleCreateTheme(w http.ResponseWriter, r *http.Request) {
	var input store.CreateThemeInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateTheme(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleUpdateTheme(w http.ResponseWriter, r *http.Request) {
	var input store.UpdateThemeInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.UpdateTheme(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (a *App) handleListProofingChecks(w http.ResponseWriter, r *http.Request) {
	items, err := a.store.GetBookProofingChecks(r.Context(), r.PathValue("bookId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"checks": items})
}
