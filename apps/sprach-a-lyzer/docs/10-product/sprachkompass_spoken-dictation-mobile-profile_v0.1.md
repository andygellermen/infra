# Sprachkompass
## Spoken Dictation / Mobile Input Profile v0.1

**Status:** Produkt-/Analyseprofil  
**Datum:** 19. August 2026

---

# 1. Ziel

Mobile Nutzer sollen ausdrücklich dazu eingeladen werden, Sprache **zu sprechen statt zu glätten**.

Vorgeschlagener CTA:

> **Sprich einfach, wie du wirklich sprichst.**  
> Diktiere deinen Gedanken so spontan, wie er gerade kommt. Gesprochene Alltagssprache kann andere Muster sichtbar machen als sorgfältig formulierter Text.

---

# 2. Warum gesprochene Alltagssprache wertvoll ist

Spontane Sprache enthält oft mehr:

- Wiederholungen,
- Modalverben,
- Generalisierungen,
- Selbstkorrekturen,
- Relativierungen,
- Gesprächspartikel,
- Satzabbrüche,
- unvollständige Propositionen,
- spontane Wahl-/Ohnmachtsformulierungen.

Damit wächst die **Evidenzmenge**.

Die Engine darf deshalb stärkere Ausschläge zeigen, wenn tatsächlich mehr relevante Muster auftreten.

---

# 3. Kein Drastik-Faktor

Nicht zulässig:

```text
spoken_input = score × 1.5
```

Zulässig:

```text
mehr echte Pattern Hits
+ lokale Wiederholung
+ spontane Pattern-Komposition
+ Modalitäts-/Pragmatiksignale
→ stärkere Dimensionsausschläge
```

So bleibt der Unterschied fachlich erklärbar.

---

# 4. InputMode

```text
TEXT
SPOKEN_DICTATION
DIRECT_AUDIO
```

## TEXT
bewusst formulierter oder geschriebener Text.

## SPOKEN_DICTATION
Text stammt aus Telefon-/Systemdiktat.

## DIRECT_AUDIO
spätere Audioanalyse mit Prosodie, Pause, Betonung und Rhythmus.

---

# 5. SPOKEN_DICTATION Features

```yaml
SpokenFeatures:
  repetition_factor
  filler_count
  self_correction_count
  fragment_count
  discourse_particle_count
  modal_density
  generalization_density
```

---

# 6. Repetition

Wiederholung darf Salienz erhöhen, aber nur nichtlinear.

Beispiel:

> „Ich muss, ich muss, ich muss das heute noch schaffen.“

Hier kann `INTERNAL_PRESSURE` stärker werden als bei einem einzelnen `muss`.

Die Verstärkung bleibt begrenzt.

---

# 7. Füll- und Diskurspartikel

Beispiele:

- äh
- ähm
- also
- irgendwie
- halt
- quasi
- eigentlich

Diese Wörter sind **keine Negativmarker**.

Sie können Hinweise liefern auf:

- HEDGING,
- SELF_CORRECTION,
- Unsicherheit,
- Gesprächsorganisation.

---

# 8. Systemdiktat ist nicht Roh-Audio

Eine Telefon-Diktierfunktion kann:

- Satzzeichen ergänzen,
- Wiederholungen entfernen,
- Fülllaute auslassen,
- Schreibweisen normalisieren.

Deshalb ist `SPOKEN_DICTATION` wertvoll, aber analytisch nicht identisch mit echter gesprochener Sprache.

Für spätere Tiefenanalyse ist `DIRECT_AUDIO` die bessere Quelle.

---

# 9. Datenschutz

Direct Audio sollte später standardmäßig:

- freiwillig sein,
- transparent verarbeitet werden,
- Audio möglichst nicht dauerhaft speichern,
- Transkript und Audio getrennt behandeln,
- lokale/temporäre Verarbeitung bevorzugen, sofern technisch sinnvoll.

---

# 10. Private / Corporate

## Persönliche Reflexion
Spontanes Diktat kann prominent angeboten werden.

## Berufsleben
Diktat ebenfalls möglich, aber mit Hinweis:

> Bitte keine vertraulichen Unternehmens-, Kunden- oder Personaldaten eingeben, sofern eure Unternehmensrichtlinien dies nicht erlauben.

---

# 11. UX-Idee

Mobile Startansicht:

```text
Was möchtest du reflektieren?

[ Text schreiben ]
[ Gedanken diktieren 🎙️ ]
```

Nach Wahl Diktat:

> „Sprich spontan. Du musst deinen Gedanken nicht erst schön formulieren.“

Das ist fachlich besonders passend, weil die Anwendung gerade die ungeglättete Sprachbewegung sichtbar machen soll.

---

# 12. Leitgedanke

> **Je spontaner die Sprache, desto mehr Muster können sichtbar werden. Nicht weil gesprochene Sprache schlechter ist – sondern weil sie weniger redaktionell geglättet ist.**
