# easy-author – Import- und Export-Anforderungen

## Ziel

Import und Export sollen Autoren Sicherheit geben. easy-author soll Manuskripte aufnehmen, in eine saubere Buchstruktur bringen und später kontrolliert in Markdown, PDF, EPUB, DOCX, HTML oder Ghost-Inhalte ausgeben können.

## Import-Ziele

### MVP-nah

- Markdown-Datei importieren
- mehrere Markdown-Dateien importieren
- Ordnerstruktur als Buchprojekt importieren

### Später

- DOCX importieren
- HTML importieren
- ZIP-Projektarchiv importieren
- Ghost-Export wieder importieren

### Bewusst nicht im MVP

- PDF als bearbeitbares Manuskript importieren

PDFs können später als Referenz-Assets gespeichert werden, aber nicht als primäres bearbeitbares Manuskript.

## Import-Assistent

Der Import-Assistent soll nicht nur importieren, sondern diagnostizieren.

### Diagnosepunkte

- Kapitel erkannt
- Kapitel ohne Titel
- doppelte Titel
- Bilder gefunden
- Bilder ohne Lizenzdaten
- Fußnoten erkannt
- Links erkannt
- ungültige Links
- leere Abschnitte
- mögliche Front-Matter-Seiten
- mögliche Back-Matter-Seiten

### Import-Protokoll

Nach dem Import soll ein Protokoll entstehen:

```text
14 Kapitel erkannt
3 Kapitel ohne eindeutigen Titel
8 Bilder übernommen
5 Bilder ohne Alt-Text
2 mögliche Back-Matter-Seiten erkannt
```

## Werkstruktur beim Import

Die Importlogik soll versuchen, Inhalte automatisch zuzuordnen:

- Front Matter
- Body/Kapitel
- Back Matter
- Assets
- Quellen

Der Autor muss jede automatische Zuordnung korrigieren können.

## Export-Ziele

### MVP-vorbereitend

- Markdown-Gesamtexport
- Projekt-ZIP
- Export-Checkliste

### Nächste Stufe

- PDF
- EPUB
- DOCX
- HTML

### Spätere Stufe

- Ghost-Post
- Ghost-Serie
- Verlags-Einreichungspaket
- Leseprobe
- Paywall-/Spenden-Version

## Export-Checkliste

Vor jedem Export prüft easy-author:

- fehlender Buchtitel
- fehlender Autor
- fehlendes Cover
- fehlendes Impressum
- leere Kapitel
- doppelte Kapiteltitel
- offene TODOs
- nicht finale Kapitel
- offene Kommentare
- ungültige interne Links
- fehlende Bildlizenzen
- fehlende Alt-Texte
- fehlende Quellenangaben

## Export-Sets

Ein Export-Set definiert, welche Teile des Werkes ausgegeben werden.

Beispiele:

- komplettes Buch
- nur Leseprobe
- nur Band 1
- nur Arbeitsblätter
- nur Ghost-Serie
- Verlagsfassung

## Formatlogik

### Markdown

- führendes Autoren- und Archivformat
- gut für Git und Portabilität
- sauberer Text ohne proprietären Ballast

### PDF

- fixiertes Layout
- wichtig für Print, Korrektur und professionelle Weitergabe
- benötigt Theme, Seitenformat, Ränder, Header/Footer

### EPUB

- reflowable Layout
- wichtig für E-Reader
- benötigt saubere Navigation, Inhaltsverzeichnis und Bildoptimierung

### DOCX

- wichtig für Verlage, Lektorat und externe Zusammenarbeit
- benötigt Formatvorlage/Reference-Docx

### Ghost/Web

- wichtig für Serien, Newsletter, Mitgliederbereiche und spätere Monetarisierung
- benötigt Web-Theme und Link-/Asset-Handling

## Vermeidungsstrategien

- Export nicht als Blackbox bauen.
- Immer ein Export-Protokoll erzeugen.
- Keine fehleranfällige PDF-Import-Automatik im MVP.
- Bilder nicht ungeprüft aus DOCX übernehmen.
- Export-Checks früh anzeigen, nicht erst beim finalen Klick.
