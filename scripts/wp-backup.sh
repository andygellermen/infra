#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="$ROOT_DIR/backups/wordpress"

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

extract_hostvar() {
  awk -F': ' -v k="$1" '$1==k {gsub(/"/,"",$2); gsub(/[[:space:]]+$/, "", $2); print $2; exit}' "$2"
}

usage() {
  cat <<USAGE
Usage:
  $0 --create <domain> [--output <file.tar.gz>]
USAGE
}

[[ ${1:-} == "--help" || ${1:-} == "-h" ]] && { usage; exit 0; }
[[ "${1:-}" == "--create" ]] || die "Usage: $0 --create <domain> [--output <file.tar.gz>]"
[[ $# -ge 2 ]] || die "Domain fehlt"
DOMAIN="$2"; shift 2

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
OUTPUT_FILE="$BACKUP_DIR/$DOMAIN/wp-backup-${DOMAIN}-${TIMESTAMP}.tar.gz"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output) OUTPUT_FILE="$2"; shift 2 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

require_cmd docker
HOSTVARS="$ROOT_DIR/ansible/hostvars/${DOMAIN}.yml"
[[ -f "$HOSTVARS" ]] || die "Hostvars fehlt: $HOSTVARS"

VOLUME="wp_${DOMAIN//./_}_html"
CONTAINER="wp-${DOMAIN//./-}"
MYSQL_CONTAINER="ghost-mysql"

docker volume ls --format '{{.Name}}' | grep -qx "$VOLUME" || die "Volume fehlt: $VOLUME"
docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER" || die "Container fehlt: $CONTAINER"
docker ps --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER" || die "MySQL Container läuft nicht: $MYSQL_CONTAINER"

DB_NAME="$(extract_hostvar wp_domain_db "$HOSTVARS")"
DB_USER="$(extract_hostvar wp_domain_usr "$HOSTVARS")"
DB_PASS="$(extract_hostvar wp_domain_pwd "$HOSTVARS")"
WP_VERSION="$(extract_hostvar wp_version "$HOSTVARS")"
[[ -n "$DB_NAME" && -n "$DB_USER" && -n "$DB_PASS" ]] || die "DB-Zugangsdaten unvollständig in Hostvars"
[[ -n "$WP_VERSION" ]] || WP_VERSION="latest"

WORKDIR="$(mktemp -d /tmp/wp-backup-${DOMAIN}.XXXXXX)"; trap 'rm -rf "$WORKDIR"' EXIT
EXPORT_ROOT="$WORKDIR/export-root"
mkdir -p "$EXPORT_ROOT"

info "Sichere vollständigen WordPress-Document-Root"
docker run --rm -v "${VOLUME}:/src:ro" -v "${EXPORT_ROOT}:/backup" alpine sh -c 'cp -a /src/. /backup/'

info "Dump DB: $DB_NAME"
docker exec -e MYSQL_PWD="$DB_PASS" "$MYSQL_CONTAINER" mysqldump --no-tablespaces -u"$DB_USER" "$DB_NAME" > "$EXPORT_ROOT/db.sql"
[[ -s "$EXPORT_ROOT/db.sql" ]] || die "DB-Dump ist leer"

# Optionales Infra-Meta für managed Instanzen
mkdir -p "$EXPORT_ROOT/_infra"
cp -a "$HOSTVARS" "$EXPORT_ROOT/_infra/hostvars.yml" || true
{
  echo "domain=$DOMAIN"
  echo "container=$CONTAINER"
  echo "volume=$VOLUME"
  echo "wp_version=$WP_VERSION"
  echo "timestamp=$TIMESTAMP"
} > "$EXPORT_ROOT/_infra/manifest.env"

mkdir -p "$(dirname "$OUTPUT_FILE")"
tar czf "$OUTPUT_FILE" -C "$EXPORT_ROOT" .
ok "WordPress-Backup erstellt: $OUTPUT_FILE"
