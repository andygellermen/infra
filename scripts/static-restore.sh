#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REDEPLOY_SCRIPT="$ROOT_DIR/scripts/static-redeploy.sh"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/static-restore.sh <domain> <backup.tar.gz|backup.tgz|backup.zip> [--restore-hostvars] [--redeploy]
USAGE
}

normalize_domain() {
  local d="$1"
  if [[ "$d" =~ ^[a-zA-Z0-9.-]+$ ]]; then
    printf '%s\n' "$d"
  else
    idn --quiet --uts46 "$d"
  fi
}

set_hostvar_value() {
  local key="$1" value="$2" file="$3"
  if grep -q "^${key}:" "$file"; then
    sed -i -E "s|^${key}:.*|${key}: \"${value}\"|" "$file"
  else
    printf '%s: "%s"\n' "$key" "$value" >> "$file"
  fi
}

extract_backup_archive() {
  local archive="$1" dest="$2"
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

detect_static_docroot() {
  local base="$1"
  local candidate

  if find "$base" -maxdepth 1 -type f \( -iname 'index.html' -o -iname 'index.htm' \) | grep -q .; then
    printf '%s\n' "$base"
    return
  fi

  candidate="$(find "$base" -mindepth 1 -maxdepth 2 -type f \( -iname 'index.html' -o -iname 'index.htm' \) | head -n1 || true)"
  if [[ -n "$candidate" ]]; then
    dirname "$candidate"
    return
  fi

  candidate="$(find "$base" -mindepth 1 -maxdepth 2 -type f \( -iname '*.html' -o -iname '*.htm' \) | head -n1 || true)"
  if [[ -n "$candidate" ]]; then
    dirname "$candidate"
    return
  fi

  die "Kein statischer Document-Root im Backup erkannt (keine HTML-Datei gefunden)."
}

ensure_static_hostvars_exists() {
  local hostvars_file="$1" domain="$2"
  [[ -f "$hostvars_file" ]] && return 0

  warn "Hostvars fehlen für ${domain}. Erzeuge minimale Hostvars für Static-Site."
  mkdir -p "$(dirname "$hostvars_file")"
  cat > "$hostvars_file" <<EOF
# Hostvars für ${domain} (Static Site)
domain: ${domain}

traefik:
  domain: ${domain}
  aliases:
    - www.${domain}

static_enabled: true
static_traefik_middleware_default: "crowdsec-default@docker"
static_basic_auth_paths: []
EOF
  info "Hostvars neu erzeugt: $hostvars_file"
}

verify_static_hostvars() {
  local hostvars_file="$1" domain="$2"
  [[ -f "$hostvars_file" ]] || die "Hostvars fehlen: $hostvars_file"
  grep -q '^static_enabled:[[:space:]]*true' "$hostvars_file" || die "Hostvars gehören nicht zu einer statischen Site: $hostvars_file"
  awk -F': ' -v expected="$domain" '
    $1=="domain" {
      gsub(/"/, "", $2)
      gsub(/[[:space:]]+$/, "", $2)
      if ($2 == expected) found=1
    }
    END { exit(found ? 0 : 1) }
  ' "$hostvars_file" || die "Domain wurde in den Hostvars nicht korrekt gesetzt: ${domain}"
}

wait_for_container_running() {
  local container="$1" timeout="${2:-60}" waited=0
  while (( waited < timeout )); do
    if docker ps --format '{{.Names}}' | grep -qx "$container"; then
      return 0
    fi
    sleep 2
    (( waited += 2 ))
  done
  return 1
}

run_post_restore_checks() {
  local domain="$1" target_dir="$2" container="static-sites"
  local public_result="" public_status="" public_redirects=""

  info "Starte Post-Restore-Selbsttest"

  [[ -f "${target_dir}/index.html" || -f "${target_dir}/index.htm" ]] \
    || warn "Kein index.html/index.htm im Zielverzeichnis gefunden; Site kann trotzdem gültig sein"

  wait_for_container_running "$container" 60 || die "Shared Static-Container läuft nicht: $container"
  ok "Shared Static-Container läuft"

  if public_result="$(curl -k -sSIL --max-redirs 10 -o /dev/null -w '%{http_code} %{num_redirects}' "https://${domain}/" 2>/dev/null)"; then
    public_status="${public_result%% *}"
    public_redirects="${public_result##* }"
    case "$public_status" in
      200|301|302|303|307|308)
        ok "Öffentlicher HTTPS-Check erfolgreich (Finalstatus ${public_status}, Redirects ${public_redirects})"
        ;;
      *)
        die "Öffentlicher HTTPS-Check fehlgeschlagen (Finalstatus ${public_status}, Redirects ${public_redirects})"
        ;;
    esac
  else
    die "Öffentlicher HTTPS-Check konnte nicht ausgeführt werden"
  fi
}

[[ ${1:-} == "--help" || ${1:-} == "-h" ]] && { usage; exit 0; }
[[ $# -ge 2 ]] || { usage; exit 1; }

DOMAIN_RAW="$1"
BACKUP_FILE="$2"
shift 2

RESTORE_HOSTVARS=0
FORCE_REDEPLOY=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --restore-hostvars) RESTORE_HOSTVARS=1; shift ;;
    --redeploy) FORCE_REDEPLOY=1; shift ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

require_cmd docker
require_cmd curl
require_cmd tar
require_cmd idn

DOMAIN="$(normalize_domain "$DOMAIN_RAW")"
[[ -f "$BACKUP_FILE" ]] || die "Backup fehlt: $BACKUP_FILE"

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
TARGET_DIR="/srv/static/${DOMAIN}"
WORKDIR="$(mktemp -d /tmp/static-restore-${DOMAIN}.XXXXXX)"
trap 'rm -rf "$WORKDIR"' EXIT

extract_backup_archive "$BACKUP_FILE" "$WORKDIR"
DOCROOT="$(detect_static_docroot "$WORKDIR")"

info "Erkannter Document-Root: ${DOCROOT#$WORKDIR/}"

if [[ "$RESTORE_HOSTVARS" -eq 1 ]]; then
  if [[ -f "$WORKDIR/_infra/hostvars.yml" ]]; then
    mkdir -p "$(dirname "$HOSTVARS_FILE")"
    cp -a "$WORKDIR/_infra/hostvars.yml" "$HOSTVARS_FILE"
    info "Hostvars aus Backup wiederhergestellt"
  elif [[ -f "$WORKDIR/files/hostvars.yml" ]]; then
    mkdir -p "$(dirname "$HOSTVARS_FILE")"
    cp -a "$WORKDIR/files/hostvars.yml" "$HOSTVARS_FILE"
    info "Hostvars aus Backup wiederhergestellt (Legacy-Pfad)"
  else
    warn "--restore-hostvars gesetzt, aber keine hostvars.yml im Backup gefunden (optional)"
  fi
fi

ensure_static_hostvars_exists "$HOSTVARS_FILE" "$DOMAIN"
set_hostvar_value "domain" "$DOMAIN" "$HOSTVARS_FILE"
if grep -q '^static_enabled:' "$HOSTVARS_FILE"; then
  sed -i -E 's|^static_enabled:.*|static_enabled: true|' "$HOSTVARS_FILE"
else
  printf 'static_enabled: true\n' >> "$HOSTVARS_FILE"
fi
verify_static_hostvars "$HOSTVARS_FILE" "$DOMAIN"

warn "Restore überschreibt statische Dateien für ${DOMAIN} unter ${TARGET_DIR}"
read -r -p "Fortfahren? (yes/no): " a
[[ "$a" == yes ]] || die "Abbruch"

info "Restore Static Document Root"
docker run --rm \
  -v "${DOCROOT}:/src:ro" \
  -v /srv/static:/srv/static \
  alpine sh -c "set -eu; mkdir -p '/srv/static/${DOMAIN}'; find '/srv/static/${DOMAIN}' -mindepth 1 -delete; cp -a /src/. '/srv/static/${DOMAIN}/'; chown -R 33:33 '/srv/static/${DOMAIN}'; find '/srv/static/${DOMAIN}' -type d -exec chmod 755 {} +; find '/srv/static/${DOMAIN}' -type f -exec chmod 644 {} +"

info "Führe Static-Redeploy aus"
"$REDEPLOY_SCRIPT" "$DOMAIN"

run_post_restore_checks "$DOMAIN" "$TARGET_DIR"

ok "Static-Restore abgeschlossen"
