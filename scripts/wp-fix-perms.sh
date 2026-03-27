#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: ./scripts/wp-fix-perms.sh <domain> [--restart]

Fixes ownership/permissions of the WordPress docker volume in-place
without running a full restore.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }

[[ ${1:-} == "--help" || ${1:-} == "-h" ]] && { usage; exit 0; }
[[ $# -ge 1 ]] || { usage; exit 1; }

DOMAIN="$1"
shift || true
RESTART=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --restart) RESTART=1; shift ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

command -v docker >/dev/null 2>&1 || die "Tool fehlt: docker"

VOLUME="wp_${DOMAIN//./_}_html"
CONTAINER="wp-${DOMAIN//./-}"

docker volume inspect "$VOLUME" >/dev/null 2>&1 || die "Volume nicht gefunden: $VOLUME"

info "Setze Ownership und Berechtigungen für Volume: $VOLUME"
docker run --rm -v "${VOLUME}:/target" alpine sh -c '
  set -eu
  chown -R 33:33 /target
  find /target -type d -exec chmod 755 {} +
  find /target -type f -exec chmod 644 {} +
  if [ -d /target/wp-content ]; then
    find /target/wp-content -type d -exec chmod 775 {} +
    find /target/wp-content -type f -exec chmod 664 {} +
  fi
'

if [[ "$RESTART" -eq 1 ]]; then
  if docker ps --format '{{.Names}}' | grep -qx "$CONTAINER"; then
    info "Starte Container neu: $CONTAINER"
    docker restart "$CONTAINER" >/dev/null
  else
    info "Container läuft nicht, kein Restart ausgeführt: $CONTAINER"
  fi
fi

ok "Permission-Fix abgeschlossen für ${DOMAIN}"
