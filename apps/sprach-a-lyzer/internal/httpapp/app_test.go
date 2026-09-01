package httpapp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/version"
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
	if response.Header().Get("X-Sprach-A-Lyzer-Version") != version.Core ||
		response.Header().Get("X-Sprach-A-Lyzer-Contract-Version") != "" {
		t.Fatalf("v1 version headers = core %q, contract %q", response.Header().Get("X-Sprach-A-Lyzer-Version"), response.Header().Get("X-Sprach-A-Lyzer-Contract-Version"))
	}
}

func TestPublicV02ResolverAndTraceContracts(t *testing.T) {
	t.Parallel()

	payload := `{
  "text":"Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.",
  "context":"PRIVATE_CONVERSATION"
}`
	handler := New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler()

	resolveResponse := httptest.NewRecorder()
	handler.ServeHTTP(resolveResponse, httptest.NewRequest(http.MethodPost, "/api/v2/resolve", strings.NewReader(payload)))
	if resolveResponse.Code != http.StatusOK {
		t.Fatalf("resolve status = %d, body = %s", resolveResponse.Code, resolveResponse.Body.String())
	}
	var resolution analysis.ResolverResult
	if err := json.NewDecoder(resolveResponse.Body).Decode(&resolution); err != nil {
		t.Fatalf("decode resolver response: %v", err)
	}
	assertVersionHeaders(t, resolveResponse, "resolver-result", "0.2")
	if resolution.ContractVersion != "0.2" || len(resolution.PropositionGraph.Nodes) != 2 {
		t.Fatalf("resolver v0.2 response = %+v", resolution)
	}

	traceResponse := httptest.NewRecorder()
	handler.ServeHTTP(traceResponse, httptest.NewRequest(http.MethodPost, "/api/v2/trace", strings.NewReader(payload)))
	if traceResponse.Code != http.StatusOK {
		t.Fatalf("trace status = %d, body = %s", traceResponse.Code, traceResponse.Body.String())
	}
	var trace analysis.TraceV02
	if err := json.NewDecoder(traceResponse.Body).Decode(&trace); err != nil {
		t.Fatalf("decode trace response: %v", err)
	}
	assertVersionHeaders(t, traceResponse, "analysis-trace", "0.2")
	if trace.ContractVersion != "0.2" || len(trace.Propositions) != 2 {
		t.Fatalf("trace v0.2 response = %+v", trace)
	}
	foundRespectfulBoundary := false
	for _, contribution := range trace.Contributions {
		if contribution.RuleID == "R-RESPECTFUL-BOUNDARY" {
			foundRespectfulBoundary = true
			if !slices.Equal(contribution.PropositionIDs, []string{"P0", "P1"}) {
				t.Fatalf("respectful-boundary proposition IDs = %v; want P0/P1", contribution.PropositionIDs)
			}
		}
	}
	if !foundRespectfulBoundary {
		t.Fatal("trace v0.2 lacks respectful-boundary contribution")
	}
}

func TestPublicV02RoutesUseStrictSharedRequestContract(t *testing.T) {
	t.Parallel()

	handler := New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler()
	for _, route := range []string{"/api/v2/resolve", "/api/v2/trace"} {
		t.Run(route, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, route, strings.NewReader(`{"text":"Hallo","person_id":"42"}`))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d; want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

func assertVersionHeaders(t *testing.T, response *httptest.ResponseRecorder, contract, contractVersion string) {
	t.Helper()
	if response.Header().Get("X-Sprach-A-Lyzer-Version") != version.Core ||
		response.Header().Get("X-Sprach-A-Lyzer-Contract") != contract ||
		response.Header().Get("X-Sprach-A-Lyzer-Contract-Version") != contractVersion {
		t.Fatalf("version headers = core %q, contract %q, contract version %q",
			response.Header().Get("X-Sprach-A-Lyzer-Version"),
			response.Header().Get("X-Sprach-A-Lyzer-Contract"),
			response.Header().Get("X-Sprach-A-Lyzer-Contract-Version"))
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
