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
  local hostvars_file="$1" domain="$2"
  [[ -f "$hostvars_file" ]] && return 0

  warn "Hostvars fehlen für ${domain}. Erzeuge minimale Hostvars für Legacy-Restore."
  local db_prefix db_user_hash db_user db_pwd
  db_prefix="$(echo "$domain" | tr '.-' '__')"
  db_user_hash="$(printf '%s' "$domain" | md5sum | awk '{print $1}')"
  db_user="${db_user_hash:0:24}_usr"
  db_pwd="$(openssl rand -hex 16)"

  mkdir -p "$(dirname "$hostvars_file")"
  cat > "$hostvars_file" <<EOF
domain: ${domain}
wp_domain_db: wp_${db_prefix}
wp_domain_usr: ${db_user}
wp_domain_pwd: ${db_pwd}
wp_table_prefix: wp_
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

require_cmd docker
[[ -f "$BACKUP_FILE" ]] || die "Backup fehlt: $BACKUP_FILE"

MYSQL_CONTAINER="ghost-mysql"
HOSTVARS="$ROOT_DIR/ansible/hostvars/${DOMAIN}.yml"
VOLUME="wp_${DOMAIN//./_}_html"
CONTAINER="wp-${DOMAIN//./-}"

WORKDIR="$(mktemp -d /tmp/wp-restore-${DOMAIN}.XXXXXX)"; trap 'rm -rf "$WORKDIR"' EXIT
extract_backup_archive "$BACKUP_FILE" "$WORKDIR"
DOCROOT="$(detect_docroot "$WORKDIR")"
SELECTED_SQL_FILE="$(select_sql_file "$DOCROOT" "$WORKDIR")"
info "Erkannter Document-Root: ${DOCROOT#$WORKDIR/}"
info "Verwende SQL-Backup: ${SELECTED_SQL_FILE#$WORKDIR/}"

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

ensure_hostvars_exists "$HOSTVARS" "$DOMAIN"
ensure_db_and_user_exists "$HOSTVARS"

DB_NAME="$(extract_hostvar wp_domain_db "$HOSTVARS")"
DB_USER="$(extract_hostvar wp_domain_usr "$HOSTVARS")"
DB_PASS="$(extract_hostvar wp_domain_pwd "$HOSTVARS")"
TABLE_PREFIX="$(extract_hostvar wp_table_prefix "$HOSTVARS")"
TARGET_WP_VERSION="$(extract_hostvar wp_version "$HOSTVARS")"
SOURCE_WP_VERSION=""
[[ -n "$SOURCE_HOSTVARS" ]] && SOURCE_WP_VERSION="$(extract_hostvar wp_version "$SOURCE_HOSTVARS" 2>/dev/null || true)"
[[ -n "$TABLE_PREFIX" ]] || TABLE_PREFIX="wp_"
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

warn "Restore überschreibt WordPress DB + Files für $DOMAIN"
info "Backup-Version: $SOURCE_WP_VERSION | Ziel-Version: $TARGET_WP_VERSION"
read -r -p "Fortfahren? (yes/no): " a
[[ "$a" == yes ]] || die "Abbruch"

docker ps --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER" || die "MySQL Container läuft nicht: $MYSQL_CONTAINER"

info "Stoppe Container: $CONTAINER"
docker stop "$CONTAINER" >/dev/null || true

info "Leere DB und importiere Dump"
{
  echo "SET FOREIGN_KEY_CHECKS=0;"
  docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -Nse "SELECT CONCAT('DROP TABLE IF EXISTS \\`', table_name, '\\`;') FROM information_schema.tables WHERE table_schema='${DB_NAME}';"
  echo "SET FOREIGN_KEY_CHECKS=1;"
} | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"
cat "$SELECTED_SQL_FILE" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

if [[ -n "$SOURCE_DOMAIN" && "$SOURCE_DOMAIN" != "$DOMAIN" ]]; then
  info "Aktualisiere WordPress-Domain in DB (${TABLE_PREFIX}options)"
  docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "UPDATE ${TABLE_PREFIX}options SET option_value='https://${DOMAIN}' WHERE option_name IN ('siteurl','home');"
fi

info "Restore Document Root"
docker volume create "$VOLUME" >/dev/null
docker run --rm -v "${VOLUME}:/target" -v "${DOCROOT}:/src:ro" alpine sh -c 'find /target -mindepth 1 -delete; cp -a /src/. /target/'

info "Starte Container: $CONTAINER"
docker start "$CONTAINER" >/dev/null || true

if [[ "$FORCE_REDEPLOY" -eq 1 ]]; then
  info "Führe gezielten Redeploy aus"
  "$REDEPLOY_SCRIPT" "$DOMAIN"
fi

if [[ "$TARGET_WP_VERSION" != "$SOURCE_WP_VERSION" ]]; then
  info "Versionen unterscheiden sich (Backup=$SOURCE_WP_VERSION, Ziel=$TARGET_WP_VERSION)."
  info "Empfehlung: kontrolliert mit ./scripts/wp-upgrade.sh ${DOMAIN} --version=<ziel> weiterziehen."
fi

ok "WordPress-Restore abgeschlossen"
