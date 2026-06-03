# easy-author – Storage Format Strategy

## 1. Grundsatz

Für easy-author bleibt **Markdown das primäre Autorenformat**. Markdown ist lesbar, portabel, langlebig und passt zur Export-Pipeline.

Gleichzeitig benötigt ein moderner Tiptap-Editor intern ein strukturiertes Dokumentmodell. Deshalb wird ein hybrides Speicherformat empfohlen.

## 2. Drei mögliche Strategien

### Variante A: Markdown-first

Markdown-Dateien sind die alleinige Wahrheit.

```text
Editor → Markdown → Datei
```

#### Vorteile

- maximal portabel
- einfach versionierbar mit Git
- menschenlesbar
- unabhängig vom Editor
- ideal für Pandoc, EPUB, PDF, DOCX
- keine starke Bindung an Tiptap

#### Nachteile

- komplexe Editorzustände schwer speicherbar
- exakte Kommentaranker fragiler
- Workflow-Verknüpfungen müssen separat gespeichert werden
- Custom-Elemente brauchen Konventionen
- Roundtrip zwischen Editor und Markdown kann bei Spezialelementen Reibung erzeugen

#### Passung

Sehr gut für klassische Kapiteltexte und langfristige Archivierung.

### Variante B: JSON-first

Tiptap/ProseMirror JSON ist die primäre Wahrheit.

```text
Editor → ProseMirror JSON → Datenbank/Datei
```

#### Vorteile

- exakt passend zum Editorzustand
- stabile Abbildung von Nodes, Marks und Custom Extensions
- Kommentare, Anker und Spezialelemente genauer modellierbar
- bessere Grundlage für kollaboratives Schreiben
- weniger Verlust zwischen Editor und Speicher

#### Nachteile

- für Menschen schlechter lesbar
- stärker abhängig vom Editor-Schema
- Git-Diffs sind schlechter verständlich
- Export nach Markdown muss aktiv gepflegt werden
- langfristige Portabilität schwächer als bei Markdown

#### Passung

Sehr gut für komplexe Editorfunktionen, Kommentare, Kollaboration und exakte Anker.

### Variante C: Hybrid

Markdown bleibt die Autoren- und Export-Wahrheit. JSON wird zusätzlich als Editor-Snapshot gespeichert.

```text
Editor
  ├─ Markdown Snapshot
  ├─ ProseMirror JSON Snapshot
  └─ Metadaten/Anker/Workflow in DB
```

#### Vorteile

- Markdown bleibt offen, lesbar und exportierbar
- JSON erhält den exakten Editorzustand
- Anker und Workflow-Elemente werden stabiler
- gute Grundlage für spätere Kollaboration
- bessere Wiederherstellung bei komplexen Editorzuständen
- weniger Lock-in als JSON-only

#### Nachteile

- Synchronisationslogik notwendig
- Konflikte zwischen Markdown und JSON müssen geregelt sein
- Speicher- und Testaufwand höher
- Import/Export-Regeln müssen sauber definiert werden

#### Passung

Sehr gut für easy-author, weil das Produkt sowohl langlebige Markdown-Dateien als auch moderne Autoren-Workflows braucht.

## 3. Empfohlene Entscheidung

```text
Primär für Autor und Export: Markdown
Primär für Editor-Wiederherstellung: Tiptap/ProseMirror JSON
Primär für Beziehungen: Datenbank
Primär für Langzeitarchiv: Markdown + Assets + project.yml
```

## 4. Praktisches Modell

Pro Kapitel:

```text
/chapters/001-vorwort.md
/editor-state/001-vorwort.tiptap.json
```

Zusätzlich in der Datenbank:

- Kapitel-Metadaten
- Workflow-Boxen
- Anker
- Kommentare
- KnowledgeLinks
- AssetUsage
- Export-Jobs

## 5. Speicher-Taktung

### Sofort-Snapshot im Browser

```text
alle 1–3 Sekunden bei Änderung
```

Ziel:

- Schutz vor Tab-Absturz
- Offline-Fortsetzung
- schneller Autosave

Speicherort:

- IndexedDB

### Server-Autosave

```text
alle 5–15 Sekunden oder nach Ruhephase
```

Ziel:

- stabile serverseitige Sicherung
- geräteübergreifendes Arbeiten
- Schutz vor Browserverlust

### Markdown-Snapshot

```text
bei Pause, Kapitelwechsel, manuellem Speichern oder alle 30–60 Sekunden
```

Ziel:

- lesbare Datei aktuell halten
- Git-freundliche Arbeitsstände
- Exportbereitschaft

### Version/Commit

```text
manuell oder nach Schreibsession
```

Ziel:

- bewusste Meilensteine
- Wiederherstellung
- Vergleich von Fassungen

### Export-Snapshot

```text
bei jedem Export
```

Ziel:

- reproduzierbarer Export
- Nachweis, welche Fassung exportiert wurde

## 6. Konfliktstrategie

Konflikte können entstehen, wenn:

- mehrere Geräte offline schreiben
- Markdown extern geändert wurde
- JSON und Markdown nicht mehr zueinander passen
- Cloud-Sync alte Fassungen zurückspielt

### Vermeidungsstrategie

- Kapitel beim Bearbeiten weich sperren
- Änderungshistorie speichern
- externe Markdown-Änderungen erkennen
- Konfliktkopie erzeugen statt überschreiben
- Diff-Ansicht anbieten
- JSON nie exportieren, ohne Markdown neu zu synchronisieren

## 7. Warum nicht Markdown-only?

Markdown-only ist angenehm, aber für easy-author zu schwach, sobald wir Folgendes ernst nehmen:

- unsichtbare Anker
- Satz-/Passagen-Kommentare
- frei konfigurierbare Workflow-Boxen
- gepinnte Clipboard-Ausschnitte
- strukturierte Assets
- kollaboratives Feedback
- stabile Editorzustände

Markdown bleibt hervorragend für Text, aber nicht für alle Arbeitszustände eines modernen Autorenstudios.

## 8. Warum nicht JSON-only?

JSON-only wäre technisch bequem für den Editor, aber strategisch ungünstig für easy-author.

Der Autor soll niemals in einem proprietären Format gefangen sein. Ein Buch muss auch ohne easy-author noch lesbar, exportierbar und archivierbar bleiben.

## 9. Empfehlung als Projektsatz

> easy-author verwendet ein hybrides Speicherformat: Markdown bleibt das primäre, menschenlesbare Autoren- und Exportformat. Tiptap/ProseMirror JSON wird ergänzend als Editor-Snapshot gespeichert, um komplexe Editorzustände, Anker und Workflow-Funktionen stabil wiederherzustellen. Beziehungen, Kommentare, Rechte und Workflow-Boxen werden in der Datenbank verwaltet.
