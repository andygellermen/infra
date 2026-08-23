# Spiritual Language Analyzer
## Automated Golden Corpus Simulation v0.1

**Status:** erste ausführbare Referenzsimulation  
**Datum:** 19. August 2026

---

# Ergebnisübersicht

| ID | WIR | VER | WER | KLA | FW | OFF | WingScore |
|---|---:|---:|---:|---:|---:|---:|---:|
| G01 | — | — | — | 53.4 | 35.9 | 46.3 | — |
| G02 | — | — | — | 53.4 | — | — | — |
| G03 | 58.1 | — | — | 68.9 | — | — | — |
| G04 | 37.9 | — | 30.8 | — | — | 42.9 | 36.9 |
| G05 | — | — | — | 53.4 | — | — | — |
| G06 | 42.7 | 38.1 | 36.8 | — | — | 41.5 | 39.7 |
| G07 | 67.3 | — | — | 65.9 | 74.9 | — | 69.4 |
| G08 | — | 56.1 | 59.6 | 64.0 | 58.1 | — | 59.6 |
| G09 | 42.7 | 43.9 | — | — | — | 36.6 | 40.8 |
| G10 | — | — | — | 54.9 | — | 53.6 | — |
| G11 | 46.2 | — | 45.1 | — | 40.7 | 46.3 | 43.9 |
| G12 | — | — | — | 53.4 | 46.5 | — | — |
| G13 | — | 56.0 | — | — | — | 58.3 | — |
| G14 | — | — | — | — | 57.7 | — | — |
| G15 | — | — | — | 58.5 | 37.9 | — | — |
| G16 | 59.6 | — | — | — | — | — | — |
| G17 | 52.3 | — | — | — | — | 41.5 | — |
| G18 | — | — | — | — | — | 59.3 | — |
| G19 | — | — | — | 56.0 | — | — | — |
| G20 | — | 33.4 | 30.3 | 59.6 | — | — | 40.6 |
| G21 | — | — | — | — | — | 59.3 | — |
| G22 | — | — | — | — | — | — | — |
| G23 | — | — | — | — | — | — | — |
| G24 | — | — | — | — | 39.7 | — | — |
| G25 | — | — | — | — | — | — | — |
| G26 | — | — | — | — | — | — | — |
| G27 | — | 59.6 | — | — | — | 59.3 | — |
| G28 | — | — | — | — | 39.7 | — | — |
| G29 | — | — | — | — | — | 59.3 | — |
| G30 | — | — | — | — | — | — | — |

---

# Detail-Traces

## G01 – Ich muss das heute unbedingt noch schaffen.

**Kontext:** `SELF_TALK`  
**Patterns:** INTERNAL_PRESSURE, URGENCY  

- **Klarheit: 53.4 %** (Confidence 0.68)
  - `specificity`: +8.0 × 0.68 – konkrete Aussage
- **Freier Wille: 35.9 %** (Confidence 0.97)
  - `must_default`: -20.0 × 0.84 – müssen
  - `must_unbedingt`: -8.0 × 0.80 – unbedingt
- **Offenheit: 46.3 %** (Confidence 0.75)
  - `must_unbedingt_open`: -8.0 × 0.75 – unbedingt
- **WingScore:** nicht ausreichend bewertbar

## G02 – Ich bin gesetzlich verpflichtet, die Unterlagen bis Freitag einzureichen.

**Kontext:** `LEGAL_ADMINISTRATIVE`  
**Patterns:** —  

- **Klarheit: 53.4 %** (Confidence 0.68)
  - `specificity`: +8.0 × 0.68 – konkrete Aussage
- **WingScore:** nicht ausreichend bewertbar

## G03 – Du musst sofort das Gebäude verlassen!

**Kontext:** `SAFETY`  
**Patterns:** SAFETY_DIRECTIVE, URGENCY  

- **Wirksamkeit: 58.1 %** (Confidence 0.82)
  - `safety_action`: +16.0 × 0.82 – müssen
- **Klarheit: 68.9 %** (Confidence 0.98)
  - `specificity`: +8.0 × 0.68 – konkrete Aussage
  - `safety_must`: +28.0 × 0.94 – müssen
- **WingScore:** nicht ausreichend bewertbar

## G04 – Ich bin einfach ein Versager.

**Kontext:** `SELF_TALK`  
**Patterns:** PREDICATIVE_LABELING, SELF_DEVALUATION  

- **Wirksamkeit: 37.9 %** (Confidence 0.90)
  - `self_devaluation`: -22.0 × 0.90 – Versager
- **Wertschätzung: 30.8 %** (Confidence 0.95)
  - `self_devaluation`: -34.0 × 0.95 – Versager
- **Offenheit: 42.9 %** (Confidence 0.82)
  - `self_devaluation`: -14.0 × 0.82 – Versager
- **WingScore:** 36.9

## G05 – Die Vereinbarung wurde heute nicht eingehalten.

**Kontext:** `WORKPLACE`  
**Patterns:** —  

- **Klarheit: 53.4 %** (Confidence 0.68)
  - `specificity`: +8.0 × 0.68 – konkrete Aussage
- **WingScore:** nicht ausreichend bewertbar

## G06 – Auf dich ist nie Verlass.

**Kontext:** `PRIVATE_CONVERSATION`  
**Patterns:** GENERALIZATION, PERSON_DEVALUATION  

- **Wirksamkeit: 42.7 %** (Confidence 0.84)
  - `absolutization`: -14.0 × 0.84 – Generalisierung
- **Verbindung: 38.1 %** (Confidence 0.88)
  - `person_devaluation`: -22.0 × 0.88 – auf dich ist ... Verlass
- **Wertschätzung: 36.8 %** (Confidence 0.90)
  - `person_devaluation`: -24.0 × 0.90 – auf dich ist ... Verlass
- **Offenheit: 41.5 %** (Confidence 0.86)
  - `absolutization`: -16.0 × 0.86 – Generalisierung
- **WingScore:** 39.7

## G07 – Nein. Ich möchte das nicht.

**Kontext:** `PRIVATE_CONVERSATION`  
**Patterns:** CHOICE_LANGUAGE, CLEAR_BOUNDARY  

- **Wirksamkeit: 67.3 %** (Confidence 0.98)
  - `choice_self`: +16.0 × 0.84 – ich ...
  - `clear_boundary`: +18.0 × 0.86 – Nein...möchte nicht
- **Klarheit: 65.9 %** (Confidence 0.94)
  - `clear_boundary`: +28.0 × 0.94 – Nein...möchte nicht
- **Freier Wille: 74.9 %** (Confidence 0.99)
  - `choice_self`: +24.0 × 0.90 – ich ...
  - `clear_boundary`: +24.0 × 0.92 – Nein...möchte nicht
- **WingScore:** 69.4

## G08 – Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.

**Kontext:** `PRIVATE_CONVERSATION`  
**Patterns:** CLEAR_BOUNDARY  

- **Verbindung: 56.1 %** (Confidence 0.82)
  - `connection_ack`: +12.0 × 0.82 – Anerkennung
- **Wertschätzung: 59.6 %** (Confidence 0.86)
  - `appreciation`: +18.0 × 0.86 – Anerkennung
- **Klarheit: 64.0 %** (Confidence 0.96)
  - `specificity`: +8.0 × 0.68 – konkrete Aussage
  - `clear_boundary`: +20.0 × 0.88 – nicht infrage
- **Freier Wille: 58.1 %** (Confidence 0.82)
  - `clear_boundary`: +16.0 × 0.82 – nicht infrage
- **WingScore:** 59.6

## G09 – Ja, aber das funktioniert doch sowieso nie.

**Kontext:** `WORKPLACE`  
**Patterns:** DEFENSIVE_OBJECTION, DISCOUNTING, GENERALIZATION  

- **Wirksamkeit: 42.7 %** (Confidence 0.84)
  - `absolutization`: -14.0 × 0.84 – Generalisierung
- **Verbindung: 43.9 %** (Confidence 0.82)
  - `ja_aber_discount`: -12.0 × 0.82 – ja, aber
- **Offenheit: 36.6 %** (Confidence 0.97)
  - `ja_aber_discount`: -10.0 × 0.82 – ja, aber
  - `absolutization`: -16.0 × 0.86 – Generalisierung
- **WingScore:** 40.8

## G10 – Ja, aber wir sollten zwischen den beiden Situationen unterscheiden.

**Kontext:** `WORKPLACE`  
**Patterns:** CONSTRUCTIVE_DIFFERENTIATION  

- **Klarheit: 54.9 %** (Confidence 0.78)
  - `ja_aber_diff`: +10.0 × 0.78 – ja, aber
- **Offenheit: 53.6 %** (Confidence 0.72)
  - `ja_aber_diff`: +8.0 × 0.72 – ja, aber
- **WingScore:** nicht ausreichend bewertbar

## G11 – Ich sollte längst weiter sein.

**Kontext:** `SELF_TALK`  
**Patterns:** INTERNALIZED_EXPECTATION, SELF_PRESSURE  

- **Wirksamkeit: 46.2 %** (Confidence 0.76)
  - `internalized_should`: -8.0 × 0.76 – ich sollte
- **Wertschätzung: 45.1 %** (Confidence 0.78)
  - `self_pressure_should`: -10.0 × 0.78 – längst/endlich
- **Freier Wille: 40.7 %** (Confidence 0.84)
  - `internalized_should`: -18.0 × 0.84 – ich sollte
- **Offenheit: 46.3 %** (Confidence 0.74)
  - `self_pressure_should`: -8.0 × 0.74 – längst/endlich
- **WingScore:** 43.9

## G12 – Du solltest heute etwas früher schlafen gehen.

**Kontext:** `FAMILY`  
**Patterns:** NORMATIVE_ADVICE  

- **Klarheit: 53.4 %** (Confidence 0.68)
  - `specificity`: +8.0 × 0.68 – konkrete Aussage
- **Freier Wille: 46.5 %** (Confidence 0.70)
  - `should_advice`: -8.0 × 0.70 – sollen
- **WingScore:** nicht ausreichend bewertbar

## G13 – Solltest du Fragen haben, melde dich jederzeit.

**Kontext:** `WORKPLACE`  
**Patterns:** CONDITIONAL_OPENING  

- **Verbindung: 56.0 %** (Confidence 0.80)
  - `conditional_sollen`: +12.0 × 0.80 – solltest du
- **Offenheit: 58.3 %** (Confidence 0.84)
  - `conditional_sollen`: +16.0 × 0.84 – solltest du
- **WingScore:** nicht ausreichend bewertbar

## G14 – Du darfst dir Zeit für die Entscheidung nehmen.

**Kontext:** `COACHING`  
**Patterns:** —  

- **Freier Wille: 57.7 %** (Confidence 0.78)
  - `permission`: +16.0 × 0.78 – dürfen
- **WingScore:** nicht ausreichend bewertbar

## G15 – Du darfst das nicht.

**Kontext:** `UNKNOWN`  
**Patterns:** PROHIBITION  

- **Klarheit: 58.5 %** (Confidence 0.86)
  - `prohibition_clear`: +16.0 × 0.86 – darf...nicht
- **Freier Wille: 37.9 %** (Confidence 0.90)
  - `prohibition`: -22.0 × 0.90 – darf...nicht
- **WingScore:** nicht ausreichend bewertbar

## G16 – Ich kann nicht alles beeinflussen, aber ich kann meinen nächsten Schritt wählen.

**Kontext:** `SELF_TALK`  
**Patterns:** —  

- **Wirksamkeit: 59.6 %** (Confidence 0.86)
  - `ability`: +18.0 × 0.86 – kann/können
- **WingScore:** nicht ausreichend bewertbar

## G17 – Ich kann sowieso nichts ändern.

**Kontext:** `SELF_TALK`  
**Patterns:** GENERALIZATION  

- **Wirksamkeit: 52.3 %** (Confidence 0.98)
  - `ability`: +18.0 × 0.86 – kann/können
  - `absolutization`: -14.0 × 0.84 – Generalisierung
- **Offenheit: 41.5 %** (Confidence 0.86)
  - `absolutization`: -16.0 × 0.86 – Generalisierung
- **WingScore:** nicht ausreichend bewertbar

## G18 – Das hat bisher nicht funktioniert. Was könnten wir beim nächsten Versuch verändern?

**Kontext:** `WORKPLACE`  
**Patterns:** OPENING_LANGUAGE  

- **Offenheit: 59.3 %** (Confidence 0.84)
  - `opening`: +18.0 × 0.84 – Öffnung
- **WingScore:** nicht ausreichend bewertbar

## G19 – Wir haben ein technisches Problem mit der Schnittstelle.

**Kontext:** `WORKPLACE`  
**Patterns:** —  
**Hinweise:** Sachlicher Problembegriff: kein pauschaler Negativscore  

- **Klarheit: 56.0 %** (Confidence 0.80)
  - `technical_issue`: +12.0 × 0.80 – technisches Problem
- **WingScore:** nicht ausreichend bewertbar

## G20 – Du bist das Problem.

**Kontext:** `PRIVATE_CONVERSATION`  
**Patterns:** PERSON_DEVALUATION, PREDICATIVE_LABELING  

- **Verbindung: 33.4 %** (Confidence 0.92)
  - `person_label`: -30.0 × 0.92 – du bist das Problem
- **Wertschätzung: 30.3 %** (Confidence 0.95)
  - `person_label`: -35.0 × 0.95 – du bist das Problem
- **Klarheit: 59.6 %** (Confidence 0.86)
  - `person_label_clear`: +18.0 × 0.86 – du bist das Problem
- **WingScore:** 40.6

## G21 – Der Fehler zeigt uns, was wir beim nächsten Versuch verändern können.

**Kontext:** `WORKPLACE`  
**Patterns:** OPENING_LANGUAGE  

- **Offenheit: 59.3 %** (Confidence 0.84)
  - `opening`: +18.0 × 0.84 – Öffnung
- **WingScore:** nicht ausreichend bewertbar

## G22 – Hast du Geld?

**Kontext:** `MODERATION`  
**Patterns:** —  
**Hinweise:** Homophonie-Link hast↔hasst erkannt  

- **WingScore:** nicht ausreichend bewertbar

## G23 – Hast du genug Zeit für dich?

**Kontext:** `WEBSITE`  
**Patterns:** —  
**Hinweise:** Homophonie-Link hast↔hasst erkannt  

- **WingScore:** nicht ausreichend bewertbar

## G24 – Wir müssen das Hindernis umfahren.

**Kontext:** `UNKNOWN`  
**Patterns:** INTERNAL_PRESSURE  

- **Freier Wille: 39.7 %** (Confidence 0.84)
  - `must_default`: -20.0 × 0.84 – müssen
- **WingScore:** nicht ausreichend bewertbar

## G25 – Der Eintritt ist frei.

**Kontext:** `PUBLIC_INFORMATION`  
**Patterns:** —  
**Hinweise:** frei als kostenlos erkannt; kein Beitrag zu Freier Wille  

- **WingScore:** nicht ausreichend bewertbar

## G26 – Er soll sehr erfolgreich sein.

**Kontext:** `PRIVATE_CONVERSATION`  
**Patterns:** REPORTED_CLAIM  
**Hinweise:** sollen als Hörensagen erkannt; kein Normativitätsmalus  

- **WingScore:** nicht ausreichend bewertbar

## G27 – Wenn du möchtest, können wir gemeinsam eine andere Möglichkeit prüfen.

**Kontext:** `COACHING`  
**Patterns:** OPENING_LANGUAGE  

- **Verbindung: 59.6 %** (Confidence 0.86)
  - `connection`: +18.0 × 0.86 – gemeinsam/miteinander
- **Offenheit: 59.3 %** (Confidence 0.84)
  - `opening`: +18.0 × 0.84 – Öffnung
- **WingScore:** nicht ausreichend bewertbar

## G28 – Du musst einfach positiv denken.

**Kontext:** `COACHING`  
**Patterns:** INTERNAL_PRESSURE  

- **Freier Wille: 39.7 %** (Confidence 0.84)
  - `must_default`: -20.0 × 0.84 – müssen
- **WingScore:** nicht ausreichend bewertbar

## G29 – Ich habe große Angst und möchte herausfinden, was mir jetzt helfen kann.

**Kontext:** `SELF_TALK`  
**Patterns:** OPENING_LANGUAGE  
**Hinweise:** Angst wird als Emotionsbenennung nicht automatisch negativ gescort  

- **Offenheit: 59.3 %** (Confidence 0.84)
  - `opening`: +18.0 × 0.84 – Öffnung
- **WingScore:** nicht ausreichend bewertbar

## G30 – Eigentlich wollte ich absagen, aber ich bin noch unsicher.

**Kontext:** `PRIVATE_CONVERSATION`  
**Patterns:** —  

- **WingScore:** nicht ausreichend bewertbar

---

# Lernpunkte

Die Simulation bestätigt die Grundarchitektur. Sie zeigt zugleich bewusst Lücken bei Syntax, Propositionen, Sense-Disambiguation und feineren Kontextregeln. Diese Fälle werden nicht künstlich gefüllt.

> **Eine gute Referenzengine darf lieber „noch nicht ausreichend bewertbar“ sagen, als fehlendes Sprachverständnis mit präzise aussehenden Zahlen zu kaschieren.**
