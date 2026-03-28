#!/usr/bin/env bash
set -euo pipefail

INVENTORY="./ansible/inventory/hosts.ini"
PLAYBOOK="./ansible/playbooks/deploy-static.yml"
HOSTVARS_DIR="./ansible/hostvars"

usage(){ echo "Usage: $0 <domain>|--all [--check-only]"; }
die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }

CHECK_ONLY=0
[[ $# -ge 1 ]] || { usage; exit 1; }
TARGET="$1"; shift || true
[[ "${1:-}" == "--check-only" ]] && CHECK_ONLY=1

if [[ "$TARGET" != "--all" ]]; then
  DOMAIN="$TARGET"
  HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
  [[ -f "$HOSTVARS_FILE" ]] || die "Hostvars nicht gefunden: $HOSTVARS_FILE"
  grep -q '^static_enabled:[[:space:]]*true' "$HOSTVARS_FILE" || die "Hostvars gehören nicht zu einer statischen Site: $HOSTVARS_FILE"

  host_ip="$(curl -fsSL https://api.ipify.org || true)"
  dns_ip="$(dig +short A "$DOMAIN" | head -n1)"
  [[ -n "$host_ip" && -n "$dns_ip" ]] || die "DNS/IP Prüfung fehlgeschlagen"
  [[ "$host_ip" == "$dns_ip" ]] || die "DNS mismatch: $DOMAIN -> $dns_ip (erwartet $host_ip)"
  ok "DNS OK für $DOMAIN"
fi

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only abgeschlossen"
  exit 0
fi

if [[ "$TARGET" == "--all" ]]; then
  ansible-playbook -i "$INVENTORY" "$PLAYBOOK"
  ok "Static-Redeploy für alle statischen Sites abgeschlossen"
else
  ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$PLAYBOOK"
  ok "Static-Redeploy abgeschlossen"
fi
