# Database / SQLite Schema v0

## Server Core

``` sql
CREATE TABLE messages (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  internet_message_id TEXT,
  thread_id TEXT,
  received_at INTEGER NOT NULL,
  sender_json TEXT NOT NULL,
  recipients_json TEXT NOT NULL,
  subject TEXT,
  preview TEXT,
  body_object_id TEXT,
  mime_object_id TEXT,
  body_hash TEXT,
  spam_score REAL,
  newsletter_id TEXT,
  version INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE message_state (
  message_id TEXT PRIMARY KEY,
  unread INTEGER NOT NULL DEFAULT 1,
  starred INTEGER NOT NULL DEFAULT 0,
  archived INTEGER NOT NULL DEFAULT 0,
  deleted INTEGER NOT NULL DEFAULT 0,
  spam INTEGER NOT NULL DEFAULT 0,
  snooze_until INTEGER,
  version INTEGER NOT NULL
);

CREATE TABLE attachments (
  id TEXT PRIMARY KEY,
  message_id TEXT NOT NULL,
  object_id TEXT NOT NULL,
  filename TEXT,
  mime_type TEXT,
  size_bytes INTEGER NOT NULL,
  sha256 TEXT NOT NULL,
  disposition TEXT,
  content_id TEXT,
  malware_state TEXT NOT NULL,
  FOREIGN KEY(message_id) REFERENCES messages(id)
);

CREATE TABLE sync_events (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  operation TEXT NOT NULL,
  device_id TEXT,
  created_at INTEGER NOT NULL,
  payload_json TEXT NOT NULL
);
```

## Client SQLite

Zusätzlich:

``` text
local_body_cache
local_attachment_cache
outbox
drafts
device_checkpoint
fts_messages
```

SQLite FTS5 ist der bevorzugte lokale Suchindex.

## Regel

Binärdaten liegen nicht als BLOB in der relationalen Mail-Tabelle.
Datenbank hält Referenzen; Content liegt im Object/Vault Store.
