#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

source "$ROOT_DIR/scripts/lib/error-notify.sh"
setup_error_notification "$(basename "$0")" "$ROOT_DIR" "$0 $*"

exec python3 "$ROOT_DIR/scripts/wp-update-report.py" "$@"
