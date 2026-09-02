# Question schemas

- `sprach-a-lyzer_question-canonical_v0.1.json` beschreibt den
  profilunabhängigen Intent einer Frage und enthält bewusst keinen sichtbaren
  Fragetext.
- `sprach-a-lyzer_question-rendering_v0.1.json` beschreibt eine sichtbare,
  profilgebundene Formulierung und darf den Construct Intent nicht verändern.
- `sprachkompass_qa-composition-contract_v0.1.json` bleibt der fachliche
  Q/A-Kompositionsvertrag.
- `sprach-a-lyzer_question-catalogue_v0.1.json` bindet exakt acht freigegebene
  Fragen, ihre Renderings, Kompositionsregeln, Prioritäten und Guardrails.
- `sprach-a-lyzer_question-answer-observation_v0.1.json` beschreibt die
  nicht-scorende Auswertung einer einzelnen Antwort.
- `sprach-a-lyzer_question-selection_v0.1.json` beschreibt fünf bis acht
  deterministisch angebotene Kandidaten.
- `sprach-a-lyzer_question-session_v0.1.json` beschreibt eine progressive
  Session mit C0–C3-Inferenz und schließt C4 aus.

Eine Frage erzeugt selbst keinen Dimensionsbeitrag. Ihr Score Bias bleibt null.

## Rendering v0.4

- `sprach-a-lyzer_question-rendering_v0.2.json` trennt Profil,
  Präsentationsmodus und Formulierungsvariante und markiert jedes Rendering
  als nicht-scoring.
- `sprach-a-lyzer_question-rendering-catalogue_v0.1.json` bindet die 40
  freigegebenen Varianten samt Qualitätsgrenzen und Guardrails.
- `sprach-a-lyzer_question-rendering-request_v0.1.json` beschreibt die
  Aktionen `DEFAULT`, `SIMPLIFY` und `REPHRASE`.
- `sprach-a-lyzer_question-rendering-result_v0.1.json` veröffentlicht
  Qualitätsdeltas und sichere Fallbacks.

Corporate-Renderings dürfen keine spirituelle Explizitheit oder
Deep-Reflective-Variante enthalten. Deep Reflective ist privat und erfordert
ein explizites Opt-in.
