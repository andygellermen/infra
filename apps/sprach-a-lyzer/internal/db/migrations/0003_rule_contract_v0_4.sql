ALTER TABLE rules DROP CONSTRAINT rules_contract_version_check;

ALTER TABLE rules
  ADD CONSTRAINT rules_contract_version_check
  CHECK (contract_version IN ('0.1', '0.2', '0.3', '0.4'));

COMMENT ON COLUMN rules.contract_version IS 'Canonical rule payload contract; Core Closure rules use Rule v0.4.';
