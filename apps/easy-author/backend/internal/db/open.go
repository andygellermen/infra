package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	defaultSQLiteBusyTimeout = 5000
	openTimeout              = 5 * time.Second
)

func OpenSQLite(dsn string) (*sql.DB, error) {
	cleanDSN := strings.TrimSpace(dsn)
	if cleanDSN == "" {
		return nil, fmt.Errorf("sqlite dsn must not be empty")
	}

	if err := ensureSQLiteParentDir(cleanDSN); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", cleanDSN)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), openTimeout)
	defer cancel()

	if err := configureSQLite(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return db, nil
}

func configureSQLite(ctx context.Context, db *sql.DB) error {
	for _, statement := range []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		fmt.Sprintf("PRAGMA busy_timeout = %d;", defaultSQLiteBusyTimeout),
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("exec %q: %w", statement, err)
		}
	}
	return nil
}

func ensureSQLiteParentDir(dsn string) error {
	if dsn == ":memory:" || strings.HasPrefix(dsn, "file:") {
		return nil
	}

	parent := filepath.Dir(dsn)
	if parent == "." || parent == "" {
		return nil
	}

	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("mkdir sqlite parent %s: %w", parent, err)
	}
	return nil
}
