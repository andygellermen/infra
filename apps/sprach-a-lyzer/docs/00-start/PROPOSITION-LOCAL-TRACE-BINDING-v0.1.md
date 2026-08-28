# Proposition-local Trace Binding v0.1

**Status:** APPROVED
**Version:** 0.1
**Stand:** 28. August 2026
**Roadmap-Ziel:** v0.2 – Context & Proposition
**Sprint:** v0.2C-A – Proposition Context & Contribution Provenance

## Ergebnis

`TargetType` und `ExpectationSource` werden nicht mehr einmal für den
Gesamttext, sondern für jede Proposition berechnet. Resolver-Fakten behalten
ihre Proposition IDs bis zur wirksamen Contribution. Damit ist erklärbar,
welcher lokale Textteil eine Regel und ihren Dimensionsbeitrag gestützt hat.

## Kanonische Laufzeitregeln

1. Jeder Proposition Node trägt seinen eigenen `target_type` und seine eigene
   `expectation_source`.
2. Der Resolver-Aggregatwert übernimmt genau einen eindeutigen, nicht
   fallbackenden Wert. Bei widersprüchlichen lokalen Werten fällt nur das
   Aggregat auf `UNKNOWN` beziehungsweise `UNSPECIFIED` zurück; die lokalen
   Werte bleiben erhalten.
3. Phrase-, Token-, Sense-, Target-, Expectation-, Relations- und
   Proposition-Feature-Fakten werden mit ihren Proposition IDs geführt.
4. `AND` und erfolgreiche `OR`-Zweige vereinigen die IDs ihrer positiven
   Evidenz. Eine Relation referenziert Quell- und Ziel-Proposition.
5. Negative Bedingungen erfinden keine positive Proposition-Provenienz.
6. Von Regeln erzeugte Pattern und ausgewählte Senses übernehmen die
   Provenienz der auslösenden Bedingung.
7. Jede tatsächlich veröffentlichte Contribution besitzt eine zum
   Contribution-Index ausgerichtete Liste von Proposition IDs.

## Vertrags- und Kompatibilitätsgrenzen

- Analysis Result v0.1 bleibt unverändert; interne Provenienz wird dort nicht
  serialisiert.
- Analysis Trace v0.1 und `Trace()` bleiben unverändert.
- Analysis Trace v0.2 ergänzt `contract_version`, proposition-lokale Kontexte
  und `proposition_ids` je Contribution.
- Resolver Result v0.2 bleibt JSON-kompatibel. Die Zuordnung eines Resolver
  Sense zu seiner Proposition ist interne Laufzeit-Provenienz.
- Policy Registry v0.5 registriert Analysis Trace v0.2. Der unveränderliche
  Resolver Catalogue v0.1 referenziert weiterhin Policy Registry v0.4.

Die CLI veröffentlicht den neuen Vertrag ausdrücklich mit:

```bash
go run ./cmd/analyze -trace-v2 \
  -context SELF_TALK \
  -text 'Ich muss das heute unbedingt noch schaffen.'
```

`-trace` bleibt der Legacy-Ausgabepfad für Trace v0.1. Die HTTP-v1-Grenze
bleibt in diesem Sprint unverändert; eine öffentliche Trace-v0.2-Route wird
erst mit einem eigenen API-Vertrag freigegeben.

## Golden Gate

[`sprach-a-lyzer_proposition-trace_v0.1.json`](../../data/golden/sprach-a-lyzer_proposition-trace_v0.1.json)
sichert einen lokalen Erwartungsfall und eine über zwei Propositionen
gespannte Phrase. Zusätzlich prüfen Unit- und Schema-Tests:

- voneinander abweichende lokale Target-/Expectation-Werte,
- konservativen Aggregat-Fallback bei widersprüchlichen Targets,
- Proposition IDs aus Catalogue-Fakten und weitergereichten Pattern/Senses,
- referenzielle Integrität aller Contribution IDs,
- unveränderte JSON-Form des Analysis Result v0.1 und Trace v0.1.

## Noch nicht enthalten

Dieser Sprint schließt nicht den gesamten Roadmap-Meilenstein v0.2. Offen
bleiben insbesondere tiefere syntaktische Rollen, proposition-lokale
Kompositionsregeln jenseits der vorhandenen Catalogue-Fakten und die später
dediziert Golden-gesicherte fachliche Kalibrierung.
