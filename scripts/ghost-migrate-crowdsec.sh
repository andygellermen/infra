#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
REDEPLOY_SCRIPT="$ROOT_DIR/scripts/ghost-redeploy.sh"
CHECK_ONLY=0
RESTART_TRAEFIK=0

usage() {
  cat <<USAGE
Usage:
  $0 [--check-only] [--restart-traefik]

Beschreibung:
  Migriert bestehende Ghost-Instanzen auf CrowdSec-Default-Middlewares.
  - ergänzt fehlende ghost_traefik_middleware_* Keys in allen Hostvars
  - führt anschließend pro Domain ghost-redeploy.sh aus

Optionen:
  --check-only       Zeigt geplante Änderungen, schreibt nichts und redeployt nicht.
  --restart-traefik  Startet Traefik einmalig nach allen erfolgreichen Redeploys neu.
  --help, -h         Hilfe anzeigen.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }

if [[ ${1:-} == "--help" || ${1:-} == "-h" ]]; then
  usage
  exit 0
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=1; shift ;;
    --restart-traefik) RESTART_TRAEFIK=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"
[[ -x "$REDEPLOY_SCRIPT" ]] || die "Redeploy-Skript fehlt/ist nicht ausführbar: $REDEPLOY_SCRIPT"

mapfile -t HOSTVAR_FILES < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' ! -name '*.example.yml' | sort)
if [[ ${#HOSTVAR_FILES[@]} -eq 0 ]]; then
  if [[ "$CHECK_ONLY" -eq 1 ]]; then
    info "Keine Hostvars-Dateien gefunden in $HOSTVARS_DIR (check-only ohne Änderungen abgeschlossen)."
    exit 0
  fi
  die "Keine Hostvars-Dateien gefunden in $HOSTVARS_DIR"
fi

DEFAULTS=(
  'ghost_traefik_middleware_default: ""'
  'ghost_traefik_middleware_admin: "crowdsec-admin@docker"'
  'ghost_traefik_middleware_api: "crowdsec-api@docker"'
  'ghost_traefik_middleware_dotghost: "crowdsec-api@docker"'
  'ghost_traefik_middleware_members_api: "crowdsec-api@docker"'
)

migrated_domains=()
skipped_domains=()
failed_domains=()

for hostvars in "${HOSTVAR_FILES[@]}"; do
  domain="$(basename "$hostvars" .yml)"
  [[ "$domain" == "templates" ]] && continue

  missing=0
  missing_lines=()
  for line in "${DEFAULTS[@]}"; do
    key="${line%%:*}"
    if ! grep -qE "^${key}:" "$hostvars"; then
      missing=1
      missing_lines+=("$line")
    fi
  done

  if [[ "$missing" -eq 1 ]]; then
    if [[ "$CHECK_ONLY" -eq 1 ]]; then
      info "[check-only] $domain: es fehlen ${#missing_lines[@]} CrowdSec-Key(s)"
    else
      printf '\n# CrowdSec middleware defaults\n' >> "$hostvars"
      for line in "${missing_lines[@]}"; do
        printf '%s\n' "$line" >> "$hostvars"
      done
      ok "$domain: CrowdSec-Defaults ergänzt"
    fi
  else
    info "$domain: CrowdSec-Defaults bereits vorhanden"
  fi

  if [[ "$CHECK_ONLY" -eq 1 ]]; then
    skipped_domains+=("$domain")
    continue
  fi

  info "$domain: starte Redeploy"
  if "$REDEPLOY_SCRIPT" "$domain"; then
    migrated_domains+=("$domain")
    ok "$domain: Redeploy erfolgreich"
  else
    failed_domains+=("$domain")
    echo "⚠️  $domain: Redeploy fehlgeschlagen" >&2
  fi
done

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only abgeschlossen. Keine Änderungen geschrieben, kein Redeploy ausgeführt."
  exit 0
fi

if [[ "$RESTART_TRAEFIK" -eq 1 ]]; then
  info "Starte Traefik einmalig neu"
  docker restart traefik >/dev/null
  ok "Traefik neugestartet"
fi

echo
ok "Migration abgeschlossen"
echo "  Erfolgreich migriert: ${#migrated_domains[@]}"
echo "  Fehlgeschlagen: ${#failed_domains[@]}"

if [[ ${#failed_domains[@]} -gt 0 ]]; then
  echo "Fehlerhafte Domains: ${failed_domains[*]}" >&2
  exit 1
fi
