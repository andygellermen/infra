# Sprach-A-Lyzer – Core Closure v0.1.0

- **Status:** RELEASED
- **Version:** 0.1.0
- **Stand:** 26. August 2026
- **Git-Tag:** `v0.1.0`
- **Owner:** Product & Engineering

## Ergebnis

Der Roadmap-Meilenstein **v0.1 – Deterministic Language Core** ist geschlossen.
Die Engine reproduziert die sechs Acceptance Cases der Golden Suite v0.2
vollständig aus einem versionierten Laufzeitkatalog. Sie benötigt dafür weder
generative KI noch versteckte fallspezifische Scoring- oder Darstellungstexte.

## Release-Vektor

| Artefakt | Version |
|---|---:|
| Core Release | 0.1.0 |
| Analysis Request / Result / Trace | 0.1 |
| Rule Contract | 0.4 |
| Foundation / Rule Set | 0.3 |
| Policy Registry | 0.3 |
| Parameter Contract / Set | 0.1 |
| Presentation Bundle | 0.2 |
| Golden Suite | 0.2 |
| PostgreSQL Schema | 3 |

Die maschinenlesbare Entsprechung liegt in
`data/seed/sprach-a-lyzer_release-manifest_v0.1.0.json`.

## Geschlossene Laufzeitlücken

- Neun Rule-v0.4-Regeln bilden Erkennung, Sense-Auflösung, Pattern,
  Contributions, Erklärungen, Reflexionsfrage, Alternativen und den
  nicht-scorenden Resonanzhinweis ab.
- Priorisierte Phasenausführung stellt neu erzeugte Pattern und Sense-Fakten
  nachfolgenden Regeln deterministisch bereit.
- `MATCHES` verwendet Go-RE2 und begrenzt Ausdrücke auf 512 Zeichen.
- Alle sichtbaren fachlichen Texte werden über versionierte Presentation Keys
  bezogen; fehlende Keys brechen die Analyse fail-closed ab.
- `STOP_RULE_CHAIN` und `stop_processing` beenden die weitere Regelverarbeitung.
- Aktivierungsänderungen im PostgreSQL-Katalog wirken ohne Codeänderung auf die
  Analyse. Standalone-CLI und Tests verwenden dasselbe eingebettete Foundation-
  Artefakt.

## Release-Gates

```bash
go test ./...
go vet ./...
```

Verbindlich grün:

- Golden-v0.2-Parität für alle sechs Vertical-Slice-Fälle über Core und HTTP,
- Aktivierungs-, Deaktivierungs- und Reaktivierungs-Smoke-Test,
- Homophonie-Stop-Guardrail ohne semantisches Scoring,
- strikte Rule-/Policy-/Seed-Vertragsprüfungen,
- Datenbankmigration und Schema-Readiness (wenn `SAL_TEST_DATABASE_URL` gesetzt
  ist).

## Versions- und Tag-Regel

Der App-Release folgt SemVer. Vertrags-, Rule-Set-, Golden- und Datenbankstände
bleiben eigenständig versioniert und werden im Release-Manifest gemeinsam
fixiert. Ein annotierter Git-Tag `vX.Y.Z` wird erst nach grünem Release-Gate auf
dem zugehörigen Release-Commit erstellt. Bestehende Tags und versionierte
Artefakte werden nicht verschoben oder überschrieben.

## Nächster Roadmap-Schritt

Als nächstes beginnt **v0.2 – Context & Proposition**. Die ausdrücklich
reservierte fachliche Kalibrierung von Regelwerten und editierbaren Parametern
erhält danach ein eigenes Review- und Golden-Zeitfenster; dieser Release nimmt
keine verdeckte Neukalibrierung vor.
