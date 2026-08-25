// Package knowledge owns the canonical linguistic knowledge catalogue.
package knowledge

import (
	"context"
	"database/sql"
	"fmt"
)

type Snapshot struct {
	Dimensions int `json:"dimensions"`
	Lexemes    int `json:"lexemes"`
	Senses     int `json:"senses"`
	Phrases    int `json:"phrases"`
}

type Repository interface {
	Snapshot(context.Context) (Snapshot, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Snapshot(ctx context.Context) (Snapshot, error) {
	return s.repository.Snapshot(ctx)
}

type PostgresRepository struct {
	database *sql.DB
}

func NewPostgresRepository(database *sql.DB) *PostgresRepository {
	return &PostgresRepository{database: database}
}

func (r *PostgresRepository) Snapshot(ctx context.Context) (Snapshot, error) {
	if r.database == nil {
		return Snapshot{}, fmt.Errorf("knowledge database is nil")
	}
	var snapshot Snapshot
	err := r.database.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*) FROM dimensions WHERE active),
  (SELECT COUNT(*) FROM lexemes WHERE status <> 'DEPRECATED'),
  (SELECT COUNT(*) FROM senses WHERE status <> 'DEPRECATED'),
  (SELECT COUNT(*) FROM phrases WHERE status <> 'DEPRECATED')`).Scan(
		&snapshot.Dimensions, &snapshot.Lexemes, &snapshot.Senses, &snapshot.Phrases,
	)
	if err != nil {
		return Snapshot{}, fmt.Errorf("load knowledge snapshot: %w", err)
	}
	return snapshot, nil
}
