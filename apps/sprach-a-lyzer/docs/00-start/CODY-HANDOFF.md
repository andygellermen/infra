# Sprachkompass / Sprach-A-Lyzer – Cody Handoff v0.3

**Status:** MVP Initiation Package  
**Stand:** 30. August 2026
**Zweck:** Verbindliche Orientierung für die technische Initiierung des ersten MVP.

## 1. START HERE

Der Sprach-A-Lyzer ist kein Wörterbuch und kein generischer LLM-Wrapper.

Ziel ist eine deterministische, erklärbare Sprachreflexions-Engine:

```text
Text / Q&A
→ Language Features
→ Sense / Proposition
→ Pattern Rules
→ Q/A Composition
→ Construct Evidence
→ Dimension Contributions
→ Assessability
→ Explainable Result
```

> **LLM proposes; Rule Engine validates/scores.**

Der MVP muss **ohne generative KI** bereits glaubwürdig, nützlich und attraktiv sein.

## 2. MVP Scope Freeze

### MUST HAVE
- Go Core Backend
- PostgreSQL
- modularer Monolith
- Knowledge Base
- Managed Import
- Rule Engine
- Sense-/Proposition-Grundstruktur
- 6 Sprachdimensionen
- Assessability
- Contribution Trace
- Question Context
- Q/A Composition Engine
- MVP Question Core Set
- Standard/Easy Question Renderings
- Corporate/Private Presentation Bundles
- Golden Test Harness
- Audit/Versioning der fachlichen Daten
- Privacy Defaults
- Analyse ohne generative KI

### OPTIONAL PREVIEW
- freiwillige KI-Vertiefung hinter Feature Flag
- lokale oder externe AI-Adapter-Schnittstelle

### OUT OF MVP
- Prosodieanalyse
- tiefe Audio-Semantik
- freie generative Coaching-Dialoge
- Graph Database
- separate Vector Database
- langfristige Persönlichkeits-/Entwicklungsprofile
- Community-Wiki
- Browser Extension
- große Mehrsprachigkeit
- vollautomatische Regelgenerierung

## 3. Verbindliche Architekturprinzipien

1. Context > isoliertes Wort.
2. Unknown/Ambiguous > Scheingenauigkeit.
3. Missing Evidence != 50 %.
4. Question score bias = 0.
5. Fragekontext darf Confidence moderat stützen, aber keinen Score erzeugen.
6. Keine Trait-/Diagnosebehauptungen.
7. Corporate != kalte Sprache; Corporate bleibt menschlich, aber nicht invasiv.
8. Private darf tiefer gehen, spirituelle Ebene bleibt optional.
9. Canonical Question Intent ist von sichtbaren Renderings getrennt.
10. KI ist optional, nicht Voraussetzung.
11. Rohtexte werden standardmäßig nicht unnötig gespeichert.
12. Golden Tests sind Release-Gate.

## 4. Canonical Dimension IDs

```text
AGENCY
CONNECTION
APPRECIATION
CLARITY
VOLITION
OPENNESS
```

`FREE_WILL` künftig intern als `VOLITION`.

## 5. Dokumentenreihenfolge für Cody

### TIER 1 – MUST READ / IMPLEMENTATION BASIS
1. [`START-HERE.md`](START-HERE.md)
2. [`DEVELOPER-HANDOFF-v0.1.md`](DEVELOPER-HANDOFF-v0.1.md)
3. [`CODY-HANDOFF.md`](CODY-HANDOFF.md)
4. [`spiritual-language-analyzer_reference-engine_v0.4.md`](../30-engine/spiritual-language-analyzer_reference-engine_v0.4.md)
5. [`spiritual-language-analyzer_reference-simulation_v0.4.1.json`](../../data/golden/spiritual-language-analyzer_reference-simulation_v0.4.1.json)
6. [`sprachkompass_construct-review_v0.1.md`](../20-domain-model/sprachkompass_construct-review_v0.1.md)
7. [`sprachkompass_qa-composition-engine_v0.1.md`](../30-engine/sprachkompass_qa-composition-engine_v0.1.md)
8. [`sprachkompass_qa-composition-contract_v0.1.json`](../../schemas/questions/sprachkompass_qa-composition-contract_v0.1.json)
9. [`sprachkompass_mvp-question-core-set_v0.1.md`](../10-product/sprachkompass_mvp-question-core-set_v0.1.md)
10. [`sprachkompass_question-rendering-rephrasing-model_v0.1.md`](../30-engine/sprachkompass_question-rendering-rephrasing-model_v0.1.md)
11. [`sprachalyzer_technology-ai-architecture-review_v0.1.md`](../70-architecture/sprachalyzer_technology-ai-architecture-review_v0.1.md)
12. [`sprachkompass_managed-import-specification_v0.1.md`](../50-import/sprachkompass_managed-import-specification_v0.1.md)
13. [`sprachkompass_managed-import-contract_v0.1.json`](../../schemas/imports/sprachkompass_managed-import-contract_v0.1.json)

### TIER 2 – TEST / QUALITY BASIS
14. [`spiritual-language-analyzer_golden-test-corpus_v0.1.md`](../40-golden/spiritual-language-analyzer_golden-test-corpus_v0.1.md)
15. [`spiritual-language-analyzer_golden-gap-report_v0.4.1.md`](../40-golden/spiritual-language-analyzer_golden-gap-report_v0.4.1.md)
16. [`sprachkompass_question-golden-corpus_v0.1.md`](../40-golden/sprachkompass_question-golden-corpus_v0.1.md)
17. [`sprachkompass_question-golden-corpus_v0.1.json`](../../data/golden/sprachkompass_question-golden-corpus_v0.1.json)
18. [`sprachkompass_qa-golden-core_v0.1.xlsx`](../../data/golden/sprachkompass_qa-golden-core_v0.1.xlsx)
19. [`sprachkompass_spoken-everyday-corpus_v0.1.md`](../40-golden/sprachkompass_spoken-everyday-corpus_v0.1.md)

### TIER 3 – DATA / IMPORT BASIS
20. [`spiritual-language-analyzer_product-concept_v1.0_consolidated-data-model.md`](../20-domain-model/spiritual-language-analyzer_product-concept_v1.0_consolidated-data-model.md)
21. [`spiritual-language-analyzer_product-concept_v1.1_scoring-engine-admin-rules.md`](../30-engine/spiritual-language-analyzer_product-concept_v1.1_scoring-engine-admin-rules.md)
22. [`sprachkompass_bulk-import-spec_v0.1.md`](../50-import/sprachkompass_bulk-import-spec_v0.1.md)
23. [`sprachkompass_bulk-import-template_v0.1.xlsx`](../../data/import-examples/sprachkompass_bulk-import-template_v0.1.xlsx)
24. [`sprachkompass_bulk-import-example_v0.1.json`](../../data/import-examples/sprachkompass_bulk-import-example_v0.1.json)
25. [`sprachkompass_coaching-question-pool_v0.1.xlsx`](../../data/seed/sprachkompass_coaching-question-pool_v0.1.xlsx)
26. [`sprachkompass_coaching-question-pool_v0.1.json`](../../data/seed/sprachkompass_coaching-question-pool_v0.1.json)

### TIER 4 – PRODUCT / GOVERNANCE CONTEXT
27. [`sprachkompass_corporate-workshop-profile_v0.1.md`](../10-product/sprachkompass_corporate-workshop-profile_v0.1.md)
28. [`sprachkompass_spoken-dictation-mobile-profile_v0.1.md`](../10-product/sprachkompass_spoken-dictation-mobile-profile_v0.1.md)
29. [`sprachalyzer-license-concept.md`](../70-architecture/sprachalyzer-license-concept.md)
30. [`NEXT-STEPS-AND-IDEAS.md`](NEXT-STEPS-AND-IDEAS.md)
31. [`ROADMAP.md`](ROADMAP.md)

## 6. Repository-Struktur

```text
/docs
  /00-start
  /10-product
  /20-domain-model
  /30-engine
  /40-golden
  /50-import
  /60-ux
  /70-architecture

/data
  /seed
  /golden
  /import-examples

/schemas
  /analysis
  /questions
  /imports
  /rules
```

## 7. Noch offener Feinschliff vor Coding

Nur als Starttickets, nicht als Blocker:

- `FREE_WILL` → `VOLITION` Migration
- Construct Review in Enum-/Schema-Namen überführen
- Question Canonical vs. Rendering DB-Schema fixieren
- Raw-text-retention Default technisch festlegen
- Feature-Flag Registry initialisieren
- NLP Adapter Interface definieren
- Golden Suites vereinheitlichen
- Knowledge-Base-Versionierung festlegen
- Corporate fallback labels einbetten
- MVP Core Question IDs als Seed markieren

## 8. Definition of Ready für Sprint 0

- Repo erstellt
- Go Module initialisiert
- PostgreSQL verfügbar
- Migration Framework gewählt
- Grundkonfiguration vorhanden
- Canonical Dimension IDs angelegt
- Golden Harness Skeleton vorhanden
- erste sechs Acceptance Cases importierbar
- Import-/Seed-Verzeichnis festgelegt
- CI startet Tests

> **Der erste MVP soll nicht alles können. Er soll beweisen, dass unsere Kernidee ohne Magie funktioniert.**

## 9. Umsetzungsstand Sprint 0A

Die kanonische Auflösung der Starttickets ist in
[`SPRINT-0A-CANONICAL-CONTRACTS-v0.1.md`](SPRINT-0A-CANONICAL-CONTRACTS-v0.1.md)
festgeschrieben.

Abgeschlossen:

- `FREE_WILL` ist nur noch Legacy-Alias; neue Verträge verwenden `VOLITION`.
- Analyse-, Ergebnis- und Trace-Verträge v0.1 sind stabil.
- Rule Contract v0.5 ist die aktuelle strikte Authoring-Grenze.
- Canonical Question und sichtbares Rendering besitzen getrennte Verträge.
- Privacy Defaults und Feature Flags sind kanonisch registriert.
- Hard und Soft Guardrails sind getrennt und maschinenlesbar.
- Versionsvektor und Golden-gesicherter Publish-Ablauf sind festgelegt.
- Elf Foundation-Regeln liegen strikt typisiert als Rule v0.5 vor.
- Seed, PostgreSQL-Regelkatalog und Provenienzfelder konsumieren Rule v0.5.
- Server/API beziehen den aktiven Foundation-Regelkatalog zur Laufzeit aus
  PostgreSQL; CLI und Tests verwenden das eingebettete kanonische Seed-Artefakt.

Weiterhin eigene Folgeschritte:

- NLP Adapter Interface,
- Vereinheitlichung der historischen Golden Suites,
- spätere fachliche Regel- und Parameterkalibrierung.

## 10. Core Closure v0.1.0

Der Deterministic Language Core ist mit
[`CORE-CLOSURE-v0.1.0.md`](CORE-CLOSURE-v0.1.0.md) abgeschlossen. Aktueller
damaliger Release-Laufzeitvektor: Rule Contract v0.4, Foundation/Rule Set v0.3,
Policy Registry v0.3, Presentation Bundle v0.2, Golden Suite v0.2 und
DB-Schema 3.

Die früher noch fest im Core verbliebenen Enrichments sind nun katalogisiert.
Der nächste Roadmap-Meilenstein ist v0.2 Context & Proposition. Die fachliche
Kalibrierung bleibt ein bewusst eigener, Golden-gesicherter Folgesprint.

## 11. Context & Proposition – erster Vertikalschnitt

Der erste additive v0.2-Schritt ist in
[`CONTEXT-PROPOSITION-VERTICAL-SLICE-v0.1.md`](CONTEXT-PROPOSITION-VERTICAL-SLICE-v0.1.md)
dokumentiert. Resolver Contract v0.2, Proposition Graph, Source Ranges,
TargetType, ExpectationSource, Modalität, Negations-Scope, Sense-/Ambiguitäts-
Ergebnisse und sieben neue Golden-Fälle sind implementiert. Der v0.1-
Analysevertrag und seine Golden-Parität bleiben unverändert.

## 12. Resolver Catalogue Contracts – Sprint v0.2B-A

Die kanonische Vertragsbasis für den Catalogue-Ausbau ist in
[`RESOLVER-CATALOGUE-CONTRACTS-v0.1.md`](RESOLVER-CATALOGUE-CONTRACTS-v0.1.md)
dokumentiert. Policy Registry v0.4 registriert die Resolver-IDs und drei neue,
nicht editierbare Hard Guardrails. Resolver Catalogue v0.1 fixiert acht
Lexemfamilien, acht Relationsklassen, priorisierte Negations-Scope-Regeln und
die Sense-Schwellen. Schema, Seed und Go-Contracts sind gegeneinander
driftgesichert. Die Laufzeitbindung folgt bewusst erst in v0.2B-B.

## 13. Resolver Catalogue Runtime Binding – Sprint v0.2B-B

Die in
[`RESOLVER-CATALOGUE-RUNTIME-BINDING-v0.1.md`](RESOLVER-CATALOGUE-RUNTIME-BINDING-v0.1.md)
dokumentierte Laufzeitbindung ist abgeschlossen. Der eingebettete Catalogue
wird je Resolution Fail-Closed geladen und validiert. Lexeme/Senses,
Thresholds, Konnektoren und Scope-Cues steuern den Resolver tatsächlich. Die
drei Resolver-Guardrails sind an Resolver- und Engine-Grenze durchgesetzt.
Der additive Runtime-Golden-Stand v0.2 hält die konservative
Schwellenkorrektur fest; die Core-Golden-Suite bleibt paritätisch.

## 14. Relations & Scope Expansion – Sprint v0.2B-C

Die in
[`RESOLVER-RELATIONS-SCOPE-EXPANSION-v0.1.md`](RESOLVER-RELATIONS-SCOPE-EXPANSION-v0.1.md)
dokumentierte Ausbaustufe ist abgeschlossen. `CAUSE`, `CONSEQUENCE`,
`CONDITION`, `ADDITION` und `CORRECTION` sind nun neben den drei bisherigen
Relationsklassen ausführbar. Rekursives Catalogue-Splitting, neun neue
Relations-/Scope-Goldens und Rule-End-to-End-Smokes sichern den Stand.
Actor-, Modalitäts- und Proposition-Scope reagieren nachweislich auf ihre
aktiven Catalogue-Cues. Damit ist die proposition-lokale Target-/Expectation-
Auflösung samt Trace-Verknüpfung als nächste Ausbaustufe vorbereitet.

## 15. Proposition Context & Contribution Provenance – Sprint v0.2C-A

Die in
[`PROPOSITION-LOCAL-TRACE-BINDING-v0.1.md`](PROPOSITION-LOCAL-TRACE-BINDING-v0.1.md)
dokumentierte Bindung ist abgeschlossen. TargetType und ExpectationSource
werden je Proposition aufgelöst; das Aggregat bleibt konservativ. Positive
Resolver- und Regelfakten tragen ihre Proposition IDs bis in jede wirksame
Contribution. Der additive Analysis Trace v0.2 veröffentlicht diese Referenzen
und ist in Policy Registry v0.5 registriert. Analysis Result v0.1, Trace v0.1
und Resolver Result v0.2 bleiben serialisiert unverändert.

## 16. Construct Ontology Contracts – Sprint v0.2C-B

Die in
[`CONSTRUCT-ONTOLOGY-CONTRACTS-v0.2.md`](CONSTRUCT-ONTOLOGY-CONTRACTS-v0.2.md)
festgeschriebene Ontology v0.2 registriert 36 Constructs in vier
Inferenzebenen. Evidenz, Nicht-Evidenz, Claim Modes, Assessability und
zulässige/verbotene Aussagen sind maschinenlesbar. Kein Construct scoret
direkt. Policy Registry v0.6 schützt die Ontology mit drei neuen Hard
Guardrails. Die vorbereiteten Runtime-Signale und Kompositionen werden erst in
v0.2C-C ausführbar gebunden.

## 17. Construct Runtime & Proposition Composition – Sprint v0.2C-C

Die in
[`CONSTRUCT-RUNTIME-PROPOSITION-COMPOSITION-v0.1.md`](CONSTRUCT-RUNTIME-PROPOSITION-COMPOSITION-v0.1.md)
dokumentierte Runtime ist abgeschlossen. Ontology-Signale werden je
Proposition ausgewertet und über `construct`- beziehungsweise `composition`-
Fakten von Rule v0.5 konsumiert. Foundation v0.4 erkennt
`RESPECTFUL_BOUNDARY`, `AGENCY_RECOVERY` und `LEARNING_RECOVERY`
propositional. Nur die bereits Golden-kalibrierte respektvolle Grenze trägt
weiterhin bei; die neuen Recovery-Pattern bleiben nicht-scoring. Policy
Registry v0.7 fixiert den Laufzeitvektor.
