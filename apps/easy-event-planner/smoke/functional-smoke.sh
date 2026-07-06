#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
APP_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

SMOKE_ROOT=${EEP_FUNCTIONAL_SMOKE_ROOT:-/private/tmp/eep-smoke}
DB_PATH=${EEP_FUNCTIONAL_SMOKE_DB_PATH:-${SMOKE_ROOT}/eep.sqlite}
CERT_DIR=${EEP_FUNCTIONAL_SMOKE_CERT_DIR:-${SMOKE_ROOT}/certificates}
PORT=${EEP_SMOKE_PORT:-18081}
ADDR="127.0.0.1:${PORT}"
BASE_URL="http://${ADDR}"
TOKEN_PEPPER=${EEP_TOKEN_PEPPER:-smoke-pepper-please-change}
TENANT_SLUG=${EEP_SEED_TENANT_SLUG:-demo}
TENANT_NAME=${EEP_SEED_TENANT_NAME:-Demo Tenant}
SESSION_COOKIE_NAME=${EEP_SESSION_COOKIE_NAME:-eep_session}
ADMIN_EMAIL=${EEP_FUNCTIONAL_SMOKE_ADMIN_EMAIL:-owner@example.com}
ADMIN_NAME=${EEP_FUNCTIONAL_SMOKE_ADMIN_NAME:-Smoke Owner}
ADMIN_ROLE=${EEP_FUNCTIONAL_SMOKE_ADMIN_ROLE:-owner}
ADMIN_SESSION_TOKEN=${EEP_FUNCTIONAL_SMOKE_ADMIN_SESSION_TOKEN:-smoke-admin-session-token}
LOG_FILE="${SMOKE_ROOT}/functional-smoke-server.log"
START_TIMEOUT_SECONDS=${EEP_FUNCTIONAL_SMOKE_START_TIMEOUT_SECONDS:-30}

die() {
  echo "functional smoke failed: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing command: $1"
}

hash_sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    printf "%s" "$1" | sha256sum | awk '{print $1}'
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    printf "%s" "$1" | shasum -a 256 | awk '{print $1}'
    return 0
  fi
  die "missing command: sha256sum or shasum"
}

cleanup() {
  if [ "${SERVER_PID:-}" != "" ]; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

require_cmd go
require_cmd curl
require_cmd sqlite3
require_cmd python3
require_cmd awk
require_cmd grep
require_cmd sed
require_cmd date
require_cmd tee

mkdir -p "$SMOKE_ROOT" "$CERT_DIR" "$APP_DIR/.gocache" "$APP_DIR/.gomodcache"
cd "$APP_DIR"

echo "== functional smoke: migrate =="
EEP_ENV=development \
EEP_HTTP_ADDR="$ADDR" \
EEP_BASE_URL="$BASE_URL" \
EEP_DB_DRIVER=sqlite \
EEP_DB_PATH="$DB_PATH" \
EEP_CERTIFICATE_STORAGE_DIR="$CERT_DIR" \
EEP_TOKEN_PEPPER="$TOKEN_PEPPER" \
GOCACHE="$APP_DIR/.gocache" \
GOMODCACHE="$APP_DIR/.gomodcache" \
go run ./cmd/migrate >/tmp/eep-functional-migrate.out 2>&1 || {
  cat /tmp/eep-functional-migrate.out
  die "migration command failed"
}
cat /tmp/eep-functional-migrate.out

echo "== functional smoke: seed tenant =="
EEP_ENV=development \
EEP_HTTP_ADDR="$ADDR" \
EEP_BASE_URL="$BASE_URL" \
EEP_DB_DRIVER=sqlite \
EEP_DB_PATH="$DB_PATH" \
EEP_CERTIFICATE_STORAGE_DIR="$CERT_DIR" \
EEP_TOKEN_PEPPER="$TOKEN_PEPPER" \
EEP_SEED_TENANT_SLUG="$TENANT_SLUG" \
EEP_SEED_TENANT_NAME="$TENANT_NAME" \
EEP_SEED_TENANT_PUBLIC_BASE_URL="${BASE_URL}/${TENANT_SLUG}" \
GOCACHE="$APP_DIR/.gocache" \
GOMODCACHE="$APP_DIR/.gomodcache" \
go run ./cmd/seed >/tmp/eep-functional-seed.out 2>&1 || {
  cat /tmp/eep-functional-seed.out
  die "seed command failed"
}
cat /tmp/eep-functional-seed.out

TENANT_ID=$(sqlite3 "$DB_PATH" "SELECT id FROM tenants WHERE slug='${TENANT_SLUG}' LIMIT 1;")
[ -n "$TENANT_ID" ] || die "tenant ${TENANT_SLUG} not found"

NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
SESSION_EXPIRES_AT="2099-01-01T00:00:00Z"
SESSION_HASH=$(hash_sha256 "${TOKEN_PEPPER}:${ADMIN_SESSION_TOKEN}")
USER_ID=$(sqlite3 "$DB_PATH" "SELECT id FROM tenant_users WHERE tenant_id='${TENANT_ID}' AND email='${ADMIN_EMAIL}' LIMIT 1;")
if [ -z "$USER_ID" ]; then
  USER_ID="usr_smoke_owner"
fi

echo "== functional smoke: bootstrap admin session =="
sqlite3 "$DB_PATH" <<SQL
INSERT OR REPLACE INTO tenant_users (
  id, tenant_id, email, name, role, status, created_at, updated_at
) VALUES (
  '${USER_ID}', '${TENANT_ID}', '${ADMIN_EMAIL}', '${ADMIN_NAME}', '${ADMIN_ROLE}', 'active', '${NOW}', '${NOW}'
);
DELETE FROM sessions WHERE session_hash = '${SESSION_HASH}' OR user_id = '${USER_ID}';
INSERT INTO sessions (
  id, tenant_id, user_id, participant_id, session_hash, expires_at, revoked_at, created_at, last_seen_at
) VALUES (
  'ses_smoke_owner', '${TENANT_ID}', '${USER_ID}', NULL, '${SESSION_HASH}', '${SESSION_EXPIRES_AT}', NULL, '${NOW}', '${NOW}'
);
SQL

echo "== functional smoke: start server =="
EEP_ENV=development \
EEP_HTTP_ADDR="$ADDR" \
EEP_BASE_URL="$BASE_URL" \
EEP_DB_DRIVER=sqlite \
EEP_DB_PATH="$DB_PATH" \
EEP_CERTIFICATE_STORAGE_DIR="$CERT_DIR" \
EEP_TOKEN_PEPPER="$TOKEN_PEPPER" \
GOCACHE="$APP_DIR/.gocache" \
GOMODCACHE="$APP_DIR/.gomodcache" \
go run ./cmd/server >"$LOG_FILE" 2>&1 &
SERVER_PID=$!

attempt=0
while :; do
  if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo "server logs:" >&2
    tail -n 200 "$LOG_FILE" >&2 || true
    die "server process exited before becoming ready (possible port conflict)"
  fi
  if curl -fsS "$BASE_URL/readyz" >/dev/null 2>&1; then
    break
  fi
  attempt=$((attempt + 1))
  if [ "$attempt" -ge "$START_TIMEOUT_SECONDS" ]; then
    echo "server logs:" >&2
    tail -n 200 "$LOG_FILE" >&2 || true
    die "server did not become ready"
  fi
  sleep 1
done

if ! kill -0 "$SERVER_PID" 2>/dev/null; then
  echo "server logs:" >&2
  tail -n 200 "$LOG_FILE" >&2 || true
  die "server process exited after ready probe (possible port conflict)"
fi

echo "== functional smoke: system endpoints =="
curl -fsS "$BASE_URL/healthz"
curl -fsS "$BASE_URL/readyz"
curl -fsS "$BASE_URL/version"

echo "== functional smoke: auth/me =="
AUTH_ME_CODE=$(curl -sS -o /tmp/eep-functional-auth-me.json -w "%{http_code}" "$BASE_URL/api/v1/auth/me" -H "Cookie: ${SESSION_COOKIE_NAME}=${ADMIN_SESSION_TOKEN}")
cat /tmp/eep-functional-auth-me.json
echo "auth/me status code: ${AUTH_ME_CODE}"
[ "$AUTH_ME_CODE" = "200" ] || die "auth/me returned HTTP ${AUTH_ME_CODE}"
python3 -c 'import json,sys; data=json.load(open("/tmp/eep-functional-auth-me.json")); assert data.get("authenticated") is True'

echo "== functional smoke: tenant embed settings =="
TENANT_SETTINGS_RESP=$(curl -fsS -X PATCH "$BASE_URL/api/v1/admin/tenant/settings" \
  -H "Content-Type: application/json" \
  -H "Cookie: ${SESSION_COOKIE_NAME}=${ADMIN_SESSION_TOKEN}" \
  -d '{"app_settings":{"allowed_embed_origins":["https://ghost.geller.men"],"event_detail_base_url":"https://www.example.com/events"}}')
printf "%s\n" "$TENANT_SETTINGS_RESP"
printf "%s" "$TENANT_SETTINGS_RESP" | python3 -c 'import json,sys; data=json.load(sys.stdin); app=data["item"]["app_settings"]; assert "https://ghost.geller.men" in app["allowed_embed_origins"]'

EVENT_SLUG="smoke-event-$(date +%s)"
STARTS_AT=$(python3 - <<'PY'
from datetime import datetime, timedelta, timezone
ts = datetime.now(timezone.utc) + timedelta(days=2)
print(ts.replace(microsecond=0).isoformat().replace("+00:00", "Z"))
PY
)

echo "== functional smoke: create event =="
CREATE_RESP=$(curl -fsS -X POST "$BASE_URL/api/v1/admin/events" \
  -H "Content-Type: application/json" \
  -H "Cookie: ${SESSION_COOKIE_NAME}=${ADMIN_SESSION_TOKEN}" \
  -d "{\"slug\":\"${EVENT_SLUG}\",\"title\":\"Smoke Event\",\"starts_at\":\"${STARTS_AT}\",\"timezone\":\"Europe/Berlin\",\"is_public\":true,\"registration_enabled\":true,\"waitlist_enabled\":true}")
printf "%s\n" "$CREATE_RESP"
EVENT_ID=$(printf "%s" "$CREATE_RESP" | python3 -c 'import json,sys; print(json.load(sys.stdin)["item"]["id"])')
[ -n "$EVENT_ID" ] || die "event id missing"

echo "== functional smoke: publish event =="
PUBLISH_RESP=$(curl -fsS -X POST "$BASE_URL/api/v1/admin/events/${EVENT_ID}/publish" \
  -H "Cookie: ${SESSION_COOKIE_NAME}=${ADMIN_SESSION_TOKEN}")
printf "%s\n" "$PUBLISH_RESP"

echo "== functional smoke: fetch event registration embed code =="
EVENT_EMBED_RESP=$(curl -fsS "$BASE_URL/api/v1/admin/events/${EVENT_ID}/embed-code" \
  -H "Cookie: ${SESSION_COOKIE_NAME}=${ADMIN_SESSION_TOKEN}")
printf "%s\n" "$EVENT_EMBED_RESP"
printf "%s" "$EVENT_EMBED_RESP" | TENANT_SLUG_EXPECTED="$TENANT_SLUG" EVENT_SLUG_EXPECTED="$EVENT_SLUG" \
  python3 -c 'import json,os,sys; data=json.load(sys.stdin); code=data.get("embed_code",""); expected="/" + os.environ["TENANT_SLUG_EXPECTED"] + "/register.js?event=" + os.environ["EVENT_SLUG_EXPECTED"]; assert expected in code; assert data.get("kind") == "registration_form"'

echo "== functional smoke: list public events =="
PUBLIC_EVENTS_RESP=$(curl -fsS "$BASE_URL/api/v1/public/${TENANT_SLUG}/events")
printf "%s\n" "$PUBLIC_EVENTS_RESP"
printf "%s" "$PUBLIC_EVENTS_RESP" | python3 -c 'import json,sys; data=json.load(sys.stdin); assert int(data.get("total",0)) >= 1'

SNIPPET_SLUG="smoke-snippet-$(date +%s)"

echo "== functional smoke: create snippet =="
SNIPPET_CREATE_RESP=$(curl -fsS -X POST "$BASE_URL/api/v1/admin/snippets" \
  -H "Content-Type: application/json" \
  -H "Cookie: ${SESSION_COOKIE_NAME}=${ADMIN_SESSION_TOKEN}" \
  -d "{\"name\":\"Smoke Snippet\",\"slug\":\"${SNIPPET_SLUG}\",\"view_type\":\"cards\",\"event_filter\":{\"events\":\"upcoming\",\"limit\":3},\"display_options\":{\"theme\":\"light\",\"register\":true},\"is_active\":true}")
printf "%s\n" "$SNIPPET_CREATE_RESP"
SNIPPET_ID=$(printf "%s" "$SNIPPET_CREATE_RESP" | python3 -c 'import json,sys; print(json.load(sys.stdin)["item"]["id"])')
[ -n "$SNIPPET_ID" ] || die "snippet id missing"

echo "== functional smoke: fetch snippet embed code =="
SNIPPET_EMBED_RESP=$(curl -fsS "$BASE_URL/api/v1/admin/snippets/${SNIPPET_ID}/embed-code" \
  -H "Cookie: ${SESSION_COOKIE_NAME}=${ADMIN_SESSION_TOKEN}")
printf "%s\n" "$SNIPPET_EMBED_RESP"
printf "%s" "$SNIPPET_EMBED_RESP" | TENANT_SLUG_EXPECTED="$TENANT_SLUG" SNIPPET_SLUG_EXPECTED="$SNIPPET_SLUG" \
  python3 -c 'import json,os,sys; data=json.load(sys.stdin); code=data.get("embed_code",""); expected="/" + os.environ["TENANT_SLUG_EXPECTED"] + "/include.js?config=" + os.environ["SNIPPET_SLUG_EXPECTED"]; assert expected in code'

echo "== functional smoke: snippet include.js =="
SNIPPET_INCLUDE_RESP=$(curl -fsS "$BASE_URL/${TENANT_SLUG}/include.js?config=${SNIPPET_SLUG}")
printf "%s\n" "$SNIPPET_INCLUDE_RESP" | sed -n '1,8p'
printf "%s" "$SNIPPET_INCLUDE_RESP" | TENANT_SLUG_EXPECTED="$TENANT_SLUG" \
  python3 -c 'import os,sys; body=sys.stdin.read(); assert "/api/v1/public/" in body; assert "/snippet/events" in body; assert os.environ["TENANT_SLUG_EXPECTED"] in body'

echo "== functional smoke: registration register.js =="
REGISTER_JS_RESP=$(curl -fsS "$BASE_URL/${TENANT_SLUG}/register.js?event=${EVENT_SLUG}")
printf "%s\n" "$REGISTER_JS_RESP" | sed -n '1,8p'
printf "%s" "$REGISTER_JS_RESP" | python3 -c 'import sys; body=sys.stdin.read(); assert "/api/v1/public/" in body; assert "/registrations/start" in body; assert "Magic Link anfordern" in body'

echo "== functional smoke: CORS preflight =="
CORS_CODE=$(curl -sS -o /tmp/eep-functional-cors.json -w "%{http_code}" -X OPTIONS "$BASE_URL/api/v1/public/${TENANT_SLUG}/registrations/start" \
  -H "Origin: https://ghost.geller.men" \
  -H "Access-Control-Request-Method: POST")
[ "$CORS_CODE" = "204" ] || die "CORS preflight returned HTTP ${CORS_CODE}"
ALLOW_ORIGIN=$(curl -sSI -X OPTIONS "$BASE_URL/api/v1/public/${TENANT_SLUG}/registrations/start" \
  -H "Origin: https://ghost.geller.men" \
  -H "Access-Control-Request-Method: POST" | tr -d '\r' | awk -F': ' 'BEGIN{IGNORECASE=1} $1=="Access-Control-Allow-Origin"{print $2; exit}')
[ "$ALLOW_ORIGIN" = "https://ghost.geller.men" ] || die "unexpected CORS allow origin: ${ALLOW_ORIGIN}"

echo "== functional smoke: snippet public events =="
SNIPPET_EVENTS_RESP=$(curl -fsS "$BASE_URL/api/v1/public/${TENANT_SLUG}/snippet/events?config=${SNIPPET_SLUG}")
printf "%s\n" "$SNIPPET_EVENTS_RESP"
printf "%s" "$SNIPPET_EVENTS_RESP" | python3 -c 'import json,sys; data=json.load(sys.stdin); assert int(data.get("total",0)) >= 1; assert data.get("view") == "cards"'

REG_EMAIL="max+$(date +%s)@example.com"

echo "== functional smoke: start registration =="
REG_START_RESP=$(curl -fsS -X POST "$BASE_URL/api/v1/public/${TENANT_SLUG}/registrations/start" \
  -H "Content-Type: application/json" \
  -d "{\"event_id\":\"${EVENT_ID}\",\"name\":\"Max Mustermann\",\"email\":\"${REG_EMAIL}\",\"participation_type\":\"onsite\",\"privacy_accepted\":true}")
printf "%s\n" "$REG_START_RESP"
REGISTRATION_ID=$(printf "%s" "$REG_START_RESP" | python3 -c 'import json,sys; data=json.load(sys.stdin); assert data.get("status")=="verification_pending"; print(data["registration_id"])')
[ -n "$REGISTRATION_ID" ] || die "registration id missing"

BODY_TEXT=$(sqlite3 "$DB_PATH" "SELECT body_text FROM email_jobs WHERE template_key='registration_verify' ORDER BY created_at DESC LIMIT 1;")
VERIFY_URL=$(printf "%s\n" "$BODY_TEXT" | grep -Eo 'https?://[^[:space:]]+' | tail -n 1)
[ -n "$VERIFY_URL" ] || die "verification URL not found in email_jobs body"

echo "== functional smoke: verify registration =="
HTTP_CODE=$(curl -sS -o /tmp/eep-functional-verify.json -w "%{http_code}" "$VERIFY_URL")
cat /tmp/eep-functional-verify.json
echo "verify status code: ${HTTP_CODE}"
[ "$HTTP_CODE" = "200" ] || die "verify endpoint returned HTTP ${HTTP_CODE}"
VERIFY_STATUS=$(python3 -c 'import json; import sys; data=json.load(open("/tmp/eep-functional-verify.json")); print(data.get("status",""))')
[ "$VERIFY_STATUS" = "confirmed" ] || [ "$VERIFY_STATUS" = "waitlist" ] || die "unexpected verify status: ${VERIFY_STATUS}"

echo "== functional smoke: admin dashboard =="
DASHBOARD_RESP=$(curl -fsS "$BASE_URL/api/v1/admin/dashboard" -H "Cookie: ${SESSION_COOKIE_NAME}=${ADMIN_SESSION_TOKEN}")
printf "%s\n" "$DASHBOARD_RESP"
printf "%s" "$DASHBOARD_RESP" | python3 -c 'import json,sys; json.load(sys.stdin)'

echo "functional smoke ok"
