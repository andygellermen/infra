# Wazuh Manager Regelstrategie fuer Infra + WordPress

## Zweck

Diese Dateien gehoeren **auf den Wazuh-Manager**, nicht auf den Agenten.
Der Agent auf dem Docker-Host sammelt Events und sendet sie an den Manager.
Der Manager bewertet sie mit Regeln, erzeugt Alerts und kann sie per E-Mail eskalieren.


## Dateien

- `local_rules.xml.example`


## Installation auf dem Wazuh-Manager

1. Datei nach `/var/ossec/etc/rules/local_rules.xml` uebernehmen oder in eure bestehende `local_rules.xml` integrieren.
2. Syntax und Treffer mit `wazuh-logtest` pruefen.
3. Wazuh-Manager neu starten.

Beispiel:

```bash
sudo cp local_rules.xml.example /var/ossec/etc/rules/local_rules.xml
sudo /var/ossec/bin/wazuh-logtest
sudo systemctl restart wazuh-manager
```


## Was die Regeln abdecken

### CrowdSec-Ereignisse

Wenn CrowdSec einen Angriff erkennt und eine Entscheidung erzeugt:

- sendet CrowdSec direkt eine SMTP-Mail
- schreibt CrowdSec ein JSON-Event in `crowdsec_alerts.ndjson`
- der Wazuh-Agent liest dieses JSON ein
- der Wazuh-Manager erzeugt daraus einen Alert

Die Beispielregeln heben insbesondere diese Szenarien hervor:

- `wordpress`
- `xmlrpc`
- `wp-login`
- `http-cve`
- `http-probing`


### Datei-Integritaet in WordPress-Volumes

Die Beispielregeln eskalieren speziell:

- Aenderungen an `wp-config.php`
- neue PHP-Dateien in `wp-content/uploads`
- Aenderungen an `mu-plugins`
- PHP-Aenderungen in Themes und Plugins
- viele WordPress-Dateiaenderungen in kurzer Zeit


### Infra-Konfigurationsaenderungen

Zusatzregel fuer:

- `ansible/hostvars`
- `ansible/playbooks`
- `scripts`
- `data/traefik`
- `data/crowdsec/config`


## Typischer Ablauf bei einem WordPress-Angriff

### 1. Brute Force / XML-RPC / HTTP-Probing

- Traefik nimmt den Request an
- CrowdSec verarbeitet die Docker-/Traefik-/WordPress-Logs
- bei Treffer blockt der Bouncer die Quelle
- CrowdSec mailt sofort
- Wazuh bekommt den CrowdSec-Alert als JSON und archiviert/korreliert ihn zentral

### 2. Webshell oder schadhafte PHP-Datei

- Angreifer schreibt eine PHP-Datei in `uploads`, `mu-plugins`, Theme oder Plugin
- Wazuh Syscheck bemerkt die Datei-Aenderung auf dem Host in Echtzeit
- der Manager erzeugt einen hochpriorisierten FIM-Alert

### 3. Manipulation an `wp-config.php`

- Wazuh Syscheck erkennt die Aenderung
- die lokale Beispielregel stuft das als kritisch ein


## Wichtige Grenze des aktuellen Setups

Wenn ein Angreifer **nur einzelne HTTP-Requests** absetzt, die **noch keine CrowdSec-Entscheidung** erzeugen und **keine Datei aendern**, dann seht ihr das aktuell primaer ueber CrowdSec selbst — nicht ueber detaillierte Wazuh-Webrequest-Regeln.

Wenn ihr auch **jeden einzelnen verdaechtigen Webrequest** zentral in Wazuh korrelieren wollt, ist der naechste sinnvolle Schritt:

- Traefik-Access-Logs in Datei oder journald zentral einsammeln
- diese Logs zusaetzlich in Wazuh auswerten


## Testen

Wazuh empfiehlt zum Testen eigener Regeln `wazuh-logtest`.

Prueft damit mindestens:

- ein Beispiel-CrowdSec-JSON-Event
- einen Beispiel-Syscheck-Event mit `wp-config.php`
- einen Beispiel-Syscheck-Event mit `wp-content/uploads/shell.php`
