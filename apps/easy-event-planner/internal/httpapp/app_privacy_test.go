package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func seedPrivacyHTTPFixture(t *testing.T, app *App, tenantID string, now time.Time) {
	t.Helper()

	mustExec := func(query string, args ...any) {
		t.Helper()
		if _, err := app.db.ExecContext(context.Background(), query, args...); err != nil {
			t.Fatalf("exec %q failed: %v", query, err)
		}
	}

	createdOld := now.AddDate(0, 0, -120).Format(time.RFC3339)
	oldStart := now.AddDate(0, 0, -46).Format(time.RFC3339)
	oldEnd := now.AddDate(0, 0, -45).Format(time.RFC3339)
	magicOldExpiry := now.AddDate(0, 0, -15).Format(time.RFC3339)
	sessionOldExpiry := now.AddDate(0, 0, -40).Format(time.RFC3339)
	emailOld := now.AddDate(0, 0, -120).Format(time.RFC3339)
	auditOld := now.AddDate(0, 0, -220).Format(time.RFC3339)

	mustExec(
		`INSERT INTO participants (id, tenant_id, email, phone, name, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"prt_http_old",
		tenantID,
		"http-old@example.com",
		"+491234567",
		"HTTP Old",
		createdOld,
		createdOld,
	)

	mustExec(
		`INSERT INTO events (id, tenant_id, slug, title, starts_at, ends_at, timezone, status, is_public, registration_enabled, waitlist_enabled, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, 'Europe/Berlin', 'completed', 1, 1, 1, ?, ?)`,
		"evt_http_old",
		tenantID,
		"http-old-event",
		"HTTP Old Event",
		oldStart,
		oldEnd,
		createdOld,
		createdOld,
	)

	mustExec(
		`INSERT INTO registrations (id, tenant_id, event_id, participant_id, status, participation_type, quantity, source, created_at, updated_at)
     VALUES (?, ?, ?, ?, 'confirmed', 'onsite', 1, 'public_page', ?, ?)`,
		"reg_http_old",
		tenantID,
		"evt_http_old",
		"prt_http_old",
		createdOld,
		createdOld,
	)

	mustExec(
		`INSERT INTO magic_links (id, tenant_id, purpose, token_hash, expires_at, created_at)
     VALUES (?, ?, 'organizer_login', ?, ?, ?)`,
		"mlk_http_old",
		tenantID,
		"token-hash-http-old",
		magicOldExpiry,
		createdOld,
	)

	mustExec(
		`INSERT INTO sessions (id, tenant_id, session_hash, expires_at, created_at)
     VALUES (?, ?, ?, ?, ?)`,
		"ses_http_old",
		tenantID,
		"session-hash-http-old",
		sessionOldExpiry,
		createdOld,
	)

	mustExec(
		`INSERT INTO email_jobs (id, tenant_id, template_key, recipient_email, subject, body_text, status, sent_at, created_at, updated_at)
     VALUES (?, ?, 'registration_confirmed', 'old@example.com', 'old', 'old', 'sent', ?, ?, ?)`,
		"eml_http_old",
		tenantID,
		emailOld,
		emailOld,
		emailOld,
	)

	mustExec(
		`INSERT INTO audit_log (id, tenant_id, action, created_at)
     VALUES (?, ?, 'old_http_audit', ?)`,
		"aud_http_old",
		tenantID,
		auditOld,
	)
}

func TestAdminPrivacyPolicyAndRunFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	seedPrivacyHTTPFixture(t, app, tenantID, time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC))

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/privacy/retention-policies", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected policy list 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	if listPayload["total"] != float64(5) {
		t.Fatalf("expected 5 policies, got %v", listPayload["total"])
	}
	items, ok := listPayload["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected policy items")
	}

	magicPolicyID := ""
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if item["data_category"] == "magic_links" {
			magicPolicyID, _ = item["id"].(string)
			break
		}
	}
	if magicPolicyID == "" {
		t.Fatalf("expected magic_links policy id")
	}

	patchPayload := map[string]any{
		"retention_days": 10,
	}
	patchBody, _ := json.Marshal(patchPayload)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/privacy/retention-policies/"+magicPolicyID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(sessionCookie)
	patchRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected policy patch 200, got %d", patchRec.Code)
	}
	patchResult := decodeBody[map[string]any](t, patchRec)
	patchedItem := patchResult["item"].(map[string]any)
	if patchedItem["retention_days"] != float64(10) {
		t.Fatalf("expected retention_days=10, got %v", patchedItem["retention_days"])
	}

	dryReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/privacy/retention-jobs/dry-run", nil)
	dryReq.AddCookie(sessionCookie)
	dryRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(dryRec, dryReq)
	if dryRec.Code != http.StatusOK {
		t.Fatalf("expected dry-run 200, got %d", dryRec.Code)
	}
	dryPayload := decodeBody[map[string]any](t, dryRec)
	dryItem := dryPayload["item"].(map[string]any)
	if dryItem["dry_run"] != true {
		t.Fatalf("expected dry_run=true, got %v", dryItem["dry_run"])
	}
	if dryItem["total_affected"].(float64) < 1 {
		t.Fatalf("expected dry-run to affect at least one record, got %v", dryItem["total_affected"])
	}

	runReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/privacy/retention-jobs/run", nil)
	runReq.AddCookie(sessionCookie)
	runRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(runRec, runReq)
	if runRec.Code != http.StatusOK {
		t.Fatalf("expected retention run 200, got %d", runRec.Code)
	}
	runPayload := decodeBody[map[string]any](t, runRec)
	runItem := runPayload["item"].(map[string]any)
	if runItem["dry_run"] != false {
		t.Fatalf("expected dry_run=false, got %v", runItem["dry_run"])
	}
	if runItem["total_executed"].(float64) < 1 {
		t.Fatalf("expected retention run to execute at least one change, got %v", runItem["total_executed"])
	}

	var remainingOldMagicLinks int
	if err := app.db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*) FROM magic_links WHERE tenant_id = ? AND id = ?`,
		tenantID,
		"mlk_http_old",
	).Scan(&remainingOldMagicLinks); err != nil {
		t.Fatalf("count old fixture magic links after run: %v", err)
	}
	if remainingOldMagicLinks != 0 {
		t.Fatalf("expected old fixture magic link removed, got %d", remainingOldMagicLinks)
	}

	jobsReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/privacy/retention-jobs", nil)
	jobsReq.AddCookie(sessionCookie)
	jobsRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(jobsRec, jobsReq)
	if jobsRec.Code != http.StatusOK {
		t.Fatalf("expected retention jobs list 200, got %d", jobsRec.Code)
	}
	jobsPayload := decodeBody[map[string]any](t, jobsRec)
	if jobsPayload["total"].(float64) < 2 {
		t.Fatalf("expected at least 2 retention jobs, got %v", jobsPayload["total"])
	}
}

func TestAdminPrivacyRequiresAuth(t *testing.T) {
	app, _, _ := setupAuthApp(t)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/privacy/retention-policies", nil)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized list 401, got %d", listRec.Code)
	}

	runReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/privacy/retention-jobs/run", nil)
	runRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(runRec, runReq)
	if runRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized run 401, got %d", runRec.Code)
	}

	jobsReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/privacy/retention-jobs", nil)
	jobsRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(jobsRec, jobsReq)
	if jobsRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized jobs 401, got %d", jobsRec.Code)
	}
}

func TestAdminPrivacyReadonlyPermissions(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE tenant_users SET role = 'readonly' WHERE email = ?`,
		"owner@example.com",
	); err != nil {
		t.Fatalf("set readonly role: %v", err)
	}
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/privacy/retention-policies", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected readonly list 200, got %d", listRec.Code)
	}

	jobsReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/privacy/retention-jobs", nil)
	jobsReq.AddCookie(sessionCookie)
	jobsRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(jobsRec, jobsReq)
	if jobsRec.Code != http.StatusOK {
		t.Fatalf("expected readonly jobs 200, got %d", jobsRec.Code)
	}

	items := decodeBody[map[string]any](t, listRec)["items"].([]any)
	policyID := items[0].(map[string]any)["id"].(string)

	patchPayload := map[string]any{"retention_days": 11}
	patchBody, _ := json.Marshal(patchPayload)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/privacy/retention-policies/"+policyID, bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(sessionCookie)
	patchRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusForbidden {
		t.Fatalf("expected readonly patch 403, got %d", patchRec.Code)
	}

	dryReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/privacy/retention-jobs/dry-run", nil)
	dryReq.AddCookie(sessionCookie)
	dryRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(dryRec, dryReq)
	if dryRec.Code != http.StatusForbidden {
		t.Fatalf("expected readonly dry-run 403, got %d", dryRec.Code)
	}

	runReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/privacy/retention-jobs/run", nil)
	runReq.AddCookie(sessionCookie)
	runRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(runRec, runReq)
	if runRec.Code != http.StatusForbidden {
		t.Fatalf("expected readonly run 403, got %d", runRec.Code)
	}
}
