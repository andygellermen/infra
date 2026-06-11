# easy-author – Editor UI Concept

## 1. Architekturentscheidung

easy-author verwendet **Tiptap als primäres Editor-Framework**. Tiptap liefert die produktive Extension-, Command- und UI-Schicht. ProseMirror bleibt als darunterliegendes Dokumentmodell bewusst Teil der Architektur und wird bei Spezialfunktionen direkt genutzt.

```text
easy-author Editor UI
  ↓
Tiptap Extensions, Commands, Menüs, Shortcuts, Nodes, Marks
  ↓
ProseMirror Document Model, Plugins, Transactions, Positions
  ↓
Markdown / JSON / HTML / Export-Pipeline
```

### Entscheidungssatz

> easy-author verwendet Tiptap als produktives, modernes Editor-Framework auf Basis von ProseMirror. Tiptap liefert die primäre Entwickler- und Extension-Schicht für das Autoren-UI. ProseMirror bleibt als zugrunde liegendes Dokumentmodell und als Zugriffsebene für Spezialfunktionen bewusst Teil der Architektur.

## 2. Ziel

Der Editor soll sich wie ein ruhiger, moderner Schreibraum anfühlen: so direkt wie Typora, aber organisatorisch stark wie ein Scrivener-artiges Autorenstudio.

Der Autor schreibt nicht in einem technischen Formular, sondern in einem lebendigen Manuskript. Markdown bleibt erhalten, wird aber visuell elegant aufgelöst.

## 3. Grundlayout

```text
┌──────────────────────┬───────────────────────────────┬──────────────────────┐
│ Linke Sidebar         │ Editor                         │ Rechte Sidebar        │
│ Struktur              │ Typora-artiges Schreiben       │ Kontext & Workflow    │
├──────────────────────┼───────────────────────────────┼──────────────────────┤
│ Buch                  │ Kapiteltext                    │ Workflow-Boxen        │
│ Kapitel               │ Markdown visuell gerendert     │ Personen              │
│ Szenen                │ Slash/Command-Menü             │ Orte                  │
│ Notizen               │ Inline-Links                   │ Ereignisse            │
│ Recherche             │ Anker & Markierungen           │ Handlungsstränge      │
│ Assets                │ Clipboard-Anbindung            │ Erinnerungen          │
│ Timeline              │ Fokusmodus                     │ Kommentare            │
└──────────────────────┴───────────────────────────────┴──────────────────────┘
```

## 4. Linke Sidebar

Die linke Sidebar dient der Navigation und Struktur.

### Bereiche

- Projektübersicht
- Buchstruktur
- Kapitel
- Szenen/Fragmente
- Notizen
- Recherche
- Assets
- Timeline
- Exporte
- Versionen
- Workflow-Boxen

### Verhalten

- einklappbar
- pinnbar
- Drag & Drop für Kapitel und Boxen
- Statusanzeige je Kapitel
- Wortanzahl je Kapitel
- Suchfeld
- globale Schnellfilter

## 5. Editor-Mitte

Die Mitte ist der eigentliche Schreibraum.

### Anforderungen

- Typora-artige Markdown-Bearbeitung
- Markdown wird visuell gerendert
- Roh-Markdown-Modus optional
- Überschriften, Listen, Zitate, Code, Tabellen visuell angenehm
- Bilder inline sichtbar
- interne Links `[[...]]`
- Anker auf Sätze, Absätze und Textpassagen
- Workflow-Markierungen ohne sichtbaren Namen erforderlich
- Fokusmodus
- Autosave-Status sichtbar

### Schreibmodi

- Rohtextmodus
- Überarbeitungsmodus
- Lektoratsmodus
- Finalisierungsmodus
- Export-Prüfmodus

## 6. Rechte Sidebar

Die rechte Sidebar zeigt den Kontext des aktuellen Kapitels und die aktiven Workflow-Elemente.

### Bereiche

- aktive Workflow-Boxen
- im Kapitel verwendete Personen
- erwähnte Orte
- Ereignisse
- Handlungsstränge
- offene Erinnerungen
- Assets im Kapitel
- Clipboard-Box
- Kommentare
- Kapitel-Checkliste
- Qualitätswarnungen

### Besonderheit

Die rechte Sidebar darf nicht nur aus sichtbaren Begriffen wie `[[Mara]]` gespeist werden. Sie muss auch Textpassagen, Sätze, Absätze und frei gesetzte Anker auswerten können.

Beispiel:

```markdown
Der Garten war still, doch in Mara begann etwas zu kippen.
```

Dieser Satz kann mit einem Workflow-Element verbunden werden, ohne dass der Autor sichtbaren Markdown einfügen muss.

## 7. Workflow-Boxen

Workflow-Boxen sind frei benennbare, frei konfigurierbare Autoren-Werkzeuge.

### Beispiele

- Personen
- Orte
- Ereignisse
- Handlungsstränge
- offene Fragen
- Erinnerungen
- Recherche
- Stilhinweise
- Kapitelaufgaben
- Motive
- Quellen
- Zitate
- Clipboard
- Lektoratsnotizen
- Marketingideen

### Eigenschaften

```yaml
id: uuid
title: string
box_type: person|place|event|thread|note|asset|clipboard|custom
function: context|collection|checklist|review|reference|automation|custom
position: left|right
collapsed: boolean
pinned: boolean
shortcut: string|null
filters: json
settings: json
```

### Verhalten

- Titel änderbar
- Funktion änderbar
- links oder rechts platzierbar
- einklappbar
- pinnbar
- per Tastatur erreichbar
- mit Kapiteln, Absätzen, Sätzen oder Textpassagen verknüpfbar
- kann automatisch relevante Inhalte vorschlagen

## 8. Anker und Textpassagen-Verknüpfung

Neben expliziten Links wie `[[Mara]]` braucht easy-author stille, unsichtbare Anker.

### Ankertypen

- Kapitelanker
- Absatzanker
- Satzanker
- Textpassagenanker
- Kommentaranker
- Assetanker
- Workflowanker

### Use Cases

- Satz mit Handlungsstrang verbinden
- Absatz mit Recherche-Notiz verbinden
- Passage als später zu überarbeiten markieren
- Motiv an einer Stelle verankern
- Testleser-Kommentar an einer exakten Passage halten
- Asset oder Quelle mit Textstelle verbinden

### Beispielhafte Bedienung

```text
Text markieren
→ Kontextmenü öffnen
→ „Anker setzen“ oder „Mit Workflow-Box verbinden“
→ Box wählen oder neue Box anlegen
```

### Technischer Hinweis

Anker sollten nicht allein über Zeichenpositionen gespeichert werden. Zeichenpositionen sind fragil, sobald Text bearbeitet wird. Besser ist ein hybrides Modell:

```text
chapter_id
block_id
text_quote
text_hash
prosemirror_position
fallback_context_before
after_context
```

Damit kann ein Anker auch nach Textänderungen meist wiedergefunden werden.

## 9. Clipboard-Box

Die Clipboard-Box sammelt bewusst kopierte Inhalte aus dem Tiptap-Editor und macht sie später schnell wiederverwendbar.

### Grundidee

Wenn der Autor im Editor Text, Zitate, Absätze, Ideen oder Formulierungen kopiert, kann easy-author diese Ausschnitte in einer Clipboard-Box sammeln.

### Funktionen

- automatische Erfassung editorinterner Kopiervorgänge
- manuelle Übernahme per Shortcut
- Einträge anpinnen
- Einträge benennen
- Einträge taggen
- Ursprungskapitel speichern
- ursprüngliche Textstelle verlinken
- per Shortcut einfügen

### Shortcuts

```text
Cmd/Ctrl + Shift + 1 → gepinnten Clipboard-Eintrag 1 einfügen
Cmd/Ctrl + Shift + 2 → gepinnten Clipboard-Eintrag 2 einfügen
...
Cmd/Ctrl + Shift + 9 → gepinnten Clipboard-Eintrag 9 einfügen
```

### ClipboardItem

```yaml
id: uuid
project_id: uuid
chapter_id: uuid|null
title: string
content_markdown: text
content_json: json|null
source_anchor_id: uuid|null
pinned_slot: 1|2|3|4|5|6|7|8|9|null
tags: string[]
created_at: datetime
updated_at: datetime
```

### Datenschutz und UX

Die Clipboard-Box sollte standardmäßig nur Kopiervorgänge innerhalb des easy-author-Editors erfassen. Systemweite Clipboard-Überwachung wäre technisch und datenschutzseitig heikel und sollte nicht Teil des MVP sein.

## 10. Schnellnotiz/Inbox

Ein zentrales Element ist eine schnelle Ablage für Gedanken.

### Shortcut

```text
Cmd/Ctrl + J → Schnellnotiz erfassen
```

### Notiztypen

- Gedanke
- später prüfen
- Figuridee
- Kapitelidee
- Recherche
- offene Entscheidung
- Marketingidee

## 11. Kapitel-Checkliste

Pro Kapitel kann eine Checkliste gepflegt werden.

Beispiel für Sachbuch:

- Hintergrund vorhanden
- Übung vorhanden
- Weitblick vorhanden
- Wort-Notiz vorhanden
- Quellen geprüft
- Bilder lizenziert
- Export getestet

## 12. Roter-Faden-Monitor

Der Autor kann Leitplanken definieren:

- Hauptaussage
- Tonalität
- Zielgruppe
- wiederkehrende Begriffe
- verbotene Begriffe
- Kapitelmuster

Die rechte Sidebar kann Hinweise anzeigen, wenn ein Kapitel davon abweicht.

## 13. UX-Prinzipien

- Schreiben zuerst, Verwaltung danach
- Seitenleisten helfen, sie dominieren nicht
- Workflow-Boxen sind flexibel, nicht starr
- keine überladene Oberfläche
- Seitenleisten einklappbar
- keine modale Komplexität beim Schreiben
- alles schnell per Tastatur erreichbar
- Markdown bleibt sichtbar, aber nicht störend
- wichtige Passagen können verankert werden, ohne den Lesefluss zu stören

## 14. Technischer Kandidat

Primärer Kandidat:

- Tiptap mit React

Darunter bewusst genutzt:

- ProseMirror Dokumentmodell
- ProseMirror Positions-/Transaction-System
- eigene Tiptap Extensions
- eigene Markdown-Serialisierung für easy-author-Spezialelemente

Nicht favorisiert für den Start:

- ProseMirror direkt als alleinige Editor-Implementierung
- MDXEditor als langfristiger Kern

## Ergänzung: Atticus-inspirierte UX-Erweiterung

Die Editor-UI wird um eine stärkere Autoren-Studio-Struktur ergänzt:

- Dashboard als Einstieg
- Werktypen beim Projektstart
- linke Sidebar mit Front Matter, Kapitel, Back Matter und Fragmenten
- Kapitelstatus direkt in der Struktur sichtbar
- Kapiteloptionen im Seitenpanel oder Dialog
- rechte Sidebar mit Workflow-Boxen, Clipboard, Ankern und Proofing-Hinweisen
- Theme-Platzhalter im Buchbereich
- wiederverwendbare Buchseiten als eigener Navigationspunkt

Der Tiptap-Schreibbereich bleibt ruhig und wird nicht durch Proofing- oder Exportlogik überfrachtet.

