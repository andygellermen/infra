# Sprachkompass – Product & Engineering Roadmap

**Status:** strategische Grobplanung  
**Stand:** 26. August 2026
**Prinzip:** Versionen markieren Fähigkeitspakete, keine festen Kalenderzusagen.

## v0.0 – Foundation
Repo, Build, PostgreSQL, Migrationen, CI, Basistabellen, Golden Harness, Seed Loader, Managed Import Skeleton.

**Exit:** erste Golden Cases laufen automatisiert.

## v0.1 – Deterministic Language Core
Token-/Phrase Recognition, Sense Candidates, Rule Engine, Contributions, Assessability, 6 Dimensions, Contribution Trace, erste 6 Vertical-Slice Acceptance Cases.

**Exit:** ein Beispielsatz kann vollständig erklärt werden.

**Status:** abgeschlossen am 26. August 2026; Release-Commit und annotierter
Tag `v0.1.0`, siehe [`CORE-CLOSURE-v0.1.0.md`](CORE-CLOSURE-v0.1.0.md).

## v0.2 – Context & Proposition
Proposition Graph, TargetType, ExpectationSource, Syntactic Scope, Negation/Modality, Ambiguity Resolver, Construct Ontology v0.2.

**Exit:** kontextabhängige Golden Cases bestehen ohne grobe Over-Assessment-Fehler.

**Status:** in Arbeit. Der erste additive Resolver-Vertikalschnitt ist mit
Proposition Graph, Source Ranges, ersten Relations-, Target-, Expectation-,
Modalitäts-, Negations- und Ambiguitätsfällen implementiert; siehe
[`CONTEXT-PROPOSITION-VERTICAL-SLICE-v0.1.md`](CONTEXT-PROPOSITION-VERTICAL-SLICE-v0.1.md).
Die Vertragsbasis für die nächste Resolver-Ausbaustufe ist mit Resolver
Catalogue v0.1, Policy Registry v0.4 und drei neuen Hard Guardrails fixiert;
siehe [`RESOLVER-CATALOGUE-CONTRACTS-v0.1.md`](RESOLVER-CATALOGUE-CONTRACTS-v0.1.md).
Die Catalogue Runtime Binding einschließlich Fail-Closed-Provider,
katalogberechneter Sense-Zustände und durchgesetzter Resolver-Guardrails ist
ebenfalls abgeschlossen; siehe
[`RESOLVER-CATALOGUE-RUNTIME-BINDING-v0.1.md`](RESOLVER-CATALOGUE-RUNTIME-BINDING-v0.1.md).

## v0.3 – Question / Answer MVP
Canonical Questions, MVP Core Questions, Question Context, Answer Relevance, Q/A Composition, C0–C3 Inference Levels, Question Golden Corpus, adaptive Auswahl.

**Exit:** 5–8 Fragen können eine erklärbare Session bilden.

## v0.4 – Rendering & Corporate/Private Profiles
Corporate/Private Standard/Easy, Deep Reflective optional, Simplify/Rephrase, Leadingness/Risk/Intimacy, Presentation Bundle Isolation.

**Exit:** gleicher Canonical Core ohne ungewolltes Profile-Leakage.

## v0.5 – Managed Knowledge Operations
XLSX/CSV/JSON Import, Mapping, Matching, Diff, Conflicts, Validation, Presets, Golden Dry Run, Commit/Rollback, Audit.

**Exit:** Fachdaten können sicher ohne direkte DB-Manipulation gepflegt werden.

## v0.6 – MVP Candidate
End-to-End UI, Privacy Defaults, no-AI Core Experience, Session Flow, Result Explanation, Feedback/Alternatives, Admin Basis.

**Exit:** Nutzer würden den Sprach-A-Lyzer auch ohne KI erneut verwenden wollen.

## v0.7 – Optional AI Enhancement
AI Adapter, Consent, KI-Erklärung, individuelle Rephrasing-Vorschläge, Core-Trace-Validierung, ggf. Local LLM Pilot.

**Exit:** KI steigert Mehrwert, ohne Core Scores autonom zu verändern.

## v0.8 – Spoken Language
Spoken Dictation, lokale ASR Option, Spoken Features, ASR Confidence, Spoken Golden Corpus.

**Exit:** gesprochene Inputs liefern reproduzierbare Zusatzinformationen.

## v0.9 – Learning Product
Question Friction Metrics, Rendering Quality Review, Session Trajectories, Variant Comparison, bessere adaptive Fragewahl.

**Exit:** Produktpflege basiert auf aggregierten Nutzungsbeobachtungen.

## v1.0 – Stable Public Core
Stabile APIs/Schemas, Migration Policy, Lizenz/Trademark/Governance, dokumentierte Golden Baseline, reproduzierbare Releases, Security/Privacy Review.

## Roadmap-Regel

Ein Feature wandert nur aus
[`NEXT-STEPS-AND-IDEAS.md`](NEXT-STEPS-AND-IDEAS.md) in die Roadmap, wenn:

- fachlicher Nutzen geklärt,
- technische Abhängigkeiten verstanden,
- Datenschutzwirkung bewertet,
- Golden-/Acceptance-Strategie vorhanden,
- MVP nicht unnötig aufgebläht wird.
