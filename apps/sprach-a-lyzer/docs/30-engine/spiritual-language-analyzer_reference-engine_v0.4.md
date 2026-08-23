# Spiritual Language Analyzer
## Reference Engine v0.4 – Definition of Done Implementation

**Status:** ausführbare Referenzengine  
**Version:** 0.4  
**Datum:** 19. August 2026

---

# 1. Implementierte Definition of Done

- [x] Proposition Graph implementiert
- [x] SenseCandidate-Ranking implementiert
- [x] Ambiguity Resolver erweitert
- [x] TargetType operationalisiert
- [x] ExpectationSource operationalisiert
- [x] Assessability v0.2
- [x] Klarheit Feature-Modell
- [x] gezielte Positive-Pattern-Kalibrierung
- [x] Overlap-/Deduplication-Regel für composed patterns
- [x] Golden-Diff v0.4
- [x] Regression der Schutzfälle geprüft

---

# 2. Neue Architekturbausteine

## Proposition Graph

Jede Teilaussage erhält eine ID. Konnektoren erzeugen Kanten:

```text
P0 --CONTRAST--> P1
P1 --CAUSE--> P2
P2 --CONDITION--> P3
```

`aber`, `trotzdem`, `weil`, `deshalb`, `wenn`, `falls` werden dadurch nicht bloß als Wörter, sondern als Beziehungen behandelt.

## SenseCandidate Ranking

Ein Sense wird nicht mehr nur durch eine harte Regex ausgewählt.

Startfaktoren:

```text
lexical_match
phrase_fit
syntax_fit
domain_fit
register_fit
discourse_fit
collocation_fit
context_fit
```

Top-Sense und Abstand zum zweitbesten Kandidaten bestimmen die Ambiguität.

## TargetType

```text
PERSON
SELF
BEHAVIOR
PROCESS
EVENT
UNKNOWN
```

## ExpectationSource

```text
LAW
INTERNALIZED
CULTURE_OR_UNSPECIFIED
OTHER_PERSON
SELF_OR_INTERNALIZED
UNSPECIFIED
```

---

# 3. Assessability v0.2

Neue Zustände:

```text
NOT_ASSESSABLE
WEAK
ASSESSABLE
STRONG
```

`WEAK` kann intern einen Wert besitzen, wird im Standard-UI später aber eher als Tendenz als als Prozentwert dargestellt.

---

# 4. Klarheit Feature-Modell

Klarheit wird aus strukturellen Features bestimmt:

```text
actor
predicate
target/reference
time
boundary
decision
```

Ein Personenlabel wie:

> „Du bist das Problem.“

ist deshalb nicht automatisch hochklar, nur weil es eindeutig formuliert ist.

---

# 5. Positive Pattern Kalibrierung

Gezielt verstärkt wurden insbesondere:

```text
RESPECTFUL_BOUNDARY
AGENCY_RECOVERY
CONDITIONAL_OPENING
LEARNING_RECOVERY
```

Keine globale Positivverstärkung.

---

# 6. Overlap / Deduplication

Wenn ein composed pattern greift, werden seine Basismarker teilweise unterdrückt.

Beispiel:

```text
ACKNOWLEDGEMENT
+ CLEAR_BOUNDARY
→ RESPECTFUL_BOUNDARY
```

`ACKNOWLEDGEMENT` und `CLEAR_BOUNDARY` werden dann nicht zusätzlich vollständig doppelt gescort.

---

# 7. Spoken Dictation Mode

Neu vorbereitet:

```text
input_mode = SPOKEN_DICTATION
```

Dabei können später berücksichtigt werden:

- Wiederholungen
- spontane Selbstkorrekturen
- Füllwörter
- Modalverben
- Generalisierungen
- fragmentarische Syntax
- Prosodie, sobald Audio direkt verfügbar ist

Wichtig:

> **Gesprochene Sprache erhält keinen pauschalen Negativ- oder Drastikfaktor.**

Stärkere Ausschläge entstehen nur, wenn tatsächlich mehr relevante Muster erkannt werden.

---

# 8. Mobile UX-Idee

Mobile Nutzer können aktiv eingeladen werden:

> **Sprich statt zu tippen**  
> Wenn du möchtest, diktiere einfach so, wie du im Alltag wirklich sprichst. Spontane Sprache kann andere Muster sichtbar machen als sorgfältig formulierter Text.

Für besonders tiefe Analyse wäre später direkte Audioaufnahme besser als reine System-Diktierung, weil Telefon-Diktierfunktionen Füllwörter, Pausen oder Interpunktion teilweise verändern können.

---

# 9. Golden Result v0.4

Geprüfte Erwartungen: **94**

Im Zielkorridor: **27 (28.7 %)**

Numerisch bewertete Nicht-n/a-Ziele: **68**

Treffer innerhalb der numerisch bewerteten Ziele: **22/68 (32.4 %)**

Gap-Verteilung:

```text
MISSING: 21
TOO_LOW: 38
TOO_HIGH: 8
OVER_ASSESSED: 0
```

v0.3 → v0.4:

```text
GAP → PASS: 8
PASS → GAP: 11
```

---

# 10. Leitgedanke

> **v0.4 verschiebt die Referenzengine von einer Sammlung guter Einzelregeln hin zu einem kleinen, erklärbaren Sprachmodell mit Aussagen, Bedeutungsalternativen, Kontextquellen und Unsicherheitslogik.**
