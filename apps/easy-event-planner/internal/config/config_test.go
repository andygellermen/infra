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
