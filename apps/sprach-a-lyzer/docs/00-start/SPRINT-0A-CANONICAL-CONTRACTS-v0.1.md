# Sprach-A-Lyzer – Sprint 0A Canonical Contracts v0.1

- **Status:** APPROVED
- **Version:** 0.1
- **Stand:** 25. August 2026
- **Owner:** Product & Engineering
- **Supersedes:** –

## Zweck

Diese Baseline trennt drei Ebenen, die nicht vermischt werden dürfen:

```text
Engine-locked Guardrails
        ↓ schützen
versionierte Contracts und IDs
        ↓ begrenzen
justierbare Rules und Parameters
```

Fachliche Regeln und Parameter sollen später bewusst kalibriert werden können.
Guardrails bleiben außerhalb des Admin-Modells und können weder durch Importe,
Rule Sets, Parameter Sets noch KI-Adapter überschrieben werden.

## Kanonische Quellen

| Gegenstand | Kanonische Quelle |
|---|---|
| Dimensionen und Legacy-Mapping | [`sprach-a-lyzer_canonical-dimensions_v0.1.md`](../20-domain-model/sprach-a-lyzer_canonical-dimensions_v0.1.md) |
| Analyse- und Trace-Grenze | [`sprach-a-lyzer_analysis-trace-contracts_v0.1.md`](../20-domain-model/sprach-a-lyzer_analysis-trace-contracts_v0.1.md) |
| IDs, Versionsvektor, Privacy, Flags, Guardrails | [`sprach-a-lyzer_policy-registry_v0.1.json`](../../data/seed/sprach-a-lyzer_policy-registry_v0.1.json) |
| Policy-Registry-Schema | [`sprach-a-lyzer_policy-registry_v0.1.json`](../../schemas/rules/sprach-a-lyzer_policy-registry_v0.1.json) |
| Strikter Rule-Contract | [`sprach-a-lyzer_rule_v0.2.json`](../../schemas/rules/sprach-a-lyzer_rule_v0.2.json) |
| Justierbarer Parameter | [`sprach-a-lyzer_parameter_v0.1.json`](../../schemas/rules/sprach-a-lyzer_parameter_v0.1.json) |
| Canonical Question | [`sprach-a-lyzer_question-canonical_v0.1.json`](../../schemas/questions/sprach-a-lyzer_question-canonical_v0.1.json) |
| Question Rendering | [`sprach-a-lyzer_question-rendering_v0.1.json`](../../schemas/questions/sprach-a-lyzer_question-rendering_v0.1.json) |

Die Go-Konstanten unter `internal/policy` sind der Compile-Time-Vertrag. Ein
Drift-Test vergleicht sie mit der versionierten Registry.

Für die in diesem Dokument geregelten IDs, Contracts, Defaults und Guardrails
hat Sprint 0A Vorrang vor älteren Konzept- und Handoff-Dokumenten. Deren übrige
fachliche Aussagen und historische Referenzwerte bleiben erhalten.

## Verbindliche ID-Gruppen

Die Registry schreibt fest:

- sechs Dimensionen mit `VOLITION`; `FREE_WILL` bleibt nur Legacy-Alias,
- vier Assessability-Zustände,
- Private- und Corporate-Profil,
- elf Analysekontexte,
- drei Input-Modi; Analyse-Request v0.1 akzeptiert weiterhin nur `TEXT`,
- fünf Resonanzmodi,
- Evidenzklassen A–E,
- vier Inferenzklassen und fünf Kausalitätsstufen,
- Target Types, Expectation Sources und Discourse Relations,
- Rule Scopes, Rule Statuses und zwölf Rule Actions.

Eine Erweiterung ist eine Vertragsänderung. Neue Pflichtwerte oder geänderte
Semantik benötigen eine neue Version und passende Golden Cases.

## Versionsvektor

Ein Ergebnis wird mindestens identifiziert durch:

```text
analysis_request
analysis_result
analysis_trace
knowledge_base
rule_set
parameter_set
presentation_bundle
golden_suite
```

Der Registry-Versionsvektor benennt zusätzlich Rule- und Question-Contracts.
Veröffentlichte Versionen werden nicht still verändert. Änderungen entstehen
als Draft und durchlaufen Contract- und Golden-Tests sowie eine Freigabe.

## Justierbare Fachlogik

Rule Sets dürfen innerhalb ihrer Contracts Conditions, Pattern-Komposition,
Priorität, Aktivierung, begrenzte Contributions, Confidence Modifier,
Explanation-/Reflection-Keys, Suppression und Caps verändern.

Parameter Sets dürfen innerhalb definierter Grenzen Assessability-Gewichte und
-Schwellen, Ambiguity Modifier, Aggregation Scale, Frequency Alpha,
Weakest-Link-Parameter und freigegebene Dimensionsgewichte verändern.

Der Parameter-Contract verlangt Typ, Default, Minimal-/Maximalwert,
Editierbarkeit, Approval-Pflicht, Status und Version. Engine-locked Guardrails
sind keine Parameter und können auf diesem Weg nicht abgeschwächt werden.

Jede fachlich wirksame Änderung folgt:

```text
Draft → Contract Validation → Golden Dry Run → Review → Approval → Publish
```

Ein früherer `PASS` darf nicht ohne dokumentierte fachliche Freigabe zu einem
`GAP` werden. Veröffentlichte Sets sind immutable; Korrekturen erzeugen neue
Versionen.

## Hard Guardrails

Die Registry enthält 18 nicht editierbare Schutzregeln. Sie sichern:

- `NOT_ASSESSABLE` und `score: null` bei fehlender Evidenz,
- begrenzte und endliche Rechenwerte,
- keine zirkulären oder unendlichen Regelketten,
- keine semantische Homophon-Vererbung,
- keine Diagnose, Trait-Behauptung oder Beschäftigten-Rangfolge,
- Question Score Bias immer null,
- keine Assessability aus einer Frage allein,
- keine Kernscore-Wirkung reiner Resonanz,
- keinen Canonical-/Private-Fallback im Corporate-Profil,
- keine autonome Scoreentscheidung durch ein LLM,
- WingScore ausschließlich als Bewertung des Textes,
- Rohtext- und Audiospeicherung nur nach ausdrücklichem Opt-in.

Sie dürfen im Admin-Bereich erklärt, aber nie bearbeitet oder deaktiviert werden.

## Soft Guardrails

Soft Guardrails verlangen Review und Freigabe:

| Parameter | Warnung oberhalb |
|---|---:|
| absoluter Base Contribution | 50.0 |
| Resonance Factor | 0.6 |
| Frequency Alpha | 0.5 |

Auch eine Änderung dieser Governance-Grenzen benötigt Approval und Golden Dry Run.

## Privacy und Feature Defaults

```text
raw text: process → analyze → discard
```

Analyse-, Rohtext- und Audiospeicherung sind standardmäßig deaktiviert;
persönliche Historie erfordert Opt-in. Managerzugriff auf individuelle
Corporate-Analysen, Beschäftigten-Ranking und HR-Auswahl bleiben deaktiviert.

Alle neun Feature Flags starten mit `false`. KI, LLMs, ASR, Prosodie sowie
Text- und Audiospeicherung sind kein stiller Bestandteil des Core-Pfads.

## Rule Contract v0.2

Rule v0.1 bleibt historischer Foundation-Vertrag. v0.2 ist die künftige
Authoring-Grenze und definiert rekursive `AND`-/`OR`-/`NOT`-Conditions,
typisierte Prädikate, zwölf Action-Typen, kanonische Dimensionen, Wertebereiche
und Provenienz. Die Foundation-Regeln werden erst in einem eigenen,
Golden-gesicherten Schritt migriert und nicht still umgedeutet.

## Canonical Question und Rendering

Der Canonical Contract enthält Intent, Konstrukte, Phase, Audience, Risiko und
Semantik, aber keinen sichtbaren Text. Das Rendering enthält den Text und
Qualitätswerte, darf jedoch den Construct Intent nicht neu definieren.
Deep-reflective Renderings erfordern Opt-in; Corporate-Renderings besitzen
spirituelle Explizitheit null.

> Eine Frage darf anders formuliert werden, aber nicht heimlich etwas anderes fragen.

## Abnahme Sprint 0A

- [x] kanonische ID-Gruppen und Versionsvektor festgeschrieben
- [x] `FREE_WILL` nur als Legacy-Alias
- [x] Analyse-/Result-/Trace v0.1 unverändert stabil
- [x] strikter Rule-Contract v0.2
- [x] begrenzter und approval-pflichtiger Parameter-Contract v0.1
- [x] Canonical Question und Rendering getrennt
- [x] Privacy- und Feature-Flag-Defaults festgeschrieben
- [x] Hard und Soft Guardrails getrennt
- [x] Contract-Fixtures, Drift- und Konsistenztests vorhanden

## Nächster Schritt

Sprint 0B darf diese Baseline konsumieren. Als nächste fachliche Arbeit folgt
die Golden-gesicherte Migration der sechs Foundation-Regeln auf Rule v0.2; die
breite Regelkalibrierung bleibt bewusst ein eigenes späteres Zeitfenster.
