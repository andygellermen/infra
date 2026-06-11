# Cody-Prompt 02 – easy-author Atticus-inspirierte UX-Erweiterung

## Kontext

Du arbeitest am Projekt `easy-author`.

Der erste Spike enthält bereits eine grundsätzlich testbare UI mit React, Tiptap, Go-Backend und SQLite. Nun soll die App UX-seitig erweitert werden, inspiriert von Autorenwerkzeugen wie Atticus, jedoch eigenständig und passend zum easy-author-Konzept.

Wichtig: Es geht nicht darum, Atticus zu kopieren. Ziel ist eine eigene, webfähige, Markdown-nahe, self-hosted Autoren-App mit Tiptap-Editor, Workflow-Boxen, Ankern, Clipboard-Box und später Export-/Publishing-Funktionen.

## Ziel dieses zweiten Spikes

Erweitere easy-author so, dass die App stärker wie ein Autoren-Studio wirkt und nicht nur wie ein Editor.

Schwerpunkte:

1. Dashboard/Projektstart
2. Werktypen
3. Front/Body/Back-Matter-Struktur
4. Kapitelstatus
5. wiederverwendbare Buchseiten vorbereiten
6. Proofing-/Export-Checkliste vorbereiten
7. Theme-Grundmodell vorbereiten
8. Dokumentation aktualisieren

## Technische Leitplanken

- Backend: Go
- Frontend: React
- Editor: Tiptap
- Datenbank: SQLite
- Keine Benutzerverwaltung in diesem Schritt
- Kein IndexedDB in diesem Schritt
- Kein echter PDF/EPUB/DOCX-Export in diesem Schritt
- Keine Zahlungsfunktionen
- Keine Cloud-Synchronisierung
- Keine komplexe Kollaboration

## Gewünschte Datenmodell-Erweiterungen

Bitte ergänze oder bereite folgende Modelle vor.

### Project

Erweitern um:

```text
status
last_opened_at
```

Statuswerte:

```text
active
paused
archived
```

### Book

Erweitern um:

```text
work_type
theme_id
cover_asset_id
```

Werktypen:

```text
book
series
freebie
course
article-series
ghost-series
```

### Chapter

Erweitern um:

```text
section_type
status
is_included_in_export
is_visible_in_toc
is_sample_content
```

`section_type`:

```text
front_matter
body
back_matter
fragment
```

`status`:

```text
idea
draft
revision
review
final
archived
```

### MasterPageTemplate / ReusableBookPage

Neues Modell vorbereiten:

```text
id
project_id nullable
title
page_type
content_markdown
content_json
is_global
created_at
updated_at
```

`page_type` Beispiele:

```text
author_bio
copyright
imprint
dedication
review_request
newsletter
also_by
donation_note
publisher_contact
custom
```

Im UI darf dies zunächst einfach als „Wiederverwendbare Buchseiten“ erscheinen.

### Theme

Neues Grundmodell vorbereiten:

```text
id
name
description
target
settings_json
created_at
updated_at
```

`target`:

```text
pdf
epub
docx
web
general
```

### ProofingCheck

Es muss nicht zwingend als persistiertes Modell umgesetzt werden, kann zunächst berechnet werden.

Check-Ergebnis:

```text
id
severity
scope
title
message
status
related_entity_type
related_entity_id
```

Severity:

```text
info
warning
error
```

Status:

```text
open
ignored
resolved
```

## API-Anforderungen

Bitte ergänze sinnvolle REST-Routen.

Mindestens:

```text
GET    /api/dashboard

PUT    /api/projects/:id

PUT    /api/books/:id

GET    /api/books/:bookId/structure

PUT    /api/chapters/:id/options

GET    /api/books/:bookId/reusable-pages
POST   /api/books/:bookId/reusable-pages
PUT    /api/reusable-pages/:id
DELETE /api/reusable-pages/:id

GET    /api/themes
POST   /api/themes
PUT    /api/themes/:id

GET    /api/books/:bookId/proofing-checks
```

Wenn bestehende Routen bereits geeignet sind, verwende sie weiter und erweitere sie sauber.

## Frontend-Anforderungen

### 1. Dashboard

Erstelle oder erweitere eine Dashboard-Ansicht.

Sie soll zeigen:

- Projektkarten
- Buchtitel
- Werktyp
- Status
- zuletzt bearbeitet
- Anzahl Kapitel
- Fortschritt grob

Aktionen:

- neues Projekt/Buch starten
- bestehendes Projekt öffnen
- Projekt archivieren vorbereiten
- Projekt duplizieren optional vorbereiten

### 2. Neues Projekt / neues Werk

Beim Anlegen soll der Nutzer auswählen können:

- Titel
- Autor optional
- Werktyp

Werktypen:

- Buch
- Buchreihe
- Freebie/Leseprobe
- Kurs
- Artikelserie
- Ghost-Serie

### 3. Werkstruktur in linker Sidebar

Die linke Sidebar soll Inhalte gruppieren nach:

```text
Front Matter
Kapitel
Back Matter
Fragmente
```

Kapitel sollen ihren Status anzeigen.

Beispiele:

```text
[Draft] Kapitel 1
[Review] Kapitel 2
[Final] Nachwort
```

### 4. Kapiteloptionen

Ergänze eine einfache Möglichkeit, Kapiteloptionen zu bearbeiten:

- section_type
- status
- is_included_in_export
- is_visible_in_toc
- is_sample_content

Das kann zunächst in einem einfachen Panel oder Dialog erfolgen.

### 5. Wiederverwendbare Buchseiten

Füge einen UI-Bereich hinzu:

```text
Wiederverwendbare Buchseiten
```

Dort sollen einfache Vorlagen erstellt und bearbeitet werden können.

Erste Seitentypen:

- Über den Autor
- Impressum
- Copyright
- Weitere Bücher
- Rezensionsbitte
- Spendenhinweis
- Verlagskontakt
- Eigene Seite

### 6. Theme-Grundmodell / Theme-Platzhalter

Füge eine einfache Theme-Auswahl oder einen Theme-Platzhalter im Buchbereich hinzu.

Für jetzt genügt:

- Theme-Name anzeigen
- Zielmedium anzeigen
- Theme später bearbeiten Hinweis

Noch kein visueller Theme Builder.

### 7. Proofing-/Export-Checkliste

Füge eine einfache Checkliste hinzu, idealerweise rechts oder in einem eigenen Bereich.

Erste Checks:

- Buchtitel fehlt
- Autor fehlt
- Kapitel leer
- Kapitel sehr lang
- doppelter Kapiteltitel
- TODO im Kapiteltext
- kein Impressum
- kein Cover
- Kapitel nicht für Export markiert

Die Checkliste soll statusartige Einträge zeigen:

- Info
- Warnung
- Kritisch

### 8. Editor beibehalten

Der bestehende Tiptap-Editor, Ankerfunktionen, Workflow-Boxen und Clipboard-Funktionen sollen erhalten bleiben.

Bitte keine große Neustrukturierung, wenn nicht nötig.

## UX-Leitlinien

- ruhig und autorenfreundlich
- keine technische Überfrachtung
- Sidebar-Logik erhalten
- neue Funktionen sichtbar, aber nicht dominant
- Schreibfluss nicht stören
- Proofing als Hilfe, nicht als Fehlermeldungswand

## Nicht-Ziele

Bitte ausdrücklich nicht umsetzen:

- echter Export nach PDF/EPUB/DOCX
- DOCX-Import
- PDF-Import
- Benutzerverwaltung
- Rollen/Rechte
- Bezahlung
- Ghost-Publishing
- Cloud-Sync
- Offline-Modus
- visuelles Theme Studio

## Dokumentation

Bitte aktualisiere oder ergänze:

```text
README.md
docs/business-case/easy-author-atticus-usability-benchmark.md
docs/business-case/easy-author-feature-backlog.md
docs/business-case/easy-author-editor-workflow-requirements.md
docs/business-case/easy-author-import-export-requirements.md
docs/business-case/easy-author-theme-studio-requirements.md
docs/business-case/easy-author-proofing-checklist.md
```

Falls die Dateien noch nicht vorhanden sind, lege sie an.

## Akzeptanzkriterien

Nach Abschluss dieses Spikes soll möglich sein:

1. Dashboard öffnen
2. neues Werk mit Werktyp erstellen
3. Werkstruktur mit Front/Body/Back-Matter sehen
4. Kapitelstatus sehen und ändern
5. Kapiteloptionen bearbeiten
6. wiederverwendbare Buchseite anlegen
7. Theme-Platzhalter sehen
8. Proofing-Checkliste sehen
9. Tiptap-Editor weiterhin nutzen
10. bestehende Workflow-Boxen, Anker und Clipboard-Funktionen bleiben erhalten

## Qualitätsanforderung

- kleine, verständliche Komponenten
- keine unnötige Komplexität
- bestehende Funktionen nicht zerstören
- Datenmodell migrationsfähig halten
- UI testbar halten
- README aktualisieren
