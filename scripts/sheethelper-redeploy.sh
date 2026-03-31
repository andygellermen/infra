#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/apps/sheet-helper"
INVENTORY="$ROOT_DIR/ansible/inventory/hosts.ini"
PLAYBOOK="$ROOT_DIR/ansible/playbooks/deploy-sheet-helper.yml"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
IMAGE_NAME="sheet-helper:latest"

usage() {
  cat <<'USAGE'
Usage: ./scripts/sheethelper-redeploy.sh [domain] [--check-only] [--build-only]

Description:
  Baut das lokale Sheet-Helper-Image und deployed den gemeinsamen Container per Ansible.
  Optional kann eine Domain angegeben werden, um vorab passende Hostvars zu validieren.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

TARGET_DOMAIN=""
CHECK_ONLY=0
BUILD_ONLY=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=1; shift ;;
    --build-only) BUILD_ONLY=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *)
      [[ -z "$TARGET_DOMAIN" ]] || die "Nur eine optionale Domain ist erlaubt."
      TARGET_DOMAIN="$1"
      shift
      ;;
  esac
done

require_cmd docker
require_cmd ansible-playbook

[[ -d "$APP_DIR" ]] || die "App-Verzeichnis fehlt: $APP_DIR"
[[ -f "$PLAYBOOK" ]] || die "Playbook fehlt: $PLAYBOOK"

if [[ -n "$TARGET_DOMAIN" ]]; then
  HOSTVARS_FILE="$HOSTVARS_DIR/${TARGET_DOMAIN}.yml"
  [[ -f "$HOSTVARS_FILE" ]] || die "Hostvars fehlen: $HOSTVARS_FILE"
  grep -q '^sheet_helper_enabled:[[:space:]]*true' "$HOSTVARS_FILE" || die "sheet_helper_enabled ist nicht aktiv in $HOSTVARS_FILE"
  grep -q '^sheet_helper_sheet_id:[[:space:]]*".\+"' "$HOSTVARS_FILE" || die "sheet_helper_sheet_id fehlt in $HOSTVARS_FILE"
  grep -q '^sheet_helper_published_url:[[:space:]]*".\+"' "$HOSTVARS_FILE" || die "sheet_helper_published_url fehlt in $HOSTVARS_FILE"
fi

info "Baue Docker-Image: $IMAGE_NAME"
docker build -t "$IMAGE_NAME" "$APP_DIR"
ok "Docker-Image gebaut: $IMAGE_NAME"

if [[ "$BUILD_ONLY" -eq 1 ]]; then
  ok "Nur Build ausgefuehrt."
  exit 0
fi

cmd=(ansible-playbook -i "$INVENTORY" "$PLAYBOOK")
[[ "$CHECK_ONLY" -eq 1 ]] && cmd+=(--check)

info "Starte Sheet-Helper-Deploy"
"${cmd[@]}"
ok "Sheet-Helper-Deploy abgeschlossen"
