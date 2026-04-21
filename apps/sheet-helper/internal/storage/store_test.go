package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/andygellermann/infra/apps/sheet-helper/internal/model"
)

func TestLookupRouteNormalizesTrailingSlash(t *testing.T) {
	ctx := context.Background()
	store, err := Open(filepath.Join(t.TempDir(), "sheet-helper.db"))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer store.Close()

	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	if err := store.ReplaceAll(ctx, []model.Route{
		{
			Domain:  "geller.men",
			Path:    "/flyer/",
			Type:    model.RouteTypeLink,
			Target:  "https://example.org",
			Enabled: true,
		},
	}, nil, nil, nil); err != nil {
		t.Fatalf("ReplaceAll returned error: %v", err)
	}

	for _, candidate := range []string{"/flyer", "/flyer/"} {
		route, found, err := store.LookupRoute(ctx, "geller.men", candidate)
		if err != nil {
			t.Fatalf("LookupRoute(%q) returned error: %v", candidate, err)
		}
		if !found {
			t.Fatalf("expected route for %q", candidate)
		}
		if route.Path != "/flyer" {
			t.Fatalf("expected stored path to be normalized, got %q", route.Path)
		}
	}
}
