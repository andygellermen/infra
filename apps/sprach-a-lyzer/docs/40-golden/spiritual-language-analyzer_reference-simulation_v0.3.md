# Spiritual Language Analyzer
## Reference Simulation v0.3

**Status:** ausführbare Referenzengine mit P0-Gap-Fixes  
**Version:** 0.3  
**Datum:** 19. August 2026

---
# 1. Neu in v0.3

- tiefere Proposition-Komposition
- stärkere positive Pattern-Contributions
- strukturierte Klarheitsbewertung
- feinere Assessability statt harter v0.2-Gates
- Ambiguity Resolver für `umfahren` / `eigentlich` / Homophonie-Hinweise
- Sense Guards für `frei`, `sollen`, `Problem`, `Fehler`

---
# 2. Ergebnisübersicht

| ID | WIR | VER | WER | KLA | FW | OFF | Wing |
|---|---:|---:|---:|---:|---:|---:|---:|
| G01 | — | — | — | — | 34.1 | 42.4 | — |
| G02 | — | — | — | — | — | — | — |
| G03 | — | — | — | 74.8 | — | — | — |
| G04 | 36.1 | — | 30.2 | 66.8 | — | — | 45.2 |
| G05 | — | — | — | 68.6 | — | — | — |
| G06 | — | 32.1 | 30.2 | — | — | 35.1 | 32.6 |
| G07 | 69.8 | — | — | 79.1 | 75.0 | — | 74.6 |
| G08 | 64.2 | 72.9 | 72.9 | 85.3 | 73.8 | 56.7 | 71.0 |
| G09 | — | — | — | — | — | 34.1 | — |
| G10 | — | — | — | — | — | — | — |
| G11 | — | — | — | — | — | — | — |
| G12 | — | — | — | — | — | — | — |
| G13 | — | — | — | 65.7 | — | 61.9 | — |
| G14 | — | — | — | 64.7 | 62.9 | — | — |
| G15 | — | — | — | 68.2 | 37.1 | — | — |
| G16 | 76.4 | 52.3 | 53.4 | 70.6 | 75.5 | 66.4 | 66.1 |
| G17 | — | — | — | — | — | — | — |
| G18 | 75.4 | 64.1 | 61.0 | 72.3 | 55.5 | 82.6 | 68.5 |
| G19 | — | — | — | 67.2 | — | — | — |
| G20 | — | 32.1 | 30.2 | 68.2 | — | — | 44.4 |
| G21 | 75.4 | 64.1 | 61.0 | 72.9 | 55.5 | 82.6 | 68.6 |
| G22 | — | — | — | — | — | — | — |
| G23 | — | — | — | — | — | — | — |
| G24 | — | — | — | — | — | — | — |
| G25 | — | — | — | — | — | — | — |
| G26 | — | — | — | — | — | — | — |
| G27 | 62.9 | 65.9 | 56.6 | 65.8 | 56.6 | 71.6 | 63.2 |
| G28 | — | — | — | — | — | — | — |
| G29 | 66.3 | — | 55.6 | 69.8 | 57.8 | 73.7 | 64.7 |
| G30 | — | — | — | — | — | — | — |

---
# 3. Architekturentscheidung

> **v0.3 versucht nicht, fehlende Sprache durch höhere Gewichte zu kompensieren. Sie erweitert zuerst Resolver und Pattern-Komposition und kalibriert erst dann Contributions.**

# 4. Noch bewusst offen

- echter Dependency Parser
- LLM-basierte Propositionserkennung
- Ironie
- Prosodie
- komplexer Negationsscope
- langfristige persönliche Resonanz

# 5. Leitgedanke

> **Eine reifere Engine erkennt mehr, aber sie weiß zugleich genauer, wann sie nichts behaupten sollte.**