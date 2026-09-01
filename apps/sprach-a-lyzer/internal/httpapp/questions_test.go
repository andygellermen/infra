package httpapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/questions"
)

func TestPublicV03QuestionAnswerContracts(t *testing.T) {
	t.Parallel()
	handler := New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler()

	selectionResponse := httptest.NewRecorder()
	handler.ServeHTTP(selectionResponse, httptest.NewRequest(http.MethodPost, "/api/v3/questions/select",
		strings.NewReader(`{"profile":"CORPORATE","construct_gaps":["OPTIONS"],"limit":5}`)))
	if selectionResponse.Code != http.StatusOK {
		t.Fatalf("selection status = %d, body = %s", selectionResponse.Code, selectionResponse.Body.String())
	}
	var selection questions.Selection
	if err := json.NewDecoder(selectionResponse.Body).Decode(&selection); err != nil {
		t.Fatal(err)
	}
	assertVersionHeaders(t, selectionResponse, "question-selection", questions.SelectionContractVersion)
	if len(selection.Candidates) != 5 || selection.Candidates[0].QuestionID != "CQ034" {
		t.Fatalf("selection = %+v", selection)
	}

	observationResponse := httptest.NewRecorder()
	handler.ServeHTTP(observationResponse, httptest.NewRequest(http.MethodPost, "/api/v3/answers/analyze",
		strings.NewReader(`{"question_id":"CQ007","answer":"Ich kann die Entscheidung nicht beeinflussen, aber ich kann morgen nachfragen."}`)))
	if observationResponse.Code != http.StatusOK {
		t.Fatalf("observation status = %d, body = %s", observationResponse.Code, observationResponse.Body.String())
	}
	var observation questions.Observation
	if err := json.NewDecoder(observationResponse.Body).Decode(&observation); err != nil {
		t.Fatal(err)
	}
	assertVersionHeaders(t, observationResponse, "question-answer-observation", questions.ObservationContractVersion)
	if observation.InferenceLevel != "C1" || observation.QuestionScoreBias != 0 || len(observation.QAPatterns) != 1 {
		t.Fatalf("observation = %+v", observation)
	}

	sessionResponse := httptest.NewRecorder()
	handler.ServeHTTP(sessionResponse, httptest.NewRequest(http.MethodPost, "/api/v3/sessions/compose", strings.NewReader(`{
  "session_id":"S-HTTP-001",
  "pairs":[
    {"question_id":"CQ007","baseline_answer":"Mein Chef ist im Urlaub.","answer":"Eigentlich gar nichts. Ich kann nur abwarten."},
    {"question_id":"CQ013","answer":"Ich kann Hilfe holen und trotzdem verantwortlich bleiben."}
  ]
}`)))
	if sessionResponse.Code != http.StatusOK {
		t.Fatalf("session status = %d, body = %s", sessionResponse.Code, sessionResponse.Body.String())
	}
	var session questions.Session
	if err := json.NewDecoder(sessionResponse.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	assertVersionHeaders(t, sessionResponse, "question-session", questions.SessionContractVersion)
	if session.InferenceLevel != "C3" || len(session.Trajectories) == 0 {
		t.Fatalf("session = %+v", session)
	}
}

func TestPublicV03QuestionRoutesAreStrictAndFailClosed(t *testing.T) {
	t.Parallel()
	handler := New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler()
	tests := []struct {
		route   string
		payload string
		status  int
	}{
		{"/api/v3/questions/select", `{"limit":5,"person_id":"42"}`, http.StatusBadRequest},
		{"/api/v3/questions/select", `{"limit":4}`, http.StatusUnprocessableEntity},
		{"/api/v3/answers/analyze", `{"question_id":"CQ999","answer":"Ja."}`, http.StatusUnprocessableEntity},
		{"/api/v3/answers/analyze", `{"question_id":"CQ007","answer":""}`, http.StatusUnprocessableEntity},
		{"/api/v3/sessions/compose", `{"session_id":"","pairs":[]}`, http.StatusUnprocessableEntity},
	}
	for _, test := range tests {
		request := httptest.NewRequest(http.MethodPost, test.route, strings.NewReader(test.payload))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.status {
			t.Errorf("%s status = %d, want %d; body = %s", test.route, response.Code, test.status, response.Body.String())
		}
	}
}
