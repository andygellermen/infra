# easy-author – Author Strategy Audit

## Ziel dieses Dokuments

Dieses Dokument ist der bewusste `Kassensturz` fuer die aktuelle `easy-author`-MVP-Oberflaeche.

Es beantwortet vier Fragen:

1. Welche Autoren-Hilfsmittel existieren bereits?
2. Wann werden sie sichtbar oder aktiv?
3. Wofuer sind sie im Schreibprozess gedacht?
4. Wo entstehen aktuell Orientierungslast, Ueberlappungen oder Reibung?

Der Fokus liegt nicht auf Technik allein, sondern auf dem echten Autorinnen-/Autoren-Flow.

## Strategischer Leitgedanke

`easy-author` hat inzwischen zwei sehr starke Ebenen:

- einen ruhigen zentralen Schreibraum
- ein kontextsensitives Assistenz-System rund um Kapitel, Wissen, Workflow und Clipboard

Die aktuelle Schwierigkeit ist nicht mehr fehlende Funktion, sondern:

- zu viele moegliche Helfer gleichzeitig
- noch nicht streng genug priorisierte Sichtbarkeit
- zu wenig mentale Fuehrung, wann welches Werkzeug wirklich benutzt werden soll

Deshalb sollte die MVP-Strategie kuenftig nach diesem Prinzip organisiert werden:

### 1. Schreiben zuerst

Alles, was den Satzfluss stoert, bleibt standardmaessig leise oder verborgen.

### 2. Kontext nur bei Anlass

Helfer erscheinen erst bei:

- Textauswahl
- Kapitel-/Workflow-Kontext
- echten Triggern im Text
- explizitem Aufruf

### 3. Werkzeuge nach Autoren-Aufgabe statt nach Datenmodell

Die Oberflaeche sollte gedanklich nicht `Objekte verwalten`, sondern diese Aufgaben unterstuetzen:

- schreiben
- sammeln
- verknuepfen
- strukturieren
- pruefen
- wiederverwenden

## Oberflaechen-Inventar

| Bereich | Hilfsmittel | Zweck | Sichtbarkeit/Aktivierung | Typisches Beispiel | Aktuelle Reibung | Empfehlung |
| --- | --- | --- | --- | --- | --- | --- |
| Linke Sidebar | `Arbeitsraum` | Projektwahl | Immer sichtbar | Zwischen Projekten wechseln | derzeit sachlich, aber noch ohne klare Prioritaet | fuer MVP ok |
| Linke Sidebar | `Buch` + `Buchdetails` | Buch-Kontext pflegen | Immer sichtbar bei Projektwahl | Titel, Beschreibung, Autor, Sichtbarkeit pflegen | sinnvoll, aber visuell noch gleichrangig zu aktiveren Bereichen | mittelfristig leicht einklappbar |
| Linke Sidebar | `Kapitel` | Kapitelwahl und Sortierung | Immer sichtbar | Kapitel oeffnen, per Drag-and-drop sortieren | zentral und nachvollziehbar | so belassen |
| Linke Sidebar | `Workflow-Boxen` | Arbeitsboxen definieren und als Ziel setzen | Immer sichtbar, plus Trigger-Vorschlaege | Personen-, Recherche- oder Erinnerungsbox aktivieren | hohe Dichte, viele Bearbeitungsfelder gleichzeitig | in Stufen gliedern: kompakt, aktiv, editierbar |
| Linke Sidebar | `Wissensbank` | Begriffe pflegen und `[[...]]`-Links einsetzen | Immer sichtbar | `[[Person:Mara]]` einfuegen | im Vollbetrieb schnell zu praesent | staerker filtern, Standard kompakter |
| Mitte | `Rich`/`Markdown` | Schreiben im bevorzugten Modus | Expliziter Moduswechsel | in Rich schreiben, Markdown exportnah pruefen | Doppellogik braucht klare Leitplanke | Markdown eher als Experten-/Exportmodus kommunizieren |
| Mitte | `Werkzeuge` | Tabellen, Zitate, Fussnoten einsetzen | Nur nach Klick auf `Werkzeuge` | Pipe-Tabelle einleiten, Fussnote setzen | bereits gut gedaempft | beibehalten |
| Mitte | `Textauswahl-Aktionen` | Anker, Clipboard, Wiki-Link direkt aus Text ableiten | Nur bei Auswahl | Passage markieren und verankern | konzeptuell stark, muss mental erklaert werden | als Primaerinteraktion beibehalten |
| Mitte | `Selection Popup` | Schnellentscheidungen am markierten Text | Nach Haltezeit bei Auswahl | `Zu Recherche`, `Clipboard`, `Neue Box` | kann ohne Einfuehrung ungewohnt wirken | als intelligentes Kernfeature positionieren |
| Mitte | `Editor-Hilfe` | Bedienlogik erklaeren | Expliziter Aufruf | Markdown/Slots/Anker nachlesen | hilfreich, aber noch eher Referenz als Fuehrung | spaeter onboarding-faehig machen |
| Mitte | `Editor-Einstellungen` | Schreibgefuehl anpassen | Expliziter Aufruf | Vollbildfarbe, Breite, Schrift einstellen | inhaltlich gut, UX-seitig eher technisch | spaeter in Presets uebersetzen |
| Mitte | `Vollbildmodus` | ablenkungsfreies Schreiben | Expliziter Aufruf | nur Manuskript sehen | bereits stark verbessert | MVP-Schluesselfunktion |
| Rechte Sidebar | `Wiki-Links im Kapitel` | sichtbare Wissensreferenzen zeigen | Kapitelabhaengig | alle `[[...]]`-Links des Kapitels sehen | inhaltlich wertvoll, aber nicht immer dringlich | Standard kompakt |
| Rechte Sidebar | `Anker im Kapitel` | unsichtbare Verknuepfungen pruefen | Kapitelabhaengig | verankerte Passagen kontrollieren | mental abstrakter als Wiki-Links | mit klareren Beispielen unterstuetzen |
| Rechte Sidebar | `Clipboard & Slots` | Snippets sammeln, pinnen, wieder einsetzen | Immer sichtbar plus Floating Palette | Zitat auf Slot 2 legen und spaeter einsetzen | stark, aber fuer neue Nutzer dicht | wichtiger Kandidat fuer sanftere Einfuehrung |
| Floating | `Clipboard-Palette` | alle Clipboard-Eintraege zentral verwalten | Explizit per FAB oder Button | Snippet fixieren, Slot vergeben | leistungsstark, aber noch zweite Bedienebene | bewusst als Fortgeschrittenen-Werkzeug markieren |
| Global | `Autosave / Save-Status` | Vertrauen und Sicherheit | Immer sichtbar | `Autosave gespeichert` sehen | noetig und gut | beibehalten |

## Autoren-Aufgaben statt Feature-Denken

Die aktuelle App laesst sich fuer den MVP besser in sechs Autorinnen-/Autoren-Aufgaben gliedern:

| Aufgabe | Primäre UI | Sekundäre UI | Ziel |
| --- | --- | --- | --- |
| Schreiben | Editor, Vollbild, Schreibmodus | Einstellungen | Text erzeugen |
| Sammeln | Clipboard, Wissensbank | Floating Clipboard | Material sichern |
| Verknuepfen | Textauswahl, Anker, Wiki-Link | Workflow-Zielbox | Bedeutung an Text haengen |
| Strukturieren | Kapitel, Buecher, Projekte | Workflow-Boxen | Ordnung halten |
| Erinnern | Workflow-Boxen, Reminder-/Research-Typen | Ankerliste | Offenes spaeter wiederfinden |
| Wiederverwenden | Slots, Clipboard-Einfuegen | Wissens-Link-Insert | Vorhandenes Material erneut nutzen |

## Wichtigste Hilfsmittel mit Praxisbeispielen

| Hilfsmittel | Wann benutze ich es? | Mini-Beispiel |
| --- | --- | --- |
| `Kapitel` | Wenn ich die Manuskriptstruktur bewege | Kapitel 4 vor Kapitel 3 ziehen |
| `Workflow-Zielbox` | Wenn eine Passage bewusst einer Denkspur zugeordnet werden soll | Absatz zu `Recherche 1969` verankern |
| `Selection Popup` | Wenn markierter Text sofort verwertet werden soll | markierten Satz direkt `Zu Personen` schicken |
| `Clipboard` | Wenn Formulierungen spaeter wiederverwendet werden sollen | Dialogsatz sichern und spaeter erneut einsetzen |
| `Slots 1-9` | Wenn wenige Snippets haeufig wiederkehren | Standard-Notiz oder Lieblingszitat schnell einsetzen |
| `Wissensbank` | Wenn Begriffe/Personen als wiederkehrende Referenzen dienen | `[[Person:Mara]]` oder `[[Ort:Alter Garten]]` einfuegen |
| `Rich/Markdown` | Wenn zwischen Schreibfokus und Quelltextkontrolle gewechselt werden muss | in Rich schreiben, in Markdown pruefen/exportieren |
| `Tabellen-Werkzeuge` | Wenn strukturierte Inhalte noetig sind | Figurenmatrix oder Zeitachse als Tabelle anlegen |
| `Zitat/Fussnote` | Wenn redaktionelle Strukturen noetig sind | Quellenzitat und passende Fussnote setzen |
| `Vollbild` | Wenn nur der Text zaehlt | alles ausblenden und fokussiert schreiben |

## Was heute schon stark ist

- Der Editor besitzt bereits echten Schreibnutzen, nicht nur CRUD-Verhalten.
- Clipboard, Slots, Anker und Wissenslinks bilden zusammen ein ungewoehnlich starkes Autoren-Werkzeugset.
- Workflow-Trigger und Kontext-Popup haben grosses Potenzial fuer eine wirklich intelligente Schreiboberflaeche.
- Vollbild, Typografie-Optionen und die gedaempfte UI bringen die App in Richtung eines ruhigen Schreibstudios.

## Aktuelle Schwachstellen

### 1. Zu viele gleichrangige Signale

Viele Bereiche wirken funktional wichtig, aber nicht hierarchisch geordnet.

Folge:

- Nutzer sehen viel
- verstehen aber nicht sofort, was zuerst relevant ist

### 2. Workflow-Boxen sind maechig, aber kognitiv teuer

Die Boxen koennen:

- Ziele sein
- Trigger tragen
- Anker aufnehmen
- Inhalte spiegeln
- manuell bearbeitet werden

Das ist stark, aber ohne klare Einfuehrung schwer lesbar.

### 3. Wissensbank, Anker und Workflow ueberlappen mental

Diese drei Systeme sind konzeptionell verschieden:

- Wissensbank = benannte Referenzobjekte
- Anker = unsichtbare Textverknuepfung
- Workflow-Box = Arbeits- oder Denkraum

In der UI ist dieser Unterschied noch nicht deutlich genug erklaert.

### 4. Clipboard ist stark, aber fast schon ein eigenes Subsystem

Pinned Slots, Palette, FAB, Direkt-Einfuegen und Capture sind wertvoll, aber fuer ein MVP muessen wir die Haupteinstiege strenger priorisieren.

## Empfohlene MVP-Nutzungsstrategie

### Kernpfad fuer Autorinnen und Autoren

1. Projekt/Buch/Kapitel waehlen
2. Im `Rich`-Editor schreiben
3. Nur bei Bedarf:
   - Text markieren
   - `Clipboard`, `Anker` oder `Wiki-Link` nutzen
4. Workflow-Boxen nur fuer aktive Schreibspuren pflegen
5. Wissensbank nur fuer wiederkehrende Entitaeten nutzen
6. Markdown-Modus nur fuer Kontrolle, Spezialfaelle oder Exportnaehe verwenden

### Empfohlene Sichtbarkeits-Priorisierung

| Priorität | Soll standardmäßig präsent sein | Soll eher auf Anforderung erscheinen |
| --- | --- | --- |
| Hoch | Kapitel, Editor, Save-Status, Vollbild | — |
| Mittel | aktive Zielbox, kompakte Wissenshinweise, letzte Clipboard-Snippets | ausfuehrliche Workflow-Bearbeitung |
| Niedrig | Trigger-Details, Slot-Verwaltung, ausfuehrliche Wissenspflege, Editor-Hilfe | Floating Clipboard, Box-Feineditierung, tiefe Settings |

## Konkrete Design-/Produktentscheidungen fuer den naechsten Schritt

### A. Drei Schichten statt einer Vollsicht

Die UI sollte bewusst in drei Ebenen lesbar werden:

1. `Schreiben`
2. `Verknuepfen`
3. `Organisieren`

### B. Workflow-Boxen staerker staffeln

Empfohlen:

- Standard: nur Titel, Status, Ankerzahl
- Aktiv: Zielbox + wichtigste Aktion
- Bearbeiten: Tags, Typ, Collapse, Details erst nach explizitem Edit

### C. Wissensbank als Referenz-Werkzeug fuehren

Die Wissensbank sollte sich subjektiv weniger wie eine zweite Sidebar und mehr wie ein optionales Nachschlage-/Insert-Werkzeug anfuehlen.

### D. Clipboard als Autoren-Sammelmappe rahmen

Nicht jede Clipboard-Funktion muss permanent sichtbar sein.
Fuer den MVP reicht als Primaerpfad:

- Text markieren
- `In Clipboard uebernehmen`
- spaeter ueber Palette oder Slots wiederverwenden

## Offene Leitfragen fuer die naechste gemeinsame Runde

Diese Fragen helfen uns beim weiteren Feinschliff:

1. Was ist fuer deinen echten Schreiballtag `taeglich`, `gelegentlich` und `selten`?
2. Welche drei Helfer sollen waehrend des Schreibens praktisch immer greifbar bleiben?
3. Welche Bereiche duerfen klar in einen `Erweitert`-Modus rutschen?
4. Soll die Wissensbank eher:
   - stiller Referenzspeicher
   - aktive Begriffswolke
   - oder halbautomatischer Story-Bundler werden?
5. Sollen Workflow-Boxen langfristig eher:
   - Denkraeume
   - Checklisten
   - semantische Trigger
   - oder Szenen-/Story-Steuerung sein?

## Konkrete Empfehlung von Cody

Fuer den MVP sollte `easy-author` zunaechst als:

> ruhiger Markdown-/Rich-Schreibraum mit intelligenter Textauswahl, Ankerung, Wissenslinks und wiederverwendbarem Clipboard

positioniert werden.

Nicht als voll ausinszeniertes Gesamt-Produktionssystem in jeder sichtbaren Ebene gleichzeitig.

Das Potenzial fuer Story-Kalender, Bundler, Regie-Planung und tiefere Workflow-Orchestrierung ist klar vorhanden — aber fuer Orientierung und Nutzbarkeit sollten diese Ebenen schrittweise ueber den Kernpfad gelegt werden.
