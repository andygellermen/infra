#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_DIR="$ROOT_DIR/apps/sprach-a-lyzer"
INVENTORY="$ROOT_DIR/ansible/inventory/hosts.ini"
PLAYBOOK="$ROOT_DIR/ansible/playbooks/deploy-sprach-a-lyzer.yml"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
IMAGE_NAME="sprach-a-lyzer:latest"

usage() {
  cat <<'USAGE'
Usage: ./scripts/sal-redeploy.sh <domain> [--check-only] [--build-only] [--skip-auth-smoke]

Builds the Sprach-A-Lyzer image, deploys PostgreSQL/migrations/seed/API through
Ansible and verifies both the auth barrier and authenticated MVP readiness.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

DOMAIN=""
CHECK_ONLY=0
BUILD_ONLY=0
SKIP_AUTH_SMOKE=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=1; shift ;;
    --build-only) BUILD_ONLY=1; shift ;;
    --skip-auth-smoke) SKIP_AUTH_SMOKE=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *)
      [[ -z "$DOMAIN" ]] || die "Nur eine Domain ist erlaubt."
      DOMAIN="$1"
      shift
      ;;
  esac
done

[[ "$DOMAIN" =~ ^[a-z0-9.-]+$ ]] || die "Ungueltige oder fehlende Domain."
require_cmd ansible-playbook
require_cmd docker
require_cmd awk

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars fehlen: $HOSTVARS_FILE (zuerst sal-add.sh ausfuehren)"
grep -Eq '^sal_enabled:[[:space:]]*true([[:space:]]|$)' "$HOSTVARS_FILE" || die "sal_enabled ist nicht aktiv."
AUTH_USERNAME="$(awk -F ': *' '/^sal_basic_auth_username:/ {gsub(/^"|"$/, "", $2); print $2; exit}' "$HOSTVARS_FILE")"
[[ "$AUTH_USERNAME" =~ ^[A-Za-z0-9._-]+$ ]] || die "Basic-Auth-Benutzername fehlt in den Hostvars."

info "Baue Docker-Image: $IMAGE_NAME"
docker build --pull -t "$IMAGE_NAME" "$APP_DIR"
ok "Docker-Image gebaut."
[[ "$BUILD_ONLY" -eq 0 ]] || exit 0

cmd=(ansible-playbook -i "$INVENTORY" "$PLAYBOOK" -e "target_domain=${DOMAIN}")
[[ "$CHECK_ONLY" -eq 1 ]] && cmd+=(--check)
info "Starte geschuetzten Sprach-A-Lyzer-Deploy fuer $DOMAIN"
"${cmd[@]}"
ok "Ansible-Deploy abgeschlossen."

if [[ "$CHECK_ONLY" -eq 0 && "$SKIP_AUTH_SMOKE" -eq 0 ]]; then
  "$ROOT_DIR/scripts/sal-smoke-check.sh" "$DOMAIN" "--username=${AUTH_USERNAME}"
fi
