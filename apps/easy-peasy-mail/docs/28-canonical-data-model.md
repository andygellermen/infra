# Canonical Data Model v1

## Grundsatz

EasyPeasyMail trennt vier Dinge strikt: 1. unveränderliche
Nachrichtendaten, 2. veränderlichen Mailbox-State, 3. gerätespezifischen
lokalen State, 4. Transport-/Verarbeitungszustände.

Damit muss ein `read=true` niemals eine komplette Nachricht neu
schreiben.

## Account

``` text
Account
- id UUID
- primary_address
- display_name
- status ACTIVE|SUSPENDED|CLOSED
- created_at
```

## Address

``` text
Address
- id
- account_id
- local_part
- domain
- kind PRIMARY|ALIAS
- status
```

## Message

Logisches, unveränderliches Nachrichtenobjekt.

``` text
Message
- id UUID
- account_id
- thread_id?
- internet_message_id?
- inbound_delivery_id?
- direction INBOUND|OUTBOUND
- received_at?
- sent_at?
- subject
- preview
- from_json
- to_json
- cc_json
- bcc_json
- reply_to_json
- headers_ref
- body_text_ref?
- body_html_ref?
- raw_mime_ref
- canonical_hash
- size_bytes
- attachment_count
- created_at
```

## MailboxState

``` text
MailboxState
- message_id
- unread bool
- starred bool
- archived bool
- spam bool
- trashed_at?
- snooze_until?
- version bigint
- updated_at
```

## Label

``` text
Label
- id
- account_id
- name
- mode USER|SYSTEM
```

`MessageLabel` ist eine eigene Relation.

## Attachment

``` text
Attachment
- id UUID
- message_id
- object_id
- filename
- mime_type
- size_bytes
- sha256
- disposition ATTACHMENT|INLINE
- content_id?
- scan_state UNKNOWN|CLEAN|SUSPICIOUS|BLOCKED
- created_at
```

## ContentObject

Physisches Objekt im Attachment Vault/Object Store.

``` text
ContentObject
- id
- account_id
- sha256
- size_bytes
- storage_key
- media_type
- ref_count_hint
- state ACTIVE|QUARANTINED|DELETE_PENDING
- created_at
```

`ref_count_hint` ist Optimierung, niemals alleinige Löschwahrheit.

## Thread

``` text
Thread
- id
- account_id
- subject_key?
- latest_message_at
- message_count
```

Threading wird aus `Message-ID`, `In-Reply-To`, `References` und
heuristischer Fallback-Logik aufgebaut.

## Draft

Draft ist absichtlich kein Message-Objekt.

``` text
Draft
- id
- account_id
- revision
- recipients_json
- subject
- body_markdown
- attachment_refs
- updated_at
- device_id
```

Erst ein erfolgreicher Send-Vorgang erzeugt die finale OUTBOUND Message.

## OutboxItem

``` text
OutboxItem
- id
- account_id
- draft_snapshot_ref
- idempotency_key UNIQUE
- state PENDING|SCHEDULED|SENDING|SENT|FAILED|CANCELLED
- scheduled_for?
- attempt_count
- next_attempt_at?
- last_error_code?
- created_at
- updated_at
```

## Device

``` text
Device
- id
- account_id
- name
- platform
- status ACTIVE|REVOKED
- last_seen_at
- created_at
```

## DeviceMessageState

Nur lokaler/cachebezogener Zustand.

``` text
DeviceMessageState
- device_id
- message_id
- body_state NONE|PREVIEW|CACHED|PINNED
- last_opened_at?
```

## NewsletterSource

``` text
NewsletterSource
- id
- account_id
- list_id?
- sender_key
- display_name
- classification AUTO|USER
- collapsed_default bool
```

## SpamDecision

``` text
SpamDecision
- message_id
- verdict ACCEPT|SPAM|QUARANTINE|REJECTED
- score?
- reasons_json
- decided_at
```

## Kern-Invarianten

-   Message Content ist nach Commit unveränderlich.
-   State-Änderungen erzeugen Events.
-   Draft und Message sind unterschiedliche Lebenszyklen.
-   Binärdaten liegen nicht als DB-BLOB im Message-Record.
-   Client-Cache-State darf niemals serverweiten Mailbox-State
    überschreiben.
