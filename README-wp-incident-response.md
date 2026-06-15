# WordPress Incident Response Runbook

## Ziel

Dieses Runbook beschreibt den sicheren Ablauf fuer eine kompromittierte WordPress-Instanz im Infra-Stack.
Es ist auf Faelle wie `heimannkunst.de` ausgelegt:

- WordPress war von aussen erreichbar
- die Instanz wurde bereits missbraucht oder manipuliert
- Schadcode kann in PHP-Dateien, Uploads, Themes, Plugins oder in der Datenbank liegen
- die betroffene Website wurde bereits isoliert oder abgeschaltet

Der wichtigste Grundsatz:

- keine "Live-Reparatur" im kompromittierten Bestand
- erst Beweise sichern, dann offline analysieren, dann sauber neu aufbauen


## Kurzfassung

1. Betroffene Instanz isoliert lassen.
2. Forensische Backups erzeugen, bevor irgendetwas geaendert wird.
3. Alte Instanz nicht bereinigen, sondern durch eine frische Instanz ersetzen.
4. Aus dem alten Stand nur gezielt Inhalte uebernehmen.
5. Vor Relaunch technische und inhaltliche Selbsttests fahren.
6. Danach Monitoring, Datei-Integritaet und Alarmierung aktivieren.


## Phase 1: Beweise sichern

### 1.1 Sofortzustand dokumentieren

Auf dem Host:

```bash
docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Image}}'
docker inspect wp-heimannkunst-de > /tmp/wp-heimannkunst-de.inspect.json
docker inspect infra-traefik > /tmp/infra-traefik.inspect.json
docker inspect crowdsec > /tmp/crowdsec.inspect.json
docker logs --tail 500 wp-heimannkunst-de > /tmp/wp-heimannkunst-de.log 2>&1 || true
docker logs --tail 500 infra-traefik > /tmp/infra-traefik.log 2>&1 || true
docker logs --tail 500 crowdsec > /tmp/crowdsec.log 2>&1 || true
docker logs --tail 500 crowdsec-bouncer-traefik > /tmp/crowdsec-bouncer-traefik.log 2>&1 || true
```

Wenn CrowdSec bereits Entscheidungen hat:

```bash
docker exec crowdsec cscli alerts list -o human > /tmp/crowdsec-alerts.txt || true
docker exec crowdsec cscli decisions list -o human > /tmp/crowdsec-decisions.txt || true
```


### 1.2 Forensisches WordPress-Backup erstellen

Ja: `wp-backup.sh` ist hier sinnvoll.
Das Skript erstellt:

- kompletten WordPress-Document-Root
- SQL-Dump der zugehoerigen Datenbank
- Hostvars
- Manifest mit Domain/Container/Version

Empfohlener Befehl:

```bash
./scripts/wp-backup.sh --create heimannkunst.de
```

Wichtig:

- dieses Backup ist ein Beweis-Snapshot des kompromittierten Zustands
- es ist **nicht** automatisch ein "sauberer Restore-Stand"
- das Backup darf danach nicht ueberschrieben oder weiterverarbeitet werden


### 1.3 Optionales Stack-Backup fuer Randspuren

Wenn du Traefik-, CrowdSec- oder Host-Konfiguration mit sichern willst:

```bash
./scripts/infra-backup.sh --create --no-mysql-dump
```

Das ist kein Muss fuer die Reinigung einer einzelnen WordPress-Site, aber hilfreich fuer die spaetere Rekonstruktion des Angriffsbilds.


## Phase 2: Kompromittierten Stand offline untersuchen

### 2.1 Backup in separates Analyse-Verzeichnis entpacken

```bash
mkdir -p /tmp/heimannkunst-forensics
tar xzf ./backups/wordpress/heimannkunst.de/wp-backup-heimannkunst.de-*.tar.gz -C /tmp/heimannkunst-forensics
cd /tmp/heimannkunst-forensics
```


### 2.2 Typische Dateispuren suchen

Verdaechtig sind insbesondere:

- PHP-Dateien in `wp-content/uploads/`
- frisch geaenderte Dateien in Core-, Theme- oder Plugin-Verzeichnissen
- obfuskiertes PHP oder obfuskiertes JavaScript
- unbekannte MU-Plugins
- Webshell-Namen wie `class.wp.php`, `wp-cached.php`, `1.php`, `x.php`, `about.php`

Praktische Checks:

```bash
find . -type f -name '*.php' | sort
find ./wp-content/uploads -type f -name '*.php' | sort
find . -type f \( -name '*.php' -o -name '*.js' \) -mtime -30 | sort
```

Nach typischen Malware-Mustern suchen:

```bash
rg -n "base64_decode|gzinflate|str_rot13|eval\\(|assert\\(|shell_exec|passthru|system\\(|preg_replace\\s*\\(.*/e|create_function|document\\.write|atob\\(|fromCharCode|unescape\\(" .
```

Auf versteckte Loader in Uploads achten:

```bash
find ./wp-content/uploads -type f | sort
```


### 2.3 WordPress-Core und Plugins gegen Original pruefen

Wenn `wp` verfuegbar ist, ist das einer der wertvollsten Checks:

```bash
wp core verify-checksums --path=/tmp/heimannkunst-forensics
wp plugin verify-checksums --all --path=/tmp/heimannkunst-forensics
```

Interpretation:

- Core-Checksum-Fehler sind ein starker Hinweis auf manipulierte Core-Dateien
- Plugin-Checksum-Fehler sind relevant bei Plugins aus dem WordPress.org-Repository
- Premium-Plugins und manuell installierte Plugins haben oft keine offiziellen Checksums und muessen manuell bewertet werden


### 2.4 Datenbank auf Injects und neue Admins pruefen

Den SQL-Dump gezielt durchsuchen:

```bash
rg -n "<script|<iframe|document\\.write|base64|onerror=|onload=|javascript:|atob\\(" db.sql
rg -n "INSERT INTO .*users|administrator|wp_capabilities|siteurl|home" db.sql
```

Wenn die Live-DB noch erreichbar ist, sind diese Abfragen sinnvoll:

```bash
docker exec -it infra-mysql mysql -uroot -p
```

Dann in MySQL:

```sql
SELECT ID, user_login, user_email, user_registered
FROM wp_users
ORDER BY ID;

SELECT user_id, meta_value
FROM wp_usermeta
WHERE meta_key='wp_capabilities';

SELECT option_name, LEFT(option_value, 300)
FROM wp_options
WHERE option_value REGEXP '<script|<iframe|document\\.write|base64|onerror=|onload=|javascript:';

SELECT ID, post_title
FROM wp_posts
WHERE post_content REGEXP '<script|<iframe|document\\.write|base64|onerror=|onload=|javascript:';
```

Zusatzpruefungen:

- unbekannte Administratoren
- manipulierte `siteurl`/`home`
- verdraechtige `active_plugins`
- unbekannte `cron`-Eintraege in `wp_options`
- SEO-Spam, Redirect-JS, fremde Iframes


### 2.5 Was du fuer die spaetere Beurteilung notieren solltest

- welche Datei war der erste klare Treffer
- welche Plugins/Themes waren installiert
- ob Core-Dateien veraendert waren
- ob `wp-content/uploads` PHP-Dateien enthielt
- ob neue Benutzer angelegt wurden
- ob JS nur in Inhalten lag oder auch in Theme/Plugin-Dateien
- ob externe Domains im Schadcode vorkamen


## Phase 3: Nicht bereinigen, sondern sauber neu aufbauen

Empfehlung:

- keine kompromittierte Instanz "gesund putzen" und weiterbetreiben
- stattdessen neue, frische WordPress-Instanz deployen
- Core, Plugins und Themes ausschliesslich neu aus vertrauenswuerdigen Quellen einspielen

Sauberer Weg:

1. neues WordPress deployen
2. DB-Credentials neu erzeugen
3. alle WordPress-Admins neu setzen
4. Salts neu setzen
5. nur gepruefte Inhalte uebernehmen

Wenn dieselbe Domain weiter genutzt wird, ist es oft sinnvoller:

- kompromittiertes Backup nur als Referenz zu behalten
- neue Produktivinstanz separat aufzubauen
- danach Inhalte kontrolliert rueckzufuehren


## Phase 4: Was aus dem alten Bestand uebernommen werden darf

Mit hoher Vorsicht:

- Medien in `wp-content/uploads/`
- Beitraege, Seiten, Kommentare
- eventuell Theme-Assets, aber nur nach manueller Pruefung

Nicht blind uebernehmen:

- gesamte Plugin-Verzeichnisse
- gesamte Theme-Verzeichnisse
- `mu-plugins`
- `wp-config.php`
- fremde PHP-Dateien in Uploads
- komplette DB ohne vorherige Pruefung


## Phase 5: Neuinstanz absichern, bevor sie wieder online geht

Pflichtpunkte:

- `xmlrpc.php` bleibt geblockt
- alle WordPress-, DB- und API-Passwoerter rotieren
- WordPress-Salts erneuern
- unnoetige Plugins entfernen
- Themes/Plugins/WordPress auf aktuellen Stand bringen
- `wp-admin` optional zusaetzlich mit Basic Auth schuetzen
- nur benoetigte Admin-Benutzer behalten

Wenn `wp` verfuegbar ist:

```bash
wp config shuffle-salts --path=/var/www/html
```


## Phase 6: Relaunch-Checkliste

Vor der Wiederfreigabe:

```bash
curl -k -I https://heimannkunst.de/
curl -k -I https://heimannkunst.de/xmlrpc.php
curl -k -I https://heimannkunst.de/wp-login.php
```

Erwartung:

- Startseite antwortet sinnvoll
- `xmlrpc.php` liefert `403` oder `401`
- kein Redirect auf fremde Domains
- keine fremden Skripte im HTML-Quelltext

Zusaetzlich:

- Browser-Test auf Startseite, Unterseiten, Kontaktformular, Medien
- Quelltext auf fremde JS-Domains pruefen
- Search Console / SEO-Spam / Redirects im Blick behalten


## Monitoring-Empfehlung fuer den Dauerbetrieb

### Empfehlung fuer euren Stack

Am sinnvollsten ist aus meiner Sicht:

- `CrowdSec` fuer Edge-/IP-Abwehr und HTTP-Angriffsdetektion
- `Wazuh` auf dem Docker-Host fuer Datei-Integritaet, Host-Logs, Docker-Events und Alarmierung per E-Mail

Warum Wazuh hier gut passt:

- erkennt geaenderte oder ersetzte Dateien
- sammelt und analysiert Logs
- kann E-Mail-Alerts senden
- kann Rootkit-/Anomalie-Scans fahren
- hat Docker-spezifische Monitoring-Funktionen

Besonders ueberwachen:

- `/home/andy/infra`
- `/home/andy/infra/data/traefik`
- `/home/andy/infra/data/crowdsec`
- Docker-Compose-/Ansible-/Script-Dateien
- `/var/lib/docker/volumes/wp_*`
- `/etc`
- `journald`
- Docker-Daemon-Logs


### Mindest-Alarmregeln

Eskalation per E-Mail bei:

- Aenderung an `wp-config.php`
- neue `.php`-Datei unter `wp-content/uploads`
- Core-Datei geaendert
- Plugin-/Theme-Datei ausserhalb eines Wartungsfensters geaendert
- neue WordPress-Admin-Benutzer
- ploetzliche Masse an `wp-login.php`- oder `xmlrpc.php`-Requests
- CrowdSec-Alerts/Decisions
- Rootcheck/FIM-Treffer auf dem Host


### Wenn Wazuh zu schwergewichtig ist

Leichtere Alternative:

- `osquery` fuer FIM und Host-Queries
- dazu `CrowdSec` fuer Edge-Detection

Aber:

- Wazuh ist fuer euren Wunsch "Datei-Integritaet + Log-Verstaendnis + E-Mail-Eskalation" deutlich runder
- osquery ist eher ein Baukasten als ein komplettes SOC-/Alerting-Werkzeug


## Praktische Entscheidung fuer `heimannkunst.de`

Wenn die Instanz bereits kompromittiert war, ist der beste Weg in der Praxis:

1. `wp-backup.sh` sofort als Beweis-Snapshot fahren
2. Log- und Docker-Metadaten sichern
3. altes Backup offline untersuchen
4. frische WordPress-Instanz neu aufbauen
5. nur gepruefte Inhalte uebernehmen
6. Monitoring und Alarmierung vor Wiederfreigabe aktivieren


## Nicht tun

- keinen Plugin-"Cleaner" blind ueber die kompromittierte Instanz laufen lassen
- keine kompromittierte Instanz direkt wieder online nehmen
- keine alten PHP-Dateien blind in die neue Instanz kopieren
- keinen Restore aus dem kompromittierten Backup als "saubere Basis" verwenden
- keine Zugangsdaten unveraendert lassen

