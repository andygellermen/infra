# Sprachkompass
## Managed Import Specification v0.1 – für Cody

**Status:** Implementierungsspezifikation  
**Datum:** 19. August 2026  
**Ziel:** Managed Bulk Import / Update für XLSX, CSV und später JSON/Sync  
**Kontext:** Admin-Bereich Sprachkompass / Sprach-a-lyzer

---

# 1. Zielbild

Der Managed Import ist **keine einfache Datei-Upload-Funktion**.

Er ist eine redaktionelle Import-Pipeline, mit der Admins große Datenmengen importieren, aktualisieren, validieren, zuordnen, vergleichen, konfliktfrei zusammenführen, wiederholen, protokollieren und später synchronisieren können.

Der Admin soll bei jedem Import jederzeit nachvollziehen können:

> **Was kommt aus der Quelldatei? Was existiert bereits in der Datenbank? Was wird neu angelegt? Was wird geändert? Was ist unklar? Was wird bewusst nicht importiert?**

---

# 2. Unterstützte Vorgangsarten

MVP:

```text
IMPORT
UPDATE
VALIDATE_ONLY
```

Später:

```text
SYNC
```

## IMPORT
Nur neue Datensätze. Bestehende Natural Keys führen zu Konflikt/Warnung.

## UPDATE
Bestehende Datensätze dürfen aktualisiert werden. Neue Datensätze können optional ebenfalls angelegt werden.

## VALIDATE_ONLY
Kein DB-Schreibvorgang. Nur Parsing, Mapping, Matching, Referenzprüfung, Guardrails, Diff und Golden-/Smoke-Test-Vorschau.

## SYNC – später
Quelle wird als maßgeblicher Datenbestand behandelt. Potentiell: neue Datensätze anlegen, vorhandene ändern, nicht mehr vorhandene deaktivieren. **Nicht im ersten MVP aktivieren.**

---

# 3. Preset-Konzept

Jeder durchgeführte Import-/Update-Vorgang kann als Preset gespeichert werden.

Standardname:

```text
{VORGANG} · {DATUM} · {UHRZEIT}
```

Beispiele:

```text
IMPORT · 19.08.2026 · 20:52
UPDATE · 21.08.2026 · 09:14
VALIDATE · 22.08.2026 · 16:07
SYNC · 04.11.2026 · 11:35
```

Optional umbenennbar:

```text
Seed DE – Modalverben v2
Google Sheet – Positive Patterns
Corporate Begriffe – Update August
```

---

# 4. Was ein Preset speichert

```yaml
ImportPreset:
  id
  name
  operation_type
  source_type
  source_sheet
  target_entity
  mapping_profile_id
  column_mapping
  matching_strategy
  conflict_policy
  import_options
  validation_options
  created_at
  created_by
  last_used_at
  usage_count
  version
```

Optional:

```yaml
source_fingerprint
original_filename
```

---

# 5. Unterschied Preset vs. Mapping-Profil

## Mapping-Profil
Beschreibt nur `Quellfeld → Zielfeld`.

Beispiel:

```text
Wort        → lemma
Bedeutung   → description
Kategorie   → pattern_class
Quelle      → source_key
```

## Import-Preset
Enthält zusätzlich Operation, Mapping-Profil, Matching-Regeln, Konfliktregeln, Importoptionen, Validierungsregeln, Ziel-Entity und Sheet-Auswahl.

---

# 6. Admin Navigation

```text
Importe
├── Neuer Import
├── Presets
├── Mapping-Profile
├── Import-Historie
├── Konflikte
└── Fehlgeschlagene Vorgänge
```

---

# 7. Wizard – Hauptablauf

```text
1 Quelle
2 Tabellenblätter / Datenbereich
3 Ziel-Entity
4 Feld-Mapping
5 Matching
6 Vorschau / Diff
7 Konflikte
8 Validierung
9 Import
10 Ergebnis / Preset speichern
```

---

# 8. Quelle

Unterstützt MVP:

```text
XLSX
CSV
JSON
```

Später:

```text
Google Sheets API
Remote URL
Object Storage
Direct Sync Connector
```

---

# 9. Tabellenblätter / Datenbereich

Bei XLSX werden Sheets erkannt, z. B.:

```text
[x] Lexemes
[x] Senses
[x] Phrases
[ ] Notes
[x] Relations
[x] DimensionContributions
```

Optional Range-Konfiguration:

```text
A1:J5000
```

---

# 10. Ziel-Entity

Automatische Vorschläge anhand von Sheet-Name, Headernamen, bekannten Import-Presets und Mapping-Profilen.

Beispiel:

```text
Sheet "Begriffe"
Vermutetes Ziel:
Lexemes   93 %
```

---

# 11. Feld-Mapping

| Quelle | Beispielwert | Ziel | Status |
|---|---|---|---|
| Wort | müssen | lemma | AUTO |
| Sprache | de-DE | language | AUTO |
| Wortart | Modalverb | part_of_speech | AUTO |
| Beschreibung | ... | description | AUTO |
| Bewertung | -20 | — | UNMAPPED |

Status:

```text
AUTO
MANUAL
UNMAPPED
IGNORED
INVALID
```

---

# 12. Mapping-Automatik

Reihenfolge:

1. exakter Name
2. Alias
3. normalisierte Schreibweise
4. bekannte Mapping-Profile
5. heuristische Ähnlichkeit

Beispiel Aliases:

```yaml
lemma:
  - wort
  - begriff
  - lexem
  - lexeme

description:
  - beschreibung
  - bedeutung
  - definition
  - erläuterung
```

---

# 13. Matching

Primär:

```text
Natural Key
```

Beispiele:

```text
lexeme_key
sense_key
phrase_key
relation_key
contribution_key
source_key
```

Sekundär, wenn Natural Key fehlt:

```text
Lexeme: language + lemma
Sense: lexeme_key + sense_key/title
Phrase: language + normalized text_or_pattern
Relation: relation_type + source_key + target_key
```

Match Confidence:

```text
EXACT
PROBABLE
AMBIGUOUS
NONE
```

Regeln:

```text
EXACT      → automatischer Match erlaubt
PROBABLE   → Admin-Bestätigung
AMBIGUOUS  → kein automatisches Merge
NONE       → neuer Datensatz
```

---

# 14. Diff-Vorschau

Jeder Datensatz erhält einen Status:

```text
NEW
UNCHANGED
CHANGED
CONFLICT
INVALID
SKIPPED
REFERENCE_MISSING
```

Beispiel:

```text
sense_key: sn_frei_liberty

Feld             Datenbank        Import
------------------------------------------------
title            Freiheit         Handlungsfreiheit
confidence       0.80             0.84
status           APPROVED         APPROVED
```

Aktionen:

```text
[ DB behalten ]
[ Import übernehmen ]
[ manuell bearbeiten ]
```

Feldweise Merge-Auflösung ist möglich.

---

# 15. Konflikt-Policy

Preset kann Standardregeln speichern:

```text
KEEP_DATABASE
USE_IMPORT
REQUIRE_MANUAL
KEEP_NEWER_VERSION
KEEP_HIGHER_EVIDENCE
```

Kritische Felder standardmäßig immer `REQUIRE_MANUAL`:

```text
evidence_class
dimension contribution value
relation_type
status = PRODUCTION
claim_type
hard_guardrail
```

---

# 16. Große Datenmengen

Statt Vollvorschau primär Summary:

```text
Datensätze gesamt: 52.486

UNCHANGED          44.209
NEW                 5.418
CHANGED             2.613
CONFLICT               97
INVALID                41
REFERENCE_MISSING      108
```

Frontend:

```text
server-side pagination
virtualized rows
lazy detail loading
```

---

# 17. Validierung

Ebenen:

```text
FILE
SCHEMA
FIELD
REFERENCE
DOMAIN
GUARDRAIL
GOLDEN
```

## FILE
Datei lesbar, Format unterstützt, keine beschädigte XLSX.

## SCHEMA
Pflichtfelder, Datentypen, bekannte Spalten, Schema-Version.

## FIELD
z. B. `confidence ∈ [0,1]`.

## REFERENCE
Referenzen dürfen auf existierende DB-Daten oder Datensätze im selben Batch zeigen.

## DOMAIN
z. B. keine semantische Vererbung über Homophonie.

## GUARDRAIL
Hard Guardrails blockieren Commit.

## GOLDEN
Golden Suite gegen simulierten neuen Datenbestand.

---

# 18. Dependency Graph

Importer bestimmt Abhängigkeiten automatisch:

```text
lx_muessen
   ↓
sn_muessen_internal_pressure
   ↓
dc_internal_pressure_fw
```

Zyklen:

```text
A → B → C → A
```

führen zu:

```text
INVALID_DEPENDENCY_CYCLE
```

---

# 19. Commit / Transaktion

Standard erst:

```text
VALIDATE_ONLY
```

danach:

```text
[ Import verbindlich durchführen ]
```

Batch atomar:

```text
BEGIN
  import lexemes
  import senses
  import phrases
  import relations
  import contributions
  validations
COMMIT
```

Bei kritischem Fehler:

```text
ROLLBACK
```

Für große Imports: Staging + Chunking.

---

# 20. Staging

Empfohlen:

```text
import_batches
import_batch_rows
import_staging_records
```

Staging enthält u. a.:

```text
raw_payload
normalized_payload
matched_target_id
diff
status
errors
warnings
```

---

# 21. Import Batch Schema

```sql
import_batches (
  id uuid primary key,
  batch_key text unique,
  operation_type text,
  status text,
  source_type text,
  original_filename text,
  source_fingerprint text,
  mapping_profile_id uuid,
  preset_id uuid,
  total_rows integer,
  new_rows integer,
  changed_rows integer,
  unchanged_rows integer,
  conflict_rows integer,
  invalid_rows integer,
  created_by uuid,
  created_at timestamptz,
  validated_at timestamptz,
  committed_at timestamptz
);
```

Status:

```text
UPLOADED
PARSED
MAPPED
MATCHED
VALIDATED
READY
RUNNING
COMPLETED
FAILED
ROLLED_BACK
CANCELLED
```

---

# 22. Import-Historie

Spalten:

```text
Name
Vorgang
Datum
Nutzer
Datei
Neu
Geändert
Konflikte
Fehler
Status
Preset
```

Detailseite:

```text
Batch Metadaten
Mapping
Matching
Diff Summary
Konfliktentscheidungen
Warnings
Errors
Golden Result
Commit Result
```

---

# 23. Preset nach Abschluss

Nach erfolgreichem Vorgang:

> Möchtest du diese Konfiguration als Preset speichern?

Standardname:

```text
UPDATE · 19.08.2026 · 20:52
```

Buttons:

```text
[ Preset speichern ]
[ Namen ändern ]
[ Nicht speichern ]
```

Optional:

```text
auto_save_import_presets = true
```

Empfehlung: `true`.

---

# 24. Preset-Wiederverwendung

Start neuer Import:

```text
[ Leerer Import ]
[ Preset verwenden ▼ ]
```

Preset lädt Operation Type, Mapping, Matching, Konfliktregeln, Validierung, Ziel-Entity und Sheet-Regeln. Die Quelldatei wird neu ausgewählt.

---

# 25. Preset-Versionierung

Änderungen nicht still überschreiben:

```text
Seed Pflege DE
v1
v2
v3
```

Status:

```text
ACTIVE
ARCHIVED
```

Später ggf.:

```text
SHARED
SYSTEM
```

---

# 26. Beispiel Import-Preset

```json
{
  "name": "UPDATE · 19.08.2026 · 20:52",
  "operation_type": "UPDATE",
  "source_type": "XLSX",
  "target_entity": "LEXEME",
  "mapping_profile": "Google Sheet – Lexemes DE",
  "matching": {
    "primary": ["lexeme_key"],
    "fallback": ["language", "lemma"]
  },
  "conflict_policy": "REQUIRE_MANUAL",
  "options": {
    "allow_insert": true,
    "allow_update": true,
    "skip_unchanged": true
  }
}
```

---

# 27. API

Upload:

```http
POST /api/v1/admin/imports
```

Parse:

```http
POST /api/v1/admin/imports/{batch_id}/parse
```

Mapping:

```http
PUT /api/v1/admin/imports/{batch_id}/mapping
```

Match:

```http
POST /api/v1/admin/imports/{batch_id}/match
```

Diff:

```http
GET /api/v1/admin/imports/{batch_id}/diff
```

Conflict Resolution:

```http
PATCH /api/v1/admin/imports/{batch_id}/rows/{row_id}
```

Validate:

```http
POST /api/v1/admin/imports/{batch_id}/validate
```

Commit:

```http
POST /api/v1/admin/imports/{batch_id}/commit
```

Presets:

```text
GET    /api/v1/admin/import-presets
POST   /api/v1/admin/import-presets
GET    /api/v1/admin/import-presets/{id}
PATCH  /api/v1/admin/import-presets/{id}
POST   /api/v1/admin/import-presets/{id}/clone
POST   /api/v1/admin/import-presets/{id}/archive
```

Mapping Profiles:

```text
GET    /api/v1/admin/mapping-profiles
POST   /api/v1/admin/mapping-profiles
PATCH  /api/v1/admin/mapping-profiles/{id}
```

---

# 28. Rollback / Audit

Änderungslog pro Batch:

```text
entity
entity_id
before
after
operation
```

Rollback:

```http
POST /api/v1/admin/imports/{batch_id}/rollback
```

Jede relevante Entscheidung auditieren:

```text
mapping changed
match manually confirmed
conflict resolved
validation override
commit
rollback
preset created
preset changed
```

---

# 29. Berechtigungen

Vorschlag:

```text
CONTRIBUTOR → Upload/Mapping/Dry Run
REVIEWER    → Konflikte bestätigen
PUBLISHER   → Commit
ADMIN       → Rollback/Settings
```

---

# 30. Duplicate Source Detection

Datei-Fingerprint:

```text
SHA-256
```

Bei identischer Datei:

> Diese Datei wurde bereits am 19.08.2026 um 20:52 importiert.

Optionen:

```text
[ Vorgang öffnen ]
[ trotzdem erneut validieren ]
[ abbrechen ]
```

UPSERT über Natural Keys muss idempotent sein.

---

# 31. Mapping Drift

Wenn ein Preset `Wort` erwartet, die neue Datei aber `Begriff` enthält:

```text
Mapping Drift erkannt
```

Alias-Vorschlag:

```text
Begriff → lemma
```

Preset erst nach Admin-Bestätigung aktualisieren.

---

# 32. Import Report

Nach Abschluss:

```text
Batch ID
Preset
Quelle
Zeit
Nutzer
Summary
Warnings
Errors
Golden Result
```

Optional Export:

```text
JSON
CSV
```

Fehlerexport:

```text
rejected_rows.csv
```

---

# 33. Fehlercodes

```text
UNKNOWN_FIELD
INVALID_VALUE
REFERENCE_NOT_FOUND
AMBIGUOUS_MATCH
HARD_GUARDRAIL_VIOLATION
DUPLICATE_NATURAL_KEY
DEPENDENCY_CYCLE
GOLDEN_REGRESSION
```

---

# 34. Definition of Done – Managed Import v0.1

- [ ] XLSX Upload
- [ ] CSV Upload
- [ ] JSON Upload
- [ ] Sheet-Erkennung
- [ ] Header-Erkennung
- [ ] Entity-Auswahl
- [ ] automatisches Feld-Mapping
- [ ] manuelles Feld-Mapping
- [ ] Mapping-Profile
- [ ] Natural-Key Matching
- [ ] sekundäres Matching
- [ ] Match Confidence
- [ ] Diff Preview
- [ ] Konfliktauflösung
- [ ] Summary für große Datenmengen
- [ ] Referenzprüfung
- [ ] Dependency Graph
- [ ] Hard Guardrails
- [ ] Golden Dry Run
- [ ] transaktionaler Commit
- [ ] Import History
- [ ] Presets
- [ ] Auto-Preset `VORGANG · DATUM · UHRZEIT`
- [ ] Preset umbenennbar
- [ ] Preset wiederverwendbar
- [ ] Preset versioniert
- [ ] Duplicate Source Detection
- [ ] Audit Log
- [ ] Fehlerexport
- [ ] Rollback Change Log
- [ ] Berechtigungen

---

# 35. Technische Priorisierung

## Sprint 1

```text
Upload
Parse
Mapping
Natural-Key Match
Diff
VALIDATE_ONLY
```

## Sprint 2

```text
Conflict UI
Commit
History
Preset Save/Load
```

## Sprint 3

```text
Golden Integration
Rollback
Mapping Profiles
Performance/Chunking
```

Später:

```text
SYNC
Google Sheets Connector
Scheduled Imports
Remote Sources
```

---

# 36. North Stars

> **Der Admin soll niemals hoffen müssen, dass ein Import richtig läuft. Er soll vor dem Commit sehen können, was passieren wird.**

> **Ein guter Bulk-Import spart Zeit, ohne Transparenz gegen Geschwindigkeit einzutauschen.**

> **Importierte Fachinformationen werden nicht allein deshalb wahr, weil sie in einer Datei stehen. Sie durchlaufen dieselben Evidenz-, Kontext- und Guardrail-Prinzipien wie alle anderen Wissensbestandteile des Sprachkompasses.**

# 37. Implementierungsstand v0.5.0

Die ausführbare Operations-Pipeline ist in
[`MANAGED-KNOWLEDGE-OPERATIONS-CLOSURE-v0.5.0.md`](../00-start/MANAGED-KNOWLEDGE-OPERATIONS-CLOSURE-v0.5.0.md)
dokumentiert. JSON/CSV/XLSX, Mapping, Matching, Diff, Konflikte,
Referenz-/Zyklusprüfung, Golden Dry Run, PostgreSQL-Commit, Rollback, History
und Audit sind geschlossen. `SYNC` und Remote Sources bleiben deaktiviert;
die visuelle Admin-Oberfläche und Preset-Verwaltung folgen im MVP Candidate.
