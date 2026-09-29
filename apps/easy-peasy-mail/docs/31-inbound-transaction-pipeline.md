# Inbound Transaction Pipeline

## Ziel

Sobald EasyPeasyMail eine Nachricht als dauerhaft übernommen betrachtet,
darf kein nachgelagerter Worker-Ausfall sie verlieren.

## Pipeline

``` text
SMTP/SES Edge
 -> durable ingress object
 -> ingress event
 -> claim processing lease
 -> parse MIME
 -> validate recipient
 -> auth/spam signals
 -> calculate hashes
 -> dedupe
 -> extract body
 -> extract attachments
 -> scan attachments
 -> commit DB transaction
 -> emit message.committed
 -> mark ingress COMPLETE
 -> lifecycle cleanup later
```

## Ingress Record

``` text
Ingress
- id
- source
- source_message_id
- raw_object_ref
- recipient
- state RECEIVED|PROCESSING|COMMITTED|QUARANTINED|FAILED
- lease_until?
- attempts
- last_error?
- created_at
```

## Processing Lease

Worker nimmt keinen Datensatz dauerhaft „weg". Er erhält nur eine
zeitlich begrenzte Lease. Stirbt der Worker, darf ein anderer
übernehmen.

## Dedupe Key

Mehrstufig: 1. eindeutige Edge/Provider Delivery ID, falls vorhanden 2.
raw MIME SHA-256 3. Internet Message-ID + recipient + content
fingerprint 4. Zeitfenster als Fallback

## Commit Boundary

DB Message + State + Attachment References + Ingress COMMITTED werden
atomar bzw. über ein Outbox/transactional-event Muster gekoppelt.

## Poison Message

Nach wiederholtem Parserfehler: `QUARANTINED`, nicht gelöscht.

Original bleibt verfügbar, damit Parser später verbessert und Nachricht
erneut verarbeitet werden kann.
