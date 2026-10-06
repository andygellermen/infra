# EasyAuthor – geschütztes Staging

Diese Anleitung bereitet ausschließlich das Review-Staging unter
`https://author.geller.men` vor. Sie ist keine öffentliche Produktivfreigabe.

## Voraussetzungen und Ersteinrichtung

- DNS-A-Record zeigt auf den Infra-Host; Traefik und sein Netz `traefik` laufen.
- Benötigt werden Docker, Ansible mit `community.docker`, `curl`, `python3`,
  `htpasswd`, `tar` und ein sauberer Git-Arbeitsbaum.
- Lokale SQLite- oder Bibliotheksdaten werden niemals übernommen.

Auf dem Infra-Host:

```bash
./scripts/easy-author-add.sh author.geller.men --username=andy
./scripts/easy-author-redeploy.sh author.geller.men
```

Die erzeugte Datei `ansible/hostvars/author.geller.men.yml` ist ignoriert,
trägt Modus `0600` und darf nicht committed werden. Das Redeploy baut beide
Images mit dem vollständigen Git-SHA, sichert bestehende Daten und führt nach
Ansible den authentifizierten Smoke aus. Beim allerersten Deploy existieren
noch keine Daten und daher auch kein Vorab-Backup.

## Prüfung und Betrieb

```bash
./scripts/easy-author-smoke-check.sh author.geller.men --username=andy
docker logs easy-author-author-geller-men-web
docker logs easy-author-author-geller-men-api
./scripts/easy-author-backup.sh author.geller.men
ls -l /srv/easy-author/author.geller.men/backups/
```

Der Smoke verlangt ohne Zugang `401`, prüft UI, Health, Projektliste und alle
fünf Kanban-Phasen, bestätigt die fehlende API-Portfreigabe und verifiziert nach
einem API-Neustart dieselbe Projekt-ID.

## Restore, Rotation und Rollback

Ein Restore ersetzt niemals laufende Daten in-place. Er prüft das Archiv,
erstellt ein Sicherheitsbackup und verlangt die Domain zweimal:

```bash
./scripts/easy-author-restore.sh author.geller.men \
  /srv/easy-author/author.geller.men/backups/easy-author-author.geller.men-YYYYMMDDTHHMMSSZ.tar.gz \
  --confirm=author.geller.men
./scripts/easy-author-rotate-secrets.sh author.geller.men --username=andy
./scripts/easy-author-redeploy.sh author.geller.men
```

Für einen Image-Rollback wird der gewünschte geprüfte Git-Stand ausgecheckt
und normal redeployed; der Daten-Rollback erfolgt ausschließlich mit dem
expliziten Restore. Keine Passwörter, bcrypt-Hashes, Hostvars, SQLite-Dateien,
Bibliotheksinhalte oder Backups gehören ins Repository.

## Lokale Validierungsgrenzen

Die vollständige statische und sandbox-basierte Prüfung läuft über:

```bash
bash scripts/verify-easy-author-deployment.sh all
```

Ein realer TLS-/Traefik-Smoke und jede Servermutation erfolgen erst beim
bewussten Staging-Deploy. Der aktuelle lokale Prüfhost stellte während der
Vorbereitung weder einen laufenden Docker-Daemon noch `ansible-playbook` bereit;
Image- und Ansible-Laufzeitprüfungen müssen daher auf dem Infra-Host nachgeholt
werden.

### Readiness-Protokoll vom 6. Oktober 2026

| Prüfung | Ergebnis |
| --- | --- |
| Deployment-Verifikator | 7/7 Gruppen bestanden |
| Shell-Syntax | alle 7 EasyAuthor-Skripte bestanden |
| Frontend | 96/96 Tests bestanden; Vite-Produktionsbuild erfolgreich |
| Backend | alle Go-Pakete bestanden |
| Compose | `docker compose config --quiet` erfolgreich |
| Geheimnis-/Datenscan | keine EasyAuthor-Hostvars, Login-Hashes, Klartextkennwörter, SQLite-, Bibliotheks- oder Backupdaten getrackt |
| Port-/Auth-Modell | API und Web ohne Host-Port; nur Web im Traefik-Netz; HTTPS-Router mit Basic Auth und entferntem Authorization-Header |
| Shellcheck | lokal nicht installiert; durch `bash -n` und Sandbox-Verhaltenstests ersetzt |
| Ansible Syntax/Check | lokal nicht ausführbar, weil `ansible-playbook` fehlt; statischer Rollenvertrag bestanden |
| Image-/Container-Smoke | lokal nicht ausführbar, weil der Docker-Daemon nicht läuft; Dockerfiles und Compose statisch geprüft |

Vor einem realen Deploy sind auf dem Infra-Host deshalb zwingend
`ansible-playbook --syntax-check`, ein Check-Mode-Lauf, beide Image-Builds und
der vollständige authentifizierte Smoke nachzuholen. Bis diese vier Prüfungen
erfolgreich sind, ist der Stand **lokal deployment-ready**, aber noch nicht als
live bereitgestellt oder produktiv freigegeben zu verstehen.
