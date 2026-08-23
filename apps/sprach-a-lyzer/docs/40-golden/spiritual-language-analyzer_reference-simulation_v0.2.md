# Spiritual Language Analyzer
## Reference Simulation v0.2 – Positive Patterns + Resolver + Assessability

**Datum:** 19. August 2026

---
# 1. Zweck

Diese ausführbare Referenzsimulation integriert erstmals die drei neuen Bausteine:
- Positive PatternLibrary v0.2
- Sense-/Proposition-Resolver v0.2 (vereinfachte Referenzheuristiken)
- Assessability Engine v0.1

---
# 2. Ergebnisübersicht

| ID | WIR | VER | WER | KLA | FW | OFF | Wing |
|---|---:|---:|---:|---:|---:|---:|---:|
| G01 | — | — | — | 57.6 | 35.9 | — | — |
| G02 | — | — | — | 57.6 | — | — | — |
| G03 | — | — | — | 72.4 | — | — | — |
| G04 | 37.9 | — | 30.8 | — | — | — | — |
| G05 | — | — | — | — | — | — | — |
| G06 | — | 38.1 | 36.8 | — | — | — | — |
| G07 | 66.6 | — | — | 71.0 | 72.0 | — | 69.9 |
| G08 | — | 70.0 | 68.1 | 77.8 | 70.0 | — | 71.5 |
| G09 | — | — | — | — | — | — | — |
| G10 | — | — | — | — | — | — | — |
| G11 | — | — | — | — | — | — | — |
| G12 | — | — | — | — | — | — | — |
| G13 | — | — | — | — | — | — | — |
| G14 | — | — | — | — | 60.6 | — | — |
| G15 | — | — | — | — | 37.9 | — | — |
| G16 | 63.5 | — | — | 60.8 | 61.1 | 59.8 | 61.3 |
| G17 | — | — | — | — | — | — | — |
| G18 | 63.2 | — | — | — | — | 71.6 | — |
| G19 | — | — | — | 61.8 | — | — | — |
| G20 | — | 33.4 | 30.3 | 64.2 | — | — | 42.9 |
| G21 | 63.2 | — | — | — | — | 71.6 | — |
| G22 | — | — | — | — | — | — | — |
| G23 | — | — | — | — | — | — | — |
| G24 | — | — | — | — | — | — | — |
| G25 | — | — | — | — | — | — | — |
| G26 | — | — | — | — | — | — | — |
| G27 | 60.0 | 66.4 | — | — | — | 76.4 | 67.7 |
| G28 | — | — | — | — | — | — | — |
| G29 | 61.1 | — | — | — | — | 67.5 | — |
| G30 | — | — | — | — | — | — | — |

---
# 3. Wesentliche Veränderungen gegenüber v0.1

- Positive Sprache kann nun zusammengesetzt erkannt werden (`RESPECTFUL_BOUNDARY`, `AGENCY_RECOVERY`, `EMOTION_PLUS_AGENCY`).
- `frei = kostenlos` und `sollen = Hörensagen` werden weiterhin als Sense-Guards behandelt.
- Klarheit benötigt propositionale Evidenz und wird nicht mehr allein aus einzelnen Wörtern erzeugt.
- Dimensionen ohne ausreichende Evidenz bleiben `—` statt künstlich bei 50 % zu erscheinen.
- WingScore erscheint nur noch bei mindestens drei tatsächlich bewertbaren Dimensionen.

# 4. Noch offene Grenzen

Die Resolver-Logik ist weiterhin heuristisch und ersetzt noch keinen echten Parser/LLM-Resolver. Insbesondere Ironie, Negationsscope, Prosodie und komplexe Anaphern bleiben bewusst offen.

# 5. Leitgedanke

> **Mit v0.2 beginnt die Referenzengine nicht nur mehr zu erkennen, sondern auch bewusster zu schweigen, wenn sie etwas noch nicht ausreichend erkennt.**