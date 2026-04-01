package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTenantsFromDir(t *testing.T) {
	dir := t.TempDir()
	content := `
SHEET_HELPER_TENANT=geller.men
SHEET_HELPER_ALIASES=www.geller.men, gellermen.de
SHEET_HELPER_COOKIE_SECRET=secret-1
SHEET_HELPER_SYNC_TOKEN=sync-1
SHEET_HELPER_SHEET_ID=sheet-1
SHEET_HELPER_PUBLISHED_URL=https://example.invalid/pubhtml
SHEET_HELPER_ROUTES_SHEET=routes
SHEET_HELPER_VCARDS_SHEET=vcard_entries
SHEET_HELPER_TEXTS_SHEET=text_entries
SHEET_HELPER_LIST_PREFIX=list_
SHEET_HELPER_STARTUP_SYNC=true
`
	if err := os.WriteFile(filepath.Join(dir, "geller.men.env"), []byte(content), 0o644); err != nil {
		t.Fatalf("write tenant env: %v", err)
	}

	tenants, err := loadTenantsFromDir(dir)
	if err != nil {
		t.Fatalf("loadTenantsFromDir returned error: %v", err)
	}

	tenant, ok := tenants["geller.men"]
	if !ok {
		t.Fatalf("expected tenant geller.men to be present")
	}
	if tenant.CookieSecret != "secret-1" {
		t.Fatalf("expected cookie secret to be parsed, got %q", tenant.CookieSecret)
	}
	if got := len(tenant.Aliases); got != 2 {
		t.Fatalf("expected 2 aliases, got %d", got)
	}
	if !tenant.StartupSync {
		t.Fatalf("expected startup sync to be true")
	}
}

func TestLoadIgnoresSeedFileInTenantDirMode(t *testing.T) {
	dir := t.TempDir()
	content := `
SHEET_HELPER_TENANT=geller.men
SHEET_HELPER_COOKIE_SECRET=secret-1
SHEET_HELPER_SYNC_TOKEN=s0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
SHEET_HELPER_SHEET_ID=sheet-1
SHEET_HELPER_PUBLISHED_URL=https://example.invalid/pubhtml
`
	if err := os.WriteFile(filepath.Join(dir, "geller.men.env"), []byte(content), 0o644); err != nil {
		t.Fatalf("write tenant env: %v", err)
	}

	t.Setenv("SHEET_HELPER_TENANT_DIR", dir)
	t.Setenv("SHEET_HELPER_SEED_FILE", "/app/testdata/seed.json")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.SeedFile != "" {
		t.Fatalf("expected seed file to be disabled in tenant dir mode, got %q", cfg.SeedFile)
	}
}
