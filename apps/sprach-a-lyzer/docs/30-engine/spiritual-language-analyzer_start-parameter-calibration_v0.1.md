# Spiritual Language Analyzer
## Start Parameter Calibration v0.1

**Status:** empfohlene Startkalibrierung nach Golden Corpus und Beispielrechnungen  
**Version:** 0.1  
**Datum:** 19. August 2026

---

# 1. Ziel

Dieses Dokument schärft die in Product Concept v1.1 vorgeschlagenen Startparameter.

Die Werte bleiben Hypothesen für den MVP und müssen mit Golden Corpus, Workshops, Coaches, skeptischen Nutzern und realen Texten weiter kalibriert werden.

---

# 2. Empfohlene Startwerte

| Parameter | v1.1 | Empfehlung v0.1 | Begründung |
|---|---:|---:|---|
| `dimension_aggregation_scale` | 80 | **80** | gute, sanfte Sättigung |
| `frequency_alpha` | 0.25 | **0.20** | Wiederholung zunächst konservativer |
| `assessability_threshold` | 0.80 | **0.75** | etwas weniger `n/a`, bleibt vorsichtig |
| Evidence A | 1.00 | **1.00** | linguistisch/korpusbasiert |
| Evidence B | 0.90 | **0.90** | kommunikative Hauptbasis |
| Evidence C | 0.65 | **0.60** | spirituell-reflexiv konservativer im automatischen Score |
| Evidence D | 0.50 | **0.40** | Community-Hypothese vor Review schwächer |
| Evidence E | 0.90 | **0.95** | persönlich bestätigte Resonanz darf für diesen Nutzer stark zählen |
| Resonance `OFF` | 0.00 | **0.00** | unverändert |
| Resonance `HINT_ONLY` | 0.00 | **0.00** | empfohlener MVP-Default |
| Resonance `MODERATE` | 0.25 | **0.20** | konservativer Start |
| Resonance `FULL` | 0.50 | **0.40** | Semantik soll dominieren |
| AUDITORY | 1.00 | **1.00** | unverändert |
| MIXED | 0.90 | **0.90** | unverändert |
| VISUAL | 0.55 | **0.60** | Lesen bleibt relevante Resonanzperspektive |
| INNER_SPEECH | 0.80 | **0.85** | persönlicher Reflexionskontext höher |
| SILENT_READING | 0.60 | **0.65** | Öffnung für Workshop-Beobachtungen |
| Weakest threshold | 0.35 | **0.30** | erst bei klarer Schieflage |
| Weakest beta | 0.20 | **0.10** | Penalty zunächst vorsichtig |
| local repetition | 1.15–1.50 | **1.10–1.35** | Übergewichtung vermeiden |

---

# 3. Resonanz im Standardmodus

Empfehlung:

```text
DEFAULT = HINT_ONLY
```

Das bedeutet:

- Homophonie und Klangresonanz werden erkannt,
- im UI kann ein Hinweis erscheinen,
- der Dimensionsscore verändert sich standardmäßig nicht.

Im Coach-/Resonanzmodus kann die Ebene aktiviert werden.

---

# 4. VISUAL und SILENT_READING

Empfehlung:

```text
VISUAL = 0.60
SILENT_READING = 0.65
```

Diese Werte wirken nur, wenn Resonanzscoring aktiviert ist.

Damit bleibt die Möglichkeit innerer oder beim Lesen aktivierter Klangassoziationen fachlich offen, ohne sie im Standardmodus numerisch zu behaupten.

---

# 5. Confidence-Kalibrierung

Empfohlene Schwellen:

```text
0.00–0.39 = niedrig
0.40–0.64 = mittel
0.65–0.84 = hoch
0.85–1.00 = sehr hoch
```

Im Standard-UI besser qualitative Labels statt scheinpräziser Prozentwerte.

---

# 6. Sense-Ambiguität

Wenn:

```text
top_sense - second_sense < 0.15
```

dann:

- Ambiguität markieren,
- Contribution Confidence um mindestens 20 % reduzieren,
- keine harte Umformulierung empfehlen.

Wenn:

```text
top_sense < 0.55
```

sollte der Treffer nicht stark scoren.

---

# 7. Contribution Caps

MVP-Empfehlung:

```text
|effective contribution per hit| <= 35
```

Ausnahmen später nur für klar definierte Sonderfälle.

Zusätzlich:

```text
|dimension contribution per sentence| <= 55
```

Damit werden Regelketten begrenzt.

---

# 8. Resonance Cap

Selbst im `FULL`-Modus:

```text
|resonance contribution per hit| <= 12
```

und:

```text
total resonance share per dimension <= 25 %
```

der effektiven Evidenzmasse.

Resonanz kann dadurch relevant sein, ohne direkte Semantik zu überstimmen.

---

# 9. Frequency Calibration

Empfehlung:

```text
frequency_alpha = 0.20
```

Wiederholung wird nur dann gemeinsam verstärkt, wenn dieselbe Bedeutung, PatternClass oder Dimensionsbewegung wiederkehrt.

---

# 10. WingScore-Kalibrierung

Für erste Tests:

```text
dimension_weight = 1.0
weakest_link_enabled = true
threshold = 0.30
beta = 0.10
```

Der WingScore sollte nur bewertbare Dimensionen einbeziehen.

---

# 11. Mindestbreite für WingScore

Empfehlung:

Ein WingScore wird nur angezeigt, wenn mindestens:

```text
3 Dimensionen
```

ausreichend bewertbar sind.

Bei weniger:

> „Für einen Gesamt-WingScore enthält der Text noch zu wenig bewertbare sprachliche Information.“

---

# 12. Parameterklassen im Admin-Panel

## Safe-to-edit
- Anzeigenamen
- Dimension Weight innerhalb enger Grenzen
- Explanation Templates
- Reflection Prompts
- Rule Enable/Disable in Draft

## Review-required
- Base Contributions
- Evidence Factors
- Frequency Alpha
- Resonance Factors
- Caps
- Assessability Threshold
- WingScore-Parameter

## Engine-locked
- Wertebereiche
- keine semantische Homophon-Vererbung
- keine zirkulären Regeln
- Confidence 0–1
- Audit / Versionierung

---

# 13. Empfohlene Admin-Grenzen

```text
base contribution: -50 .. +50
modifier factor: 0.0 .. 2.0
evidence factor: 0.0 .. 1.0
resonance mode factor: 0.0 .. 0.6
frequency alpha: 0.0 .. 0.5
aggregation scale: 40 .. 160
weakest beta: 0.0 .. 0.25
```

Werte außerhalb erfordern Developer-/Publisher-Override.

---

# 14. Kalibrierungsworkflow

```text
Draft
→ Golden Corpus komplett rechnen
→ Diff Report
→ Ausreißer prüfen
→ 5–10 reale Texte prüfen
→ Review
→ Publish
```

---

# 15. Warnschwellen

Warnung, wenn eine Draft-Version gegenüber Production:

- durchschnittlichen WingScore > 8 Punkte verschiebt,
- eine Dimension im Median > 10 Punkte verschiebt,
- `non-assessable` um > 15 Prozentpunkte verändert,
- Resonanzanteil > 25 % erreicht,
- eine einzelne Regel > 30 % aller Analysen triggert.

---

# 16. Corporate-/Workshop-Profil

Für einen Corporate-Workshop empfiehlt sich **kein eigenes fundamentales Scoring-Modell**.

Besser:

> gleiches Kernmodell + eigenes Rule/Display Profile

Beispiel `CORPORATE_WORKSHOP`:

- Resonanz standardmäßig `HINT_ONLY`
- Klarheit, Verbindung, Freier Wille prominent
- spirituell-reflexive Perspektive optional vertiefbar
- keine persönlichen Langzeitprofile
- Fokus auf reale Arbeitskommunikation

---

# 17. Zwei mögliche Betriebsweisen

## A – konservativer MVP
Resonanz nur als Hinweis.

## B – Coach-kalibrierbarer MVP
Resonanz kann pro Session aktiviert werden.

**Empfehlung:** B technisch vorbereiten, A als Default.

---

# 18. Definition of Done – erste Kalibrierung

- [x] Golden Test Corpus v0.1 definiert
- [x] erste synthetische Beispielrechnungen durchgeführt
- [x] Aggregationsfunktion plausibilisiert
- [x] Resonanz-Default festgelegt
- [x] Evidence-Faktoren geschärft
- [x] Frequency Alpha konservativer gesetzt
- [x] Weakest-Link-Penalty abgeschwächt
- [x] Contribution Caps vorgeschlagen
- [x] Resonance Caps vorgeschlagen
- [x] Sense-Ambiguitätsregeln vorgeschlagen
- [x] WingScore-Mindestbreite vorgeschlagen
- [x] Admin-Grenzen definiert
- [ ] Golden Corpus später mit implementierter Engine automatisiert rechnen
- [ ] reale Workshop-Sätze ergänzen
- [ ] Corporate-Workshop-Profil validieren
- [ ] nach Nutzerfeedback Parameter Set v0.2

---

# 19. Leitgedanke

> **Kalibrierung bedeutet nicht, die gewünschte Antwort in die Engine hineinzudrehen. Sie bedeutet, die Regeln so transparent zu justieren, dass ähnliche sprachliche Situationen zuverlässig ähnlich behandelt werden.**
