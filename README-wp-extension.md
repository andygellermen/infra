# WordPress Erweiterung (Infra Stack WP-Extension / Ziel: v1.1.0)

## Architektur-Entscheidung: „zentraler Kern“ richtig verstanden
Wir fahren **einen zentral gehärteten Betriebsstandard**, aber **nicht einen einzigen WordPress-Container für alle Domains**.

Stattdessen:
- pro Domain eine eigene WordPress-Instanz (`wp-<domain>` Container + eigenes Volume + eigene DB/User)
- alle Instanzen nutzen denselben Infra-Unterbau (Traefik, CrowdSec, MySQL-Server)
- Versionierung pro Instanz über `wp_version` in den Hostvars

Damit kombinieren wir Isolation (pro Site) mit zentraler Härtung/Wartbarkeit (gemeinsame Infrastruktur).

## Neue Skripte
- `scripts/wp-add.sh`: Legt Hostvars an, prüft DNS-A-Records und deployt WordPress via Ansible.
- `scripts/wp-backup.sh`: Erstellt WordPress-Backup (DB-Dump + Volume + Hostvars + `wp_version` im Manifest).
- `scripts/wp-delete.sh`: Entfernt WordPress-Instanz (Container/DB/User/Volume + Hostvars).
- `scripts/wp-migrate-crowdsec.sh`: Ergänzt fehlende `wp_traefik_middleware_*` Defaults in Hostvars.
- `scripts/wp-redeploy.sh`: Validiert Hostvars + DNS und startet gezielten Redeploy.
- `scripts/wp-restore.sh`: Stellt DB und `/var/www/html` aus Backup wieder her, inkl. Versions-Guard.
- `scripts/wp-upgrade.sh`: Setzt `wp_version` in Hostvars und führt Redeploy aus.

## Backup-Format für WordPress
Das Backup-Archiv (`.tar.gz`) sollte diese Struktur haben:
- `meta/manifest.env` (Domain, Container, Volume, Timestamp, `wp_version`)
- `data/db.sql` (MySQL-Dump der WP-Datenbank)
- `data/html.tar.gz` (vollständiger Inhalt des WordPress Document-Roots `/var/www/html`)
- `files/hostvars.yml` (instanzspezifische Hostvars)

> Empfehlung: Der Document-Root **muss vollständig** enthalten sein, damit Plugins/Themes/Uploads konsistent restauriert werden können.

## Restore + Versionen ohne Datenverlust (wichtig bei mehreren WP-Versionen)
`wp-restore.sh` prüft Quell- (`Backup`) und Zielversion (`hostvars`).

- **Default:** Schutz vor unbeabsichtigtem Downgrade (Abbruch, wenn Zielversion kleiner als Backup-Version).
- `--restore-hostvars`: übernimmt Hostvars aus Backup (inkl. damaliger `wp_version`).
- `--allow-version-downgrade`: explizite Freigabe eines Downgrades (nur bewusst einsetzen).

Praxis-Empfehlung:
1. Restore auf gleiche/höhere `wp_version`.
2. Danach kontrollierter `wp-upgrade.sh`/`wp-redeploy.sh` auf die gewünschte Zielversion.

## Rewrite / Redirect (Apache zu Traefik)
- `.htaccess`-Rewrite-Regeln für Pretty Permalinks bleiben in WordPress relevant.
- Domain-Redirects (inkl. 301) gehören in Traefik-Router/Middlewares, nicht in WordPress.
- In dieser Erweiterung werden Alias-Domains per Redirect-Middleware dauerhaft (`permanent=true`) auf die Primärdomain geleitet.

## Konsistenzprüfung der Methoden
- Einheitliches Naming: `wp-<action>.sh` analog `ghost-<action>.sh`.
- Einheitliches Datenmodell in Hostvars (`wp_domain_db`, `wp_domain_usr`, `wp_domain_pwd`, `wp_version`, `wp_traefik_middleware_*`).
- Einheitliche Sicherheitslogik: DNS-Check vor Deploy/Redeploy, CrowdSec-Middleware für Frontend/Admin/API, Versions-Guard im Restore.

## Container-Reuse / Wiederverwendung
- **Traefik**: vollständig wiederverwendbar; 301-Redirects sind via RedirectRegex-Middleware möglich und in den WP-Labels vorgesehen.
- **MySQL (`ghost-mysql`)**: sicher wiederverwendbar durch getrennte DBs/DB-User pro Instanz.
- **Hostvars**: erweiterbar um zusätzliche Flags (z. B. Redirect-Strategien, Middleware-Wahl, optionale Traefik-Labels).

## Caching-Hinweis
Traefik bietet kein vollwertiges Response-Cache wie Nginx FastCGI-Cache. Für aggressives Seiten-Caching empfiehlt sich:
- WordPress-Plugin/Objektcache (Redis), oder
- vorgeschalteter CDN/Cache-Layer.
