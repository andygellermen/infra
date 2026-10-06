#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fail(){ printf 'FAIL: %s\n' "$*" >&2; exit 1; }
require_text(){ grep -Eq -- "$2" "$1" || fail "$1 fehlt: $2"; }
reject_text(){ ! grep -Eq -- "$2" "$1" || fail "$1 enthaelt unerlaubt: $2"; }
file_mode(){
  if stat -c '%a' "$1" >/dev/null 2>&1; then
    stat -c '%a' "$1"
  else
    stat -f '%Lp' "$1"
  fi
}

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
  [[ "$(file_mode "$file")" == 600 ]] || fail "Hostvars sind nicht 0600"
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

verify_backup() {
  local script="$ROOT_DIR/scripts/easy-author-backup.sh" sandbox bindir root archive
  [[ -x "$script" ]] || fail "Backup-Helfer fehlt"
  require_text "$script" '\.partial'
  require_text "$script" 'tar -tf'
  require_text "$script" 'trap restart'
  sandbox="$(mktemp -d)"; trap 'rm -rf "$sandbox"' RETURN
  bindir="$sandbox/bin"; root="$sandbox/srv"; mkdir -p "$bindir" "$root/author.geller.men/data/library"
  printf 'sqlite' > "$root/author.geller.men/data/easy-author.sqlite"
  printf 'book' > "$root/author.geller.men/data/library/book.md"
  cat > "$bindir/docker" <<'SH'
#!/bin/sh
case "$1" in inspect) printf 'true\n';; stop|start) printf '%s %s\n' "$1" "$2" >> "$EASY_AUTHOR_DOCKER_LOG";; esac
SH
  chmod +x "$bindir/docker"
  archive="$(EASY_AUTHOR_SRV_ROOT="$root" EASY_AUTHOR_DOCKER_BIN="$bindir/docker" EASY_AUTHOR_DOCKER_LOG="$sandbox/docker.log" "$script" author.geller.men --retention=2)"
  [[ -f "$archive" && "$(file_mode "$archive")" == 600 ]] || fail "Backup fehlt oder hat falschen Modus"
  tar -tf "$archive" | grep -qx 'data/easy-author.sqlite' || fail "SQLite fehlt im Backup"
  grep -q '^stop ' "$sandbox/docker.log" && grep -q '^start ' "$sandbox/docker.log" || fail "API wurde nicht kontrolliert neu gestartet"
  ! find "$root" -name '*.partial' | grep -q . || fail "Partielles Archiv blieb liegen"
}

verify_restore() {
  local script="$ROOT_DIR/scripts/easy-author-restore.sh" backup="$ROOT_DIR/scripts/easy-author-backup.sh"
  local sandbox bindir root archive
  [[ -x "$script" ]] || fail "Restore-Helfer fehlt"
  require_text "$script" '--confirm='
  require_text "$script" 'tar -tf'
  require_text "$script" '\.restore-'
  sandbox="$(mktemp -d)"; trap 'rm -rf "$sandbox"' RETURN
  bindir="$sandbox/bin"; root="$sandbox/srv"; mkdir -p "$bindir" "$root/author.geller.men/data/library" "$sandbox/source/data/library"
  printf 'old' > "$root/author.geller.men/data/easy-author.sqlite"
  printf 'new' > "$sandbox/source/data/easy-author.sqlite"
  printf 'new-book' > "$sandbox/source/data/library/book.md"
  archive="$root/author.geller.men/easy-author-author.geller.men-20261006T000000Z.tar.gz"
  tar -C "$sandbox/source" -czf "$archive" data
  cat > "$bindir/docker" <<'SH'
#!/bin/sh
case "$1" in inspect) printf 'false\n';; stop|start) :;; esac
SH
  chmod +x "$bindir/docker"
  if EASY_AUTHOR_SRV_ROOT="$root" EASY_AUTHOR_DOCKER_BIN="$bindir/docker" "$script" author.geller.men "$archive" >/dev/null 2>&1; then
    fail "Restore ohne Bestaetigung wurde akzeptiert"
  fi
  EASY_AUTHOR_SRV_ROOT="$root" EASY_AUTHOR_DOCKER_BIN="$bindir/docker" EASY_AUTHOR_BACKUP_BIN="$backup" "$script" author.geller.men "$archive" --confirm=author.geller.men >/dev/null
  grep -q '^new$' "$root/author.geller.men/data/easy-author.sqlite" || fail "Restore ersetzte SQLite nicht"
}

verify_ansible() {
  local role="$ROOT_DIR/ansible/playbooks/roles/easy-author/tasks/main.yml"
  local playbook="$ROOT_DIR/ansible/playbooks/deploy-easy-author.yml"
  [[ -f "$role" && -f "$playbook" ]] || fail "EasyAuthor-Ansible fehlt"
  require_text "$role" 'no_log: true'
  require_text "$role" 'easy_author_basic_auth_password_hash'
  require_text "$role" "password_hash is match"
  require_text "$role" 'internal: true'
  require_text "$role" 'name: traefik'
  require_text "$role" 'removeheader.*true'
  require_text "$role" 'certresolver'
  require_text "$role" 'redirectscheme'
  require_text "$role" 'published_ports: \[\]'
  require_text "$role" 'loadbalancer.server.port.*8080'
  reject_text "$role" '^[[:space:]]+"?traefik\.[^:]*[{][{]'
  [[ "$(grep -c 'when: not ansible_check_mode' "$role")" -ge 2 ]] || fail "Container-Tasks sind im frischen Check-Mode nicht geschuetzt"
  require_text "$playbook" 'become: true'
  require_text "$playbook" '- easy-author'
}

verify_redeploy() {
  local script="$ROOT_DIR/scripts/easy-author-redeploy.sh"
  [[ -x "$script" ]] || fail "Redeploy-Helfer fehlt"
  require_text "$script" 'git .*rev-parse HEAD'
  require_text "$script" 'git .*diff --quiet'
  require_text "$script" 'easy_author_deploy_api_image'
  require_text "$script" 'easy_author_deploy_web_image'
  require_text "$script" 'easy-author-backup.sh'
  require_text "$script" 'easy-author-smoke-check.sh'
  require_text "$script" '--target production'
}

verify_smoke() {
  local script="$ROOT_DIR/scripts/easy-author-smoke-check.sh"
  [[ -x "$script" ]] || fail "Smoke-Helfer fehlt"
  require_text "$script" 'http://.*domain'
  require_text "$script" '401'
  require_text "$script" '/api/health'
  require_text "$script" '/api/projects'
  require_text "$script" '/api/kanban'
  require_text "$script" 'backlog.*todo.*in_progress.*review.*done'
  require_text "$script" 'DOCKER_BIN.*restart'
  require_text "$script" 'PortBindings'
  require_text "$script" 'mktemp'
  require_text "$script" 'chmod 0600'
  require_text "$script" 'read -r -s'
  require_text "$script" 'trap cleanup'
}

case "${1:-all}" in
  containers) verify_containers ;;
  hostvars) verify_hostvars ;;
  backup) verify_backup ;;
  restore) verify_restore ;;
  ansible) verify_ansible ;;
  redeploy) verify_redeploy ;;
  smoke) verify_smoke ;;
  all) verify_containers; verify_hostvars; verify_backup; verify_restore; verify_ansible; verify_redeploy; verify_smoke ;;
  *) fail "Unbekannte Prüfgruppe: $1" ;;
esac
printf 'PASS: %s\n' "${1:-all}"
