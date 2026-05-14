package privacy

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func setupPrivacyService(t *testing.T) (*Service, *sql.DB, string, time.Time) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "privacy-test.sqlite"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	migrator := db.NewMigrator(sqlDB, migrations.Files, ".")
	if _, err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	tenantRepo := tenant.NewRepository(sqlDB)
	createdTenant, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "privacy-demo",
		Name:          "Privacy Demo",
		PublicBaseURL: "https://events.example.com/privacy-demo",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	service := NewService(sqlDB)
	service.nowFn = func() time.Time { return now }

	return service, sqlDB, createdTenant.ID, now
}

func seedPrivacyFixtures(t *testing.T, sqlDB *sql.DB, tenantID string, now time.Time) {
	t.Helper()

	timestamps := map[string]string{
		"old_event_start":  now.AddDate(0, 0, -46).Format(time.RFC3339),
		"old_event_end":    now.AddDate(0, 0, -45).Format(time.RFC3339),
		"future_event":     now.AddDate(0, 0, 7).Format(time.RFC3339),
		"recent_event":     now.AddDate(0, 0, -5).Format(time.RFC3339),
		"created_old":      now.AddDate(0, 0, -120).Format(time.RFC3339),
		"created_recent":   now.AddDate(0, 0, -2).Format(time.RFC3339),
		"magic_old_expiry": now.AddDate(0, 0, -15).Format(time.RFC3339),
		"magic_recent":     now.AddDate(0, 0, -2).Format(time.RFC3339),
		"session_old":      now.AddDate(0, 0, -40).Format(time.RFC3339),
		"session_recent":   now.AddDate(0, 0, 2).Format(time.RFC3339),
		"email_old":        now.AddDate(0, 0, -120).Format(time.RFC3339),
		"email_failed":     now.AddDate(0, 0, -100).Format(time.RFC3339),
		"email_recent":     now.AddDate(0, 0, -5).Format(time.RFC3339),
		"audit_old":        now.AddDate(0, 0, -220).Format(time.RFC3339),
		"audit_recent":     now.AddDate(0, 0, -1).Format(time.RFC3339),
		"now":              now.Format(time.RFC3339),
	}

	mustExec := func(query string, args ...any) {
		t.Helper()
		if _, err := sqlDB.ExecContext(context.Background(), query, args...); err != nil {
			t.Fatalf("exec %q: %v", query, err)
		}
	}

	mustExec(
		`INSERT INTO participants (id, tenant_id, email, phone, name, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?)`,
		"prt_old",
		tenantID,
		"old@example.com",
		"+491111111",
		"Old Participant",
		timestamps["created_old"],
		timestamps["created_old"],
		"prt_mixed",
		tenantID,
		"mixed@example.com",
		"+492222222",
		"Mixed Participant",
		timestamps["created_old"],
		timestamps["created_old"],
		"prt_recent",
		tenantID,
		"recent@example.com",
		"+493333333",
		"Recent Participant",
		timestamps["created_recent"],
		timestamps["created_recent"],
	)

	mustExec(
		`INSERT INTO events (id, tenant_id, slug, title, starts_at, ends_at, timezone, status, is_public, registration_enabled, waitlist_enabled, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, 'Europe/Berlin', 'completed', 1, 1, 1, ?, ?),
            (?, ?, ?, ?, ?, NULL, 'Europe/Berlin', 'scheduled', 1, 1, 1, ?, ?),
            (?, ?, ?, ?, ?, NULL, 'Europe/Berlin', 'scheduled', 1, 1, 1, ?, ?)`,
		"evt_old",
		tenantID,
		"old-event",
		"Old Event",
		timestamps["old_event_start"],
		timestamps["old_event_end"],
		timestamps["created_old"],
		timestamps["created_old"],
		"evt_future",
		tenantID,
		"future-event",
		"Future Event",
		timestamps["future_event"],
		timestamps["created_recent"],
		timestamps["created_recent"],
		"evt_recent",
		tenantID,
		"recent-event",
		"Recent Event",
		timestamps["recent_event"],
		timestamps["created_recent"],
		timestamps["created_recent"],
	)

	mustExec(
		`INSERT INTO registrations (id, tenant_id, event_id, participant_id, status, participation_type, quantity, source, created_at, updated_at)
     VALUES (?, ?, ?, ?, 'confirmed', 'onsite', 1, 'public_page', ?, ?),
            (?, ?, ?, ?, 'confirmed', 'onsite', 1, 'public_page', ?, ?),
            (?, ?, ?, ?, 'confirmed', 'onsite', 1, 'public_page', ?, ?),
            (?, ?, ?, ?, 'confirmed', 'onsite', 1, 'public_page', ?, ?)`,
		"reg_old",
		tenantID,
		"evt_old",
		"prt_old",
		timestamps["created_old"],
		timestamps["created_old"],
		"reg_mixed_old",
		tenantID,
		"evt_old",
		"prt_mixed",
		timestamps["created_old"],
		timestamps["created_old"],
		"reg_mixed_future",
		tenantID,
		"evt_future",
		"prt_mixed",
		timestamps["created_recent"],
		timestamps["created_recent"],
		"reg_recent",
		tenantID,
		"evt_recent",
		"prt_recent",
		timestamps["created_recent"],
		timestamps["created_recent"],
	)

	mustExec(
		`INSERT INTO magic_links (id, tenant_id, purpose, token_hash, expires_at, created_at)
     VALUES (?, ?, 'organizer_login', ?, ?, ?), (?, ?, 'organizer_login', ?, ?, ?)`,
		"mlk_old",
		tenantID,
		"token-hash-old",
		timestamps["magic_old_expiry"],
		timestamps["created_old"],
		"mlk_recent",
		tenantID,
		"token-hash-recent",
		timestamps["magic_recent"],
		timestamps["created_recent"],
	)

	mustExec(
		`INSERT INTO sessions (id, tenant_id, session_hash, expires_at, created_at)
     VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)`,
		"ses_old",
		tenantID,
		"session-hash-old",
		timestamps["session_old"],
		timestamps["created_old"],
		"ses_recent",
		tenantID,
		"session-hash-recent",
		timestamps["session_recent"],
		timestamps["created_recent"],
	)

	mustExec(
		`INSERT INTO email_jobs (id, tenant_id, template_key, recipient_email, subject, body_text, status, sent_at, created_at, updated_at)
     VALUES (?, ?, 'registration_confirmed', 'old1@example.com', 'old', 'old', 'sent', ?, ?, ?),
            (?, ?, 'registration_confirmed', 'old2@example.com', 'old failed', 'old failed', 'failed', NULL, ?, ?),
            (?, ?, 'registration_confirmed', 'recent@example.com', 'recent', 'recent', 'sent', ?, ?, ?),
            (?, ?, 'registration_confirmed', 'queued@example.com', 'queued', 'queued', 'queued', NULL, ?, ?)`,
		"eml_old_sent",
		tenantID,
		timestamps["email_old"],
		timestamps["email_old"],
		timestamps["email_old"],
		"eml_old_failed",
		tenantID,
		timestamps["email_failed"],
		timestamps["email_failed"],
		"eml_recent_sent",
		tenantID,
		timestamps["email_recent"],
		timestamps["email_recent"],
		timestamps["email_recent"],
		"eml_old_queued",
		tenantID,
		timestamps["email_old"],
		timestamps["email_old"],
	)

	mustExec(
		`INSERT INTO audit_log (id, tenant_id, action, created_at)
     VALUES (?, ?, 'old_audit_action', ?), (?, ?, 'recent_audit_action', ?)`,
		"aud_old",
		tenantID,
		timestamps["audit_old"],
		"aud_recent",
		tenantID,
		timestamps["audit_recent"],
	)
}

func TestListPoliciesBootstrapsDefaults(t *testing.T) {
	service, _, tenantID, _ := setupPrivacyService(t)

	items, err := service.ListPolicies(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListPolicies returned error: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("expected 5 default policies, got %d", len(items))
	}

	byCategory := make(map[string]Policy, len(items))
	for _, item := range items {
		byCategory[item.DataCategory] = item
	}

	if byCategory[CategoryParticipantContactData].Action != ActionAnonymize {
		t.Fatalf("expected participant policy action anonymize")
	}
	if byCategory[CategoryParticipantContactData].RetentionDays != 30 {
		t.Fatalf("expected participant retention 30, got %d", byCategory[CategoryParticipantContactData].RetentionDays)
	}
	if byCategory[CategoryMagicLinks].RetentionDays != 7 {
		t.Fatalf("expected magic link retention 7, got %d", byCategory[CategoryMagicLinks].RetentionDays)
	}
	if byCategory[CategorySessions].RetentionDays != 30 {
		t.Fatalf("expected session retention 30, got %d", byCategory[CategorySessions].RetentionDays)
	}
	if byCategory[CategoryEmailJobs].RetentionDays != 90 {
		t.Fatalf("expected email jobs retention 90, got %d", byCategory[CategoryEmailJobs].RetentionDays)
	}
	if byCategory[CategoryAuditLogs].RetentionDays != 180 {
		t.Fatalf("expected audit retention 180, got %d", byCategory[CategoryAuditLogs].RetentionDays)
	}
}

func TestUpdatePolicy(t *testing.T) {
	service, _, tenantID, _ := setupPrivacyService(t)

	items, err := service.ListPolicies(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("ListPolicies returned error: %v", err)
	}

	var magicPolicy Policy
	for _, item := range items {
		if item.DataCategory == CategoryMagicLinks {
			magicPolicy = item
			break
		}
	}
	if magicPolicy.ID == "" {
		t.Fatal("expected magic link policy to exist")
	}

	enabled := false
	retentionDays := 14
	action := ActionDelete
	updated, err := service.UpdatePolicy(context.Background(), tenantID, magicPolicy.ID, UpdatePolicyParams{
		Action:        &action,
		RetentionDays: &retentionDays,
		Enabled:       &enabled,
	})
	if err != nil {
		t.Fatalf("UpdatePolicy returned error: %v", err)
	}

	if updated.RetentionDays != 14 {
		t.Fatalf("expected updated retention 14, got %d", updated.RetentionDays)
	}
	if updated.Enabled {
		t.Fatalf("expected policy to be disabled")
	}
	if updated.Action != ActionDelete {
		t.Fatalf("expected delete action, got %q", updated.Action)
	}
}

func TestUpdatePolicyNotFound(t *testing.T) {
	service, _, tenantID, _ := setupPrivacyService(t)

	retentionDays := 10
	_, err := service.UpdatePolicy(context.Background(), tenantID, "missing-policy", UpdatePolicyParams{
		RetentionDays: &retentionDays,
	})
	if !errors.Is(err, ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}
}

func TestExecuteDryRunAndRun(t *testing.T) {
	service, sqlDB, tenantID, now := setupPrivacyService(t)
	seedPrivacyFixtures(t, sqlDB, tenantID, now)

	dryResult, err := service.Execute(context.Background(), ExecuteInput{
		TenantID:    tenantID,
		ActorUserID: "usr_test",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("Execute dry-run returned error: %v", err)
	}
	if !dryResult.DryRun {
		t.Fatalf("expected dry run result")
	}
	if dryResult.TotalAffected != 6 {
		t.Fatalf("expected dry-run total_affected=6, got %d", dryResult.TotalAffected)
	}
	if dryResult.TotalExecuted != 0 {
		t.Fatalf("expected dry-run total_executed=0, got %d", dryResult.TotalExecuted)
	}

	var dryParticipantEmail sql.NullString
	var dryAnonymizedAt sql.NullString
	if err := sqlDB.QueryRowContext(
		context.Background(),
		`SELECT email, anonymized_at FROM participants WHERE id = ?`,
		"prt_old",
	).Scan(&dryParticipantEmail, &dryAnonymizedAt); err != nil {
		t.Fatalf("query participant after dry-run: %v", err)
	}
	if !dryParticipantEmail.Valid || dryParticipantEmail.String == "" {
		t.Fatalf("expected participant email untouched after dry-run")
	}
	if dryAnonymizedAt.Valid {
		t.Fatalf("expected anonymized_at to remain NULL after dry-run")
	}

	runResult, err := service.Execute(context.Background(), ExecuteInput{
		TenantID:    tenantID,
		ActorUserID: "usr_test",
		DryRun:      false,
	})
	if err != nil {
		t.Fatalf("Execute run returned error: %v", err)
	}
	if runResult.DryRun {
		t.Fatalf("expected normal run result")
	}
	if runResult.TotalAffected != 6 {
		t.Fatalf("expected run total_affected=6, got %d", runResult.TotalAffected)
	}
	if runResult.TotalExecuted != 6 {
		t.Fatalf("expected run total_executed=6, got %d", runResult.TotalExecuted)
	}
	if runResult.AuditID == "" {
		t.Fatalf("expected run audit id")
	}

	var (
		emailValue       sql.NullString
		phoneValue       sql.NullString
		nameValue        sql.NullString
		anonymizedAtRaw  sql.NullString
		mixedParticipant sql.NullString
	)
	if err := sqlDB.QueryRowContext(
		context.Background(),
		`SELECT email, phone, name, anonymized_at FROM participants WHERE id = ?`,
		"prt_old",
	).Scan(&emailValue, &phoneValue, &nameValue, &anonymizedAtRaw); err != nil {
		t.Fatalf("query anonymized participant: %v", err)
	}
	if emailValue.Valid || phoneValue.Valid || nameValue.Valid {
		t.Fatalf("expected participant PII to be anonymized")
	}
	if !anonymizedAtRaw.Valid || strings.TrimSpace(anonymizedAtRaw.String) == "" {
		t.Fatalf("expected anonymized_at to be set")
	}
	if err := sqlDB.QueryRowContext(
		context.Background(),
		`SELECT email FROM participants WHERE id = ?`,
		"prt_mixed",
	).Scan(&mixedParticipant); err != nil {
		t.Fatalf("query mixed participant: %v", err)
	}
	if !mixedParticipant.Valid || strings.TrimSpace(mixedParticipant.String) == "" {
		t.Fatalf("expected mixed participant to remain identifiable")
	}

	assertCount := func(query string, want int) {
		t.Helper()
		var got int
		if err := sqlDB.QueryRowContext(context.Background(), query, tenantID).Scan(&got); err != nil {
			t.Fatalf("query count %q failed: %v", query, err)
		}
		if got != want {
			t.Fatalf("expected count %d for query %q, got %d", want, query, got)
		}
	}

	assertCount(`SELECT COUNT(*) FROM magic_links WHERE tenant_id = ?`, 1)
	assertCount(`SELECT COUNT(*) FROM sessions WHERE tenant_id = ?`, 1)
	assertCount(`SELECT COUNT(*) FROM email_jobs WHERE tenant_id = ?`, 2)
	assertCount(`SELECT COUNT(*) FROM audit_log WHERE tenant_id = ?`, 3)

	jobs, err := service.ListJobs(context.Background(), tenantID, 10)
	if err != nil {
		t.Fatalf("ListJobs returned error: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 retention job records, got %d", len(jobs))
	}
	if jobs[0].Action != "retention_job_run" {
		t.Fatalf("expected latest action retention_job_run, got %q", jobs[0].Action)
	}
	if jobs[1].Action != "retention_job_dry_run" {
		t.Fatalf("expected second action retention_job_dry_run, got %q", jobs[1].Action)
	}
}
