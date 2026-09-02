# Sprach-A-Lyzer – Managed Knowledge Operations Closure v0.5.0

- **Status:** RELEASED
- **Version:** 0.5.0
- **Stand:** 2. September 2026
- **Git-Tag:** `v0.5.0`
- **Owner:** Product & Engineering

## Ergebnis

Der Roadmap-Meilenstein **v0.5 – Managed Knowledge Operations** ist
geschlossen. Fachdatensätze können als JSON, CSV oder XLSX geprüft, gemappt,
verglichen und nach expliziter Freigabe atomar in den Managed-Knowledge-
Bestand übernommen werden. Direkte Tabellenmanipulation ist für diesen
Arbeitsablauf nicht erforderlich.

Die Pipeline lautet:

```text
Quelle + SHA-256
→ Parse
→ automatisches/manuelles Mapping
→ Natural-Key-/Sekundärmatching
→ feldweiser Diff
→ Konflikte + kritische Feldprüfung
→ Referenzen + Dependency Graph
→ Core Golden Dry Run
→ READY / VALIDATED / FAILED
→ transaktionaler Commit oder kein Write
→ reversibler Change Log + unveränderliches Audit
```

`SYNC_LATER` ist weiterhin deaktiviert.

## Daten- und Mappinggrenze

Ein Importziel besitzt `target_entity`, einen stabilen `natural_key`, Version,
Status, Payload und optionale Referenzen. Dadurch unterstützt dieselbe
Pipeline Lexeme, Senses, Phrases, Relations, Contributions, Sources,
Question Renderings und Presentation Entries, ohne entity-spezifische
Merge-Heuristik im HTTP-Adapter zu duplizieren.

Automapping erkennt `*_key`, `key`, `id`, Version, Status und Referenzen.
Explizites `column_mapping` hat Vorrang. Legacy `FREE_WILL` wird an der
Importgrenze nach `VOLITION` normalisiert und im Plan sichtbar gemacht.

## Matching, Diff und Konflikte

- Exakte Natural Keys ergeben `EXACT`.
- Konfigurierbare Sekundärfelder können `PROBABLE` oder `AMBIGUOUS` ergeben;
  beide werden nicht automatisch zusammengeführt.
- Diffs enthalten Datenbank- und Importwert je geändertem Feld.
- Policies: `KEEP_DATABASE`, `USE_IMPORT`, `REQUIRE_MANUAL`,
  `KEEP_NEWER_VERSION`, `KEEP_HIGHER_EVIDENCE`.
- Kritische Änderungen an Evidenzklasse, Dimension, Contribution-Wert,
  Relation Type, Claim Type, Hard Guardrail oder `PRODUCTION`-Status benötigen
  eine explizite `USE_IMPORT`-Entscheidung durch mindestens `REVIEWER`.

Doppelte Natural Keys, fehlende Referenzen und Dependency-Zyklen blockieren
den Batch. Identische Quellen werden über SHA-256 erkannt.

## Golden Dry Run

Jeder Plan führt die sechs Core-Golden-Fälle v0.2 mit der aktuellen
deterministischen Engine aus. Ein Fehler markiert den Batch `FAILED` und
verhindert Commit. Kandidatenspezifische Simulationsregressionen werden
zusätzlich ausgewiesen. Das historische Bulk-Beispiel validiert zwölf Lexeme
ohne Regression.

## Transaktion, Rollback und Audit

Schema 5 ergänzt:

- `managed_knowledge_records`,
- `import_batch_rows`,
- `import_change_log`,
- erweiterte `import_batches`,
- `mapping_profiles` und `import_presets` als versionierte Operationsbasis.

Commit ist ausschließlich für `PUBLISHER` oder `ADMIN` und nur aus `READY`
zulässig. `VALIDATE_ONLY` kann niemals committed werden. Vorher-/Nachher-
Payloads werden innerhalb derselben PostgreSQL-Transaktion protokolliert.
Rollback ist ausschließlich für `ADMIN` möglich und wendet den Change Log in
umgekehrter Reihenfolge an. `audit_events` sind per Datenbank-Trigger gegen
Update und Delete geschützt.

Der HTTP-Adapter erwartet Rollen aus dem vertrauenswürdigen Admin-/Auth-
Kontext. Vor öffentlichem Betrieb muss der vorgeschaltete Auth-Adapter diese
Werte setzen; frei zugängliche Bereitstellung der Admin-Routen ist nicht
zulässig.

## Öffentliche Admin-Verträge

```text
POST /api/v5/admin/imports/prepare
POST /api/v5/admin/imports/commit
POST /api/v5/admin/imports/rollback
GET  /api/v5/admin/imports
GET  /api/v5/admin/imports/{batch_id}/audit
```

Der lokale Befehl `go run ./cmd/importctl` unterstützt sichere Dry Runs für
JSON, CSV und XLSX. Eine produktive Serverinstanz verwendet den PostgreSQL-
Adapter; Tests und lokale Einmaloperationen können den Memory-Adapter nutzen.

## Release-Vektor

| Artefakt | Version |
|---|---:|
| Core Release / HTTP API | 0.5.0 / 5 |
| Managed Import Request / Plan / Operation | 0.1 / 0.1 / 0.1 |
| Core Golden Suite | 0.2 |
| PostgreSQL Schema | 5 |
| Vorgängerfähigkeit | vollständig v0.4-kompatibel |

Die maschinenlesbare Entsprechung liegt in
`data/seed/sprach-a-lyzer_release-manifest_v0.5.0.json`.

## Closure-Gate

```bash
bash ./scripts/verify-v0.5-closure.sh
```

Das Gate umfasst alle v0.4-Nachweise sowie Contract-Parität, JSON/CSV/XLSX,
Mapping, Matching, Diff, Konflikte, Referenzen, Zyklen, Duplicate Source,
Golden Dry Run, Role Gates, `VALIDATE_ONLY`, HTTP v5, Schema 5 und – mit
PostgreSQL in CI – Commit, Rollback und Audit-Unveränderlichkeit.

## Bewusste Grenzen

- Kein `SYNC`, Remote Connector oder zeitgesteuerter Import.
- Keine fachliche Autopromotion in aktive Rule Sets; deren Veröffentlichung
  bleibt ein eigener expliziter Freigabeschritt.
- Tabellen für Presets und Mapping-Profile sind versioniert vorbereitet;
  ihre Admin-UI gehört zum v0.6-End-to-End-Produkt.
- Rohtexte aus Nutzeranalysen gehören niemals in diese Wissensimport-Pipeline.

## Nächster Roadmap-Schritt

Als nächstes folgt **v0.6 – MVP Candidate** mit End-to-End UI, Session Flow,
Result Explanation, Feedback/Alternativen und Admin-Basis.
