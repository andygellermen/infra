package golden_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/httpapp"
)

const goldenSuitePath = "../../data/golden/sprach-a-lyzer_vertical-slice_v0.2.json"

type suite struct {
	Version               string                 `json:"version"`
	AnalysisContract      string                 `json:"analysis_contract"`
	TraceContract         string                 `json:"trace_contract"`
	CanonicalDimensionIDs []analysis.DimensionID `json:"canonical_dimension_ids"`
	Cases                 []goldenCase           `json:"cases"`
}

type goldenCase struct {
	ID             string           `json:"id"`
	Request        analysis.Request `json:"request"`
	ExpectedResult analysis.Result  `json:"expected_result"`
}

func TestVerticalSliceGolden(t *testing.T) {
	t.Parallel()

	testSuite := loadSuite(t)
	validateSuite(t, testSuite)
	analyzer := analysis.NewDefault()

	for _, testCase := range testSuite.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			coreResult, err := analyzer.Analyze(testCase.Request)
			if err != nil {
				t.Fatalf("core Analyze() error: %v", err)
			}
			assertJSONEqual(t, "core analysis result", testCase.ExpectedResult, coreResult)
			assertJSONEqual(t, "core standalone trace", testCase.ExpectedResult.Trace(), coreResult.Trace())

			httpResult := analyzeHTTP(t, analyzer, testCase.Request)
			assertJSONEqual(t, "HTTP analysis result", testCase.ExpectedResult, httpResult)
			assertJSONEqual(t, "HTTP-derived standalone trace", testCase.ExpectedResult.Trace(), httpResult.Trace())
		})
	}
}

func loadSuite(t *testing.T) suite {
	t.Helper()
	file, err := os.Open(goldenSuitePath)
	if err != nil {
		t.Fatalf("open golden suite: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var testSuite suite
	if err := decoder.Decode(&testSuite); err != nil {
		t.Fatalf("decode golden suite: %v", err)
	}
	if err := requireEOF(decoder); err != nil {
		t.Fatalf("decode golden suite: %v", err)
	}
	return testSuite
}

func validateSuite(t *testing.T, testSuite suite) {
	t.Helper()
	if testSuite.Version != "0.2" || testSuite.AnalysisContract != "0.1" || testSuite.TraceContract != "0.1" {
		t.Fatalf("unexpected suite/contracts: suite=%q analysis=%q trace=%q",
			testSuite.Version, testSuite.AnalysisContract, testSuite.TraceContract)
	}
	if !slices.Equal(testSuite.CanonicalDimensionIDs, analysis.Dimensions()) {
		t.Fatalf("golden dimension IDs %v differ from core IDs %v", testSuite.CanonicalDimensionIDs, analysis.Dimensions())
	}
	wantCases := []struct {
		id      string
		text    string
		context analysis.Context
	}{
		{"VS01_INTERNAL_PRESSURE", "Ich muss das heute unbedingt noch schaffen.", analysis.ContextSelfTalk},
		{"VS02_SAFETY_DIRECTIVE", "Du musst sofort das Gebäude verlassen!", analysis.ContextSafety},
		{"VS03_FREE_OF_CHARGE", "Der Eintritt ist frei.", "PUBLIC_INFORMATION"},
		{"VS04_REPORTED_CLAIM", "Er soll sehr erfolgreich sein.", "PRIVATE_CONVERSATION"},
		{"VS05_RESPECTFUL_BOUNDARY", "Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.", "PRIVATE_CONVERSATION"},
		{"VS06_HOMOPHONE_GUARD", "Hast du Geld?", "MODERATION"},
	}
	if len(testSuite.Cases) != len(wantCases) {
		t.Fatalf("golden suite contains %d cases; want %d", len(testSuite.Cases), len(wantCases))
	}
	for index, testCase := range testSuite.Cases {
		want := wantCases[index]
		if testCase.ID != want.id || testCase.Request.Text != want.text || testCase.Request.Context != want.context {
			t.Errorf("case %d = %q/%q/%q; want %q/%q/%q", index, testCase.ID, testCase.Request.Text,
				testCase.Request.Context, want.id, want.text, want.context)
		}
		if testCase.Request.Text == "" || len(testCase.ExpectedResult.Dimensions) != 6 {
			t.Errorf("%s is not a complete six-dimension case", testCase.ID)
		}
		if testCase.ExpectedResult.Text != testCase.Request.Text || testCase.ExpectedResult.Context != testCase.Request.Context {
			t.Errorf("%s expected result does not preserve request text/context", testCase.ID)
		}
		for _, id := range analysis.Dimensions() {
			result, exists := testCase.ExpectedResult.Dimensions[id]
			if !exists {
				t.Errorf("%s lacks dimension %s", testCase.ID, id)
				continue
			}
			if result.State == analysis.NotAssessable && result.Score != nil {
				t.Errorf("%s %s is NOT_ASSESSABLE with score %v", testCase.ID, id, *result.Score)
			}
			if result.State != analysis.NotAssessable && result.Score == nil {
				t.Errorf("%s %s is %s without score", testCase.ID, id, result.State)
			}
		}
		for contributionIndex, contribution := range testCase.ExpectedResult.ContributionTrace {
			if !slices.Contains(analysis.Dimensions(), contribution.Dimension) {
				t.Errorf("%s contribution %d uses unknown dimension %q", testCase.ID, contributionIndex, contribution.Dimension)
			}
			if contribution.RuleID == "" || contribution.Evidence == "" || contribution.Reason == "" {
				t.Errorf("%s contribution %d is incomplete: %+v", testCase.ID, contributionIndex, contribution)
			}
		}
		encoded, err := json.Marshal(testCase.ExpectedResult)
		if err != nil {
			t.Fatalf("encode %s expectation: %v", testCase.ID, err)
		}
		if bytes.Contains(encoded, []byte("FREE_WILL")) {
			t.Errorf("%s contains legacy dimension ID", testCase.ID)
		}
	}
	if len(testSuite.Cases[0].ExpectedResult.ContributionTrace) != 5 {
		t.Fatalf("first vertical-slice case has %d Contributions; want 5", len(testSuite.Cases[0].ExpectedResult.ContributionTrace))
	}
}

func analyzeHTTP(t *testing.T, analyzer *analysis.Service, input analysis.Request) analysis.Result {
	t.Helper()
	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("encode HTTP request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/analyze", bytes.NewReader(payload))
	response := httptest.NewRecorder()
	httpapp.New(analyzer, readyPinger{}, 64<<10).Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d; body = %s", response.Code, response.Body.String())
	}
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	var result analysis.Result
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("decode HTTP result: %v", err)
	}
	if err := requireEOF(decoder); err != nil {
		t.Fatalf("decode HTTP result: %v", err)
	}
	return result
}

func assertJSONEqual(t *testing.T, label string, want, got any) {
	t.Helper()
	wantJSON, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("encode expected %s: %v", label, err)
	}
	gotJSON, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("encode actual %s: %v", label, err)
	}
	if !bytes.Equal(gotJSON, wantJSON) {
		t.Fatalf("%s differs from golden fixture\nwant:\n%s\n\ngot:\n%s", label, wantJSON, gotJSON)
	}
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("more than one JSON value")
	}
	return err
}

type readyPinger struct{}

func (readyPinger) PingContext(context.Context) error { return nil }
