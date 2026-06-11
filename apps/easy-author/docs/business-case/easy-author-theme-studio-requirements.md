# easy-author – Theme Studio Anforderungen

## Ziel

Das Theme Studio ermöglicht Autoren, Export- und Darstellungsvorlagen für unterschiedliche Veröffentlichungsziele zu wählen und später zu gestalten.

Im MVP wird nur das Datenmodell vorbereitet. Die vollständige visuelle Theme-Bearbeitung folgt später.

## Theme-Arten

- PDF Theme
- EPUB Theme
- DOCX Reference Theme
- Ghost/Web Theme
- Large Print Theme
- Arbeitsheft Theme
- Leseprobe Theme

## Theme-Bestandteile

### Allgemein

- Name
- Beschreibung
- Zielmedium
- Standard-Sprache
- Schriftfamilien
- Farben
- Abstandssystem

### PDF-spezifisch

- Seitenformat
- Ränder
- Beschnitt/Trim Size
- Header
- Footer
- Seitenzahlen
- Kapitelstart rechts/links/nächste Seite
- Doppelseitenlogik

### EPUB-spezifisch

- Inhaltsverzeichnis
- Kapitelnavigation
- Bildgrößen
- CSS
- responsives Verhalten

### DOCX-spezifisch

- Formatvorlagen
- Überschriftenebenen
- Absatzformate
- Fußnotenformat
- Referenzdatei später

### Ghost/Web-spezifisch

- Beitragslayout
- Seriennavigation
- Mitgliederhinweis
- Newsletter-Version
- Call-to-Action-Blöcke

## Buchbausteine

Themes müssen besondere Blöcke abbilden können:

- Zitat
- Hinweis
- Übung
- Merksatz
- Weitblick
- Wort-Notiz
- spiritueller Impuls
- Checkliste
- Quellenbox
- Autorenkommentar nur intern

## Kapiteloptionen

Pro Kapitel sollen Theme-Optionen gesetzt werden können:

- im Inhaltsverzeichnis anzeigen
- Kapitelnummer anzeigen
- Kapitelüberschrift anzeigen
- Kapitelbild anzeigen
- Header/Footer anzeigen
- Seitennummer anzeigen
- Sonderlayout verwenden
- Kapitel als Leseprobe markieren
- Kapitel vom Export ausschließen

## Theme Library

Die Theme Library verwaltet:

- System-Themes
- eigene Themes
- Projekt-Themes
- geteilte Themes
- spätere Marketplace-/SaaS-Themes

## MVP-Umfang

Im nächsten Schritt genügt:

- Theme-Entity vorbereiten
- Theme-Auswahl pro Buch speichern
- PDF/EPUB-Konfiguration als JSON-Feld vorbereiten
- UI-Platzhalter für „Theme“ anzeigen

## Späterer Ausbau

- visueller Theme Builder
- Live-Vorschau
- Exportvorschau
- Theme-Duplikation
- Theme-Import/-Export
- Font-Auswahl
- Lizenzhinweise
- Theme-Prüfung pro Exportformat

## Vermeidungsstrategien

- Keine komplexe Theme-Engine im zweiten Spike.
- Zuerst Datenmodell und UI-Ort schaffen.
- Keine eigenen Fonts im MVP.
- Exportvorschau erst bauen, wenn Exportpipeline stabil ist.
