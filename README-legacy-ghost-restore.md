# Schritt-für-Schritt: Legacy-`ghost backup` (v4) in vorbereitete Infra-Ghost-Docker-Instanz einspielen

Dieses Runbook ist für genau deinen Fall:

- **Quelle:** Legacy Ghost `v4.x` auf Ubuntu (kein Docker), Backup via `ghost backup`
- **Ziel:** Bereits vorbereitete Ghost-Instanz in der Infra (Docker, ebenfalls `v4.x`)

> Ja: Du musst die Backup-Daten in die Ziel-DB **und** in den `content`-Bereich (Docker-Volume) der Zielinstanz einspielen.

---

## 0) Voraussetzungen kurz prüfen

```bash
# Im Infra-Repo ausführen
cd /workspace/infra

DOMAIN="blog.example.com"
BACKUP_ZIP="/pfad/zum/ghost-backup-YYYY-MM-DD-HH-mm-ss.zip"
WORKDIR="/tmp/ghost-restore-${DOMAIN}"

GHOST_CONTAINER="ghost-${DOMAIN//./-}"
VOLUME_NAME="ghost_${DOMAIN//./_}_content"
MYSQL_CONTAINER="ghost-mysql"
HOSTVARS="./ansible/hostvars/${DOMAIN}.yml"
```

Pfad prüfen:

```bash
[ -f "$BACKUP_ZIP" ] || { echo "Backup-ZIP nicht gefunden: $BACKUP_ZIP"; exit 1; }
[ -f "$HOSTVARS" ] || { echo "Hostvars fehlen: $HOSTVARS"; exit 1; }
```

---

## 1) Backup auf technische Validität prüfen (wichtig)

### 1.1 ZIP-Integrität prüfen

```bash
unzip -t "$BACKUP_ZIP"
```

Erwartung: `No errors detected`.

### 1.2 Backup entpacken

```bash
rm -rf "$WORKDIR"
mkdir -p "$WORKDIR"
unzip -o "$BACKUP_ZIP" -d "$WORKDIR"
find "$WORKDIR" -maxdepth 4 -type f | sort
```

### 1.3 Pflichtinhalte prüfen

```bash
SQL_FILE=$(find "$WORKDIR" -type f -name '*.sql' | head -n1)
CONTENT_DIR=$(find "$WORKDIR" -type d -name content | head -n1)

[ -n "$SQL_FILE" ] || { echo "Kein SQL-Dump im Backup gefunden"; exit 1; }
[ -n "$CONTENT_DIR" ] || { echo "Kein content/-Ordner im Backup gefunden"; exit 1; }

# SQL darf nicht leer sein
[ -s "$SQL_FILE" ] || { echo "SQL-Dump ist leer"; exit 1; }

# Schneller Plausibilitätscheck auf typische Ghost-Tabellen
rg -n "CREATE TABLE.*posts|CREATE TABLE.*users|INSERT INTO.*posts" "$SQL_FILE" || {
  echo "Warnung: SQL enthält keine offensichtlichen Ghost-Tabellen (manuell prüfen)";
}

echo "SQL_FILE=$SQL_FILE"
echo "CONTENT_DIR=$CONTENT_DIR"
```

---

## 2) Ziel-DB-Credentials aus Hostvars laden

```bash
DB_NAME=$(awk -F': ' '/^ghost_domain_db:/ {print $2}' "$HOSTVARS")
DB_USER=$(awk -F': ' '/^ghost_domain_usr:/ {print $2}' "$HOSTVARS")
DB_PASS=$(awk -F': ' '/^ghost_domain_pwd:/ {print $2}' "$HOSTVARS")

[ -n "$DB_NAME" ] || { echo "ghost_domain_db fehlt in $HOSTVARS"; exit 1; }
[ -n "$DB_USER" ] || { echo "ghost_domain_usr fehlt in $HOSTVARS"; exit 1; }
[ -n "$DB_PASS" ] || { echo "ghost_domain_pwd fehlt in $HOSTVARS"; exit 1; }

echo "DB_NAME=$DB_NAME"
echo "DB_USER=$DB_USER"
```

DB-Verbindung testen:

```bash
docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" -e "SELECT 1" "$DB_NAME"
```

---

## 3) Safety-Net (dringend empfohlen)

Vor dem eigentlichen Restore die aktuelle Zielinstanz sichern:

```bash
# DB-Backup vom aktuellen Zielzustand
docker exec "$MYSQL_CONTAINER" mysqldump -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" > "/tmp/${DOMAIN}-pre-restore.sql"

# Content-Backup vom aktuellen Zielzustand
docker run --rm \
  -v "${VOLUME_NAME}:/data" \
  -v /tmp:/backup \
  alpine sh -c 'tar czf /backup/'"${DOMAIN}"'-content-pre-restore.tar.gz -C /data .'
```

---

## 4) Ghost stoppen (konsistent einspielen)

```bash
docker stop "$GHOST_CONTAINER"
```

---

## 5) Datenbank sauber neu befüllen

Damit keine Alt-Datenreste bleiben, zuerst DB leeren, dann importieren:

```bash
# Alle Tabellen der Ziel-DB droppen
docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -Nse "
SET FOREIGN_KEY_CHECKS=0;
SELECT CONCAT('DROP TABLE IF EXISTS `', table_name, '`;')
FROM information_schema.tables
WHERE table_schema='${DB_NAME}';
" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

# SQL importieren
cat "$SQL_FILE" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"
```

Import validieren:

```bash
docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" -e "SHOW TABLES;" "$DB_NAME" | head -n 20
```

---

## 6) `content/` in Ghost-Volume einspielen

Ja, genau hier landen die Datei-Daten aus dem Backup (Images, Themes, usw.).

Wichtig: Zielvolume zuerst leeren, damit keine alten Dateien übrig bleiben.

```bash
docker run --rm \
  -v "${VOLUME_NAME}:/target" \
  alpine sh -c 'rm -rf /target/* /target/.[!.]* /target/..?* || true'

docker run --rm \
  -v "${VOLUME_NAME}:/target" \
  -v "${CONTENT_DIR}:/source:ro" \
  alpine sh -c 'cp -a /source/. /target/'

# Kurz prüfen

docker run --rm \
  -v "${VOLUME_NAME}:/target" \
  alpine sh -c 'find /target -maxdepth 3 -type d | sort | head -n 40'
```

---

## 7) Ghost starten und Anwendung prüfen

```bash
docker start "$GHOST_CONTAINER"
docker logs --tail=150 "$GHOST_CONTAINER"
```

Dann manuell testen:

1. Frontend lädt ohne 50x
2. `/ghost` Login klappt
3. Beiträge/Seiten/Tags vorhanden
4. Bilder im Frontend rendern korrekt
5. Falls genutzt: Theme aktiv und funktionsfähig

---

## 8) Häufige Fehlerbilder

- **`ER_ACCESS_DENIED_ERROR`**: DB-Credentials in Hostvars stimmen nicht mit Ziel-DB/User überein.
- **Leere/alte Inhalte**: `content` wurde nicht korrekt geleert/kopiert.
- **Ghost startet, aber Fehler in Logs**: auf DB-Migrations-/Schemafehler prüfen; Quelle und Ziel wirklich beide v4 halten.

---

## 9) Minimal-Checkliste (wenn es schnell gehen muss)

1. ZIP-Integrität: `unzip -t`
2. SQL + `content/` vorhanden
3. Ziel-DB-Creds aus Hostvars geladen
4. Ghost stoppen
5. DB leeren + SQL importieren
6. Volume leeren + `content/` kopieren
7. Ghost starten + Logs + `/ghost` prüfen


---

## 10) FAQ: Ghost-CLI (`ghost backup`, `ghost restore`, Updates)

### Nutzt Ghost 4.x noch Ghost-CLI?
Ja — auf **klassischen VM/Bare-Metal-Installationen** (also dort, wo Ghost per CLI installiert wurde), ist Ghost-CLI weiterhin üblich.

### Gibt es in Ghost-CLI einen vollwertigen `ghost restore` Befehl?
In der Praxis für produktive Komplett-Restores (DB + `content/`) ist der zuverlässige Weg weiterhin der manuelle Restore von

1. Datenbank-Dump und
2. `content/`-Dateien,

so wie in diesem Runbook beschrieben.

### Brauchen wir beim Update auf Folge-Versionen noch Ghost-CLI?
In eurer Infra **nicht zwingend**, weil die Instanz als Docker-Container betrieben wird und die Version über das Image-Tag gesteuert wird.

- Container-Image wird über `ghost_version` gesetzt (z. B. `ghost:4`, `ghost:5`).
- Deployment/Update läuft über euer Ansible-/Script-Setup, nicht über eine lokale CLI-Installation im Container.

### Ist Ghost-CLI hier eine Erleichterung?
Für eure Docker-Infra ist Ghost-CLI meist **kein Muss** und oft nicht der kürzeste Weg.
Der robuste Restore-Pfad bleibt: Backup validieren → DB importieren → `content/` kopieren → Container starten/prüfen.
