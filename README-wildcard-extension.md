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
- [distribution.example.yml](/Users/andygellermann/Documents/Projects/infra/infra/ansible/wildcards/distribution.example.yml)

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

An Staging-Server verteilen:

```bash
./scripts/wildcard-distribute.sh example.com --config ./ansible/wildcards/distribution.example.yml
```

## Wichtig
- der Traefik-Deploy braucht jetzt Zugriff auf `ansible/secrets/secrets.yml`
- Wildcards funktionieren nur mit gesetzten AutoDNS-Zugangsdaten und passender DNS-Berechtigung
