package httpapp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/questions"
)

func (a *App) selectQuestions(response http.ResponseWriter, request *http.Request) {
	input, ok := decodeStrictJSON[questions.SelectionRequest](a, response, request)
	if !ok {
		return
	}
	result, err := a.questions.Select(input)
	if err != nil {
		writeError(response, http.StatusUnprocessableEntity, "QUESTION_SELECTION_FAILED", err.Error())
		return
	}
	writeVersionedJSON(response, "question-selection", result.ContractVersion, result)
}

func (a *App) analyzeAnswer(response http.ResponseWriter, request *http.Request) {
	input, ok := decodeStrictJSON[questions.AnswerRequest](a, response, request)
	if !ok {
		return
	}
	result, err := a.questions.Analyze(input)
	if err != nil {
		if errors.Is(err, questions.ErrUnknownQuestion) || errors.Is(err, questions.ErrEmptyAnswer) ||
			errors.Is(err, questions.ErrAnswerTooLong) || errors.Is(err, questions.ErrInvalidProfile) {
			writeError(response, http.StatusUnprocessableEntity, "INVALID_QUESTION_ANSWER", err.Error())
			return
		}
		writeError(response, http.StatusInternalServerError, "QUESTION_ANSWER_FAILED", "Die Frage-Antwort-Analyse konnte nicht ausgeführt werden.")
		return
	}
	writeVersionedJSON(response, "question-answer-observation", result.ContractVersion, result)
}

func (a *App) composeSession(response http.ResponseWriter, request *http.Request) {
	input, ok := decodeStrictJSON[questions.SessionRequest](a, response, request)
	if !ok {
		return
	}
	result, err := a.questions.ComposeSession(input)
	if err != nil {
		writeError(response, http.StatusUnprocessableEntity, "QUESTION_SESSION_FAILED", err.Error())
		return
	}
	writeVersionedJSON(response, "question-session", result.ContractVersion, result)
}

func (a *App) renderQuestion(response http.ResponseWriter, request *http.Request) {
	input, ok := decodeStrictJSON[questions.RenderRequest](a, response, request)
	if !ok {
		return
	}
	result, err := a.questions.Render(input)
	if err != nil {
		if errors.Is(err, questions.ErrUnknownQuestion) || errors.Is(err, questions.ErrInvalidProfile) ||
			errors.Is(err, questions.ErrInvalidRenderAction) {
			writeError(response, http.StatusUnprocessableEntity, "INVALID_RENDER_REQUEST", err.Error())
			return
		}
		writeError(response, http.StatusInternalServerError, "QUESTION_RENDERING_FAILED", "Die Fragenformulierung konnte nicht sicher aufgelöst werden.")
		return
	}
	writeVersionedJSON(response, "question-rendering-result", result.ContractVersion, result)
}

func decodeStrictJSON[T any](a *App, response http.ResponseWriter, request *http.Request) (T, bool) {
	var input T
	request.Body = http.MaxBytesReader(response, request.Body, a.maxRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "request body too large") {
			status = http.StatusRequestEntityTooLarge
		}
		writeError(response, status, "INVALID_REQUEST", "Der Request enthält kein gültiges JSON-Objekt.")
		return input, false
	}
	if err := ensureEOF(decoder); err != nil {
		writeError(response, http.StatusBadRequest, "INVALID_REQUEST", "Der Request darf nur ein JSON-Objekt enthalten.")
		return input, false
	}
	return input, true
}
