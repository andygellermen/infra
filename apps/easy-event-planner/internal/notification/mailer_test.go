package notification

import (
	"context"
	"strings"
	"testing"
)

func TestNewMailerDefaultsToLogProvider(t *testing.T) {
	mailer, err := NewMailer(MailerConfig{})
	if err != nil {
		t.Fatalf("NewMailer returned error: %v", err)
	}
	if _, ok := mailer.(*LogMailer); !ok {
		t.Fatalf("expected *LogMailer, got %T", mailer)
	}
}

func TestNewMailerSMTPValidation(t *testing.T) {
	_, err := NewMailer(MailerConfig{
		Provider:  ProviderSMTP,
		FromEmail: "noreply@example.com",
		SMTPHost:  "smtp.example.com",
		SMTPPort:  587,
		SMTPUser:  "smtp-user",
	})
	if err == nil {
		t.Fatal("expected validation error when smtp pass is missing")
	}
}

func TestLogMailerValidatesRecipient(t *testing.T) {
	mailer := &LogMailer{
		DefaultFromEmail: "noreply@example.com",
	}
	err := mailer.Send(context.Background(), Message{
		Recipient: "not-an-email",
		Subject:   "Test",
		BodyText:  "hello",
	})
	if err == nil {
		t.Fatal("expected error for invalid recipient")
	}
}

func TestBuildRFC822MessageMultipart(t *testing.T) {
	msg, err := buildRFC822Message(
		"Example Events <noreply@example.com>",
		"owner@example.com",
		"Subject\r\nInjected",
		"Text body",
		"<p>HTML body</p>",
	)
	if err != nil {
		t.Fatalf("buildRFC822Message returned error: %v", err)
	}
	raw := string(msg)
	if !strings.Contains(raw, "multipart/alternative") {
		t.Fatalf("expected multipart content type, got %q", raw)
	}
	if strings.Contains(raw, "\r\nInjected") {
		t.Fatalf("header injection was not sanitized: %q", raw)
	}
}
