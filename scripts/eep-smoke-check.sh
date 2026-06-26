#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/eep-smoke-check.sh <domain> [--base-url=<url>] [--retries=<n>] [--delay=<seconds>] [--timeout=<seconds>] [--insecure=<true|false>]

Checks:
  - GET /healthz   -> HTTP 200, body contains "ok"
  - GET /readyz    -> HTTP 200, body contains "ready"
  - GET /version   -> HTTP 200
USAGE
}

die(){ echo "ERROR: $*" >&2; exit 1; }
info(){ echo "INFO: $*"; }
ok(){ echo "OK: $*"; }

DOMAIN=""
BASE_URL=""
RETRIES=15
DELAY_SECONDS=2
TIMEOUT_SECONDS=10
INSECURE="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --base-url=*)
      BASE_URL="${1#*=}"
      shift
      ;;
    --retries=*)
      RETRIES="${1#*=}"
      shift
      ;;
    --delay=*)
      DELAY_SECONDS="${1#*=}"
      shift
      ;;
    --timeout=*)
      TIMEOUT_SECONDS="${1#*=}"
      shift
      ;;
    --insecure=*)
      INSECURE="${1#*=}"
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
[[ "$RETRIES" =~ ^[0-9]+$ ]] || die "--retries must be a non-negative integer"
[[ "$DELAY_SECONDS" =~ ^[0-9]+$ ]] || die "--delay must be a non-negative integer"
[[ "$TIMEOUT_SECONDS" =~ ^[0-9]+$ ]] || die "--timeout must be a non-negative integer"
[[ "$INSECURE" =~ ^(true|false)$ ]] || die "--insecure must be true or false"

if [[ -z "$BASE_URL" ]]; then
  BASE_URL="https://${DOMAIN}"
fi

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

require_cmd curl
require_cmd mktemp
require_cmd sed
require_cmd grep

CURL_ARGS=(
  --silent
  --show-error
  --location
  --max-redirs 10
  --connect-timeout "$TIMEOUT_SECONDS"
  --max-time "$TIMEOUT_SECONDS"
)

if [[ "$INSECURE" == "true" ]]; then
  CURL_ARGS+=(--insecure)
fi

check_endpoint() {
  local label="$1"
  local path="$2"
  local expected_body="${3:-}"
  local attempt=1
  local body_file
  local status_file
  local status_code=""

  body_file="$(mktemp)"
  status_file="$(mktemp)"
  trap 'rm -f "$body_file" "$status_file"' RETURN

  while (( attempt <= RETRIES )); do
    status_code="$(
      curl "${CURL_ARGS[@]}" \
        --output "$body_file" \
        --write-out '%{http_code}' \
        "${BASE_URL}${path}" >"$status_file" 2>&1 \
        && cat "$status_file" \
        || printf '000'
    )"

    if [[ "$status_code" == "200" ]]; then
      if [[ -z "$expected_body" ]] || grep -Fqi "$expected_body" "$body_file"; then
        ok "${label} -> HTTP 200"
        rm -f "$body_file" "$status_file"
        trap - RETURN
        return 0
      fi
    fi

    if (( attempt < RETRIES )); then
      sleep "$DELAY_SECONDS"
    fi
    attempt=$((attempt + 1))
  done

  echo "ERROR: ${label} failed for ${BASE_URL}${path} (last status ${status_code})" >&2
  echo "--- response body ---" >&2
  sed -n '1,20p' "$body_file" >&2 || true
  rm -f "$body_file" "$status_file"
  trap - RETURN
  return 1
}

info "Smoke check for ${BASE_URL}"
check_endpoint "healthz" "/healthz" "ok"
check_endpoint "readyz" "/readyz" "ready"
check_endpoint "version" "/version"
ok "Easy-Event-Planner smoke check succeeded for ${DOMAIN}"
