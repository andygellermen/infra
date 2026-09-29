# OpenProject im Infra-Stack

Stand: 2026-09-29. Status: vom Benutzer freigegeben; Umsetzung und Prüfgrenzen im Runbook dokumentiert.

## Ziel und Abgrenzung

OpenProject Community Edition soll auf dem vorhandenen Infra-Host durch
`scripts/op-add.sh --domain=op.example.org` installierbar werden. Update, Backup
und Wiederherstellung adressieren ebenfalls genau eine Domain. Ohne Domain
fragen die Skripte interaktiv danach; ohne Terminal brechen sie mit Usage ab.

Die bereitgestellte `OPENPROJECT_AI_BRIDGE_SPEC.md` beschreibt eine spätere,
eigenständige Integration. Deren Quellcode gehört gemäß Abschnitt 39 in ein
separates Repository. Dieser Umsetzungsschritt schafft die OpenProject-Instanz
als Grundlage. Bridge, MCP und projektspezifische Workflows folgen separat.

## Architekturentscheidung

Empfehlung: ein Docker-Compose-Projekt pro Domain, durch eine Ansible-Rolle
bereitgestellt. Es enthält OpenProject-Webprozess, Hintergrundverarbeitung,
Initialisierung/Migration, PostgreSQL und den zur gewählten Upstream-Version
gehörenden Cache. Die Dienstaufteilung folgt der offiziellen Compose-Vorlage.
Nur der Webzugang wird an das bestehende `traefik`-Netz angeschlossen.
Datenbank und Cache erhalten keine veröffentlichten Hostports.

Alternativen: Der All-in-one-Container reduziert die Containerzahl, koppelt aber
Datenbank und Anwendung stärker. Einzelne Ansible-Container-Tasks entsprechen
anderen Apps im Repo, würden jedoch mehr Upstream-Startlogik duplizieren.
Compose mit Ansible als Bereitstellungsschicht hält diese Logik zusammen.

Instanzkonfiguration: `ansible/hostvars/<domain>.yml`, aktiviert über
`op_enabled: true`. Laufzeitdateien: `/srv/openproject/<domain>/`.
Persistente Daten werden pro Instanz getrennt gehalten. Ressourcenbezeichner
verwenden eine kollisionsfreie Ableitung aus der normalisierten Domain.

Traefik übernimmt HTTPS, HTTP-Weiterleitung und den bestehenden
`letsEncrypt`-Resolver. Wildcard-Konfiguration wird über die vorhandenen
Konventionen unterstützt. Hostname und HTTPS-Erkennung werden in OpenProject
passend gesetzt. Middleware ist pro Instanz konfigurierbar.

SMTP kann die vorhandenen SES-Einstellungen übernehmen; fehlende Maildaten
werden deutlich gemeldet. Secrets werden nicht ausgegeben und mit restriktiven
Dateirechten außerhalb versionierter Dateien gespeichert. Ansible behandelt
Tasks mit Secrets mit `no_log`. Keine automatische Anlage von API-Tokens.

## Skriptverträge

### op-add.sh

- Unterstützt `--domain=...`, interaktive Eingabe und `--help`.
- Prüft Domain, Werkzeuge, Docker, DNS und vorhandene Konfiguration, bevor
  Änderungen beginnen. Eine vorhandene Instanz wird nicht überschrieben.
- Erzeugt individuelle Secrets und Hostvars, rendert Compose über Ansible,
  initialisiert die Datenbank und startet die Dienste.
- Wartet begrenzt auf Bereitschaft; meldet Fehler mit nicht-null Exitcode.
- Dokumentiert Erstzugang und Passwortwechsel entsprechend der ausgewählten
  OpenProject-Version, ohne Zugangsdaten in Logs zu schreiben.

### op-update.sh

- Unterstützt `--domain=...` und eine explizite Zielversion `--version=...`.
- Verwendet feste Release-Versionen; kein unkontrolliertes `latest`.
- Erstellt vor Migrationen ein vollständiges Backup. Scheitert es, findet kein
  Update statt. Major-Wechsel erfordern einen gesondert dokumentierten Upgradepfad;
  das erste Skript lehnt sie ebenso wie Downgrades ab.
- Prüft das Zielimage vor dem Stoppen, führt Migrationen aus und prüft danach
  die Bereitschaft. Bei Migrationsfehlern bleibt die Instanz im Wartungszustand;
  kein automatischer Start eines alten Images gegen eine migrierte Datenbank.

### op-backup.sh

- Unterstützt `--domain=...` und optional `--output=...`.
- Stoppt schreibende Anwendungsdienste für ein konsistentes Backup.
- Sichert logischen PostgreSQL-Dump, Anhänge, Hostvars, Laufzeitkonfiguration,
  benötigte Secrets und ein Manifest mit Domain, Formatversion und Imageversionen.
- Veröffentlicht das Archiv erst nach erfolgreicher Prüfung atomar. Bestehende
  Archive werden nicht überschrieben. Standardziel: `backups/openproject/<domain>/`.
- Startet auch bei Fehlern nur die vorher laufenden Dienste wieder und meldet
  fehlgeschlagene Wiederanläufe. Backup enthält Secrets und erhält Modus 0600.

### op-restore.sh

- Unterstützt `--domain=...`, `--backup=...`, `--dry-run` und `--yes`.
- Prüft Archivformat, Vollständigkeit, Prüfsummen, Domain und Version vor Eingriffen.
  Unsichere Archivpfade und Links werden zurückgewiesen; Archivinhalt wird nie
  als Shellcode eingelesen.
- Unterstützt Wiederherstellung derselben Domain auf vorhandener oder leerer
  Zielinstallation. Domainwechsel sind nicht Teil der ersten Fassung.
- Vor Überschreiben einer vorhandenen Instanz: ausdrückliche Bestätigung oder
  `--yes` sowie Sicherung ihres aktuellen Zustands.
- Stellt Konfiguration, Datenbank und Anhänge mit den gespeicherten Versionen
  wieder her. Ein anschließendes Update ist ein separater Vorgang.
- Startet erst nach vollständigem Import und prüft Bereitschaft; bei Fehlern
  bleibt der teilweise restaurierte Dienst gestoppt und meldet den Recoverypfad.

Alle mutierenden Abläufe teilen eine Sperre pro Domain, damit Update, Backup
und Restore nicht gleichzeitig dieselbe Instanz bearbeiten.

## Dateien und Prüfungen

Geplant sind die vier Skripte, eine gemeinsame Hilfsbibliothek,
`ansible/playbooks/deploy-openproject.yml`, die zugehörige Rolle mit
Compose-/Konfigurationstemplates und ein Betriebs-README. Der Skriptindex wird
ergänzt. Globale automatische Updates bleiben zunächst ohne OpenProject.

Validierung: Shell-Syntax, ShellCheck soweit verfügbar, Ansible-Syntax,
Rendering und Compose-Konfigurationsprüfung mit Testdaten. Verhaltenstests
prüfen Domainvalidierung, Instanzisolation, Sperren, Fehlerbehandlung,
Archivvalidierung und den Abbruch bei fehlgeschlagenem Backup. Ein echter
Installations-/Backup-/Restore-Rundlauf mit Testdaten ist der Integrationstest;
falls Docker hier nicht verfügbar ist, wird diese Grenze ausdrücklich berichtet.
Produktionsdeployment führt anschließend der Betreiber per op-add.sh aus.

## Quellen und Versionswahl

- [Offizielle Compose-Installation](https://www.openproject.org/docs/installation-and-operations/installation/docker-compose/)
- [Offizielle Compose-Vorlage](https://github.com/opf/openproject-docker-compose)
- [Backup](https://www.openproject.org/docs/installation-and-operations/operation/backing-up/)
- [Restore](https://www.openproject.org/docs/installation-and-operations/operation/restoring/)

Die konkrete Release-Version und kompatible PostgreSQL-/Cache-Version werden
bei der Umsetzung anhand der aktuellen offiziellen Vorlage geprüft und fest
im Template hinterlegt. Das ist eine Implementierungsprüfung, keine Vorgabe,
beliebige spätere Major-Releases automatisch zu übernehmen.
