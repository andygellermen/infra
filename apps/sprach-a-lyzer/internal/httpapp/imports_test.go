package httpapp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/managedimport"
)

func TestPublicV05ManagedImportLifecycle(t *testing.T) {
	t.Parallel()
	repository := managedimport.NewMemoryRepository()
	handler := NewWithImports(analysis.NewDefault(), pingerStub{}, 1<<20, managedimport.New(repository)).Handler()
	prepare := httptest.NewRecorder()
	payload := `{
  "batch_key":"HTTP-IMPORT-001","operation_type":"IMPORT","source_type":"JSON","source_name":"sources.json",
  "source_content":"[{\"source_key\":\"src_http\",\"version\":1,\"status\":\"APPROVED\",\"title\":\"HTTP Source\"}]",
  "target_entity":"SOURCE","conflict_policy":"USE_IMPORT","actor_id":"editor","actor_role":"REVIEWER"
}`
	handler.ServeHTTP(prepare, httptest.NewRequest(http.MethodPost, "/api/v5/admin/imports/prepare", strings.NewReader(payload)))
	if prepare.Code != http.StatusOK {
		t.Fatalf("prepare status=%d body=%s", prepare.Code, prepare.Body.String())
	}
	var plan managedimport.Plan
	if err := json.NewDecoder(prepare.Body).Decode(&plan); err != nil {
		t.Fatal(err)
	}
	assertVersionHeaders(t, prepare, "managed-import-plan", managedimport.ContractVersion)
	if plan.Status != "READY" || plan.Summary.New != 1 {
		t.Fatalf("plan=%+v", plan)
	}

	commit := httptest.NewRecorder()
	handler.ServeHTTP(commit, httptest.NewRequest(http.MethodPost, "/api/v5/admin/imports/commit", strings.NewReader(fmt.Sprintf(`{"batch_id":%q,"actor_id":"publisher","actor_role":"PUBLISHER"}`, plan.BatchID))))
	if commit.Code != http.StatusOK {
		t.Fatalf("commit status=%d body=%s", commit.Code, commit.Body.String())
	}
	assertVersionHeaders(t, commit, "managed-import-operation", managedimport.ContractVersion)

	history := httptest.NewRecorder()
	handler.ServeHTTP(history, httptest.NewRequest(http.MethodGet, "/api/v5/admin/imports", nil))
	if history.Code != http.StatusOK || !strings.Contains(history.Body.String(), `"COMPLETED"`) {
		t.Fatalf("history status=%d body=%s", history.Code, history.Body.String())
	}

	audit := httptest.NewRecorder()
	handler.ServeHTTP(audit, httptest.NewRequest(http.MethodGet, "/api/v5/admin/imports/"+plan.BatchID+"/audit", nil))
	if audit.Code != http.StatusOK || !strings.Contains(audit.Body.String(), "IMPORT_COMMITTED") {
		t.Fatalf("audit status=%d body=%s", audit.Code, audit.Body.String())
	}

	rollback := httptest.NewRecorder()
	handler.ServeHTTP(rollback, httptest.NewRequest(http.MethodPost, "/api/v5/admin/imports/rollback", strings.NewReader(fmt.Sprintf(`{"batch_id":%q,"actor_id":"admin","actor_role":"ADMIN"}`, plan.BatchID))))
	if rollback.Code != http.StatusOK || !strings.Contains(rollback.Body.String(), "ROLLED_BACK") {
		t.Fatalf("rollback status=%d body=%s", rollback.Code, rollback.Body.String())
	}
}

func TestPublicV05ManagedImportIsStrictAndRoleGuarded(t *testing.T) {
	t.Parallel()
	handler := New(analysis.NewDefault(), pingerStub{}, 64<<10).Handler()
	tests := []struct {
		route, payload string
		want           int
	}{
		{"/api/v5/admin/imports/prepare", `{"batch_key":"B","operation_type":"IMPORT","source_type":"JSON","source_name":"x","source_content":"[]","target_entity":"SOURCE","actor_role":"CONTRIBUTOR","unknown":true}`, http.StatusBadRequest},
		{"/api/v5/admin/imports/prepare", `{"batch_key":"B","operation_type":"SYNC_LATER","source_type":"JSON","source_name":"x","source_content":"[]","target_entity":"SOURCE","actor_role":"ADMIN"}`, http.StatusUnprocessableEntity},
		{"/api/v5/admin/imports/commit", `{"batch_id":"missing","actor_role":"CONTRIBUTOR"}`, http.StatusForbidden},
	}
	for _, test := range tests {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, test.route, strings.NewReader(test.payload)))
		if response.Code != test.want {
			t.Errorf("%s status=%d want=%d body=%s", test.route, response.Code, test.want, response.Body.String())
		}
	}
}
