# Prüfstand OpenProject-Integration

Stand: 2026-09-29.

## Ausgeführt

- 30 automatisierte Tests: Domain-/Versionsvalidierung, Konfigurationsschema,
  Instanzsperren, Archiv-Roundtrip und manipulierte Archive, Sonderzeichen in
  Secrets, vorhandene Instanzen, Backup-Fehler/Wiederanlauf, Update-Abbruch,
  Migrations-/Importfehler, Wartungs-Recovery und laufende Seeder.
- Echte Ansible-Templates gerendert und mit Docker Compose 2.32.4 validiert;
  Testkonfigurationen enthalten ausschließlich temporäre Testdaten.
- Ansible-Playbook-Syntaxprüfung, ShellCheck, Bash-Syntax, Python-Kompilierung,
  Git-Whitespace-Prüfung.
- Separater Code-Review; Befunde zu Recovery bei beschädigter Datenbank und
  durchgehendem Schreibstopp umgesetzt. Seeder-Guard und Setuid-Archivprüfung
  ergänzt. Backup verwendet den deployten Konfigurationsstand statt bereits
  geänderter gewünschter Hostvars.
- Feste Image-Tags gegen Docker Hub geprüft.

## Noch nicht ausgeführt

Der lokale Docker-Daemon ist nicht erreichbar. Deshalb keine Aussage über einen
bestandenen echten Installations-/Migrations-/Restore-Rundlauf. Dafür liegt
`apps/openproject/tests/integration_roundtrip.sh` bereit. Öffentliche DNS-/TLS-
Erreichbarkeit, Mailversand, UI-Arbeitspakete/Uploads und gemeinsames Bearbeiten
müssen auf dem Ziel-/Testhost geprüft werden. Produktionsdienste wurden nicht
verändert.

## Umsetzungshinweise

Die vollständigen Betriebsbefehle und Recovery-Wege stehen in `../README.md`.
Ansible dient als Renderer; die Python-Hilfsmodule orchestrieren die vier
Bash-Einstiegspunkte unter einer gemeinsamen Sperre je Domain. Die AI Bridge
wurde nicht implementiert.
