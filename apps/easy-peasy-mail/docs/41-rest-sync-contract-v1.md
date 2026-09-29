# REST / Sync Contract v1

## Konvention

JSON, UTC-Zeitstempel, versionierte API `/v1`.

## Read message

``` http
GET /v1/messages/{id}
```

Response:

``` json
{
  "id": "msg_...",
  "thread_id": "thr_...",
  "subject": "Termin",
  "preview": "Hallo...",
  "state": {
    "unread": false,
    "starred": false,
    "archived": false,
    "version": 19
  },
  "body_state": "REMOTE",
  "attachments": []
}
```

## State mutation

``` http
PATCH /v1/messages/{id}/state
If-Match: 19
Idempotency-Key: ...
```

``` json
{"unread": false}
```

Conflict:

``` http
409 STATE_VERSION_CONFLICT
```

## Sync

``` http
GET /v1/sync/events?after=918221&limit=500
```

``` json
{
  "events": [],
  "next_cursor": 918743,
  "has_more": false
}
```

Cursor ist serverseitige `seq`, Event-ID bleibt globale Eventidentität.

## Draft update

``` http
PUT /v1/drafts/{id}
If-Match: 12
```

``` json
{
  "subject": "Termin",
  "body_markdown": "Hallo ...",
  "recipients": [],
  "attachment_refs": []
}
```

Erfolg: Revision 13.

Konflikt:

``` http
409 DRAFT_CONFLICT
```

Response enthält: - current_revision - current_snapshot -
submitted_snapshot - base_revision

## Send

``` http
POST /v1/outbox
Idempotency-Key: ...
```

``` json
{
  "draft_id": "...",
  "expected_revision": 13,
  "scheduled_for": null
}
```

## Fehlerformat

``` json
{
  "error": {
    "code": "DRAFT_CONFLICT",
    "message": "Draft has changed on another device.",
    "retryable": false,
    "request_id": "..."
  }
}
```
