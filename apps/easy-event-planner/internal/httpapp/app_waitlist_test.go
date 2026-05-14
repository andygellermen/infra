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

func createWaitlistScenario(t *testing.T, app *App, tenantSlug string) (tenantID, eventID, waitlistID, firstRegistrationID string) {
	t.Helper()

	tenantID = tenantIDBySlug(t, app, tenantSlug)
	waitlistEnabled := true
	maxParticipants := 1
	eventItem := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:            "admin-waitlist-scenario",
		Title:           "Admin Waitlist Scenario",
		StartsAt:        time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		WaitlistEnabled: &waitlistEnabled,
		MaxParticipants: &maxParticipants,
	})

	registerAndVerify := func(name, email string) map[string]any {
		t.Helper()
		startPayload := map[string]any{
			"event_id":           eventItem.ID,
			"name":               name,
			"email":              email,
			"participation_type": "onsite",
			"privacy_accepted":   true,
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
		registrationID, _ := startResult["registration_id"].(string)

		token := extractVerifyTokenFromJobInHTTPTest(t, app, tenantID, registrationID)
		verifyPayload := map[string]any{"token": token}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/verify", bytes.NewReader(verifyBody))
		verifyReq.Header.Set("Content-Type", "application/json")
		verifyRec := httptest.NewRecorder()
		app.Handler().ServeHTTP(verifyRec, verifyReq)
		if verifyRec.Code != http.StatusOK {
			t.Fatalf("expected verify status 200, got %d", verifyRec.Code)
		}
		result := decodeBody[map[string]any](t, verifyRec)
		result["registration_id"] = registrationID
		return result
	}

	first := registerAndVerify("Alice", "alice@example.com")
	if first["status"] != "confirmed" {
		t.Fatalf("expected first status confirmed, got %v", first["status"])
	}
	firstRegistrationID, _ = first["registration_id"].(string)

	second := registerAndVerify("Bob", "bob@example.com")
	if second["status"] != "waitlist" {
		t.Fatalf("expected second status waitlist, got %v", second["status"])
	}
	waitlist, ok := second["waitlist"].(map[string]any)
	if !ok {
		t.Fatalf("expected waitlist payload on second registration")
	}
	waitlistID, _ = waitlist["id"].(string)
	if waitlistID == "" {
		t.Fatalf("expected waitlist id")
	}
	return tenantID, eventItem.ID, waitlistID, firstRegistrationID
}

func TestAdminWaitlistListOfferPromoteFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")
	_, eventID, waitlistID, firstRegistrationID := createWaitlistScenario(t, app, tenantSlug)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID+"/waitlist", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected waitlist list status 200, got %d", listRec.Code)
	}
	listPayload := decodeBody[map[string]any](t, listRec)
	if listPayload["total"] != float64(1) {
		t.Fatalf("expected waitlist total=1, got %v", listPayload["total"])
	}

	offerReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/waitlist/"+waitlistID+"/offer", nil)
	offerReq.AddCookie(sessionCookie)
	offerRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(offerRec, offerReq)
	if offerRec.Code != http.StatusOK {
		t.Fatalf("expected waitlist offer status 200, got %d", offerRec.Code)
	}
	offerPayload := decodeBody[map[string]any](t, offerRec)
	offeredItem := offerPayload["item"].(map[string]any)
	if offeredItem["status"] != "offered" {
		t.Fatalf("expected offered status, got %v", offeredItem["status"])
	}

	promoteReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/waitlist/"+waitlistID+"/promote", nil)
	promoteReq.AddCookie(sessionCookie)
	promoteRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(promoteRec, promoteReq)
	if promoteRec.Code != http.StatusConflict {
		t.Fatalf("expected promote conflict status 409 while full, got %d", promoteRec.Code)
	}
	promoteError := decodeBody[map[string]any](t, promoteRec)["error"].(map[string]any)
	if promoteError["code"] != "EVENT_FULL" {
		t.Fatalf("expected EVENT_FULL, got %v", promoteError["code"])
	}

	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE registrations SET status = ?, cancelled_at = ?, updated_at = ? WHERE id = ?`,
		"cancelled",
		time.Now().UTC().Format(time.RFC3339),
		time.Now().UTC().Format(time.RFC3339),
		firstRegistrationID,
	); err != nil {
		t.Fatalf("cancel first registration: %v", err)
	}

	promoteAfterFreeReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/waitlist/"+waitlistID+"/promote", nil)
	promoteAfterFreeReq.AddCookie(sessionCookie)
	promoteAfterFreeRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(promoteAfterFreeRec, promoteAfterFreeReq)
	if promoteAfterFreeRec.Code != http.StatusOK {
		t.Fatalf("expected promote status 200 after freeing capacity, got %d", promoteAfterFreeRec.Code)
	}
	promotePayload := decodeBody[map[string]any](t, promoteAfterFreeRec)
	promotedItem := promotePayload["item"].(map[string]any)
	if promotedItem["status"] != "promoted" {
		t.Fatalf("expected promoted waitlist status, got %v", promotedItem["status"])
	}
	if promotedItem["registration_status"] != "confirmed" {
		t.Fatalf("expected registration_status confirmed, got %v", promotedItem["registration_status"])
	}
}

func TestAdminWaitlistRequiresAuth(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	_, eventID, _, _ := createWaitlistScenario(t, app, tenantSlug)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID+"/waitlist", nil)
	rec := httptest.NewRecorder()
	app.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized status 401, got %d", rec.Code)
	}
}

func TestAdminWaitlistReadonlyCannotWrite(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	if _, err := app.db.ExecContext(
		context.Background(),
		`UPDATE tenant_users SET role = 'readonly' WHERE email = ?`,
		"owner@example.com",
	); err != nil {
		t.Fatalf("set readonly role: %v", err)
	}
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")
	_, eventID, waitlistID, _ := createWaitlistScenario(t, app, tenantSlug)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/events/"+eventID+"/waitlist", nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected readonly waitlist list status 200, got %d", listRec.Code)
	}

	offerReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/waitlist/"+waitlistID+"/offer", nil)
	offerReq.AddCookie(sessionCookie)
	offerRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(offerRec, offerReq)
	if offerRec.Code != http.StatusForbidden {
		t.Fatalf("expected readonly offer status 403, got %d", offerRec.Code)
	}
}
