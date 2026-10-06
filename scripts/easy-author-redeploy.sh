#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="${EASY_AUTHOR_HOSTVARS_DIR:-$ROOT_DIR/ansible/hostvars}"
SRV_ROOT="${EASY_AUTHOR_SRV_ROOT:-/srv/easy-author}"
domain=""; check_only=0; build_only=0; skip_smoke=0
die(){ echo "FEHLER: $*" >&2; exit 1; }
while [[ $# -gt 0 ]]; do
  case "$1" in --check-only) check_only=1;; --build-only) build_only=1;; --skip-auth-smoke) skip_smoke=1;; --help|-h) echo "Usage: $0 <domain> [--check-only] [--build-only] [--skip-auth-smoke]"; exit 0;; *) [[ -z "$domain" ]] || die "Nur eine Domain ist erlaubt."; domain="$1";; esac
  shift
done
[[ "$domain" == author.geller.men ]] || die "Nur author.geller.men ist fuer dieses Staging freigegeben."
hostvars="$HOSTVARS_DIR/$domain.yml"; [[ -f "$hostvars" ]] || die "Hostvars fehlen."
grep -Eq '^easy_author_enabled: true$' "$hostvars" || die "EasyAuthor ist nicht aktiviert."
git -C "$ROOT_DIR" diff --quiet && git -C "$ROOT_DIR" diff --cached --quiet || die "Der getrackte Arbeitsbaum ist nicht sauber."
sha="$(git -C "$ROOT_DIR" rev-parse HEAD)"; [[ "$sha" =~ ^[0-9a-f]{40}$ ]] || die "Git-SHA ist ungueltig."
api_repo="$(sed -nE 's/^easy_author_api_image_repository: "([a-z0-9._\/-]+)"$/\1/p' "$hostvars")"
web_repo="$(sed -nE 's/^easy_author_web_image_repository: "([a-z0-9._\/-]+)"$/\1/p' "$hostvars")"
username="$(sed -nE 's/^easy_author_basic_auth_username: "([A-Za-z0-9._-]+)"$/\1/p' "$hostvars")"
[[ -n "$api_repo" && -n "$web_repo" && -n "$username" ]] || die "Hostvars sind unvollstaendig."
api_image="$api_repo:$sha"; web_image="$web_repo:$sha"
if [[ "$check_only" -eq 0 ]]; then
  docker build --pull -t "$api_image" "$ROOT_DIR/apps/easy-author/backend"
  docker build --pull --target production -t "$web_image" "$ROOT_DIR/apps/easy-author/frontend"
fi
[[ "$build_only" -eq 0 ]] || exit 0
if [[ "$check_only" -eq 0 && -n "$(find "$SRV_ROOT/$domain/data" -mindepth 1 -print -quit 2>/dev/null)" ]]; then
  "$ROOT_DIR/scripts/easy-author-backup.sh" "$domain" >/dev/null
fi
command=(ansible-playbook -i "$ROOT_DIR/ansible/inventory/hosts.ini" "$ROOT_DIR/ansible/playbooks/deploy-easy-author.yml" -e "target_domain=$domain" -e "easy_author_deploy_api_image=$api_image" -e "easy_author_deploy_web_image=$web_image")
[[ "$check_only" -eq 0 ]] || command+=(--check)
"${command[@]}"
if [[ "$check_only" -eq 0 && "$skip_smoke" -eq 0 ]]; then "$ROOT_DIR/scripts/easy-author-smoke-check.sh" "$domain" --username="$username"; fi
