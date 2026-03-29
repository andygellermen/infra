#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INVENTORY="$ROOT_DIR/ansible/inventory/hosts.ini"
PLAYBOOK="$ROOT_DIR/ansible/playbooks/deploy-redirects.yml"
REDIRECTS_FILE="$ROOT_DIR/ansible/redirects/redirects.yml"

usage(){ echo "Usage: $0 [--check-only]"; }
die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

run_playbook_quiet() {
  local success_message="$1"
  shift
  local log_file
  log_file="$(mktemp)"

  if ANSIBLE_DISPLAY_SKIPPED_HOSTS=false "$@" >"$log_file" 2>&1; then
    local recap_line
    recap_line="$(grep -E '^localhost[[:space:]]*:' "$log_file" | tail -n1 || true)"
    [[ -n "$recap_line" ]] && info "$recap_line"
    ok "$success_message"
    rm -f "$log_file"
    return 0
  fi

  cat "$log_file" >&2
  rm -f "$log_file"
  return 1
}

CHECK_ONLY=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

require_cmd curl
require_cmd dig
require_cmd python3
require_cmd ansible-playbook

[[ -f "$REDIRECTS_FILE" ]] || die "Redirect-Konfiguration fehlt: $REDIRECTS_FILE"

REDIRECT_JSON="$(python3 - "$REDIRECTS_FILE" <<'PY'
import json
import sys
try:
    import yaml
except Exception as exc:
    print(f"❌ PyYAML fehlt: {exc}", file=sys.stderr)
    sys.exit(1)

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    payload = yaml.safe_load(fh) or {}
entries = payload.get("redirects") or []
if not isinstance(entries, list):
    print("❌ redirects muss eine Liste sein", file=sys.stderr)
    sys.exit(1)
print(json.dumps(entries))
PY
)"

HOST_IP="$(curl -fsSL https://api.ipify.org || true)"
[[ -n "$HOST_IP" ]] || die "Öffentliche Host-IP konnte nicht ermittelt werden"

mapfile -t REDIRECT_LINES < <(python3 - "$REDIRECT_JSON" <<'PY'
import json
import sys

entries = json.loads(sys.argv[1])
for item in entries:
    source = item.get("source", "")
    target = item.get("target", "")
    permanent = "true" if item.get("permanent", True) else "false"
    scheme = item.get("target_scheme", "https")
    if source and target:
        print(f"{source}|{target}|{scheme}|{permanent}")
PY
)

for line in "${REDIRECT_LINES[@]}"; do
  source="${line%%|*}"
  dns_ip="$(dig +short A "$source" | head -n1)"
  [[ -n "$dns_ip" ]] || die "DNS/IP Prüfung fehlgeschlagen für Redirect-Quelle: $source"
  [[ "$dns_ip" == "$HOST_IP" ]] || die "DNS mismatch: $source -> $dns_ip (erwartet $HOST_IP)"
done
ok "DNS OK für ${#REDIRECT_LINES[@]} Redirect-Domain(s)"

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only abgeschlossen"
  exit 0
fi

run_playbook_quiet "Redirect-Redeploy abgeschlossen" \
  ansible-playbook -i "$INVENTORY" "$PLAYBOOK" \
  || die "Redirect-Redeploy fehlgeschlagen"

python3 - "$REDIRECT_JSON" <<'PY'
import json
import time
import subprocess
import sys

entries = json.loads(sys.argv[1])

def curl_meta(url: str):
    cmd = [
        "curl", "-k", "-sS", "-o", "/dev/null",
        "-w", "%{http_code} %{redirect_url}",
        "--max-redirs", "0",
        url,
    ]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        return "000", ""
    out = result.stdout.strip().split(" ", 1)
    status = out[0] if out else "000"
    redirect_url = out[1] if len(out) > 1 else ""
    return status, redirect_url

failures = []
verified = 0
for item in entries:
    source = item.get("source", "")
    target = item.get("target", "")
    if not source or not target:
        continue
    scheme = item.get("target_scheme", "https")
    permanent = item.get("permanent", True)
    expected_prefix = f"{scheme}://{target}/"
    expected_statuses = {"301", "308"} if permanent else {"302", "307"}
    status = "000"
    redirect_url = ""
    for _ in range(6):
        status, redirect_url = curl_meta(f"https://{source}/")
        if status in expected_statuses and redirect_url.startswith(expected_prefix):
            break
        time.sleep(2)
    if status not in expected_statuses:
        failures.append(f"Redirect für {source} liefert Status {status} statt {'/'.join(sorted(expected_statuses))}")
        continue
    if not redirect_url.startswith(expected_prefix):
        failures.append(f"Redirect-Ziel für {source} ist unerwartet: {redirect_url} (erwartet Präfix {expected_prefix})")
        continue
    verified += 1

if failures:
    for message in failures:
        print(f"❌ {message}", file=sys.stderr)
    sys.exit(1)

print(f"✅ Redirect-Selbsttest erfolgreich ({verified} Domain{'s' if verified != 1 else ''})")
PY
