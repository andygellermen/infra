# Sprach-A-Lyzer – Resolver Catalogue Contracts v0.1

**Status:** APPROVED
**Version:** 0.1
**Last updated:** 26. August 2026
**Owner:** Sprach-A-Lyzer Core
**Roadmap-Ziel:** v0.2 – Context & Proposition
**Sprint:** v0.2B-A – kanonische IDs, Resolver Catalogue und Guardrails

## Ergebnis

Der Resolver besitzt nun einen maschinenlesbaren und strikt geprüften
Fachvertrag. Sprint v0.2B-A fixiert die Sprache zwischen Resolver Result v0.2,
Policy Registry v0.4 und dem neuen Resolver Catalogue v0.1, ohne die
veröffentlichte Laufzeit stillschweigend umzubinden.

## Versionierter Vertragsvektor

| Artefakt | Version | Aufgabe |
|---|---:|---|
| Resolver Result | 0.2 | Ausgabe des Context-/Proposition-Resolvers |
| Resolver Catalogue | 0.1 | Lexeme, Sense-Kandidaten, Konnektoren, Scope-Regeln und Schwellen |
| Policy Registry | 0.4 | kanonische IDs und nicht editierbare Guardrails |

Die bisherigen Seeds und Schemas bleiben unverändert erhalten. v0.4 erweitert
die Policy Registry additiv um die Resolver-IDs und drei Hard Guardrails.

## Kanonische Resolver-IDs

Policy Registry und Go-Core besitzen eine gemeinsame Quelle für:

- `Actor`: `SELF`, `OTHER_PERSON`, `GROUP_SELF`, `UNKNOWN`,
- `Modality`: `NONE`, `NECESSITY`, `POSSIBILITY`, `PERMISSION`,
  `EXPECTATION`, `INTENTION`, `PROBABILITY`,
- `NegationScope`: `NONE`, `PROPOSITION`, `MODALITY`, `ACTOR`, `AMBIGUOUS`,
- `SenseState`: `HIGH`, `MEDIUM`, `AMBIGUOUS`,
- `AmbiguityType`: `SEMANTIC`, `PHONETIC`, `ORTHOGRAPHIC`, `PRAGMATIC`,
  `SYNTACTIC`, `REGISTER`, `RESONANCE`.

Target Types, Expectation Sources und Discourse Relations werden ebenfalls aus
der Policy Registry bezogen. Die Domain-Typen sind Aliase dieser kanonischen
IDs; lokale Dubletten sind damit ausgeschlossen.

## Resolver Catalogue v0.1

Der APPROVED Seed enthält acht Lexemfamilien:

```text
MUSSEN, SOLLEN, DUERFEN, FREI,
PROBLEM, FEHLER, UMFAHREN, EIGENTLICH
```

Er registriert außerdem alle acht Relationsklassen über Konnektoren:

```text
CONTRAST, CONCESSION, CAUSE, CONSEQUENCE,
ADDITION, CONDITION, CORRECTION, DISCOUNTING
```

Die priorisierten Negationsregeln unterscheiden zunächst Akteur-, Modalitäts-
und Proposition-Scope. Unzureichende Evidenz darf später ausdrücklich zu
`AMBIGUOUS` führen.

### Sense-Schwellen

Die fachlich dokumentierten Startwerte sind:

| Zustand | Mindest-Confidence | Mindestabstand zum zweiten Kandidaten |
|---|---:|---:|
| `HIGH` | 0,75 | 0,20 |
| `MEDIUM` | 0,60 | 0,10 |
| Fallback | `AMBIGUOUS` | – |

Diese Werte klassifizieren die Sicherheit einer Sense-Auswahl. Sie sind keine
Kalibrierung der sechs Analyse-Dimensionen. Eine spätere Änderung benötigt ein
neues Catalogue-Artefakt, Approval und Golden-Nachweise.

## Hard Guardrails

1. `AMBIGUOUS_FEATURE_CANNOT_HARD_SCORE` – ein mehrdeutiges Feature darf
   keinen harten Score auslösen.
2. `PROPOSITION_SPAN_MUST_MATCH_SOURCE` – jeder Proposition-Span muss exakt
   auf seinen Abschnitt im Quelltext zeigen.
3. `RESOLVER_CANDIDATE_CANNOT_BYPASS_RULES` – Kandidaten sind Evidenz; sie
   dürfen die Rule Engine nicht umgehen.

Alle drei IDs sind in Policy Registry v0.4 nicht editierbar, im Catalogue
verpflichtend und zwischen Schema, Seed und Code driftgesichert.

## Technische Guardrails des Catalogue Imports

Der Go-Decoder verwirft unbekannte Felder und nachgestellte JSON-Werte. Er
prüft Version, Approval, Locale, eindeutige Schlüssel und Signale, gültige
Policy-IDs, begrenzte Confidences, sinnvolle Schwellen sowie den exakten Satz
der Catalogue-Guardrails.

## Bewusste Sprintgrenze

v0.2B-A definiert und prüft den Vertrag. Der bestehende deterministische
Resolver liest den Catalogue noch nicht zur Laufzeit. Dadurch ändern sich in
diesem Sprint weder Analyseergebnisse noch der öffentliche API-Vertrag.

Der nächste Schritt **v0.2B-B – Catalogue Runtime Binding** ersetzt die noch
fest codierten Resolver-Tabellen durch den geprüften Catalogue, bindet die
drei Guardrails an die tatsächlichen Ausführungspfade und sichert das Ergebnis
mit Resolver-Golden-, Core-Paritäts- und Aktivierungs-/Deaktivierungstests.

## Kanonische Dateien

- [`sprach-a-lyzer_resolver-catalogue_v0.1.json`](../../data/seed/sprach-a-lyzer_resolver-catalogue_v0.1.json)
- [`sprach-a-lyzer_resolver-catalogue_v0.1.json`](../../schemas/analysis/sprach-a-lyzer_resolver-catalogue_v0.1.json)
- [`sprach-a-lyzer_policy-registry_v0.4.json`](../../data/seed/sprach-a-lyzer_policy-registry_v0.4.json)
- [`sprach-a-lyzer_policy-registry_v0.4.json`](../../schemas/rules/sprach-a-lyzer_policy-registry_v0.4.json)
