# Wazuh Agent + CrowdSec Alerting

## Ziel

Diese Erweiterung ergänzt den Infra-Stack um zwei Bausteine:

- einen **Wazuh-Agenten auf dem Docker-Host** für Datei-Integrität, Host-Logs, Rootcheck und SCA
- **CrowdSec-Mail-Alerting** sowie eine **CrowdSec-Alert-Datei** für die spätere Wazuh-/SIEM-Aufnahme

Wichtig:

- diese Repo-Erweiterung deployt **den Wazuh-Agenten**, nicht den kompletten Wazuh-Manager
- E-Mail-Alerts aus Wazuh selbst werden **am Wazuh-Manager** konfiguriert, nicht auf dem Agenten
- die direkte Sofort-Eskalation im Stack übernimmt hier zunächst **CrowdSec per SMTP-Mail**

Ergänzende Dokumente:

- `README-wp-security-rollout.md` für die flächige Härtung aller WordPress-Instanzen
- `README-wp-mfa-plan.md` für die zentrale MFA-Strategie vor WordPress-Admin-Routen
- `wordpress-incident-readme.md` für die konkrete Incident-Dokumentation von `heimannkunst.de`


## Neue Playbooks

- `ansible/playbooks/deploy-wazuh-agent.yml`
- `ansible/playbooks/deploy-crowdsec.yml` nutzt jetzt zusätzlich `ansible/secrets/secrets.yml` für Mail-/Alerting-Variablen


## Aktivierung

In `ansible/secrets/secrets.yml` oder `ansible/secrets/secrets.yaml`:

```yaml
wazuh_enabled: true
wazuh_manager: "wazuh.example.org"
wazuh_manager_port: 1514
wazuh_registration_port: 1515
wazuh_registration_password: "bitte-stark-waehlen"

# optional
wazuh_agent_name: "infra-prod-01"
wazuh_agent_groups:
  - infra-hosts
  - docker
  - wordpress
```

Deploy:

```bash
ansible-playbook -i ./ansible/inventory/hosts.ini ./ansible/playbooks/deploy-wazuh-agent.yml
```

Hinweis:

- kanonisch im Repo bleibt `ansible/secrets/secrets.yml`
- die Loader akzeptieren jetzt aber auch `ansible/secrets/secrets.yaml`


## Aktivierungsmatrix

### 1. XML-RPC-Schutz auf WordPress

Läuft, wenn:

- die WordPress-Instanz mit der aktualisierten Rolle deployt oder redeployed wurde
- in den Hostvars `wp_xmlrpc_protection` nicht auf `off` steht

Pflicht-Config:

- keine zusätzliche Secret-Config nötig

Prüfen:

```bash
curl -k -I https://deine-domain.tld/xmlrpc.php
```

Erwartung:

- `401` oder `403`


### 2. CrowdSec Log-Auswertung + Bouncer

Läuft, wenn:

- `deploy-crowdsec.yml` ausgeführt wurde
- Traefik und Web-Container mit den CrowdSec-Labels laufen

Pflicht-Config:

- keine zusätzliche Secret-Config für die Basiserkennung nötig

Prüfen:

```bash
docker exec crowdsec cscli metrics
docker exec crowdsec cscli collections list
docker ps --format 'table {{.Names}}\t{{.Status}}' | grep crowdsec
```


### 3. CrowdSec Mail-Alerting

Läuft, wenn:

- SMTP-/SES-Daten in `ansible/secrets/secrets.yml` oder `ansible/secrets/secrets.yaml` gesetzt sind
- mindestens ein Empfänger gesetzt ist
- `deploy-crowdsec.yml` danach erneut ausgeführt wurde

Pflicht-Config:

- `ses_smtp_user`
- `ses_smtp_password`
- `ses_smtp_host`
- `ses_smtp_port`
- `ses_from`
- `crowdsec_notification_email_to` oder alternativ `infra_error_notify_to`

Prüfen:

```bash
docker exec crowdsec cscli notifications inspect email_default
```


### 4. CrowdSec JSON-Alert-Datei für Wazuh

Läuft, wenn:

- `wazuh_enabled: true` gesetzt ist
- `deploy-crowdsec.yml` erneut ausgeführt wurde

Pflicht-Config:

- `wazuh_enabled: true`

Prüfen:

```bash
ls -l /home/andy/infra/data/crowdsec/data/notifications/crowdsec_alerts.ndjson
```


### 5. Wazuh-Agent auf dem Docker-Host

Läuft, wenn:

- `wazuh_enabled: true` gesetzt ist
- `wazuh_manager` gesetzt ist
- `deploy-wazuh-agent.yml` erfolgreich gelaufen ist

Pflicht-Config:

- `wazuh_enabled: true`
- `wazuh_manager`
- optional aber stark empfohlen: `wazuh_registration_password`

Prüfen:

```bash
sudo systemctl status wazuh-agent
sudo /var/ossec/bin/wazuh-control status
```


### 6. Wazuh Alerting für WordPress-/CrowdSec-Ereignisse

Läuft, wenn:

- die Beispielregeln auf dem Wazuh-Manager eingespielt wurden
- der Wazuh-Manager neu gestartet wurde

Pflicht-Config:

- `ansible/wazuh/manager/local_rules.xml.example` auf dem Manager übernehmen

Prüfen:

```bash
sudo /var/ossec/bin/wazuh-logtest
sudo systemctl restart wazuh-manager
```


## CrowdSec Mail-Alerting

Die CrowdSec-Rolle verwendet, wenn vorhanden, automatisch eure bestehenden SES-/SMTP-Secrets:

```yaml
ses_smtp_user: "AKIA...SMTP_USER"
ses_smtp_password: "SMTP-PASSWORT"
ses_smtp_host: "email-smtp.eu-central-1.amazonaws.com"
ses_smtp_port: 587
ses_from: "Infra Bot <noreply@example.org>"

# optional explizit fuer CrowdSec
crowdsec_notification_email_to:
  - "security@example.org"
  - "ops@example.org"
```

Wenn `crowdsec_notification_email_to` nicht gesetzt ist, faellt die Rolle auf `infra_error_notify_to` zurueck.

Fuer monatliche Sammelreports koennen zusaetzlich gesetzt werden:

```yaml
wp_update_report_to: "ops@example.org"
security_digest_report_to: "security@example.org"
```


## Was der Wazuh-Agent ueberwacht

Standardmaessig:

- Systempfade: `/etc`, `/usr/bin`, `/usr/sbin`, `/bin`, `/sbin`
- Infra-Code: `/home/andy/infra/ansible/hostvars`, `/home/andy/infra/ansible/playbooks`, `/home/andy/infra/scripts`
- Traefik-/CrowdSec-Konfiguration: `/home/andy/infra/data/traefik`, `/home/andy/infra/data/crowdsec/config`
- WordPress-Volumes ueber Wildcard-Muster unter `/var/lib/docker/volumes/wp_*_html/_data`

Besonderer Fokus bei WordPress:

- `wp-config.php`
- PHP-Dateien in `mu-plugins`
- PHP-Dateien in `uploads`
- PHP-Dateien in Themes und Plugins

Ausgenommen bzw. ohne Inhalts-Diff:

- `ansible/secrets`
- `data/traefik/acme/acme.json`
- `data/crowdsec/data`


## CrowdSec -> Datei fuer Wazuh

Wenn `wazuh_enabled: true` gesetzt ist, erzeugt CrowdSec zusaetzlich:

- Host-Datei: `/home/andy/infra/data/crowdsec/data/notifications/crowdsec_alerts.ndjson`

Diese Datei wird vom Wazuh-Agenten als JSON-Logquelle eingesammelt.
So bekommt ihr:

- IP-basierte Angriffs- und Remediation-Events direkt aus CrowdSec
- Host-/Datei-Integritaets-Events direkt aus Wazuh


## Was im Angriffsfall passiert

### Einbruchversuch auf `xmlrpc.php` oder `wp-login.php`

- Traefik nimmt den Request an
- CrowdSec wertet Traefik-/WordPress-Logs aus
- bei erkannten Mustern blockt der Bouncer die Quell-IP
- CrowdSec sendet sofort eine SMTP-Mail
- CrowdSec schreibt zusaetzlich ein JSON-Event fuer Wazuh
- Wazuh archiviert und korreliert das Ereignis zentral auf dem Manager

### Erfolgreicher Dateidrop oder Manipulation im WordPress-Volume

- Wazuh Syscheck ueberwacht die WordPress-Volumes direkt auf Host-Ebene
- neue PHP-Dateien in `uploads`, Aenderungen an `wp-config.php`, `mu-plugins`, Themes oder Plugins werden als FIM-Event gemeldet
- auf dem Manager koennt ihr diese Events per `local_rules.xml` hochstufen und per Mail eskalieren

### Aenderung an Infra-Konfiguration

- Aenderungen an `ansible/hostvars`, `ansible/playbooks`, `scripts`, `data/traefik` und `data/crowdsec/config` werden ebenfalls vom Agenten erkannt

Grenze des aktuellen Setups:

- einzelne verdaechtige HTTP-Requests ohne CrowdSec-Decision und ohne Dateiaenderung seht ihr aktuell nicht als eigene Wazuh-Webalert-Regel
- dafuer ist der naechste Ausbau `Traefik`-Access-Logs -> `Wazuh`


## Infra-Skripte

Neu:

- `scripts/infra-update-all.sh` kennt jetzt `--skip-wazuh`
- `scripts/infra-setup.sh` deployt den Wazuh-Agenten automatisch mit, wenn `wazuh_enabled: true` gesetzt ist


## Wichtiger Betriebs-Hinweis

Wazuh empfiehlt, Agent und Manager versionskompatibel zu halten; der Agent sollte nicht neuer als der Manager sein.
Deshalb setzt die Rolle das Debian-Paket `wazuh-agent` standardmaessig auf `hold`.


## Zusatzdateien im Repo

- Beispiel-Secrets: `ansible/secrets/secrets.example.yml`
- Beispiel-Manager-Regeln: `ansible/wazuh/manager/local_rules.xml.example`
- Manager-Hinweise: `ansible/wazuh/manager/README.md`
- Test-Runbook: `README-wazuh-test-runbook.md`
