# Static Inline Editor

Ziel ist eine kontrollierte, versionierte und moeglichst einfache Bearbeitung statischer Webseiten direkt im Browser, ohne daraus ein vollwertiges CMS zu machen.

Die Bearbeitung soll nur fuer autorisierte Nutzer moeglich sein und nur definierte Inhaltsbereiche betreffen. Oeffentliche Besucher sehen weiterhin reine statische Seiten ohne Editor-JavaScript.

## Zielbild

Pro statischer Domain gibt es zwei Betriebsmodi:

- oeffentliche Auslieferung unter `domain.de`
- geschuetzte Bearbeitung unter `bearbeitung.domain.de`

Die Bearbeitung erfolgt inline im Browser, aber nur nach erfolgreicher Authentifizierung. Erst danach werden die benoetigten Editor-Bibliotheken geladen.

## Grundprinzip

Die statische Seite bleibt die Quelle fuer die Darstellung. Die Bearbeitungs-App ist eine separate, kleine Begleit-Anwendung in Go, die folgende Aufgaben uebernimmt:

- Authentifizierung
- Session-Management
- Freigabe fuer Edit-Modus
- Laden und Validieren bearbeitbarer Inhalte
- Vorschau, Abnahme und Speichern
- Backup vor dem Schreiben
- Git-Commit nach erfolgreichem Speichern

## Empfohlene Komponenten

### 1. Statische Site

Die statische Site liegt wie bisher auf dem Server, zum Beispiel unter:

- `/srv/static/domain.de/index.html`

Sie wird weiterhin normal ueber den bestehenden Static-Stack ausgeliefert.

### 2. Edit-App in Go

Die Edit-App laeuft zentral in einem eigenen Container oder als Erweiterung des bestehenden Infrastrukturmodells.

Aufgaben:

- Login unter `bearbeitung.domain.de`
- Setzen eines Session-Cookies
- serverseitige Pruefung, welche Domain bearbeitet werden darf
- gezieltes Injektieren von Editor-JavaScript erst nach Authentifizierung
- Entgegennahme und Validierung der Aenderungen
- Backup und Git-Workflow

### 3. Inline-Editor im Browser

Als Frontend-Bibliothek ist `ContentTools` passend.

Sie wird nur nach erfolgreichem Login geladen und initialisiert.

Vorteile:

- leichtgewichtig
- inline editing faehig
- keine schwere Admin-Oberflaeche noetig

## Sicherheitsmodell

### Oeffentliche Seite

Oeffentliche Nutzer sehen:

- keine Edit-Bibliotheken
- keine Bearbeitungscontrols
- keine Admin-Endpunkte

### Bearbeitungsseite

Der Zugriff auf `bearbeitung.domain.de` erfolgt nur nach Authentifizierung.

Nach erfolgreichem Login:

- Session-Cookie setzen
- nur dann Editor-Bibliotheken laden
- nur dann Edit-Endpoints freischalten

### Wichtig

Das Editor-JavaScript darf nicht standardmaessig auf der oeffentlichen Seite eingebunden werden. Genau das ist einer der wichtigsten Punkte fuer Sicherheit und Robustheit.

## Bearbeitungsmodell

Die App sollte nicht beliebiges HTML speichern, sondern nur klar erlaubte semantische Textbereiche.

### Fuer den MVP erlaubte Elemente

- `h1`
- `h2`
- `h3`
- `h4`
- `h5`
- `p`
- `ul`
- `ol`
- `li`
- optional spaeter `blockquote`, `a`, `strong`, `em`

### Fuer den MVP nicht direkt freigeben

- beliebige `div`-Container
- Skripte
- Inline-Events
- frei editierbare Klassen oder Styles
- Formulare
- eingebettete Drittanbieter-Snippets

## Markierung der bearbeitbaren Bereiche

Du moechtest moeglichst ohne separate `editable`-Klassen arbeiten. Das ist machbar, sollte aber kontrolliert bleiben.

Ich wuerde methodisch zwei Stufen unterscheiden.

### Stufe A: semantische Standard-Elemente

Bearbeitbar sind nur typische Textcontainer innerhalb eines klar definierten Hauptbereichs, zum Beispiel innerhalb:

- `main`
- `.content`
- `article`

Die App oder das Editor-JavaScript aktiviert dann nur:

- `main h1`
- `main h2`
- `main p`
- `main ul`
- `main ol`
- `main li`

### Stufe B: serverseitig zusaetzlich absichern

Beim Speichern validiert die Go-App noch einmal:

- welche Elemente geaendert wurden
- ob nur erlaubte Tags enthalten sind
- ob unerlaubte Attribute oder Skripte eingebracht wurden

Das ist wichtiger als die reine Frontend-Beschraenkung.

## Vorschau- und Save-Workflow

Empfohlener Ablauf:

1. Login auf `bearbeitung.domain.de`
2. Session-Cookie wird gesetzt
3. Editor-JavaScript wird geladen
4. Nutzer bearbeitet erlaubte Textbereiche inline
5. Nutzer klickt auf `Vorschau`
6. Go-App validiert und rendert die geaenderte Version als Vorschau
7. Nutzer bestaetigt `Speichern`
8. Backup der bisherigen Datei wird angelegt
9. HTML-Datei wird gezielt aktualisiert
10. `git add`
11. `git commit`

### Optional spaeter

- `git push`
- Webhook oder Redeploy
- Rollback aus der UI

## Speichermodell

Hier gibt es zwei moegliche Strategien.

### Strategie 1: ganze HTML-Datei speichern

Einfacher Start, aber hoeheres Risiko, dass zu viel geaendert wird.

### Strategie 2: nur definierte Bereiche ersetzen

Empfehlung fuer deinen Fall.

Die App:

- liest die bestehende HTML-Datei
- parst sie serverseitig
- ersetzt nur erlaubte Zielknoten
- schreibt danach die neue Datei zurueck

Das ist deutlich kontrollierter.

## Serverseitige Validierung in Go

Go eignet sich dafuer sehr gut.

Empfohlene Aufgaben der Go-Validierung:

- HTML parsen
- erlaubte Tags pruefen
- unerlaubte Tags entfernen oder ablehnen
- Attribute whitelist-basiert zulassen
- Textbereiche gegen das bestehende Dokument matchen
- Ausgabe wieder als sauberes HTML serialisieren

Wichtig:

- Browser-Input nie direkt ungeprueft auf Platte schreiben
- immer serverseitig normalisieren

## Backup-Konzept

Vor jedem Speichern:

1. bestehende Datei kopieren
2. Zeitstempel vergeben
3. unter separatem Backup-Pfad sichern

Beispiel:

- `/srv/static-backups/domain.de/2026-04-01T12-34-56_index.html`

Optional:

- nur letzte X Versionen aufbewahren
- taegliche oder woechentliche Verdichtung spaeter

## Git-Versionierung

Git ist fuer diesen Use Case sehr sinnvoll.

### Empfohlene Rolle von Git

- nachvollziehbare Aenderungshistorie
- einfache Rollbacks
- klare Commit-Messages
- optional spaeter PR- oder Review-Workflow

### Wo sollte das Repo leben

Mein Favorit:

- dort, wo die statische Site ohnehin als Quelle verwaltet wird

Wichtig ist:

- es gibt genau eine Wahrheit
- keine konkurrierenden Live-Dateien und getrennten Shadow-Repos

### Commit-Beispiel

```text
edit(domain.de): index.html inline aktualisiert
```

Oder konkreter:

```text
edit(geller.men): startseite hero-text angepasst
```

## Empfohlene Verzeichnisstruktur

Beispiel:

```text
/srv/static/domain.de/
/srv/static-backups/domain.de/
/srv/static-git/domain.de/.git
```

Alternativ:

- das Git-Repo liegt direkt auf `/srv/static/domain.de`, wenn das mit eurem Static-Workflow sauber zusammenpasst

## Authentifizierung

Fuer den MVP reicht:

- Magic-Link per E-Mail
- Session-Cookie
- Rate-Limiting

Optional spaeter:

- mehrere Nutzer
- Rollen
- Freigabe-Workflow

## Routing-Modell

Empfohlene Aufteilung:

- `domain.de` -> oeffentliche statische Site
- `bearbeitung.domain.de` -> Edit-App

Nach Login kann die Edit-App:

- die Seite mit Editor-Overlay ausliefern
- oder einen editierbaren Proxy auf die statische Seite legen

## Speicherung des Editor-Zustands

Der Browser sollte nicht direkt auf Dateien schreiben.

Stattdessen:

- `GET /edit/load?path=/index.html`
- `POST /edit/preview`
- `POST /edit/save`

Diese Endpunkte laufen alle ueber die Go-App und nur mit gueltiger Session.

## Empfohlener MVP

Ein guter erster MVP waere:

- Login auf `bearbeitung.domain.de`
- Session-Cookie
- ContentTools nur nach Auth laden
- Bearbeitung von `h1` bis `h5`, `p`, `ul`, `ol`, `li`
- Vorschau
- Backup
- Save
- Git commit

Noch nicht im ersten Wurf:

- komplexe Mehrnutzerverwaltung
- Medienverwaltung
- frei editierbare Layout-Container
- WYSIWYG fuer beliebiges HTML
- automatischer Push oder Review-Workflow

## Risiken und Gegenmassnahmen

### Risiko: zu breite Bearbeitbarkeit

Gegenmassnahme:

- nur erlaubte Elemente
- nur innerhalb definierter Hauptbereiche

### Risiko: fehlerhaftes HTML wird gespeichert

Gegenmassnahme:

- serverseitiges Parsen und Serialisieren

### Risiko: Editor-JS auf oeffentlicher Seite sichtbar

Gegenmassnahme:

- Bibliotheken erst nach Authentifizierung laden

### Risiko: versehentliches Ueberschreiben

Gegenmassnahme:

- Vorschau
- Backup
- Git commit

## Technischer Stack

Empfehlung:

- Go fuer Serverlogik
- `html/template` fuer Admin-/Login-Seiten
- `golang.org/x/net/html` oder aehnlicher Parser fuer HTML-Manipulation
- ContentTools fuer das Inline-Editing
- SQLite optional fuer Sessions oder Edit-Logs
- Git ueber sichere serverseitige Prozesse

## Mein Fazit

Die Idee ist sehr gut machbar und fuer deinen statischen Website-Betrieb sinnvoll.

Der richtige Zuschnitt ist:

- kein grosses CMS
- keine beliebige HTML-Freiheit
- sondern eine kleine, sichere Edit-Begleit-App mit:
  - Auth
  - kontrolliertem Inline-Editing
  - Vorschau
  - Backup
  - Git-Versionierung

## Naechste sinnvolle Schritte

1. exakte Bearbeitungsgrenzen fuer den MVP festlegen
2. Save- und Preview-Workflow konkret beschreiben
3. entscheiden, wo das Git-Repo liegen soll
4. danach das Go-Service-Design fuer diese Edit-App aufsetzen

## Konkretes MVP-Design

Das folgende MVP-Design ist bewusst klein, kontrolliert und auf schnelle Textkorrekturen zugeschnitten.

### Ziel des MVP

Das MVP soll genau diese Dinge koennen:

- Login auf `bearbeitung.domain.de`
- geschuetzte Bearbeitung einer bestehenden HTML-Seite
- Vorschau vor dem Speichern
- Backup vor dem Schreiben
- Git-Commit nach erfolgreichem Speichern

Nicht Ziel des MVP:

- Medienverwaltung
- Layout-Bearbeitung
- frei konfigurierbare Widgets
- Mehrbenutzer-Rechtesystem
- Workflow mit Review-Freigaben

## Laufzeitrollen

### Oeffentliche Auslieferung

Die statische Site unter `domain.de` bleibt unveraendert und wird weiterhin durch den bestehenden Static-Stack ausgeliefert.

### Edit-App

Die Edit-App unter `bearbeitung.domain.de` uebernimmt:

- Magic-Link-Login
- Session
- Laden der Seite fuer den Edit-Modus
- Validierung und Vorschau
- Speichern
- Backup
- Git-Versionierung

## Go-Service-Aufbau

Fuer den MVP wuerde ich den Service in diese Module teilen:

- `internal/config`
- `internal/auth`
- `internal/session`
- `internal/editor`
- `internal/htmlsanitize`
- `internal/storage`
- `internal/gitops`
- `internal/httpapp`

### Verantwortung der Module

`config`

- Domain-Konfiguration
- Dateipfade
- Auth-Daten
- Git-Repo-Pfad

`auth`

- Magic-Link anfordern
- E-Mail-Allowlist pruefen
- Token pruefen

`session`

- Session-Cookie
- Session-Speicherung
- Session-Pruefung

`editor`

- Laden der Ziel-Datei
- Identifikation bearbeitbarer Elemente
- Aufbereitung fuer Vorschau und Speichern

`htmlsanitize`

- Pruefung erlaubter Tags
- Pruefung erlaubter Attribute
- Entfernung oder Ablehnung unerlaubter Inhalte

`storage`

- lokale Session-Daten
- optionale Edit-Logs

`gitops`

- Backup
- `git add`
- `git commit`

`httpapp`

- HTTP-Endpunkte
- Seitenrendering
- API fuer Preview und Save

## Konfigurationsmodell

Fuer jede bearbeitbare Domain braucht die App mindestens:

```yaml
domain: example.org
editor_enabled: true
editor_login_domain: bearbeitung.example.org
editor_static_root: /srv/static/example.org
editor_backup_root: /srv/static-backups/example.org
editor_repo_root: /srv/static/example.org
editor_cookie_secret: <secret>
editor_allowed_emails:
  - andy@example.org
editor_main_selector: main
```

Zusaetzlich global:

```yaml
smtp_host: email-smtp.eu-central-1.amazonaws.com
smtp_port: 587
smtp_username: <ses-smtp-user>
smtp_password: <ses-smtp-password>
smtp_from_email: no-reply@example.org
smtp_from_name: Static Editor
magic_link_ttl: 15m
```

Optional spaeter:

- mehrere Nutzer
- Rollen
- mehr als ein editierbarer Root-Selector

## Session-Modell

Fuer den MVP reicht ein serverseitig erzeugtes Session-Cookie plus ein serverseitig gespeicherter Magic-Link-Token.

Empfehlung:

- Magic-Link-Anforderung via `POST /auth/request-link`
- Magic-Link-Pruefung via `GET /auth/verify`
- Cookie `editor_session`
- `HttpOnly`
- `Secure`
- `SameSite=Lax`

Fuer den ersten lauffaehigen Zuschnitt ist auch ein file-basierter Store okay. Mittelfristig wuerde ich fuer Sessions und Tokens trotzdem SQLite bevorzugen, weil:

- einfacher Logout
- einfache Ablaufzeiten
- spaeter besser fuer Edit-Logs
- einfache One-Time-Token-Invalidierung

## HTTP-Endpunkte des MVP

### Auth

`GET /login`

- Formular fuer Magic-Link-Anforderung

`POST /auth/request-link`

- prueft E-Mail-Allowlist
- erzeugt Magic-Link-Token
- versendet E-Mail via SMTP
- liefert generische Erfolgsmeldung

`GET /auth/verify`

- prueft Magic-Link-Token
- setzt Session-Cookie
- Redirect auf Bearbeitungsstart

`POST /auth/logout`

- Session beenden

### Editor-Ansicht

`GET /`

- Dashboard oder Liste bearbeitbarer Seiten

`GET /edit`

Query-Parameter:

- `path=/index.html`

Aufgabe:

- Ziel-Datei laden
- HTML fuer Edit-Modus ausliefern
- Editor-JS nur bei gueltiger Session laden

### Vorschau

`POST /api/preview`

Aufgabe:

- Aenderungen entgegennehmen
- validieren
- als Vorschau-HTML zurueckgeben

### Speichern

`POST /api/save`

Aufgabe:

- Aenderungen validieren
- Backup anlegen
- Datei schreiben
- Git commit erzeugen

### Health

`GET /healthz`

Aufgabe:

- einfacher Betriebscheck

## Datenform fuer Preview und Save

Ich wuerde fuer den MVP keine komplette HTML-Datei vom Browser zuruecksenden, sondern eine strukturierte Liste geaenderter Knoten.

Beispiel:

```json
{
  "path": "/index.html",
  "changes": [
    {
      "selector": "main h1:nth-of-type(1)",
      "tag": "h1",
      "html": "Neue Ueberschrift"
    },
    {
      "selector": "main p:nth-of-type(2)",
      "tag": "p",
      "html": "Aktualisierter Absatz mit <strong>Betonung</strong>."
    }
  ]
}
```

Warum so:

- kleiner Payload
- keine komplette Seite wird blind ersetzt
- serverseitig leichter validierbar

## Identifikation bearbeitbarer Knoten

Der schwierigste Teil ist nicht das Speichern, sondern das stabile Wiederfinden der richtigen HTML-Bloecke.

Fuer den MVP wuerde ich folgenden Ansatz nehmen:

1. serverseitig die HTML-Datei parsen
2. nur Knoten innerhalb des Hauptselectors, z. B. `main`, betrachten
3. erlaubte Tags filtern
4. jedem editierbaren Knoten eine stabile Editor-ID zuweisen

Beispiel:

- `data-editor-id="node-0001"`
- `data-editor-id="node-0002"`

Dann sendet der Browser nicht komplizierte Selektoren zurueck, sondern diese IDs.

Besseres Preview-/Save-Payload fuer den MVP:

```json
{
  "path": "/index.html",
  "changes": [
    {
      "id": "node-0001",
      "tag": "h1",
      "html": "Neue Ueberschrift"
    },
    {
      "id": "node-0007",
      "tag": "p",
      "html": "Aktualisierter Absatz."
    }
  ]
}
```

Das ist mein klarer Favorit.

## HTML-Sanitization-Regeln

Serverseitig erlauben:

- `h1` bis `h5`
- `p`
- `ul`
- `ol`
- `li`
- `strong`
- `em`
- `a`
- `br`

Serverseitig verbieten:

- `script`
- `style`
- `iframe`
- `form`
- `input`
- `button`
- `img` fuer den ersten MVP
- Event-Attribute wie `onclick`
- freie `style`-Attribute

Bei Links:

- nur `href`
- optional `title`
- optional `target`
- `javascript:` strikt verbieten

## Vorschau-Workflow im Detail

1. Editor sendet Aenderungen an `POST /api/preview`
2. Go-App parst Originaldatei
3. Go-App wendet Aenderungen auf die erlaubten Knoten an
4. Go-App sanitizt den Inhalt
5. Go-App rendert Vorschau-HTML
6. Browser zeigt Vorschau vor finalem Speichern

## Save-Workflow im Detail

1. Editor sendet Aenderungen an `POST /api/save`
2. Go-App validiert Session
3. Go-App parst Originaldatei
4. Go-App matcht die `data-editor-id`-Knoten
5. Go-App sanitizt Inhalte
6. Go-App erzeugt Backup
7. Go-App schreibt aktualisierte Datei
8. Go-App fuehrt Git-Workflow aus
9. Go-App liefert Ergebnis an den Browser zurueck

## Backup-Workflow

Vor jedem Save:

- Quelldatei lesen
- Backup-Datei mit Zeitstempel anlegen

Beispiel:

```text
/srv/static-backups/example.org/2026-04-01T10-15-22_index.html
```

Optional in Phase 2:

- Rotation alter Backups

## Git-Workflow

Empfohlene Reihenfolge:

```bash
git add path/to/file.html
git commit -m "edit(example.org): index.html inline aktualisiert"
```

Optional spaeter:

- `git push`
- separater Branch
- PR-Workflow

Fuer den MVP wuerde ich lokal committen, aber nicht automatisch pushen.

## Fehlerverhalten

Wenn etwas schiefgeht, sollte die App klar unterscheiden:

### Validierungsfehler

- Antwort `400`
- konkrete Rueckmeldung, welcher Block unzulaessig ist

### Auth-Fehler

- Antwort `401` oder Redirect auf Login

### Schreibfehler

- Antwort `500`
- keine halbfertigen Zwischenstaende

### Git-Fehler

- Datei bleibt gespeichert
- Fehler wird klar angezeigt
- optional Kennzeichnung "Datei gespeichert, Commit fehlgeschlagen"

## Frontend-Verhalten mit ContentTools

Der Browser bekommt im Edit-Modus:

- die HTML-Seite
- nur nach Auth die ContentTools-Bibliothek
- eine kleine eigene Initialisierungslogik

Diese Initialisierungslogik:

- sucht erlaubte Knoten
- aktiviert Bearbeitung
- sammelt Aenderungen
- sendet Vorschau oder Save an die Go-App

## Warum keine freie WYSIWYG-Seite

Weil dein Use Case kein Layout-Builder ist, sondern ein verlässliches Korrektur-Werkzeug fuer kleine Textaenderungen.

Deshalb ist das MVP absichtlich:

- klein
- whitelist-basiert
- textorientiert
- git- und backup-freundlich

## Technische Empfehlung fuer Phase 1

Ich wuerde in dieser Reihenfolge bauen:

1. Login + Session
2. Seite im Edit-Modus ausliefern
3. editierbare Knoten mit `data-editor-id` markieren
4. `POST /api/preview`
5. `POST /api/save`
6. Backup
7. Git commit

## Mein Favort fuer den ersten Implementierungszuschnitt

Ganz konkret:

- nur `index.html` im ersten Schritt
- nur ein Hauptbereich `main`
- nur `h1` bis `h5`, `p`, `ul`, `ol`, `li`, `strong`, `em`, `a`
- nur ein Login pro Domain
- Vorschau und Save serverseitig in Go
- Backup und lokaler Git-Commit

Damit bleibt das MVP klein genug, um es wirklich solide fertigzubauen.

## Implementierungs-Bauplan fuer das MVP

Der folgende Abschnitt beschreibt die erste direkt implementierbare Fassung der App.

## Verzeichnisstruktur der neuen App

Ich wuerde die App so anlegen:

```text
apps/static-inline-editor/
  cmd/server/main.go
  internal/config/
  internal/auth/
  internal/session/
  internal/editor/
  internal/htmlsanitize/
  internal/gitops/
  internal/httpapp/
  internal/model/
  templates/
  assets/
  testdata/
```

### Kurzaufteilung

`cmd/server/main.go`

- Konfiguration laden
- Storage initialisieren
- HTTP-Server starten

`internal/model`

- Requests und Responses
- Session-Modelle
- Edit-Node-Modelle

`internal/editor`

- HTML-Datei laden
- editierbare Knoten finden
- `data-editor-id` erzeugen
- Knoten ersetzen

`internal/htmlsanitize`

- Whitelist fuer Tags und Attribute
- Sanitization fuer Browser-Input

`internal/gitops`

- Backup anlegen
- Git ausfuehren

`internal/httpapp`

- Login, Edit-View, Preview, Save

## Datenmodell fuer editierbare Knoten

Die App sollte serverseitig ein internes Modell fuer bearbeitbare Nodes haben.

Beispiel:

```go
type EditableNode struct {
    ID       string
    Tag      string
    Path     string
    TextHTML string
}
```

Praktisch braucht die App fuer jeden Knoten:

- stabile ID
- Tag-Typ
- Position im Dokument
- aktueller Inhalt

## Prinzip der `data-editor-id`

Beim Laden einer editierbaren Seite parst die App das Original-HTML und markiert nur erlaubte Knoten innerhalb des Hauptbereichs.

Beispiel:

Vorher:

```html
<main>
  <h1>Hallo Welt</h1>
  <p>Ein erster Absatz.</p>
</main>
```

Fuer den Edit-Modus:

```html
<main>
  <h1 data-editor-id="node-0001">Hallo Welt</h1>
  <p data-editor-id="node-0002">Ein erster Absatz.</p>
</main>
```

Die oeffentliche Seite sollte diese Markierungen nicht zwingend bekommen. Sie koennen speziell fuer den Edit-Modus erzeugt werden.

## Regel zur ID-Erzeugung

Fuer den MVP reicht eine stabile Reihenfolge waehrend des Dokumentscans:

- `node-0001`
- `node-0002`
- `node-0003`

Wichtig:

- die Reihenfolge wird bei jedem Laden aus dem aktuellen Dokument neu bestimmt
- Preview und Save muessen sich immer auf dieselbe frisch geladene Dateiversion beziehen

Fuer spaetere Robustheit waere ein Hash-Modell moeglich, aber fuer den MVP reicht die laufende Nummer innerhalb des geparsten Dokuments.

## HTTP-Endpunkte im Detail

### `GET /login`

Response:

- HTML-Seite mit Magic-Link-Formular

### `POST /auth/request-link`

Request:

```application/x-www-form-urlencoded
email=andy@example.org
```

Response:

- generische Erfolgsantwort ohne Leaken der Freigabeliste
- bei freigegebener Adresse:
  - Magic-Link-Token erzeugen
  - E-Mail via SMTP versenden

### `GET /auth/verify`

Request:

- `token` als Query-Parameter

Response:

- Magic-Link pruefen
- Session setzen
- Redirect auf `/`

### `POST /auth/logout`

Response:

- Session entfernen
- Redirect auf `/login`

### `GET /`

Response:

- einfache Startseite mit Links auf editierbare Pfade

Fuer den MVP reicht:

- `index.html`

### `GET /edit?path=/index.html`

Response:

- HTML-Seite im Edit-Modus
- ContentTools nur bei gueltiger Session
- editierbare Knoten mit `data-editor-id`

### `POST /api/preview`

Request:

```json
{
  "path": "/index.html",
  "changes": [
    {
      "id": "node-0001",
      "tag": "h1",
      "html": "Neue Ueberschrift"
    },
    {
      "id": "node-0002",
      "tag": "p",
      "html": "Aktualisierter Absatz mit <strong>Betonung</strong>."
    }
  ]
}
```

Response:

```json
{
  "ok": true,
  "preview_html": "<main>...</main>",
  "warnings": []
}
```

### `POST /api/save`

Request:

identisch zu `preview`

Response:

```json
{
  "ok": true,
  "saved_path": "/index.html",
  "backup_path": "/srv/static-backups/example.org/2026-04-01T10-15-22_index.html",
  "git_commit": "abc1234",
  "message": "Datei gespeichert"
}
```

## Request- und Response-Modelle in Go

Empfehlung:

```go
type EditChange struct {
    ID   string `json:"id"`
    Tag  string `json:"tag"`
    HTML string `json:"html"`
}

type PreviewRequest struct {
    Path    string       `json:"path"`
    Changes []EditChange `json:"changes"`
}

type PreviewResponse struct {
    OK          bool     `json:"ok"`
    PreviewHTML string   `json:"preview_html,omitempty"`
    Warnings    []string `json:"warnings,omitempty"`
    Error       string   `json:"error,omitempty"`
}

type SaveResponse struct {
    OK         bool   `json:"ok"`
    SavedPath  string `json:"saved_path,omitempty"`
    BackupPath string `json:"backup_path,omitempty"`
    GitCommit  string `json:"git_commit,omitempty"`
    Message    string `json:"message,omitempty"`
    Error      string `json:"error,omitempty"`
}
```

## Ablauf in `editor`

Die Kernlogik in `internal/editor` sollte etwa so aussehen:

### `LoadEditableDocument(path)`

Aufgaben:

- HTML-Datei lesen
- Dokument parsen
- editierbaren Root-Bereich finden, z. B. `main`
- erlaubte Knoten markieren
- Liste von `EditableNode` erstellen
- HTML fuer den Edit-Modus rendern

### `ApplyChanges(path, changes)`

Aufgaben:

- Originaldatei erneut lesen
- erneut parsen
- editierbare Knoten mit denselben IDs erzeugen
- eingehende Aenderungen gegen vorhandene Knoten matchen
- Inhalte sanitizen
- Knoten ersetzen
- neues HTML serialisieren

### `RenderPreview(path, changes)`

Aufgaben:

- `ApplyChanges` ausfuehren, aber noch nicht schreiben
- Ergebnis-HTML fuer Vorschau zurueckgeben

### `Save(path, changes)`

Aufgaben:

- `ApplyChanges` ausfuehren
- Backup anlegen
- Datei schreiben
- Git committen

## Regelwerk fuer bearbeitbare Bereiche

Fuer den MVP wuerde ich den Root-Bereich hart konfigurieren:

- `main`

Innerhalb von `main` sind nur diese Zieltags editierbar:

- `h1`
- `h2`
- `h3`
- `h4`
- `h5`
- `p`
- `ul`
- `ol`
- `li`

Inline innerhalb dieser Bereiche zulaessig:

- `strong`
- `em`
- `a`
- `br`

## Beispiel fuer serverseitige Knoten-Erkennung

Die App traversiert den DOM-Baum und prueft:

1. Liegt der Knoten innerhalb des konfigurierten Root-Bereichs?
2. Ist der Knoten selbst ein erlaubter Block-Tag?
3. Falls ja, bekommt er eine `data-editor-id`

So bleiben Header, Navigation, Footer und Skriptbereiche automatisch ausserhalb des Bearbeitungsmodus, wenn sie nicht im Root-Bereich liegen.

## Preview-Modus im Frontend

Im Browser reicht fuer den MVP ein sehr einfacher Vorschau-Ablauf:

1. Editor sammelt geaenderte Inhalte
2. `POST /api/preview`
3. Rueckgabe-HTML wird in einem Vorschau-Container oder Modal angezeigt
4. Nutzer bestaetigt oder bricht ab

Keine aufwendige Live-Diff-Ansicht im ersten Wurf.

## Save-Modus im Frontend

Nach Klick auf `Speichern`:

1. aktuelle geaenderte Node-Liste senden
2. Button waehrenddessen deaktivieren
3. Rueckmeldung von der App anzeigen
4. bei Erfolg optional Seite neu laden

## Backup- und Git-Implementierung

### Backup

Empfohlene Funktion:

```go
func CreateBackup(srcPath, backupRoot string, ts time.Time) (string, error)
```

Verhalten:

- erzeugt Backup-Datei mit Zeitstempel
- gibt finalen Backup-Pfad zurueck

### Git

Empfohlene Funktionen:

```go
func GitAdd(repoRoot string, filePath string) error
func GitCommit(repoRoot, message string) (string, error)
```

Rueckgabe von `GitCommit`:

- Commit-SHA oder Kurz-SHA

## Fehlertoleranz beim Speichern

Empfohlener Speicheralgorithmus:

1. Original lesen
2. neues HTML berechnen
3. Backup schreiben
4. neue Datei atomar schreiben
5. Git add
6. Git commit

Beim Schreiben:

- erst in temporäre Datei
- dann Rename auf Zieldatei

Das reduziert das Risiko beschaedigter Dateien.

## MVP-Startkonfiguration

Fuer den ersten produktiven Zuschnitt wuerde ich bewusst klein bleiben:

- nur `index.html`
- nur ein Benutzer
- nur ein Root-Selector `main`
- nur ein statischer Backup-Pfad
- nur lokaler Git-Commit
- keine parallelen Bearbeitungssessions

## Akzeptanzkriterien fuer das MVP

Das MVP ist aus meiner Sicht fertig, wenn folgendes funktioniert:

1. Login auf `bearbeitung.domain.de`
2. `index.html` wird im Edit-Modus geladen
3. `h1` und `p` koennen inline geaendert werden
4. Vorschau funktioniert
5. Speichern erzeugt Backup
6. Speichern schreibt die Datei korrekt zurueck
7. Git-Commit wird erzeugt
8. oeffentliche Seite zeigt danach die neue Version

## Mein empfohlener direkter naechster Schritt

Wenn wir mit diesem MVP-Design weitergehen, wuerde ich als naechstes direkt die erste Implementierungsphase vorbereiten:

1. konkrete Config-Struktur
2. exakte Endpunkt-Signaturen
3. HTML-Markierungslogik mit `data-editor-id`
4. Sanitization-Regeln als Go-Code-Plan

## Konkrete Config-Struktur

Fuer den MVP wuerde ich die Konfiguration in zwei Ebenen teilen:

- globale Laufzeitkonfiguration der App
- tenant-spezifische Domain-Konfigurationen

### Globale App-Konfiguration

Diese Werte gelten fuer den gesamten Dienst:

```env
STATIC_EDITOR_ADDR=:8090
STATIC_EDITOR_DATA_DIR=/data
STATIC_EDITOR_TENANT_DIR=/config/tenants
STATIC_EDITOR_SESSION_TTL=12h
STATIC_EDITOR_SECURE_COOKIES=true
```

Bedeutung:

- `STATIC_EDITOR_ADDR`: HTTP-Bind-Adresse
- `STATIC_EDITOR_DATA_DIR`: optionale lokale Daten fuer Sessions oder Logs
- `STATIC_EDITOR_TENANT_DIR`: Verzeichnis fuer tenant-spezifische Konfigurationsdateien
- `STATIC_EDITOR_SESSION_TTL`: Gueltigkeit einer Session
- `STATIC_EDITOR_SECURE_COOKIES`: ob Session-Cookies nur ueber HTTPS ausgegeben werden

### Tenant-Datei pro Domain

Pro bearbeitbarer Domain gibt es eine `.env`-Datei, zum Beispiel:

```env
STATIC_EDITOR_TENANT=example.org
STATIC_EDITOR_LOGIN_DOMAIN=bearbeitung.example.org
STATIC_EDITOR_ALIASES=www.example.org
STATIC_EDITOR_STATIC_ROOT=/srv/static/example.org
STATIC_EDITOR_BACKUP_ROOT=/srv/static-backups/example.org
STATIC_EDITOR_REPO_ROOT=/srv/static/example.org
STATIC_EDITOR_ALLOWED_EMAILS=andy@example.org,redaktion@example.org
STATIC_EDITOR_COOKIE_SECRET=replace-me
STATIC_EDITOR_MAIN_SELECTOR=main
STATIC_EDITOR_ALLOWED_BLOCK_TAGS=h1,h2,h3,h4,h5,p,ul,ol,li
STATIC_EDITOR_ALLOWED_INLINE_TAGS=strong,em,a,br
STATIC_EDITOR_START_PATH=/index.html
```

Bedeutung:

- `STATIC_EDITOR_TENANT`: kanonische Hauptdomain
- `STATIC_EDITOR_LOGIN_DOMAIN`: Bearbeitungsdomain
- `STATIC_EDITOR_ALIASES`: optionale Alias-Domains
- `STATIC_EDITOR_STATIC_ROOT`: Root der statischen Dateien
- `STATIC_EDITOR_BACKUP_ROOT`: Backup-Verzeichnis
- `STATIC_EDITOR_REPO_ROOT`: Git-Repo-Wurzel
- `STATIC_EDITOR_ALLOWED_EMAILS`: freigegebene E-Mail-Adressen fuer Magic-Links
- `STATIC_EDITOR_COOKIE_SECRET`: Signatur-Secret fuer Sessions
- `STATIC_EDITOR_MAIN_SELECTOR`: editierbarer Hauptbereich
- `STATIC_EDITOR_ALLOWED_BLOCK_TAGS`: erlaubte Block-Tags
- `STATIC_EDITOR_ALLOWED_INLINE_TAGS`: erlaubte Inline-Tags
- `STATIC_EDITOR_START_PATH`: erste editierbare Zielseite

### Mein Favorit fuer den MVP

Fuer den ersten Zuschnitt wuerde ich diese Defaults annehmen:

- `STATIC_EDITOR_MAIN_SELECTOR=main`
- `STATIC_EDITOR_ALLOWED_BLOCK_TAGS=h1,h2,h3,h4,h5,p,ul,ol,li`
- `STATIC_EDITOR_ALLOWED_INLINE_TAGS=strong,em,a,br`
- `STATIC_EDITOR_START_PATH=/index.html`

So bleibt die Konfiguration klein und sauber.
