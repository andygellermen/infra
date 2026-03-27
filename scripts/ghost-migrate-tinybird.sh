#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
REDEPLOY_SCRIPT="$ROOT_DIR/scripts/ghost-redeploy.sh"
CHECK_ONLY=0
REDEPLOY=1
RESTART_TRAEFIK=0
ROTATE_TOKENS=0

usage() {
  cat <<USAGE
Usage:
  $0 [--check-only] [--no-redeploy] [--restart-traefik] [--rotate-tokens]

Beschreibung:
  Migriert bestehende Ghost-Instanzen auf Tinybird-Defaults.
  - ergänzt fehlende tinybird_* Keys in allen Hostvars
  - generiert pro Domain ein Token (wenn fehlend oder mit --rotate-tokens)
  - führt anschließend optional pro Domain ghost-redeploy.sh aus

Optionen:
  --check-only       Zeigt geplante Änderungen, schreibt nichts und redeployt nicht.
  --no-redeploy      Schreibt Hostvars, aber kein Redeploy.
  --restart-traefik  Startet Traefik einmalig nach allen erfolgreichen Redeploys neu.
  --rotate-tokens    Ersetzt vorhandene tinybird_token Werte mit neuen Tokens.
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
    --no-redeploy) REDEPLOY=0; shift ;;
    --restart-traefik) RESTART_TRAEFIK=1; shift ;;
    --rotate-tokens) ROTATE_TOKENS=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"
[[ -x "$REDEPLOY_SCRIPT" ]] || die "Redeploy-Skript fehlt/ist nicht ausführbar: $REDEPLOY_SCRIPT"
command -v openssl >/dev/null 2>&1 || die "Tool fehlt: openssl"

mapfile -t HOSTVAR_FILES < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' ! -name '*.example.yml' | sort)
if [[ ${#HOSTVAR_FILES[@]} -eq 0 ]]; then
  if [[ "$CHECK_ONLY" -eq 1 ]]; then
    info "Keine Hostvars-Dateien gefunden in $HOSTVARS_DIR (check-only ohne Änderungen abgeschlossen)."
    exit 0
  fi
  die "Keine Hostvars-Dateien gefunden in $HOSTVARS_DIR"
fi

migrated_domains=()
failed_domains=()
changed_domains=()

for hostvars in "${HOSTVAR_FILES[@]}"; do
  domain="$(basename "$hostvars" .yml)"
  [[ "$domain" == "templates" ]] && continue

  datasource="ghost_pageviews_$(echo "$domain" | tr '.-' '__')"
  generated_token="$(openssl rand -hex 24)"

  defaults=(
    'tinybird_enabled: true'
    'tinybird_api_url: "https://api.tinybird.co"'
    'tinybird_workspace: "main"'
    "tinybird_datasource: \"${datasource}\""
    "tinybird_token: \"${generated_token}\""
    'tinybird_events_endpoint: "/v0/events?name=pageviews"'
  )

  missing=0
  missing_lines=()
  for line in "${defaults[@]}"; do
    key="${line%%:*}"
    if ! grep -qE "^${key}:" "$hostvars"; then
      missing=1
      missing_lines+=("$line")
    fi
  done

  changed=0
  if [[ "$missing" -eq 1 ]]; then
    changed=1
    if [[ "$CHECK_ONLY" -eq 1 ]]; then
      info "[check-only] $domain: es fehlen ${#missing_lines[@]} Tinybird-Key(s)"
    else
      printf '\n# Tinybird analytics defaults\n' >> "$hostvars"
      for line in "${missing_lines[@]}"; do
        printf '%s\n' "$line" >> "$hostvars"
      done
      ok "$domain: Tinybird-Defaults ergänzt"
    fi
  else
    info "$domain: Tinybird-Defaults bereits vorhanden"
  fi

  if [[ "$ROTATE_TOKENS" -eq 1 ]]; then
    changed=1
    if [[ "$CHECK_ONLY" -eq 1 ]]; then
      info "[check-only] $domain: tinybird_token würde rotiert"
    else
      sed -i -E "s#^tinybird_token:.*#tinybird_token: \"${generated_token}\"#" "$hostvars"
      ok "$domain: tinybird_token rotiert"
    fi
  fi

  if [[ "$changed" -eq 1 ]]; then
    changed_domains+=("$domain")
  fi

  if [[ "$CHECK_ONLY" -eq 1 || "$REDEPLOY" -eq 0 || "$changed" -eq 0 ]]; then
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

if [[ "$RESTART_TRAEFIK" -eq 1 && "$REDEPLOY" -eq 1 && ${#changed_domains[@]} -gt 0 ]]; then
  info "Starte Traefik einmalig neu"
  docker restart infra-traefik >/dev/null
  ok "Traefik neugestartet"
fi

echo
ok "Tinybird-Migration abgeschlossen"
echo "  Domains mit Hostvars-Änderungen: ${#changed_domains[@]}"
echo "  Erfolgreiche Redeploys: ${#migrated_domains[@]}"
echo "  Fehlgeschlagen: ${#failed_domains[@]}"

if [[ ${#failed_domains[@]} -gt 0 ]]; then
  echo "Fehlerhafte Domains: ${failed_domains[*]}" >&2
  exit 1
fi
