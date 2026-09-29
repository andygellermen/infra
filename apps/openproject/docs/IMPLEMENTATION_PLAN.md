# OpenProject Infra Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** OpenProject Community Edition pro Domain mit vier Skripten installieren, aktualisieren, sichern und wiederherstellen.

**Architecture:** Ansible rendert ein eigenes Compose-Projekt je Domain. Bash-Einstiegspunkte verwenden gemeinsame Betriebsfunktionen; Python verarbeitet strukturierte Konfiguration und prüft Backup-Archive. Traefik terminiert TLS, PostgreSQL und Cache bleiben im Instanznetz.

**Tech Stack:** Bash, Python 3 mit PyYAML, Ansible, Docker Compose v2, offizielle OpenProject-Images, PostgreSQL, Memcached.

**Spec:** [Freigegebener Integrationsentwurf](INFRA_INTEGRATION_DESIGN.md), vom Benutzer am 2026-09-29 bestätigt.

## Global Constraints

- Alle vier Skripte unterstützen `--domain=...`, interaktive Domain-Abfrage und `--help`; ohne Terminal und Domain Abbruch.
- Instanzkonfiguration: `ansible/hostvars/<domain>.yml`, aktiviert über `op_enabled: true`.
- Laufzeitdateien: `/srv/openproject/<domain>/`; Backups: `backups/openproject/<domain>/`.
- Keine Hostports für PostgreSQL oder Cache; feste Release-Versionen statt `latest`.
- Backup vor jedem Update; erste Fassung lehnt Major-Wechsel und Downgrades ab.
- Restore nur für dieselbe Domain, mit gespeicherten Versionen; Überschreiben benötigt Bestätigung oder `--yes` und vorheriges Backup.
- Secrets nicht in Git oder Logs; Backup-Dateien Modus 0600; Ansible-Tasks mit Secrets verwenden `no_log`.
- Sperre je Domain; keine parallelen Mutationen derselben Instanz.
- AI Bridge bleibt ein späteres eigenständiges Projekt. Keine Änderung globaler automatischer Updates.

## Review Focus

1. Domainnamen mit Bindestrichen/Punkten dürfen nicht auf dasselbe Compose-Projekt abgebildet werden (Task 1).
2. Abbruch nach dem Stoppen muss beim Backup den vorherigen Laufzustand wiederherstellen; Restore-Fehler dürfen keine halbfertige Instanz starten (Tasks 3/4).
3. Bösartige Archive mit Links, absoluten Pfaden, doppelten Einträgen oder `..` dürfen keine Zieldateien verändern (Task 4).
4. Geheimnisse mit `$`, Anführungszeichen oder Zeilenumbrüchen müssen sicher serialisiert oder klar abgelehnt werden (Task 2).
5. Zweite Domain, fehlendes Zielimage und fehlerhafte Migration müssen von der laufenden Nachbarinstanz isoliert bleiben (Tasks 2/5).

## Task 1: CLI, Konfiguration und Instanzsperren

**Files:** `scripts/lib/openproject.sh`, `scripts/lib/openproject_config.py`, `apps/openproject/tests/test_cli.py`, `apps/openproject/tests/test_config.py`.

**Interfaces:** `op_parse_args "$@"` setzt validierte Optionsvariablen; `op_select_domain` normalisiert und validiert die Domain; `op_lock` erwirbt die Instanzsperre; `op_compose ...` adressiert ausschließlich das ausgewählte Projekt. Python-CLI bietet `validate-domain`, `site-id`, `read-config` und `write-config`; JSON/YAML werden ausschließlich mit sicheren Parsern gelesen.

- [x] Tests zuerst: `test_reject_path_domain`, `test_missing_domain_noninteractive`, `test_unknown_option`, `test_distinct_site_ids`, `test_lock_contention`. Jeweils nicht-null Exitcode ohne Docker-Mutation; `a-b.example.org` und `a.b-example.org` ergeben unterschiedliche IDs.
- [x] Tests mit `python3 -m unittest discover -s apps/openproject/tests -v` ausführen und erwartete fehlende Implementierung feststellen.
- [x] Gemeinsame Funktionen implementieren; Pfade vom Skriptstandort ableiten, Argumente als Arrays übergeben, keine Konfiguration mit `source` oder `eval` ausführen. Projektname ist `op-` plus vollständige SHA-256 der normalisierten Domain. Sperren über Linux `flock`; Tests verwenden isolierte Pfade.
- [x] Tests erneut ausführen; erwartetes Ergebnis: alle Task-1-Tests erfolgreich.

## Task 2: Ansible-Deployment und Installation

**Files:** `scripts/op-add.sh`, `ansible/playbooks/deploy-openproject.yml`, `ansible/playbooks/roles/openproject/{defaults,tasks}/main.yml`, `ansible/playbooks/roles/openproject/templates/compose.yml.j2`, `ansible/playbooks/roles/openproject/templates/openproject.env.j2`, `apps/openproject/tests/test_deployment.py`.

**Interfaces:** Playbook akzeptiert genau `target_domain` und rendert die Instanz. `op_deploy` aus der Shellbibliothek führt es für diese Domain aus. `op_wait_ready` wartet begrenzt auf den Web-Healthcheck und prüft den Laufzustand der Hintergrunddienste.

- [x] Tests zuerst: `test_target_domain_only`, `test_no_database_host_ports`, `test_tls_labels`, `test_secret_serialization`, `test_existing_instance_not_overwritten`, `test_seeder_failure_prevents_start` mit gerenderten Testkonfigurationen und aufgezeichneten Docker-Aufrufen.
- [x] Testlauf durchführen und erwartete Fehler dokumentieren.
- [x] Aktuelle offizielle Release-Tags und kompatible PostgreSQL-/Cache-Version prüfen und als feste Defaults eintragen. Web, Worker, Cron und Seeder folgen den Upstream-Kommandos. Seeder läuft einmalig und muss vor Anwendungsstart erfolgreich enden.
- [x] Kollaboratives Editing explizit konfigurieren: sofern in der gewählten Version standardmäßig aktiv, den offiziellen Hocuspocus-Dienst samt dedizierter Traefik-WebSocket-Route ergänzen; dessen interner Callback und gemeinsames Secret werden geprüft. Diese Versionsabhängigkeit nicht stillschweigend ignorieren.
- [x] `op-add.sh` implementieren: Voraussetzungen und DNS prüfen, Secrets generieren, Hostvars sicher anlegen, Playbook ausführen, Bereitschaft melden. Unterstützt zusätzlich `--version=...`, `--wildcard-domain=...`, `--dns-account=...`, `--skip-dns-check`. SES aus den vorhandenen Ansible-Secrets übernehmen; fehlende SMTP-Konfiguration klar melden.
- [x] Tests, Ansible-Syntax und `docker compose config --quiet` auf gerenderten Testdaten ausführen. Erstzugang anhand des gewählten Releases im Runbook festhalten.

## Task 3: Konsistentes Backup

**Files:** `scripts/op-backup.sh`, `scripts/lib/openproject-backup.sh`, `scripts/lib/openproject_archive.py`, `apps/openproject/tests/test_backup.py`.

**Interfaces:** `op_backup_locked OUTPUT` setzt eine gehaltene Domain-Sperre voraus und erstellt ein Archiv ohne erneute Sperranforderung; Update und Restore verwenden dieselbe Funktion. Archivformat v1 enthält `manifest.json`, `database.dump`, `assets.tar`, `hostvars.yml`, `compose.yml`, `openproject.env`. Manifest enthält Domain, Versionen/Image-Digests, Zeitstempel und SHA-256 pro Nutzdatei.

- [x] Tests zuerst: `test_dump_failure_restores_running_services`, `test_stopped_service_stays_stopped`, `test_restart_failure_nonzero`, `test_existing_output_preserved`, `test_incomplete_archive_not_published`.
- [x] Tests ausführen und erwartete Fehler feststellen.
- [x] Nur laufende schreibende Dienste stoppen; PostgreSQL bleibt für `pg_dump --format=custom` verfügbar. Anhänge und Konfiguration sichern; Secrets nicht auf stdout schreiben. Dateien in privatem Staging sammeln und Archiv nach Prüfung atomar ohne Überschreiben veröffentlichen. Traps für Fehler/Signale und Wiederanlauf einrichten.
- [x] Tests erfolgreich ausführen; Dateimodus 0600 und Manifest-Prüfsummen verifizieren.

## Task 4: Validierter Restore

**Files:** `scripts/op-restore.sh`, Erweiterung `scripts/lib/openproject_archive.py`, `apps/openproject/tests/test_restore.py`.

**Interfaces:** Python-CLI `validate ARCHIVE DOMAIN` prüft das vollständige Archiv einschließlich `assets.tar`; `extract ARCHIVE DEST DOMAIN` extrahiert nur nach Validierung in ein privates Staging-Verzeichnis. `--dry-run` prüft ohne Deployment oder Datenänderungen.

- [x] Tests zuerst: `test_reject_traversal`, `test_reject_links`, `test_reject_duplicate_members`, `test_checksum_mismatch`, `test_wrong_domain`, `test_dry_run_no_mutation`, `test_restore_failure_stays_stopped`, `test_pre_restore_backup_failure_aborts`.
- [x] Tests ausführen und erwartete Fehler feststellen.
- [x] Archive als nicht vertrauenswürdige Daten behandeln: Pfade, Dateitypen, Größenlimit, Manifest und Checksummen prüfen. Keine Shellauswertung. Bestehende Instanz erst nach Bestätigung und erfolgreicher Vorsicherung stoppen. Leeres Ziel anhand gesicherter Konfiguration vorbereiten, DB/Anhänge importieren, danach Hintergrunddienste und Web starten. Restore initialisiert keine Beispieldaten und führt kein Versionsupgrade aus.
- [x] Tests erfolgreich ausführen; prüfen, dass jede Fehlerphase einen klaren Wiederherstellungshinweis und nicht-null Exitcode liefert.

## Task 5: Update und Betriebsnachweis

**Files:** `scripts/op-update.sh`, `apps/openproject/tests/test_update.py`, `apps/openproject/tests/integration_roundtrip.sh`, `apps/openproject/README.md`, `README-infra-scripts.md`.

**Interfaces:** `op-update.sh --domain=... --version=X.Y.Z` verwendet Task-1-Sperre, Task-3-Backup und Task-2-Deployment/Healthcheck. Ohne Zielversion erfolgt eine interaktive Abfrage; ohne Terminal Abbruch.

- [x] Tests zuerst: `test_major_change_rejected`, `test_downgrade_rejected`, `test_missing_image_preserves_running_site`, `test_failed_backup_prevents_migration`, `test_failed_migration_no_old_image_restart`, `test_other_domain_untouched`.
- [x] Tests ausführen und erwartete Fehler feststellen.
- [x] Update implementieren: Version validieren, Image vorab laden, Backup erstellen, schreibende Dienste stoppen, Zielkonfiguration rendern, migrieren, Bereitschaft prüfen. Fehler nach Migration führen zum Wartungszustand und Hinweis auf das konkrete Backup; kein vermeintliches Rollback allein durch Imagewechsel.
- [x] Runbook mit Voraussetzungen, vier Befehlen, Erstzugang, Mail, Wartungsunterbrechung, Backup-Geheimnissen und Recoverypfad schreiben; Skriptindex ergänzen.
- [x] Shell-Syntax, ShellCheck, gesamte Verhaltenstests, Ansible-Syntax und Compose-Rendering ausführen. Benötigte Python-Werkzeuge isoliert installieren, ohne globale Python-Pakete zu verändern.
- [ ] Integrationstest auf separater Testinstanz: Installation, Test-Arbeitspaket mit Anhang, Backup, kontrollierte Datenänderung, Restore und Nachweis ursprünglicher Daten/Anhang. Anschließend Update auf eine kompatible neuere Version, sofern verfügbar. Kein Produktionsdeployment.
- [x] Änderungen und Anforderungen abschließend abgleichen, `git diff --check` ausführen und Prüfgrenzen berichten.

## Lokaler Befund und Ausführung

Am 2026-09-29 ist die Docker-CLI vorhanden, der lokale Docker-Daemon jedoch
nicht erreichbar. Ansible und ShellCheck fehlen im PATH. Statische Prüfungen
und Verhaltenstests können vorbereitet und mit isolierten Werkzeugen ausgeführt
werden. Ein echter Container-Rundlauf darf erst als bestanden gelten, wenn ein
Docker-Daemon tatsächlich erreichbar ist.

Empfohlene Ausführung: direkt in dieser Sitzung, Aufgaben nacheinander, da alle
Betriebsskripte dieselben Konfigurations-, Sperr- und Archivverträge verwenden.
Der Plan wurde freigegeben und umgesetzt; offene Laufzeitvalidierung siehe unten.

## Umsetzung und begruendete Abweichungen

Design, Plan und direkte Ausführung wurden vom Benutzer freigegeben; Commit und
Push sind ebenfalls beauftragt. Umsetzung auf `feat/openproject-infra` im vorhandenen
Checkout, damit die anfänglich unversionierten Spezifikationen erhalten bleiben.

- Die Bash-Einstiegspunkte delegieren an vier kleine Python-Module (`config`,
  `archive`, `runtime`, `cli`), statt Parsing, Trap-Logik und JSON-Verarbeitung
  zwischen Bash und Python aufzuteilen. `fcntl.flock` liefert dieselbe
  Betriebssystem-Sperre auf Linux und macOS; keine zusätzliche Shell-Lockbibliothek.
- Versionsdefaults liegen zentral in `openproject_config.py`; Ansible rendert
  ausschließlich übergebene, vorher validierte Daten. Deshalb kein separates
  Rollen-Defaults-File und keine Python-CLI für interne Einzeloperationen.
- Neue Installation und Migration verwenden den offiziellen Seeder als einmaligen
  Compose-Tools-Dienst. Hocuspocus erhält eine eigene WebSocket-Route und einen
  instanzspezifischen internen Web-Hostnamen.
- Die gemeinsame Traefik-Konfiguration bleibt Betreiberaufgabe; Wildcard-Metadaten
  werden im bestehenden Hostvars-Format gespeichert. Die Skripte verändern den
  globalen Proxy nicht automatisch.
- Review-Korrektur: Update-/Restore-Backups lassen schreibende Dienste nach dem
  Snapshot gestoppt. Bei Backup-Fehlern wird ihr vorheriger Zustand wiederhergestellt.
- Review-Korrektur: `--recover` ermöglicht nach einem beschädigten Restore die
  erneute Wiederherstellung aus dem ursprünglichen Archiv ohne neuen Dump.
  Der Modus verlangt Wartungsmarker und gestoppte Schreiber einschließlich Seeder.
  `op-add --resume` gilt ausschließlich für fehlgeschlagene Neuinstallationen.
- Der ausführbare Container-Rundlauftest verwendet eine Probezeile in PostgreSQL
  und eine Datei im Attachment-Volume. Ein zusätzlich dokumentierter Browser-Test
  prüft echte Arbeitspakete/Uploads, TLS und Mail. Der Docker-Daemon ist lokal nicht
  verfügbar; diese echten Laufzeittests bleiben ausdrücklich offen.

- Backup sichert den privaten, beim Rendering gespeicherten Hostvars-Stand im
  Instanzverzeichnis. So gelangen noch nicht deployte Änderungen nicht als
  vermeintlich passende Konfiguration in ein Recovery-Archiv.

Abschlussprüfung: siehe [VALIDATION.md](VALIDATION.md).
