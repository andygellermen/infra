#!/usr/bin/env bash
set -euo pipefail

HOSTVARS_DIR="./ansible/hostvars"
INVENTORY="./ansible/inventory"
PLAYBOOK="./ansible/playbooks/deploy-ghost.yml"

usage() {
  cat <<USAGE
Usage: $0 <domain> --version=<major|major.minor|major.minor.patch|latest> [--force-major-jump] [--dry-run]

Beispiele:
  $0 blog.example.com --version=5
  $0 blog.example.com --version=6.18
  $0 blog.example.com --version=6.18.2
  $0 blog.example.com --version=latest --force-major-jump
USAGE
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

if [[ $# -lt 2 ]]; then
  usage
  exit 1
fi

domain=""
target_version=""
force_major_jump="false"
dry_run="false"

for arg in "$@"; do
  case "$arg" in
    --version=*)
      target_version="${arg#*=}"
      ;;
    --version)
      die "Bitte --version=<major|major.minor|major.minor.patch|latest> verwenden (z. B. --version=5, --version=6.18 oder --version=latest)."
      ;;
    --force-major-jump)
      force_major_jump="true"
      ;;
    --dry-run)
      dry_run="true"
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    --*)
      die "Unbekannte Option: $arg"
      ;;
    *)
      if [[ -z "$domain" ]]; then
        domain="$arg"
      else
        die "Unerwartetes Argument: $arg"
      fi
      ;;
  esac
done

[[ -n "$domain" ]] || die "Bitte eine Domain angeben."
[[ -n "$target_version" ]] || die "Bitte --version=<major|major.minor|major.minor.patch|latest> angeben."

if [[ "$target_version" != "latest" ]] && ! [[ "$target_version" =~ ^[0-9]+(\.[0-9]+){0,2}$ ]]; then
  die "Ungültige Zielversion '$target_version'. Erlaubt sind 'latest', Major (z. B. 6), Minor (z. B. 6.18) oder Patch (z. B. 6.18.2)."
fi

hostvars_file="${HOSTVARS_DIR}/${domain}.yml"
[[ -f "$hostvars_file" ]] || die "Hostvars-Datei fehlt: $hostvars_file"

current_version="$(awk -F':' '/^ghost_version:/ {gsub(/["[:space:]]/,"",$2); print $2}' "$hostvars_file" | tail -n 1)"
[[ -n "$current_version" ]] || die "ghost_version konnte nicht aus $hostvars_file gelesen werden."

info "Aktuelle Version für ${domain}: ${current_version}"
info "Zielversion: ${target_version}"

if [[ "$current_version" == "$target_version" ]]; then
  info "Die Zielversion ist bereits gesetzt. Es ist nichts zu tun."
  exit 0
fi

current_major=""
target_major=""

if [[ "$current_version" =~ ^[0-9]+(\.[0-9]+){0,2}$ ]]; then
  current_major="${current_version%%.*}"
fi

if [[ "$target_version" =~ ^[0-9]+(\.[0-9]+){0,2}$ ]]; then
  target_major="${target_version%%.*}"
fi

if [[ -n "$current_major" ]] && [[ -n "$target_major" ]]; then
  expected_next=$((current_major + 1))
  if (( target_major > expected_next )) && [[ "$force_major_jump" != "true" ]]; then
    die "Sprung von Major ${current_major} auf ${target_major} ist größer als +1. Nutze --force-major-jump oder upgrade schrittweise."
  fi
fi

backup_file="${hostvars_file}.bak.$(date +%Y%m%d%H%M%S)"
cp "$hostvars_file" "$backup_file"
info "Backup erstellt: $backup_file"

if grep -q '^ghost_version:' "$hostvars_file"; then
  sed -i -E "s/^ghost_version: .*/ghost_version: \"${target_version}\"/" "$hostvars_file"
else
  printf '\nghost_version: "%s"\n' "$target_version" >> "$hostvars_file"
fi

success "ghost_version in ${hostvars_file} auf ${target_version} gesetzt."

if [[ "$dry_run" == "true" ]]; then
  info "Dry-Run aktiv: ansible-playbook wird nicht ausgeführt."
  exit 0
fi

info "Starte Deployment"
ansible-playbook -i "$INVENTORY" -e "target_domain=${domain}" "$PLAYBOOK"

success "Upgrade-Deployment für ${domain} abgeschlossen 🎉"
