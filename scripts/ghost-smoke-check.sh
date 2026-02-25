#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<USAGE
Usage:
  $0 <domain>

Checks:
  - GET /                (Site)
  - GET /ghost/          (Admin route)
  - GET /ghost/api/admin/site/ (API reachability)
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }

DOMAIN="${1:-}"
if [[ "$DOMAIN" == "--help" || "$DOMAIN" == "-h" || -z "$DOMAIN" ]]; then
  usage
  [[ -n "$DOMAIN" ]] && exit 0 || exit 1
fi

check_url() {
  local label="$1"
  local url="$2"
  local expected_regex="$3"
  local code

  code="$(curl -k -sS -o /tmp/ghost-smoke-body.$$ -w '%{http_code}' "$url" || echo 000)"
  if [[ "$code" =~ $expected_regex ]]; then
    ok "$label -> HTTP $code"
  else
    echo "❌ $label -> HTTP $code (erwartet: $expected_regex)" >&2
    echo "--- Body (erste 20 Zeilen) ---" >&2
    sed -n '1,20p' /tmp/ghost-smoke-body.$$ >&2 || true
    rm -f /tmp/ghost-smoke-body.$$
    return 1
  fi
}

info "Smoke-Check für https://${DOMAIN}"

# Site should be reachable (200 normally; sometimes 301/302 depending on config)
check_url "Frontend /" "https://${DOMAIN}/" '^(200|301|302)$'

# Admin route should not be 404 at edge; often 200 or redirect
check_url "Admin /ghost/" "https://${DOMAIN}/ghost/" '^(200|301|302)$'

# API endpoint should be reachable (usually 401/403/200, but not 404 from router mismatch)
check_url "API /ghost/api/admin/site/" "https://${DOMAIN}/ghost/api/admin/site/" '^(200|401|403)$'

rm -f /tmp/ghost-smoke-body.$$
ok "Smoke-Check erfolgreich für ${DOMAIN}"
