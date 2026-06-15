# Wazuh + CrowdSec Test-Runbook

## Ziel

Dieses Runbook prueft kontrolliert, dass die neuen Mechanismen wirklich arbeiten:

- XML-RPC-Schutz
- CrowdSec Detection + Mail-Alerting
- CrowdSec JSON-Alert-Datei
- Wazuh-Agent
- Wazuh WordPress-FIM-Regeln


## Voraussetzungen

Vor dem Test sollten diese Punkte erledigt sein:

1. `ansible/secrets/secrets.yml` ist aus `ansible/secrets/secrets.example.yml` abgeleitet.
   Alternativ funktioniert jetzt auch `ansible/secrets/secrets.yaml`.
2. `deploy-crowdsec.yml` wurde ausgefuehrt.
3. `deploy-wazuh-agent.yml` wurde ausgefuehrt.
4. Die Manager-Regeln aus `ansible/wazuh/manager/local_rules.xml.example` wurden auf dem Wazuh-Manager eingespielt.


## Test 1: XML-RPC-Schutz

```bash
curl -k -I https://deine-domain.tld/xmlrpc.php
```

Erwartung:

- `401` oder `403`


## Test 2: CrowdSec lebt und sammelt Daten

```bash
docker exec crowdsec cscli metrics
docker exec crowdsec cscli collections list
docker exec crowdsec cscli bouncers list
```

Erwartung:

- Collections wie `crowdsecurity/wordpress`, `crowdsecurity/traefik`, `crowdsecurity/http-cve`
- Bouncer `traefik-bouncer`


## Test 3: Wazuh-Agent verbunden

Auf dem Docker-Host:

```bash
sudo systemctl status wazuh-agent --no-pager
sudo /var/ossec/bin/wazuh-control status
```

Auf dem Wazuh-Manager oder im Dashboard:

- Agent ist verbunden
- Agent-Gruppen wie `infra-hosts`, `docker`, `wordpress` sind sichtbar


## Test 4: Wazuh FIM fuer WordPress-Dateien

Auf dem Docker-Host eine harmlose Testdatei anlegen:

```bash
sudo touch /var/lib/docker/volumes/wp_deine_domain_tld_html/_data/wp-content/uploads/wazuh-test.php
```

Erwartung:

- Wazuh erzeugt einen FIM-Alert
- mit den Beispielregeln sollte der Event als kritisch/WordPress-Webshell-nah auffallen

Danach Testdatei wieder entfernen:

```bash
sudo rm -f /var/lib/docker/volumes/wp_deine_domain_tld_html/_data/wp-content/uploads/wazuh-test.php
```


## Test 5: CrowdSec -> Wazuh Übergabe

Prüfen, ob die Alert-Datei existiert:

```bash
ls -l /home/andy/infra/data/crowdsec/data/notifications/crowdsec_alerts.ndjson
tail -n 5 /home/andy/infra/data/crowdsec/data/notifications/crowdsec_alerts.ndjson
```

Erwartung:

- bei einer CrowdSec-Entscheidung erscheinen JSON-Zeilen mit Feldern wie `scenario`, `value`, `type`


## Test 6: CrowdSec Mail-Alerting

Prüfen:

```bash
docker exec crowdsec cscli notifications inspect email_default
```

Erwartung:

- Plugin `email_default` ist geladen

Praxisnaher Funktionstest:

- einen bewusst begrenzten Angriffstest von einer Test-IP gegen `wp-login.php` oder `xmlrpc.php` fahren
- danach prüfen:
  - CrowdSec Decision vorhanden
  - Mail eingegangen
  - JSON-Zeile in `crowdsec_alerts.ndjson`
  - Wazuh-Event im Manager/Dashboard


## Deployment-Reihenfolge bei Änderungen

Wenn du neue Security-Secrets ergänzt, ist die sichere Reihenfolge:

```bash
ansible-playbook -i ./ansible/inventory/hosts.ini ./ansible/playbooks/deploy-crowdsec.yml
ansible-playbook -i ./ansible/inventory/hosts.ini ./ansible/playbooks/deploy-wazuh-agent.yml
```

Falls du WordPress-Hostvars angepasst hast:

```bash
./scripts/wp-redeploy.sh deine-domain.tld
```


## Was du noch konfigurieren musst

Pflicht fuer kompletten Betrieb:

- `ansible/secrets/secrets.yml` anlegen
- alternativ `ansible/secrets/secrets.yaml`
- SMTP-/SES-Zugang setzen
- `wazuh_enabled: true`
- `wazuh_manager` setzen
- Manager-Regeln auf dem Wazuh-Manager uebernehmen

Optional, aber empfohlen:

- `wazuh_registration_password`
- Wazuh-Manager-Mail-Alerting fuer High-Severity-Events
- spaeter `Traefik`-Access-Logs ebenfalls nach Wazuh geben


## Noch nicht automatisch enthalten

Dieses Repo liefert aktuell **nicht**:

- den kompletten Wazuh-Manager
- die Wazuh-Manager-Mail-Konfiguration
- eine tiefe Auswertung jedes einzelnen Traefik-Requests in Wazuh

Das naechste sinnvolle Ausbaupaket waere deshalb:

1. `Traefik`-Access-Logs explizit fuer Wazuh einsammeln
2. Manager-seitige E-Mail-Regeln fuer FIM/CrowdSec-High-Severity setzen
3. optional aktive Responses im Wazuh-Manager definieren
