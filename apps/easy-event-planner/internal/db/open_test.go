package db

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenSQLiteCreatesParentDirectory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "data", "easy-event-planner.sqlite")

	sqlDB, err := Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := sqlDB.PingContext(context.Background()); err != nil {
		t.Fatalf("database ping failed: %v", err)
	}
}

func TestOpenSQLiteAppliesPragmas(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "eep.sqlite")

	sqlDB, err := Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	var foreignKeys int
	if err := sqlDB.QueryRow(`PRAGMA foreign_keys;`).Scan(&foreignKeys); err != nil {
		t.Fatalf("query foreign_keys pragma: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("expected foreign_keys pragma to be 1, got %d", foreignKeys)
	}

	var journalMode string
	if err := sqlDB.QueryRow(`PRAGMA journal_mode;`).Scan(&journalMode); err != nil {
		t.Fatalf("query journal_mode pragma: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("expected journal_mode wal, got %q", journalMode)
	}
}
