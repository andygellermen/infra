package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const openTimeout = 5 * time.Second

const RequiredSchemaVersion int64 = 3

func Open(ctx context.Context, databaseURL string) (*sql.DB, error) {
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	database.SetMaxOpenConns(10)
	database.SetMaxIdleConns(5)
	database.SetConnMaxIdleTime(5 * time.Minute)
	database.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, openTimeout)
	defer cancel()
	if err := database.PingContext(pingCtx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return database, nil
}

type SchemaPinger struct {
	database        *sql.DB
	requiredVersion int64
}

func NewSchemaPinger(database *sql.DB, requiredVersion int64) SchemaPinger {
	return SchemaPinger{database: database, requiredVersion: requiredVersion}
}

func (p SchemaPinger) PingContext(ctx context.Context) error {
	if p.database == nil {
		return fmt.Errorf("database is nil")
	}
	if err := p.database.PingContext(ctx); err != nil {
		return err
	}
	var version sql.NullInt64
	if err := p.database.QueryRowContext(ctx, `SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil {
		return fmt.Errorf("query schema version: %w", err)
	}
	if !version.Valid || version.Int64 < p.requiredVersion {
		return fmt.Errorf("schema version %d is below required version %d", version.Int64, p.requiredVersion)
	}
	return nil
}
