#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
INVENTORY="$ROOT_DIR/ansible/inventory/hosts.ini"
PLAYBOOK="$ROOT_DIR/ansible/playbooks/deploy-wordpress.yml"

CHECK_ONLY=0
REDEPLOY=0
ROTATE_EXISTING=0
TARGET_DOMAIN=""

usage() {
  cat <<'EOF'
Usage: ./scripts/wp-salt-rotate.sh [--check-only] [--redeploy] [--rotate-existing] [--domain <domain>]

Default:
  ergänzt fehlende WordPress-Salts, ohne bestehende zu ändern

Optionen:
  --check-only        nur anzeigen, was geändert würde
  --redeploy          geänderte Instanzen direkt neu deployen
  --rotate-existing   bestehende Salts vollständig neu erzeugen
  --domain <domain>   nur eine einzelne Domain bearbeiten
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
    --rotate-existing) ROTATE_EXISTING=1 ;;
    --domain)
      shift
      [[ $# -gt 0 ]] || die "Bitte Domain nach --domain angeben"
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
if [[ "$REDEPLOY" -eq 1 ]]; then
  require_cmd ansible-playbook
fi

[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"

mapfile -t files < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' | sort)
(( ${#files[@]} > 0 )) || die "Keine Hostvars-Dateien gefunden"

processed=0
changed=0
redeployed=0
changed_domains=()

for file in "${files[@]}"; do
  grep -q '^wp_domain_db:' "$file" || continue
  domain="$(basename "$file" .yml)"
  [[ -n "$TARGET_DOMAIN" && "$domain" != "$TARGET_DOMAIN" ]] && continue
  (( processed += 1 ))

  result="$(
    python3 - "$file" "$domain" "$ROTATE_EXISTING" "$CHECK_ONLY" <<'PY'
import secrets
import sys
from pathlib import Path

try:
    import yaml
except Exception as exc:
    print(f"ERROR|PyYAML fehlt: {exc}")
    sys.exit(0)

path = Path(sys.argv[1])
domain = sys.argv[2]
rotate_existing = sys.argv[3] == "1"
check_only = sys.argv[4] == "1"

keys = [
    "wp_auth_key",
    "wp_secure_auth_key",
    "wp_logged_in_key",
    "wp_nonce_key",
    "wp_auth_salt",
    "wp_secure_auth_salt",
    "wp_logged_in_salt",
    "wp_nonce_salt",
]

data = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
missing = [key for key in keys if not str(data.get(key, "") or "").strip()]
should_rotate = rotate_existing or bool(missing)

if not should_rotate:
    print(f"OK|{domain}|keine Änderung")
    sys.exit(0)

new_values = {key: secrets.token_urlsafe(48) for key in keys}
if not check_only:
    data.update(new_values)
    path.write_text(yaml.safe_dump(data, sort_keys=False, allow_unicode=True), encoding="utf-8")

reason = "fehlende Salts ergänzt" if missing and not rotate_existing else "Salts rotiert"
print(f"CHANGED|{domain}|{reason}")
PY
  )"

  status="${result%%|*}"
  rest="${result#*|}"
  if [[ "$status" == "ERROR" ]]; then
    die "$rest"
  elif [[ "$status" == "CHANGED" ]]; then
    changed_domains+=("$domain")
    (( changed += 1 ))
    info "${rest#*|}"
  else
    info "${rest#*|}"
  fi
done

(( processed > 0 )) || die "Keine passenden WordPress-Hostvars gefunden"

if [[ "$REDEPLOY" -eq 1 && ${#changed_domains[@]} -gt 0 ]]; then
  for domain in "${changed_domains[@]}"; do
    if [[ "$CHECK_ONLY" -eq 1 ]]; then
      info "würde ${domain} redeployen"
    else
      warn "${domain}: Salt-Rotation invalidiert bestehende Logins/Sessions"
      ansible-playbook -i "$INVENTORY" -e "target_domain=${domain}" "$PLAYBOOK"
      (( redeployed += 1 ))
    fi
  done
fi

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only abgeschlossen: ${processed} Instanz(en), ${changed} Änderung(en)"
else
  ok "Salt-Verwaltung abgeschlossen: ${processed} Instanz(en), ${changed} geändert"
  if [[ "$REDEPLOY" -eq 1 ]]; then
    ok "${redeployed} Instanz(en) redeployed"
  fi
fi
