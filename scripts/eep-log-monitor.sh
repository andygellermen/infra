#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

source "$ROOT_DIR/scripts/lib/error-notify.sh"
setup_error_notification "eep-log-monitor" "$ROOT_DIR" "$0 $*"

python3 "$ROOT_DIR/scripts/eep-log-monitor.py" "$@"
