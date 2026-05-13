package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const migrationTableDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at TEXT NOT NULL
);`

var migrationFileNamePattern = regexp.MustCompile(`^(\d+)_([a-z0-9_]+)\.sql$`)

type Migration struct {
	Version int
	Name    string
	File    string
	SQL     string
}

type MigrationResult struct {
	Applied []string
	Total   int
}

type Migrator struct {
	db   *sql.DB
	fsys fs.FS
	dir  string
	now  func() time.Time
}

func NewMigrator(sqlDB *sql.DB, migrationFS fs.FS, migrationDir string) *Migrator {
	return &Migrator{
		db:   sqlDB,
		fsys: migrationFS,
		dir:  migrationDir,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (m *Migrator) Up(ctx context.Context) (MigrationResult, error) {
	if m.db == nil {
		return MigrationResult{}, fmt.Errorf("migrator database is nil")
	}

	if _, err := m.db.ExecContext(ctx, migrationTableDDL); err != nil {
		return MigrationResult{}, fmt.Errorf("ensure schema_migrations: %w", err)
	}

	migrations, err := loadMigrations(m.fsys, m.dir)
	if err != nil {
		return MigrationResult{}, err
	}

	appliedVersions, err := m.fetchAppliedVersions(ctx)
	if err != nil {
		return MigrationResult{}, err
	}

	result := MigrationResult{Total: len(migrations)}
	for _, migration := range migrations {
		if _, ok := appliedVersions[migration.Version]; ok {
			continue
		}

		if err := m.applyMigration(ctx, migration); err != nil {
			return MigrationResult{}, err
		}
		result.Applied = append(result.Applied, migration.File)
	}

	return result, nil
}

func (m *Migrator) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.File, err)
	}

	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("exec migration %s: %w", migration.File, err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO schema_migrations(version, name, applied_at) VALUES(?, ?, ?)`,
		migration.Version,
		migration.Name,
		m.now().Format(time.RFC3339),
	); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", migration.File, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.File, err)
	}

	return nil
}

func (m *Migrator) fetchAppliedVersions(ctx context.Context) (map[int]struct{}, error) {
	rows, err := m.db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]struct{})
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration version: %w", err)
		}
		applied[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return applied, nil
}

func loadMigrations(migrationFS fs.FS, migrationDir string) ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFS, migrationDir)
	if err != nil {
		return nil, fmt.Errorf("read migration dir %q: %w", migrationDir, err)
	}

	var migrations []Migration
	seenVersions := make(map[int]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		matches := migrationFileNamePattern.FindStringSubmatch(name)
		if len(matches) != 3 {
			continue
		}

		version, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", name, err)
		}

		if existing, ok := seenVersions[version]; ok {
			return nil, fmt.Errorf("duplicate migration version %d in %q and %q", version, existing, name)
		}
		seenVersions[version] = name

		sqlPath := name
		if migrationDir != "." {
			sqlPath = strings.TrimRight(migrationDir, "/") + "/" + name
		}
		content, err := fs.ReadFile(migrationFS, sqlPath)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", sqlPath, err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    matches[2],
			File:    name,
			SQL:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}
