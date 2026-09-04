#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"

usage() {
  cat <<'USAGE'
Usage: ./scripts/sal-rotate-secrets.sh <domain>

Rotates both protected-MVP secrets without printing them: the PostgreSQL role
password and the bcrypt Basic-Auth hash. Run this on the Infra server, then run
sal-redeploy.sh immediately.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

DOMAIN=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --help|-h) usage; exit 0 ;;
    *)
      [[ -z "$DOMAIN" ]] || die "Nur eine Domain ist erlaubt."
      DOMAIN="$1"
      shift
      ;;
  esac
done

[[ "$DOMAIN" =~ ^[a-z0-9.-]+$ ]] || die "Ungueltige oder fehlende Domain."
require_cmd awk
require_cmd docker
require_cmd htpasswd
require_cmd openssl
require_cmd python3

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
POSTGRES_CONTAINER="sprach-a-lyzer-${DOMAIN//./-}-postgres"
[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars fehlen: $HOSTVARS_FILE"
docker inspect "$POSTGRES_CONTAINER" >/dev/null 2>&1 || die "PostgreSQL-Container fehlt: $POSTGRES_CONTAINER"

OLD_POSTGRES_PASSWORD="$(awk -F ': *' '/^sal_postgres_password:/ {gsub(/^"|"$/, "", $2); print $2; exit}' "$HOSTVARS_FILE")"
[[ "$OLD_POSTGRES_PASSWORD" =~ ^[a-f0-9]{64}$ ]] || die "Bestehendes PostgreSQL-Kennwort ist ungueltig."

AUTH_USERNAME="$(awk -F ': *' '/^sal_basic_auth_username:/ {gsub(/^"|"$/, "", $2); print $2; exit}' "$HOSTVARS_FILE")"
[[ "$AUTH_USERNAME" =~ ^[A-Za-z0-9._-]+$ ]] || die "Basic-Auth-Benutzername fehlt in den Hostvars."

info "Bitte jetzt das neue Zugriffskennwort fuer ${AUTH_USERNAME} zweimal eingeben."
AUTH_LINE="$(htpasswd -nB -C 12 "$AUTH_USERNAME")"
NEW_AUTH_HASH="${AUTH_LINE#*:}"
NEW_POSTGRES_PASSWORD="$(openssl rand -hex 32)"
DB_CHANGED=0

alter_database_password() {
  local password="$1"
  printf "ALTER ROLE sprachalyzer WITH PASSWORD '%s';\n" "$password" \
    | docker exec -i "$POSTGRES_CONTAINER" \
        psql -X -v ON_ERROR_STOP=1 -U sprachalyzer -d sprachalyzer >/dev/null
}

rollback_on_error() {
  local status=$?
  trap - ERR INT TERM
  if [[ "$DB_CHANGED" -eq 1 ]]; then
    echo "⚠️  Hostvars-Aktualisierung fehlgeschlagen; stelle das bisherige Datenbankkennwort wieder her." >&2
    alter_database_password "$OLD_POSTGRES_PASSWORD" || true
  fi
  unset AUTH_LINE NEW_AUTH_HASH NEW_POSTGRES_PASSWORD OLD_POSTGRES_PASSWORD
  exit "$status"
}
trap rollback_on_error ERR INT TERM

alter_database_password "$NEW_POSTGRES_PASSWORD"
DB_CHANGED=1

SAL_HOSTVARS_FILE="$HOSTVARS_FILE" \
SAL_NEW_POSTGRES_PASSWORD="$NEW_POSTGRES_PASSWORD" \
SAL_NEW_AUTH_HASH="$NEW_AUTH_HASH" \
python3 - <<'PY'
import os
import re
import tempfile
from pathlib import Path

path = Path(os.environ["SAL_HOSTVARS_FILE"])
text = path.read_text(encoding="utf-8")
replacements = {
    "sal_postgres_password": os.environ["SAL_NEW_POSTGRES_PASSWORD"],
    "sal_basic_auth_password_hash": os.environ["SAL_NEW_AUTH_HASH"],
}
for key, value in replacements.items():
    text, count = re.subn(
        rf"(?m)^{re.escape(key)}:[ \t]*.*$",
        f'{key}: "{value}"',
        text,
    )
    if count != 1:
        raise SystemExit(f"Hostvars-Feld {key} wurde nicht genau einmal gefunden")

fd, temporary_name = tempfile.mkstemp(prefix=f".{path.name}.", dir=path.parent)
try:
    with os.fdopen(fd, "w", encoding="utf-8") as handle:
        handle.write(text)
        handle.flush()
        os.fsync(handle.fileno())
    os.chmod(temporary_name, 0o600)
    os.replace(temporary_name, path)
finally:
    if os.path.exists(temporary_name):
        os.unlink(temporary_name)
PY

DB_CHANGED=0
trap - ERR INT TERM
unset AUTH_LINE NEW_AUTH_HASH NEW_POSTGRES_PASSWORD OLD_POSTGRES_PASSWORD

ok "PostgreSQL- und Basic-Auth-Geheimnisse wurden sicher rotiert."
info "Jetzt unmittelbar ausfuehren: ./scripts/sal-redeploy.sh $DOMAIN"
