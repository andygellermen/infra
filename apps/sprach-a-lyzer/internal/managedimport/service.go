package managedimport

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
)

var (
	ErrNotReady          = errors.New("import batch is not ready")
	ErrForbidden         = errors.New("actor role is not authorized")
	ErrValidateOnly      = errors.New("validate-only batch cannot be committed")
	ErrAlreadyRolledBack = errors.New("import batch is already rolled back")
)

type Repository interface {
	Existing(context.Context, string) (map[string]Record, error)
	KnownKeys(context.Context) (map[string]bool, error)
	DuplicateFingerprint(context.Context, string) (bool, error)
	SavePlan(context.Context, Plan) error
	LoadPlan(context.Context, string) (Plan, error)
	Commit(context.Context, Plan, string) (int, error)
	Rollback(context.Context, Plan, string) (int, error)
	History(context.Context) ([]Plan, error)
	Audit(context.Context, string) ([]AuditEvent, error)
}

type GoldenRunner interface {
	Run(context.Context, string, []Record) GoldenResult
}

type deterministicGolden struct{}

func (deterministicGolden) Run(_ context.Context, _ string, records []Record) GoldenResult {
	var suite struct {
		Cases []struct {
			ID       string           `json:"id"`
			Request  analysis.Request `json:"request"`
			Expected analysis.Result  `json:"expected_result"`
		} `json:"cases"`
	}
	result := GoldenResult{Passed: true, Errors: []string{}}
	if err := json.Unmarshal(assets.VerticalSliceGoldenV02, &suite); err != nil {
		return GoldenResult{Passed: false, Errors: []string{"GOLDEN_SUITE_INVALID"}}
	}
	engine := analysis.NewDefault()
	for _, testCase := range suite.Cases {
		actual, err := engine.Analyze(testCase.Request)
		actualJSON, _ := json.Marshal(actual)
		expectedJSON, _ := json.Marshal(testCase.Expected)
		if err != nil || !bytes.Equal(actualJSON, expectedJSON) {
			result.Passed = false
			result.Errors = append(result.Errors, "GOLDEN_REGRESSION:"+testCase.ID)
		}
	}
	result.Cases = len(suite.Cases)
	for _, record := range records {
		if regression, _ := record.Payload["golden_regression"].(bool); regression {
			result.Passed = false
			result.Errors = append(result.Errors, "CANDIDATE_GOLDEN_REGRESSION:"+record.NaturalKey)
		}
	}
	return result
}

type Service struct {
	repository Repository
	golden     GoldenRunner
	now        func() time.Time
}

func New(repository Repository) *Service { return NewWithGolden(repository, deterministicGolden{}) }
func NewWithGolden(repository Repository, golden GoldenRunner) *Service {
	return &Service{repository: repository, golden: golden, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Prepare(ctx context.Context, request PrepareRequest) (Plan, error) {
	if s == nil || s.repository == nil || s.golden == nil {
		return Plan{}, fmt.Errorf("managed import dependencies unavailable")
	}
	request.OperationType = strings.ToUpper(strings.TrimSpace(request.OperationType))
	request.SourceType = strings.ToUpper(strings.TrimSpace(request.SourceType))
	request.TargetEntity = strings.ToUpper(strings.TrimSpace(request.TargetEntity))
	request.ActorRole = strings.ToUpper(strings.TrimSpace(request.ActorRole))
	request.ConflictPolicy = strings.ToUpper(strings.TrimSpace(request.ConflictPolicy))
	if request.ConflictPolicy == "" {
		request.ConflictPolicy = "REQUIRE_MANUAL"
	}
	if request.OperationType == "SYNC_LATER" {
		return Plan{}, fmt.Errorf("SYNC_NOT_ENABLED")
	}
	if !contains([]string{"IMPORT", "UPDATE", "VALIDATE_ONLY"}, request.OperationType) || request.BatchKey == "" || request.SourceName == "" || request.TargetEntity == "" {
		return Plan{}, fmt.Errorf("INVALID_IMPORT_REQUEST")
	}
	if request.ActorRole != "CONTRIBUTOR" && request.ActorRole != "REVIEWER" && request.ActorRole != "PUBLISHER" && request.ActorRole != "ADMIN" {
		return Plan{}, ErrForbidden
	}
	if !contains([]string{"KEEP_DATABASE", "USE_IMPORT", "REQUIRE_MANUAL", "KEEP_NEWER_VERSION", "KEEP_HIGHER_EVIDENCE"}, request.ConflictPolicy) {
		return Plan{}, fmt.Errorf("INVALID_CONFLICT_POLICY")
	}
	data, err := sourceBytes(request)
	if err != nil {
		return Plan{}, err
	}
	if len(data) == 0 || len(data) > 10<<20 {
		return Plan{}, fmt.Errorf("INVALID_SOURCE_SIZE")
	}
	rawRows, err := parseSource(request, data)
	if err != nil {
		return Plan{}, err
	}
	existing, err := s.repository.Existing(ctx, request.TargetEntity)
	if err != nil {
		return Plan{}, err
	}
	known, err := s.repository.KnownKeys(ctx)
	if err != nil {
		return Plan{}, err
	}
	digest := sha256.Sum256(data)
	fingerprint := hex.EncodeToString(digest[:])
	duplicate, err := s.repository.DuplicateFingerprint(ctx, fingerprint)
	if err != nil {
		return Plan{}, err
	}
	rows := make([]Row, 0, len(rawRows))
	batchKeys := map[string]bool{}
	for index, raw := range rawRows {
		row := Row{RowNumber: index + 1, Raw: raw, MatchConfidence: "NONE", Status: "INVALID", Diff: []FieldDiff{}, Errors: []string{}, Warnings: []string{}}
		record, mappingWarnings := mapRecord(raw, request)
		row.Warnings = append(row.Warnings, mappingWarnings...)
		row.NaturalKey = record.NaturalKey
		row.Normalized = &record
		if record.NaturalKey == "" {
			row.Errors = append(row.Errors, "UNKNOWN_FIELD:natural_key")
		}
		if record.Version < 1 {
			row.Errors = append(row.Errors, "INVALID_VALUE:version")
		}
		if !contains([]string{"DRAFT", "REVIEW", "APPROVED", "PRODUCTION", "ARCHIVED"}, record.Status) {
			row.Errors = append(row.Errors, "INVALID_VALUE:status")
		}
		if batchKeys[record.NaturalKey] {
			row.Errors = append(row.Errors, "DUPLICATE_NATURAL_KEY")
		}
		batchKeys[record.NaturalKey] = true
		if len(row.Errors) > 0 {
			rows = append(rows, row)
			continue
		}
		current, exact := existing[record.NaturalKey]
		if exact {
			row.MatchedKey, row.MatchConfidence = record.NaturalKey, "EXACT"
		} else if key, confidence := secondaryMatch(record, existing, request.SecondaryMatchFields); key != "" {
			row.MatchedKey, row.MatchConfidence = key, confidence
			current = existing[key]
		}
		if row.MatchConfidence == "AMBIGUOUS" || row.MatchConfidence == "PROBABLE" {
			row.Status = "CONFLICT"
			row.Errors = append(row.Errors, "AMBIGUOUS_MATCH")
			rows = append(rows, row)
			continue
		}
		if row.MatchConfidence == "NONE" {
			row.Status = "NEW"
			if request.OperationType == "UPDATE" && !request.AllowInsert {
				row.Status = "CONFLICT"
				row.Errors = append(row.Errors, "INSERT_NOT_ALLOWED")
			}
		} else {
			row.Diff = diffRecords(current, record)
			if len(row.Diff) == 0 {
				row.Status = "UNCHANGED"
			} else if request.OperationType == "IMPORT" {
				row.Status = "CONFLICT"
				row.Errors = append(row.Errors, "DUPLICATE_NATURAL_KEY")
			} else if !request.AllowUpdate {
				row.Status = "CONFLICT"
				row.Errors = append(row.Errors, "UPDATE_NOT_ALLOWED")
			} else {
				row.Status = "CHANGED"
			}
		}
		if row.Status == "CHANGED" {
			resolution := strings.ToUpper(request.Resolutions[record.NaturalKey])
			critical := hasCriticalDiff(row.Diff)
			switch {
			case critical && (resolution != "USE_IMPORT" || request.ActorRole == "CONTRIBUTOR"):
				row.Status = "CONFLICT"
				row.Errors = append(row.Errors, "HARD_GUARDRAIL_VIOLATION")
				row.Resolution = "REQUIRE_MANUAL"
			case resolution == "KEEP_DATABASE" || request.ConflictPolicy == "KEEP_DATABASE":
				row.Status = "SKIPPED"
				row.Resolution = "KEEP_DATABASE"
			case resolution == "USE_IMPORT" || request.ConflictPolicy == "USE_IMPORT":
				row.Resolution = resolution
				if row.Resolution == "" {
					row.Resolution = "USE_IMPORT"
				}
			case request.ConflictPolicy == "KEEP_NEWER_VERSION" && record.Version > current.Version:
				row.Resolution = "USE_IMPORT"
			case request.ConflictPolicy == "KEEP_HIGHER_EVIDENCE" && evidenceRank(record.Payload["evidence_class"]) < evidenceRank(current.Payload["evidence_class"]):
				row.Resolution = "USE_IMPORT"
			default:
				row.Status = "CONFLICT"
				row.Errors = append(row.Errors, "CONFLICT_REQUIRES_MANUAL")
				row.Resolution = "REQUIRE_MANUAL"
			}
		}
		rows = append(rows, row)
	}
	for index := range rows {
		if rows[index].Normalized == nil || len(rows[index].Errors) > 0 {
			continue
		}
		for _, reference := range rows[index].Normalized.References {
			if !batchKeys[reference] && !known[reference] {
				rows[index].Status = "REFERENCE_MISSING"
				rows[index].Errors = append(rows[index].Errors, "REFERENCE_NOT_FOUND:"+reference)
			}
		}
	}
	if cycle := dependencyCycle(rows); len(cycle) > 0 {
		for index := range rows {
			if contains(cycle, rows[index].NaturalKey) {
				rows[index].Status = "INVALID"
				rows[index].Errors = append(rows[index].Errors, "DEPENDENCY_CYCLE")
			}
		}
	}
	candidates := []Record{}
	for _, row := range rows {
		if row.Normalized != nil && (row.Status == "NEW" || row.Status == "CHANGED" || row.Status == "UNCHANGED") {
			candidates = append(candidates, *row.Normalized)
		}
	}
	golden := s.golden.Run(ctx, request.TargetEntity, candidates)
	if !golden.Passed {
		for index := range rows {
			if rows[index].Status == "NEW" || rows[index].Status == "CHANGED" {
				rows[index].Status = "INVALID"
				rows[index].Errors = append(rows[index].Errors, "GOLDEN_REGRESSION")
			}
		}
	}
	summary := summarize(rows)
	status := "READY"
	if request.OperationType == "VALIDATE_ONLY" {
		status = "VALIDATED"
	}
	if summary.Conflicts+summary.Invalid+summary.ReferenceMissing > 0 || !golden.Passed {
		status = "FAILED"
	}
	plan := Plan{ContractVersion: ContractVersion, BatchID: newID(), BatchKey: request.BatchKey, OperationType: request.OperationType, Status: status, SourceType: request.SourceType, SourceName: request.SourceName, SourceFingerprint: fingerprint, DuplicateSource: duplicate, TargetEntity: request.TargetEntity, ActorID: request.ActorID, Rows: rows, Summary: summary, Golden: golden, CreatedAt: s.now()}
	if err := s.repository.SavePlan(ctx, plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (s *Service) Commit(ctx context.Context, request CommitRequest) (OperationResult, error) {
	if request.ActorRole != "PUBLISHER" && request.ActorRole != "ADMIN" {
		return OperationResult{}, ErrForbidden
	}
	plan, err := s.repository.LoadPlan(ctx, request.BatchID)
	if err != nil {
		return OperationResult{}, err
	}
	if plan.OperationType == "VALIDATE_ONLY" {
		return OperationResult{}, ErrValidateOnly
	}
	if plan.Status != "READY" {
		return OperationResult{}, ErrNotReady
	}
	count, err := s.repository.Commit(ctx, plan, request.ActorID)
	if err != nil {
		return OperationResult{}, err
	}
	return OperationResult{ContractVersion: ContractVersion, BatchID: plan.BatchID, Status: "COMPLETED", ChangedRecords: count, OccurredAt: s.now()}, nil
}

func (s *Service) Rollback(ctx context.Context, request CommitRequest) (OperationResult, error) {
	if request.ActorRole != "ADMIN" {
		return OperationResult{}, ErrForbidden
	}
	plan, err := s.repository.LoadPlan(ctx, request.BatchID)
	if err != nil {
		return OperationResult{}, err
	}
	if plan.Status == "ROLLED_BACK" {
		return OperationResult{}, ErrAlreadyRolledBack
	}
	if plan.Status != "COMPLETED" {
		return OperationResult{}, ErrNotReady
	}
	count, err := s.repository.Rollback(ctx, plan, request.ActorID)
	if err != nil {
		return OperationResult{}, err
	}
	return OperationResult{ContractVersion: ContractVersion, BatchID: plan.BatchID, Status: "ROLLED_BACK", ChangedRecords: count, OccurredAt: s.now()}, nil
}

func (s *Service) History(ctx context.Context) ([]Plan, error) { return s.repository.History(ctx) }
func (s *Service) Audit(ctx context.Context, batchID string) ([]AuditEvent, error) {
	return s.repository.Audit(ctx, batchID)
}

func mapRecord(raw map[string]any, request PrepareRequest) (Record, []string) {
	mapped := map[string]any{}
	warnings := []string{}
	for source, value := range raw {
		target := request.ColumnMapping[source]
		if target == "" {
			target = autoField(source, request.TargetEntity)
			if target != source {
				warnings = append(warnings, "AUTO_MAPPING:"+source+"->"+target)
			}
		}
		if target != "" && target != "-" {
			mapped[target] = value
		}
	}
	naturalKey := stringValue(mapped["natural_key"])
	version, _ := strconv.Atoi(stringValue(mapped["version"]))
	if number, ok := mapped["version"].(json.Number); ok {
		version, _ = strconv.Atoi(number.String())
	}
	if version == 0 {
		version = 1
	}
	status := strings.ToUpper(stringValue(mapped["status"]))
	if status == "" {
		status = "DRAFT"
	}
	references := stringSlice(mapped["references"])
	delete(mapped, "natural_key")
	delete(mapped, "version")
	delete(mapped, "status")
	delete(mapped, "references")
	normalizeDimensions(mapped)
	return Record{NaturalKey: naturalKey, Version: version, Status: status, Payload: mapped, References: references}, warnings
}

func autoField(source, target string) string {
	normalized := strings.ToLower(strings.TrimSpace(source))
	normalized = strings.NewReplacer(" ", "_", "-", "_", ".", "_").Replace(normalized)
	aliases := map[string]string{"key": "natural_key", "id": "natural_key", "naturalkey": "natural_key", "natural_key": "natural_key", "version": "version", "status": "status", "references": "references", "reference_keys": "references"}
	if value := aliases[normalized]; value != "" {
		return value
	}
	targetKey := strings.ToLower(target) + "_key"
	if normalized == targetKey || strings.HasSuffix(normalized, "_key") && !strings.Contains(normalized, "source_key") && !strings.Contains(normalized, "target_key") {
		return "natural_key"
	}
	return normalized
}

func secondaryMatch(record Record, existing map[string]Record, fields []string) (string, string) {
	if len(fields) == 0 {
		return "", "NONE"
	}
	matches := []string{}
	for key, candidate := range existing {
		equal := true
		for _, field := range fields {
			if !reflect.DeepEqual(record.Payload[field], candidate.Payload[field]) {
				equal = false
				break
			}
		}
		if equal {
			matches = append(matches, key)
		}
	}
	if len(matches) == 1 {
		return matches[0], "PROBABLE"
	}
	if len(matches) > 1 {
		return matches[0], "AMBIGUOUS"
	}
	return "", "NONE"
}

func diffRecords(before, after Record) []FieldDiff {
	diffs := []FieldDiff{}
	if before.Version != after.Version {
		diffs = append(diffs, FieldDiff{"version", before.Version, after.Version})
	}
	if before.Status != after.Status {
		diffs = append(diffs, FieldDiff{"status", before.Status, after.Status})
	}
	if !reflect.DeepEqual(before.References, after.References) {
		diffs = append(diffs, FieldDiff{"references", before.References, after.References})
	}
	keys := map[string]bool{}
	for key := range before.Payload {
		keys[key] = true
	}
	for key := range after.Payload {
		keys[key] = true
	}
	ordered := []string{}
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Strings(ordered)
	for _, key := range ordered {
		if !reflect.DeepEqual(before.Payload[key], after.Payload[key]) {
			diffs = append(diffs, FieldDiff{key, before.Payload[key], after.Payload[key]})
		}
	}
	return diffs
}

func hasCriticalDiff(diffs []FieldDiff) bool {
	for _, diff := range diffs {
		if contains([]string{"evidence_class", "dimension", "value", "relation_type", "claim_type", "hard_guardrail"}, diff.Field) || diff.Field == "status" && diff.Import == "PRODUCTION" {
			return true
		}
	}
	return false
}
func evidenceRank(value any) int {
	switch strings.ToUpper(stringValue(value)) {
	case "A":
		return 1
	case "B":
		return 2
	case "C":
		return 3
	case "D":
		return 4
	case "E":
		return 5
	default:
		return 99
	}
}
func summarize(rows []Row) Summary {
	value := Summary{Total: len(rows)}
	for _, row := range rows {
		switch row.Status {
		case "NEW":
			value.New++
		case "CHANGED":
			value.Changed++
		case "UNCHANGED":
			value.Unchanged++
		case "CONFLICT":
			value.Conflicts++
		case "INVALID":
			value.Invalid++
		case "REFERENCE_MISSING":
			value.ReferenceMissing++
		}
	}
	return value
}
func dependencyCycle(rows []Row) []string {
	graph := map[string][]string{}
	for _, row := range rows {
		if row.Normalized != nil {
			graph[row.NaturalKey] = row.Normalized.References
		}
	}
	visiting, visited := map[string]bool{}, map[string]bool{}
	var cycle []string
	var visit func(string) bool
	visit = func(node string) bool {
		if visiting[node] {
			cycle = append(cycle, node)
			return true
		}
		if visited[node] {
			return false
		}
		visiting[node] = true
		for _, next := range graph[node] {
			if _, ok := graph[next]; ok && visit(next) {
				cycle = append(cycle, node)
				return true
			}
		}
		visiting[node] = false
		visited[node] = true
		return false
	}
	for node := range graph {
		if visit(node) {
			return cycle
		}
	}
	return nil
}
func normalizeDimensions(value map[string]any) {
	for key, item := range value {
		if key == "dimension" && item == "FREE_WILL" {
			value[key] = "VOLITION"
		}
		if nested, ok := item.(map[string]any); ok {
			normalizeDimensions(nested)
		}
	}
}
func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(value)
	}
}
func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []any:
		result := []string{}
		for _, item := range typed {
			if text := stringValue(item); text != "" {
				result = append(result, text)
			}
		}
		return result
	case []string:
		return typed
	case string:
		if strings.TrimSpace(typed) == "" {
			return []string{}
		}
		parts := strings.Split(typed, ",")
		for index := range parts {
			parts[index] = strings.TrimSpace(parts[index])
		}
		return parts
	default:
		return []string{}
	}
}
func contains[T comparable](values []T, want T) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
func newID() string {
	value := make([]byte, 16)
	_, _ = rand.Read(value)
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	encoded := hex.EncodeToString(value)
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:]
}
