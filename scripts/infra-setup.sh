#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ANSIBLE_DIR="$ROOT_DIR/ansible"
PUBLIC_IP="$(curl -fsSL https://api.ipify.org || true)"

log() { printf '\n[%s] %s\n' "$(date +%H:%M:%S)" "$*"; }
die() { echo "❌ $*" >&2; exit 1; }
require_root() { [[ ${EUID:-$(id -u)} -eq 0 ]] || die "Bitte als root ausführen (sudo)."; }

have_cmd() { command -v "$1" >/dev/null 2>&1; }

install_docker() {
  if have_cmd docker; then
    log "Docker bereits installiert"
    return
  fi
  log "Installiere Docker"
  apt-get update -y
  apt-get install -y ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  . /etc/os-release
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable" > /etc/apt/sources.list.d/docker.list
  apt-get update -y
  apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
}

install_ansible() {
  if have_cmd ansible-playbook; then
    log "Ansible bereits installiert"
    return
  fi
  log "Installiere Ansible"
  apt-get update -y
  apt-get install -y software-properties-common
  add-apt-repository --yes --update ppa:ansible/ansible
  apt-get install -y ansible
}

install_mysql_client() {
  if have_cmd mysql; then
    log "MySQL Client bereits installiert"
    return
  fi
  log "Installiere MySQL Client"
  apt-get update -y
  apt-get install -y mysql-client
}

install_prereqs() {
  for bin in dig jq python3; do
    if ! have_cmd "$bin"; then
      log "Installiere Voraussetzung: $bin"
      apt-get update -y
      case "$bin" in
        dig) apt-get install -y dnsutils ;;
        jq) apt-get install -y jq ;;
        python3) apt-get install -y python3 ;;
      esac
    fi
  done

  if ! ansible-galaxy collection list 2>/dev/null | rg -q '^community.docker\b'; then
    log "Installiere Ansible Collection community.docker"
    ansible-galaxy collection install community.docker
  fi
}

resolve_a_record() {
  local domain="$1"
  dig +short A "$domain" | head -n1
}

verify_domain_points_here() {
  local domain="$1"
  local a_record
  a_record="$(resolve_a_record "$domain")"

  [[ -n "$a_record" ]] || die "Kein A-Record für ${domain} gefunden."
  [[ -n "$PUBLIC_IP" ]] || die "Öffentliche Host-IP konnte nicht ermittelt werden (api.ipify.org nicht erreichbar)."

  if [[ "$a_record" != "$PUBLIC_IP" ]]; then
    die "DNS-Fehler: ${domain} zeigt auf ${a_record}, erwartet wird ${PUBLIC_IP}. Deployment abgebrochen."
  fi

  log "DNS OK: ${domain} -> ${a_record}"
}

deploy_stack() {
  local portainer_domain="$1"

  log "Deploy MySQL"
  ansible-playbook -i "$ANSIBLE_DIR/inventory/hosts.ini" "$ANSIBLE_DIR/playbooks/deploy-mysql.yml"

  log "Deploy Traefik"
  ansible-playbook -i "$ANSIBLE_DIR/inventory/hosts.ini" "$ANSIBLE_DIR/playbooks/deploy-traefik.yml"

  log "Deploy CrowdSec"
  ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/hosts.ini" \
    -e "crowdsec_base_dir=$ROOT_DIR/data/crowdsec" \
    "$ANSIBLE_DIR/playbooks/deploy-crowdsec.yml"

  verify_domain_points_here "$portainer_domain"

  log "Deploy Portainer"
  ansible-playbook \
    -i "$ANSIBLE_DIR/inventory/hosts.ini" \
    -e "portainer={domain:'$portainer_domain'}" \
    "$ANSIBLE_DIR/playbooks/deploy-portainer.yml"
}

main() {
  require_root
  cd "$ROOT_DIR"

  echo "🌙 Infra Initial-Setup (Docker, Ansible, MySQL, Traefik, Portainer, CrowdSec)"
  read -r -p "Portainer Domain (z.B. portainer.example.com): " PORTAINER_DOMAIN
  [[ -n "${PORTAINER_DOMAIN:-}" ]] || die "Portainer Domain darf nicht leer sein."

  install_docker
  install_ansible
  install_mysql_client
  install_prereqs
  deploy_stack "$PORTAINER_DOMAIN"

  cat <<MSG

✅ Initial-Setup abgeschlossen.

Nächste Schritte:
1) Für jede Website vor Deployment DNS-A-Record prüfen (wird abgebrochen, wenn IP nicht passt).
2) Ghost-Deployment weiterhin via ./scripts/ghost-add.sh <domain>
3) WordPress/Static Sites bitte mit Traefik-Middleware 'crowdsec-default@docker' labeln.

MSG
}

main "$@"
