# Spiritual Language Analyzer
## Positive PatternLibrary v0.2

**Status:** MVP-Fachbibliothek  
**Version:** 0.2  
**Datum:** 19. August 2026

---

# 1. Ziel

Die bisherige Referenzengine erkennt belastende Muster bereits besser als stärkende Sprachbewegungen.

Diese Bibliothek gleicht diese Schieflage aus.

Grundsatz:

> **Stärkende Sprache ist nicht einfach das Gegenteil belastender Wörter. Sie zeigt eigenständige Muster wie Anerkennung, Wahl, Verantwortung, klare Grenze, Lernhaltung und Öffnung.**

---

# 2. Pattern-Kategorien

## A – Selbstwirksamkeit

### CHOICE_LANGUAGE
Beispiele:
- ich entscheide
- ich wähle
- ich möchte
- ich kann
- mein nächster Schritt

Primäre Dimensionen:
- Wirksamkeit ↑
- Freier Wille ↑

Sekundär:
- Klarheit ↑

### RESPONSIBILITY_LANGUAGE
Beispiele:
- ich übernehme Verantwortung
- ich kümmere mich darum
- ich prüfe meinen Anteil
- ich entscheide, wie ich weiter vorgehe

Dimensionen:
- Wirksamkeit ↑
- Freier Wille ↑
- Klarheit ↑

Wichtig: Verantwortung ≠ Schuldübernahme.

### REALISTIC_LIMIT
Beispiele:
- das kann ich nicht beeinflussen
- dieser Teil liegt außerhalb meines Einflusses
- darauf habe ich keinen direkten Zugriff

Allein kein positiver Marker.

Wird positiv relevant, wenn anschließend eigener Handlungsspielraum benannt wird.

Beispiel:
> „Das kann ich nicht beeinflussen, aber ich kann meinen nächsten Schritt wählen.“

Relation:
```text
REALISTIC_LIMIT
+ CHOICE_LANGUAGE
→ AGENCY_RECOVERY
```

### AGENCY_RECOVERY
Pattern für reale Begrenzung + eigener Gestaltungsspielraum.

Dimensionen:
- Wirksamkeit ↑↑
- Freier Wille ↑
- Offenheit ↑

---

# 3. Verbindung

### ACKNOWLEDGEMENT
Beispiele:
- ich verstehe deinen Punkt
- ich sehe, dass dir das wichtig ist
- danke für deine Rückmeldung
- ich kann nachvollziehen, dass ...

Dimensionen:
- Verbindung ↑
- Wertschätzung ↑

### PERSPECTIVE_RECOGNITION
Beispiele:
- aus deiner Sicht
- ich sehe deine Perspektive
- wir bewerten das unterschiedlich
- beide Sichtweisen haben unterschiedliche Ausgangspunkte

Dimensionen:
- Verbindung ↑
- Offenheit ↑
- Wertschätzung ↑

### COLLABORATIVE_LANGUAGE
Beispiele:
- gemeinsam
- miteinander
- lass uns prüfen
- wie können wir ...
- was brauchen wir dafür?

Dimensionen:
- Verbindung ↑
- Offenheit ↑
- Wirksamkeit ↑

### REPAIR_LANGUAGE
Beispiele:
- ich möchte das Missverständnis klären
- lass uns noch einmal anschauen, wo wir auseinanderliegen
- ich möchte meinen Anteil korrigieren

Dimensionen:
- Verbindung ↑
- Wirksamkeit ↑
- Offenheit ↑

---

# 4. Wertschätzung

### PERSON_BEHAVIOR_SEPARATION
Beispiel:
> „Die Vereinbarung wurde nicht eingehalten.“

statt:
> „Auf dich ist nie Verlass.“

Merkmal:
Verhalten oder Ereignis wird beschrieben, nicht die Person als Ganzes etikettiert.

Dimensionen:
- Wertschätzung ↑
- Klarheit ↑

### NEED_RESPECT
Beispiele:
- mir ist wichtig ...
- ich nehme wahr, dass dir ... wichtig ist
- ich respektiere deine Entscheidung

Dimensionen:
- Wertschätzung ↑
- Verbindung ↑

### SELF_APPRECIATION
Beispiele:
- ich darf mir Zeit geben
- ich habe einen Fehler gemacht, ohne mich darauf zu reduzieren
- ich erkenne an, was ich bereits geschafft habe

Dimensionen:
- Wertschätzung ↑
- Wirksamkeit ↑
- Offenheit ↑

---

# 5. Klarheit

### CLEAR_BOUNDARY
Beispiele:
- nein
- ich möchte das nicht
- das kommt für mich nicht infrage
- ich stimme dem nicht zu

Dimensionen:
- Klarheit ↑↑
- Freier Wille ↑
- Wirksamkeit ↑

Wichtig: Nicht automatisch Verbindung ↓.

### CLEAR_REQUEST
Beispiele:
- bitte sende mir die Unterlagen bis Freitag
- ich wünsche mir, dass ...
- kannst du bis 15 Uhr ...?

Dimensionen:
- Klarheit ↑
- Verbindung ↑/neutral
- Freier Wille ↑ gegenüber verdecktem Druck

### CLEAR_DECISION
Beispiele:
- ich entscheide mich für ...
- wir haben entschieden ...
- meine Entscheidung ist ...

Dimensionen:
- Klarheit ↑
- Wirksamkeit ↑
- Freier Wille ↑

### SPECIFICITY
Merkmale:
- Akteur eindeutig
- Handlung eindeutig
- Referent eindeutig
- Zeitpunkt/Frist konkret
- Grenze/Entscheidung explizit

Dimension:
- Klarheit ↑

---

# 6. Freier Wille

### CONSENT_LANGUAGE
Beispiele:
- wenn du möchtest
- wenn es für dich passt
- du kannst entscheiden
- du darfst dir Zeit nehmen

Dimensionen:
- Freier Wille ↑
- Wertschätzung ↑
- Verbindung ↑

### OPTIONALITY
Beispiele:
- eine Möglichkeit wäre
- du könntest
- wir können A oder B prüfen

Dimensionen:
- Freier Wille ↑
- Offenheit ↑

### OWNED_COMMITMENT
Beispiele:
- ich sage zu
- ich entscheide mich, diese Verpflichtung zu erfüllen
- ich übernehme diese Aufgabe

Dimensionen:
- Freier Wille ↑
- Wirksamkeit ↑
- Klarheit ↑

---

# 7. Offenheit

### OPENING_LANGUAGE
Beispiele:
- was wäre noch möglich?
- welche weitere Perspektive gibt es?
- was könnten wir verändern?
- ich bin offen für ...
- lass uns prüfen ...

Dimensionen:
- Offenheit ↑↑
- Wirksamkeit ↑
- Verbindung ↑ bei gemeinsamem Bezug

### LEARNING_FRAME
Beispiele:
- was lernen wir daraus?
- der Fehler zeigt uns ...
- beim nächsten Versuch können wir ...
- was würden wir anders machen?

Dimensionen:
- Offenheit ↑
- Wirksamkeit ↑
- Wertschätzung ↑/neutral

### CONSTRUCTIVE_DIFFERENTIATION
Beispiele:
- wir sollten zwischen A und B unterscheiden
- das gilt in diesem Kontext, aber nicht zwingend im anderen

Dimensionen:
- Klarheit ↑
- Offenheit ↑
- Verbindung ↑/neutral

### CONDITIONAL_OPENING
Beispiele:
- solltest du Fragen haben, ...
- wenn du möchtest, ...
- falls eine andere Idee auftaucht, ...

Dimensionen:
- Offenheit ↑
- Verbindung ↑
- Freier Wille ↑/neutral

---

# 8. Mehrdimensionale positive Patterns

## RESPECTFUL_BOUNDARY
```text
ACKNOWLEDGEMENT
+ CLEAR_BOUNDARY
```

Beispiel:
> „Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.“

Erwartung:
- Verbindung ↑
- Wertschätzung ↑
- Klarheit ↑
- Freier Wille ↑

## EMOTION_PLUS_AGENCY
```text
EMOTION_ACKNOWLEDGEMENT
+ OPENING_LANGUAGE
```

Beispiel:
> „Ich habe große Angst und möchte herausfinden, was mir jetzt helfen kann.“

Erwartung:
- emotionale Benennung nicht bestrafen
- Wirksamkeit ↑
- Offenheit ↑
- Klarheit ↑

## LEARNING_RECOVERY
```text
ERROR_DESCRIPTION
+ LEARNING_FRAME
```

Beispiel:
> „Der Fehler zeigt uns, was wir beim nächsten Versuch verändern können.“

Erwartung:
- Offenheit ↑
- Wirksamkeit ↑
- Wertschätzung neutral bis ↑

---

# 9. Pattern-Komposition

Patterns dürfen neue Patterns erzeugen.

Beispiel:
```text
REALISTIC_LIMIT
+ CHOICE_LANGUAGE
→ AGENCY_RECOVERY
```

oder:
```text
ACKNOWLEDGEMENT
+ CLEAR_BOUNDARY
→ RESPECTFUL_BOUNDARY
```

Damit vermeiden wir hunderte Spezialregeln.

---

# 10. Default Contributions v0.2

Noch keine endgültigen Werte.

| Pattern | WIR | VER | WER | KLA | FW | OFF |
|---|---:|---:|---:|---:|---:|---:|
| CHOICE_LANGUAGE | +18 | 0 | 0 | +8 | +22 | +8 |
| RESPONSIBILITY_LANGUAGE | +20 | +4 | +4 | +10 | +14 | +8 |
| ACKNOWLEDGEMENT | 0 | +18 | +16 | +4 | 0 | +6 |
| COLLABORATIVE_LANGUAGE | +10 | +18 | +8 | +6 | +4 | +14 |
| PERSON_BEHAVIOR_SEPARATION | 0 | +6 | +18 | +14 | 0 | +6 |
| CLEAR_BOUNDARY | +14 | 0 | +4 | +24 | +20 | 0 |
| CONSENT_LANGUAGE | +4 | +12 | +12 | +6 | +20 | +12 |
| OPENING_LANGUAGE | +10 | +10 | +4 | +6 | +6 | +22 |
| LEARNING_FRAME | +16 | +8 | +8 | +8 | +4 | +20 |
| CONSTRUCTIVE_DIFFERENTIATION | +4 | +8 | +6 | +18 | +2 | +14 |
| CONDITIONAL_OPENING | +4 | +14 | +10 | +8 | +12 | +18 |
| AGENCY_RECOVERY | +24 | +4 | +6 | +12 | +20 | +18 |
| RESPECTFUL_BOUNDARY | +8 | +20 | +18 | +22 | +18 | +4 |

Diese Werte gehören ins Admin-Panel und werden über Golden Tests kalibriert.

---

# 11. Guardrails

Positive Pattern-Erkennung darf nicht zu „Positivitätszwang“ führen.

Insbesondere:
- Angst
- Trauer
- Kritik
- Ablehnung
- Grenze
- Fehler
- reale Einschränkung

sind nicht automatisch belastende Sprache.

---

# 12. Leitgedanke

> **Stärkende Sprache ist nicht Schönfärberei. Sie verbindet Realität mit Klarheit, Würde, Wahl und Handlungsspielraum.**
