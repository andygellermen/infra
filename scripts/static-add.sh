#!/usr/bin/env bash
set -euo pipefail

ANSIBLE_PLAYBOOK="./ansible/playbooks/deploy-static.yml"
INVENTORY="./ansible/inventory"
HOSTVARS_DIR="./ansible/hostvars"

usage() {
  cat <<'USAGE'
Usage: ./scripts/static-add.sh <domain> [alias1 alias2 ...] [--protected-path=/private-folder/] [--protected-realm="Protected Area"]
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

PROTECTED_PATH=""
PROTECTED_REALM="Protected Area"
args=()

for arg in "$@"; do
  case "$arg" in
    --protected-path=*) PROTECTED_PATH="${arg#*=}" ;;
    --protected-realm=*) PROTECTED_REALM="${arg#*=}" ;;
    --help|-h) usage; exit 0 ;;
    *) args+=("$arg") ;;
  esac
done

command -v idn >/dev/null 2>&1 || die "Tool fehlt: idn"
command -v dig >/dev/null 2>&1 || die "Tool fehlt: dig"
command -v curl >/dev/null 2>&1 || die "Tool fehlt: curl"

DOMAIN="$(normalize_domain "${args[0]}")"
ALIASES=()
for a in "${args[@]:1}"; do
  ALIASES+=("$(normalize_domain "$a")")
done

verify_domain_points_here "$DOMAIN"
for alias in "${ALIASES[@]}"; do verify_domain_points_here "$alias"; done

mkdir -p "$HOSTVARS_DIR"
HOSTVARS_FILE="${HOSTVARS_DIR}/${DOMAIN}.yml"

{
  echo "# Hostvars für ${DOMAIN} (Static Site)"
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
  echo "static_enabled: true"
  echo "static_traefik_middleware_default: \"crowdsec-default@docker\""
  if [[ -n "$PROTECTED_PATH" ]]; then
    sanitized_path="$(printf '%s' "$PROTECTED_PATH" | sed -E 's#^/##; s#/$##; s#[^A-Za-z0-9._-]+#-#g')"
    [[ -n "$sanitized_path" ]] || sanitized_path="private"
    echo "static_basic_auth_paths:"
    echo "  - path: \"${PROTECTED_PATH}\""
    echo "    realm: \"${PROTECTED_REALM}\""
    echo "    auth_file: \"/srv/static-auth/${DOMAIN}-${sanitized_path}.htpasswd\""
  else
    echo "static_basic_auth_paths: []"
  fi
} > "$HOSTVARS_FILE"

info "Hostvars geschrieben: $HOSTVARS_FILE"
ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$ANSIBLE_PLAYBOOK"
ok "Static-Site-Setup für ${DOMAIN} abgeschlossen"
