#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REDEPLOY_SCRIPT="$ROOT_DIR/scripts/wp-redeploy.sh"

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }
extract_hostvar(){ awk -F': ' -v k="$1" '$1==k {gsub(/"/,"",$2); gsub(/[[:space:]]+$/, "", $2); print $2; exit}' "$2"; }
extract_secret(){ awk -F': ' -v k="$1" '$1==k {gsub(/"/,"",$2); gsub(/[[:space:]]+$/, "", $2); print $2; exit}' "$2"; }

usage() {
  cat <<USAGE
Usage:
  $0 <domain> <backup.tar.gz|backup.tgz|backup.zip> [--restore-hostvars] [--allow-version-downgrade] [--php-version=<major.minor>] [--redeploy]
USAGE
}

version_to_int() {
  local v="$1" a b c
  [[ "$v" == "latest" || -z "$v" ]] && { echo 999999999; return; }
  IFS='.' read -r a b c <<< "$v"
  a="${a:-0}"; b="${b:-0}"; c="${c:-0}"
  printf '%03d%03d%03d\n' "$a" "$b" "$c"
}

set_hostvar_value() {
  local key="$1" value="$2" file="$3"
  if grep -q "^${key}:" "$file"; then
    sed -i -E "s|^${key}:.*|${key}: \"${value}\"|" "$file"
  else
    printf '%s: "%s"\n' "$key" "$value" >> "$file"
  fi
}

extract_wp_config_domain() {
  local config_file="$1"
  [[ -f "$config_file" ]] || return 0
  grep -E "define\(['\"]WP_(HOME|SITEURL)['\"],[[:space:]]*['\"]https?://[^/'\"]+" "$config_file" \
    | sed -E "s/.*https?:\/\/([^/'\"]+).*/\1/" \
    | head -n1 || true
}

set_wp_config_constant() {
  local file="$1" constant="$2" value="$3" escaped_value
  [[ -f "$file" ]] || return 0
  escaped_value="$(printf '%s' "$value" | sed "s/[&|']/\\\\&/g")"

  if grep -Eq "define\([[:space:]]*['\"]${constant}['\"]" "$file"; then
    sed -i -E "s|define\([[:space:]]*['\"]${constant}['\"][[:space:]]*,[[:space:]]*['\"][^'\"]*['\"][[:space:]]*\);|define('${constant}', '${escaped_value}');|" "$file"
  else
    if grep -Eq "^/\* That's all, stop editing! Happy (publishing|blogging)\. \*/" "$file"; then
      sed -i -E "/^\/\* That's all, stop editing! Happy (publishing|blogging)\. \*\//i define('${constant}', '${escaped_value}');" "$file"
    elif grep -q "require_once(ABSPATH . 'wp-settings.php');" "$file"; then
      sed -i "/require_once(ABSPATH \. 'wp-settings\.php');/i define('${constant}', '${escaped_value}');" "$file"
    else
      printf "\ndefine('%s', '%s');\n" "$constant" "$value" >> "$file"
    fi
  fi
}

set_wp_config_table_prefix() {
  local file="$1" value="$2" escaped_value
  [[ -f "$file" ]] || return 0
  escaped_value="$(printf '%s' "$value" | sed "s/[&|']/\\\\&/g")"

  if grep -Eq "^[[:space:]]*\\\$table_prefix[[:space:]]*=" "$file"; then
    sed -i -E "s|^[[:space:]]*\\\$table_prefix[[:space:]]*=.*|\\\$table_prefix = '${escaped_value}';|" "$file"
  else
    if grep -Eq "^/\* That's all, stop editing! Happy (publishing|blogging)\. \*/" "$file"; then
      sed -i -E "/^\/\* That's all, stop editing! Happy (publishing|blogging)\. \*\//i \\\$table_prefix = '${escaped_value}';" "$file"
    elif grep -q "require_once(ABSPATH . 'wp-settings.php');" "$file"; then
      sed -i "/require_once(ABSPATH \. 'wp-settings\.php');/i \\\$table_prefix = '${escaped_value}';" "$file"
    else
      printf "\n\$table_prefix = '%s';\n" "$value" >> "$file"
    fi
  fi
}

ensure_wp_config_proxy_ssl_block() {
  local file="$1" tmp_file block
  [[ -f "$file" ]] || return 0
  tmp_file="$(mktemp "${file}.proxy.XXXXXX")"
  block="$(cat <<'EOF'
define('FORCE_SSL_ADMIN', true);
if (isset($_SERVER['HTTP_X_FORWARDED_PROTO']) && $_SERVER['HTTP_X_FORWARDED_PROTO'] === 'https') {
    $_SERVER['HTTPS'] = 'on';
}
EOF
)"

  awk '
    BEGIN { skip = 0 }
    /^define\('\''FORCE_SSL_ADMIN'\'',[[:space:]]*true\);$/ { skip = 1; next }
    skip && /^if[[:space:]]*\(isset\(\$_SERVER\['\''HTTP_X_FORWARDED_PROTO'\''\]\)[[:space:]]*&&[[:space:]]*\$_SERVER\['\''HTTP_X_FORWARDED_PROTO'\''\][[:space:]]*===[[:space:]]*'\''https'\''\)[[:space:]]*\{$/ { next }
    skip && /^[[:space:]]*\$_SERVER\['\''HTTPS'\''\][[:space:]]*=[[:space:]]*'\''on'\'';$/ { next }
    skip && /^\}[[:space:]]*$/ { skip = 0; next }
    { print }
  ' "$file" > "$tmp_file"
  mv "$tmp_file" "$file"

  if grep -Eq "^/\* That's all, stop editing! Happy (publishing|blogging)\. \*/" "$file"; then
    tmp_file="$(mktemp "${file}.proxy.XXXXXX")"
    awk -v block="$block" '
      /^\/\* That\x27s all, stop editing! Happy (publishing|blogging)\. \*\// {
        print block
        print ""
      }
      { print }
    ' "$file" > "$tmp_file"
    mv "$tmp_file" "$file"
  elif grep -q "require_once(ABSPATH . 'wp-settings.php');" "$file"; then
    tmp_file="$(mktemp "${file}.proxy.XXXXXX")"
    awk -v block="$block" '
      /require_once\(ABSPATH \. '\''wp-settings\.php'\''\);/ {
        print block
        print ""
      }
      { print }
    ' "$file" > "$tmp_file"
    mv "$tmp_file" "$file"
  else
    printf "\n%s\n" "$block" >> "$file"
  fi
}

verify_wp_config_constant() {
  local file="$1" constant="$2" expected_value="$3"
  [[ -f "$file" ]] || die "wp-config.php fehlt: $file"
  local expected_line
  expected_line="define('${constant}', '${expected_value}');"
  grep -Fqx "$expected_line" "$file" || grep -Fq "$expected_line" "$file" \
    || die "${constant} wurde in wp-config.php nicht korrekt gesetzt (erwartet: ${expected_value})"
}

verify_wp_config_table_prefix() {
  local file="$1" expected_value="$2"
  [[ -f "$file" ]] || die "wp-config.php fehlt: $file"
  grep -Eq "^[[:space:]]*\\\$table_prefix[[:space:]]*=[[:space:]]*['\"]${expected_value//\//\\/}['\"][[:space:]]*;" "$file" \
    || die "\$table_prefix wurde in wp-config.php nicht korrekt gesetzt (erwartet: ${expected_value})"
}

wait_for_container_running() {
  local container="$1" timeout="${2:-60}" waited=0
  while (( waited < timeout )); do
    if docker ps --format '{{.Names}}' | grep -qx "$container"; then
      return 0
    fi
    sleep 2
    (( waited += 2 ))
  done
  return 1
}

run_post_restore_checks() {
  local container="$1" domain="$2"
  local headers="" first_status="" location="" public_result="" public_status="" public_redirects=""

  info "Starte Post-Restore-Selbsttest"

  wait_for_container_running "$container" 60 || die "Container ist nach Restore/Redeploy nicht im Status 'running': $container"

  docker exec "$container" php -l /var/www/html/wp-config.php >/dev/null \
    || die "Syntaxfehler in laufender wp-config.php erkannt"
  ok "wp-config.php Syntaxcheck erfolgreich"

  local attempt
  for attempt in {1..20}; do
    headers="$(docker exec "$container" sh -lc "curl -sSI -H 'Host: ${domain}' -H 'X-Forwarded-Proto: https' http://127.0.0.1/" 2>/dev/null || true)"
    first_status="$(printf '%s\n' "$headers" | awk 'toupper($1) ~ /^HTTP/ {print $2; exit}')"
    location="$(printf '%s\n' "$headers" | awk 'tolower($1)=="location:" {sub(/\r$/, "", $2); print $2; exit}')"
    [[ -n "$first_status" ]] && break
    sleep 2
  done

  [[ -n "$first_status" ]] || die "Kein interner HTTP-Response vom WordPress-Container erhalten"

  case "$first_status" in
    200|301|302|303|307|308) ;;
    *) die "Unerwarteter interner HTTP-Status nach Restore: ${first_status}" ;;
  esac

  if [[ "$location" == "https://${domain}/" && "$first_status" =~ ^30[1278]$ ]]; then
    die "Canonical-Redirect-Schleife erkannt: interner Proxy-Request leitet auf dieselbe Ziel-URL weiter"
  fi
  ok "Interner Proxy-/HTTP-Check erfolgreich (Status ${first_status}${location:+, Location ${location}})"

  if public_result="$(curl -k -sSIL --max-redirs 10 -o /dev/null -w '%{http_code} %{num_redirects}' "https://${domain}/" 2>/dev/null)"; then
    public_status="${public_result%% *}"
    public_redirects="${public_result##* }"
    case "$public_status" in
      200|301|302|303|307|308)
        ok "Öffentlicher HTTPS-Check erfolgreich (Finalstatus ${public_status}, Redirects ${public_redirects})"
        ;;
      *)
        warn "Öffentlicher HTTPS-Check meldet Finalstatus ${public_status} (Redirects ${public_redirects})"
        ;;
    esac
  else
    warn "Öffentlicher HTTPS-Check konnte nicht ausgeführt werden"
  fi
}

extract_domain_from_sql() {
  local sql_file="$1"
  local line

  line="$(grep -Ei "(siteurl|home)" "$sql_file" | grep -Eo "https?://[^'\" )]+" | head -n1 || true)"
  if [[ -z "$line" ]]; then
    line="$(grep -Eo "https?://[^'\" )]+" "$sql_file" | head -n1 || true)"
  fi
  [[ -n "$line" ]] || return 0

  printf '%s\n' "$line" | sed -E 's#https?://([^/:]+).*#\1#'
}

ensure_hostvars_exists() {
  local hostvars_file="$1" domain="$2" table_prefix="${3:-wp_}"
  [[ -f "$hostvars_file" ]] && return 0

  warn "Hostvars fehlen für ${domain}. Erzeuge minimale Hostvars für Legacy-Restore."
  local db_name db_user_hash db_user db_pwd
  db_user_hash="$(printf '%s' "$domain" | md5sum | awk '{print $1}')"
  db_name="wp_${db_user_hash:0:24}_db"
  db_user="${db_user_hash:0:24}_usr"
  db_pwd="$(openssl rand -hex 16)"

  mkdir -p "$(dirname "$hostvars_file")"
  cat > "$hostvars_file" <<EOF
domain: ${domain}
wp_domain_db: ${db_name}
wp_domain_usr: ${db_user}
wp_domain_pwd: ${db_pwd}
wp_table_prefix: ${table_prefix}
wp_version: "latest"
EOF
  info "Hostvars neu erzeugt: $hostvars_file"
}

ensure_db_and_user_exists() {
  local hostvars_file="$1"
  local db_name db_user db_pwd secrets_file mysql_root_password
  db_name="$(extract_hostvar wp_domain_db "$hostvars_file")"
  db_user="$(extract_hostvar wp_domain_usr "$hostvars_file")"
  db_pwd="$(extract_hostvar wp_domain_pwd "$hostvars_file")"
  secrets_file="$ROOT_DIR/ansible/secrets/secrets.yml"
  [[ -f "$secrets_file" ]] || die "Secrets-Datei fehlt für DB-Initialisierung: $secrets_file"

  mysql_root_password="$(extract_secret mysql_root_password "$secrets_file")"
  [[ -n "$mysql_root_password" ]] || die "mysql_root_password fehlt in $secrets_file"

  info "Stelle DB/User sicher: ${db_name} / ${db_user}"
  docker exec -e MYSQL_PWD="$mysql_root_password" "$MYSQL_CONTAINER" mysql -uroot -e "CREATE DATABASE IF NOT EXISTS \`${db_name}\`;"
  docker exec -e MYSQL_PWD="$mysql_root_password" "$MYSQL_CONTAINER" mysql -uroot -e "CREATE USER IF NOT EXISTS '${db_user}'@'%' IDENTIFIED BY '${db_pwd}';"
  docker exec -e MYSQL_PWD="$mysql_root_password" "$MYSQL_CONTAINER" mysql -uroot -e "GRANT ALL PRIVILEGES ON \`${db_name}\`.* TO '${db_user}'@'%'; FLUSH PRIVILEGES;"
}

get_mysql_root_password() {
  local secrets_file mysql_root_password
  secrets_file="$ROOT_DIR/ansible/secrets/secrets.yml"
  [[ -f "$secrets_file" ]] || die "Secrets-Datei fehlt: $secrets_file"
  mysql_root_password="$(extract_secret mysql_root_password "$secrets_file")"
  [[ -n "$mysql_root_password" ]] || die "mysql_root_password fehlt in $secrets_file"
  printf '%s\n' "$mysql_root_password"
}

detect_wp_table_prefix_from_sql() {
  local sql_file="$1"
  local table

  table="$(grep -Eo "CREATE TABLE \`[^\\\`]+\`" "$sql_file" | sed -E "s/CREATE TABLE \`([^\\\`]+)\`/\\1/" | grep -E "_options$" | head -n1 || true)"
  if [[ -z "$table" ]]; then
    table="$(grep -Eo "INSERT INTO \`[^\\\`]+\`" "$sql_file" | sed -E "s/INSERT INTO \`([^\\\`]+)\`/\\1/" | grep -E "_options$" | head -n1 || true)"
  fi
  [[ -n "$table" ]] || return 0
  printf '%s\n' "${table%options}"
}

extract_backup_archive() {
  local archive="$1" dest="$2"
  case "$archive" in
    *.tar.gz|*.tgz)
      tar xzf "$archive" -C "$dest"
      ;;
    *.zip)
      require_cmd unzip
      unzip -q "$archive" -d "$dest"
      ;;
    *)
      die "Unbekanntes Backup-Format: $archive (erwartet .tar.gz, .tgz oder .zip)"
      ;;
  esac
}

detect_docroot() {
  local base="$1"
  if [[ -f "$base/wp-config.php" ]]; then
    printf '%s\n' "$base"
    return
  fi

  local candidate
  candidate="$(find "$base" -mindepth 1 -maxdepth 2 -type f -name 'wp-config.php' | head -n1 || true)"
  if [[ -n "$candidate" ]]; then
    dirname "$candidate"
    return
  fi

  # Legacy-Format fallback: data/html.tar.gz
  if [[ -f "$base/data/html.tar.gz" ]]; then
    local legacy="$base/_legacy-docroot"
    mkdir -p "$legacy"
    tar xzf "$base/data/html.tar.gz" -C "$legacy"
    printf '%s\n' "$legacy"
    return
  fi

  die "Kein WordPress-Document-Root im Backup erkannt (wp-config.php fehlt)."
}

select_sql_file() {
  local docroot="$1"
  local backup_root="$2"
  local sql_files=()

  # Zielbild: SQL direkt im Document-Root
  mapfile -t sql_files < <(find "$docroot" -maxdepth 1 -type f -name '*.sql' | sort)
  if [[ ${#sql_files[@]} -eq 0 ]]; then
    # Fallbacks für ältere Backups
    mapfile -t sql_files < <(find "$docroot" -type f -name '*.sql' | sort)
  fi
  if [[ ${#sql_files[@]} -eq 0 ]]; then
    mapfile -t sql_files < <(find "$backup_root" -type f -name '*.sql' | sort)
  fi

  [[ ${#sql_files[@]} -gt 0 ]] || die "Kein SQL-File im Backup gefunden."

  if [[ ${#sql_files[@]} -eq 1 ]]; then
    printf '%s\n' "${sql_files[0]}"
    return
  fi

  echo "Mehrere SQL-Dateien gefunden:"
  local i=1
  for f in "${sql_files[@]}"; do
    echo "  $i) ${f#$backup_root/}"
    ((i++))
  done
  echo "Bitte wähle das Datenbank-Backup zur Wiederherstellung 1...${#sql_files[@]} (standard: 1)"
  read -r selection
  [[ -n "$selection" ]] || selection=1
  [[ "$selection" =~ ^[0-9]+$ ]] || die "Ungültige Auswahl: $selection"
  (( selection >= 1 && selection <= ${#sql_files[@]} )) || die "Auswahl außerhalb des Bereichs"
  printf '%s\n' "${sql_files[$((selection-1))]}"
}

[[ ${1:-} == "--help" || ${1:-} == "-h" ]] && { usage; exit 0; }
[[ $# -ge 2 ]] || { usage; exit 1; }

DOMAIN="$1"; BACKUP_FILE="$2"; shift 2
RESTORE_HOSTVARS=0
ALLOW_DOWNGRADE=0
FORCE_REDEPLOY=0
PHP_VERSION=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --restore-hostvars) RESTORE_HOSTVARS=1; shift ;;
    --allow-version-downgrade) ALLOW_DOWNGRADE=1; shift ;;
    --redeploy) FORCE_REDEPLOY=1; shift ;;
    --php-version=*) PHP_VERSION="${1#*=}"; shift ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

require_cmd curl
require_cmd docker
[[ -f "$BACKUP_FILE" ]] || die "Backup fehlt: $BACKUP_FILE"

MYSQL_CONTAINER="infra-mysql"
HOSTVARS="$ROOT_DIR/ansible/hostvars/${DOMAIN}.yml"
VOLUME="wp_${DOMAIN//./_}_html"
CONTAINER="wp-${DOMAIN//./-}"

WORKDIR="$(mktemp -d /tmp/wp-restore-${DOMAIN}.XXXXXX)"; trap 'rm -rf "$WORKDIR"' EXIT
extract_backup_archive "$BACKUP_FILE" "$WORKDIR"
DOCROOT="$(detect_docroot "$WORKDIR")"
SELECTED_SQL_FILE="$(select_sql_file "$DOCROOT" "$WORKDIR")"
DETECTED_TABLE_PREFIX="$(detect_wp_table_prefix_from_sql "$SELECTED_SQL_FILE" || true)"
info "Erkannter Document-Root: ${DOCROOT#$WORKDIR/}"
info "Verwende SQL-Backup: ${SELECTED_SQL_FILE#$WORKDIR/}"
[[ -n "$DETECTED_TABLE_PREFIX" ]] && info "Erkannter WP-Tabellenprefix aus SQL: ${DETECTED_TABLE_PREFIX}"

if [[ "$RESTORE_HOSTVARS" -eq 1 ]]; then
  if [[ -f "$WORKDIR/_infra/hostvars.yml" ]]; then
    mkdir -p "$(dirname "$HOSTVARS")"
    cp -a "$WORKDIR/_infra/hostvars.yml" "$HOSTVARS"
    info "Hostvars aus Backup wiederhergestellt"
  elif [[ -f "$WORKDIR/files/hostvars.yml" ]]; then
    mkdir -p "$(dirname "$HOSTVARS")"
    cp -a "$WORKDIR/files/hostvars.yml" "$HOSTVARS"
    info "Hostvars aus Backup wiederhergestellt (Legacy-Pfad)"
  else
    warn "--restore-hostvars gesetzt, aber keine hostvars.yml im Backup gefunden (optional)"
  fi
fi

SOURCE_HOSTVARS=""
if [[ -f "$WORKDIR/_infra/hostvars.yml" ]]; then
  SOURCE_HOSTVARS="$WORKDIR/_infra/hostvars.yml"
elif [[ -f "$WORKDIR/files/hostvars.yml" ]]; then
  SOURCE_HOSTVARS="$WORKDIR/files/hostvars.yml"
fi

SOURCE_DOMAIN=""
[[ -n "$SOURCE_HOSTVARS" ]] && SOURCE_DOMAIN="$(extract_hostvar domain "$SOURCE_HOSTVARS" 2>/dev/null || true)"
if [[ -z "$SOURCE_DOMAIN" ]]; then
  SOURCE_DOMAIN="$(extract_domain_from_sql "$SELECTED_SQL_FILE" || true)"
  [[ -n "$SOURCE_DOMAIN" ]] && info "Backup-Domain aus SQL erkannt: $SOURCE_DOMAIN"
fi

if [[ -n "$SOURCE_DOMAIN" && "$SOURCE_DOMAIN" != "$DOMAIN" ]]; then
  echo "Die gewählte WordPress-Domain ('${DOMAIN}') entspricht nicht der Domain aus dem Backup ('${SOURCE_DOMAIN}') soll die neue Domain ('${DOMAIN}') migriert werden? (yes/NO)"
  read -r choice
  if [[ "$choice" != "yes" ]]; then
    warn "Nutze Domain aus dem Backup: ${SOURCE_DOMAIN}"
    DOMAIN="$SOURCE_DOMAIN"
  else
    warn "Domain-Migration aktiv: ${SOURCE_DOMAIN} -> ${DOMAIN}"
  fi
fi

HOSTVARS="$ROOT_DIR/ansible/hostvars/${DOMAIN}.yml"
VOLUME="wp_${DOMAIN//./_}_html"
CONTAINER="wp-${DOMAIN//./-}"

if [[ ! -f "$HOSTVARS" && -n "$SOURCE_HOSTVARS" ]]; then
  mkdir -p "$(dirname "$HOSTVARS")"
  cp -a "$SOURCE_HOSTVARS" "$HOSTVARS"
  set_hostvar_value "domain" "$DOMAIN" "$HOSTVARS"
fi

ensure_hostvars_exists "$HOSTVARS" "$DOMAIN" "${DETECTED_TABLE_PREFIX:-wp_}"
ensure_db_and_user_exists "$HOSTVARS"

DB_NAME="$(extract_hostvar wp_domain_db "$HOSTVARS")"
DB_USER="$(extract_hostvar wp_domain_usr "$HOSTVARS")"
DB_PASS="$(extract_hostvar wp_domain_pwd "$HOSTVARS")"
TABLE_PREFIX="$(extract_hostvar wp_table_prefix "$HOSTVARS")"
TARGET_WP_VERSION="$(extract_hostvar wp_version "$HOSTVARS")"
SOURCE_WP_VERSION=""
[[ -n "$SOURCE_HOSTVARS" ]] && SOURCE_WP_VERSION="$(extract_hostvar wp_version "$SOURCE_HOSTVARS" 2>/dev/null || true)"
[[ -n "$TABLE_PREFIX" ]] || TABLE_PREFIX="wp_"
if [[ -n "${DETECTED_TABLE_PREFIX:-}" && "$TABLE_PREFIX" != "$DETECTED_TABLE_PREFIX" ]]; then
  warn "Setze wp_table_prefix aus SQL: ${TABLE_PREFIX} -> ${DETECTED_TABLE_PREFIX}"
  TABLE_PREFIX="$DETECTED_TABLE_PREFIX"
  set_hostvar_value "wp_table_prefix" "$TABLE_PREFIX" "$HOSTVARS"
fi
[[ -n "$TARGET_WP_VERSION" ]] || TARGET_WP_VERSION="latest"
[[ -n "$SOURCE_WP_VERSION" ]] || SOURCE_WP_VERSION="$TARGET_WP_VERSION"

if [[ -n "$PHP_VERSION" ]]; then
  [[ "$PHP_VERSION" =~ ^[0-9]+\.[0-9]+$ ]] || die "Ungültige PHP-Version: $PHP_VERSION (erwartet z. B. 8.2)"
  if [[ "$TARGET_WP_VERSION" == "latest" ]]; then
    set_hostvar_value "wp_image_tag" "php${PHP_VERSION}-apache" "$HOSTVARS"
  else
    set_hostvar_value "wp_image_tag" "${TARGET_WP_VERSION}-php${PHP_VERSION}-apache" "$HOSTVARS"
  fi
  info "Hostvars aktualisiert: wp_image_tag=$(extract_hostvar wp_image_tag "$HOSTVARS")"
  FORCE_REDEPLOY=1
fi

src_int="$(version_to_int "$SOURCE_WP_VERSION")"
tgt_int="$(version_to_int "$TARGET_WP_VERSION")"
if [[ "$ALLOW_DOWNGRADE" -ne 1 && "$tgt_int" -lt "$src_int" ]]; then
  die "Version-Downgrade erkannt (Backup: $SOURCE_WP_VERSION -> Ziel: $TARGET_WP_VERSION). Entweder wp_version anheben oder --allow-version-downgrade setzen."
fi

CONFIG_DOMAIN="$(extract_wp_config_domain "$DOCROOT/wp-config.php")"
if [[ -n "$CONFIG_DOMAIN" && "$CONFIG_DOMAIN" != "$DOMAIN" ]]; then
  warn "Passe wp-config.php Domain an: ${CONFIG_DOMAIN} -> ${DOMAIN}"
  sed -i "s|${CONFIG_DOMAIN}|${DOMAIN}|g" "$DOCROOT/wp-config.php"
fi

if [[ -f "$DOCROOT/wp-config.php" ]]; then
  info "Setze DB-Zugangsdaten in wp-config.php auf Restore-/Hostvars-Werte"
  set_wp_config_constant "$DOCROOT/wp-config.php" "DB_NAME" "$DB_NAME"
  set_wp_config_constant "$DOCROOT/wp-config.php" "DB_USER" "$DB_USER"
  set_wp_config_constant "$DOCROOT/wp-config.php" "DB_PASSWORD" "$DB_PASS"
  set_wp_config_constant "$DOCROOT/wp-config.php" "DB_HOST" "infra-mysql"
  set_wp_config_constant "$DOCROOT/wp-config.php" "WP_HOME" "https://${DOMAIN}"
  set_wp_config_constant "$DOCROOT/wp-config.php" "WP_SITEURL" "https://${DOMAIN}"
  set_wp_config_table_prefix "$DOCROOT/wp-config.php" "$TABLE_PREFIX"
  ensure_wp_config_proxy_ssl_block "$DOCROOT/wp-config.php"
  verify_wp_config_constant "$DOCROOT/wp-config.php" "DB_NAME" "$DB_NAME"
  verify_wp_config_constant "$DOCROOT/wp-config.php" "DB_USER" "$DB_USER"
  verify_wp_config_constant "$DOCROOT/wp-config.php" "DB_PASSWORD" "$DB_PASS"
  verify_wp_config_constant "$DOCROOT/wp-config.php" "DB_HOST" "infra-mysql"
  verify_wp_config_constant "$DOCROOT/wp-config.php" "WP_HOME" "https://${DOMAIN}"
  verify_wp_config_constant "$DOCROOT/wp-config.php" "WP_SITEURL" "https://${DOMAIN}"
  verify_wp_config_table_prefix "$DOCROOT/wp-config.php" "$TABLE_PREFIX"
fi

warn "Restore überschreibt WordPress DB + Files für $DOMAIN"
info "Backup-Version: $SOURCE_WP_VERSION | Ziel-Version: $TARGET_WP_VERSION"
read -r -p "Fortfahren? (yes/no): " a
[[ "$a" == yes ]] || die "Abbruch"

docker ps --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER" || die "MySQL Container läuft nicht: $MYSQL_CONTAINER"

CONTAINER_EXISTS=0
if docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER"; then
  CONTAINER_EXISTS=1
  info "Stoppe Container: $CONTAINER"
  docker stop "$CONTAINER" >/dev/null || true
else
  warn "Container ${CONTAINER} existiert noch nicht. Restore schreibt DB/Volume und triggert anschließend Redeploy."
  FORCE_REDEPLOY=1
fi

info "Leere DB und importiere Dump"
MYSQL_ROOT_PASSWORD="$(get_mysql_root_password)"
docker exec -e MYSQL_PWD="$MYSQL_ROOT_PASSWORD" "$MYSQL_CONTAINER" mysql -uroot -e "DROP DATABASE IF EXISTS \`${DB_NAME}\`; CREATE DATABASE \`${DB_NAME}\`;"
docker exec -e MYSQL_PWD="$MYSQL_ROOT_PASSWORD" "$MYSQL_CONTAINER" mysql -uroot -e "GRANT ALL PRIVILEGES ON \`${DB_NAME}\`.* TO '${DB_USER}'@'%'; FLUSH PRIVILEGES;"
cat "$SELECTED_SQL_FILE" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

if [[ -n "$SOURCE_DOMAIN" && "$SOURCE_DOMAIN" != "$DOMAIN" ]]; then
  if docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -Nse "SHOW TABLES LIKE '${TABLE_PREFIX}options';" | grep -qx "${TABLE_PREFIX}options"; then
    info "Aktualisiere WordPress-Domain in DB (${TABLE_PREFIX}options)"
    docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "UPDATE ${TABLE_PREFIX}options SET option_value='https://${DOMAIN}' WHERE option_name IN ('siteurl','home');"
  else
    warn "Tabelle ${TABLE_PREFIX}options nicht gefunden - überspringe Domain-Update in DB."
  fi
fi

info "Restore Document Root"
docker volume create "$VOLUME" >/dev/null
docker run --rm -v "${VOLUME}:/target" -v "${DOCROOT}:/src:ro" alpine sh -c 'find /target -mindepth 1 -delete; cp -a /src/. /target/'
docker run --rm -v "${VOLUME}:/target" alpine sh -c '
  set -eu
  # WordPress official Apache image runs as www-data (uid/gid 33).
  chown -R 33:33 /target
  # Ensure Apache can traverse directories and read .htaccess/files.
  find /target -type d -exec chmod 755 {} +
  find /target -type f -exec chmod 644 {} +
  # Keep wp-content writable for plugin/theme updates and uploads.
  if [ -d /target/wp-content ]; then
    find /target/wp-content -type d -exec chmod 775 {} +
    find /target/wp-content -type f -exec chmod 664 {} +
  fi
'

if [[ "$CONTAINER_EXISTS" -eq 1 ]]; then
  info "Starte Container: $CONTAINER"
  docker start "$CONTAINER" >/dev/null || true
fi

if [[ "$FORCE_REDEPLOY" -eq 1 ]]; then
  info "Führe gezielten Redeploy aus"
  "$REDEPLOY_SCRIPT" "$DOMAIN"
fi

if [[ "$TARGET_WP_VERSION" != "$SOURCE_WP_VERSION" ]]; then
  info "Versionen unterscheiden sich (Backup=$SOURCE_WP_VERSION, Ziel=$TARGET_WP_VERSION)."
  info "Empfehlung: kontrolliert mit ./scripts/wp-upgrade.sh ${DOMAIN} --version=<ziel> weiterziehen."
fi

run_post_restore_checks "$CONTAINER" "$DOMAIN"

ok "WordPress-Restore abgeschlossen"
