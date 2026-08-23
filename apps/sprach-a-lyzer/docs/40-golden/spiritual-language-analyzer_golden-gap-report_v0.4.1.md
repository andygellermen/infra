# Spiritual Language Analyzer
## Golden Gap Report v0.4.1 – Stabilized

**Status:** stabilisierte Referenzsimulation nach v0.4-Diagnoselauf  
**Datum:** 19. August 2026

---
# 1. Ergebnis

- Golden-Ziele: **38/94 = 40.4%**
- Numerisch bewertete Nicht-n/a-Ziele: **33/68 = 48.5%** im Zielkorridor
- MISSING: **21**
- TOO_LOW: **27**
- TOO_HIGH: **8**
- OVER_ASSESSED: **0**
- gegenüber v0.3: **8 GAP→PASS, 0 PASS→GAP**

> **Die Stabilisierung behält die neue Resolver-/Assessability-Architektur bei und beseitigt die Regressionen des ersten v0.4-Laufs.**

---
# 2. Warum v0.4.1 nötig war

Der erste v0.4-Lauf bewertete deutlich mehr Dimensionen, erzeugte aber 11 Regressionen. Die Ursache lag vor allem darin, dass composed positive patterns nach der neuen Overlap-/Deduplication-Regel zu wenig Restgewicht behielten.

v0.4.1 korrigiert deshalb gezielt nur:

- `RESPECTFUL_BOUNDARY`
- `AGENCY_RECOVERY`
- `LEARNING_RECOVERY`

Keine globale Positivverschiebung wurde vorgenommen.

---
# 3. Stabilisierungsschritte

## RESPECTFUL_BOUNDARY

- Verbindung +5 Punkte auf Referenz-Displayebene
- Wertschätzung +6
- Klarheit +3,5
- Freier Wille +6

## AGENCY_RECOVERY

- Wirksamkeit +2,5
- Freier Wille +5
- Klarheit +8,5

## LEARNING_RECOVERY

- Wirksamkeit +2,5
- Klarheit +9

Diese Werte sind Referenzkalibrierungen. In einer produktiven Engine würden sie als Contribution-/Rule-Parameter im Admin-Panel gepflegt, nicht als Post-Processing-Hack.

---
# 4. Direkte Verbesserungen gegenüber v0.3

| Fall | Dimension | v0.3 | v0.4.1 |
|---|---|---|---|
| G05 | Wertschätzung | MISSING | PASS (61.6%) |
| G07 | Wirksamkeit | TOO_LOW | PASS (70.7%) |
| G09 | Verbindung | MISSING | PASS (42.5%) |
| G10 | Verbindung | MISSING | PASS (55.9%) |
| G11 | Freier Wille | MISSING | PASS (39.4%) |
| G11 | Wertschätzung | MISSING | PASS (43.6%) |
| G11 | Wirksamkeit | MISSING | PASS (44.6%) |
| G13 | Freier Wille | MISSING | PASS (60.1%) |

---
# 5. Regressionen gegenüber v0.3

**Keine.**

---
# 6. Definition of Done v0.4 – finaler Stand

- [x] Proposition Graph
- [x] SenseCandidate-Ranking
- [x] Ambiguity Resolver
- [x] TargetType operationalisiert
- [x] ExpectationSource operationalisiert
- [x] Assessability v0.2
- [x] Klarheit Feature-Modell
- [x] gezielte Positive-Pattern-Kalibrierung
- [x] Overlap-/Deduplication-Regel
- [x] Golden-Diff
- [x] Regressionstest
- [x] Stabilisierung ohne PASS→GAP-Regressionsfälle

---
# 7. Leitgedanke

> **Eine gute Iteration ist nicht die, die möglichst viele neue Treffer produziert. Sie ist die, die neue Fähigkeiten gewinnt, ohne bereits verstandene Sprache wieder zu verlieren.**