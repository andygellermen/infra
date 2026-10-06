# easy-author

`easy-author` ist der erste lauffaehige MVP-Spike fuer ein self-hosted Autoren-Studio mit Go-Backend, React/Tiptap-Frontend und SQLite-Speicherung.

Die App setzt bewusst auf eine Hybridstrategie:

- Markdown bleibt das fuehrende Autoren- und Exportformat.
- Tiptap-/ProseMirror-JSON bleibt als Editor-Snapshot erhalten.
- SQLite speichert Beziehungen, Workflow-Boxen, Anker und Clipboard-Items.

## Struktur

```text
apps/easy-author/
  backend/
  frontend/
  docs/
  docker-compose.yml
  README.md
```

## Enthalten im ersten Spike

- REST-API fuer Projekte, Buecher, Kapitel, Workflow-Boxen, Anker und Clipboard
- Erste Wissensbank mit `[[...]]`-Links fuer Personen, Orte, Ereignisse und weitere Knowledge-Typen
- SQLite-Initialisierung mit Demo-Daten beim ersten Start
- Markdown-Snapshots pro Kapitel unter `backend/data/library/...`
- Ruhiger, editorzentrierter Schreibraum mit einer bei Bedarf eingeblendeten Kontrollleiste
- Buchbezogene Arbeitsansichten `Clean & Free`, `Intense` und `Review`
- React-Frontend mit optionalen Werkzeug- und Kontextbereichen
- Tiptap-Editor mit Autosave, manuellem Speichern, Tabellen-Tools, Zitat-/Fussnoten-Helfern und Slot-Shortcuts `Cmd/Ctrl + Shift + 1-9`
- Umschaltbarer Editor zwischen Rich-Ansicht und rohem Markdown pro Kapitel
- Editor-Hilfe und Editor-Einstellungen im zentralen Schreibbereich fuer Markdown-Support, Tabellen, Clipboard-Slots und Workflow-Anker
- Rechte Sidebar fuer Wiki-Link-Kontext, Anker, Clipboard und gepinnte Slots
- Dauerhafte Textkontexte fuer Kommentare, Aufgaben, Verknuepfungen und Clipboard-Bezuege
- Threadfaehige Kommentare und ein synchronisiertes Kanban mit fuenf Phasen

## UI-MVP bedienen

Die Drei-Punkte-Schaltflaeche rechts oben blendet die Kontrollleiste ein. Sie
verschwindet nach 60 Sekunden Inaktivitaet, bleibt bei Fokus, Mausinteraktion
oder geoeffnetem Dialog jedoch stehen. Alle zentralen Aktionen sind als native
Schaltflaechen per Tabulator erreichbar; `Escape` schliesst offene Arbeits- und
Hilfsansichten.

Beim ersten Anlegen eines Buches wird eine Arbeitsansicht gewaehlt. Der
Buchstandard wird dauerhaft gespeichert; ein spaeterer Wechsel fuer die
laufende Sitzung veraendert ihn nicht ungefragt:

- `Clean & Free` zeigt den Schreibtext mit minimaler Umgebung.
- `Intense` stellt Struktur, Recherche und Verknuepfungen bereit.
- `Review` fokussiert Kommentare, Aufgaben und Pruefhinweise.

Hell, Dunkel oder das aktuelle Systemfarbschema gelten global. Typografie kann
global vorgegeben und pro Buch feldweise ueberschrieben werden. Nicht
ueberschriebene Schriftarten, H1-H6-Groessen, Fliesstext-, Zitat- und
Tabellengroessen sowie Textbreite, Erstzeileneinzug, Zeilen- und Absatzabstand
erben weiterhin den Gesamtstandard.

Eine Textauswahl kann mit mehreren unabhaengigen Kontexten verbunden werden.
Kommentar-Threads besitzen Antworten und bleiben nach dem Erledigen als graue
Markierung erhalten; ihre Anzeige kann in den Editor-Einstellungen abgeschaltet
werden. Das explizite Loeschen eines Kontextes entfernt keine Geschwister am
selben Textanker.

Das Kanban wird ueber `Kanban oeffnen` in der temporaeren Kontrollleiste
erreicht. Es bietet `Backlog`, `To-do`, `In Arbeit`, `Pruefung` und `Fertig`.
Karten lassen sich ziehen, mit den Pfeilschaltflaechen oder fokussiert mittels
`Alt` + `Pfeil links/rechts` bewegen. Die Buecheruebersicht erlaubt eine
Mehrfachauswahl; ihre kontrastreichen Buchfarben gelten nur fuer die aktuelle
Ansicht und werden nicht in den Buechern gespeichert. Je Phase werden zunaechst
12 und nach bewusstem Nachladen hoechstens 20 Karten dargestellt.

## Lokales Setup

### Schnellstart

```bash
cd apps/easy-author
./dev.sh
```

Das Script startet zuerst das Go-Backend, wartet auf den Healthcheck und startet danach automatisch das Frontend.

### Backend

```bash
cd apps/easy-author/backend
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go mod tidy
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go run ./cmd/server
```

Backend-URL: `http://127.0.0.1:8086`

Healthcheck:

```bash
curl -fsS http://127.0.0.1:8086/api/health
```

### Frontend

Voraussetzung: `Node 20.x`. Das Frontend wurde in diesem Repo erfolgreich mit `node v20.20.2` gebaut.

Wenn lokal noch ein altes globales `node` aktiv ist, hilft zum Beispiel:

```bash
export PATH="/usr/local/opt/node@20/bin:$PATH"
node --version
```

Oder mit `nvm` direkt aus der mitgelieferten Datei:

```bash
cd apps/easy-author/frontend
nvm use
```

Wenn `nvm` in `zsh` geladen ist, wird die passende Version beim Betreten des Frontend-Ordners automatisch aus `.nvmrc` aktiviert.

```bash
cd apps/easy-author/frontend
npm install
npm run dev
npm run test:markdown
npm run test:editor
npm test
```

Wenn du nicht beide Prozesse getrennt starten willst, nutze stattdessen einfach:

```bash
cd apps/easy-author
./dev.sh
```

Frontend-URL: `http://127.0.0.1:5173`

Die Vite-Konfiguration proxyt `/api` automatisch auf das lokale Go-Backend.

### Letzte MVP-Smoke-Pruefung (2026-10-06)

- Backend lokal gestartet; `/api/health` antwortete mit `status: ok`.
- Projektliste blieb lesbar und der Kanban-Endpunkt lieferte alle fuenf Phasen.
- Vite-Frontend startete auf Port 5173; Startseite und API-Proxy waren erreichbar.
- `docker compose config --quiet` war erfolgreich. Ein Container-Smoke konnte
  in der Pruefumgebung nicht zusaetzlich gestartet werden, weil dort kein
  Docker-Daemon lief; dies ist eine Umgebungsgrenze, kein Compose-Fehler.

### Markdown-Workflow

- Jedes Kapitel kann zwischen `Rich` und `Markdown` umgeschaltet werden.
- Im Markdown-Modus bleibt Markdown die fuehrende Quelle; beim Speichern wird der Editor-Snapshot neu erzeugt.
- `[[...]]`-Wiki-Links, einfache Pipe-Tabellen, Blockquotes, Fussnoten, ausgewaehlte Textstellen, Clipboard-Einfuegen und manuelles Speichern funktionieren in beiden Modi.
- Im Rich-Editor stehen Tabellenwerkzeuge fuer Einfuegen, Zeilen/Spalten und Kopfzeilen bereit; abgetippte Pipe-Tabellen werden beim Weiterschreiben automatisch in Rich-Tabellen uebernommen.
- `Zitat` schaltet Blockquotes um; `Fussnote` erzeugt im Rich-Editor Referenz plus Notizblock und im Markdown-Modus eine passende `[^1]`-Struktur.
- `npm run test:editor` deckt Editor-Smoke-Tests fuer Laden, Moduswechsel, Speichern, Textauswahl, Anker, Wissensbank-Link-Einfuegen, Clipboard, Mehrfach-Slot-Pinning, Loeschen/Freigeben von Slots, offene Wiki-Referenzen sowie mehrfache Fehler-, Anchor-/Clipboard-, Workflow-/Knowledge-, Create-/Switch-Recovery-, Retry-, Save-Konflikte, parallele Workflow-Sidebar-Aktionen und volle Mehrfach-Workflow-Last inklusive Autosave im Markdown-Modus ab.

## Datenbank und Inhalte

- SQLite-Datei: `apps/easy-author/backend/data/easy-author.sqlite`
- Markdown-Snapshots: `apps/easy-author/backend/data/library/<project>/<book>/chapters/*.md`
- Beim ersten Backend-Start wird ein Demo-Projekt mit Buch, Kapitel und Workflow-Boxen angelegt.

## Orientierung und UI-Kassensturz

- Strategische UI-/Autoren-Uebersicht: `apps/easy-author/docs/business-case/easy-author-author-strategy-audit.md`

## API-Ueberblick

- `GET /api/health`
- `GET/POST /api/projects`
- `GET /api/projects/:id`
- `GET/POST /api/projects/:projectId/knowledge-items`
- `GET /api/books/:id`
- `POST /api/projects/:projectId/books`
- `GET/POST /api/books/:bookId/chapters`
- `GET/PUT /api/chapters/:id`
- `GET/POST /api/books/:bookId/workflow-boxes`
- `PUT /api/workflow-boxes/:id`
- `PUT /api/knowledge-items/:id`
- `GET/POST /api/chapters/:chapterId/anchors`
- `DELETE /api/anchors/:id`
- `GET/POST /api/books/:bookId/clipboard`
- `PUT/DELETE /api/clipboard/:id`
- `GET/PUT /api/books/:bookId/presentation`
- `GET/POST /api/chapters/:chapterId/contexts`
- `POST /api/threads/:threadId/messages`
- `PUT /api/threads/:threadId`
- `DELETE /api/contexts/:contextId`
- `GET /api/kanban?book_ids=...`
- `POST /api/books/:bookId/work-items`
- `PUT /api/work-items/:workItemId/phase`

## Bekannte Einschraenkungen des Spikes

- Kein Login, keine Benutzerverwaltung, keine Kollaboration
- Kein vollstaendiger Markdown-Roundtrip fuer alle Sonderfaelle; der Parser/Serializer deckt aktuell Ueberschriften, verschachtelte Listen, Zitate, Code-Fences, Trennlinien, harte Umbrueche, Escaping und Basis-Inline-Markup ab
- Kein PDF/EPUB/DOCX-Export
- Kein Asset-Management und keine Wissensbank-Entitaeten ausser Workflow-Boxen
- Die globale Kanban-Ansicht ist fuer grosse Bibliotheken bewusst auf maximal 20 sichtbare Karten je Phase begrenzt

## Naechste sinnvolle Schritte

1. Markdown-Roundtrip fuer verschachtelte Listen, Escaping und Sonderfaelle weiter haerten
2. Wissensbank-Objekte und sichtbare `[[...]]`-Links weiter vertiefen
3. Workflow-Boxen mit Kapitel- und Textfilterlogik ausbauen
4. Asset-Verwaltung und Export-Jobs nachziehen
5. Spaetere Benutzer- und Kommentarfluesse auf die vorhandenen Anchor-Modelle aufsetzen
