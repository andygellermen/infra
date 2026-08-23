# Sprachkompass
## Spoken Language Resolver & Simulation v0.1

**Status:** erste ausführbare Alltagssprache-Analyse  
**Datum:** 19. August 2026

---

# 1. Ziel

Diese Referenzsimulation prüft, welche zusätzlichen Signale in spontaner Alltagssprache gegenüber geglätteter Schriftsprache sichtbar werden.

Sie verwendet **keinen pauschalen Drastikfaktor**.

---

# 2. Erkannte Spoken Features

```text
filler_count
generalizer_count
modal_count
repeated_tokens
self_corrections
fragments
repetition_factor
```

Füllwörter senken nicht den Score. Sie können lediglich die Sicherheit mancher automatischer Deutungen etwas reduzieren.

---

# 3. Paarvergleich

## S01 – Arbeitsdruck

**Geschrieben:** Ich möchte die Aufgabe heute noch fertigstellen.

**Gesprochen:** Also ich muss das heute unbedingt noch irgendwie schaffen, sonst wird das wieder nichts.

Spoken Features: Modalverben **1**, Generalisierer **1**, Füll-/Diskurspartikel **2**, Wiederholungen **{'das': 2}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | 58.7% | 41.5% | -17.2 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | — | — | +0.0 |
| Freier Wille | 60.8% | 35.7% | -25.1 |
| Offenheit | — | 37.2% | -12.8 |

## S02 – Meeting-Widerspruch

**Geschrieben:** Ich sehe deinen Punkt, würde aber eine andere Lösung bevorzugen.

**Gesprochen:** Ja, aber ganz ehrlich, das funktioniert doch sowieso nie so, das haben wir doch immer so gehabt.

Spoken Features: Modalverben **0**, Generalisierer **3**, Füll-/Diskurspartikel **1**, Wiederholungen **{'das': 2, 'doch': 2}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | — | 41.9% | -8.1 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | — | — | +0.0 |
| Freier Wille | — | — | +0.0 |
| Offenheit | — | 39.2% | -10.8 |

## S03 – Selbstbewertung

**Geschrieben:** Ich bin mit meinem Fortschritt noch nicht zufrieden.

**Gesprochen:** Ich sollte eigentlich längst viel weiter sein, keine Ahnung, warum ich das immer noch nicht hinbekomme.

Spoken Features: Modalverben **1**, Generalisierer **1**, Füll-/Diskurspartikel **1**, Wiederholungen **{'ich': 2}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | — | 40.3% | -9.7 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | 46.7% | -3.3 |
| Klarheit | — | — | +0.0 |
| Freier Wille | — | 40.3% | -9.7 |
| Offenheit | — | 43.5% | -6.5 |

## S04 – Grenze setzen

**Geschrieben:** Ich möchte das heute nicht entscheiden.

**Gesprochen:** Nee, also ich weiß nicht, ich kann das jetzt irgendwie nicht entscheiden, ich möchte eigentlich nicht.

Spoken Features: Modalverben **1**, Generalisierer **0**, Füll-/Diskurspartikel **4**, Wiederholungen **{'ich': 3, 'nicht': 3}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | 58.7% | 50.7% | -8.0 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | — | — | +0.0 |
| Freier Wille | 60.8% | 61.1% | +0.3 |
| Offenheit | — | 45.4% | -4.6 |

## S05 – Fehler im Team

**Geschrieben:** Die Vereinbarung wurde nicht eingehalten. Lass uns klären, was wir ändern können.

**Gesprochen:** Das ist schon wieder schiefgelaufen, irgendwie hält sich hier nie jemand an irgendwas und am Ende muss ich mich kümmern.

Spoken Features: Modalverben **1**, Generalisierer **1**, Füll-/Diskurspartikel **1**, Wiederholungen **{}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | 55.5% | 45.4% | -10.1 |
| Verbindung | 59.8% | — | -9.8 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | — | — | +0.0 |
| Freier Wille | — | 46.1% | -3.9 |
| Offenheit | 58.2% | 41.0% | -17.2 |

## S06 – Veränderung/KI

**Geschrieben:** Die Veränderung verunsichert mich, und ich möchte verstehen, welche Möglichkeiten ich habe.

**Gesprochen:** Ganz ehrlich, ich hab Angst, dass das mit der KI irgendwann alles übernimmt und wir dann sowieso keine Wahl mehr haben.

Spoken Features: Modalverben **0**, Generalisierer **2**, Füll-/Diskurspartikel **1**, Wiederholungen **{}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | 59.3% | 39.6% | -19.7 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | — | 52.1% | +2.1 |
| Freier Wille | 61.6% | 37.1% | -24.5 |
| Offenheit | — | 35.6% | -14.4 |

## S07 – Ressourcen

**Geschrieben:** Wir können heute nur einen Teil umsetzen und priorisieren deshalb.

**Gesprochen:** Wir können das alles unmöglich schaffen, wir haben nie genug Leute und es kommt ja immer noch mehr dazu.

Spoken Features: Modalverben **1**, Generalisierer **4**, Füll-/Diskurspartikel **0**, Wiederholungen **{'wir': 2}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | — | 32.6% | -17.4 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | — | — | +0.0 |
| Freier Wille | — | — | +0.0 |
| Offenheit | — | 33.2% | -16.8 |

## S08 – Feedback

**Geschrieben:** Ich wünsche mir, dass du Absprachen verlässlicher einhältst.

**Gesprochen:** Du musst echt endlich zuverlässiger werden, auf dich ist ja nie Verlass.

Spoken Features: Modalverben **1**, Generalisierer **1**, Füll-/Diskurspartikel **0**, Wiederholungen **{}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | — | 45.3% | -4.7 |
| Verbindung | — | 30.0% | -20.0 |
| Wertschätzung | — | 30.0% | -20.0 |
| Klarheit | — | — | +0.0 |
| Freier Wille | — | 34.7% | -15.3 |
| Offenheit | — | 36.5% | -13.5 |

## S09 – Entscheidung

**Geschrieben:** Ich möchte noch prüfen, welche Option besser passt.

**Gesprochen:** Also eigentlich weiß ich nicht, vielleicht sollte ich noch mal schauen, aber wahrscheinlich bringt das eh nichts.

Spoken Features: Modalverben **1**, Generalisierer **1**, Füll-/Diskurspartikel **2**, Wiederholungen **{'ich': 2}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | 58.7% | 36.0% | -22.7 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | — | — | +0.0 |
| Freier Wille | 60.8% | — | -10.8 |
| Offenheit | — | 35.5% | -14.5 |

## S10 – Lernhaltung

**Geschrieben:** Der Fehler zeigt mir, was ich beim nächsten Mal verändern kann.

**Gesprochen:** Okay, das war Mist, aber gut, jetzt weiß ich wenigstens, was ich beim nächsten Mal anders machen kann.

Spoken Features: Modalverben **1**, Generalisierer **0**, Füll-/Diskurspartikel **2**, Wiederholungen **{'ich': 2}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | — | 79.1% | +29.1 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | — | 65.1% | +15.1 |
| Freier Wille | — | — | +0.0 |
| Offenheit | — | 83.2% | +33.2 |

## S11 – Teamkonflikt

**Geschrieben:** Wir bewerten die Situation unterschiedlich. Ich möchte deinen Blick besser verstehen.

**Gesprochen:** Du verstehst mich einfach nicht, wir reden doch jedes Mal aneinander vorbei und irgendwie bringt das alles nichts.

Spoken Features: Modalverben **0**, Generalisierer **3**, Füll-/Diskurspartikel **1**, Wiederholungen **{}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | 58.7% | 33.1% | -25.6 |
| Verbindung | — | 38.9% | -11.1 |
| Wertschätzung | — | 44.9% | -5.1 |
| Klarheit | — | — | +0.0 |
| Freier Wille | 60.8% | — | -10.8 |
| Offenheit | — | 31.9% | -18.1 |

## S12 – Eigene Entscheidung

**Geschrieben:** Ich habe entschieden, die Aufgabe nicht zu übernehmen.

**Gesprochen:** Ich weiß nicht, ich soll das wohl wieder machen, obwohl ich eigentlich gar keine Zeit habe.

Spoken Features: Modalverben **1**, Generalisierer **0**, Füll-/Diskurspartikel **1**, Wiederholungen **{'ich': 3}**.

| Dimension | geschrieben | gesprochen | Δ |
|---|---:|---:|---:|
| Wirksamkeit | 60.8% | 47.7% | -13.1 |
| Verbindung | — | — | +0.0 |
| Wertschätzung | — | — | +0.0 |
| Klarheit | 60.8% | — | -10.8 |
| Freier Wille | 62.9% | 43.1% | -19.8 |
| Offenheit | — | — | +0.0 |


---

# 4. Erste Erkenntnisse

## A – Spontane Sprache liefert häufiger zusätzliche Evidenz

Besonders sichtbar werden:

- `müssen`
- `sollen`
- `immer`
- `nie`
- `sowieso`
- `keine Wahl`
- spontane Personalisierungen
- Wiederholungen

## B – Deutlichere Ausschläge entstehen aus mehr Treffern

Die gesprochene Variante kann dadurch stärker in Richtung:

- Ohnmacht
- Zwang
- Begrenzung
- Trennung

ausschlagen.

Umgekehrt kann spontane Sprache genauso deutlich positive Muster zeigen:

- Entscheidung
- Lernbewegung
- Anerkennung
- Offenheit
- eigener nächster Schritt

## C – Füllwörter sind keine Fehler

`also`, `irgendwie`, `eigentlich`, `halt`, `ähm` werden nicht moralisch oder negativ bewertet.

Sie können pragmatische Funktionen anzeigen und bei Mehrdeutigkeit die Confidence beeinflussen.

---

# 5. Wichtige Produktentscheidung

Für echte Tiefenanalyse sollten wir später unterscheiden:

```text
SPOKEN_DICTATION
DIRECT_AUDIO
```

Telefon-Diktat ist bereits wertvoll, kann aber Rohsprache glätten.

Direktes Audio könnte zusätzlich erfassen:

- Pausen
- Betonung
- Lautstärkeverlauf
- Sprechtempo
- Wiederholung
- Prosodie

Diese Signale dürfen nur als beobachtbare Merkmale bzw. klar gekennzeichnete Interpretationshinweise verwendet werden.

---

# 6. Mobile UX

Vorschlag:

> **Sprich einfach, wie du wirklich sprichst.**  
> Du musst deinen Gedanken nicht erst formulieren. Je spontaner du sprichst, desto mehr sprachliche Muster können sichtbar werden.

Nach der Analyse:

> **Spontane Sprachmuster**  
> In deiner gesprochenen Formulierung wurden zusätzliche Wiederholungen, Modalverben oder Verallgemeinerungen erkannt. Dadurch unterscheidet sich das Reflexionsprofil von einer geglätteten Textfassung.

---

# 7. Spirituell-reflexive Ebene

Im privaten Vertiefungsprofil kann zusätzlich erklärt werden:

> Manche Menschen erleben bestimmte Klang-, Wiederholungs- oder Ausdrucksmuster als besonders resonant. Diese Perspektive kann zusätzlich eingeblendet werden.

Diese Ebene bleibt von linguistisch beobachtbaren Befunden getrennt.

---

# 8. Leitgedanke

> **Gesprochene Sprache ist kein schlechterer Text. Sie ist ein anderer, oft unmittelbarer Zugang zu sprachlichen Gewohnheiten.**
