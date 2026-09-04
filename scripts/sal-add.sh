#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
source "$ROOT_DIR/scripts/lib/dns-check.sh"

usage() {
  cat <<'USAGE'
Usage: ./scripts/sal-add.sh <domain> [--username=<name>] [--skip-dns-check]

Creates ignored, mode-0600 hostvars with a generated PostgreSQL password and
an interactively entered bcrypt Basic-Auth password. Run this on the Infra server.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

DOMAIN=""
AUTH_USERNAME="andy"
SKIP_DNS_CHECK=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --username=*) AUTH_USERNAME="${1#*=}"; shift ;;
    --skip-dns-check) SKIP_DNS_CHECK=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *)
      [[ -z "$DOMAIN" ]] || die "Nur eine Domain ist erlaubt."
      DOMAIN="$1"
      shift
      ;;
  esac
done

[[ "$DOMAIN" =~ ^[a-z0-9.-]+$ ]] || die "Ungueltige oder fehlende Domain."
[[ "$AUTH_USERNAME" =~ ^[A-Za-z0-9._-]+$ ]] || die "Ungueltiger Basic-Auth-Benutzername."
require_cmd curl
require_cmd dig
require_cmd htpasswd
require_cmd openssl

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
[[ ! -e "$HOSTVARS_FILE" ]] || die "Hostvars existieren bereits: $HOSTVARS_FILE"

if [[ "$SKIP_DNS_CHECK" -eq 0 ]]; then
  HOST_IP="$(curl -fsSL https://api.ipify.org)"
  verify_domain_resolves_to_host_ipv4 "$DOMAIN" "$HOST_IP"
else
  info "DNS-Pruefung uebersprungen."
fi

info "Bitte jetzt das Zugriffskennwort fuer ${AUTH_USERNAME} zweimal eingeben."
AUTH_LINE="$(htpasswd -nB -C 12 "$AUTH_USERNAME")"
AUTH_HASH="${AUTH_LINE#*:}"
POSTGRES_PASSWORD="$(openssl rand -hex 32)"

mkdir -p "$HOSTVARS_DIR"
umask 077
{
  printf '# Hostvars fuer %s (Sprach-A-Lyzer protected MVP)\n' "$DOMAIN"
  printf 'domain: %s\n\n' "$DOMAIN"
  printf 'traefik:\n  domain: %s\n  aliases: []\n\n' "$DOMAIN"
  printf 'sal_enabled: true\n'
  printf 'sal_image: "sprach-a-lyzer:latest"\n'
  printf 'sal_postgres_image: "postgres:17-alpine"\n'
  printf 'sal_postgres_password: "%s"\n' "$POSTGRES_PASSWORD"
  printf 'sal_max_request_bytes: 65536\n\n'
  printf 'sal_basic_auth_username: "%s"\n' "$AUTH_USERNAME"
  printf 'sal_basic_auth_password_hash: "%s"\n' "$AUTH_HASH"
  printf 'sal_basic_auth_realm: "Sprach-A-Lyzer MVP"\n'
  printf 'sal_traefik_middleware_default: ""\n\n'
  printf 'sal_smoke_retries: 20\n'
  printf 'sal_smoke_delay_seconds: 3\n'
} > "$HOSTVARS_FILE"
chmod 0600 "$HOSTVARS_FILE"
unset AUTH_LINE AUTH_HASH POSTGRES_PASSWORD

ok "Geschuetzte Hostvars erstellt: $HOSTVARS_FILE"
info "Naechster Schritt: ./scripts/sal-redeploy.sh $DOMAIN"
