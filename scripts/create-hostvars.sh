#!/bin/bash
# create-hostvars.sh
# Erstellt / aktualisiert eine hostvars-Datei für Ghost-Domain (inkl. Alias + DB + IDN)

set -e

if [[ -z "$1" ]]; then
    echo "Usage: $0 <main-domain> [alias-domain1] [alias-domain2] [...]"
    exit 1
fi

# Prüfe idn
if ! command -v idn >/dev/null 2>&1; then
    echo "Fehler: Das 'idn'-Tool fehlt."
    echo "Installiere es mit: sudo apt install idn"
    exit 1
fi

# --- Domains ---
main_domain=$(idn --quiet "$1")
shift

alias_domains=()
alias_domains+=("www.${main_domain}")

for alias in "$@"; do
    puny=$(idn --quiet "$alias")
    alias_domains+=("$puny")
    alias_domains+=("www.$puny")
done

# --- Hostvars-Pfad ---
hostvars_dir="./ansible/hostvars"
mkdir -p "$hostvars_dir"
file="${hostvars_dir}/${main_domain}.yml"

echo "📝 Erstelle oder aktualisiere: $file"

# --- DB-Daten (kompatibel zu blog.geller.men.yml) ---
db_prefix=$(echo "$main_domain" | tr '.-' '__')
ghost_domain_db="ghost_${db_prefix}"
ghost_user_hash=$(printf '%s' "$main_domain" | md5sum | awk '{print $1}')
ghost_domain_usr="${ghost_user_hash:0:24}_usr"
ghost_domain_pwd=$(openssl rand -hex 16)

# --- YAML schreiben ---
cat > "$file" <<EOF
# Hostvars für ${main_domain}

domain: ${main_domain}

traefik:
  domain: ${main_domain}
  aliases:
EOF

for a in "${alias_domains[@]}"; do
  echo "    - ${a}" >> "$file"
done

cat >> "$file" <<EOF

# ghost_mysql:
ghost_domain_db: ${ghost_domain_db}
ghost_domain_usr: ${ghost_domain_usr}
ghost_domain_pwd: ${ghost_domain_pwd}

ghost_env: production
EOF

echo "✅ Hostvars-Datei erzeugt: $file"
