# Sprach-A-Lyzer – Question / Answer MVP Closure v0.3.0

- **Status:** RELEASED
- **Version:** 0.3.0
- **Stand:** 1. September 2026
- **Git-Tag:** `v0.3.0`
- **Owner:** Product & Engineering

## Ergebnis

Der Roadmap-Meilenstein **v0.3 – Question / Answer MVP** ist geschlossen.
Ein strikt validierter Runtime-Katalog stellt acht kanonische Fragen mit je
einem privaten und Corporate-Rendering bereit. Antworten werden zusammen mit
dem unveränderten Ergebnis des deterministischen Sprach-Cores ausgewertet;
Fragen und Q/A-Evidenz erzeugen selbst keinen Score.

Adaptive Auswahl bietet fünf bis acht Fragen anhand der dokumentierten
Gewichte für Informationsgewinn, Construct-Lücke, bisherige Relevanz, Phase
und Präferenz an. Redundanz, Leadingness und nicht freigegebenes Risiko wirken
als Abzüge. Die Engine bietet Kandidaten ausschließlich an und stellt keine
Frage ungefragt.

Progressive Sessions unterscheiden die Inferenzebenen:

- `C0`: sprachliche Beobachtung ohne Frageaussage,
- `C1`: fragekonditionierte Assoziation,
- `C2`: vorsichtig formulierte Elicitation-Assoziation mit Baseline,
- `C3`: Veränderung sprachlicher Evidenz innerhalb derselben Session.

`C4`-Kausalbehauptungen werden nicht erzeugt.

## Kanonischer MVP-Satz

Der ausführbare Katalog enthält `CQ007`, `CQ008`, `CQ009`, `CQ013`, `CQ021`,
`CQ023`, `CQ024` und `CQ034`. Intent, sichtbares Rendering,
Kompositionsregeln und adaptive Priorität bleiben getrennte Datenbereiche.

Der historische Question-Golden-Korpus v0.1 bleibt unverändert erhalten. In
21 seiner 54 Fälle weicht `expected_primary_construct` vom kanonischen
Core-Set ab. Ein Regressionstest dokumentiert diese bekannte Inkonsistenz
explizit; sie wird nicht stillschweigend in den neuen Runtime-Katalog
übernommen. Die kohärente Laufzeitbaseline ist
`data/golden/sprach-a-lyzer_question-answer-runtime_v0.2.json` mit 25 Fällen.

## Release-Vektor

| Artefakt | Version |
|---|---:|
| Core Release / HTTP API | 0.3.0 / 3 |
| Canonical Question / Rendering | 0.1 / 0.1 |
| Question Catalogue | 0.1 |
| Q/A Observation / Selection / Session | 0.1 / 0.1 / 0.1 |
| Q/A Runtime Golden | 0.2 |
| Analysis Result / Trace | 0.1 / 0.2 |
| Resolver Result / Catalogue | 0.2 / 0.1 |
| Construct Ontology | 0.2 |
| Policy Registry / PostgreSQL Schema | 0.7 / 4 |

Die vollständige maschinenlesbare Entsprechung liegt in
`data/seed/sprach-a-lyzer_release-manifest_v0.3.0.json`.

## Öffentliche Zugänge

```text
POST /api/v3/questions/select  → Question Selection v0.1
POST /api/v3/answers/analyze   → Q/A Observation v0.1
POST /api/v3/sessions/compose  → Question Session v0.1
```

Alle Requests sind strikt: unbekannte Felder, unbekannte Question IDs,
leere Antworten und ungültige Sessiongrößen werden abgewiesen. Rohtexte
werden durch diese Adapter nicht persistiert. Der lokale Befehl `go run
./cmd/qa` bietet dieselbe eingebettete Runtime ohne Datenbank an.

## Closure-Gate

```bash
bash ./scripts/verify-v0.3-closure.sh
```

Das Gate führt zuerst sämtliche v0.2-Closure-Nachweise aus. Zusätzlich prüft
es Schema-/Go-Parität, 25 Q/A-Golden-Fälle, unveränderte Core-Analyse,
adaptive Auswahl, C0–C3 ohne C4, Katalogaktivierung/-deaktivierung,
Fail-Closed-Verhalten, HTTP v3 und die historischen Release-Manifeste.

## Bewusste Grenzen

- Keine neue fachliche Score-Kalibrierung und kein Frage-Score-Bias.
- Q/A-Dimensionsevidenz ist erklärend und immer `scoring: false`.
- Keine Diagnose-, Trait-, Ranking- oder Kausalbehauptungen.
- Keine langfristigen Personenprofile oder Speicherung von Antworttexten.
- Der MVP enthält Standard-Renderings; Easy und Deep Reflective bleiben v0.4.

## Nächster Roadmap-Schritt

Als nächstes folgt **v0.4 – Rendering & Corporate/Private Profiles**. Der
kanonische Question Core und seine Q/A-Semantik bleiben dabei unverändert;
ausgebaut werden die isolierten Präsentationsvarianten und ihre Leakage-Gates.
