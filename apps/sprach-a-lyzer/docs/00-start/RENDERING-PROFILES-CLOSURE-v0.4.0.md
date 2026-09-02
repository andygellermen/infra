# Sprach-A-Lyzer – Rendering & Corporate/Private Profiles Closure v0.4.0

- **Status:** RELEASED
- **Version:** 0.4.0
- **Stand:** 2. September 2026
- **Git-Tag:** `v0.4.0`
- **Owner:** Product & Engineering

## Ergebnis

Der Roadmap-Meilenstein **v0.4 – Rendering & Corporate/Private Profiles** ist
geschlossen. Der unveränderte Acht-Fragen-Core wird durch einen eigenen,
fail-closed Rendering-Katalog ergänzt. Alle Fragen besitzen redaktionell
freigegebene Private- und Corporate-Varianten in Standard und Easy. Eine
Deep-Reflective-Variante existiert ausschließlich privat und erfordert ein
explizites Opt-in.

Der Katalog enthält 40 Renderings. `DEFAULT`, `SIMPLIFY` und `REPHRASE`
werden deterministisch aufgelöst. Fehlt eine freigegebene sichere Variante
oder das erforderliche Opt-in, liefert die Runtime das Standard-Rendering mit
einem maschinenlesbaren Fallback-Grund. Sie erzeugt keine Formulierung live.

## Isolationsmodell

```text
Canonical Question v0.1
        │ unveränderte Question ID + Construct Intent
        ▼
Question Rendering Catalogue v0.1
        ├── PRIVATE / STANDARD
        ├── PRIVATE / EASY
        ├── PRIVATE / DEEP_REFLECTIVE (Opt-in)
        ├── CORPORATE / STANDARD
        └── CORPORATE / EASY
```

Jedes Rendering trägt Leadingness, Specificity, Intimacy,
Spiritual Explicitness und Relational Warmth. Die Runtime berechnet die
Abweichungen zum jeweiligen Profil-Standard und verwirft Varianten außerhalb
der freigegebenen Grenzen.

## Guardrails

- Question ID und Construct Intent müssen dem kanonischen v0.3-Katalog
  entsprechen.
- Renderings und Qualitätsinformationen sind immer `scoring: false`.
- Corporate hat `spiritual_explicitness = 0`, begrenzte Intimität und niemals
  einen Deep-Reflective-Modus.
- Deep Reflective ist ausschließlich `PRIVATE` und `requires_opt_in: true`.
- Diagnose-, Trait-, Ranking- und Leistungsbewertungssprache wird abgewiesen.
- Fehlende Varianten fallen auf das freigegebene Profil-Standard-Rendering
  zurück; kanonische Schlüssel werden nicht als sichtbare Texte ausgegeben.

Golden- und Paritätstests beweisen zusätzlich, dass Private und Corporate
dieselben Propositionen, Pattern, Dimensionswerte und wirksamen
Contributions erzeugen. Nur die Darstellung darf variieren.

## Öffentlicher Vertrag

```text
POST /api/v4/questions/render → Question Rendering Result v0.1
```

Beispiel:

```json
{
  "question_id": "CQ009",
  "profile": "CORPORATE",
  "action": "SIMPLIFY"
}
```

Der lokale Befehl `go run ./cmd/qa -render` verwendet denselben eingebetteten
Katalog. Die bestehenden APIs v1 bis v3 bleiben unverändert verfügbar.

## Release-Vektor

| Artefakt | Version |
|---|---:|
| Core Release / HTTP API | 0.4.0 / 4 |
| Canonical Question / Question Catalogue | 0.1 / 0.1 |
| Question Rendering / Rendering Catalogue | 0.2 / 0.1 |
| Question Rendering Result / Golden | 0.1 / 0.1 |
| Q/A Observation / Selection / Session | 0.1 / 0.1 / 0.1 |
| Presentation Bundle | 0.2 |
| Policy Registry / PostgreSQL Schema | 0.7 / 4 |

Die vollständige Entsprechung liegt in
`data/seed/sprach-a-lyzer_release-manifest_v0.4.0.json`.

## Closure-Gate

```bash
bash ./scripts/verify-v0.4-closure.sh
```

Das Gate führt zunächst sämtliche v0.3-Nachweise aus und ergänzt Schema-/Go-
Parität, acht Rendering-Goldens, vollständige Profilabdeckung,
Core-Score-Parität, Corporate-Leakage-Schutz, Deep-Opt-in,
Aktivierung/Deaktivierung/Reaktivierung, HTTP v4 und historische
Release-Manifest-Tests.

## Bewusste Folgestufen

Der Release persistiert keine Rendering-Interaktionsereignisse. Managed
Import, redaktionelle Freigabeworkflows und Rollback folgen mit v0.5;
aggregierte Friction-/Variant-Metriken und experimentelle Vergleiche sind
gemäß Roadmap Bestandteil von v0.9. Damit werden aus Rephrase-Requests weder
psychologische Scores noch voreilige Kausalbehauptungen abgeleitet.

## Nächster Roadmap-Schritt

Als nächstes folgt **v0.5 – Managed Knowledge Operations**: sichere
XLSX/CSV/JSON-Imports mit Mapping, Diff, Konflikten, Golden Dry Run,
Commit/Rollback und Audit.
