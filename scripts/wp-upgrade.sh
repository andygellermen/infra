#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
REDEPLOY_SCRIPT="$ROOT_DIR/scripts/wp-redeploy.sh"

usage() {
  cat <<USAGE
Usage: $0 <domain> --version=<tag|latest> [--dry-run]
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }

[[ $# -ge 2 ]] || { usage; exit 1; }
DOMAIN="$1"; shift
TARGET_VERSION=""
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --version=*) TARGET_VERSION="${arg#*=}" ;;
    --dry-run) DRY_RUN=1 ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $arg" ;;
  esac
done

[[ -n "$TARGET_VERSION" ]] || die "Bitte --version=<tag|latest> angeben."
HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars fehlt: $HOSTVARS_FILE"

current_version="$(awk -F': ' '/^wp_version:/ {gsub(/["[:space:]]/,"",$2); print $2; exit}' "$HOSTVARS_FILE")"
[[ -n "$current_version" ]] || current_version="latest"
info "Aktuelle Version: $current_version"
info "Zielversion: $TARGET_VERSION"

cp -a "$HOSTVARS_FILE" "$HOSTVARS_FILE.bak.$(date +%Y%m%d-%H%M%S)"
if grep -q '^wp_version:' "$HOSTVARS_FILE"; then
  sed -i -E "s/^wp_version: .*/wp_version: \"${TARGET_VERSION}\"/" "$HOSTVARS_FILE"
else
  printf '\nwp_version: "%s"\n' "$TARGET_VERSION" >> "$HOSTVARS_FILE"
fi

ok "wp_version in Hostvars aktualisiert"

if [[ "$DRY_RUN" -eq 1 ]]; then
  ok "Dry-run aktiv: kein Redeploy ausgeführt"
  exit 0
fi

"$REDEPLOY_SCRIPT" "$DOMAIN"
ok "WordPress-Upgrade-/Redeploy abgeschlossen"
