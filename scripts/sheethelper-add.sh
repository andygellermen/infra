#!/usr/bin/env bash
set -euo pipefail

HOSTVARS_DIR="./ansible/hostvars"

usage() {
  cat <<'USAGE'
Usage: ./scripts/sheethelper-add.sh <domain> [alias1 alias2 ...] [--sheet-id=<id>] [--published-url=<url>] [--cookie-secret=<secret>] [--sync-token=<token>] [--routes-sheet=<name>] [--vcards-sheet=<name>] [--texts-sheet=<name>] [--list-prefix=<prefix>] [--theme=<name>] [--wildcard-domain=<apex-domain>] [--dns-account=<key>]
USAGE
}

die(){ echo "❌ Fehler: $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }

normalize_domain() {
  local d="$1"
  if [[ "$d" =~ ^[a-zA-Z0-9.-]+$ ]]; then
    printf '%s\n' "$d"
  else
    idn --quiet --uts46 "$d"
  fi
}

[[ $# -ge 1 ]] || { usage; exit 1; }

SHEET_ID=""
PUBLISHED_URL=""
COOKIE_SECRET=""
SYNC_TOKEN=""
ROUTES_SHEET="routes"
VCARDS_SHEET="vcard_entries"
TEXTS_SHEET="text_entries"
LIST_PREFIX="list_"
THEME="clean"
WILDCARD_DOMAIN=""
DNS_ACCOUNT=""
args=()

for arg in "$@"; do
  case "$arg" in
    --sheet-id=*) SHEET_ID="${arg#*=}" ;;
    --published-url=*) PUBLISHED_URL="${arg#*=}" ;;
    --cookie-secret=*) COOKIE_SECRET="${arg#*=}" ;;
    --sync-token=*) SYNC_TOKEN="${arg#*=}" ;;
    --routes-sheet=*) ROUTES_SHEET="${arg#*=}" ;;
    --vcards-sheet=*) VCARDS_SHEET="${arg#*=}" ;;
    --texts-sheet=*) TEXTS_SHEET="${arg#*=}" ;;
    --list-prefix=*) LIST_PREFIX="${arg#*=}" ;;
    --theme=*) THEME="${arg#*=}" ;;
    --wildcard-domain=*) WILDCARD_DOMAIN="${arg#*=}" ;;
    --dns-account=*) DNS_ACCOUNT="${arg#*=}" ;;
    --help|-h) usage; exit 0 ;;
    *) args+=("$arg") ;;
  esac
done

command -v idn >/dev/null 2>&1 || die "Tool fehlt: idn"

if [[ -z "$COOKIE_SECRET" ]] && command -v openssl >/dev/null 2>&1; then
  COOKIE_SECRET="$(openssl rand -hex 24)"
fi
if [[ -z "$SYNC_TOKEN" ]] && command -v openssl >/dev/null 2>&1; then
  SYNC_TOKEN="$(openssl rand -hex 24)"
fi

DOMAIN="$(normalize_domain "${args[0]}")"
ALIASES=()
for a in "${args[@]:1}"; do
  ALIASES+=("$(normalize_domain "$a")")
done

if [[ -n "$WILDCARD_DOMAIN" ]]; then
  WILDCARD_DOMAIN="$(normalize_domain "$WILDCARD_DOMAIN")"
fi

TLS_MODE="standard"
if [[ -n "$WILDCARD_DOMAIN" ]]; then
  TLS_MODE="wildcard"
fi

mkdir -p "$HOSTVARS_DIR"
HOSTVARS_FILE="${HOSTVARS_DIR}/${DOMAIN}.yml"

if [[ -e "$HOSTVARS_FILE" ]]; then
  die "Hostvars-Datei existiert bereits: $HOSTVARS_FILE"
fi

{
  echo "# Hostvars fuer ${DOMAIN} (Sheet Helper)"
  echo "domain: ${DOMAIN}"
  echo
  echo "traefik:"
  echo "  domain: ${DOMAIN}"
  echo "  aliases:"
  echo "    - www.${DOMAIN}"
  for a in "${ALIASES[@]}"; do
    echo "    - ${a}"
    echo "    - www.${a}"
  done
  echo
  echo "sheet_helper_enabled: true"
  echo "sheet_helper_mode: \"public_csv\""
  echo "sheet_helper_sheet_id: \"${SHEET_ID}\""
  echo "sheet_helper_published_url: \"${PUBLISHED_URL}\""
  echo "sheet_helper_cookie_secret: \"${COOKIE_SECRET}\""
  echo "sheet_helper_sync_token: \"${SYNC_TOKEN}\""
  echo "sheet_helper_routes_sheet: \"${ROUTES_SHEET}\""
  echo "sheet_helper_vcards_sheet: \"${VCARDS_SHEET}\""
  echo "sheet_helper_texts_sheet: \"${TEXTS_SHEET}\""
  echo "sheet_helper_default_list_prefix: \"${LIST_PREFIX}\""
  echo "sheet_helper_theme: \"${THEME}\""
  echo "sheet_helper_sync_mode: \"hybrid\""
  echo "sheet_helper_sync_interval: \"15m\""
  echo "sheet_helper_click_sync_interval: \"24h\""
  echo "sheet_helper_allow_text: true"
  echo "sheet_helper_allow_vcard: true"
  echo "sheet_helper_allow_list: true"
  echo "sheet_helper_allow_link: true"
  echo "sheet_helper_require_rate_limit: true"
  echo "sheet_helper_container_name: \"sheet-helper\""
  echo "sheet_helper_data_dir: \"/srv/sheet-helper\""
  echo
  echo "tls_mode: \"${TLS_MODE}\""
  echo "tls_wildcard_domain: \"${WILDCARD_DOMAIN}\""
  echo "tls_dns_account: \"${DNS_ACCOUNT}\""
} > "$HOSTVARS_FILE"

info "Hostvars geschrieben: $HOSTVARS_FILE"
info "Naechster Schritt: Blattnamen pruefen und spaeter die Deploy-Integration aktivieren."
ok "Sheet-Helper-Hostvars fuer ${DOMAIN} vorbereitet"
