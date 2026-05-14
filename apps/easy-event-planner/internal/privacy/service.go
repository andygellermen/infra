package privacy

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	ActionAnonymize = "anonymize"
	ActionDelete    = "delete"

	CategoryParticipantContactData = "participant_contact_data"
	CategoryMagicLinks             = "magic_links"
	CategorySessions               = "sessions"
	CategoryEmailJobs              = "email_jobs"
	CategoryAuditLogs              = "audit_logs"
)

var (
	ErrPolicyNotFound = errors.New("retention policy not found")
)

type Policy struct {
	ID            string
	TenantID      string
	DataCategory  string
	Action        string
	RetentionDays int
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UpdatePolicyParams struct {
	Action        *string
	RetentionDays *int
	Enabled       *bool
}

type ExecuteInput struct {
	TenantID    string
	ActorUserID string
	RequestIP   string
	UserAgent   string
	DryRun      bool
}

type CategoryResult struct {
	PolicyID      string
	DataCategory  string
	Action        string
	RetentionDays int
	Enabled       bool
	CutoffAt      time.Time
	Affected      int
	Executed      int
	Note          string
}

type ExecuteResult struct {
	DryRun        bool
	StartedAt     time.Time
	FinishedAt    time.Time
	TotalAffected int
	TotalExecuted int
	AuditID       string
	Items         []CategoryResult
}

type JobRecord struct {
	ID            string
	TenantID      string
	Action        string
	ActorUserID   string
	DryRun        bool
	TotalAffected int
	TotalExecuted int
	CreatedAt     time.Time
	Items         []CategoryResult
}

type Service struct {
	db    *sql.DB
	nowFn func() time.Time
	idFn  func(prefix string) string
}

type rowScanner interface {
	Scan(dest ...any) error
}

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type defaultPolicy struct {
	Category         string
	Action           string
	RetentionDays    int
	UseTenantDefault bool
}

var defaultPolicies = []defaultPolicy{
	{Category: CategoryParticipantContactData, Action: ActionAnonymize, UseTenantDefault: true},
	{Category: CategoryMagicLinks, Action: ActionDelete, RetentionDays: 7},
	{Category: CategorySessions, Action: ActionDelete, RetentionDays: 30},
	{Category: CategoryEmailJobs, Action: ActionDelete, RetentionDays: 90},
	{Category: CategoryAuditLogs, Action: ActionDelete, RetentionDays: 180},
}

func NewService(sqlDB *sql.DB) *Service {
	return &Service{
		db:    sqlDB,
		nowFn: func() time.Time { return time.Now().UTC() },
		idFn:  defaultID,
	}
}

func (s *Service) ListPolicies(ctx context.Context, tenantID string) ([]Policy, error) {
	if s.db == nil {
		return nil, fmt.Errorf("privacy service database is nil")
	}
	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}
	if err := s.ensureDefaultPolicies(ctx, tenant); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, data_category, action, retention_days, enabled, created_at, updated_at
     FROM retention_policies
     WHERE tenant_id = ?
     ORDER BY data_category ASC`,
		tenant,
	)
	if err != nil {
		return nil, fmt.Errorf("list retention policies: %w", err)
	}
	defer rows.Close()

	items := make([]Policy, 0)
	for rows.Next() {
		item, scanErr := scanPolicy(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate retention policies: %w", err)
	}
	return items, nil
}

func (s *Service) UpdatePolicy(ctx context.Context, tenantID, policyID string, params UpdatePolicyParams) (Policy, error) {
	if s.db == nil {
		return Policy{}, fmt.Errorf("privacy service database is nil")
	}
	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return Policy{}, fmt.Errorf("tenant id must not be empty")
	}
	id := strings.TrimSpace(policyID)
	if id == "" {
		return Policy{}, fmt.Errorf("policy id must not be empty")
	}

	if err := s.ensureDefaultPolicies(ctx, tenant); err != nil {
		return Policy{}, err
	}

	current, err := s.getPolicyByID(ctx, tenant, id)
	if err != nil {
		return Policy{}, err
	}

	updated := current
	hasChange := false
	if params.Action != nil {
		action, actionErr := normalizeAction(*params.Action)
		if actionErr != nil {
			return Policy{}, actionErr
		}
		updated.Action = action
		hasChange = true
	}
	if params.RetentionDays != nil {
		retentionDays := *params.RetentionDays
		if retentionDays <= 0 {
			return Policy{}, fmt.Errorf("retention_days must be > 0")
		}
		updated.RetentionDays = retentionDays
		hasChange = true
	}
	if params.Enabled != nil {
		updated.Enabled = *params.Enabled
		hasChange = true
	}
	if !hasChange {
		return Policy{}, fmt.Errorf("at least one field must be set for update")
	}

	now := s.nowFn().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE retention_policies
     SET action = ?, retention_days = ?, enabled = ?, updated_at = ?
     WHERE tenant_id = ? AND id = ?`,
		updated.Action,
		updated.RetentionDays,
		boolToInt(updated.Enabled),
		now,
		tenant,
		id,
	)
	if err != nil {
		return Policy{}, fmt.Errorf("update retention policy: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return Policy{}, fmt.Errorf("read retention policy update rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return Policy{}, ErrPolicyNotFound
	}
	return s.getPolicyByID(ctx, tenant, id)
}

func (s *Service) Execute(ctx context.Context, input ExecuteInput) (ExecuteResult, error) {
	if s.db == nil {
		return ExecuteResult{}, fmt.Errorf("privacy service database is nil")
	}
	tenant := strings.TrimSpace(input.TenantID)
	if tenant == "" {
		return ExecuteResult{}, fmt.Errorf("tenant id must not be empty")
	}
	if err := s.ensureDefaultPolicies(ctx, tenant); err != nil {
		return ExecuteResult{}, err
	}

	policies, err := s.ListPolicies(ctx, tenant)
	if err != nil {
		return ExecuteResult{}, err
	}

	startedAt := s.nowFn().UTC()
	result := ExecuteResult{
		DryRun:    input.DryRun,
		StartedAt: startedAt,
		Items:     make([]CategoryResult, 0, len(policies)),
	}

	var executor sqlExecutor = s.db
	var tx *sql.Tx
	if !input.DryRun {
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return ExecuteResult{}, fmt.Errorf("begin retention run transaction: %w", err)
		}
		executor = tx
	}

	for _, policy := range policies {
		item := CategoryResult{
			PolicyID:      policy.ID,
			DataCategory:  policy.DataCategory,
			Action:        policy.Action,
			RetentionDays: policy.RetentionDays,
			Enabled:       policy.Enabled,
			CutoffAt:      startedAt.AddDate(0, 0, -policy.RetentionDays),
		}

		if !policy.Enabled {
			item.Note = "policy_disabled"
			result.Items = append(result.Items, item)
			continue
		}

		categoryResult, categoryErr := s.executePolicy(ctx, executor, tenant, policy, startedAt, input.DryRun)
		if categoryErr != nil {
			if tx != nil {
				_ = tx.Rollback()
			}
			return ExecuteResult{}, categoryErr
		}
		result.Items = append(result.Items, categoryResult)
		result.TotalAffected += categoryResult.Affected
		result.TotalExecuted += categoryResult.Executed
	}

	result.FinishedAt = s.nowFn().UTC()
	auditID, err := s.insertRunAudit(
		ctx,
		executor,
		tenant,
		strings.TrimSpace(input.ActorUserID),
		strings.TrimSpace(input.RequestIP),
		strings.TrimSpace(input.UserAgent),
		result,
	)
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		return ExecuteResult{}, err
	}
	result.AuditID = auditID

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return ExecuteResult{}, fmt.Errorf("commit retention run transaction: %w", err)
		}
	}

	return result, nil
}

func (s *Service) ListJobs(ctx context.Context, tenantID string, limit int) ([]JobRecord, error) {
	if s.db == nil {
		return nil, fmt.Errorf("privacy service database is nil")
	}
	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return nil, fmt.Errorf("tenant id must not be empty")
	}
	if limit <= 0 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, tenant_id, action, COALESCE(actor_user_id, ''), COALESCE(details_json, ''), created_at
     FROM audit_log
     WHERE tenant_id = ?
       AND action IN ('retention_job_run', 'retention_job_dry_run')
     ORDER BY created_at DESC
     LIMIT ?`,
		tenant,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list retention jobs: %w", err)
	}
	defer rows.Close()

	items := make([]JobRecord, 0)
	for rows.Next() {
		var (
			item       JobRecord
			detailsRaw string
			createdRaw string
		)
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.Action,
			&item.ActorUserID,
			&detailsRaw,
			&createdRaw,
		); err != nil {
			return nil, fmt.Errorf("scan retention job: %w", err)
		}
		createdAt, err := time.Parse(time.RFC3339, createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse retention job created_at: %w", err)
		}
		item.CreatedAt = createdAt.UTC()

		details := parseAuditDetails(detailsRaw)
		item.DryRun = details.DryRun
		item.TotalAffected = details.TotalAffected
		item.TotalExecuted = details.TotalExecuted
		item.Items = details.Items
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate retention jobs: %w", err)
	}
	return items, nil
}

func (s *Service) executePolicy(
	ctx context.Context,
	executor sqlExecutor,
	tenantID string,
	policy Policy,
	now time.Time,
	dryRun bool,
) (CategoryResult, error) {
	item := CategoryResult{
		PolicyID:      policy.ID,
		DataCategory:  policy.DataCategory,
		Action:        policy.Action,
		RetentionDays: policy.RetentionDays,
		Enabled:       policy.Enabled,
		CutoffAt:      now.AddDate(0, 0, -policy.RetentionDays),
	}

	switch policy.DataCategory {
	case CategoryParticipantContactData:
		affected, err := countParticipantCandidates(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Affected = affected
		if dryRun {
			return item, nil
		}
		executed, err := anonymizeParticipants(ctx, executor, tenantID, item.CutoffAt, now)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Executed = executed
		return item, nil
	case CategoryMagicLinks:
		affected, err := countMagicLinks(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Affected = affected
		if dryRun {
			return item, nil
		}
		executed, err := deleteMagicLinks(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Executed = executed
		return item, nil
	case CategorySessions:
		affected, err := countSessions(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Affected = affected
		if dryRun {
			return item, nil
		}
		executed, err := deleteSessions(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Executed = executed
		return item, nil
	case CategoryEmailJobs:
		affected, err := countEmailJobs(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Affected = affected
		if dryRun {
			return item, nil
		}
		executed, err := deleteEmailJobs(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Executed = executed
		return item, nil
	case CategoryAuditLogs:
		affected, err := countAuditLogs(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Affected = affected
		if dryRun {
			return item, nil
		}
		executed, err := deleteAuditLogs(ctx, executor, tenantID, item.CutoffAt)
		if err != nil {
			return CategoryResult{}, err
		}
		item.Executed = executed
		return item, nil
	default:
		item.Note = "category_not_supported"
		return item, nil
	}
}

func (s *Service) ensureDefaultPolicies(ctx context.Context, tenantID string) error {
	tenant := strings.TrimSpace(tenantID)
	if tenant == "" {
		return fmt.Errorf("tenant id must not be empty")
	}

	retentionDays := 30
	row := s.db.QueryRowContext(
		ctx,
		`SELECT default_retention_days FROM tenant_settings WHERE tenant_id = ? LIMIT 1`,
		tenant,
	)
	var configured int
	err := row.Scan(&configured)
	switch {
	case err == nil:
		if configured > 0 {
			retentionDays = configured
		}
	case errors.Is(err, sql.ErrNoRows):
		// keep fallback
	default:
		return fmt.Errorf("query default retention days: %w", err)
	}

	now := s.nowFn().UTC().Format(time.RFC3339)
	for _, def := range defaultPolicies {
		days := def.RetentionDays
		if def.UseTenantDefault {
			days = retentionDays
		}
		if days <= 0 {
			days = 30
		}

		if _, err := s.db.ExecContext(
			ctx,
			`INSERT INTO retention_policies (
        id, tenant_id, data_category, action, retention_days, enabled, created_at, updated_at
      ) VALUES (?, ?, ?, ?, ?, 1, ?, ?)
      ON CONFLICT(tenant_id, data_category) DO NOTHING`,
			s.idFn("rtp"),
			tenant,
			def.Category,
			def.Action,
			days,
			now,
			now,
		); err != nil {
			return fmt.Errorf("insert default retention policy for %s: %w", def.Category, err)
		}
	}
	return nil
}

func (s *Service) getPolicyByID(ctx context.Context, tenantID, policyID string) (Policy, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, data_category, action, retention_days, enabled, created_at, updated_at
     FROM retention_policies
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenantID,
		policyID,
	)
	item, err := scanPolicy(row)
	if err != nil {
		return Policy{}, err
	}
	return item, nil
}

func (s *Service) insertRunAudit(
	ctx context.Context,
	executor sqlExecutor,
	tenantID string,
	actorUserID string,
	requestIP string,
	userAgent string,
	result ExecuteResult,
) (string, error) {
	action := "retention_job_run"
	if result.DryRun {
		action = "retention_job_dry_run"
	}

	categories := make([]map[string]any, 0, len(result.Items))
	for _, item := range result.Items {
		categoryPayload := map[string]any{
			"policy_id":      item.PolicyID,
			"data_category":  item.DataCategory,
			"action":         item.Action,
			"retention_days": item.RetentionDays,
			"enabled":        item.Enabled,
			"cutoff_at":      item.CutoffAt.UTC().Format(time.RFC3339),
			"affected":       item.Affected,
			"executed":       item.Executed,
		}
		if strings.TrimSpace(item.Note) != "" {
			categoryPayload["note"] = item.Note
		}
		categories = append(categories, categoryPayload)
	}

	details := map[string]any{
		"dry_run":        result.DryRun,
		"started_at":     result.StartedAt.UTC().Format(time.RFC3339),
		"finished_at":    result.FinishedAt.UTC().Format(time.RFC3339),
		"total_affected": result.TotalAffected,
		"total_executed": result.TotalExecuted,
		"categories":     categories,
	}
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return "", fmt.Errorf("marshal retention audit details: %w", err)
	}

	auditID := s.idFn("aud")
	_, err = executor.ExecContext(
		ctx,
		`INSERT INTO audit_log (
      id, tenant_id, actor_user_id, actor_participant_id, action, entity_type, entity_id,
      details_json, request_ip, user_agent, created_at
    ) VALUES (?, ?, ?, NULL, ?, 'privacy', NULL, ?, ?, ?, ?)`,
		auditID,
		nullable(tenantID),
		nullable(actorUserID),
		action,
		nullable(string(detailsJSON)),
		nullable(requestIP),
		nullable(userAgent),
		s.nowFn().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return "", fmt.Errorf("insert retention audit log: %w", err)
	}
	return auditID, nil
}

func scanPolicy(row rowScanner) (Policy, error) {
	var (
		item       Policy
		enabledInt int
		createdRaw string
		updatedRaw string
	)
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.DataCategory,
		&item.Action,
		&item.RetentionDays,
		&enabledInt,
		&createdRaw,
		&updatedRaw,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Policy{}, ErrPolicyNotFound
		}
		return Policy{}, fmt.Errorf("scan retention policy: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, createdRaw)
	if err != nil {
		return Policy{}, fmt.Errorf("parse retention policy created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedRaw)
	if err != nil {
		return Policy{}, fmt.Errorf("parse retention policy updated_at: %w", err)
	}

	item.Action = strings.ToLower(strings.TrimSpace(item.Action))
	item.Enabled = enabledInt == 1
	item.CreatedAt = createdAt.UTC()
	item.UpdatedAt = updatedAt.UTC()
	return item, nil
}

func normalizeAction(raw string) (string, error) {
	action := strings.ToLower(strings.TrimSpace(raw))
	switch action {
	case ActionAnonymize, ActionDelete:
		return action, nil
	default:
		return "", fmt.Errorf("unsupported action %q", raw)
	}
}

func countParticipantCandidates(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	row := executor.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
     FROM participants p
     WHERE p.tenant_id = ?
       AND p.anonymized_at IS NULL
       AND EXISTS (
         SELECT 1
         FROM registrations r
         JOIN events e ON e.id = r.event_id
         WHERE r.tenant_id = p.tenant_id
           AND r.participant_id = p.id
           AND e.tenant_id = r.tenant_id
           AND datetime(COALESCE(e.ends_at, e.starts_at)) <= datetime(?)
       )
       AND NOT EXISTS (
         SELECT 1
         FROM registrations r2
         JOIN events e2 ON e2.id = r2.event_id
         WHERE r2.tenant_id = p.tenant_id
           AND r2.participant_id = p.id
           AND e2.tenant_id = r2.tenant_id
           AND datetime(COALESCE(e2.ends_at, e2.starts_at)) > datetime(?)
       )`,
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
		cutoff.UTC().Format(time.RFC3339),
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count participant anonymization candidates: %w", err)
	}
	return count, nil
}

func anonymizeParticipants(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time, now time.Time) (int, error) {
	result, err := executor.ExecContext(
		ctx,
		`UPDATE participants
     SET name = NULL,
         email = NULL,
         phone = NULL,
         anonymized_at = ?,
         updated_at = ?
     WHERE tenant_id = ?
       AND anonymized_at IS NULL
       AND id IN (
         SELECT p.id
         FROM participants p
         WHERE p.tenant_id = ?
           AND p.anonymized_at IS NULL
           AND EXISTS (
             SELECT 1
             FROM registrations r
             JOIN events e ON e.id = r.event_id
             WHERE r.tenant_id = p.tenant_id
               AND r.participant_id = p.id
               AND e.tenant_id = r.tenant_id
               AND datetime(COALESCE(e.ends_at, e.starts_at)) <= datetime(?)
           )
           AND NOT EXISTS (
             SELECT 1
             FROM registrations r2
             JOIN events e2 ON e2.id = r2.event_id
             WHERE r2.tenant_id = p.tenant_id
               AND r2.participant_id = p.id
               AND e2.tenant_id = r2.tenant_id
               AND datetime(COALESCE(e2.ends_at, e2.starts_at)) > datetime(?)
           )
       )`,
		now.UTC().Format(time.RFC3339),
		now.UTC().Format(time.RFC3339),
		tenantID,
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("anonymize participants: %w", err)
	}
	return rowsAffected(result, "anonymize participants")
}

func countMagicLinks(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	return countByQuery(
		ctx,
		executor,
		`SELECT COUNT(*)
     FROM magic_links
     WHERE tenant_id = ?
       AND datetime(expires_at) <= datetime(?)`,
		"count retention magic links",
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
	)
}

func deleteMagicLinks(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	result, err := executor.ExecContext(
		ctx,
		`DELETE FROM magic_links
     WHERE tenant_id = ?
       AND datetime(expires_at) <= datetime(?)`,
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("delete retention magic links: %w", err)
	}
	return rowsAffected(result, "delete retention magic links")
}

func countSessions(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	return countByQuery(
		ctx,
		executor,
		`SELECT COUNT(*)
     FROM sessions
     WHERE tenant_id = ?
       AND datetime(expires_at) <= datetime(?)`,
		"count retention sessions",
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
	)
}

func deleteSessions(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	result, err := executor.ExecContext(
		ctx,
		`DELETE FROM sessions
     WHERE tenant_id = ?
       AND datetime(expires_at) <= datetime(?)`,
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("delete retention sessions: %w", err)
	}
	return rowsAffected(result, "delete retention sessions")
}

func countEmailJobs(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	return countByQuery(
		ctx,
		executor,
		`SELECT COUNT(*)
     FROM email_jobs
     WHERE tenant_id = ?
       AND status IN ('sent', 'failed')
       AND datetime(COALESCE(sent_at, updated_at, created_at)) <= datetime(?)`,
		"count retention email jobs",
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
	)
}

func deleteEmailJobs(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	result, err := executor.ExecContext(
		ctx,
		`DELETE FROM email_jobs
     WHERE tenant_id = ?
       AND status IN ('sent', 'failed')
       AND datetime(COALESCE(sent_at, updated_at, created_at)) <= datetime(?)`,
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("delete retention email jobs: %w", err)
	}
	return rowsAffected(result, "delete retention email jobs")
}

func countAuditLogs(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	return countByQuery(
		ctx,
		executor,
		`SELECT COUNT(*)
     FROM audit_log
     WHERE tenant_id = ?
       AND datetime(created_at) <= datetime(?)`,
		"count retention audit logs",
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
	)
}

func deleteAuditLogs(ctx context.Context, executor sqlExecutor, tenantID string, cutoff time.Time) (int, error) {
	result, err := executor.ExecContext(
		ctx,
		`DELETE FROM audit_log
     WHERE tenant_id = ?
       AND datetime(created_at) <= datetime(?)`,
		tenantID,
		cutoff.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("delete retention audit logs: %w", err)
	}
	return rowsAffected(result, "delete retention audit logs")
}

func countByQuery(ctx context.Context, executor sqlExecutor, query string, label string, args ...any) (int, error) {
	row := executor.QueryRowContext(ctx, query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("%s: %w", label, err)
	}
	return count, nil
}

func rowsAffected(result sql.Result, label string) (int, error) {
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("%s rows affected: %w", label, err)
	}
	return int(affected), nil
}

type auditDetails struct {
	DryRun        bool
	TotalAffected int
	TotalExecuted int
	Items         []CategoryResult
}

func parseAuditDetails(raw string) auditDetails {
	value := strings.TrimSpace(raw)
	if value == "" {
		return auditDetails{}
	}

	var payload struct {
		DryRun        bool `json:"dry_run"`
		TotalAffected int  `json:"total_affected"`
		TotalExecuted int  `json:"total_executed"`
		Categories    []struct {
			PolicyID      string `json:"policy_id"`
			DataCategory  string `json:"data_category"`
			Action        string `json:"action"`
			RetentionDays int    `json:"retention_days"`
			Enabled       bool   `json:"enabled"`
			CutoffAt      string `json:"cutoff_at"`
			Affected      int    `json:"affected"`
			Executed      int    `json:"executed"`
			Note          string `json:"note"`
		} `json:"categories"`
	}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return auditDetails{}
	}

	items := make([]CategoryResult, 0, len(payload.Categories))
	for _, category := range payload.Categories {
		cutoff, err := time.Parse(time.RFC3339, category.CutoffAt)
		if err != nil {
			cutoff = time.Time{}
		}
		items = append(items, CategoryResult{
			PolicyID:      category.PolicyID,
			DataCategory:  category.DataCategory,
			Action:        category.Action,
			RetentionDays: category.RetentionDays,
			Enabled:       category.Enabled,
			CutoffAt:      cutoff.UTC(),
			Affected:      category.Affected,
			Executed:      category.Executed,
			Note:          category.Note,
		})
	}

	return auditDetails{
		DryRun:        payload.DryRun,
		TotalAffected: payload.TotalAffected,
		TotalExecuted: payload.TotalExecuted,
		Items:         items,
	}
}

func boolToInt(flag bool) int {
	if flag {
		return 1
	}
	return 0
}

func nullable(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func defaultID(prefix string) string {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return fmt.Sprintf("%s_%x", prefix, random)
}
