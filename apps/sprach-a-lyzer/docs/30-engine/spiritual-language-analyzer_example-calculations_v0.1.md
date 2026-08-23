# Spiritual Language Analyzer
## Example Calculations v0.1

**Status:** synthetische Referenzrechnungen zur Kalibrierung  
**Version:** 0.1  
**Datum:** 19. August 2026

---

# 1. Wichtiger Hinweis

Diese Berechnungen sind **keine Ergebnisse einer bereits implementierten SLA-Engine**. Die Contributions wurden als fachlich plausible Testwerte gesetzt, um Formel, Sättigung und Confidence zu prüfen.

Verwendete Formel:

```text
effective_i = contribution_i × confidence_i
S = Σ effective_i
raw = 100 × tanh(S / 80)
display = (raw + 100) / 2
```

Dimension Confidence:

```text
1 - Π(1 - confidence_i)
```

---

# 2. Beispielrechnungen

## G01 – Ich muss das heute unbedingt noch schaffen.

- **Freier Wille:** Σ effektiv = -21.9; Raw = -26.7; Anzeige = **36.7 %**; Confidence = 0.96
- **Klarheit:** Σ effektiv = 15.8; Raw = 19.5; Anzeige = **59.8 %**; Confidence = 0.88
- **Wirksamkeit:** Σ effektiv = -1.5; Raw = -1.9; Anzeige = **49.1 %**; Confidence = 0.90
- **Offenheit:** Σ effektiv = -7.4; Raw = -9.2; Anzeige = **45.4 %**; Confidence = 0.74

## G07 – Nein. Ich möchte das nicht.

- **Freier Wille:** Σ effektiv = 26.3; Raw = 31.8; Anzeige = **65.9 %**; Confidence = 0.94
- **Klarheit:** Σ effektiv = 33.6; Raw = 39.7; Anzeige = **69.8 %**; Confidence = 0.96
- **Wirksamkeit:** Σ effektiv = 18.0; Raw = 22.1; Anzeige = **61.1 %**; Confidence = 0.90
- **Wertschätzung:** Σ effektiv = 1.1; Raw = 1.4; Anzeige = **50.7 %**; Confidence = 0.55

## G09 – Ja, aber das funktioniert doch sowieso nie.

- **Verbindung:** Σ effektiv = -9.8; Raw = -12.2; Anzeige = **43.9 %**; Confidence = 0.82
- **Offenheit:** Σ effektiv = -24.4; Raw = -29.6; Anzeige = **35.2 %**; Confidence = 0.98
- **Wirksamkeit:** Σ effektiv = -14.1; Raw = -17.4; Anzeige = **41.3 %**; Confidence = 0.88
- **Klarheit:** Σ effektiv = 5.6; Raw = 7.0; Anzeige = **53.5 %**; Confidence = 0.70

## G11 – Ich sollte längst weiter sein.

- **Freier Wille:** Σ effektiv = -20.7; Raw = -25.3; Anzeige = **37.3 %**; Confidence = 0.96
- **Wertschätzung:** Σ effektiv = -9.6; Raw = -11.9; Anzeige = **44.0 %**; Confidence = 0.80
- **Wirksamkeit:** Σ effektiv = -7.8; Raw = -9.7; Anzeige = **45.1 %**; Confidence = 0.78
- **Offenheit:** Σ effektiv = -7.6; Raw = -9.5; Anzeige = **45.3 %**; Confidence = 0.76

## G16 – Ich kann nicht alles beeinflussen, aber ich kann meinen nächsten Schritt wählen.

- **Wirksamkeit:** Σ effektiv = 32.6; Raw = 38.7; Anzeige = **69.3 %**; Confidence = 0.99
- **Freier Wille:** Σ effektiv = 22.1; Raw = 26.9; Anzeige = **63.5 %**; Confidence = 0.92
- **Offenheit:** Σ effektiv = 15.5; Raw = 19.1; Anzeige = **59.6 %**; Confidence = 0.86
- **Klarheit:** Σ effektiv = 17.6; Raw = 21.7; Anzeige = **60.8 %**; Confidence = 0.88

## G20 – Du bist das Problem.

- **Wertschätzung:** Σ effektiv = -33.2; Raw = -39.3; Anzeige = **30.3 %**; Confidence = 0.95
- **Verbindung:** Σ effektiv = -27.6; Raw = -33.2; Anzeige = **33.4 %**; Confidence = 0.92
- **Klarheit:** Σ effektiv = 17.6; Raw = 21.7; Anzeige = **60.8 %**; Confidence = 0.88
- **Offenheit:** Σ effektiv = -7.2; Raw = -9.0; Anzeige = **45.5 %**; Confidence = 0.72

## G22 – Hast du Geld?

**Hinweis:** Homophonie `hast ↔ hasst` läuft im Defaultmodus `HINT_ONLY` und verändert hier keinen Dimensionswert.

- **Klarheit:** Σ effektiv = 9.0; Raw = 11.2; Anzeige = **55.6 %**; Confidence = 0.90

## G29 – Ich habe große Angst und möchte herausfinden, was mir jetzt helfen kann.

- **Wirksamkeit:** Σ effektiv = 17.6; Raw = 21.7; Anzeige = **60.8 %**; Confidence = 0.88
- **Offenheit:** Σ effektiv = 22.1; Raw = 26.9; Anzeige = **63.5 %**; Confidence = 0.92
- **Klarheit:** Σ effektiv = 15.5; Raw = 19.1; Anzeige = **59.6 %**; Confidence = 0.86
- **Wertschätzung:** Σ effektiv = 5.6; Raw = 7.0; Anzeige = **53.5 %**; Confidence = 0.70


# 3. Contribution Trace – G01

Text:

> „Ich muss das heute unbedingt noch schaffen.“

Vereinfachter Trace für **Freier Wille**:

```text
„ich muss“
  Base Contribution        -20
  Confidence               0.82
  Effective               -16.40

„unbedingt“
  zusätzlicher Pressure     -7
  Confidence               0.78
  Effective                -5.46

S                         -21.86
raw ≈ -26.7
display ≈ 36.7 %
```

Das passt zur Golden-Erwartung „Freier Wille niedrig“.

---

# 4. Beobachtungen

## A – Sanfte Sättigung
Einzelne mittlere Marker erzeugen keine extremen Scores.

## B – Neutralität
50 % bedeutet mathematisch neutral. Wenn für eine Dimension keine ausreichende Evidenz existiert, sollte jedoch `nicht bewertbar` statt 50 % erscheinen.

## C – Confidence ist keine Positivität
Ein sicher erkannter belastender Marker besitzt hohe Confidence und kann dennoch einen niedrigen Dimensionswert erzeugen.

## D – Resonanz im Default
Für den MVP ist `HINT_ONLY` als Default sinnvoll. Homophonien können sichtbar sein, ohne den semantischen Hauptscore zu verzerren.

## E – WingScore
Der WingScore sollte ausschließlich aus bewertbaren Dimensionen berechnet werden und Confidence berücksichtigen.

---

# 5. Leitgedanke

> **Die erste Mathematik soll uns nicht beweisen, dass unser Modell richtig ist. Sie soll uns zeigen, wo unsere fachlichen Annahmen noch nicht sauber genug sind.**
