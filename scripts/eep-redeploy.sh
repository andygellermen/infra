#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/apps/easy-event-planner"
INVENTORY="$ROOT_DIR/ansible/inventory/hosts.ini"
PLAYBOOK="$ROOT_DIR/ansible/playbooks/deploy-easy-event-planner.yml"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
IMAGE_NAME="easy-event-planner:latest"

usage() {
  cat <<'USAGE'
Usage: ./scripts/eep-redeploy.sh [domain|--all] [--check-only] [--build-only]

Description:
  Baut das lokale Easy-Event-Planner-Image und deployed per Ansible.
  Ohne Domain wird standardmaessig --all verwendet (alle eep_enabled-Instanzen).
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

TARGET_DOMAIN=""
CHECK_ONLY=0
BUILD_ONLY=0
DEPLOY_ALL=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    --all)
      [[ -z "$TARGET_DOMAIN" ]] || die "--all kann nicht mit einer Domain kombiniert werden."
      DEPLOY_ALL=1
      shift
      ;;
    --check-only)
      CHECK_ONLY=1
      shift
      ;;
    --build-only)
      BUILD_ONLY=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      [[ -z "$TARGET_DOMAIN" ]] || die "Nur eine optionale Domain ist erlaubt."
      TARGET_DOMAIN="$1"
      DEPLOY_ALL=0
      shift
      ;;
  esac
done

require_cmd docker
require_cmd ansible-playbook
require_cmd find
require_cmd grep

[[ -d "$APP_DIR" ]] || die "App-Verzeichnis fehlt: $APP_DIR"
[[ -f "$PLAYBOOK" ]] || die "Playbook fehlt: $PLAYBOOK"
[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"

if [[ "$DEPLOY_ALL" -eq 1 ]]; then
  EEP_HOSTVARS=()
  while IFS= read -r file; do
    grep -q '^eep_enabled:[[:space:]]*true' "$file" && EEP_HOSTVARS+=("$file")
  done < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' | sort)
  [[ ${#EEP_HOSTVARS[@]} -gt 0 ]] || die "Keine eep_enabled-Hostvars gefunden."
  info "Gefundene eep-Instanzen: ${#EEP_HOSTVARS[@]}"
else
  HOSTVARS_FILE="$HOSTVARS_DIR/${TARGET_DOMAIN}.yml"
  [[ -f "$HOSTVARS_FILE" ]] || die "Hostvars fehlen: $HOSTVARS_FILE"
  grep -q '^eep_enabled:[[:space:]]*true' "$HOSTVARS_FILE" || die "eep_enabled ist nicht aktiv in $HOSTVARS_FILE"
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
if [[ "$DEPLOY_ALL" -eq 0 ]]; then
  cmd+=(-e "target_domain=${TARGET_DOMAIN}")
fi

if [[ "$DEPLOY_ALL" -eq 1 ]]; then
  info "Starte Easy-Event-Planner-Deploy fuer alle Instanzen"
else
  info "Starte Easy-Event-Planner-Deploy fuer ${TARGET_DOMAIN}"
fi
"${cmd[@]}"
ok "Easy-Event-Planner-Deploy abgeschlossen"
