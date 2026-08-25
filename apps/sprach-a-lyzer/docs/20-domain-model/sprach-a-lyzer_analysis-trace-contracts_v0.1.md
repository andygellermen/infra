# Sprach-A-Lyzer – Analyse- und Trace-Verträge v0.1

**Status:** APPROVED  
**Version:** 0.1  
**Stand:** 24. August 2026  
**Owner:** Product & Engineering

## Zweck

Die Verträge stabilisieren die Grenze zwischen HTTP-Adapter, deterministischem
Core, künftiger Persistenz und späteren Oberflächen. Go-Typen sind der
Kompilierungsvertrag; JSON-Schemas sind der sprachunabhängige Austauschvertrag.

| Vertrag | Go-Typ | JSON-Schema |
|---|---|---|
| Analyse-Request | `analysis.Request` | `sprach-a-lyzer_analysis-request_v0.1.json` |
| Analyse-Ergebnis | `analysis.Result` | `sprach-a-lyzer_analysis-result_v0.1.json` |
| Explainability-Trace | `analysis.Trace` | `sprach-a-lyzer_analysis-trace_v0.1.json` |

Die zugehörigen Untertypen – etwa Dimensionsergebnis, Proposition,
Sense-Auflösung, Contribution und Assessability-Trace – werden ebenfalls über
das Package `internal/analysis` veröffentlicht.

## Verbindliche Invarianten

- Jeder Ergebnis- und Trace-Vertrag enthält exakt die sechs kanonischen
  Dimensionen einschließlich `VOLITION`.
- `FREE_WILL` ist in neuen Analyse- und Trace-Daten unzulässig.
- `NOT_ASSESSABLE` besitzt im Analyse-Ergebnis immer `score: null`.
- Leere Sammlungen werden als JSON-Arrays beziehungsweise -Objekte und nicht
  als `null` ausgegeben.
- Contribution-Indizes sind nullbasiert und verweisen ausschließlich auf
  Contributions derselben Dimension.
- Resonanzhinweise bleiben von semantischen Contributions getrennt.

## Trace v0.1

Der eigenständige Trace wird deterministisch aus einem `analysis.Result`
abgeleitet. Er enthält:

- die Contribution-Einträge in stabiler Reihenfolge,
- den finalen Assessability-Zustand jeder Dimension,
- den finalen Assessability-Wert und
- die Indizes der zugehörigen Contributions.

Die Konzeptdokumente sehen künftig weitere Zwischenwerte wie Evidence Mass,
Pattern Diversity oder Ambiguity Modifier vor. Solange die Engine diese Werte
nicht berechnet, werden sie nicht mit scheinpräzisen Nullwerten in v0.1
aufgenommen.

## Versionierung

Optionale, rückwärtskompatible Ergänzungen dürfen innerhalb der Go-API erst
nach Schema-Prüfung erfolgen. Neue Pflichtfelder, geänderte Semantik,
Dimensionsänderungen oder zusätzliche Trace-Stufen benötigen eine neue
Vertragsversion. Bestehende v0.1-Schemas werden danach nicht still verändert.
