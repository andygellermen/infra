#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXPORT_SCRIPT="$ROOT_DIR/scripts/wildcard-export.sh"
DISTRIBUTE_SCRIPT="$ROOT_DIR/scripts/wildcard-distribute.sh"
DEFAULT_STATE_DIR="$ROOT_DIR/data/wildcard-distribution-state"
FIELD_SEPARATOR=$'\x1f'
source "$ROOT_DIR/scripts/lib/error-notify.sh"
setup_error_notification "$(basename "$0")" "$ROOT_DIR" "$0 $*"

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/wildcard-distribute-on-change.sh <apex-domain> [--config /pfad/export.yml] [--state-dir /pfad/state] [--dry-run] [--force]
  ./scripts/wildcard-distribute-on-change.sh --all [--config /pfad/export.yml] [--state-dir /pfad/state] [--dry-run] [--force]
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

read_state_value() {
  local file="$1"
  local key="$2"
  [[ -f "$file" ]] || return 1
  awk -F'=' -v wanted="$key" '$1 == wanted { print substr($0, index($0, "=") + 1); exit }' "$file"
}

build_state_file_path() {
  local domain="$1"
  local acme_file="$2"
  local acme_marker hash

  acme_marker="${acme_file:-default-acme}"
  hash="$(printf '%s' "$acme_marker" | openssl dgst -sha256 | awk '{print $NF}' | cut -c1-16)"
  printf '%s/%s.%s.state\n' "$STATE_DIR" "$domain" "$hash"
}

compute_cert_fingerprint() {
  local cert_file="$1"
  openssl x509 -in "$cert_file" -outform DER | openssl dgst -sha256 | awk '{print $NF}'
}

read_cert_expiry() {
  local cert_file="$1"
  openssl x509 -in "$cert_file" -noout -enddate | sed 's/^notAfter=//'
}

validate_exported_cert() {
  local cert_file="$1"
  local domain="$2"

  openssl x509 -in "$cert_file" -noout >/dev/null 2>&1 || die "Exportiertes Zertifikat für ${domain} ist ungültig"
  openssl x509 -in "$cert_file" -noout -ext subjectAltName 2>/dev/null | grep -Fq "DNS:*.${domain}" \
    || die "Exportiertes Zertifikat enthält keine Wildcard-SAN für *.${domain}"
}

write_state_file() {
  local state_file="$1"
  local domain="$2"
  local acme_file="$3"
  local fingerprint="$4"
  local not_after="$5"
  local tmp_file

  tmp_file="$(mktemp "${state_file}.tmp.XXXXXX")"
  cat > "$tmp_file" <<EOF
domain=${domain}
acme_file=${acme_file}
fingerprint=${fingerprint}
not_after=${not_after}
updated_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
EOF
  mv "$tmp_file" "$state_file"
}

CONFIG_FILE=""
STATE_DIR="$DEFAULT_STATE_DIR"
RUN_ALL=0
DRY_RUN=0
FORCE=0
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
    --state-dir=*)
      STATE_DIR="${1#*=}"
      shift
      ;;
    --state-dir)
      shift
      [[ $# -gt 0 ]] || die "Fehlender Wert für --state-dir"
      STATE_DIR="$1"
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
    --force)
      FORCE=1
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

require_cmd python3
require_cmd openssl
require_cmd mktemp
[[ -x "$EXPORT_SCRIPT" ]] || die "Export-Skript fehlt oder ist nicht ausführbar: $EXPORT_SCRIPT"
[[ -x "$DISTRIBUTE_SCRIPT" ]] || die "Distributions-Skript fehlt oder ist nicht ausführbar: $DISTRIBUTE_SCRIPT"

if [[ "$DRY_RUN" -eq 0 ]]; then
  mkdir -p "$STATE_DIR"
fi

LIST_ARGS=(--list-exports)
RUN_ARGS=()

if [[ -n "$CONFIG_FILE" ]]; then
  LIST_ARGS+=(--config "$CONFIG_FILE")
  RUN_ARGS+=(--config "$CONFIG_FILE")
fi

if [[ "$RUN_ALL" -eq 1 ]]; then
  LIST_ARGS+=(--all)
else
  LIST_ARGS+=("${ARGS[0]}")
fi

TMP_DIRS=()
cleanup() {
  local tmp_dir
  for tmp_dir in "${TMP_DIRS[@]:-}"; do
    [[ -n "$tmp_dir" ]] || continue
    rm -rf "$tmp_dir"
  done
}
trap cleanup EXIT

EXPORT_LINES=()
while IFS= read -r line; do
  [[ -n "$line" ]] || continue
  EXPORT_LINES+=("$line")
done < <("$DISTRIBUTE_SCRIPT" "${LIST_ARGS[@]}")

[[ ${#EXPORT_LINES[@]} -gt 0 ]] || die "Keine Wildcard-Exporte zur Prüfung gefunden"

TOTAL_EXPORTS=0
CHANGED_EXPORTS=0
SKIPPED_EXPORTS=0

for line in "${EXPORT_LINES[@]}"; do
  IFS="$FIELD_SEPARATOR" read -r domain acme_file <<< "$line"
  TOTAL_EXPORTS=$((TOTAL_EXPORTS + 1))

  export_dir="$(mktemp -d "/tmp/wildcard-on-change-${domain}.XXXXXX")"
  TMP_DIRS+=("$export_dir")

  export_args=("$domain" "--output-dir=$export_dir")
  if [[ -n "$acme_file" ]]; then
    export_args+=("--acme-file=$acme_file")
  fi
  "$EXPORT_SCRIPT" "${export_args[@]}" >/dev/null

  cert_file="$export_dir/fullchain.pem"
  validate_exported_cert "$cert_file" "$domain"
  fingerprint="$(compute_cert_fingerprint "$cert_file")"
  not_after="$(read_cert_expiry "$cert_file")"
  state_file="$(build_state_file_path "$domain" "$acme_file")"
  previous_fingerprint="$(read_state_value "$state_file" fingerprint 2>/dev/null || true)"

  if [[ "$FORCE" -eq 0 && -n "$previous_fingerprint" && "$previous_fingerprint" == "$fingerprint" ]]; then
    SKIPPED_EXPORTS=$((SKIPPED_EXPORTS + 1))
    info "Kein Zertifikatswechsel für ${domain}; Verteilung wird übersprungen"
    continue
  fi

  change_reason="neues Zertifikat erkannt"
  if [[ "$FORCE" -eq 1 ]]; then
    change_reason="Verteilung per --force angefordert"
  elif [[ -n "$previous_fingerprint" ]]; then
    change_reason="Zertifikats-Fingerprint hat sich geändert"
  fi

  CHANGED_EXPORTS=$((CHANGED_EXPORTS + 1))
  if [[ "$DRY_RUN" -eq 1 ]]; then
    info "[dry-run] ${domain}: ${change_reason}; Verteilung würde ausgelöst"
    continue
  fi

  info "${domain}: ${change_reason}; starte Verteilung"
  "$DISTRIBUTE_SCRIPT" "$domain" "${RUN_ARGS[@]}"
  write_state_file "$state_file" "$domain" "$acme_file" "$fingerprint" "$not_after"
done

if [[ "$DRY_RUN" -eq 1 ]]; then
  ok "Dry-Run abgeschlossen: ${TOTAL_EXPORTS} Export(e) geprüft, ${CHANGED_EXPORTS} Verteilung(en) würden ausgelöst, ${SKIPPED_EXPORTS} unverändert"
else
  ok "Wildcard-Change-Check abgeschlossen: ${TOTAL_EXPORTS} Export(e) geprüft, ${CHANGED_EXPORTS} Verteilung(en) ausgelöst, ${SKIPPED_EXPORTS} unverändert"
fi
