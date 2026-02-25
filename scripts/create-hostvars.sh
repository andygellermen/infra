#!/bin/bash
# create-hostvars.sh
# Erstellt / aktualisiert eine hostvars-Datei für Ghost-Domain (inkl. Alias + DB + IDN)

set -e

ghost_version="latest"

usage() {
    echo "Usage: $0 <main-domain> [alias-domain1] [alias-domain2] [--version=<major|major.minor|major.minor.patch|latest>]"
}

main_domain=""
args=()
for arg in "$@"; do
    case "$arg" in
        --version=*)
            ghost_version="${arg#*=}"
            ;;
        --version)
            echo "Fehler: Bitte --version=<major|major.minor|major.minor.patch|latest> verwenden (z. B. --version=4 oder --version=latest)."
            usage
            exit 1
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
    echo "Fehler: Ungültige Ghost-Version '$ghost_version'. Erlaubt sind 'latest' oder Major-/Minor-/Patch-Versionen (z. B. 6, 6.18, 6.18.2)."
    exit 1
fi

main_domain="${args[0]}"
alias_args=("${args[@]:1}")

# Prüfe idn
if ! command -v idn >/dev/null 2>&1; then
    echo "Fehler: Das 'idn'-Tool fehlt."
    echo "Installiere es mit: sudo apt install idn"
    exit 1
fi

# --- Domains ---
main_domain=$(idn --quiet "$main_domain")

alias_domains=()
alias_domains+=("www.${main_domain}")

for alias in "${alias_args[@]}"; do
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

ghost_version: "${ghost_version}"
ghost_env: production

ghost_traefik_middleware_default: "crowdsec-default@docker"
ghost_traefik_middleware_admin: "crowdsec-admin@docker"
ghost_traefik_middleware_api: "crowdsec-api@docker"
ghost_traefik_middleware_dotghost: "crowdsec-api@docker"
ghost_traefik_middleware_members_api: "crowdsec-api@docker"
EOF

echo "✅ Hostvars-Datei erzeugt: $file"
