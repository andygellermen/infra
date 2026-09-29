# SQL DDL v1

> PostgreSQL-orientierter Serverentwurf. SQLite Client nutzt eine
> reduzierte Projektion.

``` sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    primary_address TEXT NOT NULL UNIQUE,
    display_name TEXT,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE','SUSPENDED','CLOSED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE devices (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    name TEXT NOT NULL,
    platform TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('ACTIVE','REVOKED')),
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE threads (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    subject_key TEXT,
    latest_message_at TIMESTAMPTZ,
    message_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    thread_id UUID REFERENCES threads(id),
    internet_message_id TEXT,
    inbound_delivery_id TEXT,
    direction TEXT NOT NULL CHECK (direction IN ('INBOUND','OUTBOUND')),
    received_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    subject TEXT,
    preview TEXT,
    from_json JSONB NOT NULL,
    to_json JSONB NOT NULL,
    cc_json JSONB,
    bcc_json JSONB,
    reply_to_json JSONB,
    headers_ref TEXT,
    body_text_ref TEXT,
    body_html_ref TEXT,
    raw_mime_ref TEXT NOT NULL,
    canonical_hash TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    attachment_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_messages_delivery
ON messages(account_id, inbound_delivery_id)
WHERE inbound_delivery_id IS NOT NULL;

CREATE INDEX ix_messages_account_received
ON messages(account_id, received_at DESC);

CREATE INDEX ix_messages_thread
ON messages(thread_id, received_at);

CREATE TABLE mailbox_state (
    message_id UUID PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    unread BOOLEAN NOT NULL DEFAULT TRUE,
    starred BOOLEAN NOT NULL DEFAULT FALSE,
    archived BOOLEAN NOT NULL DEFAULT FALSE,
    spam BOOLEAN NOT NULL DEFAULT FALSE,
    trashed_at TIMESTAMPTZ,
    snooze_until TIMESTAMPTZ,
    version BIGINT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE labels (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    name TEXT NOT NULL,
    mode TEXT NOT NULL CHECK (mode IN ('USER','SYSTEM')),
    UNIQUE(account_id, name)
);

CREATE TABLE message_labels (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    label_id UUID NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY(message_id, label_id)
);

CREATE TABLE content_objects (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    sha256 TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_key TEXT NOT NULL UNIQUE,
    media_type TEXT,
    state TEXT NOT NULL CHECK (state IN ('ACTIVE','QUARANTINED','DELETE_PENDING')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(account_id, sha256)
);

CREATE TABLE attachments (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    object_id UUID NOT NULL REFERENCES content_objects(id),
    filename TEXT,
    mime_type TEXT,
    size_bytes BIGINT NOT NULL,
    sha256 TEXT NOT NULL,
    disposition TEXT NOT NULL CHECK (disposition IN ('ATTACHMENT','INLINE')),
    content_id TEXT,
    scan_state TEXT NOT NULL CHECK (scan_state IN ('UNKNOWN','CLEAN','SUSPICIOUS','BLOCKED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ix_attachments_message ON attachments(message_id);

CREATE TABLE drafts (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    revision BIGINT NOT NULL DEFAULT 1,
    recipients_json JSONB NOT NULL,
    subject TEXT,
    body_markdown TEXT NOT NULL DEFAULT '',
    attachment_refs JSONB NOT NULL DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by_device UUID REFERENCES devices(id)
);

CREATE TABLE draft_revisions (
    draft_id UUID NOT NULL REFERENCES drafts(id) ON DELETE CASCADE,
    revision BIGINT NOT NULL,
    base_revision BIGINT,
    device_id UUID REFERENCES devices(id),
    snapshot_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(draft_id, revision)
);

CREATE TABLE outbox (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    draft_snapshot_ref TEXT NOT NULL,
    idempotency_key UUID NOT NULL UNIQUE,
    state TEXT NOT NULL CHECK (state IN ('PENDING','SCHEDULED','SENDING','RETRY_WAIT','SENT','FAILED','CANCELLED')),
    scheduled_for TIMESTAMPTZ,
    version BIGINT NOT NULL DEFAULT 1,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ,
    last_error_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ix_outbox_due
ON outbox(state, scheduled_for, next_attempt_at);

CREATE TABLE sync_events (
    seq BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    stream TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    entity_version BIGINT,
    device_id UUID REFERENCES devices(id),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    payload JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX ix_sync_events_account_seq
ON sync_events(account_id, seq);

CREATE TABLE device_checkpoints (
    device_id UUID PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    last_seq BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ingress (
    id UUID PRIMARY KEY,
    source TEXT NOT NULL,
    source_message_id TEXT,
    raw_object_ref TEXT NOT NULL,
    recipient TEXT NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('RECEIVED','PROCESSING','COMMITTED','QUARANTINED','FAILED')),
    lease_until TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(source, source_message_id)
);

CREATE TABLE send_ledger (
    idempotency_key UUID PRIMARY KEY,
    outbox_id UUID NOT NULL REFERENCES outbox(id),
    provider TEXT NOT NULL,
    provider_message_id TEXT,
    attempt_started_at TIMESTAMPTZ NOT NULL,
    attempt_finished_at TIMESTAMPTZ,
    result TEXT
);
```

## Noch bewusst offen

-   Partitionierung von `sync_events` erst nach Messdaten.
-   Adressen ggf. später normalisieren statt JSONB.
-   Newsletter-Tabellen kommen mit Classifier-Spezifikation.
