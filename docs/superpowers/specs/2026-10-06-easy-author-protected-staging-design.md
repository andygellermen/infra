# EasyAuthor – Geschütztes Staging-Deployment

## 1. Ziel und Abgrenzung

Der geprüfte EasyAuthor UI-MVP wird unter `https://author.geller.men` als
geschützte Einzelinstanz für ein reales Autorenreview bereitgestellt. Diese
Instanz ist weder eine öffentliche Produktfreigabe noch eine SaaS-Umgebung.

Der vorhandene lokale Test bleibt lokal. Da bisher keine EasyAuthor-Instanz
auf dem Server existiert, gibt es dort weder Altdaten zu entfernen noch eine
Online-Datenbank zu migrieren. Das erste Deployment erzeugt eine frische
Serverinstanz mit den idempotenten Demo-Daten des Backends. Lokale Buchdaten
werden nicht ungefragt übertragen.

Nicht Bestandteil dieses Schritts sind:

- EasyReader, AddToBook und Webimporte,
- öffentliches Benutzerkonto oder Mehrmandantenbetrieb,
- Produktivfreigabe ohne vorgeschaltete Basic Auth,
- PostgreSQL-Migration,
- Übernahme lokaler Demo- oder Buchdaten.

## 2. Sicherheitsmodell

- Traefik veröffentlicht ausschließlich HTTPS; HTTP wird permanent auf HTTPS
  umgeleitet.
- Eine eigene Basic-Auth-Middleware schützt die vollständige Domain
  einschließlich UI, API und Healthcheck.
- Traefik entfernt den `Authorization`-Header vor der Weitergabe.
- Das Klartextkennwort wird verdeckt abgefragt und weder committed noch in
  Hostvariablen gespeichert. Nur ein bcrypt-Hash mit angemessenem Kostenfaktor
  liegt in einer ignorierten Datei mit Modus `0600`.
- Nur der Webcontainer hängt zusätzlich im gemeinsamen Traefik-Netz. Das
  Go-Backend besitzt keinen Host-Port und liegt ausschließlich im privaten
  Instanznetz.
- Die SQLite-Datei und Markdown-Bibliothek liegen in einem expliziten,
  serverseitigen Datenverzeichnis unter `/srv/easy-author/author.geller.men`.
- Container laufen mit `restart_policy: unless-stopped`; unnötige Privilegien,
  Host-Netzwerk und Docker-Socket-Mounts sind ausgeschlossen.
- Geheimnisse und komplette Nutzdaten werden in Ansible-Ausgaben mit
  `no_log` geschützt.

## 3. Laufzeittopologie

```text
Browser
  ↓ HTTPS + Basic Auth
Traefik
  ↓
easy-author-author-geller-men-web (nginx, Traefik + internes Netz)
  ├── statische React-Anwendung
  └── /api/* → easy-author-author-geller-men-api:8086
                    ↓
        /srv/easy-author/author.geller.men/data
          ├── easy-author.sqlite
          └── library/
```

Das Frontend wird für Staging als Produktions-Bundle gebaut und von nginx
ausgeliefert. Der bisherige Vite-Entwicklungsserver ist nicht Bestandteil der
Serverlaufzeit. nginx liefert unbekannte UI-Routen über `index.html` aus und
proxyt `/api/` ohne CORS-Abhängigkeit an das interne Backend.

Der API-Container erhält:

- `EASY_AUTHOR_ADDR=0.0.0.0:8086`,
- `EASY_AUTHOR_DB_PATH=/srv/easy-author/data/easy-author.sqlite`,
- `EASY_AUTHOR_LIBRARY_DIR=/srv/easy-author/data/library`,
- den origin-spezifischen Wert `https://author.geller.men`, soweit das Backend
  ihn für Browseranfragen benötigt.

## 4. Deployment-Bausteine

Das bestehende geschützte Sprach-A-Lyzer-Muster wird für EasyAuthor angepasst,
ohne dessen PostgreSQL- oder Seed-spezifische Abläufe zu kopieren.

Vorgesehen sind:

- `ansible/playbooks/deploy-easy-author.yml`,
- Rolle `ansible/playbooks/roles/easy-author/`,
- ignorierte Hostvariablen `ansible/hostvars/author.geller.men.yml`,
- Hostvariablen-Vorlage ohne echte Geheimnisse,
- `scripts/easy-author-add.sh`,
- `scripts/easy-author-redeploy.sh`,
- `scripts/easy-author-smoke-check.sh`,
- `scripts/easy-author-backup.sh`,
- `scripts/easy-author-restore.sh`,
- `scripts/easy-author-rotate-secrets.sh`,
- Betriebsdokumentation im EasyAuthor-Verzeichnis.

Alle Helper-Skripte verwenden strikten Shellmodus, validieren Domain und
Werkzeuge, geben keine Geheimnisse aus und unterstützen `--help`. Destruktive
Restore- oder Rückfallaktionen benötigen eine ausdrückliche Bestätigung und
arbeiten nur auf einem exakt aufgelösten Instanzpfad.

## 5. Hostvariablen

Die ignorierte Datei enthält mindestens:

- `domain`,
- `easy_author_enabled: true`,
- unveränderliche Image-Tags beziehungsweise den zu deployenden Git-Stand,
- Basic-Auth-Benutzername, bcrypt-Hash und Realm,
- optionale zusätzliche Traefik-Middleware,
- Daten-, Backup- und Aufbewahrungsparameter,
- Smoke-Retry- und Timeoutwerte.

Sie enthält niemals das Klartextkennwort. Die Erstellung erfolgt ausschließlich
interaktiv über `easy-author-add.sh`, einschließlich DNS-Prüfung gegen die
öffentliche IPv4-Adresse des Infra-Servers. `--skip-dns-check` bleibt eine
bewusste Ausnahme für Vorbereitungen vor dem DNS-Umschalten.

## 6. Build und Bereitstellung

`easy-author-redeploy.sh` führt in dieser Reihenfolge aus:

1. Voraussetzungen, Hostvariablen und Arbeitsbaum prüfen.
2. Backend- und Webimage mit einem nachvollziehbaren Tag des Git-Commits bauen.
3. Vorhandene Instanz und Datenlage ermitteln.
4. Bei bestehender Datenlage ein verifiziertes Vorab-Backup erstellen.
5. Ansible im Check-Modus oder real ausführen.
6. Privates Netz und persistente Verzeichnisse idempotent anlegen.
7. API-Container mit dem Datenverzeichnis starten; `Store.Init` führt die
   additive SQLite-Migration aus.
8. API-Health intern prüfen.
9. Webcontainer starten und Traefik-Labels anwenden.
10. Unauthentifizierten und authentifizierten Smoke-Test ausführen.
11. Persistenz durch kontrollierten API-Neustart und erneuten Read-Test prüfen.

Ein Deployment darf nicht auf einen schmutzigen oder nicht nachvollziehbaren
Git-Stand zeigen. `--check-only`, `--build-only` und `--skip-auth-smoke` werden
für sichere Vorbereitung angeboten; der reale Standardpfad überspringt weder
Backup noch Auth-Smoke.

## 7. Backup, Restore und Rückfall

Vor jedem Redeploy mit vorhandener Datenbank entsteht ein zeitgestempeltes,
nicht überschreibbares Archiv. Für eine konsistente SQLite-Sicherung wird der
API-Container kurz angehalten, das gesamte Datenverzeichnis einschließlich
Markdown-Bibliothek archiviert und anschließend wieder gestartet. Für die
geschützte Einzelinstanz ist diese kurze Wartungspause vertretbar.

Das Backup-Skript prüft:

- erwarteten Instanzpfad und Dateitypen,
- erfolgreichen Archivabschluss,
- Lesbarkeit des Archivs,
- restriktive Rechte,
- eine definierte Aufbewahrung, ohne das jüngste erfolgreiche Backup zu
  löschen.

Restore ist ein separates, bewusstes Kommando. Es erstellt zunächst ein
Sicherungsbackup des aktuellen Zustands, stoppt API und Web, entpackt in ein
neues temporäres Verzeichnis und tauscht erst nach erfolgreicher Prüfung den
Datenpfad aus. Ein Fehlschlag lässt den bisherigen Zustand unangetastet.

Ein Anwendungsrollback verwendet den zuvor deployten Image-Tag. Ein
Datenrollback erfolgt niemals automatisch, weil additive Migrationen bereits
neue Daten enthalten können. Er benötigt den expliziten Restore-Ablauf.

## 8. Smoke- und Abnahmekriterien

Automatisch geprüft werden:

1. `http://author.geller.men` leitet auf HTTPS um.
2. HTTPS ohne Zugangsdaten antwortet mit `401`.
3. HTTPS mit Zugangsdaten liefert die React-Startseite mit `200`.
4. `/api/health` ist authentifiziert erreichbar und meldet `status: ok`.
5. `/api/projects` liefert eine gültige JSON-Struktur.
6. `/api/kanban?limit=12&include_done=true` liefert alle fünf Phasen.
7. Der API-Container veröffentlicht keinen Host-Port.
8. SQLite-Datei und Bibliothek liegen im vorgesehenen Persistenzpfad.
9. Nach einem API-Neustart bleiben Projekt- und Kanban-Antworten lesbar.

Passwörter werden verdeckt über `/dev/tty`, eine ausdrücklich gewählte
stdin-Option oder eine kurzlebige Umgebungsvariable eingelesen und unmittelbar
danach entfernt.

## 9. Fehlerbehandlung und Beobachtbarkeit

- Ein fehlgeschlagener Build verändert die laufende Instanz nicht.
- Ein fehlgeschlagenes Backup verhindert den Redeploy.
- Ein fehlgeschlagener interner Healthcheck verhindert das Umschalten des
  Webcontainers.
- Ein fehlgeschlagener externer Auth-Smoke wird klar gemeldet; die vorherige
  Image-Version und das Vorab-Backup bleiben erhalten.
- Betriebsbefehle dokumentieren Containerstatus, begrenzte Logauszüge,
  Backup-Liste und erneuten Smoke-Test.
- Keine Logausgabe darf Basic-Auth-Zugangsdaten, Hostvariablen oder Manuskripte
  enthalten.

## 10. Einführung und Review

Die Umsetzung erfolgt zunächst im isolierten EasyAuthor-Feature-Zweig und wird
lokal mit Shell-Syntaxprüfungen, Ansible-Syntax-/Check-Modus, Docker-Builds und
Container-Smokes validiert. Danach wird der gesamte Stand kontrolliert in
`main` integriert und normal zu `origin` gepusht.

Erst anschließend werden auf dem Infra-Server DNS, ignorierte Hostvariablen und
das reale Deployment vorbereitet. Der erste Onlinegang endet mit dem
automatischen Smoke-Test. Danach beginnt das manuelle Autorenreview; eine
öffentliche Freigabe ist ausdrücklich nicht impliziert.
