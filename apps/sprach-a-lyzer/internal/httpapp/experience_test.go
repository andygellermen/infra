package httpapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/experience"
)

func TestPublicV06MVPExperienceIsTransientNoAIAndExplainable(t *testing.T) {
	t.Parallel()
	handler := New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler()
	request := httptest.NewRequest(http.MethodPost, "/api/v6/experience/analyze", strings.NewReader(`{"text":"Ich muss das heute unbedingt noch schaffen.","context":"SELF_TALK","profile":"PRIVATE","language_level":"STANDARD"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var result experience.Result
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	assertVersionHeaders(t, response, "mvp-experience-result", experience.ContractVersion)
	if result.ExperienceMode != "CORE_NO_AI" || result.Privacy.RawTextStored || result.Privacy.AnalysisStored || result.Privacy.ExternalTransfer || result.Privacy.AIUsed || len(result.Dimensions) != 6 || len(result.SuggestedQuestions) != 5 {
		t.Fatalf("experience=%+v", result)
	}
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Content-Security-Policy") == "" {
		t.Fatalf("privacy/security headers=%v", response.Header())
	}
}

func TestPublicV06ExperienceContractIsStrictAndFailClosed(t *testing.T) {
	t.Parallel()
	handler := New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler()
	for _, body := range []string{
		`{"text":"Hallo","store_raw_text":true}`,
		`{"text":"Hallo","profile":"TEAM_RANKING"}`,
		`{"text":"Hallo","language_level":"GENERATIVE"}`,
		`{"text":"Hallo","context":"PERSONALITY_DIAGNOSIS"}`,
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v6/experience/analyze", strings.NewReader(body)))
		if response.Code < 400 {
			t.Fatalf("body=%s status=%d", body, response.Code)
		}
	}
}

func TestMVPProductAndReadOnlyAdminShellAreEmbedded(t *testing.T) {
	t.Parallel()
	handler := New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler()
	for _, test := range []struct {
		path, contentType, marker string
	}{
		{"/", "text/html", "Was zeigt sich in deinem Satz?"},
		{"/admin", "text/html", "Transparenz vor Veränderung."},
		{"/app.css", "text/css", ".workspace-card"},
		{"/app.js", "text/javascript", "/api/v6/experience/analyze"},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK || !strings.HasPrefix(response.Header().Get("Content-Type"), test.contentType) || !strings.Contains(response.Body.String(), test.marker) {
			t.Fatalf("GET %s: status=%d type=%s", test.path, response.Code, response.Header().Get("Content-Type"))
		}
	}
	admin := httptest.NewRecorder()
	handler.ServeHTTP(admin, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if strings.Contains(admin.Body.String(), "/commit") || strings.Contains(admin.Body.String(), "/rollback") {
		t.Fatal("read-only admin shell exposes write operations")
	}
	product := httptest.NewRecorder()
	handler.ServeHTTP(product, httptest.NewRequest(http.MethodGet, "/", nil))
	if strings.Contains(product.Body.String(), "https://") || strings.Contains(product.Body.String(), "http://") {
		t.Fatal("product shell loads an external resource")
	}
	javascript := httptest.NewRecorder()
	handler.ServeHTTP(javascript, httptest.NewRequest(http.MethodGet, "/app.js", nil))
	for _, forbidden := range []string{"localStorage", "sessionStorage", "innerHTML", "eval("} {
		if strings.Contains(javascript.Body.String(), forbidden) {
			t.Fatalf("product javascript contains forbidden browser capability %q", forbidden)
		}
	}
	if !strings.Contains(javascript.Header().Get("Content-Security-Policy"), "connect-src 'self'") {
		t.Fatal("embedded product assets do not retain the self-only connection policy")
	}
}
