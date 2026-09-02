package managedimport

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestManagedImportJSONCommitRollbackAndAudit(t *testing.T) {
	t.Parallel()
	repository := NewMemoryRepository()
	repository.Records["LEXEME"] = map[string]Record{
		"lx_alt":    {NaturalKey: "lx_alt", Version: 1, Status: "APPROVED", Payload: map[string]any{"lemma": "alt", "description": "Vorher"}, References: []string{"lx_parent"}},
		"lx_parent": {NaturalKey: "lx_parent", Version: 1, Status: "APPROVED", Payload: map[string]any{"lemma": "parent"}},
		"lx_other":  {NaturalKey: "lx_other", Version: 1, Status: "APPROVED", Payload: map[string]any{"lemma": "other"}},
	}
	service := New(repository)
	plan, err := service.Prepare(context.Background(), PrepareRequest{
		BatchKey: "B-JSON-001", OperationType: "UPDATE", SourceType: "JSON", SourceName: "lexemes.json",
		TargetEntity: "LEXEME", ActorID: "editor", ActorRole: "REVIEWER", AllowInsert: true, AllowUpdate: true,
		ConflictPolicy: "USE_IMPORT", SourceContent: `{"records":[
	  {"lexeme_key":"lx_alt","version":2,"status":"APPROVED","lemma":"alt","description":"Nachher","references":["lx_parent","lx_other"]},
  {"lexeme_key":"lx_neu","version":1,"status":"REVIEW","lemma":"neu","references":["lx_alt"]}
]}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Status != "READY" || plan.Summary.Changed != 1 || plan.Summary.New != 1 || !plan.Golden.Passed {
		t.Fatalf("plan = %+v", plan)
	}
	if !containsField(plan.Rows[0].Diff, "references") {
		t.Fatalf("references missing from diff: %+v", plan.Rows[0].Diff)
	}
	if _, err := service.Commit(context.Background(), CommitRequest{BatchID: plan.BatchID, ActorID: "publisher", ActorRole: "CONTRIBUTOR"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("unauthorized commit error = %v", err)
	}
	committed, err := service.Commit(context.Background(), CommitRequest{BatchID: plan.BatchID, ActorID: "publisher", ActorRole: "PUBLISHER"})
	if err != nil || committed.Status != "COMPLETED" || committed.ChangedRecords != 2 {
		t.Fatalf("commit = %+v, %v", committed, err)
	}
	if repository.Records["LEXEME"]["lx_alt"].Version != 2 || repository.Records["LEXEME"]["lx_neu"].NaturalKey == "" {
		t.Fatalf("records after commit = %+v", repository.Records)
	}
	rolledBack, err := service.Rollback(context.Background(), CommitRequest{BatchID: plan.BatchID, ActorID: "admin", ActorRole: "ADMIN"})
	if err != nil || rolledBack.Status != "ROLLED_BACK" || rolledBack.ChangedRecords != 2 {
		t.Fatalf("rollback = %+v, %v", rolledBack, err)
	}
	if repository.Records["LEXEME"]["lx_alt"].Version != 1 || repository.Records["LEXEME"]["lx_neu"].NaturalKey != "" {
		t.Fatalf("records after rollback = %+v", repository.Records)
	}
	events, _ := repository.Audit(context.Background(), plan.BatchID)
	if got := []string{events[0].EventType, events[1].EventType, events[2].EventType}; !reflect.DeepEqual(got, []string{"IMPORT_PREPARED", "IMPORT_COMMITTED", "IMPORT_ROLLED_BACK"}) {
		t.Fatalf("audit events = %v", got)
	}
}

func containsField(diffs []FieldDiff, field string) bool {
	for _, diff := range diffs {
		if diff.Field == field {
			return true
		}
	}
	return false
}

func TestValidateOnlyCSVMappingCompatibilityAndDuplicateDetection(t *testing.T) {
	t.Parallel()
	repository := NewMemoryRepository()
	service := New(repository)
	request := PrepareRequest{
		BatchKey: "B-CSV-001", OperationType: "VALIDATE_ONLY", SourceType: "CSV", SourceName: "contributions.csv",
		TargetEntity: "DIMENSION_CONTRIBUTION", ActorID: "editor", ActorRole: "CONTRIBUTOR",
		SourceContent: "contribution_key,version,status,dimension,value\ndc_one,1,APPROVED,FREE_WILL,-10\n",
	}
	plan, err := service.Prepare(context.Background(), request)
	if err != nil || plan.Status != "VALIDATED" || plan.Rows[0].Normalized.Payload["dimension"] != "VOLITION" {
		t.Fatalf("CSV validation plan = %+v, %v", plan, err)
	}
	if _, err := service.Commit(context.Background(), CommitRequest{BatchID: plan.BatchID, ActorRole: "PUBLISHER"}); !errors.Is(err, ErrValidateOnly) {
		t.Fatalf("validate-only commit error = %v", err)
	}
	request.BatchKey = "B-CSV-002"
	duplicate, err := service.Prepare(context.Background(), request)
	if err != nil || !duplicate.DuplicateSource {
		t.Fatalf("duplicate plan = %+v, %v", duplicate, err)
	}
}

func TestConflictsReferencesCyclesAndGoldenRegressionBlockCommit(t *testing.T) {
	t.Parallel()
	repository := NewMemoryRepository()
	repository.Records["RELATION"] = map[string]Record{"rel_one": {NaturalKey: "rel_one", Version: 1, Status: "APPROVED", Payload: map[string]any{"relation_type": "RELATED"}}}
	service := New(repository)
	critical, err := service.Prepare(context.Background(), PrepareRequest{
		BatchKey: "B-CONFLICT", OperationType: "UPDATE", SourceType: "JSON", SourceName: "relations.json", TargetEntity: "RELATION",
		ActorRole: "CONTRIBUTOR", AllowUpdate: true, ConflictPolicy: "USE_IMPORT",
		SourceContent: `[{"relation_key":"rel_one","version":2,"status":"APPROVED","relation_type":"CAUSE"}]`,
	})
	if err != nil || critical.Status != "FAILED" || critical.Rows[0].Status != "CONFLICT" || !contains(critical.Rows[0].Errors, "HARD_GUARDRAIL_VIOLATION") {
		t.Fatalf("critical plan = %+v, %v", critical, err)
	}
	blocked, err := service.Prepare(context.Background(), PrepareRequest{
		BatchKey: "B-BLOCKED", OperationType: "IMPORT", SourceType: "JSON", SourceName: "blocked.json", TargetEntity: "SOURCE",
		ActorRole: "REVIEWER", ConflictPolicy: "USE_IMPORT",
		SourceContent: `[
 {"source_key":"a","version":1,"status":"REVIEW","references":["b"]},
 {"source_key":"b","version":1,"status":"REVIEW","references":["a"]},
 {"source_key":"c","version":1,"status":"REVIEW","references":["missing"]},
 {"source_key":"d","version":1,"status":"REVIEW","golden_regression":true}
]`,
	})
	if err != nil || blocked.Status != "FAILED" || blocked.Golden.Passed || blocked.Summary.Invalid == 0 || blocked.Summary.ReferenceMissing != 1 {
		t.Fatalf("blocked plan = %+v, %v", blocked, err)
	}
	if _, err := service.Commit(context.Background(), CommitRequest{BatchID: blocked.BatchID, ActorRole: "PUBLISHER"}); !errors.Is(err, ErrNotReady) {
		t.Fatalf("blocked commit error = %v", err)
	}
}

func TestXLSXFirstSheetParsing(t *testing.T) {
	t.Parallel()
	data := minimalXLSX(t)
	rows, err := parseSource(PrepareRequest{SourceType: "XLSX"}, data)
	if err != nil || len(rows) != 1 || rows[0]["lexeme_key"] != "lx_xlsx" {
		t.Fatalf("XLSX rows = %+v, %v", rows, err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if decoded, err := sourceBytes(PrepareRequest{SourceBase64: encoded}); err != nil || !bytes.Equal(decoded, data) {
		t.Fatal("XLSX base64 round trip failed")
	}
}

func minimalXLSX(t *testing.T) []byte {
	t.Helper()
	buffer := &bytes.Buffer{}
	archive := zip.NewWriter(buffer)
	write := func(name, value string) {
		file, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write([]byte(value)); err != nil {
			t.Fatal(err)
		}
	}
	write("xl/sharedStrings.xml", `<?xml version="1.0"?><sst><si><t>lexeme_key</t></si><si><t>version</t></si><si><t>status</t></si><si><t>lx_xlsx</t></si><si><t>APPROVED</t></si></sst>`)
	write("xl/worksheets/sheet1.xml", `<?xml version="1.0"?><worksheet><sheetData><row><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c><c r="C1" t="s"><v>2</v></c></row><row><c r="A2" t="s"><v>3</v></c><c r="B2"><v>1</v></c><c r="C2" t="s"><v>4</v></c></row></sheetData></worksheet>`)
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestManualCriticalResolutionRequiresReviewer(t *testing.T) {
	t.Parallel()
	repository := NewMemoryRepository()
	repository.Records["RELATION"] = map[string]Record{"rel_one": {NaturalKey: "rel_one", Version: 1, Status: "APPROVED", Payload: map[string]any{"relation_type": "RELATED"}}}
	plan, err := New(repository).Prepare(context.Background(), PrepareRequest{
		BatchKey: "B-REVIEW", OperationType: "UPDATE", SourceType: "JSON", SourceName: "relations.json", TargetEntity: "RELATION",
		ActorRole: "REVIEWER", AllowUpdate: true, ConflictPolicy: "REQUIRE_MANUAL", Resolutions: map[string]string{"rel_one": "USE_IMPORT"},
		SourceContent: `[{"relation_key":"rel_one","version":2,"status":"APPROVED","relation_type":"CAUSE"}]`,
	})
	if err != nil || plan.Status != "READY" || plan.Rows[0].Status != "CHANGED" || plan.Rows[0].Resolution != "USE_IMPORT" {
		t.Fatalf("reviewed plan = %+v, %v", plan, err)
	}
}

func TestUnsupportedSyncAndInvalidSourceFailClosed(t *testing.T) {
	t.Parallel()
	service := New(NewMemoryRepository())
	_, err := service.Prepare(context.Background(), PrepareRequest{BatchKey: "B", OperationType: "SYNC_LATER", SourceType: "JSON", SourceName: "x", TargetEntity: "SOURCE", ActorRole: "ADMIN", SourceContent: "[]"})
	if err == nil || !strings.Contains(err.Error(), "SYNC_NOT_ENABLED") {
		t.Fatalf("sync error = %v", err)
	}
}
