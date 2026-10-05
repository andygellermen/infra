#!/usr/bin/env bash
set -euo pipefail
umask 077
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
exec "${OPAI_BRIDGE_PYTHON:-python3}" "$ROOT_DIR/scripts/lib/openproject_ai_bridge_cli.py" status "$@"
