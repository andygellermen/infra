#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<USAGE
Usage:
  $0 --restore <infra-backup.tar.gz> [--yes] [--run-setup]
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

restore_volume() {
  local archive="$1"
  local volume="$2"
  docker volume create "$volume" >/dev/null
  docker run --rm -v "${volume}:/target" -v "$(dirname "$archive"):/backup:ro" alpine \
    sh -c "find /target -mindepth 1 -delete; tar xzf /backup/$(basename "$archive") -C /target"
}

ASSUME_YES=0
RUN_SETUP=0
BACKUP_FILE=""
if [[ ${1:-} == "--help" || ${1:-} == "-h" ]]; then
  usage
  exit 0
fi
[[ $# -ge 2 ]] || { usage; exit 1; }

if [[ "$1" != "--restore" ]]; then usage; exit 1; fi
BACKUP_FILE="$2"; shift 2

while [[ $# -gt 0 ]]; do
  case "$1" in
    --yes) ASSUME_YES=1; shift ;;
    --run-setup) RUN_SETUP=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ -f "$BACKUP_FILE" ]] || die "Backup nicht gefunden: $BACKUP_FILE"
require_cmd docker
require_cmd tar

if [[ "$RUN_SETUP" -eq 1 ]]; then
  info "Führe infra-setup.sh vorab aus"
  "$ROOT_DIR/scripts/infra-setup.sh"
fi

WORKDIR="$(mktemp -d /tmp/infra-restore.XXXXXX)"
trap 'rm -rf "$WORKDIR"' EXIT

tar xzf "$BACKUP_FILE" -C "$WORKDIR"

if [[ "$ASSUME_YES" -ne 1 ]]; then
  echo "⚠️  Restore überschreibt Hostvars/Secrets/Traefik/CrowdSec-Dateien und Docker-Volumes."
  read -r -p "Fortfahren? (yes/no): " a
  [[ "$a" == "yes" ]] || die "Abgebrochen"
fi

info "Stelle dateibasierte Konfigurationen wieder her"
if [[ -d "$WORKDIR/files" ]]; then
  cp -a "$WORKDIR/files/." "$ROOT_DIR/"
fi

if [[ -d "$WORKDIR/volumes" ]]; then
  while IFS= read -r vol_archive; do
    vol_name="$(basename "$vol_archive" .tar.gz)"
    info "Restore Volume: $vol_name"
    restore_volume "$vol_archive" "$vol_name"
  done < <(find "$WORKDIR/volumes" -type f -name '*.tar.gz' | sort)
fi

if [[ -f "$WORKDIR/mysql/all-databases.sql" ]] && docker ps --format '{{.Names}}' | grep -qx 'ghost-mysql'; then
  info "Importiere MySQL all-databases Dump"
  docker exec -i ghost-mysql sh -c 'exec mysql -uroot -p"$MYSQL_ROOT_PASSWORD"' < "$WORKDIR/mysql/all-databases.sql"
fi

ok "Infra-Restore abgeschlossen"
echo "🔁 Empfehlung: Deploy-Reihenfolge anstoßen: deploy-mysql -> deploy-traefik -> deploy-crowdsec -> Ghost/Portainer"
