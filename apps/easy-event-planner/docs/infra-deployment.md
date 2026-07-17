# Easy-Event-Planner – Infra & Deployment Konzept

## Zielarchitektur

```text
Internet -> Traefik -> easy-event-planner Container -> SQLite Volume -> SES/PayPal/Jobs
```

Domain: `events.geller.men`. Mandanten zunächst per Pfad: `/customerxyz`.

## Container

MVP: ein Container für API, Admin UI, Public Pages, Snippet JS und einfache Jobs. Später trennbar in API, Worker und Admin.

## Volumes

```text
/data/easy-event-planner.sqlite
/uploads
/certificates
```

SQLite im WAL-Modus. Backups über kontrollierte SQLite-Backup-Strategie, nicht blind während Schreiblast kopieren.

## Environment

```env
EEP_ENV=production
EEP_BASE_URL=https://events.geller.men
EEP_HTTP_ADDR=:8080
EEP_DB_DRIVER=sqlite
EEP_DB_PATH=/data/easy-event-planner.sqlite
EEP_CERTIFICATE_STORAGE_DIR=/certificates
EEP_SESSION_SECRET=...
EEP_TOKEN_PEPPER=...
EEP_INFRA_SYNC_TOKEN=...
EEP_MAIL_PROVIDER=ses
EEP_SES_REGION=eu-north-1
EEP_SES_SMTP_HOST=email-smtp.eu-north-1.amazonaws.com
EEP_SES_SMTP_PORT=587
EEP_SES_SMTP_USER=...
EEP_SES_SMTP_PASS=...
EEP_MAIL_FROM=noreply@events.geller.men
EEP_PAYPAL_MODE=sandbox
EEP_PAYPAL_CLIENT_ID=...
EEP_PAYPAL_CLIENT_SECRET=...
EEP_PAYPAL_WEBHOOK_ID=...
EEP_SEED_SETTINGS_JSON={"allowed_embed_origins":["*"],"event_detail_base_url":"https://www.geller.men/events"}
```

Hinweis: `EEP_SEED_SETTINGS_JSON` ist der saubere Infra-Hebel fuer Tenant-spezifische Embed-Regeln wie CORS-Freigaben (`allowed_embed_origins`) und redaktionelle Detailseiten (`event_detail_base_url`).

Fuer mandantenfaehige Custom-Domains rendert der Infra-Stack zusaetzlich einen internen Edge-Sync:

- `EEP_INFRA_SYNC_TOKEN` schuetzt die internen Export-/Refresh-Endpunkte fuer Domain-Bindings.
- `scripts/eep-domain-bindings-sync.sh <domain>` liest die freigegebenen Domain-Bindings aus EEP und schreibt daraus eine Traefik-File-Config.
- Ein systemd-Timer `eep-domain-bindings-sync@<domain>.timer` haelt Routing, Zertifikatsbereitstellung und Status-Refresh automatisch nach.

## Docker Compose Beispiel

```yaml
services:
  easy-event-planner:
    image: ghcr.io/andygellermann/easy-event-planner:latest
    restart: unless-stopped
    env_file: .env
    volumes:
      - ./data:/data
      - ./uploads:/uploads
      - ./certificates:/certificates
    networks: [proxy]
    labels:
      - traefik.enable=true
      - traefik.http.routers.eep.rule=Host(`events.geller.men`)
      - traefik.http.routers.eep.entrypoints=websecure
      - traefik.http.routers.eep.tls=true
      - traefik.http.services.eep.loadbalancer.server.port=8080
networks:
  proxy:
    external: true
```

## Ansible Rolle

```text
roles/easy_event_planner/
  tasks/main.yml
  templates/docker-compose.yml.j2
  templates/env.j2
  defaults/main.yml
```

## Deployment-Schritte

Verzeichnisse erstellen, `.env` aus Secrets generieren, Compose rendern, Image pullen, Container starten, Healthcheck prüfen, Migrationen ausführen, Smoke-Test.

## Sicherheit

Secrets per Ansible Vault, HTTPS-only, Secure Cookies, Webhook-Verifikation, keine Tokens im Log, Admin-Routen geschützt, Backups verschlüsseln.

## Akzeptanzkriterien

```text
[x] Container startet
[x] /healthz antwortet
[x] Traefik routet HTTPS
[x] SQLite persistiert
[x] Migrationen laufen
[x] Snippet-URL liefert JS
[x] Backup läuft
```
