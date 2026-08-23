# Spiritual Language Analyzer
## Remaining Gap Cause Clustering & v0.4 Action List

**Status:** Priorisierte Ursachenanalyse der verbleibenden Golden-Gaps  
**Version:** 0.1  
**Datum:** 19. August 2026

---
# 1. Ausgangslage

Verbleibende Gaps in Golden Gap Report v0.3: **64**.

Ziel dieses Dokuments ist nicht, die Zahlen durch globale Gewichtungsänderungen passend zu machen. Stattdessen werden die Abweichungen nach ihrer **wahrscheinlichen fachlichen Ursache** gruppiert.

> **Resolver vor Kalibrierung, Kalibrierung vor kosmetischer Trefferoptimierung.**

---
# 2. Clusterübersicht

| Priorität | Ursachencluster | Gaps |
|---:|---|---:|
| 1 | R3 Sense-/Pragmatik-Resolver zu flach | 14 |
| 2 | C1 Positive Contribution zu schwach | 13 |
| 3 | A1 Assessability zu streng / Evidenz zu dünn | 13 |
| 4 | R2 Proposition-/Discourse-Komposition fehlt | 6 |
| 5 | R4 Kontext-/Target-/Expectation-Auflösung fehlt | 6 |
| 6 | C2 Klarheitsbeitrag zu schwach | 6 |
| 7 | R1 Ambiguitäts-/Sense-Resolver fehlt | 3 |
| 8 | C5 Belastender Contribution zu stark | 3 |

---
# 3. Detailcluster

## R3 Sense-/Pragmatik-Resolver zu flach

**Anzahl:** 14

Betroffene Gaps:
- G02 · Klarheit · Soll 70–100% · Ist — · MISSING
- G02 · Freier Wille · Soll 40–70% · Ist — · MISSING
- G02 · Wirksamkeit · Soll 55–85% · Ist — · MISSING
- G02 · Offenheit · Soll 40–70% · Ist — · MISSING
- G12 · Freier Wille · Soll 40–70% · Ist — · MISSING
- G12 · Klarheit · Soll 70–100% · Ist — · MISSING
- G13 · Verbindung · Soll 70–100% · Ist — · MISSING
- G13 · Freier Wille · Soll 50–100% · Ist — · MISSING
- G14 · Offenheit · Soll 70–100% · Ist — · MISSING
- G14 · Verbindung · Soll 70–100% · Ist — · MISSING
- G14 · Wertschätzung · Soll 70–100% · Ist — · MISSING
- G19 · Wertschätzung · Soll 45–55% · Ist — · MISSING
- G19 · Offenheit · Soll 45–55% · Ist — · MISSING
- G25 · Freier Wille · Soll 45–55% · Ist — · MISSING

Empfohlene Maßnahmen:
- SenseCandidate-Scoring mit phrase_fit, syntax_fit, discourse_fit umsetzen
- sollen/dürfen/frei/eigentlich/Problem/Fehler mit klaren Resolver-Regeln hinterlegen
- Top-Sense-Gap und Ambiguity-Cap operationalisieren

## C1 Positive Contribution zu schwach

**Anzahl:** 13

Betroffene Gaps:
- G07 · Wirksamkeit · Soll 70–100% · Ist 69.8% · TOO_LOW
- G07 · Klarheit · Soll 85–100% · Ist 79.1% · TOO_LOW
- G13 · Offenheit · Soll 70–100% · Ist 61.9% · TOO_LOW
- G13 · Klarheit · Soll 70–100% · Ist 65.7% · TOO_LOW
- G14 · Freier Wille · Soll 70–100% · Ist 62.9% · TOO_LOW
- G16 · Offenheit · Soll 70–100% · Ist 66.4% · TOO_LOW
- G18 · Verbindung · Soll 70–100% · Ist 64.1% · TOO_LOW
- G27 · Freier Wille · Soll 70–100% · Ist 56.6% · TOO_LOW
- G27 · Offenheit · Soll 85–100% · Ist 71.6% · TOO_LOW
- G27 · Verbindung · Soll 70–100% · Ist 65.9% · TOO_LOW
- G27 · Wertschätzung · Soll 70–100% · Ist 56.6% · TOO_LOW
- G29 · Wirksamkeit · Soll 70–100% · Ist 66.3% · TOO_LOW
- G29 · Klarheit · Soll 70–100% · Ist 69.8% · TOO_LOW

Empfohlene Maßnahmen:
- nur bereits sicher erkannte positive Patterns um 10–25 % nachkalibrieren
- Composed Patterns stärker als Einzelmarker gewichten
- Positive Contribution Caps prüfen

## A1 Assessability zu streng / Evidenz zu dünn

**Anzahl:** 13

Betroffene Gaps:
- G01 · Wirksamkeit · Soll 40–70% · Ist — · MISSING
- G01 · Klarheit · Soll 70–100% · Ist — · MISSING
- G04 · Offenheit · Soll 0–40% · Ist — · MISSING
- G05 · Wertschätzung · Soll 45–80% · Ist — · MISSING
- G06 · Klarheit · Soll 40–70% · Ist — · MISSING
- G07 · Wertschätzung · Soll 45–55% · Ist — · MISSING
- G11 · Freier Wille · Soll 0–40% · Ist — · MISSING
- G11 · Wertschätzung · Soll 25–55% · Ist — · MISSING
- G11 · Wirksamkeit · Soll 25–55% · Ist — · MISSING
- G11 · Offenheit · Soll 0–40% · Ist — · MISSING
- G17 · Wirksamkeit · Soll 0–25% · Ist — · MISSING
- G17 · Freier Wille · Soll 0–40% · Ist — · MISSING
- G17 · Offenheit · Soll 0–40% · Ist — · MISSING

Empfohlene Maßnahmen:
- WEAK-Zustand praktisch nutzen statt sofort null
- eine starke komponierte Regel als ausreichende Evidenz zulassen
- Dimension-spezifische Mindestkriterien statt globalem Gate

## R2 Proposition-/Discourse-Komposition fehlt

**Anzahl:** 6

Betroffene Gaps:
- G09 · Verbindung · Soll 25–55% · Ist — · MISSING
- G09 · Wirksamkeit · Soll 0–40% · Ist — · MISSING
- G09 · Klarheit · Soll 40–70% · Ist — · MISSING
- G10 · Klarheit · Soll 70–100% · Ist — · MISSING
- G10 · Offenheit · Soll 70–100% · Ist — · MISSING
- G10 · Verbindung · Soll 55–85% · Ist — · MISSING

Empfohlene Maßnahmen:
- Cross-sentence/Clause Proposition Graph einführen
- Discourse Relations für CONTRAST, CONCESSION, CONDITION, ADDITION robuster ableiten
- Kompositionsregeln RESPECTFUL_BOUNDARY, AGENCY_RECOVERY, LEARNING_RECOVERY mit Proposition-IDs statt Regex auslösen

## R4 Kontext-/Target-/Expectation-Auflösung fehlt

**Anzahl:** 6

Betroffene Gaps:
- G03 · Wirksamkeit · Soll 70–100% · Ist — · MISSING
- G03 · Freier Wille · Soll 50–100% · Ist — · MISSING
- G28 · Freier Wille · Soll 0–40% · Ist — · MISSING
- G28 · Verbindung · Soll 25–55% · Ist — · MISSING
- G28 · Wertschätzung · Soll 25–55% · Ist — · MISSING
- G28 · Klarheit · Soll 70–100% · Ist — · MISSING

Empfohlene Maßnahmen:
- TargetType-Resolver für PERSON/BEHAVIOR/PROCESS/OBJECT
- ExpectationSource für SELF/LAW/INTERNALIZED/OTHER_PERSON
- Kontext-Caps für Safety, Legal, Family, Coaching operationalisieren

## C2 Klarheitsbeitrag zu schwach

**Anzahl:** 6

Betroffene Gaps:
- G03 · Klarheit · Soll 85–100% · Ist 74.8% · TOO_LOW
- G04 · Klarheit · Soll 70–100% · Ist 66.8% · TOO_LOW
- G05 · Klarheit · Soll 70–100% · Ist 68.6% · TOO_LOW
- G15 · Klarheit · Soll 70–100% · Ist 68.2% · TOO_LOW
- G19 · Klarheit · Soll 70–100% · Ist 67.2% · TOO_LOW
- G20 · Klarheit · Soll 70–100% · Ist 68.2% · TOO_LOW

Empfohlene Maßnahmen:
- Structured Clarity als Feature-Score statt einzelner Contribution modellieren
- actor/action/reference/boundary/time einzeln bewerten

## R1 Ambiguitäts-/Sense-Resolver fehlt

**Anzahl:** 3

Betroffene Gaps:
- G30 · Klarheit · Soll 40–70% · Ist — · MISSING
- G30 · Freier Wille · Soll 40–70% · Ist — · MISSING
- G30 · Offenheit · Soll 55–85% · Ist — · MISSING

Empfohlene Maßnahmen:
- AmbiguityProfile ausführbar machen
- Homographen mit Pronunciation/Stress/Sense-Kandidaten koppeln
- Bei text-only Ambiguität Confidence begrenzen und Rewrite unterdrücken

## C5 Belastender Contribution zu stark

**Anzahl:** 3

Betroffene Gaps:
- G04 · Wertschätzung · Soll 0–25% · Ist 30.2% · TOO_HIGH
- G20 · Wertschätzung · Soll 0–25% · Ist 30.2% · TOO_HIGH
- G20 · Verbindung · Soll 0–25% · Ist 32.1% · TOO_HIGH

Empfohlene Maßnahmen:
- negative Einzelhit-Caps prüfen
- Target-/Kontextmodifier vor Aggregation anwenden

---
# 4. Priorisierte v0.4 Resolver-Liste

## P0.1 – Proposition Graph

- Propositionen mit IDs und Source-Spans erzeugen
- Relationen zwischen Propositionen explizit speichern
- `aber`, `trotzdem`, `wenn`, `weil`, `deshalb` nicht als Wortmarker, sondern als Relation behandeln
- Kompositionspatterns auf Propositionen anwenden

## P0.2 – SenseCandidate Resolver

- mehrere Sense-Kandidaten mit Scores statt harter Regex-Auswahl
- phrase_fit, syntax_fit, domain_fit, discourse_fit einführen
- Ambiguität anhand Top1–Top2-Abstand bewerten

## P0.3 – TargetType / ExpectationSource

- Person vs. Verhalten vs. Prozess sauber unterscheiden
- `sollen` / `müssen` um Quelle der Erwartung/Pflicht ergänzen
- Safety/Legal/Coaching-Kontext als Modifier und nicht als pauschale Ersetzung

## P0.4 – Assessability 0.2

- WEAK als sichtbare Tendenz implementieren
- komponierte High-Confidence-Patterns dürfen Dimension assessable machen
- dimensionsspezifische Gates statt eines globalen Schwellenwerts

## P0.5 – Klarheit 0.3

- Klarheit aus Feature-Vektor bilden: actor, predicate, target, reference, time, boundary, decision
- keine starke Klarheit allein durch Labeling oder einzelne Marker

---
# 5. Priorisierte Kalibrierungsliste

Kalibrierung beginnt erst nach P0.1–P0.5.

## K1 – Positive Composed Patterns

Zuerst prüfen:
`RESPECTFUL_BOUNDARY`, `AGENCY_RECOVERY`, `CONDITIONAL_OPENING`, `LEARNING_RECOVERY`.

Vorgehen:
- nur bei hoher Pattern-Confidence erhöhen
- +10 %, +15 %, +20 %, +25 % als Teststufen
- Golden-Diff nach jeder Stufe
- keine globale positive Verstärkung

## K2 – Negative Einzelhit-Caps

Prüfen:
`SELF_DEVALUATION`, `PERSON_DEVALUATION`, `GENERALIZATION`, `INTERNAL_PRESSURE`.

Ziel:
- klare Belastungen sichtbar lassen
- einzelne Phrasen nicht komplette Dimension dominieren lassen

## K3 – Klarheits-Caps

- `PREDICATIVE_LABELING` nicht mit Klarheit gleichsetzen
- Feature-basierte Klarheit stärker als Keyword-/Pattern-Klarheit gewichten

## K4 – Assessability Thresholds

- `WEAK`: 0.50–0.64
- `ASSESSABLE`: 0.65–0.79
- `STRONG`: ≥0.80
- Werte im Test Lab sichtbar, im Standard-UI abhängig vom Profil

---
# 6. Was wir ausdrücklich nicht tun sollten

- keine globale Erhöhung aller positiven Contributions
- keine Absenkung der Assessability nur für bessere Trefferquote
- keine automatische Wertung von Emotionen als negativ
- keine Resonanzwerte nutzen, um fehlende semantische Evidenz zu ersetzen
- keine Homophonie semantisch vererben

---
# 7. Definition of Done für v0.4

- [ ] Proposition Graph implementiert
- [ ] SenseCandidate-Ranking implementiert
- [ ] Ambiguity Resolver erweitert
- [ ] TargetType operationalisiert
- [ ] ExpectationSource operationalisiert
- [ ] Assessability v0.2
- [ ] Klarheit Feature-Modell
- [ ] gezielte Positive-Pattern-Kalibrierung
- [ ] Overlap-/Deduplication-Regel für composed patterns
- [ ] Golden-Diff v0.4
- [ ] keine Regression der Schutzfälle

---
# 8. Leitgedanke

> **Jeder verbleibende Gap soll entweder durch besseres Sprachverständnis, bessere Unsicherheitsbehandlung oder gezielte fachliche Kalibrierung erklärbar geschlossen werden – niemals durch kosmetisches Tuning.**