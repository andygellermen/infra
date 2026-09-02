ALTER TABLE import_batches
  ADD COLUMN batch_key TEXT UNIQUE,
  ADD COLUMN source_type TEXT NOT NULL DEFAULT 'JSON' CHECK (source_type IN ('JSON', 'CSV', 'XLSX')),
  ADD COLUMN target_entity TEXT NOT NULL DEFAULT 'KNOWLEDGE_RECORD',
  ADD COLUMN actor_id TEXT,
  ADD COLUMN total_rows INTEGER NOT NULL DEFAULT 0 CHECK (total_rows >= 0),
  ADD COLUMN new_rows INTEGER NOT NULL DEFAULT 0 CHECK (new_rows >= 0),
  ADD COLUMN changed_rows INTEGER NOT NULL DEFAULT 0 CHECK (changed_rows >= 0),
  ADD COLUMN unchanged_rows INTEGER NOT NULL DEFAULT 0 CHECK (unchanged_rows >= 0),
  ADD COLUMN conflict_rows INTEGER NOT NULL DEFAULT 0 CHECK (conflict_rows >= 0),
  ADD COLUMN invalid_rows INTEGER NOT NULL DEFAULT 0 CHECK (invalid_rows >= 0),
  ADD COLUMN validated_at TIMESTAMPTZ,
  ADD COLUMN committed_at TIMESTAMPTZ,
  ADD COLUMN rolled_back_at TIMESTAMPTZ;

CREATE INDEX idx_import_batches_source_sha256 ON import_batches(source_sha256);

CREATE TABLE managed_knowledge_records (
  entity_type TEXT NOT NULL,
  natural_key TEXT NOT NULL,
  payload JSONB NOT NULL,
  references JSONB NOT NULL DEFAULT '[]'::jsonb,
  version INTEGER NOT NULL CHECK (version > 0),
  status TEXT NOT NULL CHECK (status IN ('DRAFT', 'REVIEW', 'APPROVED', 'PRODUCTION', 'ARCHIVED')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (entity_type, natural_key)
);

CREATE TABLE import_batch_rows (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  batch_id UUID NOT NULL REFERENCES import_batches(id) ON DELETE CASCADE,
  row_number INTEGER NOT NULL CHECK (row_number > 0),
  natural_key TEXT,
  raw_payload JSONB NOT NULL,
  normalized_payload JSONB,
  matched_natural_key TEXT,
  match_confidence TEXT NOT NULL CHECK (match_confidence IN ('EXACT', 'PROBABLE', 'AMBIGUOUS', 'NONE')),
  row_status TEXT NOT NULL CHECK (row_status IN ('NEW', 'UNCHANGED', 'CHANGED', 'CONFLICT', 'INVALID', 'SKIPPED', 'REFERENCE_MISSING')),
  diff JSONB NOT NULL DEFAULT '{}'::jsonb,
  errors JSONB NOT NULL DEFAULT '[]'::jsonb,
  warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
  resolution TEXT CHECK (resolution IN ('KEEP_DATABASE', 'USE_IMPORT', 'REQUIRE_MANUAL')),
  UNIQUE (batch_id, row_number)
);

CREATE TABLE import_change_log (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  batch_id UUID NOT NULL REFERENCES import_batches(id),
  entity_type TEXT NOT NULL,
  natural_key TEXT NOT NULL,
  operation TEXT NOT NULL CHECK (operation IN ('INSERT', 'UPDATE')),
  before_payload JSONB,
  after_payload JSONB NOT NULL,
  rolled_back BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE mapping_profiles (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  target_entity TEXT NOT NULL,
  column_mapping JSONB NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'ARCHIVED')),
  version INTEGER NOT NULL CHECK (version > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (name, version)
);

CREATE TABLE import_presets (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  operation_type TEXT NOT NULL CHECK (operation_type IN ('IMPORT', 'UPDATE', 'VALIDATE_ONLY')),
  source_type TEXT NOT NULL CHECK (source_type IN ('JSON', 'CSV', 'XLSX')),
  target_entity TEXT NOT NULL,
  configuration JSONB NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('ACTIVE', 'ARCHIVED')),
  version INTEGER NOT NULL CHECK (version > 0),
  usage_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (name, version)
);

CREATE OR REPLACE FUNCTION reject_audit_mutation() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'audit_events are immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_events_immutable
BEFORE UPDATE OR DELETE ON audit_events
FOR EACH ROW EXECUTE FUNCTION reject_audit_mutation();

COMMENT ON TABLE import_batch_rows IS 'Normalized staging rows; never production user analysis text.';
COMMENT ON TABLE import_change_log IS 'Reversible before/after records for one committed managed import.';
