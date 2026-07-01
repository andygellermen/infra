#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$ROOT_DIR/scripts/lib/error-notify.sh"

usage() {
  cat <<'USAGE'
Usage: ./scripts/check-shared-ses.sh [--domain=<domain>] [--check-login]

Prueft die gemeinsamen SES-Secrets aus ansible/secrets/secrets.yml
und vergleicht sie optional mit der gerenderten EEP-Env-Datei.

Optionen:
  --domain=<domain>   Vergleicht mit /srv/easy-event-planner/<domain>/easy-event-planner.env
  --check-login       Testet den SMTP-Login mit den Shared-Secrets (ohne Testmail)
USAGE
}

mask_value() {
  local value="${1:-}" length
  length=${#value}
  if [[ "$length" -eq 0 ]]; then
    printf '(empty)\n'
    return
  fi
  if [[ "$length" -le 8 ]]; then
    printf '[set len=%d]\n' "$length"
    return
  fi
  printf '%s…%s (len=%d)\n' "${value:0:4}" "${value: -4}" "$length"
}

extract_env_value() {
  local key="$1" file="$2"
  awk -F= -v k="$key" '$1 == k {print substr($0, index($0, "=") + 1); exit}' "$file"
}

compare_value() {
  local label="$1" expected="$2" actual="$3"
  if [[ "$expected" == "$actual" ]]; then
    printf '  %-22s %s\n' "$label" "ok"
  else
    printf '  %-22s %s\n' "$label" "DIFF"
  fi
}

DOMAIN=""
CHECK_LOGIN=0

for arg in "$@"; do
  case "$arg" in
    --domain=*) DOMAIN="${arg#*=}" ;;
    --check-login) CHECK_LOGIN=1 ;;
    --help|-h) usage; exit 0 ;;
    *) echo "❌ Unbekannte Option: $arg" >&2; usage; exit 1 ;;
  esac
done

SECRETS_FILE="$(resolve_infra_secrets_file "$ROOT_DIR" || true)"
[[ -n "$SECRETS_FILE" ]] || { echo "❌ Keine Secrets-Datei gefunden unter $ROOT_DIR/ansible/secrets/" >&2; exit 1; }

SES_SMTP_HOST="$(extract_top_level_secret ses_smtp_host "$SECRETS_FILE")"
SES_SMTP_PORT="$(extract_top_level_secret ses_smtp_port "$SECRETS_FILE")"
SES_SMTP_USER="$(extract_top_level_secret ses_smtp_user "$SECRETS_FILE")"
SES_SMTP_PASSWORD="$(extract_top_level_secret ses_smtp_password "$SECRETS_FILE")"
SES_SMTP_SECURE="$(extract_top_level_secret ses_smtp_secure "$SECRETS_FILE")"
SES_SMTP_REQUIRE_TLS="$(extract_top_level_secret ses_smtp_requireTLS "$SECRETS_FILE")"
SES_FROM="$(extract_top_level_secret ses_from "$SECRETS_FILE")"

[[ -n "$SES_SMTP_HOST" ]] || SES_SMTP_HOST="email-smtp.eu-north-1.amazonaws.com"
[[ -n "$SES_SMTP_PORT" ]] || SES_SMTP_PORT="587"
[[ -n "$SES_SMTP_REQUIRE_TLS" ]] || SES_SMTP_REQUIRE_TLS="true"

echo "Shared SES config"
echo "  secrets_file:         $SECRETS_FILE"
echo "  ses_smtp_host:        $SES_SMTP_HOST"
echo "  ses_smtp_port:        $SES_SMTP_PORT"
echo "  ses_smtp_secure:      ${SES_SMTP_SECURE:-false}"
echo "  ses_smtp_requireTLS:  ${SES_SMTP_REQUIRE_TLS}"
echo "  ses_from:             ${SES_FROM:-"(empty)"}"
echo "  ses_smtp_user:        $(mask_value "$SES_SMTP_USER")"
echo "  ses_smtp_password:    $(mask_value "$SES_SMTP_PASSWORD")"

if [[ -n "$DOMAIN" ]]; then
  ENV_FILE="/srv/easy-event-planner/$DOMAIN/easy-event-planner.env"
  if [[ ! -f "$ENV_FILE" ]]; then
    echo
    echo "⚠️  EEP-Env-Datei nicht gefunden: $ENV_FILE"
  else
    echo
    echo "EEP runtime env ($DOMAIN)"
    EEP_MAIL_PROVIDER="$(extract_env_value EEP_MAIL_PROVIDER "$ENV_FILE")"
    EEP_MAIL_FROM="$(extract_env_value EEP_MAIL_FROM "$ENV_FILE")"
    EEP_SES_SMTP_HOST="$(extract_env_value EEP_SES_SMTP_HOST "$ENV_FILE")"
    EEP_SES_SMTP_PORT="$(extract_env_value EEP_SES_SMTP_PORT "$ENV_FILE")"
    EEP_SES_SMTP_USER="$(extract_env_value EEP_SES_SMTP_USER "$ENV_FILE")"
    EEP_SES_SMTP_PASS="$(extract_env_value EEP_SES_SMTP_PASS "$ENV_FILE")"
    echo "  env_file:             $ENV_FILE"
    echo "  EEP_MAIL_PROVIDER:    ${EEP_MAIL_PROVIDER:-"(empty)"}"
    echo "  EEP_MAIL_FROM:        ${EEP_MAIL_FROM:-"(empty)"}"
    echo "  EEP_SES_SMTP_HOST:    ${EEP_SES_SMTP_HOST:-"(empty)"}"
    echo "  EEP_SES_SMTP_PORT:    ${EEP_SES_SMTP_PORT:-"(empty)"}"
    echo "  EEP_SES_SMTP_USER:    $(mask_value "$EEP_SES_SMTP_USER")"
    echo "  EEP_SES_SMTP_PASS:    $(mask_value "$EEP_SES_SMTP_PASS")"
    echo
    echo "Comparison"
    compare_value "provider" "ses" "${EEP_MAIL_PROVIDER:-}"
    compare_value "smtp_host" "$SES_SMTP_HOST" "${EEP_SES_SMTP_HOST:-}"
    compare_value "smtp_port" "$SES_SMTP_PORT" "${EEP_SES_SMTP_PORT:-}"
    compare_value "smtp_user" "$SES_SMTP_USER" "${EEP_SES_SMTP_USER:-}"
    compare_value "smtp_pass" "$SES_SMTP_PASSWORD" "${EEP_SES_SMTP_PASS:-}"
  fi
fi

if [[ "$CHECK_LOGIN" -eq 1 ]]; then
  echo
  echo "SMTP login check"
  python3 "$ROOT_DIR/scripts/check-smtp-login.py" \
    --smtp-host "$SES_SMTP_HOST" \
    --smtp-port "$SES_SMTP_PORT" \
    --smtp-user "$SES_SMTP_USER" \
    --smtp-password "$SES_SMTP_PASSWORD" \
    --smtp-secure "${SES_SMTP_SECURE:-}" \
    --smtp-require-tls "${SES_SMTP_REQUIRE_TLS:-true}"
fi
