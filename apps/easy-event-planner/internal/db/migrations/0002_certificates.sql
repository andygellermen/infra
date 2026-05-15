CREATE TABLE certificates (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  registration_id TEXT NOT NULL,
  participant_id TEXT NOT NULL,
  event_id TEXT NOT NULL,
  certificate_number TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'issued',
  issued_at TEXT NOT NULL,
  attended_at TEXT,
  file_path TEXT NOT NULL,
  file_sha256 TEXT NOT NULL,
  verification_code_hash TEXT NOT NULL UNIQUE,
  verification_code_hint TEXT,
  revoked_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, registration_id),
  UNIQUE (tenant_id, certificate_number),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE,
  FOREIGN KEY (participant_id) REFERENCES participants(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);

CREATE INDEX idx_certificates_participant_issued
  ON certificates(tenant_id, participant_id, issued_at DESC);
CREATE INDEX idx_certificates_event_issued
  ON certificates(tenant_id, event_id, issued_at DESC);
