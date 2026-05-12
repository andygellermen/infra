#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
APP_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
PORT=${EEP_SMOKE_PORT:-18080}
ADDR="127.0.0.1:${PORT}"
BASE_URL="http://${ADDR}"
LOG_FILE="${APP_DIR}/smoke/smoke.log"

cleanup() {
  if [ "${SERVER_PID:-}" != "" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

cd "$APP_DIR"
mkdir -p .gocache .gomodcache

EEP_HTTP_ADDR="$ADDR" \
EEP_BASE_URL="$BASE_URL" \
GOCACHE="$APP_DIR/.gocache" \
GOMODCACHE="$APP_DIR/.gomodcache" \
go run ./cmd/server >"$LOG_FILE" 2>&1 &
SERVER_PID=$!

attempt=0
until curl -fsS "$BASE_URL/healthz" >/dev/null 2>&1; do
  attempt=$((attempt + 1))
  if [ "$attempt" -ge 30 ]; then
    echo "smoke failed: server did not become healthy"
    exit 1
  fi
  sleep 1
done

curl -fsS "$BASE_URL/healthz"
curl -fsS "$BASE_URL/readyz"
curl -fsS "$BASE_URL/version"

echo "smoke ok"
