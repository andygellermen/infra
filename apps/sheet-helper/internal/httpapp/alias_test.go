package httpapp

import (
	"testing"

	"github.com/andygellermann/infra/apps/sheet-helper/internal/config"
)

func TestCanonicalDomainUsesAliases(t *testing.T) {
	app := &App{
		cfg: config.Config{
			Tenants: map[string]config.TenantConfig{
				"geller.men": {
					Domain:  "geller.men",
					Aliases: []string{"www.geller.men"},
				},
			},
		},
	}

	if got := app.canonicalDomain("www.geller.men"); got != "geller.men" {
		t.Fatalf("expected alias to canonicalize to geller.men, got %q", got)
	}
	if got := app.canonicalDomain("geller.men"); got != "geller.men" {
		t.Fatalf("expected canonical domain to stay unchanged, got %q", got)
	}
}
