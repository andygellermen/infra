package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultEnv               = "development"
	defaultAddr              = ":8080"
	defaultBaseURL           = "http://localhost:8080"
	defaultDBDriver          = "sqlite"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultVersion           = "dev"
	appName                  = "easy-event-planner"
)

type Config struct {
	AppName           string
	Env               string
	Addr              string
	BaseURL           string
	Version           string
	DBDriver          string
	DBPath            string
	ReadHeaderTimeout time.Duration
	ShutdownTimeout   time.Duration
}

func Load(buildVersion string) (Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, fmt.Errorf("getwd: %w", err)
	}

	cfg := Config{
		AppName:           appName,
		Env:               strings.TrimSpace(getenv("EEP_ENV", defaultEnv)),
		Addr:              strings.TrimSpace(getenv("EEP_HTTP_ADDR", defaultAddr)),
		BaseURL:           strings.TrimSpace(getenv("EEP_BASE_URL", defaultBaseURL)),
		Version:           resolveVersion(buildVersion),
		DBDriver:          strings.ToLower(strings.TrimSpace(getenv("EEP_DB_DRIVER", defaultDBDriver))),
		DBPath:            strings.TrimSpace(getenv("EEP_DB_PATH", filepath.Join(cwd, "data", "easy-event-planner.sqlite"))),
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		ShutdownTimeout:   defaultShutdownTimeout,
	}

	if cfg.Env == "" {
		return Config{}, fmt.Errorf("EEP_ENV must not be empty")
	}
	if cfg.Addr == "" {
		return Config{}, fmt.Errorf("EEP_HTTP_ADDR must not be empty")
	}
	if cfg.BaseURL == "" {
		return Config{}, fmt.Errorf("EEP_BASE_URL must not be empty")
	}
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return Config{}, err
	}
	if cfg.DBPath == "" {
		return Config{}, fmt.Errorf("EEP_DB_PATH must not be empty")
	}
	if err := validateDBDriver(cfg.DBDriver); err != nil {
		return Config{}, err
	}

	cfg.ReadHeaderTimeout, err = parseDurationEnv("EEP_HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.ShutdownTimeout, err = parseDurationEnv("EEP_HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func resolveVersion(buildVersion string) string {
	fromEnv := strings.TrimSpace(os.Getenv("EEP_VERSION"))
	if fromEnv != "" {
		return fromEnv
	}
	if strings.TrimSpace(buildVersion) != "" {
		return strings.TrimSpace(buildVersion)
	}
	return defaultVersion
}

func parseDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be > 0", key)
	}
	return parsed, nil
}

func validateBaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse EEP_BASE_URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("EEP_BASE_URL must include scheme and host")
	}
	return nil
}

func validateDBDriver(driver string) error {
	switch driver {
	case "sqlite", "postgres":
		return nil
	default:
		return fmt.Errorf("unsupported EEP_DB_DRIVER %q", driver)
	}
}

func getenv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
