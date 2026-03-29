# Static Sites Erweiterung (Infra Stack Static-Extension / Ziel: v1.6.1)

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
- `scripts/static-redeploy.sh`: DNS-Check und Redeploy der Shared-Static-Instanz, optional auch gesammelt via `--all`, inkl. interaktiver Verwaltung geschützter Verzeichnisse
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

### Patch-Hinweis v1.5.1
In `v1.5.1` wurde die Hostvars-Prüfung in `static-restore.sh` robuster gemacht:

- akzeptiert jetzt sowohl unquotierte als auch quotierte `domain:`-Einträge
- behebt einen Restore-Abbruch bei frisch erzeugten Minimal-Hostvars ohne Backup-Hostvars

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
    username: "editor"
    password_hash: "$2y$..."
    auth_file: "/srv/static-auth/example.com-private-folder.htpasswd"
```

## Hostvars-Beispiel mit Alias-Domains und geschütztem Bereich
Für eine produktivere Konfiguration kann eine statische Site z. B. so aussehen:

```yaml
domain: faz-pfalz.de

traefik:
  domain: faz-pfalz.de
  aliases:
    - www.faz-pfalz.de
    - faz-pfalz.com
    - www.faz-pfalz.com

static_enabled: true
static_traefik_middleware_default: "crowdsec-default@docker"
static_basic_auth_paths:
  - path: "/private-folder/"
    realm: "Interner Bereich"
    username: "andy"
    password_hash: "$2y$..."
    auth_file: "/srv/static-auth/faz-pfalz.de-private-folder.htpasswd"
```

Das bewirkt:
- `faz-pfalz.de` ist die Primärdomain der Site
- `www.faz-pfalz.de`, `faz-pfalz.com` und `www.faz-pfalz.com` werden per Traefik auf die Primärdomain umgeleitet
- für alle in `traefik.aliases` eingetragenen Domains werden beim Redeploy ebenfalls TLS-Router mit `letsEncrypt` konfiguriert
- der Pfad `/private-folder/` wird per Nginx Basic Auth geschützt

Wichtig in der Praxis:
- jede Alias-Domain braucht einen funktionierenden DNS-A-Record auf den Server
- `username` und `password_hash` können direkt in den Hostvars gepflegt werden
- die eigentlichen HTML-Dateien liegen weiterhin unter `/srv/static/faz-pfalz.de/`

## Passwortgeschützte Unterverzeichnisse
Ein privater Bereich wie `/private-folder/` wird direkt in Nginx über Basic Auth geschützt.

Vorteile:
- stabil und bewährt
- kein Konflikt mit Traefik nötig
- Schutz nur für den gewünschten Unterpfad

Wichtig:
- die `.htpasswd`-Datei liegt **nicht** im Webroot
- sie kann automatisch aus `username` und `password_hash` in den Hostvars erzeugt werden
- alternativ kann weiterhin eine vorhandene `auth_file` genutzt werden

Beispiel zum Anlegen:
```bash
sudo apt install apache2-utils
sudo htpasswd -cB /srv/static-auth/example.com-private-folder.htpasswd andy
```

## Interaktive Auth-Verwaltung per `static-redeploy.sh`
Bei einem Redeploy einer einzelnen statischen Site unterstützt `static-redeploy.sh` jetzt die interaktive Verwaltung geschützter Verzeichnisse.

Ablauf:
- vorhandene geschützte Pfade werden nacheinander geprüft
- existiert ein eingetragener Pfad nicht, zeigt das Script die vorhandene Ordnerstruktur an und fordert direkt zur Korrektur auf
- für bestehende geschützte Pfade kann der Schutz aufgehoben werden, ohne den Ordnerinhalt anzutasten
- alternativ kann das Kennwort für einen bestehenden Schutz neu gesetzt werden
- danach kann in einer Schleife jeweils ein weiteres Verzeichnis geschützt werden
- Benutzername und Passwort-Hash werden in den Hostvars hinterlegt
- die zugehörige `.htpasswd`-Datei wird beim Deploy automatisch erzeugt

Empfohlene Hostvars-Struktur:

```yaml
static_basic_auth_paths:
  - path: "/private-folder/"
    realm: "Secure-Example"
    username: "andy"
    password_hash: ""
```

Praktischer Hinweis:
- ein leerer `password_hash` ist als Platzhalter erlaubt
- beim nächsten `static-redeploy.sh <domain>` wird das Passwort dann interaktiv abgefragt und als Hash gespeichert
- `static-redeploy.sh --all` führt bewusst **keine** interaktive Passwortverwaltung aus

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
- Das Deploy-/Redeploy-Playbook läuft mit `become: true`, weil die Shared-Static-Instanz gezielt Host-Pfade unter `/srv/` verwaltet.

### Patch-Hinweis v1.5.2
In `v1.5.2` wurde `deploy-static.yml` auf `become: true` umgestellt:

- behebt Berechtigungsfehler beim Anlegen von `/srv/static-auth` und `/srv/static-nginx`
- passt zur Architektur, weil die Shared-Static-Instanz Host-Verzeichnisse unter `/srv/` verwaltet

### Patch-Hinweis v1.5.3
In `v1.5.3` wurde der öffentliche HTTPS-Selbsttest in `static-restore.sh` entschärft:

- nutzt jetzt einen echten `GET` statt `HEAD`
- bewertet abweichende Finalstatus nur noch als Warnung statt als Restore-Abbruch
- reduziert Fehlalarme bei statischen Seiten, die im Browser korrekt funktionieren

### Patch-Hinweis v1.5.4
In `v1.5.4` wurde die Static-README um ein ausführlicheres Hostvars-Beispiel ergänzt:

- Primärdomain plus mehrere Alias-Domains
- automatischer TLS-/Redirect-Kontext für Aliase
- geschützter `/private-folder/` mit Basic Auth

### Patch-Hinweis v1.6.0
In `v1.6.0` wurde die Static-Auth-Verwaltung deutlich ausgebaut:

- `static-add.sh` legt bei Bedarf Auth-Platzhalter mit `username` und `password_hash` an
- `static-redeploy.sh` verwaltet geschützte Pfade interaktiv
- vorhandene Ordner werden als Hilfestellung angezeigt und auf Gültigkeit geprüft
- Schutz kann aufgehoben oder mit neuem Kennwort versehen werden
- zusätzliche geschützte Verzeichnisse können in einer Schleife ergänzt werden
- `.htpasswd`-Dateien werden aus den Hostvars automatisch erzeugt

### Patch-Hinweis v1.6.1
In `v1.6.1` wurde die `htpasswd`-Prüfung in `static-redeploy.sh` entschärft:

- `htpasswd` wird nicht mehr pauschal vor jedem Single-Domain-Redeploy verlangt
- das Tool wird erst dann benötigt, wenn tatsächlich ein neues Kennwort bzw. ein neuer Hash erzeugt werden soll
- bestehende geschützte Bereiche können damit auch ohne installiertes `apache2-utils` unverändert redeployed werden
