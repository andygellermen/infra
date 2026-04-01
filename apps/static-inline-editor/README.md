# Static Inline Editor App

Erstes Scaffold fuer die geplante Go-Begleit-App zum sicheren Inline-Editing statischer Webseiten.

## Stand heute

Aktuell enthalten:

- globale App-Konfiguration
- tenant-spezifische `.env`-Dateien
- Go-Config-Loader
- Basismodell fuer bearbeitbare Domains
- minimaler HTTP-Server
- Healthcheck und Tenant-Debug-Endpunkt

Noch nicht enthalten:

- Login
- Session-Management
- Edit-Modus
- HTML-Markierung mit `data-editor-id`
- Preview und Save
- Backup und Git-Commit

## Lokales Testen

```bash
cd apps/static-inline-editor
go run ./cmd/server
```

Optional mit Tenant-Datei:

```bash
mkdir -p tenants
cat > tenants/example.org.env <<'EOF'
STATIC_EDITOR_TENANT=example.org
STATIC_EDITOR_LOGIN_DOMAIN=bearbeitung.example.org
STATIC_EDITOR_ALIASES=www.example.org
STATIC_EDITOR_STATIC_ROOT=/srv/static/example.org
STATIC_EDITOR_BACKUP_ROOT=/srv/static-backups/example.org
STATIC_EDITOR_REPO_ROOT=/srv/static/example.org
STATIC_EDITOR_USERNAME=admin
STATIC_EDITOR_PASSWORD_HASH=$2y$example
STATIC_EDITOR_COOKIE_SECRET=replace-me
STATIC_EDITOR_MAIN_SELECTOR=main
STATIC_EDITOR_ALLOWED_BLOCK_TAGS=h1,h2,h3,h4,h5,p,ul,ol,li
STATIC_EDITOR_ALLOWED_INLINE_TAGS=strong,em,a,br
STATIC_EDITOR_START_PATH=/index.html
EOF

STATIC_EDITOR_TENANT_DIR=./tenants go run ./cmd/server
```

Dann pruefen:

- `http://localhost:8090/healthz`
- `http://localhost:8090/debug/tenants`

## Naechste Schritte

1. Login und Session-Cookie
2. `GET /edit?path=/index.html`
3. HTML-Parsing und `data-editor-id`
4. `POST /api/preview`
5. `POST /api/save`
