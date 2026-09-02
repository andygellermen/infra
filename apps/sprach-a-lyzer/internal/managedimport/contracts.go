// Package managedimport owns safe, reviewable knowledge import operations.
package managedimport

import "time"

const ContractVersion = "0.1"

type PrepareRequest struct {
	BatchKey             string            `json:"batch_key"`
	OperationType        string            `json:"operation_type"`
	SourceType           string            `json:"source_type"`
	SourceName           string            `json:"source_name"`
	SourceContent        string            `json:"source_content,omitempty"`
	SourceBase64         string            `json:"source_base64,omitempty"`
	SourceCollection     string            `json:"source_collection,omitempty"`
	SourceSheet          string            `json:"source_sheet,omitempty"`
	TargetEntity         string            `json:"target_entity"`
	ColumnMapping        map[string]string `json:"column_mapping,omitempty"`
	SecondaryMatchFields []string          `json:"secondary_match_fields,omitempty"`
	ConflictPolicy       string            `json:"conflict_policy,omitempty"`
	Resolutions          map[string]string `json:"resolutions,omitempty"`
	AllowInsert          bool              `json:"allow_insert,omitempty"`
	AllowUpdate          bool              `json:"allow_update,omitempty"`
	ActorID              string            `json:"actor_id"`
	ActorRole            string            `json:"actor_role"`
}

type Record struct {
	NaturalKey string         `json:"natural_key"`
	Version    int            `json:"version"`
	Status     string         `json:"status"`
	Payload    map[string]any `json:"payload"`
	References []string       `json:"references"`
}

type FieldDiff struct {
	Field    string `json:"field"`
	Database any    `json:"database,omitempty"`
	Import   any    `json:"import,omitempty"`
}

type Row struct {
	RowNumber       int            `json:"row_number"`
	NaturalKey      string         `json:"natural_key,omitempty"`
	Raw             map[string]any `json:"raw"`
	Normalized      *Record        `json:"normalized,omitempty"`
	MatchedKey      string         `json:"matched_natural_key,omitempty"`
	MatchConfidence string         `json:"match_confidence"`
	Status          string         `json:"status"`
	Diff            []FieldDiff    `json:"diff"`
	Errors          []string       `json:"errors"`
	Warnings        []string       `json:"warnings"`
	Resolution      string         `json:"resolution,omitempty"`
}

type Summary struct {
	Total            int `json:"total"`
	New              int `json:"new"`
	Changed          int `json:"changed"`
	Unchanged        int `json:"unchanged"`
	Conflicts        int `json:"conflicts"`
	Invalid          int `json:"invalid"`
	ReferenceMissing int `json:"reference_missing"`
}

type GoldenResult struct {
	Passed bool     `json:"passed"`
	Cases  int      `json:"cases"`
	Errors []string `json:"errors"`
}

type Plan struct {
	ContractVersion   string       `json:"contract_version"`
	BatchID           string       `json:"batch_id"`
	BatchKey          string       `json:"batch_key"`
	OperationType     string       `json:"operation_type"`
	Status            string       `json:"status"`
	SourceType        string       `json:"source_type"`
	SourceName        string       `json:"source_name"`
	SourceFingerprint string       `json:"source_fingerprint"`
	DuplicateSource   bool         `json:"duplicate_source"`
	TargetEntity      string       `json:"target_entity"`
	ActorID           string       `json:"actor_id"`
	Rows              []Row        `json:"rows"`
	Summary           Summary      `json:"summary"`
	Golden            GoldenResult `json:"golden_dry_run"`
	CreatedAt         time.Time    `json:"created_at"`
}

type CommitRequest struct {
	BatchID   string `json:"batch_id"`
	ActorID   string `json:"actor_id"`
	ActorRole string `json:"actor_role"`
}

type OperationResult struct {
	ContractVersion string    `json:"contract_version"`
	BatchID         string    `json:"batch_id"`
	Status          string    `json:"status"`
	ChangedRecords  int       `json:"changed_records"`
	OccurredAt      time.Time `json:"occurred_at"`
}

type AuditEvent struct {
	EventType string         `json:"event_type"`
	EntityID  string         `json:"entity_id"`
	ActorID   string         `json:"actor_id"`
	Detail    map[string]any `json:"detail"`
	CreatedAt time.Time      `json:"created_at"`
}
