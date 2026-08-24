package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment     string
	HTTPAddr        string
	DatabaseURL     string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MaxRequestBytes int64
}

func Load() (Config, error) {
	cfg := Config{
		Environment:     valueOrDefault("SAL_ENV", "development"),
		HTTPAddr:        valueOrDefault("SAL_HTTP_ADDR", ":8080"),
		DatabaseURL:     strings.TrimSpace(os.Getenv("SAL_DATABASE_URL")),
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    15 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		MaxRequestBytes: 64 << 10,
	}

	if value := strings.TrimSpace(os.Getenv("SAL_MAX_REQUEST_BYTES")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed < 1024 {
			return Config{}, fmt.Errorf("SAL_MAX_REQUEST_BYTES must be an integer of at least 1024")
		}
		cfg.MaxRequestBytes = parsed
	}
	return cfg, nil
}

func (c Config) ValidateDatabase() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("SAL_DATABASE_URL must not be empty")
	}
	if !strings.HasPrefix(c.DatabaseURL, "postgres://") && !strings.HasPrefix(c.DatabaseURL, "postgresql://") {
		return fmt.Errorf("SAL_DATABASE_URL must use postgres:// or postgresql://")
	}
	return nil
}

func valueOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
