# easy-author Review API

## 1. Ziel der Review API

Die `easy-author Review API` ist eine provider-neutrale Schnittstelle für strukturierte Reviews von Texten, Kapiteln, Büchern und Autoren-Workflows.

Sie ermöglicht externen Systemen, KI-Anbietern, lokalen Modellen oder menschlichen Reviewern, Rückmeldungen als strukturierte ReviewItems an `easy-author` zurückzugeben.

Die API verändert Texte nicht direkt. Sie erzeugt Vorschläge, Kommentare und Hinweise, über die der Autor bewusst entscheidet.

## 2. Designprinzipien

- Provider-neutral
- Human-in-Control
- Keine direkte Textänderung ohne Autorentscheidung
- Strukturierte Rückgaben
- Ankerfähige Kommentare
- Kapitel- und buchweite Reviews möglich
- Datenschutzfreundlich durch expliziten Scope
- Erweiterbar für KI, Lektorat, Verlag und menschliche Reviewer

## 3. Kernobjekte

Die Review API arbeitet mit folgenden Kernobjekten:

- `ReviewSession`
- `ReviewRequest`
- `ReviewResponse`
- `ReviewItem`
- `AuthorDecision`
- `ReviewProvider`
- `ReviewScope`
- `ReviewContext`

## 4. ReviewScope

Ein Review kann sich auf unterschiedliche Bereiche beziehen.

Mögliche Scopes:

```text
selection
paragraph
chapter
multiple_chapters
book
asset
workflow_box
export_package
```

## 5. ReviewType

Mögliche Review-Typen:

```text
spelling
grammar
style
structure
readability
consistency
red_thread
chapter_checklist
asset_license
source_check
publication_readiness
sensitivity
custom
```

## 6. ReviewRequest

Ein ReviewRequest beschreibt, was geprüft werden soll.

Beispiel:

```json
{
  "session_id": "rev_sess_001",
  "book_id": "book_001",
  "chapter_id": "chapter_004",
  "scope": "chapter",
  "review_types": ["style", "structure", "red_thread"],
  "language": "de",
  "author_intent": "warm, klar, spirituell, mutmachend, nicht therapeutisch",
  "content_format": "markdown",
  "content": "# Kapitel 4\n\n...",
  "include_context": {
    "book_metadata": true,
    "chapter_metadata": true,
    "anchors": true,
    "workflow_boxes": true,
    "assets": false,
    "comments": false
  },
  "constraints": {
    "no_direct_rewrite": true,
    "suggestions_only": true,
    "preserve_author_voice": true
  }
}
```

## 7. ReviewResponse

Ein ReviewResponse enthält strukturierte ReviewItems.

Beispiel:

```json
{
  "session_id": "rev_sess_001",
  "provider": "openai-compatible",
  "status": "completed",
  "summary": "Das Kapitel ist gut verständlich, enthält aber zwei Brüche im Ton und eine Wiederholung.",
  "review_items": [
    {
      "type": "style_suggestion",
      "severity": "medium",
      "selected_text": "Die Angst wird funktional reduziert.",
      "comment": "Diese Formulierung wirkt technischer als der übrige Text.",
      "suggestion": "Du könntest den Satz wärmer und bildhafter formulieren.",
      "suggested_action": "revise_sentence"
    }
  ]
}
```

## 8. ReviewItem-Typen

Mögliche Typen:

```text
comment
question
warning
correction
style_suggestion
structure_suggestion
alternative_text
consistency_issue
source_issue
asset_issue
workflow_hint
publication_check
```

## 9. Severity

Mögliche Gewichtungen:

```text
info
low
medium
high
critical
```

Beispiele:

- `info`: hilfreicher Hinweis
- `low`: kleine Verbesserung
- `medium`: relevante Überarbeitung empfehlenswert
- `high`: wichtiger Struktur- oder Verständnishinweis
- `critical`: rechtliches, lizenzbezogenes oder veröffentlichungskritisches Thema

## 10. AuthorDecision

Jedes ReviewItem kann eine Autorentscheidung erhalten.

Mögliche Entscheidungen:

```text
pending
accepted
rejected
deferred
converted_to_note
converted_to_anchor
converted_to_task
copied_to_clipboard
asked_followup
resolved
```

## 11. API-Routen

### Health

```http
GET /api/review/health
```

### Review Sessions

```http
GET    /api/review/sessions
POST   /api/review/sessions
GET    /api/review/sessions/:id
PUT    /api/review/sessions/:id
DELETE /api/review/sessions/:id
```

### Start Review

```http
POST /api/review/sessions/:id/run
```

### Review Items

```http
GET  /api/review/sessions/:id/items
POST /api/review/sessions/:id/items
GET  /api/review/items/:id
PUT  /api/review/items/:id
```

### Author Decisions

```http
POST /api/review/items/:id/decision
GET  /api/review/items/:id/decisions
```

### Provider Adapter

```http
GET  /api/review/providers
POST /api/review/providers/test
```

## 12. Provider Adapter Interface

Provider Adapter sollen austauschbar sein.

Konzeptionelles Interface:

```go
type ReviewProvider interface {
    Name() string
    Capabilities() []ReviewCapability
    RunReview(ctx context.Context, req ReviewRequest) (ReviewResponse, error)
}
```

## 13. Speicherung

Die Review API speichert:

- ReviewSession
- ReviewRequest-Metadaten
- ReviewItems
- AuthorDecisions
- Provider-Metadaten
- Statusinformationen

Ob vollständige Inhalte gespeichert werden, muss konfigurierbar sein.

Empfohlener Standard:

- ReviewItems speichern
- Entscheidungen speichern
- gesendete Volltexte nicht dauerhaft speichern, außer Autor aktiviert Review-Archiv

## 14. Datenschutzstufen

Mögliche Einstellungen:

```text
no_ai
local_only
selection_only
chapter_with_confirmation
book_with_explicit_consent
external_provider_allowed
```

## 15. Nicht-Ziele der ersten Umsetzung

- Keine automatische Textänderung
- Keine direkte Provider-Bindung
- Kein vollständiger KI-Chat
- Keine automatische Veröffentlichung
- Kein automatisches Buchlektorat im Hintergrund
- Keine versteckte Datenübertragung

## 16. MVP-Empfehlung

Erste Umsetzung:

- ReviewSession anlegen
- ReviewItems manuell oder durch MockProvider erzeugen
- ReviewItems in rechter Sidebar anzeigen
- AuthorDecision speichern
- ReviewItem in Note, Anchor, Task oder Clipboard überführen
- Provider-Schnittstelle vorbereiten, aber noch keinen externen Provider erzwingen

