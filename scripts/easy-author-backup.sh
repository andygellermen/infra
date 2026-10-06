#!/usr/bin/env bash
set -euo pipefail

SRV_ROOT="${EASY_AUTHOR_SRV_ROOT:-/srv/easy-author}"
DOCKER_BIN="${EASY_AUTHOR_DOCKER_BIN:-docker}"
die(){ echo "FEHLER: $*" >&2; exit 1; }
domain=""; retention=7; allow_empty=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --retention=*) retention="${1#*=}" ;;
    --allow-empty) allow_empty=1 ;;
    --help|-h) echo "Usage: $0 <domain> [--retention=<n>] [--allow-empty]"; exit 0 ;;
    *) [[ -z "$domain" ]] || die "Nur eine Domain ist erlaubt."; domain="$1" ;;
  esac
  shift
done
[[ "$domain" =~ ^[a-z0-9]+([.-][a-z0-9]+)*$ ]] || die "Ungueltige Domain."
[[ "$retention" =~ ^[1-9][0-9]*$ ]] || die "Ungueltige Aufbewahrungszahl."
instance="$SRV_ROOT/$domain"; data="$instance/data"; backups="$instance/backups"
if [[ ! -d "$data" || -z "$(find "$data" -mindepth 1 -print -quit 2>/dev/null)" ]]; then
  [[ "$allow_empty" -eq 1 ]] || die "Keine bestehenden Daten zum Sichern."
  exit 0
fi
mkdir -p "$backups"; chmod 0700 "$backups"
id="${domain//./-}"; api="easy-author-$id-api"; stopped=0; partial=""
restart(){ local status=$?; [[ "$stopped" -eq 0 ]] || "$DOCKER_BIN" start "$api" >/dev/null 2>&1 || true; [[ -z "$partial" ]] || rm -f "$partial"; exit "$status"; }
trap restart EXIT INT TERM
if "$DOCKER_BIN" inspect "$api" >/dev/null 2>&1 && [[ "$("$DOCKER_BIN" inspect -f '{{.State.Running}}' "$api")" == true ]]; then
  "$DOCKER_BIN" stop "$api" >/dev/null; stopped=1
fi
stamp="$(date -u +%Y%m%dT%H%M%SZ)"; final="$backups/easy-author-$domain-$stamp.tar.gz"; partial="$final.partial"
tar -C "$instance" -czf "$partial" data
tar -tf "$partial" >/dev/null
chmod 0600 "$partial"; mv "$partial" "$final"; partial=""
if [[ "$stopped" -eq 1 ]]; then "$DOCKER_BIN" start "$api" >/dev/null; stopped=0; fi
find "$backups" -maxdepth 1 -type f -name "easy-author-$domain-*.tar.gz" -print \
  | sort -r | sed -n "$((retention + 1)),\$p" \
  | while IFS= read -r expired; do rm -f -- "$expired"; done
trap - EXIT INT TERM
printf '%s\n' "$final"
