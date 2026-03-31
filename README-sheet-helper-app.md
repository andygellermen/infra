# Google-Sheet Helper App

Ziel ist eine zentrale, leichtgewichtige Web-Applikation in Go, die je nach Domain und Pfad unterschiedliche Helferlein ausliefert. Die Inhalte werden aus Google Sheets synchronisiert und lokal in einer Cache-DB gehalten.

Die App ist ausdruecklich fuer unsensible Daten gedacht. Sie ist kein Secret-Store und kein Ersatz fuer sichere Passwort- oder Schluesselverwaltung.

## Zielbild

Eine zentrale App verarbeitet mehrere Domains und mehrere Modi:

- `link`: Weiterleitung
- `text`: Text- oder Code-Anzeige
- `vcard`: digitale Visitenkarte mit `.vcf`-Download
- `list`: geschuetzte oder offene Link-/Download-Liste

Die aktive Konfiguration ergibt sich aus:

- `Host` der Anfrage
- URL-Pfad
- Domain-spezifischer Konfiguration aus Hostvars oder Umgebungsvariablen

## Warum dieses Modell

Das bisherige Google-Sheet-Prototyping ist fuer einen MVP gut geeignet, sollte aber nicht direkt bei jedem Request gegen Google laufen. Sinnvoller ist:

1. Google Sheet wird gelesen oder per Trigger synchronisiert.
2. Die Daten werden lokal in SQLite normalisiert gespeichert.
3. Requests werden ausschliesslich aus dem lokalen Cache bedient.
4. Klicks und Zugriffsdaten werden lokal gesammelt und periodisch aggregiert.

Vorteile:

- schnellere Antwortzeiten
- weniger Abhaengigkeit von Google
- weniger API-Limits
- einfacher Betrieb in einem zentralen Container
- gute Anschlussfaehigkeit an bestehende Ansible-Hostvars

## Empfohlenes Datenmodell

### 1. Haupttabelle `routes`

Diese Tabelle beschreibt pro Domain/Pfad genau einen Eintrag.

| Domain | Path | Type | Passphrase | Target | Title | Description | ListSheet | Enabled |
|--------|------|------|------------|--------|-------|-------------|-----------|---------|
| geller.men | / | link |  | https://andy.geller.men/ | Geller Start | Standard-Weiterleitung |  | true |
| geller.men | /api | text |  | G2Y5DKY-QJW%X9)4NZMbs8Og9FirzFC)QRdZvyUL | API-Key | Dein persoenlicher API-Schluessel |  | true |
| geller.men | /andy | vcard |  | Andy Gellermann, Agile & Change Management Consultant & Coach | Andy Gellermann | Dit bin ick! |  | true |
| geller.men | /tabelle | link |  | https://pad.tchncs.de/sheet/#/3/sheet/edit/e095063826d81dc248a7cc8de125530a/ | Tabelle |  |  | true |
| geller.men | /agileebooks | list | scrum | eBook Downloads | Agile eBooks | eBook-Sammlung "Agilisten 2022" | agileebooks | true |

#### Spaltenbedeutung

- `Domain`: zustaendige Domain ohne Schema
- `Path`: Pfad inklusive fuehrendem `/`
- `Type`: `link`, `text`, `vcard`, `list`
- `Passphrase`: optionaler Zugriffsschutz fuer `text` und `list`, bei Bedarf auch fuer `vcard`
- `Target`: je nach Typ das Ziel
- `Title`: Seitentitel oder Anzeigename
- `Description`: Untertitel, Hilfetext oder Beschreibung
- `ListSheet`: Name oder Schluessel eines zweiten Tabellenblatts fuer `list`
- `Enabled`: einfacher Aktiv-/Inaktiv-Schalter

Wichtige Empfehlung:

- `Path` statt vollstaendigem `Quell-Link` speichern
- `Domain` und `Path` getrennt halten

Das vereinfacht Routing, Validierung und spaetere Multi-Domain-Konfiguration erheblich.

### 2. Detailtabelle `vcard_entries`

Fuer `vcard` sollte nicht alles in der Haupttabelle leben. Besser ist eine eigene Detailtabelle oder ein eigenes Tabellenblatt.

| Domain | Path | FullName | Organization | JobTitle | Email | PhoneMobile | Address | Website | ImageURL |
|--------|------|----------|--------------|----------|-------|-------------|---------|---------|----------|
| geller.men | /andy | Andy Gellermann | Geller.men | Agile Coach | andy@gellermann.berlin | 00491732159150 | Peter-Hille-Str. 109A, 12587 Berlin, Germany | https://geller.men | https://www.gravatar.com/avatar/... |

Hinweise:

- Dateiformat fuer den Download: `.vcf`
- `Target` aus der Haupttabelle kann die Kurzbeschreibung oder der Default-Text fuer die Profilseite bleiben
- `ImageURL` darf z. B. Gravatar, eigenes CDN oder lokales Asset sein

### 3. Detailtabelle `list_<name>`

Fuer `list` sollte jedes Listenblatt einen klaren Aufbau haben. Beispiel fuer `list_agileebooks`:

| Sort | Label | URL | Description | Category | Password | Enabled |
|------|-------|-----|-------------|----------|----------|---------|
| 10 | Scrum Pocket Guide | https://example.org/file1.pdf | Kompakter Einstieg | Scrum |  | true |
| 20 | Kanban Basics | https://example.org/file2.pdf | PDF-Download | Kanban |  | true |

Hinweise:

- `ListSheet` aus `routes` verweist auf dieses Blatt
- optional kann pro Eintrag nochmals ein eigenes Passwort gepflegt werden
- `Sort` erlaubt stabile Reihenfolge

### 4. Detailtabelle `text_entries`

Wenn `text` mehr werden soll als nur ein einzelner Token, lohnt sich ebenfalls eine eigene Struktur:

| Domain | Path | ContentType | Content | CopyHint | ExpiresAt | Enabled |
|--------|------|-------------|---------|----------|-----------|---------|
| geller.men | /api | text/plain | G2Y5DKY-QJW%X9)4NZMbs8Og9FirzFC)QRdZvyUL | API-Key fuer Demo-Zwecke |  | true |

Hinweise:

- `ContentType` z. B. `text/plain`, `text/markdown`, `code`
- `ExpiresAt` optional fuer temporäre Anzeigen

## Mapping der vier Szenarien

### `link`

Verhalten:

- Lookup ueber `Domain + Path`
- `302` oder `301` Redirect
- Klick zaehlen

Pflichtfelder:

- `Type`
- `Target`

Optionale Erweiterungen:

- `StatusCode`
- `UTMAppend`
- `TrackClicks`

### `text`

Verhalten:

- HTML-Seite mit sauberer Darstellung
- optional Passwortabfrage vor Anzeige
- optional Copy-to-Clipboard

Pflichtfelder:

- `Type`
- `Target` oder Detailinhalt aus `text_entries`

Empfehlung:

- keine echten Secrets
- eher Demo-Tokens, temporäre Codes, Zugangshinweise, Freigabetexte

### `vcard`

Verhalten:

- schoene Profilseite
- `.vcf`-Download
- optional QR-Code spaeter

Pflichtfelder:

- `Type`
- Eintrag in `vcard_entries`

### `list`

Verhalten:

- Landingpage fuer Link- oder Download-Sammlung
- optional Passwortschutz
- Eintraege kommen aus zugeordnetem Listenblatt

Pflichtfelder:

- `Type`
- `ListSheet`

## Sync-Modell

Es gibt drei sinnvolle Sync-Wege.

### A. Polling

Die Go-App zieht alle X Minuten die Tabellen neu.

Vorteile:

- sehr einfach
- robust

Nachteile:

- Aenderungen sind nicht sofort sichtbar

### B. Manueller Trigger

Ein Google-Apps-Script oder ein externer Trigger ruft nach einer Aenderung einen Sync-Endpunkt der Go-App auf.

Beispiel:

- `POST /internal/sync/{tenant}`

Vorteile:

- schneller sichtbar
- trotzdem einfache Architektur

Nachteile:

- braucht kleinen Trigger-Mechanismus

### C. Hybrid

Trigger bei manuellen Aenderungen plus taeglicher Vollsync.

Das ist mein Favorit.

## Klick-Tracking

Klicks sollten lokal gespeichert werden, nicht direkt ins Sheet.

Empfohlenes lokales Schema:

| Timestamp | Domain | Path | Type | Target | Referrer | UserAgentHash |
|-----------|--------|------|------|--------|----------|---------------|

Ruecksync ins Sheet besser aggregiert:

| Domain | Path | TotalClicks | LastClickedAt | ClicksToday |
|--------|------|-------------|---------------|-------------|

So bleibt das Sheet uebersichtlich und die App bleibt performant.

## Passwortschutz

Passphrasen sollten in der App nicht im Klartext weiterverarbeitet werden als noetig.

Empfehlung:

- im Google Sheet fuer den MVP noch Klartext tolerierbar, wenn wirklich nur unsensibel
- in der lokal synchronisierten DB immer nur Hash speichern
- Zugriff per Session-Cookie nach erfolgreicher Eingabe
- Rate-Limiting pro IP/Route

Wenn ihr es sauberer wollt, kann spaeter bereits im Sheet ein vorberechneter Hash abgelegt werden.

## Domain-Konfiguration

Die Domain-spezifische Steuerung gehoert nicht ins Sheet allein, sondern in Hostvars oder ENV.

Beispiel:

```yaml
sheet_helper_domains:
  - domain: geller.men
    sheet_id: "abc123"
    public_sheet: true
    sync_mode: "hybrid"
    theme: "clean"
    allow_text: true
    allow_vcard: true
```

Damit bleibt Infrastruktur von Content getrennt.

## Empfehlung fuer den technischen Stack

Go:

- Router: `chi`
- Templates: Standardbibliothek `html/template`
- DB: `sqlite`
- Migrationen: einfache SQL-Migrationen
- CSS: Bootstrap oder Pico CSS
- optional etwas `htmx` oder `alpine.js`

Bewusst nicht noetig:

- kein schweres Vue-SPA fuer den MVP
- keine komplexe verteilte Architektur
- kein direkter Google-Schreibzugriff bei jedem Klick

## Konkrete Bewertung deines Prototyps

Dein bisheriges Modell ist gut als Startpunkt, aber ich wuerde drei Dinge aendern:

1. `Quell-Link` in `Domain` und `Path` aufteilen.
2. `vcard`-Felder in ein separates Detailblatt verschieben.
3. `list` nicht ueber Freitext im `Ziel/Text` aufloesen, sondern ueber ein klares `ListSheet`.

Beibehalten wuerde ich:

- ein gemeinsames Routing-Sheet als Zentrale
- einfache Typen pro Zeile
- optionale Passphrase
- pragmatischen Fokus auf unsensible Inhalte

## Nächste sinnvolle Schritte

1. Das Routing-Sheet finalisieren.
2. Die drei Detailblaetter `vcard_entries`, `text_entries` und mindestens ein `list_*`-Beispiel definieren.
3. Die Konfiguration fuer Domains in Ansible-Hostvars beschreiben.
4. Danach die Go-App mit lokalem SQLite-Cache als MVP bauen.

## MVP-Scope

Ein guter erster MVP waere:

- `link`
- `text`
- `vcard`
- lokaler Cache
- manueller oder periodischer Sync
- Klickzaehlung lokal

`list` wuerde ich direkt mitplanen, aber nur dann sofort bauen, wenn ihr sie wirklich als ersten Use Case braucht.

## Kopierbare Google-Sheets-Vorlagen

Die folgenden Tabellen sind so formuliert, dass du sie direkt als Ausgangspunkt fuer Google Sheets verwenden kannst. Fuer den Start reichen vier Blaetter:

- `routes`
- `vcard_entries`
- `text_entries`
- `list_agileebooks`

### Blatt `routes`

Das ist das zentrale Routing-Blatt. Jede Zeile beschreibt genau einen oeffentlichen Pfad.

| Domain | Path | Type | Passphrase | Target | Title | Description | ListSheet | Enabled |
|--------|------|------|------------|--------|-------|-------------|-----------|---------|
| geller.men | / | link |  | https://andy.geller.men/ | Geller Start | Standard-Weiterleitung |  | true |
| geller.men | /api | text |  | api-demo-key-001 | API-Key | Dein persoenlicher API-Schluessel |  | true |
| geller.men | /andy | vcard |  | Andy Gellermann, Agile & Change Management Consultant & Coach | Andy Gellermann | Dit bin ick! |  | true |
| geller.men | /tabelle | link |  | https://pad.tchncs.de/sheet/#/3/sheet/edit/e095063826d81dc248a7cc8de125530a/ | Tabelle | Gemeinsame Online-Tabelle |  | true |
| geller.men | /agileebooks | list | scrum | eBook Downloads | Agile eBooks | eBook-Sammlung "Agilisten 2022" | list_agileebooks | true |

Hinweise:

- `Domain + Path` muss eindeutig sein.
- `Path` sollte immer mit `/` beginnen.
- Fuer `link` ist `Target` die Ziel-URL.
- Fuer `text` ist `Target` ein kurzer Inhalt oder Platzhalter, wenn der eigentliche Inhalt aus `text_entries` kommt.
- Fuer `vcard` ist `Target` ein kurzer Intro-Text fuer die Profilseite.
- Fuer `list` ist `ListSheet` verpflichtend.

### Blatt `vcard_entries`

Dieses Blatt enthaelt die eigentlichen Kontaktdaten fuer den Typ `vcard`.

| Domain | Path | FullName | Organization | JobTitle | Email | PhoneMobile | Address | Website | ImageURL | Note | Enabled |
|--------|------|----------|--------------|----------|-------|-------------|---------|---------|----------|------|---------|
| geller.men | /andy | Andy Gellermann | Geller.men | Agile Coach | andy@gellermann.berlin | 00491732159150 | Peter-Hille-Str. 109A, 12587 Berlin, Germany | https://geller.men | https://www.gravatar.com/avatar/example | Agile & Change Management Consultant & Coach | true |

Hinweise:

- `Domain + Path` muss zu einem `vcard`-Eintrag in `routes` passen.
- `ImageURL` ist optional.
- `Note` eignet sich fuer Untertitel, Freitext oder Kurzbeschreibung auf der HTML-Seite.

### Blatt `text_entries`

Dieses Blatt lohnt sich, sobald `text` mehr als ein simpler Einzeiler ist.

| Domain | Path | ContentType | Content | CopyHint | ExpiresAt | Enabled |
|--------|------|-------------|---------|----------|-----------|---------|
| geller.men | /api | text/plain | G2Y5DKY-QJW%X9)4NZMbs8Og9FirzFC)QRdZvyUL | API-Key fuer Demo-Zwecke |  | true |
| geller.men | /zugang | text/markdown | Bitte nur fuer den vereinbarten Zweck verwenden. Zugang gueltig bis Projektende. | Text markieren und kopieren |  | true |

Hinweise:

- `ContentType` kann fuer die Darstellung genutzt werden, etwa `text/plain`, `text/markdown` oder `code`.
- `ExpiresAt` ist optional und kann spaeter fuer zeitgesteuerte Inhalte dienen.
- Fuer einen MVP kann `text` auch ausschliesslich aus `routes.Target` kommen.

### Blatt `list_agileebooks`

Dieses Blatt ist ein Beispiel fuer einen Listenmodus mit Downloads oder Links.

| Sort | Label | URL | Description | Category | Password | Enabled |
|------|-------|-----|-------------|----------|----------|---------|
| 10 | Scrum Pocket Guide | https://example.org/downloads/scrum-pocket-guide.pdf | Kompakter Einstieg in Scrum | Scrum |  | true |
| 20 | Kanban Basics | https://example.org/downloads/kanban-basics.pdf | Uebersicht fuer Einsteiger | Kanban |  | true |
| 30 | Agile Retrospektiven | https://example.org/downloads/retrospektiven.pdf | Sammlung mit Moderationsideen | Facilitation |  | true |

Hinweise:

- Der Blattname muss zu `ListSheet` aus `routes` passen.
- `Sort` regelt die Anzeige-Reihenfolge.
- `Password` ist optional, wenn einzelne Links in einer Liste separat geschuetzt werden sollen.
- Fuer den MVP wuerde ich meist nur die ganze Liste schuetzen, nicht einzelne Zeilen.

## Minimale Startvariante

Wenn du moeglichst schnell loslegen willst, reicht anfangs sogar dieses reduzierte Modell:

- `routes`
- `vcard_entries`
- ein einziges `list_*`-Blatt nach Bedarf

`text_entries` kann zunaechst entfallen, solange kurze Inhalte direkt in `routes.Target` stehen.

## Review nach dem ersten Zuschnitt

Nach dem Aufbereiten deiner Prototyp-Tabelle wuerde ich gemeinsam mit dir noch auf diese Punkte schauen:

1. Soll `text` wirklich frei im Sheet stehen oder lieber spaeter mit Ablaufdatum und optionalem Passwort abgesichert werden?
2. Soll `list` nur Downloads und Links zeigen oder auch kleine Metadaten wie Kategorie, Tags und Dateigroesse?
3. Soll `vcard` nur einzelne Personen abbilden oder spaeter auch Team-/Firma-Seiten koennen?
4. Soll der Sync nur manuell per Trigger laufen oder zunaechst einfach alle 15 Minuten pollen?

## Meine aktuelle Empfehlung

Fuer einen sauberen MVP wuerde ich heute exakt so starten:

- Google Sheet als Content-Quelle
- Go-App als zentrale Laufzeit
- SQLite als lokaler Cache
- `routes` als Hauptblatt
- `vcard_entries` und `list_*` direkt aktiv
- `text_entries` vorbereitet, aber nur bei Bedarf genutzt
- manueller Sync-Trigger plus taeglicher Vollsync
- lokales Click-Tracking nur fuer `link`

Damit bleibt das Ganze klein, klar und betrieblich angenehm, ohne deine spaeteren Optionen zu verbauen.
