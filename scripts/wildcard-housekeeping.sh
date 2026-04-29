#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ACME_FILE="/home/andy/infra/data/traefik/acme/acme.json"
EXPORT_SCRIPT="$ROOT_DIR/scripts/wildcard-export.sh"
RESTART_TRAEFIK=0

usage() {
  cat <<'USAGE'
Usage: ./scripts/wildcard-housekeeping.sh <apex-domain> [--restart-traefik]
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
    --restart-traefik)
      RESTART_TRAEFIK=1
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
[[ -x "$EXPORT_SCRIPT" ]] || die "Export-Skript fehlt oder ist nicht ausführbar: $EXPORT_SCRIPT"
[[ -f "$ACME_FILE" ]] || die "ACME-Datei nicht gefunden: $ACME_FILE"

TMP_DIR="$(mktemp -d "/tmp/wildcard-housekeeping-${DOMAIN}.XXXXXX")"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

if ! export_output="$("$EXPORT_SCRIPT" "$DOMAIN" --acme-file="$ACME_FILE" --output-dir="$TMP_DIR" 2>&1)"; then
  if [[ "$export_output" == *"Kein Wildcard-Zertifikat"* ]]; then
    info "Noch kein exportierbares Wildcard-Zertifikat für ${DOMAIN} vorhanden. Housekeeping übersprungen."
    echo "ACME_CHANGED=0"
    exit 0
  fi
  printf '%s\n' "$export_output" >&2
  die "Wildcard-Export für ${DOMAIN} fehlgeschlagen"
fi

CERT_JSON="$(python3 - "$DOMAIN" "$TMP_DIR/fullchain.pem" <<'PY'
import json
import ssl
import sys

domain = sys.argv[1].lower()
wildcard = f"*.{domain}"
certificate = ssl._ssl._test_decode_cert(sys.argv[2])
sans = sorted({value.lower() for key, value in certificate.get("subjectAltName", []) if key == "DNS"})
print(json.dumps({
    "domain": domain,
    "wildcard": wildcard,
    "sans": sans,
    "covers_apex": domain in sans,
    "covers_wildcard": wildcard in sans,
}))
PY
)"

if ! CERT_JSON="$CERT_JSON" python3 - <<'PY'
import json
import os
import sys

payload = json.loads(os.environ["CERT_JSON"])
if not payload.get("covers_wildcard"):
    print(
        f"❌ Exportiertes Zertifikat für {payload['domain']} enthält kein SAN {payload['wildcard']}",
        file=sys.stderr,
    )
    sys.exit(1)
PY
then
  die "Exportiertes Zertifikat für ${DOMAIN} ist kein gültiges Wildcard-Zertifikat"
fi

summary="$(
  CERT_JSON="$CERT_JSON" python3 - "$ACME_FILE" <<'PY'
import json
import os
import sys
from datetime import datetime

acme_file = sys.argv[1]
cert_meta = json.loads(os.environ["CERT_JSON"])
domain = cert_meta["domain"]
wildcard = cert_meta["wildcard"]
covers_apex = bool(cert_meta["covers_apex"])
covers_wildcard = bool(cert_meta["covers_wildcard"])

with open(acme_file, "r", encoding="utf-8") as fh:
    payload = json.load(fh)

removed_entries = []
changed = False

def covers_name(name: str) -> bool:
    name = (name or "").lower()
    if name == domain:
        return covers_apex
    suffix = f".{domain}"
    if not covers_wildcard or not name.endswith(suffix):
        return False
    prefix = name[: -len(suffix)]
    return prefix != "" and "." not in prefix

def prune(node):
    global changed
    if isinstance(node, dict):
        for key, value in node.items():
            if key.lower() == "certificates" and isinstance(value, list):
                kept = []
                for cert in value:
                    dom = cert.get("domain") or cert.get("Domain") or {}
                    main = dom.get("main") or dom.get("Main")
                    sans = dom.get("sans") or dom.get("Sans") or dom.get("SANs") or []
                    names = [name for name in [main, *sans] if name]
                    lower_names = {name.lower() for name in names}
                    if wildcard in lower_names:
                        kept.append(cert)
                        continue
                    if lower_names and all(covers_name(name) for name in lower_names):
                        removed_entries.append(sorted(lower_names))
                        changed = True
                        continue
                    kept.append(cert)
                node[key] = kept
            else:
                prune(value)
    elif isinstance(node, list):
        for item in node:
            prune(item)

prune(payload)

backup_file = ""
if changed:
    timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
    backup_file = f"{acme_file}.bak.{timestamp}"
    with open(acme_file, "r", encoding="utf-8") as fh:
        original = fh.read()
    with open(backup_file, "w", encoding="utf-8") as fh:
        fh.write(original)
    with open(acme_file, "w", encoding="utf-8") as fh:
        json.dump(payload, fh, indent=2)
        fh.write("\n")
    os.chmod(acme_file, 0o600)

print(json.dumps({
    "changed": changed,
    "backup_file": backup_file,
    "removed_entries": removed_entries,
    "cert_sans": cert_meta["sans"],
}))
PY
)"

changed="$(SUMMARY_JSON="$summary" python3 - <<'PY'
import json
import os
summary = json.loads(os.environ["SUMMARY_JSON"])
print("1" if summary.get("changed") else "0")
PY
)"

if [[ "$changed" == "1" ]]; then
  info "Wildcard-Zertifikat für ${DOMAIN} bestätigt"
  info "Zertifikatsnamen: $(SUMMARY_JSON="$summary" python3 - <<'PY'
import json
import os
summary = json.loads(os.environ["SUMMARY_JSON"])
print(", ".join(summary.get("cert_sans", [])))
PY
)"
  info "ACME-Backup erstellt: $(SUMMARY_JSON="$summary" python3 - <<'PY'
import json
import os
summary = json.loads(os.environ["SUMMARY_JSON"])
print(summary.get("backup_file", ""))
PY
)"
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    info "Bereinigt: ${line}"
  done < <(SUMMARY_JSON="$summary" python3 - <<'PY'
import json
import os
summary = json.loads(os.environ["SUMMARY_JSON"])
for entry in summary.get("removed_entries", []):
    print(", ".join(entry))
PY
)
  ok "ACME-Housekeeping für ${DOMAIN} abgeschlossen"
  if [[ "$RESTART_TRAEFIK" -eq 1 ]]; then
    info "Starte Traefik neu"
    docker restart infra-traefik >/dev/null
    ok "Traefik neu gestartet"
  fi
  echo "ACME_CHANGED=1"
else
  info "Für ${DOMAIN} waren keine bereinigbaren Einzelzertifikate mehr vorhanden"
  echo "ACME_CHANGED=0"
fi
