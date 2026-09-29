# Backup & Restore Strategy

## Ziel

Backup ist erst dann Backup, wenn Restore regelmäßig funktioniert.

## Ebenen

### PostgreSQL

-   regelmäßige Dumps / physische Backups
-   Point-in-Time-Recovery prüfen

### Object/Vault Store

-   Versioning/Retention nach Bedarf
-   zweite unabhängige Kopie

### Konfiguration

-   Ansible/Repo reproduzierbar
-   Secrets separat gesichert

### Client

Clientcache ist kein Backup.

## Restore Drill

Mindestens: 1. leere Testumgebung 2. DB Restore 3. Object Store
Restore/Attach 4. Integrity Sweep 5. Event/State Check 6. Stichproben
`.eml` Rekonstruktion 7. Outbox-Sicherheitsprüfung

## RPO/RTO

Pilotwerte explizit dokumentieren und später anhand realer Anforderungen
schärfen.
