# Redirect Erweiterung

Für einfache Domain-Weiterleitungen gibt es eine zentrale Redirect-Konfiguration unter [redirects.yml](/Users/andygellermann/Documents/Projects/infra/infra/ansible/redirects/redirects.yml).

## Modell
- ein gemeinsamer Container: `domain-redirects`
- Traefik-Router pro Redirect-Domain
- Weiterleitung per `301` oder optional temporär
- TLS für die Redirect-Domain automatisch über `letsEncrypt`
- eigener Traefik-Service auf Port `80`, damit Redirects zuverlässig statt `404` greifen

## Konfiguration
Beispiel:

```yaml
redirects:
  - source: md.mf2go.de
    target: mietdachboxen.de
    permanent: true

  - source: mf.mf2go.de
    target: mietfirma.de
    permanent: true
```

Optional:
- `target_scheme`: Standard ist `https`
- `permanent`: Standard ist `true`

## Redeploy

```bash
./scripts/redirect-redeploy.sh
```

Check-only:

```bash
./scripts/redirect-redeploy.sh --check-only
```

## Selbsttest
Nach dem Redeploy prüft das Script automatisch:
- DNS der Redirect-Quellen zeigt auf den aktuellen Server
- Redirect-Quelle liefert den erwarteten Redirect-Status
- `Location` zeigt auf das erwartete Ziel
- der Check wartet nach dem Deploy kurz auf Traefik-/Label-Übernahme, um Fehlalarme direkt nach dem Rollout zu vermeiden

## Integration
- `redeploy-all-web.sh` unterstützt jetzt zusätzlich `--only=redirect`
- bei `--only=all` werden die Redirects ebenfalls mitredeployed
