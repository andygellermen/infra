# Sprach-A-Lyzer

Der Sprach-A-Lyzer ist der gemeinsame, deterministische und erklärbare
Analyse-Core für die Produktprofile **Sprachkompass** (Corporate) und
**MeineSprache** (Private).

## Einstieg

1. [START HERE](docs/00-start/START-HERE.md)
2. [Cody Handoff](docs/00-start/CODY-HANDOFF.md)
3. [Developer Handoff](docs/00-start/DEVELOPER-HANDOFF-v0.1.md)
4. [Roadmap](docs/00-start/ROADMAP.md)
5. [Documentation Manifest](docs/00-start/DOCUMENTATION-MANIFEST.md)

## Repository-Struktur

```text
docs/
  00-start/          Einstieg, Handoffs, Roadmap und Dokumentenregeln
  10-product/        Produktkonzepte und Produktprofile
  20-domain-model/   Fachliches Datenmodell, Konstrukte und Pattern
  30-engine/         Resolver, Scoring, Assessability und Q/A-Engine
  40-golden/         Golden-Korpora, Simulationen und Gap-Berichte
  50-import/         Bulk- und Managed-Import-Spezifikationen
  60-ux/             UX- und Admin-Wireframes
  70-architecture/   Technologie-, Datenschutz- und Lizenzkonzepte

data/
  seed/              Kanonische Start- und Redaktionsdaten
  golden/            Maschinenlesbare Golden- und Simulationsdaten
  import-examples/   Importvorlagen und Beispiel-Batches

schemas/
  analysis/          Reserviert für Analyseverträge
  questions/         Q/A-Verträge
  imports/           Importverträge
  rules/             Reserviert für Regelverträge
```

## Artefaktregeln

- Markdown ist die lesbare fachliche Dokumentation.
- JSON ist das kanonische Austausch- und Testformat.
- CSV und XLSX sind redaktionelle Arbeits- und Importformate.
- Ausführbarer Anwendungscode wird außerhalb von `docs/`, `data/` und
  `schemas/` angelegt.
- Bestehende Versionen werden nicht überschrieben oder gelöscht.

## Noch vor der fachlichen Implementierung

Die in [CODY-HANDOFF](docs/00-start/CODY-HANDOFF.md) festgehaltenen Starttickets
bleiben bestehen. Insbesondere wird die fachliche Migration von `FREE_WILL` zu
`VOLITION` separat und testbar durchgeführt.
