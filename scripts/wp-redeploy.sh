#!/usr/bin/env bash
set -euo pipefail

INVENTORY="./ansible/inventory/hosts.ini"
PLAYBOOK="./ansible/playbooks/deploy-wordpress.yml"
HOSTVARS_DIR="./ansible/hostvars"

usage(){ echo "Usage: $0 <domain> [--check-only]"; }
die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

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

run_playbook_quiet() {
  local success_message="$1"
  shift
  local log_file
  log_file="$(mktemp)"

  if ANSIBLE_DISPLAY_SKIPPED_HOSTS=false "$@" >"$log_file" 2>&1; then
    local recap_line
    recap_line="$(grep -E '^localhost[[:space:]]*:' "$log_file" | tail -n1 || true)"
    [[ -n "$recap_line" ]] && info "$recap_line"
    ok "$success_message"
    rm -f "$log_file"
    return 0
  fi

  cat "$log_file" >&2
  rm -f "$log_file"
  return 1
}

run_post_redeploy_checks() {
  local container="$1" domain="$2"
  local headers="" first_status="" location="" public_result="" public_status="" public_redirects=""

  info "Starte WordPress-Selbsttest nach Redeploy"

  wait_for_container_running "$container" 60 || die "Container ist nach Redeploy nicht im Status 'running': $container"

  docker exec "$container" php -l /var/www/html/wp-config.php >/dev/null \
    || die "Syntaxfehler in laufender wp-config.php erkannt"
  ok "wp-config.php Syntaxcheck erfolgreich"

  local attempt
  for attempt in {1..20}; do
    headers="$(docker exec "$container" sh -lc "curl -sSI -H 'Host: ${domain}' -H 'X-Forwarded-Proto: https' http://127.0.0.1/" 2>/dev/null || true)"
    first_status="$(printf '%s\n' "$headers" | awk 'toupper($1) ~ /^HTTP/ {print $2; exit}')"
    location="$(printf '%s\n' "$headers" | awk 'tolower($1)==\"location:\" {sub(/\r$/, \"\", $2); print $2; exit}')"
    [[ -n "$first_status" ]] && break
    sleep 2
  done

  [[ -n "$first_status" ]] || die "Kein interner HTTP-Response vom WordPress-Container erhalten"

  case "$first_status" in
    200|301|302|303|307|308) ;;
    *) die "Unerwarteter interner HTTP-Status nach Redeploy: ${first_status}" ;;
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

CHECK_ONLY=0
[[ $# -ge 1 ]] || { usage; exit 1; }
DOMAIN="$1"; shift || true
[[ "${1:-}" == "--check-only" ]] && CHECK_ONLY=1

require_cmd curl
require_cmd dig
require_cmd docker
require_cmd ansible-playbook

HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
[[ -f "$HOSTVARS_FILE" ]] || die "Hostvars nicht gefunden: $HOSTVARS_FILE"
CONTAINER="wp-${DOMAIN//./-}"

host_ip="$(curl -fsSL https://api.ipify.org || true)"
dns_ip="$(dig +short A "$DOMAIN" | head -n1)"
[[ -n "$host_ip" && -n "$dns_ip" ]] || die "DNS/IP Prüfung fehlgeschlagen"
[[ "$host_ip" == "$dns_ip" ]] || die "DNS mismatch: $DOMAIN -> $dns_ip (erwartet $host_ip)"
ok "DNS OK für $DOMAIN"

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only abgeschlossen"
  exit 0
fi

run_playbook_quiet "WordPress-Redeploy abgeschlossen" \
  ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$PLAYBOOK" \
  || die "WordPress-Redeploy fehlgeschlagen"
run_post_redeploy_checks "$CONTAINER" "$DOMAIN"
