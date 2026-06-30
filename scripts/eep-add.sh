#!/usr/bin/env bash
set -euo pipefail

HOSTVARS_DIR="./ansible/hostvars"

source "./scripts/lib/dns-check.sh"

usage() {
  cat <<'USAGE'
Usage: ./scripts/eep-add.sh <domain> [alias1 alias2 ...] [--tenant-slug=<slug>] [--tenant-name=<name>] [--base-url=<url>] [--mail-provider=<log|smtp|ses>] [--mail-from=<email>] [--mail-from-name=<name>] [--seed-enabled=<true|false>] [--wildcard-domain=<apex-domain>] [--dns-account=<key>] [--skip-dns-check]
USAGE
}

die(){ echo "❌ Fehler: $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }

normalize_domain() {
  local d="$1"
  if [[ "$d" =~ ^[a-zA-Z0-9.-]+$ ]]; then
    printf '%s\n' "$d"
    return
  fi
  command -v idn >/dev/null 2>&1 || die "Domain enthaelt Nicht-ASCII-Zeichen, aber Tool fehlt: idn"
  idn --quiet --uts46 "$d"
}

normalize_bool() {
  local value
  value="$(printf '%s' "${1:-}" | tr '[:upper:]' '[:lower:]')"
  case "$value" in
    1|true|yes|on) printf 'true\n' ;;
    0|false|no|off) printf 'false\n' ;;
    *) die "Ungueltiger Bool-Wert: $value (erlaubt: true|false)" ;;
  esac
}

verify_domain_points_here() {
  local domain="$1" host_ip
  host_ip="$(curl -fsSL https://api.ipify.org || true)"
  [[ -n "$host_ip" ]] || die "Oeffentliche Host-IP konnte nicht ermittelt werden."
  verify_domain_resolves_to_host_ipv4 "$domain" "$host_ip"
}

[[ $# -ge 1 ]] || { usage; exit 1; }

TENANT_SLUG="demo"
TENANT_NAME="Demo Tenant"
BASE_URL=""
MAIL_PROVIDER="ses"
MAIL_FROM=""
MAIL_FROM_NAME="Easy Event Planner"
SEED_ENABLED="true"
SEED_TENANT_TIMEZONE="Europe/Berlin"
SEED_TENANT_LOCALE="de-DE"
SEED_RETENTION_DAYS="30"
SEED_PAYPAL_MODE="disabled"
WILDCARD_DOMAIN=""
DNS_ACCOUNT=""
SKIP_DNS_CHECK=0
args=()

for arg in "$@"; do
  case "$arg" in
    --tenant-slug=*) TENANT_SLUG="${arg#*=}" ;;
    --tenant-name=*) TENANT_NAME="${arg#*=}" ;;
    --base-url=*) BASE_URL="${arg#*=}" ;;
    --mail-provider=*) MAIL_PROVIDER="${arg#*=}" ;;
    --mail-from=*) MAIL_FROM="${arg#*=}" ;;
    --mail-from-name=*) MAIL_FROM_NAME="${arg#*=}" ;;
    --seed-enabled=*) SEED_ENABLED="$(normalize_bool "${arg#*=}")" ;;
    --wildcard-domain=*) WILDCARD_DOMAIN="${arg#*=}" ;;
    --dns-account=*) DNS_ACCOUNT="${arg#*=}" ;;
    --skip-dns-check) SKIP_DNS_CHECK=1 ;;
    --help|-h) usage; exit 0 ;;
    *) args+=("$arg") ;;
  esac
done

command -v dig >/dev/null 2>&1 || die "Tool fehlt: dig"
command -v curl >/dev/null 2>&1 || die "Tool fehlt: curl"
command -v openssl >/dev/null 2>&1 || die "Tool fehlt: openssl"

DOMAIN="$(normalize_domain "${args[0]}")"
ALIASES=()
for alias in "${args[@]:1}"; do
  ALIASES+=("$(normalize_domain "$alias")")
done

if [[ -n "$WILDCARD_DOMAIN" ]]; then
  WILDCARD_DOMAIN="$(normalize_domain "$WILDCARD_DOMAIN")"
fi

[[ -n "$TENANT_SLUG" ]] || die "tenant slug darf nicht leer sein."
[[ -n "$TENANT_NAME" ]] || die "tenant name darf nicht leer sein."

MAIL_PROVIDER="$(printf '%s' "$MAIL_PROVIDER" | tr '[:upper:]' '[:lower:]')"
case "$MAIL_PROVIDER" in
  log|smtp|ses) ;;
  *) die "mail provider muss log, smtp oder ses sein." ;;
esac

TLS_MODE="standard"
if [[ -n "$WILDCARD_DOMAIN" ]]; then
  TLS_MODE="wildcard"
fi

if [[ -z "$BASE_URL" ]]; then
  BASE_URL="https://${DOMAIN}"
fi
BASE_URL="${BASE_URL%/}"

SEED_TENANT_PUBLIC_BASE_URL="${BASE_URL}/${TENANT_SLUG}"

mkdir -p "$HOSTVARS_DIR"
HOSTVARS_FILE="${HOSTVARS_DIR}/${DOMAIN}.yml"
[[ ! -e "$HOSTVARS_FILE" ]] || die "Hostvars-Datei existiert bereits: $HOSTVARS_FILE"

if [[ "$SKIP_DNS_CHECK" -eq 0 ]]; then
  verify_domain_points_here "$DOMAIN"
  if (( ${#ALIASES[@]} > 0 )); then
    for alias in "${ALIASES[@]}"; do
      verify_domain_points_here "$alias"
    done
  fi
else
  info "Ueberspringe DNS-Checks (--skip-dns-check)."
fi

TOKEN_PEPPER="$(openssl rand -hex 32)"

{
  echo "# Hostvars fuer ${DOMAIN} (Easy Event Planner)"
  echo "domain: ${DOMAIN}"
  echo
  echo "traefik:"
  echo "  domain: ${DOMAIN}"
  echo "  aliases:"
  echo "    - www.${DOMAIN}"
  if (( ${#ALIASES[@]} > 0 )); then
    for alias in "${ALIASES[@]}"; do
      echo "    - ${alias}"
      echo "    - www.${alias}"
    done
  fi
  echo
  echo "eep_enabled: true"
  echo "eep_image: \"easy-event-planner:latest\""
  echo "eep_env: \"production\""
  echo "eep_http_addr: \":8080\""
  echo "eep_base_url: \"${BASE_URL}\""
  echo "eep_db_driver: \"sqlite\""
  echo "eep_db_path: \"/data/easy-event-planner.sqlite\""
  echo "eep_certificate_storage_dir: \"/certificates\""
  echo "eep_token_pepper: \"${TOKEN_PEPPER}\""
  echo "eep_session_cookie_name: \"eep_session\""
  echo "eep_secure_cookies: true"
  echo
  echo "eep_mail_provider: \"${MAIL_PROVIDER}\""
  echo "eep_mail_from: \"${MAIL_FROM}\""
  echo "eep_mail_from_name: \"${MAIL_FROM_NAME}\""
  echo "eep_ses_region: \"eu-north-1\""
  echo "eep_ses_smtp_host: \"\""
  echo "eep_ses_smtp_port: 587"
  echo "eep_ses_smtp_user: \"\""
  echo "eep_ses_smtp_pass: \"\""
  echo
  echo "eep_paypal_use_real_api: false"
  echo "eep_paypal_client_id: \"\""
  echo "eep_paypal_client_secret: \"\""
  echo "eep_paypal_webhook_id: \"\""
  echo "eep_paypal_sandbox_api_base_url: \"https://api-m.sandbox.paypal.com\""
  echo "eep_paypal_live_api_base_url: \"https://api-m.paypal.com\""
  echo "eep_paypal_http_timeout: \"15s\""
  echo
  echo "eep_email_worker_poll_interval: \"3s\""
  echo "eep_email_worker_batch_size: 10"
  echo
  echo "eep_seed_enabled: ${SEED_ENABLED}"
  echo "eep_seed_tenant_slug: \"${TENANT_SLUG}\""
  echo "eep_seed_tenant_name: \"${TENANT_NAME}\""
  echo "eep_seed_tenant_public_base_url: \"${SEED_TENANT_PUBLIC_BASE_URL}\""
  echo "eep_seed_tenant_timezone: \"${SEED_TENANT_TIMEZONE}\""
  echo "eep_seed_tenant_locale: \"${SEED_TENANT_LOCALE}\""
  echo "eep_seed_retention_days: ${SEED_RETENTION_DAYS}"
  echo "eep_seed_sender_email: \"${MAIL_FROM}\""
  echo "eep_seed_sender_name: \"${MAIL_FROM_NAME}\""
  echo "eep_seed_paypal_mode: \"${SEED_PAYPAL_MODE}\""
  echo
  echo "eep_traefik_middleware_default: \"\""
  echo
  echo "tls_mode: \"${TLS_MODE}\""
  echo "tls_wildcard_domain: \"${WILDCARD_DOMAIN}\""
  echo "tls_dns_account: \"${DNS_ACCOUNT}\""
} > "$HOSTVARS_FILE"

info "Hostvars geschrieben: $HOSTVARS_FILE"
ok "Easy-Event-Planner-Hostvars fuer ${DOMAIN} vorbereitet"
info "Naechster Schritt: ./scripts/eep-redeploy.sh ${DOMAIN}"
