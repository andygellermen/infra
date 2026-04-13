#!/usr/bin/env bash

extract_top_level_secret() {
  local key="$1" file="$2"
  awk -F': ' -v k="$key" '
    $0 ~ "^[A-Za-z0-9_]+:" && $1 == k {
      value = substr($0, index($0, ":") + 1)
      sub(/^[[:space:]]+/, "", value)
      sub(/[[:space:]]+$/, "", value)
      gsub(/^"/, "", value)
      gsub(/"$/, "", value)
      gsub(/^'\''/, "", value)
      gsub(/'\''$/, "", value)
      print value
      exit
    }
  ' "$file"
}

setup_error_notification() {
  local script_name="$1"
  local root_dir="$2"
  local command_line="${3:-$0}"

  INFRA_NOTIFY_SCRIPT_NAME="$script_name"
  INFRA_NOTIFY_ROOT_DIR="$root_dir"
  INFRA_NOTIFY_COMMAND_LINE="$command_line"
  INFRA_NOTIFY_SECRETS_FILE="$root_dir/ansible/secrets/secrets.yml"
  INFRA_NOTIFY_HOSTNAME="$(hostname -f 2>/dev/null || hostname 2>/dev/null || printf 'unknown-host')"
  INFRA_NOTIFY_TIMESTAMP="$(date '+%Y-%m-%d %H:%M:%S %Z')"
  INFRA_NOTIFY_LOG_DIR="$root_dir/logs/error-notify"
  INFRA_NOTIFY_LOG_FILE="$INFRA_NOTIFY_LOG_DIR/${script_name}-$(date +%Y%m%d-%H%M%S)-$$.log"

  mkdir -p "$INFRA_NOTIFY_LOG_DIR"
  exec > >(tee -a "$INFRA_NOTIFY_LOG_FILE") 2>&1

  trap '__infra_notify_on_exit $?' EXIT
}

__infra_notify_on_exit() {
  local rc="$1"
  trap - EXIT

  if [[ "$rc" -eq 0 ]]; then
    rm -f "$INFRA_NOTIFY_LOG_FILE"
    exit 0
  fi

  if [[ ! -f "$INFRA_NOTIFY_SECRETS_FILE" ]]; then
    echo "⚠️  Fehler-Notification übersprungen: Secrets-Datei fehlt: $INFRA_NOTIFY_SECRETS_FILE" >&2
    exit "$rc"
  fi

  local notify_to notify_from smtp_host smtp_port smtp_user smtp_password smtp_secure smtp_require_tls subject_prefix
  notify_to="$(extract_top_level_secret infra_error_notify_to "$INFRA_NOTIFY_SECRETS_FILE")"
  notify_from="$(extract_top_level_secret infra_error_notify_from "$INFRA_NOTIFY_SECRETS_FILE")"
  smtp_host="$(extract_top_level_secret ses_smtp_host "$INFRA_NOTIFY_SECRETS_FILE")"
  smtp_port="$(extract_top_level_secret ses_smtp_port "$INFRA_NOTIFY_SECRETS_FILE")"
  smtp_user="$(extract_top_level_secret ses_smtp_user "$INFRA_NOTIFY_SECRETS_FILE")"
  smtp_password="$(extract_top_level_secret ses_smtp_password "$INFRA_NOTIFY_SECRETS_FILE")"
  smtp_secure="$(extract_top_level_secret ses_smtp_secure "$INFRA_NOTIFY_SECRETS_FILE")"
  smtp_require_tls="$(extract_top_level_secret ses_smtp_requireTLS "$INFRA_NOTIFY_SECRETS_FILE")"
  subject_prefix="$(extract_top_level_secret infra_error_notify_subject_prefix "$INFRA_NOTIFY_SECRETS_FILE")"

  [[ -n "$notify_to" ]] || exit "$rc"
  [[ -n "$notify_from" ]] || notify_from="$(extract_top_level_secret ses_from "$INFRA_NOTIFY_SECRETS_FILE")"
  [[ -n "$notify_from" ]] || notify_from="Infra <noreply@localhost>"
  [[ -n "$smtp_host" ]] || smtp_host="email-smtp.eu-north-1.amazonaws.com"
  [[ -n "$smtp_port" ]] || smtp_port="587"
  [[ -n "$smtp_require_tls" ]] || smtp_require_tls="true"
  [[ -n "$subject_prefix" ]] || subject_prefix="[infra]"

  local summary tail_file
  summary="$(cat <<EOF
Zeit: ${INFRA_NOTIFY_TIMESTAMP}
Host: ${INFRA_NOTIFY_HOSTNAME}
Skript: ${INFRA_NOTIFY_SCRIPT_NAME}
Exit-Code: ${rc}
Arbeitsverzeichnis: ${PWD}
Befehl: ${INFRA_NOTIFY_COMMAND_LINE}
Logdatei: ${INFRA_NOTIFY_LOG_FILE}
EOF
)"

  tail_file="$(mktemp /tmp/infra-error-tail.XXXXXX)"
  tail -n 200 "$INFRA_NOTIFY_LOG_FILE" > "$tail_file" || true

  python3 "$INFRA_NOTIFY_ROOT_DIR/scripts/send-error-notification.py" \
    --smtp-host "$smtp_host" \
    --smtp-port "$smtp_port" \
    --smtp-user "$smtp_user" \
    --smtp-password "$smtp_password" \
    --smtp-secure "$smtp_secure" \
    --smtp-require-tls "$smtp_require_tls" \
    --mail-from "$notify_from" \
    --mail-to "$notify_to" \
    --subject "${subject_prefix} Fehler in ${INFRA_NOTIFY_SCRIPT_NAME} auf ${INFRA_NOTIFY_HOSTNAME}" \
    --summary "$summary" \
    --log-file "$tail_file" \
    || echo "⚠️  Fehler-Notification konnte nicht per SMTP versendet werden" >&2

  rm -f "$tail_file"
  exit "$rc"
}
