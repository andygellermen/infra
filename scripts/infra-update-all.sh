#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ANSIBLE_DIR="$ROOT_DIR/ansible"
INVENTORY="$ANSIBLE_DIR/inventory/hosts.ini"

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/infra-update-all.sh [--skip-mysql] [--skip-traefik] [--skip-crowdsec] [--skip-portainer] [--portainer-domain=<fqdn>]

Description:
  Führt ein Update/Redeploy der zentralen Infra-Container durch:
  MySQL, Traefik, CrowdSec und Portainer.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }

SKIP_MYSQL=0
SKIP_TRAEFIK=0
SKIP_CROWDSEC=0
SKIP_PORTAINER=0
PORTAINER_DOMAIN=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-mysql) SKIP_MYSQL=1; shift ;;
    --skip-traefik) SKIP_TRAEFIK=1; shift ;;
    --skip-crowdsec) SKIP_CROWDSEC=1; shift ;;
    --skip-portainer) SKIP_PORTAINER=1; shift ;;
    --portainer-domain=*) PORTAINER_DOMAIN="${1#*=}"; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ -f "$INVENTORY" ]] || die "Inventory nicht gefunden: $INVENTORY"

run_playbook() {
  local playbook="$1"
  shift || true
  ansible-playbook -i "$INVENTORY" "$playbook" "$@"
}

if [[ "$SKIP_MYSQL" -eq 0 ]]; then
  info "Update MySQL (infra-mysql)"
  run_playbook "$ANSIBLE_DIR/playbooks/deploy-mysql.yml"
  ok "MySQL aktualisiert"
fi

if [[ "$SKIP_TRAEFIK" -eq 0 ]]; then
  info "Update Traefik (infra-traefik)"
  run_playbook "$ANSIBLE_DIR/playbooks/deploy-traefik.yml"
  ok "Traefik aktualisiert"
fi

if [[ "$SKIP_CROWDSEC" -eq 0 ]]; then
  info "Update CrowdSec"
  run_playbook "$ANSIBLE_DIR/playbooks/deploy-crowdsec.yml"
  ok "CrowdSec aktualisiert"
fi

if [[ "$SKIP_PORTAINER" -eq 0 ]]; then
  [[ -n "$PORTAINER_DOMAIN" ]] || die "Für Portainer bitte --portainer-domain=<fqdn> setzen oder --skip-portainer nutzen."
  info "Update Portainer (infra-portainer)"
  run_playbook "$ANSIBLE_DIR/playbooks/deploy-portainer.yml" -e "portainer={domain:'$PORTAINER_DOMAIN'}"
  ok "Portainer aktualisiert"
fi

ok "Infra-Update abgeschlossen."
