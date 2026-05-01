#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR=""
CLEAN=0

usage() {
  cat <<USAGE
Usage:
  $0 <domain> [--output-dir /pfad] [--clean]

Beschreibung:
  Kopiert die offiziellen Tinybird-Projektdateien aus dem laufenden Ghost-Container
  in ein lokales Arbeitsverzeichnis. Das ist die Grundlage fuer den anschliessenden
  Tinybird-Cloud-Deploy per CLI.

Optionen:
  --output-dir DIR  Zielverzeichnis fuer die kopierten Tinybird-Dateien.
                    Default: ./data/ghost-native-analytics/<domain>/tinybird
  --clean           Loescht ein bestehendes Zielverzeichnis vor dem Sync.
  --help, -h        Hilfe anzeigen.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

if [[ ${1:-} == "--help" || ${1:-} == "-h" ]]; then
  usage
  exit 0
fi

DOMAIN="${1:-}"
[[ -n "$DOMAIN" ]] || { usage; exit 1; }
shift || true

while [[ $# -gt 0 ]]; do
  case "$1" in
    --output-dir)
      [[ $# -ge 2 ]] || die "--output-dir braucht einen Pfad"
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --output-dir=*)
      OUTPUT_DIR="${1#*=}"
      shift
      ;;
    --clean)
      CLEAN=1
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

require_cmd docker

CONTAINER="ghost-${DOMAIN//./-}"
SOURCE_DIR="/var/lib/ghost/current/core/server/data/tinybird"
OUTPUT_DIR="${OUTPUT_DIR:-$ROOT_DIR/data/ghost-native-analytics/$DOMAIN/tinybird}"

docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER" || die "Ghost-Container nicht gefunden: $CONTAINER"

mkdir -p "$(dirname "$OUTPUT_DIR")"
if [[ "$CLEAN" -eq 1 && -d "$OUTPUT_DIR" ]]; then
  info "Bereinige bestehendes Zielverzeichnis: $OUTPUT_DIR"
  rm -rf "$OUTPUT_DIR"
fi
mkdir -p "$OUTPUT_DIR"

info "Kopiere Tinybird-Projektdateien aus $CONTAINER:$SOURCE_DIR"
docker cp "$CONTAINER:$SOURCE_DIR/." "$OUTPUT_DIR" >/dev/null 2>&1 || die "Tinybird-Dateien konnten nicht aus dem Ghost-Container kopiert werden. Pruefe Ghost 6+ und den Pfad $SOURCE_DIR."

find "$OUTPUT_DIR" -mindepth 1 -print -quit >/dev/null 2>&1 || die "Sync abgeschlossen, aber das Zielverzeichnis ist leer: $OUTPUT_DIR"

ok "Tinybird-Projekt synchronisiert: $OUTPUT_DIR"
