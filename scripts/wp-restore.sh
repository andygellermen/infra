#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }
extract_hostvar(){ awk -F': ' -v k="$1" '$1==k {gsub(/"/,"",$2); gsub(/[[:space:]]+$/, "", $2); print $2; exit}' "$2"; }

usage() {
  cat <<USAGE
Usage:
  $0 --restore <domain> <backup.tar.gz> [--yes] [--restore-hostvars] [--allow-version-downgrade]
USAGE
}

version_to_int() {
  local v="$1" a b c
  [[ "$v" == "latest" || -z "$v" ]] && { echo 999999999; return; }
  IFS='.' read -r a b c <<< "$v"
  a="${a:-0}"; b="${b:-0}"; c="${c:-0}"
  printf '%03d%03d%03d\n' "$a" "$b" "$c"
}

[[ ${1:-} == "--help" || ${1:-} == "-h" ]] && { usage; exit 0; }
[[ "${1:-}" == "--restore" ]] || die "Usage: $0 --restore <domain> <backup.tar.gz> [--yes]"
[[ $# -ge 3 ]] || die "Parameter fehlen"

DOMAIN="$2"; BACKUP_FILE="$3"; shift 3
ASSUME_YES=0
RESTORE_HOSTVARS=0
ALLOW_DOWNGRADE=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes) ASSUME_YES=1; shift ;;
    --restore-hostvars) RESTORE_HOSTVARS=1; shift ;;
    --allow-version-downgrade) ALLOW_DOWNGRADE=1; shift ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

require_cmd docker
[[ -f "$BACKUP_FILE" ]] || die "Backup fehlt: $BACKUP_FILE"

HOSTVARS="$ROOT_DIR/ansible/hostvars/${DOMAIN}.yml"
VOLUME="wp_${DOMAIN//./_}_html"
MYSQL_CONTAINER="ghost-mysql"
CONTAINER="wp-${DOMAIN//./-}"

WORKDIR="$(mktemp -d /tmp/wp-restore-${DOMAIN}.XXXXXX)"; trap 'rm -rf "$WORKDIR"' EXIT
tar xzf "$BACKUP_FILE" -C "$WORKDIR"
[[ -f "$WORKDIR/data/db.sql" && -f "$WORKDIR/data/html.tar.gz" ]] || die "Backup unvollständig (db.sql/html.tar.gz fehlt)"

if [[ "$RESTORE_HOSTVARS" -eq 1 ]]; then
  [[ -f "$WORKDIR/files/hostvars.yml" ]] || die "hostvars.yml fehlt im Backup"
  mkdir -p "$(dirname "$HOSTVARS")"
  cp -a "$WORKDIR/files/hostvars.yml" "$HOSTVARS"
  info "Hostvars aus Backup wiederhergestellt"
fi

[[ -f "$HOSTVARS" ]] || die "Hostvars fehlt: $HOSTVARS"
DB_NAME="$(extract_hostvar wp_domain_db "$HOSTVARS")"
DB_USER="$(extract_hostvar wp_domain_usr "$HOSTVARS")"
DB_PASS="$(extract_hostvar wp_domain_pwd "$HOSTVARS")"
TARGET_WP_VERSION="$(extract_hostvar wp_version "$HOSTVARS")"
SOURCE_WP_VERSION="$(extract_hostvar wp_version "$WORKDIR/files/hostvars.yml" 2>/dev/null || true)"
[[ -n "$TARGET_WP_VERSION" ]] || TARGET_WP_VERSION="latest"
[[ -n "$SOURCE_WP_VERSION" ]] || SOURCE_WP_VERSION="$TARGET_WP_VERSION"

src_int="$(version_to_int "$SOURCE_WP_VERSION")"
tgt_int="$(version_to_int "$TARGET_WP_VERSION")"
if [[ "$ALLOW_DOWNGRADE" -ne 1 && "$tgt_int" -lt "$src_int" ]]; then
  die "Version-Downgrade erkannt (Backup: $SOURCE_WP_VERSION -> Ziel: $TARGET_WP_VERSION). Entweder wp_version anheben oder --allow-version-downgrade setzen."
fi

if [[ "$ASSUME_YES" -ne 1 ]]; then
  echo "⚠️  Restore überschreibt WordPress DB + Files für $DOMAIN"
  echo "ℹ️  Backup-Version: $SOURCE_WP_VERSION | Ziel-Version: $TARGET_WP_VERSION"
  read -r -p "Fortfahren? (yes/no): " a
  [[ "$a" == yes ]] || die "Abbruch"
fi

docker ps --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER" || die "MySQL Container läuft nicht: $MYSQL_CONTAINER"

info "Stoppe Container: $CONTAINER"
docker stop "$CONTAINER" >/dev/null || true

info "Leere DB und importiere Dump"
{
  echo "SET FOREIGN_KEY_CHECKS=0;"
  docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -Nse "SELECT CONCAT('DROP TABLE IF EXISTS \\`', table_name, '\\`;') FROM information_schema.tables WHERE table_schema='${DB_NAME}';"
  echo "SET FOREIGN_KEY_CHECKS=1;"
} | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"
cat "$WORKDIR/data/db.sql" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

info "Restore Document Root"
docker volume create "$VOLUME" >/dev/null
docker run --rm -v "${VOLUME}:/target" -v "${WORKDIR}/data:/backup:ro" alpine sh -c 'find /target -mindepth 1 -delete; tar xzf /backup/html.tar.gz -C /target'

info "Starte Container: $CONTAINER"
docker start "$CONTAINER" >/dev/null || true

if [[ "$TARGET_WP_VERSION" != "$SOURCE_WP_VERSION" ]]; then
  info "Versionen unterscheiden sich (Backup=$SOURCE_WP_VERSION, Ziel=$TARGET_WP_VERSION)."
  info "Empfehlung: anschließend ./scripts/wp-redeploy.sh $DOMAIN ausführen, damit der Container garantiert mit target wp_version läuft."
fi

ok "WordPress-Restore abgeschlossen"
