#!/usr/bin/env bash
# Run on a disposable Linux Docker host, never on an existing production domain.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
exec "${OP_PYTHON:-python3}" "$ROOT_DIR/apps/openproject/tests/integration_roundtrip.py" "$@"
