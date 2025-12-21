#!/usr/bin/env bash
set -euo pipefail

ANSIBLE_PLAYBOOK="./ansible/playbooks/deploy-ghost.yml"
INVENTORY="./ansible/inventory"
HOSTVARS_DIR="./ansible/hostvars"
CREATE_HOSTVARS="./scripts/create-hostvars.sh"

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

# =========================
# Parameter
# =========================
if [[ $# -lt 1 ]]; then
  die "Usage: $0 <domain> [alias1 alias2 ...]"
fi

DOMAIN_RAW="$1"
shift
ALIASES_RAW=("$@")

echo "🚀 Starte Ghost-Setup für ${DOMAIN_RAW}"

# =========================
# idn vorhanden?
# =========================
if ! command -v idn >/dev/null 2>&1; then
  die "Das 'idn'-Tool fehlt. Installiere es mit: sudo apt install idn"
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
# Hostvars erzeugen
# =========================
info "Erstelle oder aktualisiere Hostvars für ${DOMAIN_PUNY}"

mkdir -p "$HOSTVARS_DIR"

"$CREATE_HOSTVARS" \
  "$DOMAIN_PUNY" \
  "${ALIASES_PUNY[@]}"

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
