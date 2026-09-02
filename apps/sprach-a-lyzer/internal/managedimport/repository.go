package managedimport

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

type change struct {
	Entity, Key, Operation string
	Before                 *Record
	After                  Record
	RolledBack             bool
}

type MemoryRepository struct {
	mu      sync.Mutex
	Records map[string]map[string]Record
	Plans   map[string]Plan
	Changes map[string][]change
	Events  []AuditEvent
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{Records: map[string]map[string]Record{}, Plans: map[string]Plan{}, Changes: map[string][]change{}, Events: []AuditEvent{}}
}
func (r *MemoryRepository) Existing(_ context.Context, entity string) (map[string]Record, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := map[string]Record{}
	for key, value := range r.Records[entity] {
		result[key] = cloneRecord(value)
	}
	return result, nil
}
func (r *MemoryRepository) KnownKeys(context.Context) (map[string]bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := map[string]bool{}
	for _, values := range r.Records {
		for key := range values {
			result[key] = true
		}
	}
	return result, nil
}
func (r *MemoryRepository) DuplicateFingerprint(_ context.Context, fingerprint string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, plan := range r.Plans {
		if plan.SourceFingerprint == fingerprint && (plan.Status == "COMPLETED" || plan.Status == "VALIDATED") {
			return true, nil
		}
	}
	return false, nil
}
func (r *MemoryRepository) SavePlan(_ context.Context, plan Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.Plans {
		if existing.BatchKey == plan.BatchKey {
			return fmt.Errorf("duplicate batch key %s", plan.BatchKey)
		}
	}
	r.Plans[plan.BatchID] = plan
	r.Events = append(r.Events, audit("IMPORT_PREPARED", plan.BatchID, plan.ActorID, map[string]any{"status": plan.Status, "fingerprint": plan.SourceFingerprint}))
	return nil
}
func (r *MemoryRepository) LoadPlan(_ context.Context, id string) (Plan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	plan, ok := r.Plans[id]
	if !ok {
		return Plan{}, sql.ErrNoRows
	}
	return plan, nil
}
func (r *MemoryRepository) Commit(_ context.Context, plan Plan, actor string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Records[plan.TargetEntity] == nil {
		r.Records[plan.TargetEntity] = map[string]Record{}
	}
	count := 0
	changes := []change{}
	for _, row := range plan.Rows {
		if row.Normalized == nil || row.Status != "NEW" && row.Status != "CHANGED" {
			continue
		}
		record := cloneRecord(*row.Normalized)
		current, exists := r.Records[plan.TargetEntity][row.MatchedKey]
		operation := "INSERT"
		var before *Record
		if exists {
			operation = "UPDATE"
			copied := cloneRecord(current)
			before = &copied
			delete(r.Records[plan.TargetEntity], row.MatchedKey)
		}
		r.Records[plan.TargetEntity][record.NaturalKey] = record
		changes = append(changes, change{plan.TargetEntity, record.NaturalKey, operation, before, record, false})
		count++
	}
	plan.Status = "COMPLETED"
	r.Plans[plan.BatchID] = plan
	r.Changes[plan.BatchID] = changes
	r.Events = append(r.Events, audit("IMPORT_COMMITTED", plan.BatchID, actor, map[string]any{"changed_records": count}))
	return count, nil
}
func (r *MemoryRepository) Rollback(_ context.Context, plan Plan, actor string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	changes := r.Changes[plan.BatchID]
	count := 0
	for index := len(changes) - 1; index >= 0; index-- {
		item := &changes[index]
		if item.RolledBack {
			continue
		}
		if item.Operation == "INSERT" {
			delete(r.Records[item.Entity], item.Key)
		} else if item.Before != nil {
			r.Records[item.Entity][item.Before.NaturalKey] = cloneRecord(*item.Before)
		}
		item.RolledBack = true
		count++
	}
	r.Changes[plan.BatchID] = changes
	plan.Status = "ROLLED_BACK"
	r.Plans[plan.BatchID] = plan
	r.Events = append(r.Events, audit("IMPORT_ROLLED_BACK", plan.BatchID, actor, map[string]any{"changed_records": count}))
	return count, nil
}
func (r *MemoryRepository) History(context.Context) ([]Plan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]Plan, 0, len(r.Plans))
	for _, plan := range r.Plans {
		result = append(result, plan)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	return result, nil
}
func (r *MemoryRepository) Audit(_ context.Context, id string) ([]AuditEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := []AuditEvent{}
	for _, event := range r.Events {
		if event.EntityID == id {
			result = append(result, event)
		}
	}
	return result, nil
}

type PostgresRepository struct{ database *sql.DB }

func NewPostgresRepository(database *sql.DB) *PostgresRepository {
	return &PostgresRepository{database: database}
}
func (r *PostgresRepository) Existing(ctx context.Context, entity string) (map[string]Record, error) {
	if r.database == nil {
		return nil, fmt.Errorf("managed import database is nil")
	}
	rows, err := r.database.QueryContext(ctx, `SELECT natural_key,payload,dependency_refs,version,status FROM managed_knowledge_records WHERE entity_type=$1`, entity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]Record{}
	for rows.Next() {
		var record Record
		var payload, references []byte
		if err := rows.Scan(&record.NaturalKey, &payload, &references, &record.Version, &record.Status); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payload, &record.Payload); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(references, &record.References); err != nil {
			return nil, err
		}
		result[record.NaturalKey] = record
	}
	return result, rows.Err()
}
func (r *PostgresRepository) KnownKeys(ctx context.Context) (map[string]bool, error) {
	rows, err := r.database.QueryContext(ctx, `SELECT natural_key FROM managed_knowledge_records`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]bool{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		result[key] = true
	}
	return result, rows.Err()
}
func (r *PostgresRepository) DuplicateFingerprint(ctx context.Context, fingerprint string) (bool, error) {
	var exists bool
	err := r.database.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM import_batches WHERE source_sha256=$1 AND status IN ('COMPLETED','VALIDATED'))`, fingerprint).Scan(&exists)
	return exists, err
}
func (r *PostgresRepository) SavePlan(ctx context.Context, plan Plan) error {
	payload, _ := json.Marshal(plan)
	errorsPayload, _ := json.Marshal(plan.Golden.Errors)
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO import_batches(id,batch_key,operation_type,status,source_type,source_name,source_sha256,target_entity,actor_id,total_rows,new_rows,changed_rows,unchanged_rows,conflict_rows,invalid_rows,diff_payload,error_summary,validated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,CASE WHEN $4 IN ('READY','VALIDATED') THEN now() END)`, plan.BatchID, plan.BatchKey, plan.OperationType, plan.Status, plan.SourceType, plan.SourceName, plan.SourceFingerprint, plan.TargetEntity, plan.ActorID, plan.Summary.Total, plan.Summary.New, plan.Summary.Changed, plan.Summary.Unchanged, plan.Summary.Conflicts, plan.Summary.Invalid+plan.Summary.ReferenceMissing, payload, errorsPayload)
	if err != nil {
		return err
	}
	for _, row := range plan.Rows {
		raw, _ := json.Marshal(row.Raw)
		normalized, _ := json.Marshal(row.Normalized)
		diff, _ := json.Marshal(row.Diff)
		errs, _ := json.Marshal(row.Errors)
		warnings, _ := json.Marshal(row.Warnings)
		_, err = tx.ExecContext(ctx, `INSERT INTO import_batch_rows(batch_id,row_number,natural_key,raw_payload,normalized_payload,matched_natural_key,match_confidence,row_status,diff,errors,warnings,resolution) VALUES($1,$2,NULLIF($3,''),$4,$5,NULLIF($6,''),$7,$8,$9,$10,$11,NULLIF($12,''))`, plan.BatchID, row.RowNumber, row.NaturalKey, raw, normalized, row.MatchedKey, row.MatchConfidence, row.Status, diff, errs, warnings, row.Resolution)
		if err != nil {
			return err
		}
	}
	detail, _ := json.Marshal(map[string]any{"status": plan.Status, "fingerprint": plan.SourceFingerprint})
	_, err = tx.ExecContext(ctx, `INSERT INTO audit_events(event_type,entity_type,entity_id,actor_id,detail) VALUES('IMPORT_PREPARED','IMPORT_BATCH',$1,$2,$3)`, plan.BatchID, plan.ActorID, detail)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (r *PostgresRepository) LoadPlan(ctx context.Context, id string) (Plan, error) {
	var payload []byte
	var status string
	err := r.database.QueryRowContext(ctx, `SELECT diff_payload,status FROM import_batches WHERE id=$1`, id).Scan(&payload, &status)
	if err != nil {
		return Plan{}, err
	}
	var plan Plan
	if err := json.Unmarshal(payload, &plan); err != nil {
		return Plan{}, err
	}
	plan.Status = status
	return plan, nil
}
func (r *PostgresRepository) Commit(ctx context.Context, plan Plan, actor string) (int, error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	count := 0
	for _, row := range plan.Rows {
		if row.Normalized == nil || row.Status != "NEW" && row.Status != "CHANGED" {
			continue
		}
		record := *row.Normalized
		var beforePayload, beforeReferences []byte
		var beforeVersion int
		var beforeStatus string
		scanErr := tx.QueryRowContext(ctx, `SELECT payload,dependency_refs,version,status FROM managed_knowledge_records WHERE entity_type=$1 AND natural_key=$2 FOR UPDATE`, plan.TargetEntity, row.MatchedKey).Scan(&beforePayload, &beforeReferences, &beforeVersion, &beforeStatus)
		operation := "INSERT"
		var before any
		if scanErr == nil {
			operation = "UPDATE"
			before = map[string]any{"natural_key": row.MatchedKey, "version": beforeVersion, "status": beforeStatus, "payload": json.RawMessage(beforePayload), "references": json.RawMessage(beforeReferences)}
		} else if scanErr != sql.ErrNoRows {
			return 0, scanErr
		}
		payload, _ := json.Marshal(record.Payload)
		references, _ := json.Marshal(record.References)
		_, err = tx.ExecContext(ctx, `INSERT INTO managed_knowledge_records(entity_type,natural_key,payload,dependency_refs,version,status) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(entity_type,natural_key) DO UPDATE SET payload=EXCLUDED.payload,dependency_refs=EXCLUDED.dependency_refs,version=EXCLUDED.version,status=EXCLUDED.status,updated_at=now()`, plan.TargetEntity, record.NaturalKey, payload, references, record.Version, record.Status)
		if err != nil {
			return 0, err
		}
		beforeJSON, _ := json.Marshal(before)
		afterJSON, _ := json.Marshal(record)
		_, err = tx.ExecContext(ctx, `INSERT INTO import_change_log(batch_id,entity_type,natural_key,operation,before_payload,after_payload) VALUES($1,$2,$3,$4,$5,$6)`, plan.BatchID, plan.TargetEntity, record.NaturalKey, operation, nullableJSON(beforeJSON, before), afterJSON)
		if err != nil {
			return 0, err
		}
		count++
	}
	detail, _ := json.Marshal(map[string]any{"changed_records": count})
	stateResult, err := tx.ExecContext(ctx, `UPDATE import_batches SET status='COMPLETED',committed_at=now(),completed_at=now() WHERE id=$1 AND status='READY'`, plan.BatchID)
	if err != nil {
		return 0, err
	}
	if affected, _ := stateResult.RowsAffected(); affected != 1 {
		return 0, ErrNotReady
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO audit_events(event_type,entity_type,entity_id,actor_id,detail) VALUES('IMPORT_COMMITTED','IMPORT_BATCH',$1,$2,$3)`, plan.BatchID, actor, detail); err != nil {
		return 0, err
	}
	return count, tx.Commit()
}
func (r *PostgresRepository) Rollback(ctx context.Context, plan Plan, actor string) (int, error) {
	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id,entity_type,natural_key,operation,before_payload FROM import_change_log WHERE batch_id=$1 AND NOT rolled_back ORDER BY id DESC FOR UPDATE`, plan.BatchID)
	if err != nil {
		return 0, err
	}
	type item struct {
		id                     int64
		entity, key, operation string
		before                 []byte
	}
	items := []item{}
	for rows.Next() {
		var value item
		if err := rows.Scan(&value.id, &value.entity, &value.key, &value.operation, &value.before); err != nil {
			rows.Close()
			return 0, err
		}
		items = append(items, value)
	}
	rows.Close()
	for _, value := range items {
		if value.operation == "INSERT" {
			_, err = tx.ExecContext(ctx, `DELETE FROM managed_knowledge_records WHERE entity_type=$1 AND natural_key=$2`, value.entity, value.key)
		} else {
			var before Record
			if err = json.Unmarshal(value.before, &before); err == nil {
				payload, _ := json.Marshal(before.Payload)
				references, _ := json.Marshal(before.References)
				_, err = tx.ExecContext(ctx, `UPDATE managed_knowledge_records SET payload=$3,dependency_refs=$4,version=$5,status=$6,updated_at=now() WHERE entity_type=$1 AND natural_key=$2`, value.entity, value.key, payload, references, before.Version, before.Status)
			}
		}
		if err != nil {
			return 0, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE import_change_log SET rolled_back=true WHERE id=$1`, value.id); err != nil {
			return 0, err
		}
	}
	detail, _ := json.Marshal(map[string]any{"changed_records": len(items)})
	stateResult, err := tx.ExecContext(ctx, `UPDATE import_batches SET status='ROLLED_BACK',rolled_back_at=now() WHERE id=$1 AND status='COMPLETED'`, plan.BatchID)
	if err != nil {
		return 0, err
	}
	if affected, _ := stateResult.RowsAffected(); affected != 1 {
		return 0, ErrNotReady
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO audit_events(event_type,entity_type,entity_id,actor_id,detail) VALUES('IMPORT_ROLLED_BACK','IMPORT_BATCH',$1,$2,$3)`, plan.BatchID, actor, detail)
	if err != nil {
		return 0, err
	}
	return len(items), tx.Commit()
}
func (r *PostgresRepository) History(ctx context.Context) ([]Plan, error) {
	rows, err := r.database.QueryContext(ctx, `SELECT diff_payload,status FROM import_batches ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Plan{}
	for rows.Next() {
		var payload []byte
		var status string
		if err := rows.Scan(&payload, &status); err != nil {
			return nil, err
		}
		var plan Plan
		if err := json.Unmarshal(payload, &plan); err != nil {
			return nil, err
		}
		plan.Status = status
		result = append(result, plan)
	}
	return result, rows.Err()
}
func (r *PostgresRepository) Audit(ctx context.Context, id string) ([]AuditEvent, error) {
	rows, err := r.database.QueryContext(ctx, `SELECT event_type,entity_id,COALESCE(actor_id,''),detail,created_at FROM audit_events WHERE entity_type='IMPORT_BATCH' AND entity_id=$1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []AuditEvent{}
	for rows.Next() {
		var event AuditEvent
		var detail []byte
		if err := rows.Scan(&event.EventType, &event.EntityID, &event.ActorID, &detail, &event.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(detail, &event.Detail); err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func cloneRecord(value Record) Record {
	data, _ := json.Marshal(value)
	var result Record
	_ = json.Unmarshal(data, &result)
	return result
}
func audit(kind, id, actor string, detail map[string]any) AuditEvent {
	return AuditEvent{EventType: kind, EntityID: id, ActorID: actor, Detail: detail, CreatedAt: time.Now().UTC()}
}
func nullableJSON(value []byte, original any) any {
	if original == nil {
		return nil
	}
	return value
}
