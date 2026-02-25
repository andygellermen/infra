#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INVENTORY="$ROOT_DIR/ansible/inventory/hosts.ini"
PLAYBOOK="$ROOT_DIR/ansible/playbooks/deploy-ghost.yml"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
CHECK_ONLY=0
RESTART_TRAEFIK=0

usage() {
  cat <<USAGE
Usage:
  $0 <domain> [--check-only] [--restart-traefik]

Beschreibung:
  Validiert hostvars + DNS A-Records (Domain + Aliase) und führt anschließend
  den gezielten Ghost-Redeploy aus.

Optionen:
  --check-only       Nur Validierung, kein Redeploy.
  --restart-traefik  Restart von Traefik nach erfolgreichem Redeploy.
  --help, -h         Hilfe anzeigen.
USAGE
}

die() { echo "❌ $*" >&2; exit 1; }
info() { echo "ℹ️  $*"; }
ok() { echo "✅ $*"; }
require_cmd() { command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

check_repo_for_conflict_markers() {
  local matches
  matches="$(rg -n "^(<<<<<<<|=======|>>>>>>>)" "$ROOT_DIR/ansible" "$ROOT_DIR/scripts" || true)"

  if [[ -n "$matches" ]]; then
    echo "❌ Merge-Konfliktmarker im Repository gefunden:" >&2
    echo "$matches" >&2
    die "Bitte Konflikte auflösen, committen und Redeploy erneut starten."
  fi
}

normalize_domain() {
  local d="$1"
  if [[ "$d" =~ ^[a-zA-Z0-9.-]+$ ]]; then
    printf '%s\n' "$d"
    return
  fi

  if command -v idn >/dev/null 2>&1; then
    idn --quiet --uts46 "$d"
  else
    die "Domain enthält Nicht-ASCII-Zeichen, aber 'idn' fehlt. Installiere: sudo apt install idn"
  fi
}

extract_scalar() {
  local key="$1" file="$2"
  awk -F': ' -v k="$key" '$1==k {gsub(/"/,"",$2); gsub(/[[:space:]]+$/, "", $2); print $2; exit}' "$file"
}

extract_aliases() {
  local file="$1"
  awk '
    /^traefik:/ {in_traefik=1; next}
    in_traefik && /^[^[:space:]]/ {in_traefik=0}
    in_traefik && /^[[:space:]]{2}aliases:[[:space:]]*$/ {in_aliases=1; next}
    in_aliases && /^[[:space:]]{4}-[[:space:]]+/ {
      line=$0
      sub(/^[[:space:]]{4}-[[:space:]]+/, "", line)
      gsub(/[[:space:]]+$/, "", line)
      print line
      next
    }
    in_aliases && !/^[[:space:]]{4}-[[:space:]]+/ {in_aliases=0}
  ' "$file"
}

resolve_a_record() {
  local domain="$1"
  dig +short A "$domain" | head -n1
}

verify_a_record_matches_host() {
  local domain="$1" host_ip="$2"
  local dns_ip
  dns_ip="$(resolve_a_record "$domain")"

  [[ -n "$dns_ip" ]] || die "Kein A-Record für ${domain} gefunden."

  if [[ "$dns_ip" != "$host_ip" ]]; then
    die "DNS-Fehler: ${domain} zeigt auf ${dns_ip}, erwartet wird ${host_ip}."
  fi

  ok "DNS OK: ${domain} -> ${dns_ip}"
}

if [[ ${1:-} == "--help" || ${1:-} == "-h" ]]; then
  usage
  exit 0
fi

DOMAIN="${1:-}"
[[ -n "$DOMAIN" ]] || { usage; exit 1; }
shift || true

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=1; shift ;;
    --restart-traefik) RESTART_TRAEFIK=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

require_cmd dig
require_cmd curl

DOMAIN="$(normalize_domain "$DOMAIN")"
HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars nicht gefunden: $HOSTVARS_FILE"
[[ -f "$PLAYBOOK" ]] || die "Playbook nicht gefunden: $PLAYBOOK"
[[ -f "$INVENTORY" ]] || die "Inventory nicht gefunden: $INVENTORY"

check_repo_for_conflict_markers

HOSTVARS_DOMAIN="$(extract_scalar domain "$HOSTVARS_FILE")"
DB_NAME="$(extract_scalar ghost_domain_db "$HOSTVARS_FILE")"
DB_USER="$(extract_scalar ghost_domain_usr "$HOSTVARS_FILE")"
DB_PASS="$(extract_scalar ghost_domain_pwd "$HOSTVARS_FILE")"

[[ -n "$HOSTVARS_DOMAIN" ]] || die "Pflichtwert 'domain' fehlt in $HOSTVARS_FILE"
[[ -n "$DB_NAME" ]] || die "Pflichtwert 'ghost_domain_db' fehlt in $HOSTVARS_FILE"
[[ -n "$DB_USER" ]] || die "Pflichtwert 'ghost_domain_usr' fehlt in $HOSTVARS_FILE"
[[ -n "$DB_PASS" ]] || die "Pflichtwert 'ghost_domain_pwd' fehlt in $HOSTVARS_FILE"

HOSTVARS_DOMAIN="$(normalize_domain "$HOSTVARS_DOMAIN")"
if [[ "$HOSTVARS_DOMAIN" != "$DOMAIN" ]]; then
  die "Domain-Mismatch: Argument=${DOMAIN}, hostvars.domain=${HOSTVARS_DOMAIN}"
fi

HOST_IP="$(curl -fsSL https://api.ipify.org || true)"
[[ -n "$HOST_IP" ]] || die "Öffentliche Host-IP konnte nicht ermittelt werden (api.ipify.org)."
info "Host-IP erkannt: $HOST_IP"

mapfile -t ALIASES < <(extract_aliases "$HOSTVARS_FILE")

verify_a_record_matches_host "$DOMAIN" "$HOST_IP"

if [[ ${#ALIASES[@]} -eq 0 ]]; then
  info "Keine Aliase in hostvars gefunden."
else
  info "Prüfe ${#ALIASES[@]} Alias-Domain(s)"
  declare -A seen=()
  for alias in "${ALIASES[@]}"; do
    alias="$(normalize_domain "$alias")"
    [[ -n "$alias" ]] || continue
    if [[ -n "${seen[$alias]:-}" ]]; then
      info "Alias doppelt, übersprungen: $alias"
      continue
    fi
    seen[$alias]=1
    verify_a_record_matches_host "$alias" "$HOST_IP"
  done
fi

ok "Integritätsprüfung erfolgreich: $HOSTVARS_FILE"

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only: kein Redeploy ausgeführt"
  exit 0
fi

require_cmd ansible-playbook
info "Starte Ghost-Redeploy für ${DOMAIN}"
ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$PLAYBOOK"
ok "Ghost-Redeploy abgeschlossen"

if [[ "$RESTART_TRAEFIK" -eq 1 ]]; then
  info "Starte Traefik neu"
  docker restart traefik >/dev/null
  ok "Traefik neugestartet"
fi

cat <<MSG

Hinweis zu TLS/Let's Encrypt:
- Für Alias-Domains sind Zertifikate relevant (nicht irrelevant).
- Da die Router-Regel Host(main + aliases) enthält, zieht Traefik die benötigten Zertifikate/SANs nach,
  sobald Traffic für diese Hosts eingeht und DNS korrekt auf den Host zeigt.

MSG
