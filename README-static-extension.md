# Static Sites Erweiterung (Infra Stack Static-Extension / Ziel: v1.5.0)

## Zielbild
Für kleine reine HTML-Seiten wird eine gemeinsame, leichtgewichtige Nginx-Instanz betrieben:

- ein gemeinsamer Container: `static-sites`
- mehrere Domains über Traefik-Router auf denselben Container
- Inhalte direkt auf dem Host unter `/srv/static/<domain>/`
- Bearbeitung per SSH/SFTP direkt im Host-Dateisystem

Das ist bewusst einfacher und robuster als WordPress:

- keine Datenbank
- keine App-Logik
- keine `WP_HOME`-/Proxy-Probleme
- Redirects liegen klar bei Traefik, Zugriffsschutz für Unterpfade bei Nginx

## Verzeichnisstruktur auf dem Host
- Inhalte einer Site: `/srv/static/<domain>/`
- Passwortdateien: `/srv/static-auth/`
- generierte Nginx-Konfiguration: `/srv/static-nginx/conf.d/sites.conf`

Damit lassen sich Inhalte bequem per SSH/SFTP pflegen, ohne im Container arbeiten zu müssen.

## Skripte
- `scripts/static-add.sh`: legt Hostvars an und deployt oder aktualisiert den Shared-Container
- `scripts/static-backup.sh`: erstellt ein Backup des statischen Document-Roots inkl. optionaler Hostvars-Metadaten
- `scripts/static-redeploy.sh`: DNS-Check und Redeploy der Shared-Static-Instanz, optional auch gesammelt via `--all`
- `scripts/static-delete.sh`: entfernt die Domain aus den Hostvars und deployt die Shared-Instanz neu
- `scripts/static-restore.sh`: stellt eine statische Site aus `.tar.gz`, `.tgz` oder `.zip` wieder her und führt danach einen HTTPS-Selbsttest aus

## Backup einer statischen Site (neu in v1.5.0)
Aufruf:

```bash
./scripts/static-backup.sh --create domain.de
```

Optional:

```bash
./scripts/static-backup.sh --create domain.de --output /pfad/zum/backup.tar.gz
```

Das Backup enthält:
- den vollständigen statischen Document-Root
- optional `_infra/hostvars.yml`, falls Hostvars vorhanden sind
- ein kleines `_infra/manifest.env`

## Restore einer statischen Site (neu in v1.4.0)
Aufruf:

```bash
./scripts/static-restore.sh domain.de backup.zip
```

Optional:

```bash
./scripts/static-restore.sh domain.de backup.zip --restore-hostvars
```

Verhalten:
- erkennt automatisch den statischen Document-Root im Archiv
- schreibt die Inhalte nach `/srv/static/<domain>/`
- erzeugt bei Bedarf minimale Hostvars für die statische Site
- kann optional Hostvars aus dem Backup übernehmen
- führt danach einen Shared-Static-Redeploy aus
- prüft abschließend die öffentliche HTTPS-Erreichbarkeit

Das Script ist bewusst einfacher als `wp-restore.sh`, weil bei reinen HTML-Sites weder Datenbank noch App-Migrationslogik nötig sind.

## Hostvars-Beispiel
```yaml
domain: example.com

traefik:
  domain: example.com
  aliases:
    - www.example.com

static_enabled: true
static_traefik_middleware_default: "crowdsec-default@docker"
static_basic_auth_paths:
  - path: "/private-folder/"
    realm: "Protected Area"
    auth_file: "/srv/static-auth/example.com-private-folder.htpasswd"
```

## Passwortgeschützte Unterverzeichnisse
Ein privater Bereich wie `/private-folder/` wird direkt in Nginx über Basic Auth geschützt.

Vorteile:
- stabil und bewährt
- kein Konflikt mit Traefik nötig
- Schutz nur für den gewünschten Unterpfad

Wichtig:
- die `.htpasswd`-Datei liegt **nicht** im Webroot
- sie muss vor dem Deploy existieren

Beispiel zum Anlegen:
```bash
sudo apt install apache2-utils
sudo htpasswd -cB /srv/static-auth/example.com-private-folder.htpasswd andy
```

## DNS und Redirects
- `http -> https` erfolgt über Traefik
- Alias-Domains wie `www.example.com` werden per Traefik dauerhaft auf die Primärdomain geleitet
- anders als bei WordPress gibt es hier normalerweise keine fragilen app-internen Redirect-Loops

## SSH/SFTP-Arbeitsweise
Empfohlener Workflow:

1. per SSH/SFTP auf den Server verbinden
2. Inhalte unter `/srv/static/<domain>/` bearbeiten
3. Änderungen sind direkt live, solange nur Dateien geändert wurden
4. nur bei Hostvars-/Routing-/Auth-Änderungen `static-redeploy.sh` ausführen

## Hinweise
- Der Shared-Container ist ideal für kleine, anspruchslose HTML-Sites.
- Für 5 bis 10 statische Sites ist dieses Modell in der Regel sehr gut geeignet.
- Falls später deutlich mehr Sonderlogik nötig wird, kann jederzeit auf dedizierte Container pro Domain umgestellt werden.
