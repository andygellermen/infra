# Easy-Event-Planner – Datenmodell und SQLite-DDL

## Prinzipien

- `tenant_id` trennt Mandanten.
- IDs sind `TEXT` und werden als UUID/ULID erzeugt.
- Zeitfelder werden als ISO-8601-Text gespeichert.
- Booleans werden als `INTEGER` gespeichert.
- Statuswerte sind `TEXT`.
- SQLite reicht für MVP; Modell bleibt PostgreSQL-fähig.

## Initiales DDL

```sql
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;

CREATE TABLE tenants (
  id TEXT PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  public_base_url TEXT NOT NULL,
  default_timezone TEXT NOT NULL DEFAULT 'Europe/Berlin',
  default_locale TEXT NOT NULL DEFAULT 'de-DE',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE tenant_settings (
  tenant_id TEXT PRIMARY KEY,
  sender_email TEXT,
  sender_name TEXT,
  paypal_mode TEXT NOT NULL DEFAULT 'disabled',
  paypal_client_id TEXT,
  paypal_merchant_id TEXT,
  default_retention_days INTEGER NOT NULL DEFAULT 30,
  settings_json TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE tenant_users (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  email TEXT NOT NULL,
  name TEXT,
  role TEXT NOT NULL DEFAULT 'event_manager',
  status TEXT NOT NULL DEFAULT 'active',
  last_login_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, email),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE magic_links (
  id TEXT PRIMARY KEY,
  tenant_id TEXT,
  user_id TEXT,
  participant_id TEXT,
  purpose TEXT NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  redirect_path TEXT,
  expires_at TEXT NOT NULL,
  used_at TEXT,
  request_ip TEXT,
  user_agent TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES tenant_users(id) ON DELETE CASCADE
);

CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  tenant_id TEXT,
  user_id TEXT,
  participant_id TEXT,
  session_hash TEXT NOT NULL UNIQUE,
  expires_at TEXT NOT NULL,
  revoked_at TEXT,
  created_at TEXT NOT NULL,
  last_seen_at TEXT,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES tenant_users(id) ON DELETE CASCADE
);

CREATE TABLE event_series (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  slug TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  default_location_name TEXT,
  default_address TEXT,
  default_online_url TEXT,
  is_public INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, slug),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE events (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  series_id TEXT,
  slug TEXT NOT NULL,
  title TEXT NOT NULL,
  subtitle TEXT,
  description TEXT,
  starts_at TEXT NOT NULL,
  ends_at TEXT,
  timezone TEXT NOT NULL DEFAULT 'Europe/Berlin',
  location_name TEXT,
  address TEXT,
  online_url TEXT,
  participation_mode TEXT NOT NULL DEFAULT 'onsite',
  status TEXT NOT NULL DEFAULT 'draft',
  is_public INTEGER NOT NULL DEFAULT 0,
  registration_enabled INTEGER NOT NULL DEFAULT 1,
  waitlist_enabled INTEGER NOT NULL DEFAULT 1,
  max_participants INTEGER,
  change_note TEXT,
  cancelled_reason TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, slug),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (series_id) REFERENCES event_series(id) ON DELETE SET NULL
);

CREATE TABLE event_tickets (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  event_id TEXT NOT NULL,
  name TEXT NOT NULL,
  ticket_type TEXT NOT NULL DEFAULT 'standard',
  price_cents INTEGER NOT NULL DEFAULT 0,
  currency TEXT NOT NULL DEFAULT 'EUR',
  max_quantity INTEGER,
  donation_enabled INTEGER NOT NULL DEFAULT 0,
  donation_min_cents INTEGER,
  donation_suggested_cents INTEGER,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);

CREATE TABLE participants (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  email TEXT,
  phone TEXT,
  name TEXT,
  email_verified_at TEXT,
  phone_verified_at TEXT,
  anonymized_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, email),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE registrations (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  event_id TEXT NOT NULL,
  participant_id TEXT NOT NULL,
  ticket_id TEXT,
  status TEXT NOT NULL DEFAULT 'verification_pending',
  participation_type TEXT NOT NULL DEFAULT 'onsite',
  quantity INTEGER NOT NULL DEFAULT 1,
  reserved_until TEXT,
  confirmed_at TEXT,
  cancelled_at TEXT,
  cancellation_reason TEXT,
  attended_at TEXT,
  source TEXT NOT NULL DEFAULT 'public_page',
  invite_id TEXT,
  privacy_accepted_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
  FOREIGN KEY (participant_id) REFERENCES participants(id) ON DELETE CASCADE,
  FOREIGN KEY (ticket_id) REFERENCES event_tickets(id) ON DELETE SET NULL
);

CREATE TABLE waitlist_entries (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  event_id TEXT NOT NULL,
  registration_id TEXT NOT NULL,
  position INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'waiting',
  offered_at TEXT,
  offer_expires_at TEXT,
  accepted_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
  FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
);

CREATE TABLE payments (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  registration_id TEXT NOT NULL,
  provider TEXT NOT NULL DEFAULT 'paypal',
  provider_order_id TEXT,
  provider_capture_id TEXT,
  status TEXT NOT NULL DEFAULT 'created',
  amount_cents INTEGER NOT NULL,
  currency TEXT NOT NULL DEFAULT 'EUR',
  donation_amount_cents INTEGER,
  raw_provider_response TEXT,
  paid_at TEXT,
  refunded_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
);

CREATE TABLE paypal_webhook_events (
  id TEXT PRIMARY KEY,
  tenant_id TEXT,
  paypal_event_id TEXT NOT NULL UNIQUE,
  event_type TEXT NOT NULL,
  resource_id TEXT,
  payload_json TEXT NOT NULL,
  verified_at TEXT,
  processed_at TEXT,
  processing_status TEXT NOT NULL DEFAULT 'received',
  error_message TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL
);

CREATE TABLE invitation_links (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  event_id TEXT,
  series_id TEXT,
  code TEXT NOT NULL,
  label TEXT,
  invite_type TEXT NOT NULL,
  discount_type TEXT,
  discount_value INTEGER,
  max_uses INTEGER,
  used_count INTEGER NOT NULL DEFAULT 0,
  max_uses_per_email INTEGER,
  starts_at TEXT,
  expires_at TEXT,
  is_shareable INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, code),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
  FOREIGN KEY (series_id) REFERENCES event_series(id) ON DELETE CASCADE
);

CREATE TABLE discount_redemptions (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  invitation_link_id TEXT,
  registration_id TEXT NOT NULL,
  participant_email TEXT,
  discount_amount_cents INTEGER NOT NULL DEFAULT 0,
  redeemed_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (invitation_link_id) REFERENCES invitation_links(id) ON DELETE SET NULL,
  FOREIGN KEY (registration_id) REFERENCES registrations(id) ON DELETE CASCADE
);

CREATE TABLE email_jobs (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  template_key TEXT NOT NULL,
  recipient_email TEXT NOT NULL,
  subject TEXT NOT NULL,
  body_text TEXT NOT NULL,
  body_html TEXT,
  status TEXT NOT NULL DEFAULT 'queued',
  scheduled_for TEXT,
  sent_at TEXT,
  error_message TEXT,
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE snippet_configs (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  view_type TEXT NOT NULL DEFAULT 'cards',
  event_filter_json TEXT,
  display_options_json TEXT,
  is_active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, slug),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE calendar_feeds (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  user_id TEXT,
  feed_type TEXT NOT NULL DEFAULT 'organizer',
  token_hash TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL DEFAULT 'active',
  last_accessed_at TEXT,
  created_at TEXT NOT NULL,
  rotated_at TEXT,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
  FOREIGN KEY (user_id) REFERENCES tenant_users(id) ON DELETE CASCADE
);

CREATE TABLE retention_policies (
  id TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  data_category TEXT NOT NULL,
  action TEXT NOT NULL DEFAULT 'anonymize',
  retention_days INTEGER NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE (tenant_id, data_category),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE TABLE audit_log (
  id TEXT PRIMARY KEY,
  tenant_id TEXT,
  actor_user_id TEXT,
  actor_participant_id TEXT,
  action TEXT NOT NULL,
  entity_type TEXT,
  entity_id TEXT,
  details_json TEXT,
  request_ip TEXT,
  user_agent TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE SET NULL
);
```

## Indizes

```sql
CREATE INDEX idx_events_tenant_starts_at ON events(tenant_id, starts_at);
CREATE INDEX idx_events_tenant_status ON events(tenant_id, status);
CREATE INDEX idx_registrations_event_status ON registrations(event_id, status);
CREATE INDEX idx_waitlist_event_status ON waitlist_entries(event_id, status, position);
CREATE INDEX idx_magic_links_hash ON magic_links(token_hash);
CREATE INDEX idx_sessions_hash ON sessions(session_hash);
CREATE INDEX idx_email_jobs_status ON email_jobs(status, scheduled_for);
CREATE INDEX idx_audit_tenant_created ON audit_log(tenant_id, created_at);
```

## Kapazitätszählung

Aktive Platzbindung:

```text
confirmed
reserved mit gültigem reserved_until
payment_pending mit gültigem reserved_until
```

SQL-Idee:

```sql
SELECT COALESCE(SUM(quantity), 0)
FROM registrations
WHERE event_id = ?
  AND status IN ('confirmed', 'reserved', 'payment_pending')
  AND (reserved_until IS NULL OR reserved_until > ?);
```

## Anonymisierung

```sql
UPDATE participants
SET name = NULL,
    email = NULL,
    phone = NULL,
    anonymized_at = ?
WHERE tenant_id = ? AND id = ?;
```

Registrierungen bleiben für Statistik und Kapazitätsauswertung erhalten, verlieren aber ihren Personenbezug.
