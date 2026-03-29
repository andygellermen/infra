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
- [wildcard-distribute.sh](/Users/andygellermann/Documents/Projects/infra/infra/scripts/wildcard-distribute.sh)
- [distribution.example.yml](/Users/andygellermann/Documents/Projects/infra/infra/ansible/wildcards/distribution.example.yml)

## Secrets
Die AutoDNS-Zugangsdaten liegen unsynced in `ansible/secrets/secrets.yml`:

```yaml
traefik_autodns_user: "tech-user"
traefik_autodns_password: "super-secret"
traefik_autodns_context: "4"
traefik_autodns_endpoint: "https://api.autodns.com/v1/"
traefik_autodns_user_agent: "infra-traefik-wildcard/1.0"
traefik_autodns_http_timeout: "30"
traefik_autodns_polling_interval: "2"
traefik_autodns_propagation_timeout: "180"
traefik_autodns_ttl: "600"
```

## Hostvars
Für Ghost-, WordPress- und Static-Domains kann Wildcard-TLS direkt in den Hostvars aktiviert werden:

```yaml
tls_mode: "wildcard"
tls_wildcard_domain: "example.com"
```

Dann verwenden die Deployments automatisch den Resolver `letsEncryptDns`.

## Add-Skripte
Beim Anlegen neuer Domains kann die Wildcard direkt mitgegeben werden:

```bash
./scripts/ghost-add.sh blog.example.com --wildcard-domain=example.com
./scripts/wp-add.sh app.example.com --wildcard-domain=example.com
./scripts/static-add.sh docs.example.com --wildcard-domain=example.com
```

Bestehende Restore- und Redeploy-Skripte übernehmen die Wildcard-Konfiguration automatisch, sobald sie in den Hostvars steht.

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
```

## Export und Verteilung
Wildcard-Zertifikat aus Traefiks `acme.json` exportieren:

```bash
./scripts/wildcard-export.sh example.com
```

An Staging-Server verteilen:

```bash
./scripts/wildcard-distribute.sh example.com --config ./ansible/wildcards/distribution.example.yml
```

## Wichtig
- der Traefik-Deploy braucht jetzt Zugriff auf `ansible/secrets/secrets.yml`
- Wildcards funktionieren nur mit gesetzten AutoDNS-Zugangsdaten und passender DNS-Berechtigung
