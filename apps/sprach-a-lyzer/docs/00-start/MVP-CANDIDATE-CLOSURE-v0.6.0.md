# Sprach-A-Lyzer – MVP Candidate Closure v0.6.0

- **Status:** RELEASED
- **Version:** 0.6.0
- **Stand:** 3. September 2026
- **Git-Tag:** `v0.6.0`
- **Owner:** Product & Engineering

## Ergebnis

Der erste sichtbare MVP Candidate ist geschlossen. Die bisherigen Core-,
Resolver-, Q/A-, Rendering- und Managed-Knowledge-Fähigkeiten bilden nun einen
gemeinsamen End-to-End-Produktpfad für **MeineSprache** und
**Sprachkompass**.

Die Web-Oberfläche unter `/` bietet:

- private oder berufliche Profilwahl,
- Kontext und Standard-/Easy-Sprache,
- deterministische Analyse ohne KI,
- sechs Dimensionen mit sichtbarer Assessability,
- Muster und Contribution Trace,
- redaktionell freigegebene Reflexion und Alternativen,
- fünf adaptive Anschlussfragen,
- eine progressive, transiente Q/A-Session,
- lokales Sitzungsfeedback ohne Telemetrie.

## Privacy- und Aussagegrenzen

Der v0.6-Pfad speichert weder Rohtext noch Analyse noch Antworten. Es gibt
keine externe Übertragung und keine KI-Ausführung. Der Response enthält einen
expliziten Privacy Receipt mit vier negativen Zuständen. Nicht assessable
Dimensionen werden nicht mit einem künstlichen Mittelwert gefüllt.

Die Erklärung beschreibt Sprache im Text, niemals Eigenschaften, Diagnose,
Eignung oder Leistung eines Menschen. Private/Corporate und Standard/Easy
bleiben reine Darstellungsentscheidungen; Core-Ergebnisse sind paritätsgleich.

## Öffentlicher Vertrag

```text
GET  /
GET  /admin
POST /api/v6/experience/analyze
```

MVP Experience Request und Result sind als strikte JSON-Schemas v0.1
versioniert. Bestehende APIs v1–v5 bleiben kompatibel.

## Admin-Basis

Die Admin-Hülle zeigt Readiness, Core-/Privacy-Modus sowie die sicheren
Knowledge-, Test- und Audit-Grenzen. Sie enthält absichtlich keine
schreibenden Operations. Der produktive Publish-/Rollback-Zugang bleibt bis
zu einem vertrauenswürdigen Auth-Adapter außerhalb der öffentlichen UI.

## Sicherheitsheader

Alle Antworten verwenden `Cache-Control: no-store`, `nosniff`, Frame Denial,
No-Referrer, eine restriktive Permissions Policy und eine self-only Content
Security Policy. Die Oberfläche lädt keine externen Schriften, Skripte,
Bilder oder Tracker.

## Golden und Release-Gate

`data/golden/sprach-a-lyzer_mvp-experience_v0.1.json` sichert drei
End-to-End-Fälle für Internal Pressure, Safety und Respectful Boundary. Das
Gate prüft zusätzlich:

- bitgleiche Core-Parität,
- sechs Dimensionsansichten,
- profilisolierte Labels und Easy-Fragen,
- strikt abgelehnte Speicher-/Ranking-Felder,
- Privacy Receipt und No-AI-Modus,
- eingebettete Produkt- und Admin-Artefakte,
- alle historischen v0.1–v0.5 Closure-Gates.

```bash
bash ./scripts/verify-v0.6-closure.sh
```

## Release-Vektor

| Artefakt | Version |
|---|---:|
| Core Release / HTTP API | 0.6.0 / 6 |
| MVP Experience Request / Result | 0.1 / 0.1 |
| MVP Experience Golden | 0.1 |
| Analysis / Trace | 0.1 / 0.2 |
| Question Session / Rendering | 0.1 / 0.2 |
| Managed Import | 0.1 |
| PostgreSQL Schema | 5 |

## Bewusste Grenzen

- Kein KI-Adapter, KI-Consent oder generatives Rephrasing; das folgt frühestens
  in v0.7.
- Kein Audio oder ASR; das bleibt v0.8.
- Keine persistente Sessionhistorie, Telemetrie oder Trajectory-Langzeitakte.
- Keine schreibende Admin-UI ohne vorgelagerte Authentisierung.
- Die offene Kalibrierung historischer Gap-Fälle bleibt getrennte Facharbeit
  und wird nicht durch UI-Heuristiken überdeckt.

## Nächster Roadmap-Schritt

Als nächstes folgt **v0.7 – Optional AI Enhancement**. Vor dessen Umsetzung
soll der No-AI-Candidate praktisch erprobt und sein eigenständiger Nutzen
bewertet werden.
