package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTenantsFromDir(t *testing.T) {
	dir := t.TempDir()
	content := `
STATIC_EDITOR_TENANT=example.org
STATIC_EDITOR_LOGIN_DOMAIN=bearbeitung.example.org
STATIC_EDITOR_ALIASES=www.example.org
STATIC_EDITOR_STATIC_ROOT=/srv/static/example.org
STATIC_EDITOR_BACKUP_ROOT=/srv/static-backups/example.org
STATIC_EDITOR_REPO_ROOT=/srv/static/example.org
STATIC_EDITOR_USERNAME=admin
STATIC_EDITOR_PASSWORD_HASH=$2y$example
STATIC_EDITOR_COOKIE_SECRET=secret-1
STATIC_EDITOR_MAIN_SELECTOR=main
STATIC_EDITOR_ALLOWED_BLOCK_TAGS=h1,h2,p
STATIC_EDITOR_ALLOWED_INLINE_TAGS=strong,em
STATIC_EDITOR_START_PATH=/index.html
`
	if err := os.WriteFile(filepath.Join(dir, "example.org.env"), []byte(content), 0o644); err != nil {
		t.Fatalf("write tenant env: %v", err)
	}

	tenants, err := loadTenantsFromDir(dir)
	if err != nil {
		t.Fatalf("loadTenantsFromDir returned error: %v", err)
	}

	tenant, ok := tenants["example.org"]
	if !ok {
		t.Fatalf("expected tenant example.org to be present")
	}
	if tenant.LoginDomain != "bearbeitung.example.org" {
		t.Fatalf("unexpected login domain %q", tenant.LoginDomain)
	}
	if tenant.MainSelector != "main" {
		t.Fatalf("unexpected main selector %q", tenant.MainSelector)
	}
	if got := len(tenant.AllowedBlockTags); got != 3 {
		t.Fatalf("expected 3 allowed block tags, got %d", got)
	}
}
