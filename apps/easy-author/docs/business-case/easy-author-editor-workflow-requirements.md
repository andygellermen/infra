# easy-author – Editor- und Workflow-Anforderungen

## Ziel

Der Editor ist nicht nur ein Texteingabefeld. Er ist der zentrale Arbeitsraum des Autors. easy-author kombiniert einen Tiptap-basierten Typora-artigen Editor mit Seitenleisten, Workflow-Boxen, Ankern, Clipboard-Slots und Autorenkontext.

## Grundlayout

```text
Linke Sidebar | Schreibbereich | Rechte Sidebar
```

### Linke Sidebar

Zweck: Navigation und Werkstruktur.

Inhalte:

- Projekt
- Buch
- Front Matter
- Kapitel
- Szenen
- Back Matter
- Fragmente
- wiederverwendbare Buchseiten
- Export-Sets

### Schreibbereich

Zweck: fokussiertes Schreiben.

Funktionen:

- Tiptap-Editor
- Typora-artige Darstellung
- Markdown-nahe Eingabe
- Kapitelüberschrift
- Autosave-Status
- Wortzählung
- Kapitelstatus
- Shortcut-Unterstützung
- Anker setzen
- Clipboard übernehmen

### Rechte Sidebar

Zweck: Kontext und Autorenworkflow.

Inhalte:

- Workflow-Boxen
- Anker zur aktuellen Textstelle
- Clipboard-Box
- angepinnte Clipboard-Slots 1–9
- Personen/Orte/Ereignisse später
- Assets im aktuellen Kapitel später
- Proofing-Hinweise später

## Tiptap-Entscheidung

easy-author verwendet Tiptap als primäre Editor-Schicht. ProseMirror bleibt als darunterliegendes Dokumentmodell bewusst Teil der Architektur.

Begründung:

- schneller MVP-Fortschritt
- gute Extension-Struktur
- geeignet für typora-/notion-artige Bearbeitung
- Kommentar-, Mention-, Node- und Shortcut-Konzepte gut erweiterbar
- Zugriff auf ProseMirror bleibt für Spezialfälle möglich

## Workflow-Boxen

Workflow-Boxen sind frei konfigurierbare Arbeitscontainer. Sie können links oder rechts sichtbar sein und vom Autor angepasst werden.

### Eigenschaften

- Titel änderbar
- Typ änderbar
- einklappbar
- sortierbar
- optional als Kontextbox rechts
- optional als Navigationsbox links

### Initiale Typen

- notes
- persons
- events
- threads
- reminders
- research
- clipboard
- proofing
- custom

## Anker

Ein Anker verbindet eine Textstelle mit einer Workflow-Box oder einem Workflow-Objekt.

### Ankerarten

- sentence
- passage
- invisible
- manual

### Wichtige Anforderung

Ein Anker darf auch dann gesetzt werden, wenn im Text kein sichtbarer Name, kein Link und keine Markierung erscheinen soll.

Beispiel:

Ein Satz beschreibt eine innere Wandlung der Hauptfigur. Der Name der Figur wird nicht genannt. Trotzdem kann der Satz mit der Workflow-Box „Figurenentwicklung“ verbunden werden.

## Clipboard-Box

Die Clipboard-Box sammelt ausgewählte Inhalte aus dem Editor.

### Funktionen

- markierten Text übernehmen
- Inhalte pinnen
- Slot 1–9 zuweisen
- Slot per Shortcut einfügen
- Quelle optional merken
- Kapitelbezug optional speichern

### Shortcuts

- Cmd/Ctrl + Shift + 1–9: gepinnten Slot einfügen
- Cmd/Ctrl + Shift + C: markierten Text in Clipboard-Box übernehmen
- Cmd/Ctrl + Shift + A: Anker setzen

## Kapitelstatus

Jedes Kapitel erhält einen Status:

- idea
- draft
- revision
- review
- final
- archived

Der Status dient nicht nur der Anzeige, sondern später auch der Export- und Review-Logik.

## Proofing-Hinweise im Editor

Im ersten Schritt sollen Proofing-Hinweise einfach in der rechten Sidebar erscheinen.

Beispiele:

- Kapitel ist leer
- Kapitel ist sehr lang
- TODO im Text gefunden
- Kapitel ohne Status
- Asset ohne Lizenzdaten
- interner Link ungültig

## Vermeidungsstrategien

### Keine Überladung im Editor

Der Schreibbereich bleibt ruhig. Komplexe Funktionen gehören in Seitenleisten oder Dialoge.

### Markdown nicht verschmutzen

Workflow-Informationen werden nicht ungefragt in sichtbares Markdown geschrieben. Anker und Workflowbeziehungen werden separat gespeichert.

### Exportfähigkeit schützen

Editorfunktionen müssen exportierbar oder bewusst als Autorenmaterial markierbar sein.
