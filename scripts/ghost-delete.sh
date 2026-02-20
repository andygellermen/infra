#!/usr/bin/env bash
set -euo pipefail

ANSIBLE_PLAYBOOK="./ansible/playbooks/delete-ghost.yml"
INVENTORY="./ansible/inventory"
HOSTVARS_DIR="./ansible/hostvars"

usage() {
  echo "Usage: $0 <domain>"
}

die() {
  echo "❌ Fehler: $*" >&2
  exit 1
}

info() {
  echo "ℹ️  $*"
}

success() {
  echo "✅ $*"
}

if [[ $# -ne 1 ]]; then
  usage
  exit 1
fi

DOMAIN="$1"
CONTAINER_NAME="ghost-${DOMAIN//./-}"
VOLUME_NAME="ghost_${DOMAIN//./_}_content"
HOSTVARS_FILE="${HOSTVARS_DIR}/${DOMAIN}.yml"

info "Prüfe Docker-Container: ${CONTAINER_NAME}"

if docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER_NAME"; then
  info "Stoppe und entferne Ghost-Container..."
  docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
  success "Container entfernt"
else
  info "Kein Ghost-Container gefunden"
fi

if [[ -f "$HOSTVARS_FILE" ]]; then
  info "Starte Ansible-Löschung (DB + DB-User + Volume)"
  ansible-playbook \
    -i "$INVENTORY" \
    -e "target_domain=${DOMAIN}" \
    "$ANSIBLE_PLAYBOOK"
else
  info "Keine Hostvars gefunden (${HOSTVARS_FILE}) – überspringe DB-Löschung via Ansible"
fi

# Fallback: verwaiste Content-Volumes entfernen
if docker volume inspect "$VOLUME_NAME" >/dev/null 2>&1; then
  info "Entferne Ghost-Content-Volume: ${VOLUME_NAME}"
  docker volume rm -f "$VOLUME_NAME" >/dev/null 2>&1 || true
  success "Volume entfernt"
else
  info "Kein Ghost-Content-Volume gefunden"
fi

if [[ -f "$HOSTVARS_FILE" ]]; then
  rm -f "$HOSTVARS_FILE"
  success "Hostvars gelöscht: $HOSTVARS_FILE"
else
  info "Keine Hostvars-Datei gefunden für ${DOMAIN}"
fi

info "Warte kurz, damit Docker Ressourcen freigibt..."
sleep 2

success "Ghost-Löschung für ${DOMAIN} abgeschlossen"
