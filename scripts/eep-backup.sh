#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${EEP_BACKUP_DIR:-$ROOT_DIR/backups/eep}"
HOSTVARS_DIR="${EEP_HOSTVARS_DIR:-$ROOT_DIR/ansible/hostvars}"
SITE_ROOT="${EEP_SITE_ROOT:-/srv/easy-event-planner}"

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

extract_hostvar() {
  local key="$1"
  local file="$2"
  awk -F': ' -v k="$key" '$1==k {value=$2; gsub(/^"|"$/, "", value); print value; exit}' "$file"
}

container_running() {
  local name="$1"
  docker ps --format '{{.Names}}' | grep -qx "$name"
}

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/eep-backup.sh --create <domain> [--output <file.tar.gz>]
USAGE
}

normalize_domain() {
  local domain="$1"
  if [[ "$domain" =~ ^[a-zA-Z0-9.-]+$ ]]; then
    printf '%s\n' "$domain"
  else
    command -v idn >/dev/null 2>&1 || die "Domain enthaelt Nicht-ASCII-Zeichen, aber Tool fehlt: idn"
    idn --quiet --uts46 "$domain"
  fi
}

[[ ${1:-} == "--help" || ${1:-} == "-h" ]] && { usage; exit 0; }
[[ "${1:-}" == "--create" ]] || die "Usage: $0 --create <domain> [--output <file.tar.gz>]"
[[ $# -ge 2 ]] || die "Domain fehlt"

DOMAIN="$(normalize_domain "$2")"
shift 2

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
OUTPUT_FILE="$BACKUP_DIR/$DOMAIN/eep-backup-${DOMAIN}-${TIMESTAMP}.tar.gz"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output) OUTPUT_FILE="$2"; shift 2 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ ! -e "$OUTPUT_FILE" ]] || die "Backup-Zieldatei existiert bereits: $OUTPUT_FILE"

require_cmd docker
require_cmd tar
require_cmd awk
require_cmd grep

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
SITE_DIR="$SITE_ROOT/${DOMAIN}"

[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars fehlt: $HOSTVARS_FILE"
grep -Eq '^eep_enabled:[[:space:]]*true([[:space:]]|$)' "$HOSTVARS_FILE" || die "Hostvars gehoeren nicht zu einer aktiven EEP-Instanz: $HOSTVARS_FILE"
[[ -d "$SITE_DIR" ]] || die "EEP-Site-Verzeichnis fehlt: $SITE_DIR"

HOST_UID="$(id -u)"
HOST_GID="$(id -g)"

WORKDIR="$(mktemp -d /tmp/eep-backup-${DOMAIN}.XXXXXX)"
APP_CONTAINER="$(extract_hostvar eep_container_name "$HOSTVARS_FILE")"
WORKER_CONTAINER="$(extract_hostvar eep_worker_container_name "$HOSTVARS_FILE")"
[[ -n "$APP_CONTAINER" ]] || APP_CONTAINER="easy-event-planner-${DOMAIN//./-}"
[[ -n "$WORKER_CONTAINER" ]] || WORKER_CONTAINER="easy-event-planner-worker-${DOMAIN//./-}"
APP_WAS_RUNNING=0
WORKER_WAS_RUNNING=0

cleanup() {
  local rc="$?"
  local restart_failed=0
  trap - EXIT
  if [[ "$APP_WAS_RUNNING" -eq 1 ]]; then
    info "Starte EEP-App nach Backup: $APP_CONTAINER"
    docker start "$APP_CONTAINER" >/dev/null || {
      warn "EEP-App konnte nicht neu gestartet werden: $APP_CONTAINER"
      restart_failed=1
    }
  fi
  if [[ "$WORKER_WAS_RUNNING" -eq 1 ]]; then
    info "Starte EEP-Worker nach Backup: $WORKER_CONTAINER"
    docker start "$WORKER_CONTAINER" >/dev/null || {
      warn "EEP-Worker konnte nicht neu gestartet werden: $WORKER_CONTAINER"
      restart_failed=1
    }
  fi
  rm -rf "$WORKDIR"
  if [[ "$rc" -eq 0 && "$restart_failed" -ne 0 ]]; then
    rc=1
  fi
  exit "$rc"
}
trap cleanup EXIT

if container_running "$WORKER_CONTAINER"; then
  WORKER_WAS_RUNNING=1
  info "Stoppe EEP-Worker fuer konsistente SQLite-Sicherung: $WORKER_CONTAINER"
  docker stop --time 30 "$WORKER_CONTAINER" >/dev/null
fi
if container_running "$APP_CONTAINER"; then
  APP_WAS_RUNNING=1
  info "Stoppe EEP-App fuer konsistente SQLite-Sicherung: $APP_CONTAINER"
  docker stop --time 30 "$APP_CONTAINER" >/dev/null
fi

EXPORT_ROOT="$WORKDIR/export-root"
mkdir -p "$EXPORT_ROOT/_infra"

info "Sichere EEP-Site-Verzeichnis: $SITE_DIR"
docker run --rm \
  -e HOST_UID="$HOST_UID" \
  -e HOST_GID="$HOST_GID" \
  -v "${SITE_DIR}:/src:ro" \
  -v "${EXPORT_ROOT}:/backup" \
  alpine \
  sh -c 'cp -a /src/. /backup/ && chown -R "$HOST_UID:$HOST_GID" /backup'

cp -a "$HOSTVARS_FILE" "$EXPORT_ROOT/_infra/hostvars.yml"
{
  echo "domain=$DOMAIN"
  echo "site_dir=$SITE_DIR"
  echo "type=easy-event-planner"
  echo "timestamp=$TIMESTAMP"
} > "$EXPORT_ROOT/_infra/manifest.env"

mkdir -p "$(dirname "$OUTPUT_FILE")"
tar czf "$OUTPUT_FILE" -C "$EXPORT_ROOT" .
tar tzf "$OUTPUT_FILE" >/dev/null
ok "EEP-Backup erstellt: $OUTPUT_FILE"
