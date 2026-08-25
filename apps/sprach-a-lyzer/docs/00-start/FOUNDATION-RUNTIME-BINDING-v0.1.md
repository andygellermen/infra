# Sprach-A-Lyzer – Foundation Runtime Binding v0.1

- **Status:** APPROVED
- **Version:** 0.1
- **Stand:** 25. August 2026
- **Owner:** Product & Engineering
- **Basiert auf:** Foundation Rule Migration v0.1

## Ergebnis

Die deterministische Engine verwendet den Rule-v0.3-Katalog nun zur Laufzeit:

- Server/API lesen den aktiven `PRODUCTION` Rule Set aus PostgreSQL.
- CLI, Unit- und Golden-Tests verwenden den zur Build-Zeit unverändert
  eingebetteten Foundation Seed v0.2.
- Aktivierung, Condition Matching, Pattern-Aktionen, die im Seed bereits
  typisierten Contributions, Sinnauswahl und Resonanzhinweise stammen aus dem
  Katalog.
- Ein fehlender, leerer oder ungültiger Katalog beendet die Analyse fail-closed.
- Eine deaktivierte Regel kann weder Pattern noch Contributions erzeugen.

## Deterministische Ausführung

Der Runtime-Evaluator unterstützt rekursive `AND`-, `OR`- und `NOT`-Bäume,
kanonische Text-, Token-, Kontext- und Input-Mode-Fakten sowie die in den sechs
Foundation-Regeln tatsächlich verwendeten Vergleichsoperatoren. Deutsche
Flexionsformen `muss`, `musst` und `müssen` werden für das kanonische Token
`müssen` zusammengeführt.

Aktiv ausgeführt werden:

- `ADD_PATTERN`
- `ADD_CONTRIBUTION`
- `SELECT_SENSE`
- `ADD_RESONANCE_HINT` mit zwingendem `semantic_score: false`

Andere zwar registrierte, aber im Foundation Set nicht verwendete Aktionen oder
Runtime-Felder werden nicht still ignoriert, sondern fail-closed abgelehnt.

## Bewusst noch deterministisch im Core

Golden-Erklärtexte, Assessability-Enrichments, zusammengesetzte Pattern,
Alternativformulierungen und Reflexionsfragen bleiben vorerst kontrollierte
Core-Bausteine. Ebenso bleibt `REPORTED_CLAIM` ein Resolver-Fall außerhalb der
sechs Foundation-Regeln. Ihre spätere Überführung benötigt einen eigenen
versionierten Contract und ist keine Voraussetzung für diese Laufzeitanbindung.

## Smoke-Test

```bash
go test ./...

go run ./cmd/analyze \
  -context SELF_TALK \
  -text 'Ich muss das heute unbedingt noch schaffen.'

go run ./cmd/analyze \
  -context MODERATION \
  -text 'Hast du Geld?'
```

Der PostgreSQL-Integrationstest schaltet `R-INTERNAL-PRESSURE` temporär ab und
weist nach, dass der API-Composition-Root die Änderung unmittelbar im
Laufzeitkatalog berücksichtigt. Die Regel wird durch Test-Cleanup reaktiviert.
