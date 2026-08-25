# Sprach-A-Lyzer – Implementation Baseline v0.1

**Status:** APPROVED  
**Version:** 0.1  
**Stand:** 24. August 2026  
**Owner:** Product & Engineering  
**Supersedes:** –

## Zweck

Diese Baseline friert den verbindlichen Umfang des ersten ausführbaren
Vertical Slice ein. Ältere Dokumente und Daten bleiben unverändert und dienen
weiterhin als versionierte fachliche Herkunft.

## Verbindliche Quellen

| Quelle | Status in dieser Baseline | Verwendung |
|---|---|---|
| `START-HERE.md` | APPROVED | Schutzregeln und sechs Acceptance Cases |
| `CODY-HANDOFF.md` | APPROVED | MVP Scope, Architekturprinzipien, kanonische IDs |
| `DEVELOPER-HANDOFF-v0.1.md` | REFERENCE | Pipeline und erwartete Fallsemantik |
| `spiritual-language-analyzer_reference-engine_v0.4.md` | REFERENCE | Resolver- und Assessability-Modell |
| `spiritual-language-analyzer_reference-simulation_v0.4.1.json` | REFERENCE | numerische Referenz, nicht blind zu kopierende Produktwahrheit |

Bei Widersprüchen gilt die Reihenfolge dieser Tabelle von oben nach unten.

## Kanonische Dimensionen

```text
AGENCY
CONNECTION
APPRECIATION
CLARITY
VOLITION
OPENNESS
```

`FREE_WILL` ist ausschließlich ein Legacy-Alias. Neue API-Ergebnisse,
Regeln und persistierte Daten verwenden `VOLITION`. Importgrenzen dürfen
`FREE_WILL` annehmen und müssen es deterministisch auf `VOLITION` abbilden.
Versionierte Quelldateien werden dafür nicht nachträglich verändert.

## Umfang des Vertical Slice

Der Slice implementiert:

- Textnormalisierung
- kontextabhängige Sense-Auflösung für die sechs Acceptance Cases
- Pattern-Erkennung
- sechs Dimensionen mit expliziter Assessability
- `null` bei fehlender Evidenz
- Contribution Trace
- Reflexionsfrage und Alternativen, sofern fachlich belegt
- Resonanzhinweis ohne semantische Scorewirkung
- maschinenlesbare Golden Tests

Nicht Bestandteil dieses Slice sind PostgreSQL-Persistenz, HTTP-API, UI,
Managed Import, Q/A Sessions, Audio und generative KI. Diese Fähigkeiten
folgen nach dem nachweislich laufenden Core.

## Abnahmekriterium

Alle sechs Fälle aus `START-HERE.md` laufen reproduzierbar über dieselbe
öffentliche Analysefunktion. Jeder fachliche Score ist durch Contributions
erklärbar; nicht belegte Dimensionen besitzen keinen Score.
