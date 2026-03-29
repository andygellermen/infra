#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXPORT_SCRIPT="$ROOT_DIR/scripts/wildcard-export.sh"
DEFAULT_CONFIG="$ROOT_DIR/ansible/wildcards/distribution.example.yml"

usage() {
  cat <<'USAGE'
Usage: ./scripts/wildcard-distribute.sh <apex-domain> [--config /pfad/distribution.yml]
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

normalize_domain() {
  local domain="$1"
  if [[ "$domain" =~ ^[A-Za-z0-9.-]+$ ]]; then
    printf '%s\n' "${domain,,}"
  else
    require_cmd idn
    idn --quiet --uts46 "$domain" | tr '[:upper:]' '[:lower:]'
  fi
}

CONFIG_FILE="$DEFAULT_CONFIG"
ARGS=()
for arg in "$@"; do
  case "$arg" in
    --config=*) CONFIG_FILE="${arg#*=}" ;;
    --help|-h) usage; exit 0 ;;
    *) ARGS+=("$arg") ;;
  esac
done

[[ ${#ARGS[@]} -eq 1 ]] || { usage; exit 1; }
require_cmd python3
require_cmd scp
require_cmd ssh

DOMAIN="$(normalize_domain "${ARGS[0]}")"
[[ -f "$CONFIG_FILE" ]] || die "Distributions-Konfiguration fehlt: $CONFIG_FILE"

EXPORT_DIR="$(mktemp -d "/tmp/wildcard-distribute-${DOMAIN}.XXXXXX")"
cleanup() {
  rm -rf "$EXPORT_DIR"
}
trap cleanup EXIT

"$EXPORT_SCRIPT" "$DOMAIN" --output-dir="$EXPORT_DIR" >/dev/null

mapfile -t TARGET_LINES < <(python3 - "$DOMAIN" "$CONFIG_FILE" <<'PY'
import sys
try:
    import yaml
except Exception as exc:
    print(f"❌ PyYAML fehlt: {exc}", file=sys.stderr)
    sys.exit(1)

domain = sys.argv[1]
config_file = sys.argv[2]
with open(config_file, "r", encoding="utf-8") as fh:
    payload = yaml.safe_load(fh) or {}
entries = payload.get("wildcard_distribution") or []
for entry in entries:
    if (entry.get("wildcard_domain") or "").lower() != domain:
        continue
    for target in entry.get("targets") or []:
        host = target.get("host", "")
        user = target.get("user", "root")
        remote_dir = target.get("remote_dir", "")
        port = str(target.get("port", 22))
        if host and remote_dir:
            print(f"{host}|{user}|{remote_dir}|{port}")
PY
)

[[ ${#TARGET_LINES[@]} -gt 0 ]] || die "Keine Verteilziele für ${DOMAIN} in ${CONFIG_FILE} gefunden"

for line in "${TARGET_LINES[@]}"; do
  host="${line%%|*}"
  rest="${line#*|}"
  user="${rest%%|*}"
  rest="${rest#*|}"
  remote_dir="${rest%%|*}"
  port="${rest##*|}"

  info "Verteile Wildcard-Zertifikat an ${user}@${host}:${remote_dir}"
  ssh -p "$port" "${user}@${host}" "mkdir -p '$remote_dir'"
  scp -P "$port" "$EXPORT_DIR/fullchain.pem" "${user}@${host}:${remote_dir}/fullchain.pem.tmp"
  scp -P "$port" "$EXPORT_DIR/privkey.pem" "${user}@${host}:${remote_dir}/privkey.pem.tmp"
  ssh -p "$port" "${user}@${host}" \
    "mv '${remote_dir}/fullchain.pem.tmp' '${remote_dir}/fullchain.pem' && \
     mv '${remote_dir}/privkey.pem.tmp' '${remote_dir}/privkey.pem' && \
     chmod 0644 '${remote_dir}/fullchain.pem' && chmod 0600 '${remote_dir}/privkey.pem'"
done

ok "Wildcard-Zertifikat für ${DOMAIN} auf ${#TARGET_LINES[@]} Zielsystem(e) verteilt"
