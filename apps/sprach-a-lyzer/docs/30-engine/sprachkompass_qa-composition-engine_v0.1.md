# Sprachkompass – Q/A Composition Engine v0.1

**Status:** Fachlich-technische Referenzspezifikation  
**Datum:** 20. August 2026  
**Ziel:** Fragekontext und Antwortanalyse kontrolliert komponieren, ohne Frageeffekte mit Personenmerkmalen oder Kausalität zu verwechseln.

---

## 1. Kernidee

Der Sprach-A-Lyzer analysiert nicht mehr nur eine isolierte Antwort.

Er analysiert:

```text
QUESTION
   ↓
QuestionContext
   ↓
ANSWER
   ↓
AnswerRelevance
   ↓
Language Analysis
   ↓
Question × Answer Composition
   ↓
Construct Evidence
   ↓
Dimension Evidence
   ↓
Explanation / Reflection
```

Die Frage ist damit ein **Kontextanker**, kein Score-Geber.

> **Die Frage darf Interpretation präzisieren. Sie darf kein Ergebnis vorwegnehmen.**

---

## 2. Sanfte Kausalität

Fragen beeinflussen, worüber und wie Menschen antworten. Deshalb ist der Q/A-Kontext nicht neutral.

Wir unterscheiden fünf Inferenzstufen:

### C0 – OBSERVATIONAL
Nur die Antwort wird beobachtet.

Beispiel:

> „Ich kann sowieso nichts ändern.“

Zulässig:

> Generalisierung und geringe Handlungssprache wurden in dieser Äußerung erkannt.

Nicht zulässig:

> Die Person ist grundsätzlich ohnmächtig.

### C1 – QUESTION_CONDITIONED_RELEVANCE
Die Antwort bezieht sich nachweisbar auf ein durch die Frage aktiviertes Konstrukt.

Frage:

> „Was davon liegt heute tatsächlich in deinem Einflussbereich?“

Antwort:

> „Eigentlich gar nichts. Ich kann nur abwarten.“

Zulässig:

> In der Antwort auf die Frage nach dem eigenen Einfluss zeigt sich wenig wahrgenommener Handlungsspielraum.

Das ist die Standardstufe für die Q/A Composition Engine.

### C2 – ELICITATION_ASSOCIATION
Die Frage hat plausibel dazu beigetragen, dass ein bestimmtes Sprachmuster sichtbar wurde.

Zulässige Formulierung:

> Die Frage nach Wahlmöglichkeiten hat eine Antwort hervorgerufen, in der mehrere Alternativen explizit benannt wurden.

Nicht zulässig:

> Die Frage hat den Nutzer offener gemacht.

### C3 – WITHIN_SESSION_TEMPORAL_ASSOCIATION
Zwischen früheren und späteren Antworten verändert sich die Sprache nach bestimmten Reflexionsfragen.

Zulässig:

> In späteren Antworten werden häufiger konkrete Wahl- und Handlungsmöglichkeiten formuliert.

Nicht zulässig:

> Die Folgefragen haben die Selbstwirksamkeit des Nutzers erhöht.

### C4 – CAUSAL_EFFECT
Echte Ursache-Wirkungs-Behauptung.

Beispiel:

> Diese Frage erhöht Selbstwirksamkeit.

**Im Produkt standardmäßig nicht zulässig.**

Dafür wären geeignete Forschungsdesigns erforderlich, z. B.:

- randomisierte Vergleichsgruppen
- alternative Frageformulierungen
- ausreichende Stichprobe
- vorab definierte Outcomes
- Kontrolle von Reihenfolge-/Kontexteffekten

---

## 3. Leitregel

> **Fragekonditionierte Evidenz ist Evidenz über die Antwort im Kontext der Frage – nicht automatisch Evidenz über eine stabile Eigenschaft der Person.**

---

## 4. Eingabemodell

```yaml
QuestionContext:
  question_id
  phase
  audience
  primary_construct
  secondary_constructs
  expected_dimensions
  context_relevance_prior
  leadingness
  specificity
  risk_level

AnswerContext:
  answer_text
  input_mode
  answer_relevance
  answer_completeness
  ambiguity
  response_independence
  detected_patterns
  sense_candidates
  propositions
  target_types
  dimension_contributions
```

---

## 5. Neue Laufzeitobjekte

### QuestionAnswerObservation

```yaml
QuestionAnswerObservation:
  question_id
  answer_id
  answer_relevance
  construct_fit
  context_gain
  causal_level
  construct_evidence[]
  dimension_evidence[]
  disconfirming_evidence[]
  assessability
  explanation_keys[]
```

### ConstructEvidence

```yaml
ConstructEvidence:
  construct_key
  direction
  strength
  confidence
  source_patterns[]
  question_supported
  independently_observable
```

---

## 6. Question Score Bias

Immer:

```text
question_score_bias = 0.0
```

Eine Frage darf niemals direkt:

- einen positiven Dimensionswert addieren
- einen negativen Dimensionswert addieren
- Assessability erzeugen
- einen WingScore anheben oder senken

---

## 7. Answer Relevance

`answer_relevance ∈ [0,1]`

Startheuristik:

```text
0.00–0.34  OFF_TOPIC
0.35–0.54  WEAK
0.55–0.74  RELEVANT
0.75–1.00  STRONGLY_RELEVANT
```

Bei `answer_relevance < 0.35`:

```text
QuestionContext prior = disabled
```

Die Antwort darf weiterhin normal sprachlich analysiert werden.

---

## 8. Construct Fit

`construct_fit ∈ [0,1]`

Bewertet, ob die tatsächlich erkannten Propositionen das erwartete Konstrukt berühren.

Beispiel Frage zu `LOCUS_OF_CONTROL`:

- „Ich kann meinen Teil vorbereiten.“ → hoher Fit
- „Mein Chef ist im Urlaub.“ → niedriger Fit
- „Ich bin müde.“ → sehr niedriger Fit

---

## 9. Leadingness

`leadingness ∈ [0,1]`

Fragen, die bereits eine bestimmte Bewertung nahelegen, bekommen einen höheren Wert.

Beispiel:

Niedrig:

> „Welche Möglichkeiten siehst du?“

Höher:

> „Welche deiner negativen Glaubenssätze hindern dich daran, frei zu sein?“

Hohe Leadingness reduziert den Q/A-Kontextgewinn.

Spirituell-reflexive Varianten dürfen deshalb **nicht automatisch stärkere Context Priors** erhalten.

---

## 10. Response Independence

`response_independence ∈ [0,1]`

Grobe Einschätzung, wie viel eigenständige Information die Antwort enthält.

Beispiel:

Frage:

> „Fühlst du dich frei?“

Antwort:

> „Ja.“

→ geringe Independence.

Frage:

> „Welche Möglichkeiten siehst du?“

Antwort:

> „Ich könnte A tun, B verschieben oder C ganz lassen.“

→ hohe Independence.

---

## 11. Context Gain

Der Fragekontext darf nur die **Confidence einer bereits in der Antwort vorhandenen Interpretation** moderat erhöhen.

Startformel:

```text
K =
  answer_relevance
  × construct_fit
  × specificity
  × (1 - leadingness)
  × response_independence
```

Dann:

```text
context_gain = min(0.15, 0.15 × K)
```

und:

```text
qa_confidence =
  min(1.0, base_answer_confidence × (1 + context_gain))
```

Wichtig:

```text
base contribution VALUE bleibt unverändert.
```

Die Frage beeinflusst also zunächst nur Confidence, nicht Valenz oder Rohstärke.

---

## 12. Harte Obergrenze

Der Q/A-Kontext darf eine Interpretation maximal um:

```text
+15 % relative Confidence
```

verstärken.

Dieser Wert ist ein **Startparameter**, kein wissenschaftlich etablierter Wert.

Er wird im Golden Corpus kalibriert.

---

## 13. Keine Assessability aus der Frage allein

Unzulässig:

```text
Answer evidence = weak
Question context = strong
→ ASSESSABLE
```

Stattdessen:

```text
question context can support
but cannot satisfy independent evidence gate
```

Mindestens eine eigenständig beobachtbare Antwort-Evidenz muss vorhanden sein.

---

## 14. Confirming vs. Disconfirming Evidence

Der Fragekontext darf nicht nur bestätigende Evidenz suchen.

Frage:

> „Was liegt in deinem Einflussbereich?“

Antwort:

> „Die Entscheidung des Kunden nicht. Aber meine Vorbereitung, meine Rückfrage und meinen Zeitplan kann ich beeinflussen.“

Erwartet:

- REALISTIC_LIMIT
- CHOICE_LANGUAGE
- AGENCY_RECOVERY

Eine gegenteilige Antwort muss ebenso sauber erkannt werden.

---

## 15. Construct Coverage

Eine Frage kann mehrere Konstrukte berühren, aber nur eines ist primär.

Beispiel:

```text
primary:
LOCUS_OF_CONTROL

secondary:
AGENCY
CLARITY
```

Sekundärkonstrukte erhalten keinen automatischen Prior.

Sie werden nur aktiviert, wenn Antwortmuster sie tatsächlich stützen.

---

## 16. Question × Pattern Compositions

Neue regelbare Kompositionen:

```text
QUESTION(LOCUS_OF_CONTROL)
+ ANSWER(NO_CHOICE)
→ LOW_PERCEIVED_INFLUENCE

QUESTION(LOCUS_OF_CONTROL)
+ ANSWER(REALISTIC_LIMIT + CHOICE_LANGUAGE)
→ DIFFERENTIATED_AGENCY

QUESTION(VALUES)
+ ANSWER(OWNED_COMMITMENT)
→ VALUE_ALIGNED_COMMITMENT

QUESTION(IDENTITY_NARRATIVE)
+ ANSWER(SELF_DEVALUATION)
→ IDENTITY_FUSION_SIGNAL

QUESTION(IDENTITY_NARRATIVE)
+ ANSWER(PERSON_BEHAVIOR_SEPARATION)
→ FLEXIBLE_SELF_DESCRIPTION

QUESTION(OPTIONS)
+ ANSWER(MULTIPLE_OPTIONS)
→ OPTION_GENERATION

QUESTION(AMBIVALENCE)
+ ANSWER(BOTH_SIDES_EXPLICIT)
→ ARTICULATED_AMBIVALENCE
```

Diese erzeugen **ConstructEvidence**, nicht automatisch neue Persönlichkeitsscores.

---

## 17. Folgefragenwahl

Adaptive Fragewahl darf auf beobachteten Antwortmustern beruhen.

Beispiel:

```text
GENERALIZATION
+ LOW_AGENCY_LANGUAGE
→ candidate next question:
  EXCEPTIONS / LOCUS_OF_CONTROL
```

Aber:

> Das System soll Fragen **anbieten**, nicht den Nutzer in eine erwartete Antwort führen.

---

## 18. Adaptive Selection Score

Startmodell:

```text
selection_score =
  0.30 × information_gain
+ 0.25 × construct_gap
+ 0.20 × answer_relevance_history
+ 0.15 × phase_fit
+ 0.10 × user_preference_fit
- 0.30 × redundancy
- 0.35 × leadingness
- 0.50 × risk_without_opt_in
```

Alle Gewichte sind Startparameter.

---

## 19. Information Gain

Hoher Information Gain:

- zwei plausible Deutungen können getrennt werden
- bisher nicht assessable Konstrukte können mit geeigneter Antwort Evidenz erhalten
- Widersprüche können geklärt werden

Niedriger Information Gain:

- Frage wiederholt bereits geklärte Inhalte
- Frage ist nur semantische Variation
- Frage zielt auf Konstrukt, das bereits stark evidenzbasiert ist

---

## 20. Sanfte Kausalität im UI

Empfohlene Formulierungen:

**Gut:**

> „Auf die Frage nach deinem Einfluss beschreibst du mehrere konkrete Handlungsmöglichkeiten.“

> „Im Verlauf deiner Antworten taucht häufiger selbstgewählte Handlungssprache auf.“

> „Diese Folgefrage hat eine andere Perspektive in deiner Antwort sichtbar gemacht.“

**Vermeiden:**

> „Die Frage hat deine Selbstwirksamkeit erhöht.“

> „Du bist jetzt freier.“

> „Dein Glaubenssatz wurde aufgelöst.“

---

## 21. Verlaufsauswertung

Für mehrere Q/A-Paare:

```yaml
SessionTrajectory:
  construct_key
  observations[]
  language_pattern_change
  confidence
  interpretation_level: C3
```

Mögliche Messgrößen:

- Anteil konkreter Wahlformulierungen
- absolute Generalisierungen
- Person↔Verhalten-Trennung
- konkrete nächste Schritte
- Perspektivenvielfalt
- owned commitment
- realistic limits

Nicht:

- „Persönlichkeitsfortschritt“
- „spiritueller Entwicklungsgrad“

---

## 22. Corporate Guardrails

Corporate:

- keine intime Offenlegung erzwingen
- keine P4-HIGH-Frage ohne explizite Wahl
- keine Individual-Rankings
- keine Manager-Einsicht in persönliche Q/A-Verläufe
- system-/rollen-/verhaltensbezogene Interpretation bevorzugen
- keine psychologische Diagnostik

---

## 23. Private / spirituell-reflexive Ebene

Spirituell-reflexive Varianten können:

- andere Metaphern
- Sinnfragen
- Resonanzfragen
- innere Stimmigkeit
- Identitäts-/Werte-Reflexion

anbieten.

Sie ändern nicht:

- wissenschaftlichen Evidenzstatus
- Core Score
- Assessability-Gate
- Kausalitätsstufe

---

## 24. Pseudocode

```text
analyzeQA(question, answer):

  q = loadQuestionContext(question)
  a = analyzeLanguage(answer)

  relevance = resolveAnswerRelevance(q, a)

  if relevance < 0.35:
      return composeFreeLanguageOnly(a)

  fit = resolveConstructFit(q, a)
  leadingness = q.leadingness
  independence = resolveResponseIndependence(q, a)
  specificity = q.specificity

  K = relevance * fit * specificity * (1-leadingness) * independence
  gain = min(0.15, 0.15*K)

  for evidence in a.detectedEvidence:
      if evidence.matches(q.primaryConstruct):
          evidence.confidence *= (1+gain)

  qaPatterns = composeQuestionAnswerPatterns(q, a)

  assessability = assess(
      answerEvidence=a.evidence,
      qaEvidence=qaPatterns,
      questionAloneCannotQualify=true
  )

  return QuestionAnswerObservation(...)
```

---

## 25. Definition of Done v0.1

- [ ] QuestionContext runtime model
- [ ] AnswerRelevance resolver
- [ ] ConstructFit resolver
- [ ] Leadingness metadata
- [ ] ResponseIndependence heuristic
- [ ] ContextGain with hard cap
- [ ] question_score_bias hardcoded to zero
- [ ] question alone cannot satisfy assessability
- [ ] confirming + disconfirming evidence
- [ ] Question×Pattern composition rules
- [ ] adaptive question candidate scoring
- [ ] causal inference levels C0–C4
- [ ] UI wording for C1–C3
- [ ] SessionTrajectory without trait claims
- [ ] Corporate privacy/opt-in guardrails
- [ ] Golden Corpus in CI

---

## 26. North Star

> **Wir wollen sichtbar machen, was eine Frage in einer Antwort hervortreten lässt – ohne daraus mehr Ursache, Wahrheit oder Persönlichkeit abzuleiten, als die Daten tragen.**
