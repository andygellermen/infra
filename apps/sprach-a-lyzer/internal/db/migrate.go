package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	migrationLockID   = 83920471
	migrationTableDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version BIGINT PRIMARY KEY,
  name TEXT NOT NULL,
  checksum TEXT NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`
)

var migrationFileNamePattern = regexp.MustCompile(`^(\d+)_([a-z0-9_]+)\.sql$`)

type Migration struct {
	Version  int64
	Name     string
	File     string
	SQL      string
	Checksum string
}

type MigrationResult struct {
	Applied []string `json:"applied"`
	Total   int      `json:"total"`
}

type Migrator struct {
	db   *sql.DB
	fsys fs.FS
	dir  string
}

func NewMigrator(database *sql.DB, migrationFS fs.FS, migrationDir string) *Migrator {
	return &Migrator{db: database, fsys: migrationFS, dir: migrationDir}
}

func (m *Migrator) Up(ctx context.Context) (MigrationResult, error) {
	if m.db == nil {
		return MigrationResult{}, fmt.Errorf("migrator database is nil")
	}
	migrations, err := loadMigrations(m.fsys, m.dir)
	if err != nil {
		return MigrationResult{}, err
	}

	connection, err := m.db.Conn(ctx)
	if err != nil {
		return MigrationResult{}, fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Close()

	if _, err := connection.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, migrationLockID); err != nil {
		return MigrationResult{}, fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = connection.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, migrationLockID)
	}()

	if _, err := connection.ExecContext(ctx, migrationTableDDL); err != nil {
		return MigrationResult{}, fmt.Errorf("ensure schema_migrations: %w", err)
	}
	applied, err := fetchApplied(ctx, connection)
	if err != nil {
		return MigrationResult{}, err
	}

	result := MigrationResult{Total: len(migrations), Applied: []string{}}
	for _, migration := range migrations {
		if checksum, ok := applied[migration.Version]; ok {
			if checksum != migration.Checksum {
				return MigrationResult{}, fmt.Errorf("migration %s changed after application", migration.File)
			}
			continue
		}
		if err := applyMigration(ctx, connection, migration); err != nil {
			return MigrationResult{}, err
		}
		result.Applied = append(result.Applied, migration.File)
	}
	return result, nil
}

func fetchApplied(ctx context.Context, connection *sql.Conn) (map[int64]string, error) {
	rows, err := connection.QueryContext(ctx, `SELECT version, checksum FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	result := make(map[int64]string)
	for rows.Next() {
		var version int64
		var checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		result[version] = checksum
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}
	return result, nil
}

func applyMigration(ctx context.Context, connection *sql.Conn, migration Migration) error {
	tx, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", migration.File, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("execute migration %s: %w", migration.File, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations(version, name, checksum) VALUES($1, $2, $3)`,
		migration.Version, migration.Name, migration.Checksum,
	); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.File, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", migration.File, err)
	}
	return nil
}

func loadMigrations(migrationFS fs.FS, migrationDir string) ([]Migration, error) {
	entries, err := fs.ReadDir(migrationFS, migrationDir)
	if err != nil {
		return nil, fmt.Errorf("read migration directory %q: %w", migrationDir, err)
	}

	seen := make(map[int64]string)
	result := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFileNamePattern.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue
		}
		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", entry.Name(), err)
		}
		if existing, ok := seen[version]; ok {
			return nil, fmt.Errorf("duplicate migration version %d in %q and %q", version, existing, entry.Name())
		}
		seen[version] = entry.Name()

		path := entry.Name()
		if migrationDir != "." {
			path = strings.TrimRight(migrationDir, "/") + "/" + entry.Name()
		}
		content, err := fs.ReadFile(migrationFS, path)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", path, err)
		}
		digest := sha256.Sum256(content)
		result = append(result, Migration{
			Version: version, Name: matches[2], File: entry.Name(), SQL: string(content),
			Checksum: hex.EncodeToString(digest[:]),
		})
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Version < result[j].Version })
	return result, nil
}
