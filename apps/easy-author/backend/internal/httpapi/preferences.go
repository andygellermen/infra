package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/model"
)

type updateBookPresentationInput struct {
	DefaultWorkView     string          `json:"default_work_view"`
	TypographyOverrides json.RawMessage `json:"typography_overrides"`
}

func (a *App) handleGetBookPresentation(w http.ResponseWriter, r *http.Request) {
	presentation, err := a.store.GetBookPresentation(r.Context(), r.PathValue("bookId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, presentation)
}

func (a *App) handleUpdateBookPresentation(w http.ResponseWriter, r *http.Request) {
	var input updateBookPresentationInput
	if err := decodeJSON(r, &input); err != nil {
		writeClientError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.DefaultWorkView != "clean" && input.DefaultWorkView != "intense" && input.DefaultWorkView != "review" {
		writeClientError(w, http.StatusBadRequest, "default_work_view must be clean, intense, or review")
		return
	}
	if len(input.TypographyOverrides) == 0 {
		input.TypographyOverrides = json.RawMessage(`{}`)
	}
	var typographyObject map[string]json.RawMessage
	if err := json.Unmarshal(input.TypographyOverrides, &typographyObject); err != nil || typographyObject == nil {
		writeClientError(w, http.StatusBadRequest, "typography_overrides must be a JSON object")
		return
	}
	presentation, err := a.store.UpdateBookPresentation(r.Context(), model.BookPresentation{
		BookID:              r.PathValue("bookId"),
		DefaultWorkView:     input.DefaultWorkView,
		TypographyOverrides: input.TypographyOverrides,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, presentation)
}
