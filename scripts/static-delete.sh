#!/usr/bin/env bash
set -euo pipefail

INVENTORY="./ansible/inventory/hosts.ini"
PLAYBOOK="./ansible/playbooks/deploy-static.yml"
HOSTVARS_DIR="./ansible/hostvars"

usage(){ echo "Usage: $0 <domain>"; }
die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }

[[ $# -ge 1 ]] || { usage; exit 1; }
DOMAIN="$1"
HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"

[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars nicht gefunden: $HOSTVARS_FILE"
grep -q '^static_enabled:[[:space:]]*true' "$HOSTVARS_FILE" || die "Hostvars gehören nicht zu einer statischen Site: $HOSTVARS_FILE"

timestamp="$(date +%Y%m%d%H%M%S)"
backup_file="${HOSTVARS_FILE}.bak.${timestamp}"
cp "$HOSTVARS_FILE" "$backup_file"
rm "$HOSTVARS_FILE"

warn "Hostvars entfernt. Inhalte unter /srv/static/${DOMAIN}/ und Auth-Dateien bleiben unverändert erhalten."
info "Backup der Hostvars: $backup_file"

ansible-playbook -i "$INVENTORY" "$PLAYBOOK"
ok "Static-Site aus dem Shared-Container entfernt"
