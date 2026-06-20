# WordPress Incident Report – `heimannkunst.de`

## Zweck

Diese Datei dokumentiert den Sicherheitsvorfall rund um `heimannkunst.de`, die forensischen Erkenntnisse, die Bereinigungsschritte und die bereits umgesetzten Härtungsmaßnahmen.

Sie ist bewusst als eigenständige Incident-Dokumentation gedacht und ergänzt das generische Runbook in `README-wp-incident-response.md`.


## Kurzfazit

- Die WordPress-Instanz war kompromittiert.
- Ein Weiterbetrieb des alten Bestands war nicht vertretbar.
- Die saubere Wiederherstellung erfolgte deshalb als **Neuaufbau** statt als Live-Bereinigung.
- Die neue Instanz läuft wieder, `xmlrpc.php` ist auf Traefik-Ebene geblockt und Theme/Plugins wurden aus vertrauenswürdigen Quellen frisch eingespielt.
- Der zusätzliche Browser-Passwortschutz bleibt bis zur Kundenabnahme aktiv.


## Betroffene Instanz

- Domain: `heimannkunst.de`
- Anwendung: WordPress im Docker-Stack hinter Traefik
- Status vor Bereinigung: kompromittiert
- Status nach Bereinigung: neu aufgebaut, funktional getestet, noch mit zusätzlichem Passwortschutz


## Zeitlinie

Die folgenden Zeitpunkte wurden im Verlauf der Untersuchung sichtbar:

- **7. Juni 2026, ca. 17:10 Uhr**  
  Auffälligkeiten rund um `insert-headers-and-footers` bzw. `WPCode`-Artefakte; im Webroot/Upload-Kontext entstanden passende Dateien und Verzeichnisse.

- **13. Juni 2026, ca. 13:25 Uhr**  
  Eine Manipulation an `wp-content/themes/Avada/header.php` wurde zeitlich eingegrenzt.

- **20. Juni 2026**  
  Die kompromittierte Produktivinstanz wurde isoliert, ein forensisches Backup erzeugt und die Bereinigung/Neuinstallation durchgeführt.


## Bestätigte Indikatoren

Folgende Indikatoren wurden im Incident-Verlauf bestätigt:

- `xmlrpc.php` war von außen erreichbar und nicht vorgeschaltet blockiert.
- `wp-blog-header.php` war manipuliert und enthielt obfuskierten PHP-/JavaScript-Schadcode.
- `wp-content/themes/Avada/header.php` war verändert.
- Im Upload-Kontext existierten `WPCode`-/Snippet-Artefakte.
- In der Datenbank wurden zugehörige Spuren aus dem Snippet-/Code-Umfeld untersucht.
- Ein kompromittierter Altbestand ließ sich nicht vertrauenswürdig „im Betrieb“ säubern.


## Wahrscheinliche Angriffspfade

Die Untersuchung spricht für eines oder mehrere dieser Szenarien:

1. Missbrauch eines WordPress-Admin-Zugangs oder einer kompromittierten Session
2. Brute-Force/Credential-Stuffing gegen `xmlrpc.php` oder `wp-login.php`
3. Nachgelagerter Upload oder Austausch von Theme-/Plugin-Dateien nach erfolgreicher Anmeldung
4. Missbrauch eines Code-/Snippet-Mechanismus im Admin-Bereich

Wichtig:

- Das Vorhandensein eines Plugins wie `insert-headers-and-footers` bzw. `WPCode` ist **nicht automatisch** ein Beweis, dass genau dieses Plugin selbst die Schwachstelle war.
- Im Incident war aber gerade diese Funktionsklasse besonders relevant, weil sie Code-Injektion im Admin-Kontext erleichtert.


## Forensische Maßnahmen

Durchgeführt bzw. bestätigt:

- kompromittierte Instanz isoliert
- Maintenance-Stack vorgeschaltet
- betroffenen WordPress-Container gestoppt
- forensisches Backup erstellt:

```bash
./scripts/wp-backup.sh --create heimannkunst.de
```

- Offline-Analyse in einem getrennten Arbeitsverzeichnis
- Traefik-/WordPress-/Datei-/DB-Spuren ausgewertet
- Änderungszeitpunkte einzelner Dateien eingegrenzt
- Schadcode-Funde im Webroot und Theme-Kontext dokumentiert


## Bereinigungsstrategie

Die Bereinigung erfolgte bewusst **nicht** durch Reparatur des kompromittierten Bestands, sondern durch Neuaufbau:

1. saubere WordPress-Instanz neu deployt
2. inhaltlich geprüften Datenbankstand aus März 2026 als saubere Basis verwendet
3. Theme `Avada` aus vertrauenswürdiger Quelle frisch installiert
4. Premium-Plugins aus vertrauenswürdiger Quelle frisch installiert
5. alte kompromittierte Theme-/Plugin-Bestände **nicht** blind übernommen
6. Frontend, Shop und Seitenstruktur funktional getestet


## Technische Fixes im Infra-Stack

Im Zuge des Incidents wurden mehrere Stack-Verbesserungen umgesetzt:

- `xmlrpc.php`-Blockade direkt in Traefik
- robustere WordPress-Hostvar-Normalisierung für Primärdomain/Aliase
- Korrektur einer fehlerhaften Traefik-Redirect-Regex
- Entfernen eines ungültigen `ipallowlist.rejectstatuscode`-Labels
- PHP-Upload-Limits für größere Theme-/Plugin-ZIPs
- Deaktivierung des WordPress-Dateieditors per `DISALLOW_FILE_EDIT`
- Wazuh-Agent-Rolle für Datei-Integrität und Host-Monitoring
- CrowdSec-Mail-Alerting plus JSON-Alert-Datei für Wazuh/SIEM


## Aktueller Sicherheitsstatus

Stand nach Wiederherstellung:

- Website erreichbar
- Browser-Passwortschutz aktiv
- `xmlrpc.php` liefert extern `403`
- Theme/Plugins neu installiert
- WordPress-Admin-Passwörter erneuert bzw. Rotation angestoßen
- Funktionstest für Frontend/Shop/Seitenstruktur erfolgreich


## Noch offene oder empfohlene Folgearbeiten

Die folgenden Punkte sind nach so einem Incident weiterhin wichtig:

- Kundenabnahme der reparierten Seite
- zusätzlicher Passwortschutz erst danach entfernen
- **WordPress-Salts** neu setzen
- alle **Application Passwords**, API-Keys, SMTP-Zugänge, WooCommerce-/Webhook-Secrets und ähnliche Integrationen prüfen bzw. rotieren
- alle nicht mehr benötigten Admin- oder Editor-Accounts deaktivieren
- falls verfügbar: MFA für Admin-Zugänge aktivieren
- frischen **sauberen** Backup-Stand nach Abnahme erzeugen
- Wazuh-/CrowdSec-Testlauf mit echten Testereignissen durchführen
- Härtungsänderungen auf alle übrigen WordPress-Instanzen ausrollen


## Lessons Learned

- kompromittierte WordPress-Instanzen sollten im Zweifel **neu aufgebaut** werden
- `xmlrpc.php` sollte standardmäßig geblockt oder nur sehr eng allowlisted werden
- Code-/Snippet-Funktionen im Admin-Kontext brauchen besondere Aufmerksamkeit
- Datei-Integrität auf Docker-Volumes ist für WordPress-Forensik und Früherkennung sehr wertvoll
- reine Verfügbarkeit reicht nicht aus; entscheidend sind auch Alarmierung und nachvollziehbare Zeitlinien


## Verwandte Dokumente

- `README-wp-incident-response.md`
- `README-wp-extension.md`
- `README-wazuh-extension.md`
- `README-wazuh-test-runbook.md`
