#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${EEP_BACKUP_DIR:-$ROOT_DIR/backups/eep}"
HOSTVARS_DIR="${EEP_HOSTVARS_DIR:-$ROOT_DIR/ansible/hostvars}"
SITE_ROOT="${EEP_SITE_ROOT:-/srv/easy-event-planner}"

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

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

require_cmd docker
require_cmd tar

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
SITE_DIR="$SITE_ROOT/${DOMAIN}"

[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars fehlt: $HOSTVARS_FILE"
grep -Eq '^eep_enabled:[[:space:]]*true([[:space:]]|$)' "$HOSTVARS_FILE" || die "Hostvars gehoeren nicht zu einer aktiven EEP-Instanz: $HOSTVARS_FILE"
[[ -d "$SITE_DIR" ]] || die "EEP-Site-Verzeichnis fehlt: $SITE_DIR"

HOST_UID="$(id -u)"
HOST_GID="$(id -g)"

WORKDIR="$(mktemp -d /tmp/eep-backup-${DOMAIN}.XXXXXX)"
trap 'rm -rf "$WORKDIR"' EXIT
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
ok "EEP-Backup erstellt: $OUTPUT_FILE"
