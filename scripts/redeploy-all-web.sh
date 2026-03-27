#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
GHOST_REDEPLOY="$ROOT_DIR/scripts/ghost-redeploy.sh"
WP_REDEPLOY="$ROOT_DIR/scripts/wp-redeploy.sh"

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/redeploy-all-web.sh [--check-only] [--only=all|ghost|wp] [--continue-on-error]

Description:
  Redeployt alle Web-Container (Ghost + WordPress) basierend auf Hostvars.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }

CHECK_ONLY=0
ONLY="all"
CONTINUE_ON_ERROR=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=1; shift ;;
    --only=*) ONLY="${1#*=}"; shift ;;
    --continue-on-error) CONTINUE_ON_ERROR=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ "$ONLY" =~ ^(all|ghost|wp)$ ]] || die "--only muss all|ghost|wp sein."
[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"
[[ -x "$GHOST_REDEPLOY" ]] || die "Script nicht ausführbar: $GHOST_REDEPLOY"
[[ -x "$WP_REDEPLOY" ]] || die "Script nicht ausführbar: $WP_REDEPLOY"

mapfile -t HOSTVAR_FILES < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' | sort)
[[ ${#HOSTVAR_FILES[@]} -gt 0 ]] || die "Keine Hostvars-Dateien in $HOSTVARS_DIR gefunden."

ghost_domains=()
wp_domains=()

for file in "${HOSTVAR_FILES[@]}"; do
  domain="$(basename "$file" .yml)"
  grep -q '^ghost_domain_db:' "$file" && ghost_domains+=("$domain")
  grep -q '^wp_domain_db:' "$file" && wp_domains+=("$domain")
done

run_redeploy() {
  local kind="$1" domain="$2" script="$3"
  local cmd=("$script" "$domain")
  [[ "$CHECK_ONLY" -eq 1 ]] && cmd+=("--check-only")

  info "Starte ${kind}-Redeploy: ${domain}"
  if "${cmd[@]}"; then
    ok "${kind} erfolgreich: ${domain}"
    return 0
  fi

  warn "${kind} fehlgeschlagen: ${domain}"
  return 1
}

failed=()

if [[ "$ONLY" == "all" || "$ONLY" == "ghost" ]]; then
  info "Ghost-Domains: ${#ghost_domains[@]}"
  for d in "${ghost_domains[@]}"; do
    if ! run_redeploy "Ghost" "$d" "$GHOST_REDEPLOY"; then
      failed+=("ghost:${d}")
      [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || exit 1
    fi
  done
fi

if [[ "$ONLY" == "all" || "$ONLY" == "wp" ]]; then
  info "WordPress-Domains: ${#wp_domains[@]}"
  for d in "${wp_domains[@]}"; do
    if ! run_redeploy "WordPress" "$d" "$WP_REDEPLOY"; then
      failed+=("wp:${d}")
      [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || exit 1
    fi
  done
fi

if [[ ${#failed[@]} -gt 0 ]]; then
  warn "Redeploy mit Fehlern beendet:"
  for entry in "${failed[@]}"; do
    warn "  - $entry"
  done
  exit 1
fi

ok "Alle gewünschten Redeploys abgeschlossen."
