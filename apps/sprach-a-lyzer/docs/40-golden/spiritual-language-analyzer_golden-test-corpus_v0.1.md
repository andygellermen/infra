# Spiritual Language Analyzer
## Golden Test Corpus v0.1

**Status:** Fachlicher Referenzkorpus für Scoring-Validierung  
**Version:** 0.1  
**Datum:** 19. August 2026  
**Bezug:** Product Concept v1.1 – Scoring Engine

---

# 1. Zweck

Der Golden Test Corpus enthält bewusst unterschiedliche, teilweise extreme und mehrdeutige Beispielsätze.

Er dient nicht dazu, „richtige“ Sprache festzulegen. Er soll überprüfen, ob Sense-Erkennung, Kontextregeln, PatternClasses, Dimension Contributions, Confidence, Resonanzebene und später der WingScore plausibel und reproduzierbar zusammenspielen.

Dimensionen:
- WIR = Wirksamkeit
- VER = Verbindung
- WER = Wertschätzung
- KLA = Klarheit
- FW = Freier Wille
- OFF = Offenheit

Bewertung:
- `hoch` ≈ 70–100 %
- `mittel` ≈ 40–70 %
- `niedrig` ≈ 0–40 %
- `n/a` = nicht ausreichend bewertbar

---

# 2. Golden Cases

## G01 – Innerer Druck
**Text:** „Ich muss das heute unbedingt noch schaffen.“
**Kontext:** SELF_TALK  
**Patterns:** INTERNAL_PRESSURE, URGENCY  
**Erwartung:** WIR mittel; KLA hoch; FW niedrig; OFF niedrig bis mittel; VER/WER n/a.

## G02 – Reale Verpflichtung
**Text:** „Ich bin gesetzlich verpflichtet, die Unterlagen bis Freitag einzureichen.“
**Kontext:** LEGAL_ADMINISTRATIVE  
**Patterns:** EXTERNAL_OBLIGATION  
**Erwartung:** KLA hoch; FW mittel; WIR mittel bis hoch; OFF mittel.

## G03 – Sicherheitsanweisung
**Text:** „Du musst sofort das Gebäude verlassen!“
**Kontext:** SAFETY  
**Patterns:** SAFETY_DIRECTIVE, URGENCY  
**Erwartung:** KLA sehr hoch; WIR hoch; FW nicht negativ bestrafen.

## G04 – Selbstabwertung
**Text:** „Ich bin einfach ein Versager.“
**Kontext:** SELF_TALK  
**Patterns:** SELF_DEVALUATION, PREDICATIVE_LABELING  
**Erwartung:** WER sehr niedrig; WIR niedrig; OFF niedrig; KLA hoch.

## G05 – Verhalten statt Identität
**Text:** „Die Vereinbarung wurde heute nicht eingehalten.“
**Kontext:** WORKPLACE  
**Patterns:** BEHAVIOR_DESCRIPTION  
**Erwartung:** KLA hoch; WER neutral bis hoch.

## G06 – Personenabwertung
**Text:** „Auf dich ist nie Verlass.“
**Kontext:** PRIVATE_CONVERSATION  
**Patterns:** PERSON_DEVALUATION, GENERALIZATION  
**Erwartung:** WER niedrig; VER niedrig; OFF niedrig; KLA mittel.

## G07 – Klare Grenze
**Text:** „Nein. Ich möchte das nicht.“
**Kontext:** PRIVATE_CONVERSATION  
**Patterns:** CHOICE_LANGUAGE, CLEAR_BOUNDARY  
**Erwartung:** WIR hoch; KLA sehr hoch; FW hoch; WER neutral; VER kontextabhängig.

## G08 – Wertschätzende Grenze
**Text:** „Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.“
**Kontext:** PRIVATE_CONVERSATION  
**Patterns:** ACKNOWLEDGEMENT, CLEAR_BOUNDARY  
**Erwartung:** VER hoch; WER hoch; KLA hoch; FW hoch; OFF mittel.

## G09 – „Ja, aber“ als Abwehr
**Text:** „Ja, aber das funktioniert doch sowieso nie.“
**Kontext:** WORKPLACE  
**Patterns:** DISCOUNTING, DEFENSIVE_OBJECTION, GENERALIZATION  
**Erwartung:** OFF niedrig; VER niedrig bis mittel; WIR niedrig; KLA mittel.

## G10 – „Ja, aber“ konstruktiv
**Text:** „Ja, aber wir sollten zwischen den beiden Situationen unterscheiden.“
**Kontext:** WORKPLACE  
**Patterns:** PARTIAL_AGREEMENT, CONSTRUCTIVE_DIFFERENTIATION  
**Erwartung:** KLA hoch; OFF hoch; VER mittel bis hoch.

## G11 – Internalisiertes Sollen
**Text:** „Ich sollte längst weiter sein.“
**Kontext:** SELF_TALK  
**Patterns:** INTERNALIZED_EXPECTATION, SELF_PRESSURE  
**Erwartung:** FW niedrig; WER niedrig bis mittel; WIR niedrig bis mittel; OFF niedrig.

## G12 – Fürsorgliche Empfehlung
**Text:** „Du solltest heute etwas früher schlafen gehen.“
**Kontext:** FAMILY  
**Patterns:** NORMATIVE_ADVICE  
**Erwartung:** FW mittel; VER/WER stark kontextabhängig; KLA hoch.

## G13 – Konditionales Sollen
**Text:** „Solltest du Fragen haben, melde dich jederzeit.“
**Kontext:** WORKPLACE  
**Patterns:** CONDITIONAL_OPENING, OPENING_LANGUAGE  
**Erwartung:** VER hoch; OFF hoch; FW hoch/neutral; KLA hoch.

## G14 – Dürfen als Freiheit
**Text:** „Du darfst dir Zeit für die Entscheidung nehmen.“
**Kontext:** COACHING  
**Patterns:** CHOICE_LANGUAGE  
**Erwartung:** FW hoch; OFF hoch; VER hoch; WER hoch.

## G15 – Dürfen als Verbot
**Text:** „Du darfst das nicht.“
**Kontext:** UNKNOWN  
**Patterns:** PROHIBITION  
**Erwartung:** FW niedrig; KLA hoch; VER/WER ohne Kontext n/a.

## G16 – Möglichkeit statt Ohnmacht
**Text:** „Ich kann nicht alles beeinflussen, aber ich kann meinen nächsten Schritt wählen.“
**Kontext:** SELF_TALK  
**Patterns:** REALISTIC_LIMIT, CHOICE_LANGUAGE, RESPONSIBILITY_LANGUAGE  
**Erwartung:** WIR hoch; FW hoch; OFF hoch; KLA hoch.

## G17 – Generalisierte Ohnmacht
**Text:** „Ich kann sowieso nichts ändern.“
**Kontext:** SELF_TALK  
**Patterns:** GENERALIZATION, PERCEIVED_NO_CHOICE  
**Erwartung:** WIR sehr niedrig; FW niedrig; OFF niedrig.

## G18 – Lösungsoffenheit
**Text:** „Das hat bisher nicht funktioniert. Was könnten wir beim nächsten Versuch verändern?“
**Kontext:** WORKPLACE  
**Patterns:** OPENING_LANGUAGE, RESPONSIBILITY_LANGUAGE  
**Erwartung:** OFF hoch; WIR hoch; VER hoch; KLA hoch.

## G19 – Problem als Sachbegriff
**Text:** „Wir haben ein technisches Problem mit der Schnittstelle.“
**Kontext:** WORKPLACE  
**Patterns:** ISSUE_FRAMING  
**Erwartung:** KLA hoch; WER neutral; OFF neutral; kein pauschaler Negativscore für „Problem“.

## G20 – Problem als Personenlabel
**Text:** „Du bist das Problem.“
**Kontext:** PRIVATE_CONVERSATION  
**Patterns:** PREDICATIVE_LABELING, PERSON_DEVALUATION  
**Erwartung:** WER sehr niedrig; VER sehr niedrig; KLA hoch.

## G21 – Fehler als Lernereignis
**Text:** „Der Fehler zeigt uns, was wir beim nächsten Versuch verändern können.“
**Kontext:** WORKPLACE  
**Patterns:** LEARNING_FRAME, OPENING_LANGUAGE  
**Erwartung:** OFF hoch; WIR hoch; VER mittel bis hoch; KLA hoch.

## G22 – Homophonie auditiv
**Text:** „Hast du Geld?“
**Kontext:** MODERATION + AUDITORY + EXTERNAL_SPEECH  
**Relation:** hast ↔ hasst = HOMOPHONE  
**Erwartung:** Semantik neutral; Resonanzhinweis sichtbar, im Defaultmodus ohne Dimensionsmalus.

## G23 – Homophonie visuell
**Text:** „Hast du genug Zeit für dich?“
**Kontext:** WEBSITE + VISUAL + SILENT_READING  
**Relation:** hast ↔ hasst  
**Erwartung:** Semantik neutral; Resonanzhinweis optional; geringere Relevanz als G22.

## G24 – Homograph
**Text:** „Wir müssen das Hindernis umfahren.“
**Kontext:** UNKNOWN  
**Ambiguität:** umfahren / umfahren  
**Erwartung:** Sense Confidence nur hoch, wenn Syntax/Kontext die Lesart ausreichend auflöst; sonst Ambiguität anzeigen.

## G25 – Polysemie
**Text:** „Der Eintritt ist frei.“
**Kontext:** PUBLIC_INFORMATION  
**Sense:** kostenlos  
**Erwartung:** kein positiver Beitrag zur Dimension Freier Wille allein durch „frei“.

## G26 – Hörensagen
**Text:** „Er soll sehr erfolgreich sein.“
**Kontext:** PRIVATE_CONVERSATION  
**Sense:** REPORTED_CLAIM  
**Erwartung:** FW irrelevant; KLA/Confidence mittel; kein Normativitätsmalus.

## G27 – Offene Einladung
**Text:** „Wenn du möchtest, können wir gemeinsam eine andere Möglichkeit prüfen.“
**Kontext:** COACHING  
**Patterns:** CHOICE_LANGUAGE, OPENING_LANGUAGE, CONNECTION  
**Erwartung:** FW hoch; OFF sehr hoch; VER hoch; WER hoch.

## G28 – Toxic-Positivity-Falle
**Text:** „Du musst einfach positiv denken.“
**Kontext:** COACHING  
**Patterns:** EXTERNAL_EXPECTATION, PERSON_DIRECTIVE, POTENTIAL_INVALIDATION  
**Erwartung:** FW niedrig; VER niedrig bis mittel; WER niedrig bis mittel; KLA hoch.

## G29 – Emotion anerkennen
**Text:** „Ich habe große Angst und möchte herausfinden, was mir jetzt helfen kann.“
**Kontext:** SELF_TALK  
**Patterns:** EMOTION_ACKNOWLEDGEMENT, OPENING_LANGUAGE  
**Erwartung:** WIR hoch; OFF hoch; KLA hoch; keine Negativwertung von „Angst“ an sich.

## G30 – Mehrdeutiges „eigentlich“
**Text:** „Eigentlich wollte ich absagen, aber ich bin noch unsicher.“
**Kontext:** PRIVATE_CONVERSATION  
**Patterns:** HEDGING, IMPLICIT_CONTRAST, UNCERTAIN_COMMITMENT  
**Erwartung:** KLA mittel; FW mittel; OFF mittel bis hoch; Confidence mittel.

---

# 3. Freigabekriterium

Eine neue Rule-Set-Version darf die Golden Cases nur dann wesentlich verändern, wenn die fachliche Änderung dokumentiert ist, der neue Zielbereich bewusst beschlossen wurde und keine unbeabsichtigten Regressionen in anderen Fällen entstehen.

---

# 4. Leitgedanke

> **Der Golden Corpus ist kein Sprachgesetz. Er ist unser gemeinsamer Referenzspiegel für die Frage, ob die Engine das meint, was wir fachlich meinen.**
