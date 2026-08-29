# Construct Runtime & Proposition Composition v0.1

**Status:** APPROVED
**Version:** 0.1
**Stand:** 30. August 2026
**Roadmap-Ziel:** v0.2 – Context & Proposition
**Sprint:** v0.2C-C – Construct Runtime & Proposition Composition

## Ergebnis

Construct Ontology v0.2 ist fail-closed an die deterministische Engine
gebunden. Runtime-Signale werden je Proposition ausgewertet; daraus entstehende
Construct-Evidenz behält ihre Proposition IDs. Nur katalogisierte
Kompositionen werden als Rule-Fakten veröffentlicht.

Die Ontology erzeugt selbst weder Contributions noch Scores. Rule Contract
v0.5 ergänzt ausschließlich die Bedingungsfelder `construct` und
`composition`; neue Scoring-Aktionen existieren nicht.

## Ausführbare Kompositionen

### RESPECTFUL_BOUNDARY

`PERSPECTIVE_TAKING` vor `BOUNDARY_CLARITY`, maximal eine Proposition Abstand.
Die bestehende Foundation-Regel verwendet jetzt dieses Composition-Fakt statt
einer globalen Drei-Phrasen-Bedingung. Ihre sechs Contributions und deren
Golden-Werte bleiben unverändert; alle verweisen auf `P0` und `P1`.

### AGENCY_RECOVERY

`CONTROL_PRESSURE_INTERPRETATION` vor `CONTEXTUAL_AGENCY`, verbunden durch
`CONTRAST` oder `CORRECTION`. Die Foundation veröffentlicht zunächst nur das
Pattern `AGENCY_RECOVERY`; es entsteht keine neue Contribution.

### LEARNING_RECOVERY

`PERSON_BEHAVIOR_LABELING` vor `ARTICULATED_LEARNING`, ebenfalls verbunden
durch `CONTRAST` oder `CORRECTION`. Die explizite Lernphrase ist positive
Evidenz. Ein ambiger Sense allein genügt nicht. Auch diese Komposition erzeugt
zunächst ausschließlich ein Pattern.

## Runtime- und Guardrail-Grenzen

- Der Ontology Provider wird vor jeder Analyse vollständig validiert.
- Fehlender, nicht ladbarer oder ungültiger Katalog bricht die Analyse ab.
- Runtime-Signale akzeptieren nur registrierte Actor-, Modality-, Target-,
  Expectation-, Relations- und Proposition-Feature-Werte.
- Arbeitshypothesen bleiben als `HYPOTHESIS_ONLY` gekennzeichnet.
- Reflective Constructs besitzen keine Runtime-Signale und keinen Core Score.
- Kompositionen verlangen explizite Evidenz für alle benötigten Constructs.
- Proposition-Reihenfolge, maximaler Abstand und erforderliche Relation werden
  vor Veröffentlichung geprüft.

## Versionierter Laufzeitvektor

- Construct Ontology v0.2
- Rule Contract v0.5
- Foundation / Rule Set v0.4 mit elf Regeln
- Policy Registry v0.7
- Analysis Result v0.1 unverändert
- Analysis Trace v0.1 und v0.2 unverändert
- Resolver Result v0.2 unverändert

Foundation v0.3, Rule v0.4 und Policy Registries bis v0.6 bleiben als
unveränderliche historische Artefakte erhalten.

## Golden Gate

[`sprach-a-lyzer_construct-composition-runtime_v0.1.json`](../../data/golden/sprach-a-lyzer_construct-composition-runtime_v0.1.json)
sichert alle drei Kompositionen, ihre lokalen Construct-Fakten, Proposition
IDs und die Nicht-Scoring-Grenze. Zusätzlich bestehen:

- Composition-Aktivierungs-/Deaktivierungs-Smoke-Test,
- Provider-Fail-Closed-Tests,
- Rule-v0.5- und Policy-v0.7-Driftprüfungen,
- Resolver-Sense-Provenienztest für wiederholte Lexeme,
- vollständige Core-Golden-Parität.

## Nächster Meilensteinschritt

Der fachliche v0.2-Kern ist damit weitgehend geschlossen. Vor `v0.2.0` fehlen
noch die öffentliche Resolver-/Trace-HTTP-Versionierung, ein konsolidiertes
v0.2-Closure-Gate und das versionierte Release Manifest.
