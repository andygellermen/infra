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
	defaultTokenPepper       = "dev-only-change-me"
	defaultSessionCookieName = "eep_session"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
	defaultSessionTTL        = 12 * time.Hour
	defaultMagicLinkTTL      = 15 * time.Minute
	defaultRegistrationTTL   = 30 * time.Minute
	defaultWaitlistTTL       = 24 * time.Hour
	defaultCertificateTTL    = 30 * time.Minute
	defaultRateLimitWindow   = 15 * time.Minute
	defaultRateLimitRequests = 5
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
	TokenPepper       string
	SessionCookieName string
	SecureCookies     bool
	SessionTTL        time.Duration
	MagicLinkTTL      time.Duration
	RegistrationTTL   time.Duration
	WaitlistOfferTTL  time.Duration
	CertificateTTL    time.Duration
	AuthRateLimit     int
	AuthRateWindow    time.Duration
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
		TokenPepper:       strings.TrimSpace(getenv("EEP_TOKEN_PEPPER", defaultTokenPepper)),
		SessionCookieName: strings.TrimSpace(getenv("EEP_SESSION_COOKIE_NAME", defaultSessionCookieName)),
		SecureCookies:     getenvBool("EEP_SECURE_COOKIES", strings.EqualFold(strings.TrimSpace(getenv("EEP_ENV", defaultEnv)), "production")),
		ReadHeaderTimeout: defaultReadHeaderTimeout,
		ShutdownTimeout:   defaultShutdownTimeout,
		SessionTTL:        defaultSessionTTL,
		MagicLinkTTL:      defaultMagicLinkTTL,
		RegistrationTTL:   defaultRegistrationTTL,
		WaitlistOfferTTL:  defaultWaitlistTTL,
		CertificateTTL:    defaultCertificateTTL,
		AuthRateLimit:     defaultRateLimitRequests,
		AuthRateWindow:    defaultRateLimitWindow,
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
	if cfg.TokenPepper == "" {
		return Config{}, fmt.Errorf("EEP_TOKEN_PEPPER must not be empty")
	}
	if cfg.SessionCookieName == "" {
		return Config{}, fmt.Errorf("EEP_SESSION_COOKIE_NAME must not be empty")
	}

	cfg.ReadHeaderTimeout, err = parseDurationEnv("EEP_HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.ShutdownTimeout, err = parseDurationEnv("EEP_HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.SessionTTL, err = parseDurationEnv("EEP_SESSION_TTL", defaultSessionTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.MagicLinkTTL, err = parseDurationEnv("EEP_MAGIC_LINK_TTL", defaultMagicLinkTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.RegistrationTTL, err = parseDurationEnv("EEP_REGISTRATION_MAGIC_LINK_TTL", defaultRegistrationTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.WaitlistOfferTTL, err = parseDurationEnv("EEP_WAITLIST_OFFER_TTL", defaultWaitlistTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.CertificateTTL, err = parseDurationEnv("EEP_CERTIFICATE_ACCESS_TTL", defaultCertificateTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.AuthRateWindow, err = parseDurationEnv("EEP_AUTH_RATE_LIMIT_WINDOW", defaultRateLimitWindow)
	if err != nil {
		return Config{}, err
	}
	cfg.AuthRateLimit, err = parseIntEnv("EEP_AUTH_RATE_LIMIT_REQUESTS", defaultRateLimitRequests)
	if err != nil {
		return Config{}, err
	}
	if cfg.AuthRateLimit <= 0 {
		return Config{}, fmt.Errorf("EEP_AUTH_RATE_LIMIT_REQUESTS must be > 0")
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

func parseIntEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
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
