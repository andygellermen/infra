#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="$ROOT_DIR/backups/static"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/static-backup.sh --create <domain> [--output <file.tar.gz>]
USAGE
}

[[ ${1:-} == "--help" || ${1:-} == "-h" ]] && { usage; exit 0; }
[[ "${1:-}" == "--create" ]] || die "Usage: $0 --create <domain> [--output <file.tar.gz>]"
[[ $# -ge 2 ]] || die "Domain fehlt"
DOMAIN="$2"; shift 2

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
OUTPUT_FILE="$BACKUP_DIR/$DOMAIN/static-backup-${DOMAIN}-${TIMESTAMP}.tar.gz"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output) OUTPUT_FILE="$2"; shift 2 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

require_cmd tar

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
SITE_DIR="/srv/static/${DOMAIN}"

[[ -d "$SITE_DIR" ]] || die "Statisches Site-Verzeichnis fehlt: $SITE_DIR"
[[ -f "$HOSTVARS_FILE" ]] || warn "Hostvars fehlen: $HOSTVARS_FILE (Backup enthält dann nur die Inhalte)"

WORKDIR="$(mktemp -d /tmp/static-backup-${DOMAIN}.XXXXXX)"
trap 'rm -rf "$WORKDIR"' EXIT
EXPORT_ROOT="$WORKDIR/export-root"
mkdir -p "$EXPORT_ROOT"

info "Sichere statischen Document-Root: $SITE_DIR"
cp -a "$SITE_DIR/." "$EXPORT_ROOT/"

mkdir -p "$EXPORT_ROOT/_infra"
if [[ -f "$HOSTVARS_FILE" ]]; then
  cp -a "$HOSTVARS_FILE" "$EXPORT_ROOT/_infra/hostvars.yml"
fi

{
  echo "domain=$DOMAIN"
  echo "site_dir=$SITE_DIR"
  echo "type=static"
  echo "timestamp=$TIMESTAMP"
} > "$EXPORT_ROOT/_infra/manifest.env"

mkdir -p "$(dirname "$OUTPUT_FILE")"
tar czf "$OUTPUT_FILE" -C "$EXPORT_ROOT" .
ok "Static-Backup erstellt: $OUTPUT_FILE"
