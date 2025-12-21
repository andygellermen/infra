
# 🌀 Ghost Backup & Restore Toolkit

Willkommen im Restore-Tempel deines Ghost CMS Docker-Systems.  
Dieses Toolkit ermöglicht dir die einfache Wiederherstellung gelöschter oder geänderter Ghost-Websites – vollständig automatisiert, abgesichert und protokolliert.

---

## 📜 `ghost-restore.sh`

Wiederherstellung einer Ghost-Instanz aus einem `.tar.gz`-Backup.

### 🔧 Syntax

```bash
./scripts/ghost-restore.sh [domain] [options]
```

---

## 🏷️ Verfügbare Optionen / Flags

| Flag | Beschreibung |
|------|--------------|
| `--force` | Erzwingt die Wiederherstellung und ersetzt eine bereits existierende Instanz ohne Rückfrage. |
| `--dry-run` | Führt keine Wiederherstellung durch. Prüft nur, ob das gewählte Backup vollständig und gültig ist. |
| `--purge` | (Geplant für `ghost-delete.sh`) Entfernt _endgültig_ inkl. Datenbank, Docker-Volume und Hostvars. |
| `--select` | Öffnet ein interaktives Menü zur Auswahl eines Backups aus dem Backup-Ordner. |
| `--help` | Zeigt diese Übersicht an. |

---

## 📂 Backup-Verzeichnisstruktur

Backups werden im folgenden Format abgelegt:

```
infra/backups/ghost/<domain>/<timestamp>.tar.gz
```

### Inhalt eines gültigen Backups:

- Docker Volume (Ghost Content)
- MySQL Dump
- `hostvars/<domain>.yml`

---

## 📓 Logs

Alle Wiederherstellungsaktionen werden protokolliert unter:

```
infra/logs/ghost-restore/<domain>/<timestamp>.log
```

---

## ⚠️ Sicherheit & Hinweise

- Keine Verschlüsselung, kein Passwortschutz: bitte Backups sicher verwahren.
- Die `--dry-run`-Option kann verwendet werden, um Backups vor der Wiederherstellung zu prüfen.
- Im Restore-Prozess wird überprüft, ob `hostvars/<domain>.yml` im Backup enthalten ist. Fehlt diese Datei ➤ Abbruch.

---

Bleibe bei deiner Macht. Restore mit Bedacht.
