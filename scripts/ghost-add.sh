#!/usr/bin/env bash
set -euo pipefail

ANSIBLE_PLAYBOOK="./ansible/playbooks/deploy-ghost.yml"
INVENTORY="./ansible/inventory"
HOSTVARS_DIR="./ansible/hostvars"
CREATE_HOSTVARS="./scripts/create-hostvars.sh"

DEFAULT_GHOST_VERSION="latest"

usage() {
  cat <<EOF
Usage: $0 <domain> [alias1 alias2 ...] [--version=<major|latest>]

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
      die "Bitte --version=<major|latest> verwenden (z. B. --version=4 oder --version=latest)."
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

if [[ "$ghost_version" != "latest" ]] && ! [[ "$ghost_version" =~ ^[0-9]+$ ]]; then
  die "Ungültige Ghost-Version '$ghost_version'. Erlaubt sind 'latest' oder numerische Major-Versionen (z. B. 4, 5, 6)."
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
