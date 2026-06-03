# easy-author – Workflow Instances & Author Cockpit

## 1. Ziel

Workflow-Instanzen machen easy-author zu mehr als einem Editor. Sie sind frei konfigurierbare Arbeitsboxen, die den Autor beim Schreiben, Erinnern, Strukturieren, Wiederverwenden und Überarbeiten unterstützen.

Eine Workflow-Box kann eine klassische Notizbox sein, aber auch eine Personenliste, eine Ereignissammlung, eine Clipboard-Ablage, eine Kapitel-Checkliste, ein Recherchefenster oder ein Lektoratswerkzeug.

## 2. Grundprinzip

```text
Workflow-Box = frei benennbare, konfigurierbare Autorenfunktion
```

Jede Box kann:

- einen eigenen Titel haben
- eine eigene Funktion haben
- links oder rechts platziert werden
- eingeklappt oder sichtbar sein
- angepinnt werden
- eigene Shortcuts erhalten
- mit Kapiteln, Absätzen, Sätzen oder Textpassagen verknüpft werden
- automatisch Inhalte aus dem aktuellen Kapitel vorschlagen

## 3. Warum das wichtig ist

Autoren denken nicht immer linear. Während des Schreibens entstehen:

- Nebenfiguren
- offene Fragen
- lose Formulierungen
- Recherchepunkte
- Erinnerungen
- Motive
- wiederverwendbare Textbausteine
- Kapitelideen
- Verlags- oder Marketinggedanken

Diese Gedanken dürfen den Schreibfluss nicht bremsen. Sie müssen schnell abgelegt und später intelligent wieder auffindbar sein.

## 4. Box-Typen

### Strukturboxen

- Buchstruktur
- Kapitel
- Szenen
- Fragmente
- Export-Sets

### Wissensboxen

- Personen
- Orte
- Ereignisse
- Handlungsstränge
- Motive
- Begriffe
- Quellen

### Workflowboxen

- Schnellnotizen
- offene Fragen
- Erinnerungen
- Kapitelaufgaben
- Recherche
- Lektorat
- Stilprüfung
- Roter-Faden-Monitor

### Spezialboxen

- Clipboard
- Asset-Verwendung
- Kommentarübersicht
- Testleserfeedback
- Veröffentlichungsstatus
- Monetarisierung

## 5. Verknüpfung ohne Namensnennung

Ein wichtiges Prinzip: Eine Workflow-Verknüpfung darf nicht davon abhängig sein, dass ein Name oder Begriff im Text sichtbar erwähnt wird.

Beispiele:

```text
Ein Absatz kann mit „Handlungsstrang: Angst“ verbunden werden,
auch wenn das Wort Angst dort gar nicht vorkommt.
```

```text
Ein Satz kann mit „Figurentwicklung: Mara wird mutiger“ verbunden werden,
auch wenn Mara im Satz nicht genannt wird.
```

```text
Eine Textpassage kann mit „später lektorieren“ verbunden werden,
ohne sichtbares TODO im Manuskript.
```

## 6. Anker-Modell

### Sichtbare Anker

```markdown
[[Mara]]
[[Ort:Alter Garten]]
[[Motiv:Angst]]
```

### Unsichtbare Anker

Unsichtbare Anker entstehen durch Markieren einer Textstelle und Verknüpfen mit einer Workflow-Box.

```text
Text markieren
→ Anker setzen
→ Workflow-Box wählen
→ optional Notiz ergänzen
```

### Anchor Entity

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

## 7. WorkflowLink Entity

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

## 8. Clipboard-Box

Die Clipboard-Box ist eine besondere Workflow-Box. Sie sammelt bewusst kopierte Inhalte aus dem easy-author-Editor.

### Funktionen

- kopierte Textteile sammeln
- Ursprung speichern
- Eintrag benennen
- Eintrag anpinnen
- per Shortcut abrufen
- Eintrag mit Kapitel oder Projekt verknüpfen
- als Textbaustein wiederverwenden

### Shortcut-Logik

```text
Cmd/Ctrl + Shift + 1 bis 9
```

Die Slots 1 bis 9 sind bewusst begrenzt. Das hält den Workflow schnell und verhindert eine überladene Shortcut-Logik.

### ClipboardItem Entity

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

## 9. Intelligente Hilfe

Workflow-Boxen sollten nicht nur passiv anzeigen, sondern aktiv helfen.

Mögliche Vorschläge:

- „Diese Passage ähnelt einer früheren Notiz.“
- „Dieser Absatz könnte zum Handlungsstrang X gehören.“
- „Diese Figur wurde länger nicht weiterentwickelt.“
- „Diese Quelle ist noch nicht vollständig dokumentiert.“
- „Dieser Clipboard-Eintrag wurde mehrfach verwendet.“
- „Dieses Kapitel hat offene Anker ohne Klärung.“

## 10. MVP-Schnitt

Für den ersten MVP sollten enthalten sein:

- Workflow-Box-Grundmodell
- frei benennbare Boxen
- linke/rechte Platzierung
- Collapse/Pin-Funktion
- Satz-/Passagenanker
- Clipboard-Box mit 9 Slots
- Schnellnotiz-Box
- Verknüpfung mit Kapiteln

Später:

- KI-gestützte Vorschläge
- automatische Inkonsistenzprüfung
- Testleser-Anker
- erweiterte Review-Workflows
- kollaborative Boxen pro Rolle
