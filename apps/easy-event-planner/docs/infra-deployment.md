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

## Betriebs-Runbook fuer eine neue Beta-Instanz

Eine neue Instanz besitzt eine eigene Domain, eigene Container und eine eigene SQLite-Datei. Sie wird mit genau einem ersten Mandanten angelegt; weitere Mandanten innerhalb derselben Instanz sind davon zu unterscheiden.

Vom Infra-Repository auf dem Docker-Host aus:

```bash
cd /home/andy/infra

./scripts/eep-add.sh beta-events.example.org \
  --tenant-slug=kundin \
  --tenant-name="Name der Kundin" \
  --mail-provider=ses \
  --mail-from=events@example.org \
  --mail-from-name="Easy Event Planner" \
  --admin-email=kundin@example.org \
  --admin-name="Name der Kundin"
```

Voraussetzung ist, dass die Domain bereits auf den Host zeigt. Nur fuer eine bewusst vor DNS vorbereitete Konfiguration darf `--skip-dns-check` verwendet werden.

Danach `ansible/hostvars/beta-events.example.org.yml` pruefen. Wichtig sind insbesondere:

- `eep_base_url`, Absender und erster Admin
- ein eigener, vom Skript erzeugter `eep_token_pepper`
- `eep_seed_settings_json` mit den erlaubten Embed-Urspruengen
- `eep_paypal_use_real_api: false` fuer den ersten Beta-Test

Die SES-Zugangsdaten liegen gemeinsam und verschluesselt in `ansible/secrets/secrets.yml`:

```yaml
ses_smtp_host: "email-smtp.eu-central-1.amazonaws.com"
ses_smtp_port: 587
ses_smtp_user: "..."
ses_smtp_password: "..."
ses_from: "EEP <events@example.org>"
```

Die Absenderdomain beziehungsweise Absenderadresse muss in Amazon SES verifiziert sein. Befindet sich SES noch in der Sandbox, muessen auch Empfaengeradressen verifiziert sein.

Deployment und direkte Kontrolle:

```bash
./scripts/eep-redeploy.sh beta-events.example.org
./scripts/eep-smoke-check.sh beta-events.example.org
docker ps --filter name=easy-event-planner-beta-events-example-org
docker logs --tail 100 easy-event-planner-beta-events-example-org
docker logs --tail 100 easy-event-planner-worker-beta-events-example-org
```

Das Seed-Kommando ist wiederholbar, sollte nach erfolgreicher Erstanlage aber mit `eep_seed_enabled: false` in den Hostvars abgeschaltet werden, damit spaetere Deployments nicht unbemerkt Seed-Daten nachpflegen.

## Backup und Restore testen

Das Backup stoppt Worker und App kurz kontrolliert, kopiert danach SQLite samt den uebrigen Instanzdateien und startet zuvor laufende Container wieder. Dadurch entsteht eine konsistente Sicherung; waehrenddessen gibt es ein kurzes Wartungsfenster.

Das Archiv enthaelt personenbezogene Veranstaltungsdaten und die gerenderte Env-Datei. Es wird daher mit restriktiven Dateirechten erzeugt und muss bei externer Aufbewahrung zusaetzlich verschluesselt werden.

```bash
cd /home/andy/infra
./scripts/eep-backup.sh --create beta-events.example.org
ls -lh ./backups/eep/beta-events.example.org/
tar tzf ./backups/eep/beta-events.example.org/eep-backup-beta-events.example.org-YYYYMMDD-HHMMSS.tar.gz | head
```

Ein echter Restore-Test ueberschreibt die aktuelle Instanz. Deshalb zuerst mit einer leeren beziehungsweise entbehrlichen Beta-Instanz und in einem angekuendigten Wartungsfenster testen. Das Restore-Skript erzeugt vorab automatisch ein zusaetzliches Safety-Backup und fuehrt anschliessend den Smoke-Test aus:

```bash
./scripts/eep-restore.sh beta-events.example.org \
  ./backups/eep/beta-events.example.org/eep-backup-beta-events.example.org-YYYYMMDD-HHMMSS.tar.gz \
  --yes --redeploy
```

Danach mindestens Admin-Login, ein Testevent, Registrierung, E-Mail-Zustellung, Snippet und Teilnehmerportal praktisch pruefen. Ein erfolgreich entpacktes Archiv allein ist noch kein bestandener Restore-Test.

## Stuendliches Log- und Verfuegbarkeitsmonitoring

`scripts/eep-log-monitor.sh` prueft App und Worker, Container-Restarts, `/healthz`, `/readyz`, `/version` sowie nur die seit dem letzten Lauf neu hinzugekommenen ERROR-/WARN-Logzeilen. Tokens, Zugangsdaten und E-Mail-Adressen werden aus Mail-Logauszuegen entfernt. Zustandsaenderungen und Entwarnungen werden gemeldet; unveraenderte Probleme und alte Logzeilen erzeugen keine identischen Mails.

Optionale dedizierte Mailwerte in `ansible/secrets/secrets.yml` (sonst werden `infra_error_notify_*` und die gemeinsamen `ses_*`-Werte verwendet):

```yaml
eep_log_monitor_to: "betrieb@example.org"
eep_log_monitor_from: "EEP Monitor <events@example.org>"
eep_log_monitor_subject_prefix: "[EEP Beta]"
```

Einrichtung testen:

```bash
./scripts/eep-log-monitor.sh --domain beta-events.example.org --dry-run
./scripts/eep-log-monitor.sh --domain beta-events.example.org --test-mail
```

Cronjob fuer alle EEP-Instanzen jeweils sieben Minuten nach der vollen Stunde (als Benutzer mit Docker-Zugriff und Leserecht auf die Secrets):

```cron
7 * * * * /bin/sh -lc 'umask 077; mkdir -p /home/andy/infra/data/eep-log-monitor && cd /home/andy/infra && ./scripts/eep-log-monitor.sh --all >> /home/andy/infra/data/eep-log-monitor/cron.log 2>&1'
```

Installation beispielsweise mit `crontab -e`, danach pruefen:

```bash
crontab -l
tail -n 100 /home/andy/infra/data/eep-log-monitor/cron.log
```

Mit `--no-warnings` werden nur Fehler erfasst. `--dry-run` verschickt keine Mail und veraendert den gespeicherten Cursor nicht; `--no-mail` arbeitet dagegen bewusst zustandsbehaftet. Der Monitor verwaltet Cursor und Zustandsvergleich unter `data/eep-log-monitor/state.json`; diese Laufzeitdaten gehoeren nicht ins Git-Repository.

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
