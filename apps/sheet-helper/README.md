# Sheet Helper App

Kleine Go-Anwendung fuer mehrzweckartige Web-Helferlein auf Basis von Google-Sheet-Daten mit lokalem SQLite-Cache.

## Lokales Testen

Die App laeuft lokal mit Seed-Daten aus [testdata/seed.json](/Users/andygellermann/Documents/Projects/infra/infra/apps/sheet-helper/testdata/seed.json).

```bash
cd apps/sheet-helper
go run ./cmd/server
```

Danach im Browser testen:

- `http://localhost:8080/` -> Redirect
- `http://localhost:8080/api` -> Textseite
- `http://localhost:8080/andy` -> VCard-Seite
- `http://localhost:8080/agileebooks` -> passwortgeschuetzte Liste, Passphrase `scrum`

## Wichtige Umgebungsvariablen

```bash
SHEET_HELPER_ADDR=:8080
SHEET_HELPER_DB_PATH=./sheet-helper.db
SHEET_HELPER_SEED_FILE=./testdata/seed.json
SHEET_HELPER_COOKIE_SECRET=dev-only-change-me
SHEET_HELPER_TENANT=localhost
SHEET_HELPER_SYNC_TOKEN=replace-me
SHEET_HELPER_SHEET_ID=
SHEET_HELPER_PUBLISHED_URL=
SHEET_HELPER_ROUTES_SHEET=routes
SHEET_HELPER_VCARDS_SHEET=vcard_entries
SHEET_HELPER_TEXTS_SHEET=text_entries
SHEET_HELPER_LIST_PREFIX=list_
SHEET_HELPER_STARTUP_SYNC=false
```

## Architekturstand heute

- HTTP-Server mit Go-Standardbibliothek
- SQLite als lokaler Cache und Event-Speicher
- Seed-Sync fuer lokale Entwicklung
- Public-Google-Sheet-Sync ueber feste Blattnamen
- Modi: `link`, `text`, `vcard`, `list`
- einfacher Cookie-basierter Passwortschutz fuer unsensible Inhalte
- geschuetzter Sync-Endpunkt unter `/internal/sync/{tenant}`

## Erster echter Sheet-Test

Mit deinem echten Google Sheet laeuft der lokale Test jetzt ohne Seed-Datei.

Beispiel fuer `geller.men`:

```bash
cd apps/sheet-helper
SHEET_HELPER_ADDR=:8080 \
SHEET_HELPER_DB_PATH=./sheet-helper.db \
SHEET_HELPER_COOKIE_SECRET=dev-only-change-me \
SHEET_HELPER_TENANT=geller.men \
SHEET_HELPER_SYNC_TOKEN=replace-me \
SHEET_HELPER_SHEET_ID=1oqfAB0CtF2RT6zBlP8voWtLqfQeCTLscI4Q5lxmnUUQ \
SHEET_HELPER_PUBLISHED_URL='https://docs.google.com/spreadsheets/d/e/2PACX-1vSp5LbYzjAz1wgHFbfTUvflGc1YRarHgn_KpNEIf4uo5GEcDEdyGeZ32e52_btzCFzD1XuvFyXa5gw5/pubhtml' \
SHEET_HELPER_ROUTES_SHEET=routes \
SHEET_HELPER_VCARDS_SHEET=vcard_entries \
SHEET_HELPER_TEXTS_SHEET=text_entries \
SHEET_HELPER_LIST_PREFIX=list_ \
SHEET_HELPER_STARTUP_SYNC=true \
go run ./cmd/server
```

Danach lokal testen:

- `http://localhost:8080/api`
- `http://localhost:8080/andy`
- `http://localhost:8080/agileebooks`

Fuer lokales Routing gegen `geller.men` kannst du vorerst mit `curl` den `Host`-Header setzen:

```bash
curl -i -H 'Host: geller.men' http://127.0.0.1:8080/api
curl -i -H 'Host: geller.men' http://127.0.0.1:8080/andy
```

Der Sync-Endpunkt fuer das Apps Script sieht dann so aus:

```bash
curl -i \
  -X POST \
  -H 'X-Sheet-Helper-Token: replace-me' \
  http://127.0.0.1:8080/internal/sync/geller.men
```

Hinweis:

- Fuer den oeffentlichen Import ist aktuell `SHEET_HELPER_PUBLISHED_URL` der entscheidende Wert.
- Die App liest daraus automatisch die Blattnamen und zugehoerigen `gid`-Werte aus, nutzt nach aussen aber weiter deine Namenskonvention.

## Container

Es gibt ein erstes Container-Grundgeruest in [Dockerfile](/Users/andygellermann/Documents/Projects/infra/infra/apps/sheet-helper/Dockerfile).

Beispiel:

```bash
cd apps/sheet-helper
docker build -t sheet-helper:dev .
docker run --rm -p 8080:8080 sheet-helper:dev
```

## Hostvars-Vorlage

Eine erste Vorlage fuer spaetere Domain-spezifische Ansible-Hostvars liegt unter [sheet-helper-hostvars.j2](/Users/andygellermann/Documents/Projects/infra/infra/ansible/hostvars/templates/sheet-helper-hostvars.j2).

Die Idee dahinter:

- Infrastrukturwerte bleiben in Hostvars
- Inhalte bleiben in Google Sheets
- die App wird spaeter zentral betrieben und pro Domain konfiguriert

## Hostvars bequem anlegen

Mit [sheethelper-add.sh](/Users/andygellermann/Documents/Projects/infra/infra/scripts/sheethelper-add.sh) kannst du eine neue Hostvars-Datei mit sinnvollen Defaults erzeugen.

Beispiele:

```bash
./scripts/sheethelper-add.sh geller.men
./scripts/sheethelper-add.sh geller.men --sheet-id=abc123
./scripts/sheethelper-add.sh geller.men --sheet-id=abc123 --routes-sheet=routes --vcards-sheet=vcard_entries --texts-sheet=text_entries --list-prefix=list_
./scripts/sheethelper-add.sh team.example team.example.net --wildcard-domain=example
```

Das Skript erzeugt aktuell nur die Hostvars-Datei. Die eigentliche Ansible-Deploy-Rolle fuer den zentralen Sheet-Helper folgt als naechster Schritt.

## Google-Sheets-Trigger vorbereiten

Fuer spaetere Sync-Callbacks liegt eine Google-Apps-Script-Vorlage unter [sync-trigger.js](/Users/andygellermann/Documents/Projects/infra/infra/apps/sheet-helper/google-apps-script/sync-trigger.js).

Die zugehoerige Kurzanleitung findest du in [google-apps-script/README.md](/Users/andygellermann/Documents/Projects/infra/infra/apps/sheet-helper/google-apps-script/README.md).

## Naechste Schritte

- Google-Sheets-Import statt JSON-Seed
- Domain-spezifische Konfiguration via Hostvars oder ENV
- Aggregation und Ruecksync von Klicks
- Containerisierung und Ansible-Rolle
