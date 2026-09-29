# Retention & Deletion Model

## Nachricht

``` text
ACTIVE
 -> TRASHED
 -> PURGE_ELIGIBLE
 -> PURGED
```

## Warum zweistufig

Synchronisation, versehentliche Löschung und offline Geräte benötigen
ein Tombstone-Fenster.

## Tombstone

Enthält nur die zur Synchronisation nötige Identität und
Purge-Information. Ein offline Gerät darf eine bereits gelöschte
Nachricht nicht wieder „auferstehen" lassen.

## Attachment

Attachment-Objekt darf erst purged werden, wenn: - keine aktive Referenz
mehr existiert, - Retention abgelaufen ist, - kein Legal/Recovery Hold
existiert, - Integrity Sweep dies bestätigt.

## Account Closure

Export anbieten -\> definierte Grace Period -\> Purge Pipeline.

## Backup

Backup-Retention und produktive Löschung sind getrennte Lebenszyklen.
Dokumentation muss transparent festhalten, wann Daten auch aus Backups
verschwinden.
