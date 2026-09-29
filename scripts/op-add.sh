#!/usr/bin/env bash
set -euo pipefail
umask 077
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec "${OP_PYTHON:-python3}" "$ROOT_DIR/scripts/lib/openproject_cli.py" add "$@"
