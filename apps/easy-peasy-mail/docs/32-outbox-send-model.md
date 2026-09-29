# Outbox & Scheduled Send Model

## Zustandsautomat

``` text
DRAFT
  |
  v
PENDING
  | schedule
  v
SCHEDULED
  | due
  v
SENDING
  |------ transient error ------> RETRY_WAIT
  |                                  |
  |                                  +----> SENDING
  |
  +------ success ----------------> SENT
  |
  +------ permanent error --------> FAILED
```

## Send Command

``` text
outbox_id
account_id
snapshot_ref
idempotency_key
scheduled_for?
expected_version
```

## Snapshot

Beim Übergang Draft -\> Outbox wird ein unveränderlicher
Versand-Snapshot erzeugt. Spätere Draft-Änderungen verändern den bereits
geplanten Send nicht.

## Retry

Exponential Backoff mit Jitter, begrenzt durch Fehlerklasse. Permanente
Adress-/Policyfehler werden nicht endlos retried.

## Crash während Send

Nach unbekanntem Provider-Ergebnis wird nicht blind erneut gesendet.
Provider Message ID / idempotente Providerintegration bzw. Send Ledger
werden geprüft.

## Send Ledger

``` text
idempotency_key UNIQUE
provider
provider_message_id?
attempt_started_at
attempt_finished_at?
result
```

## Scheduled Send

Server ist Autorität für geplanten Versand. Client darf ausgeschaltet
sein.
