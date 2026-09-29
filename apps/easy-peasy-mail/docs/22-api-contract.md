# API Contract v0 --- Entwurf

## Prinzipien

-   versioniert
-   idempotente Mutationen
-   Cursor statt Offset für Sync
-   Bodies/Attachments getrennt von Listen-Metadaten
-   Offline-first-fähig

## Kernendpunkte

``` text
POST   /v1/auth/session
GET    /v1/mailboxes
GET    /v1/messages?cursor=
GET    /v1/messages/{id}
GET    /v1/messages/{id}/body
PATCH  /v1/messages/{id}/state

GET    /v1/threads/{id}

GET    /v1/attachments/{id}
POST   /v1/attachments/{id}/download-token

GET    /v1/sync/events?after=
POST   /v1/sync/events

POST   /v1/drafts
PATCH  /v1/drafts/{id}
POST   /v1/outbox
POST   /v1/outbox/{id}/schedule
DELETE /v1/outbox/{id}/schedule

GET    /v1/newsletters
GET    /v1/search?q=
```

## Idempotency

Send/Upload/Event-Mutationen tragen:

``` text
Idempotency-Key: <uuid>
```

## Body Response

Body-Endpunkt liefert bewusst nicht automatisch Attachments.

## Event Cursor

Clients speichern einen durable `last_applied_event_id`. Events dürfen
erneut geliefert werden; Anwendung muss idempotent sein.
