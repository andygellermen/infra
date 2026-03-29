# WordPress Erweiterung (Infra Stack WP-Extension / Ziel: v1.5.0)

## Architektur-Entscheidung: „zentraler Kern“ richtig verstanden
Wir fahren **einen zentral gehärteten Betriebsstandard**, aber **nicht einen einzigen WordPress-Container für alle Domains**.

Stattdessen:
- pro Domain eine eigene WordPress-Instanz (`wp-<domain>` Container + eigenes Volume + eigene DB/User)
- alle Instanzen nutzen denselben Infra-Unterbau (Traefik, CrowdSec, MySQL-Server)
- Versionierung pro Instanz über `wp_version` in den Hostvars (optional `wp_image_tag` für PHP-Variante)

Damit kombinieren wir Isolation (pro Site) mit zentraler Härtung/Wartbarkeit (gemeinsame Infrastruktur).

## Neue Skripte
- `scripts/wp-add.sh`: Legt Hostvars an, prüft DNS-A-Records und deployt WordPress via Ansible.
- `scripts/wp-backup.sh`: Erstellt WordPress-Backup (DB-Dump + Volume + Hostvars + `wp_version` im Manifest).
- `scripts/wp-delete.sh`: Entfernt WordPress-Instanz (Container/DB/User/Volume + Hostvars).
- `scripts/wp-fix-perms.sh`: Korrigiert Dateirechte im bestehenden WP-Volume ohne vollständigen Restore (nützlich bei `.htaccess`-Forbidden).
- `scripts/wp-migrate-crowdsec.sh`: Ergänzt fehlende `wp_traefik_middleware_*` Defaults in Hostvars.
- `scripts/wp-redeploy.sh`: Validiert Hostvars + DNS und startet gezielten Redeploy.
- `scripts/wp-restore.sh`: Stellt DB und `/var/www/html` aus Backup wieder her, inkl. Versions-, Domain- und `wp-config.php`-Guard.
- `scripts/wp-upgrade.sh`: Setzt `wp_version` in Hostvars und führt Redeploy aus.

## Backup-Format für WordPress
Erwartetes Backup-Format (neu):
- Das Archiv enthält den **vollständigen WordPress-Document-Root** direkt auf Top-Level.
- Die DB-SQL-Datei liegt ebenfalls direkt im Document-Root (z. B. `db.sql`).
- `manifest` und/oder `hostvars` sind **optional** (z. B. unter `_infra/`), damit auch Legacy-WordPress-Backups ohne Infra-Metadaten nutzbar bleiben.

> Empfehlung: Der Document-Root **muss vollständig** enthalten sein, damit Plugins/Themes/Uploads konsistent restauriert werden können.

## Restore bei Domain-Migration (neu)
`wp-restore.sh` prüft die Domain-Konsistenz über:
- falls vorhanden: `domain` aus optionaler hostvars-Datei im Backup (`_infra/hostvars.yml` oder Legacy `files/hostvars.yml`)
- ohne hostvars: Fallback-Domain-Ermittlung aus der ausgewählten SQL-Datei (siteurl/home)
- Tabellenprefix wird aus der ausgewählten SQL-Datei erkannt und als `wp_table_prefix` in Hostvars übernommen
- fehlen lokale Hostvars komplett, erzeugt `wp-restore.sh` minimale Hostvars automatisch und initialisiert DB/User aus `ansible/secrets/secrets.yml`
- `WP_HOME`/`WP_SITEURL` werden auf die bestätigte Ziel-Domain gesetzt
- `wp-config.php` wird auf Restore-Sollwerte gehärtet (`DB_*`, `WP_HOME`, `WP_SITEURL`, `$table_prefix`, Proxy-/HTTPS-Block)

Verhalten:
- **Default:** interaktive Abfrage bei Domain-Mismatch, ob auf die neue Domain migriert werden soll (`yes`) oder die Backup-Domain verwendet wird (`NO`).
- bei erlaubter Migration: `siteurl`/`home` in `${wp_table_prefix}options` werden auf `https://<ziel-domain>` gesetzt
- bei jedem Restore werden `WP_HOME` und `WP_SITEURL` zusätzlich in `wp-config.php` auf `https://<ziel-domain>` gesetzt
- existiert der Ziel-Container noch nicht, schreibt der Restore zunächst DB/Volume und führt danach automatisiert einen Redeploy aus

## Restore-Härtung in `wp-config.php` (neu in v1.2.0)
`wp-restore.sh` überschreibt mitgebrachte Legacy-Werte in `wp-config.php` bewusst mit den zur Ziel-Instanz passenden Werten. Das schützt vor klassischen Restore-Fehlern wie:

- alte DB-Credentials aus dem Backup
- `localhost:3306` statt `infra-mysql`
- fehlende oder falsche `WP_HOME`/`WP_SITEURL`
- falscher `$table_prefix`
- fehlende HTTPS-Erkennung hinter Traefik

Beim Restore werden daher aktiv gesetzt:

- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_HOST=infra-mysql`
- `WP_HOME=https://<ziel-domain>`
- `WP_SITEURL=https://<ziel-domain>`
- `$table_prefix`
- Proxy-/HTTPS-Block für `HTTP_X_FORWARDED_PROTO`

Zusätzlich prüft `wp-restore.sh` nach dem Schreiben explizit, ob diese Werte wirklich in `wp-config.php` angekommen sind. Wenn nicht, bricht der Restore mit einer klaren Fehlermeldung ab, statt einen still fehlerhaften Zustand zu hinterlassen.

### Patch-Hinweis v1.2.2
In `v1.2.2` wurde die Einfügelogik für `wp-config.php` weiter gehärtet:

- unterstützt jetzt auch ältere `wp-config.php`-Varianten mit Kommentar `Happy blogging.` statt `Happy publishing.`
- fällt notfalls auf Einfügen direkt vor `require_once(ABSPATH . 'wp-settings.php');` zurück
- verhindert damit, dass `WP_HOME` und `WP_SITEURL` zu spät in der Datei landen und erst nach dem WordPress-Bootstrap wirksam würden

Dieser Fall war relevant, weil verspätet gesetzte URL-Konstanten WordPress-Redirect-Schleifen trotz korrekt restaurierter DB und Domain verursachen können.

### Patch-Hinweis v1.2.3
In `v1.2.3` wurde der Proxy-/HTTPS-Block in `wp-config.php` nochmals robuster gemacht:

- Literal-Einfügung ohne fehleranfällige Shell-Interpolation von `$_SERVER[...]`
- verhindert kaputte Zeilen wie `if (isset() && === 'https')`
- schützt damit vor PHP-Parse-Errors in `wp-config.php`

### Patch-Hinweis v1.2.4
In `v1.2.4` führt `wp-restore.sh` nach jedem erfolgreichen Restore einen automatisierten Selbsttest aus:

- Syntaxcheck der laufenden `wp-config.php` via `php -l`
- interner HTTP-Check direkt im Container mit `Host: <domain>` und `X-Forwarded-Proto: https`
- explizite Erkennung einer WordPress-Canonical-Redirect-Schleife auf dieselbe Ziel-URL
- zusätzlicher öffentlicher HTTPS-Check gegen `https://<domain>/`

Dadurch werden die zuletzt aufgetretenen Fehlerbilder deutlich früher erkannt:

- kaputte `wp-config.php`
- fehlerhafte Proxy-/HTTPS-Erkennung hinter Traefik
- Redirect-Loops direkt nach dem Restore

### Patch-Hinweis v1.2.5
In `v1.2.5` wurde die Header-Auswertung im Post-Restore-Selbsttest korrigiert:

- robustere `awk`-Auswertung des `Location`-Headers
- behebt einen Quoting-Fehler, der den Selbsttest trotz erfolgreichem Restore abbrechen konnte

### Patch-Hinweis v1.2.6
In `v1.2.6` wurde die Ignore-Liste für lokale/remote Arbeitskopien ergänzt:

- Backup-Artefakte wie `ansible/hostvars/*.yml.bak*`
- typische Editor-Reste wie `*~`
- zusätzliche temporäre Dateien wie `*.tmp` und `*.bak`

Damit sinkt das Risiko, versehentlich lokale Restore- oder Hostvars-Sicherungen ins Repository zu committen.

### Patch-Hinweis v1.5.0
In `v1.5.0` wurde die Domain-Erkennung beim WordPress-Restore präzisiert:

- bevorzugte Ermittlung der Backup-Domain direkt aus den `siteurl`- und `home`-Einträgen des `${wp_table_prefix}options`-Dumps
- keine beliebige URL-Erkennung mehr aus anderen Tabellenbereichen

Zusätzlich setzt `wp-restore.sh` bei einer Domain-Migration in `${wp_table_prefix}options` jetzt bewusst:

- `option_value='https://<ziel-domain>'`
- `autoload='yes'`

für die Optionen `siteurl` und `home`.

Hinweis:
- `autoload='yes'` ist der normale und saubere WordPress-Standard für diese Kernoptionen
- ein falscher `autoload`-Wert ist nicht zwangsläufig die Hauptursache eines Redirect-Loops
- er kann aber zusammen mit Cache-/Plugin-Effekten zu inkonsistentem Laufzeitverhalten beitragen und wird deshalb beim Restore auf den Sollzustand zurückgeführt

## Post-Restore-Test-Szenario (neu in v1.2.4)
Nach dem Restore wird die Zielinstanz automatisiert in dieser Reihenfolge geprüft:

1. `wp-config.php` ist im laufenden Container syntaktisch gültig.
2. WordPress beantwortet einen internen Request im Proxy-Kontext (`Host` + `X-Forwarded-Proto`) mit einem sinnvollen Status.
3. Eine Selbst-Umleitung auf exakt dieselbe HTTPS-URL wird als Fehler gewertet.
4. Ein zusätzlicher öffentlicher HTTPS-Check prüft die final erreichbare Ziel-URL.

Der Restore bricht bei den kritischen Prüfungen bewusst ab, damit kein scheinbar „erfolgreicher“ Restore mit defekter Laufzeit-Konfiguration stehen bleibt.

## Restore + Versionen ohne Datenverlust (wichtig bei mehreren WP-Versionen)
`wp-restore.sh` prüft Quell- (`Backup`) und Zielversion (`hostvars`).

Zusätzlich:
- Backup-Archive können als `.tar.gz`, `.tgz` oder `.zip` wiederhergestellt werden.
- Bei mehreren SQL-Dateien im Archiv erfolgt eine interaktive Auswahl in der Konsole (`1...n`, Standard `1`). Dabei wird zuerst im Document-Root gesucht.

- **Default:** Schutz vor unbeabsichtigtem Downgrade (Abbruch, wenn Zielversion kleiner als Backup-Version).
- `--restore-hostvars`: übernimmt Hostvars aus Backup (inkl. damaliger `wp_version`).
- `--allow-version-downgrade`: explizite Freigabe eines Downgrades (nur bewusst einsetzen).
- `--php-version=<major.minor>`: setzt `wp_image_tag` im Hostvars und triggert optionalen Redeploy.

Praxis-Empfehlung:
1. Restore auf gleiche/höhere `wp_version`.
2. Danach kontrollierter `wp-upgrade.sh`/`wp-redeploy.sh` auf die gewünschte Zielversion.

## Alte/fragile WP (z. B. PHP 7.4) als statisches HTML konservieren
Wenn JS-MP3-Player erhalten bleiben muss, reicht ein „nur HTML“-Dump häufig nicht. Sinnvoller Ablauf:
1. Legacy-WP in isoliertem Container mit passender PHP-Version starten.
2. Headless-Crawl mit JS-Ausführung verwenden (z. B. Playwright-Renderer statt reinem `wget`).
3. Assets absolut mitnehmen (`/wp-content/uploads`, JS-Bundles, Audio-Dateien, API-Endpunkte).
4. Ergebnis als statische Site hinter Traefik/CDN ausliefern.

## Rewrite / Redirect (Apache zu Traefik)
- `.htaccess`-Rewrite-Regeln für Pretty Permalinks bleiben in WordPress relevant.
- Domain-Redirects (inkl. 301) gehören in Traefik-Router/Middlewares, nicht in WordPress.
- In dieser Erweiterung werden Alias-Domains per Redirect-Middleware dauerhaft (`permanent=true`) auf die Primärdomain geleitet.

## Canonical Redirect in WordPress
Der WordPress-Canonical-Redirect ist grundsätzlich sinnvoll:

- normalisiert URLs (`http`/`https`, Slash-Varianten, Duplicate URLs)
- hilft bei SEO und konsistenten internen Links

Im Restore-Kontext gilt aber:

- **Default-Empfehlung:** aktiviert lassen
- **nicht** pauschal per MU-Plugin deaktivieren
- nur temporär als Diagnose- oder Notfallmaßnahme einsetzen, wenn eine Redirect-Schleife isoliert werden muss

Die eigentliche Zielsetzung der Restore-Härtung ist daher nicht, `redirect_canonical` abzuschalten, sondern WordPress nach dem Restore so konsistent zu konfigurieren, dass der Canonical-Redirect wieder korrekt arbeiten kann.

## Konsistenzprüfung der Methoden
- Einheitliches Naming: `wp-<action>.sh` analog `ghost-<action>.sh`.
- Einheitliches Datenmodell in Hostvars (`wp_domain_db`, `wp_domain_usr`, `wp_domain_pwd`, `wp_version`, `wp_traefik_middleware_*`).
- Einheitliche Sicherheitslogik: DNS-Check vor Deploy/Redeploy, CrowdSec-Middleware für Frontend/Admin/API, Versions-Guard + Domain-Guard im Restore sowie Verifikation der kritischen `wp-config.php`-Werte nach dem Restore.

## Patch-Hinweis v1.6.5
In `v1.6.5` wurde der Redeploy-/Restore-Ablauf der WordPress-Skripte beim Bedienkomfort und bei der Verifikation nachgeschärft:

- `wp-redeploy.sh` fasst erfolgreiche Ansible-Läufe jetzt kompakter zusammen
- nach einem Redeploy wird direkt ein WordPress-Selbsttest ausgeführt:
  `wp-config.php`-Syntax, interner Proxy-/HTTP-Check, Schleifen-Erkennung und öffentlicher HTTPS-Check
- `wp-restore.sh` profitiert bei gezielten Folge-Deployments automatisch vom kompakteren Redeploy-Verhalten

## Passwort-Schutz für WordPress-Bereiche (neu in v1.7.0)
`wp-redeploy.sh` verwaltet optionalen Passwort-Schutz für typische WordPress-Bereiche jetzt interaktiv direkt über Traefik-Basic-Auth-Middlewares.

Unterstützte Bereiche:
- `frontend`: gesamte Website `/`
- `admin`: `wp-admin` und `wp-login.php`
- `api`: `wp-json`

Ablauf:
- vorhandene Schutz-Einträge werden erkannt und können geändert oder aufgehoben werden
- neue Schutz-Einträge können interaktiv ergänzt werden
- bei Auswahl von `frontend` erscheint eine ausdrückliche Warnung, weil damit die gesamte Website geschützt wird
- Benutzername und Passwort-Hash werden in `wp_basic_auth_scopes` in den Hostvars hinterlegt
- nach dem Redeploy prüft ein Selbsttest die aktivierten Bereiche automatisch

Empfohlene Hostvars-Struktur:

```yaml
wp_basic_auth_scopes:
  - scope: "admin"
    realm: "Protected Admin"
    username: "andy"
    password_hash: ""
```

Praktischer Hinweis:
- ein leerer `password_hash` ist als Platzhalter erlaubt
- beim nächsten `wp-redeploy.sh <domain>` oder `wp-restore.sh <domain> <backup>` wird das Passwort dann interaktiv abgefragt und als Hash gespeichert
- `wp-redeploy.sh --check-only` bleibt bewusst ohne interaktive Passwortverwaltung

## Patch-Hinweis v1.7.1
In `v1.7.1` wurde der öffentliche Restore-Selbsttest für WordPress an den neuen Frontend-Passwortschutz angepasst:

- ein aktiver Frontend-Schutz darf den öffentlichen HTTPS-Check jetzt sauber mit `401` beantworten
- `wp-restore.sh` bewertet diesen Fall deshalb nicht mehr als Fehlverdacht, sondern als bestätigten Passwortschutz

## Versionspflege
- Aktueller Stand dieser WordPress-Erweiterung: `v1.5.0`
- Praxisregel: Nach jedem erfolgreichen, produktiv relevanten Patch die Stack-Version bewusst erhöhen, damit Restore-/Betriebszustände leichter identifizierbar bleiben.
- Empfohlenes Vorgehen:
  - Patch fertigstellen
  - `VERSION` anheben
  - relevante README kurz nachziehen

## Container-Reuse / Wiederverwendung
- **Traefik**: vollständig wiederverwendbar; 301-Redirects sind via RedirectRegex-Middleware möglich und in den WP-Labels vorgesehen.
- **MySQL (`infra-mysql`)**: sicher wiederverwendbar durch getrennte DBs/DB-User pro Instanz.
- **Hostvars**: erweiterbar um zusätzliche Flags (z. B. Redirect-Strategien, Middleware-Wahl, optionale Traefik-Labels).

## Caching-Empfehlung (leichtgewichtig + robust)
Empfohlene Reihenfolge:
1. WordPress-Page-Cache Plugin (Datei-Cache) als schneller Start.
2. Redis Object Cache für dynamische Last.
3. Optional CDN/Edge-Cache vor Traefik.

Traefik selbst bietet kein vollwertiges Response-Cache wie Nginx FastCGI-Cache.
