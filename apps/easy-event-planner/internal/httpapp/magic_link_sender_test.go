package httpapp

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/auth"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
)

func TestBuildMagicLinkMailContentOrganizerLogin(t *testing.T) {
	expiresAt := time.Date(2026, 5, 24, 21, 15, 0, 0, time.UTC)
	message := auth.MagicLinkMessage{
		ToEmail:    "owner@example.com",
		TenantSlug: "demo",
		Purpose:    auth.PurposeOrganizerLogin,
		VerifyURL:  "http://127.0.0.1:18080/api/v1/auth/magic-link/verify?token=abc123",
		ExpiresAt:  expiresAt,
	}

	subject, bodyText, bodyHTML := buildMagicLinkMailContent(message)
	if subject != "Dein Login-Link fuer Easy Event Planner" {
		t.Fatalf("expected organizer subject, got %q", subject)
	}
	if !strings.Contains(bodyText, message.VerifyURL) {
		t.Fatalf("expected body text to include verify url")
	}
	if !strings.Contains(bodyText, "Gueltig bis: 2026-05-24T21:15:00Z (UTC)") {
		t.Fatalf("expected body text to include expiry, got %q", bodyText)
	}
	if !strings.Contains(bodyHTML, "<a href=\""+message.VerifyURL+"\">") {
		t.Fatalf("expected body html to include verify link")
	}
}

func TestBuildMagicLinkMetadataJSON(t *testing.T) {
	expiresAt := time.Date(2026, 5, 24, 21, 15, 0, 0, time.UTC)
	message := auth.MagicLinkMessage{
		ToEmail:    "owner@example.com",
		TenantSlug: "demo",
		Purpose:    auth.PurposeOrganizerLogin,
		VerifyURL:  "http://127.0.0.1:18080/api/v1/auth/magic-link/verify?token=abc123",
		ExpiresAt:  expiresAt,
	}

	metadataJSON := buildMagicLinkMetadataJSON(message)
	if metadataJSON == "" {
		t.Fatalf("expected metadata json to be non-empty")
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(metadataJSON), &payload); err != nil {
		t.Fatalf("unmarshal metadata json: %v", err)
	}

	if payload["purpose"] != auth.PurposeOrganizerLogin {
		t.Fatalf("expected purpose %q, got %v", auth.PurposeOrganizerLogin, payload["purpose"])
	}
	if payload["tenant_slug"] != "demo" {
		t.Fatalf("expected tenant_slug demo, got %v", payload["tenant_slug"])
	}
	if payload["verify_url"] != message.VerifyURL {
		t.Fatalf("expected verify_url %q, got %v", message.VerifyURL, payload["verify_url"])
	}
	if payload["expires_at"] != "2026-05-24T21:15:00Z" {
		t.Fatalf("expected expires_at timestamp, got %v", payload["expires_at"])
	}
}

func TestBuildMagicLinkSenderFallsBackToLogSender(t *testing.T) {
	cfg := config.Config{
		Env:           "development",
		MailProvider:  "smtp",
		MailFromEmail: "noreply@example.com",
		SESPort:       587,
		// Missing host -> invalid smtp config.
	}

	sender := buildMagicLinkSender(cfg)
	if _, ok := sender.(*auth.LogSender); !ok {
		t.Fatalf("expected fallback sender to be *auth.LogSender, got %T", sender)
	}
}
