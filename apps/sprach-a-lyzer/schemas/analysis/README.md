# Analysis schemas

- `sprach-a-lyzer_analysis-request_v0.1.json` beschreibt den Request für
  `POST /api/v1/analyze`.
- `sprach-a-lyzer_analysis-result_v0.1.json` beschreibt den öffentlichen
  Ergebnisvertrag des deterministischen Vertical Slice.
- `sprach-a-lyzer_analysis-trace_v0.1.json` beschreibt den eigenständigen,
  aus dem Ergebnis ableitbaren Explainability-Vertrag. Seine Contribution-
  Indizes sind nullbasiert.
- `sprach-a-lyzer_resolver-result_v0.2.json` beschreibt den additive aufgebauten
  Context-/Proposition-Resolver vor dem Scoring. Der veröffentlichte v0.1-
  Analysevertrag bleibt dadurch unverändert.
- Neue Verträge verwenden ausschließlich die kanonische Dimension
  `VOLITION`; `FREE_WILL` wird nur an Legacy-Importgrenzen akzeptiert.

Die zugehörigen Go-Typen werden über `internal/analysis` veröffentlicht. Der
Trace v0.1 enthält nur bereits deterministisch verfügbare Werte. Zusätzliche
Zwischenfaktoren des künftigen Assessability-Modells benötigen eine neue
Vertragsversion und werden nicht mit künstlichen Nullwerten vorweggenommen.
