# Infra Stack Releases

## Infra-Härtung Roadmap
- HSTS als nachgelagerte HTTPS-Härtung evaluieren und stufenweise einführen.
- Einstieg bewusst defensiv: zunächst `Strict-Transport-Security: max-age=86400` ohne `includeSubDomains` und ohne `preload`.
- Erst nach erfolgreicher Prüfung aller betroffenen Hosts/Subdomains schrittweise auf längere Laufzeiten und ggf. `includeSubDomains` erhöhen.
- Ghost Native Analytics auf dem bestehenden Ansible-/Traefik-Stack pilotieren, bevor ein breiter Rollout auf weitere produktive Ghost-Instanzen erfolgt.

## 1.1.0 (planned) — Infra Stack WP-Extension
- WordPress Betriebs- und Verwaltungs-Skripte ergänzt (`wp-add`, `wp-backup`, `wp-delete`, `wp-migrate-crowdsec`, `wp-redeploy`, `wp-restore`, `wp-upgrade`).
- Neue Ansible-Playbooks/Rollen für WordPress Deploy/Delete.
- Einheitliche Hostvars-Flags für WordPress + Traefik/CrowdSec-Middleware.
- Backup/Restore-Standard für WordPress dokumentiert.
- Restore-Härtung mit Domain-Migrations-Checks und optionalem `wp_image_tag` via `--php-version`.

## 1.0.0 (current baseline)
- Stabiler Infra-Kern mit Traefik, MySQL, CrowdSec, Portainer.
- Produktive Ghost-Automation inkl. Deploy, Backup, Restore, Redeploy, Upgrades.
