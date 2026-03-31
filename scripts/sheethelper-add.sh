#!/usr/bin/env bash
set -euo pipefail

HOSTVARS_DIR="./ansible/hostvars"

usage() {
  cat <<'USAGE'
Usage: ./scripts/sheethelper-add.sh <domain> [alias1 alias2 ...] [--sheet-id=<id>] [--routes-gid=<gid>] [--vcards-gid=<gid>] [--texts-gid=<gid>] [--theme=<name>] [--wildcard-domain=<apex-domain>] [--dns-account=<key>]
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
ROUTES_GID=""
VCARDS_GID=""
TEXTS_GID=""
THEME="clean"
WILDCARD_DOMAIN=""
DNS_ACCOUNT=""
args=()

for arg in "$@"; do
  case "$arg" in
    --sheet-id=*) SHEET_ID="${arg#*=}" ;;
    --routes-gid=*) ROUTES_GID="${arg#*=}" ;;
    --vcards-gid=*) VCARDS_GID="${arg#*=}" ;;
    --texts-gid=*) TEXTS_GID="${arg#*=}" ;;
    --theme=*) THEME="${arg#*=}" ;;
    --wildcard-domain=*) WILDCARD_DOMAIN="${arg#*=}" ;;
    --dns-account=*) DNS_ACCOUNT="${arg#*=}" ;;
    --help|-h) usage; exit 0 ;;
    *) args+=("$arg") ;;
  esac
done

command -v idn >/dev/null 2>&1 || die "Tool fehlt: idn"

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
  echo "sheet_helper_routes_gid: \"${ROUTES_GID}\""
  echo "sheet_helper_vcards_gid: \"${VCARDS_GID}\""
  echo "sheet_helper_texts_gid: \"${TEXTS_GID}\""
  echo "sheet_helper_default_list_prefix: \"list_\""
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
info "Naechster Schritt: Sheet-IDs/GIDs pruefen und spaeter die Deploy-Integration aktivieren."
ok "Sheet-Helper-Hostvars fuer ${DOMAIN} vorbereitet"
