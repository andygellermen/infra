// Package rules owns versioned rule catalogues and their publication state.
package rules

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Catalogue struct {
	Version string       `json:"version"`
	Rules   []Definition `json:"rules"`
}

type Repository interface {
	Active(context.Context) (Catalogue, error)
}

type Service struct {
	repository Repository
}

func New(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Active(ctx context.Context) (Catalogue, error) {
	return s.repository.Active(ctx)
}

type PostgresRepository struct {
	database *sql.DB
}

func NewPostgresRepository(database *sql.DB) *PostgresRepository {
	return &PostgresRepository{database: database}
}

func (r *PostgresRepository) Active(ctx context.Context) (Catalogue, error) {
	if r.database == nil {
		return Catalogue{}, fmt.Errorf("rules database is nil")
	}
	var catalogue Catalogue
	if err := r.database.QueryRowContext(ctx,
		`SELECT version FROM rule_sets WHERE status = 'PRODUCTION'`,
	).Scan(&catalogue.Version); err != nil {
		return Catalogue{}, fmt.Errorf("load production rule set: %w", err)
	}
	rows, err := r.database.QueryContext(ctx, `
SELECT r.contract_version, r.id::text, r.rule_key, r.name, r.description,
       r.priority, r.enabled, r.scope, r.status, r.version, COALESCE(r.evidence_class, ''),
       r.source_keys, r.condition_tree, r.actions, r.confidence_modifier,
       r.stop_processing
FROM rules r
JOIN rule_set_rules rsr ON rsr.rule_id = r.id
JOIN rule_sets rs ON rs.id = rsr.rule_set_id
WHERE rs.status = 'PRODUCTION' AND r.enabled
ORDER BY r.priority DESC, r.rule_key`)
	if err != nil {
		return Catalogue{}, fmt.Errorf("load production rules: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var definition Definition
		var sourceKeys, condition, actions json.RawMessage
		if err := rows.Scan(
			&definition.ContractVersion, &definition.ID, &definition.Key, &definition.Name,
			&definition.Description, &definition.Priority, &definition.Enabled,
			&definition.Scope, &definition.Status, &definition.Version,
			&definition.EvidenceClass, &sourceKeys, &condition, &actions,
			&definition.ConfidenceModifier, &definition.StopProcessing,
		); err != nil {
			return Catalogue{}, fmt.Errorf("scan production rule: %w", err)
		}
		if err := json.Unmarshal(sourceKeys, &definition.SourceKeys); err != nil {
			return Catalogue{}, fmt.Errorf("decode rule %s sources: %w", definition.Key, err)
		}
		if err := json.Unmarshal(condition, &definition.Condition); err != nil {
			return Catalogue{}, fmt.Errorf("decode rule %s condition: %w", definition.Key, err)
		}
		if err := json.Unmarshal(actions, &definition.Actions); err != nil {
			return Catalogue{}, fmt.Errorf("decode rule %s actions: %w", definition.Key, err)
		}
		if err := definition.Validate(); err != nil {
			return Catalogue{}, fmt.Errorf("validate production rule: %w", err)
		}
		catalogue.Rules = append(catalogue.Rules, definition)
	}
	if err := rows.Err(); err != nil {
		return Catalogue{}, fmt.Errorf("iterate production rules: %w", err)
	}
	return catalogue, nil
}
