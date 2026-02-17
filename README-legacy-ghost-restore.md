# Restore eines Legacy-Backups (`ghost backup`) in eine vorbereitete Ghost-Docker-Instanz (Infra)

Dieses Runbook beschreibt den Restore-Pfad für folgenden Fall:

- **Quelle:** Legacy Ghost `v4.x` auf Ubuntu (ohne Docker), Backup erstellt mit `ghost backup`
- **Ziel:** Bereits vorbereitete Infra-Instanz mit Ghost `v4.x` im Docker-Container

> Voraussetzung: Domain/Hostvars/Secrets und die Zielinstanz wurden bereits angelegt.

## 1) Zielparameter setzen

```bash
DOMAIN="blog.example.com"
BACKUP_ZIP="/pfad/zum/backup/ghost-backup-YYYY-MM-DD-HH-mm-ss.zip"
WORKDIR="/tmp/ghost-restore-${DOMAIN}"
```

Die Infra-Namen sind:

- Ghost-Container: `ghost-${DOMAIN//./-}`
- Content-Volume: `ghost_${DOMAIN//./_}_content`
- MySQL-Container: `ghost-mysql`

## 2) Backup entpacken und Struktur prüfen

```bash
mkdir -p "$WORKDIR"
unzip -o "$BACKUP_ZIP" -d "$WORKDIR"
find "$WORKDIR" -maxdepth 4 -type f | sort
```

Erwartet werden mindestens:

- ein SQL-Dump (z. B. `*.sql`)
- `content/` (Images, Themes, ggf. `settings/routes.yaml`)

## 3) Ziel-DB-Werte aus Hostvars laden

```bash
HOSTVARS="./ansible/hostvars/${DOMAIN}.yml"
DB_NAME=$(awk -F': ' '/^ghost_domain_db:/ {print $2}' "$HOSTVARS")
DB_USER=$(awk -F': ' '/^ghost_domain_usr:/ {print $2}' "$HOSTVARS")
DB_PASS=$(awk -F': ' '/^ghost_domain_pwd:/ {print $2}' "$HOSTVARS")

echo "DB_NAME=$DB_NAME"
echo "DB_USER=$DB_USER"
```

## 4) Ghost kurz stoppen (konsistenter Restore)

```bash
docker stop "ghost-${DOMAIN//./-}"
```

## 5) DB-Dump in Ziel-DB importieren

SQL-Datei im entpackten Backup suchen:

```bash
SQL_FILE=$(find "$WORKDIR" -type f -name '*.sql' | head -n1)
[ -n "$SQL_FILE" ] || { echo "Kein SQL-Dump gefunden"; exit 1; }

echo "Importiere: $SQL_FILE"
cat "$SQL_FILE" | docker exec -i ghost-mysql \
  mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"
```

## 6) Content in das Docker-Volume kopieren

Pfad `content/` aus dem Backup in das Volume übernehmen:

```bash
CONTENT_DIR=$(find "$WORKDIR" -type d -name content | head -n1)
[ -n "$CONTENT_DIR" ] || { echo "Kein content/-Ordner gefunden"; exit 1; }

VOLUME_NAME="ghost_${DOMAIN//./_}_content"

docker run --rm \
  -v "${VOLUME_NAME}:/target" \
  -v "${CONTENT_DIR}:/source:ro" \
  alpine sh -c 'cp -a /source/. /target/'
```

## 7) Ghost starten und prüfen

```bash
docker start "ghost-${DOMAIN//./-}"
docker logs --tail=100 "ghost-${DOMAIN//./-}"
```

Danach im Browser prüfen:

- Frontend lädt
- `/ghost` Login möglich
- Beiträge/Bilder/Seiten vorhanden

## 8) Troubleshooting (typisch)

- **`ER_BAD_DB_ERROR` / Login-Fehler:** Hostvars-Werte und importierte DB stimmen nicht überein.
- **Leere Seite/fehlende Bilder:** `content/` wurde nicht vollständig in das Volume kopiert.
- **Versionsthema:** Quelle und Ziel müssen auf Ghost `v4.x` bleiben.

## 9) Optional: Rollback-Vorsorge vor Restore

Vor dem Import ein Sicherungsdump der Ziel-DB erstellen:

```bash
docker exec ghost-mysql mysqldump -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" > "/tmp/${DOMAIN}-pre-restore.sql"
```

Und Content-Volume archivieren:

```bash
docker run --rm -v "ghost_${DOMAIN//./_}_content:/data" -v /tmp:/backup alpine \
  sh -c 'tar czf /backup/'"${DOMAIN}"'-content-pre-restore.tar.gz -C /data .'
```
