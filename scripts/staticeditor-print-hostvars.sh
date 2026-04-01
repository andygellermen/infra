#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: ./scripts/staticeditor-print-hostvars.sh <domain> [allowed-email1 allowed-email2 ...] [--login-domain=<domain>] [--repo-root=<path>] [--static-root=<path>] [--backup-root=<path>] [--cookie-secret=<secret>]

Description:
  Gibt einen fertigen Hostvars-Block fuer den Static Inline Editor aus, den du in eine bestehende Domain-Datei unter ansible/hostvars/<domain>.yml einkopieren kannst.
USAGE
}

die(){ echo "❌ Fehler: $*" >&2; exit 1; }

[[ $# -ge 1 ]] || { usage; exit 1; }

LOGIN_DOMAIN=""
REPO_ROOT=""
STATIC_ROOT=""
BACKUP_ROOT=""
COOKIE_SECRET=""
args=()

for arg in "$@"; do
  case "$arg" in
    --login-domain=*) LOGIN_DOMAIN="${arg#*=}" ;;
    --repo-root=*) REPO_ROOT="${arg#*=}" ;;
    --static-root=*) STATIC_ROOT="${arg#*=}" ;;
    --backup-root=*) BACKUP_ROOT="${arg#*=}" ;;
    --cookie-secret=*) COOKIE_SECRET="${arg#*=}" ;;
    --help|-h) usage; exit 0 ;;
    *) args+=("$arg") ;;
  esac
done

DOMAIN="${args[0]}"
shifted=("${args[@]:1}")

if [[ -z "$COOKIE_SECRET" ]] && command -v openssl >/dev/null 2>&1; then
  COOKIE_SECRET="$(openssl rand -hex 24)"
fi

LOGIN_DOMAIN="${LOGIN_DOMAIN:-bearbeitung.${DOMAIN}}"
STATIC_ROOT="${STATIC_ROOT:-/srv/static/${DOMAIN}}"
BACKUP_ROOT="${BACKUP_ROOT:-/srv/static-backups/${DOMAIN}}"
REPO_ROOT="${REPO_ROOT:-$STATIC_ROOT}"

echo "# --- Static Inline Editor ---"
echo "static_editor_enabled: true"
echo "static_editor_login_domain: \"${LOGIN_DOMAIN}\""
echo "static_editor_aliases: []"
echo "static_editor_static_root: \"${STATIC_ROOT}\""
echo "static_editor_backup_root: \"${BACKUP_ROOT}\""
echo "static_editor_repo_root: \"${REPO_ROOT}\""
echo "static_editor_allowed_emails:"
for email in "${shifted[@]}"; do
  echo "  - \"${email}\""
done
echo "static_editor_cookie_secret: \"${COOKIE_SECRET}\""
echo "static_editor_start_path: \"/index.html\""
echo "static_editor_main_selector: \"main\""
echo "static_editor_allowed_block_tags: \"h1,h2,h3,h4,h5,p,ul,ol,li\""
echo "static_editor_allowed_inline_tags: \"strong,em,a,br\""
echo "static_editor_git_push_on_save: false"
echo "static_editor_git_remote: \"origin\""
echo "static_editor_git_branch: \"\""
echo "static_editor_git_author_name: \"Static Inline Editor\""
echo "static_editor_smtp_host: \"\""
echo "static_editor_smtp_port: 587"
echo "static_editor_smtp_username: \"\""
echo "static_editor_smtp_password: \"\""
echo "static_editor_smtp_from_email: \"\""
echo "static_editor_smtp_from_name: \"Static Inline Editor\""
