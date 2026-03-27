# Infra Stack Releases

## 1.1.0 (planned) — Infra Stack WP-Extension
- WordPress Betriebs- und Verwaltungs-Skripte ergänzt (`wp-add`, `wp-backup`, `wp-delete`, `wp-migrate-crowdsec`, `wp-redeploy`, `wp-restore`, `wp-upgrade`).
- Neue Ansible-Playbooks/Rollen für WordPress Deploy/Delete.
- Einheitliche Hostvars-Flags für WordPress + Traefik/CrowdSec-Middleware.
- Backup/Restore-Standard für WordPress dokumentiert.
- Restore-Härtung mit Domain-Migrations-Checks und optionalem `wp_image_tag` via `--php-version`.

## 1.0.0 (current baseline)
- Stabiler Infra-Kern mit Traefik, MySQL, CrowdSec, Portainer.
- Produktive Ghost-Automation inkl. Deploy, Backup, Restore, Redeploy, Upgrades.
