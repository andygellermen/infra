#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MONTHLY_ROOT="${ROOT_DIR}/backups/monthly"
GHOST_OUTPUT_DIR="${MONTHLY_ROOT}/ghost"
WP_OUTPUT_DIR="${MONTHLY_ROOT}/wordpress"
KEEP_COUNT="${KEEP_COUNT:-3}"

source "$ROOT_DIR/scripts/lib/error-notify.sh"
setup_error_notification "$(basename "$0")" "$ROOT_DIR" "$0 $*"

usage() {
  cat <<'EOF'
Usage: ./scripts/monthly-site-backup.sh [--check-only] [--keep <count>]

Description:
  Erstellt monatliche Sammel-Backups fuer alle Ghost- und WordPress-Instanzen
  und behaelt pro Domain nur die neuesten <count> Archive.
EOF
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }

CHECK_ONLY=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only)
      CHECK_ONLY=1
      shift
      ;;
    --keep)
      [[ $# -ge 2 ]] || die "Bitte Anzahl nach --keep angeben."
      KEEP_COUNT="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "Unbekannte Option: $1"
      ;;
  esac
done

[[ "$KEEP_COUNT" =~ ^[0-9]+$ ]] || die "--keep muss eine Zahl sein"
(( KEEP_COUNT >= 1 )) || die "--keep muss mindestens 1 sein"

mkdir -p "$GHOST_OUTPUT_DIR" "$WP_OUTPUT_DIR"

prune_backup_tree() {
  local root="$1" pattern="$2"
  [[ -d "$root" ]] || return 0

  while IFS= read -r domain_dir; do
    mapfile -t files < <(find "$domain_dir" -maxdepth 1 -type f -name "$pattern" | sort -r)
    (( ${#files[@]} > KEEP_COUNT )) || continue
    for file in "${files[@]:KEEP_COUNT}"; do
      if [[ "$CHECK_ONLY" -eq 1 ]]; then
        info "würde altes Backup löschen: $file"
      else
        rm -f "$file"
        info "altes Backup gelöscht: $file"
      fi
    done
  done < <(find "$root" -mindepth 1 -maxdepth 1 -type d | sort)
}

ghost_rc=0
wp_rc=0

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  info "Dry-Run: Ghost-Monatsbackups"
  "$ROOT_DIR/scripts/ghost-backup-all.sh" --output-dir "$GHOST_OUTPUT_DIR" --dry-run --continue-on-error || ghost_rc=$?
  info "Dry-Run: WordPress-Monatsbackups"
  "$ROOT_DIR/scripts/wp-backup-all.sh" --output-dir "$WP_OUTPUT_DIR" --dry-run --continue-on-error || wp_rc=$?
else
  info "Starte Ghost-Monatsbackups"
  "$ROOT_DIR/scripts/ghost-backup-all.sh" --output-dir "$GHOST_OUTPUT_DIR" --continue-on-error || ghost_rc=$?
  info "Starte WordPress-Monatsbackups"
  "$ROOT_DIR/scripts/wp-backup-all.sh" --output-dir "$WP_OUTPUT_DIR" --continue-on-error || wp_rc=$?
fi

prune_backup_tree "$GHOST_OUTPUT_DIR" 'ghost-backup-*.tar.gz'
prune_backup_tree "$WP_OUTPUT_DIR" 'wp-backup-*.tar.gz'

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only für Monatsbackups abgeschlossen"
  exit 0
fi

if (( ghost_rc != 0 || wp_rc != 0 )); then
  warn "Monatsbackup beendet mit Fehlern (Ghost RC=$ghost_rc, WordPress RC=$wp_rc)"
  exit 1
fi

ok "Monatsbackups für Ghost und WordPress abgeschlossen"
