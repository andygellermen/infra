#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="$ROOT_DIR/backups/ghost"

usage() {
  cat <<USAGE
Usage:
  $0 --create <domain> [--output <file.tar.gz>]
  $0 --restore <domain> <file.tar.gz> [--yes]
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

extract_traefik_aliases() {
  local file="$1"
  awk '
    /^traefik:/ { in_traefik=1; next }
    in_traefik && /^  aliases:/ { in_aliases=1; next }
    in_traefik && in_aliases && /^    - / { sub(/^    - /, "", $0); print; next }
    in_traefik && in_aliases && !/^    - / { in_aliases=0 }
    in_traefik && !/^  / { in_traefik=0 }
  ' "$file"
}

build_domain_list_csv() {
  local domain="$1"
  local hostvars="$2"
  local aliases=()
  local raw

  while IFS= read -r raw; do
    [[ -n "$raw" ]] && aliases+=("$raw")
  done < <(extract_traefik_aliases "$hostvars")

  python3 - "$domain" "${aliases[@]}" <<'PY'
import sys
values = [v.strip() for v in sys.argv[1:] if v and v.strip()]
seen = set()
out = []
for v in values:
    if v not in seen:
        seen.add(v)
        out.append(v)
print(','.join(out))
PY
}

backup_acme_for_domains() {
  local source_acme="$1"
  local target_file="$2"
  local domains_csv="$3"

  python3 - "$source_acme" "$target_file" "$domains_csv" <<'PY'
import json, pathlib, sys

src = pathlib.Path(sys.argv[1])
dst = pathlib.Path(sys.argv[2])
domains = {d.strip().lower() for d in sys.argv[3].split(',') if d.strip()}

if not src.exists() or not domains:
    sys.exit(0)

data = json.loads(src.read_text())
filtered = {}

for resolver, payload in data.items():
    certs = []
    for cert in payload.get("Certificates", []):
        dom = cert.get("domain") or {}
        names = [dom.get("main", "")] + (dom.get("sans") or [])
        names = {n.lower() for n in names if n}
        if names & domains:
            certs.append(cert)
    if certs:
        clone = dict(payload)
        clone["Certificates"] = certs
        filtered[resolver] = clone

if filtered:
    dst.write_text(json.dumps(filtered, indent=2))
PY
}

restore_acme_for_domains() {
  local backup_acme="$1"
  local target_acme="$2"
  local domains_csv="$3"

  python3 - "$backup_acme" "$target_acme" "$domains_csv" <<'PY'
import json, pathlib, sys

backup_path = pathlib.Path(sys.argv[1])
target_path = pathlib.Path(sys.argv[2])
domains = {d.strip().lower() for d in sys.argv[3].split(',') if d.strip()}

if not backup_path.exists() or not domains:
    sys.exit(0)

incoming = json.loads(backup_path.read_text())
current = {}
if target_path.exists() and target_path.stat().st_size > 0:
    current = json.loads(target_path.read_text())

for resolver, payload in incoming.items():
    existing_payload = dict(current.get(resolver, {}))
    existing_certs = list(existing_payload.get("Certificates", []))

    kept = []
    for cert in existing_certs:
        dom = cert.get("domain") or {}
        names = [dom.get("main", "")] + (dom.get("sans") or [])
        names = {n.lower() for n in names if n}
        if not (names & domains):
            kept.append(cert)

    incoming_certs = list(payload.get("Certificates", []))
    merged_payload = dict(existing_payload)
    merged_payload.update({k: v for k, v in payload.items() if k != "Certificates"})
    merged_payload["Certificates"] = kept + incoming_certs
    current[resolver] = merged_payload

target_path.parent.mkdir(parents=True, exist_ok=True)
target_path.write_text(json.dumps(current, indent=2))
PY

  chmod 600 "$target_acme" || true
}

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

    ACME_FILE="$ROOT_DIR/data/traefik/acme/acme.json"
    DOMAINS_CSV="$(build_domain_list_csv "$DOMAIN" "$HOSTVARS")"
    printf '%s\n' "$DOMAINS_CSV" > "$WORKDIR/meta/acme-domains.csv"
    if [[ -n "$DOMAINS_CSV" ]]; then
      backup_acme_for_domains "$ACME_FILE" "$WORKDIR/files/traefik-acme.json" "$DOMAINS_CSV"
    fi

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
    while [[ $# -gt 0 ]]; do
      case "$1" in
        --yes) ASSUME_YES=1; shift ;;
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

    [[ -f "$WORKDIR/data/db.sql" ]] || die "db.sql fehlt im Backup"
    [[ -f "$WORKDIR/data/content.tar.gz" ]] || die "content.tar.gz fehlt im Backup"

    if [[ "$ASSUME_YES" -ne 1 ]]; then
      echo "⚠️  Restore überschreibt Ghost-DB + Content für $DOMAIN"
      read -r -p "Fortfahren? (yes/no): " a
      [[ "$a" == "yes" ]] || die "Abgebrochen"
    fi

    mkdir -p "$(dirname "$HOSTVARS")"
    cp -a "$WORKDIR/files/hostvars.yml" "$HOSTVARS"
    ensure_crowdsec_hostvars_defaults "$HOSTVARS"

    DB_NAME="$(extract_hostvar ghost_domain_db "$HOSTVARS")"
    DB_USER="$(extract_hostvar ghost_domain_usr "$HOSTVARS")"
    DB_PASS="$(extract_hostvar ghost_domain_pwd "$HOSTVARS")"

    info "Safety-Backup vor Restore"
    SAFETY_DIR="$ROOT_DIR/backups/ghost/${DOMAIN}/safety-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$SAFETY_DIR"
    if docker ps --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER"; then
      mysql_dump_cmd "$DB_PASS" "$MYSQL_CONTAINER" "$DB_USER" "$DB_NAME" > "$SAFETY_DIR/pre-restore.sql" || true
    fi
    docker run --rm -v "${VOLUME}:/src:ro" -v "${SAFETY_DIR}:/backup" alpine sh -c 'tar czf /backup/pre-content.tar.gz -C /src .' || true

    info "Stoppe Container: $CONTAINER"
    docker stop "$CONTAINER" >/dev/null || true

    info "Leere DB & importiere Dump"
    docker exec "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -Nse "
SET FOREIGN_KEY_CHECKS=0;
SELECT CONCAT('DROP TABLE IF EXISTS `', table_name, '`;')
FROM information_schema.tables
WHERE table_schema='${DB_NAME}';" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

    cat "$WORKDIR/data/db.sql" | docker exec -i "$MYSQL_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"

    info "Restore Content-Volume"
    restore_volume "$VOLUME" "$WORKDIR/data/content.tar.gz"

    info "Stelle Traefik/CrowdSec Files optional wieder her"
    if [[ -f "$WORKDIR/files/traefik-acme.json" ]]; then
      ACME_TARGET="$ROOT_DIR/data/traefik/acme/acme.json"
      ACME_DOMAINS="$(cat "$WORKDIR/meta/acme-domains.csv" 2>/dev/null || true)"
      restore_acme_for_domains "$WORKDIR/files/traefik-acme.json" "$ACME_TARGET" "$ACME_DOMAINS"
    fi
    [[ -d "$WORKDIR/files/crowdsec" ]] && cp -a "$WORKDIR/files/crowdsec" "$ROOT_DIR/data/"

    info "Starte Container: $CONTAINER"
    docker start "$CONTAINER" >/dev/null || true

    ok "Restore abgeschlossen"
    echo "📄 Safety-Backup: $SAFETY_DIR"
    ;;

  *)
    usage
    exit 1
    ;;
esac
