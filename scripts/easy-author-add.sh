#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="${EASY_AUTHOR_HOSTVARS_DIR:-$ROOT_DIR/ansible/hostvars}"
source "$ROOT_DIR/scripts/lib/dns-check.sh"

usage(){ echo "Usage: $0 <domain> [--username=<name>] [--skip-dns-check]"; }
die(){ echo "FEHLER: $*" >&2; exit 1; }
ok(){ echo "OK: $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

domain=""; username="andy"; skip_dns=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --username=*) username="${1#*=}" ;;
    --skip-dns-check) skip_dns=1 ;;
    --help|-h) usage; exit 0 ;;
    *) [[ -z "$domain" ]] || die "Nur eine Domain ist erlaubt."; domain="$1" ;;
  esac
  shift
done
[[ "$domain" =~ ^[a-z0-9]+([.-][a-z0-9]+)*$ ]] || die "Ungueltige oder fehlende Domain."
[[ "$username" =~ ^[A-Za-z0-9._-]+$ ]] || die "Ungueltiger Benutzername."
require_cmd htpasswd
if [[ "$skip_dns" -eq 0 ]]; then
  require_cmd curl; require_cmd dig
  host_ip="$(curl -fsSL https://api.ipify.org)"
  verify_domain_resolves_to_host_ipv4 "$domain" "$host_ip"
fi

mkdir -p "$HOSTVARS_DIR"
hostvars_file="$HOSTVARS_DIR/$domain.yml"
[[ ! -e "$hostvars_file" ]] || die "Hostvars existieren bereits: $hostvars_file"
umask 077
temporary="$(mktemp "$HOSTVARS_DIR/.$domain.yml.XXXXXX")"
cleanup(){ rm -f "$temporary"; }
trap cleanup EXIT INT TERM
auth_line="$(htpasswd -nB -C 12 "$username")"
auth_hash="${auth_line#*:}"
[[ "$auth_hash" =~ ^\$2[aby]\$[0-9]{2}\$ ]] || die "htpasswd lieferte keinen gueltigen bcrypt-Hash."
cat > "$temporary" <<EOF
# EasyAuthor protected staging; niemals committen.
domain: "$domain"
easy_author_enabled: true
easy_author_api_image_repository: "easy-author-api"
easy_author_web_image_repository: "easy-author-web"
easy_author_basic_auth_username: "$username"
easy_author_basic_auth_password_hash: "$auth_hash"
easy_author_basic_auth_realm: "EasyAuthor Staging"
easy_author_instance_root: "/srv/easy-author/$domain"
easy_author_backup_retention: 7
easy_author_smoke_retries: 20
easy_author_smoke_delay_seconds: 3
EOF
chmod 0600 "$temporary"
mv "$temporary" "$hostvars_file"
trap - EXIT INT TERM
unset auth_line auth_hash
ok "Geschuetzte Hostvars erstellt: $hostvars_file"
