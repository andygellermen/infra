# easy-author – Proofing- und Veröffentlichungscheckliste

## Ziel

Die Proofing-Checkliste hilft Autoren, ein Werk vor Export, Veröffentlichung oder Review fachlich, formal und technisch zu prüfen.

Sie soll schrittweise wachsen: zuerst einfache Regeln, später exportformat-spezifische Prüfungen.

## Prüfungsebenen

1. Manuskript
2. Struktur
3. Assets
4. Links und Quellen
5. Kommentare und Review
6. Export
7. Veröffentlichung

## Manuskriptprüfung

- Buchtitel vorhanden
- Autor vorhanden
- Untertitel optional geprüft
- Kapitel vorhanden
- keine leeren Kapitel
- keine doppelten Kapiteltitel
- keine offenen TODOs
- Kapitelstatus gesetzt
- finale Kapitel markiert
- sehr lange Kapitel markiert

## Strukturprüfung

- Front Matter vorhanden, falls erforderlich
- Back Matter vorhanden, falls erforderlich
- Impressum vorhanden
- Copyright-Seite vorhanden
- Inhaltsverzeichnis generierbar
- Kapitelreihenfolge geprüft
- Leseprobe definiert, falls benötigt
- Export-Sets geprüft

## Assetprüfung

- Cover vorhanden
- Bilder vorhanden
- Alt-Texte vorhanden
- Quellen erfasst
- Urheber erfasst
- Lizenz erfasst
- Verwendung pro Kapitel nachvollziehbar
- Bildqualität für Export ausreichend

## Link- und Quellenprüfung

- externe Links gültig
- interne Links gültig
- Ankerlinks gültig
- Fußnoten verweisen korrekt
- Endnoten verweisen korrekt
- Quellenverzeichnis vollständig
- Abrufdatum optional erfasst

## Kommentar- und Reviewprüfung

- offene Kommentare vorhanden?
- Kommentare für Export ausgeblendet?
- Reviewer-Freigaben vollständig?
- Testleserfeedback verarbeitet?
- interne Autoren-Notizen entfernt oder als intern markiert?

## PDF-Prüfung

- Seitenformat korrekt
- Ränder korrekt
- Header/Footer korrekt
- Seitenzahlen korrekt
- Kapitelstarts korrekt
- Doppelseitenansicht geprüft
- Cover/erste Seite geprüft
- Fußnotenlayout geprüft

## EPUB-Prüfung

- Inhaltsverzeichnis korrekt
- Kapitelnavigation korrekt
- Bilder skaliert
- Links funktionieren
- Reflow-Verhalten geprüft
- Fuß-/Endnoten funktionieren
- Metadaten vollständig

## DOCX-Prüfung

- Formatvorlagen korrekt
- Überschriftenebenen korrekt
- Fußnoten korrekt
- Kommentare optional enthalten/entfernt
- Verlagsfassung vollständig

## Ghost/Web-Prüfung

- Titel korrekt
- Slug korrekt
- Auszug/Excerpt vorhanden
- Beitragsbild vorhanden
- Mitgliederstatus korrekt
- interne Links angepasst
- CTA/Newsletter-Hinweise geprüft

## UI-Anforderung

Die Checkliste soll in der rechten Sidebar oder in einem eigenen Export-/Proofing-Bereich sichtbar sein.

Status je Checkpunkt:

- ok
- warning
- error
- ignored
- not_applicable

## MVP-Regeln

Für den nächsten UX-Spike genügen einfache Prüfungen:

- leeres Kapitel
- sehr langes Kapitel
- doppelter Kapiteltitel
- TODO im Text
- fehlender Buchtitel
- fehlender Autor
- fehlendes Cover
- fehlendes Impressum
- Asset ohne Lizenzdaten

## Vermeidungsstrategie

Proofing darf den Schreibfluss nicht stören. Hinweise sollen sichtbar, aber nicht aggressiv sein. Kritische Exporthindernisse werden erst beim Export deutlich blockierend.
