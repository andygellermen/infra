#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<USAGE
Usage: $0 <domain>

Diagnose Ghost/Traefik routing for ActivityPub endpoint.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }

DOMAIN="${1:-}"
[[ -n "$DOMAIN" ]] || { usage; exit 1; }

GHOST_CONTAINER="ghost-${DOMAIN//./-}"
[[ "$(docker ps --format '{{.Names}}' | rg -x "$GHOST_CONTAINER" || true)" == "$GHOST_CONTAINER" ]] || die "Ghost-Container nicht gefunden: $GHOST_CONTAINER"

info "Ghost-Container: $GHOST_CONTAINER"

info "1) DNS-Auflösung im Ghost-Container"
docker exec "$GHOST_CONTAINER" getent hosts "$DOMAIN" || true

RESOLVED_IP="$(docker exec "$GHOST_CONTAINER" getent hosts "$DOMAIN" 2>/dev/null | awk '{print $1; exit}')"
if [[ -n "$RESOLVED_IP" ]]; then
  info "2) Container mit IP ${RESOLVED_IP} im Docker-Netz suchen"
  docker ps -q | while read -r cid; do
    name="$(docker inspect -f '{{.Name}}' "$cid" | sed 's#^/##')"
    ips="$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}' "$cid")"
    if [[ " $ips " == *" $RESOLVED_IP "* ]]; then
      echo "   -> ${name} (${ips})"
    fi
  done
fi

info "3) TCP-Test aus Ghost zu ${DOMAIN}:443"
docker exec "$GHOST_CONTAINER" sh -lc "(echo >/dev/tcp/${DOMAIN}/443) >/dev/null 2>&1 && echo '   -> TCP 443 erreichbar' || echo '   -> TCP 443 NICHT erreichbar'"

info "4) HTTP-Test von Host via Public URL"
curl -sS -o /tmp/ghost-route-body.txt -D /tmp/ghost-route-headers.txt "https://${DOMAIN}/.ghost/activitypub/v1/site/" || true
awk 'NR==1{print "   -> " $0}' /tmp/ghost-route-headers.txt || true

info "5) Kürzlicher Traefik-Accesslog-Ausschnitt"
docker logs --tail 50 traefik 2>&1 | rg "\.ghost/activitypub/v1/site/| GET / HTTP" || true

info "Fertig."
