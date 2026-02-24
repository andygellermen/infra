# 📘 Dokumentation der Ghost-Infra-Skripte

### ghost-add.sh

**Beschreibung:**  
Dieses Skript erstellt eine neue Ghost-Instanz inklusive Docker-Container, Datenbankeintrag, Zertifikatseinrichtung über Traefik sowie der passenden `hostvars` Datei.

**Syntax:**
```bash
./scripts/ghost-add.sh DOMAIN [ALIAS] [--version=<major|major.minor|major.minor.patch|latest>]
```

**Parameter:**
- `DOMAIN` – Die Hauptdomain, z. B. `blog.example.com`
- `ALIAS` – (optional) Alias-Domain, z. B. `www.blog.example.com`
- `--version=<major|major.minor|major.minor.patch|latest>` – (optional) setzt die gewünschte Ghost-Version, z. B. `--version=4`, `--version=6.18` oder `--version=6.18.2`; ohne Angabe wird `latest` verwendet

**Vorgänge:**
- Docker-Container mit Labels für Traefik wird erzeugt
- Datenbank wird erstellt
- `hostvars/DOMAIN.yml` inkl. ALIAS wird automatisch generiert
- Zertifikat via Let's Encrypt wird beantragt


### ghost-upgrade.sh

**Beschreibung:**  
Hebt eine bestehende Ghost-Instanz auf eine neue Version an, indem `ghost_version` in den Hostvars angepasst und anschließend das Deployment neu ausgeführt wird.

**Syntax:**
```bash
./scripts/ghost-upgrade.sh DOMAIN --version=<major|major.minor|major.minor.patch|latest> [--force-major-jump] [--dry-run]
```

**Parameter:**
- `DOMAIN` – Die bestehende Ghost-Domain
- `--version=<major|major.minor|major.minor.patch|latest>` – Zielversion, z. B. `--version=5`, `--version=6.18` oder `--version=6.18.2`
- `--force-major-jump` – erlaubt Sprünge größer als +1 Major-Version
- `--dry-run` – schreibt nur die Hostvars (inkl. Backup), ohne Deployment

**Features:**
- Liest und validiert die aktuelle `ghost_version`
- Verhindert standardmäßig große Versionssprünge (z. B. 4 → 6)
- Erstellt automatisch ein Backup der Hostvars-Datei (`.bak.<timestamp>`)
- Führt danach ein reguläres `ansible-playbook` Deployment aus

### ghost-delete.sh

**Beschreibung:**  
Dieses Skript entfernt eine bestehende Ghost-Instanz inklusive Datenbank und Hostvars. Optional mit Backup & vollständigem Löschen.

**Syntax:**
```bash
./scripts/ghost-delete.sh DOMAIN [--purge]
```

**Parameter:**
- `DOMAIN` – Die zu entfernende Ghost-Domain
- `--purge` – (optional) löscht alle zugehörigen Daten, inkl. Backups

**Features:**
- Sicheres Entfernen des Containers und der DB
- Optionaler Backup vor Löschung
- Log-Eintrag in `/logs`
- Interaktive Bestätigung bei gefährlichen Operationen

### create-hostvars.sh

**Beschreibung:**  
Erstellt eine passende `hostvars` Datei für eine neue Ghost-Domain automatisch.

**Syntax:**
```bash
./scripts/create-hostvars.sh DOMAIN [ALIAS] [--version=<major|major.minor|major.minor.patch|latest>]
```

**Parameter:**
- `DOMAIN` – Hauptdomain
- `ALIAS` – (optional) Aliasdomain
- `--version=<major|major.minor|major.minor.patch|latest>` – (optional) gewünschte Ghost-Version für den Container-Tag; ohne Angabe wird `latest` verwendet

**Features:**
- Validiert Eingaben (inkl. Punycode bei Umlauten)
- Schreibt in `ansible/hostvars`
- Warnung bei bestehenden Dateien

### Amazon SES (Standard für Ghost-Mail)

Die Ghost-Container verwenden standardmäßig Amazon SES als SMTP-Transport. Lege die Zugangsdaten einmalig in `ansible/secrets/secrets.yml` ab, damit sie bei jeder Neuanlage automatisch genutzt werden. Beispiel:

```yaml
ghost_ses_smtp_user: "AKIA...SMTP_USER"
ghost_ses_smtp_password: "DEIN_SMTP_PASSWORT"
ghost_ses_from: "Ghost <noreply@deine-domain.tld>"

# Optional (Defaults werden verwendet, wenn nicht gesetzt)
ghost_ses_smtp_host: "email-smtp.eu-central-1.amazonaws.com"
ghost_ses_smtp_port: 587
ghost_ses_smtp_secure: false
```

Wenn `ghost_ses_from` nicht gesetzt ist, wird automatisch `noreply@<domain>` verwendet. Individuelle Abweichungen kannst du pro Instanz im jeweiligen `ansible/hostvars/<domain>.yml` überschreiben.


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


**Hinweis zu Node.js:**
Die Node.js-Version wird automatisch durch das gewählte offizielle Ghost-Docker-Image bestimmt (z. B. `ghost:4`, `ghost:5`, `ghost:6`). Dadurch ist immer die zur Ghost-Version passende Node-Laufzeit enthalten.

### Ghost-Version auf nächste Major-Version anheben

1. Hostvars der Domain anpassen (`ansible/hostvars/<domain>.yml`):
   ```yaml
   ghost_version: "5"
   ```
2. Deployment erneut ausführen:
   ```bash
   ./scripts/ghost-add.sh <domain> --version=5
   ```
   oder alternativ direkt:
   ```bash
   ansible-playbook -i ./ansible/inventory -e "target_domain=<domain>" ./ansible/playbooks/deploy-ghost.yml
   ```
3. Anschließend Ghost-Admin unter `/ghost` prüfen und ggf. Migrationshinweise im Dashboard bestätigen.

### infra-setup.sh

**Beschreibung:**  
Initialisiert den kompletten Infra-Stack auf einem frischen Host und installiert fehlende Basis-Tools automatisch (Docker, Ansible, MySQL-Client, dnsutils, jq, Python 3, community.docker Collection).

**Syntax:**
```bash
sudo ./scripts/infra-setup.sh
```

**Funktionen:**
- Fragt die Portainer-Domain interaktiv ab.
- Deployt nacheinander MySQL, Traefik, CrowdSec und Portainer via Ansible.
- Prüft den A-Record der Portainer-Domain gegen die öffentliche Host-IP und bricht bei Abweichung ab.
- Richtet CrowdSec + Traefik-Bouncer ein, inkl. Middleware-Namen:
  - `crowdsec-default@docker`
  - `crowdsec-admin@docker`
  - `crowdsec-api@docker`

**Hinweis für weitere Container (z. B. WordPress / statische Seiten):**
- Hänge die passenden Traefik-Router an mindestens `crowdsec-default@docker`.
- Für WordPress-Backend explizit zusätzliche Router für `/wp-admin` und `/wp-login.php` mit `crowdsec-admin@docker` verwenden.
- Für APIs (Ghost/WordPress) Router mit `crowdsec-api@docker` verwenden.
