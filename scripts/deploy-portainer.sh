#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DOMAIN="${1:-}"

if [[ -z "$DOMAIN" ]]; then
  read -r -p "Portainer Domain (z.B. portainer.example.com): " DOMAIN
fi

[[ -n "$DOMAIN" ]] || { echo "❌ Domain fehlt"; exit 1; }

echo "🚀 Deploy Portainer via Ansible (${DOMAIN})"

command -v ansible-playbook >/dev/null 2>&1 || {
  echo "❌ ansible-playbook nicht gefunden"
  exit 1
}

ansible-playbook \
  -i "$ROOT_DIR/ansible/inventory/hosts.ini" \
  -e "portainer={domain:'$DOMAIN'}" \
  "$ROOT_DIR/ansible/playbooks/deploy-portainer.yml"

echo "✅ Portainer verfügbar unter https://${DOMAIN}"
