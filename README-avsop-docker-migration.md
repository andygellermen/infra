# AVSoP Docker-Migrationsplan

## Ziel

Diese Notiz beschreibt die empfohlene erste Rettungs- und Migrationsstufe fuer die Legacy-AVSoP in eine nicht oeffentliche Docker-Umgebung.

Ziel ist **nicht** sofortige Modernisierung, sondern eine **moeglichst originalgetreue, kontrollierbare Ersatzinstanz** fuer:

- Ausfalluebungen
- Smoke-Tests fuer die spaetere `Mietmachine`
- kontrollierte Legacy-Weiterfuehrung bis zur Ablösung


## Aktuelles Legacy-Profil

- Host-OS: `Ubuntu 18.04.5`
- Webserver: `Apache 2.4.29`
- PHP: `7.2.24` via `mod_php`
- Datenbank: `MySQL 5.7.33`
- Webroot: `/var/www/avs`
- Codegroesse: ca. `538M`
- relevante Datenbanken:
  - `autovermietung_miethaenger` (~`2.8G`)
  - `autovermietung_anhaenger-berlin` (~`127M`)
  - `autovermietung_miethaenger-test` (~`2.4G`, historisches Artefakt)
  - `autovermietung_anhaenger-berlin-test` (~`125M`, historisches Artefakt)
  - `autovermietung_miethaenger_experimental` (~`60M`)

Wesentliche Laufzeitmerkmale aus dem Legacy-Code:

- Datenbankauswahl erfolgt ueber `$_SERVER["HTTP_HOST"]`
- Stage-Erkennung erfolgt ueber `$_SERVER["SERVER_ADDR"]`
- HTTPS wird app-intern erwartet
- Frontend/Backend teilen sich dieselbe Codebasis
- Frontend-Verhalten wird u. a. ueber `.htaccess_frontend` gesteuert


## Getroffene Entscheidungen

- `/var/www/admin` gehoert **nicht** in die erste AVSoP-Rettungsmigration.
- Die bisherigen `*-test`-Datenbanken gelten als Altlast eines frueheren Staging-Versuchs und sind **nicht** die bevorzugte Basis.
- DMS-Synchronisation wird in Phase 1 **nur manuell** in Richtung Docker-Stage ausgefuehrt.
- Die Ersatzinstanz bleibt **nonpublic** und ist nur ueber Proxy/VPN/IP-Whitelist erreichbar.
- TLS wird **nicht** im AVSoP-Container selbst terminiert, sondern vor dem Stack.


## Empfohlene Phase-1-Architektur

### Host

- frischer Docker-Host, bevorzugt `Debian 12`
- keine produktiven WordPress-/Ghost-/Custom-App-Lasten auf demselben Stage-Stack

### Container

- `avsop-web`
  - `Apache + mod_php`
  - Legacy-PHP moeglichst nah an `7.2`
  - bind-mount oder Image mit `/var/www/avs`
- `avsop-db`
  - `mysql:5.7`
  - isoliert, ohne externes Port-Publishing
- `avsop-cron`
  - gleiche Codebasis wie `avsop-web`
  - Cronjobs anfangs **deaktiviert** bzw. nur manuell
- `avsop-mailpit` (optional, empfohlen)
  - faengt Mails in Phase 1 lokal ab
  - verhindert versehentlichen Versand echter Legacy-Mails


## Datenbankstrategie fuer Phase 1

Die Ersatzinstanz soll zunaechst **dieselben Schemanamen** wie die produktive AVSoP nutzen, jedoch auf einem **isolierten Stage-MySQL**:

- `autovermietung_miethaenger`
- `autovermietung_anhaenger-berlin`

Begruendung:

- reduziert App-Patches in `config/data.inc`
- entspricht dem tatsaechlich genutzten Zustand besser als die alten `*-test`-Artefakte
- erleichtert den spaeteren Notfallbetrieb als echte Ersatzinstanz

Die `*-test`- und `*_experimental`-Schemas koennen spaeter optional zusaetzlich uebernommen werden, sind aber **nicht** der Startpunkt.


## Hostnamen fuer die Stage

Da die Datenbankauswahl in `config/data.inc` per `strpos()` auf dem Hostnamen basiert, sollten Stage-Hostnamen bewusst so gewaehlt werden, dass sie die bestehende Logik treffen, ohne neue Sonderfaelle einzufuehren.

Geeignet sind zum Beispiel:

- `miethaenger-stage.internal`
- `anhaenger-berlin-stage.internal`
- `backend-miethaenger-stage.internal`
- `backend-anhaenger-berlin-stage.internal`

Wichtig:

- Hostnamen mit `miethaenger-test` oder `anhaenger-berlin-test` nur verwenden, wenn bewusst die historischen `*-test`-Schemas getestet werden sollen.


## Zwingende Sicherheitsleitplanken

### Reverse Proxy / Zugang

- keine oeffentliche Freischaltung
- Zugriff nur ueber:
  - VPN
  - feste Admin-IP-Whitelist
  - oder Proxy mit Basic Auth + IP-Restriktion

### Im Webstack blockieren

Zusatzschutz vor dem Legacy-Code:

- `/.git`
- `/.idea`
- `/config`
- `/queries`
- `/replication`
- Dateien mit:
  - `*.sql`
  - `*.inc`
  - `*.sh`

### HTTPS hinter Proxy

Die App erwartet `$_SERVER["HTTPS"] == "on"`.

Der vorgeschaltete Proxy muss deshalb:

- TLS terminieren
- `X-Forwarded-Proto: https` setzen

Und der AVSoP-Webcontainer muss diesen Header in ein fuer PHP sichtbares `HTTPS=on` ueberfuehren.


## Cron- und Nebenprozesse

Cronjobs werden in Phase 1 **nicht automatisch** aktiviert.

Erst nach erfolgreichem Basis-Smoke-Test werden sie einzeln freigeschaltet, z. B.:

- reine Leselogs
- Queue-/Reportjobs
- PDF-/Mailjobs

Vorlaeufig ausgeschlossen:

- automatische DMS-Synchronisation
- produktionsnahe Replikationsprozesse
- Mails an echte Empfaenger ohne vorherige Testumleitung


## DMS-Strategie fuer Phase 1

- DMS-Daten werden nur **manuell** zur Docker-Stage synchronisiert
- bevorzugt einmaliger oder bewusst angestossener `rsync`
- keine automatische Richtungsentscheidung per Host-IP wie im Legacy-Skript

Ziel:

- saubere, nachvollziehbare Ersatzinstanz
- kein versehentliches Rueckschreiben in andere Umgebungen


## Minimaler Code-Patch fuer Containerbetrieb

Fuer eine robuste Docker-Stage ist ein kleiner, rueckwaertskompatibler Patch empfehlenswert:

- `AVS_STAGE`
- `AVS_DB_HOST`
- `AVS_DB_USER`
- `AVS_DB_PASSWORD`
- optional `AVS_FORCE_HTTPS=1`

Funktion:

- wenn Umgebungsvariablen gesetzt sind, nutzen
- sonst unveraendert auf die bestehende Legacy-Logik zurueckfallen

Damit bleibt der Altserver lauffaehig, waehrend die Docker-Stage kontrollierbar wird.


## Empfohlene Migrationsreihenfolge

1. frischen Stage-Docker-Host bereitstellen
2. `mysql:5.7`-Container mit isoliertem Datenvolume aufsetzen
3. produktive AVSoP-Schemas als Stage-Kopie importieren
4. `Apache + mod_php`-Legacy-Container erstellen
5. nur internen Proxy-Zugang aktivieren
6. Smoke-Tests ohne Cronjobs fahren
7. DMS manuell synchronisieren
8. weitere Smoke-Tests mit echten Dokument-/Upload-Pfaden
9. erst danach einzelne Cronjobs kontrolliert freischalten


## Smoke-Test-Mindestumfang

### Backend

- Login erreichbar
- Session bleibt stabil
- Navigation/Kernmasken laden
- PDF-Vorschau oder PDF-Export funktioniert
- Upload in das DMS-/Dateiverzeichnis funktioniert

### Frontend

- Startseite laedt
- `.html`-Rewrite auf `frontend.php` funktioniert
- Buchungscode-/Pay-Routen brechen nicht sofort

### Datenbank

- Verbindung steht sauber
- beide Hauptmandanten liefern Daten
- keine versehentliche Verbindung zu externer Live-DB

### Mail

- Mails landen waehrend Phase 1 in einer Testsenke wie `Mailpit`


## Noch offene Folgearbeit

- Dockerfile/Compose fuer `Apache + mod_php` konkret ausarbeiten
- Legacy-Cronjobs klassifizieren
- minimalen Env-Patch in `config/data.inc` / `config/starter.php` vorbereiten
- Proxy-Regeln fuer die nonpublic Stage festlegen
- spaeter Test-Stage zur Security-Stage/Wazuh-Heimat umwidmen, sobald die AVSoP-Rettungsinstanz belastbar ist
