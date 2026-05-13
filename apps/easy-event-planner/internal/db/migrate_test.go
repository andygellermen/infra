package db

import (
	"context"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestMigratorUpAppliesAllMigrations(t *testing.T) {
	sqlDB, err := Open("sqlite", filepath.Join(t.TempDir(), "migrate.sqlite"))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	migrationFS := fstest.MapFS{
		"0001_create_widgets.sql": {
			Data: []byte(`CREATE TABLE widgets (id TEXT PRIMARY KEY, title TEXT NOT NULL);`),
		},
		"0002_add_index.sql": {
			Data: []byte(`CREATE INDEX idx_widgets_title ON widgets(title);`),
		},
	}

	migrator := NewMigrator(sqlDB, migrationFS, ".")
	result, err := migrator.Up(context.Background())
	if err != nil {
		t.Fatalf("Up returned error: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("expected total migrations 2, got %d", result.Total)
	}
	if len(result.Applied) != 2 {
		t.Fatalf("expected applied migrations 2, got %d", len(result.Applied))
	}

	var count int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected schema_migrations count 2, got %d", count)
	}
}

func TestMigratorUpIsIdempotent(t *testing.T) {
	sqlDB, err := Open("sqlite", filepath.Join(t.TempDir(), "migrate.sqlite"))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	migrationFS := fstest.MapFS{
		"0001_create_widgets.sql": {
			Data: []byte(`CREATE TABLE widgets (id TEXT PRIMARY KEY);`),
		},
	}
	migrator := NewMigrator(sqlDB, migrationFS, ".")

	first, err := migrator.Up(context.Background())
	if err != nil {
		t.Fatalf("first Up returned error: %v", err)
	}
	if len(first.Applied) != 1 {
		t.Fatalf("expected first run to apply one migration, got %d", len(first.Applied))
	}

	second, err := migrator.Up(context.Background())
	if err != nil {
		t.Fatalf("second Up returned error: %v", err)
	}
	if len(second.Applied) != 0 {
		t.Fatalf("expected second run to apply no migrations, got %d", len(second.Applied))
	}
}
