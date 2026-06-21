#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
GHOST_UPGRADE_SCRIPT="$ROOT_DIR/scripts/ghost-upgrade.sh"

source "$ROOT_DIR/scripts/lib/error-notify.sh"
source "$ROOT_DIR/scripts/lib/ghost-image-tag.sh"
setup_error_notification "$(basename "$0")" "$ROOT_DIR" "$0 $*"

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/ghost-upgrade-all.sh --version=<major|major.minor|major.minor.patch|latest> [--force-major-jump] [--dry-run] [--continue-on-error]

Description:
  Führt ein Bulk-Upgrade für alle Ghost-Domains im Stack aus.
  Die Domains werden automatisch über ansible/hostvars/*.yml erkannt
  (Dateien mit ghost_domain_db).

Beispiele:
  ./scripts/ghost-upgrade-all.sh --version=6
  ./scripts/ghost-upgrade-all.sh --version=6.18.2
  ./scripts/ghost-upgrade-all.sh --version=latest --continue-on-error
  ./scripts/ghost-upgrade-all.sh --version=5 --force-major-jump --dry-run
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }

TARGET_VERSION=""
FORCE_MAJOR_JUMP=0
DRY_RUN=0
CONTINUE_ON_ERROR=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version=*) TARGET_VERSION="${1#*=}"; shift ;;
    --version)
      die "Bitte --version=<major|major.minor|major.minor.patch|latest> verwenden."
      ;;
    --force-major-jump) FORCE_MAJOR_JUMP=1; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --continue-on-error) CONTINUE_ON_ERROR=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ -n "$TARGET_VERSION" ]] || die "Bitte --version=<major|major.minor|major.minor.patch|latest> angeben."
if [[ "$TARGET_VERSION" != "latest" ]] && ! [[ "$TARGET_VERSION" =~ ^[0-9]+(\.[0-9]+){0,2}$ ]]; then
  die "Ungültige Zielversion '$TARGET_VERSION'. Erlaubt sind 'latest', Major (z. B. 6), Minor (z. B. 6.18) oder Patch (z. B. 6.18.2)."
fi

[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"
[[ -x "$GHOST_UPGRADE_SCRIPT" ]] || die "Script nicht ausführbar: $GHOST_UPGRADE_SCRIPT"

cd "$ROOT_DIR"

HOSTVAR_FILES=()
while IFS= read -r hostvar_file; do
  HOSTVAR_FILES+=("$hostvar_file")
done < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' | sort)
[[ ${#HOSTVAR_FILES[@]} -gt 0 ]] || die "Keine Hostvars-Dateien in $HOSTVARS_DIR gefunden."

ghost_domains=()
for file in "${HOSTVAR_FILES[@]}"; do
  if grep -q '^ghost_domain_db:' "$file"; then
    ghost_domains+=("$(basename "$file" .yml)")
  fi
done

[[ ${#ghost_domains[@]} -gt 0 ]] || die "Keine Ghost-Domains gefunden (Marker: ghost_domain_db)."

info "Gefundene Ghost-Domains: ${#ghost_domains[@]}"
info "Zielversion: $TARGET_VERSION"
[[ "$FORCE_MAJOR_JUMP" -eq 1 ]] && info "Option aktiv: --force-major-jump"
[[ "$DRY_RUN" -eq 1 ]] && info "Option aktiv: --dry-run"
[[ "$CONTINUE_ON_ERROR" -eq 1 ]] && info "Option aktiv: --continue-on-error"

info "Pruefe Verfuegbarkeit von ghost:${TARGET_VERSION}"
validate_ghost_image_tag_or_die "$TARGET_VERSION" "Zielversion"
export INFRA_GHOST_TAG_PREVALIDATED="$TARGET_VERSION"

failed=()
succeeded=0

run_upgrade() {
  local domain="$1"
  local cmd=("$GHOST_UPGRADE_SCRIPT" "$domain" "--version=$TARGET_VERSION")

  [[ "$FORCE_MAJOR_JUMP" -eq 1 ]] && cmd+=("--force-major-jump")
  [[ "$DRY_RUN" -eq 1 ]] && cmd+=("--dry-run")

  info "Starte Upgrade: $domain"
  if "${cmd[@]}"; then
    ok "Upgrade erfolgreich: $domain"
    return 0
  fi

  warn "Upgrade fehlgeschlagen: $domain"
  return 1
}

for domain in "${ghost_domains[@]}"; do
  if run_upgrade "$domain"; then
    succeeded=$((succeeded + 1))
    continue
  fi

  failed+=("$domain")
  if [[ "$CONTINUE_ON_ERROR" -eq 0 ]]; then
    warn "Abbruch nach erstem Fehler (nutze --continue-on-error für Batch-Fortsetzung)."
    break
  fi
done

if [[ ${#failed[@]} -gt 0 ]]; then
  warn "Bulk-Upgrade mit Fehlern beendet."
  warn "Erfolgreich: $succeeded"
  warn "Fehlgeschlagen: ${#failed[@]}"
  for domain in "${failed[@]}"; do
    warn "  - $domain"
  done
  exit 1
fi

ok "Bulk-Upgrade abgeschlossen: $succeeded/${#ghost_domains[@]} Domains erfolgreich."
