# WordPress Erweiterung (Infra Stack WP-Extension / Ziel: v1.1.0)

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
- `scripts/wp-migrate-crowdsec.sh`: Ergänzt fehlende `wp_traefik_middleware_*` Defaults in Hostvars.
- `scripts/wp-redeploy.sh`: Validiert Hostvars + DNS und startet gezielten Redeploy.
- `scripts/wp-restore.sh`: Stellt DB und `/var/www/html` aus Backup wieder her, inkl. Versions- und Domain-Guard.
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
- optional `WP_HOME/WP_SITEURL` in `wp-config.php`

Verhalten:
- **Default:** interaktive Abfrage bei Domain-Mismatch, ob auf die neue Domain migriert werden soll (`yes`) oder die Backup-Domain verwendet wird (`NO`).
- bei erlaubter Migration: `siteurl`/`home` in `${wp_table_prefix}options` werden auf `https://<ziel-domain>` gesetzt

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

## Konsistenzprüfung der Methoden
- Einheitliches Naming: `wp-<action>.sh` analog `ghost-<action>.sh`.
- Einheitliches Datenmodell in Hostvars (`wp_domain_db`, `wp_domain_usr`, `wp_domain_pwd`, `wp_version`, `wp_traefik_middleware_*`).
- Einheitliche Sicherheitslogik: DNS-Check vor Deploy/Redeploy, CrowdSec-Middleware für Frontend/Admin/API, Versions-Guard + Domain-Guard im Restore.

## Container-Reuse / Wiederverwendung
- **Traefik**: vollständig wiederverwendbar; 301-Redirects sind via RedirectRegex-Middleware möglich und in den WP-Labels vorgesehen.
- **MySQL (`ghost-mysql`)**: sicher wiederverwendbar durch getrennte DBs/DB-User pro Instanz.
- **Hostvars**: erweiterbar um zusätzliche Flags (z. B. Redirect-Strategien, Middleware-Wahl, optionale Traefik-Labels).

## Caching-Empfehlung (leichtgewichtig + robust)
Empfohlene Reihenfolge:
1. WordPress-Page-Cache Plugin (Datei-Cache) als schneller Start.
2. Redis Object Cache für dynamische Last.
3. Optional CDN/Edge-Cache vor Traefik.

Traefik selbst bietet kein vollwertiges Response-Cache wie Nginx FastCGI-Cache.
