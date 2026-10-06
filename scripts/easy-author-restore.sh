#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRV_ROOT="${EASY_AUTHOR_SRV_ROOT:-/srv/easy-author}"
DOCKER_BIN="${EASY_AUTHOR_DOCKER_BIN:-docker}"
BACKUP_BIN="${EASY_AUTHOR_BACKUP_BIN:-$ROOT_DIR/scripts/easy-author-backup.sh}"
die(){ echo "FEHLER: $*" >&2; exit 1; }
domain=""; archive=""; confirmation=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --confirm=*) confirmation="${1#*=}" ;;
    --help|-h) echo "Usage: $0 <domain> <archive> --confirm=<domain>"; exit 0 ;;
    *) if [[ -z "$domain" ]]; then domain="$1"; elif [[ -z "$archive" ]]; then archive="$1"; else die "Zu viele Argumente."; fi ;;
  esac
  shift
done
[[ "$domain" =~ ^[a-z0-9]+([.-][a-z0-9]+)*$ ]] || die "Ungueltige Domain."
[[ "$confirmation" == "$domain" ]] || die "Restore erfordert --confirm=$domain."
[[ -f "$archive" ]] || die "Archiv fehlt."
[[ "$(basename "$archive")" == easy-author-"$domain"-*.tar.gz ]] || die "Archivname passt nicht zur Instanz."
while IFS= read -r member; do
  [[ "$member" != /* && "$member" != *'../'* && "$member" != '..' ]] || die "Unsicherer Archivpfad: $member"
done < <(tar -tf "$archive")
tar -tf "$archive" | grep -qx 'data/easy-author.sqlite' || die "SQLite fehlt im Archiv."
tar -tf "$archive" | grep -Eq '^data/library(/|$)' || die "Bibliothek fehlt im Archiv."
instance="$SRV_ROOT/$domain"; data="$instance/data"; stage="$instance/.restore-$$"; previous="$instance/.restore-previous-$$"
mkdir -p "$instance"
"$BACKUP_BIN" "$domain" --allow-empty >/dev/null || die "Sicherheitsbackup fehlgeschlagen."
rm -rf "$stage"; mkdir -m 0700 "$stage"
cleanup(){ rm -rf "$stage"; }
trap cleanup EXIT INT TERM
tar -C "$stage" -xzf "$archive"
[[ -f "$stage/data/easy-author.sqlite" && -d "$stage/data/library" ]] || die "Extrahierter Datenstand ist unvollstaendig."
id="${domain//./-}"; api="easy-author-$id-api"; web="easy-author-$id-web"; running=()
for container in "$web" "$api"; do
  if "$DOCKER_BIN" inspect "$container" >/dev/null 2>&1 && [[ "$("$DOCKER_BIN" inspect -f '{{.State.Running}}' "$container")" == true ]]; then
    "$DOCKER_BIN" stop "$container" >/dev/null; running+=("$container")
  fi
done
if [[ -e "$data" ]]; then mv "$data" "$previous"; fi
if ! mv "$stage/data" "$data"; then
  [[ ! -e "$previous" ]] || mv "$previous" "$data"
  for container in "${running[@]-}"; do [[ -z "$container" ]] || "$DOCKER_BIN" start "$container" >/dev/null || true; done
  die "Atomarer Datenaustausch fehlgeschlagen."
fi
rm -rf "$previous"; chmod -R u+rwX,go-rwx "$data"
for container in "${running[@]-}"; do [[ -z "$container" ]] || "$DOCKER_BIN" start "$container" >/dev/null; done
trap - EXIT INT TERM; rm -rf "$stage"
echo "OK: Restore fuer $domain abgeschlossen."
