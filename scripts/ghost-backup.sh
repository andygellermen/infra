#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="$ROOT_DIR/backups/ghost"

usage() {
  cat <<USAGE
Usage:
  $0 --create <domain> [--output <file.tar.gz>]
  $0 --restore <domain> <file.tar.gz> [--yes] [--content-only]
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

ensure_crowdsec_hostvars_defaults() {
  local hostvars="$1"
  local defaults=(
    'ghost_traefik_middleware_default: "crowdsec-default@docker"'
    'ghost_traefik_middleware_admin: "crowdsec-admin@docker"'
    'ghost_traefik_middleware_api: "crowdsec-api@docker"'
    'ghost_traefik_middleware_dotghost: "crowdsec-api@docker"'
    'ghost_traefik_middleware_members_api: "crowdsec-api@docker"'
  )
  local line key

  for line in "${defaults[@]}"; do
    key="${line%%:*}"
    if ! grep -qE "^${key}:" "$hostvars"; then
      printf '%s\n' "$line" >> "$hostvars"
    fi
  done
}

mysql_dump_cmd() {
  # --no-tablespaces verhindert PROCESS-Privilege Fehler bei eingeschränkten DB-Usern
  docker exec -e MYSQL_PWD="$1" "$2" mysqldump --no-tablespaces -u"$3" "$4"
}

extract_hostvar() {
  local key="$1" file="$2"
  awk -F': ' -v k="$key" '$1==k {gsub(/"/,"",$2); gsub(/[[:space:]]+$/, "", $2); print $2; exit}' "$file"
}

backup_volume() {
  local volume="$1" out="$2"
  docker run --rm -v "${volume}:/src:ro" -v "${out}:/backup" alpine sh -c 'tar czf /backup/content.tar.gz -C /src .'
}

restore_volume() {
  local volume="$1" archive="$2"
  docker volume create "$volume" >/dev/null
  docker run --rm -v "${volume}:/target" -v "$(dirname "$archive"):/backup:ro" alpine \
    sh -c 'find /target -mindepth 1 -delete; tar xzf /backup/content.tar.gz -C /target'
}

if [[ ${1:-} == "--help" || ${1:-} == "-h" ]]; then
  usage
  exit 0
fi

[[ $# -ge 2 ]] || { usage; exit 1; }
ACTION="$1"; shift

case "$ACTION" in
  --create)
    DOMAIN="$1"; shift
    TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
    OUTPUT_FILE="$BACKUP_DIR/$DOMAIN/ghost-backup-${DOMAIN}-${TIMESTAMP}.tar.gz"
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --output) OUTPUT_FILE="$2"; shift 2 ;;
        *) die "Unbekannte Option: $1" ;;
      esac
    done

    require_cmd docker
    mkdir -p "$(dirname "$OUTPUT_FILE")"

    HOSTVARS="$ROOT_DIR/ansible/hostvars/${DOMAIN}.yml"
    CONTAINER="ghost-${DOMAIN//./-}"
    VOLUME="ghost_${DOMAIN//./_}_content"
    MYSQL_CONTAINER="ghost-mysql"

    [[ -f "$HOSTVARS" ]] || die "Hostvars fehlt: $HOSTVARS"
    docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER" || die "Container fehlt: $CONTAINER"
    docker volume ls --format '{{.Name}}' | grep -qx "$VOLUME" || die "Volume fehlt: $VOLUME"
    docker ps --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER" || die "MySQL Container läuft nicht: $MYSQL_CONTAINER"

    DB_NAME="$(extract_hostvar ghost_domain_db "$HOSTVARS")"
    DB_USER="$(extract_hostvar ghost_domain_usr "$HOSTVARS")"
    DB_PASS="$(extract_hostvar ghost_domain_pwd "$HOSTVARS")"

    WORKDIR="$(mktemp -d /tmp/ghost-backup-${DOMAIN}.XXXXXX)"
    trap 'rm -rf "$WORKDIR"' EXIT
    mkdir -p "$WORKDIR"/{meta,data,files}

    info "Dump DB: $DB_NAME"
    if ! mysql_dump_cmd "$DB_PASS" "$MYSQL_CONTAINER" "$DB_USER" "$DB_NAME" > "$WORKDIR/data/db.sql"; then
      die "MySQL-Dump fehlgeschlagen (DB=${DB_NAME}, User=${DB_USER}). Prüfe Berechtigungen/Verbindung."
    fi
    [[ -s "$WORKDIR/data/db.sql" ]] || die "MySQL-Dump ist leer: $WORKDIR/data/db.sql"

    info "Backup Content-Volume: $VOLUME"
    backup_volume "$VOLUME" "$WORKDIR/data"

    cp -a "$HOSTVARS" "$WORKDIR/files/hostvars.yml"
    [[ -d "$ROOT_DIR/data/crowdsec" ]] && cp -a "$ROOT_DIR/data/crowdsec" "$WORKDIR/files/crowdsec"


    {
      echo "domain=$DOMAIN"
      echo "container=$CONTAINER"
      echo "volume=$VOLUME"
      echo "timestamp=$TIMESTAMP"
    } > "$WORKDIR/meta/manifest.env"

    tar czf "$OUTPUT_FILE" -C "$WORKDIR" .
    ok "Ghost-Backup erstellt: $OUTPUT_FILE"
    ;;

  --restore)
    DOMAIN="$1"; BACKUP_FILE="$2"; shift 2
    ASSUME_YES=0
    CONTENT_ONLY=0
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --yes) ASSUME_YES=1; shift ;;
        --content-only) CONTENT_ONLY=1; shift ;;
        *) die "Unbekannte Option: $1" ;;
      esac
    done

    require_cmd docker
    [[ -f "$BACKUP_FILE" ]] || die "Backup nicht gefunden: $BACKUP_FILE"

    CONTAINER="ghost-${DOMAIN//./-}"
    VOLUME="ghost_${DOMAIN//./_}_content"
    MYSQL_CONTAINER="ghost-mysql"
    HOSTVARS="$ROOT_DIR/ansible/hostvars/${DOMAIN}.yml"

    WORKDIR="$(mktemp -d /tmp/ghost-restore-${DOMAIN}.XXXXXX)"
    trap 'rm -rf "$WORKDIR"' EXIT
    tar xzf "$BACKUP_FILE" -C "$WORKDIR"

    [[ -f "$WORKDIR/data/content.tar.gz" ]] || die "content.tar.gz fehlt im Backup"
    if [[ "$CONTENT_ONLY" -eq 0 ]]; then
      [[ -f "$WORKDIR/data/db.sql" ]] || die "db.sql fehlt im Backup (oder --content-only verwenden)"
    fi

    if [[ "$ASSUME_YES" -ne 1 ]]; then
      if [[ "$CONTENT_ONLY" -eq 1 ]]; then
        echo "⚠️  Restore überschreibt nur Ghost-Content für $DOMAIN (--content-only)"
      else
        echo "⚠️  Restore überschreibt Ghost-DB + Content für $DOMAIN"
      fi
      read -r -p "Fortfahren? (yes/no): " a
      [[ "$a" == "yes" ]] || die "Abgebrochen"
    fi

    if [[ "$CONTENT_ONLY" -eq 0 ]]; then
      [[ -f "$WORKDIR/files/hostvars.yml" ]] || die "hostvars.yml fehlt im Backup (oder --content-only verwenden)"
      mkdir -p "$(dirname "$HOSTVARS")"
      cp -a "$WORKDIR/files/hostvars.yml" "$HOSTVARS"
      ensure_crowdsec_hostvars_defaults "$HOSTVARS"
    else
      info "--content-only aktiv: Hostvars/Domain-Setup bleiben unverändert"
    fi

    DB_NAME="$(extract_hostvar ghost_domain_db "$HOSTVARS")"
    DB_USER="$(extract_hostvar ghost_domain_usr "$HOSTVARS")"
    DB_PASS="$(extract_hostvar ghost_domain_pwd "$HOSTVARS")"

    info "Safety-Backup vor Restore"
    SAFETY_DIR="$ROOT_DIR/backups/ghost/${DOMAIN}/safety-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$SAFETY_DIR"
    if [[ "$CONTENT_ONLY" -eq 0 ]] && docker ps --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER"; then
      mysql_dump_cmd "$DB_PASS" "$MYSQL_CONTAINER" "$DB_USER" "$DB_NAME" > "$SAFETY_DIR/pre-restore.sql" || true
    fi
    docker run --rm -v "${VOLUME}:/src:ro" -v "${SAFETY_DIR}:/backup" alpine sh -c 'tar czf /backup/pre-content.tar.gz -C /src .' || true

    info "Stoppe Container: $CONTAINER"
    docker stop "$CONTAINER" >/dev/null || true

    if [[ "$CONTENT_ONLY" -eq 0 ]]; then
      info "Leere DB & importiere Dump"
      docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -Nse "
SET FOREIGN_KEY_CHECKS=0;
SELECT CONCAT('DROP TABLE IF EXISTS `', table_name, '`;')
FROM information_schema.tables
WHERE table_schema='${DB_NAME}';" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

      cat "$WORKDIR/data/db.sql" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"
    else
      info "--content-only aktiv: DB-Reset und SQL-Import übersprungen"
    fi

    info "Restore Content-Volume"
    restore_volume "$VOLUME" "$WORKDIR/data/content.tar.gz"

    if [[ "$CONTENT_ONLY" -eq 0 ]]; then
      info "Stelle CrowdSec Files optional wieder her (TLS-Zertifikate werden nicht aus Backup importiert)"
      [[ -d "$WORKDIR/files/crowdsec" ]] && cp -a "$WORKDIR/files/crowdsec" "$ROOT_DIR/data/"
    else
      info "--content-only aktiv: CrowdSec/Host-Konfiguration bleibt unverändert"
    fi

    info "Starte Container: $CONTAINER"
    docker start "$CONTAINER" >/dev/null || true

    ok "Restore abgeschlossen"
    [[ "$CONTENT_ONLY" -eq 1 ]] && echo "🧬 Klon-Modus: Nur Ghost-Content wurde wiederhergestellt."
    echo "📄 Safety-Backup: $SAFETY_DIR"
    echo "🔐 TLS-Hinweis: Zertifikate werden durch Traefik/Let's Encrypt neu erstellt (ggf. Traefik neu starten und Domain aufrufen)."
    ;;

  *)
    usage
    exit 1
    ;;
esac
