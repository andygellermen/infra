package httpapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

type pingerStub struct{ err error }

func (p pingerStub) PingContext(context.Context) error { return p.err }

func TestAnalyzeUsesCoreEngine(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", strings.NewReader(`{
  "text":"Ich muss das heute unbedingt noch schaffen.",
  "context":"SELF_TALK",
  "locale":"de-DE",
  "input_mode":"TEXT",
  "presentation_profile":"PRIVATE",
  "analysis_mode":"STANDARD"
}`))
	response := httptest.NewRecorder()
	New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var result analysis.Result
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result.ResolvedSenses) != 1 || result.ResolvedSenses[0].Sense != "INTERNAL_PRESSURE" {
		t.Fatalf("unexpected analysis response: %+v", result)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("analysis response must not be cacheable")
	}
}

func TestAnalyzeRejectsUnknownField(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", strings.NewReader(`{"text":"Hallo","person_id":"42"}`))
	response := httptest.NewRecorder()
	New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", response.Code, http.StatusBadRequest)
	}
}

func TestAnalyzeRejectsUnsupportedInputMode(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", strings.NewReader(`{"text":"Hallo","input_mode":"DIRECT_AUDIO"}`))
	response := httptest.NewRecorder()
	New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d; want %d", response.Code, http.StatusUnprocessableEntity)
	}
}

func TestReadinessReflectsDatabase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "ready", want: http.StatusOK},
		{name: "unavailable", err: errors.New("down"), want: http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			response := httptest.NewRecorder()
			New(analysis.NewDefault(), pingerStub{err: test.err}, 64<<10).Handler().ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d; want %d", response.Code, test.want)
			}
		})
	}
}
