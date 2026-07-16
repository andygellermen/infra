CREATE TABLE tenant_domain_bindings (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  domain_host TEXT NOT NULL,
  base_path TEXT NOT NULL DEFAULT '/',
  status TEXT NOT NULL DEFAULT 'pending_dns',
  is_primary INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (domain_host, base_path),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_tenant_domain_bindings_tenant_id
  ON tenant_domain_bindings (tenant_id);

CREATE INDEX idx_tenant_domain_bindings_domain_host
  ON tenant_domain_bindings (domain_host);
