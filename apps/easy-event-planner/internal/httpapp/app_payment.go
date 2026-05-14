package httpapp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/payment"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

func (a *App) handlePublicPayPalCreateOrder(w http.ResponseWriter, r *http.Request, tenantItem tenant.Tenant) {
	if a.paymentService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Payment-Service ist nicht verfuegbar.")
		return
	}

	var request struct {
		RegistrationID      string `json:"registration_id"`
		AmountCents         int    `json:"amount_cents"`
		Currency            string `json:"currency"`
		DonationAmountCents int    `json:"donation_amount_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Ungueltige Anfrage.")
		return
	}

	result, err := a.paymentService.CreatePayPalOrder(r.Context(), payment.CreatePayPalOrderInput{
		TenantID:            tenantItem.ID,
		RegistrationID:      request.RegistrationID,
		AmountCents:         request.AmountCents,
		Currency:            request.Currency,
		DonationAmountCents: request.DonationAmountCents,
	})
	if err != nil {
		a.writePayPalError(w, err)
		return
	}

	var donationAmount any
	if result.DonationAmountCents != nil {
		donationAmount = *result.DonationAmountCents
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok": true,
		"payment": map[string]any{
			"id":                    result.PaymentID,
			"tenant_id":             result.TenantID,
			"registration_id":       result.RegistrationID,
			"provider":              result.Provider,
			"provider_order_id":     result.ProviderOrderID,
			"status":                result.Status,
			"amount_cents":          result.AmountCents,
			"currency":              result.Currency,
			"donation_amount_cents": donationAmount,
			"approve_url":           result.ApproveURL,
			"paypal_mode":           result.PayPalMode,
			"paypal_client_id":      result.PayPalClientID,
		},
		"registration": map[string]any{
			"id":             result.RegistrationID,
			"status":         result.RegistrationStatus,
			"reserved_until": result.RegistrationReserved.UTC().Format(time.RFC3339),
		},
	})
}

func (a *App) handlePayPalWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Methode nicht erlaubt.")
		return
	}
	if a.paymentService == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Payment-Service ist nicht verfuegbar.")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Webhook-Payload konnte nicht gelesen werden.")
		return
	}

	headers := payment.PayPalWebhookHeaders{
		TransmissionID:   headerValue(r, "Paypal-Transmission-Id", "PAYPAL-TRANSMISSION-ID"),
		TransmissionTime: headerValue(r, "Paypal-Transmission-Time", "PAYPAL-TRANSMISSION-TIME"),
		TransmissionSig:  headerValue(r, "Paypal-Transmission-Sig", "PAYPAL-TRANSMISSION-SIG"),
		CertURL:          headerValue(r, "Paypal-Cert-Url", "PAYPAL-CERT-URL"),
		AuthAlgo:         headerValue(r, "Paypal-Auth-Algo", "PAYPAL-AUTH-ALGO"),
	}

	result, err := a.paymentService.ProcessPayPalWebhook(r.Context(), headers, body)
	if err != nil {
		a.writePayPalError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"event_id":        result.EventID,
		"event_type":      result.EventType,
		"payment_id":      result.PaymentID,
		"registration_id": result.RegistrationID,
		"payment_status":  result.PaymentStatus,
		"processed":       result.Processed,
		"ignored":         result.Ignored,
		"duplicate":       result.Duplicate,
	})
}

func (a *App) writePayPalError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, payment.ErrPayPalDisabled):
		writeAPIError(w, http.StatusConflict, "PAYMENT_REQUIRED", "PayPal ist fuer diesen Mandanten nicht aktiviert.")
	case errors.Is(err, payment.ErrPayPalConfigInvalid):
		writeAPIError(w, http.StatusConflict, "PAYMENT_REQUIRED", "PayPal-Konfiguration ist unvollstaendig.")
	case errors.Is(err, payment.ErrPayPalCredentialsMissing):
		writeAPIError(w, http.StatusConflict, "PAYMENT_REQUIRED", "PayPal-Credentials sind nicht vollstaendig konfiguriert.")
	case errors.Is(err, payment.ErrRegistrationNotFound):
		writeAPIError(w, http.StatusNotFound, "REGISTRATION_NOT_FOUND", "Anmeldung nicht gefunden.")
	case errors.Is(err, payment.ErrRegistrationStateInvalid):
		writeAPIError(w, http.StatusConflict, "REGISTRATION_STATE_INVALID", "Anmeldung kann nicht fuer Zahlung reserviert werden.")
	case errors.Is(err, payment.ErrInvalidCreateOrderInput):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, payment.ErrWebhookPayloadInvalid):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Webhook-Payload ist ungueltig.")
	case errors.Is(err, payment.ErrWebhookSignatureInvalid):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_WEBHOOK_SIGNATURE", "Webhook-Signatur ist ungueltig.")
	default:
		message := strings.TrimSpace(err.Error())
		if message == "" {
			message = "Payment konnte nicht verarbeitet werden."
		}
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", message)
	}
}

func headerValue(r *http.Request, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(r.Header.Get(key))
		if value != "" {
			return value
		}
	}
	return ""
}
