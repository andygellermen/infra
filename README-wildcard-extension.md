# Wildcard Erweiterung

Diese Erweiterung bereitet den Stack für automatisch erneuerbare Wildcard-Zertifikate mit Traefik per DNS-01 vor.

## Modell
- Traefik behält den bisherigen Resolver `letsEncrypt` für normale HTTP-01-Zertifikate
- zusätzlich gibt es `letsEncryptDns` für Wildcard-Zertifikate per DNS-01
- für InterNetX/Schlundtech wird der native `autodns`-Provider von Traefik/lego verwendet
- dadurch ist kein eigenes DNS-Hook-Script mehr nötig

## Wichtige Dateien
- [deploy-traefik.yml](/Users/andygellermann/Documents/Projects/infra/infra/ansible/playbooks/deploy-traefik.yml)
- [main.yml](/Users/andygellermann/Documents/Projects/infra/infra/ansible/playbooks/roles/traefik/tasks/main.yml)
- [wildcard-export.sh](/Users/andygellermann/Documents/Projects/infra/infra/scripts/wildcard-export.sh)
- [wildcard-housekeeping.sh](/Users/andygellermann/Documents/Projects/infra/infra/scripts/wildcard-housekeeping.sh)
- [wildcard-distribute.sh](/Users/andygellermann/Documents/Projects/infra/infra/scripts/wildcard-distribute.sh)
- [wildcard-distribute-on-change.sh](/Users/andygellermann/Documents/Projects/infra/infra/scripts/wildcard-distribute-on-change.sh)
- [export.example.yml](/Users/andygellermann/Documents/Projects/infra/infra/ansible/wildcards/export.example.yml)
- [distribution.example.yml](/Users/andygellermann/Documents/Projects/infra/infra/ansible/wildcards/distribution.example.yml)
- [wildcard-distribute-on-change.service](/Users/andygellermann/Documents/Projects/infra/infra/ansible/wildcards/systemd/wildcard-distribute-on-change.service)
- [wildcard-distribute-on-change.path](/Users/andygellermann/Documents/Projects/infra/infra/ansible/wildcards/systemd/wildcard-distribute-on-change.path)

## Secrets
Die AutoDNS-Zugangsdaten liegen unsynced in `ansible/secrets/secrets.yml`.

Ein einzelner Legacy-/Fallback-Account bleibt möglich:

```yaml
traefik_autodns_user: "tech-user"
traefik_autodns_password: "super-secret"
traefik_autodns_context: "10"
traefik_autodns_endpoint: "https://api.autodns.com/v1/"
traefik_autodns_user_agent: "infra-traefik-wildcard/1.0"
traefik_autodns_http_timeout: "30"
traefik_autodns_polling_interval: "2"
traefik_autodns_propagation_timeout: "180"
traefik_autodns_ttl: "600"
```

Für mehrere betreute Schlundtech-/AutoDNS-Accounts ist diese Struktur empfohlen:

```yaml
traefik_autodns_default_account: "account-a"

traefik_autodns_accounts:
  account-a:
    user: "tech-user-a"
    password: "secret-a"
    context: "10"
    endpoint: "https://api.autodns.com/v1/"
    user_agent: "infra-traefik-wildcard/1.0"
    http_timeout: "30"
    polling_interval: "2"
    propagation_timeout: "180"
    ttl: "600"

  account-b:
    user: "tech-user-b"
    password: "secret-b"
    context: "10"
```

Wichtig:
- pro Wildcard-Nutzung kann derselbe Traefik-Prozess aktuell nur genau einen `tls_dns_account` gleichzeitig aktiv bedienen
- wenn mehrere verschiedene Wildcard-Accounts parallel aktiv wären, bricht der Traefik-Deploy mit einer klaren Fehlermeldung ab
- laut InternetX-/Schlundtech-Rückmeldung läuft unser Live-Account mit dem persönlichen AutoDNS-Kontext `10`, nicht mit dem generischen Produktionswert `4`

## Hostvars
Für Ghost-, WordPress- und Static-Domains kann Wildcard-TLS direkt in den Hostvars aktiviert werden:

```yaml
tls_mode: "wildcard"
tls_wildcard_domain: "example.com"
tls_dns_account: "account-a"
```

Dann verwenden die Deployments automatisch den Resolver `letsEncryptDns`.

## Add-Skripte
Beim Anlegen neuer Domains kann die Wildcard direkt mitgegeben werden:

```bash
./scripts/ghost-add.sh blog.example.com --wildcard-domain=example.com --dns-account=account-a
./scripts/wp-add.sh app.example.com --wildcard-domain=example.com --dns-account=account-a
./scripts/static-add.sh docs.example.com --wildcard-domain=example.com --dns-account=account-a
```

Bestehende Restore- und Redeploy-Skripte übernehmen die Wildcard-Konfiguration automatisch, sobald sie in den Hostvars steht.
Die Restore-Skripte können sie jetzt auch direkt setzen:

```bash
./scripts/wp-restore.sh app.example.com backup.zip --wildcard-domain=example.com --dns-account=account-a
./scripts/static-restore.sh docs.example.com backup.zip --wildcard-domain=example.com --dns-account=account-a
./scripts/ghost-restore.sh blog.example.com backup.zip --wildcard-domain=example.com --dns-account=account-a --redeploy
```

## Redirects
Redirect-Einträge unterstützen optional ebenfalls:

```yaml
redirects:
  - source: app.example.com
    aliases:
      - www.app.example.com
    target: www.example.com
    permanent: true
    tls_mode: wildcard
    tls_wildcard_domain: example.com
    tls_dns_account: account-a
```

## Export und Verteilung
Wildcard-Zertifikat aus Traefiks `acme.json` exportieren:

```bash
./scripts/wildcard-export.sh example.com
```

Bestehende Einzelzertifikate unterhalb derselben Apex-Domain lassen sich nach erfolgreicher Wildcard-Ausstellung sicher bereinigen:

```bash
./scripts/wildcard-housekeeping.sh example.com --restart-traefik
```

Dabei gilt:
- das Skript bereinigt nur Zertifikate, deren Namen vollständig durch das tatsächlich exportierte Wildcard-Zertifikat gedeckt sind
- Aliase außerhalb der Wildcard-Apex-Domain, z. B. fremde Punycode-Domains, bleiben bewusst erhalten und laufen separat weiter
- der Traefik-Deploy führt dieses Housekeeping inzwischen automatisch für konfigurierte Wildcard-Apex-Domains aus, sobald das Wildcard-Zertifikat bereits existiert

Für die Verteilung braucht der Zielserver keine zusätzlichen Hostvars im Infra-Stack, solange dort nur Zertifikatsdateien abgelegt und lokal weiterverarbeitet werden. Die Ausstellung und Erneuerung bleiben auf dem zentralen Traefik-Host; im Repo pflegen wir nur die Export-/Verteilkonfiguration.

Zentrale Konfiguration:

```yaml
wildcard_exports:
  - source_domain: example.com
    acme_file: /home/andy/infra/data/traefik/acme/acme.json
    targets:
      - name: staging-nginx-1
        host: staging-1.example.net
        user: root
        identity_file: /home/andy/.ssh/id_wildcard_staging
        remote_fullchain_path: /etc/nginx/ssl/example.com/fullchain.pem
        remote_privkey_path: /etc/nginx/ssl/example.com/privkey.pem
        post_deploy_command: "systemctl reload nginx"
```

Wichtige Felder:
- `source_domain`: welche Wildcard aus `acme.json` exportiert werden soll
- `acme_file`: optionaler Pfad zur Quell-`acme.json`
- `remote_fullchain_path` und `remote_privkey_path`: exakte Zielpfade auf dem Server
- `remote_dir`: Kurzform, falls beide Dateien klassisch als `fullchain.pem` und `privkey.pem` in einem Verzeichnis landen sollen
- `identity_file`: optionaler separater SSH-Key pro Zielserver
- `post_deploy_command`: optionaler lokaler Hook auf dem Zielserver, z. B. für `nginx`-/`caddy`-Reload oder ein eigenes Install-Skript

Gezielt an eine Apex-Domain verteilen:

```bash
./scripts/wildcard-distribute.sh example.com --config ./ansible/wildcards/export.example.yml
```

Alle Einträge aus der zentralen Datei abarbeiten, z. B. für Cron:

```bash
./scripts/wildcard-distribute.sh --all --config ./ansible/wildcards/export.example.yml
```

Nur den Ablauf prüfen, ohne Dateien zu exportieren oder per SSH zu verteilen:

```bash
./scripts/wildcard-distribute.sh --all --config ./ansible/wildcards/export.example.yml --dry-run
```

Cron-Beispiel:

```cron
17 3 * * * /bin/sh -lc 'cd /home/andy/infra && ./scripts/wildcard-distribute.sh --all --config ./ansible/wildcards/export.yml >> /home/andy/infra/logs/wildcard-distribute.log 2>&1'
```

## Erneuerungsgekoppelte Verteilung
Damit die Zielserver nur nach einer echten Zertifikatserneuerung beliefert werden, gibt es zusätzlich einen Change-Trigger:

```bash
./scripts/wildcard-distribute-on-change.sh --all --config ./ansible/wildcards/export.yml
```

Das Skript arbeitet so:
- liest alle konfigurierten `source_domain`-Einträge aus `export.yml`
- exportiert pro Wildcard das aktuelle Zertifikat aus `acme.json`
- berechnet einen SHA-256-Fingerprint des Leaf-Zertifikats
- vergleicht ihn mit einer State-Datei unter `data/wildcard-distribution-state/`
- ruft nur bei echter Änderung das Verteilskript auf

Vorteil:
- keine tägliche Blind-Verteilung
- keine unnötigen SSH-Transfers
- keine überflüssigen Reloads auf den Zielservern
- trotzdem sofortige Reaktion, sobald Traefik das Zertifikat erneuert hat

Manuell testen:

```bash
./scripts/wildcard-distribute-on-change.sh --all --config ./ansible/wildcards/export.yml --dry-run
```

Erzwungene Verteilung trotz identischem Fingerprint:

```bash
./scripts/wildcard-distribute-on-change.sh --all --config ./ansible/wildcards/export.yml --force
```

## Systemd Watcher
Für den produktiven Betrieb ist `systemd.path` die sauberste Kopplung:

1. `acme.json` wird von `wildcard-distribute-on-change.path` überwacht
2. bei Änderung startet `wildcard-distribute-on-change.service`
3. der Service prüft den Fingerprint und verteilt nur bei echter Änderung

Beispiel-Dateien:
- [wildcard-distribute-on-change.service](/Users/andygellermann/Documents/Projects/infra/infra/ansible/wildcards/systemd/wildcard-distribute-on-change.service)
- [wildcard-distribute-on-change.path](/Users/andygellermann/Documents/Projects/infra/infra/ansible/wildcards/systemd/wildcard-distribute-on-change.path)

Installationsbeispiel auf dem Traefik-/ACME-Host:

```bash
sudo cp ./ansible/wildcards/systemd/wildcard-distribute-on-change.service /etc/systemd/system/
sudo cp ./ansible/wildcards/systemd/wildcard-distribute-on-change.path /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now wildcard-distribute-on-change.path
```

Optionaler Sicherheitsgurt als Fallback-Cron, z. B. einmal täglich:

```cron
23 4 * * * /bin/sh -lc 'cd /home/andy/infra && ./scripts/wildcard-distribute-on-change.sh --all --config ./ansible/wildcards/export.yml >> /home/andy/infra/logs/wildcard-distribute-on-change.log 2>&1'
```

Auch dieser Fallback verteilt nur bei geändertem Fingerprint.

## Wichtig
- der Traefik-Deploy braucht jetzt Zugriff auf `ansible/secrets/secrets.yml`
- Wildcards funktionieren nur mit gesetzten AutoDNS-Zugangsdaten und passender DNS-Berechtigung
