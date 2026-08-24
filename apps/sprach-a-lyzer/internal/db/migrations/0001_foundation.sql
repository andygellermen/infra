CREATE TABLE dimensions (
  dimension_id TEXT PRIMARY KEY CHECK (dimension_id IN (
    'AGENCY', 'CONNECTION', 'APPRECIATION', 'CLARITY', 'VOLITION', 'OPENNESS'
  )),
  slug TEXT NOT NULL UNIQUE,
  positive_label TEXT NOT NULL,
  negative_label TEXT NOT NULL,
  short_description TEXT NOT NULL DEFAULT '',
  default_weight NUMERIC(8,4) NOT NULL DEFAULT 1 CHECK (default_weight >= 0),
  default_visible BOOLEAN NOT NULL DEFAULT true,
  active BOOLEAN NOT NULL DEFAULT true,
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sources (
  id UUID PRIMARY KEY,
  source_key TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  source_type TEXT NOT NULL,
  citation TEXT,
  evidence_class CHAR(1) CHECK (evidence_class IN ('A', 'B', 'C', 'D', 'E')),
  status TEXT NOT NULL DEFAULT 'DRAFT',
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE lexemes (
  id UUID PRIMARY KEY,
  language VARCHAR(10) NOT NULL,
  lemma TEXT NOT NULL,
  part_of_speech TEXT,
  status TEXT NOT NULL DEFAULT 'DRAFT',
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (language, lemma, part_of_speech)
);

CREATE TABLE senses (
  id UUID PRIMARY KEY,
  lexeme_id UUID NOT NULL REFERENCES lexemes(id) ON DELETE CASCADE,
  sense_key TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  description TEXT,
  register_name TEXT,
  domain_name TEXT,
  evidence_class CHAR(1) CHECK (evidence_class IN ('A', 'B', 'C', 'D', 'E')),
  status TEXT NOT NULL DEFAULT 'DRAFT',
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_senses_lexeme_id ON senses(lexeme_id);

CREATE TABLE phrases (
  id UUID PRIMARY KEY,
  language VARCHAR(10) NOT NULL,
  phrase_key TEXT NOT NULL UNIQUE,
  surface_text TEXT NOT NULL,
  normalized_text TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'DRAFT',
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_phrases_normalized_text ON phrases(language, normalized_text);

CREATE TABLE pattern_classes (
  pattern_key TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'DRAFT',
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE dimension_contributions (
  id UUID PRIMARY KEY,
  owner_type TEXT NOT NULL CHECK (owner_type IN ('SENSE', 'PHRASE', 'PATTERN_CLASS', 'RULE')),
  owner_id TEXT NOT NULL,
  dimension_id TEXT NOT NULL REFERENCES dimensions(dimension_id),
  base_value NUMERIC(8,4) NOT NULL CHECK (base_value BETWEEN -100 AND 100),
  confidence NUMERIC(6,5) NOT NULL CHECK (confidence BETWEEN 0 AND 1),
  assessability NUMERIC(6,5) NOT NULL CHECK (assessability BETWEEN 0 AND 1),
  evidence_class CHAR(1) CHECK (evidence_class IN ('A', 'B', 'C', 'D', 'E')),
  context_condition JSONB NOT NULL DEFAULT '{}'::jsonb,
  active BOOLEAN NOT NULL DEFAULT true,
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (owner_type, owner_id, dimension_id, version)
);

CREATE INDEX idx_dimension_contributions_owner ON dimension_contributions(owner_type, owner_id);
CREATE INDEX idx_dimension_contributions_dimension ON dimension_contributions(dimension_id);

CREATE TABLE rule_sets (
  id UUID PRIMARY KEY,
  version TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL CHECK (status IN ('DRAFT', 'TESTING', 'APPROVED', 'PRODUCTION', 'ARCHIVED')),
  changelog TEXT NOT NULL DEFAULT '',
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_rule_sets_one_production
  ON rule_sets ((status)) WHERE status = 'PRODUCTION';

CREATE TABLE rules (
  id UUID PRIMARY KEY,
  rule_key TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  priority INTEGER NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT true,
  scope TEXT NOT NULL DEFAULT 'TEXT',
  condition_tree JSONB NOT NULL,
  actions JSONB NOT NULL,
  confidence_modifier NUMERIC(6,5) NOT NULL DEFAULT 1 CHECK (confidence_modifier BETWEEN 0 AND 1),
  stop_processing BOOLEAN NOT NULL DEFAULT false,
  status TEXT NOT NULL DEFAULT 'DRAFT',
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE rule_set_rules (
  rule_set_id UUID NOT NULL REFERENCES rule_sets(id) ON DELETE CASCADE,
  rule_id UUID NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
  PRIMARY KEY (rule_set_id, rule_id)
);

CREATE TABLE parameters (
  parameter_key TEXT PRIMARY KEY,
  category TEXT NOT NULL,
  default_value JSONB NOT NULL,
  min_value JSONB,
  max_value JSONB,
  description TEXT NOT NULL DEFAULT '',
  editable BOOLEAN NOT NULL DEFAULT true,
  requires_approval BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE parameter_sets (
  id UUID PRIMARY KEY,
  version TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL CHECK (status IN ('DRAFT', 'TESTING', 'APPROVED', 'PRODUCTION', 'ARCHIVED')),
  changelog TEXT NOT NULL DEFAULT '',
  published_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_parameter_sets_one_production
  ON parameter_sets ((status)) WHERE status = 'PRODUCTION';

CREATE TABLE parameter_set_values (
  parameter_set_id UUID NOT NULL REFERENCES parameter_sets(id) ON DELETE CASCADE,
  parameter_key TEXT NOT NULL REFERENCES parameters(parameter_key),
  value JSONB NOT NULL,
  PRIMARY KEY (parameter_set_id, parameter_key)
);

CREATE TABLE presentation_bundles (
  id UUID PRIMARY KEY,
  profile TEXT NOT NULL CHECK (profile IN ('PRIVATE', 'CORPORATE')),
  locale VARCHAR(10) NOT NULL,
  version TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('DRAFT', 'APPROVED', 'PRODUCTION', 'ARCHIVED')),
  fallbacks JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (profile, locale, version)
);

CREATE TABLE presentation_entries (
  bundle_id UUID NOT NULL REFERENCES presentation_bundles(id) ON DELETE CASCADE,
  canonical_key TEXT NOT NULL,
  display_value TEXT NOT NULL,
  PRIMARY KEY (bundle_id, canonical_key)
);

CREATE TABLE golden_test_cases (
  case_id TEXT PRIMARY KEY,
  suite_version TEXT NOT NULL,
  input_payload JSONB NOT NULL,
  expected_payload JSONB NOT NULL,
  active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE import_batches (
  id UUID PRIMARY KEY,
  operation_type TEXT NOT NULL CHECK (operation_type IN ('IMPORT', 'UPDATE', 'VALIDATE_ONLY', 'SYNC_LATER')),
  status TEXT NOT NULL CHECK (status IN (
    'UPLOADED', 'PARSED', 'MAPPED', 'MATCHED', 'VALIDATED', 'READY',
    'RUNNING', 'COMPLETED', 'FAILED', 'ROLLED_BACK', 'CANCELLED'
  )),
  source_name TEXT NOT NULL,
  source_sha256 TEXT NOT NULL,
  mapping_profile JSONB,
  diff_payload JSONB,
  error_summary JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  completed_at TIMESTAMPTZ
);

CREATE TABLE audit_events (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_type TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  actor_id TEXT,
  detail JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id, created_at DESC);

COMMENT ON TABLE golden_test_cases IS 'Versioned test inputs and expectations; never production user analyses.';
COMMENT ON TABLE import_batches IS 'Managed import staging metadata; not a store for analysis raw text.';
