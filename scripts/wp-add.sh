#!/usr/bin/env bash
set -euo pipefail

ANSIBLE_PLAYBOOK="./ansible/playbooks/deploy-wordpress.yml"
INVENTORY="./ansible/inventory"
HOSTVARS_DIR="./ansible/hostvars"
HOSTVARS_NORMALIZER="./scripts/wp-normalize-hostvars.py"
DEFAULT_WP_VERSION="latest"

usage() {
  cat <<USAGE
Usage: $0 <domain> [alias1 alias2 ...] [--version=<tag|latest>] [--wildcard-domain=<apex-domain>] [--dns-account=<key>]
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

resolve_a_record(){ dig +short A "$1" | head -n1; }
verify_domain_points_here() {
  local domain="$1" host_ip dns_ip
  host_ip="$(curl -fsSL https://api.ipify.org || true)"
  dns_ip="$(resolve_a_record "$domain")"
  [[ -n "$host_ip" ]] || die "Öffentliche Host-IP konnte nicht ermittelt werden."
  [[ -n "$dns_ip" ]] || die "Kein A-Record für ${domain} gefunden."
  [[ "$dns_ip" == "$host_ip" ]] || die "DNS-Fehler: ${domain} zeigt auf ${dns_ip}, erwartet wird ${host_ip}."
}

[[ $# -ge 1 ]] || { usage; exit 1; }
wp_version="$DEFAULT_WP_VERSION"
wildcard_domain=""
dns_account=""
args=()
for arg in "$@"; do
  case "$arg" in
    --version=*) wp_version="${arg#*=}" ;;
    --wildcard-domain=*) wildcard_domain="${arg#*=}" ;;
    --dns-account=*) dns_account="${arg#*=}" ;;
    --help|-h) usage; exit 0 ;;
    *) args+=("$arg") ;;
  esac
done

DOMAIN="$(normalize_domain "${args[0]}")"
if [[ -n "$wildcard_domain" ]]; then
  wildcard_domain="$(normalize_domain "$wildcard_domain")"
fi
tls_mode="standard"
if [[ -n "$wildcard_domain" ]]; then
  tls_mode="wildcard"
fi
ALIASES=()
for a in "${args[@]:1}"; do
  ALIASES+=("$(normalize_domain "$a")")
done

command -v idn >/dev/null 2>&1 || die "Tool fehlt: idn"
command -v dig >/dev/null 2>&1 || die "Tool fehlt: dig"
command -v curl >/dev/null 2>&1 || die "Tool fehlt: curl"
command -v python3 >/dev/null 2>&1 || die "Tool fehlt: python3"
[[ -f "$HOSTVARS_NORMALIZER" ]] || die "Hostvars-Normalizer fehlt: $HOSTVARS_NORMALIZER"

verify_domain_points_here "$DOMAIN"
for alias in "${ALIASES[@]}"; do verify_domain_points_here "$alias"; done

mkdir -p "$HOSTVARS_DIR"
HOSTVARS_FILE="${HOSTVARS_DIR}/${DOMAIN}.yml"
wp_hash="$(printf '%s' "$DOMAIN" | md5sum | awk '{print $1}')"
wp_db="wp_${wp_hash:0:24}_db"
wp_user="${wp_hash:0:24}_usr"
wp_pwd="$(openssl rand -hex 16)"

{
  echo "# Hostvars für ${DOMAIN} (WordPress)"
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
  echo "wp_domain_db: ${wp_db}"
  echo "wp_domain_usr: ${wp_user}"
  echo "wp_domain_pwd: ${wp_pwd}"
  echo "wp_table_prefix: wp_"
  echo "wp_version: \"${wp_version}\""
  echo "wp_traefik_middleware_default: \"crowdsec-default@docker\""
  echo "wp_traefik_middleware_admin: \"crowdsec-admin@docker\""
  echo "wp_traefik_middleware_api: \"crowdsec-api@docker\""
  echo "tls_mode: \"${tls_mode}\""
  echo "tls_wildcard_domain: \"${wildcard_domain}\""
  echo "tls_dns_account: \"${dns_account}\""
} > "$HOSTVARS_FILE"

python3 "$HOSTVARS_NORMALIZER" "$HOSTVARS_FILE" "$DOMAIN" "${ALIASES[@]}"
info "Hostvars geschrieben: $HOSTVARS_FILE"
ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$ANSIBLE_PLAYBOOK"
ok "WordPress-Setup für ${DOMAIN} abgeschlossen"
