# easy-author AI Review Data Model

## 1. Ziel

Dieses Dokument ergänzt das bestehende Datenmodell von `easy-author` um Review-Funktionen.

Die neuen Objekte ermöglichen KI-gestützte oder menschliche Reviews, ohne dass Textänderungen automatisch übernommen werden.

## 2. Neue Entitäten

- ReviewSession
- ReviewItem
- AuthorDecision
- ReviewProvider
- ReviewRequestLog

## 3. ReviewSession

Eine ReviewSession bündelt einen Review-Vorgang.

### Felder

```text
id
book_id
chapter_id
scope
status
provider_key
review_types
language
author_intent
privacy_level
created_by
created_at
updated_at
started_at
completed_at
summary
error_message
```

### Feldbeschreibung

| Feld | Beschreibung |
|---|---|
| `id` | Eindeutige ID der ReviewSession |
| `book_id` | Zugeordnetes Buch |
| `chapter_id` | Optional zugeordnetes Kapitel |
| `scope` | selection, chapter, book usw. |
| `status` | draft, running, completed, failed, cancelled |
| `provider_key` | genutzter Provider oder MockProvider |
| `review_types` | Liste der Review-Typen |
| `language` | Sprache, z. B. de |
| `author_intent` | gewünschter Stil / Zielrichtung |
| `privacy_level` | Datenschutzstufe |
| `summary` | Zusammenfassung des Reviews |
| `error_message` | Fehlertext bei fehlgeschlagenem Review |

### Statuswerte

```text
draft
waiting_for_consent
running
completed
failed
cancelled
archived
```

## 4. ReviewItem

Ein ReviewItem ist ein einzelner Kommentar, Hinweis oder Vorschlag.

### Felder

```text
id
review_session_id
book_id
chapter_id
anchor_id
workflow_box_id
type
severity
status
selected_text
start_offset
end_offset
context_before
context_after
comment
suggestion
alternative_text
rationale
suggested_action
created_at
updated_at
```

### Feldbeschreibung

| Feld | Beschreibung |
|---|---|
| `id` | Eindeutige ID |
| `review_session_id` | Zugehörige ReviewSession |
| `book_id` | Zugehöriges Buch |
| `chapter_id` | Zugehöriges Kapitel |
| `anchor_id` | Optionaler Textanker |
| `workflow_box_id` | Optionale Workflow-Box |
| `type` | Art des ReviewItems |
| `severity` | Gewichtung |
| `status` | Bearbeitungsstatus |
| `selected_text` | betroffener Textauszug |
| `start_offset` | Startposition, wenn verfügbar |
| `end_offset` | Endposition, wenn verfügbar |
| `context_before` | Kontext vor der Textstelle |
| `context_after` | Kontext nach der Textstelle |
| `comment` | Review-Kommentar |
| `suggestion` | Verbesserungsvorschlag |
| `alternative_text` | konkrete alternative Formulierung |
| `rationale` | Begründung des Hinweises |
| `suggested_action` | empfohlene Folgeaktion |

### Typen

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
custom
```

### Severity

```text
info
low
medium
high
critical
```

### Status

```text
open
accepted
rejected
deferred
converted
resolved
archived
```

## 5. AuthorDecision

Eine AuthorDecision dokumentiert die bewusste Entscheidung des Autors zu einem ReviewItem.

### Felder

```text
id
review_item_id
decision
decision_note
result_anchor_id
result_workflow_box_id
result_clipboard_item_id
result_chapter_id
created_by
created_at
```

### Entscheidungstypen

```text
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

### Bedeutung

| Entscheidung | Bedeutung |
|---|---|
| `accepted` | Vorschlag wird übernommen oder als akzeptiert markiert |
| `rejected` | Vorschlag wird abgelehnt |
| `deferred` | später prüfen |
| `converted_to_note` | als Autoren-Notiz speichern |
| `converted_to_anchor` | als Anker speichern |
| `converted_to_task` | als Workflow-Aufgabe speichern |
| `copied_to_clipboard` | in Clipboard-Box übernehmen |
| `asked_followup` | Folgefrage an Review-Partner/KI gestellt |
| `resolved` | erledigt ohne direkte Übernahme |

## 6. ReviewProvider

Ein ReviewProvider beschreibt eine mögliche Review-Quelle.

### Felder

```text
id
key
name
provider_type
is_enabled
is_external
supports_structured_output
supports_streaming
supports_local_only
created_at
updated_at
```

### provider_type

```text
mock
openai_compatible
local_model
human_reviewer
custom_webhook
```

## 7. ReviewRequestLog

Optionales Log für Transparenz.

### Felder

```text
id
review_session_id
provider_key
scope
content_hash
sent_content_stored
request_metadata_json
response_metadata_json
created_at
```

Empfehlung:

- Volltexte nicht standardmäßig speichern.
- Hash und Metadaten speichern.
- Volltext-Archiv nur nach bewusster Aktivierung durch den Autor.

## 8. Beziehungen

```text
Book 1:n ReviewSession
Chapter 1:n ReviewSession
ReviewSession 1:n ReviewItem
ReviewItem 1:n AuthorDecision
ReviewItem n:1 Anchor optional
ReviewItem n:1 WorkflowBox optional
ReviewProvider 1:n ReviewSession
```

## 9. Integration mit bestehenden Objekten

### Anchor

Ein ReviewItem kann einen bestehenden Anchor nutzen oder nach Autorentscheidung einen neuen Anchor erzeugen.

### WorkflowBox

ReviewItems können in WorkflowBoxen überführt werden, z. B.:

- Überarbeitung
- Quellen prüfen
- Stil prüfen
- Leitmotive
- Parkplatz

### ClipboardItem

Alternative Formulierungen können in die Clipboard-Box übernommen und bei Bedarf einem Slot 1–9 zugeordnet werden.

### Chapter

Kapitel können einen ReviewStatus erhalten:

```text
not_reviewed
review_requested
reviewed
changes_pending
approved
```

## 10. Migrationsempfehlung für SQLite MVP

Neue Tabellen:

```sql
review_sessions
review_items
author_decisions
review_providers
review_request_logs
```

Die Tabellen sollten unabhängig von externen KI-Anbietern funktionieren.

Im ersten Schritt kann ein `mock` Provider ReviewItems lokal erzeugen, um UI und Datenfluss zu testen.

