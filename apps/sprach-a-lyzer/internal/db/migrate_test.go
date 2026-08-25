package db

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadMigrationsSortsAndChecksums(t *testing.T) {
	t.Parallel()

	migrations, err := loadMigrations(fstest.MapFS{
		"0002_second.sql": {Data: []byte("SELECT 2;")},
		"0001_first.sql":  {Data: []byte("SELECT 1;")},
		"README.md":       {Data: []byte("ignored")},
	}, ".")
	if err != nil {
		t.Fatalf("loadMigrations() error: %v", err)
	}
	if len(migrations) != 2 || migrations[0].Version != 1 || migrations[1].Version != 2 {
		t.Fatalf("unexpected migrations: %+v", migrations)
	}
	if len(migrations[0].Checksum) != 64 {
		t.Fatalf("checksum length = %d; want 64", len(migrations[0].Checksum))
	}
}

func TestLoadMigrationsRejectsDuplicateVersion(t *testing.T) {
	t.Parallel()

	_, err := loadMigrations(fstest.MapFS{
		"0001_first.sql": {Data: []byte("SELECT 1;")},
		"01_second.sql":  {Data: []byte("SELECT 2;")},
	}, ".")
	if err == nil || !strings.Contains(err.Error(), "duplicate migration version") {
		t.Fatalf("loadMigrations() error = %v; want duplicate version", err)
	}
}
