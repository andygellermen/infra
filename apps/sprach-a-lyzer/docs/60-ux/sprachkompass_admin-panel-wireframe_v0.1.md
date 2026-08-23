# Sprachkompass / Spiritual Language Analyzer
## Admin Panel Wireframe v0.1

**Status:** textueller Low-Fidelity-Wireframe  
**Datum:** 19. August 2026

---

# 1. Navigationsstruktur

```text
┌──────────────────────────────────────────────────────────────┐
│ Sprachkompass Admin                           Ruleset v0.1 DRAFT │
├──────────────┬───────────────────────────────────────────────┤
│ Dashboard    │                                               │
│ Knowledge    │              Arbeitsbereich                   │
│ Dimensions   │                                               │
│ Rules        │                                               │
│ Patterns     │                                               │
│ Relations    │                                               │
│ Resonance    │                                               │
│ Mappings     │                                               │
│ Parameters   │                                               │
│ Test Lab     │                                               │
│ Rule Sets    │                                               │
│ Sources      │                                               │
│ Audit        │                                               │
└──────────────┴───────────────────────────────────────────────┘
```

# 2. Dashboard

```text
[Production Ruleset v0.1] [Draft v0.2] [Golden Tests 27/30]

Score Drift          +1.8
Non-assessable       14 %
Resonance Share       0 % (Default HINT_ONLY)
Open Reviews           6

[Run Golden Corpus] [Compare Draft] [Open Test Lab]
```

# 3. Mapping Manager

Neu aufgrund Corporate-/Privat-Profil:

```text
Profile: [PRIVAT ▼] [BERUFSLEBEN ▼] [COACHING ▼]

Canonical Concept      Display Text                 Visibility
RESONANCE               Resonanz                     visible
RESONANCE               Wahrnehmungswirkung          corporate
ENERGY                   Energie                      private/deep
ENERGY                   Aktivierung                  corporate
CONSCIOUSNESS            Bewusstsein                  private/deep
CONSCIOUSNESS            Reflexionsfähigkeit          corporate

[+ Mapping] [Preview Profile]
```

Kernregel: Mapping verändert **keine** Relation, Dimension oder Punktzahl.

# 4. Rule Builder

```text
Rule: amplify_unbedingt_muessen                 [Enabled ✓]
Priority: [200]

WENN
  [Lexeme] [equals] [müssen]                  [+ AND]
  [Token]  [within 3 tokens] [unbedingt]       [+ AND]
  [Context][not equals] [SAFETY]

DANN
  [Freier Wille] [MULTIPLY] [1.35] Richtung Zwang
  [Pattern] [ADD] [URGENCY]

Confidence [0.80]      Stop chain [ ]

[Test Rule] [Save Draft]
```

# 5. Dimension Manager

```text
Freier Wille
Negative Pole: Zwang
Default Weight: 1.0
Visible: ✓
Corporate Label: Freier Wille
Private Label: Freier Wille

[Contributions] [Rules] [Golden Cases]
```

# 6. Resonance Manager

```text
Default Mode: HINT_ONLY

OFF        0.00
HINT_ONLY  0.00
MODERATE   0.20
FULL       0.40

Caps:
Per hit       12
Per dimension 25 % evidence share

Perception:
Auditory 1.00 | Mixed 0.90 | Visual 0.60
Inner speech 0.85 | Silent reading 0.65
```

# 7. Parameter Registry

```text
Search: [____________________]

frequency_alpha                 0.20   REVIEW REQUIRED
aggregation_scale              80      REVIEW REQUIRED
assessability_threshold         0.75   REVIEW REQUIRED
evidence.B                      0.90   REVIEW REQUIRED
wing.weakest_beta               0.10   REVIEW REQUIRED
ui.confidence_mode              LABEL  SAFE
```

# 8. Test Lab

```text
Text:
┌─────────────────────────────────────────────────────────────┐
│ Ja, aber ich sollte das eigentlich längst können.          │
└─────────────────────────────────────────────────────────────┘

Profile: [CORPORATE_WORKSHOP ▼]
Context: [WORKPLACE ▼]
Resonance: [HINT_ONLY ▼]
Ruleset: [DRAFT v0.2 ▼]

[ANALYSE STARTEN]

Detected:
  ja, aber       DISCOUNTING
  ich sollte     INTERNALIZED_EXPECTATION
  eigentlich     HEDGING
  längst         SELF_PRESSURE

Dimensions:
  Wirksamkeit      38 %   medium confidence
  Freier Wille     31 %   high confidence
  Offenheit        39 %   medium confidence

Contribution Trace                     [expand all]
WingScore: not enough assessable dimensions
```

# 9. Golden Corpus Compare

```text
Case   Production   Draft   Δ     Status
G01       36.7       39.1  +2.4   OK
G09       31.2       45.8 +14.6   ⚠ review
G22        n/a        n/a    —    OK

[Only regressions ✓] [Dimension ▼] [Export Diff]
```

# 10. Publish Flow

```text
DRAFT
  ↓
AUTOMATED TESTS
  ↓
GOLDEN CORPUS
  ↓
REVIEWER APPROVAL
  ↓
PUBLISH
  ↓
MONITOR

[Rollback] always available
```

# 11. Roles

```text
Viewer       read only
Contributor knowledge proposals
Rule Editor  draft rules
Reviewer     approve tests/reviews
Publisher    publish/rollback
Admin        system configuration
```

# 12. MVP-Scope

Für den ersten Admin-MVP:

- Dashboard
- Mapping Manager
- Dimensions
- Rules
- Resonance
- Parameters
- Test Lab
- Rule Sets
- Golden Corpus Compare
- Audit

# 13. Leitgedanke

> **Der Admin soll die Fachlogik verändern können, ohne Programmierer zu sein – aber niemals ohne Transparenz, Test und Rückweg.**
