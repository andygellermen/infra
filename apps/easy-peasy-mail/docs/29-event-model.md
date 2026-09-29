# Event Model v1

## Ziel

Events synchronisieren Zustandsänderungen. Sie sind **nicht** die
primäre Ablage des Mailcontents.

## Event Envelope

``` json
{
  "event_id": "01J...",
  "account_id": "acc_...",
  "stream": "mailbox",
  "entity_type": "message_state",
  "entity_id": "msg_...",
  "operation": "message.read",
  "entity_version": 18,
  "device_id": "dev_...",
  "occurred_at": "2026-09-25T12:00:00Z",
  "payload": {}
}
```

## Event-ID

Zeitlich sortierbare IDs wie UUIDv7 oder ULID sind geeignet. Die
konkrete Wahl bleibt Implementierungsentscheidung.

## Streams

``` text
mailbox
draft
outbox
attachment
newsletter
device
system
```

## Wichtige Events

``` text
message.committed
message.read
message.unread
message.starred
message.unstarred
message.archived
message.restored
message.trashed
message.spam_marked

label.added
label.removed

draft.created
draft.updated
draft.deleted

outbox.created
outbox.scheduled
outbox.send_started
outbox.sent
outbox.send_failed
outbox.cancelled

attachment.committed
attachment.quarantined
attachment.deleted

newsletter.classified
newsletter.unclassified

device.registered
device.revoked
```

## Delivery Semantics

At-least-once.

Warum: Exactly-once über verteilte Komponenten ist unnötig
teuer/kompliziert. Stattdessen sind Events und Mutationen idempotent.

## Reihenfolge

Reihenfolge wird pro Account/Stream über einen monotonen Cursor
garantiert bzw. rekonstruiert. Globale Reihenfolge über alle Accounts
ist nicht nötig.

## Replay

Ein Client kann ab `last_event_id` erneut lesen. Ein Snapshot plus Delta
verhindert unbegrenzt wachsende Replay-Zeiten.

## Event Retention

Kurzlebige Relay-Kopie und langlebigerer Core-Journal sind getrennte
Dinge. Relay darf verfallen; Recovery darf davon nicht abhängen.
