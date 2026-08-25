# Sprachkompass – Repository Documentation Manifest

**Stand:** 21. August 2026

## Root-Einstieg

```text
README.md
docs/00-start/START-HERE.md
docs/00-start/CODY-HANDOFF.md
docs/00-start/DEVELOPER-HANDOFF-v0.1.md
docs/00-start/SPRINT-0A-CANONICAL-CONTRACTS-v0.1.md
docs/00-start/FOUNDATION-RULE-MIGRATION-v0.1.md
docs/00-start/FOUNDATION-RUNTIME-BINDING-v0.1.md
docs/00-start/ROADMAP.md
docs/00-start/NEXT-STEPS-AND-IDEAS.md
```

Alle Pfade sind relativ zum App-Root `apps/sprach-a-lyzer/` angegeben.

## Dokumentstatus

```text
DRAFT
REVIEW
APPROVED
REFERENCE
SUPERSEDED
ARCHIVED
```

Zusätzlich je Dokument:

```text
version
last_updated
owner
supersedes
```

## Regel für Cody

Implementierungsrelevant ist nur:

- APPROVED
- REFERENCE

`DRAFT` darf als fachlicher Kontext gelesen, aber nicht stillschweigend als verbindliche Spezifikation implementiert werden.

## Versionierung

Keine alten Spezifikationen löschen.

```text
v0.1 → SUPERSEDED by v0.2
```

## Golden Files

Golden Inputs und erwartete Ergebnisse gehören versioniert ins Repository und in CI.

## Datenartefakte

```text
XLSX = redaktionelles Arbeitsformat
JSON = kanonisches Austausch-/Testformat
DB   = Laufzeit-/Produktionsspeicher
```

## Ablageorte

```text
docs/    = lesbare Spezifikationen und fachliche Dokumentation
data/    = kanonische Seed-, Golden- und Importartefakte
schemas/ = maschinenlesbare Verträge und künftige JSON-Schemas
```
