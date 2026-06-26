# Easy-Event-Planner – Pre-Live Roadmap

Stand: 2026-05-15

## Ziel

Diese Liste sammelt die sicherheits- und betriebskritischen Aufgaben vor dem produktiven Live-Betrieb.

## P0 – Vor Live-Betrieb (verbindlich)

- [ ] Betriebsmodus final festlegen: `SQLite` als Startphase oder Migration auf `Postgres` vor Go-Live.
- [ ] Netzwerkgrenzen fixieren: App nur hinter Traefik, kein direkter Public-Port auf die App.
- [ ] DB-Isolation sicherstellen: keine oeffentliche DB-Portfreigabe (bei Postgres kein `0.0.0.0:5432`).
- [ ] Secrets-Haertung: `EEP_TOKEN_PEPPER` und weitere Secrets nur aus Secret-Store/Ansible-Vault, keine Defaults.
- [ ] Proxy-Trust-Regel: `X-Forwarded-*` nur akzeptieren, wenn Request ueber trusted reverse proxy kommt.
- [ ] Cookie-/Session-Haertung pruefen: `Secure`, `HttpOnly`, `SameSite`, Session-TTL und Logout-Verhalten.
- [ ] Backup-Strategie verbindlich machen: RPO/RTO, Retention, Verschluesselung, Offsite-Speicher.
- [ ] Restore-Runbook dokumentieren: kompletter Wiederanlauf und Teil-Restore mit klaren Schritten.
- [ ] Disaster-Recovery-Testlauf in Staging: Backup einspielen, Smoke-Test, Freigabe protokollieren.
- [ ] Logging/Monitoring minimal produktionsreif: Healthchecks, Error-Alarmierung, Audit-Trails verifizieren.

## P1 – Kurz nach Live-Betrieb

- [ ] Automatisierte Backup-Integritaetspruefung (regelmaessig, mit Alerting bei Fehlern).
- [ ] Geplanter Restore-Drill (monatlich oder quartalsweise) mit dokumentierter Dauer und Findings.
- [ ] Sicherheits-Header-Haertung im HTTP-Layer (z. B. HSTS/CSP/X-Frame-Options je nach UI-Plan).
- [ ] Rate-Limit-/Abuse-Review fuer oeffentliche Endpunkte mit Lastprofil.
- [ ] Include-Haertung: bei config-basierten Snippets nur ein Parameter (`config`) erlauben, keine Query-Overrides.
- [ ] Include-Multitenancy absichern: Tenant-Trennung fuer `include.js` und Snippet-Config in Tests und Monitoring explizit pruefen.
- [ ] Opaque Include-Slug/Token einfuehren (kryptischer Wert statt sprechendem Slug) und rotierbar machen.
- [ ] Snippet-Event-Response-Caching mit Invalidation bei DB-Aenderungen (ETag/Last-Modified oder in-memory Cache).

## Konkrete Infra-Umsetzungspunkte

- [x] `scripts/infra-backup.sh` um Easy-Event-Planner-Datenpfade/Volumes erweitern.
- [x] `scripts/infra-restore.sh` um Easy-Event-Planner-Restore-Schritte erweitern.
- [x] App-spezifische Skripte `eep-backup.sh` und `eep-restore.sh` einfuehren (inkl. Safety-Backup vor Restore).
- [x] Post-Restore-Smoke-Flow standardisieren (`/healthz`, `/readyz`, Kern-Endpunkte).
