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

verify_hostvars() {
  local add="$ROOT_DIR/scripts/easy-author-add.sh"
  local rotate="$ROOT_DIR/scripts/easy-author-rotate-secrets.sh"
  local template="$ROOT_DIR/ansible/hostvars/templates/easy-author-hostvars.j2"
  [[ -x "$add" && -x "$rotate" && -f "$template" ]] || fail "EasyAuthor-Hostvars-Helfer fehlen"
  require_text "$add" 'umask 077'
  require_text "$add" 'htpasswd -nB'
  require_text "$add" 'verify_domain_resolves_to_host_ipv4'
  require_text "$add" 'mv.*hostvars_file'
  reject_text "$template" '\$2[aby]\$[0-9][0-9]\$'
  require_text "$rotate" 'os.replace'
  require_text "$rotate" 'easy_author_basic_auth_password_hash'

  local sandbox bindir hostdir output file
  sandbox="$(mktemp -d)"
  trap 'rm -rf "$sandbox"' RETURN
  bindir="$sandbox/bin"; hostdir="$sandbox/hostvars"
  mkdir -p "$bindir" "$hostdir"
  cat > "$bindir/htpasswd" <<'SH'
#!/bin/sh
printf '%s:%s\n' "${4:-tester}" '$2y$12$abcdefghijklmnopqrstuv12345678901234567890123456789'
SH
  chmod +x "$bindir/htpasswd"
  output="$(PATH="$bindir:$PATH" EASY_AUTHOR_HOSTVARS_DIR="$hostdir" "$add" author.geller.men --username=tester --skip-dns-check)"
  file="$hostdir/author.geller.men.yml"
  [[ -f "$file" ]] || fail "Hostvars wurden nicht erstellt"
  [[ "$(stat -f '%Lp' "$file" 2>/dev/null || stat -c '%a' "$file")" == 600 ]] || fail "Hostvars sind nicht 0600"
  require_text "$file" '^easy_author_basic_auth_username: "tester"$'
  require_text "$file" '^easy_author_basic_auth_password_hash: "\$2y\$12\$'
  ! printf '%s\n' "$output" | grep -Eq '\$2y\$|123456789' || fail "Geheimnis wurde ausgegeben"
  if PATH="$bindir:$PATH" EASY_AUTHOR_HOSTVARS_DIR="$hostdir" "$add" author.geller.men --skip-dns-check >/dev/null 2>&1; then
    fail "Doppelte Hostvars wurden akzeptiert"
  fi
  PATH="$bindir:$PATH" EASY_AUTHOR_HOSTVARS_DIR="$hostdir" "$rotate" author.geller.men --username=reviewer >/dev/null
  require_text "$file" '^easy_author_basic_auth_username: "reviewer"$'
  require_text "$file" '^easy_author_basic_auth_password_hash: "\$2y\$12\$'
}

case "${1:-all}" in
  containers) verify_containers ;;
  hostvars) verify_hostvars ;;
  all) verify_containers; verify_hostvars ;;
  *) fail "Unbekannte Prüfgruppe: $1" ;;
esac
printf 'PASS: %s\n' "${1:-all}"
