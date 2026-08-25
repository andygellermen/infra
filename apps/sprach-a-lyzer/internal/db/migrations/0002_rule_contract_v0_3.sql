ALTER TABLE rules
  ADD COLUMN contract_version TEXT NOT NULL DEFAULT '0.1',
  ADD COLUMN evidence_class CHAR(1) CHECK (evidence_class IN ('A', 'B', 'C', 'D', 'E')),
  ADD COLUMN source_keys JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE rules
  ADD CONSTRAINT rules_contract_version_check
  CHECK (contract_version IN ('0.1', '0.2', '0.3'));

ALTER TABLE rules DROP CONSTRAINT rules_rule_key_key;
ALTER TABLE rules ADD CONSTRAINT rules_rule_key_version_key UNIQUE (rule_key, version);

COMMENT ON COLUMN rules.contract_version IS 'Canonical rule payload contract; migrated Foundation rules use Rule v0.3.';
COMMENT ON COLUMN rules.source_keys IS 'Stable provenance keys; source content remains in the knowledge catalogue.';
