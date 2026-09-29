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

func TestReplaceTenantPreservesOtherTenantAndIsolatesLists(t *testing.T) {
	ctx := context.Background()
	store, err := Open(filepath.Join(t.TempDir(), "sheet-helper.db"))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer store.Close()

	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	for _, tenant := range []struct {
		domain string
		path   string
		label  string
	}{
		{domain: "first.example", path: "/first", label: "First tenant"},
		{domain: "second.example", path: "/second", label: "Second tenant"},
	} {
		err := store.ReplaceTenant(ctx, tenant.domain,
			[]model.Route{{Path: tenant.path, Type: model.RouteTypeList, ListSheet: "links", Enabled: true}},
			nil,
			nil,
			[]model.ListItem{{SheetName: "links", Label: tenant.label, Enabled: true}},
		)
		if err != nil {
			t.Fatalf("ReplaceTenant(%q) returned error: %v", tenant.domain, err)
		}
	}

	for _, want := range []struct {
		domain string
		path   string
		label  string
	}{
		{domain: "first.example", path: "/first", label: "First tenant"},
		{domain: "second.example", path: "/second", label: "Second tenant"},
	} {
		if _, found, err := store.LookupRoute(ctx, want.domain, want.path); err != nil {
			t.Fatalf("LookupRoute(%q) returned error: %v", want.domain, err)
		} else if !found {
			t.Fatalf("expected route for %q to survive the other tenant sync", want.domain)
		}

		items, err := store.ListItems(ctx, want.domain, "links")
		if err != nil {
			t.Fatalf("ListItems(%q) returned error: %v", want.domain, err)
		}
		if len(items) != 1 || items[0].Label != want.label {
			t.Fatalf("ListItems(%q) = %#v, want one item labeled %q", want.domain, items, want.label)
		}
	}

	if err := store.ReplaceTenant(ctx, "first.example",
		[]model.Route{{Path: "/replacement", Type: model.RouteTypeLink, Target: "https://example.org", Enabled: true}},
		nil, nil, nil,
	); err != nil {
		t.Fatalf("second ReplaceTenant returned error: %v", err)
	}

	if _, found, err := store.LookupRoute(ctx, "first.example", "/first"); err != nil {
		t.Fatalf("LookupRoute for stale route returned error: %v", err)
	} else if found {
		t.Fatal("expected stale route of replaced tenant to be removed")
	}
	if _, found, err := store.LookupRoute(ctx, "second.example", "/second"); err != nil {
		t.Fatalf("LookupRoute for other tenant returned error: %v", err)
	} else if !found {
		t.Fatal("expected route of other tenant to remain after replacement")
	}
}

func TestInitSchemaMigratesExistingListItemsToTenantSchema(t *testing.T) {
	ctx := context.Background()
	store, err := Open(filepath.Join(t.TempDir(), "sheet-helper.db"))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer store.Close()

	if _, err := store.db.ExecContext(ctx, `CREATE TABLE list_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sheet_name TEXT NOT NULL,
		sort_order INTEGER NOT NULL DEFAULT 0,
		label TEXT NOT NULL DEFAULT '',
		url TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT '',
		password TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1
	)`); err != nil {
		t.Fatalf("create legacy list_items table: %v", err)
	}

	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("InitSchema returned error for legacy schema: %v", err)
	}
	hasDomain, err := store.tableHasColumn(ctx, "list_items", "domain")
	if err != nil {
		t.Fatalf("tableHasColumn returned error: %v", err)
	}
	if !hasDomain {
		t.Fatal("expected InitSchema to add domain to legacy list_items table")
	}
}
