# Sprachkompass / Spiritual Language Analyzer
## Developer Handoff v0.1 – für Cody

**Status:** Implementierungsübergabe  
**Datum:** 19. August 2026  
**Fachstand:** Reference Engine v0.4.1  
**Ziel:** Aus dem validierten Fachmodell einen ersten implementierbaren MVP bauen.

---

# 1. Executive Summary

Der Sprachkompass ist **kein Wörterbuch mit Positiv-/Negativwerten**.

Er ist eine erklärbare Analyse-Engine für sprachliche Muster.

Kernpipeline:

```text
Input
→ Normalisierung
→ Token/Phrase Matching
→ Proposition Graph
→ SenseCandidate Resolver
→ Context / Target / Expectation
→ Pattern Detection
→ Rule Engine
→ Contributions
→ Assessability
→ Dimension Aggregation
→ WingScore / Corporate Metric
→ Explanation
→ Reflection Prompt
→ Alternatives
```

Die Engine bewertet **Texte und sprachliche Muster, nicht Menschen**.

---

# 2. Sechs Kerndimensionen

Canonical IDs:

```text
AGENCY
CONNECTION
APPRECIATION
CLARITY
FREE_WILL
OPENNESS
```

Fachliche Achsen:

```text
Ohnmacht     ↔ Wirksamkeit
Trennung     ↔ Verbindung
Abwertung    ↔ Wertschätzung
Unklarheit   ↔ Klarheit
Zwang        ↔ Freier Wille
Begrenzung   ↔ Offenheit
```

Intern bipolar.

Standardanzeige kann nur den positiven Pol zeigen.

---

# 3. Wichtigste Architekturentscheidung

## Eine fachliche Engine

Es existiert **keine Private Engine** und **keine Corporate Engine**.

Es existieren:

```text
Canonical Knowledge
Canonical Rules
Canonical Scoring
```

und darauf:

```text
Presentation Profiles
```

---

# 4. Betriebsdomains

Beispiel:

```text
private.example.tld
corporate.example.tld
```

Jede Domain lädt nur ihr eigenes Presentation Bundle.

## Private

```yaml
profile: PRIVATE
wing_score_visible: true
resonance_ui: available
spiritual_reflection_layer: available
```

## Corporate

```yaml
profile: CORPORATE
wing_score_visible: false_as_label
canonical_metric: WING_SCORE
display_metric: REFLECTION_INDEX
resonance_default: HINT_ONLY
private_bundle_shipped: false
```

WICHTIG:

> Das Corporate Frontend darf niemals auf interne Canonical Labels zurückfallen.

Fallback gehört in das Corporate Bundle selbst.

---

# 5. Presentation Mapping

Canonical concept:

```text
RESONANCE
```

Private:

```text
Resonanz
```

Corporate:

```text
Wahrnehmungswirkung
```

Canonical:

```text
WING_SCORE
```

Private:

```text
WingScore
```

Corporate:

```text
Wirkungsprofil
```

Mapping verändert niemals:

- Rule
- Contribution
- Relation
- Score
- Evidence
- Confidence

---

# 6. Mapping Bundle Contract

```json
{
  "profile": "CORPORATE",
  "locale": "de-DE",
  "version": "1.0.0",
  "labels": {
    "METRIC_WING_SCORE": "Wirkungsprofil",
    "RESONANCE": "Wahrnehmungswirkung",
    "SPIRITUAL_DEVELOPMENT": "persönliche Entwicklung",
    "CONSCIOUSNESS": "Reflexionsfähigkeit"
  },
  "fallbacks": {
    "METRIC_WING_SCORE": "Wirkungsprofil",
    "UNKNOWN_METRIC": "Reflexionswert"
  }
}
```

Frontend-Regel:

```text
render(canonicalKey):
    profileBundle[key]
    ?? profileBundle.fallbacks[key]
    ?? genericCorporateFallback
```

NIE:

```text
?? canonicalKey
```

---

# 7. Input Modes

```text
TEXT
SPOKEN_DICTATION
DIRECT_AUDIO
```

## TEXT
Normaler Text.

## SPOKEN_DICTATION
Text stammt aus System-/Telefon-Diktat.

## DIRECT_AUDIO
Später: Audio + Transkript + Prosodie.

---

# 8. Spoken Input

Spoken Dictation darf keinen pauschalen Negativfaktor besitzen.

Nicht:

```text
spoken_score = text_score * 1.5
```

Sondern:

```text
spoken language
→ more observable patterns
→ more contributions
→ potentially stronger result
```

Spoken Features:

```yaml
SpokenFeatures:
  filler_count
  discourse_particle_count
  repetition_factor
  repeated_tokens
  self_correction_count
  fragment_count
  modal_density
  generalization_density
```

---

# 9. Epistemische Schichten

Jede fachliche Aussage besitzt Provenienz.

```text
A = linguistic/corpus
B = communicative interpretation
C = spiritual-reflective hypothesis
D = community hypothesis
E = personal association
```

Wichtig:

Homophonie ist linguistisch beschreibbar.

Eine behauptete unbewusste/spirituelle Wirkung ist **keine linguistisch etablierte Tatsache** und wird separat gespeichert.

---

# 10. Core Entities

Persistiert:

```text
Lexeme
WordForm
Pronunciation
Sense
Phrase
PhrasePattern
Relation
RelationClaim
Collocation
UsageProfile
ResonanceProfile
AmbiguityProfile
PatternClass
Dimension
DimensionContribution
Alternative
AlternativeQuality
Source
Rule
RuleSet
ParameterSet
PresentationBundle
GoldenTestCase
```

Runtime:

```text
AnalysisContext
Proposition
PropositionEdge
SenseCandidate
AnalysisHit
ContributionTrace
AssessabilityResult
DimensionResult
AnalysisResult
SpokenFeatures
```

---

# 11. PostgreSQL Tabellen – MVP

Empfohlener Start:

```text
lexemes
word_forms
pronunciations
senses
phrases
phrase_patterns
relations
relation_claims
pattern_classes
dimensions
dimension_contributions
rules
rule_sets
parameters
parameter_sets
presentation_bundles
presentation_entries
golden_test_cases
sources
```

Analyseergebnisse optional:

```text
analyses
analysis_dimension_results
analysis_traces
```

Für Privacy-MVP kann Analysepersistenz zunächst deaktiviert sein.

---

# 12. Lexeme

```sql
lexemes (
  id uuid primary key,
  language varchar(10) not null,
  lemma text not null,
  part_of_speech text,
  status text not null,
  version integer not null,
  created_at timestamptz,
  updated_at timestamptz
);
```

---

# 13. Sense

```sql
senses (
  id uuid primary key,
  lexeme_id uuid references lexemes(id),
  sense_key text not null,
  title text not null,
  description text,
  register text,
  domain text,
  evidence_class char(1),
  status text,
  version integer
);
```

---

# 14. Relation

```sql
relations (
  id uuid primary key,
  relation_type text not null,
  source_type text not null,
  source_id uuid not null,
  target_type text not null,
  target_id uuid not null,
  directionality text not null,
  strength numeric,
  confidence numeric,
  evidence_class char(1),
  active boolean default true,
  version integer
);
```

Keine FK für polymorphe source/target IDs im ersten MVP; Integrität auf Service-Ebene prüfen.

---

# 15. RelationClaim

```sql
relation_claims (
  id uuid primary key,
  relation_id uuid references relations(id),
  claim_type text not null,
  statement text not null,
  evidence_class char(1),
  confidence numeric,
  status text,
  version integer
);
```

Claim Types:

```text
LINGUISTIC_FACT
COMMUNICATIVE_EFFECT
RESONANCE_HYPOTHESIS
COACHING_OBSERVATION
COMMUNITY_OBSERVATION
PERSONAL_ASSOCIATION
```

---

# 16. Rules

```sql
rules (
  id uuid primary key,
  rule_key text unique not null,
  name text not null,
  description text,
  priority integer not null,
  enabled boolean default true,
  scope text,
  condition_tree jsonb not null,
  actions jsonb not null,
  confidence_modifier numeric default 1,
  stop_processing boolean default false,
  status text,
  version integer
);
```

---

# 17. Rule Set

```sql
rule_sets (
  id uuid primary key,
  version text unique not null,
  status text not null,
  changelog text,
  created_by uuid,
  approved_by uuid,
  published_at timestamptz
);
```

Join:

```text
rule_set_rules(rule_set_id, rule_id)
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

# 18. Parameter Registry

```sql
parameters (
  key text primary key,
  category text,
  default_value jsonb,
  min_value jsonb,
  max_value jsonb,
  description text,
  editable boolean,
  requires_approval boolean
);
```

ParameterSet:

```text
parameter_sets
parameter_set_values
```

---

# 19. Condition Tree JSON

Beispiel:

```json
{
  "op": "AND",
  "children": [
    {
      "field": "selected_sense",
      "operator": "EQUALS",
      "value": "INTERNAL_PRESSURE"
    },
    {
      "field": "nearby_tokens",
      "operator": "CONTAINS",
      "value": "unbedingt"
    },
    {
      "op": "NOT",
      "child": {
        "field": "context",
        "operator": "EQUALS",
        "value": "SAFETY"
      }
    }
  ]
}
```

---

# 20. Rule Action JSON

```json
[
  {
    "type": "ADD_CONTRIBUTION",
    "dimension": "FREE_WILL",
    "value": -8
  },
  {
    "type": "ADD_PATTERN",
    "pattern": "URGENCY"
  }
]
```

Supported:

```text
ADD_CONTRIBUTION
MULTIPLY_CONTRIBUTION
CAP_MIN
CAP_MAX
SET
INVERT
SUPPRESS
ADD_PATTERN
ADD_EXPLANATION
ADD_REFLECTION_PROMPT
MARK_NON_ASSESSABLE
STOP_RULE_CHAIN
```

---

# 21. Proposition Graph

Runtime representation:

```json
{
  "nodes": [
    {
      "id": "P0",
      "text": "Ich verstehe, dass dir das wichtig ist",
      "features": {
        "actor": "SELF",
        "predicate": true,
        "target": true
      }
    },
    {
      "id": "P1",
      "text": "Für mich kommt diese Lösung nicht infrage",
      "features": {
        "actor": "SELF",
        "boundary": true
      }
    }
  ],
  "edges": [
    {
      "source": "P0",
      "target": "P1",
      "relation": "CONCESSION"
    }
  ]
}
```

MVP Relations:

```text
CONTRAST
CONCESSION
CAUSE
CONSEQUENCE
ADDITION
CONDITION
CORRECTION
DISCOUNTING
```

---

# 22. SenseCandidate

```json
{
  "sense_id": "INTERNALIZED_EXPECTATION",
  "features": {
    "lexical_match": 1.0,
    "phrase_fit": 0.9,
    "syntax_fit": 0.8,
    "domain_fit": 0.5,
    "register_fit": 0.6,
    "discourse_fit": 0.7,
    "collocation_fit": 0.9,
    "context_fit": 0.8
  },
  "score": 0.81
}
```

Top1 / Top2:

```text
gap >= 0.12 and top >= 0.72 → HIGH
gap >= 0.07 and top >= 0.58 → MEDIUM
otherwise → AMBIGUOUS
```

Startwerte sind Parameter.

---

# 23. TargetType

```text
PERSON
SELF
BEHAVIOR
EVENT
OBJECT
PROCESS
IDEA
GROUP
INSTITUTION
UNKNOWN
```

Beispiele:

```text
"Du bist das Problem." → PERSON
"Die Vereinbarung wurde nicht eingehalten." → BEHAVIOR
"technisches Problem" → PROCESS
```

---

# 24. ExpectationSource

```text
SELF
OTHER_PERSON
GROUP
INSTITUTION
LAW
CULTURE
UNSPECIFIED
INTERNALIZED
```

---

# 25. Pattern Composition

Nicht nur einzelne Marker.

Beispiel:

```text
ACKNOWLEDGEMENT
+ CLEAR_BOUNDARY
→ RESPECTFUL_BOUNDARY
```

```text
REALISTIC_LIMIT
+ CHOICE_LANGUAGE
→ AGENCY_RECOVERY
```

```text
ERROR_DESCRIPTION
+ LEARNING_FRAME
→ LEARNING_RECOVERY
```

Composed Pattern darf Komponenten teilweise suppressen, damit kein Double Counting entsteht.

---

# 26. Contribution

Intern:

```json
{
  "dimension": "FREE_WILL",
  "value": -20,
  "confidence": 0.84,
  "source_type": "PATTERN",
  "source_key": "INTERNAL_PRESSURE",
  "evidence_class": "B"
}
```

Wertebereich fachlich:

```text
-100 ... +100
```

Einzelne MVP Base Contributions typischerweise:

```text
-50 ... +50
```

---

# 27. Aggregation

```text
effective_i = value_i × confidence_i

S = Σ effective_i

raw =
100 × tanh(S / aggregation_scale)

display =
(raw + 100) / 2
```

Start:

```text
aggregation_scale = 80
```

---

# 28. Assessability

Fehlende Evidenz ist nicht 50 %.

Runtime states:

```text
NOT_ASSESSABLE
WEAK
ASSESSABLE
STRONG
```

Startlogik:

```text
evidence_mass
independent_hits
pattern_diversity
context_confidence
sense_confidence
proposition_coverage
ambiguity
```

Standard-UI:

```text
NOT_ASSESSABLE → keine Zahl
WEAK           → Tendenz
ASSESSABLE     → Zahl
STRONG         → Zahl
```

---

# 29. WingScore

Nur wenn mindestens drei Dimensionen:

```text
ASSESSABLE or STRONG
```

und ausreichende Confidence besitzen.

Nicht:

```text
simple average
```

Start:

```text
confidence weighted mean
+ cautious weakest-link penalty
```

WingScore bewertet **den analysierten Text**, niemals den Menschen.

---

# 30. Resonance Layer

Default:

```text
PRIVATE = HINT_ONLY or user selectable
CORPORATE = HINT_ONLY
```

Modes:

```text
OFF
HINT_ONLY
MODERATE
FULL
PERSONALIZED
```

Wichtig:

Resonanz allein macht im Standard keine Kerndimension assessable.

---

# 31. Homophonie

Beispiel:

```text
hast ↔ hasst
```

Canonical:

```text
WordForm(hast)
--HOMOPHONE-->
WordForm(hasst)
```

Keine semantische Vererbung.

Nicht:

```text
hast.score += hassen.score
```

Optional:

```text
RelationClaim(RESONANCE_HYPOTHESIS)
```

---

# 32. API – Analyze

```http
POST /api/v1/analyze
```

Request:

```json
{
  "text": "Ich sollte eigentlich längst weiter sein.",
  "locale": "de-DE",
  "context": "SELF_TALK",
  "input_mode": "TEXT",
  "presentation_profile": "PRIVATE",
  "analysis_mode": "STANDARD"
}
```

---

# 33. Analyze Response

```json
{
  "analysis_id": "optional",
  "versions": {
    "knowledge_base": "0.1",
    "rule_set": "0.1",
    "parameter_set": "0.1",
    "engine": "0.4.1"
  },
  "patterns": [
    {
      "key": "INTERNALIZED_EXPECTATION",
      "span": [0, 10],
      "confidence": 0.84
    }
  ],
  "dimensions": {
    "AGENCY": {
      "assessability": "ASSESSABLE",
      "score": 44.6,
      "confidence": 0.76
    },
    "FREE_WILL": {
      "assessability": "ASSESSABLE",
      "score": 39.4,
      "confidence": 0.84
    }
  },
  "metric": {
    "key": "WING_SCORE",
    "value": null,
    "reason": "INSUFFICIENT_DIMENSIONS"
  },
  "reflection_prompts": [
    "Wessen Erwartung hörst du in dieser Formulierung?"
  ],
  "alternatives": [],
  "trace_available": true
}
```

---

# 34. API – Trace

```http
GET /api/v1/analyses/{id}/trace
```

Admin / Fachmodus.

Response:

```json
{
  "contributions": [
    {
      "source": "INTERNALIZED_EXPECTATION",
      "dimension": "FREE_WILL",
      "base": -20,
      "confidence": 0.84,
      "effective": -16.8
    }
  ]
}
```

---

# 35. API – Admin Rules

```text
GET    /api/v1/admin/rules
POST   /api/v1/admin/rules
PATCH  /api/v1/admin/rules/{id}
POST   /api/v1/admin/rule-sets
POST   /api/v1/admin/rule-sets/{id}/test
POST   /api/v1/admin/rule-sets/{id}/publish
POST   /api/v1/admin/rule-sets/{id}/rollback
```

---

# 36. API – Test Lab

```http
POST /api/v1/admin/test-lab/analyze
```

Request kann explizit Versionen wählen:

```json
{
  "text": "...",
  "rule_set": "draft-17",
  "parameter_set": "calibration-9",
  "profile": "PRIVATE"
}
```

---

# 37. Golden Tests

Golden Tests müssen automatisiert in CI laufen.

Beispiel:

```yaml
id: G03
text: "Du musst sofort das Gebäude verlassen!"
context: SAFETY
expected:
  patterns:
    - SAFETY_DIRECTIVE
  dimensions:
    CLARITY:
      min: 85
    AGENCY:
      min: 70
  forbidden:
    - strong_free_will_penalty
```

---

# 38. Regression Gate

Publish darf blockiert werden, wenn:

```text
previous PASS → new GAP
```

ohne explizite fachliche Freigabe.

MVP-Regel:

```text
zero unapproved golden regressions
```

---

# 39. Domain Bundle Tests

Corporate CI Test:

```text
bundle must not contain:
WingScore
Resonanz
Energie
spirituell
```

Ausnahmen nur, wenn bewusst in transparenter Querverlinkungstext-Datei freigegeben.

Noch besser:

private bundle wird gar nicht in Corporate Build Context kopiert.

---

# 40. Frontend-MVP

## Analyze Screen

```text
[ Textarea ]

[ Text schreiben ] [ Gedanken diktieren ]

Kontext:
[ Persönlich ] / Corporate Domain automatisch Berufsleben

[ Analysieren ]
```

Result:

```text
markierter Text
6 Dimensionskarten
WingScore / Corporate Metric
Top 3 Hinweise
1 Reflexionsfrage
2–4 Alternativen
```

---

# 41. Mobile Spoken UX

CTA:

> Sprich einfach, wie du wirklich sprichst.

Subtext:

> Du musst deinen Gedanken nicht erst schön formulieren. Spontane Sprache kann andere Muster sichtbar machen als sorgfältig geschriebener Text.

System-Diktat MVP:

```text
SPOKEN_DICTATION
```

Audio später:

```text
DIRECT_AUDIO
```

---

# 42. Privacy Defaults

MVP:

```text
analysis_storage = OFF by default
raw_audio_storage = OFF
personal_history = explicit opt-in
```

Corporate:

```text
manager_access_to_individual_analysis = NEVER
employee_ranking = NEVER
HR_selection_use = NEVER
```

---

# 43. Hard Guardrails – Code

Nicht im Admin überschreibbar:

```text
confidence ∈ [0,1]
score bounded
no circular rule chain
no semantic homophone inheritance
no person diagnosis
no employee ranking
missing evidence != 50
```

---

# 44. Soft Guardrails – Admin Warning

Beispiele:

```text
resonance factor > 0.6
base contribution abs > 50
frequency alpha > 0.5
```

Warnung + Review Required.

---

# 45. Recommended Backend Stack

Kein Muss, aber passend:

```text
Go
PostgreSQL
REST API
JSONB Rule Trees
```

Admin UI:

```text
React / Vue / similar
```

MVP zunächst modularer Monolith.

Nicht sofort Microservices bauen.

Module:

```text
analysis
knowledge
resolver
rules
scoring
presentation
admin
```

---

# 46. Suggested Go Packages

```text
/internal/analysis
/internal/knowledge
/internal/resolver
/internal/rules
/internal/scoring
/internal/assessability
/internal/presentation
/internal/admin
/internal/golden
```

---

# 47. Core Interfaces – Go

```go
type Resolver interface {
    Resolve(ctx context.Context, input AnalysisInput) (ResolverResult, error)
}

type RuleEngine interface {
    Evaluate(ctx context.Context, in RuleContext) ([]RuleEffect, error)
}

type Scorer interface {
    Score(ctx context.Context, in ScoringInput) (ScoringResult, error)
}

type PresentationMapper interface {
    Map(profile string, locale string, result AnalysisResult) PresentedResult
}
```

---

# 48. AnalysisInput

```go
type AnalysisInput struct {
    Text                string
    Locale              string
    Context             string
    InputMode           string
    AnalysisMode        string
    PresentationProfile string
}
```

---

# 49. MVP Milestones

## M0 – Skeleton

- Go project
- PostgreSQL migrations
- health endpoint
- configuration
- unit test foundation

## M1 – Knowledge + Rules

- dimensions
- pattern classes
- rules
- parameters
- seed data

## M2 – Basic Analyzer

- token/phrase matching
- first senses
- contributions
- aggregation
- assessability

## M3 – Proposition + Resolver

- proposition graph
- sense candidates
- target type
- expectation source

## M4 – API + UI

- analyze endpoint
- result UI
- trace
- presentation bundles

## M5 – Admin/Test Lab

- rule editor
- parameters
- golden tests
- draft/publish

## M6 – Spoken Dictation

- input mode
- spoken features
- pair tests

---

# 50. MVP Seed Words

Start:

```text
müssen
sollen
dürfen
aber
eigentlich
Problem
Schuld
Fehler
frei
versuchen
```

Plus:

```text
immer
nie
keine Wahl
ich kann nicht
du bist das Problem
```

---

# 51. First Acceptance Scenarios

## A

Input:

> Ich muss das heute unbedingt noch schaffen.

Expected:

```text
INTERNAL_PRESSURE
URGENCY
FREE_WILL below neutral
OPENNESS below neutral
```

## B

> Du musst sofort das Gebäude verlassen!

Context:

```text
SAFETY
```

Expected:

```text
SAFETY_DIRECTIVE
high CLARITY
no strong coercion penalty
```

## C

> Der Eintritt ist frei.

Expected:

```text
frei → FREE_OF_CHARGE
no FREE_WILL bonus
```

## D

> Er soll sehr erfolgreich sein.

Expected:

```text
REPORTED_CLAIM
no normative penalty
```

## E

> Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.

Expected:

```text
RESPECTFUL_BOUNDARY
CONNECTION high
APPRECIATION high
CLARITY high
FREE_WILL high
```

## F

> Hast du Geld?

Expected:

```text
HOMOPHONE(hast,hasst)
semantic score unaffected
optional resonance hint
```

---

# 52. Out of Scope for first coding sprint

Nicht sofort:

- Direct Audio
- Prosody
- full community wiki
- reputation system
- browser extension
- website analysis
- long-term personal trends
- multi-language
- automatic rule generation
- exhaustive German NLP

---

# 53. What Cody should build first

Empfohlene Reihenfolge:

```text
1. Domain models
2. Rule representation
3. parameter registry
4. deterministic analysis pipeline
5. golden test harness
6. analyze REST endpoint
7. minimal result UI
8. only then admin UI
```

Warum?

Ohne Golden Harness wird jede spätere fachliche Änderung riskant.

---

# 54. Golden Harness is mandatory

Command idea:

```bash
go test ./internal/golden/...
```

oder:

```bash
sprachkompass golden test --ruleset draft
```

Output:

```text
94 expectations
38 pass
21 missing
27 too low
8 too high
0 over-assessed
0 regressions
```

Die konkrete aktuelle Referenzzahl stammt aus v0.4.1 und ist kein Produktionsziel.

---

# 55. Definition of Done – Coding Foundation

- [ ] Repository bootstrapped
- [ ] migrations run
- [ ] canonical dimension IDs
- [ ] presentation profile loader
- [ ] Corporate bundle has embedded fallback
- [ ] rule JSON schema
- [ ] parameter registry
- [ ] seed import
- [ ] analysis input contract
- [ ] contribution type
- [ ] aggregation
- [ ] assessability
- [ ] analyze endpoint
- [ ] trace endpoint
- [ ] golden harness
- [ ] six acceptance scenarios pass
- [ ] private/corporate bundle separation test
- [ ] no raw canonical fallback in Corporate UI

---

# 56. Fachlicher North Star

> **Die Engine soll lieber transparent unsicher sein als präzise falsch.**

---

# 57. Produkt-North-Star

> **Sprache sichtbar machen. Reflexion ermöglichen. Entwicklung unterstützen.**

---

# 58. Implementierungsleitgedanke

> **Komplexität gehört in Engine, Wissen und Regeln. Die Nutzererfahrung soll leicht, verständlich und freiwillig bleiben.**
