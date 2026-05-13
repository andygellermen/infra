package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/tenant"
)

var buildVersion = "dev"

func main() {
	cfg, err := config.Load(buildVersion)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	sqlDB, err := db.Open(cfg.DBDriver, cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	migrator := db.NewMigrator(sqlDB, migrations.Files, ".")
	if _, err := migrator.Up(context.Background()); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	seedInput, err := loadSeedInput(cfg.BaseURL)
	if err != nil {
		log.Fatalf("load seed input: %v", err)
	}

	repo := tenant.NewRepository(sqlDB)
	result, err := repo.SeedTenant(context.Background(), seedInput)
	if err != nil {
		log.Fatalf("seed tenant: %v", err)
	}

	action := "updated"
	if result.Created {
		action = "created"
	}
	log.Printf(
		"%s tenant slug=%s id=%s public_base_url=%s retention_days=%d paypal_mode=%s",
		action,
		result.Tenant.Slug,
		result.Tenant.ID,
		result.Tenant.PublicBaseURL,
		result.Settings.DefaultRetentionDays,
		result.Settings.PayPalMode,
	)
}

func loadSeedInput(baseURL string) (tenant.SeedInput, error) {
	slug := strings.TrimSpace(getenv("EEP_SEED_TENANT_SLUG", "demo"))
	if slug == "" {
		return tenant.SeedInput{}, fmt.Errorf("EEP_SEED_TENANT_SLUG must not be empty")
	}

	publicBaseURL := strings.TrimSpace(getenv("EEP_SEED_TENANT_PUBLIC_BASE_URL", ""))
	if publicBaseURL == "" {
		publicBaseURL = strings.TrimRight(baseURL, "/") + "/" + strings.Trim(slug, "/")
	}

	retentionDays, err := getenvInt("EEP_SEED_RETENTION_DAYS", tenant.DefaultRetentionDays)
	if err != nil {
		return tenant.SeedInput{}, err
	}

	return tenant.SeedInput{
		Slug:            slug,
		Name:            strings.TrimSpace(getenv("EEP_SEED_TENANT_NAME", "Demo Tenant")),
		PublicBaseURL:   publicBaseURL,
		DefaultTimezone: strings.TrimSpace(getenv("EEP_SEED_TENANT_TIMEZONE", tenant.DefaultTimezone)),
		DefaultLocale:   strings.TrimSpace(getenv("EEP_SEED_TENANT_LOCALE", tenant.DefaultLocale)),
		Status:          strings.TrimSpace(getenv("EEP_SEED_TENANT_STATUS", tenant.DefaultStatus)),
		Settings: tenant.TenantSettingsInput{
			SenderEmail:          strings.TrimSpace(getenv("EEP_SEED_SENDER_EMAIL", "")),
			SenderName:           strings.TrimSpace(getenv("EEP_SEED_SENDER_NAME", "")),
			PayPalMode:           strings.TrimSpace(getenv("EEP_SEED_PAYPAL_MODE", tenant.DefaultPayPalMode)),
			PayPalClientID:       strings.TrimSpace(getenv("EEP_SEED_PAYPAL_CLIENT_ID", "")),
			PayPalMerchantID:     strings.TrimSpace(getenv("EEP_SEED_PAYPAL_MERCHANT_ID", "")),
			DefaultRetentionDays: retentionDays,
			SettingsJSON:         strings.TrimSpace(getenv("EEP_SEED_SETTINGS_JSON", "")),
		},
	}, nil
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}
