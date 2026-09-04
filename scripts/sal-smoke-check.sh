#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: ./scripts/sal-smoke-check.sh <domain> --username=<name> [--password-stdin]

Without SAL_SMOKE_PASSWORD or --password-stdin the password is requested from
the controlling terminal and is never written to disk or command history.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }

DOMAIN=""
AUTH_USERNAME=""
PASSWORD_STDIN=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --username=*) AUTH_USERNAME="${1#*=}"; shift ;;
    --password-stdin) PASSWORD_STDIN=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *)
      [[ -z "$DOMAIN" ]] || die "Nur eine Domain ist erlaubt."
      DOMAIN="$1"
      shift
      ;;
  esac
done

[[ "$DOMAIN" =~ ^[a-z0-9.-]+$ ]] || die "Ungueltige oder fehlende Domain."
[[ "$AUTH_USERNAME" =~ ^[A-Za-z0-9._-]+$ ]] || die "--username fehlt oder ist ungueltig."
command -v curl >/dev/null 2>&1 || die "Tool fehlt: curl"

unauth_status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' "https://${DOMAIN}/")"
[[ "$unauth_status" == "401" ]] || die "Ohne Auth wurde HTTP ${unauth_status} statt 401 geliefert."
ok "Unauthentifizierter Zugriff wird mit HTTP 401 abgewiesen."

if [[ "$PASSWORD_STDIN" -eq 1 ]]; then
  IFS= read -r AUTH_PASSWORD
elif [[ -n "${SAL_SMOKE_PASSWORD:-}" ]]; then
  AUTH_PASSWORD="$SAL_SMOKE_PASSWORD"
elif [[ -r /dev/tty ]]; then
  IFS= read -r -s -p "Basic-Auth-Kennwort fuer ${AUTH_USERNAME}: " AUTH_PASSWORD </dev/tty
  printf '\n' >/dev/tty
else
  die "Kein Kennwort verfuegbar; SAL_SMOKE_PASSWORD setzen oder --password-stdin verwenden."
fi
[[ -n "$AUTH_PASSWORD" ]] || die "Das Kennwort darf nicht leer sein."

product_status="$(curl --silent --show-error --user "${AUTH_USERNAME}:${AUTH_PASSWORD}" --output /dev/null --write-out '%{http_code}' "https://${DOMAIN}/")"
[[ "$product_status" == "200" ]] || die "Authentifizierte Produktseite liefert HTTP ${product_status} statt 200."

ready_body="$(curl --fail-with-body --silent --show-error --user "${AUTH_USERNAME}:${AUTH_PASSWORD}" "https://${DOMAIN}/health/ready")"
unset AUTH_PASSWORD SAL_SMOKE_PASSWORD
[[ "$ready_body" == *'"status":"ready"'* ]] || die "Readiness-Antwort ist unerwartet: $ready_body"
ok "Produktseite und Datenbank-Readiness sind authentifiziert erreichbar."
