#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
REDIRECTS_FILE="$ROOT_DIR/ansible/redirects/redirects.yml"
ACME_FILE="/home/andy/infra/data/traefik/acme/acme.json"
GHOST_DELETE="$ROOT_DIR/scripts/ghost-delete.sh"
WP_DELETE="$ROOT_DIR/scripts/wp-delete.sh"
STATIC_DELETE="$ROOT_DIR/scripts/static-delete.sh"
REDIRECT_REDEPLOY="$ROOT_DIR/scripts/redirect-redeploy.sh"

usage() {
  cat <<'USAGE'
Usage: ./scripts/delete-domain.sh <domain> [--yes]
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

normalize_domain() {
  local domain="$1"
  if [[ "$domain" =~ ^[A-Za-z0-9.-]+$ ]]; then
    printf '%s\n' "${domain,,}"
  else
    require_cmd idn
    idn --quiet --uts46 "$domain" | tr '[:upper:]' '[:lower:]'
  fi
}

YES=0
ARGS=()
for arg in "$@"; do
  case "$arg" in
    --yes) YES=1 ;;
    --help|-h) usage; exit 0 ;;
    *) ARGS+=("$arg") ;;
  esac
done

[[ ${#ARGS[@]} -eq 1 ]] || { usage; exit 1; }

require_cmd python3
require_cmd docker

DOMAIN="$(normalize_domain "${ARGS[0]}")"
HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
TMP_INFO="$(mktemp)"
TMP_DOMAINS="$(mktemp)"
cleanup() {
  rm -f "$TMP_INFO" "$TMP_DOMAINS"
}
trap cleanup EXIT

python3 - "$DOMAIN" "$HOSTVARS_FILE" "$REDIRECTS_FILE" "$TMP_INFO" "$TMP_DOMAINS" <<'PY'
import json
import os
import sys

try:
    import yaml
except Exception as exc:
    print(f"❌ PyYAML fehlt: {exc}", file=sys.stderr)
    sys.exit(1)

domain = sys.argv[1]
hostvars_file = sys.argv[2]
redirects_file = sys.argv[3]
info_path = sys.argv[4]
domains_path = sys.argv[5]

info = {
    "domain": domain,
    "type": "",
    "hostvars_exists": False,
    "aliases": [],
    "redirect_source": False,
    "redirect_aliases": [],
}
domains_to_prune = {domain}

if os.path.exists(hostvars_file):
    with open(hostvars_file, "r", encoding="utf-8") as fh:
        payload = yaml.safe_load(fh) or {}
    info["hostvars_exists"] = True
    if payload.get("static_enabled") is True:
        info["type"] = "static"
    elif "wp_version" in payload or "wp_domain_db" in payload:
        info["type"] = "wordpress"
    elif "ghost_version" in payload or "ghost_mysql_db" in payload:
        info["type"] = "ghost"
    aliases = (((payload.get("traefik") or {}).get("aliases")) or [])
    info["aliases"] = [alias for alias in aliases if alias]
    domains_to_prune.update(info["aliases"])

if os.path.exists(redirects_file):
    with open(redirects_file, "r", encoding="utf-8") as fh:
        payload = yaml.safe_load(fh) or {}
    for entry in payload.get("redirects") or []:
        source = entry.get("source")
        aliases = [alias for alias in (entry.get("aliases") or []) if alias]
        if source == domain:
            info["redirect_source"] = True
            domains_to_prune.add(source)
            domains_to_prune.update(aliases)
        elif domain in aliases:
            info["redirect_aliases"].append(source or "")
            domains_to_prune.add(domain)

with open(info_path, "w", encoding="utf-8") as fh:
    json.dump(info, fh)

with open(domains_path, "w", encoding="utf-8") as fh:
    json.dump(sorted(domains_to_prune), fh)
PY

INFO_JSON="$(cat "$TMP_INFO")"
DOMAINS_JSON="$(cat "$TMP_DOMAINS")"

TYPE="$(python3 - "$INFO_JSON" <<'PY'
import json, sys
print((json.loads(sys.argv[1]).get("type") or "").strip())
PY
)"

HOSTVARS_EXISTS="$(python3 - "$INFO_JSON" <<'PY'
import json, sys
print("yes" if json.loads(sys.argv[1]).get("hostvars_exists") else "no")
PY
)"

REDIRECT_MATCH="$(python3 - "$INFO_JSON" <<'PY'
import json, sys
payload = json.loads(sys.argv[1])
print("yes" if payload.get("redirect_source") or payload.get("redirect_aliases") else "no")
PY
)"

[[ "$HOSTVARS_EXISTS" == "yes" || "$REDIRECT_MATCH" == "yes" ]] || die "Keine passende Domain-Konfiguration gefunden: $DOMAIN"

if [[ "$TYPE" == "ghost" ]]; then
  info "Erkannter Domain-Typ: Ghost"
elif [[ "$TYPE" == "wordpress" ]]; then
  info "Erkannter Domain-Typ: WordPress"
elif [[ "$TYPE" == "static" ]]; then
  info "Erkannter Domain-Typ: Static"
fi

python3 - "$INFO_JSON" <<'PY'
import json, sys
payload = json.loads(sys.argv[1])
aliases = payload.get("aliases") or []
if aliases:
    print("ℹ️  Zusätzliche Alias-Domains: " + ", ".join(aliases))
if payload.get("redirect_source"):
    print(f"ℹ️  Redirect-Quelle wird entfernt: {payload['domain']}")
for source in payload.get("redirect_aliases") or []:
    if source:
        print(f"ℹ️  Redirect-Alias wird aus Eintrag entfernt: {payload['domain']} (Quelle: {source})")
PY

if [[ "$YES" -ne 1 ]]; then
  warn "Dieser Vorgang entfernt die Domain-Konfiguration und bereinigt ACME-Zertifikate."
  read -r -p "Domain ${DOMAIN} wirklich löschen? (yes/NO): " confirm
  [[ "$confirm" == "yes" ]] || die "Abgebrochen"
fi

if [[ "$TYPE" == "ghost" ]]; then
  "$GHOST_DELETE" "$DOMAIN"
elif [[ "$TYPE" == "wordpress" ]]; then
  "$WP_DELETE" "$DOMAIN"
elif [[ "$TYPE" == "static" ]]; then
  "$STATIC_DELETE" "$DOMAIN"
  if [[ -d "/srv/static/${DOMAIN}" ]]; then
    info "Entferne statische Dateien: /srv/static/${DOMAIN}"
    rm -rf "/srv/static/${DOMAIN}"
    ok "Statische Dateien entfernt"
  fi
  find /srv/static-auth -maxdepth 1 -type f -name "${DOMAIN}*.htpasswd" -delete 2>/dev/null || true
else
  info "Keine Ghost-/WordPress-/Static-Instanz für ${DOMAIN} erkannt"
fi

if [[ -f "$REDIRECTS_FILE" ]]; then
  python3 - "$DOMAIN" "$REDIRECTS_FILE" <<'PY'
import os
import sys
from datetime import datetime

try:
    import yaml
except Exception as exc:
    print(f"❌ PyYAML fehlt: {exc}", file=sys.stderr)
    sys.exit(1)

domain = sys.argv[1]
redirects_file = sys.argv[2]
with open(redirects_file, "r", encoding="utf-8") as fh:
    payload = yaml.safe_load(fh) or {}
entries = payload.get("redirects") or []
new_entries = []
changed = False
for entry in entries:
    source = entry.get("source")
    aliases = [alias for alias in (entry.get("aliases") or []) if alias != domain]
    if source == domain:
        changed = True
        continue
    if len(aliases) != len(entry.get("aliases") or []):
        entry = dict(entry)
        if aliases:
            entry["aliases"] = aliases
        else:
            entry.pop("aliases", None)
        changed = True
    new_entries.append(entry)
if changed:
    timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
    backup_file = f"{redirects_file}.bak.{timestamp}"
    with open(redirects_file, "r", encoding="utf-8") as fh:
        original = fh.read()
    with open(backup_file, "w", encoding="utf-8") as fh:
        fh.write(original)
    with open(redirects_file, "w", encoding="utf-8") as fh:
        yaml.safe_dump({"redirects": new_entries}, fh, sort_keys=False, allow_unicode=True)
    print(backup_file)
PY
fi >"$TMP_INFO" || die "Redirect-Konfiguration konnte nicht bereinigt werden"

if [[ -s "$TMP_INFO" ]]; then
  info "Redirect-Konfiguration bereinigt"
  info "Backup der Redirect-Datei: $(cat "$TMP_INFO")"
  "$REDIRECT_REDEPLOY"
fi

if [[ -f "$ACME_FILE" ]]; then
  if ! python3 - "$DOMAINS_JSON" "$ACME_FILE" <<'PY' >"$TMP_INFO"; then
import json
import os
import sys
from datetime import datetime

domains = set(json.loads(sys.argv[1]))
acme_file = sys.argv[2]

with open(acme_file, "r", encoding="utf-8") as fh:
    payload = json.load(fh)

removed = 0

def prune(node):
    global removed
    if isinstance(node, dict):
        for key, value in list(node.items()):
            if key.lower() == "certificates" and isinstance(value, list):
                kept = []
                for cert in value:
                    domain = cert.get("domain") or cert.get("Domain") or {}
                    main = domain.get("main") or domain.get("Main")
                    sans = domain.get("sans") or domain.get("SANs") or domain.get("Sans") or []
                    names = {name for name in [main, *sans] if name}
                    if names & domains:
                        removed += 1
                    else:
                        kept.append(cert)
                node[key] = kept
            else:
                prune(value)
    elif isinstance(node, list):
        for item in node:
            prune(item)

prune(payload)

timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
backup_file = f"{acme_file}.bak.{timestamp}"
with open(acme_file, "r", encoding="utf-8") as fh:
    original = fh.read()
with open(backup_file, "w", encoding="utf-8") as fh:
    fh.write(original)
with open(acme_file, "w", encoding="utf-8") as fh:
    json.dump(payload, fh, indent=2)
    fh.write("\n")
os.chmod(acme_file, 0o600)

print(f"{removed}|{backup_file}")
PY
    die "ACME-Bereinigung fehlgeschlagen"
  fi
else
  info "Keine ACME-Datei gefunden unter ${ACME_FILE}"
fi

if [[ -s "$TMP_INFO" ]]; then
  ACME_REMOVED="$(cut -d'|' -f1 "$TMP_INFO")"
  ACME_BACKUP="$(cut -d'|' -f2- "$TMP_INFO")"
  info "ACME-Backup erstellt: ${ACME_BACKUP}"
  if [[ "$ACME_REMOVED" -gt 0 ]]; then
    ok "ACME-Zertifikatseinträge entfernt: ${ACME_REMOVED}"
  else
    info "Keine passenden ACME-Zertifikatseinträge gefunden"
  fi
  info "Starte Traefik neu"
  docker restart infra-traefik >/dev/null
  ok "Traefik neu gestartet"
fi

ok "Domain-Löschung für ${DOMAIN} abgeschlossen"
