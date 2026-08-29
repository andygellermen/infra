# Construct Ontology Contracts v0.2

**Status:** APPROVED
**Version:** 0.2
**Stand:** 30. August 2026
**Roadmap-Ziel:** v0.2 – Context & Proposition
**Sprint:** v0.2C-B – Construct Ontology Contracts

## Ergebnis

Der bisherige Construct Review ist in einen strikten, maschinenlesbaren
Fachvertrag überführt. Die Ontology umfasst 36 kanonische IDs in vier Ebenen:

1. `LANGUAGE_FEATURE` – direkt beobachtbare Sprachmerkmale,
2. `CONTEXTUAL_CONSTRUCT` – vorsichtig inferierbare Aussagen zum aktuellen
   Äußerungskontext,
3. `WORKING_HYPOTHESIS` – ausschließlich qualifizierte Arbeitshypothesen,
4. `REFLECTIVE` – freiwillige Reflexionsangebote ohne Core Score.

Jeder Eintrag definiert positive Evidenz, Nicht-Evidenz, eine zulässige
Aussage, verbotene Aussagen, Claim Mode, Assessability-Anforderungen und
nicht-scoring Dimension Links. Die ursprünglichen sechs Dimensionen bleiben
damit die einzige Aggregations- und Darstellungsschicht.

## Kanonische Grenzen

- `core_scoring` ist für jeden Construct `false`.
- Fehlende Evidenz erzeugt weder Construct noch neutralen Ersatzwert.
- `WORKING_HYPOTHESIS` darf nur als Hypothese gekennzeichnet erscheinen.
- `REFLECTIVE` benötigt Fragekontext und darf niemals den Core Score verändern.
- Dimension Links beschreiben fachliche Nähe, keine automatische Contribution.
- Nur eine separat versionierte Regel darf aus Construct-Evidenz eine
  Contribution erzeugen.

Policy Registry v0.6 ergänzt dazu drei unveränderliche Guardrails:

- `CONSTRUCT_REQUIRES_EXPLICIT_EVIDENCE`
- `WORKING_HYPOTHESIS_REQUIRES_HEDGING`
- `REFLECTIVE_CONSTRUCT_CANNOT_SCORE`

Policy Registry v0.5 und alle älteren Verträge bleiben unverändert erhalten.

## Runtime-Vorbereitung

Sechs Constructs enthalten erste deterministische Runtime-Signale:

- `PERSPECTIVE_TAKING`
- `BOUNDARY_CLARITY`
- `CONTEXTUAL_AGENCY`
- `ARTICULATED_LEARNING`
- `CONTROL_PRESSURE_INTERPRETATION`
- `PERSON_BEHAVIOR_LABELING`

Daraus sind drei nicht direkt scorende Kompositionen definiert:
`RESPECTFUL_BOUNDARY`, `AGENCY_RECOVERY` und `LEARNING_RECOVERY`.
Die Ausführung und Regelbindung ist in Sprint v0.2C-C erfolgt; siehe
[`CONSTRUCT-RUNTIME-PROPOSITION-COMPOSITION-v0.1.md`](CONSTRUCT-RUNTIME-PROPOSITION-COMPOSITION-v0.1.md).

## Kanonische Artefakte

- [`sprach-a-lyzer_construct-ontology_v0.2.json`](../../data/seed/sprach-a-lyzer_construct-ontology_v0.2.json)
- [`sprach-a-lyzer_construct-ontology_v0.2.json`](../../schemas/constructs/sprach-a-lyzer_construct-ontology_v0.2.json)
- [`sprach-a-lyzer_policy-registry_v0.6.json`](../../data/seed/sprach-a-lyzer_policy-registry_v0.6.json)
- [`sprach-a-lyzer_policy-registry_v0.6.json`](../../schemas/rules/sprach-a-lyzer_policy-registry_v0.6.json)

Schema-, Seed-, Go-ID- und Historientests bilden das Contract-Golden-Gate.
