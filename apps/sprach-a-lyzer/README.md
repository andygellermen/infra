# Sprach-A-Lyzer

Der Sprach-A-Lyzer ist der gemeinsame, deterministische und erklärbare
Analyse-Core für die Produktprofile **Sprachkompass** (Corporate) und
**MeineSprache** (Private).

## Einstieg

1. [START HERE](docs/00-start/START-HERE.md)
2. [Cody Handoff](docs/00-start/CODY-HANDOFF.md)
3. [Developer Handoff](docs/00-start/DEVELOPER-HANDOFF-v0.1.md)
4. [Roadmap](docs/00-start/ROADMAP.md)
5. [Documentation Manifest](docs/00-start/DOCUMENTATION-MANIFEST.md)
6. [Implementation Baseline v0.1](docs/00-start/IMPLEMENTATION-BASELINE-v0.1.md)
7. [Foundation v0.0](docs/00-start/FOUNDATION-v0.0.md)
8. [Modular Monolith](docs/70-architecture/sprach-a-lyzer_modular-monolith_v0.1.md)
9. [Canonical Dimensions](docs/20-domain-model/sprach-a-lyzer_canonical-dimensions_v0.1.md)
10. [Analyse- und Trace-Verträge](docs/20-domain-model/sprach-a-lyzer_analysis-trace-contracts_v0.1.md)
11. [Vertical-Slice Golden Gate](docs/40-golden/sprach-a-lyzer_vertical-slice-golden_v0.2.md)
12. [Sprint 0A Canonical Contracts](docs/00-start/SPRINT-0A-CANONICAL-CONTRACTS-v0.1.md)
13. [Foundation Rule Migration](docs/00-start/FOUNDATION-RULE-MIGRATION-v0.1.md)
14. [Foundation Runtime Binding](docs/00-start/FOUNDATION-RUNTIME-BINDING-v0.1.md)
15. [Core Closure v0.1.0](docs/00-start/CORE-CLOSURE-v0.1.0.md)
16. [Context & Proposition Vertical Slice v0.1](docs/00-start/CONTEXT-PROPOSITION-VERTICAL-SLICE-v0.1.md)
17. [Resolver Catalogue Contracts v0.1](docs/00-start/RESOLVER-CATALOGUE-CONTRACTS-v0.1.md)
18. [Resolver Catalogue Runtime Binding v0.1](docs/00-start/RESOLVER-CATALOGUE-RUNTIME-BINDING-v0.1.md)
19. [Resolver Relations & Scope Expansion v0.1](docs/00-start/RESOLVER-RELATIONS-SCOPE-EXPANSION-v0.1.md)
20. [Proposition-local Trace Binding v0.1](docs/00-start/PROPOSITION-LOCAL-TRACE-BINDING-v0.1.md)
21. [Construct Ontology Contracts v0.2](docs/00-start/CONSTRUCT-ONTOLOGY-CONTRACTS-v0.2.md)

## Repository-Struktur

```text
docs/
  00-start/          Einstieg, Handoffs, Roadmap und Dokumentenregeln
  10-product/        Produktkonzepte und Produktprofile
  20-domain-model/   Fachliches Datenmodell, Konstrukte und Pattern
  30-engine/         Resolver, Scoring, Assessability und Q/A-Engine
  40-golden/         Golden-Korpora, Simulationen und Gap-Berichte
  50-import/         Bulk- und Managed-Import-Spezifikationen
  60-ux/             UX- und Admin-Wireframes
  70-architecture/   Technologie-, Datenschutz- und Lizenzkonzepte

data/
  seed/              Kanonische Start- und Redaktionsdaten
  golden/            Maschinenlesbare Golden- und Simulationsdaten
  import-examples/   Importvorlagen und Beispiel-Batches

schemas/
  analysis/          Versionierte Analyse- und Trace-Verträge
  constructs/        Construct-Ontologie und Kompositionsverträge
  questions/         Q/A-Verträge
  imports/           Importverträge
  rules/             Regel-, Policy- und Guardrail-Verträge
```

## Artefaktregeln

- Markdown ist die lesbare fachliche Dokumentation.
- JSON ist das kanonische Austausch- und Testformat.
- CSV und XLSX sind redaktionelle Arbeits- und Importformate.
- Ausführbarer Anwendungscode wird außerhalb von `docs/`, `data/` und
  `schemas/` angelegt.
- Bestehende Versionen werden nicht überschrieben oder gelöscht.

## Fachliche Kompatibilität

Neue Verträge und der Go-Core verwenden die kanonische Dimension `VOLITION`.
Ältere, versionierte Fachdaten mit `FREE_WILL` bleiben unverändert und werden
an Importgrenzen testbar auf `VOLITION` abgebildet.

JSON- und CSV-Artefakte können ohne Veränderung ihrer Quelldatei geprüft und
normalisiert werden:

```bash
go run ./cmd/normalize-dimensions \
  -input data/seed/sprachkompass_coaching-question-pool_v0.1.json \
  > /tmp/question-pool-canonical.json
```

## Ausführbarer Vertical Slice

Der erste Go-Core implementiert die sechs Acceptance Cases aus `START-HERE`
mit Sense-Auflösung, Pattern-Erkennung, Assessability und Contribution Trace.

```bash
go test ./...

go run ./cmd/analyze \
  -context SELF_TALK \
  -text 'Ich muss das heute unbedingt noch schaffen.'

go run ./cmd/resolve \
  -context PRIVATE_CONVERSATION \
  -text 'Eigentlich wollte ich absagen, aber ich bin noch unsicher.'

go run ./cmd/analyze \
  -trace \
  -context SELF_TALK \
  -text 'Ich muss das heute unbedingt noch schaffen.'

go run ./cmd/analyze \
  -trace-v2 \
  -context SELF_TALK \
  -text 'Ich muss das heute unbedingt noch schaffen.'
```

Die maschinenlesbaren Request-, Ergebnis- und Trace-Verträge liegen unter
`schemas/analysis/`, die Golden Suite unter
`data/golden/sprach-a-lyzer_vertical-slice_v0.2.json`.

Die neun ausführbaren Foundation-Regeln liegen ohne fachliche Neukalibrierung
im strikten Rule-v0.4-Format unter
`data/seed/sprach-a-lyzer_foundation_v0.3.json`. Die Seed-Pipeline verwirft
unbekannte Felder und nicht registrierte Conditions oder Aktionen.

Im Serverpfad liest die Engine den aktiven `PRODUCTION` Rule Set aus PostgreSQL.
Standalone-CLI und Tests verwenden denselben Foundation Seed als eingebettetes
Build-Artefakt; dadurch benötigen lokale Smoke-Tests keine Datenbank. Der
geschlossene Core-Stand ist durch das Release-Manifest v0.1.0 und den
annotierten Git-Tag `v0.1.0` fixiert.

Der Context-/Proposition-Resolver lädt Resolver Catalogue v0.1 als
eingebettetes, strikt validiertes Runtime-Artefakt. `AMBIGUOUS`-Senses und
diagnostische Pattern-Kandidaten werden an der Rule-Engine-Grenze nicht zu
scorenden Fakten; Proposition-Spans werden vor jeder Rückgabe gegen den
Quelltext geprüft.

## PostgreSQL und HTTP-API

Die lokale Entwicklungsumgebung führt PostgreSQL, Migration, Seed und API in
der korrekten Reihenfolge aus:

```bash
docker compose up --build
```

Danach:

```bash
curl --fail-with-body http://localhost:8080/health/ready

curl --fail-with-body \
  -H 'Content-Type: application/json' \
  -d '{"text":"Ich muss das heute unbedingt noch schaffen.","context":"SELF_TALK"}' \
  http://localhost:8080/api/v1/analyze
```

Der Analyse-Request wird standardmäßig nicht persistiert. Migrationen,
Seed-Daten und Betriebsdetails sind in `docs/00-start/FOUNDATION-v0.0.md`
dokumentiert.

## Modulstruktur

Der Server bleibt ein gemeinsam deploybarer Go-Prozess. Die fachlichen
Bereiche sind dennoch durch Fassaden und Repository-Ports getrennt:

```text
internal/app/           Composition Root
internal/analysis/      öffentliche Analyse-Fassade
internal/resolver/      Context-, Proposition-, Sense- und Scope-Auflösung
internal/knowledge/     kanonischer Wissensbestand
internal/rules/         Rule Sets und Regelkatalog
internal/presentation/  Profile, Labels und sichere Fallbacks
internal/dimension/     kanonischer Shared Kernel und Legacy-Mapping
internal/httpapp/       HTTP-Adapter
internal/db/            PostgreSQL-Adapter
```

Ein automatischer Architekturtest schützt diese Abhängigkeitsrichtung.
