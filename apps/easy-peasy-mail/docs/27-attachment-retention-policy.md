# Attachment Retention Policy

## Inbound Attachments

Standardmäßig in den zentralen Attachment Vault extrahieren, aber
Original-MIME oder rekonstruierbare Darstellung gemäß Retention-Policy
erhalten.

## Client Cache

-   kleine häufig genutzte Dateien: LRU Cache
-   `pin offline`: nie automatisch entfernen
-   große Dateien: standardmäßig remote
-   Cache-Limit pro Gerät konfigurierbar

## Vault Lifecycle

Mögliche Klassen:

``` text
ACTIVE
ARCHIVED
QUARANTINED
DELETED_PENDING
PURGED
```

## Löschen

Eine Mail-Löschung erzeugt zunächst Tombstone/Retention. Erst nach
Ablauf wird geprüft, ob ein Attachment-Objekt noch von anderen
Nachrichten referenziert wird.

## Dedupe

Reference Counting darf nicht alleinige Sicherheitslogik sein;
periodischer Integrity Sweep prüft DB-Referenzen gegen Object Store.

## Exit

Nutzer müssen Mail + Attachments in interoperabler Form exportieren
können, z. B. rekonstruierte `.eml` plus optional Maildir.
