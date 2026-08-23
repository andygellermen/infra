# Spiritual Language Analyzer
## Sense-/Proposition-Resolver v0.2

**Status:** MVP-Fach- und Laufzeitkonzept  
**Version:** 0.2  
**Datum:** 19. August 2026

---

# 1. Ziel

Der Resolver soll vor dem Scoring bestimmen:

1. welche Bedeutung eines Ausdrucks im konkreten Satz gemeint ist,
2. welche Teilaussagen (Propositionen) vorliegen,
3. wie diese Propositionen zueinander stehen,
4. worauf sich Bewertungen oder Erwartungen richten,
5. welche Kontext- und Pragmatiksignale den Dimensionsbeitrag verändern.

Grundsatz:

> **Erst Bedeutung und Aussagebeziehung klären – dann scoren.**

---

# 2. Resolver-Pipeline

```text
Text
→ Token / Morphologie
→ Phrase Matching
→ Candidate Senses
→ Syntax-/Dependency-Hinweise
→ Proposition Segmentation
→ Discourse Relations
→ TargetType / ExpectationSource
→ Context Fit
→ Sense Ranking
→ Ambiguity Handling
→ Pattern Emission
→ Scoring
```

---

# 3. SenseCandidate

```yaml
SenseCandidate:
  sense_id
  lexical_match
  phrase_fit
  syntax_fit
  domain_fit
  register_fit
  discourse_fit
  context_fit
  collocation_fit
  negative_evidence
  final_confidence
```

---

# 4. Ranking v0.2

Empfohlene gewichtete Startlogik:

```text
sense_score =
0.20 lexical_match
+ 0.20 phrase_fit
+ 0.15 syntax_fit
+ 0.10 domain_fit
+ 0.10 register_fit
+ 0.10 discourse_fit
+ 0.10 collocation_fit
+ 0.05 context_fit
- negative_evidence
```

Die Werte bleiben Admin-/Testparameter.

---

# 5. Ambiguitätsentscheidung

```text
top_score >= 0.75
AND top_score - second_score >= 0.20
→ HIGH_CONFIDENCE_SENSE
```

```text
top_score >= 0.60
AND gap >= 0.10
→ MEDIUM_CONFIDENCE_SENSE
```

sonst:

```text
AMBIGUOUS
```

Bei `AMBIGUOUS`:
- Beitrag abschwächen,
- keine harte Rewrite-Empfehlung,
- ggf. mehrere Lesarten anzeigen.

---

# 6. Proposition

```yaml
Proposition:
  proposition_id
  source_span
  subject
  predicate
  object
  complements
  modality
  negation
  tense
  target_type
  expectation_source
  confidence
```

---

# 7. Proposition Segmentation

Beispiel:

> „Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.“

P1:
```text
Ich verstehe, dass dir das wichtig ist.
```

P2:
```text
Für mich kommt diese Lösung nicht infrage.
```

Zusätzlich:

```text
ACKNOWLEDGEMENT(P1)
CLEAR_BOUNDARY(P2)
```

Komposition:

```text
P1 + P2
→ RESPECTFUL_BOUNDARY
```

---

# 8. „aber“ propositionell

Beispiel:

> „Ja, aber das funktioniert sowieso nie.“

P0:
```text
ja
```

P1:
```text
das funktioniert sowieso nie
```

Relation:

```text
DISCOUNTS_PREVIOUS_PROPOSITION
```

Pattern:
```text
DISCOUNTING
GENERALIZATION
```

Beispiel konstruktiv:

> „Ja, aber wir sollten zwischen den beiden Situationen unterscheiden.“

P1:
```text
Zustimmung
```

P2:
```text
Differenzierungsbedarf
```

Relation:
```text
DISCOURSE_CONTRAST
```

Pattern:
```text
CONSTRUCTIVE_DIFFERENTIATION
```

Der Resolver muss deshalb nicht `aber` bewerten, sondern die **Relation zwischen den Aussagen**.

---

# 9. Resolver für „müssen“

Candidate Senses:

```text
EXTERNAL_NECESSITY
FORMAL_OBLIGATION
SAFETY_NECESSITY
INTERNAL_PRESSURE
SOCIAL_NORM
EMPHATIC_RECOMMENDATION
EPISTEMIC_INFERENCE
```

Entscheidungsmerkmale:

## SAFETY_NECESSITY
Signale:
- sofort
- Gefahr
- Brand
- verlassen
- Schutz
- Notfall

## FORMAL_OBLIGATION
Signale:
- gesetzlich
- verpflichtet
- Vertrag
- vorgeschrieben
- Frist

## EPISTEMIC_INFERENCE
Signale:
- muss wohl
- muss schon
- muss bereits
- das muss bedeuten

## INTERNAL_PRESSURE
Signale:
- ich muss
- unbedingt
- immer
- endlich
- perfekt
- schaffen

Default bei Unsicherheit:
```text
REQUIRES_CONTEXT
```

---

# 10. Resolver für „sollen“

Candidate Senses:

```text
DIRECTIVE
ADVICE
SOCIAL_NORM
INTERNALIZED_EXPECTATION
EXPECTED_OUTCOME
REPORTED_CLAIM
CONDITIONAL_OPENING
```

Regeln:

```text
"er/sie soll" + Eigenschaft
→ REPORTED_CLAIM
```

```text
"solltest du" + Frage/Hilfe/Bedarf
→ CONDITIONAL_OPENING
```

```text
"ich sollte" + längst/endlich/eigentlich
→ INTERNALIZED_EXPECTATION
```

```text
"man sollte"
→ SOCIAL_NORM + ExpectationSource UNSPECIFIED/CULTURE
```

---

# 11. Resolver für „dürfen“

Candidate Senses:

```text
PERMISSION
PROHIBITION
POLITE_REQUEST
PROBABILITY
RHETORICAL_LEGITIMATION
```

Regeln:

```text
dürfen + nicht
→ PROHIBITION
```

```text
"darf ich" + Frage
→ POLITE_REQUEST
```

```text
"dürfte" + Ergebnis/Einschätzung
→ PROBABILITY
```

---

# 12. Resolver für „frei“

Candidate Senses:

```text
LIBERTY
AVAILABLE
FREE_OF_CHARGE
UNBOUND
FREE_FROM
```

Beispiele:

```text
Eintritt ist frei
→ FREE_OF_CHARGE
```

```text
Termin ist frei
→ AVAILABLE
```

```text
ich bin frei von ...
→ FREE_FROM
```

Nur `LIBERTY` darf direkt die Dimension Freier Wille beeinflussen.

---

# 13. Resolver für „Problem“

Candidate Senses / Target:

```text
TECHNICAL_ISSUE
TASK_DIFFICULTY
PERSONAL_CHALLENGE
PERSON_LABEL
SCIENTIFIC_PROBLEM
```

Regeln:

```text
technisches Problem + System/Schnittstelle/Code
→ TECHNICAL_ISSUE
```

```text
"du bist das Problem"
→ PERSON_LABEL
→ TargetType PERSON
```

---

# 14. Resolver für „Fehler“

```text
TECHNICAL_ERROR
FACTUAL_ERROR
BEHAVIOR_ERROR
LEARNING_EVENT
IDENTITY_LABEL
```

```text
"im Code ist ein Fehler"
→ TECHNICAL_ERROR
```

```text
"ich habe einen Fehler gemacht"
→ BEHAVIOR_ERROR
```

```text
"du bist ein Fehler"
→ IDENTITY_LABEL
```

```text
"der Fehler zeigt uns ..."
→ LEARNING_EVENT
```

---

# 15. Resolver für „eigentlich“

PragmaticFunctions:

```text
HEDGE
ORIGINAL_INTENTION
IMPLICIT_CONTRAST
POLITENESS_SOFTENER
SELF_CORRECTION
ATTITUDE_MARKER
```

Beispiel:

> „Eigentlich wollte ich absagen, aber ich bin noch unsicher.“

Resolver:
```text
ORIGINAL_INTENTION
+ HEDGE
+ UNCERTAIN_COMMITMENT
```

Kein pauschaler Klarheitsmalus.

---

# 16. Homograph „umfahren“

Candidate Senses:

```text
DRIVE_AROUND
KNOCK_DOWN_BY_DRIVING
```

Entscheidungsmerkmale:
- Syntax
- Objektsemantik
- Prosodie/Akzent, falls Audio
- Kontext

Text-only:

> „Wir müssen das Hindernis umfahren.“

`Hindernis` erhöht stark die Wahrscheinlichkeit von `DRIVE_AROUND`.

Wenn Confidence nicht ausreichend:
```text
AMBIGUOUS
```

---

# 17. TargetType Resolver

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

Beispiel:

> „Du bist unzuverlässig.“

Target:
```text
PERSON
```

> „Die Abgabe war unzuverlässig organisiert.“

Target:
```text
PROCESS
```

Dies verändert besonders:
- Wertschätzung
- Verbindung

---

# 18. ExpectationSource Resolver

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

Beispiel:

> „Ich sollte längst weiter sein.“

wenn keine externe Quelle genannt:
```text
INTERNALIZED
```

Confidence ggf. mittel, nicht absolut.

---

# 19. DiscourseRelation Resolver

MVP-Typen:

```text
CONTRAST
CONCESSION
CAUSE
CONSEQUENCE
CORRECTION
ADDITION
CONDITION
DISCOUNTING
```

Beispiele:

```text
aber → candidate CONTRAST
trotzdem → CONCESSION
weil → CAUSE
deshalb → CONSEQUENCE
wenn/falls → CONDITION
```

Die Relation wird durch propositionale Bedeutung validiert.

---

# 20. Composition Rules

### Respectful Boundary

```text
P1 = ACKNOWLEDGEMENT
P2 = CLEAR_BOUNDARY
distance <= 2 propositions
→ RESPECTFUL_BOUNDARY
```

### Agency Recovery

```text
P1 = REALISTIC_LIMIT
P2 = CHOICE_LANGUAGE / RESPONSIBILITY_LANGUAGE
relation = CONTRAST or ADDITION
→ AGENCY_RECOVERY
```

### Learning Recovery

```text
P1 = ERROR_DESCRIPTION / FAILURE_EVENT
P2 = LEARNING_FRAME
→ LEARNING_RECOVERY
```

### Ja-aber Discount

```text
P1 = AGREEMENT
P2 = ABSOLUTIZATION / REJECTION / NO_CHOICE
relation = CONTRAST
→ DISCOUNTING
```

---

# 21. Negation Scope

Negation muss an Proposition/Sense hängen.

Beispiel:

> „Du musst das nicht heute entscheiden.“

Nicht:
```text
NEGATE entire sentence
```

Sondern:
```text
NEGATES obligation scope
```

Beispiel:

> „Nicht du musst das entscheiden.“

Negation fokussiert den Akteur, nicht zwingend die Verpflichtung.

Für MVP:
- Syntaxindikatoren nutzen,
- bei unklarer Scope Confidence reduzieren.

---

# 22. Modalität und Zeitbezug

```yaml
Modality:
  necessity
  possibility
  permission
  expectation
  intention
  probability
```

Diese Eigenschaften unterstützen die Dimension:
- Freier Wille
- Klarheit
- Wirksamkeit

---

# 23. Resolver Output

```yaml
ResolverResult:
  propositions
  selected_senses
  ambiguity_profiles
  discourse_relations
  target_types
  expectation_sources
  pragmatic_functions
  pattern_candidates
  overall_resolver_confidence
```

---

# 24. LLM/NLP-Aufgabenteilung

Rule-/NLP-seitig gut lösbar:
- Morphologie
- bekannte Phrase
- Negation
- Konnektor
- einfache Syntax
- Lexem/Sense-Kandidaten

LLM-geeignet:
- Propositionparaphrase
- TargetType bei komplexem Satz
- ExpectationSource
- Ironieverdacht
- feinere Sense-Auswahl
- discourse/pragmatic interpretation

Grundsatz:
```text
LLM proposes
Rule Engine validates / scores
```

---

# 25. Fallback

Wenn Resolver unsicher:

```text
resolver_confidence < threshold
→ no hard score from ambiguous feature
```

UI:
> „Diese Passage lässt mehrere Lesarten zu.“

---

# 26. Testfälle

Pflichtfälle:
- G02 Verpflichtung
- G08 Respectful Boundary
- G09/G10 Ja-aber
- G11/G13 sollen
- G14/G15 dürfen
- G16 Agency Recovery
- G19/G20 Problem
- G21 Fehler/Lernen
- G24 umfahren
- G25 frei
- G26 Hörensagen
- G30 eigentlich

---

# 27. Leitgedanke

> **Ein Resolver ist dann gut, wenn er nicht möglichst schnell eine Bedeutung auswählt, sondern zuverlässig erkennt, wann Sprache mehrdeutig bleibt.**
