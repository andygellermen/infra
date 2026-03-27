#!/usr/bin/env bash
set -euo pipefail

ANSIBLE_PLAYBOOK="./ansible/playbooks/delete-wordpress.yml"
INVENTORY="./ansible/inventory"
HOSTVARS_DIR="./ansible/hostvars"

usage(){ echo "Usage: $0 <domain>"; }
die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }

[[ $# -eq 1 ]] || { usage; exit 1; }
DOMAIN="$1"
HOSTVARS_FILE="${HOSTVARS_DIR}/${DOMAIN}.yml"

if [[ -f "$HOSTVARS_FILE" ]]; then
  ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$ANSIBLE_PLAYBOOK"
else
  die "Hostvars fehlt: $HOSTVARS_FILE"
fi

rm -f "$HOSTVARS_FILE"
ok "WordPress-Löschung für ${DOMAIN} abgeschlossen"
