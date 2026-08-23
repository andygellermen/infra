# Spiritual Language Analyzer
## Assessability Engine v0.1

**Status:** MVP-Fachlogik  
**Version:** 0.1  
**Datum:** 19. August 2026

---

# 1. Ziel

Die Assessability Engine entscheidet **vor der Ausgabe**, ob eine Dimension für den vorliegenden Text überhaupt sinnvoll bewertet werden kann.

Grundsatz:

> **Nicht jede Dimension ist in jedem Text sichtbar. Fehlende Evidenz ist nicht Neutralität.**

Deshalb gilt:

```text
keine ausreichende Evidenz
≠ 50 %
```

sondern:

```text
assessable = false
```

---

# 2. Warum das notwendig ist

Beispiel:

> „Der Termin beginnt um 10 Uhr.“

Klarheit:
- sinnvoll bewertbar.

Verbindung:
- kaum sinnvoll bewertbar.

Wertschätzung:
- kaum sinnvoll bewertbar.

Ein künstlicher Wert von 50 % würde Scheingenauigkeit erzeugen.

---

# 3. Assessability Inputs

Für jede Dimension werden betrachtet:

```text
evidence_mass
independent_hits
pattern_diversity
sense_confidence
context_confidence
ambiguity
target_relevance
proposition_coverage
contradiction_level
```

---

# 4. Evidence Mass

Jeder Contribution-Hit besitzt:

```text
effective_evidence =
abs(base_contribution)
× contribution_confidence
× evidence_factor
```

Die Dimension erhält:

```text
evidence_mass = Σ effective_evidence
```

Aber:

Mehrere identische Wiederholungen zählen nicht vollständig unabhängig.

---

# 5. Independent Hits

Beispiele:

Nicht unabhängig:

> „muss, muss, muss“

Teilweise derselbe Marker.

Eher unabhängig:

- `ich muss`
- `unbedingt`
- `keine Wahl`
- `immer`

Die Engine zählt deshalb:

```text
independent_hit_count
```

---

# 6. Pattern Diversity

Mehrere unterschiedliche PatternClasses erhöhen Assessability.

Beispiel Freier Wille:

```text
INTERNAL_PRESSURE
+ NO_CHOICE
+ EXTERNAL_EXPECTATION
```

liefert stärkere Evidenz als drei identische `muss`-Treffer.

---

# 7. Mindestvoraussetzungen v0.1

Eine Dimension ist grundsätzlich bewertbar, wenn mindestens eine Bedingung erfüllt ist:

## Regel A
```text
independent_hit_count >= 2
AND average_confidence >= 0.65
```

## Regel B
```text
one_high_strength_hit
AND confidence >= 0.85
AND context_fit >= 0.75
```

## Regel C
```text
composed_pattern
AND composition_confidence >= 0.80
```

---

# 8. High-Strength-Hit

Beispiele:

```text
SELF_DEVALUATION
PERSON_DEVALUATION
CLEAR_BOUNDARY
NO_CHOICE
SAFETY_DIRECTIVE
```

Ein solcher Hit kann allein eine Dimension bewertbar machen.

---

# 9. Dimension-Specific Minimums

## Wirksamkeit
relevante Evidenz:
- agency
- no-choice
- responsibility
- self-efficacy
- action ownership

## Verbindung
relevante Evidenz:
- acknowledgement
- attack
- collaboration
- separation
- dialogue

## Wertschätzung
relevante Evidenz:
- person labeling
- respect
- appreciation
- self-devaluation
- behavior/person separation

## Klarheit
relevante Evidenz:
- proposition completeness
- explicit actor/action
- clear boundary
- explicit decision
- ambiguity

## Freier Wille
relevante Evidenz:
- choice
- obligation
- expectation
- prohibition
- consent

## Offenheit
relevante Evidenz:
- alternatives
- absolutization
- opening language
- learning frame
- no-choice

---

# 10. Klarheit als Sonderfall

Klarheit darf nicht über einfache Wortmarker wie `heute`, `nicht` oder `morgen` allein bewertet werden.

Assessability für Klarheit setzt mindestens voraus:

```text
proposition_detected = true
```

Zusätzlich mindestens zwei der folgenden Merkmale:

```text
actor_identified
action_or_state_identified
target_or_object_identified
time_or_condition_identified
boundary_or_decision_identified
reference_resolution_sufficient
```

---

# 11. Ambiguity Penalty

```text
LOW ambiguity     → ×1.00
MEDIUM ambiguity  → ×0.80
HIGH ambiguity    → ×0.55
VERY_HIGH         → ×0.30
```

Diese Gewichtung betrifft Assessability, nicht automatisch den Dimensionswert.

---

# 12. Context Confidence

Wenn eine Dimensionswirkung stark vom Kontext abhängt:

```text
context_confidence < 0.55
→ dimension cannot become HIGH assessability
```

Beispiel:

> „Du solltest früher schlafen.“

Ohne Beziehungskontext:
- Freier Wille teilweise bewertbar,
- Wertschätzung/Verbindung nur eingeschränkt.

---

# 13. Target Relevance

Beispiel:

> „Du bist das Problem.“

TargetType:
```text
PERSON
```

Wertschätzung:
- hoch relevant.

> „Wir haben ein technisches Problem.“

TargetType:
```text
OBJECT / PROCESS
```

Wertschätzung:
- nicht sinnvoll aus dem Wort `Problem` ableiten.

---

# 14. Assessability Score

Intern:

```text
assessability_score =
0.30 evidence_mass_score
+ 0.20 independent_hits_score
+ 0.15 pattern_diversity_score
+ 0.15 context_confidence
+ 0.10 sense_confidence
+ 0.10 proposition_coverage
```

Danach Ambiguity Modifier.

Werte:
```text
0.00 – 1.00
```

---

# 15. Thresholds

Empfehlung:

```text
0.00–0.44 = NOT_ASSESSABLE
0.45–0.64 = WEAK
0.65–0.79 = ASSESSABLE
0.80–1.00 = STRONG
```

Standard-UI:

- `NOT_ASSESSABLE` → keine Zahl
- `WEAK` → optional „Tendenz“
- `ASSESSABLE` → Wert anzeigen
- `STRONG` → Wert anzeigen

---

# 16. Weak Mode

Bei `WEAK`:

Nicht:
> „Freier Wille 42 %“

Besser:
> „Es gibt eine leichte Tendenz in Richtung …“

Optional kann der Prozentwert nur im Fach-/Adminmodus erscheinen.

---

# 17. Contradiction Handling

Wenn starke positive und negative Contributions gleichzeitig vorhanden sind:

```text
contradiction_ratio =
min(pos_mass, neg_mass) / max(pos_mass, neg_mass)
```

Bei hohem Wert:

- Dimension bleibt bewertbar,
- Confidence kann sinken,
- UI zeigt „gemischte Tendenz“.

Dies ist etwas anderes als fehlende Evidenz.

---

# 18. Neutral versus nicht bewertbar

## Neutral
Es existiert belastbare Evidenz in beide Richtungen oder echte mittlere Ausprägung.

```text
score ≈ 50 %
assessable = true
```

## Nicht bewertbar
Es fehlt relevante Evidenz.

```text
score = null
assessable = false
```

Diese Trennung ist fundamental.

---

# 19. Resonanz und Assessability

Resonanz allein sollte im Defaultmodus **keine Dimension assessable machen**.

Regel:

```text
if semantic_evidence_mass == 0
and only_resonance_evidence == true
and profile != DEEP_RESONANCE:
    assessable = false
```

Im vertieften Resonanzprofil kann ein eigener Resonanzwert sichtbar werden.

---

# 20. WingScore Gate

WingScore nur, wenn:

```text
assessable_dimension_count >= 3
```

und:

```text
average_assessability >= 0.65
```

Optional zusätzlich:

```text
no critical dimension exclusively WEAK
```

---

# 21. Corporate Gate

Corporate-Profil:

- strengere Assessability für persönliche Dimensionen,
- keine Interpretation aus schwachen Kontextsignalen,
- keine Resonanz zur Herstellung von Assessability,
- kein WingScore-Label im UI.

Canonical metric kann intern weiter existieren.

---

# 22. Private Gate

Private-/Deep-Reflection-Profil:

- `WEAK`-Tendenzen dürfen optional sichtbar sein,
- Resonanzhinweise können vertieft werden,
- WingScore sichtbar,
- Nutzer kann Unsicherheitsdetails öffnen.

---

# 23. Explainability

Wenn nicht bewertbar:

Beispiel:
> „Für Verbindung enthält dieser Ausschnitt noch zu wenig sprachliche Hinweise.“

Wenn schwach:
> „Es gibt einzelne Hinweise, die Einordnung bleibt jedoch kontextabhängig.“

Wenn gemischt:
> „Der Text enthält sowohl verbindende als auch trennende Sprachbewegungen.“

---

# 24. Admin Trace

```yaml
AssessabilityTrace:
  dimension
  evidence_mass
  independent_hits
  pattern_diversity
  context_confidence
  sense_confidence
  proposition_coverage
  ambiguity_modifier
  contradiction_ratio
  final_assessability
  state
```

---

# 25. Beispiel G22

> „Hast du Geld?“

Klarheit:
- Proposition erkannt
- Frage klar
- Akteur/Adressat erkennbar
→ ggf. assessable

Verbindung:
- keine ausreichende Evidenz
→ NOT_ASSESSABLE

Freier Wille:
- keine semantische Evidenz
→ NOT_ASSESSABLE

Homophonie:
- Resonanzhinweis
→ macht keine Dimension assessable.

---

# 26. Beispiel G08

> „Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.“

Verbindung:
- ACKNOWLEDGEMENT
- RESPECTFUL_BOUNDARY
→ STRONG

Wertschätzung:
- ACKNOWLEDGEMENT
- NEED_RESPECT
→ STRONG

Klarheit:
- Proposition
- klare Grenze
→ STRONG

Freier Wille:
- klare eigene Position
→ ASSESSABLE/STRONG

---

# 27. Beispiel G19

> „Wir haben ein technisches Problem mit der Schnittstelle.“

Klarheit:
- propositionale Sachbeschreibung
→ ASSESSABLE

Wertschätzung:
- keine relevante Personen-/Beziehungsevidenz
→ NOT_ASSESSABLE

Offenheit:
- `Problem` allein reicht nicht
→ NOT_ASSESSABLE

---

# 28. Implementierungs-Pseudocode

```text
for dimension in dimensions:
    evidence = collect_contributions(dimension)

    if evidence.empty:
        return NOT_ASSESSABLE

    features = calculate_assessability_features(evidence, context)

    assessability =
        weighted_sum(features)
        * ambiguity_modifier

    if only_resonance and profile_not_deep:
        assessability = 0

    state = classify(assessability)

    if state == NOT_ASSESSABLE:
        score = null
```

---

# 29. Parameter Registry

Admin-editierbar:

```text
assessability_weight_evidence_mass
assessability_weight_independent_hits
assessability_weight_pattern_diversity
assessability_weight_context
assessability_weight_sense
assessability_weight_proposition

threshold_weak
threshold_assessable
threshold_strong

ambiguity_modifier_low
ambiguity_modifier_medium
ambiguity_modifier_high
```

---

# 30. Guardrails

Engine-locked:

- fehlende Evidenz darf nicht zu 50 % werden,
- Resonanz allein macht im Standard keine Kern-Dimension bewertbar,
- `null` bleibt gültiges Ergebnis,
- WingScore darf nicht aus weniger als Mindestdimensionen entstehen.

---

# 31. Testpflichtfälle

- G01: VER/WER n/a
- G05: WER nicht künstlich durch „nicht“ erzeugen
- G12: VER/WER kontextabhängig
- G15: VER/WER n/a
- G19: WER/OFF n/a
- G22/G23: Resonanz ohne Dimensions-Assessability
- G25: `frei` nicht automatisch FW
- G26: `soll` Hörensagen nicht normativ scoren

---

# 32. Leitgedanke

> **Ein fehlender Wert ist manchmal die präziseste Analyse, die ein verantwortungsvolles System geben kann.**
