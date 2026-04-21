#!/usr/bin/env bash

resolve_ipv4_addresses_system() {
  local domain="$1"

  python3 - "$domain" <<'PY'
import ipaddress
import socket
import sys

domain = sys.argv[1]

try:
    infos = socket.getaddrinfo(domain, None, family=socket.AF_INET, type=socket.SOCK_STREAM)
except socket.gaierror:
    sys.exit(0)

addresses = []
for info in infos:
    try:
        addr = info[4][0]
        ipaddress.IPv4Address(addr)
    except Exception:
        continue
    if addr not in addresses:
        addresses.append(addr)

for addr in addresses:
    print(addr)
PY
}

lookup_non_a_records() {
  local domain="$1"
  local rr output joined

  for rr in CNAME AAAA MX TXT; do
    output="$(dig +short "$rr" "$domain" 2>/dev/null | sed '/^[[:space:]]*$/d' || true)"
    [[ -n "$output" ]] || continue
    joined="$(printf '%s\n' "$output" | paste -sd',' -)"
    printf '%s=%s\n' "$rr" "$joined"
  done
}

join_by_comma_space() {
  local first="${1:-}"
  shift || true
  printf '%s' "$first"
  for item in "$@"; do
    printf ', %s' "$item"
  done
}

verify_domain_resolves_to_host_ipv4() {
  local domain="$1" host_ip="$2"
  local dns_ips=() other_records=()
  local line

  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    dns_ips+=("$line")
  done < <(resolve_ipv4_addresses_system "$domain")

  if [[ ${#dns_ips[@]} -eq 0 ]]; then
    while IFS= read -r line; do
      [[ -n "$line" ]] || continue
      other_records+=("$line")
    done < <(lookup_non_a_records "$domain")

    if [[ ${#other_records[@]} -gt 0 ]]; then
      die "Kein IPv4-A-Record für ${domain} gefunden. Vorhandene Records: $(join_by_comma_space "${other_records[@]}"). Der Name existiert damit bereits explizit; ein Wildcard-A-Record greift für diesen Hostnamen dann nicht. Bitte expliziten A-Record setzen oder kollidierende Records prüfen."
    fi

    die "Kein IPv4-A-Record für ${domain} gefunden."
  fi

  local dns_joined
  dns_joined="$(join_by_comma_space "${dns_ips[@]}")"

  local dns_ip
  for dns_ip in "${dns_ips[@]}"; do
    if [[ "$dns_ip" == "$host_ip" ]]; then
      ok "DNS OK: ${domain} -> ${dns_joined}"
      return 0
    fi
  done

  die "DNS-Fehler: ${domain} zeigt auf ${dns_joined}, erwartet wird ${host_ip}."
}
