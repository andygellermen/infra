# OpenProject im Infra-Stack

Die vier Skripte verwalten getrennte OpenProject-Community-Instanzen je Domain.
Traefik übernimmt HTTPS; jede Instanz erhält PostgreSQL, Memcached, Web, Worker,
Cron und Hocuspocus für gemeinsames Bearbeiten. Die AI Bridge aus
[Chads Spezifikation](docs/OPENPROJECT_AI_BRIDGE_SPEC.md) ist ein separates Folgeprojekt.
Ihre private Installation wird mit den
[Bridge-Lifecycle-Skripten](docs/OPENPROJECT_AI_BRIDGE_OPERATIONS.md) gestartet,
gestoppt, geprüft und aktualisiert.

## Voraussetzungen

Auf dem **Linux-Infra-Host** ausführen, mit Root-Rechten:

- Docker Engine und Compose **ab 2.30**; der Docker-Dienst muss laufen.
- Bestehender Traefik-Stack mit Docker-Netz `traefik`, EntryPoints `web` und
  `websecure`, Resolver `letsEncrypt` bzw. eingerichtetem `letsEncryptDns`.
- Python ab 3.9 mit PyYAML, Ansible Core und Bash. Beispielsweise auf Debian/Ubuntu:
  `sudo apt-get install python3-yaml ansible-core`.
- Genügend Arbeitsspeicher für die OpenProject-Dienste und freier Speicher für
  Daten, temporäre unkomprimierte Sicherungen und Archive. Die Archivprüfung
  benötigt zusätzlichen Platz; maximal 100 GiB unkomprimierter Archivinhalt.
- Ein auf den Infra-Host zeigender DNS-A-Record. Ein vorhandener AAAA-Record
  muss ebenfalls korrekt sein; das Skript prüft den IPv4-Pfad.

Festgelegte Ausgangsversionen: OpenProject `17.8.0-slim`, PostgreSQL `17.11`,
Memcached `1.6.45-alpine`, Hocuspocus `17.8.0`. Die Tags wurden am 2026-09-29
in der Hersteller-/Docker-Registry geprüft. Images werden vor Eingriffen geladen.
Es gibt keine automatischen Major-Upgrades.

## Installation

Im Repo auf dem Infra-Host:

```bash
sudo ./scripts/op-add.sh --domain=op.domain.de
```

Ohne `--domain` fragt das Skript interaktiv danach. Ohne Terminal ist der
Parameter verpflichtend. Alle Skripte bieten `--help` und funktionieren auch
beim Aufruf über einen absoluten Pfad aus einem anderen Arbeitsverzeichnis.

Optionen:

```bash
sudo ./scripts/op-add.sh --domain=op.domain.de --version=17.8.0
sudo ./scripts/op-add.sh --domain=op.domain.de --wildcard-domain=domain.de --dns-account=primary
```

`--skip-dns-check` ist für gezielte Test-/Proxy-Installationen vorgesehen.
Wildcard-DNS-Provider und Zugangsdaten müssen bereits in Traefik eingerichtet
sein. `--dns-account` schreibt die Metadaten für die vorhandene Traefik-Rolle;
das Skript verändert oder startet den gemeinsam genutzten Traefik nicht neu.

Hostvars liegen unter `ansible/hostvars/op.domain.de.yml`, Laufzeitdateien unter
`/srv/openproject/op.domain.de/`, jeweils mit privaten Rechten. Das anfängliche
Admin-Kennwort steht ausschließlich in den privaten Hostvars unter
`op.admin_password`. Benutzername: `admin`. Öffne die Datei in einem privaten
Editor und ändere das Kennwort beim ersten Login. Keine Zugangsdaten in Git,
Tickets oder öffentliche Logs kopieren.

Die Bereitschaftsprüfung prüft den internen Web-Healthcheck und die laufenden
Hintergrunddienste. Anschließend `https://op.domain.de` im Browser öffnen und
TLS, Login, ein Arbeitspaket mit Dateianhang sowie gemeinsames Bearbeiten testen.
Ein grüner interner Healthcheck bestätigt noch keine öffentliche DNS-/TLS-Verbindung.

## Konfiguration und Mail

Die Hostvars enthalten ein strikt validiertes `op`-Objekt. Änderungen dort werden
mit `op-update.sh --version=<aktuell installierte Version>` angewendet; auch dieser
Redeploy erstellt zuerst ein Backup. Unbekannte Felder werden abgewiesen.

Bei der Erstanlage werden `ses_smtp_host`, `ses_smtp_port`, `ses_smtp_user`,
`ses_smtp_password` und `ses_from` aus `ansible/secrets/secrets.yml` bzw.
`secrets.yaml` übernommen, sofern vorhanden. Die Datei muss lesbares YAML sein;
verschlüsselte Vault-Dateien werden nicht automatisch entschlüsselt.

Alternativ nach der Anlage folgende `op`-Felder setzen und erneut deployen:
`smtp_host`, `smtp_port`, `smtp_user`, `smtp_password`, `smtp_from`,
`smtp_authentication` (`plain`, `login`, `cram_md5`), `smtp_starttls` (boolesch).
Ohne SMTP-Konfiguration wird ein Hinweis ausgegeben; Mailversand ist dann nicht
als eingerichtet zu betrachten. In OpenProject eine Testmail versenden.

Umgebungsdateien verwenden Compose `format: raw`, sodass `$`, Anführungszeichen
und `#` in Kennwörtern erhalten bleiben. Steuerzeichen und Jinja-Ausdrücke wie
`{{ ... }}` werden abgewiesen. Secrets werden nicht als Shellcode eingelesen.
`op.middleware` kann eine vorhandene Traefik-Middleware-Liste enthalten; sie gilt
für Web und WebSocket-Route. Datenbank und Cache veröffentlichen keine Hostports.

## Update

```bash
sudo ./scripts/op-update.sh --domain=op.domain.de --version=17.8.0
```

Eine neuere vollständige Release-Version im selben Major kann angegeben werden.
Ohne Version fragt das Skript interaktiv danach. Major-Wechsel und Downgrades
werden abgewiesen. PostgreSQL wird über dieses Skript nicht automatisch auf
neue Versionsnummern gehoben; Datenbank-Upgrades benötigen einen eigenen Ablauf.

Ablauf: Zielimages laden, schreibende Dienste stoppen, vollständiges Backup
prüfen, Konfiguration rendern, Migration/Seeder ausführen, Dienste starten und
Bereitschaft prüfen. Die Dienste bleiben zwischen Snapshot und Migration
angehalten. Ein Backup-Fehler verhindert das Update.

Schlägt die Migration fehl, erfolgt kein automatischer Start eines alten Images
gegen die möglicherweise migrierte Datenbank. `MAINTENANCE` im Laufzeitverzeichnis
enthält den Recovery-Verweis. Wiederherstellung siehe unten.

## Backup

```bash
sudo ./scripts/op-backup.sh --domain=op.domain.de
sudo ./scripts/op-backup.sh --domain=op.domain.de --output=/backups/op-manual.tar.gz
```

Das Skript unterbricht schreibende Dienste kurz und startet anschließend nur
zuvor laufende Dienste wieder, auch bei einem Backup-Fehler. Scheitert der
Wiederanlauf, endet es mit einem Fehlercode. Standardziel:
`backups/openproject/<domain>/op-<UTC-Zeitstempel>.tar.gz`.

Archivformat 1 enthält PostgreSQL-Custom-Dump, Anhänge, den tatsächlich deployten
Hostvars-Stand aus dem Instanzverzeichnis (auch bei bereits bearbeiteten Repo-Hostvars), gerenderte
Compose-/Umgebungsdateien und ein Manifest mit Domain, Version, Image-IDs und
SHA-256-Prüfsummen. Archive werden erst nach vollständiger Rückleseprüfung
veröffentlicht und überschreiben keine vorhandene Datei.

**Backups enthalten Secrets und sind nicht verschlüsselt.** Dateimodus ist 0600.
Für externe Ablage Zugriffsschutz und Verschlüsselung vorsehen. Images selbst
sind nicht enthalten; Restore benötigt die gespeicherten Release-Tags aus der
Registry bzw. dem lokalen Image-Cache. Image-IDs dienen als Provenienznachweis.

## Restore

Zuerst ohne Änderungen prüfen:

```bash
sudo ./scripts/op-restore.sh --domain=op.domain.de --backup=/backups/op-manual.tar.gz --dry-run
```

Danach wiederherstellen:

```bash
sudo ./scripts/op-restore.sh --domain=op.domain.de --backup=/backups/op-manual.tar.gz
# Fuer bewusst nicht-interaktiven Betrieb:
sudo ./scripts/op-restore.sh --domain=op.domain.de --backup=/backups/op-manual.tar.gz --yes
```

Unterstützt werden eine leere Zielinstallation oder dieselbe vorhandene Domain.
Bei vorhandener Instanz wird vor Änderungen bestätigt und ihr aktueller Zustand
gesichert. PostgreSQL-Versionen müssen beim Überschreiben übereinstimmen.
Domainumzüge werden abgewiesen. Externe Dateispeicher sind nicht Teil dieser
Integration oder ihrer Sicherung.

Archivpfade, Links, Duplikate, Prüfsummen, Domain und Konfiguration werden vor
Eingriffen geprüft. Gesicherte Compose-Dateien werden nicht ausgeführt; Ansible
rendert sie aus den validierten Hostvars neu. Restore importiert Daten und Anhänge
und startet die gespeicherte OpenProject-Version ohne Seeder/Versionsmigration.

### Wiederanlauf nach Fehlern

Fehlgeschlagene **Neuinstallation** mit denselben Secrets wiederholen:

```bash
sudo ./scripts/op-add.sh --domain=op.domain.de --resume
```

Dabei gelten die gespeicherten Hostvars; Versions-/Wildcard-Optionen der
Erstanlage werden nicht erneut übernommen. `--resume` ist nur bei einer als
Neuinstallation markierten Wartungsinstanz erlaubt.

Nach fehlgeschlagenem **Update oder Restore** den ursprünglichen, vollständigen
Backup-Pfad aus `MAINTENANCE` verwenden:

```bash
sudo ./scripts/op-restore.sh --domain=op.domain.de --backup=/backups/original.tar.gz --recover --yes
```

`--recover` verlangt einen Wartungsmarker und gestoppte Anwendungsdienste. Es
überspringt ausdrücklich die neue Sicherung der bereits beschädigten Datenbank;
damit funktioniert ein Wiederholungsversuch auch nach fehlgeschlagenem DB-Import.
Der ursprüngliche Recovery-Verweis bleibt bei erneutem Fehlschlag erhalten.
Bei gesunder laufender Instanz ist dieser Modus nicht erlaubt.

Alle Mutationen halten eine Dateisperre je Domain. Direkte Docker-/Ansible-Aufrufe
umgehen diese Sperre; während der Skripte keine manuellen Änderungen vornehmen.
Das Ansible-Playbook ist eine interne Render-Komponente und startet keine Dienste.
Bei Fehlern werden Container-Ausgaben wegen möglicher Secrets nicht in das
Terminal übernommen; Dienste/Logs gezielt und privat auf dem Host untersuchen.

## Tests und Prüfstand

```bash
python3 -m venv /tmp/op-checks
/tmp/op-checks/bin/pip install PyYAML ansible-core
PATH=/tmp/op-checks/bin:$PATH OP_PYTHON=/tmp/op-checks/bin/python \
  /tmp/op-checks/bin/python -m unittest discover -s apps/openproject/tests -v
```

Verhaltenstests decken unter anderem Sperren, Domainisolation, Archivangriffe,
Backup-/Import-/Migrationsfehler und Secret-Rendering ab. Der Render-Test verwendet
echtes Ansible und `docker compose config`, benötigt aber keinen Docker-Daemon.

Auf einem separaten Linux-Testhost mit Traefik-Netz und einer **neuen Testdomain**:

```bash
sudo ./apps/openproject/tests/integration_roundtrip.sh \
  --domain=op-test.domain.de --confirm-disposable
```

Der Rundlauf installiert, sichert eine Probezeile in PostgreSQL und eine Datei im
Attachment-Volume, verändert beides, restauriert und prüft die ursprünglichen
Inhalte. Optional `--update-version=X.Y.Z` ergänzt ein Update. Instanz und Archive
bleiben zur Prüfung erhalten; es gibt keine automatische Löschung. Zusätzlich
den Browser-Test mit echtem Arbeitspaket und Anhang durchführen.

**Lokale Grenze am 2026-09-29:** Docker-CLI/Compose-Rendering verfügbar, Docker-Daemon
nicht erreichbar. Der echte Container-Rundlauf und öffentliche TLS-/Mail-/UI-Tests
sind hier nicht ausgeführt. Die Integration ist deshalb noch nicht auf einem
laufenden OpenProject-Stack end-to-end verifiziert.

Für isolierte Tests können `OP_SITE_ROOT`, `OP_HOSTVARS_DIR`, `OP_BACKUP_DIR` und
`OP_PYTHON` überschrieben werden. Auf dem Produktivhost immer konsistente Pfade
verwenden, insbesondere für die gemeinsame Instanzsperre.

## Referenzen

- [Offizielle Compose-Installation](https://www.openproject.org/docs/installation-and-operations/installation/docker-compose/)
- [Offizielle Dienstaufteilung](https://github.com/opf/openproject-docker-compose/tree/stable/17)
- [OpenProject-Konfiguration](https://www.openproject.org/docs/installation-and-operations/configuration/environment/)
- [Backup/Restore](https://www.openproject.org/docs/installation-and-operations/operation/restoring/)
- [Freigegebener Entwurf](docs/INFRA_INTEGRATION_DESIGN.md)
- [Umsetzungsplan und Abweichungen](docs/IMPLEMENTATION_PLAN.md)
