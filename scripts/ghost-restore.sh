#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<USAGE
Usage:
  $0 --list
  $0 <domain> <backup.zip> [--dry-run] [--yes] [--allow-major-mismatch]

Beispiele:
  $0 --list
  $0 blog.example.com /backups/ghost-backup-2025-12-31-00-44-11.zip --dry-run
  $0 blog.example.com /backups/ghost-backup-2025-12-31-00-44-11.zip --yes

Optionen:
  --list                   Listet vorhandene Ghost-Container (docker ps -a).
  --dry-run                Führt nur Validierung durch, ohne Restore.
  --yes                    Kein interaktiver Bestätigungs-Dialog.
  --allow-major-mismatch   Erlaubt Restore trotz Versions-Major-Mismatch.
  --help, -h               Hilfe anzeigen.
USAGE
}

die() {
  echo "❌ Fehler: $*" >&2
  exit 1
}

info() {
  echo "ℹ️  $*"
}

ok() {
  echo "✅ $*"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "Benötigtes Tool fehlt: $1"
}

trim() {
  local v="$1"
  v="${v#\"}"
  v="${v%\"}"
  printf '%s' "$v" | sed 's/[[:space:]]*$//'
}

extract_hostvar() {
  local key="$1"
  local file="$2"
  local value
  value="$(awk -F': ' -v k="$key" '$1==k {print $2; exit}' "$file")"
  trim "$value"
}

list_ghost_containers() {
  echo "📦 Verfügbare Ghost-Container:"
  docker ps -a --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}' | {
    read -r header || true
    echo "$header"
    grep '^ghost-' || true
  }
}

DOMAIN=""
BACKUP_ZIP=""
DRY_RUN=0
ASSUME_YES=0
ALLOW_MAJOR_MISMATCH=0

if [[ $# -eq 0 ]]; then
  usage
  exit 1
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --list)
      require_cmd docker
      list_ghost_containers
      exit 0
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --yes)
      ASSUME_YES=1
      shift
      ;;
    --allow-major-mismatch)
      ALLOW_MAJOR_MISMATCH=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    --*)
      die "Unbekannte Option: $1"
      ;;
    *)
      if [[ -z "$DOMAIN" ]]; then
        DOMAIN="$1"
      elif [[ -z "$BACKUP_ZIP" ]]; then
        BACKUP_ZIP="$1"
      else
        die "Zu viele Positionsargumente. Erwartet: <domain> <backup.zip>"
      fi
      shift
      ;;
  esac
done

[[ -n "$DOMAIN" ]] || die "Domain fehlt."
[[ -n "$BACKUP_ZIP" ]] || die "Backup-ZIP fehlt."

require_cmd docker
require_cmd unzip
require_cmd awk
require_cmd rg
require_cmd sed

HOSTVARS_FILE="./ansible/hostvars/${DOMAIN}.yml"
[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars nicht gefunden: $HOSTVARS_FILE"
[[ -f "$BACKUP_ZIP" ]] || die "Backup nicht gefunden: $BACKUP_ZIP"

CONTAINER_NAME="ghost-${DOMAIN//./-}"
VOLUME_NAME="ghost_${DOMAIN//./_}_content"
MYSQL_CONTAINER="ghost-mysql"

if ! docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER_NAME"; then
  die "Zielcontainer existiert nicht: $CONTAINER_NAME (nutze '$0 --list')"
fi

if ! docker ps -a --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER"; then
  die "MySQL-Container fehlt: $MYSQL_CONTAINER"
fi

if ! docker volume ls --format '{{.Name}}' | grep -qx "$VOLUME_NAME"; then
  die "Ghost-Volume fehlt: $VOLUME_NAME"
fi

DB_NAME="$(extract_hostvar ghost_domain_db "$HOSTVARS_FILE")"
DB_USER="$(extract_hostvar ghost_domain_usr "$HOSTVARS_FILE")"
DB_PASS="$(extract_hostvar ghost_domain_pwd "$HOSTVARS_FILE")"
TARGET_GHOST_VERSION="$(extract_hostvar ghost_version "$HOSTVARS_FILE")"

[[ -n "$DB_NAME" ]] || die "ghost_domain_db fehlt in $HOSTVARS_FILE"
[[ -n "$DB_USER" ]] || die "ghost_domain_usr fehlt in $HOSTVARS_FILE"
[[ -n "$DB_PASS" ]] || die "ghost_domain_pwd fehlt in $HOSTVARS_FILE"

info "Prüfe ZIP-Integrität"
unzip -t "$BACKUP_ZIP" >/dev/null
ok "ZIP ist konsistent"

WORKDIR="$(mktemp -d /tmp/ghost-restore-${DOMAIN}.XXXXXX)"
trap 'rm -rf "$WORKDIR"' EXIT

info "Entpacke Backup nach $WORKDIR"
unzip -o "$BACKUP_ZIP" -d "$WORKDIR" >/dev/null

SQL_FILE="$(find "$WORKDIR" -type f -name '*.sql' | head -n1 || true)"
CONTENT_DIR="$(find "$WORKDIR" -type d -name content | head -n1 || true)"
VERSION_JSON="$(find "$WORKDIR" -type f -path '*/data/content-from-v*-on-*.json' | head -n1 || true)"

[[ -n "$SQL_FILE" ]] || die "Keine SQL-Datei im Backup gefunden"
[[ -s "$SQL_FILE" ]] || die "SQL-Datei ist leer: $SQL_FILE"
[[ -n "$CONTENT_DIR" ]] || die "Kein content/ Ordner im Backup gefunden"

SOURCE_GHOST_VERSION=""
SOURCE_GHOST_MAJOR=""
if [[ -n "$VERSION_JSON" ]]; then
  SOURCE_GHOST_VERSION="$(rg -o '"version"\s*:\s*"[0-9]+\.[0-9]+\.[0-9]+"' "$VERSION_JSON" | head -n1 | sed -E 's/.*"([0-9]+\.[0-9]+\.[0-9]+)"/\1/' || true)"
  if [[ -n "$SOURCE_GHOST_VERSION" ]]; then
    SOURCE_GHOST_MAJOR="${SOURCE_GHOST_VERSION%%.*}"
  fi
fi

TARGET_GHOST_MAJOR=""
if [[ "$TARGET_GHOST_VERSION" =~ ^[0-9]+$ ]]; then
  TARGET_GHOST_MAJOR="$TARGET_GHOST_VERSION"
else
  IMAGE_TAG="$(docker inspect --format '{{.Config.Image}}' "$CONTAINER_NAME" | awk -F: '{print $2}' || true)"
  if [[ "$IMAGE_TAG" =~ ^[0-9]+(\.[0-9]+)?(\.[0-9]+)?$ ]]; then
    TARGET_GHOST_MAJOR="${IMAGE_TAG%%.*}"
  fi
fi

info "Restore-Ziel: $DOMAIN"
info "Container: $CONTAINER_NAME | Volume: $VOLUME_NAME | MySQL: $MYSQL_CONTAINER"
info "SQL: $SQL_FILE"
info "Content: $CONTENT_DIR"
[[ -n "$SOURCE_GHOST_VERSION" ]] && info "Quelle Ghost-Version (aus Backup): $SOURCE_GHOST_VERSION"
[[ -n "$TARGET_GHOST_VERSION" ]] && info "Ziel Ghost-Version (hostvars): $TARGET_GHOST_VERSION"

if [[ -n "$SOURCE_GHOST_MAJOR" && -n "$TARGET_GHOST_MAJOR" && "$SOURCE_GHOST_MAJOR" != "$TARGET_GHOST_MAJOR" ]]; then
  if [[ "$ALLOW_MAJOR_MISMATCH" -eq 0 ]]; then
    die "Versions-Major-Mismatch: Quelle v${SOURCE_GHOST_MAJOR} vs Ziel v${TARGET_GHOST_MAJOR}. Nutze --allow-major-mismatch falls bewusst."
  fi
  info "⚠️  Major-Mismatch wurde durch --allow-major-mismatch freigegeben"
fi

info "Prüfe DB-Login"
docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" -e 'SELECT 1' "$DB_NAME" >/dev/null
ok "DB-Login erfolgreich"

if [[ "$DRY_RUN" -eq 1 ]]; then
  ok "Dry-Run abgeschlossen. Keine Änderungen durchgeführt."
  exit 0
fi

if [[ "$ASSUME_YES" -ne 1 ]]; then
  echo "⚠️  Es werden DB und Content der Zielinstanz überschrieben: $DOMAIN"
  read -r -p "Fortfahren? (yes/no): " answer
  [[ "$answer" == "yes" ]] || die "Abgebrochen"
fi

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
SAFETY_DIR="/tmp/ghost-restore-safety/${DOMAIN}/${TIMESTAMP}"
mkdir -p "$SAFETY_DIR"

info "Erzeuge Safety-Backups unter $SAFETY_DIR"
docker exec "$MYSQL_CONTAINER" mysqldump -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" > "${SAFETY_DIR}/pre-restore.sql"
docker run --rm -v "${VOLUME_NAME}:/data" -v "${SAFETY_DIR}:/backup" alpine \
  sh -c 'tar czf /backup/pre-restore-content.tar.gz -C /data .'
ok "Safety-Backups erstellt"

info "Stoppe Ghost-Container"
docker stop "$CONTAINER_NAME" >/dev/null || true

info "Leere Ziel-Datenbank"
docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -Nse "
SET FOREIGN_KEY_CHECKS=0;
SELECT CONCAT('DROP TABLE IF EXISTS \\\`', table_name, '\\\`;')
FROM information_schema.tables
WHERE table_schema='${DB_NAME}';
" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

info "Importiere SQL"
cat "$SQL_FILE" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

info "Leere Ghost-Content-Volume"
docker run --rm -v "${VOLUME_NAME}:/target" alpine sh -c 'find /target -mindepth 1 -delete'

info "Kopiere content/ in Volume"
docker run --rm \
  -v "${VOLUME_NAME}:/target" \
  -v "${CONTENT_DIR}:/source:ro" \
  alpine sh -c 'cp -a /source/. /target/'

info "Starte Ghost-Container"
docker start "$CONTAINER_NAME" >/dev/null
sleep 2

ok "Restore abgeschlossen"
echo "📄 Safety-Backups: $SAFETY_DIR"
echo "🔎 Logs prüfen: docker logs --tail=150 $CONTAINER_NAME"
