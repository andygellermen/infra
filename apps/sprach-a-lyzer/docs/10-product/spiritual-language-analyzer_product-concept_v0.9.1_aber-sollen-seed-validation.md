# Spiritual Language Analyzer
## Product Concept v0.9.1 – Ergänzende Seed-Validierung: „Ja, aber …“ und „sollen“

**Status:** MVP-Modellvalidierung / Ergänzung zu v0.9  
**Version:** 0.9.1  
**Datum:** 18. August 2026  
**Arbeitstitel:** Spiritual Language Analyzer (SLA)

---

# 1. Ziel dieser Ergänzung

Diese Version ergänzt die Seed-Validierung v0.9 um zwei besonders alltagsrelevante Testfelder:

1. die Konstruktion **„Ja, aber …“**
2. das Modalverb **„sollen“**

Beide Fälle sind für die spätere Analyse besonders wertvoll, weil sie nicht nur lexikalische Bedeutung transportieren, sondern stark von:

- Gesprächsverlauf,
- Sprecherhaltung,
- Beziehung,
- Betonung,
- sozialer Erwartung,
- impliziter Fremdsteuerung,
- und pragmatischer Funktion

abhängen.

---

# 2. Testfall A – „aber“ vertieft

## 2.1 Grundfunktion

„aber“ markiert typischerweise einen Kontrast, Einwand, eine Einschränkung oder Korrektur.

Es ist deshalb nicht grundsätzlich negativ.

Beispiele:

> „Das ist schwierig, aber lösbar.“

> „Ich stimme dir zu, aber bei diesem Punkt sehe ich es anders.“

Hier kann „aber“ Klarheit und Differenzierung sogar erhöhen.

---

# 3. Sonderfall „Ja, aber …“

Die Phrase:

> **„Ja, aber …“**

verdient einen eigenen Phrase-/Pragmatik-Eintrag.

Warum?

Das einleitende **„Ja“** signalisiert formal Zustimmung.

Das anschließende **„aber“** kann diese Zustimmung jedoch:

- relativieren,
- begrenzen,
- teilweise aufheben,
- oder als bloße Gesprächskonvention erscheinen lassen.

Damit entsteht häufig eine kommunikative Spannung zwischen:

> **formaler Zustimmung**

und

> **tatsächlichem Einwand**.

---

# 4. Typische Funktionen von „Ja, aber …“

## JAB1 – echte Teilzustimmung

> „Ja, aber beim zweiten Punkt sehe ich es anders.“

Hier ist die Zustimmung real und der Einwand klar abgegrenzt.

### mögliche Wirkung

- Klarheit ↑
- Verbindung neutral bis ↑
- Offenheit neutral

---

## JAB2 – scheinbare Zustimmung

> „Ja, aber das funktioniert doch sowieso nicht.“

Das „Ja“ kann kommunikativ nahezu bedeutungslos werden.

### mögliche Wirkung

- Verbindung ↓
- Offenheit ↓
- Klarheit mittel
- möglicher Widerstand ↑

---

## JAB3 – Abwehr

> „Ja, aber du verstehst meine Situation nicht.“

Hier kann die Phrase einen Gegenangriff oder eine defensive Bewegung einleiten.

### mögliche Wirkung

- Verbindung ↓
- Offenheit ↓
- Wertschätzung kontextabhängig

---

## JAB4 – Selbstbegrenzung

> „Ja, aber ich kann das nicht.“

Hier wird eine zuvor mögliche Perspektive unmittelbar begrenzt.

### mögliche Wirkung

- Wirksamkeit ↓
- Offenheit ↓
- Freier Wille ↓ oder kontextabhängig

---

## JAB5 – konstruktive Differenzierung

> „Ja, aber wir sollten zwischen zwei Situationen unterscheiden.“

Hier unterstützt die Konstruktion Differenzierung.

### mögliche Wirkung

- Klarheit ↑
- Offenheit ↑
- Verbindung neutral bis ↑

---

# 5. Warum „Ja, aber“ nicht pauschal negativ bewertet werden darf

Die Konstruktion kann:

- Widerstand,
- Relativierung,
- Abwehr

anzeigen.

Sie kann aber ebenso:

- Präzision,
- Differenzierung,
- gesunde Abgrenzung

ermöglichen.

Daraus folgt:

> **„Ja, aber“ ist ein hoch kontextabhängiges Diskursmuster und kein Negativmarker.**

---

# 6. Neue PhraseDefinition

```yaml
Phrase:
  canonical_text: "ja, aber"
  phrase_type: DISCOURSE_PATTERN
  pragmatic_functions:
    - PARTIAL_AGREEMENT
    - OBJECTION
    - DISCOUNTING
    - DEFENSIVE_SHIFT
    - DIFFERENTIATION
```

---

# 7. Relevante Relations

```text
"ja" PRECEDES "aber"

"aber" MAY_REDUCE_PREVIOUS_AGREEMENT

"aber" DISCOURSE_CONTRAST previous_proposition next_proposition
```

Für die Taxonomie empfiehlt sich zusätzlich:

```text
DISCOUNTS_PREVIOUS_PROPOSITION
```

Diese Relation ist präziser als ein allgemeines `REDUCES`.

---

# 8. Propositionale Analyse

Beispiel:

> „Ja, das ist eine gute Idee, aber wir haben dafür keine Zeit.“

Proposition A:

> Das ist eine gute Idee.

Proposition B:

> Wir haben dafür keine Zeit.

Relation:

```text
DISCOURSE_CONTRAST
```

Zusätzlich kann geprüft werden:

> Wird Proposition A durch Proposition B kommunikativ faktisch entwertet?

Wenn ja:

```text
DISCOUNTS_PREVIOUS_PROPOSITION
```

---

# 9. PatternClasses für „Ja, aber“

```text
PARTIAL_AGREEMENT
DEFENSIVE_OBJECTION
DISCOUNTING
SELF_LIMITATION
CONSTRUCTIVE_DIFFERENTIATION
```

---

# 10. Reflexionsimpulse für „Ja, aber“

Der Analyzer könnte je nach Kontext fragen:

> „Bleibt deine Zustimmung auch nach dem ‚aber‘ noch bestehen?“

> „Möchtest du zustimmen und ergänzen – oder eigentlich widersprechen?“

> „Wäre ‚und gleichzeitig‘ hier näher an deiner eigentlichen Aussage?“

Wichtig:

„und“ ist **nicht automatisch besser** als „aber“.

Beispiel:

> „Ich verstehe deine Sicht, und ich lehne den Vorschlag ab.“

kann verbindender wirken.

Aber:

> „Ich stimme zu, und du liegst trotzdem falsch.“

löst die kommunikative Spannung nicht automatisch.

---

# 11. Alternative Konnektoren als Reflexionsoption

Je nach Bedeutung:

- und
- gleichzeitig
- zugleich
- dennoch
- trotzdem
- allerdings
- während
- hingegen
- auf der anderen Seite

Diese Alternativen besitzen jeweils eigene pragmatische Wirkungen.

---

# 12. Testfall B – „sollen“

## Typ

Modalverb

„sollen“ ist für das SLA besonders interessant, weil es häufig eine **Quelle außerhalb des Sprechers** voraussetzt.

Es kann ausdrücken:

- Auftrag,
- Erwartung,
- Empfehlung,
- soziale Norm,
- moralische Forderung,
- indirekte Information,
- Vermutung / Hörensagen.

---

# 13. Usage Senses von „sollen“

## S1 – Auftrag / Anweisung

> „Du sollst morgen um acht Uhr kommen.“

Quelle:

eine andere Person oder Institution.

### Dimensionen

- Freier Wille ↓, kontextabhängig
- Klarheit ↑
- Verbindung kontextabhängig

---

## S2 – Empfehlung

> „Du solltest mehr schlafen.“

Kann fürsorglich gemeint sein.

Kann aber auch bevormundend wirken.

### Dimensionen

- Freier Wille leicht ↓ bis neutral
- Wertschätzung kontextabhängig
- Verbindung kontextabhängig

---

## S3 – soziale Norm

> „So etwas sollte man nicht tun.“

Quelle:

gesellschaftliche oder moralische Erwartung.

### Dimensionen

- Freier Wille ↓
- Offenheit ↓
- Klarheit mittel

---

## S4 – Selbstanforderung

> „Ich sollte endlich produktiver sein.“

Hier wird eine äußere oder internalisierte Norm auf das Selbst angewendet.

### mögliche Wirkung

- Freier Wille ↓
- Wirksamkeit ↓ oder ambivalent
- Wertschätzung ↓ bei Selbstabwertung
- Offenheit ↓

---

## S5 – Erwartung / Zielwert

> „Die Lieferung sollte morgen eintreffen.“

Hier bedeutet „sollte“ eher Erwartung oder Wahrscheinlichkeit.

### Dimensionen

- Klarheit mittel
- Freier Wille irrelevant

---

## S6 – Hörensagen / indirekte Information

> „Er soll sehr reich sein.“

Bedeutung:

Es wird berichtet oder behauptet, ohne volle Gewissheit.

### Dimensionen

- Klarheit / Evidenzrelevanz
- Freier Wille irrelevant

---

## S7 – hypothetische Empfehlung

> „Solltest du Hilfe brauchen, melde dich.“

Hier ist „sollen“ Teil einer konditionalen Konstruktion.

### mögliche Wirkung

- Verbindung ↑
- Offenheit ↑
- Freier Wille neutral bis ↑

---

# 14. Warum „sollen“ nicht wie „müssen“ behandelt werden darf

„müssen“ markiert typischerweise Notwendigkeit.

„sollen“ markiert häufig:

> **eine erwartete, gewünschte oder übermittelte Handlung aus einer anderen Quelle.**

Das macht für das Coaching eine neue Frage relevant:

> **Wessen Erwartung spricht hier eigentlich?**

---

# 15. Neue Kontextdimension: Expectation Source

Für Modalverben wie „sollen“ empfiehlt sich ein neues Analyseattribut:

```text
EXPECTATION_SOURCE
```

Mögliche Werte:

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

# 16. Warum EXPECTATION_SOURCE wichtig ist

Beispiel:

> „Ich sollte mehr Sport machen.“

Die Quelle kann sein:

- eigener Wunsch,
- ärztliche Empfehlung,
- gesellschaftliche Norm,
- Partner,
- internalisierte Selbstanforderung.

Diese Varianten können sprachlich identisch aussehen, aber unterschiedlich wirken.

---

# 17. Testphrase „ich sollte“

Beispiel:

> „Ich sollte endlich mehr leisten.“

Marker:

- sollte
- endlich
- mehr

Mögliche PatternClasses:

```text
INTERNALIZED_EXPECTATION
SELF_PRESSURE
COMPARATIVE_ESCALATION
URGENCY
```

Mögliche Dimensionen:

- Freier Wille ↓
- Wirksamkeit ambivalent
- Wertschätzung ↓
- Offenheit ↓

---

# 18. Testphrase „du solltest“

Beispiel:

> „Du solltest wirklich vernünftiger sein.“

Mögliche PatternClasses:

```text
EXTERNAL_EXPECTATION
PERSON_EVALUATION
ADVICE
POTENTIAL_PATRONIZING
```

Mögliche Dimensionen:

- Freier Wille ↓
- Verbindung ↓
- Wertschätzung ↓
- Klarheit ↑

---

# 19. Testphrase „man sollte“

Beispiel:

> „Man sollte immer dankbar sein.“

Besonderheiten:

- unklare Quelle,
- soziale Norm,
- Generalisierung durch „immer“.

PatternClasses:

```text
SOCIAL_NORM
GENERALIZATION
UNSPECIFIED_AUTHORITY
```

Dimensionen:

- Freier Wille ↓
- Offenheit ↓
- Klarheit teilweise ↓, weil die Quelle unklar bleibt

---

# 20. Testphrase „solltest du …“

Beispiel:

> „Solltest du Fragen haben, melde dich.“

Hier wirkt „sollen“ nicht primär normativ.

Pragmatic Function:

```text
CONDITIONAL_OPENING
```

Mögliche Wirkung:

- Verbindung ↑
- Offenheit ↑
- Freier Wille neutral

Dies bestätigt erneut:

> **Phrase und Syntax schlagen Lexem.**

---

# 21. Testphrase „das soll …“

Beispiel:

> „Das soll helfen.“

Mögliche Bedeutungen:

- Zweck
- Behauptung
- Erwartung
- Skepsis, je nach Betonung

Beispiel:

> „Das soll helfen?“

kann Zweifel oder Ironie transportieren.

Damit wird Prosodie relevant.

---

# 22. Neue Modellanforderung: Prosody Cue

Für gesprochene Sprache können:

- Betonung,
- Tonhöhe,
- Pausen,
- Satzmelodie

die pragmatische Funktion verändern.

Für den MVP muss Prosodie noch nicht vollautomatisch analysiert werden.

Das Datenmodell sollte sie jedoch vorsehen:

```text
PROSODY_CUE
```

Mögliche Attribute:

```text
stress
intonation
pause
emphasis
question_contour
ironic_contour
confidence
```

---

# 23. Beispiel: „Das soll helfen.“

neutral:

> Aussage über Zweck oder Erwartung.

betont / skeptisch:

> „DAS soll helfen?“

kann Ironie oder Zweifel ausdrücken.

Der geschriebene Text allein kann dies nur begrenzt erkennen.

Daher:

> **Prosodie ist ein zukünftiger Confidence- und Pragmatikfaktor.**

---

# 24. Vergleich „müssen“ und „sollen“

| Merkmal | müssen | sollen |
|---|---|---|
| Kern | Notwendigkeit | Erwartung / Auftrag |
| typische Quelle | Situation / Pflicht | häufig extern |
| Freier Wille | häufig direkt relevant | häufig indirekt relevant |
| zentrale Frage | „Ist es wirklich notwendig?“ | „Wessen Erwartung ist das?“ |
| innere Variante | „Ich muss …“ | „Ich sollte …“ |
| soziale Norm | möglich | sehr typisch |
| Hörensagen | nein | ja |
| Empfehlung | selten | typisch |

---

# 25. Neue PatternClasses

Aus „sollen“ ergeben sich:

```text
EXTERNAL_EXPECTATION
INTERNALIZED_EXPECTATION
UNSPECIFIED_AUTHORITY
NORMATIVE_ADVICE
CONDITIONAL_OPENING
REPORTED_CLAIM
EXPECTED_OUTCOME
```

---

# 26. Aktualisierte Relation Taxonomy v1.1.1

Neu empfohlen:

```text
DISCOUNTS_PREVIOUS_PROPOSITION
HAS_EXPECTATION_SOURCE
HAS_PRAGMATIC_FUNCTION
HAS_PROSODY_CUE
```

Damit bleibt die Taxonomie weiterhin evolutionär versionierbar.

---

# 27. Aktualisiertes Kontextmodell

Zusätzlich:

```yaml
Context:
  expectation_source:
    - SELF
    - OTHER_PERSON
    - GROUP
    - INSTITUTION
    - LAW
    - CULTURE
    - UNSPECIFIED
    - INTERNALIZED
```

---

# 28. Scoring-Relevanz von „sollen“

Ein späterer Rohwert darf nicht einfach lauten:

```text
sollen = Freier Wille -20
```

Stattdessen:

```text
Sense
+ Phrase
+ Expectation Source
+ Target Type
+ Register
+ Negation
+ Intensifier
+ Syntax
+ Prosody Confidence
→ Dimensionsbeitrag
```

---

# 29. Beispielanalyse A

Text:

> „Ja, aber ich sollte das eigentlich längst können.“

Erkannte Elemente:

```text
ja, aber
ich sollte
eigentlich
längst
können
```

PatternClasses:

```text
DISCOUNTING
INTERNALIZED_EXPECTATION
HEDGING
TIME_PRESSURE
SELF_EVALUATION
```

Mögliche Dimensionen:

- Wirksamkeit ↓
- Freier Wille ↓
- Wertschätzung ↓
- Offenheit ↓
- Klarheit mittel

Dieser Satz eignet sich hervorragend als späterer Testfall für die Scoring-Engine.

---

# 30. Beispielanalyse B

Text:

> „Ja, aber du solltest verstehen, dass ich keine Wahl habe.“

Erkannte Elemente:

```text
ja, aber
du solltest
keine Wahl
```

Mögliche PatternClasses:

```text
DEFENSIVE_OBJECTION
EXTERNAL_EXPECTATION
PERCEIVED_NO_CHOICE
```

Dimensionen:

- Verbindung ↓
- Freier Wille ↓
- Offenheit ↓
- Klarheit ↑

Auch hier muss Kontext die Intensität bestimmen.

---

# 31. Beispielanalyse C

Text:

> „Solltest du eine andere Idee haben, bin ich offen dafür.“

PatternClasses:

```text
CONDITIONAL_OPENING
OPENING_LANGUAGE
```

Dimensionen:

- Offenheit ↑
- Verbindung ↑
- Freier Wille ↑/neutral
- Wertschätzung ↑
- Klarheit ↑

Dies demonstriert, warum „sollen“ keinen statischen Negativwert besitzen darf.

---

# 32. Ergebnis der Ergänzung

Die beiden Testfelder bestätigen das bestehende Modell, führen aber zu drei wertvollen Ergänzungen:

1. **ExpectationSource**
2. **ProsodyCue**
3. **DISCOUNTS_PREVIOUS_PROPOSITION**

Damit können soziale Erwartung, indirekte Fremdsteuerung und konfliktbezogene „Ja, aber“-Konstruktionen wesentlich präziser beschrieben werden.

---

# 33. Status des Datenmodells nach v0.9.1

Das Modell bleibt:

> **fachlich ausreichend stabil für die erste Scoring-Engine.**

Die neuen Ergänzungen erweitern das Modell, verändern seine Grundarchitektur aber nicht mehr.

Das ist ein positives Reifezeichen.

---

# 34. Nächste Schritte Richtung v1.0

- [x] „aber“ vertieft
- [x] „Ja, aber“ als eigene Phrase modelliert
- [x] konstruktive und destruktive Varianten getrennt
- [x] propositionale Entwertung berücksichtigt
- [x] „sollen“ mit sieben Usage Senses modelliert
- [x] ExpectationSource eingeführt
- [x] soziale und internalisierte Erwartung getrennt
- [x] Konditionalgebrauch berücksichtigt
- [x] Hörensagen berücksichtigt
- [x] ProsodyCue als zukünftige Erweiterung ergänzt
- [x] Relation Taxonomy v1.1.1 ergänzt
- [ ] Datenmodell v1.0 konsolidieren
- [ ] Scoring-Logik v0.1
- [ ] Confidence-Modell
- [ ] erste vollständige Analysefälle
- [ ] danach Product Concept / Data Model v1.0

---

# 35. Leitgedanke

> **Ein „aber“ kann Verbindung relativieren – oder notwendige Differenzierung schaffen.  
> Ein „sollen“ kann Fremderwartung tragen – oder eine offene Möglichkeit einleiten.  
> Erst Kontext, Quelle und Funktion machen aus einem Wort eine Wirkung.**
