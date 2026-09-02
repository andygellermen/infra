package httpapp

import (
	"errors"
	"net/http"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/experience"
)

func (a *App) analyzeExperience(response http.ResponseWriter, request *http.Request) {
	input, ok := decodeStrictJSON[experience.Request](a, response, request)
	if !ok {
		return
	}
	result, err := a.experience.Analyze(input)
	if err != nil {
		if errors.Is(err, experience.ErrEmptyText) || errors.Is(err, experience.ErrTextTooLong) ||
			errors.Is(err, experience.ErrInvalidProfile) || errors.Is(err, experience.ErrInvalidLanguage) ||
			errors.Is(err, experience.ErrInvalidContext) {
			writeError(response, http.StatusUnprocessableEntity, "INVALID_EXPERIENCE_REQUEST", err.Error())
			return
		}
		writeError(response, http.StatusInternalServerError, "EXPERIENCE_FAILED", "Die transiente Core-Erfahrung konnte nicht erzeugt werden.")
		return
	}
	writeVersionedJSON(response, "mvp-experience-result", result.ContractVersion, result)
}
