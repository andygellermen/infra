# Sprach-A-Lyzer – Context & Proposition Closure v0.2.0

- **Status:** RELEASED
- **Version:** 0.2.0
- **Stand:** 1. September 2026
- **Git-Tag:** `v0.2.0`
- **Owner:** Product & Engineering

## Ergebnis

Der Roadmap-Meilenstein **v0.2 – Context & Proposition** ist geschlossen. Der
deterministische Core löst Propositionen, Source Ranges, Relationsklassen,
Actor, TargetType, ExpectationSource, Modalität, Negations-Scope und
Ambiguitäten kataloggestützt auf. Proposition-lokale Evidenz bleibt bis zum
Contribution Trace erhalten.

Construct Ontology v0.2 ist fail-closed an die Runtime gebunden. Ihre drei
Kompositionen sind ausführbar; ausschließlich der bereits zuvor kalibrierte
`RESPECTFUL_BOUNDARY`-Fall trägt zum Score bei. `AGENCY_RECOVERY` und
`LEARNING_RECOVERY` bleiben erklärbare, nicht-scorende Pattern.

## Release-Vektor

| Artefakt | Version |
|---|---:|
| Core Release | 0.2.0 |
| HTTP API | 2 |
| Analysis Request / Result | 0.1 |
| Analysis Trace | 0.2 |
| Resolver Result / Catalogue | 0.2 / 0.1 |
| Construct Ontology | 0.2 |
| Rule Contract / Foundation Rule Set | 0.5 / 0.4 |
| Policy Registry | 0.7 |
| Parameter Contract / Set | 0.1 / 0.1 |
| Presentation Bundle | 0.2 |
| Core / Resolver Golden Suite | 0.2 / 0.3 |
| Proposition Trace / Construct Composition Golden | 0.1 / 0.1 |
| PostgreSQL Schema | 4 |

Die maschinenlesbare Entsprechung liegt in
`data/seed/sprach-a-lyzer_release-manifest_v0.2.0.json`.

## Öffentliche HTTP-Verträge

Die v0.1-Analysegrenze bleibt unverändert:

```text
POST /api/v1/analyze → Analysis Result v0.1
```

Die neuen additiven v0.2-Routen sind:

```text
POST /api/v2/resolve → Resolver Result v0.2
POST /api/v2/trace   → Analysis Trace v0.2
```

Alle Routen akzeptieren weiterhin Analysis Request v0.1, wenden dieselben
Defaults und Validierungen an und persistieren den übergebenen Text nicht. Die
v0.2-Antworten veröffentlichen ihre Vertragsversion sowohl im JSON-Feld
`contract_version` als auch über die Header
`X-Sprach-A-Lyzer-Contract` und
`X-Sprach-A-Lyzer-Contract-Version`. Der Header
`X-Sprach-A-Lyzer-Version` bezeichnet den Core Release.

## Closure-Gate

Das konsolidierte Gate wird lokal und in GitHub Actions identisch ausgeführt:

```bash
bash ./scripts/verify-v0.2-closure.sh
```

Es prüft:

- Schema-/Go-Contract-Parität für Resolver Result und Analysis Trace,
- öffentliche HTTP-v1-Rückwärtskompatibilität und HTTP-v2-Versionierung,
- sechs unveränderte Core-Golden-Fälle,
- Resolver-Golden v0.3 einschließlich Relations- und Scope-Matrix,
- proposition-lokale Trace- und Construct-Composition-Goldens,
- Rule-, Resolver- und Composition-Aktivierung/Deaktivierung/Reaktivierung,
- eingebettete Migrationen bis Datenbankschema 4,
- aktuellen und historischen Release-Manifest-Vektor.

Die vollständige CI führt anschließend zusätzlich `go vet ./...` und
`go test -race ./...` mit PostgreSQL 17 aus.

## Guardrails und bewusste Grenzen

- Analysis Result v0.1 und Trace v0.1 bleiben serialisiert unverändert.
- Ambige Resolver-Kandidaten können die Rule-Guardrails nicht umgehen.
- Construct-Evidenz erzeugt ohne katalogisierte Rule keine Contribution.
- Reflective Constructs und Recovery-Pattern erhalten keinen verdeckten Score.
- Fehlerhafte oder nicht verfügbare Catalogue-/Ontology-Provider brechen
  fail-closed ab.
- Dieser Release enthält keine neue fachliche Regel- oder
  Parameterkalibrierung.

## Versions- und Tag-Regel

Der annotierte Tag `v0.2.0` wird erst nach vollständig grünem Closure-Gate auf
dem Release-Commit erstellt. Das Manifest v0.1.0 und der Tag `v0.1.0` bleiben
unverändert erhalten.

## Nächster Roadmap-Schritt

Als nächstes beginnt **v0.3 – Question / Answer MVP**. Vor seiner Umsetzung
wird das Capability-Paket in kanonische Question-, Session- und
Inference-Verträge zerlegt. Die reservierte fachliche Kalibrierung bleibt ein
eigenes, Golden-gesichertes Arbeitsfenster.
