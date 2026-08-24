package httpapp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/engine"
)

const maxTextRunes = 10_000

type Analyzer interface {
	Analyze(domain.AnalysisRequest) (domain.AnalysisResult, error)
}

type Pinger interface {
	PingContext(context.Context) error
}

type App struct {
	analyzer        Analyzer
	database        Pinger
	maxRequestBytes int64
	handler         http.Handler
}

func New(analyzer Analyzer, database Pinger, maxRequestBytes int64) *App {
	app := &App{analyzer: analyzer, database: database, maxRequestBytes: maxRequestBytes}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", app.live)
	mux.HandleFunc("GET /health/ready", app.ready)
	mux.HandleFunc("POST /api/v1/analyze", app.analyze)
	app.handler = securityHeaders(mux)
	return app
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) live(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) ready(response http.ResponseWriter, request *http.Request) {
	if a.database == nil || a.database.PingContext(request.Context()) != nil {
		writeJSON(response, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(response, http.StatusOK, map[string]string{"status": "ready"})
}

func (a *App) analyze(response http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(response, request.Body, a.maxRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	var input domain.AnalysisRequest
	if err := decoder.Decode(&input); err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "request body too large") {
			status = http.StatusRequestEntityTooLarge
		}
		writeError(response, status, "INVALID_REQUEST", "Der Request enthält kein gültiges Analyseobjekt.")
		return
	}
	if err := ensureEOF(decoder); err != nil {
		writeError(response, http.StatusBadRequest, "INVALID_REQUEST", "Der Request darf nur ein JSON-Objekt enthalten.")
		return
	}
	applyRequestDefaults(&input)
	if validationCode, message := validateRequest(input); validationCode != "" {
		writeError(response, http.StatusUnprocessableEntity, validationCode, message)
		return
	}

	result, err := a.analyzer.Analyze(input)
	if err != nil {
		if errors.Is(err, engine.ErrEmptyText) {
			writeError(response, http.StatusUnprocessableEntity, "EMPTY_TEXT", "Der Analysetext darf nicht leer sein.")
			return
		}
		writeError(response, http.StatusInternalServerError, "ANALYSIS_FAILED", "Die Analyse konnte nicht ausgeführt werden.")
		return
	}
	response.Header().Set("X-Sprach-A-Lyzer-Version", "0.1.0")
	writeJSON(response, http.StatusOK, result)
}

func applyRequestDefaults(input *domain.AnalysisRequest) {
	input.Locale = strings.TrimSpace(input.Locale)
	if input.Locale == "" {
		input.Locale = "de-DE"
	}
	input.InputMode = strings.ToUpper(strings.TrimSpace(input.InputMode))
	if input.InputMode == "" {
		input.InputMode = "TEXT"
	}
	input.PresentationProfile = strings.ToUpper(strings.TrimSpace(input.PresentationProfile))
	if input.PresentationProfile == "" {
		input.PresentationProfile = "PRIVATE"
	}
	input.AnalysisMode = strings.ToUpper(strings.TrimSpace(input.AnalysisMode))
	if input.AnalysisMode == "" {
		input.AnalysisMode = "STANDARD"
	}
}

func validateRequest(input domain.AnalysisRequest) (string, string) {
	if strings.TrimSpace(input.Text) == "" {
		return "EMPTY_TEXT", "Der Analysetext darf nicht leer sein."
	}
	if !utf8.ValidString(input.Text) || utf8.RuneCountInString(input.Text) > maxTextRunes {
		return "TEXT_TOO_LONG", "Der Analysetext darf höchstens 10.000 Zeichen enthalten."
	}
	if input.Locale != "de-DE" {
		return "UNSUPPORTED_LOCALE", "Der erste Vertical Slice unterstützt ausschließlich de-DE."
	}
	if input.InputMode != "TEXT" {
		return "UNSUPPORTED_INPUT_MODE", "Der erste Vertical Slice unterstützt ausschließlich TEXT."
	}
	if input.PresentationProfile != "PRIVATE" && input.PresentationProfile != "CORPORATE" {
		return "UNSUPPORTED_PRESENTATION_PROFILE", "Das Präsentationsprofil muss PRIVATE oder CORPORATE sein."
	}
	if input.AnalysisMode != "STANDARD" {
		return "UNSUPPORTED_ANALYSIS_MODE", "Der erste Vertical Slice unterstützt ausschließlich STANDARD."
	}
	return "", ""
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json; charset=utf-8")
		response.Header().Set("Cache-Control", "no-store")
		response.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(response, request)
	})
}

func writeError(response http.ResponseWriter, status int, code, message string) {
	writeJSON(response, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}
