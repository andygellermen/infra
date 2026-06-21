#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
GHOST_BACKUP_SCRIPT="$ROOT_DIR/scripts/ghost-backup.sh"
BACKUP_BASE_DIR="$ROOT_DIR/backups/ghost"

source "$ROOT_DIR/scripts/lib/error-notify.sh"
setup_error_notification "$(basename "$0")" "$ROOT_DIR" "$0 $*"

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/ghost-backup-all.sh [--output-dir <dir>] [--dry-run] [--continue-on-error]

Description:
  Erstellt Backups fuer alle Ghost-Domains im Stack.
  Die Domains werden automatisch ueber ansible/hostvars/*.yml erkannt
  (Dateien mit ghost_domain_db).

Beispiele:
  ./scripts/ghost-backup-all.sh
  ./scripts/ghost-backup-all.sh --output-dir /tmp/ghost-bulk-backups
  ./scripts/ghost-backup-all.sh --dry-run
  ./scripts/ghost-backup-all.sh --continue-on-error
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }

OUTPUT_DIR="$BACKUP_BASE_DIR"
DRY_RUN=0
CONTINUE_ON_ERROR=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output-dir)
      [[ $# -ge 2 ]] || die "Bitte Pfad nach --output-dir angeben."
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --continue-on-error)
      CONTINUE_ON_ERROR=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      die "Unbekannte Option: $1"
      ;;
  esac
done

[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"
[[ -x "$GHOST_BACKUP_SCRIPT" ]] || die "Script nicht ausfuehrbar: $GHOST_BACKUP_SCRIPT"

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

RUN_TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
mkdir -p "$OUTPUT_DIR"

info "Gefundene Ghost-Domains: ${#ghost_domains[@]}"
info "Output-Verzeichnis: $OUTPUT_DIR"
info "Run-Timestamp: $RUN_TIMESTAMP"
[[ "$DRY_RUN" -eq 1 ]] && info "Option aktiv: --dry-run"
[[ "$CONTINUE_ON_ERROR" -eq 1 ]] && info "Option aktiv: --continue-on-error"

failed=()
succeeded=0
created_files=()

run_backup() {
  local domain="$1"
  local output_file="$OUTPUT_DIR/$domain/ghost-backup-${domain}-${RUN_TIMESTAMP}.tar.gz"
  local cmd=("$GHOST_BACKUP_SCRIPT" "--create" "$domain" "--output" "$output_file")

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "Dry-Run: ${cmd[*]}"
    return 0
  fi

  info "Starte Backup: $domain"
  if "${cmd[@]}"; then
    ok "Backup erfolgreich: $domain"
    created_files+=("$output_file")
    return 0
  fi

  warn "Backup fehlgeschlagen: $domain"
  return 1
}

for domain in "${ghost_domains[@]}"; do
  if run_backup "$domain"; then
    succeeded=$((succeeded + 1))
    continue
  fi

  failed+=("$domain")
  if [[ "$CONTINUE_ON_ERROR" -eq 0 ]]; then
    warn "Abbruch nach erstem Fehler (nutze --continue-on-error fuer Batch-Fortsetzung)."
    break
  fi
done

if [[ "$DRY_RUN" -eq 1 ]]; then
  ok "Dry-Run abgeschlossen: $succeeded/${#ghost_domains[@]} Domains geplant."
  exit 0
fi

if [[ ${#failed[@]} -gt 0 ]]; then
  warn "Bulk-Backup mit Fehlern beendet."
  warn "Erfolgreich: $succeeded"
  warn "Fehlgeschlagen: ${#failed[@]}"
  for domain in "${failed[@]}"; do
    warn "  - $domain"
  done
  exit 1
fi

ok "Bulk-Backup abgeschlossen: $succeeded/${#ghost_domains[@]} Domains erfolgreich."
if [[ ${#created_files[@]} -gt 0 ]]; then
  info "Erstellte Backup-Dateien:"
  for file in "${created_files[@]}"; do
    info "  - $file"
  done
fi
