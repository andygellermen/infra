#!/usr/bin/env bash
set -euo pipefail

ACME_FILE="/home/andy/infra/data/traefik/acme/acme.json"

usage() {
  cat <<'USAGE'
Usage: ./scripts/wildcard-export.sh <apex-domain> [--output-dir /tmp/wildcard-export]
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

normalize_domain() {
  local domain="$1"
  if [[ "$domain" =~ ^[A-Za-z0-9.-]+$ ]]; then
    printf '%s\n' "$domain" | tr '[:upper:]' '[:lower:]'
  else
    require_cmd idn
    idn --quiet --uts46 "$domain" | tr '[:upper:]' '[:lower:]'
  fi
}

OUTPUT_DIR=""
ARGS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --acme-file=*)
      ACME_FILE="${1#*=}"
      shift
      ;;
    --acme-file)
      shift
      [[ $# -gt 0 ]] || die "Fehlender Wert für --acme-file"
      ACME_FILE="$1"
      shift
      ;;
    --output-dir=*)
      OUTPUT_DIR="${1#*=}"
      shift
      ;;
    --output-dir)
      shift
      [[ $# -gt 0 ]] || die "Fehlender Wert für --output-dir"
      OUTPUT_DIR="$1"
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      ARGS+=("$1")
      shift
      ;;
  esac
done

[[ ${#ARGS[@]} -eq 1 ]] || { usage; exit 1; }
require_cmd python3

DOMAIN="$(normalize_domain "${ARGS[0]}")"
OUTPUT_DIR="${OUTPUT_DIR:-$(mktemp -d "/tmp/wildcard-${DOMAIN}.XXXXXX")}"
mkdir -p "$OUTPUT_DIR"

[[ -f "$ACME_FILE" ]] || die "ACME-Datei nicht gefunden: $ACME_FILE"

python3 - "$DOMAIN" "$ACME_FILE" "$OUTPUT_DIR" <<'PY'
import base64
import json
import os
import sys

domain = sys.argv[1]
acme_file = sys.argv[2]
output_dir = sys.argv[3]
wildcard = f"*.{domain}"

with open(acme_file, "r", encoding="utf-8") as fh:
    payload = json.load(fh)

matches = []

def walk(node):
    if isinstance(node, dict):
        for key, value in node.items():
            if key.lower() == "certificates" and isinstance(value, list):
                for cert in value:
                    dom = cert.get("domain") or cert.get("Domain") or {}
                    main = dom.get("main") or dom.get("Main")
                    sans = dom.get("sans") or dom.get("Sans") or dom.get("SANs") or []
                    names = {name for name in [main, *sans] if name}
                    if wildcard in names:
                        matches.append(cert)
            else:
                walk(value)
    elif isinstance(node, list):
        for item in node:
            walk(item)

walk(payload)

if not matches:
    print(f"❌ Kein Wildcard-Zertifikat für {domain} in {acme_file} gefunden", file=sys.stderr)
    sys.exit(1)

cert = matches[0]
certificate = cert.get("certificate") or cert.get("Certificate")
key = cert.get("key") or cert.get("Key")
if not certificate or not key:
    print(f"❌ Zertifikatsdaten für {domain} unvollständig", file=sys.stderr)
    sys.exit(1)

fullchain_pem = base64.b64decode(certificate)
privkey_pem = base64.b64decode(key)

with open(os.path.join(output_dir, "fullchain.pem"), "wb") as fh:
    fh.write(fullchain_pem)
with open(os.path.join(output_dir, "privkey.pem"), "wb") as fh:
    fh.write(privkey_pem)
PY

ok "Wildcard-Zertifikat exportiert nach ${OUTPUT_DIR}"
info "Dateien: ${OUTPUT_DIR}/fullchain.pem und ${OUTPUT_DIR}/privkey.pem"
