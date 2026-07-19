ALTER TABLE tenant_domain_bindings ADD COLUMN overview_enabled INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tenant_domain_bindings ADD COLUMN event_detail_enabled INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tenant_domain_bindings ADD COLUMN registration_embed_enabled INTEGER NOT NULL DEFAULT 1;
ALTER TABLE tenant_domain_bindings ADD COLUMN organizer_calendar_enabled INTEGER NOT NULL DEFAULT 1;
