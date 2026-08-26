# Sprach-A-Lyzer – Context & Proposition Vertical Slice v0.1

- **Status:** APPROVED
- **Version:** 0.1
- **Stand:** 26. August 2026
- **Roadmap-Ziel:** v0.2 – Context & Proposition
- **Basiert auf:** Core Release v0.1.0

## Ergebnis

Der erste additive v0.2-Vertikalschnitt ist implementiert. Er führt vor dem
Scoring einen deterministischen Resolver aus, ohne den veröffentlichten
Analyse-, Ergebnis- oder Trace-Vertrag v0.1 zu verändern.

Der Resolver liefert gemäß Contract v0.2:

- Proposition-Nodes mit stabilen IDs und Quellspannen,
- gerichtete Graph-Kanten mit Diskursrelation und Konnektor,
- Akteur, Predicate-/Target-/Time-/Boundary-/Decision-Features,
- Modalität und Negations-Scope,
- TargetType und ExpectationSource,
- ausgewählte Senses mit Confidence, Abstand und Zustand,
- explizite Ambiguitätsprofile,
- Pattern-Kandidaten sowie eine Resolver-Gesamtconfidence.

`source_start` und `source_end` sind nullbasierte UTF-8-Bytepositionen; das
Intervall ist halb-offen. `text[source_start:source_end]` muss daher exakt dem
Node-Text entsprechen.

## Laufzeitanbindung

Die Rule Engine konsumiert Resolver-Fakten vor der Regelausführung. Rule-v0.4-
Conditions können dadurch die bereits vorgesehenen Felder verwenden:

```text
selected_sense
target_type
expectation_source
discourse_relation
proposition_feature
```

Ein Smoke-Test weist die Kombination von TargetType, Sense und Proposition-
Feature sowie von Erwartungsquelle, Relation und Zeitfeature nach. Resolver-
Pattern werden nicht ungeprüft als ausgeführte Rule-Pattern übernommen; damit
bleiben Aktivierung und Deaktivierung der Scoring-Regeln wirksam.

## Verträge und Bedienung

- Contract:
  `schemas/analysis/sprach-a-lyzer_resolver-result_v0.2.json`
- Golden Suite:
  `data/golden/sprach-a-lyzer_context-proposition_v0.1.json`
- CLI:

```bash
go run ./cmd/resolve \
  -context PRIVATE_CONVERSATION \
  -text 'Eigentlich wollte ich absagen, aber ich bin noch unsicher.'
```

Der Resolver ist zusätzlich über die Analyse-Fassade als `Resolve` verfügbar.
Andere Feature-Module dürfen weder Resolver noch Engine oder Domain direkt
importieren; der Architekturtest schützt diese Grenze.

## Golden-Abdeckung

Die sieben initialen Fälle sichern:

- gesetzliche Erwartungsquelle,
- respektvolle Grenze mit Konzession,
- diskontierendes „Ja, aber“,
- internalisierte Erwartung,
- Personenlabel und TargetType `PERSON`,
- semantische Mehrdeutigkeit von `müssen` bei eindeutigem `umfahren`,
- negierte Erlaubnis mit Negations-Scope `MODALITY`.

Parallel bleibt die vollständige Golden-v0.2-Parität des Core Release v0.1.0
erhalten.

## Bewusste Grenze dieses Schritts

Dies ist noch nicht der Abschluss des Roadmap-Meilensteins v0.2. Noch offen:

- weitere Relationsmarker, Satzgefüge und komplexere Scope-Fälle,
- ein versionierter Sense-/Connector-Katalog statt aller Startheuristiken im
  Code,
- proposition-lokale Target-/Expectation-Auflösung für komplexe Texte,
- Construct Ontology v0.2,
- öffentliche HTTP-Versionierung für den Resolver-Vertrag,
- Verknüpfung von Contribution Trace und Proposition IDs,
- ein eigener Release-Manifest- und Tag-Stand `v0.2.0` erst nach Closure.

## Empfohlener nächster Sprint

**v0.2B – Resolver Catalogue & Scope Expansion**:

1. Sense-, Konnektor- und Scope-Definitionen strikt versionieren,
2. `CAUSE`, `CONSEQUENCE`, `CONDITION`, `ADDITION` und `CORRECTION` Golden-sichern,
3. Actor-, Modalitäts- und Negations-Scope-Schutzfälle erweitern,
4. proposition-lokale Target-/Expectation-Fakten erzeugen,
5. diese Fakten über aktivierbare Regeln und Trace-Referenzen end-to-end prüfen.
