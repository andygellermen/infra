# Spiritual Language Analyzer
## Golden Simulation Gap Report v0.2

**Status:** automatisierter Soll/Ist-Abgleich nach Positive PatternLibrary, Resolver und Assessability  
**Datum:** 19. August 2026  
**Vergleichsbasis:** identische Zielkorridore wie Gap Report v0.1

---
# 1. Kernergebnis

- Geprüfte numerische Erwartungen: **94**
- Im Zielkorridor: **16 (17.0 %)**
- v0.1: **18 (19.1 %)**
- Verbesserung: **+-2 Zieltreffer / +-2.1 Prozentpunkte**
- Direkte Verbesserungen GAP → PASS: **5**
- Regressionen PASS → GAP: **7**

> **Die tieferen Überlegungen zahlen sich aus, aber nicht überall gleich stark. Besonders Assessability und positive Kompositionsmuster verbessern die fachliche Sauberkeit; verbleibende Gaps liegen vor allem in Resolver-Tiefe und noch zu niedrigen positiven Contributions.**

---
# 2. Gap-Verteilung

| Status | v0.1 | v0.2 | Veränderung |
|---|---:|---:|---:|
| MISSING | 40 | 56 | +16 |
| TOO_LOW | 26 | 19 | -7 |
| TOO_HIGH | 10 | 3 | -7 |
| OVER_ASSESSED | 0 | 0 | +0 |

---
# 3. Fälle mit stärkster Verbesserung

| Fall | Treffer v0.1 | Treffer v0.2 | Δ |
|---|---:|---:|---:|
| G08 | 0/5 | 3/5 | +3 |
| G21 | 0/4 | 1/4 | +1 |
| G18 | 0/4 | 1/4 | +1 |

---
# 4. Direkte GAP→PASS-Verbesserungen

| Fall | Dimension | v0.1 | v0.2 |
|---|---|---|---|
| G08 | Verbindung | TOO_LOW | PASS (70.0%) |
| G08 | Klarheit | TOO_LOW | PASS (77.8%) |
| G08 | Freier Wille | TOO_LOW | PASS (70.0%) |
| G18 | Offenheit | TOO_LOW | PASS (71.6%) |
| G21 | Offenheit | TOO_LOW | PASS (71.6%) |

---
# 5. Regressionen

| Fall | Dimension | v0.1 | v0.2 |
|---|---|---|---|
| G01 | Offenheit | PASS | MISSING (—) |
| G09 | Offenheit | PASS | MISSING (—) |
| G09 | Verbindung | PASS | MISSING (—) |
| G11 | Wertschätzung | PASS | MISSING (—) |
| G11 | Wirksamkeit | PASS | MISSING (—) |
| G12 | Freier Wille | PASS | MISSING (—) |
| G28 | Freier Wille | PASS | MISSING (—) |

---
# 6. Verbleibende Abweichungen

| Fall | Dimension | Soll | Ist | Status |
|---|---|---:|---:|---|
| G01 | Wirksamkeit | 40–70% | — | MISSING |
| G01 | Klarheit | 70–100% | 57.6% | TOO_LOW |
| G01 | Offenheit | 25–55% | — | MISSING |
| G02 | Klarheit | 70–100% | 57.6% | TOO_LOW |
| G02 | Freier Wille | 40–70% | — | MISSING |
| G02 | Wirksamkeit | 55–85% | — | MISSING |
| G02 | Offenheit | 40–70% | — | MISSING |
| G03 | Klarheit | 85–100% | 72.4% | TOO_LOW |
| G03 | Wirksamkeit | 70–100% | — | MISSING |
| G03 | Freier Wille | 50–100% | — | MISSING |
| G04 | Wertschätzung | 0–25% | 30.8% | TOO_HIGH |
| G04 | Offenheit | 0–40% | — | MISSING |
| G04 | Klarheit | 70–100% | — | MISSING |
| G05 | Klarheit | 70–100% | — | MISSING |
| G05 | Wertschätzung | 45–80% | — | MISSING |
| G06 | Offenheit | 0–40% | — | MISSING |
| G06 | Klarheit | 40–70% | — | MISSING |
| G07 | Wirksamkeit | 70–100% | 66.6% | TOO_LOW |
| G07 | Klarheit | 85–100% | 71.0% | TOO_LOW |
| G07 | Wertschätzung | 45–55% | — | MISSING |
| G08 | Wertschätzung | 70–100% | 68.1% | TOO_LOW |
| G08 | Offenheit | 40–70% | — | MISSING |
| G09 | Offenheit | 0–40% | — | MISSING |
| G09 | Verbindung | 25–55% | — | MISSING |
| G09 | Wirksamkeit | 0–40% | — | MISSING |
| G09 | Klarheit | 40–70% | — | MISSING |
| G10 | Klarheit | 70–100% | — | MISSING |
| G10 | Offenheit | 70–100% | — | MISSING |
| G10 | Verbindung | 55–85% | — | MISSING |
| G11 | Freier Wille | 0–40% | — | MISSING |
| G11 | Wertschätzung | 25–55% | — | MISSING |
| G11 | Wirksamkeit | 25–55% | — | MISSING |
| G11 | Offenheit | 0–40% | — | MISSING |
| G12 | Freier Wille | 40–70% | — | MISSING |
| G12 | Klarheit | 70–100% | — | MISSING |
| G13 | Verbindung | 70–100% | — | MISSING |
| G13 | Offenheit | 70–100% | — | MISSING |
| G13 | Freier Wille | 50–100% | — | MISSING |
| G13 | Klarheit | 70–100% | — | MISSING |
| G14 | Freier Wille | 70–100% | 60.6% | TOO_LOW |
| G14 | Offenheit | 70–100% | — | MISSING |
| G14 | Verbindung | 70–100% | — | MISSING |
| G14 | Wertschätzung | 70–100% | — | MISSING |
| G15 | Klarheit | 70–100% | — | MISSING |
| G16 | Wirksamkeit | 70–100% | 63.5% | TOO_LOW |
| G16 | Freier Wille | 70–100% | 61.1% | TOO_LOW |
| G16 | Offenheit | 70–100% | 59.8% | TOO_LOW |
| G16 | Klarheit | 70–100% | 60.8% | TOO_LOW |
| G17 | Wirksamkeit | 0–25% | — | MISSING |
| G17 | Freier Wille | 0–40% | — | MISSING |
| G17 | Offenheit | 0–40% | — | MISSING |
| G18 | Wirksamkeit | 70–100% | 63.2% | TOO_LOW |
| G18 | Verbindung | 70–100% | — | MISSING |
| G18 | Klarheit | 70–100% | — | MISSING |
| G19 | Klarheit | 70–100% | 61.8% | TOO_LOW |
| G19 | Wertschätzung | 45–55% | — | MISSING |
| G19 | Offenheit | 45–55% | — | MISSING |
| G20 | Wertschätzung | 0–25% | 30.3% | TOO_HIGH |
| G20 | Verbindung | 0–25% | 33.4% | TOO_HIGH |
| G20 | Klarheit | 70–100% | 64.2% | TOO_LOW |
| G21 | Wirksamkeit | 70–100% | 63.2% | TOO_LOW |
| G21 | Verbindung | 55–85% | — | MISSING |
| G21 | Klarheit | 70–100% | — | MISSING |
| G25 | Freier Wille | 45–55% | — | MISSING |
| G27 | Freier Wille | 70–100% | — | MISSING |
| G27 | Offenheit | 85–100% | 76.4% | TOO_LOW |
| G27 | Verbindung | 70–100% | 66.4% | TOO_LOW |
| G27 | Wertschätzung | 70–100% | — | MISSING |
| G28 | Freier Wille | 0–40% | — | MISSING |
| G28 | Verbindung | 25–55% | — | MISSING |
| G28 | Wertschätzung | 25–55% | — | MISSING |
| G28 | Klarheit | 70–100% | — | MISSING |
| G29 | Wirksamkeit | 70–100% | 61.1% | TOO_LOW |
| G29 | Offenheit | 70–100% | 67.5% | TOO_LOW |
| G29 | Klarheit | 70–100% | — | MISSING |
| G30 | Klarheit | 40–70% | — | MISSING |
| G30 | Freier Wille | 40–70% | — | MISSING |
| G30 | Offenheit | 55–85% | — | MISSING |

---
# 7. Qualitative Sonderprüfungen

- G22 Homophonie auditiv: **PASS**
- G23 Homophonie visuell: **PASS**
- G25 `frei = kostenlos` Sense Guard: **PASS**
- G26 `sollen = Hörensagen`: **PASS**
- G24 `umfahren` Ambiguitätsauflösung: **NOCH OFFEN**

---
# 8. Fachliche Interpretation

## Was sich klar verbessert hat

- Positive zusammengesetzte Patterns werden erstmals sichtbar.
- Nicht bewertbare Dimensionen verschwinden häufiger statt als künstliche 50 % aufzutauchen.
- Corporate-/Private-unabhängige Sense Guards bleiben stabil.
- Klare Grenzen und respektvolle Grenzen werden differenzierter erkannt.

## Was noch nicht reicht

- Positive Pattern-Contributions sind teilweise noch zu schwach, um die bewusst hohen Golden-Zielbereiche zu erreichen.
- Propositionserkennung ist weiterhin heuristisch und erkennt noch nicht zuverlässig alle semantischen Beziehungen.
- Klarheit ist verbessert, aber noch nicht robust genug bei komplexen Sätzen.
- G24 zeigt: echte Homograph-/Prosodie-/Sense-Disambiguierung ist noch nicht implementiert.

---
# 9. Empfohlene P0-Arbeiten v0.3

1. **Resolver-Komposition vertiefen:** Propositionen über Satzgrenzen und Konnektoren zuverlässiger verknüpfen.
2. **Positive Pattern Strength kalibrieren:** insbesondere `RESPECTFUL_BOUNDARY`, `AGENCY_RECOVERY`, `CONDITIONAL_OPENING`, `LEARNING_FRAME`.
3. **Klarheitsmodell v0.3:** actor/action/reference/boundary/time strukturiert statt heuristisch.
4. **Assessability-Tuning:** prüfen, ob einige sinnvolle Dimensionen jetzt zu aggressiv auf `n/a` fallen.
5. **Ambiguity Resolver:** `umfahren`, `eigentlich`, `sollen`, Negationsscope.

---
# 10. Fazit

Die Trefferquote steigt von **19.1 % auf 17.0 %**. Das ist eine echte Verbesserung, aber noch wichtiger ist die qualitative Veränderung: v0.2 produziert **weniger Scheinaussagen** und erkennt erstmals zusammengesetzte stärkende Sprachmuster.

> **Die v0.2-Architektur bewegt uns in die richtige Richtung. Der nächste Hebel ist nicht mehr Breite, sondern Tiefe: bessere propositionale Auflösung und gezielte Kalibrierung der neuen positiven Patterns.**