#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
HOSTVARS_NORMALIZER="$ROOT_DIR/scripts/wp-normalize-hostvars.py"
INVENTORY="$ROOT_DIR/ansible/inventory/hosts.ini"
PLAYBOOK="$ROOT_DIR/ansible/playbooks/deploy-wordpress.yml"

CHECK_ONLY=0
REDEPLOY=0
FORCE_XMLRPC_BLOCK=0
TARGET_DOMAIN=""

usage() {
  cat <<'EOF'
Usage: ./scripts/wp-rollout-hardening.sh [--check-only] [--redeploy] [--force-xmlrpc-block] [--domain <domain>]

  --check-only          Nur prüfen und geplante Änderungen ausgeben
  --redeploy            Nach Hostvars-Anpassungen jede WordPress-Instanz redeployen
  --force-xmlrpc-block  Setzt wp_xmlrpc_protection explizit auf "block"
  --domain <domain>     Nur eine einzelne Domain bearbeiten
EOF
}

die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=1 ;;
    --redeploy) REDEPLOY=1 ;;
    --force-xmlrpc-block) FORCE_XMLRPC_BLOCK=1 ;;
    --domain)
      shift
      [[ $# -gt 0 ]] || die "Bitte eine Domain nach --domain angeben"
      TARGET_DOMAIN="$1"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "Unbekanntes Argument: $1"
      ;;
  esac
  shift
done

require_cmd python3
[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"
[[ -f "$HOSTVARS_NORMALIZER" ]] || die "Hostvars-Normalizer fehlt: $HOSTVARS_NORMALIZER"
if [[ "$REDEPLOY" -eq 1 ]]; then
  require_cmd ansible-playbook
fi

defaults=(
  'wp_traefik_middleware_default: "crowdsec-default@docker"'
  'wp_traefik_middleware_admin: "crowdsec-admin@docker"'
  'wp_traefik_middleware_api: "crowdsec-api@docker"'
  'wp_xmlrpc_protection: "block"'
  'wp_xmlrpc_allowed_cidrs: []'
)

set_scalar_value() {
  local file="$1" key="$2" value="$3"
  if grep -qE "^${key}:" "$file"; then
    if [[ "$CHECK_ONLY" -eq 1 ]]; then
      info "$file: würde ${key} auf ${value} setzen"
    else
      python3 - "$file" "$key" "$value" <<'PY'
import sys
from pathlib import Path

path = Path(sys.argv[1])
key = sys.argv[2]
value = sys.argv[3]
lines = path.read_text(encoding="utf-8").splitlines()
updated = []
done = False
for line in lines:
    if not done and line.startswith(f"{key}:"):
        updated.append(f'{key}: "{value}"')
        done = True
    else:
        updated.append(line)
path.write_text("\n".join(updated) + "\n", encoding="utf-8")
PY
    fi
  else
    if [[ "$CHECK_ONLY" -eq 1 ]]; then
      info "$file: würde ${key}: \"${value}\" ergänzen"
    else
      printf '%s: "%s"\n' "$key" "$value" >> "$file"
    fi
  fi
}

append_line_if_missing() {
  local file="$1" line="$2" key
  key="${line%%:*}"
  if grep -qE "^${key}:" "$file"; then
    return 0
  fi
  if [[ "$CHECK_ONLY" -eq 1 ]]; then
    info "$file: würde ${line} ergänzen"
  else
    printf '%s\n' "$line" >> "$file"
  fi
}

mapfile -t files < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' | sort)
(( ${#files[@]} > 0 )) || die "Keine Hostvars-Dateien gefunden"

processed=0
redeployed=0

for file in "${files[@]}"; do
  grep -q '^wp_domain_db:' "$file" || continue
  domain="$(basename "$file" .yml)"
  [[ -n "$TARGET_DOMAIN" && "$domain" != "$TARGET_DOMAIN" ]] && continue

  (( processed += 1 ))
  info "Prüfe WordPress-Hostvars für ${domain}"

  if [[ "$CHECK_ONLY" -eq 1 ]]; then
    info "$file: würde Traefik-Aliase für ${domain} normalisieren"
  else
    python3 "$HOSTVARS_NORMALIZER" "$file" "$domain"
  fi

  for line in "${defaults[@]}"; do
    append_line_if_missing "$file" "$line"
  done

  if grep -qE '^wp_xmlrpc_protection:\s*["'\'']?(off|allowlist)["'\'']?\s*$' "$file"; then
    if [[ "$FORCE_XMLRPC_BLOCK" -eq 1 ]]; then
      set_scalar_value "$file" "wp_xmlrpc_protection" "block"
      if [[ "$CHECK_ONLY" -eq 1 ]]; then
        info "$file: vorhandene XML-RPC-Ausnahme würde auf block gesetzt"
      else
        warn "${domain}: vorhandene XML-RPC-Ausnahme wurde auf block gesetzt"
      fi
    else
      warn "${domain}: wp_xmlrpc_protection ist explizit nicht 'block' gesetzt"
    fi
  fi

  if [[ "$REDEPLOY" -eq 1 ]]; then
    if [[ "$CHECK_ONLY" -eq 1 ]]; then
      info "würde ${domain} redeployen"
    else
      ansible-playbook -i "$INVENTORY" -e "target_domain=${domain}" "$PLAYBOOK"
      (( redeployed += 1 ))
    fi
  fi
done

(( processed > 0 )) || die "Keine passenden WordPress-Hostvars gefunden"

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only abgeschlossen für ${processed} WordPress-Instanz(en)"
else
  ok "WordPress-Hardening auf ${processed} Instanz(en) angewendet"
  if [[ "$REDEPLOY" -eq 1 ]]; then
    ok "${redeployed} WordPress-Instanz(en) redeployed"
  else
    info "Noch kein Redeploy durchgeführt. Dafür: ./scripts/wp-rollout-hardening.sh --redeploy"
  fi
fi
