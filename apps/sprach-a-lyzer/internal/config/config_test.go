package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("SAL_ENV", "")
	t.Setenv("SAL_HTTP_ADDR", "")
	t.Setenv("SAL_DATABASE_URL", "")
	t.Setenv("SAL_MAX_REQUEST_BYTES", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.MaxRequestBytes != 64<<10 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestValidateDatabase(t *testing.T) {
	t.Parallel()

	if err := (Config{}).ValidateDatabase(); err == nil {
		t.Fatal("empty database URL unexpectedly accepted")
	}
	if err := (Config{DatabaseURL: "postgres://localhost/sprachalyzer"}).ValidateDatabase(); err != nil {
		t.Fatalf("valid database URL rejected: %v", err)
	}
}
