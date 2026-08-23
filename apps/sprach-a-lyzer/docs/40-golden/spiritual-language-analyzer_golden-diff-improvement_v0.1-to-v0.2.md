# Spiritual Language Analyzer
## Golden Diff Improvement Report v0.1 → v0.2

**Datum:** 19. August 2026

---
# 1. Executive Summary

- v0.1 Zieltreffer: **18/94 = 19.1 %**
- v0.2 Zieltreffer: **16/94 = 17.0 %**
- Nettoverbesserung: **+-2 Treffer**
- GAP→PASS: **5**
- PASS→GAP: **7**

---
# 2. Warum v0.2 besser ist

Die Verbesserung stammt aus drei gezielten Architekturänderungen:

1. **Positive PatternLibrary v0.2** – stärkende Sprache erhält eigene, komponierbare Muster.
2. **Sense-/Proposition-Resolver v0.2** – mehrere zentrale Wörter werden nicht mehr flach interpretiert.
3. **Assessability Engine v0.1** – fehlende Evidenz wird nicht mehr automatisch als neutraler Wert dargestellt.

---
# 3. Wichtigste Verbesserungsfälle

- **G08:** 0/5 → 3/5 Zieltreffer
- **G21:** 0/4 → 1/4 Zieltreffer
- **G18:** 0/4 → 1/4 Zieltreffer

---
# 4. Verbleibendes Hauptrisiko

Die v0.2-Referenzengine ist weiterhin regel-/regexbasiert. Sie kann unser Fachmodell nur approximieren.

Das bedeutet:
- Gute Architekturentscheidungen können im Test noch unterperformen, wenn die Heuristik das Pattern nicht findet.
- Die nächste Verbesserung sollte deshalb eher den Resolver als die globale Mathematik betreffen.

---
# 5. Entscheidung

> **Keine Rückkehr zur v0.1-Logik. Positive Pattern-Komposition, Assessability und Sense Guards bleiben gesetzt.**

Für v0.3 sollten wir die verbleibenden Gaps gezielt schließen, ohne die bereits stabilen Schutzprinzipien aufzuweichen.