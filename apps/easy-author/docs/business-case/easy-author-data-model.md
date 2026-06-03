# easy-author – Data Model Draft

## 1. Grundsatz

Die Inhalte sollen offen und langlebig bleiben. Markdown-Dateien, Assets und Themes dürfen nicht in einer proprietären Datenbank verschwinden. Die Datenbank verwaltet Metadaten, Beziehungen, Rechte, Kommentare und Jobs.

## 2. Speicherstrategie

### Dateisystem

- Markdown-Kapitel
- Assets
- Themes
- Export-Artefakte
- Projekt-Sicherungen

### Datenbank

- Benutzer
- Projekte
- Kapitel-Metadaten
- Wissensbank-Objekte
- Asset-Metadaten
- Kommentare
- Rechte
- Export-Jobs
- Planer

## 3. Projektstruktur auf Dateiebene

```text
/project-id
  project.yml
  manuscript.yml
  /chapters
    001-vorwort.md
    002-kapitel-1.md
  /notes
  /assets
    /original
    /optimized
  /themes
    pdf-default.css
    epub-default.css
  /exports
  /snapshots
```

## 4. Zentrale Entitäten

### User

```yaml
id: uuid
name: string
email: string
status: active|inactive
created_at: datetime
```

### Project

```yaml
id: uuid
owner_id: uuid
title: string
subtitle: string
author_name: string
language: string
status: draft|review|published|archived
visibility: private|registered|public|paid
created_at: datetime
updated_at: datetime
```

### Chapter

```yaml
id: uuid
project_id: uuid
title: string
slug: string
path: string
sort_order: integer
status: idea|draft|review|final
word_count: integer
created_at: datetime
updated_at: datetime
```

### KnowledgeItem

```yaml
id: uuid
project_id: uuid
type: person|location|event|thread|motif|term|reminder|research_note
name: string
summary: text
body: text
tags: string[]
created_at: datetime
updated_at: datetime
```

### KnowledgeLink

```yaml
id: uuid
project_id: uuid
knowledge_item_id: uuid
chapter_id: uuid
anchor_text: string
position: json
created_at: datetime
```

### Asset

```yaml
id: uuid
project_id: uuid
file_path: string
mime_type: string
title: string
description: text
alt_text: text
source: string
author: string
license_name: string
license_url: string
rights_note: text
created_at: datetime
updated_at: datetime
```

### AssetUsage

```yaml
id: uuid
asset_id: uuid
chapter_id: uuid
usage_type: inline|cover|gallery|export_only
created_at: datetime
```

### Comment

```yaml
id: uuid
project_id: uuid
chapter_id: uuid
author_id: uuid
visibility: author_only|review_group|public
comment_type: note|question|correction|review
body: text
anchor: json
status: open|resolved|archived
created_at: datetime
updated_at: datetime
```

### ExportProfile

```yaml
id: uuid
project_id: uuid
name: string
format: pdf|epub|docx|markdown
page_format: A4|A5|Letter|custom
margins: json
theme_id: uuid
settings: json
```

### ExportJob

```yaml
id: uuid
project_id: uuid
profile_id: uuid
status: queued|running|success|failed
result_path: string
log: text
created_at: datetime
finished_at: datetime
```

### PlannerEvent

```yaml
id: uuid
project_id: uuid
title: string
event_type: writing|review|export|publish|marketing|custom
start_at: datetime
end_at: datetime
status: planned|done|cancelled
```

## 5. Beziehungen

```text
User 1:n Project
Project 1:n Chapter
Project 1:n KnowledgeItem
Project 1:n Asset
Project 1:n ExportProfile
Project 1:n PlannerEvent
Chapter n:m KnowledgeItem via KnowledgeLink
Chapter n:m Asset via AssetUsage
Chapter 1:n Comment
```

## 6. Markdown-Verlinkung

Interne Links sollten einfach bleiben:

```markdown
[[Mara]]
[[Ort:Alter Garten]]
[[Ereignis:Erste Begegnung]]
```

Beim Speichern kann ein Parser diese Links erkennen und KnowledgeLinks aktualisieren.

## 7. Vermeidungsstrategie

- Markdown bleibt primäre Inhaltsquelle.
- Datenbank speichert Beziehungen, nicht den alleinigen Textwert.
- Server-Speicher ist die Wahrheit.
- Browser-Speicher dient nur für Offline/Autosave.
- Export-Artefakte sind reproduzierbar und nicht die Primärquelle.

## 8. Ergänzung: Workflow-Boxen und Anker

### WorkflowBox

```yaml
id: uuid
project_id: uuid
title: string
box_type: person|place|event|thread|note|asset|clipboard|review|custom
function: context|collection|checklist|reference|review|automation|custom
position: left|right
sort_order: integer
collapsed: boolean
pinned: boolean
shortcut: string|null
settings: json
created_at: datetime
updated_at: datetime
```

### Anchor

```yaml
id: uuid
project_id: uuid
chapter_id: uuid
anchor_type: chapter|block|sentence|selection|comment|asset|workflow
block_id: string|null
text_quote: text
text_hash: string
position_json: json
context_before: text
context_after: text
created_at: datetime
updated_at: datetime
```

### WorkflowLink

```yaml
id: uuid
project_id: uuid
workflow_box_id: uuid
anchor_id: uuid|null
knowledge_item_id: uuid|null
asset_id: uuid|null
comment_id: uuid|null
link_type: mention|anchor|reference|task|review|source|clipboard
status: active|resolved|archived
created_at: datetime
updated_at: datetime
```

### ClipboardItem

```yaml
id: uuid
project_id: uuid
chapter_id: uuid|null
source_anchor_id: uuid|null
title: string
content_markdown: text
content_json: json|null
plain_text: text
pinned_slot: integer|null
tags: string[]
created_at: datetime
updated_at: datetime
```

## 9. Ergänzung: Speicherformate

Pro Kapitel wird ein hybrides Modell empfohlen:

```text
/chapters/001-vorwort.md
/editor-state/001-vorwort.tiptap.json
```

Markdown bleibt das primäre Autoren- und Exportformat. Tiptap/ProseMirror JSON dient als Editor-Snapshot für die genaue Wiederherstellung von Editorzustand, Markierungen, Custom-Nodes und komplexeren Workflow-Verknüpfungen.
