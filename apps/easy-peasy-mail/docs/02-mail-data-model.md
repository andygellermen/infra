# Mail-Datenmodell

## Message

``` text
id
internet_message_id
thread_id
account_id
received_at
from
to/cc/bcc
subject
preview
body_ref
body_hash
body_state       NONE | PREVIEW | CACHED | PINNED
attachment_count
auth_results
spam_score
newsletter_id?
created_at
```

## Device State

``` text
message_id
device_id
read_state
local_body_state
local_attachment_state
last_seen_version
```

## User Mail State

``` text
message_id
read
starred
archived
deleted
spam
labels[]
snooze_until?
version
```

## Attachment

``` text
id
message_id
filename
mime_type
size
sha256
storage_ref
local_state
malware_state
```

## Sync Event

``` text
event_id
account_id
entity_id
entity_type
operation
version
device_id
timestamp
payload
```

## Newsletter Bundle

``` text
bundle_id
sender/domain
list_id
display_name
unread_count
latest_message_id
collapsed=true
```

### Grundsatz

Mailinhalt und Mailzustand werden getrennt. Dadurch müssen kleine
Aktionen wie `read`, `archive` oder `spam` nicht komplette
MIME-Nachrichten replizieren.
