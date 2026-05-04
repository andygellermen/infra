#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXPORT_SCRIPT="$ROOT_DIR/scripts/wildcard-export.sh"
source "$ROOT_DIR/scripts/lib/error-notify.sh"
setup_error_notification "$(basename "$0")" "$ROOT_DIR" "$0 $*"

pick_default_config() {
  local candidate
  for candidate in \
    "$ROOT_DIR/ansible/wildcards/export.yml" \
    "$ROOT_DIR/ansible/wildcards/export.example.yml" \
    "$ROOT_DIR/ansible/wildcards/distribution.yml" \
    "$ROOT_DIR/ansible/wildcards/distribution.example.yml"
  do
    if [[ -f "$candidate" ]]; then
      printf '%s\n' "$candidate"
      return 0
    fi
  done

  printf '%s\n' "$ROOT_DIR/ansible/wildcards/export.example.yml"
}

DEFAULT_CONFIG="$(pick_default_config)"

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/wildcard-distribute.sh <apex-domain> [--config /pfad/export.yml] [--dry-run]
  ./scripts/wildcard-distribute.sh --all [--config /pfad/export.yml] [--dry-run]
  ./scripts/wildcard-distribute.sh --all [--config /pfad/export.yml] --list-exports
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

normalize_domain() {
  local domain="$1"
  if [[ "$domain" =~ ^[A-Za-z0-9.-]+$ ]]; then
    printf '%s\n' "$domain" | tr '[:upper:]' '[:lower:]'
  else
    require_cmd idn
    idn --quiet --uts46 "$domain" | tr '[:upper:]' '[:lower:]'
  fi
}

upload_file_via_ssh() {
  local local_file="$1"
  local remote_path="$2"
  shift 2

  ssh "$@" "cat > \"$remote_path\"" < "$local_file"
}

validate_export_artifacts() {
  local export_dir="$1"
  local domain="$2"
  local cert_pub key_pub

  cert_pub="$(mktemp "/tmp/wildcard-cert-pub.${domain}.XXXXXX")"
  key_pub="$(mktemp "/tmp/wildcard-key-pub.${domain}.XXXXXX")"

  if ! openssl x509 -in "$export_dir/fullchain.pem" -noout >/dev/null 2>&1; then
    rm -f "$cert_pub" "$key_pub"
    die "Exportiertes Zertifikat für ${domain} ist ungültig"
  fi

  if ! openssl pkey -in "$export_dir/privkey.pem" -pubout -out "$key_pub" >/dev/null 2>&1; then
    rm -f "$cert_pub" "$key_pub"
    die "Exportierter Private Key für ${domain} ist ungültig"
  fi

  if ! openssl x509 -in "$export_dir/fullchain.pem" -pubkey -noout > "$cert_pub" 2>/dev/null; then
    rm -f "$cert_pub" "$key_pub"
    die "Öffentlicher Schlüssel aus Zertifikat für ${domain} konnte nicht gelesen werden"
  fi

  if ! cmp -s "$cert_pub" "$key_pub"; then
    rm -f "$cert_pub" "$key_pub"
    die "Zertifikat und Private Key passen für ${domain} nicht zusammen"
  fi

  if ! openssl x509 -in "$export_dir/fullchain.pem" -noout -ext subjectAltName 2>/dev/null | grep -Fq "DNS:*.${domain}"; then
    rm -f "$cert_pub" "$key_pub"
    die "Exportiertes Zertifikat enthält keine Wildcard-SAN für *.${domain}"
  fi

  rm -f "$cert_pub" "$key_pub"
}

CONFIG_FILE="$DEFAULT_CONFIG"
DRY_RUN=0
RUN_ALL=0
LIST_EXPORTS=0
ARGS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --config=*)
      CONFIG_FILE="${1#*=}"
      shift
      ;;
    --config)
      shift
      [[ $# -gt 0 ]] || die "Fehlender Wert für --config"
      CONFIG_FILE="$1"
      shift
      ;;
    --all)
      RUN_ALL=1
      shift
      ;;
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --list-exports)
      LIST_EXPORTS=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      ARGS+=("$1")
      shift
      ;;
  esac
done

if [[ "$RUN_ALL" -eq 1 && ${#ARGS[@]} -gt 0 ]]; then
  die "--all kann nicht mit einer einzelnen Domain kombiniert werden"
fi

if [[ "$RUN_ALL" -eq 0 && ${#ARGS[@]} -ne 1 ]]; then
  usage
  exit 1
fi

if [[ "$LIST_EXPORTS" -eq 0 ]]; then
  require_cmd python3
  require_cmd ssh
  require_cmd openssl
fi

DOMAIN=""
if [[ "$RUN_ALL" -eq 0 ]]; then
  DOMAIN="$(normalize_domain "${ARGS[0]}")"
fi
[[ -f "$CONFIG_FILE" ]] || die "Distributions-Konfiguration fehlt: $CONFIG_FILE"

TMP_DIRS=()
cleanup() {
  local tmp_dir
  for tmp_dir in "${TMP_DIRS[@]:-}"; do
    [[ -n "$tmp_dir" ]] || continue
    rm -rf "$tmp_dir"
  done
}
trap cleanup EXIT

MODE="domain"
if [[ "$RUN_ALL" -eq 1 ]]; then
  MODE="all"
fi

PARSER_ENGINE=""
if python3 -c 'import yaml' >/dev/null 2>&1; then
  PARSER_ENGINE="python"
elif command -v ruby >/dev/null 2>&1; then
  PARSER_ENGINE="ruby"
else
  die "Für die Distributions-Konfiguration wird entweder PyYAML oder Ruby mit YAML-Support benötigt"
fi

FIELD_SEPARATOR=$'\x1f'
TARGET_LINES=()
if [[ "$PARSER_ENGINE" == "python" ]]; then
while IFS= read -r line; do
  TARGET_LINES+=("$line")
done < <(python3 - "$MODE" "$DOMAIN" "$CONFIG_FILE" <<'PY'
import sys
try:
    import yaml
except Exception as exc:
    print(f"❌ PyYAML fehlt: {exc}", file=sys.stderr)
    sys.exit(1)

mode = sys.argv[1]
selected_domain = sys.argv[2]
config_file = sys.argv[3]
with open(config_file, "r", encoding="utf-8") as fh:
    payload = yaml.safe_load(fh) or {}

rows = []

for entry in payload.get("wildcard_exports") or []:
    source_domain = (entry.get("source_domain") or "").strip().lower()
    if not source_domain:
        continue
    if mode != "all" and source_domain != selected_domain:
        continue
    acme_file = (entry.get("acme_file") or "").strip()
    for target in entry.get("targets") or []:
        host = (target.get("host") or "").strip()
        user = str(target.get("user") or "root").strip() or "root"
        port = str(target.get("port") or 22).strip() or "22"
        identity_file = (target.get("identity_file") or "").strip()
        known_hosts_file = (target.get("known_hosts_file") or "").strip()
        remote_dir = (target.get("remote_dir") or "").strip()
        remote_fullchain_path = (target.get("remote_fullchain_path") or "").strip()
        remote_privkey_path = (target.get("remote_privkey_path") or "").strip()
        if remote_dir:
            remote_fullchain_path = remote_fullchain_path or f"{remote_dir.rstrip('/')}/fullchain.pem"
            remote_privkey_path = remote_privkey_path or f"{remote_dir.rstrip('/')}/privkey.pem"
        fullchain_mode = str(target.get("fullchain_mode") or "0644").strip() or "0644"
        privkey_mode = str(target.get("privkey_mode") or "0600").strip() or "0600"
        owner = str(target.get("owner") or "").strip()
        group = str(target.get("group") or "").strip()
        post_deploy_command = str(target.get("post_deploy_command") or "").strip()
        label = (target.get("name") or f"{user}@{host}").strip()
        if host and remote_fullchain_path and remote_privkey_path:
            rows.append([
                source_domain,
                acme_file,
                label,
                host,
                user,
                port,
                identity_file,
                known_hosts_file,
                remote_fullchain_path,
                remote_privkey_path,
                fullchain_mode,
                privkey_mode,
                owner,
                group,
                post_deploy_command,
            ])

for entry in payload.get("wildcard_distribution") or []:
    source_domain = (entry.get("wildcard_domain") or "").strip().lower()
    if not source_domain:
        continue
    if mode != "all" and source_domain != selected_domain:
        continue
    for target in entry.get("targets") or []:
        host = (target.get("host") or "").strip()
        user = str(target.get("user") or "root").strip() or "root"
        port = str(target.get("port") or 22).strip() or "22"
        remote_dir = (target.get("remote_dir") or "").strip()
        if host and remote_dir:
            rows.append([
                source_domain,
                "",
                f"{user}@{host}",
                host,
                user,
                port,
                "",
                "",
                f"{remote_dir.rstrip('/')}/fullchain.pem",
                f"{remote_dir.rstrip('/')}/privkey.pem",
                "0644",
                "0600",
                "",
                "",
                "",
            ])

for row in rows:
    print("\x1f".join(row))
PY
)
else
while IFS= read -r line; do
  TARGET_LINES+=("$line")
done < <(ruby - "$MODE" "$DOMAIN" "$CONFIG_FILE" <<'RUBY'
require "yaml"

mode = ARGV[0]
selected_domain = ARGV[1]
config_file = ARGV[2]
payload = YAML.load_file(config_file) || {}
rows = []

(payload["wildcard_exports"] || []).each do |entry|
  source_domain = entry["source_domain"].to_s.strip.downcase
  next if source_domain.empty?
  next if mode != "all" && source_domain != selected_domain

  acme_file = entry["acme_file"].to_s.strip
  (entry["targets"] || []).each do |target|
    host = target["host"].to_s.strip
    user = target["user"].to_s.strip
    user = "root" if user.empty?
    port = target["port"].to_s.strip
    port = "22" if port.empty?
    identity_file = target["identity_file"].to_s.strip
    known_hosts_file = target["known_hosts_file"].to_s.strip
    remote_dir = target["remote_dir"].to_s.strip
    remote_fullchain_path = target["remote_fullchain_path"].to_s.strip
    remote_privkey_path = target["remote_privkey_path"].to_s.strip
    if !remote_dir.empty?
      remote_fullchain_path = "#{remote_dir.sub(%r{/$}, "")}/fullchain.pem" if remote_fullchain_path.empty?
      remote_privkey_path = "#{remote_dir.sub(%r{/$}, "")}/privkey.pem" if remote_privkey_path.empty?
    end
    fullchain_mode = target["fullchain_mode"].to_s.strip
    fullchain_mode = "0644" if fullchain_mode.empty?
    privkey_mode = target["privkey_mode"].to_s.strip
    privkey_mode = "0600" if privkey_mode.empty?
    owner = target["owner"].to_s.strip
    group = target["group"].to_s.strip
    post_deploy_command = target["post_deploy_command"].to_s.strip
    label = target["name"].to_s.strip
    label = "#{user}@#{host}" if label.empty?

    next if host.empty? || remote_fullchain_path.empty? || remote_privkey_path.empty?

    rows << [
      source_domain,
      acme_file,
      label,
      host,
      user,
      port,
      identity_file,
      known_hosts_file,
      remote_fullchain_path,
      remote_privkey_path,
      fullchain_mode,
      privkey_mode,
      owner,
      group,
      post_deploy_command,
    ]
  end
end

(payload["wildcard_distribution"] || []).each do |entry|
  source_domain = entry["wildcard_domain"].to_s.strip.downcase
  next if source_domain.empty?
  next if mode != "all" && source_domain != selected_domain

  (entry["targets"] || []).each do |target|
    host = target["host"].to_s.strip
    user = target["user"].to_s.strip
    user = "root" if user.empty?
    port = target["port"].to_s.strip
    port = "22" if port.empty?
    remote_dir = target["remote_dir"].to_s.strip
    next if host.empty? || remote_dir.empty?

    remote_dir = remote_dir.sub(%r{/$}, "")
    rows << [
      source_domain,
      "",
      "#{user}@#{host}",
      host,
      user,
      port,
      "",
      "",
      "#{remote_dir}/fullchain.pem",
      "#{remote_dir}/privkey.pem",
      "0644",
      "0600",
      "",
      "",
      "",
    ]
  end
end

rows.each do |row|
  puts row.join("\x1f")
end
RUBY
)
fi

if [[ "$RUN_ALL" -eq 1 ]]; then
  [[ ${#TARGET_LINES[@]} -gt 0 ]] || die "Keine Verteilziele in ${CONFIG_FILE} gefunden"
else
  [[ ${#TARGET_LINES[@]} -gt 0 ]] || die "Keine Verteilziele für ${DOMAIN} in ${CONFIG_FILE} gefunden"
fi

if [[ "$LIST_EXPORTS" -eq 1 ]]; then
  seen_exports=$'\n'
  for line in "${TARGET_LINES[@]}"; do
    IFS="$FIELD_SEPARATOR" read -r source_domain acme_file _rest <<< "$line"
    export_key="${source_domain}${FIELD_SEPARATOR}${acme_file}"
    case "$seen_exports" in
      *$'\n'"$export_key"$'\n'*) ;;
      *)
        printf '%s%s%s\n' "$source_domain" "$FIELD_SEPARATOR" "$acme_file"
        seen_exports+="${export_key}"$'\n'
        ;;
    esac
  done
  exit 0
fi

CURRENT_EXPORT_KEY=""
CURRENT_EXPORT_DIR=""
TARGET_COUNT=0

for line in "${TARGET_LINES[@]}"; do
  IFS="$FIELD_SEPARATOR" read -r source_domain acme_file label host user port identity_file known_hosts_file remote_fullchain_path remote_privkey_path fullchain_mode privkey_mode owner group post_deploy_command <<< "$line"

  export_key="${source_domain}|${acme_file}"
  if [[ "$export_key" != "$CURRENT_EXPORT_KEY" ]]; then
    CURRENT_EXPORT_KEY="$export_key"
    CURRENT_EXPORT_DIR="$(mktemp -d "/tmp/wildcard-distribute-${source_domain}.XXXXXX")"
    TMP_DIRS+=("$CURRENT_EXPORT_DIR")

    if [[ "$DRY_RUN" -eq 1 ]]; then
      if [[ -n "$acme_file" ]]; then
        info "[dry-run] Export würde ${source_domain} aus ${acme_file} laden"
      else
        info "[dry-run] Export würde ${source_domain} aus der Standard-ACME-Datei laden"
      fi
    else
      export_args=("$source_domain" "--output-dir=$CURRENT_EXPORT_DIR")
      if [[ -n "$acme_file" ]]; then
        export_args+=("--acme-file=$acme_file")
      fi
      "$EXPORT_SCRIPT" "${export_args[@]}" >/dev/null
      validate_export_artifacts "$CURRENT_EXPORT_DIR" "$source_domain"
    fi
  fi

  TARGET_COUNT=$((TARGET_COUNT + 1))

  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "[dry-run] ${source_domain} -> ${label} (${user}@${host}:${remote_fullchain_path}, ${remote_privkey_path})"
    if [[ -n "$post_deploy_command" ]]; then
      info "[dry-run] Post-Deploy auf ${label}: ${post_deploy_command}"
    fi
    continue
  fi

  info "Verteile Wildcard-Zertifikat ${source_domain} an ${label}"

  ssh_args=(-p "$port" -o BatchMode=yes)
  if [[ -n "$identity_file" ]]; then
    ssh_args+=(-i "$identity_file" -o IdentitiesOnly=yes)
  fi
  if [[ -n "$known_hosts_file" ]]; then
    ssh_args+=(-o "UserKnownHostsFile=${known_hosts_file}")
  fi
  ssh_args+=("${user}@${host}")

  remote_fullchain_tmp="${remote_fullchain_path}.tmp.$$"
  remote_privkey_tmp="${remote_privkey_path}.tmp.$$"

  upload_file_via_ssh "$CURRENT_EXPORT_DIR/fullchain.pem" "$remote_fullchain_tmp" "${ssh_args[@]}"
  upload_file_via_ssh "$CURRENT_EXPORT_DIR/privkey.pem" "$remote_privkey_tmp" "${ssh_args[@]}"

  ssh "${ssh_args[@]}" /bin/sh -s -- \
    "$remote_fullchain_path" \
    "$remote_privkey_path" \
    "$remote_fullchain_tmp" \
    "$remote_privkey_tmp" \
    "$fullchain_mode" \
    "$privkey_mode" \
    "$owner" \
    "$group" \
    "$post_deploy_command" <<'REMOTE'
set -eu

remote_fullchain_path="$1"
remote_privkey_path="$2"
remote_fullchain_tmp="$3"
remote_privkey_tmp="$4"
fullchain_mode="$5"
privkey_mode="$6"
owner="$7"
group="$8"
post_deploy_command="$9"

mkdir -p "$(dirname "$remote_fullchain_path")"
mkdir -p "$(dirname "$remote_privkey_path")"

mv "$remote_fullchain_tmp" "$remote_fullchain_path"
mv "$remote_privkey_tmp" "$remote_privkey_path"
chmod "$fullchain_mode" "$remote_fullchain_path"
chmod "$privkey_mode" "$remote_privkey_path"

if [ -n "$owner" ] || [ -n "$group" ]; then
  chown_target="$owner"
  if [ -n "$group" ]; then
    chown_target="${chown_target}:$group"
  fi
  if [ -z "$owner" ]; then
    chown_target=":$group"
  fi
  chown "$chown_target" "$remote_fullchain_path" "$remote_privkey_path"
fi

if [ -n "$post_deploy_command" ]; then
  /bin/sh -lc "$post_deploy_command"
fi
REMOTE
done

if [[ "$DRY_RUN" -eq 1 ]]; then
  ok "Dry-Run abgeschlossen: ${TARGET_COUNT} Zielsystem(e) geprüft"
else
  if [[ "$RUN_ALL" -eq 1 ]]; then
    ok "Wildcard-Zertifikate für ${TARGET_COUNT} Zielsystem(e) verteilt"
  else
    ok "Wildcard-Zertifikat für ${DOMAIN} auf ${TARGET_COUNT} Zielsystem(e) verteilt"
  fi
fi
