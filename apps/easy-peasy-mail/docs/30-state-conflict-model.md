# State & Conflict Model

## Grundidee

Nicht jede Eigenschaft benötigt dieselbe Konfliktstrategie.

## Read State

`read/unread` ist ein skalarer Nutzerzustand. Server akzeptiert Mutation
gegen bekannte Version und erzeugt neue Version.

Bei konkurrierenden Änderungen gewinnt zunächst die serverseitig
akzeptierte höhere Version. UI kann bei Bedarf den jüngsten Nutzerwunsch
erneut senden.

## Starred / Archived

Wie Read State: versionierte skalare Mutationen.

## Labels

Keine komplette Label-Liste überschreiben.

``` text
label.added(message, label)
label.removed(message, label)
```

Dadurch löschen zwei Geräte nicht gegenseitig unabhängige
Label-Änderungen.

## Delete

Delete ist zunächst Tombstone:

``` text
ACTIVE -> TRASHED -> PURGE_ELIGIBLE -> PURGED
```

Restore ist bis `PURGED` möglich.

## Draft

Keine automatische Textverschmelzung paralleler Draft-Edits.

Draft besitzt `revision`. Bei Konflikt: - Server lehnt stale revision
ab. - Client erhält aktuelle Revision. - UI bietet beide Fassungen bzw.
„Kopie behalten".

## Send

Send besitzt einen unveränderlichen `idempotency_key`. Dasselbe
Send-Kommando darf beliebig oft retried werden und maximal eine logische
Sendung erzeugen.

## Scheduled Send

Änderung eines bereits geplanten Sends erzeugt neue Outbox-Version.
Worker sendet nur, wenn die erwartete Version noch aktuell ist.

## Device Revocation

Revoked Device Events dürfen keine neuen autorisierten Mutationen mehr
erzeugen.
