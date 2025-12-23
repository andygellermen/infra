#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "🚀 Deploy Portainer via Ansible"

command -v ansible-playbook >/dev/null 2>&1 || {
  echo "❌ ansible-playbook nicht gefunden"
  exit 1
}

ansible-playbook \
  -i "$ROOT_DIR/ansible/inventory/hosts.ini" \
  "$ROOT_DIR/ansible/playbooks/deploy-portainer.yml"

echo "✅ Portainer verfügbar unter https://infra.geller.men"
