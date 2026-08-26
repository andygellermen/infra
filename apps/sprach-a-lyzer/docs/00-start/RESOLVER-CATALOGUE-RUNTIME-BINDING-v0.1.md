# Sprach-A-Lyzer – Resolver Catalogue Runtime Binding v0.1

**Status:** APPROVED
**Version:** 0.1
**Last updated:** 26. August 2026
**Owner:** Sprach-A-Lyzer Core
**Roadmap-Ziel:** v0.2 – Context & Proposition
**Sprint:** v0.2B-B – Catalogue Runtime Binding

## Ergebnis

Der deterministische Context-/Proposition-Resolver konsumiert Resolver
Catalogue v0.1 nun im tatsächlichen Laufzeitpfad. Der Catalogue ist kein bloßes
Begleitdokument mehr: Ohne gültigen APPROVED Stand findet keine Resolution
statt.

## Laufzeitkette

```text
embedded Resolver Catalogue v0.1
  → strikter Decoder
  → CatalogueProvider.Active
  → erneute Fail-Closed-Validierung je Resolution
  → Runtime-Indizes
  → Proposition-/Sense-/Scope-Resolution
  → Span-Guardrail
  → gefilterte Resolver-Fakten
  → Rule Engine
```

Standalone-CLI, Tests und Server verwenden dasselbe eingebettete kanonische
Artefakt. Der Provider bildet zugleich den Test- und späteren Repository-Seam.
Für diesen Sprint ist keine Datenbankmigration nötig; es gibt weiterhin genau
eine produktive, reproduzierbare Resolver-Konfiguration im Build.

## Kataloggebundene Funktionen

- Lexemformen entscheiden, ob eine der acht Familien aktiv ist.
- Ein Resolver darf nur Senses aus der aktiven Lexemdefinition ausgeben.
- `HIGH`, `MEDIUM` und `AMBIGUOUS` werden aus den Catalogue-Schwellen
  `0,75/0,20` und `0,60/0,10` berechnet.
- Konnektormarker liefern Relation und Confidence für Proposition Edges.
- Scope-Cues werden nach Catalogue-Priorität ausgewertet. Mehrwort-Cues dürfen
  grammatische Zwischenwörter enthalten, bleiben aber in ihrer Reihenfolge.
- Catalogue-, Provider- oder Validierungsfehler brechen die Analyse
  Fail-Closed ab.

Die deterministischen Entscheidungsrezepte bleiben Code. Fachliche Vokabulare,
zulässige Ausgaben, Schwellen, Relationsmarker und Scope-Cues stammen dagegen
aus dem versionierten Catalogue.

## Durchgesetzte Guardrails

### `AMBIGUOUS_FEATURE_CANNOT_HARD_SCORE`

Nur `HIGH`- und `MEDIUM`-Senses werden zu adressierbaren Rule-Fakten.
`AMBIGUOUS` bleibt im Resolver Result sichtbar, kann aber weder eine
`selected_sense`-Condition noch eine Contribution auslösen.

### `PROPOSITION_SPAN_MUST_MATCH_SOURCE`

Vor Rückgabe prüft der Resolver für jeden halb-offenen Bytebereich:

```text
result.text[source_start:source_end] == proposition.text
```

Ein Fehler verwirft das gesamte Resolver Result.

### `RESOLVER_CANDIDATE_CANNOT_BYPASS_RULES`

`pattern_candidates` sind diagnostische Kandidaten. Die Engine übernimmt sie
nicht in ihren aktiven Pattern-Fakten. Erst eine aktivierte, validierte Regel
darf daraus ein Pattern oder eine Contribution erzeugen.

## Bewusste Golden-Weiterentwicklung

Resolver Golden v0.1 bleibt unverändert erhalten. Der additive Runtime-Stand
v0.2 basiert darauf und überschreibt nur die drei Resultate, deren bisher
manuell gesetzter Zustand den jetzt verbindlichen Schwellen widersprach:

| Fall | Zustand im Runtime-Stand |
|---|---|
| `CP04_INTERNALIZED_EXPECTATION` | `MEDIUM` |
| `CP05_PERSON_TARGET` | `AMBIGUOUS` |
| `CP07_NEGATED_PERMISSION` | `AMBIGUOUS` |

Bei den beiden mehrdeutigen Fällen sinkt die Resolver-Gesamtconfidence auf
`0,72`. Dies ist eine konservative Guardrail-Korrektur, keine Kalibrierung der
sechs Dimensionen. Die öffentliche Core-Golden-Suite v0.2 bleibt vollständig
paritätisch.

## Test- und Abnahmeumfang

- eingebetteter Catalogue und striktes Fail-Closed-Laden,
- Sense-Deaktivierung und anschließende Reaktivierungsparität,
- veränderte Threshold-, Connector- und Scope-Varianten,
- ungültiger und nicht verfügbarer Catalogue,
- absichtlich fehlerhafter Proposition-Span,
- blockierte Rule-Condition und Contribution für `AMBIGUOUS`,
- blockierter Pattern-Candidate-Bypass,
- sieben Context-/Proposition-Runtime-Goldens,
- sechs unveränderte Core-Goldens,
- bestehender Rule-Aktivierungs-/Deaktivierungs-Smoke-Test.

## Kanonische Dateien

- [`sprach-a-lyzer_resolver-catalogue_v0.1.json`](../../data/seed/sprach-a-lyzer_resolver-catalogue_v0.1.json)
- [`sprach-a-lyzer_context-proposition-catalogue-runtime_v0.2.json`](../../data/golden/sprach-a-lyzer_context-proposition-catalogue-runtime_v0.2.json)
- [`RESOLVER-CATALOGUE-CONTRACTS-v0.1.md`](RESOLVER-CATALOGUE-CONTRACTS-v0.1.md)

## Nächster sinnvoller Schritt

v0.2B-C erweitert die Golden-Abdeckung für `CAUSE`, `CONSEQUENCE`,
`CONDITION`, `ADDITION` und `CORRECTION` sowie Actor- und Scope-Schutzfälle.
Danach folgen proposition-lokale Target-/Expectation-Fakten und deren
Trace-Verknüpfung, bevor der Meilenstein v0.2 geschlossen werden kann.
