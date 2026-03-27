#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
REDEPLOY_SCRIPT="$ROOT_DIR/scripts/wp-redeploy.sh"
CHECK_ONLY=0

[[ "${1:-}" == "--check-only" ]] && CHECK_ONLY=1

defaults=(
  'wp_traefik_middleware_default: "crowdsec-default@docker"'
  'wp_traefik_middleware_admin: "crowdsec-admin@docker"'
  'wp_traefik_middleware_api: "crowdsec-api@docker"'
)

mapfile -t files < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' | sort)
for f in "${files[@]}"; do
  grep -q '^wp_domain_db:' "$f" || continue
  domain="$(basename "$f" .yml)"
  for line in "${defaults[@]}"; do
    key="${line%%:*}"
    grep -qE "^${key}:" "$f" || { [[ "$CHECK_ONLY" -eq 1 ]] || printf '%s\n' "$line" >> "$f"; }
  done
  [[ "$CHECK_ONLY" -eq 1 ]] || "$REDEPLOY_SCRIPT" "$domain"
done

echo "✅ WordPress CrowdSec-Migration abgeschlossen"
