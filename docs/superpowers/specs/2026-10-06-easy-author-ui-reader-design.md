# EasyAuthor UI, EasyReader und AddToBook – Design

## 1. Ziel und Leitbild

EasyAuthor wird zu einem ruhigen Autorenarbeitsraum, in dem der Text die
visuelle Hauptrolle spielt. Die Ablenkungsarmut von Typora dient als Leitbild,
ohne die bestehenden Autorenwerkzeuge aufzugeben. Zusatzfunktionen erscheinen
nur im passenden Arbeitsmodus oder durch eine konkrete Handlung.

Die erste Ausbaustufe umfasst den UI-Umbau, Kommentare und Aufgaben sowie eine
Kanban-Ansicht. EasyReader und die Browser-Erweiterung AddToBook werden bereits
in den gemeinsamen Schnittstellen berücksichtigt, aber erst nach dem stabilen
UI-MVP umgesetzt. Obsidian-inspirierte Erweiterungen bleiben eine spätere,
separate Produktphase.

Erfolg bedeutet:

- Ein Buch kann ohne dauerhaft sichtbare Nebenwerkzeuge geschrieben werden.
- Jedes Buch öffnet sich in seinem passenden Arbeitskontext.
- Kommentare, Aufgaben und Quellen bleiben präzise mit Textpassagen verbunden.
- Manuskript, Kommentaransicht und Kanban zeigen stets denselben Vorgangsstatus.
- Importierte Quellen bleiben unverändert und langfristig nachvollziehbar.

## 2. Gestaltungsgrundsätze

1. **Schreiben zuerst:** Der Editor ist Zentrum und Standardinteraktion.
2. **Kontext statt Dauerpräsenz:** Werkzeuge erscheinen bei Auswahl, Fokus oder
   ausdrücklichem Aufruf.
3. **Eine fachliche Einheit, mehrere Ansichten:** Kommentar-Thread, Kanban-Karte
   und Textmarkierung sind keine voneinander getrennten Kopien.
4. **Globale Umgebung, buchbezogene Arbeit:** Farbschema und allgemeines
   Bedienverhalten gelten global; Arbeitsansichten gelten pro Buch.
5. **Unveränderliche Quellen:** Eigene Annotationen liegen als separate Ebene
   über einem importierten Quellenstand.
6. **Farbe unterstützt, ersetzt aber keine Beschriftung:** Status, Buch und
   Auswahlzustand bleiben auch ohne Farbwahrnehmung verständlich.

## 3. Ruhiger Schreibraum

### 3.1 Grundzustand

Der Schreibraum zeigt primär Manuskript, Kapitelbezeichnung und einen sehr
zurückhaltenden Speicherstatus. Globale Navigation, Einstellungen und
Spezialwerkzeuge beanspruchen keine dauerhafte Fläche.

### 3.2 Temporäre Kontrollleiste

Rechts oben befindet sich ein dezenter Drei-Punkte-Auslöser. Die vollständige
Kontrollleiste erscheint:

- bei Annäherung an den oberen Fensterrand,
- nach Aktivierung des Drei-Punkte-Auslösers,
- über ein Tastenkürzel oder per Tastaturfokus.

Nach 60 Sekunden ohne Interaktion wird sie ausgeblendet. Der Timer pausiert,
solange Fokus oder Mauszeiger in der Leiste liegen. Offene Menüs, Dialoge oder
wichtige Hinweise dürfen nicht automatisch verschwinden.

Die Leiste enthält:

- links Buch, Kapitel und Rückkehr zur Übersicht,
- mittig den aktuellen Arbeitsmodus,
- rechts Suche beziehungsweise Befehlspalette, Erscheinungsbild,
  Einstellungen und weitere globale Aktionen.

Textbezogene Aktionen wie Kommentieren, Verknüpfen, Hervorheben oder Übernahme
ins Clipboard gehören nicht in diese Leiste. Sie erscheinen an der Auswahl
oder über die Befehlspalette.

### 3.3 Erscheinungsbild

Hell, Dunkel und Systemeinstellung sind globale Präferenzen. Ein Buchwechsel
darf keinen unerwarteten Wechsel des Farbschemas verursachen. Schrift,
Textbreite sowie Fokus- und Schreibmaschinenmodus sind über das
Erscheinungsbild-Menü erreichbar.

## 4. Buchbezogene Arbeitsansichten

Beim Anlegen eines Buches erscheint vor weichgezeichnetem und abgedunkeltem
Hintergrund ein Auswahl-Widget mit drei Karten. Die ungefähr 50 Pixel großen
Symbole sind Teil größerer, verständlich beschrifteter Auswahlflächen.

1. **Clean & Free:** Manuskript, minimale Navigation und Speicherstatus.
2. **Intense:** Recherche, Struktur, Wissensbank und Story-Verknüpfungen.
3. **Review:** Kommentare, Aufgaben, Vorschläge und Prüfhinweise.

Die Auswahl wird als Standard des Buches gespeichert. Ein temporärer Wechsel
für die laufende Sitzung verändert diesen Standard nicht ungefragt. Alle Modi
arbeiten auf demselben Manuskript und verändern ausschließlich Sichtbarkeit und
Erreichbarkeit der Werkzeuge.

Der bestätigte interaktive Entwurf liegt unter
`apps/easy-author/docs/mockups/work-view-selector.html`.

## 5. Textmarkierungen und Kommentare

### 5.1 Gemeinsamer Einstieg

Eine Textauswahl öffnet ein kleines Kontextmenü. Es bietet mindestens:

- Kommentar anlegen,
- Markdown-Hervorhebung,
- Passage verknüpfen,
- ins Knowledge-Clipboard übernehmen,
- weitere Aktionen über einen nachgeordneten Einstieg.

### 5.2 Bedeutungen

- Eine gelbe Markierung ist inhaltliche Markdown-Hervorhebung und Teil des
  Manuskripts.
- Eine hellgrüne Markierung kennzeichnet einen offenen Kommentar-Thread und ist
  redaktionelle Metainformation.
- Ein erledigter Thread bleibt hell- beziehungsweise dunkelgrau erkennbar. Die
  Anzeige erledigter Markierungen ist einstellbar.
- Nur bewusstes Löschen entfernt Markierung und Thread.

Die Farbtöne besitzen angepasste Varianten für Hell und Dunkel. Symbole,
Beschriftungen und zugängliche Zustände verhindern, dass Bedeutung allein von
Farbe abhängt.

### 5.3 Schwebende Kommentar-Threads

Kommentarblasen erscheinen freihängend rechts neben der zugehörigen Passage.
Nahe Blasen werden gestaffelt oder gruppiert. Ein Klick öffnet einen
threadfähigen Dialog im Messenger-Stil. Antworten gehören zum selben Thread und
erzeugen keine separaten Aufgaben.

Ein Thread kennt mindestens `offen`, `geplant`, `in_arbeit`, `pruefung`,
`erledigt` und `geloescht`. Statusänderungen wirken gleichzeitig auf
Textmarkierung, Kommentarblase und Kanban-Karte.

Textanker kombinieren stabile Blockkennung, Textzitat, Kontext davor und
danach, Prüfsumme und Dokumentposition. Eine beschädigte Zuordnung wird nicht
stillschweigend verworfen, sondern zur Prüfung markiert.

## 6. Kanban-Arbeitsansicht

### 6.1 Vorgänge und Phasen

Das Kanban ist eine eigene Ansicht für Aufgaben, Kommentar-Threads und
Erinnerungen. Ein Thread entspricht genau einer Karte. Die fünf Spalten sind:

1. **Backlog:** neu und ungeprüft; roter Akzent.
2. **To-do:** geprüft und eingeplant; gelber beziehungsweise amberfarbener
   Akzent.
3. **In Arbeit:** aktive Bearbeitung; oranger Akzent.
4. **Prüfung:** Lösung wartet auf Bestätigung; grüner Akzent.
5. **Fertig:** abgeschlossen; grauer Akzent.

Spalten und Karten bleiben überwiegend neutral. Farben erscheinen als
Statuskante, Kopfakzent oder vergleichbar kleine Markierung. Dringlichkeit ist
eine separate Eigenschaft und wird nicht aus der Phase abgeleitet.

Drag-and-drop ändert den Status. Eine gleichwertige Tastaturbedienung muss
vorhanden sein. Ein Klick auf die Quellenangabe einer Karte führt zur
verknüpften Manuskript- oder Quellenpassage.

### 6.2 Buch- und globale Ansicht

Innerhalb eines Buches zeigt das Kanban ausschließlich dessen Vorgänge. Die
Bücherübersicht bietet versuchsweise ein globales Kanban über mehrere Bücher.
Beide Ansichten verwenden dasselbe Daten- und Statusmodell und unterscheiden
sich nur durch Filter.

Links zeigt eine Bücherliste je Buch drei verdichtete Zähler:

- offen: Backlog und To-do,
- in Arbeit: In Arbeit und Prüfung,
- fertig: abgeschlossene Vorgänge.

Vor jedem Titel steht ein runder Mehrfachauswahl-Schalter. Mehrere Bücher
können gemeinsam gewählt werden. Bei Mehrfachauswahl erhält jedes gewählte Buch
temporär eine möglichst gut unterscheidbare Farbe aus einer kuratierten,
kontrastreichen Palette. Die Farbe wird nicht im Buch gespeichert.

Während die Auswahl besteht, behalten bereits gewählte Bücher ihre Farbe. Ein
neues Buch erhält die am besten unterscheidbare verbleibende Farbe. Kreis und
Buchfähnchen der Karten verwenden dieselbe Farbe. Die Fähnchen ragen leicht
über die Karte hinaus und verwenden ungefähr 60 bis 65 Prozent der normalen
Kartenschrift. Bei genau einem Buch werden sie ausgeblendet.

### 6.3 Mengenbegrenzung und Sortierung

Pro Status werden zunächst ungefähr zwölf und höchstens 10 bis 20 Karten
angezeigt. Weitere Karten werden bewusst nachgeladen. Die Gesamtzahl bleibt im
Spaltenkopf sichtbar. Die Standardsortierung berücksichtigt:

1. hohe Priorität,
2. Überfälligkeit,
3. letzte Aktualisierung,
4. ältere reguläre Vorgänge.

Erledigte Vorgänge treten zurück und können eingeklappt oder gefiltert werden.

## 7. EasyReader und Quellenbibliothek

### 7.1 Quellenobjekt

Ein Webimport erzeugt ein eigenständiges Quellenobjekt und fügt den Inhalt
nicht ungeordnet in ein Manuskript ein. Der vollständige relevante
Seiteninhalt wird in Markdown überführt. Dazu gehören Überschriften, Absätze,
Listen, Zitate, Tabellen, Links und Artikelbilder. Navigation, Werbung,
Cookie-Dialoge, Footer, Skripte, Formulare und Layout-CSS werden entfernt.

Jede Revision enthält mindestens:

- bereinigtes `content.md`,
- lokal gespeicherte Medien,
- Original- und kanonische URL,
- Titel, Autor und Veröffentlichungsdatum, soweit ermittelbar,
- Domain, Sprache und exakten Importzeitpunkt,
- Inhaltsprüfsumme,
- Import- und Extraktionsversion,
- ursprüngliche Bildadressen, Formate, Alt-Texte und erkennbare Rechteangaben.

Eine Revision ist nach dem Import unveränderlich. Kommentare, Markierungen,
Aufgaben und Verknüpfungen werden getrennt gespeichert.

### 7.2 Bilder

Alle Bilder des relevanten Inhalts werden im Shoot-and-forget-Prozess
automatisch lokal gesichert. Übliche Rasterbilder werden nach WebP konvertiert,
wenn dies ohne relevanten Funktionsverlust sinnvoll ist. SVG bleibt nach
Sicherheitsbereinigung als Vektor erhalten. Animationen dürfen nicht
unbemerkt verloren gehen. Ursprungs-URL, Ausgangsformat und Prüfsumme bleiben
dokumentiert.

Nicht mehr eingebundene Bilder können manuell bereinigt werden. Eine Löschung
ist gesperrt, solange eine Revision oder Annotation darauf verweist.

### 7.3 Reader-Erlebnis und Annotationen

EasyReader zeigt Quellen ohne Einzelseiten als echten, ablenkungsarmen
Fließtext. Markierungen bleiben über Sitzungen hinweg erhalten und zeigen ihre
Ziele, etwa Buch, Kapitel, Absatz, Satz, Knowledge-Objekt oder Kanban-Aufgabe.
Navigation funktioniert in beide Richtungen zwischen Quelle und Ziel.

Eine markierte Quellenpassage kann ins Knowledge-Clipboard übernommen werden.
Der Eintrag bewahrt exaktes Zitat, Quellenrevision, Passage, Herkunftsdaten und
Verknüpfungsziele. Eine spätere Übernahme ins Manuskript verliert den
Quellennachweis nicht.

### 7.4 Diff und Revisionen

Die erste Version prüft Quellen ausschließlich nach manuellem Aufruf. EasyReader
ruft die URL erneut ab, verarbeitet sie mit derselben Pipeline und vergleicht
normalisiertes Markdown blockweise sowie innerhalb geänderter Absätze.

Die Ansicht unterscheidet hinzugefügte, entfernte und geänderte Passagen. Der
Vergleich kann ohne Speicherung geschlossen oder als neue unveränderliche
Revision gesichert werden. Der bisherige Stand wird nie überschrieben.

Annotationen bleiben an ihrer Ursprungsrevision erhalten. Für eine neue
Revision versucht EasyReader eine Wiederzuordnung:

- sicher gefunden: auch im neuen Stand anzeigen,
- wahrscheinlich gefunden: Zuordnung vorschlagen,
- entfernt oder stark verändert: zur Prüfung markieren.

## 8. AddToBook-Browser-Erweiterung

AddToBook übergibt eine geöffnete Webseite mit möglichst wenig Reibung an
EasyAuthor:

1. Erweiterung öffnen.
2. Zielbuch oder allgemeinen Eingang wählen.
3. erkannten Hauptinhalt kurz prüfen.
4. Quelle senden.
5. optional in EasyReader öffnen oder eine Aufgabe anlegen.

Die Erweiterung benötigt möglichst geringe Browserrechte. Clientseitige
Extraktion ist kein Vertrauensanker: EasyAuthor behandelt jeden Import als
nicht vertrauenswürdig, bereinigt ihn erneut und übernimmt keine aktiven
Inhalte. Fehlerhafte, unvollständige oder nicht zugängliche Seiten erhalten
einen nachvollziehbaren Importstatus statt eines stillen Teilerfolgs.

## 9. Paketierung

### Paket 1 – Ruhiger Schreibraum

Editorzentrierung, temporäre Kontrollleiste, Drei-Punkte-Auslöser, globale
Farbdarstellung, Fokus- und Tastaturverhalten.

### Paket 2 – Buchbezogene Arbeitsansichten

Auswahl-Widget, Clean & Free, Intense, Review, Speicherung und temporärer
Wechsel.

### Paket 3 – Markierungen und Kommentar-Threads

Auswahlmenü, gelbe Hervorhebung, grüne und graue Kommentarmarkierung,
schwebende Blasen, Threads und robuste Anker.

### Paket 4 – Kanban-Arbeitsansicht

Fünf Phasen, Drag-and-drop, gemeinsame Statuslogik, Buchfilter,
Mehrfachauswahl, temporäre Kontrastfarben und begrenzte globale Ansicht.

### Paket 5 – Stabilisierung des UI-MVP

Responsive Verhalten, Barrierefreiheit, Migration, Zustandswiederherstellung,
realistische Lastfälle, automatisierte Tests und Dokumentation.

### Paket 6 – EasyReader und Quellenbibliothek

Unveränderliche Markdown-Revisionen, lokale Medien, Fließtext-Reader,
Annotationen, Knowledge-Clipboard und manueller Diff.

### Paket 7 – AddToBook

Browser-Erweiterung, Hauptinhaltsextraktion, Zielauswahl, sicherer Importvertrag
und Übergabe an EasyReader.

## 10. Architekturgrenzen und Datenfluss

Die Pakete 1 bis 5 implementieren den UI-MVP. Sie müssen bereits gemeinsame
fachliche Schnittstellen für `DocumentAnchor`, `CommentThread`, `WorkItem`,
`KnowledgeClipboardItem` und ein allgemeines Verknüpfungsziel vorsehen. Paket 6
ergänzt `Source` und `SourceRevision`; Paket 7 verwendet ausschließlich den
definierten Importvertrag.

```text
Manuskript oder Quellenrevision
        ↓ Passage markieren
DocumentAnchor
        ↓
CommentThread / KnowledgeClipboardItem / WorkItem
        ↓
Manuskriptansicht ↔ EasyReader ↔ Kanban
```

Jede Einheit hat eine eindeutige Verantwortung:

- `DocumentAnchor` findet eine Passage wieder.
- `CommentThread` verwaltet den Dialog.
- `WorkItem` verwaltet Phase, Priorität und Planung.
- `KnowledgeClipboardItem` bewahrt einen wiederverwendbaren Ausschnitt samt
  Herkunft.
- `SourceRevision` bewahrt einen unveränderlichen Quellenstand.

## 11. Fehler- und Schutzregeln

- Automatisches Ausblenden darf keine laufende Interaktion abbrechen.
- Ein beschädigter Anker bleibt als prüfbedürftiger Verweis erhalten.
- Statusänderungen werden atomar zwischen Thread und Kanban synchronisiert.
- Importiertes HTML, SVG und Medien gelten als nicht vertrauenswürdig.
- Fehlgeschlagene Bildimporte werden sichtbar dokumentiert.
- Eine neue Quellenrevision überschreibt niemals eine ältere.
- Manuelle Medienlöschung prüft Referenzen und historische Revisionen.
- Externe Inhalte werden nicht automatisch veröffentlicht oder in einen
  Manuskriptexport aufgenommen.
- Quellen-, Urheber- und Lizenzangaben werden soweit verfügbar bewahrt und bei
  fehlenden Angaben kenntlich gemacht.

## 12. Prüfstrategie

Jedes Paket erhält Komponenten-, Integrations- und End-to-End-Prüfungen.
Besonders wichtig sind:

- Ein- und Ausblendlogik der Kontrollleiste mit Maus und Tastatur,
- Persistenz und temporärer Wechsel der Arbeitsansichten,
- Markierungssemantik in Hell und Dunkel,
- Wiederfinden von Ankern nach Textänderungen,
- Thread- und Kanban-Statussynchronisation,
- Drag-and-drop sowie Tastaturalternative,
- Mehrfachauswahl und stabile temporäre Buchfarben,
- Begrenzung und Sortierung großer Boards,
- sichere und reproduzierbare Webextraktion,
- lokale Bildreferenzen und Formatbehandlung,
- unveränderliche Quellenrevisionen,
- Diff sowie Wiederzuordnung bestehender Annotationen.

Nach Paket 1 und 2 erfolgt eine gemeinsame Erfahrungsprüfung der tatsächlichen
Ablenkungsarmut. Paket 6 und 7 beginnen erst, wenn Paket 5 stabil abgeschlossen
ist.

## 13. Nicht Bestandteil dieser Spezifikation

- automatische zeitgesteuerte Quellenprüfungen,
- ungefragte Aktualisierung gespeicherter Quellen,
- vollständige Website-Archivierung einschließlich Navigation und Werbung,
- Obsidian-artiges Plugin-System,
- automatische KI-Zusammenfassung oder KI-Einarbeitung von Quellen,
- Veröffentlichung oder Produktivbereitstellung.
