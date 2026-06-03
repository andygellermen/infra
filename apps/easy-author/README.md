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
- SQLite-Initialisierung mit Demo-Daten beim ersten Start
- Markdown-Snapshots pro Kapitel unter `backend/data/library/...`
- React-Frontend mit dreispaltigem Autoren-Cockpit
- Tiptap-Editor mit Autosave, manuellem Speichern und Slot-Shortcuts `Cmd/Ctrl + Shift + 1-9`
- Rechte Sidebar fuer Anker, Clipboard und gepinnte Slots

## Lokales Setup

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

```bash
cd apps/easy-author/frontend
npm install
npm run dev
```

Frontend-URL: `http://127.0.0.1:5173`

Die Vite-Konfiguration proxyt `/api` automatisch auf das lokale Go-Backend.

## Datenbank und Inhalte

- SQLite-Datei: `apps/easy-author/backend/data/easy-author.sqlite`
- Markdown-Snapshots: `apps/easy-author/backend/data/library/<project>/<book>/chapters/*.md`
- Beim ersten Backend-Start wird ein Demo-Projekt mit Buch, Kapitel und Workflow-Boxen angelegt.

## API-Ueberblick

- `GET /api/health`
- `GET/POST /api/projects`
- `GET /api/projects/:id`
- `GET /api/books/:id`
- `POST /api/projects/:projectId/books`
- `GET/POST /api/books/:bookId/chapters`
- `GET/PUT /api/chapters/:id`
- `GET/POST /api/books/:bookId/workflow-boxes`
- `PUT /api/workflow-boxes/:id`
- `GET/POST /api/chapters/:chapterId/anchors`
- `DELETE /api/anchors/:id`
- `GET/POST /api/books/:bookId/clipboard`
- `PUT/DELETE /api/clipboard/:id`

## Bekannte Einschraenkungen des Spikes

- Kein Login, keine Benutzerverwaltung, keine Kollaboration
- Kein echter Markdown-Roundtrip fuer alle Sonderfaelle; der Serializer deckt den StarterKit-Grundumfang ab
- Kein PDF/EPUB/DOCX-Export
- Kein Asset-Management und keine Wissensbank-Entitaeten ausser Workflow-Boxen
- Keine Kapitel-Reihenfolge per Drag-and-drop

## Naechste sinnvolle Schritte

1. Saubere Markdown-Import/Export-Pipeline mit robusterem Roundtrip ausbauen
2. Wissensbank-Objekte und sichtbare `[[...]]`-Links ergaenzen
3. Workflow-Boxen mit Kapitel- und Textfilterlogik vertiefen
4. Asset-Verwaltung und Export-Jobs nachziehen
5. Spaetere Benutzer- und Kommentarfluesse auf die vorhandenen Anchor-Modelle aufsetzen
