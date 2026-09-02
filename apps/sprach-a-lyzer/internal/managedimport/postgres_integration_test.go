package managedimport

import (
	"context"
	"os"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db/migrations"
)

func TestPostgresManagedImportCommitRollbackAndImmutableAudit(t *testing.T) {
	databaseURL := os.Getenv("SAL_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SAL_TEST_DATABASE_URL is not configured")
	}
	ctx := context.Background()
	database, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := db.NewMigrator(database, migrations.Files, ".").Up(ctx); err != nil {
		t.Fatal(err)
	}
	repository := NewPostgresRepository(database)
	service := New(repository)
	key := "src_" + newID()
	parentKey := "src_" + newID()
	batchKey := "PG-" + newID()
	plan, err := service.Prepare(ctx, PrepareRequest{BatchKey: batchKey, OperationType: "IMPORT", SourceType: "JSON", SourceName: "postgres.json", TargetEntity: "SOURCE", ConflictPolicy: "USE_IMPORT", ActorID: "reviewer", ActorRole: "REVIEWER", SourceContent: `[{"source_key":"` + parentKey + `","version":1,"status":"APPROVED","title":"Parent"},{"source_key":"` + key + `","version":1,"status":"APPROVED","title":"Postgres integration","references":["` + parentKey + `"]}]`})
	if err != nil || plan.Status != "READY" {
		t.Fatalf("prepare=%+v %v", plan, err)
	}
	t.Cleanup(func() {
		_, _ = database.Exec(`DELETE FROM import_change_log WHERE batch_id=$1`, plan.BatchID)
		_, _ = database.Exec(`DELETE FROM import_batches WHERE id=$1`, plan.BatchID)
		_, _ = database.Exec(`DELETE FROM managed_knowledge_records WHERE entity_type='SOURCE' AND natural_key=$1`, key)
		_, _ = database.Exec(`DELETE FROM managed_knowledge_records WHERE entity_type='SOURCE' AND natural_key=$1`, parentKey)
	})
	committed, err := service.Commit(ctx, CommitRequest{BatchID: plan.BatchID, ActorID: "publisher", ActorRole: "PUBLISHER"})
	if err != nil || committed.ChangedRecords != 2 {
		t.Fatalf("commit=%+v %v", committed, err)
	}
	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM managed_knowledge_records WHERE entity_type='SOURCE' AND natural_key=$1`, key).Scan(&count); err != nil || count != 1 {
		t.Fatalf("record count=%d %v", count, err)
	}
	existing, err := repository.Existing(ctx, "SOURCE")
	if err != nil || len(existing[key].References) != 1 || existing[key].References[0] != parentKey {
		t.Fatalf("persisted references=%+v %v", existing[key].References, err)
	}
	rolledBack, err := service.Rollback(ctx, CommitRequest{BatchID: plan.BatchID, ActorID: "admin", ActorRole: "ADMIN"})
	if err != nil || rolledBack.ChangedRecords != 2 {
		t.Fatalf("rollback=%+v %v", rolledBack, err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM managed_knowledge_records WHERE entity_type='SOURCE' AND natural_key=$1`, key).Scan(&count); err != nil || count != 0 {
		t.Fatalf("record count after rollback=%d %v", count, err)
	}
	events, err := repository.Audit(ctx, plan.BatchID)
	if err != nil || len(events) != 3 {
		t.Fatalf("audit=%+v %v", events, err)
	}
	if _, err := database.Exec(`UPDATE audit_events SET event_type='TAMPERED' WHERE entity_type='IMPORT_BATCH' AND entity_id=$1`, plan.BatchID); err == nil {
		t.Fatal("immutable audit accepted update")
	}
}
