# 🌀 Ghost Backup Restore (`scripts/ghost-restore.sh`)

Helper-Skript für den Restore eines Legacy-Backups (`ghost backup`) in eine **bestehende** Ghost-Instanz der Infra.

## Voraussetzungen

- Zielinstanz existiert bereits (`ghost-<domain>` Container + Hostvars)
- Backup-ZIP liegt auf dem Zielserver
- Hostvars enthalten:
  - `ghost_domain_db`
  - `ghost_domain_usr`
  - `ghost_domain_pwd`
  - optional `ghost_version`
- Tools: `docker`, `unzip`, `awk`, `sed`, `grep` (optional: `rg` für schnellere Mustererkennung)

## Verwendung

```bash
# 1) Verfügbare Ghost-Container anzeigen
./scripts/ghost-restore.sh --list

# 2) Validierung ohne Änderungen
./scripts/ghost-restore.sh blog.example.com /backups/ghost-backup.zip --dry-run

# 3) Restore durchführen
./scripts/ghost-restore.sh blog.example.com /backups/ghost-backup.zip --yes

# 4) Nur content/ einspielen (wenn Backup keine SQL enthält)
./scripts/ghost-restore.sh blog.example.com /backups/ghost-backup.zip --content-only --yes
```

## Was das Skript macht

1. Prüft ZIP-Integrität (`unzip -t`)
2. Entpackt Backup in ein Temp-Verzeichnis
3. Validiert, dass SQL-Dump + `content/` existieren
4. Liest DB-Credentials aus `ansible/hostvars/<domain>.yml`
5. Liest Quellversion aus `data/content-from-v*-on-*.json` (wenn vorhanden)
6. Prüft Major-Version Quelle vs. Ziel (Abbruch bei Mismatch, außer `--allow-major-mismatch`)
7. Erstellt Safety-Backups (DB + Content)
8. Stoppt Ghost-Container
9. Leert Ziel-DB und importiert SQL
10. Leert Ghost-Volume und kopiert `content/`
11. Startet Ghost neu

## Optionen

- `--list`: Zeigt bestehende Ghost-Container (`docker ps -a`) 
- `--dry-run`: Nur Checks, keine Änderung
- `--yes`: Keine Rückfrage
- `--allow-major-mismatch`: Restore trotz Versions-Major-Mismatch erzwingen
- `--content-only`: Nur `content/` einspielen (DB bleibt unverändert)

## Hinweise

- Standardmäßig wird bei Versions-Major-Mismatch abgebrochen.
- Safety-Backups landen unter:

```bash
/tmp/ghost-restore-safety/<domain>/<timestamp>/
```


## Warum kam „Keine SQL-Datei im Backup gefunden“?

Viele `ghost backup`-Archive enthalten einen JSON-Export unter `data/*.json` statt eines SQL-Dumps.
Das Skript meldet das jetzt klar und bietet zwei Wege:

1. `--content-only` nutzen (nur Dateien wie Images/Themes einspielen),
2. JSON im Ghost-Admin importieren (`/ghost` → Settings → Labs → Import content).

## Troubleshooting: `syntax error near unexpected token '<<<'`

Dieser Fehler kommt typischerweise von einem nicht sauber aufgelösten Merge-Konflikt
(`<<<<<<<`, `=======`, `>>>>>>>`) in `scripts/ghost-restore.sh`.

Prüfen:

```bash
rg -n "^<<<<<<<|^=======|^>>>>>>>" ./scripts/ghost-restore.sh
```

Wenn Treffer vorhanden sind, die Datei auf den aktuellen Stand zurücksetzen:

```bash
git checkout -- ./scripts/ghost-restore.sh
chmod +x ./scripts/ghost-restore.sh
bash -n ./scripts/ghost-restore.sh
```

Danach den Dry-Run erneut starten:

```bash
./scripts/ghost-restore.sh <domain> <backup.zip> --dry-run
```
