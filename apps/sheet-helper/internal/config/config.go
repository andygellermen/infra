package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Addr         string
	DBPath       string
	SeedFile     string
	CookieSecret string
}

func Load() (Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("getwd: %w", err)
	}

	cfg := Config{
		Addr:         getenv("SHEET_HELPER_ADDR", ":8080"),
		DBPath:       getenv("SHEET_HELPER_DB_PATH", filepath.Join(cwd, "sheet-helper.db")),
		SeedFile:     getenv("SHEET_HELPER_SEED_FILE", filepath.Join(cwd, "testdata", "seed.json")),
		CookieSecret: getenv("SHEET_HELPER_COOKIE_SECRET", "dev-only-change-me"),
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
