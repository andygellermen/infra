package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

func (a *App) handleGetKanban(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	query := store.KanbanQuery{}
	if values.Has("book_ids") {
		raw := values.Get("book_ids")
		if strings.TrimSpace(raw) == "" {
			writeClientError(w, http.StatusBadRequest, "book_ids must not be empty")
			return
		}
		for _, id := range strings.Split(raw, ",") {
			id = strings.TrimSpace(id)
			if id == "" {
				writeClientError(w, http.StatusBadRequest, "book_ids contains an empty id")
				return
			}
			query.BookIDs = append(query.BookIDs, id)
		}
	}
	if values.Has("limit") {
		limit, err := strconv.Atoi(values.Get("limit"))
		if err != nil || limit < 1 || limit > 20 {
			writeClientError(w, http.StatusBadRequest, "limit must be between 1 and 20")
			return
		}
		query.LimitPerPhase = limit
	}
	if values.Has("include_done") {
		switch values.Get("include_done") {
		case "true":
			query.IncludeDone = true
		case "false":
		default:
			writeClientError(w, http.StatusBadRequest, "include_done must be true or false")
			return
		}
	}
	result, err := a.store.ListKanban(r.Context(), query)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) handleCreateWorkItem(w http.ResponseWriter, r *http.Request) {
	var input store.CreateWorkItemInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.CreateWorkItem(r.Context(), r.PathValue("bookId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (a *App) handleMoveWorkItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Phase string `json:"phase"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := a.store.MoveWorkItem(r.Context(), r.PathValue("id"), input.Phase)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
