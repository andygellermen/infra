# Sprach-A-Lyzer – Resolver Relations & Scope Expansion v0.1

**Status:** APPROVED
**Version:** 0.1
**Last updated:** 28. August 2026
**Owner:** Sprach-A-Lyzer Core
**Roadmap-Ziel:** v0.2 – Context & Proposition
**Sprint:** v0.2B-C – Relations-, Actor- und Scope-Golden-Ausbau

## Ergebnis

Der kataloggebundene Resolver verarbeitet nun alle acht in Resolver Catalogue
v0.1 registrierten Discourse Relations ausführbar. Nach `CONTRAST`,
`CONCESSION` und `DISCOUNTING` sind jetzt auch folgende Klassen durch
Resolver-Goldens und Rule-Engine-Smokes abgesichert:

```text
CAUSE, CONSEQUENCE, CONDITION, ADDITION, CORRECTION
```

Actor-Fokus und Negations-Scope besitzen eine eigene Schutzmatrix. Der Sprint
ändert weder Resolver Contract v0.2 noch Catalogue v0.1 oder Policy Registry
v0.4. Die veröffentlichten Artefakte bleiben unverändert.

## Proposition-Splitting

Die Satzzerlegung arbeitet rekursiv über die aktiven Catalogue-Konnektoren:

- Infix-Marker trennen linke und rechte Proposition.
- `CONDITION`-Präfixe wie `wenn` und `falls` trennen den eingeleiteten
  Bedingungssatz am Komma von der Hauptproposition.
- Präfix-Marker zwischen Sätzen, etwa `deshalb`, verbinden die vorherige mit
  der aktuellen Proposition.
- Mehrere Marker in einem Satz erzeugen mehrere geordnete Nodes und Edges.
- Der mehrdeutige Marker `und` trennt nur bei einem expliziten neuen Actor;
  nominale Koordinationen wie „Brot und Butter“ bleiben eine Proposition.
- Relation und Edge-Confidence stammen ausschließlich aus dem Catalogue.
- Jeder erzeugte Node besteht weiterhin den Source-Span-Guardrail.

Beispiel:

```text
Wenn du Zeit hast, sprechen wir und wir planen morgen.

P0 --CONDITION/wenn--> P1 --ADDITION/und--> P2
```

## Relationsmatrix

| Relation | Golden-Beispiel | Marker | Confidence |
|---|---|---|---:|
| `CAUSE` | „…, weil ich krank bin.“ | `weil` | 0,90 |
| `CONSEQUENCE` | „Deshalb bleibe ich …“ | `deshalb` | 0,90 |
| `CONDITION` | „Wenn du Zeit hast, …“ | `wenn` | 0,90 |
| `ADDITION` | „… und ich melde mich …“ | `und` | 0,82 |
| `CORRECTION` | „… sondern ein Lernschritt.“ | `sondern` | 0,90 |

Alle fünf Relationswerte werden zusätzlich als `discourse_relation`-Fakten
von aktivierten Rule-v0.4-Conditions konsumiert. Der Test endet erst bei einem
von der jeweiligen Regel publizierten Pattern.

## Actor- und Scope-Schutzmatrix

| Form | Actor | Negation Scope |
|---|---|---|
| „Nicht ich muss …“ | `SELF` | `ACTOR` |
| „Nicht du musst …“ | `OTHER_PERSON` | `ACTOR` |
| „Du musst das nicht …“ | `OTHER_PERSON` | `MODALITY` |
| „Du entscheidest das nicht.“ | `OTHER_PERSON` | `PROPOSITION` |

Die Scope-Regeln werden weiterhin in Catalogue-Priorität ausgewertet. Eigene
Runtime-Varianten entfernen Actor- beziehungsweise Modalitäts-Cues und weisen
nach, dass der Resolver dann auf den katalogisierten Proposition-Fallback
zurückfällt. Damit sind die Ergebnisse nicht durch versteckte Code-Defaults
vorbestimmt.

## Golden-Suite v0.3

[`sprach-a-lyzer_relations-scope_v0.3.json`](../../data/golden/sprach-a-lyzer_relations-scope_v0.3.json)
ist ein eigenständiges, strikt dekodiertes Artefakt mit neun Fällen. Es ergänzt
die unveränderten Context-/Proposition-Goldens v0.1 und den additiven
Catalogue-Runtime-Stand v0.2.

Geprüft werden:

- Node-Reihenfolge und Node-Text,
- Actor, Negation, Negations-Scope und Modalität,
- Edge-Endpunkte, Marker, Relation und Confidence,
- Graphintegrität und exakte Source-Spans.

## Regression und Release-Grenze

Die sechs Core-Goldens, der öffentliche Analysevertrag v0.1 und die HTTP-v1-
Ausgabe bleiben unverändert. Ein Git-Release-Tag `v0.2.0` wird weiterhin erst
nach vollständiger Closure des Roadmap-Meilensteins gesetzt.

## Nächster sinnvoller Schritt

Der nächste v0.2-Teilsprint lokalisiert `TargetType` und `ExpectationSource`
pro Proposition und verbindet Resolver-Fakten mit Proposition IDs im
Contribution Trace. Anschließend folgen Construct Ontology v0.2, öffentliche
Resolver-HTTP-Versionierung und das v0.2-Closure-Gate.
