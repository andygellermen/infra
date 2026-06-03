# easy-author – Infra & Deployment

## 1. Ziel

easy-author soll in den bestehenden Infra-Stack passen: Docker, Traefik, Portainer, Ansible und perspektivisch eigene Mandanten-/SaaS-Fähigkeit.

## 2. Zielarchitektur

```text
Internet
  ↓
Traefik
  ↓
easy-author-web
  ↓
easy-author-api
  ↓
easy-author-worker
  ↓
Storage / DB / Export Tools
```

## 3. Container

### easy-author-web

- React/Vite/Next optional
- Editor UI
- Leseräume
- Admin UI

### easy-author-api

- Go Backend
- REST/JSON oder später GraphQL
- Auth
- Projekte
- Kapitel
- Assets
- Kommentare
- Exporte

### easy-author-worker

- Export-Jobs
- Bildoptimierung
- Sync-Jobs
- Qualitätsprüfungen

### easy-author-db

MVP:

- SQLite möglich

Produktiv/SaaS:

- PostgreSQL empfohlen

### easy-author-storage

- lokales Volume
- später S3/MinIO

### easy-author-rclone optional

- Cloud-Sync
- WebDAV
- Google Drive
- OneDrive
- Dropbox

## 4. Volumes

```yaml
volumes:
  easy_author_projects:
  easy_author_assets:
  easy_author_exports:
  easy_author_db:
```

## 5. Traefik Labels

Beispielhafte Zielstruktur:

```yaml
traefik.enable: true
traefik.http.routers.easy-author.rule: Host(`author.example.com`)
traefik.http.routers.easy-author.entrypoints: websecure
traefik.http.routers.easy-author.tls.certresolver: letsencrypt
```

## 6. Authentifizierung

MVP:

- E-Mail + Passwort oder Magic Link

Später:

- Passkeys
- OAuth optional
- Mandantenlogin
- rollenbasierte Rechte

## 7. Storage-Modell

Server-Speicher ist die Wahrheit.

Browser-Speicher:

- IndexedDB für Offline-Entwürfe
- Autosave-Fallback
- Konfliktprüfung beim Reconnect

## 8. Backup

Mindestens:

- Datenbankdump
- Projektordner
- Assets
- Exportprofile
- Themes

Optional:

- Git-Commits pro Projekt
- tägliche Snapshots
- Remote-Backup via Restic/Rclone

## 9. Sicherheit

- Cloud-Tokens verschlüsseln
- Export-Worker isolieren
- Upload-Dateien prüfen
- Dateipfade normalisieren
- keine direkten Shell-Aufrufe aus Userdaten
- Rollen konsequent prüfen

## 10. Entwicklungsumgebung

```text
/docker-compose.dev.yml
/frontend
/backend
/worker
/docs
```

## 11. Produktionsumgebung

Deployment via Ansible:

- Hostvars pro Instanz
- Domain
- Datenbankmodus
- Storage-Pfad
- SMTP/SES optional
- Backup-Ziele
- Traefik-Router

## 12. Vermeidungsstrategie

Nicht sofort SaaS-Komplexität erzwingen. Zuerst Single-Instance/self-hosted stabilisieren. Danach Mandantenfähigkeit vorbereiten und erst später abrechenbare Plattformlogik aktivieren.
