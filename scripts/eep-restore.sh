#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REDEPLOY_SCRIPT="$ROOT_DIR/scripts/eep-redeploy.sh"
BACKUP_SCRIPT="$ROOT_DIR/scripts/eep-backup.sh"
SMOKE_SCRIPT="$ROOT_DIR/scripts/eep-smoke-check.sh"
BACKUP_DIR="${EEP_BACKUP_DIR:-$ROOT_DIR/backups/eep}"
HOSTVARS_DIR="${EEP_HOSTVARS_DIR:-$ROOT_DIR/ansible/hostvars}"
SITE_ROOT="${EEP_SITE_ROOT:-/srv/easy-event-planner}"

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/eep-restore.sh <domain> <backup.tar.gz|backup.tgz|backup.zip> [--restore-hostvars] [--yes] [--redeploy] [--skip-smoke] [--wildcard-domain=<apex-domain>] [--dns-account=<key>]
USAGE
}

normalize_domain() {
  local domain="$1"
  if [[ "$domain" =~ ^[a-zA-Z0-9.-]+$ ]]; then
    printf '%s\n' "$domain"
  else
    command -v idn >/dev/null 2>&1 || die "Domain enthaelt Nicht-ASCII-Zeichen, aber Tool fehlt: idn"
    idn --quiet --uts46 "$domain"
  fi
}

trim_value() {
  local value="$1"
  value="${value#\"}"
  value="${value%\"}"
  printf '%s' "$value" | sed 's/[[:space:]]*$//'
}

extract_hostvar() {
  local key="$1"
  local file="$2"
  local value
  value="$(awk -F': ' -v k="$key" '$1==k {print $2; exit}' "$file")"
  trim_value "$value"
}

set_hostvar_value() {
  local key="$1"
  local value="$2"
  local file="$3"
  if grep -q "^${key}:" "$file"; then
    sed -i -E "s|^${key}:.*|${key}: \"${value}\"|" "$file"
  else
    printf '%s: "%s"\n' "$key" "$value" >> "$file"
  fi
}

extract_backup_archive() {
  local archive="$1"
  local dest="$2"
  case "$archive" in
    *.tar.gz|*.tgz)
      tar xzf "$archive" -C "$dest"
      ;;
    *.zip)
      require_cmd unzip
      unzip -q "$archive" -d "$dest"
      ;;
    *)
      die "Unbekanntes Backup-Format: $archive (erwartet .tar.gz, .tgz oder .zip)"
      ;;
  esac
}

detect_backup_root() {
  local base="$1"
  local candidate

  if [[ -f "$base/_infra/manifest.env" || -f "$base/easy-event-planner.env" || -d "$base/data" ]]; then
    printf '%s\n' "$base"
    return
  fi

  candidate="$(find "$base" -mindepth 1 -maxdepth 3 -type f \( -path '*/_infra/manifest.env' -o -path '*/easy-event-planner.env' \) | head -n1 || true)"
  [[ -n "$candidate" ]] || die "Backup-Struktur konnte nicht erkannt werden"
  dirname "$candidate" | sed 's|/_infra$||'
}

verify_eep_hostvars() {
  local hostvars_file="$1"
  [[ -f "$hostvars_file" ]] || die "Hostvars fehlen: $hostvars_file"
  grep -Eq '^eep_enabled:[[:space:]]*true([[:space:]]|$)' "$hostvars_file" || die "Hostvars gehoeren nicht zu einer aktiven EEP-Instanz: $hostvars_file"
}

restore_site_tree() {
  local source_dir="$1"
  local domain="$2"
  local host_uid
  local host_gid

  host_uid="$(id -u)"
  host_gid="$(id -g)"

  docker run --rm \
    -e HOST_UID="$host_uid" \
    -e HOST_GID="$host_gid" \
    -v "${SITE_ROOT}:/srv/eep-root" \
    -v "${source_dir}:/src:ro" \
    alpine \
    sh -c "set -eu; mkdir -p '/srv/eep-root/${domain}'; find '/srv/eep-root/${domain}' -mindepth 1 -delete; cd /src; for entry in .??* .[!.]* *; do [ \"\$entry\" = '_infra' ] && continue; [ -e \"\$entry\" ] || continue; cp -a \"\$entry\" \"/srv/eep-root/${domain}/\"; done; chown -R \"\$HOST_UID:\$HOST_GID\" '/srv/eep-root/${domain}'"
}

container_exists() {
  local name="$1"
  docker ps -a --format '{{.Names}}' | grep -qx "$name"
}

container_running() {
  local name="$1"
  docker ps --format '{{.Names}}' | grep -qx "$name"
}

stop_container_if_present() {
  local name="$1"
  if container_running "$name"; then
    info "Stoppe Container: $name"
    docker stop "$name" >/dev/null
  fi
}

start_container_if_present() {
  local name="$1"
  if container_exists "$name"; then
    info "Starte Container: $name"
    docker start "$name" >/dev/null || true
  fi
}

run_post_restore_smoke() {
  local hostvars_file="$1"
  local domain="$2"
  local base_url retries delay timeout insecure

  [[ -x "$SMOKE_SCRIPT" ]] || warn "Smoke-Skript fehlt oder ist nicht ausführbar: $SMOKE_SCRIPT"
  [[ -x "$SMOKE_SCRIPT" ]] || return 0

  base_url="$(extract_hostvar eep_smoke_base_url "$hostvars_file")"
  [[ -n "$base_url" ]] || base_url="$(extract_hostvar eep_base_url "$hostvars_file")"
  [[ -n "$base_url" ]] || base_url="https://${domain}"

  retries="$(extract_hostvar eep_smoke_retries "$hostvars_file")"
  delay="$(extract_hostvar eep_smoke_delay_seconds "$hostvars_file")"
  timeout="$(extract_hostvar eep_smoke_timeout_seconds "$hostvars_file")"
  insecure="$(extract_hostvar eep_smoke_insecure "$hostvars_file")"

  [[ -n "$retries" ]] || retries="15"
  [[ -n "$delay" ]] || delay="2"
  [[ -n "$timeout" ]] || timeout="10"
  [[ "$insecure" == "true" ]] || insecure="false"

  "$SMOKE_SCRIPT" "$domain" \
    --base-url="$base_url" \
    --retries="$retries" \
    --delay="$delay" \
    --timeout="$timeout" \
    --insecure="$insecure"
}

[[ ${1:-} == "--help" || ${1:-} == "-h" ]] && { usage; exit 0; }
[[ $# -ge 2 ]] || { usage; exit 1; }

DOMAIN="$(normalize_domain "$1")"
BACKUP_FILE="$2"
shift 2

RESTORE_HOSTVARS=0
ASSUME_YES=0
FORCE_REDEPLOY=0
SKIP_SMOKE=0
WILDCARD_DOMAIN=""
DNS_ACCOUNT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --restore-hostvars) RESTORE_HOSTVARS=1; shift ;;
    --yes) ASSUME_YES=1; shift ;;
    --redeploy) FORCE_REDEPLOY=1; shift ;;
    --skip-smoke) SKIP_SMOKE=1; shift ;;
    --wildcard-domain=*) WILDCARD_DOMAIN="$(normalize_domain "${1#*=}")"; shift ;;
    --dns-account=*) DNS_ACCOUNT="${1#*=}"; shift ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ -f "$BACKUP_FILE" ]] || die "Backup nicht gefunden: $BACKUP_FILE"

require_cmd docker
require_cmd tar
require_cmd sed
require_cmd awk
require_cmd grep

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
SITE_DIR="$SITE_ROOT/${DOMAIN}"
WORKDIR="$(mktemp -d /tmp/eep-restore-${DOMAIN}.XXXXXX)"
trap 'rm -rf "$WORKDIR"' EXIT

extract_backup_archive "$BACKUP_FILE" "$WORKDIR"
BACKUP_ROOT="$(detect_backup_root "$WORKDIR")"

[[ -d "$BACKUP_ROOT/data" || -f "$BACKUP_ROOT/easy-event-planner.env" ]] || die "EEP-Backup enthaelt weder data/ noch easy-event-planner.env"

if [[ "$RESTORE_HOSTVARS" -eq 1 ]]; then
  [[ -f "$BACKUP_ROOT/_infra/hostvars.yml" ]] || die "hostvars.yml fehlt im Backup (oder ohne --restore-hostvars ausfuehren)"
fi

if [[ "$ASSUME_YES" -ne 1 ]]; then
  if [[ "$RESTORE_HOSTVARS" -eq 1 ]]; then
    echo "⚠️  Restore überschreibt EEP-Dateien und Hostvars für $DOMAIN"
  else
    echo "⚠️  Restore überschreibt EEP-Dateien für $DOMAIN (Hostvars bleiben unverändert)"
  fi
  read -r -p "Fortfahren? (yes/no): " answer
  [[ "$answer" == "yes" ]] || die "Abgebrochen"
fi

if [[ "$RESTORE_HOSTVARS" -eq 1 ]]; then
  mkdir -p "$(dirname "$HOSTVARS_FILE")"
  cp -a "$BACKUP_ROOT/_infra/hostvars.yml" "$HOSTVARS_FILE"
  info "Hostvars aus Backup wiederhergestellt"
  FORCE_REDEPLOY=1
fi

verify_eep_hostvars "$HOSTVARS_FILE"

if [[ -n "$WILDCARD_DOMAIN" ]]; then
  set_hostvar_value "tls_mode" "wildcard" "$HOSTVARS_FILE"
  set_hostvar_value "tls_wildcard_domain" "$WILDCARD_DOMAIN" "$HOSTVARS_FILE"
  [[ -n "$DNS_ACCOUNT" ]] && set_hostvar_value "tls_dns_account" "$DNS_ACCOUNT" "$HOSTVARS_FILE"
  FORCE_REDEPLOY=1
  info "Wildcard-TLS aktiviert: *.${WILDCARD_DOMAIN}${DNS_ACCOUNT:+ via DNS-Account ${DNS_ACCOUNT}}"
elif [[ -n "$DNS_ACCOUNT" ]]; then
  set_hostvar_value "tls_dns_account" "$DNS_ACCOUNT" "$HOSTVARS_FILE"
  FORCE_REDEPLOY=1
fi

APP_CONTAINER="$(extract_hostvar eep_container_name "$HOSTVARS_FILE")"
WORKER_CONTAINER="$(extract_hostvar eep_worker_container_name "$HOSTVARS_FILE")"
[[ -n "$APP_CONTAINER" ]] || APP_CONTAINER="easy-event-planner-${DOMAIN//./-}"
[[ -n "$WORKER_CONTAINER" ]] || WORKER_CONTAINER="easy-event-planner-worker-${DOMAIN//./-}"

SAFETY_DIR="$BACKUP_DIR/${DOMAIN}/safety-$(date +%Y%m%d-%H%M%S)"
if [[ -d "$SITE_DIR" && -x "$BACKUP_SCRIPT" ]]; then
  mkdir -p "$SAFETY_DIR"
  info "Erzeuge Safety-Backup vor Restore"
  EEP_SITE_ROOT="$SITE_ROOT" EEP_HOSTVARS_DIR="$HOSTVARS_DIR" EEP_BACKUP_DIR="$BACKUP_DIR" \
    "$BACKUP_SCRIPT" --create "$DOMAIN" --output "$SAFETY_DIR/pre-restore.tar.gz" || warn "Safety-Backup konnte nicht vollständig erstellt werden"
fi

stop_container_if_present "$APP_CONTAINER"
stop_container_if_present "$WORKER_CONTAINER"

info "Stelle EEP-Dateien wieder her nach $SITE_DIR"
restore_site_tree "$BACKUP_ROOT" "$DOMAIN"

if [[ "$FORCE_REDEPLOY" -eq 1 ]]; then
  [[ -x "$REDEPLOY_SCRIPT" ]] || die "Redeploy-Skript fehlt oder ist nicht ausführbar: $REDEPLOY_SCRIPT"
  info "Starte Redeploy nach Restore"
  "$REDEPLOY_SCRIPT" "$DOMAIN"
else
  start_container_if_present "$APP_CONTAINER"
  start_container_if_present "$WORKER_CONTAINER"
fi

if [[ "$SKIP_SMOKE" -eq 0 ]]; then
  run_post_restore_smoke "$HOSTVARS_FILE" "$DOMAIN"
fi

ok "EEP-Restore abgeschlossen"
[[ -d "$SAFETY_DIR" ]] && echo "📄 Safety-Backup: $SAFETY_DIR"
