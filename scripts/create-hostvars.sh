#!/bin/bash
# create-hostvars.sh
# Erstellt / aktualisiert eine hostvars-Datei für Ghost-Domain (inkl. Alias + DB + IDN)

set -e

ghost_version="latest"
wildcard_domain=""
dns_account=""

usage() {
    echo "Usage: $0 <main-domain> [alias-domain1] [alias-domain2] [--version=<major|major.minor|major.minor.patch|latest>] [--wildcard-domain=<apex-domain>] [--dns-account=<key>]"
}

main_domain=""
args=()
for arg in "$@"; do
    case "$arg" in
        --version=*)
            ghost_version="${arg#*=}"
            ;;
        --wildcard-domain=*)
            wildcard_domain="${arg#*=}"
            ;;
        --dns-account=*)
            dns_account="${arg#*=}"
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
if [[ -n "$wildcard_domain" ]]; then
    wildcard_domain=$(idn --quiet "$wildcard_domain")
fi

tls_mode="standard"
if [[ -n "$wildcard_domain" ]]; then
    tls_mode="wildcard"
fi

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
tinybird_token=$(openssl rand -hex 24)
tinybird_datasource=$(echo "$main_domain" | tr '.-' '__')

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

ghost_traefik_middleware_default: ""
ghost_traefik_middleware_admin: "crowdsec-admin@docker"
ghost_traefik_middleware_api: "crowdsec-api@docker"
ghost_traefik_middleware_dotghost: "crowdsec-api@docker"
ghost_traefik_middleware_members_api: "crowdsec-api@docker"

# Optionaler Ghost-Backend-Hinweis fuer den Ghost Image Editor
# ghost_image_editor_notice_enabled: true
# ghost_image_editor_notice_install_url: "https://github.com/andygellermen/ghost-image-editor/"
# ghost_image_editor_notice_remind_hours: 24
# ghost_image_editor_notice_extra_title: "Wichtiger Hinweis"
# ghost_image_editor_notice_extra_text: |
#   Dieser Hinweistext kann pro Ghost-Domain individuell gesetzt werden.
#   Er wird im Modal unterhalb der Standard-Vorteile angezeigt.

tls_mode: "${tls_mode}"
tls_wildcard_domain: "${wildcard_domain}"
tls_dns_account: "${dns_account}"

# Legacy Tinybird custom env defaults
# Hinweis: Das ist nicht automatisch Ghost Native Analytics.
tinybird_enabled: true
tinybird_api_url: "https://api.tinybird.co"
tinybird_workspace: "main"
tinybird_datasource: "ghost_pageviews_${tinybird_datasource}"
tinybird_token: "${tinybird_token}"
tinybird_events_endpoint: "/v0/events?name=pageviews"

# Ghost Native Analytics (offizieller Ghost-/Tinybird-Pfad)
ghost_native_analytics_enabled: false
ghost_native_analytics_profile: ""
ghost_native_analytics_tracker_datasource: "analytics_events"
ghost_traefik_middleware_native_analytics: "crowdsec-api@docker"
EOF

echo "✅ Hostvars-Datei erzeugt: $file"
