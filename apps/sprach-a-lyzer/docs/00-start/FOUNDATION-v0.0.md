# Sprach-A-Lyzer – Foundation v0.0

**Status:** APPROVED  
**Phase:** IMPLEMENTED  
**Stand:** 24. August 2026

## Gelieferter Unterbau

- Go-Modul und deterministischer Analyse-Core
- modularer Monolith mit expliziten Fachmodulen und zentraler Composition Root
- PostgreSQL-Verbindung über `database/sql` und `pgx`
- unveränderliche, eingebettete SQL-Migrationen mit SHA-256-Prüfsumme
- transaktionale Migrationen mit PostgreSQL Advisory Lock
- versionierte Basisobjekte für Knowledge, Rules, Parameters und Presentation
- Managed-Import-Staging und Audit-Grundtabelle
- idempotenter Foundation- und Golden-Seed-Lauf
- getrennte Private- und Corporate-Präsentationsbundles mit eigenen Fallbacks
- `POST /api/v1/analyze`
- Liveness und migrationsbewusste Readiness
- Docker-/Compose-Entwicklungsumgebung
- Golden-, HTTP-, Migrations-, Seed- und Guardrail-Tests in CI

## Privacy Default

Der Analyse-Endpunkt verarbeitet den Text nur im Arbeitsspeicher. Die
Foundation enthält absichtlich keine Tabellen für Analysen, Analyseergebnisse
oder Rohtexte. Request Bodies werden nicht protokolliert und Antworten tragen
`Cache-Control: no-store`.

## Befehle

```bash
go run ./cmd/migrate
go run ./cmd/seed
go run ./cmd/server
```

Die Befehle erwarten `SAL_DATABASE_URL`. Migration und Seed sind voneinander
getrennt, wiederholbar und können als Deployment-Jobs ausgeführt werden.

## HTTP

```http
GET  /health/live
GET  /health/ready
POST /api/v1/analyze
```

Readiness wird erst gemeldet, wenn PostgreSQL erreichbar ist und mindestens
die vom Binary benötigte Schemaversion angewendet wurde.
