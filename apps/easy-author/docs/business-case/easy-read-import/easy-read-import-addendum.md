# easy-author – Easy Read Import Addendum

## Zweck dieses Addendums

Dieses Addendum ergänzt das bisherige Korrektur- und Optimierungskonzept für die easy-author App um eine wichtige Erweiterung des Moduls **Easy Read Import**.

Easy Read Import soll nicht nur Webseiten, Artikel oder längere Texte bereinigt importieren. Es soll importierte Inhalte zugleich als **Recherchequelle**, **Lesedokument**, **verknüpfbare Arbeitsgrundlage** und optional als **Kanban-Aufgabe** nutzbar machen.

Der Kern ist:

> Lesen, importieren, einordnen, verknüpfen, bearbeiten, als Aufgabe sichtbar machen und später bewusst in das eigene Buchprojekt einarbeiten.

---

## Grundprinzip

Importierte Texte dürfen nicht ungeordnet im Manuskript landen.

Stattdessen sollen sie als eigenständige, nachvollziehbare Quellenobjekte verwaltet werden.

Ein importierter Text kann wahlweise:

1. als eigenes Buch oder eigenes Lesedokument angelegt werden,
2. als inhaltlich verknüpfte Quelle einem bestehenden Buch zugeordnet werden,
3. einem bestimmten Kapitel zugeordnet werden,
4. mit einer Handlung, Person, Gruppe, Beziehung, einem Ereignis oder Motiv verknüpft werden,
5. als Kanban-Ticket für spätere Bearbeitung sichtbar gemacht werden,
6. mit Aufgaben, Fälligkeiten, Meta-Beschreibung und Bearbeitungsstatus versehen werden.

---

## Warum diese Ergänzung wichtig ist

easy-author soll nicht nur Schreibfläche sein, sondern ein Autoren-Arbeitsraum.

Viele Autoren lesen, recherchieren, markieren, sammeln und ordnen lange bevor ein Text final im Manuskript landet.

Easy Read Import soll diesen natürlichen Arbeitsfluss unterstützen:

```text
Webseite / Artikel / Quelle
→ bereinigter Import
→ Lesedokument
→ Markierung / Kommentar / Todo
→ Verknüpfung mit Buchobjekten
→ Kanban-Karte
→ spätere Einordnung ins Manuskript
```

## Vertragsgrenze nach dem UI-MVP

Die Pakete 1 bis 5 stellen bereits die wiederverwendbaren Manuskriptverträge
bereit. EasyReader und AddToBook dürfen diese Verträge erweitern, aber nicht
durch parallele Quellvarianten duplizieren.

### Bereits implementierte Verträge

- `DocumentAnchor` lokalisiert eine Passage über Kapitel, Blockkennung,
  Textzitat, Start-/Endposition, Kontext davor und danach sowie Prüfsumme.
- `AnchorContext` ordnet einem Anker unabhängig Kommentar, Aufgabe,
  Verknüpfung, Clipboard-Einfügung oder Clipboard-Quelle zu.
- `CommentThread` bewahrt Status und geordnete Dialognachrichten.
- `WorkItem` bewahrt Projekt, Buch, Kapitel, Anker, optionalen Thread, Art,
  Titel, Phase, Priorität und Fälligkeit. Ein Thread besitzt höchstens eine
  Karte; Phasenänderungen sind atomar synchronisiert.
- Das heutige `ClipboardItem` bewahrt Buch-/Kapitelbezug, Inhalt,
  Inhaltstyp, Quellanker und Slot. Es ist die vorhandene Manuskriptfunktion,
  aber noch nicht der vollständige Quellenvertrag `KnowledgeClipboardItem`.

### Paket 6 ergänzt

`Source` beschreibt die dauerhaft identifizierbare Quelle. Mindestens
vorzusehen sind ID, Original- und kanonische URL, Domain, aktueller
Importstatus, zugeordnete Bücher und Zeitstempel.

`SourceRevision` ist unveränderlich und benötigt mindestens:

- Quellen-ID und fortlaufende Revision,
- bereinigtes Markdown beziehungsweise Pfad zu `content.md`,
- Titel, Autor, Veröffentlichungsdatum und Sprache, soweit ermittelbar,
- exakten Importzeitpunkt, Inhaltsprüfsumme und Extraktionsversion,
- lokale Medienreferenzen mit Original-URL, Format, WebP-/SVG-Ziel,
  Alt-Text, Prüfsumme und erkennbarem Rechtehinweis,
- Status für vollständige, teilweise oder fehlgeschlagene Extraktion.

`KnowledgeClipboardItem` erweitert den heutigen Clipboard-Vertrag um
`source_revision_id`, exaktes Zitat, einen `DocumentAnchor` innerhalb der
Quellenrevision, Herkunftsmetadaten und mehrere Verknüpfungsziele. Die spätere
Übernahme ins Manuskript darf diese Provenienz nicht verlieren.

Für Quellenanker wird `DocumentAnchor` medienneutral erweitert: Statt nur
`chapter_id` wird ein typisiertes Dokumentziel aus Manuskriptkapitel oder
Quellenrevision benötigt. Die vorhandene Wiederfindelogik aus Zitat, Kontext,
Position und Prüfsumme bleibt maßgeblich.

### Paket 7 verwendet ausschließlich den Importvertrag

AddToBook übergibt keine bereits vertrauenswürdigen Datenbankobjekte, sondern
einen Importauftrag mit URL, kanonischer URL, Browser-Metadaten, extrahiertem
Kandidateninhalt, Bildkandidaten, Extraktionsversion und optionalem Zielbuch.
EasyAuthor bereinigt und validiert alles erneut.

Die Erweiterung darf zur Erkennung bekannter Seiten einen URL-Abgleich und den
Hash des normalisierten relevanten Inhalts anfragen. Nur eine abweichende
Prüfsumme löst den proaktiven Revisionsdialog aus; ohne Bestätigung wird weder
eine Revision angelegt noch eine bestehende verändert.

### Ausdrücklich noch nicht implementiert

- `Source`, `SourceRevision` und revisionsgebundene Quellenanker,
- Hauptinhaltsextraktion und Markdown-Normalisierung von Webseiten,
- Download, Bereinigung und WebP-Konvertierung von Bildern,
- manueller Diff und Wiederzuordnung von Annotationen,
- Browser-Erweiterung, Berechtigungsmodell und bekannte-Seiten-Pop-up,
- automatische oder zeitgesteuerte Quellenprüfung.

Damit bleibt der UI-MVP unabhängig von der späteren Importpipeline, während
EasyReader Manuskriptansicht und Kanban ohne Modellbruch wiederverwenden kann.
