package httpapp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
)

func TestAdminInvitationCreateAndPublicResolve(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "invite-resolve-event",
		Title:    "Invite Resolve Event",
		StartsAt: time.Now().UTC().Add(72 * time.Hour).Format(time.RFC3339),
	})

	createPayload := map[string]any{
		"event_id":       eventItem.ID,
		"code":           "friends25",
		"invite_type":    "discount_percent",
		"discount_value": 25,
		"max_uses":       10,
		"status":         "active",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invitations", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create invitation status 201, got %d", createRec.Code)
	}
	created := decodeBody[map[string]any](t, createRec)
	item := created["item"].(map[string]any)
	if item["code"] != "FRIENDS25" {
		t.Fatalf("expected normalized code FRIENDS25, got %v", item["code"])
	}

	resolvePayload := map[string]any{
		"event_id":          eventItem.ID,
		"invite_code":       "friends25",
		"participant_email": "guest@example.com",
		"base_amount_cents": 8000,
	}
	resolveBody, _ := json.Marshal(resolvePayload)
	resolveReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/invitations/resolve", bytes.NewReader(resolveBody))
	resolveReq.Header.Set("Content-Type", "application/json")
	resolveRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(resolveRec, resolveReq)
	if resolveRec.Code != http.StatusOK {
		t.Fatalf("expected resolve status 200, got %d", resolveRec.Code)
	}
	resolved := decodeBody[map[string]any](t, resolveRec)
	invite := resolved["invite"].(map[string]any)
	if invite["discount_amount_cents"] != float64(2000) {
		t.Fatalf("expected discount amount 2000, got %v", invite["discount_amount_cents"])
	}
	if invite["final_amount_cents"] != float64(6000) {
		t.Fatalf("expected final amount 6000, got %v", invite["final_amount_cents"])
	}
}

func TestPublicRegistrationStartAppliesInvitationCode(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:     "invite-start-event",
		Title:    "Invite Start Event",
		StartsAt: time.Now().UTC().Add(72 * time.Hour).Format(time.RFC3339),
	})

	createPayload := map[string]any{
		"event_id":           eventItem.ID,
		"code":               "voucher1000",
		"invite_type":        "voucher_fixed",
		"discount_type":      "fixed",
		"discount_value":     1000,
		"max_uses":           5,
		"max_uses_per_email": 1,
		"is_shareable":       false,
		"status":             "active",
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invitations", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected create invitation status 201, got %d", createRec.Code)
	}

	startPayload := map[string]any{
		"event_id":            eventItem.ID,
		"name":                "Invite Guest",
		"email":               "invite-guest@example.com",
		"participation_type":  "onsite",
		"invite_code":         "voucher1000",
		"invite_amount_cents": 3500,
		"privacy_accepted":    true,
	}
	startBody, _ := json.Marshal(startPayload)
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/start", bytes.NewReader(startBody))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusAccepted {
		t.Fatalf("expected start status 202, got %d", startRec.Code)
	}
	startResult := decodeBody[map[string]any](t, startRec)
	invite := startResult["invite"].(map[string]any)
	if invite["code"] != "VOUCHER1000" {
		t.Fatalf("expected invite code VOUCHER1000 in response, got %v", invite["code"])
	}
	if invite["discount_amount_cents"] != float64(1000) {
		t.Fatalf("expected discount amount 1000, got %v", invite["discount_amount_cents"])
	}
	if invite["final_amount_cents"] != float64(2500) {
		t.Fatalf("expected final amount 2500, got %v", invite["final_amount_cents"])
	}

	registrationID, _ := startResult["registration_id"].(string)
	if registrationID == "" {
		t.Fatalf("expected registration id")
	}

	var inviteID string
	if err := app.db.QueryRowContext(
		context.Background(),
		`SELECT COALESCE(invite_id, '') FROM registrations WHERE tenant_id = ? AND id = ?`,
		tenantID,
		registrationID,
	).Scan(&inviteID); err != nil {
		t.Fatalf("query registration invite_id: %v", err)
	}
	if inviteID == "" {
		t.Fatalf("expected invite_id to be stored on registration")
	}
}
