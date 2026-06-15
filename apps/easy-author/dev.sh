#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_URL="http://127.0.0.1:8086/api/health"

cleanup() {
  if [[ -n "${BACKEND_PID:-}" ]]; then
    kill "$BACKEND_PID" >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT INT TERM

echo "→ easy-author Backend startet ..."
(
  cd "$BACKEND_DIR"
  export GOCACHE="$BACKEND_DIR/.gocache"
  export GOMODCACHE="$BACKEND_DIR/.gomodcache"
  mkdir -p "$GOCACHE" "$GOMODCACHE"
  go run ./cmd/server
) &
BACKEND_PID=$!

echo "→ Warte auf Backend unter $BACKEND_URL"
for _ in {1..40}; do
  if curl -fsS "$BACKEND_URL" >/dev/null 2>&1; then
    echo "✓ Backend ist bereit"
    break
  fi
  sleep 0.5
done

if ! curl -fsS "$BACKEND_URL" >/dev/null 2>&1; then
  echo "✗ Backend konnte nicht gestartet werden."
  exit 1
fi

echo "→ easy-author Frontend startet ..."
cd "$FRONTEND_DIR"

if [[ -s "$HOME/.nvm/nvm.sh" ]]; then
  # shellcheck disable=SC1090
  . "$HOME/.nvm/nvm.sh"
fi

if command -v nvm >/dev/null 2>&1; then
  nvm use >/dev/null
fi

npm install >/dev/null
exec npm run dev
