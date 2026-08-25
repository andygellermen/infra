# Sprach-A-Lyzer – Vertical-Slice Golden Gate v0.2

**Status:** APPROVED  
**Version:** 0.2  
**Stand:** 24. August 2026  
**Owner:** Product & Engineering

## Zweck

Die Suite `sprach-a-lyzer_vertical-slice_v0.2.json` friert die vollständige
Ausgabe der sechs verbindlichen Sätze aus `START-HERE.md` ein. Sie ersetzt
nicht die historische v0.1-Datei, sondern verschärft deren Mindestprüfungen zu
einem exakten Regressionstest.

Für jeden Fall enthält sie:

- Request mit Text und Kontext,
- Propositionen und Sense-Auflösung,
- Pattern in stabiler Reihenfolge,
- Zustand, Score, Confidence und Assessability aller sechs Dimensionen,
- vollständige Contributions mit Regel, Evidenz, Dimension, Delta und Grund,
- Reflexionsfrage, Alternativen, Resonanzhinweise und Notizen.

## Abnahmekette

Jeder Fall läuft über dieselbe Suite zweimal:

```text
Golden Request → Analysis Service → exakter Result- und Trace-Vergleich
Golden Request → POST /api/v1/analyze → exakter Result- und Trace-Vergleich
```

Damit schützt das Gate sowohl die Engine als auch JSON-Decoding, HTTP-Defaults
und Serialisierung. Unbekannte Felder in der Suite werden abgelehnt.

## Erster vollständiger Satz

```bash
go run ./cmd/analyze \
  -context SELF_TALK \
  -text 'Ich muss das heute unbedingt noch schaffen.'

go run ./cmd/analyze \
  -trace \
  -context SELF_TALK \
  -text 'Ich muss das heute unbedingt noch schaffen.'
```

Der Trace enthält fünf Contributions zu `VOLITION`, `OPENNESS` und `CLARITY`.
Die übrigen drei Dimensionen bleiben mangels Evidenz `NOT_ASSESSABLE` mit
`score: null`.

## Release Gate

Lokal:

```bash
go test ./internal/golden -run '^TestVerticalSliceGolden$' -count=1
```

Der GitHub-Actions-Workflow führt diesen Befehl als benannten Golden Gate aus.
Danach folgt die vollständige Testsuite mit aktiviertem Race Detector. Ein
Golden-Unterschied stoppt damit Push- und Pull-Request-Prüfungen.
