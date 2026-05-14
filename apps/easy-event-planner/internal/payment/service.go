package payment

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"
)

const (
	ProviderPayPal = "paypal"

	PaymentStatusCreated        = "created"
	PaymentStatusPaymentPending = "payment_pending"
	PaymentStatusPaid           = "paid"
	PaymentStatusFailed         = "failed"
	PaymentStatusCancelled      = "cancelled"
	PaymentStatusRefunded       = "refunded"

	WebhookStatusReceived  = "received"
	WebhookStatusProcessed = "processed"
	WebhookStatusIgnored   = "ignored"
	WebhookStatusFailed    = "failed"

	registrationStatusReserved       = "reserved"
	registrationStatusPaymentPending = "payment_pending"
	registrationStatusConfirmed      = "confirmed"
	registrationStatusExpired        = "expired"

	defaultReservationTTL = 15 * time.Minute
	defaultCurrency       = "EUR"
)

var (
	ErrInvalidCreateOrderInput  = errors.New("invalid create order input")
	ErrPayPalDisabled           = errors.New("paypal is disabled for tenant")
	ErrPayPalConfigInvalid      = errors.New("paypal configuration is invalid")
	ErrPayPalCredentialsMissing = errors.New("paypal credentials are missing")
	ErrRegistrationNotFound     = errors.New("registration not found")
	ErrRegistrationStateInvalid = errors.New("registration state does not allow payment")
	ErrWebhookPayloadInvalid    = errors.New("paypal webhook payload is invalid")
	ErrWebhookSignatureInvalid  = errors.New("paypal webhook signature is invalid")
)

type Config struct {
	ReservationTTL       time.Duration
	FallbackClientSecret string
	FallbackWebhookID    string
	FallbackClientID     string
	SandboxAPIBaseURL    string
	LiveAPIBaseURL       string
	HTTPTimeout          time.Duration
	UseRealPayPalAPI     bool
}

type Service struct {
	db         *sql.DB
	cfg        Config
	nowFn      func() time.Time
	idFn       func(prefix string) string
	orderFn    func(ctx context.Context, input createOrderProviderInput) (createOrderProviderResult, error)
	verifyFn   func(ctx context.Context, input verifyWebhookProviderInput) (verifyWebhookProviderResult, error)
	httpClient *http.Client
}

type CreatePayPalOrderInput struct {
	TenantID            string
	RegistrationID      string
	AmountCents         int
	Currency            string
	DonationAmountCents int
}

type CreatePayPalOrderResult struct {
	PaymentID            string
	TenantID             string
	RegistrationID       string
	Provider             string
	ProviderOrderID      string
	Status               string
	AmountCents          int
	Currency             string
	DonationAmountCents  *int
	PayPalMode           string
	PayPalClientID       string
	ApproveURL           string
	RegistrationStatus   string
	RegistrationReserved time.Time
}

type ProcessPayPalWebhookResult struct {
	EventID        string
	EventType      string
	PaymentID      string
	RegistrationID string
	PaymentStatus  string
	Processed      bool
	Ignored        bool
	Duplicate      bool
}

type PayPalWebhookHeaders struct {
	TransmissionID   string
	TransmissionTime string
	TransmissionSig  string
	CertURL          string
	AuthAlgo         string
}

type tenantPayPalSettings struct {
	Mode         string
	ClientID     string
	ClientSecret string
	MerchantID   string
	WebhookID    string
}

type registrationForPayment struct {
	ID            string
	TenantID      string
	Status        string
	ReservedUntil string
}

type paymentLookup struct {
	ID              string
	TenantID        string
	RegistrationID  string
	ProviderOrderID string
	Status          string
}

type createOrderProviderInput struct {
	Mode           string
	ClientID       string
	ClientSecret   string
	OrderRef       string
	Currency       string
	AmountCents    int
	RegistrationID string
}

type createOrderProviderResult struct {
	OrderID     string
	Status      string
	ApproveURL  string
	RawResponse string
}

type verifyWebhookProviderInput struct {
	Mode      string
	ClientID  string
	Secret    string
	WebhookID string
	Headers   PayPalWebhookHeaders
	Payload   any
}

type verifyWebhookProviderResult struct {
	VerificationStatus string `json:"verification_status"`
}

type payPalWebhookPayload struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	Resource  struct {
		ID                string `json:"id"`
		CustomID          string `json:"custom_id"`
		SupplementaryData struct {
			RelatedIDs struct {
				OrderID   string `json:"order_id"`
				CaptureID string `json:"capture_id"`
			} `json:"related_ids"`
		} `json:"supplementary_data"`
	} `json:"resource"`
}

func NewService(sqlDB *sql.DB, cfg Config) *Service {
	ttl := cfg.ReservationTTL
	if ttl <= 0 {
		ttl = defaultReservationTTL
	}
	sandboxBase := strings.TrimSpace(cfg.SandboxAPIBaseURL)
	if sandboxBase == "" {
		sandboxBase = "https://api-m.sandbox.paypal.com"
	}
	liveBase := strings.TrimSpace(cfg.LiveAPIBaseURL)
	if liveBase == "" {
		liveBase = "https://api-m.paypal.com"
	}
	httpTimeout := cfg.HTTPTimeout
	if httpTimeout <= 0 {
		httpTimeout = 15 * time.Second
	}
	client := &http.Client{Timeout: httpTimeout}

	service := &Service{
		db: sqlDB,
		cfg: Config{
			ReservationTTL:       ttl,
			FallbackClientSecret: strings.TrimSpace(cfg.FallbackClientSecret),
			FallbackWebhookID:    strings.TrimSpace(cfg.FallbackWebhookID),
			FallbackClientID:     strings.TrimSpace(cfg.FallbackClientID),
			SandboxAPIBaseURL:    strings.TrimRight(sandboxBase, "/"),
			LiveAPIBaseURL:       strings.TrimRight(liveBase, "/"),
			HTTPTimeout:          httpTimeout,
			UseRealPayPalAPI:     cfg.UseRealPayPalAPI,
		},
		nowFn:      func() time.Time { return time.Now().UTC() },
		idFn:       defaultID,
		httpClient: client,
	}
	service.orderFn = service.defaultCreatePayPalOrder
	service.verifyFn = service.defaultVerifyPayPalWebhookSignature
	return service
}

func (s *Service) CreatePayPalOrder(ctx context.Context, input CreatePayPalOrderInput) (CreatePayPalOrderResult, error) {
	if s.db == nil {
		return CreatePayPalOrderResult{}, fmt.Errorf("payment service database is nil")
	}

	tenantID := strings.TrimSpace(input.TenantID)
	registrationID := strings.TrimSpace(input.RegistrationID)
	if tenantID == "" {
		return CreatePayPalOrderResult{}, fmt.Errorf("%w: tenant id must not be empty", ErrInvalidCreateOrderInput)
	}
	if registrationID == "" {
		return CreatePayPalOrderResult{}, fmt.Errorf("%w: registration id must not be empty", ErrInvalidCreateOrderInput)
	}
	if input.AmountCents <= 0 {
		return CreatePayPalOrderResult{}, fmt.Errorf("%w: amount_cents must be > 0", ErrInvalidCreateOrderInput)
	}
	if input.DonationAmountCents < 0 || input.DonationAmountCents > input.AmountCents {
		return CreatePayPalOrderResult{}, fmt.Errorf("%w: donation_amount_cents must be between 0 and amount_cents", ErrInvalidCreateOrderInput)
	}

	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = defaultCurrency
	}
	if len(currency) != 3 {
		return CreatePayPalOrderResult{}, fmt.Errorf("%w: currency must be a 3-letter code", ErrInvalidCreateOrderInput)
	}

	settings, err := s.lookupTenantPayPalSettings(ctx, tenantID)
	if err != nil {
		return CreatePayPalOrderResult{}, err
	}
	if settings.Mode == "disabled" {
		return CreatePayPalOrderResult{}, ErrPayPalDisabled
	}
	clientID := strings.TrimSpace(settings.ClientID)
	if clientID == "" {
		clientID = strings.TrimSpace(s.cfg.FallbackClientID)
	}
	if clientID == "" {
		return CreatePayPalOrderResult{}, ErrPayPalConfigInvalid
	}
	clientSecret := strings.TrimSpace(settings.ClientSecret)
	if clientSecret == "" {
		clientSecret = strings.TrimSpace(s.cfg.FallbackClientSecret)
	}
	if clientSecret == "" {
		return CreatePayPalOrderResult{}, ErrPayPalCredentialsMissing
	}

	now := s.nowFn().UTC()
	reservedUntil := now.Add(s.cfg.ReservationTTL).UTC()
	paymentID := s.idFn("pay")
	orderRef := s.idFn("ppord")

	providerResult, err := s.orderFn(ctx, createOrderProviderInput{
		Mode:           settings.Mode,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		OrderRef:       orderRef,
		Currency:       currency,
		AmountCents:    input.AmountCents,
		RegistrationID: registrationID,
	})
	if err != nil {
		return CreatePayPalOrderResult{}, fmt.Errorf("create paypal order with provider: %w", err)
	}
	if strings.TrimSpace(providerResult.OrderID) == "" {
		return CreatePayPalOrderResult{}, fmt.Errorf("provider returned empty order id")
	}
	if strings.TrimSpace(providerResult.Status) == "" {
		providerResult.Status = strings.ToUpper(PaymentStatusCreated)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CreatePayPalOrderResult{}, fmt.Errorf("begin paypal order transaction: %w", err)
	}

	regItem, err := s.lookupRegistrationForPaymentTx(ctx, tx, tenantID, registrationID)
	if err != nil {
		_ = tx.Rollback()
		return CreatePayPalOrderResult{}, err
	}
	if !canCreateOrderForRegistrationStatus(regItem.Status) {
		_ = tx.Rollback()
		return CreatePayPalOrderResult{}, ErrRegistrationStateInvalid
	}

	rawProviderResponse := strings.TrimSpace(providerResult.RawResponse)
	if rawProviderResponse == "" {
		rawProviderResponse = buildProviderResponseFallbackJSON(providerResult)
	}
	nowRFC3339 := now.Format(time.RFC3339)
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO payments (
      id, tenant_id, registration_id, provider, provider_order_id, provider_capture_id,
      status, amount_cents, currency, donation_amount_cents, raw_provider_response,
      paid_at, refunded_at, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, NULL, ?, ?, ?, ?, ?, NULL, NULL, ?, ?)`,
		paymentID,
		tenantID,
		registrationID,
		ProviderPayPal,
		providerResult.OrderID,
		PaymentStatusCreated,
		input.AmountCents,
		currency,
		nullableInt(input.DonationAmountCents),
		rawProviderResponse,
		nowRFC3339,
		nowRFC3339,
	); err != nil {
		_ = tx.Rollback()
		return CreatePayPalOrderResult{}, fmt.Errorf("insert payment record: %w", err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE registrations
     SET status = ?, reserved_until = ?, confirmed_at = NULL, updated_at = ?
     WHERE id = ? AND tenant_id = ?`,
		registrationStatusPaymentPending,
		reservedUntil.Format(time.RFC3339),
		nowRFC3339,
		registrationID,
		tenantID,
	); err != nil {
		_ = tx.Rollback()
		return CreatePayPalOrderResult{}, fmt.Errorf("update registration payment_pending status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return CreatePayPalOrderResult{}, fmt.Errorf("commit paypal order transaction: %w", err)
	}

	var donation *int
	if input.DonationAmountCents > 0 {
		value := input.DonationAmountCents
		donation = &value
	}

	return CreatePayPalOrderResult{
		PaymentID:            paymentID,
		TenantID:             tenantID,
		RegistrationID:       registrationID,
		Provider:             ProviderPayPal,
		ProviderOrderID:      providerResult.OrderID,
		Status:               PaymentStatusCreated,
		AmountCents:          input.AmountCents,
		Currency:             currency,
		DonationAmountCents:  donation,
		PayPalMode:           settings.Mode,
		PayPalClientID:       clientID,
		ApproveURL:           providerResult.ApproveURL,
		RegistrationStatus:   registrationStatusPaymentPending,
		RegistrationReserved: reservedUntil,
	}, nil
}

func (s *Service) ProcessPayPalWebhook(ctx context.Context, headers PayPalWebhookHeaders, payload []byte) (ProcessPayPalWebhookResult, error) {
	if s.db == nil {
		return ProcessPayPalWebhookResult{}, fmt.Errorf("payment service database is nil")
	}
	raw := strings.TrimSpace(string(payload))
	if raw == "" {
		return ProcessPayPalWebhookResult{}, ErrWebhookPayloadInvalid
	}

	var webhook payPalWebhookPayload
	if err := json.Unmarshal([]byte(raw), &webhook); err != nil {
		return ProcessPayPalWebhookResult{}, ErrWebhookPayloadInvalid
	}

	eventID := strings.TrimSpace(webhook.ID)
	eventType := strings.ToUpper(strings.TrimSpace(webhook.EventType))
	if eventID == "" || eventType == "" {
		return ProcessPayPalWebhookResult{}, ErrWebhookPayloadInvalid
	}

	var webhookPayload map[string]any
	if err := json.Unmarshal([]byte(raw), &webhookPayload); err != nil {
		return ProcessPayPalWebhookResult{}, ErrWebhookPayloadInvalid
	}

	resourceID := strings.TrimSpace(webhook.Resource.ID)
	orderID := strings.TrimSpace(webhook.Resource.SupplementaryData.RelatedIDs.OrderID)
	captureID := strings.TrimSpace(webhook.Resource.SupplementaryData.RelatedIDs.CaptureID)
	if orderID == "" && strings.HasPrefix(eventType, "CHECKOUT.ORDER.") {
		orderID = resourceID
	}
	if captureID == "" && strings.HasPrefix(eventType, "PAYMENT.CAPTURE.") {
		captureID = resourceID
	}

	now := s.nowFn().UTC()
	nowRFC3339 := now.Format(time.RFC3339)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ProcessPayPalWebhookResult{}, fmt.Errorf("begin paypal webhook transaction: %w", err)
	}

	paymentItem, foundPayment, err := s.lookupPaymentForWebhookTx(ctx, tx, orderID, captureID)
	if err != nil {
		_ = tx.Rollback()
		return ProcessPayPalWebhookResult{}, err
	}

	webhookID := s.idFn("ppwh")
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO paypal_webhook_events (
      id, tenant_id, paypal_event_id, event_type, resource_id, payload_json,
      verified_at, processed_at, processing_status, error_message, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, NULL, NULL, ?, NULL, ?)`,
		webhookID,
		nullableString(paymentItem.TenantID),
		eventID,
		eventType,
		nullableString(resourceID),
		raw,
		WebhookStatusReceived,
		nowRFC3339,
	); err != nil {
		_ = tx.Rollback()
		if isUniqueWebhookEventError(err) {
			return ProcessPayPalWebhookResult{
				EventID:   eventID,
				EventType: eventType,
				Duplicate: true,
			}, nil
		}
		return ProcessPayPalWebhookResult{}, fmt.Errorf("insert paypal webhook event: %w", err)
	}

	result := ProcessPayPalWebhookResult{
		EventID:   eventID,
		EventType: eventType,
	}

	verifySettings := tenantPayPalSettings{
		ClientID:     strings.TrimSpace(s.cfg.FallbackClientID),
		ClientSecret: strings.TrimSpace(s.cfg.FallbackClientSecret),
		WebhookID:    strings.TrimSpace(s.cfg.FallbackWebhookID),
	}
	if foundPayment {
		settings, settingsErr := s.lookupTenantPayPalSettingsTx(ctx, tx, paymentItem.TenantID)
		if settingsErr != nil && !errors.Is(settingsErr, ErrPayPalDisabled) {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, settingsErr
		}
		if settingsErr == nil {
			verifySettings.Mode = settings.Mode
			if strings.TrimSpace(settings.ClientID) != "" {
				verifySettings.ClientID = strings.TrimSpace(settings.ClientID)
			}
			if strings.TrimSpace(settings.ClientSecret) != "" {
				verifySettings.ClientSecret = strings.TrimSpace(settings.ClientSecret)
			}
			if strings.TrimSpace(settings.WebhookID) != "" {
				verifySettings.WebhookID = strings.TrimSpace(settings.WebhookID)
			}
		}
	}

	if err := s.verifyWebhookSignature(ctx, verifySettings, headers, webhookPayload); err != nil {
		status := WebhookStatusFailed
		errorMessage := "verification_failed"
		if errors.Is(err, ErrWebhookSignatureInvalid) {
			errorMessage = "invalid_signature"
		}
		if errors.Is(err, ErrPayPalCredentialsMissing) || errors.Is(err, ErrPayPalConfigInvalid) {
			errorMessage = "verification_config_missing"
		}
		if _, updErr := tx.ExecContext(
			ctx,
			`UPDATE paypal_webhook_events
       SET processed_at = ?, processing_status = ?, error_message = ?
       WHERE id = ?`,
			nowRFC3339,
			status,
			errorMessage,
			webhookID,
		); updErr != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("mark failed paypal webhook verification: %w", updErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return ProcessPayPalWebhookResult{}, fmt.Errorf("commit failed paypal webhook verification: %w", commitErr)
		}
		return ProcessPayPalWebhookResult{}, err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE paypal_webhook_events SET verified_at = ? WHERE id = ?`,
		nowRFC3339,
		webhookID,
	); err != nil {
		_ = tx.Rollback()
		return ProcessPayPalWebhookResult{}, fmt.Errorf("mark paypal webhook verified: %w", err)
	}

	if !foundPayment {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE paypal_webhook_events
       SET processed_at = ?, processing_status = ?, error_message = ?
       WHERE id = ?`,
			nowRFC3339,
			WebhookStatusIgnored,
			"payment_not_found",
			webhookID,
		); err != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("mark paypal webhook as ignored: %w", err)
		}
		if err := tx.Commit(); err != nil {
			return ProcessPayPalWebhookResult{}, fmt.Errorf("commit ignored paypal webhook event: %w", err)
		}
		result.Ignored = true
		return result, nil
	}

	result.PaymentID = paymentItem.ID
	result.RegistrationID = paymentItem.RegistrationID

	action := webhookActionForEventType(eventType)
	switch action {
	case "paid":
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE payments
       SET status = ?, provider_capture_id = CASE WHEN ? <> '' THEN ? ELSE provider_capture_id END,
           raw_provider_response = ?, paid_at = COALESCE(paid_at, ?), updated_at = ?
       WHERE id = ?`,
			PaymentStatusPaid,
			captureID,
			captureID,
			raw,
			nowRFC3339,
			nowRFC3339,
			paymentItem.ID,
		); err != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("update payment paid status: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE registrations
       SET status = ?, confirmed_at = COALESCE(confirmed_at, ?), reserved_until = NULL, updated_at = ?
       WHERE id = ? AND tenant_id = ? AND status IN (?, ?, ?)`,
			registrationStatusConfirmed,
			nowRFC3339,
			nowRFC3339,
			paymentItem.RegistrationID,
			paymentItem.TenantID,
			registrationStatusPaymentPending,
			registrationStatusReserved,
			registrationStatusConfirmed,
		); err != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("confirm registration on payment webhook: %w", err)
		}
		result.PaymentStatus = PaymentStatusPaid
		result.Processed = true
	case "failed":
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE payments
       SET status = ?, provider_capture_id = CASE WHEN ? <> '' THEN ? ELSE provider_capture_id END,
           raw_provider_response = ?, updated_at = ?
       WHERE id = ?`,
			PaymentStatusFailed,
			captureID,
			captureID,
			raw,
			nowRFC3339,
			paymentItem.ID,
		); err != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("update payment failed status: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE registrations
       SET status = ?, reserved_until = NULL, updated_at = ?
       WHERE id = ? AND tenant_id = ? AND status = ?`,
			registrationStatusExpired,
			nowRFC3339,
			paymentItem.RegistrationID,
			paymentItem.TenantID,
			registrationStatusPaymentPending,
		); err != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("expire registration on failed payment webhook: %w", err)
		}
		result.PaymentStatus = PaymentStatusFailed
		result.Processed = true
	case "cancelled":
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE payments
       SET status = ?, raw_provider_response = ?, updated_at = ?
       WHERE id = ?`,
			PaymentStatusCancelled,
			raw,
			nowRFC3339,
			paymentItem.ID,
		); err != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("update payment cancelled status: %w", err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE registrations
       SET status = ?, reserved_until = NULL, updated_at = ?
       WHERE id = ? AND tenant_id = ? AND status = ?`,
			registrationStatusExpired,
			nowRFC3339,
			paymentItem.RegistrationID,
			paymentItem.TenantID,
			registrationStatusPaymentPending,
		); err != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("expire registration on cancelled payment webhook: %w", err)
		}
		result.PaymentStatus = PaymentStatusCancelled
		result.Processed = true
	case "refunded":
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE payments
       SET status = ?, raw_provider_response = ?, refunded_at = COALESCE(refunded_at, ?), updated_at = ?
       WHERE id = ?`,
			PaymentStatusRefunded,
			raw,
			nowRFC3339,
			nowRFC3339,
			paymentItem.ID,
		); err != nil {
			_ = tx.Rollback()
			return ProcessPayPalWebhookResult{}, fmt.Errorf("update payment refunded status: %w", err)
		}
		result.PaymentStatus = PaymentStatusRefunded
		result.Processed = true
	default:
		result.Ignored = true
	}

	processingStatus := WebhookStatusProcessed
	errorMessage := any(nil)
	if result.Ignored && !result.Processed {
		processingStatus = WebhookStatusIgnored
		errorMessage = "event_type_not_handled"
	}
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE paypal_webhook_events
     SET processed_at = ?, processing_status = ?, error_message = ?
     WHERE id = ?`,
		nowRFC3339,
		processingStatus,
		errorMessage,
		webhookID,
	); err != nil {
		_ = tx.Rollback()
		return ProcessPayPalWebhookResult{}, fmt.Errorf("update paypal webhook processing status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ProcessPayPalWebhookResult{}, fmt.Errorf("commit paypal webhook transaction: %w", err)
	}
	return result, nil
}

func (s *Service) lookupTenantPayPalSettings(ctx context.Context, tenantID string) (tenantPayPalSettings, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(paypal_mode, 'disabled'), COALESCE(paypal_client_id, ''), COALESCE(paypal_merchant_id, ''), COALESCE(settings_json, '')
     FROM tenant_settings
     WHERE tenant_id = ?`,
		tenantID,
	)
	return scanTenantPayPalSettingsRow(row)
}

func (s *Service) lookupTenantPayPalSettingsTx(ctx context.Context, tx *sql.Tx, tenantID string) (tenantPayPalSettings, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(paypal_mode, 'disabled'), COALESCE(paypal_client_id, ''), COALESCE(paypal_merchant_id, ''), COALESCE(settings_json, '')
     FROM tenant_settings
     WHERE tenant_id = ?`,
		tenantID,
	)
	return scanTenantPayPalSettingsRow(row)
}

func scanTenantPayPalSettingsRow(row interface{ Scan(dest ...any) error }) (tenantPayPalSettings, error) {
	var settings tenantPayPalSettings
	var settingsJSON string
	if err := row.Scan(&settings.Mode, &settings.ClientID, &settings.MerchantID, &settingsJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tenantPayPalSettings{}, ErrPayPalDisabled
		}
		return tenantPayPalSettings{}, fmt.Errorf("query tenant paypal settings: %w", err)
	}
	clientSecret, webhookID, err := parsePayPalSettingsJSON(settingsJSON)
	if err != nil {
		return tenantPayPalSettings{}, fmt.Errorf("%w: %v", ErrPayPalConfigInvalid, err)
	}
	settings.ClientSecret = clientSecret
	settings.WebhookID = webhookID
	settings.Mode = strings.ToLower(strings.TrimSpace(settings.Mode))
	switch settings.Mode {
	case "sandbox", "live":
		return settings, nil
	case "", "disabled":
		return tenantPayPalSettings{}, ErrPayPalDisabled
	default:
		return tenantPayPalSettings{}, fmt.Errorf("%w: mode %q", ErrPayPalConfigInvalid, settings.Mode)
	}
}

func parsePayPalSettingsJSON(raw string) (clientSecret string, webhookID string, err error) {
	content := strings.TrimSpace(raw)
	if content == "" {
		return "", "", nil
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(content), &payload); err != nil {
		return "", "", fmt.Errorf("parse tenant settings_json: %w", err)
	}

	clientSecret = firstSettingsString(
		payload,
		"paypal_client_secret",
		"paypalClientSecret",
		"client_secret",
		"clientSecret",
	)
	webhookID = firstSettingsString(
		payload,
		"paypal_webhook_id",
		"paypalWebhookID",
		"webhook_id",
		"webhookId",
	)

	if paypalRaw, ok := payload["paypal"]; ok {
		if paypalMap, ok := paypalRaw.(map[string]any); ok {
			if clientSecret == "" {
				clientSecret = firstSettingsString(
					paypalMap,
					"client_secret",
					"clientSecret",
					"paypal_client_secret",
				)
			}
			if webhookID == "" {
				webhookID = firstSettingsString(
					paypalMap,
					"webhook_id",
					"webhookId",
					"paypal_webhook_id",
				)
			}
		}
	}

	return strings.TrimSpace(clientSecret), strings.TrimSpace(webhookID), nil
}

func firstSettingsString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			continue
		}
		text = strings.TrimSpace(text)
		if text != "" {
			return text
		}
	}
	return ""
}

func (s *Service) lookupRegistrationForPaymentTx(ctx context.Context, tx *sql.Tx, tenantID, registrationID string) (registrationForPayment, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT id, tenant_id, status, COALESCE(reserved_until, '')
     FROM registrations
     WHERE tenant_id = ? AND id = ?
     LIMIT 1`,
		tenantID,
		registrationID,
	)
	var item registrationForPayment
	if err := row.Scan(&item.ID, &item.TenantID, &item.Status, &item.ReservedUntil); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return registrationForPayment{}, ErrRegistrationNotFound
		}
		return registrationForPayment{}, fmt.Errorf("query registration for payment: %w", err)
	}
	return item, nil
}

func (s *Service) lookupPaymentForWebhookTx(ctx context.Context, tx *sql.Tx, orderID, captureID string) (paymentLookup, bool, error) {
	tryLookup := func(query string, args ...any) (paymentLookup, bool, error) {
		row := tx.QueryRowContext(ctx, query, args...)
		var item paymentLookup
		if err := row.Scan(&item.ID, &item.TenantID, &item.RegistrationID, &item.ProviderOrderID, &item.Status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return paymentLookup{}, false, nil
			}
			return paymentLookup{}, false, err
		}
		return item, true, nil
	}

	if strings.TrimSpace(captureID) != "" {
		item, found, err := tryLookup(
			`SELECT id, tenant_id, registration_id, COALESCE(provider_order_id, ''), status
       FROM payments
       WHERE provider = ? AND provider_capture_id = ?
       ORDER BY created_at DESC
       LIMIT 1`,
			ProviderPayPal,
			captureID,
		)
		if err != nil {
			return paymentLookup{}, false, fmt.Errorf("query payment by capture id: %w", err)
		}
		if found {
			return item, true, nil
		}
	}

	if strings.TrimSpace(orderID) == "" {
		return paymentLookup{}, false, nil
	}
	item, found, err := tryLookup(
		`SELECT id, tenant_id, registration_id, COALESCE(provider_order_id, ''), status
     FROM payments
     WHERE provider = ? AND provider_order_id = ?
     ORDER BY created_at DESC
     LIMIT 1`,
		ProviderPayPal,
		orderID,
	)
	if err != nil {
		return paymentLookup{}, false, fmt.Errorf("query payment by order id: %w", err)
	}
	return item, found, nil
}

func canCreateOrderForRegistrationStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case registrationStatusReserved, registrationStatusPaymentPending, registrationStatusConfirmed:
		return true
	default:
		return false
	}
}

func webhookActionForEventType(eventType string) string {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case "PAYMENT.CAPTURE.COMPLETED":
		return "paid"
	case "PAYMENT.CAPTURE.DENIED", "PAYMENT.CAPTURE.DECLINED", "PAYMENT.CAPTURE.FAILED":
		return "failed"
	case "CHECKOUT.ORDER.CANCELLED", "CHECKOUT.ORDER.VOIDED":
		return "cancelled"
	case "PAYMENT.CAPTURE.REFUNDED":
		return "refunded"
	default:
		return "ignore"
	}
}

func (s *Service) verifyWebhookSignature(ctx context.Context, settings tenantPayPalSettings, headers PayPalWebhookHeaders, payload any) error {
	if !s.cfg.UseRealPayPalAPI {
		return nil
	}

	clientID := strings.TrimSpace(settings.ClientID)
	clientSecret := strings.TrimSpace(settings.ClientSecret)
	webhookID := strings.TrimSpace(settings.WebhookID)
	if clientID == "" {
		return ErrPayPalConfigInvalid
	}
	if clientSecret == "" || webhookID == "" {
		return ErrPayPalCredentialsMissing
	}

	mode := strings.ToLower(strings.TrimSpace(settings.Mode))
	tryModes := make([]string, 0, 2)
	switch mode {
	case "sandbox", "live":
		tryModes = append(tryModes, mode)
	default:
		tryModes = append(tryModes, "sandbox", "live")
	}

	var lastErr error
	for _, candidateMode := range tryModes {
		verifyResult, err := s.verifyFn(ctx, verifyWebhookProviderInput{
			Mode:      candidateMode,
			ClientID:  clientID,
			Secret:    clientSecret,
			WebhookID: webhookID,
			Headers:   headers,
			Payload:   payload,
		})
		if err != nil {
			lastErr = err
			continue
		}
		if strings.EqualFold(strings.TrimSpace(verifyResult.VerificationStatus), "SUCCESS") {
			return nil
		}
		lastErr = ErrWebhookSignatureInvalid
	}

	if lastErr == nil {
		return ErrWebhookSignatureInvalid
	}
	if errors.Is(lastErr, ErrWebhookSignatureInvalid) {
		return lastErr
	}
	return fmt.Errorf("%w: %v", ErrWebhookSignatureInvalid, lastErr)
}

func (s *Service) defaultCreatePayPalOrder(ctx context.Context, input createOrderProviderInput) (createOrderProviderResult, error) {
	if !s.cfg.UseRealPayPalAPI {
		return createPayPalOrderStub(input)
	}

	baseURL, err := s.apiBaseURL(input.Mode)
	if err != nil {
		return createOrderProviderResult{}, err
	}
	accessToken, err := s.fetchPayPalAccessToken(ctx, baseURL, input.ClientID, input.ClientSecret)
	if err != nil {
		return createOrderProviderResult{}, err
	}

	createPayload := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{
			{
				"reference_id": input.OrderRef,
				"custom_id":    input.RegistrationID,
				"amount": map[string]any{
					"currency_code": strings.ToUpper(strings.TrimSpace(input.Currency)),
					"value":         centsToPayPalAmount(input.AmountCents),
				},
			},
		},
	}
	bodyBytes, err := json.Marshal(createPayload)
	if err != nil {
		return createOrderProviderResult{}, fmt.Errorf("marshal paypal create-order request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v2/checkout/orders", bytes.NewReader(bodyBytes))
	if err != nil {
		return createOrderProviderResult{}, fmt.Errorf("build paypal create-order request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return createOrderProviderResult{}, fmt.Errorf("paypal create-order request failed: %w", err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return createOrderProviderResult{}, fmt.Errorf("read paypal create-order response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return createOrderProviderResult{}, fmt.Errorf("paypal create-order returned status %d: %s", resp.StatusCode, trimmedPayload(rawResp))
	}

	var parsed struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}
	if err := json.Unmarshal(rawResp, &parsed); err != nil {
		return createOrderProviderResult{}, fmt.Errorf("decode paypal create-order response: %w", err)
	}
	approveURL := ""
	for _, link := range parsed.Links {
		if strings.EqualFold(strings.TrimSpace(link.Rel), "approve") {
			approveURL = strings.TrimSpace(link.Href)
			break
		}
	}
	if strings.TrimSpace(parsed.ID) == "" {
		return createOrderProviderResult{}, fmt.Errorf("paypal create-order response missing id")
	}
	if strings.TrimSpace(parsed.Status) == "" {
		parsed.Status = "CREATED"
	}
	return createOrderProviderResult{
		OrderID:     strings.TrimSpace(parsed.ID),
		Status:      strings.ToUpper(strings.TrimSpace(parsed.Status)),
		ApproveURL:  approveURL,
		RawResponse: string(rawResp),
	}, nil
}

func (s *Service) defaultVerifyPayPalWebhookSignature(ctx context.Context, input verifyWebhookProviderInput) (verifyWebhookProviderResult, error) {
	if !s.cfg.UseRealPayPalAPI {
		return verifyWebhookProviderResult{VerificationStatus: "SUCCESS"}, nil
	}

	baseURL, err := s.apiBaseURL(input.Mode)
	if err != nil {
		return verifyWebhookProviderResult{}, err
	}
	clientID := strings.TrimSpace(input.ClientID)
	secret := strings.TrimSpace(input.Secret)
	webhookID := strings.TrimSpace(input.WebhookID)
	if clientID == "" {
		return verifyWebhookProviderResult{}, ErrPayPalConfigInvalid
	}
	if secret == "" || webhookID == "" {
		return verifyWebhookProviderResult{}, ErrPayPalCredentialsMissing
	}

	headers := PayPalWebhookHeaders{
		TransmissionID:   strings.TrimSpace(input.Headers.TransmissionID),
		TransmissionTime: strings.TrimSpace(input.Headers.TransmissionTime),
		TransmissionSig:  strings.TrimSpace(input.Headers.TransmissionSig),
		CertURL:          strings.TrimSpace(input.Headers.CertURL),
		AuthAlgo:         strings.TrimSpace(input.Headers.AuthAlgo),
	}
	if headers.TransmissionID == "" || headers.TransmissionTime == "" || headers.TransmissionSig == "" || headers.CertURL == "" || headers.AuthAlgo == "" {
		return verifyWebhookProviderResult{}, ErrWebhookSignatureInvalid
	}

	accessToken, err := s.fetchPayPalAccessToken(ctx, baseURL, clientID, secret)
	if err != nil {
		return verifyWebhookProviderResult{}, err
	}

	payload := map[string]any{
		"transmission_id":   headers.TransmissionID,
		"transmission_time": headers.TransmissionTime,
		"cert_url":          headers.CertURL,
		"auth_algo":         headers.AuthAlgo,
		"transmission_sig":  headers.TransmissionSig,
		"webhook_id":        webhookID,
		"webhook_event":     input.Payload,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return verifyWebhookProviderResult{}, fmt.Errorf("marshal paypal verify-signature request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseURL+"/v1/notifications/verify-webhook-signature",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return verifyWebhookProviderResult{}, fmt.Errorf("build paypal verify-signature request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return verifyWebhookProviderResult{}, fmt.Errorf("paypal verify-signature request failed: %w", err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return verifyWebhookProviderResult{}, fmt.Errorf("read paypal verify-signature response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return verifyWebhookProviderResult{}, fmt.Errorf("paypal verify-signature returned status %d: %s", resp.StatusCode, trimmedPayload(rawResp))
	}

	var result verifyWebhookProviderResult
	if err := json.Unmarshal(rawResp, &result); err != nil {
		return verifyWebhookProviderResult{}, fmt.Errorf("decode paypal verify-signature response: %w", err)
	}
	if strings.TrimSpace(result.VerificationStatus) == "" {
		return verifyWebhookProviderResult{}, ErrWebhookSignatureInvalid
	}
	return result, nil
}

func (s *Service) fetchPayPalAccessToken(ctx context.Context, baseURL, clientID, clientSecret string) (string, error) {
	form := neturl.Values{}
	form.Set("grant_type", "client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build paypal token request: %w", err)
	}
	req.SetBasicAuth(strings.TrimSpace(clientID), strings.TrimSpace(clientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("paypal token request failed: %w", err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", fmt.Errorf("read paypal token response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= 300 {
		return "", fmt.Errorf("paypal token returned status %d: %s", resp.StatusCode, trimmedPayload(rawResp))
	}

	var tokenPayload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(rawResp, &tokenPayload); err != nil {
		return "", fmt.Errorf("decode paypal token response: %w", err)
	}
	if strings.TrimSpace(tokenPayload.AccessToken) == "" {
		return "", fmt.Errorf("paypal token response missing access_token")
	}
	return strings.TrimSpace(tokenPayload.AccessToken), nil
}

func (s *Service) apiBaseURL(mode string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	switch normalized {
	case "sandbox":
		return strings.TrimRight(strings.TrimSpace(s.cfg.SandboxAPIBaseURL), "/"), nil
	case "live":
		return strings.TrimRight(strings.TrimSpace(s.cfg.LiveAPIBaseURL), "/"), nil
	default:
		return "", fmt.Errorf("%w: mode %q", ErrPayPalConfigInvalid, mode)
	}
}

func createPayPalOrderStub(input createOrderProviderInput) (createOrderProviderResult, error) {
	orderID := strings.TrimSpace(input.OrderRef)
	if orderID == "" {
		orderID = defaultID("ppord")
	}
	base := "https://www.paypal.com"
	if strings.EqualFold(strings.TrimSpace(input.Mode), "sandbox") {
		base = "https://www.sandbox.paypal.com"
	}
	approveURL := fmt.Sprintf("%s/checkoutnow?token=%s", base, orderID)

	response := map[string]any{
		"id":     orderID,
		"status": "CREATED",
		"links": []map[string]string{
			{
				"rel":  "approve",
				"href": approveURL,
			},
		},
	}
	rawBytes, err := json.Marshal(response)
	if err != nil {
		return createOrderProviderResult{}, fmt.Errorf("marshal provider create-order payload: %w", err)
	}
	return createOrderProviderResult{
		OrderID:     orderID,
		Status:      "CREATED",
		ApproveURL:  approveURL,
		RawResponse: string(rawBytes),
	}, nil
}

func centsToPayPalAmount(cents int) string {
	if cents < 0 {
		cents = 0
	}
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func trimmedPayload(payload []byte) string {
	raw := strings.TrimSpace(string(payload))
	if raw == "" {
		return ""
	}
	const maxLen = 512
	if len(raw) > maxLen {
		return raw[:maxLen] + "..."
	}
	return raw
}

func buildProviderResponseFallbackJSON(result createOrderProviderResult) string {
	payload := map[string]any{
		"id":          result.OrderID,
		"status":      result.Status,
		"approve_url": result.ApproveURL,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func isUniqueWebhookEventError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: paypal_webhook_events.paypal_event_id")
}

func nullableInt(value int) any {
	if value <= 0 {
		return nil
	}
	return value
}

func nullableString(value string) any {
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
