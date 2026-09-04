#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ROLE="$ROOT_DIR/ansible/playbooks/roles/sprach-a-lyzer/tasks/main.yml"

bash -n \
  "$ROOT_DIR/scripts/sal-add.sh" \
  "$ROOT_DIR/scripts/sal-rotate-secrets.sh" \
  "$ROOT_DIR/scripts/sal-redeploy.sh" \
  "$ROOT_DIR/scripts/sal-smoke-check.sh"

test -f "$ROOT_DIR/ansible/playbooks/deploy-sprach-a-lyzer.yml"
test -f "$ROOT_DIR/ansible/hostvars/templates/sprach-a-lyzer-hostvars.j2"
git -C "$ROOT_DIR" check-ignore --quiet --no-index ansible/hostvars/sal.geller.men.yml

grep -q 'internal: true' "$ROLE"
grep -q 'basicauth.users' "$ROLE"
grep -q 'basicauth.removeheader' "$ROLE"
grep -q 'status_code: 401' "$ROLE"
grep -q 'tls.certresolver' "$ROLE"
grep -q 'sprach-a-lyzer-migrate' "$ROLE"
grep -q 'sprach-a-lyzer-seed' "$ROLE"
grep -q 'htpasswd -nB -C 12' "$ROOT_DIR/scripts/sal-add.sh"
grep -q 'htpasswd -nB -C 12' "$ROOT_DIR/scripts/sal-rotate-secrets.sh"
grep -q 'ALTER ROLE sprachalyzer' "$ROOT_DIR/scripts/sal-rotate-secrets.sh"
grep -q 'loop: "{{ sal_runtime_sites }}"' "$ROLE"

if grep -Eq 'traefik\.(http|docker)[^:]*[{][{]' "$ROLE"; then
  echo "Dynamic Traefik label keys must be constructed as a rendered map" >&2
  exit 1
fi

if grep -Eq 'published_ports:|^[[:space:]]+ports:' "$ROLE"; then
  echo "Sprach-A-Lyzer deployment must not publish host ports" >&2
  exit 1
fi

if grep -q 'purge_networks:' "$ROLE"; then
  echo "Sprach-A-Lyzer deployment must remain compatible with the server Docker collection" >&2
  exit 1
fi
