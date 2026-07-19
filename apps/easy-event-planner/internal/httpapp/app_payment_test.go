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
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func createRegistrationForPaymentHTTPTest(t *testing.T, app *App, tenantSlug, eventID, email, expectedStatus string) string {
	t.Helper()
	tenantID := tenantIDBySlug(t, app, tenantSlug)

	startPayload := map[string]any{
		"event_id":           eventID,
		"name":               "Paying Person",
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
		t.Fatalf("expected registration start status 202, got %d", startRec.Code)
	}
	startResult := decodeBody[map[string]any](t, startRec)
	registrationID, _ := startResult["registration_id"].(string)
	if registrationID == "" {
		t.Fatalf("expected registration id")
	}

	verifyToken := extractVerifyTokenFromJobInHTTPTest(t, app, tenantID, registrationID)
	verifyPayload := map[string]any{"token": verifyToken}
	verifyBody, _ := json.Marshal(verifyPayload)
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/registrations/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("expected registration verify status 200, got %d", verifyRec.Code)
	}
	verifyResult := decodeBody[map[string]any](t, verifyRec)
	if verifyResult["status"] != expectedStatus {
		t.Fatalf("expected %s registration, got %v", expectedStatus, verifyResult["status"])
	}
	return registrationID
}

func createConfirmedRegistrationForPaymentHTTPTest(t *testing.T, app *App, tenantSlug, eventID, email string) string {
	t.Helper()
	return createRegistrationForPaymentHTTPTest(t, app, tenantSlug, eventID, email, "confirmed")
}

func createReservedRegistrationForPaymentHTTPTest(t *testing.T, app *App, tenantSlug, eventID, email string) string {
	t.Helper()
	return createRegistrationForPaymentHTTPTest(t, app, tenantSlug, eventID, email, "reserved")
}

func TestPublicPayPalCreateOrderAndWebhookFlow(t *testing.T) {
	app, sender, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	sessionCookie := loginSessionCookie(t, app, sender, tenantSlug, "owner@example.com")

	if _, err := app.tenantRepo.UpsertSettings(context.Background(), tenant.UpsertTenantSettingsParams{
		TenantID: tenantID,
		Settings: tenant.TenantSettingsInput{
			PayPalMode:       "sandbox",
			PayPalClientID:   "sandbox-client-id",
			PayPalMerchantID: "sandbox-merchant-id",
			SettingsJSON:     `{"paypal_client_secret":"sandbox-secret","paypal_webhook_id":"wh_sandbox_http"}`,
		},
	}); err != nil {
		t.Fatalf("upsert tenant paypal settings: %v", err)
	}

	published := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:             "paypal-public-event",
		Title:            "PayPal Public Event",
		StartsAt:         time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		TicketName:       "Standard",
		PriceCents:       intPtr(4600),
		Currency:         "EUR",
		DonationEnabled:  boolPtr(true),
		DonationMinCents: intPtr(200),
	})
	registrationID := createReservedRegistrationForPaymentHTTPTest(
		t,
		app,
		tenantSlug,
		published.ID,
		"payment-user@example.com",
	)

	createOrderPayload := map[string]any{
		"registration_id":       registrationID,
		"amount_cents":          4900,
		"currency":              "EUR",
		"donation_amount_cents": 300,
	}
	createOrderBody, _ := json.Marshal(createOrderPayload)
	createOrderReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/payments/paypal/create-order", bytes.NewReader(createOrderBody))
	createOrderReq.Header.Set("Content-Type", "application/json")
	createOrderRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createOrderRec, createOrderReq)
	if createOrderRec.Code != http.StatusCreated {
		t.Fatalf("expected create-order status 201, got %d", createOrderRec.Code)
	}
	createOrderResult := decodeBody[map[string]any](t, createOrderRec)
	paymentPayload, ok := createOrderResult["payment"].(map[string]any)
	if !ok {
		t.Fatalf("expected payment payload")
	}
	providerOrderID, _ := paymentPayload["provider_order_id"].(string)
	paymentID, _ := paymentPayload["id"].(string)
	if paymentID == "" || providerOrderID == "" {
		t.Fatalf("expected payment id and provider order id in response")
	}
	if paymentPayload["status"] != "created" {
		t.Fatalf("expected payment status created, got %v", paymentPayload["status"])
	}
	registrationPayload := createOrderResult["registration"].(map[string]any)
	if registrationPayload["status"] != "payment_pending" {
		t.Fatalf("expected registration status payment_pending, got %v", registrationPayload["status"])
	}

	webhookPayload := map[string]any{
		"id":         "wh_http_001",
		"event_type": "PAYMENT.CAPTURE.COMPLETED",
		"resource": map[string]any{
			"id": "CAPTURE_HTTP_001",
			"supplementary_data": map[string]any{
				"related_ids": map[string]any{
					"order_id":   providerOrderID,
					"capture_id": "CAPTURE_HTTP_001",
				},
			},
		},
	}
	webhookBody, _ := json.Marshal(webhookPayload)
	webhookReq := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/paypal", bytes.NewReader(webhookBody))
	webhookReq.Header.Set("Content-Type", "application/json")
	webhookRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(webhookRec, webhookReq)
	if webhookRec.Code != http.StatusOK {
		t.Fatalf("expected webhook status 200, got %d", webhookRec.Code)
	}
	webhookResult := decodeBody[map[string]any](t, webhookRec)
	if webhookResult["processed"] != true {
		t.Fatalf("expected processed=true, got %v", webhookResult["processed"])
	}
	if webhookResult["payment_id"] != paymentID {
		t.Fatalf("expected webhook payment id %q, got %v", paymentID, webhookResult["payment_id"])
	}
	if webhookResult["payment_status"] != "paid" {
		t.Fatalf("expected webhook payment_status paid, got %v", webhookResult["payment_status"])
	}

	duplicateReq := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/paypal", bytes.NewReader(webhookBody))
	duplicateReq.Header.Set("Content-Type", "application/json")
	duplicateRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(duplicateRec, duplicateReq)
	if duplicateRec.Code != http.StatusOK {
		t.Fatalf("expected duplicate webhook status 200, got %d", duplicateRec.Code)
	}
	duplicateResult := decodeBody[map[string]any](t, duplicateRec)
	if duplicateResult["duplicate"] != true {
		t.Fatalf("expected duplicate=true, got %v", duplicateResult["duplicate"])
	}

	adminGetReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/registrations/"+registrationID, nil)
	adminGetReq.AddCookie(sessionCookie)
	adminGetRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(adminGetRec, adminGetReq)
	if adminGetRec.Code != http.StatusOK {
		t.Fatalf("expected admin registration get status 200, got %d", adminGetRec.Code)
	}
	adminGetPayload := decodeBody[map[string]any](t, adminGetRec)
	item := adminGetPayload["item"].(map[string]any)
	if item["status"] != "confirmed" {
		t.Fatalf("expected registration confirmed after paid webhook, got %v", item["status"])
	}
	if item["payment_status"] != "paid" {
		t.Fatalf("expected payment_status paid after webhook, got %v", item["payment_status"])
	}
}

func TestPublicPayPalCreateOrderDisabled(t *testing.T) {
	app, _, tenantSlug := setupAuthApp(t)
	tenantID := tenantIDBySlug(t, app, tenantSlug)
	published := createPublishedEventForRegistrationHTTP(t, app, tenantID, event.CreateEventParams{
		Slug:       "paypal-disabled-event",
		Title:      "PayPal Disabled Event",
		StartsAt:   time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		TicketName: "Standard",
		PriceCents: intPtr(1990),
		Currency:   "EUR",
	})
	registrationID := createReservedRegistrationForPaymentHTTPTest(
		t,
		app,
		tenantSlug,
		published.ID,
		"payment-disabled@example.com",
	)

	createOrderPayload := map[string]any{
		"registration_id": registrationID,
		"amount_cents":    1990,
		"currency":        "EUR",
	}
	createOrderBody, _ := json.Marshal(createOrderPayload)
	createOrderReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/"+tenantSlug+"/payments/paypal/create-order", bytes.NewReader(createOrderBody))
	createOrderReq.Header.Set("Content-Type", "application/json")
	createOrderRec := httptest.NewRecorder()
	app.Handler().ServeHTTP(createOrderRec, createOrderReq)
	if createOrderRec.Code != http.StatusConflict {
		t.Fatalf("expected create-order conflict status 409, got %d", createOrderRec.Code)
	}
	errorPayload := decodeBody[map[string]any](t, createOrderRec)["error"].(map[string]any)
	if errorPayload["code"] != "PAYMENT_REQUIRED" {
		t.Fatalf("expected PAYMENT_REQUIRED, got %v", errorPayload["code"])
	}
}

func intPtr(value int) *int {
	return &value
}
