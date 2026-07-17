ALTER TABLE tenant_domain_bindings ADD COLUMN verification_token TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_domain_bindings ADD COLUMN dns_verified_at TEXT;
ALTER TABLE tenant_domain_bindings ADD COLUMN routing_verified_at TEXT;
ALTER TABLE tenant_domain_bindings ADD COLUMN last_dns_check_at TEXT;
ALTER TABLE tenant_domain_bindings ADD COLUMN last_dns_error TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_domain_bindings ADD COLUMN last_routing_check_at TEXT;
ALTER TABLE tenant_domain_bindings ADD COLUMN last_routing_error TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_domain_bindings ADD COLUMN ssl_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE tenant_domain_bindings ADD COLUMN ssl_certificate_issuer TEXT NOT NULL DEFAULT '';
ALTER TABLE tenant_domain_bindings ADD COLUMN ssl_certificate_expires_at TEXT;
ALTER TABLE tenant_domain_bindings ADD COLUMN last_ssl_check_at TEXT;
ALTER TABLE tenant_domain_bindings ADD COLUMN last_ssl_error TEXT NOT NULL DEFAULT '';

UPDATE tenant_domain_bindings
SET verification_token = lower(hex(randomblob(12)))
WHERE verification_token = '';

UPDATE tenant_domain_bindings
SET ssl_status = 'valid'
WHERE status = 'active'
  AND ssl_status = 'pending';
