# Spiritual Language Analyzer
## Product Concept v1.1 – Scoring Engine v0.1 & Admin Rule Configuration

**Status:** Fachlicher Entwurf / MVP-Scoring-Grundlage  
**Version:** 1.1  
**Datum:** 18. August 2026  
**Arbeitstitel:** Spiritual Language Analyzer (SLA)

---

# 1. Ziel

Dieses Dokument definiert eine erste **ausbaufähige, transparente und administrierbare Berechnungslogik**.

Grundsatz:

> **Fachliche Entscheidungen gehören in konfigurierbare Regeln – nicht dauerhaft in Programmcode.**

---

# 2. Trennung von Engine und Fachkonfiguration

## Engine
Die Engine weiß **wie** gerechnet wird:
- Bedingungen prüfen
- Modifier anwenden
- Werte begrenzen
- Confidence aggregieren
- Wiederholungen normalisieren
- Dimensionen berechnen

## Fachkonfiguration
Die Fachkonfiguration weiß **was** gilt:
- „unbedingt“ verstärkt `müssen`
- `nicht müssen` verändert den Grundbeitrag
- reale Verpflichtung relativiert Zwang
- `ja, aber` kann vorherige Zustimmung relativieren

---

# 3. Pipeline

```text
INPUT
→ Form-/Phrase-Erkennung
→ Sense Disambiguation
→ Context Resolution
→ Rule Matching
→ Base Contributions
→ Modifier Chain
→ Frequency / Repetition
→ Resonance Layer
→ Confidence
→ Dimension Aggregation
→ WingScore
→ Explanation
```

---

# 4. Contribution

```yaml
Contribution:
  dimension: FREE_WILL
  base_value: -20
  confidence: 0.82
  source: phrase_ich_muss
  evidence_class: B
```

Rohbereich:

```text
-100 bis +100
```

`-100` = starker negativer Pol  
`0` = neutral  
`+100` = starker positiver Pol

---

# 5. Anzeige als Prozentwert

Neutraler Rohwert:

```text
0
```

Transformation:

```text
percentage = (raw_score + 100) / 2
```

Damit:

```text
0 %   = starker negativer Pol
50 %  = neutral / gemischt
100 % = starker positiver Pol
```

---

# 6. Rule Object

```yaml
Rule:
  rule_id
  name
  description
  priority
  enabled
  scope
  condition_tree
  actions
  confidence_modifier
  stop_processing
  version
  status
```

---

# 7. Condition Tree

Unterstützte Operatoren:

```text
AND
OR
NOT
EQUALS
NOT_EQUALS
IN
NOT_IN
GREATER_THAN
LESS_THAN
CONTAINS
MATCHES_PATTERN
WITHIN_TOKENS
BEFORE
AFTER
HAS_RELATION
HAS_PATTERN_CLASS
HAS_SENSE
```

Beispiel:

```yaml
AND:
  - phrase == "ich muss"
  - contains_token == "unbedingt"
  - context != SAFETY
```

---

# 8. Rule Actions

```text
ADD_CONTRIBUTION
MULTIPLY_CONTRIBUTION
REDUCE_CONTRIBUTION
INVERT_CONTRIBUTION
SET_CONFIDENCE
MODIFY_CONFIDENCE
ADD_PATTERN
ADD_EXPLANATION
ADD_REFLECTION_PROMPT
ADD_ALTERNATIVE_CLASS
MARK_NON_ASSESSABLE
STOP_RULE_CHAIN
```

---

# 9. Beispiel: „unbedingt müssen“

```yaml
Rule:
  name: amplify_unbedingt_muessen
  priority: 200
  conditions:
    AND:
      - lexeme == müssen
      - nearby_token == unbedingt
      - sense != SAFETY_NECESSITY
  actions:
    - MULTIPLY_CONTRIBUTION:
        dimension: FREE_WILL
        factor: 1.35
    - ADD_PATTERN:
        class: URGENCY
```

---

# 10. Beispiel: gesetzliche Verpflichtung

```yaml
Rule:
  name: legal_obligation_reduces_free_will_penalty
  priority: 300
  conditions:
    OR:
      - register == LEGAL_ADMINISTRATIVE
      - expectation_source == LAW
      - pattern == EXTERNAL_OBLIGATION
  actions:
    - MULTIPLY_CONTRIBUTION:
        dimension: FREE_WILL
        factor: 0.35
```

---

# 11. Beispiel: „nicht müssen“

```yaml
Rule:
  name: negate_muessen
  priority: 400
  conditions:
    AND:
      - lexeme == müssen
      - negated == true
  actions:
    - INVERT_CONTRIBUTION:
        dimension: FREE_WILL
        factor: 0.6
```

---

# 12. Beispiel: „Ja, aber“

```yaml
Rule:
  name: ja_aber_discounting
  priority: 220
  conditions:
    AND:
      - phrase == "ja, aber"
      - next_proposition_conflicts_with_previous == true
  actions:
    - ADD_CONTRIBUTION:
        dimension: CONNECTION
        value: -12
    - ADD_CONTRIBUTION:
        dimension: OPENNESS
        value: -10
    - ADD_PATTERN:
        class: DISCOUNTING
```

---

# 13. Modifier Chain

Reihenfolge:

```text
1. Base Contribution
2. Sense Modifier
3. Phrase Modifier
4. Negation Modifier
5. Intensifier Modifier
6. Context Modifier
7. Target Modifier
8. Expectation Source Modifier
9. Ambiguity Modifier
10. Resonance Modifier
11. Frequency Modifier
12. Confidence Modifier
```

---

# 14. Prioritäten

Beispiel:

```text
SAFETY_RULE = 900
LEGAL_RULE = 800
NEGATION_RULE = 700
PHRASE_RULE = 600
LEXEME_RULE = 300
DEFAULT_RULE = 100
```

Optional:

```text
stop_processing = true
```

---

# 15. Rule Scopes

```text
GLOBAL
LANGUAGE
REGISTER
DIMENSION
LEXEME
SENSE
PHRASE
PATTERN
COACH_PROFILE
USER_PROFILE
WORKSHOP
```

---

# 16. Rule Profiles

Beispiele:

```text
DEFAULT
COACHING
WORKSHOP
DEEP_REFLECTION
BUSINESS
EDUCATION
RESONANCE
```

Ein Profil steuert:
- aktive Regeln
- Gewichtungen
- sichtbare Dimensionen
- Resonanzstärke
- Erklärungstiefe

---

# 17. Sense Confidence

Skala:

```text
0.00 – 1.00
```

Erste Formel:

```text
sense_confidence =
lexical_match
× context_fit
× ambiguity_factor
```

Beispiel:

```text
0.98 × 0.90 × 0.85 ≈ 0.75
```

---

# 18. Ambiguity Factor

Startwerte:

```text
unambiguous = 1.00
low = 0.90
medium = 0.75
high = 0.55
very_high = 0.35
```

---

# 19. Contribution Confidence

```text
contribution_confidence =
sense_confidence
× rule_confidence
× context_confidence
× evidence_factor
```

---

# 20. Evidence Factor

Startwerte:

```text
A linguistic/corpus = 1.00
B communicative = 0.90
C spiritual-reflective = 0.65
D community hypothesis = 0.50
E personal = 0.90 for this user
```

Diese Faktoren bewerten nicht den menschlichen Wert einer Perspektive, sondern nur deren automatisches Scoring-Gewicht.

---

# 21. Resonance Modes

```text
OFF
HINT_ONLY
MODERATE
FULL
PERSONALIZED
```

Startfaktoren:

```text
OFF          = 0.00
HINT_ONLY    = 0.00 scoring / visible hint
MODERATE     = 0.25
FULL         = 0.50
PERSONALIZED = user-specific
```

---

# 22. Resonance Weight

Perception:

```text
AUDITORY = 1.00
MIXED = 0.90
VISUAL = 0.55
UNKNOWN = 0.50
```

Mental activation:

```text
EXTERNAL_SPEECH = 1.00
READ_ALOUD = 1.00
INNER_SPEECH = 0.80
SILENT_READING = 0.60
UNKNOWN = 0.50
```

Formel:

```text
resonance_effect =
base_resonance
× resonance_mode
× perception_weight
× mental_activation_weight
× evidence_factor
× confidence
```

---

# 23. Frequency Model

Nicht linear.

V0.1:

```text
frequency_factor = 1 + alpha × ln(count)
```

für `count >= 1`.

Default:

```text
alpha = 0.25
```

Ungefähr:

```text
1 → 1.00
2 → 1.17
3 → 1.27
5 → 1.40
10 → 1.58
```

---

# 24. Textlängennormalisierung

```text
normalized_count =
count / max(1, token_count / 100)
```

Lokale rhetorische Wiederholung:

```text
local_repetition_factor = 1.15 – 1.50
```

---

# 25. Dimensionsaggregation

Schritt A:

```text
effective_contribution =
value × confidence
```

Schritt B:

```text
S = Σ effective_contribution
```

Schritt C:

```text
raw_dimension =
100 × tanh(S / scale)
```

Default:

```text
scale = 80
```

Ergebnis:

```text
-100 bis +100
```

---

# 26. Warum tanh

Vorteile:
- annähernd linear nahe 0
- natürliche Sättigung
- verhindert extreme Ausreißer
- lange Texte dominieren weniger

---

# 27. Assessability

```text
if total_confidence_mass < threshold:
    assessable = false
```

Default:

```text
threshold = 0.8
```

---

# 28. Dimension Confidence

V0.1:

```text
dimension_confidence =
1 - Π(1 - contribution_confidence_i)
```

Bei hoher Ambiguität:

```text
dimension_confidence =
min(dimension_confidence, ambiguity_cap)
```

---

# 29. Contribution Trace

Jeder finale Wert bleibt erklärbar.

```yaml
ContributionTrace:
  source_span
  rule_id
  base_value
  modifiers
  final_value
  confidence
  explanation
```

Beispiel:

```text
base: -20
sense modifier: ×1.2
unbedingt: ×1.35
context: ×0.9
confidence: 0.82
effective: -23.9
```

---

# 30. WingScore v0.1

Der WingScore ist **kein einfacher Mittelwert**.

Für jede bewertbare Dimension:

```text
positive_score_i = dimension_percentage_i / 100
```

Gewichtet:

```text
weighted_i =
positive_score_i
× dimension_weight_i
× dimension_confidence_i
```

Basis:

```text
base_wing =
Σ weighted_i
/
Σ(dimension_weight_i × dimension_confidence_i)
```

---

# 31. Weakest-Link Penalty

Starke negative Ausreißer dürfen nicht vollständig weggemittelt werden.

```text
lowest_dimension = min(positive_score_i)
```

```text
penalty =
beta × max(0, threshold - lowest_dimension)
```

Default:

```text
threshold = 0.35
beta = 0.20
```

Final:

```text
wing_score =
100 × clamp(base_wing - penalty, 0, 1)
```

---

# 32. WingScore-Sprache

Nicht:

> „Dein Text ist 72 % gut.“

Sondern:

> **WingScore 72**

> „Der Text zeigt insgesamt eine eher starke Ausprägung der ausgewählten Sprachqualitäten.“

---

# 33. Modifier Types

```text
ADD
MULTIPLY
CAP_MIN
CAP_MAX
SET
INVERT
SUPPRESS
OVERRIDE
```

---

# 34. Kausale Regelketten

Beispiel:

```text
"ich sollte"
→ INTERNALIZED_EXPECTATION

INTERNALIZED_EXPECTATION + "endlich"
→ SELF_PRESSURE

SELF_PRESSURE
→ Freier Wille -15
→ Wertschätzung -10
→ Offenheit -8
```

---

# 35. Schutz vor Rule Explosion

Strategien:

1. generische PatternClasses
2. Rule Templates
3. PhrasePatterns statt Einzelsätze
4. Prioritätsgruppen
5. Deprecation alter Regeln

---

# 36. Rule Templates

```text
INTENSIFIER_TEMPLATE
NEGATION_TEMPLATE
EXTERNAL_OBLIGATION_TEMPLATE
PERSON_LABELING_TEMPLATE
HOMOPHONE_RESONANCE_TEMPLATE
```

---

# 37. Admin-Panel – Hauptbereiche

```text
Dashboard
Knowledge
Dimensions
Rules
Patterns
Relations
Resonance
Parameters
Test Lab
Rule Sets
Sources
Audit
Users & Roles
```

---

# 38. Dimension Manager

Pflegbar:

- positive / negative Labels
- Beschreibungen
- Default Weight
- Sichtbarkeit
- Aktivierung
- Reihenfolge

---

# 39. Rule Builder

Visueller Wenn/Dann-Editor.

Beispiel:

```text
WENN
  Lexem = müssen
UND
  "unbedingt" innerhalb 3 Tokens
UND NICHT
  Kontext = SAFETY

DANN
  Freier Wille × 1.35 Richtung Zwang
  Pattern URGENCY hinzufügen
```

UI:
- Objektart
- Operator
- Wert
- AND/OR-Gruppen
- Aktion
- Dimension
- Faktor
- Confidence
- Priorität
- Stop Processing

---

# 40. Test Lab

Admin gibt einen Satz ein:

> „Ich muss das unbedingt heute schaffen.“

Ausgabe:
- erkannte Senses
- Phrasen
- Regeln
- Contribution Trace
- Dimensionen
- WingScore

Regeländerungen können gegen denselben Satz neu simuliert werden.

---

# 41. Before/After Comparison

```text
Production Rules
vs.
Draft Rules
```

Ausgabe:
- WingScore Δ
- Dimension Δ
- Confidence Δ
- Triggered Rules
- neue Regressionen

---

# 42. Rule Set Versioning

```yaml
RuleSet:
  ruleset_id
  version
  status
  created_by
  approved_by
  changelog
  published_at
```

Status:

```text
DRAFT
TESTING
APPROVED
PRODUCTION
ARCHIVED
```

---

# 43. Publish Workflow

```text
Edit
→ Draft
→ Automated Tests
→ Golden Corpus
→ Review
→ Approve
→ Publish
→ Monitor
```

Keine direkte Live-Änderung.

---

# 44. Rollback

Jede veröffentlichte Version muss rückrollbar sein.

---

# 45. Rule Tests

```yaml
RuleTest:
  input_text
  context
  expected_dimension_range
  expected_patterns
  forbidden_patterns
  expected_rule_trigger
```

---

# 46. Golden Test Corpus

MVP-Empfehlung:

```text
50–100 Sätze
```

Kategorien:
- müssen
- sollen
- dürfen
- ja, aber
- Selbstabwertung
- reale Pflicht
- Gefahr
- klare Grenze
- Homophonie
- Resonanz
- Polysemie
- Ironie
- unklare Fälle

---

# 47. Feature Flags

```text
enable_resonance_scoring
enable_prosody
enable_expectation_source
enable_weakest_link_penalty
```

---

# 48. Audit Log

```text
who
what
old value
new value
when
why
ruleset version
```

---

# 49. Rollen

```text
VIEWER
CONTRIBUTOR
RULE_EDITOR
REVIEWER
PUBLISHER
ADMIN
```

---

# 50. Hard Guardrails

Nicht frei überschreibbar:

- Score bleibt im Wertebereich
- Confidence 0–1
- keine Division durch 0
- keine unendlichen Modifier-Ketten
- keine zirkulären Regeln
- keine automatische Diagnose
- keine semantische Homophon-Vererbung

---

# 51. Soft Guardrails

Beispiel:

Wenn:

```text
Resonance Modifier > 0.8
```

Warnung:

> „Resonanzbeiträge werden damit ähnlich stark wie direkte semantische Beiträge gewichtet.“

---

# 52. Parameter Registry

```yaml
Parameter:
  key
  category
  value
  min
  max
  default
  description
  editable
  requires_approval
```

Pflegbare Parameter:
- base contribution
- rule priority
- modifier factor
- confidence modifier
- ambiguity factor
- evidence factor
- resonance factor
- perception weight
- mental activation weight
- frequency alpha
- aggregation scale
- assessability threshold
- weakest-link threshold
- weakest-link beta
- dimension weights

---

# 53. Scoring Snapshot

Jede Analyse kann reproduzierbar machen:

```text
ruleset_version
parameter_set_version
knowledge_base_version
model_version
```

---

# 54. Rolle der KI

```text
LLM/NLP = Interpretation
Rule Engine = Bewertung
Knowledge Base = Fachwissen
```

KI hilft bei:
- Sense Disambiguation
- Kontext
- Propositionen
- ExpectationSource
- TargetType
- Erklärung
- Alternativen

KI sollte keine freien Punktwerte erfinden.

---

# 55. Unknown Handling

Bei Unsicherheit:

```text
assessable = false
```

oder:

> „Diese Passage ist stark kontextabhängig.“

---

# 56. Konflikt mehrerer Senses

Bei:

```text
Sense A 0.52
Sense B 0.44
```

MVP-Empfehlung:

> Score abschwächen + Ambiguität anzeigen.

---

# 57. Reflection Prompt Rules

Beispiel:

```text
IF INTERNALIZED_EXPECTATION
THEN
"Wessen Erwartung hörst du in dieser Formulierung?"
```

---

# 58. Alternative Rules

Beispiel:

```text
IF müssen + EXTERNAL_OBLIGATION
THEN alternative_type = EXTERNAL_REQUIREMENT
```

Nicht:

```text
replace müssen with möchten
```

---

# 59. MVP-Formelübersicht

```text
base contribution
× sense modifier
× phrase modifier
× negation/intensifier
× context modifier
× ambiguity modifier
× resonance modifier
× repetition modifier
× confidence
→ effective contribution
```

Dann:

```text
Σ contributions
→ tanh aggregation
→ raw dimension -100..+100
→ display 0..100 %
```

Dann:

```text
dimension values
× weights
× confidence
→ WingScore
```

mit optionalem Weakest-Link-Penalty.

---

# 60. Warum nicht alles multiplizieren

## Additiv
für unabhängige Effekte.

## Multiplikativ
für Verstärkung eines bestehenden Effekts.

## Override / Cap
für Sonderkontexte wie Safety.

---

# 61. MVP-Admin-Panel

Für den ersten MVP reichen:

- Dimensionsverwaltung
- Parameterverwaltung
- Rule Builder
- Rule Set Versioning
- Test Lab
- Contribution Trace
- Golden Test Cases
- Publish / Rollback

---

# 62. Noch nicht im ersten Admin-MVP

- komplexes A/B Testing
- vollwertiges Community-Moderationssystem
- Graph-Explorer
- automatische Regelgenerierung
- Prosodie-Editor
- Mehrsprachigkeitsverwaltung

---

# 63. Technische Logik

Mögliche Module:

```text
analysis-service
rule-engine
knowledge-service
scoring-service
admin-api
admin-ui
```

Für MVP können sie zunächst modular in einem Backend leben.

---

# 64. Beispiel Parameter Set v0.1

```yaml
frequency_alpha: 0.25
dimension_aggregation_scale: 80
assessability_threshold: 0.80

evidence_factors:
  A: 1.00
  B: 0.90
  C: 0.65
  D: 0.50
  E: 0.90

resonance:
  OFF: 0.00
  HINT_ONLY: 0.00
  MODERATE: 0.25
  FULL: 0.50

perception:
  AUDITORY: 1.00
  MIXED: 0.90
  VISUAL: 0.55
  UNKNOWN: 0.50

mental_activation:
  EXTERNAL_SPEECH: 1.00
  READ_ALOUD: 1.00
  INNER_SPEECH: 0.80
  SILENT_READING: 0.60
  UNKNOWN: 0.50

wing_score:
  weakest_link_threshold: 0.35
  weakest_link_beta: 0.20
```

Alle Werte sind **Startwerte für Tests**, keine endgültigen Wahrheiten.

---

# 65. Validierungsstrategie

1. Expertenschätzung als Start
2. Golden Test Corpus
3. Workshop-Feedback
4. Coach-Feedback
5. skeptische Testgruppe
6. statistische Analyse
7. Anpassung
8. Versionierung

---

# 66. Vermeidungsstrategien

## Zu viele Regeln
- Templates
- PatternClasses
- Deprecation
- Regelbibliothek

## Unverständliche Scores
- Contribution Trace
- Explainability
- wenige Modifier pro Treffer

## Übergewichtete Resonanz
- separater Modus
- Caps
- Evidenzfaktoren
- Warnungen

## Riskante Admin-Änderung
- Draft/Test/Publish
- Approval
- Rollback
- Audit

## Lange Texte dominieren
- Normalisierung
- logarithmische Frequenz
- tanh-Sättigung

---

# 67. Nächster sinnvoller Schritt

Jetzt sollte ein **Golden Test Corpus v0.1** mit zunächst 20–30 gezielten Sätzen entstehen.

Damit können wir:
- Regeln prüfen
- Startparameter kalibrieren
- Contribution Traces bewerten
- erste Dimensionswerte vergleichen
- WingScore praktisch beurteilen

---

# 68. Definition of Done – Scoring Engine v0.1 Konzept

- [x] Engine/Fachkonfiguration getrennt
- [x] Contribution-Modell
- [x] Rohwertbereich
- [x] Rule Object
- [x] Condition Tree
- [x] Rule Actions
- [x] Modifier-Reihenfolge
- [x] Prioritäten
- [x] Rule Profiles
- [x] Sense Confidence
- [x] Evidenzfaktoren
- [x] Resonance Modifier
- [x] Frequency-Funktion
- [x] Aggregationsfunktion
- [x] Dimension Confidence
- [x] WingScore v0.1
- [x] Weakest-Link-Penalty
- [x] Contribution Trace
- [x] Admin Rule Builder
- [x] Test Lab
- [x] Rule Set Versionierung
- [x] Rollback
- [x] Golden Tests vorgesehen
- [x] Audit
- [x] Guardrails
- [ ] Golden Test Corpus v0.1
- [ ] Beispielrechnungen
- [ ] Startparameter kalibrieren
- [ ] Admin-UI-Wireframe
- [ ] technische Implementierung spezifizieren

---

# 69. Leitgedanke

> **Die Engine soll rechnen.  
> Das Fachmodell soll entscheiden, warum.  
> Das Admin-Panel soll ermöglichen, dieses Warum kontrolliert weiterzuentwickeln.**
