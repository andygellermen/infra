#!/bin/bash

# create-hostvars.sh
# Erstellt / aktualisiert eine hostvars-Datei für Ghost-Domain (inkl. Alias-Support und IDN)

set -e

if [[ -z "$1" ]]; then
    echo "Usage: $0 <main-domain> [alias-domain1] [alias-domain2] [...]"
    exit 1
fi

# Prüfe ob 'idn' vorhanden ist
if ! command -v idn >/dev/null 2>&1; then
    echo "Fehler: Das 'idn'-Tool fehlt. Installiere es mit:"
    echo "       sudo apt install idn"
    exit 1
fi

main_domain=$(idn --quiet "$1")
shift

# Verarbeite Alias-Domains
alias_domains=()
# Hauptdomain-www-Alias immer einfügen
alias_domains+=("www.${main_domain}")

for alias in "$@"; do
    punycode=$(idn --quiet "$alias")
    alias_domains+=("$punycode")
    alias_domains+=("www.$punycode")
done

# Datei vorbereiten
hostvars_dir="./ansible/hostvars"
mkdir -p "$hostvars_dir"
file="$hostvars_dir/${main_domain}.yml"

echo "📝 Erstelle oder aktualisiere: $file"

# Generiere DB-Zugangsdaten
db_name="ghost_$(echo "$main_domain" | tr '.' '_' | tr '-' '_')"
db_user="${db_name}_usr"
db_pass=$(openssl rand -hex 12)

# Schreibe YAML-Datei
cat > "$file" <<EOF
traefik:
  domain: ${main_domain}
  aliases:
EOF

for a in "${alias_domains[@]}"; do
  echo "    - ${a}" >> "$file"
done

cat >> "$file" <<EOF

ghost_db_name: ${db_name}
ghost_db_user: ${db_user}
ghost_db_password: ${db_pass}
ghost_env: production
EOF

echo "✅ Hostvars-Datei erzeugt: $file"
