# Spiritual Language Analyzer
## Product Concept v1.0 – Konsolidiertes Datenmodell

**Status:** Konsolidierter MVP-Fachstand  
**Version:** 1.0  
**Datum:** 18. August 2026  
**Arbeitstitel:** Spiritual Language Analyzer (SLA)

---

# 1. Ziel dieses Dokuments

Dieses Dokument konsolidiert alle bisher entwickelten fachlichen Datenstrukturen zu einem stabilen **Datenmodell v1.0**.

Es bildet die Grundlage für:

- Seed-Daten,
- Analysepipeline,
- Scoring-Logik,
- Confidence-Modell,
- WingScore,
- Browser-Integration,
- Workshop-Modus,
- Community-Erweiterungen.

Das Modell soll:

1. sprachlich differenziert genug sein,
2. technisch implementierbar bleiben,
3. später erweiterbar sein,
4. Bedeutungs- und Resonanzebenen sauber trennen,
5. möglichst viele künftige Erweiterungen über Daten statt über neue Kernobjekte ermöglichen.

---

# 2. Grundprinzip

Die Plattform speichert nicht einfach Wörter mit positiven oder negativen Werten.

Sie speichert ein **vernetztes Sprachwissensmodell**.

Zentrale Struktur:

```text
LEXEME
  ├── WORD_FORM
  ├── PRONUNCIATION
  ├── SENSE
  ├── PHRASE
  ├── RELATION
  ├── RELATION_CLAIM
  ├── COLLOCATION
  ├── USAGE_PROFILE
  ├── RESONANCE_PROFILE
  ├── AMBIGUITY_PROFILE
  ├── PATTERN_CLASS
  └── DIMENSION_CONTRIBUTION
```

Zur Analysezeit entstehen zusätzliche Laufzeitobjekte.

---

# 3. Vier Datenebenen

## 3.1 Stammdaten

Stabile redaktionell gepflegte Objekte:

- Lexem
- Wortform
- Aussprache
- Sense
- Phrase
- PatternClass
- Dimension

---

## 3.2 Relationsdaten

Verbindungen zwischen Objekten:

- Homophonie
- Synonymie
- Kollokation
- Diskursrelation
- Resonanz
- Dimensionswirkung

---

## 3.3 Nutzungs- und Evidenzdaten

Kontextualisierung:

- Register
- Häufigkeit
- Medium
- Wahrnehmungskanal
- mentale Aktivierung
- Quellen
- Evidenzklassen
- Confidence

---

## 3.4 Laufzeitdaten

Entstehen erst bei konkreter Analyse:

- erkannter Sense
- Kontext
- TargetType
- ExpectationSource
- Proposition
- AnalysisHit
- Dimensionsbeitrag
- Resonanzrelevanz
- AlternativeQuality
- finaler Dimensionswert

---

# 4. Kernobjekt: Lexeme

```yaml
Lexeme:
  lexeme_id
  lemma
  language
  part_of_speech
  subtype
  grammatical_features
  canonical_form
  editorial_status
  active
  version
  created_at
  updated_at
```

Beispiel:

```yaml
lemma: müssen
part_of_speech: VERB
subtype: MODALVERB
```

---

# 5. WordForm

```yaml
WordForm:
  form_id
  lexeme_id
  surface_form
  grammatical_features
  register
  frequency_hint
  editorial_status
```

Beispiele:

```text
muss
musst
müssen
musste
gemusst
```

Relations können direkt auf Wortformen zeigen.

Das ist insbesondere für Homophonie wichtig.

---

# 6. Pronunciation

```yaml
Pronunciation:
  pronunciation_id
  owner_node_id
  phonetic_form
  ipa
  locale
  dialect
  stress_pattern
  confidence
  source_ids
```

`owner_node_id` kann auf:

- Lexeme
- WordForm
- Phrase

zeigen.

---

# 7. Sense

```yaml
Sense:
  sense_id
  lexeme_id
  sense_key
  title
  short_description
  detailed_description
  domain
  register
  usage_notes
  editorial_status
  evidence_class
  version
```

Ein Sense besitzt eigene:

- Dimensionsbeiträge,
- Kontextregeln,
- Kollokationen,
- Relations.

---

# 8. Phrase

```yaml
Phrase:
  phrase_id
  canonical_text
  language
  phrase_type
  register
  short_description
  editorial_status
  version
```

Beispiele:

```text
ja, aber
ich muss
du musst
nicht müssen
keine Wahl
```

---

# 9. PhrasePattern

```yaml
PhrasePattern:
  pattern_id
  pattern_expression
  slots
  order_constraints
  token_distance
  register
  pragmatic_function
  active
```

Beispiel:

```text
[PRONOUN_SELF] + müssen
```

---

# 10. Proposition

`Proposition` ist primär ein Laufzeitobjekt.

Es bildet bedeutungstragende Teilaussagen ab.

```yaml
Proposition:
  proposition_id
  source_span
  normalized_content
  predicate
  arguments
  target_type
  confidence
```

Beispiel:

> „Ich verstehe dich, aber du liegst falsch.“

wird zerlegt in zwei Propositionen.

---

# 11. Relation

```yaml
Relation:
  relation_id
  relation_type
  source_node_id
  target_node_id
  directionality
  strength
  confidence
  locale
  register
  perception_channel
  mental_activation
  context_conditions
  active
  editorial_status
  version
```

---

# 12. Directionality

```text
SYMMETRIC
DIRECTED
```

Beispiele:

```text
HOMOPHONE → meist SYMMETRIC
HYPERNYM → DIRECTED
AMPLIFIES → DIRECTED
```

---

# 13. Relation Taxonomy v1.0

## Form / Orthografie

```text
SAME_LEMMA_FORM
ORTHOGRAPHIC_VARIANT
HOMOGRAPH
COMPOUND_COMPONENT
TYPOGRAPHIC_NEIGHBOR
```

## Klang

```text
HOMOPHONE
PHONETIC_NEIGHBOR
PRONUNCIATION_VARIANT
HETERONYM_READING
PHRASE_HOMOPHONE
SOUND_RESONANCE
```

## Bedeutung

```text
HAS_SENSE
POLYSEMY_RELATED_SENSE
HOMONYM
SYNONYM
ANTONYM
HYPERNYM
HYPONYM
SEMANTIC_ASSOCIATION
METAPHOR_RELATION
```

## Phrase / Gebrauch

```text
PART_OF_PHRASE
COLLOCATES_WITH
IDIOM_COMPONENT
AMPLIFIES
REDUCES
NEGATES
SHIFTS_MEANING
REQUIRES_CONTEXT
```

## Diskurs

```text
DISCOURSE_CONTRAST
DISCOURSE_CONCESSION
DISCOURSE_CORRECTION
DISCOURSE_CAUSE
DISCOURSE_CONSEQUENCE
DISCOUNTS_PREVIOUS_PROPOSITION
```

## Dimension

```text
AFFECTS_DIMENSION
MODIFIES_DIMENSION_EFFECT
```

## Reflexion / Resonanz

```text
SPIRITUAL_RESONANCE
PERSONAL_RESONANCE
COMMUNITY_RESONANCE
COACHING_ASSOCIATION
CULTURAL_ASSOCIATION
```

---

# 14. RelationClaim

Eine Relation kann mehrere Aussagen mit unterschiedlicher Evidenz besitzen.

```yaml
RelationClaim:
  claim_id
  relation_id
  claim_type
  statement
  evidence_class
  confidence
  source_ids
  context_conditions
  editorial_status
  version
```

`claim_type`:

```text
LINGUISTIC_FACT
COMMUNICATIVE_EFFECT
RESONANCE_HYPOTHESIS
COACHING_OBSERVATION
COMMUNITY_OBSERVATION
PERSONAL_ASSOCIATION
```

---

# 15. Evidenzklassen

```text
A = linguistisch / korpusbasiert
B = kommunikativ / psychologisch plausibilisiert
C = spirituell-reflexiv
D = Community-Hypothese
E = persönlich / nutzerspezifisch
```

Evidenz beschreibt die Herkunft einer Aussage, nicht ihren menschlichen Wert.

---

# 16. Confidence

```text
0.00 – 1.00
```

Confidence ist getrennt von Wirkung bzw. Stärke.

Beispiel:

```text
strength = 0.8
confidence = 0.4
```

= potenziell starke, aber unsichere Wirkung.

---

# 17. PerceptionChannel

```text
AUDITORY
VISUAL
MIXED
UNKNOWN
```

---

# 18. MentalActivation

```text
EXTERNAL_SPEECH
SILENT_READING
READ_ALOUD
INNER_SPEECH
DICTATED
UNKNOWN
```

Diese Ebene ist insbesondere für Klangrelationen relevant.

---

# 19. Register

MVP-Taxonomie:

```text
PRIVATE_CONVERSATION
FAMILY
FRIENDS
WORKPLACE
LEADERSHIP
COACHING
WORKSHOP
PUBLIC_SPEECH
MODERATION
EDUCATION
SOCIAL_MEDIA
MESSAGING
EMAIL
JOURNALISM
LITERATURE
ACADEMIC
LEGAL_ADMINISTRATIVE
ADVERTISING
WEBSITE
SELF_TALK
UNKNOWN
```

---

# 20. PragmaticFunction

```text
HEDGE
SOFTENER
INTENSIFIER
FOCUS_MARKER
ATTITUDE_MARKER
POLITENESS_MARKER
CHALLENGE_MARKER
DISTANCING_MARKER
PARTIAL_AGREEMENT
OBJECTION
DIFFERENTIATION
CONDITIONAL_OPENING
```

---

# 21. PatternClass

```yaml
PatternClass:
  pattern_class_id
  slug
  positive_label
  neutral_label
  short_description
  affected_dimensions
  evidence_class
```

MVP-Beispiele:

```text
GENERALIZATION
ABSOLUTIZATION
NEGATION
URGENCY
INTERNAL_PRESSURE
EXTERNAL_OBLIGATION
SOCIAL_NORM
SELF_DEVALUATION
PERSON_DEVALUATION
CHOICE_LANGUAGE
RESPONSIBILITY_LANGUAGE
OPENING_LANGUAGE
CONTRAST
DISCOUNTING
HEDGING
SELF_CORRECTION
PREDICATIVE_LABELING
IDENTITY_DEVALUATION
LEARNING_FRAME
UNCERTAIN_COMMITMENT
EXTERNAL_EXPECTATION
INTERNALIZED_EXPECTATION
UNSPECIFIED_AUTHORITY
NORMATIVE_ADVICE
REPORTED_CLAIM
EXPECTED_OUTCOME
```

---

# 22. TargetType

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

Wichtig für:

- Wertschätzung,
- Verbindung,
- Wirksamkeit.

---

# 23. ExpectationSource

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

Besonders wichtig für:

- sollen,
- müssen,
- dürfen,
- soziale Normen.

---

# 24. ProsodyCue

Für gesprochene Sprache:

```yaml
ProsodyCue:
  stress
  intonation
  pause
  emphasis
  question_contour
  ironic_contour
  confidence
```

Im MVP zunächst optional / vorbereitet.

---

# 25. Collocation

```yaml
Collocation:
  collocation_id
  source_node_id
  target_node_id
  direction
  token_distance
  order_constraint
  normalized_frequency
  corpus_id
  register
  perception_channel
  effect_type
  effect_strength
  confidence
```

---

# 26. CollocationEffectType

```text
AMPLIFY_POSITIVE
AMPLIFY_NEGATIVE
REDUCE_POSITIVE
REDUCE_NEGATIVE
NEGATE
SHIFT_MEANING
CONTEXT_DEPENDENT
NEUTRAL
```

---

# 27. UsageProfile

```yaml
UsageProfile:
  usage_id
  owner_node_id
  corpus_id
  perception_channel
  mental_activation
  register
  genre
  region
  period
  normalized_frequency
  confidence
```

---

# 28. AmbiguityProfile

```yaml
AmbiguityProfile:
  ambiguity_id
  owner_node_id
  ambiguity_types
  possible_interpretations
  disambiguation_features
  default_confidence
  user_notice_threshold
```

AmbiguityTypes:

```text
SEMANTIC
PHONETIC
ORTHOGRAPHIC
PRAGMATIC
SYNTACTIC
REGISTER
RESONANCE
```

---

# 29. ResonanceProfile

```yaml
ResonanceProfile:
  resonance_id
  relation_id
  resonance_type
  default_relevance
  auditory_weight
  visual_weight
  mixed_weight
  external_speech_weight
  silent_reading_weight
  read_aloud_weight
  inner_speech_weight
  register_weights
  repetition_weight
  personal_override_allowed
  evidence_class
  confidence
  editorial_note
```

---

# 30. Resonanzgrundsatz

Homophonie, Klangnähe oder spirituelle Resonanz werden **nicht automatisch semantisch vererbt**.

Falsch:

```text
hast inherits meaning/effect from hasst
```

Richtig:

```text
hast HOMOPHONE_OF hasst
relation HAS ResonanceProfile
```

---

# 31. Dimension

MVP-Kernachsen:

```text
Ohnmacht ↔ Wirksamkeit
Trennung ↔ Verbindung
Abwertung ↔ Wertschätzung
Unklarheit ↔ Klarheit
Zwang ↔ Freier Wille
Begrenzung ↔ Offenheit
```

Standard-UI kann nur positive Attribute zeigen.

---

# 32. Dimension-Objekt

```yaml
Dimension:
  dimension_id
  slug
  positive_label
  negative_label
  short_description
  detailed_description
  default_weight
  default_visible
  active
  version
```

---

# 33. DimensionContribution

```yaml
DimensionContribution:
  contribution_id
  owner_node_id
  dimension_id
  base_value
  confidence
  evidence_class
  context_condition
  perception_channel
  mental_activation
  register
  target_type
  expectation_source
  active
```

Owner kann sein:

- Sense
- Phrase
- Relation
- Collocation
- PatternClass

---

# 34. Warum Contributions nicht nur am Lexem hängen

Ein Lexem ist meist zu allgemein.

Beispiel:

```text
müssen
```

kann sein:

- Pflicht,
- innerer Druck,
- Sicherheit,
- Schlussfolgerung.

Der Dimensionsbeitrag gehört daher möglichst an:

- Sense,
- Phrase,
- PatternClass.

---

# 35. Avoidability

```text
YES
NO
CONTEXT_DEPENDENT
```

Beschreibt, ob ein sprachliches Phänomen sinnvoll vermeidbar ist.

---

# 36. Alternative

```yaml
Alternative:
  alternative_id
  source_pattern
  replacement_text
  intent_type
  register
  context_conditions
  editorial_status
```

IntentTypes:

```text
OWN_DECISION
EXTERNAL_REQUIREMENT
CLEAR_EXPECTATION
REQUEST
SAFETY_DIRECTIVE
PRIORITY
COMMITMENT
REFLECT
COMPARE
KEEP
REPHRASE
```

---

# 37. AlternativeQuality

```yaml
AlternativeQuality:
  alternative_id
  semantic_fidelity
  naturalness
  clarity
  register_fit
  dimension_improvement
  resonance_change
  confidence
```

---

# 38. Source

```yaml
Source:
  source_id
  source_type
  title
  author
  publication
  url
  accessed_at
  reliability_note
  status
```

SourceTypes:

```text
DICTIONARY
CORPUS
SCIENTIFIC
BOOK
ARTICLE
WORKSHOP_OBSERVATION
COMMUNITY
EDITORIAL
OTHER
```

---

# 39. Laufzeitobjekt: AnalysisContext

```yaml
AnalysisContext:
  language
  register
  perception_channel
  mental_activation
  speaker_role
  recipient_role
  target_type
  expectation_source
  user_mode
  resonance_mode
  location_locale
  temporal_context
```

---

# 40. Laufzeitobjekt: AnalysisHit

```yaml
AnalysisHit:
  hit_id
  source_span
  detected_form
  lexeme_id
  selected_sense_id
  sense_confidence
  matched_phrase_ids
  relation_ids
  relation_claim_ids
  pattern_class_ids
  target_type
  expectation_source
  ambiguity_profile
  context_confidence
  dimension_contributions
  resonance_relevance
  explanation
  alternative_candidates
```

---

# 41. Laufzeitobjekt: DimensionResult

```yaml
DimensionResult:
  dimension_id
  raw_score
  normalized_score
  confidence
  evidence_mix
  contributing_hits
  positive_label
  negative_label
  assessable
```

---

# 42. Laufzeitobjekt: AnalysisResult

```yaml
AnalysisResult:
  analysis_id
  context
  hits
  propositions
  dimension_results
  resonance_summary
  detected_patterns
  alternatives
  overall_confidence
  wing_score
```

---

# 43. Persistenz versus Laufzeit

## Persistieren

- Lexeme
- WordForm
- Pronunciation
- Sense
- Phrase
- Relation
- RelationClaim
- PatternClass
- Dimension
- Collocation
- UsageProfile
- AmbiguityProfile
- ResonanceProfile
- Alternative
- Source

## Primär Laufzeit

- Proposition
- AnalysisContext
- AnalysisHit
- DimensionResult
- AnalysisResult

Optional können Analysen später mit Einwilligung gespeichert werden.

---

# 44. Redaktionsstatus

Für Community und Qualitätssicherung:

```text
DRAFT
PROPOSED
REVIEW
VERIFIED
PUBLISHED
DEPRECATED
REJECTED
```

---

# 45. Versionierung

Jedes fachlich relevante Objekt sollte besitzen:

```text
version
created_at
updated_at
created_by
reviewed_by
```

Später zusätzlich:

```text
supersedes_id
change_reason
```

---

# 46. Community-Erweiterbarkeit

Das Modell unterstützt später:

- Vorschläge,
- Ergänzungen,
- alternative Claims,
- neue Senses,
- Quellen,
- Resonanzbeobachtungen,
- PatternClasses,
- Coach-spezifische Ergänzungen.

Community-Daten verändern den veröffentlichten Kern nicht direkt.

---

# 47. Technische Implementierungsstrategie

Für den MVP ist keine Graphdatenbank zwingend erforderlich.

Empfehlung:

> **PostgreSQL mit relationalem Kern und generischen Relations-Tabellen.**

Warum?

- solide Transaktionen,
- gut versionierbar,
- performant,
- JSONB für flexible Metadaten,
- ausreichend für Relation Graph v1.

Später kann bei Bedarf ergänzt werden:

- Graph-Projektion,
- Suchindex,
- Embedding-/Vector-Index.

---

# 48. Mögliche Tabellenstruktur

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
collocations
usage_profiles
ambiguity_profiles
resonance_profiles
alternatives
alternative_qualities
sources
object_sources
```

Laufzeitobjekte müssen nicht zwingend Tabellen besitzen.

---

# 49. Generic Node Reference

Mehrere Objekte benötigen generische Verweise.

Möglichkeiten:

## Variante A

```text
source_type
source_id
```

## Variante B

zentrale `nodes`-Tabelle.

Für MVP empfiehlt sich:

> **Variante A**, sofern die Implementierung übersichtlich bleibt.

Eine zentrale Graph-Node-Schicht kann später ergänzt werden.

---

# 50. Mindestvalidierung für Seed-Daten

Ein veröffentlichter Seed-Eintrag sollte mindestens besitzen:

- Lexem
- eine relevante Wortform
- mindestens einen Sense
- Kurzbeschreibung
- mindestens eine Phrase oder Kontextregel
- mindestens einen Dimensionsbeitrag oder klare Neutralität
- Evidenzklasse
- Confidence
- Quellenstatus
- Ambiguitätshinweis falls nötig
- Resonanzlink falls relevant

---

# 51. MVP-Seed-Objekte

Aktuell validiert:

```text
müssen
hast / hasst
aber
ja, aber
eigentlich
dürfen
Problem
Schuld
Fehler
frei
versuchen
sollen
```

Zusätzlich Ambiguitätstests:

```text
fort / Ford
Seite / Saite
mehr / Meer
wieder / wider
umfahren / umfahren
übersetzen / übersetzen
modern / modern
```

---

# 52. Datenmodell-Reifegrad

Nach den bisherigen Tests gilt das Modell als:

> **MVP-ready für die erste Berechnungslogik.**

Neue Begriffe sollten nun überwiegend durch:

- neue Senses,
- neue Relations,
- neue PatternClasses,
- neue Contributions

integrierbar sein.

Neue fundamentale Kernobjekte sollten nur noch selten nötig werden.

---

# 53. Bewusste Nicht-Ziele v1.0

Nicht Teil des Kernmodells:

- Persönlichkeitsdiagnostik
- psychologische Diagnosen
- automatische Wahrheitsbewertung
- automatische moralische Klassifikation
- deterministische Unterbewusstseinsbehauptungen
- vollständige Prosodieanalyse
- vollständige Dialektmodellierung
- automatische Mehrsprachigkeitsrelationen

---

# 54. Schutzprinzipien des Datenmodells

## 54.1 Bedeutung vor Oberfläche

Keine Bewertung nur aufgrund eines Wortes.

## 54.2 Sense vor Score

Erst Bedeutung disambiguieren.

## 54.3 Phrase vor Einzelmarker

Phrasen können Einzelwerte überstimmen.

## 54.4 Kontext vor Standardwert

Reale Situationen modifizieren Beiträge.

## 54.5 Resonanz getrennt von Semantik

Klangbeziehungen sind keine Bedeutungsvererbung.

## 54.6 Evidenz sichtbar

Interpretation und Fakt bleiben unterscheidbar.

## 54.7 Nutzer entscheidet

Resonanz- und Tiefenmodi bleiben konfigurierbar.

## 54.8 Natürlichkeit schützen

Alternative Sprache darf nicht künstlich werden.

---

# 55. Nächste Phase: Berechnungslogik v0.1

Das Datenmodell unterstützt jetzt die folgende Pipeline:

```text
Input
→ Token/Form Recognition
→ Phrase Matching
→ Sense Disambiguation
→ Proposition Detection
→ Relation Resolution
→ Pattern Detection
→ Context Resolution
→ Base Dimension Contributions
→ Modifiers
→ Resonance Layer
→ Frequency/Repetition
→ Confidence
→ Dimension Aggregation
→ WingScore
```

---

# 56. Berechnungslogik: empfohlene Reihenfolge

## 1. Sense Confidence
## 2. Base Contribution
## 3. Phrase Modifier
## 4. Negation / Intensifier
## 5. Context Modifier
## 6. Target / Expectation Modifier
## 7. Ambiguity Modifier
## 8. Resonance Modifier
## 9. Frequency / Repetition Modifier
## 10. Dimension Aggregation
## 11. Dimension Confidence
## 12. WingScore

---

# 57. Definition of Done – Datenmodell v1.0

- [x] Lexeme definiert
- [x] WordForm definiert
- [x] Pronunciation definiert
- [x] Sense definiert
- [x] Phrase / PhrasePattern definiert
- [x] Proposition berücksichtigt
- [x] Relation Taxonomy konsolidiert
- [x] RelationClaim definiert
- [x] Evidenzmodell definiert
- [x] Confidence getrennt
- [x] Register definiert
- [x] Wahrnehmungskanal definiert
- [x] mentale Aktivierung definiert
- [x] PragmaticFunction definiert
- [x] PatternClass definiert
- [x] TargetType definiert
- [x] ExpectationSource definiert
- [x] Prosody vorbereitet
- [x] Collocation definiert
- [x] UsageProfile definiert
- [x] AmbiguityProfile definiert
- [x] ResonanceProfile definiert
- [x] Dimension / Contribution definiert
- [x] Alternative / AlternativeQuality definiert
- [x] AnalysisContext definiert
- [x] AnalysisHit definiert
- [x] DimensionResult definiert
- [x] AnalysisResult definiert
- [x] Persistenz-/Laufzeittrennung definiert
- [x] Community-Erweiterbarkeit berücksichtigt
- [x] technische MVP-Persistenzstrategie vorgeschlagen
- [x] Scoring-Pipeline vorbereitet

---

# 58. Leitgedanke

> **Das Datenmodell soll die Tiefe von Sprache bewahren, ohne die Leichtigkeit der Nutzererfahrung zu verlieren.  
> Komplexität gehört in die Engine – Erkenntnis gehört zum Menschen.**
