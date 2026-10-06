#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail(){ printf 'FAIL: %s\n' "$*" >&2; exit 1; }
require_text(){ grep -Eq "$2" "$1" || fail "$1 fehlt: $2"; }
reject_text(){ ! grep -Eq "$2" "$1" || fail "$1 enthaelt unerlaubt: $2"; }

verify_containers() {
  local frontend="$ROOT_DIR/apps/easy-author/frontend/Dockerfile"
  local nginx="$ROOT_DIR/apps/easy-author/frontend/nginx.conf"
  local backend="$ROOT_DIR/apps/easy-author/backend/Dockerfile"
  local compose="$ROOT_DIR/apps/easy-author/docker-compose.yml"
  [[ -f "$nginx" ]] || fail "nginx.conf fehlt"
  require_text "$frontend" 'AS development'
  require_text "$frontend" 'AS build'
  require_text "$frontend" 'AS production'
  require_text "$frontend" 'nginx-unprivileged'
  require_text "$frontend" 'COPY --from=build /app/dist'
  require_text "$frontend" 'COPY nginx.conf'
  require_text "$frontend" 'EXPOSE 8080'
  reject_text <(sed -n '/AS production/,$p' "$frontend") 'npm run dev'
  require_text "$nginx" 'try_files.*index.html'
  require_text "$nginx" 'proxy_pass http://easy-author-api:8086'
  require_text "$nginx" 'X-Content-Type-Options'
  require_text "$nginx" 'Referrer-Policy'
  require_text "$nginx" 'Content-Security-Policy'
  require_text "$compose" 'target: development'
  require_text "$compose" '"5173:5173"'
  require_text "$compose" '"8086:8086"'
  require_text "$backend" '^EXPOSE 8086$'
  reject_text "$backend" 'apt-get|apk add|COPY .*data'
}

case "${1:-all}" in
  containers) verify_containers ;;
  all) verify_containers ;;
  *) fail "Unbekannte Prüfgruppe: $1" ;;
esac
printf 'PASS: %s\n' "${1:-all}"
