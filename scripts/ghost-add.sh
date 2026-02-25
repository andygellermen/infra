#!/usr/bin/env bash
set -euo pipefail

ANSIBLE_PLAYBOOK="./ansible/playbooks/deploy-ghost.yml"
INVENTORY="./ansible/inventory"
HOSTVARS_DIR="./ansible/hostvars"
CREATE_HOSTVARS="./scripts/create-hostvars.sh"

DEFAULT_GHOST_VERSION="latest"

usage() {
  cat <<EOF
Usage: $0 <domain> [alias1 alias2 ...] [--version=<major|major.minor|major.minor.patch|latest>]

Beispiele:
  $0 blog.example.com
  $0 blog.example.com alias.example.com --version=4
EOF
}

die() {
  echo "❌ Fehler: $*" >&2
  exit 1
}

info() {
  echo "ℹ️  $*"
}

success() {
  echo "✅ $*"
}

resolve_a_record() {
  local domain="$1"
  dig +short A "$domain" | head -n1
}

verify_domain_points_here() {
  local domain="$1"
  local host_ip dns_ip

  host_ip="$(curl -fsSL https://api.ipify.org || true)"
  dns_ip="$(resolve_a_record "$domain")"

  [[ -n "$host_ip" ]] || die "Öffentliche Host-IP konnte nicht ermittelt werden."
  [[ -n "$dns_ip" ]] || die "Kein A-Record für ${domain} gefunden."

  if [[ "$dns_ip" != "$host_ip" ]]; then
    die "DNS-Fehler: ${domain} zeigt auf ${dns_ip}, erwartet wird ${host_ip}. Zertifikats-Deployment wird abgebrochen."
  fi

  info "DNS OK: ${domain} -> ${dns_ip}"
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

ghost_version="$DEFAULT_GHOST_VERSION"
args=()
for arg in "$@"; do
  case "$arg" in
    --version=*)
      ghost_version="${arg#*=}"
      ;;
    --version)
      die "Bitte --version=<major|major.minor|major.minor.patch|latest> verwenden (z. B. --version=4 oder --version=latest)."
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      args+=("$arg")
      ;;
  esac
done

if [[ ${#args[@]} -lt 1 ]]; then
  usage
  exit 1
fi

if [[ "$ghost_version" != "latest" ]] && ! [[ "$ghost_version" =~ ^[0-9]+(\.[0-9]+){0,2}$ ]]; then
  die "Ungültige Ghost-Version '$ghost_version'. Erlaubt sind 'latest' oder Major-/Minor-/Patch-Versionen (z. B. 4, 6.18, 6.18.2)."
fi

DOMAIN_RAW="${args[0]}"
ALIASES_RAW=("${args[@]:1}")

echo "🚀 Starte Ghost-Setup für ${DOMAIN_RAW} (Ghost ${ghost_version})"

# =========================
# idn vorhanden?
# =========================
if ! command -v idn >/dev/null 2>&1; then
  die "Das 'idn'-Tool fehlt. Installiere es mit: sudo apt install idn"
fi
if ! command -v dig >/dev/null 2>&1; then
  die "Das 'dig'-Tool fehlt. Installiere es mit: sudo apt install dnsutils"
fi
if ! command -v curl >/dev/null 2>&1; then
  die "Das 'curl'-Tool fehlt. Installiere es mit: sudo apt install curl"
fi

# =========================
# Domain validieren & normalisieren
# =========================
normalize_domain() {
  local d="$1"

  # ASCII-Domain → direkt zurück
  if [[ "$d" =~ ^[a-zA-Z0-9.-]+$ ]]; then
    echo "$d"
    return 0
  fi

  # Nicht-ASCII → idn
  local p
  p="$(printf '%s' "$d" | idn --quiet --uts46 2>/dev/null || true)"

  [[ -z "$p" ]] && return 1
  echo "$p"
}

DOMAIN_PUNY="$(normalize_domain "$DOMAIN_RAW")" \
  || die "Ungültige Domain: '$DOMAIN_RAW'"

# =========================
# Aliase normalisieren
# =========================
ALIASES_PUNY=()
for a in "${ALIASES_RAW[@]}"; do
  [[ -z "$a" ]] && continue
  p="$(normalize_domain "$a")" \
    || die "Ungültige Alias-Domain: '$a'"
  ALIASES_PUNY+=("$p")
done

# =========================
# DNS A-Record prüfen (Pflicht für LE-Zertifikate)
# =========================
verify_domain_points_here "$DOMAIN_PUNY"
for alias in "${ALIASES_PUNY[@]}"; do
  verify_domain_points_here "$alias"
done

# =========================
# Hostvars erzeugen
# =========================
info "Erstelle oder aktualisiere Hostvars für ${DOMAIN_PUNY}"

mkdir -p "$HOSTVARS_DIR"

"$CREATE_HOSTVARS" \
  "$DOMAIN_PUNY" \
  "${ALIASES_PUNY[@]}" \
  "--version=${ghost_version}"

HOSTVARS_FILE="${HOSTVARS_DIR}/${DOMAIN_PUNY}.yml"

[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars-Datei wurde nicht erzeugt"

success "Hostvars-Datei erzeugt: $HOSTVARS_FILE"

# =========================
# Ansible Deployment
# =========================
info "Starte Ansible Deployment"

ansible-playbook \
  -i "$INVENTORY" \
  -e "target_domain=${DOMAIN_PUNY}" \
  "$ANSIBLE_PLAYBOOK"

# =========================
# Traefik Reload
# =========================
info "Starte Traefik neu zur Zertifikatsprüfung..."
docker restart traefik >/dev/null

success "Ghost-Setup für ${DOMAIN_PUNY} abgeschlossen 🎉"
