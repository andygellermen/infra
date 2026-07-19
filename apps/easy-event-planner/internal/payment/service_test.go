package payment

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/event"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func setupPaymentService(t *testing.T, paypalMode string) (*Service, *sql.DB, tenant.Tenant, string) {
	t.Helper()

	sqlDB, err := db.Open("sqlite", filepath.Join(t.TempDir(), "payment-service.sqlite"))
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
	tenantItem, err := tenantRepo.CreateTenant(context.Background(), tenant.CreateTenantParams{
		Slug:          "pay-tenant",
		Name:          "Pay Tenant",
		PublicBaseURL: "http://localhost:8080/pay-tenant",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	clientID := ""
	settingsJSON := ""
	if paypalMode != "disabled" {
		clientID = "client_test_sandbox"
		settingsJSON = `{"paypal_client_secret":"secret_test_sandbox","paypal_webhook_id":"wh_test_sandbox"}`
	}
	if _, err := tenantRepo.UpsertSettings(context.Background(), tenant.UpsertTenantSettingsParams{
		TenantID: tenantItem.ID,
		Settings: tenant.TenantSettingsInput{
			PayPalMode:       paypalMode,
			PayPalClientID:   clientID,
			PayPalMerchantID: "merchant_test",
			SettingsJSON:     settingsJSON,
		},
	}); err != nil {
		t.Fatalf("upsert tenant settings: %v", err)
	}

	eventRepo := event.NewRepository(sqlDB)
	createdEvent, err := eventRepo.CreateEvent(context.Background(), tenantItem.ID, event.CreateEventParams{
		Slug:             "paid-workshop",
		Title:            "Paid Workshop",
		StartsAt:         time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		TicketName:       "Standard",
		PriceCents:       intPtr(2500),
		Currency:         "EUR",
		DonationEnabled:  boolPtr(true),
		DonationMinCents: intPtr(100),
	})
	if err != nil {
		t.Fatalf("create event: %v", err)
	}
	if _, err := eventRepo.PublishEvent(context.Background(), tenantItem.ID, createdEvent.ID); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	participantID := "par_pay_001"
	registrationID := "reg_pay_001"
	if _, err := sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO participants (id, tenant_id, email, phone, name, created_at, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, ?)`,
		participantID,
		tenantItem.ID,
		"pay-test@example.com",
		"+491701234567",
		"Pay Test",
		now,
		now,
	); err != nil {
		t.Fatalf("insert participant: %v", err)
	}
	if _, err := sqlDB.ExecContext(
		context.Background(),
		`INSERT INTO registrations (
      id, tenant_id, event_id, participant_id, status, participation_type, quantity,
      source, privacy_accepted_at, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?)`,
		registrationID,
		tenantItem.ID,
		createdEvent.ID,
		participantID,
		registrationStatusConfirmed,
		event.ParticipationModeOnsite,
		"public_page",
		now,
		now,
		now,
	); err != nil {
		t.Fatalf("insert registration: %v", err)
	}

	return NewService(sqlDB, Config{}), sqlDB, tenantItem, registrationID
}

func TestCreatePayPalOrderAndWebhookFlow(t *testing.T) {
	service, sqlDB, tenantItem, registrationID := setupPaymentService(t, "sandbox")
	service.SetParticipantCalendarURLBuilder(func(tenantSlug, tenantID, registrationID, participantID string) string {
		return "https://events.example.com/api/v1/public/" + tenantSlug + "/registrations/" + registrationID + "/calendar.ics?token=pay-test"
	})

	createResult, err := service.CreatePayPalOrder(context.Background(), CreatePayPalOrderInput{
		TenantID:            tenantItem.ID,
		RegistrationID:      registrationID,
		AmountCents:         2900,
		Currency:            "eur",
		DonationAmountCents: 400,
	})
	if err != nil {
		t.Fatalf("create paypal order: %v", err)
	}
	if createResult.ProviderOrderID == "" {
		t.Fatalf("expected provider order id")
	}
	if createResult.Status != PaymentStatusCreated {
		t.Fatalf("expected payment status created, got %q", createResult.Status)
	}
	if createResult.RegistrationStatus != registrationStatusPaymentPending {
		t.Fatalf("expected registration status payment_pending, got %q", createResult.RegistrationStatus)
	}
	if createResult.DonationAmountCents == nil || *createResult.DonationAmountCents != 400 {
		t.Fatalf("expected donation amount 400, got %v", createResult.DonationAmountCents)
	}

	var registrationStatus string
	var reservedUntilRaw string
	if err := sqlDB.QueryRowContext(
		context.Background(),
		`SELECT status, COALESCE(reserved_until, '')
     FROM registrations
     WHERE tenant_id = ? AND id = ?`,
		tenantItem.ID,
		registrationID,
	).Scan(&registrationStatus, &reservedUntilRaw); err != nil {
		t.Fatalf("query registration after create order: %v", err)
	}
	if registrationStatus != registrationStatusPaymentPending {
		t.Fatalf("expected registration status payment_pending in db, got %q", registrationStatus)
	}
	if reservedUntilRaw == "" {
		t.Fatalf("expected reserved_until to be set after order creation")
	}

	webhookPayload := map[string]any{
		"id":         "wh_test_001",
		"event_type": "PAYMENT.CAPTURE.COMPLETED",
		"resource": map[string]any{
			"id": "CAPTURE_TEST_001",
			"supplementary_data": map[string]any{
				"related_ids": map[string]any{
					"order_id":   createResult.ProviderOrderID,
					"capture_id": "CAPTURE_TEST_001",
				},
			},
		},
	}
	webhookBody, _ := json.Marshal(webhookPayload)
	webhookResult, err := service.ProcessPayPalWebhook(context.Background(), PayPalWebhookHeaders{}, webhookBody)
	if err != nil {
		t.Fatalf("process webhook: %v", err)
	}
	if !webhookResult.Processed {
		t.Fatalf("expected processed webhook")
	}
	if webhookResult.PaymentStatus != PaymentStatusPaid {
		t.Fatalf("expected payment status paid, got %q", webhookResult.PaymentStatus)
	}

	var paymentStatus string
	var captureID string
	var paidAtRaw string
	if err := sqlDB.QueryRowContext(
		context.Background(),
		`SELECT status, COALESCE(provider_capture_id, ''), COALESCE(paid_at, '')
     FROM payments
     WHERE id = ?`,
		createResult.PaymentID,
	).Scan(&paymentStatus, &captureID, &paidAtRaw); err != nil {
		t.Fatalf("query payment after webhook: %v", err)
	}
	if paymentStatus != PaymentStatusPaid {
		t.Fatalf("expected payment status paid in db, got %q", paymentStatus)
	}
	if captureID != "CAPTURE_TEST_001" {
		t.Fatalf("expected capture id CAPTURE_TEST_001, got %q", captureID)
	}
	if paidAtRaw == "" {
		t.Fatalf("expected paid_at to be set")
	}

	var confirmedAtRaw string
	if err := sqlDB.QueryRowContext(
		context.Background(),
		`SELECT status, COALESCE(reserved_until, ''), COALESCE(confirmed_at, '')
     FROM registrations
     WHERE id = ?`,
		registrationID,
	).Scan(&registrationStatus, &reservedUntilRaw, &confirmedAtRaw); err != nil {
		t.Fatalf("query registration after webhook: %v", err)
	}
	if registrationStatus != registrationStatusConfirmed {
		t.Fatalf("expected registration status confirmed after paid webhook, got %q", registrationStatus)
	}
	if reservedUntilRaw != "" {
		t.Fatalf("expected reserved_until to be cleared after paid webhook")
	}
	if confirmedAtRaw == "" {
		t.Fatalf("expected confirmed_at to be set after paid webhook")
	}
	var emailTemplate string
	var emailBody string
	var metadataJSON string
	if err := sqlDB.QueryRowContext(
		context.Background(),
		`SELECT template_key, body_text, COALESCE(metadata_json, '')
     FROM email_jobs
     WHERE tenant_id = ?
     ORDER BY created_at DESC
     LIMIT 1`,
		tenantItem.ID,
	).Scan(&emailTemplate, &emailBody, &metadataJSON); err != nil {
		t.Fatalf("query latest email job after paid webhook: %v", err)
	}
	if emailTemplate != "registration_confirmed" {
		t.Fatalf("expected registration_confirmed email template, got %q", emailTemplate)
	}
	if !strings.Contains(emailBody, "wieder ab") {
		t.Fatalf("expected payment confirmation mail to contain cancel hint, got %q", emailBody)
	}
	if !strings.Contains(emailBody, "/calendar.ics?token=pay-test") {
		t.Fatalf("expected payment confirmation mail to contain calendar URL, got %q", emailBody)
	}
	if !strings.Contains(metadataJSON, "\"participant_cancel_deadline_hours\"") {
		t.Fatalf("expected payment confirmation metadata to contain cancel deadline, got %q", metadataJSON)
	}

	duplicateResult, err := service.ProcessPayPalWebhook(context.Background(), PayPalWebhookHeaders{}, webhookBody)
	if err != nil {
		t.Fatalf("process duplicate webhook: %v", err)
	}
	if !duplicateResult.Duplicate {
		t.Fatalf("expected duplicate webhook marker")
	}
}

func TestCreatePayPalOrderDisabledMode(t *testing.T) {
	service, _, tenantItem, registrationID := setupPaymentService(t, "disabled")

	_, err := service.CreatePayPalOrder(context.Background(), CreatePayPalOrderInput{
		TenantID:       tenantItem.ID,
		RegistrationID: registrationID,
		AmountCents:    1000,
		Currency:       "EUR",
	})
	if !errors.Is(err, ErrPayPalDisabled) {
		t.Fatalf("expected ErrPayPalDisabled, got %v", err)
	}
}

func TestCreatePayPalOrderRealAPI(t *testing.T) {
	service, _, tenantItem, registrationID := setupPaymentService(t, "sandbox")

	tokenCalls := 0
	orderCalls := 0
	service.httpClient = &http.Client{
		Transport: roundTripFn(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/v1/oauth2/token":
				tokenCalls++
				if req.Method != http.MethodPost {
					t.Fatalf("expected POST /v1/oauth2/token, got %s", req.Method)
				}
				if got := strings.TrimSpace(req.Header.Get("Authorization")); got == "" || !strings.HasPrefix(got, "Basic ") {
					t.Fatalf("expected basic auth header for token endpoint, got %q", got)
				}
				expectedBasic := "Basic " + base64.StdEncoding.EncodeToString([]byte("client_test_sandbox:secret_test_sandbox"))
				if strings.TrimSpace(req.Header.Get("Authorization")) != expectedBasic {
					t.Fatalf("expected basic auth %q, got %q", expectedBasic, strings.TrimSpace(req.Header.Get("Authorization")))
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"access_token":"token-123","token_type":"Bearer"}`)),
					Request:    req,
				}, nil
			case "/v2/checkout/orders":
				orderCalls++
				if req.Method != http.MethodPost {
					t.Fatalf("expected POST /v2/checkout/orders, got %s", req.Method)
				}
				if got := strings.TrimSpace(req.Header.Get("Authorization")); got != "Bearer token-123" {
					t.Fatalf("expected bearer token header, got %q", got)
				}
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("read order request body: %v", err)
				}
				var payload map[string]any
				if err := json.Unmarshal(bodyBytes, &payload); err != nil {
					t.Fatalf("decode order request body: %v", err)
				}
				expectedAmount := "25.00"
				amount := deepString(payload, "purchase_units", "0", "amount", "value")
				if amount != expectedAmount {
					t.Fatalf("expected amount %q, got %q", expectedAmount, amount)
				}
				return &http.Response{
					StatusCode: http.StatusCreated,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"id":"O-REAL-123","status":"CREATED","links":[{"rel":"approve","href":"https://paypal.example/approve/O-REAL-123"}]}`)),
					Request:    req,
				}, nil
			default:
				t.Fatalf("unexpected path called: %s", req.URL.Path)
				return nil, nil
			}
		}),
	}

	service.cfg.UseRealPayPalAPI = true
	service.cfg.SandboxAPIBaseURL = "https://api-m.sandbox.paypal.test"
	service.cfg.LiveAPIBaseURL = "https://api-m.paypal.test"
	service.orderFn = service.defaultCreatePayPalOrder

	result, err := service.CreatePayPalOrder(context.Background(), CreatePayPalOrderInput{
		TenantID:       tenantItem.ID,
		RegistrationID: registrationID,
		AmountCents:    2500,
		Currency:       "EUR",
	})
	if err != nil {
		t.Fatalf("create paypal order in real mode: %v", err)
	}
	if result.ProviderOrderID != "O-REAL-123" {
		t.Fatalf("expected provider order id O-REAL-123, got %q", result.ProviderOrderID)
	}
	if result.ApproveURL != "https://paypal.example/approve/O-REAL-123" {
		t.Fatalf("expected approve url from provider response, got %q", result.ApproveURL)
	}
	if tokenCalls != 1 {
		t.Fatalf("expected 1 token call, got %d", tokenCalls)
	}
	if orderCalls != 1 {
		t.Fatalf("expected 1 order call, got %d", orderCalls)
	}
}

func TestProcessPayPalWebhookRejectsInvalidSignatureInRealMode(t *testing.T) {
	service, _, tenantItem, registrationID := setupPaymentService(t, "sandbox")

	createResult, err := service.CreatePayPalOrder(context.Background(), CreatePayPalOrderInput{
		TenantID:       tenantItem.ID,
		RegistrationID: registrationID,
		AmountCents:    2500,
		Currency:       "EUR",
	})
	if err != nil {
		t.Fatalf("create paypal order: %v", err)
	}

	service.cfg.UseRealPayPalAPI = true
	service.verifyFn = func(context.Context, verifyWebhookProviderInput) (verifyWebhookProviderResult, error) {
		return verifyWebhookProviderResult{VerificationStatus: "FAILURE"}, nil
	}

	webhookPayload := map[string]any{
		"id":         "wh_invalid_sig_001",
		"event_type": "PAYMENT.CAPTURE.COMPLETED",
		"resource": map[string]any{
			"id": "CAPTURE_SIG_FAIL_001",
			"supplementary_data": map[string]any{
				"related_ids": map[string]any{
					"order_id":   createResult.ProviderOrderID,
					"capture_id": "CAPTURE_SIG_FAIL_001",
				},
			},
		},
	}
	webhookBody, _ := json.Marshal(webhookPayload)
	_, err = service.ProcessPayPalWebhook(context.Background(), PayPalWebhookHeaders{
		TransmissionID:   "tid-1",
		TransmissionTime: "2026-05-15T10:00:00Z",
		TransmissionSig:  "sig-1",
		CertURL:          "https://example.test/cert.pem",
		AuthAlgo:         "SHA256withRSA",
	}, webhookBody)
	if !errors.Is(err, ErrWebhookSignatureInvalid) {
		t.Fatalf("expected ErrWebhookSignatureInvalid, got %v", err)
	}

	var status, errorMessage string
	if err := service.db.QueryRowContext(
		context.Background(),
		`SELECT processing_status, COALESCE(error_message, '')
     FROM paypal_webhook_events
     WHERE paypal_event_id = ?`,
		"wh_invalid_sig_001",
	).Scan(&status, &errorMessage); err != nil {
		t.Fatalf("query webhook event status: %v", err)
	}
	if status != WebhookStatusFailed {
		t.Fatalf("expected webhook status failed, got %q", status)
	}
	if errorMessage != "invalid_signature" {
		t.Fatalf("expected webhook error invalid_signature, got %q", errorMessage)
	}
}

type roundTripFn func(req *http.Request) (*http.Response, error)

func (fn roundTripFn) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func deepString(payload map[string]any, path ...string) string {
	var current any = payload
	for _, segment := range path {
		switch typed := current.(type) {
		case map[string]any:
			current = typed[segment]
		case []any:
			index := 0
			if _, err := fmt.Sscanf(segment, "%d", &index); err != nil {
				return ""
			}
			if index < 0 || index >= len(typed) {
				return ""
			}
			current = typed[index]
		default:
			return ""
		}
	}
	text, _ := current.(string)
	return strings.TrimSpace(text)
}

func intPtr(value int) *int {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
