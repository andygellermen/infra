# Static Inline Editor App

Erstes Scaffold fuer die geplante Go-Begleit-App zum sicheren Inline-Editing statischer Webseiten.

## Stand heute

Aktuell enthalten:

- globale App-Konfiguration
- tenant-spezifische `.env`-Dateien
- Go-Config-Loader
- SMTP-/Magic-Link-Konfigurationsrahmen
- Basismodell fuer bearbeitbare Domains
- file-basierten Magic-Link- und Session-Store
- Magic-Link-Anforderung und Verifikation
- Session-Cookie
- einfache Login- und Startseite
- erster Edit-Endpunkt mit HTML-Markierung
- ContentTools nur im geschuetzten Edit-Fall
- Preview- und Save-Endpunkte
- Backup vor dem Schreiben
- Git-Commit nach erfolgreichem Save
- optionaler Git-Push nach erfolgreichem Save
- minimaler HTTP-Server
- Healthcheck und Tenant-Debug-Endpunkt

Noch nicht enthalten:

- bewusst gesteuerter Push erst bei "Bearbeitung beenden"

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
STATIC_EDITOR_ALLOWED_EMAILS=andy@example.org
STATIC_EDITOR_COOKIE_SECRET=replace-me
STATIC_EDITOR_MAIN_SELECTOR=main
STATIC_EDITOR_ALLOWED_BLOCK_TAGS=h1,h2,h3,h4,h5,p,ul,ol,li
STATIC_EDITOR_ALLOWED_INLINE_TAGS=strong,em,a,br
STATIC_EDITOR_START_PATH=/index.html
EOF

STATIC_EDITOR_SMTP_HOST=email-smtp.eu-central-1.amazonaws.com \
STATIC_EDITOR_SMTP_PORT=587 \
STATIC_EDITOR_SMTP_USERNAME=replace-me \
STATIC_EDITOR_SMTP_PASSWORD=replace-me \
STATIC_EDITOR_SMTP_FROM_EMAIL=no-reply@example.org \
STATIC_EDITOR_MAGIC_LINK_TTL=15m \
STATIC_EDITOR_GIT_AUTHOR_NAME='Static Inline Editor' \
STATIC_EDITOR_GIT_AUTHOR_EMAIL=18n1ylzby6v4t2pmwufj6jsoeeomh9@bots.bitbucket.org \
STATIC_EDITOR_GIT_PUSH_ON_SAVE=true \
STATIC_EDITOR_GIT_REMOTE=origin \
STATIC_EDITOR_GIT_HTTP_USERNAME=x-token-auth \
STATIC_EDITOR_GIT_HTTP_PASSWORD=replace-me \
STATIC_EDITOR_TENANT_DIR=./tenants \
go run ./cmd/server
```

Dann pruefen:

- `http://localhost:8090/healthz`
- `http://localhost:8090/debug/tenants`
- `http://localhost:8090/login`
- `curl -H 'Host: bearbeitung.example.org' -X POST http://localhost:8090/auth/request-link -d 'email=andy@example.org'`
- nach erfolgreichem Magic-Link: `http://localhost:8090/edit?path=/index.html`

## Betriebsgeruest

Neu vorbereitet sind ausserdem:

- Dockerfile fuer `static-inline-editor:latest`
- Ansible-Playbook `ansible/playbooks/deploy-static-inline-editor.yml`
- Ansible-Rolle `ansible/playbooks/roles/static-inline-editor/tasks/main.yml`
- Hostvars-Vorlage `ansible/hostvars/templates/static-inline-editor-hostvars.j2`
- Helferskript `scripts/staticeditor-print-hostvars.sh`
- Redeploy-Skript `scripts/staticeditor-redeploy.sh`

Fuer bestehende statische Domains ist der einfachste Weg:

1. Hostvars-Block erzeugen:
   `./scripts/staticeditor-print-hostvars.sh example.org andy@example.org`
2. Den ausgegebenen Block in `ansible/hostvars/example.org.yml` uebernehmen.
3. Danach das gemeinsame Editor-Image deployen:
   `./scripts/staticeditor-redeploy.sh example.org`

Wichtig:

- `static_editor_login_domain` sollte auf die Bearbeitungsdomain zeigen, zum Beispiel `bearbeitung.example.org`
- `static_editor_repo_root` sollte auf das Git-Repo der statischen Seite zeigen
- `static_editor_static_root` und `static_editor_repo_root` duerfen identisch sein
- fuer Bitbucket-Repository-Tokens kann `origin` als normale HTTPS-URL bestehen bleiben; der Editor sendet Username und Token ueber `STATIC_EDITOR_GIT_HTTP_USERNAME` und `STATIC_EDITOR_GIT_HTTP_PASSWORD`
- fuer Bitbucket sollte `STATIC_EDITOR_GIT_AUTHOR_EMAIL` auf die von Bitbucket ausgegebene Bot-Adresse gesetzt werden

## Naechste Schritte

1. optionaler "Bearbeitung beenden"-Knopf statt sofortigem Push
2. Dateiliste oder Start-Dashboard fuer mehrere Seiten
3. bessere Preview-Abnahme ohne Browser-Alert
4. feinere Sanitization-Regeln fuer Links und Spezialfaelle
5. Rollback aus vorhandenen Backups
