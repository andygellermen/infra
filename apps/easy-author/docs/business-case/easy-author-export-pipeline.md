# easy-author – Export Pipeline

## 1. Ziel

Die Export-Pipeline erzeugt aus einem strukturierten Buchprojekt hochwertige Ausgaben in Markdown, PDF, EPUB und DOCX.

## 2. Grundprozess

```text
Projekt laden
→ Kapitel nach Reihenfolge einsammeln
→ Includes und interne Links auflösen
→ Metadaten anwenden
→ Assets prüfen
→ Theme auswählen
→ Exportprofil anwenden
→ Ausgabe erzeugen
→ Ergebnis speichern
→ Exportprotokoll anzeigen
```

## 3. Exportformate

### Markdown

- gesamtes Manuskript als eine Markdown-Datei
- optional Kapitel einzeln
- optional ZIP-Projektarchiv

### PDF

- PDF für Druck/Lesen
- Seitenformat einstellbar
- Ränder einstellbar
- Theme/CSS auswählbar
- Cover optional
- Inhaltsverzeichnis optional

### EPUB

- EPUB mit Metadaten
- Cover
- Inhaltsverzeichnis
- CSS-Theme
- Bildoptimierung

### DOCX

- Word-Dokument
- Formatvorlagen über Referenz-DOCX
- geeignet für Lektorat und Verlagseinreichung

## 4. Exportprofile

Ein Projekt kann mehrere Exportprofile besitzen:

- Buch PDF A5
- Arbeitsfassung A4
- EPUB Standard
- Verlag DOCX
- Leseprobe PDF
- Ghost-Serie Markdown

## 5. PDF-Einstellungen

- Seitenformat: A4, A5, Letter, Custom
- Ränder: oben, unten, links, rechts
- Schriftgröße
- Zeilenhöhe
- Kopf-/Fußzeile optional
- Seitenzahlen
- Inhaltsverzeichnis
- Kapitelstart auf neuer Seite

## 6. Themes

Themes sind wiederverwendbare Layoutpakete.

```text
/themes
  /pdf
    calm-book.css
    workbook.css
  /epub
    clean-epub.css
  /docx
    reference.docx
```

## 7. Export-Qualitätscheck

Vor dem Export:

- fehlender Titel
- fehlender Autor
- fehlendes Cover
- leere Kapitel
- offene TODOs
- nicht aufgelöste interne Links
- fehlende Assets
- fehlende Alt-Texte
- fehlende Lizenzdaten
- Kommentare noch offen

## 8. Technische Bausteine

Empfohlen:

- Pandoc für Markdown, EPUB, DOCX und allgemeine Konvertierung
- HTML/CSS-basierter PDF-Renderer für typografisch steuerbare PDFs
- optional LaTeX-Pipeline für sehr professionelle Drucklayouts
- Worker-Container für isolierte Exporte

## 9. Export-Historie

Jeder Export erzeugt einen Datensatz:

- Format
- Profil
- Zeitpunkt
- Status
- Dateipfad
- Log
- Prüfwarnungen

## 10. Vermeidungsstrategie

PDF-Layout darf nicht zu früh perfektioniert werden. Der erste Schwerpunkt liegt auf zuverlässigem Export. Feintypografie, Normseiten und Druck-PDFs folgen als Ausbauphase.

## Ergänzung: Exportvorbereitung vor Exportimplementierung

Vor der eigentlichen Exportpipeline wird eine Proofing-/Export-Checkliste eingeführt. Diese prüft einfache Voraussetzungen wie Buchtitel, Autor, leere Kapitel, doppelte Kapiteltitel, TODOs, fehlendes Cover, fehlendes Impressum und Asset-Lizenzdaten. Der echte Export nach PDF, EPUB und DOCX bleibt eine spätere Umsetzungsstufe.

