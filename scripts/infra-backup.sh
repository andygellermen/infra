#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="$ROOT_DIR/backups/infra"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
OUTPUT_FILE="${BACKUP_DIR}/infra-backup-${TIMESTAMP}.tar.gz"
INCLUDE_MYSQL_DUMP=1

usage() {
  cat <<USAGE
Usage:
  $0 --create [--output <file.tar.gz>] [--no-mysql-dump]

Examples:
  $0 --create
  $0 --create --output /tmp/infra-full.tar.gz
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

backup_volume() {
  local volume="$1"
  local out="$2"
  docker run --rm -v "${volume}:/src:ro" -v "${out}:/backup" alpine \
    sh -c "tar czf /backup/${volume}.tar.gz -C /src ."
}

if [[ ${1:-} == "--help" || ${1:-} == "-h" ]]; then
  usage
  exit 0
fi

[[ $# -ge 1 ]] || { usage; exit 1; }
ACTION="$1"; shift

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output)
      OUTPUT_FILE="$2"; shift 2 ;;
    --no-mysql-dump)
      INCLUDE_MYSQL_DUMP=0; shift ;;
    --help|-h)
      usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ "$ACTION" == "--create" ]] || { usage; exit 1; }

require_cmd docker
require_cmd tar
mkdir -p "$BACKUP_DIR"

WORKDIR="$(mktemp -d /tmp/infra-backup.XXXXXX)"
trap 'rm -rf "$WORKDIR"' EXIT
mkdir -p "$WORKDIR"/{volumes,files,mysql,meta}

info "Sammle Metadaten"
docker ps -a --format '{{.Names}} {{.Image}} {{.Status}}' > "$WORKDIR/meta/containers.txt" || true
docker volume ls --format '{{.Name}}' > "$WORKDIR/meta/volumes.txt" || true
printf 'timestamp=%s\n' "$TIMESTAMP" > "$WORKDIR/meta/manifest.env"
printf 'root_dir=%s\n' "$ROOT_DIR" >> "$WORKDIR/meta/manifest.env"

info "Sichere relevante Dateien"
for path in ansible/hostvars ansible/secrets data/traefik data/crowdsec; do
  if [[ -e "$ROOT_DIR/$path" ]]; then
    mkdir -p "$WORKDIR/files/$(dirname "$path")"
    cp -a "$ROOT_DIR/$path" "$WORKDIR/files/$path"
  fi
done

info "Ermittle relevante Docker-Volumes"
mapfile -t all_volumes < <(docker volume ls --format '{{.Name}}')
volumes_to_backup=()
for v in "${all_volumes[@]}"; do
  case "$v" in
    mysql_data|portainer_data|ghost_*_content|traefik*|crowdsec*) volumes_to_backup+=("$v") ;;
  esac
done

for v in "${volumes_to_backup[@]}"; do
  info "Backup Volume: $v"
  backup_volume "$v" "$WORKDIR/volumes"
done

if [[ "$INCLUDE_MYSQL_DUMP" -eq 1 ]] && docker ps --format '{{.Names}}' | grep -qx 'ghost-mysql'; then
  info "Exportiere MySQL Dump (all databases)"
  docker exec ghost-mysql sh -c 'exec mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" --all-databases --single-transaction --quick --lock-tables=false' > "$WORKDIR/mysql/all-databases.sql" || {
    info "⚠️ MySQL-Dump fehlgeschlagen, fahre ohne Dump fort"
    rm -f "$WORKDIR/mysql/all-databases.sql"
  }
fi

mkdir -p "$(dirname "$OUTPUT_FILE")"
info "Erzeuge Archiv: $OUTPUT_FILE"
tar czf "$OUTPUT_FILE" -C "$WORKDIR" .
ok "Infra-Backup erstellt: $OUTPUT_FILE"
