#!/bin/bash
set -e

DOMAIN="$1"

if [[ -z "$DOMAIN" ]]; then
  echo "Usage: $0 <domain> [--purge]"
  exit 1
fi

HOSTVARS="ansible/hostvars/$DOMAIN.yml"

if [[ ! -f "$HOSTVARS" ]]; then
  echo "❌ Hostvars nicht gefunden: $HOSTVARS"
  exit 1
fi

echo "⚠️  Lösche Ghost-Instanz für $DOMAIN"
read -p "Wirklich löschen? [y/N] " confirm
[[ "$confirm" != "y" ]] && exit 0

# Container
docker rm -f "ghost-$DOMAIN" 2>/dev/null || true

# Volume
docker volume rm "ghost_${DOMAIN//./_}_content" 2>/dev/null || true

# DB (über Ansible oder mysql-client)
ansible-playbook \
  -i ./ansible/inventory \
  -e target_domain="$DOMAIN" \
  ./ansible/playbooks/delete-ghost.yml

# Hostvars
rm -f "$HOSTVARS"

echo "🗑️  Ghost $DOMAIN entfernt"
