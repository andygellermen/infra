package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Addr            string
	DBPath          string
	SeedFile        string
	CookieSecret    string
	SyncToken       string
	Tenant          string
	SheetID         string
	PublishedURL    string
	RoutesSheet     string
	VCardsSheet     string
	TextsSheet      string
	DefaultListPref string
	StartupSync     bool
}

func Load() (Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("getwd: %w", err)
	}

	cfg := Config{
		Addr:            getenv("SHEET_HELPER_ADDR", ":8080"),
		DBPath:          getenv("SHEET_HELPER_DB_PATH", filepath.Join(cwd, "sheet-helper.db")),
		SeedFile:        getenv("SHEET_HELPER_SEED_FILE", filepath.Join(cwd, "testdata", "seed.json")),
		CookieSecret:    getenv("SHEET_HELPER_COOKIE_SECRET", "dev-only-change-me"),
		SyncToken:       getenv("SHEET_HELPER_SYNC_TOKEN", ""),
		Tenant:          getenv("SHEET_HELPER_TENANT", "localhost"),
		SheetID:         getenv("SHEET_HELPER_SHEET_ID", ""),
		PublishedURL:    getenv("SHEET_HELPER_PUBLISHED_URL", ""),
		RoutesSheet:     getenv("SHEET_HELPER_ROUTES_SHEET", "routes"),
		VCardsSheet:     getenv("SHEET_HELPER_VCARDS_SHEET", "vcard_entries"),
		TextsSheet:      getenv("SHEET_HELPER_TEXTS_SHEET", "text_entries"),
		DefaultListPref: getenv("SHEET_HELPER_LIST_PREFIX", "list_"),
		StartupSync:     getenvBool("SHEET_HELPER_STARTUP_SYNC", false),
	}

	if cfg.SheetID != "" {
		cfg.SeedFile = strings.TrimSpace(getenv("SHEET_HELPER_SEED_FILE", ""))
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvBool(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}
