#!/bin/sh
set -eu

OVERLAY_DIR="/opt/ghost-admin-overlay"
ADMIN_DIR="/var/lib/ghost/current/core/built/admin"
ASSET_DIR="${ADMIN_DIR}/assets"
INDEX_FILE="${ADMIN_DIR}/index.html"
HELPER_SRC="${OVERLAY_DIR}/ghost-image-editor-helper.js"
HELPER_TARGET="${ASSET_DIR}/ghost-image-editor-helper.js"
HELPER_MARKER="ghost-image-editor-helper-loader"
HELPER_TAG="<script id=\"${HELPER_MARKER}\" src=\"assets/ghost-image-editor-helper.js\"></script>"

if [ -f "$HELPER_SRC" ] && [ -f "$INDEX_FILE" ]; then
  mkdir -p "$ASSET_DIR"
  cp "$HELPER_SRC" "$HELPER_TARGET"

  if ! grep -q "$HELPER_MARKER" "$INDEX_FILE"; then
    tmp_file="$(mktemp)"

    if grep -qi '</body>' "$INDEX_FILE"; then
      sed "s#</body>#${HELPER_TAG}</body>#I" "$INDEX_FILE" > "$tmp_file"
    else
      cat "$INDEX_FILE" > "$tmp_file"
      printf '%s\n' "$HELPER_TAG" >> "$tmp_file"
    fi

    cat "$tmp_file" > "$INDEX_FILE"
    rm -f "$tmp_file"
  fi
fi

cd /var/lib/ghost
exec docker-entrypoint.sh node current/index.js
