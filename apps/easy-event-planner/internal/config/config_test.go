package config

import (
	"path/filepath"
	"testing"
	"time"
)

func clearEEPEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"EEP_ENV",
		"EEP_HTTP_ADDR",
		"EEP_BASE_URL",
		"EEP_VERSION",
		"EEP_DB_DRIVER",
		"EEP_DB_PATH",
		"EEP_TOKEN_PEPPER",
		"EEP_SESSION_COOKIE_NAME",
		"EEP_SECURE_COOKIES",
		"EEP_MAIL_PROVIDER",
		"EEP_MAIL_FROM",
		"EEP_MAIL_FROM_NAME",
		"EEP_PAYPAL_USE_REAL_API",
		"EEP_PAYPAL_CLIENT_ID",
		"EEP_PAYPAL_CLIENT_SECRET",
		"EEP_PAYPAL_WEBHOOK_ID",
		"EEP_PAYPAL_SANDBOX_API_BASE_URL",
		"EEP_PAYPAL_LIVE_API_BASE_URL",
		"EEP_PAYPAL_HTTP_TIMEOUT",
		"EEP_SES_REGION",
		"EEP_SES_SMTP_HOST",
		"EEP_SES_SMTP_PORT",
		"EEP_SES_SMTP_USER",
		"EEP_SES_SMTP_PASS",
		"EEP_EMAIL_WORKER_POLL_INTERVAL",
		"EEP_EMAIL_WORKER_BATCH_SIZE",
		"EEP_SESSION_TTL",
		"EEP_MAGIC_LINK_TTL",
		"EEP_REGISTRATION_MAGIC_LINK_TTL",
		"EEP_WAITLIST_OFFER_TTL",
		"EEP_CERTIFICATE_ACCESS_TTL",
		"EEP_AUTH_RATE_LIMIT_WINDOW",
		"EEP_AUTH_RATE_LIMIT_REQUESTS",
		"EEP_HTTP_READ_HEADER_TIMEOUT",
		"EEP_HTTP_SHUTDOWN_TIMEOUT",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEEPEnv(t)

	cfg, err := Load("1.2.3")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	actualCwd, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	expectedDBPath := filepath.Join(actualCwd, "data", "easy-event-planner.sqlite")

	if cfg.AppName != "easy-event-planner" {
		t.Fatalf("expected AppName easy-event-planner, got %q", cfg.AppName)
	}
	if cfg.Env != "development" {
		t.Fatalf("expected default env development, got %q", cfg.Env)
	}
	if cfg.Addr != ":8080" {
		t.Fatalf("expected default addr :8080, got %q", cfg.Addr)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Fatalf("expected default base url, got %q", cfg.BaseURL)
	}
	if cfg.Version != "1.2.3" {
		t.Fatalf("expected version 1.2.3, got %q", cfg.Version)
	}
	if cfg.DBDriver != "sqlite" {
		t.Fatalf("expected db driver sqlite, got %q", cfg.DBDriver)
	}
	if cfg.DBPath != expectedDBPath {
		t.Fatalf("expected db path %q, got %q", expectedDBPath, cfg.DBPath)
	}
	if cfg.TokenPepper != "dev-only-change-me" {
		t.Fatalf("expected default token pepper dev-only-change-me, got %q", cfg.TokenPepper)
	}
	if cfg.SessionCookieName != "eep_session" {
		t.Fatalf("expected default cookie name eep_session, got %q", cfg.SessionCookieName)
	}
	if cfg.SecureCookies {
		t.Fatalf("expected secure cookies to default to false in development")
	}
	if cfg.MailProvider != "log" {
		t.Fatalf("expected default mail provider log, got %q", cfg.MailProvider)
	}
	if cfg.MailFromEmail != "noreply@example.com" {
		t.Fatalf("expected default mail from noreply@example.com, got %q", cfg.MailFromEmail)
	}
	if cfg.MailFromName != "" {
		t.Fatalf("expected default mail from name empty, got %q", cfg.MailFromName)
	}
	if cfg.PayPalUseRealAPI {
		t.Fatalf("expected paypal real api default false")
	}
	if cfg.PayPalClientID != "" {
		t.Fatalf("expected empty paypal client id by default, got %q", cfg.PayPalClientID)
	}
	if cfg.PayPalClientSecret != "" {
		t.Fatalf("expected empty paypal client secret by default")
	}
	if cfg.PayPalWebhookID != "" {
		t.Fatalf("expected empty paypal webhook id by default, got %q", cfg.PayPalWebhookID)
	}
	if cfg.PayPalSandboxAPIBaseURL != "https://api-m.sandbox.paypal.com" {
		t.Fatalf("expected default paypal sandbox base url, got %q", cfg.PayPalSandboxAPIBaseURL)
	}
	if cfg.PayPalLiveAPIBaseURL != "https://api-m.paypal.com" {
		t.Fatalf("expected default paypal live base url, got %q", cfg.PayPalLiveAPIBaseURL)
	}
	if cfg.PayPalHTTPTimeout != 15*time.Second {
		t.Fatalf("expected default paypal http timeout 15s, got %s", cfg.PayPalHTTPTimeout)
	}
	if cfg.SESRegion != "eu-north-1" {
		t.Fatalf("expected default ses region eu-north-1, got %q", cfg.SESRegion)
	}
	if cfg.SESPort != 587 {
		t.Fatalf("expected default ses smtp port 587, got %d", cfg.SESPort)
	}
	if cfg.EmailWorkerPollInterval != 3*time.Second {
		t.Fatalf("expected default worker poll interval 3s, got %s", cfg.EmailWorkerPollInterval)
	}
	if cfg.EmailWorkerBatchSize != 10 {
		t.Fatalf("expected default worker batch size 10, got %d", cfg.EmailWorkerBatchSize)
	}
	if cfg.SessionTTL != 12*time.Hour {
		t.Fatalf("expected default session ttl 12h, got %s", cfg.SessionTTL)
	}
	if cfg.MagicLinkTTL != 15*time.Minute {
		t.Fatalf("expected default magic link ttl 15m, got %s", cfg.MagicLinkTTL)
	}
	if cfg.RegistrationTTL != 30*time.Minute {
		t.Fatalf("expected default registration ttl 30m, got %s", cfg.RegistrationTTL)
	}
	if cfg.WaitlistOfferTTL != 24*time.Hour {
		t.Fatalf("expected default waitlist ttl 24h, got %s", cfg.WaitlistOfferTTL)
	}
	if cfg.CertificateTTL != 30*time.Minute {
		t.Fatalf("expected default certificate ttl 30m, got %s", cfg.CertificateTTL)
	}
	if cfg.AuthRateLimit != 5 {
		t.Fatalf("expected default auth rate limit 5, got %d", cfg.AuthRateLimit)
	}
	if cfg.AuthRateWindow != 15*time.Minute {
		t.Fatalf("expected default auth rate window 15m, got %s", cfg.AuthRateWindow)
	}
	if cfg.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("expected read header timeout 5s, got %s", cfg.ReadHeaderTimeout)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("expected shutdown timeout 10s, got %s", cfg.ShutdownTimeout)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	clearEEPEnv(t)
	t.Setenv("EEP_ENV", "production")
	t.Setenv("EEP_HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("EEP_BASE_URL", "https://events.example.com")
	t.Setenv("EEP_VERSION", "2026.05.13")
	t.Setenv("EEP_DB_DRIVER", "postgres")
	t.Setenv("EEP_DB_PATH", "/var/lib/eep/eep.db")
	t.Setenv("EEP_TOKEN_PEPPER", "pepper-123")
	t.Setenv("EEP_SESSION_COOKIE_NAME", "eep_auth")
	t.Setenv("EEP_SECURE_COOKIES", "true")
	t.Setenv("EEP_MAIL_PROVIDER", "ses")
	t.Setenv("EEP_MAIL_FROM", "events@example.com")
	t.Setenv("EEP_MAIL_FROM_NAME", "Events Team")
	t.Setenv("EEP_PAYPAL_USE_REAL_API", "true")
	t.Setenv("EEP_PAYPAL_CLIENT_ID", "paypal-client-id")
	t.Setenv("EEP_PAYPAL_CLIENT_SECRET", "paypal-client-secret")
	t.Setenv("EEP_PAYPAL_WEBHOOK_ID", "WH-123")
	t.Setenv("EEP_PAYPAL_SANDBOX_API_BASE_URL", "https://sandbox-paypal.example.test")
	t.Setenv("EEP_PAYPAL_LIVE_API_BASE_URL", "https://live-paypal.example.test")
	t.Setenv("EEP_PAYPAL_HTTP_TIMEOUT", "11s")
	t.Setenv("EEP_SES_REGION", "eu-central-1")
	t.Setenv("EEP_SES_SMTP_HOST", "email-smtp.eu-central-1.amazonaws.com")
	t.Setenv("EEP_SES_SMTP_PORT", "2525")
	t.Setenv("EEP_SES_SMTP_USER", "smtp-user")
	t.Setenv("EEP_SES_SMTP_PASS", "smtp-pass")
	t.Setenv("EEP_EMAIL_WORKER_POLL_INTERVAL", "7s")
	t.Setenv("EEP_EMAIL_WORKER_BATCH_SIZE", "25")
	t.Setenv("EEP_SESSION_TTL", "8h")
	t.Setenv("EEP_MAGIC_LINK_TTL", "10m")
	t.Setenv("EEP_REGISTRATION_MAGIC_LINK_TTL", "20m")
	t.Setenv("EEP_WAITLIST_OFFER_TTL", "18h")
	t.Setenv("EEP_CERTIFICATE_ACCESS_TTL", "40m")
	t.Setenv("EEP_AUTH_RATE_LIMIT_WINDOW", "20m")
	t.Setenv("EEP_AUTH_RATE_LIMIT_REQUESTS", "8")
	t.Setenv("EEP_HTTP_READ_HEADER_TIMEOUT", "3s")
	t.Setenv("EEP_HTTP_SHUTDOWN_TIMEOUT", "15s")

	cfg, err := Load("ignored")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Env != "production" {
		t.Fatalf("expected env production, got %q", cfg.Env)
	}
	if cfg.Addr != "127.0.0.1:9090" {
		t.Fatalf("expected addr override, got %q", cfg.Addr)
	}
	if cfg.BaseURL != "https://events.example.com" {
		t.Fatalf("expected base url override, got %q", cfg.BaseURL)
	}
	if cfg.Version != "2026.05.13" {
		t.Fatalf("expected version from env, got %q", cfg.Version)
	}
	if cfg.DBDriver != "postgres" {
		t.Fatalf("expected db driver postgres, got %q", cfg.DBDriver)
	}
	if cfg.DBPath != "/var/lib/eep/eep.db" {
		t.Fatalf("expected db path override, got %q", cfg.DBPath)
	}
	if cfg.TokenPepper != "pepper-123" {
		t.Fatalf("expected token pepper override, got %q", cfg.TokenPepper)
	}
	if cfg.SessionCookieName != "eep_auth" {
		t.Fatalf("expected session cookie override, got %q", cfg.SessionCookieName)
	}
	if !cfg.SecureCookies {
		t.Fatalf("expected secure cookies override to true")
	}
	if cfg.MailProvider != "ses" {
		t.Fatalf("expected mail provider ses, got %q", cfg.MailProvider)
	}
	if cfg.MailFromEmail != "events@example.com" {
		t.Fatalf("expected mail from override, got %q", cfg.MailFromEmail)
	}
	if cfg.MailFromName != "Events Team" {
		t.Fatalf("expected mail from name override, got %q", cfg.MailFromName)
	}
	if !cfg.PayPalUseRealAPI {
		t.Fatalf("expected paypal real api override to true")
	}
	if cfg.PayPalClientID != "paypal-client-id" {
		t.Fatalf("expected paypal client id override, got %q", cfg.PayPalClientID)
	}
	if cfg.PayPalClientSecret != "paypal-client-secret" {
		t.Fatalf("expected paypal client secret override, got %q", cfg.PayPalClientSecret)
	}
	if cfg.PayPalWebhookID != "WH-123" {
		t.Fatalf("expected paypal webhook id override, got %q", cfg.PayPalWebhookID)
	}
	if cfg.PayPalSandboxAPIBaseURL != "https://sandbox-paypal.example.test" {
		t.Fatalf("expected paypal sandbox base url override, got %q", cfg.PayPalSandboxAPIBaseURL)
	}
	if cfg.PayPalLiveAPIBaseURL != "https://live-paypal.example.test" {
		t.Fatalf("expected paypal live base url override, got %q", cfg.PayPalLiveAPIBaseURL)
	}
	if cfg.PayPalHTTPTimeout != 11*time.Second {
		t.Fatalf("expected paypal http timeout 11s, got %s", cfg.PayPalHTTPTimeout)
	}
	if cfg.SESRegion != "eu-central-1" {
		t.Fatalf("expected ses region override, got %q", cfg.SESRegion)
	}
	if cfg.SESHost != "email-smtp.eu-central-1.amazonaws.com" {
		t.Fatalf("expected ses host override, got %q", cfg.SESHost)
	}
	if cfg.SESPort != 2525 {
		t.Fatalf("expected ses port override 2525, got %d", cfg.SESPort)
	}
	if cfg.SESUser != "smtp-user" {
		t.Fatalf("expected ses user override, got %q", cfg.SESUser)
	}
	if cfg.SESPass != "smtp-pass" {
		t.Fatalf("expected ses pass override, got %q", cfg.SESPass)
	}
	if cfg.EmailWorkerPollInterval != 7*time.Second {
		t.Fatalf("expected worker poll interval 7s, got %s", cfg.EmailWorkerPollInterval)
	}
	if cfg.EmailWorkerBatchSize != 25 {
		t.Fatalf("expected worker batch size 25, got %d", cfg.EmailWorkerBatchSize)
	}
	if cfg.SessionTTL != 8*time.Hour {
		t.Fatalf("expected session ttl 8h, got %s", cfg.SessionTTL)
	}
	if cfg.MagicLinkTTL != 10*time.Minute {
		t.Fatalf("expected magic link ttl 10m, got %s", cfg.MagicLinkTTL)
	}
	if cfg.RegistrationTTL != 20*time.Minute {
		t.Fatalf("expected registration ttl 20m, got %s", cfg.RegistrationTTL)
	}
	if cfg.WaitlistOfferTTL != 18*time.Hour {
		t.Fatalf("expected waitlist ttl 18h, got %s", cfg.WaitlistOfferTTL)
	}
	if cfg.CertificateTTL != 40*time.Minute {
		t.Fatalf("expected certificate ttl 40m, got %s", cfg.CertificateTTL)
	}
	if cfg.AuthRateLimit != 8 {
		t.Fatalf("expected auth rate limit 8, got %d", cfg.AuthRateLimit)
	}
	if cfg.AuthRateWindow != 20*time.Minute {
		t.Fatalf("expected auth rate window 20m, got %s", cfg.AuthRateWindow)
	}
	if cfg.ReadHeaderTimeout != 3*time.Second {
		t.Fatalf("expected read header timeout 3s, got %s", cfg.ReadHeaderTimeout)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Fatalf("expected shutdown timeout 15s, got %s", cfg.ShutdownTimeout)
	}
}

func TestLoadRejectsInvalidBaseURL(t *testing.T) {
	clearEEPEnv(t)
	t.Setenv("EEP_BASE_URL", "events.example.com")

	_, err := Load("dev")
	if err == nil {
		t.Fatal("expected error for invalid base url")
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	clearEEPEnv(t)
	t.Setenv("EEP_HTTP_SHUTDOWN_TIMEOUT", "nope")

	_, err := Load("dev")
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestLoadRejectsUnsupportedMailProvider(t *testing.T) {
	clearEEPEnv(t)
	t.Setenv("EEP_MAIL_PROVIDER", "pigeon")

	_, err := Load("dev")
	if err == nil {
		t.Fatal("expected error for unsupported mail provider")
	}
}
