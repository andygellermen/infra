#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/eep-domain-bindings-sync.sh <domain> [--base-url=<url>] [--token=<token>] [--output=<file>] [--service=<traefik-service>] [--certresolver=<name>] [--refresh-after-render=<true|false>] [--insecure=<true|false>]

Description:
  Liest die internen EEP-Domain-Bindings, rendert daraus eine Traefik-File-Config
  fuer kundenspezifische Public-Domains und stoesst optional direkt einen DNS/TLS-Refresh
  in der App an.
USAGE
}

die(){ echo "ERROR: $*" >&2; exit 1; }
info(){ echo "INFO: $*"; }
ok(){ echo "OK: $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "missing command: $1"; }

DOMAIN=""
BASE_URL="${EEP_EDGE_SYNC_BASE_URL:-}"
TOKEN="${EEP_EDGE_SYNC_TOKEN:-}"
OUTPUT_FILE="${EEP_EDGE_SYNC_OUTPUT_FILE:-}"
DOCKER_SERVICE="${EEP_EDGE_SYNC_DOCKER_SERVICE:-}"
CERTRESOLVER="${EEP_EDGE_SYNC_CERTRESOLVER:-letsEncrypt}"
REFRESH_AFTER_RENDER="${EEP_EDGE_SYNC_REFRESH_AFTER_RENDER:-true}"
INSECURE="${EEP_EDGE_SYNC_INSECURE:-false}"
TIMEOUT_SECONDS="${EEP_EDGE_SYNC_TIMEOUT_SECONDS:-20}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --base-url=*)
      BASE_URL="${1#*=}"
      shift
      ;;
    --token=*)
      TOKEN="${1#*=}"
      shift
      ;;
    --output=*)
      OUTPUT_FILE="${1#*=}"
      shift
      ;;
    --service=*)
      DOCKER_SERVICE="${1#*=}"
      shift
      ;;
    --certresolver=*)
      CERTRESOLVER="${1#*=}"
      shift
      ;;
    --refresh-after-render=*)
      REFRESH_AFTER_RENDER="${1#*=}"
      shift
      ;;
    --insecure=*)
      INSECURE="${1#*=}"
      shift
      ;;
    --timeout=*)
      TIMEOUT_SECONDS="${1#*=}"
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    --*)
      die "unknown option: $1"
      ;;
    *)
      [[ -z "$DOMAIN" ]] || die "only one domain is allowed"
      DOMAIN="$1"
      shift
      ;;
  esac
done

[[ -n "$DOMAIN" ]] || { usage; exit 1; }
[[ "$REFRESH_AFTER_RENDER" =~ ^(true|false)$ ]] || die "--refresh-after-render must be true or false"
[[ "$INSECURE" =~ ^(true|false)$ ]] || die "--insecure must be true or false"
[[ "$TIMEOUT_SECONDS" =~ ^[0-9]+$ ]] || die "--timeout must be a non-negative integer"

require_cmd curl
require_cmd mktemp
require_cmd python3
require_cmd install
require_cmd dirname

if [[ -z "$BASE_URL" ]]; then
  BASE_URL="https://${DOMAIN}"
fi
if [[ -z "$TOKEN" ]]; then
  die "missing EEP edge sync token"
fi
if [[ -z "$OUTPUT_FILE" ]]; then
  OUTPUT_FILE="/home/andy/infra/data/traefik/eep-domain-bindings-${DOMAIN//./-}.yml"
fi
if [[ -z "$DOCKER_SERVICE" ]]; then
  DOCKER_SERVICE="easy-event-planner-${DOMAIN//./-}@docker"
fi

TMP_JSON="$(mktemp)"
TMP_YAML="$(mktemp)"
cleanup() {
  rm -f "$TMP_JSON" "$TMP_YAML"
}
trap cleanup EXIT

CURL_ARGS=(
  --silent
  --show-error
  --fail
  --location
  --connect-timeout "$TIMEOUT_SECONDS"
  --max-time "$TIMEOUT_SECONDS"
  --header "Authorization: Bearer ${TOKEN}"
)
if [[ "$INSECURE" == "true" ]]; then
  CURL_ARGS+=(--insecure)
fi

EXPORT_URL="${BASE_URL%/}/api/v1/internal/infra/domain-bindings/export"
REFRESH_URL="${BASE_URL%/}/api/v1/internal/infra/domain-bindings/refresh"

info "Fetch domain binding export from ${EXPORT_URL}"
curl "${CURL_ARGS[@]}" --output "$TMP_JSON" "$EXPORT_URL"

python3 - "$TMP_JSON" "$TMP_YAML" "$DOCKER_SERVICE" "$CERTRESOLVER" <<'PY'
import json
import re
import sys

src_path, out_path, service_name, certresolver = sys.argv[1:]
with open(src_path, "r", encoding="utf-8") as fh:
    payload = json.load(fh)

eligible = []
for item in payload.get("items", []):
    if not item.get("edge_enabled"):
        continue
    domain = str(item.get("domain", "")).strip().lower()
    if not domain:
        continue
    base_path = str(item.get("base_path", "/")).strip() or "/"
    if not base_path.startswith("/"):
        base_path = "/" + base_path
    eligible.append({"domain": domain, "base_path": base_path})


def sanitize(text: str) -> str:
    cleaned = re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")
    return cleaned or "route"


lines = [
    "http:",
    "  middlewares:",
    "    eep-edge-https-redirect:",
    "      redirectScheme:",
    "        scheme: https",
    "        permanent: true",
]

if not eligible:
    lines.append("  routers: {}")
else:
    lines.append("  routers:")
    seen = set()
    for index, item in enumerate(eligible, start=1):
        suffix = sanitize(f"{item['domain']}-{item['base_path']}")
        if suffix in seen:
            suffix = f"{suffix}-{index}"
        seen.add(suffix)

        rule = f"Host(`{item['domain']}`)"
        priority = 200
        if item["base_path"] != "/":
            rule += f" && PathPrefix(`{item['base_path']}`)"
            priority = 260

        https_router = f"eep-edge-{suffix}"
        http_router = f"{https_router}-web"
        lines.extend(
            [
                f"    {https_router}:",
                "      entryPoints:",
                "        - websecure",
                f"      rule: \"{rule}\"",
                f"      priority: {priority}",
                "      tls:",
                f"        certResolver: {certresolver}",
                f"      service: {service_name}",
                f"    {http_router}:",
                "      entryPoints:",
                "        - web",
                f"      rule: \"{rule}\"",
                f"      priority: {priority}",
                "      middlewares:",
                "        - eep-edge-https-redirect",
                "      service: noop@internal",
            ]
        )

with open(out_path, "w", encoding="utf-8") as fh:
    fh.write("\n".join(lines) + "\n")
PY

install -d "$(dirname "$OUTPUT_FILE")"
install -m 0644 "$TMP_YAML" "$OUTPUT_FILE"
ok "Rendered Traefik domain binding config to ${OUTPUT_FILE}"

if [[ "$REFRESH_AFTER_RENDER" == "true" ]]; then
  info "Trigger domain binding refresh at ${REFRESH_URL}"
  curl "${CURL_ARGS[@]}" --output /dev/null --request POST "$REFRESH_URL"
  ok "Triggered EEP domain binding refresh"
fi
