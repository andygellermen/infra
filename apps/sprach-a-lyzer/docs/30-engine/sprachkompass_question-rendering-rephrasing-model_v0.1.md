# Sprachkompass – Question Rendering & Rephrasing Model v0.1

**Stand:** 21. August 2026  
**Ziel:** Fragen nutzerseitig verständlicher und passender formulieren lassen, ohne Konstrukt, Aussageziel oder Bewertungsrichtung unbemerkt zu verändern.

## 1. Produktidee

Unter jeder Coaching-Frage stehen zunächst:

```text
[ einfacher formulieren ]
[ anders fragen ]
```

Optional später:

```text
[ kürzer ]
[ mit Beispiel ]
[ warum wird das gefragt? ]
```

Diese Aktionen dienen zugleich Barrierearmut, individueller Verständlichkeit, adaptiver Dialogführung, Qualitätsmessung und redaktioneller Produktpflege.

## 2. Zentrale Trennung

### SIMPLIFY
Verändert Sprache und Satzbau, aber möglichst nicht Perspektive oder Konstrukt.

Original:
> Welche Erwartungen beeinflussen deine Entscheidung gerade – und welche davon möchtest du selbst übernehmen?

Einfach:
> Welche Erwartungen wirken auf deine Entscheidung? Welche davon sind wirklich deine eigenen?

### REPHRASE
Darf die Perspektive wechseln, muss aber denselben Canonical Question Intent respektieren.

Beispiel:
> Wenn niemand etwas von dir erwarten würde: Wie würdest du selbst entscheiden?

Diese Variante kann höhere Leadingness besitzen und wird deshalb separat geprüft.

## 3. Canonical Question Model

```yaml
QuestionCanonical:
  question_id
  construct_intent
  secondary_constructs
  phase
  audience_scope
  risk_level
  canonical_semantics
  prohibited_semantics
```

Die sichtbare Frage ist nur ein Rendering dieses Canonical Intent.

## 4. Question Rendering

```yaml
QuestionRendering:
  rendering_id
  question_id
  rendering_profile
  language_level
  text
  word_count
  sentence_count
  syntactic_complexity
  leadingness
  specificity
  intimacy_level
  spiritual_explicitness
  relational_warmth
  status
  version
```

## 5. Rendering Profiles

```text
CORPORATE_STANDARD
CORPORATE_EASY
PRIVATE_STANDARD
PRIVATE_EASY
DEEP_REFLECTIVE
```

Später ggf.:

```text
VERY_EASY
YOUTH
AUDIO_SHORT
SCREEN_READER_OPTIMIZED
```

## 6. Sprachlevel

MVP:

```text
STANDARD
EASY
VERY_EASY
```

**STANDARD:** normale erwachsene Alltagssprache.

**EASY:**
- kürzere Sätze
- möglichst eine Denkoperation pro Satz
- weniger Nominalstil
- konkrete Verben
- wenig Nebensätze
- Fremdwörter vermeiden

**VERY_EASY:** optional; nicht automatisch mit formaler „Leichter Sprache“ gleichsetzen.

## 7. Rephrase Types

```text
SIMPLIFY
ALTERNATIVE_PERSPECTIVE
SHORTEN
ADD_EXAMPLE
EXPLAIN_INTENT
```

`EXPLAIN_INTENT` erklärt knapp, warum die Frage gestellt wird.

## 8. Qualitätsregeln

Jede Variante wird geprüft auf:

```text
same_construct_intent
semantic_equivalence
leadingness_delta
risk_delta
intimacy_delta
spiritual_explicitness_delta
```

## 9. Hard Guardrails

Nicht automatisch freigeben, wenn:

```text
construct_intent_changed = true
leadingness_delta > configured_limit
risk_level_increased_without_opt_in = true
spiritual_explicitness_increased_without_profile_permission = true
new_diagnostic_claim = true
new_trait_assumption = true
```

## 10. Schlechte Vereinfachung

Original:
> Welche Möglichkeiten siehst du?

Unzulässig:
> Welche guten Möglichkeiten hast du?

`gut` erzeugt bereits eine Bewertungsrichtung.

## 11. User Interaction Event

```yaml
QuestionRenderingEvent:
  event_id
  session_id
  question_id
  from_rendering_id
  action
  resulting_rendering_id
  timestamp
```

Actions:

```text
SIMPLIFY_REQUESTED
REPHRASE_REQUESTED
SHORTER_REQUESTED
EXAMPLE_REQUESTED
INTENT_EXPLANATION_REQUESTED
ANSWERED
SKIPPED
```

## 12. Meta-Signale

Diese Events erzeugen **keine psychologischen Scores**.

```text
QUESTION_COMPREHENSION_SIGNAL
QUESTION_FRICTION_SIGNAL
QUESTION_REPHRASE_PREFERENCE
```

Nutzung: UX, Produktpflege, adaptive Frageauswahl, Barrierearmut.

## 13. Admin Analytics

Beispiel:

```text
CQ017 – Standard Corporate

views:              1.240
simplify_requested:   286
rephrase_requested:    91
skipped:               47

simplify_rate:       23.1 %
rephrase_rate:        7.3 %
skip_rate:            3.8 %
```

Fragen mit hoher Rephrase-/Simplify-Rate werden für redaktionellen Review markiert.

## 14. Corporate/Private Renderings

Gleicher Canonical Intent, andere Einladung:

Canonical `VALUES`

Corporate:
> Was ist dir bei dieser Entscheidung besonders wichtig?

Private:
> Was ist dir persönlich daran wirklich wichtig?

Deep:
> Welcher Wert in dir möchte bei dieser Entscheidung nicht länger übergangen werden?

Der Analyseintent bleibt vergleichbar, Ton und Tiefe unterscheiden sich.

## 15. Q/A Composition Integration

Die Q/A Composition Engine erhält immer:

```yaml
question_id
rendering_id
canonical_construct_intent
leadingness
specificity
language_level
```

Nicht nur den sichtbaren Fragetext.

## 16. A/B-Lernfähigkeit

Vergleichbar werden u. a.:

- Antwortlänge
- Answer Relevance
- Construct Fit
- Abbruchquote
- Simplify Rate
- Rephrase Rate
- Response Independence
- freiwillige Verständlichkeitsbewertung

Keine Kausalitätsbehauptung ohne geeignetes Forschungsdesign.

## 17. KI-Unterstützung

MVP bevorzugt redaktionell freigegebene Renderings.

Später:

```text
LLM proposes
Rule/Validation Engine checks
Approved rendering served
```

Live-Rephrases müssen mindestens prüfen:

- Construct Intent
- Leadingness
- Risk
- Intimacy
- Spiritual Explicitness
- Länge/Lesbarkeit

## 18. Datenmodell

```text
questions
question_renderings
question_rendering_events
question_rendering_metrics
question_rendering_reviews
```

Natural Keys:

```text
CQ001
CQ001:CORPORATE_STANDARD:v1
CQ001:CORPORATE_EASY:v1
CQ001:PRIVATE_STANDARD:v1
```

## 19. Definition of Done v0.1

- [ ] Canonical Question von Rendering getrennt
- [ ] STANDARD/EASY
- [ ] SIMPLIFY/REPHRASE
- [ ] Leadingness pro Rendering
- [ ] Risk-/Intimacy-Delta
- [ ] User Events
- [ ] Admin Metrics
- [ ] Quality Review Trigger
- [ ] Q/A Composition erhält `rendering_id`
- [ ] Managed Import unterstützt Renderings
- [ ] keine psychologische Interpretation von Rephrase-Requests
- [ ] Golden Cases für semantische Äquivalenz

## 20. North Star

> **Eine Frage darf einfacher oder anders werden – aber sie darf dabei nicht heimlich etwas anderes fragen.**
