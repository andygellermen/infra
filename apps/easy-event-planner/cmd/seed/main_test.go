package main

import (
	"strings"
	"testing"
)

func clearSeedEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"EEP_SEED_TENANT_SLUG",
		"EEP_SEED_TENANT_NAME",
		"EEP_SEED_TENANT_PUBLIC_BASE_URL",
		"EEP_SEED_TENANT_TIMEZONE",
		"EEP_SEED_TENANT_LOCALE",
		"EEP_SEED_TENANT_STATUS",
		"EEP_SEED_SENDER_EMAIL",
		"EEP_SEED_SENDER_NAME",
		"EEP_SEED_ADMIN_EMAIL",
		"EEP_SEED_ADMIN_NAME",
		"EEP_SEED_ADMIN_ROLE",
		"EEP_SEED_PAYPAL_MODE",
		"EEP_SEED_PAYPAL_CLIENT_ID",
		"EEP_SEED_PAYPAL_MERCHANT_ID",
		"EEP_SEED_RETENTION_DAYS",
		"EEP_SEED_SETTINGS_JSON",
	} {
		t.Setenv(key, "")
	}
}

func TestLoadSeedInputDefaults(t *testing.T) {
	clearSeedEnv(t)

	input, err := loadSeedInput("https://events.example.com")
	if err != nil {
		t.Fatalf("loadSeedInput returned error: %v", err)
	}

	if input.Slug != "demo" {
		t.Fatalf("expected default slug demo, got %q", input.Slug)
	}
	if input.Name != "Demo Tenant" {
		t.Fatalf("expected default name Demo Tenant, got %q", input.Name)
	}
	if input.PublicBaseURL != "https://events.example.com/demo" {
		t.Fatalf("expected derived public base url, got %q", input.PublicBaseURL)
	}
	if input.Settings.DefaultRetentionDays != 30 {
		t.Fatalf("expected default retention days 30, got %d", input.Settings.DefaultRetentionDays)
	}
	if input.Settings.PayPalMode != "disabled" {
		t.Fatalf("expected default paypal mode disabled, got %q", input.Settings.PayPalMode)
	}
}

func TestLoadSeedInputOverrides(t *testing.T) {
	clearSeedEnv(t)
	t.Setenv("EEP_SEED_TENANT_SLUG", "customer-xyz")
	t.Setenv("EEP_SEED_TENANT_NAME", "Customer XYZ")
	t.Setenv("EEP_SEED_TENANT_PUBLIC_BASE_URL", "https://events.example.com/customer-xyz")
	t.Setenv("EEP_SEED_RETENTION_DAYS", "45")
	t.Setenv("EEP_SEED_PAYPAL_MODE", "sandbox")
	t.Setenv("EEP_SEED_SENDER_EMAIL", "noreply@example.com")
	t.Setenv("EEP_SEED_ADMIN_EMAIL", "owner@example.com")
	t.Setenv("EEP_SEED_ADMIN_NAME", "Owner Example")
	t.Setenv("EEP_SEED_ADMIN_ROLE", "admin")
	t.Setenv("EEP_SEED_SETTINGS_JSON", `{"brand":"oak"}`)

	input, err := loadSeedInput("https://events.example.com")
	if err != nil {
		t.Fatalf("loadSeedInput returned error: %v", err)
	}

	if input.Slug != "customer-xyz" {
		t.Fatalf("expected slug override customer-xyz, got %q", input.Slug)
	}
	if input.Name != "Customer XYZ" {
		t.Fatalf("expected name override Customer XYZ, got %q", input.Name)
	}
	if input.Settings.DefaultRetentionDays != 45 {
		t.Fatalf("expected retention override 45, got %d", input.Settings.DefaultRetentionDays)
	}
	if input.Settings.PayPalMode != "sandbox" {
		t.Fatalf("expected paypal mode sandbox, got %q", input.Settings.PayPalMode)
	}
	if input.Settings.SenderEmail != "noreply@example.com" {
		t.Fatalf("expected sender email override, got %q", input.Settings.SenderEmail)
	}
	if input.AdminUser.Email != "owner@example.com" {
		t.Fatalf("expected admin email override, got %q", input.AdminUser.Email)
	}
	if input.AdminUser.Name != "Owner Example" {
		t.Fatalf("expected admin name override, got %q", input.AdminUser.Name)
	}
	if input.AdminUser.Role != "admin" {
		t.Fatalf("expected admin role override, got %q", input.AdminUser.Role)
	}
	if !strings.Contains(input.Settings.SettingsJSON, "brand") {
		t.Fatalf("expected settings json override, got %q", input.Settings.SettingsJSON)
	}
}

func TestLoadSeedInputRejectsInvalidRetentionDays(t *testing.T) {
	clearSeedEnv(t)
	t.Setenv("EEP_SEED_RETENTION_DAYS", "invalid")

	if _, err := loadSeedInput("https://events.example.com"); err == nil {
		t.Fatalf("expected error for invalid EEP_SEED_RETENTION_DAYS")
	}
}
