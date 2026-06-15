# 📘 Dokumentation der Ghost-Infra-Skripte

### ghost-add.sh

**Beschreibung:**  
Dieses Skript erstellt eine neue Ghost-Instanz inklusive Docker-Container, Datenbankeintrag, Zertifikatseinrichtung über Traefik sowie der passenden `hostvars` Datei.

**Syntax:**
```bash
./scripts/ghost-add.sh DOMAIN [ALIAS] [--version=<major|major.minor|major.minor.patch|latest>]
```

**Parameter:**
- `DOMAIN` – Die Hauptdomain, z. B. `blog.example.com`
- `ALIAS` – (optional) Alias-Domain, z. B. `www.blog.example.com`
- `--version=<major|major.minor|major.minor.patch|latest>` – (optional) setzt die gewünschte Ghost-Version, z. B. `--version=4`, `--version=6.18` oder `--version=6.18.2`; ohne Angabe wird `latest` verwendet

**Vorgänge:**
- Docker-Container mit Labels für Traefik wird erzeugt
- Datenbank wird erstellt
- `hostvars/DOMAIN.yml` inkl. ALIAS wird automatisch generiert
- Zertifikat via Let's Encrypt wird beantragt


### ghost-upgrade.sh

**Beschreibung:**  
Hebt eine bestehende Ghost-Instanz auf eine neue Version an, indem `ghost_version` in den Hostvars angepasst und anschließend das Deployment neu ausgeführt wird. Ist die Zielversion bereits gesetzt, wird trotzdem ein Redeploy ausgelöst, damit z. B. `latest`-Instanzen aktualisiert werden können.

**Syntax:**
```bash
./scripts/ghost-upgrade.sh DOMAIN --version=<major|major.minor|major.minor.patch|latest> [--force-major-jump] [--dry-run]
```

**Parameter:**
- `DOMAIN` – Die bestehende Ghost-Domain
- `--version=<major|major.minor|major.minor.patch|latest>` – Zielversion, z. B. `--version=5`, `--version=6.18` oder `--version=6.18.2`
- `--force-major-jump` – erlaubt Sprünge größer als +1 Major-Version
- `--dry-run` – schreibt nur die Hostvars (inkl. Backup), ohne Deployment

**Features:**
- Liest und validiert die aktuelle `ghost_version`
- Verhindert standardmäßig große Versionssprünge (z. B. 4 → 6)
- Erstellt automatisch ein Backup der Hostvars-Datei (`.bak.<timestamp>`)
- Führt auch bei unveränderter Zielversion ein Redeploy aus, damit Floating-Tags wie `latest` neu gezogen werden
- Führt danach ein reguläres `ansible-playbook` Deployment aus

### ghost-upgrade-all.sh

**Beschreibung:**  
Führt ein Bulk-Upgrade für alle Ghost-Domains aus, die in `ansible/hostvars/*.yml` per `ghost_domain_db` erkannt werden. Intern wird pro Domain `ghost-upgrade.sh` aufgerufen.

**Syntax:**
```bash
./scripts/ghost-upgrade-all.sh --version=<major|major.minor|major.minor.patch|latest> [--force-major-jump] [--dry-run] [--continue-on-error]
```

**Parameter:**
- `--version=<major|major.minor|major.minor.patch|latest>` – Zielversion für alle Ghost-Instanzen
- `--force-major-jump` – erlaubt Major-Sprünge größer als +1 (pro Domain)
- `--dry-run` – führt nur die Hostvars-Änderungen/Prüflogik aus, ohne Deployment
- `--continue-on-error` – setzt den Batch bei Domain-Fehlern fort und listet Fehler am Ende gesammelt

**Beispiele:**
```bash
./scripts/ghost-upgrade-all.sh --version=6
./scripts/ghost-upgrade-all.sh --version=6.18.2
./scripts/ghost-upgrade-all.sh --version=latest --continue-on-error
./scripts/ghost-upgrade-all.sh --version=5 --force-major-jump --dry-run
```

### ghost-delete.sh

**Beschreibung:**  
Dieses Skript entfernt eine bestehende Ghost-Instanz inklusive Datenbank und Hostvars. Optional mit Backup & vollständigem Löschen.

**Syntax:**
```bash
./scripts/ghost-delete.sh DOMAIN [--purge]
```

**Parameter:**
- `DOMAIN` – Die zu entfernende Ghost-Domain
- `--purge` – (optional) löscht alle zugehörigen Daten, inkl. Backups

**Features:**
- Sicheres Entfernen des Containers und der DB
- Optionaler Backup vor Löschung
- Log-Eintrag in `/logs`
- Interaktive Bestätigung bei gefährlichen Operationen

### delete-domain.sh

**Beschreibung:**  
Zentraler Löschbefehl für Domains. Erkennt Ghost-, WordPress-, Static- und Redirect-Konfigurationen, entfernt die jeweilige Domain und bereinigt anschließend verwaiste ACME-Zertifikatseinträge aus Traefiks `acme.json`.

**Syntax:**
```bash
./scripts/delete-domain.sh DOMAIN [--yes]
```

**Features:**
- erkennt den Domain-Typ über `ansible/hostvars/DOMAIN.yml`
- ruft passend `ghost-delete.sh`, `wp-delete.sh` oder `static-delete.sh` auf
- entfernt Redirect-Einträge aus `ansible/redirects/redirects.yml`
- bereinigt passende Zertifikats-Einträge aus `/home/andy/infra/data/traefik/acme/acme.json`
- erstellt Backups von `redirects.yml` und `acme.json`
- startet Traefik nach der ACME-Bereinigung neu

### wildcard-export.sh

**Beschreibung:**  
Exportiert ein vorhandenes Wildcard-Zertifikat für eine Apex-Domain aus Traefiks `acme.json`.

**Syntax:**
```bash
./scripts/wildcard-export.sh example.com [--output-dir /tmp/wildcard-export]
```

### wildcard-housekeeping.sh

**Beschreibung:**  
Prüft, ob für eine Apex-Domain bereits ein echtes Wildcard-Zertifikat in Traefiks `acme.json` vorliegt, und bereinigt anschließend nur die davon vollständig abgedeckten Einzelzertifikate. Optional kann Traefik danach direkt neu gestartet werden.

**Syntax:**
```bash
./scripts/wildcard-housekeeping.sh example.com [--restart-traefik]
```

### wildcard-distribute.sh

**Beschreibung:**  
Verteilt ein exportiertes Wildcard-Zertifikat an definierte Zielserver, z. B. für Staging-Systeme. Die zentrale Mapping-Datei kann exakte Zielpfade, einen separaten SSH-Key und optionale Post-Deploy-Kommandos pro Server enthalten.

**Syntax:**
```bash
./scripts/wildcard-distribute.sh example.com [--config ./ansible/wildcards/export.yml] [--dry-run]
./scripts/wildcard-distribute.sh --all [--config ./ansible/wildcards/export.yml] [--dry-run]
```

**Konfigurationsmodell (`ansible/wildcards/export.yml`):**
```yaml
wildcard_exports:
  - source_domain: example.com
    acme_file: /home/andy/infra/data/traefik/acme/acme.json
    targets:
      - host: staging-1.example.net
        user: root
        identity_file: /home/andy/.ssh/id_wildcard_staging
        remote_fullchain_path: /etc/nginx/ssl/example.com/fullchain.pem
        remote_privkey_path: /etc/nginx/ssl/example.com/privkey.pem
        post_deploy_command: "systemctl reload nginx"
```

**Features:**
- verarbeitet einen einzelnen Wildcard-Eintrag oder alle Einträge per `--all`
- unterstützt exakte Zielpfade oder alternativ `remote_dir`
- erlaubt separate SSH-Keys pro Zielserver über `identity_file`
- kann auf dem Zielserver optional einen lokalen Reload-/Install-Hook ausführen
- `--dry-run` zeigt den Plan an, ohne Export oder SSH-Transfer auszuführen

### wildcard-distribute-on-change.sh

**Beschreibung:**  
Prüft die konfigurierten Wildcard-Exporte gegen einen Fingerprint-State und löst die Verteilung nur dann aus, wenn sich das eigentliche Zertifikat seit dem letzten erfolgreichen Lauf geändert hat. Gedacht als Trigger hinter `systemd.path` oder als sparsamer Fallback-Cron.

**Syntax:**
```bash
./scripts/wildcard-distribute-on-change.sh example.com [--config ./ansible/wildcards/export.yml] [--state-dir ./data/wildcard-distribution-state] [--dry-run] [--force]
./scripts/wildcard-distribute-on-change.sh --all [--config ./ansible/wildcards/export.yml] [--state-dir ./data/wildcard-distribution-state] [--dry-run] [--force]
```

**Features:**
- berechnet pro `source_domain` einen SHA-256-Fingerprint des aktuellen Leaf-Zertifikats
- schreibt den letzten erfolgreichen Stand in `data/wildcard-distribution-state`
- ruft `wildcard-distribute.sh` nur bei echter Zertifikatsänderung auf
- `--dry-run` zeigt, welche Verteilungen aktuell ausgelöst würden
- `--force` verteilt trotz unverändertem Fingerprint erneut
- eignet sich direkt für `systemd.path` auf `data/traefik/acme/acme.json`

Hinweis:
- Wildcard-Restores für WordPress, Static und Ghost unterstützen zusätzlich `--wildcard-domain=<apex-domain>` und `--dns-account=<key>`.

### create-hostvars.sh

**Beschreibung:**  
Erstellt eine passende `hostvars` Datei für eine neue Ghost-Domain automatisch.

**Syntax:**
```bash
./scripts/create-hostvars.sh DOMAIN [ALIAS] [--version=<major|major.minor|major.minor.patch|latest>]
```

**Parameter:**
- `DOMAIN` – Hauptdomain
- `ALIAS` – (optional) Aliasdomain
- `--version=<major|major.minor|major.minor.patch|latest>` – (optional) gewünschte Ghost-Version für den Container-Tag; ohne Angabe wird `latest` verwendet

**Features:**
- Validiert Eingaben (inkl. Punycode bei Umlauten)
- Schreibt in `ansible/hostvars`
- Warnung bei bestehenden Dateien
- Ergänzt standardmäßig Legacy-/Custom-Tinybird-Defaults (`tinybird_*`) pro neuer Domain
- Ergänzt Platzhalter für Ghost Native Analytics (`ghost_native_analytics_*`) pro neuer Domain

### Amazon SES (Standard für Ghost-Mail)

Die Ghost-Container verwenden standardmäßig Amazon SES als SMTP-Transport. Lege die Zugangsdaten einmalig in `ansible/secrets/secrets.yml` ab, damit sie bei jeder Neuanlage automatisch genutzt werden. Beispiel:

```yaml
ses_smtp_user: "AKIA...SMTP_USER"
ses_smtp_password: "DEIN_SMTP_PASSWORT"
ses_from: "Infra <noreply@deine-domain.tld>"

# Optional (Defaults werden verwendet, wenn nicht gesetzt)
ses_smtp_host: "email-smtp.eu-central-1.amazonaws.com"
ses_smtp_port: 587
ses_smtp_secure: false
```

Wenn `ses_from` nicht gesetzt ist, wird automatisch `noreply@<domain>` verwendet. Individuelle Abweichungen kannst du pro Instanz im jeweiligen `ansible/hostvars/<domain>.yml` überschreiben.

## Fehler-Notifications fuer Batch-Skripte

Autonom laufende Batch-Skripte koennen bei Fehlern SMTP-Mails ueber Amazon SES versenden. Dafuer werden die vorhandenen SES-Zugangsdaten aus `ansible/secrets/secrets.yml` verwendet.

Neue optionale Secrets:

```yaml
infra_error_notify_to: "andy@example.org"
infra_error_notify_from: "Infra Bot <noreply@example.org>"
infra_error_notify_subject_prefix: "[infra]"
```

Hinweise:
- `infra_error_notify_to` aktiviert die Fehler-Mailings. Ohne diesen Key bleibt die Notification still deaktiviert.
- `infra_error_notify_from` ist optional. Ohne diesen Key wird `ses_from` als Absender verwendet.
- SMTP-Host, Port, User und Passwort kommen weiterhin aus `ses_smtp_*`.

Aktuell angebunden sind:
- `scripts/infra-update-all.sh`
- `scripts/infra-backup.sh`
- `scripts/redeploy-all-web.sh`
- `scripts/ghost-backup-all.sh`
- `scripts/ghost-upgrade-all.sh`
- `scripts/wildcard-distribute.sh`

## Domain Health Monitor

Fuer einen einfachen Cron-basierten Aussencheck gibt es jetzt
`scripts/domain-health-monitor.py`.

Was das Skript prueft:
- HTTPS-Erreichbarkeit pro Domain
- TLS-Handshake und Zertifikatsgueltigkeit
- HTTP-Status am Pfad `/`
- Redirect-Ziele fuer Eintraege aus `ansible/redirects/redirects.yml`
- optionale Zertifikatswarnung vor Ablauf

Automatische Discovery:
- alle oeffentlichen Domains aus `ansible/hostvars/*.yml`
- zusaetzliche `traefik.aliases`
- Redirect-Quellen und Redirect-Aliase aus `ansible/redirects/redirects.yml`

Standardverhalten:
- normale Sites gelten als gesund bei `2xx`, `3xx`, `401` oder `403`
- Redirect-Domains muessen den konfigurierten Redirect liefern
- neue Fehler oder geaenderte Fehlerbilder loesen genau eine Mail aus
- unveraenderte Dauerfehler werden nicht alle 15 Minuten erneut gemailt
- Recoveries koennen optional ebenfalls gemailt werden

State-Datei:
- `data/domain-health-monitor/state.json`

Beispiele:

```bash
# normaler Lauf mit Mailversand bei Aenderungen
./scripts/domain-health-monitor.py

# nur pruefen, keine Mail
./scripts/domain-health-monitor.py --no-mail

# Zertifikate erst ab 7 Tagen Restlaufzeit warnen
./scripts/domain-health-monitor.py --warn-cert-days 7
```

Cron-Beispiel alle 15 Minuten:

```cron
*/15 * * * * /bin/sh -lc 'mkdir -p /home/andy/infra/data/domain-health-monitor && cd /home/andy/infra && /usr/bin/env python3 ./scripts/domain-health-monitor.py >> /home/andy/infra/data/domain-health-monitor/cron.log 2>&1'
```

Mail-Konfiguration:
- verwendet `infra_error_notify_to`
- verwendet `infra_error_notify_from` oder faellt auf `ses_from` zurueck
- SMTP-Zugangsdaten kommen aus `ses_smtp_*`

Voraussetzungen:
- `python3`

## Static Inline Editor

Beim Static Inline Editor kommen SMTP-Zugangsdaten fuer Magic Links jetzt ebenfalls global aus `ses_*` statt aus einzelnen Domain-Hostvars.

Wichtig:
- `static_editor_smtp_*` wird nicht mehr in Hostvars gepflegt
- Speichern erzeugt standardmaessig ein Dateibackup, aber keinen Git-Commit
- dadurch bleibt der Editor fuer seltene Kleinaenderungen bewusst schlank und ohne Git-Zwang
- fuer den Static Editor heisst das Ruecksicherungsziel jetzt `static_editor_undo_backups`


# 🌀 Ghost Backup & Restore Toolkit

Willkommen im Restore-Tempel deines Ghost CMS Docker-Systems.  
Dieses Toolkit ermöglicht dir die einfache Wiederherstellung gelöschter oder geänderter Ghost-Websites – vollständig automatisiert, abgesichert und protokolliert.

---

## 📜 `ghost-restore.sh`

Wiederherstellung einer Ghost-Instanz aus einem `.tar.gz`-Backup.

### 🔧 Syntax

```bash
./scripts/ghost-restore.sh [domain] [options]
```

---

## 🏷️ Verfügbare Optionen / Flags

| Flag | Beschreibung |
|------|--------------|
| `--force` | Erzwingt die Wiederherstellung und ersetzt eine bereits existierende Instanz ohne Rückfrage. |
| `--dry-run` | Führt keine Wiederherstellung durch. Prüft nur, ob das gewählte Backup vollständig und gültig ist. |
| `--purge` | (Geplant für `ghost-delete.sh`) Entfernt _endgültig_ inkl. Datenbank, Docker-Volume und Hostvars. |
| `--select` | Öffnet ein interaktives Menü zur Auswahl eines Backups aus dem Backup-Ordner. |
| `--help` | Zeigt diese Übersicht an. |

---

## 📂 Backup-Verzeichnisstruktur

Backups werden im folgenden Format abgelegt:

```
infra/backups/ghost/<domain>/<timestamp>.tar.gz
```

### Inhalt eines gültigen Backups:

- Docker Volume (Ghost Content)
- MySQL Dump
- `hostvars/<domain>.yml`

---

## 📓 Logs

Alle Wiederherstellungsaktionen werden protokolliert unter:

```
infra/logs/ghost-restore/<domain>/<timestamp>.log
```

---

## ⚠️ Sicherheit & Hinweise

- Keine Verschlüsselung, kein Passwortschutz: bitte Backups sicher verwahren.
- Die `--dry-run`-Option kann verwendet werden, um Backups vor der Wiederherstellung zu prüfen.
- Im Restore-Prozess wird überprüft, ob `hostvars/<domain>.yml` im Backup enthalten ist. Fehlt diese Datei ➤ Abbruch.

---

Bleibe bei deiner Macht. Restore mit Bedacht.


**Hinweis zu Node.js:**
Die Node.js-Version wird automatisch durch das gewählte offizielle Ghost-Docker-Image bestimmt (z. B. `ghost:4`, `ghost:5`, `ghost:6`). Dadurch ist immer die zur Ghost-Version passende Node-Laufzeit enthalten.

### Ghost-Version auf nächste Major-Version anheben

1. Hostvars der Domain anpassen (`ansible/hostvars/<domain>.yml`):
   ```yaml
   ghost_version: "5"
   ```
2. Deployment erneut ausführen:
   ```bash
   ./scripts/ghost-add.sh <domain> --version=5
   ```
   oder alternativ direkt:
   ```bash
   ansible-playbook -i ./ansible/inventory -e "target_domain=<domain>" ./ansible/playbooks/deploy-ghost.yml
   ```
3. Anschließend Ghost-Admin unter `/ghost` prüfen und ggf. Migrationshinweise im Dashboard bestätigen.

### infra-setup.sh

**Beschreibung:**  
Initialisiert den kompletten Infra-Stack auf einem frischen Host und installiert fehlende Basis-Tools automatisch (Docker, Ansible, MySQL-Client, dnsutils, jq, Python 3, community.docker Collection).

**Syntax:**
```bash
sudo ./scripts/infra-setup.sh
```

**Funktionen:**
- Fragt die Portainer-Domain interaktiv ab.
- Deployt nacheinander MySQL, Traefik, CrowdSec und Portainer via Ansible.
- Prüft den A-Record der Portainer-Domain gegen die öffentliche Host-IP und bricht bei Abweichung ab.
- Richtet CrowdSec + Traefik-Bouncer ein, inkl. Middleware-Namen:
  - `crowdsec-default@docker`
  - `crowdsec-admin@docker`
  - `crowdsec-api@docker`

**Hinweis für weitere Container (z. B. WordPress / statische Seiten):**
- Hänge die passenden Traefik-Router an mindestens `crowdsec-default@docker`.
- Für WordPress-Backend explizit zusätzliche Router für `/wp-admin` und `/wp-login.php` mit `crowdsec-admin@docker` verwenden.
- Für APIs (Ghost/WordPress) Router mit `crowdsec-api@docker` verwenden.

### infra-backup.sh

**Beschreibung:**  
Erstellt ein Gesamt-Backup des Infra-Stacks als `tar.gz` (Docker-Volumes + relevante Konfigurationen + optional MySQL all-databases Dump).

**Syntax:**
```bash
./scripts/infra-backup.sh --create [--output /pfad/infra-backup.tar.gz] [--no-mysql-dump]
```

**Enthaltene Bestandteile (wenn vorhanden):**
- Docker Volumes: `mysql_data`, `portainer_data`, `ghost_*_content`, sowie Volumes mit Präfix `traefik*`/`crowdsec*`
- Dateibasierte Konfigurationen: `ansible/hostvars`, `ansible/secrets`, `data/traefik`, `data/crowdsec`
- Optional: `infra-mysql` Full-Dump via `mysqldump --all-databases`

### infra-restore.sh

**Beschreibung:**  
Stellt ein Gesamt-Backup wieder her (Dateien + Volumes + optional MySQL-Dump-Import). Kann vorab optional `infra-setup.sh` starten.

**Syntax:**
```bash
./scripts/infra-restore.sh --restore /pfad/infra-backup.tar.gz [--yes] [--run-setup]
```

**Hinweis:**
- Restore überschreibt Konfigurationen und Volume-Inhalte.
- Für produktive Systeme zuerst mit frischem Infra-Backup absichern.

### infra-update-all.sh

**Beschreibung:**  
Aktualisiert die zentralen Infrastruktur-Bausteine per Ansible (`infra-mysql`, `infra-traefik`, CrowdSec, Wazuh-Agent und `infra-portainer`).

**Syntax:**
```bash
# Vollständiges Infra-Update inkl. Portainer
./scripts/infra-update-all.sh --portainer-domain=portainer.example.com

# Einzelne Dienste überspringen
./scripts/infra-update-all.sh --skip-portainer --skip-crowdsec --skip-wazuh
```


### ghost-migrate-crowdsec.sh

**Beschreibung:**  
Migrationsskript für bestehende Ghost-Instanzen. Ergänzt fehlende CrowdSec-Middleware-Defaults in `ansible/hostvars/*.yml` und führt anschließend je Domain einen `ghost-redeploy.sh` aus.

**Syntax:**
```bash
# Nur prüfen (keine Änderungen)
./scripts/ghost-migrate-crowdsec.sh --check-only

# Migration + Redeploy aller Ghost-Domains
./scripts/ghost-migrate-crowdsec.sh

# Optional: Traefik am Ende einmalig neu starten
./scripts/ghost-migrate-crowdsec.sh --restart-traefik
```

**Hinweis:**
- Das ist der einfachste Weg, CrowdSec nachträglich für bestehende Ghost-Container zu aktivieren.
- Voraussetzung: DNS/Hostvars sind gültig, da intern `ghost-redeploy.sh` aufgerufen wird.

### ghost-migrate-tinybird.sh

**Beschreibung:**  
Migrationsskript für bestehende Ghost-Instanzen. Ergänzt fehlende Legacy-/Custom-Tinybird-Defaults in `ansible/hostvars/*.yml`, generiert pro Domain ein eigenes `tinybird_token` und führt optional je Domain `ghost-redeploy.sh` aus.

**Syntax:**
```bash
# Nur prüfen (keine Änderungen)
./scripts/ghost-migrate-tinybird.sh --check-only

# Migration + Redeploy aller Ghost-Domains
./scripts/ghost-migrate-tinybird.sh

# Nur Hostvars schreiben, ohne Redeploy
./scripts/ghost-migrate-tinybird.sh --no-redeploy

# Optional vorhandene Tokens rotieren
./scripts/ghost-migrate-tinybird.sh --rotate-tokens
```

**Empfohlener Setup-Ablauf:**
1. Tinybird-seitig den Ziel-Workspace und die gewünschte Event-/Datasource-Struktur vorbereiten.
2. Für neue Ghost-Domains reichen die automatisch erzeugten `tinybird_*` Hostvars in `create-hostvars.sh` normalerweise aus.
3. Für bestehende Ghost-Domains zuerst prüfen:
```bash
./scripts/ghost-migrate-tinybird.sh --check-only
```
4. Dann die fehlenden Tinybird-Defaults in die Hostvars schreiben und Ghost neu deployen:
```bash
./scripts/ghost-migrate-tinybird.sh
```
5. Anschließend pro Domain Smoke-Check fahren:
```bash
./scripts/ghost-smoke-check.sh <domain>
```
6. Optional Tokens rotieren:
```bash
./scripts/ghost-migrate-tinybird.sh --rotate-tokens
```

**Wichtige Hostvars/Env-Werte:**
- `tinybird_enabled`
- `tinybird_api_url`
- `tinybird_workspace`
- `tinybird_datasource`
- `tinybird_token`
- `tinybird_events_endpoint`

Hinweis:
- Diese Infra übergibt die Tinybird-Werte an die Ghost-Container-Umgebung. Die eigentliche Event-Erzeugung bzw. ein Theme-/Frontend-Hook, der Nutzungsdaten an Tinybird sendet, muss auf Ghost-/Theme-Seite ebenfalls vorhanden sein.
- Wenn direkt gegen die Tinybird Events API gesendet wird, muss `tinybird_token` ein echter Tinybird-Token mit Schreibrechten für die Zieldatenquelle sein; ein lokal generierter Zufallswert reicht dafür nicht aus.
- `tinybird_api_url` muss zum API-Host der Tinybird-Workspace-Region passen. Der Default `https://api.tinybird.co` ist nur korrekt, wenn eure Workspace-Region genau diesen Host verwendet.
- Wichtige Abgrenzung: Diese Hostvars-/Env-Vorbereitung ist nicht automatisch identisch mit Ghost 6 Native Analytics. Laut offizieller Ghost-Doku funktioniert die native Tinybird-Analytics für Self-Hosting nur über den Docker-Preview-/`ghost-docker`-Setup mit zusätzlichen `tinybird-*` Diensten und dem separaten Traffic-Analytics-Service.

**Hinweis:**
- Ideal für den Nachzug bei bereits produktiven Domains.
- Für einen kontrollierten Rollout zuerst `--check-only`, danach regulär ausführen.

### ghost-native-analytics-sync.sh

**Beschreibung:**  
Kopiert die offiziellen Tinybird-Projektdateien aus einem laufenden Ghost-Container in ein lokales Arbeitsverzeichnis. Das ist der vorbereitende Operator-Schritt fuer Ghost Native Analytics auf dem bestehenden Ansible-/Traefik-Stack.

**Syntax:**
```bash
# Standardziel unter ./data/ghost-native-analytics/<domain>/tinybird
./scripts/ghost-native-analytics-sync.sh <domain>

# Bestehendes Zielverzeichnis vorher bereinigen
./scripts/ghost-native-analytics-sync.sh <domain> --clean

# Eigenes Zielverzeichnis verwenden
./scripts/ghost-native-analytics-sync.sh <domain> --output-dir /tmp/ghost-tinybird
```

**Ghost Native Analytics (Pilot-Pfad):**
1. In `ansible/secrets/secrets.yml` ein Profil unter `ghost_native_analytics_profiles` anlegen:
```yaml
ghost_native_analytics_profiles:
  woelfe_workspace:
    api_url: https://api.europe-west2.gcp.tinybird.co
    workspace_id: <tinybird-workspace-id>
    admin_token: <tinybird-admin-token>
    tracker_token: <tinybird-tracker-token>
```
2. In den Hostvars der Pilot-Domain aktivieren:
```yaml
ghost_native_analytics_enabled: true
ghost_native_analytics_profile: woelfe_workspace
ghost_native_analytics_tracker_datasource: analytics_events
ghost_traefik_middleware_native_analytics: crowdsec-api@docker
```
3. `./scripts/ghost-native-analytics-sync.sh <domain>` ausfuehren, damit die Tinybird-Datenfiles aus dem Ghost-Container vorliegen.
4. Das synchronisierte Tinybird-Projekt anschliessend bewusst per Tinybird CLI gegen den Ziel-Workspace deployen.
5. Danach `./scripts/ghost-redeploy.sh <domain>` ausfuehren.

**Was die Rolle jetzt automatisch uebernimmt, sobald `ghost_native_analytics_enabled: true` gesetzt ist:**
- offizieller Ghost-Env-Pfad via `tinybird__tracker__endpoint`, `tinybird__adminToken`, `tinybird__workspaceId`, `tinybird__tracker__datasource`, `tinybird__stats__endpoint`
- separater `ghost-analytics-<domain>` Container auf Basis von `ghost/traffic-analytics`
- eigene Traefik-Route fuer `/.ghost/analytics/**` mit hoeherer Prioritaet als die allgemeine `/.ghost`-Route
- eigenes Salt-Volume pro Ghost-Domain fuer den Analytics-Proxy

**Wichtige Abgrenzung:**
- Der eigentliche Tinybird-Cloud-Deploy der synchronisierten Datenfiles ist derzeit bewusst noch ein expliziter Operator-Schritt.
- So bleibt der produktive Runtime-Rollout von Ghost + Traefik entkoppelt vom heikleren Tinybird-CLI-/Workspace-Deploy.

### ghost-backup.sh

**Beschreibung:**  
Selektives All-in-One Backup/Restore für eine einzelne Ghost-Instanz inkl. DB, Content-Volume, Hostvars und optional CrowdSec-Dateien.

**Syntax:**
```bash
# Backup
./scripts/ghost-backup.sh --create <domain> [--output /pfad/ghost-backup.tar.gz]

# Restore
./scripts/ghost-backup.sh --restore <domain> <pfad/ghost-backup.tar.gz> [--yes] [--content-only] [--restore-hostvars]
```

**Backup-Inhalt:**
- SQL-Dump der Ghost-Datenbank (gemäß `ansible/hostvars/<domain>.yml`)
- Der Dump nutzt `mysqldump --no-tablespaces`, damit kein zusätzliches `PROCESS`-Privilege nötig ist.
- Ghost Content-Volume (`ghost_<domain>_content`)
- Hostvars der Domain
- **Keine** TLS-Zertifikate (`acme.json`) im Backup: Zertifikate werden nach Restore von Traefik/Let's Encrypt neu ausgestellt
- Optionaler best-effort Snapshot von `data/crowdsec` (`crowdsec.tar.gz`), ohne das Backup bei Root-/Permission-Dateien scheitern zu lassen


**Restore-Modi:**
- Standard: DB + Content, aber bestehende Hostvars bleiben unverändert (sicheres Default).
- `--restore-hostvars`: stellt zusätzlich Hostvars (und optional CrowdSec-Dateien) aus dem Backup wieder her.
- `--content-only`: **nur** Ghost-Content-Volume wird wiederhergestellt; Domain-Setup/Hostvars/DB/CrowdSec bleiben unverändert. Ideal zum Duplizieren in bestehende Ziel-Instanzen.

### ghost-backup-all.sh

**Beschreibung:**  
Erstellt in einem Lauf Backups für alle Ghost-Instanzen. Die Domains werden automatisch über `ansible/hostvars/*.yml` erkannt (Marker: `ghost_domain_db`).

**Syntax:**
```bash
./scripts/ghost-backup-all.sh [--output-dir <dir>] [--dry-run] [--continue-on-error]
```

**Parameter:**
- `--output-dir <dir>` – Zielbasisverzeichnis für alle erzeugten Backups (Default: `./backups/ghost`)
- `--dry-run` – zeigt nur den geplanten Backup-Lauf inkl. Zielpfaden, ohne Backups zu erstellen
- `--continue-on-error` – setzt den Batch bei Domain-Fehlern fort und listet Fehler am Ende gesammelt

**Beispiele:**
```bash
./scripts/ghost-backup-all.sh
./scripts/ghost-backup-all.sh --output-dir /tmp/ghost-bulk-backups
./scripts/ghost-backup-all.sh --dry-run
./scripts/ghost-backup-all.sh --continue-on-error
```

### ghost-smoke-check.sh

**Beschreibung:**  
Schneller Smoke-Check für eine Ghost-Domain nach Redeploy/Restore. Prüft Frontend-, Admin- und API-Erreichbarkeit über HTTPS, um Routing-/Middleware-Regressions schnell zu erkennen.

**Syntax:**
```bash
./scripts/ghost-smoke-check.sh <domain>
```

**Geprüfte Endpunkte:**
- `/` (Frontend)
- `/ghost/` (Admin-Route)
- `/ghost/api/admin/site/` (API-Erreichbarkeit)

### ghost-redeploy.sh

**Beschreibung:**  
Hilfsskript für bestehende Ghost-Instanzen nach Änderungen in `ansible/hostvars/<domain>.yml` (z. B. neue Alias-Domain). Vor dem Redeploy werden Integrität und DNS-Matching geprüft.

**Syntax:**
```bash
# Validieren + Redeploy
./scripts/ghost-redeploy.sh <domain>

# Nur validieren
./scripts/ghost-redeploy.sh <domain> --check-only

# Optional mit Traefik-Restart danach
./scripts/ghost-redeploy.sh <domain> --restart-traefik
```

**Prüfungen:**
- Pflichtwerte in Hostvars: `domain`, `ghost_domain_db`, `ghost_domain_usr`, `ghost_domain_pwd`
- Domain-Matching zwischen Argument und `hostvars.domain`
- DNS-IPv4-Matching (Hauptdomain + alle Aliase) gegen die öffentliche Host-IP

**Hinweis zu Wildcard-DNS:**
- Ein `*.example.com A ...` reicht nicht, wenn der konkrete Hostname bereits eigene Records wie `MX` oder `TXT` hat. In diesem Fall greift der Wildcard-A-Record für genau diesen Hostnamen nicht mehr, und ein expliziter `A`-Record ist nötig.

**Tinybird-Nachzug (Bestandsdomains):**
- `ghost-redeploy.sh` ist der empfohlene Weg, um Änderungen in Hostvars (inkl. `tinybird_*`) live in die Container-Umgebung zu übernehmen.
- Typischer Ablauf:
  1. `./scripts/ghost-migrate-tinybird.sh --check-only`
  2. `./scripts/ghost-migrate-tinybird.sh`
  3. Optional: `./scripts/ghost-smoke-check.sh <domain>`

**TLS/Let's Encrypt Hinweis:**
- Alias-Domains sind **relevant** für Zertifikate.
- Nach erfolgreichem Redeploy zieht Traefik die Zertifikate für die Host-Regeln nach (bei korrekt gesetztem DNS und eingehendem Traffic).

### Optional: Backend-Hinweis fuer den Ghost Image Editor

Fuer Ghost kann optional ein Admin-Modal aktiviert werden, das nur im Backend fuer eingeloggte Nutzer erscheint, wenn das Browser-Plugin `ghost-image-editor` im aktuellen Browser nicht aktiv ist.

Hostvars-Beispiel:

```yaml
ghost_image_editor_notice_enabled: true
ghost_image_editor_notice_install_url: "https://github.com/andygellermen/ghost-image-editor/"
ghost_image_editor_notice_remind_hours: 24
ghost_image_editor_notice_extra_title: "Nur im Ghost-Backend sichtbar"
ghost_image_editor_notice_extra_text: |
  Bitte installiere das Browser-Plugin im bevorzugten Redaktions-Browser.
  Nach der Installation die Ghost-Admin-Seite einmal neu laden.
```

Danach wie gewohnt ausrollen:

```bash
./scripts/ghost-redeploy.sh <domain>
```

Optional koennen die Zusatztexte spaeter pro Domain direkt in `ansible/hostvars/<domain>.yml` angepasst und mit demselben Redeploy uebernommen werden.

**CrowdSec-Routen (Ghost):**
- Standardseiten: standardmäßig **ohne** CrowdSec-Middleware (optional via `ghost_traefik_middleware_default: "crowdsec-default@docker"`)
- Admin: `/ghost` über `crowdsec-admin@docker`
- API-Hotspots: `/ghost/api`, `/.ghost`, `/members/api` über `crowdsec-api@docker`
- Diese Middleware-Defaults werden bei neuen Hostvars automatisch gesetzt und bei Restore alter Backups ergänzt.

### eep-add.sh

**Beschreibung:**  
Erzeugt eine neue Hostvars-Datei fuer eine Easy-Event-Planner-Domain inkl. DNS-Pruefung, Token-Pepper, Traefik-Aliases und optionalem Wildcard-TLS-Setup.

**Syntax:**
```bash
./scripts/eep-add.sh <domain> [alias1 alias2 ...] [--tenant-slug=<slug>] [--tenant-name=<name>] [--base-url=<url>] [--mail-provider=<log|smtp|ses>] [--mail-from=<email>] [--mail-from-name=<name>] [--seed-enabled=<true|false>] [--wildcard-domain=<apex-domain>] [--dns-account=<key>] [--skip-dns-check]
```

**Hinweise:**
- Schreibt nach `ansible/hostvars/<domain>.yml`.
- Fuegt standardmaessig `www.<domain>` als Alias hinzu.
- Das eigentliche Deployment startet anschliessend mit `./scripts/eep-redeploy.sh`.
- `--skip-dns-check` ist hilfreich, wenn du die Hostvars lokal vorbereitest, aber DNS noch nicht final auf den Zielhost zeigt.

### eep-redeploy.sh

**Beschreibung:**  
Baut das lokale Easy-Event-Planner-Image und deployed per Ansible entweder eine einzelne Domain oder alle `eep_enabled`-Instanzen.

**Syntax:**
```bash
# Alle EEP-Instanzen deployen
./scripts/eep-redeploy.sh --all

# Einzelne Domain deployen
./scripts/eep-redeploy.sh events.geller.men

# Nur Build
./scripts/eep-redeploy.sh --build-only

# Check-Mode
./scripts/eep-redeploy.sh --all --check-only
```

### redeploy-all-web.sh

**Beschreibung:**  
Massen-Redeploy für alle Web-Container auf Basis vorhandener Hostvars (`ghost_domain_db` / `wp_domain_db` / `eep_enabled` etc.).

**Syntax:**
```bash
# Alle Ghost + WordPress Domains redeployen
./scripts/redeploy-all-web.sh

# Nur prüfen (kein Redeploy)
./scripts/redeploy-all-web.sh --check-only

# Nur Ghost oder nur WordPress
./scripts/redeploy-all-web.sh --only=ghost
./scripts/redeploy-all-web.sh --only=wp

# Nur Easy-Event-Planner
./scripts/redeploy-all-web.sh --only=eep

# Parallelisiert (z. B. 4 gleichzeitige Redeploys)
./scripts/redeploy-all-web.sh --parallel=4

# Bei Fehlern weiterlaufen und am Ende gesammelt reporten
./scripts/redeploy-all-web.sh --continue-on-error
```
