# Sprach-A-Lyzer – Technology & AI Architecture Review v0.1

**Stand:** 21. August 2026  
**Zweck:** Prüfung der bisherigen Technologieentscheidungen mit Fokus auf Datenbank, deterministische Analyse, optionale KI, Datenschutz und Audio.

## 1. Strategische Kernentscheidung

Der Sprach-A-Lyzer sollte **ohne generative KI einen eigenständigen, attraktiven und erklärbaren Nutzwert liefern**.

KI wird nicht zur Voraussetzung für:

- Basisanalyse
- Pattern-Erkennung
- Q/A Composition
- Dimensionsauswertung
- Contribution Trace
- Golden Testing
- Frageauswahl
- Feedback auf Basis freigegebener Templates

KI ist eine **optionale Vertiefungsschicht**.

## 2. Warum diese Trennung sinnvoll ist

- **Vertrauen:** Persönliche Sprache ist hochsensibel.
- **Erklärbarkeit:** Regel-/Pattern-basierte Ergebnisse lassen sich exakt begründen.
- **Datenschutz:** Kein externer KI-Provider muss persönliche Texte erhalten.
- **Kosten:** Core Analyse bleibt günstig und berechenbar.
- **Verfügbarkeit:** Core funktioniert auch bei KI-Ausfall oder offline.
- **Produktschärfe:** KI ergänzt den Sprach-A-Lyzer, ersetzt ihn nicht.

## 3. Produktmodi

### CORE – ohne generative KI

```text
Text / Q&A
↓
Tokenizer / Normalizer
↓
Phrase Matcher
↓
Morphology / Syntax
↓
Sense Resolver
↓
Proposition Graph
↓
Pattern Rules
↓
Q/A Composition
↓
Construct Evidence
↓
Dimension Scoring
↓
Template-based Explanation
```

### ENHANCED – freiwillige KI

```text
Core Result
+ User opt-in
↓
LLM / local or external
↓
individualisierte Erklärung
alternative Formulierungen
tieferer Reflexionsdialog
```

### AUDIO – später

```text
Audio
↓
local/offline ASR preferred
↓
Transcript
↓
Core Engine
+
optional Prosody Features
```

## 4. Kann der Core ohne KI attraktiv genug sein?

Ja – wenn er nicht wie ein Wörterbuch wirkt.

Der Core sollte mindestens liefern:

1. konkrete Textmarkierungen
2. nachvollziehbare Pattern-Erklärungen
3. Q/A-Kontext
4. sechs Dimensionsprofile mit Assessability
5. Contribution Trace
6. passende Reflexionsfrage
7. redaktionell freigegebene Alternativformulierungen
8. Verlauf / Vorher-Nachher-Sprachmuster
9. adaptive Fragewahl
10. klare Unsicherheitsanzeige

## 5. Grenzen ohne KI

Schwieriger bleiben:

- offene Paraphrasen
- sehr freie semantische Äquivalenz
- komplexe Ironie/Sarkasmus
- pragmatische Implikaturen
- hochflexible Mehrdeutigkeit
- natürlich klingende individuelle Restaurierung
- freie Coaching-Folgefragen
- semantische Similarity über große Ausdrucksvarianten

Diese Grenzen sollten sichtbar kommuniziert werden.

## 6. Rolle von KI

KI ist besonders wertvoll für:

- natürlichsprachliche Erklärung des Core Trace
- individuelle Rephrasing-Vorschläge
- Ambiguity Assistance
- adaptive Formulierung einer freigegebenen nächsten Frage
- Long-form Session-Zusammenfassung

Grundsatz:

> **LLM proposes; Rule Engine validates/scores.**

## 7. KI darf nicht allein entscheiden

Kein LLM darf autonom:

- Dimensionenscores festlegen
- Menschen diagnostizieren
- WingScore erzeugen
- Evidenzklassen ändern
- Corporate Guardrails umgehen
- spirituelle Hypothesen als Tatsachen ausgeben

## 8. Consent-Modell

Sichtbare Wahl:

```text
[ Basisanalyse ohne KI ]
[ Analyse vertiefen ]
```

Bei Vertiefung:

```text
Diese optionale Analyse verwendet ein Sprachmodell, um bereits erkannte
Muster individueller zu erklären und Formulierungsalternativen zu erzeugen.

[ lokal auf diesem Gerät, wenn verfügbar ]
[ geschützter Server ]
[ externer KI-Dienst – nur nach ausdrücklicher Zustimmung ]
```

Keine versteckte Weitergabe.

## 9. Local-first AI

Lokale LLM-Inferenz ist technisch realistisch, z. B. über `llama.cpp`.

Vorteile:

- Text verlässt das Gerät nicht
- bessere Vertrauensposition
- Offline-Betrieb

Nachteile:

- Hardwareabhängigkeit
- Qualität kleiner Modelle
- Modellverteilung/Updates
- RAM-/Energiebedarf

## 10. Backend: Go bleibt sinnvoll

Go bleibt geeignet für:

- API
- Rule Engine
- Scoring
- Import Pipeline
- Sessions
- Auth
- Audit
- Orchestrierung
- Concurrency
- Deployment

Aber:

> **Nicht alles muss in Go implementiert sein.**

## 11. NLP nicht künstlich auf Pure-Go beschränken

Für anspruchsvolles Deutsch ist ein separater NLP-Adapter sinnvoll.

```text
Go Core
   │
   ├── deterministic lexical/rule processing
   │
   └── NLP Adapter
         ├── spaCy service
         ├── future ONNX/local transformer
         └── optional LLM adapter
```

## 12. Option A – Go + spaCy Sidecar

**Favorit für MVP/Qualität.**

Vorteile:

- deutsche Tokenisierung/POS/Morphologie
- Dependency Parsing
- Lemmatisierung
- NER
- klare Servicegrenze

Nachteile:

- zweiter Runtime-Stack
- Modelle sind nicht speziell auf spontane Coaching-Sprache trainiert

Deshalb: eigener Golden Corpus bleibt zwingend.

## 13. Option B – Go + ONNX Runtime

Später attraktiv, wenn geeignete deutsche Modelle sauber als ONNX verfügbar und validiert sind.

Vorteile:

- weniger Python-Abhängigkeit
- lokalisierbare Inferenz

Nachteile:

- höherer Engineering-Aufwand
- Tokenizer-/Model-Plumbing
- eigene Modellpflege

## 14. Option C – Pure Go

Geeignet für:

- Normalisierung
- Token-/Phrase-Matching
- Rule Engine
- einfache Feature Detection

Nicht als alleinige Grundlage für anspruchsvolle deutsche Syntax/Sense-Auflösung empfohlen.

## 15. Datenbankreview

**PostgreSQL bleibt sehr passend.**

Gründe:

- relationale Integrität
- Transaktionen
- JSONB
- GIN-Indizes
- Full Text Search
- Audit/History
- Import Staging
- Versionierung
- später `pgvector`
- etablierte Backups/PITR

## 16. Datenmodell-Empfehlung

### Relational speichern

```text
lexemes
senses
phrases
relations
dimensions
constructs
questions
question_renderings
rules
rule_sets
sources
golden_cases
```

### JSONB verwenden

```text
condition_tree
rule_actions
analysis_trace
resolver_candidates
import_raw_payload
diff_payload
experimental metadata
```

Nicht alles in JSONB schieben.

## 17. Graph DB?

Noch nicht.

Erst erwägen bei echten Anforderungen wie:

- tiefe mehrstufige Traversalen
- graphweite Pfadalgorithmen
- Knowledge Discovery über sehr viele Relationstypen

Für MVP ist eine Graphdatenbank zusätzliche Komplexität ohne klaren Nutzen.

## 18. Vector DB?

Nicht für den Core nötig.

Falls später benötigt für:

- Similar Question Search
- Similar Phrase Search
- Semantic Retrieval
- RAG

zuerst `pgvector` statt separater Vektordatenbank prüfen.

## 19. SQLite?

Für den zentralen Server-Core: PostgreSQL bleibt besser.

Für eine spätere lokale/private App kann SQLite sehr attraktiv sein:

```text
Server: PostgreSQL
Device: SQLite
```

Lokale Nutzung:

- Offline Sessions
- lokale Analysehistorie
- private Knowledge Packs
- local-first Datenschutz

## 20. Audio / Speech-to-Text

Audio und Prosodie getrennt behandeln.

### Phase 1
Diktat / Transkript.

### Phase 2
lokale ASR.

### Phase 3
Prosodie separat.

## 21. Offline ASR

`whisper.cpp` ermöglicht Whisper-basierte lokale/offline Inferenz auf vielen Plattformen.

```text
Audio
↓
local whisper.cpp adapter
↓
transcript + timestamps
↓
SpokenFeatures
↓
Core Engine
```

Das kann Datenschutzbedenken erheblich reduzieren.

## 22. Prosodie ist ein eigener Baustein

Später mögliche Features:

- Pausen
- Sprechtempo
- Betonungs-/Lautstärkeverlauf
- Wiederholung
- Tonhöhenverlauf
- Rhythmus

Zulässig:

> „Das Wort wurde deutlich stärker betont.“

Nicht zulässig:

> „Du bist innerlich aggressiv.“

## 23. ASR-Unsicherheit

Soweit verfügbar speichern:

```text
asr_confidence
token_confidence
audio_quality
```

Unsichere Transkripte reduzieren Sense-/Pattern-Confidence.

## 24. Privacy Modes

### PRIVATE_LOCAL
- möglichst lokal
- keine Speicherung standardmäßig
- lokale KI optional

### PRIVATE_CLOUD
- Core serverseitig
- explizite Speicherung optional

### CORPORATE
- minimierte Speicherung
- keine Manager-Einzelanalysen
- kein Rohtext-Logging standardmäßig
- KI nur explizit

## 25. Sensitive Text Lifecycle

Default:

```text
raw text:
process → analyze → discard
```

Wenn Verlauf gewünscht:

```text
explicit opt-in
```

Alternative:

```text
store derived features only
```

z. B. Pattern Counts, Construct Evidence, Dimensions – ohne vollständigen Rohtext.

## 26. Zielarchitektur v0.2

```text
                 CLIENT
                   │
        ┌──────────┴──────────┐
        │                     │
   Text/Q&A UI            Audio UI
        │                     │
        │               Local ASR optional
        │                     │
        └──────────┬──────────┘
                   ▼
               GO API
                   │
     ┌─────────────┼─────────────┐
     │             │             │
 Rule/Score     NLP Adapter    AI Adapter
   Engine           │             │
     │          spaCy/ONNX     disabled
     │                         local LLM
     │                         cloud LLM
     │
     └─────────────┬─────────────┘
                   ▼
              PostgreSQL
                   │
              pgvector later
```

## 27. Feature Flags

```text
AI_EXPLANATION
AI_REPHRASING
AI_ADAPTIVE_QUESTION
LOCAL_LLM
CLOUD_LLM
LOCAL_ASR
PROSODY_ANALYSIS
STORE_RAW_TEXT
STORE_AUDIO
```

Alle unabhängig schaltbar.

## 28. MVP Empfehlung

### Unbedingt
- deterministische Textanalyse
- Question Context
- Sense/Proposition Resolver
- Patterns
- Construct Evidence
- Dimensions
- Contribution Trace
- Template Feedback
- adaptive Fragen aus freigegebenem Pool

### Optional Preview
- KI-Erklärung nach User Consent

### Später
- generatives Coaching
- Audio-Prosodie
- Vector Search
- Graph DB

## 29. Entscheidender Produkttest

Vor größerer KI-Integration:

> **Würden Nutzer den Sprach-A-Lyzer wieder benutzen, wenn die optionale KI-Schaltfläche nicht existiert?**

Wenn nein:
Der Core ist noch nicht stark genug.

Wenn ja:
KI ist echter Zusatznutzen statt Krücke.

## 30. Produktmetriken

### Core
- Erkenntnis-Relevanz
- Verständlichkeit
- Trace-Nutzung
- Alternative gewählt?
- nächste Frage beantwortet?
- Session Completion
- Wiederkehr

### KI-Vertiefung
- Opt-in Rate
- wahrgenommener Mehrwert
- lokale vs. Cloud-Präferenz
- Abbruch nach Consent
- Qualität der Alternativen

## 31. Architekturentscheidung

### Beibehalten
- Go als Core Backend
- PostgreSQL
- modularer Monolith
- Rule Engine
- Golden Tests
- Presentation Bundles
- Managed Import

### Ergänzen
- klarer NLP Adapter
- spaCy Sidecar als realistische MVP-Option
- KI Adapter als optionales Plugin
- Local-first Privacy Mode
- lokale ASR-Option
- Feature Flags

### Nicht erzwingen
- Pure-Go-NLP
- KI in jedem Analyseweg
- Vector DB
- Graph DB
- Audio-Speicherung

## 32. North Stars

> **Der Sprach-A-Lyzer muss ohne KI glaubwürdig, nützlich und erklärbar sein. KI darf ihn persönlicher und flexibler machen – aber nicht erst zu einem Produkt machen.**

> **Der Nutzer entscheidet nicht nur, was er analysieren lässt, sondern auch, welche technische Tiefe er dafür zulässt.**
