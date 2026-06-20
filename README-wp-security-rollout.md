# WordPress Security Rollout & Monitoring

## Ziel

Diese Datei bündelt die WordPress-spezifischen Sicherheitsmechanismen im Stack und beschreibt, wie die Härtung auf alle bestehenden WordPress-Instanzen ausgerollt wird.


## Aktive Schutzmechanismen

### Traefik

- blockiert `xmlrpc.php` vor WordPress
- erzwingt HTTPS
- verarbeitet Primärdomain und Aliase konsistent
- kann pro Bereich zusätzlichen Basic-Auth-Schutz vor WordPress schalten

### CrowdSec

- liest Docker-/Traefik-/Apache-/WordPress-Signale
- nutzt Collections für `traefik`, `apache2`, `wordpress` und `http-cve`
- erstellt Decisions gegen auffällige Quell-IP-Adressen
- versendet Alert-Mails per SMTP/SES
- schreibt JSON-Alerts für Wazuh nach `data/crowdsec/data/notifications/crowdsec_alerts.ndjson`

### Wazuh

- überwacht WordPress-Docker-Volumes auf Host-Ebene
- meldet neue PHP-Dateien in `uploads`
- meldet Änderungen an `wp-config.php`
- meldet Änderungen an `mu-plugins`, Theme- und Plugin-PHP-Dateien
- meldet Massenänderungen im WordPress-Volume
- überwacht zusätzlich Infra-Code und Traefik-/CrowdSec-Konfiguration

### WordPress-Rolle

- setzt `FORCE_SSL_ADMIN`
- setzt standardmäßig `DISALLOW_FILE_EDIT`
- mountet eine PHP-Upload-Konfiguration für größere Theme-/Plugin-ZIPs


## Angriffsszenarien und erwartete Signale

### 1. XML-RPC-Probing oder Brute Force

- Schutz: Traefik blockiert `xmlrpc.php`
- Sichtbarkeit: Traefik/CrowdSec sehen die Requests
- Erwartetes Signal: `403` für den Request, CrowdSec-Entscheidung bei entsprechendem Muster, Mail + Wazuh-CrowdSec-Event

### 2. Brute Force gegen `wp-login.php`

- Schutz: CrowdSec-Collections + optionaler Basic-Auth-Schutz vor dem Admin
- Sichtbarkeit: Traefik-/Apache-/WordPress-Logsignale
- Erwartetes Signal: CrowdSec-Decision, Mail-Alert, JSON-Event für Wazuh

### 3. Upload einer Webshell nach erfolgreicher Anmeldung

- Schutz: kein vollständiger Präventionsschutz auf Dateiebene, aber starke Früherkennung
- Sichtbarkeit: Wazuh FIM auf WordPress-Volumes
- Erwartetes Signal: kritischer Wazuh-Alert bei neuer PHP-Datei in `uploads`

### 4. Austausch von Theme- oder Plugin-Dateien

- Schutz: nur begrenzt präventiv; `DISALLOW_FILE_EDIT` reduziert den Admin-internen Editierpfad
- Sichtbarkeit: Wazuh FIM
- Erwartetes Signal: Wazuh-Alert für Theme-/Plugin-PHP-Änderung

### 5. Manipulation von `wp-config.php`

- Schutz: keine direkte Prävention, aber sehr gute Detektion
- Sichtbarkeit: Wazuh FIM
- Erwartetes Signal: kritischer Wazuh-Alert

### 6. Massenänderung im WordPress-Volume

- Schutz: keine direkte Prävention
- Sichtbarkeit: Wazuh korreliert mehrere Syscheck-Ereignisse
- Erwartetes Signal: Wazuh-Regel für „multiple WordPress file changes“

### 7. Änderung an Infra-Automation oder Reverse-Proxy-Konfiguration

- Schutz: keine direkte Prävention
- Sichtbarkeit: Wazuh überwacht `ansible/hostvars`, `ansible/playbooks`, `scripts`, `data/traefik`, `data/crowdsec/config`
- Erwartetes Signal: Wazuh-Alert auf Infra-Ebene


## Verbleibende Lücken

Das aktuelle Setup ist stark, aber nicht allwissend:

- nicht jeder einzelne Probe-Request erzeugt sofort einen Wazuh-Alert
- ein erfolgreicher Missbrauch rein innerhalb der Datenbank ohne Dateispur ist schwerer sichtbar
- Wazuh wertet derzeit nicht automatisch jeden Traefik-Access-Log-Eintrag als Einzelereignis aus

Deshalb sind weiterhin wichtig:

- regelmäßige Plugin-/Theme-/Core-Updates
- Admin-MFA
- Rotationsprozesse für Passwörter, Salts und Integrations-Secrets
- saubere und getestete Backups


## Rollout auf alle WordPress-Instanzen

### 1. Globale Security-Komponenten aktualisieren

```bash
ansible-playbook -i ./ansible/inventory/hosts.ini ./ansible/playbooks/deploy-crowdsec.yml
ansible-playbook -i ./ansible/inventory/hosts.ini ./ansible/playbooks/deploy-wazuh-agent.yml
```

### 2. WordPress-Hostvars prüfen

```bash
./scripts/wp-rollout-hardening.sh --check-only
```

Das Skript:

- normalisiert `traefik.aliases`
- ergänzt fehlende CrowdSec-Middleware-Defaults
- ergänzt fehlende XML-RPC-Defaults
- kann bestehende Instanzen anschließend gesammelt redeployen

### 3. Härtung auf alle WordPress-Instanzen anwenden

```bash
./scripts/wp-rollout-hardening.sh
```

### 4. Härtung ausrollen und Container neu deployen

```bash
./scripts/wp-rollout-hardening.sh --redeploy
```

Wenn bewusst alte XML-RPC-Ausnahmen zurückgebaut werden sollen:

```bash
./scripts/wp-rollout-hardening.sh --force-xmlrpc-block --redeploy
```


## Validierung nach dem Rollout

Stichproben pro Domain:

```bash
curl -k -i https://deine-domain.tld/xmlrpc.php | sed -n '1,20p'
curl -k -i https://deine-domain.tld/wp-login.php | sed -n '1,20p'
```

Global:

```bash
docker exec crowdsec cscli metrics
docker exec crowdsec cscli decisions list -o human
sudo systemctl status wazuh-agent --no-pager
```

Optionaler Testlauf:

- `README-wazuh-test-runbook.md` Schritt für Schritt abarbeiten


## Was häufig noch übersehen wird

- WordPress-Salts nach einem Incident erneuern
- WooCommerce-/Webhook-/SMTP-/API-Secrets rotieren
- alte Admin- und Editor-Accounts ausmisten
- MFA für Admin-Zugänge aktivieren
- frischen sauberen Backup-Stand nach erfolgreicher Abnahme erstellen
- einen festen Security-Update-Zyklus für Core, Themes und Plugins definieren
