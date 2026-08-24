// Package presentation owns profile-specific labels and safe fallbacks.
package presentation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Bundle struct {
	Profile   string            `json:"profile"`
	Locale    string            `json:"locale"`
	Version   string            `json:"version"`
	Entries   map[string]string `json:"entries"`
	Fallbacks map[string]string `json:"fallbacks"`
}

func (b Bundle) Resolve(canonicalKey string) string {
	if value := b.Entries[canonicalKey]; value != "" {
		return value
	}
	if value := b.Fallbacks[canonicalKey]; value != "" {
		return value
	}
	if value := b.Fallbacks["UNKNOWN_CONCEPT"]; value != "" {
		return value
	}
	return "Sprachmuster"
}

type Repository interface {
	Load(context.Context, string, string) (Bundle, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Bundle(ctx context.Context, profile, locale string) (Bundle, error) {
	return s.repository.Load(ctx, profile, locale)
}

type PostgresRepository struct {
	database *sql.DB
}

func NewPostgresRepository(database *sql.DB) *PostgresRepository {
	return &PostgresRepository{database: database}
}

func (r *PostgresRepository) Load(ctx context.Context, profile, locale string) (Bundle, error) {
	if r.database == nil {
		return Bundle{}, fmt.Errorf("presentation database is nil")
	}
	bundle := Bundle{Profile: profile, Locale: locale, Entries: map[string]string{}}
	var bundleID string
	var fallbacks json.RawMessage
	err := r.database.QueryRowContext(ctx, `
SELECT id, version, fallbacks
FROM presentation_bundles
WHERE profile = $1 AND locale = $2 AND status = 'PRODUCTION'
ORDER BY created_at DESC
LIMIT 1`, profile, locale).Scan(&bundleID, &bundle.Version, &fallbacks)
	if err != nil {
		return Bundle{}, fmt.Errorf("load presentation bundle %s/%s: %w", profile, locale, err)
	}
	if err := json.Unmarshal(fallbacks, &bundle.Fallbacks); err != nil {
		return Bundle{}, fmt.Errorf("decode presentation fallbacks: %w", err)
	}
	rows, err := r.database.QueryContext(ctx, `
SELECT canonical_key, display_value
FROM presentation_entries
WHERE bundle_id = $1`, bundleID)
	if err != nil {
		return Bundle{}, fmt.Errorf("load presentation entries: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return Bundle{}, fmt.Errorf("scan presentation entry: %w", err)
		}
		bundle.Entries[key] = value
	}
	if err := rows.Err(); err != nil {
		return Bundle{}, fmt.Errorf("iterate presentation entries: %w", err)
	}
	return bundle, nil
}
