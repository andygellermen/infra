#!/usr/bin/env bash
set -euo pipefail

CURL_BIN="${EASY_AUTHOR_CURL_BIN:-curl}"
DOCKER_BIN="${EASY_AUTHOR_DOCKER_BIN:-docker}"
die(){ echo "FEHLER: $*" >&2; exit 1; }
domain=""; username=""; password_stdin=0; skip_restart=0
while [[ $# -gt 0 ]]; do
  case "$1" in --username=*) username="${1#*=}";; --password-stdin) password_stdin=1;; --skip-restart) skip_restart=1;; --help|-h) echo "Usage: $0 <domain> --username=<name> [--password-stdin] [--skip-restart]"; exit 0;; *) [[ -z "$domain" ]] || die "Nur eine Domain ist erlaubt."; domain="$1";; esac
  shift
done
[[ "$domain" =~ ^[a-z0-9]+([.-][a-z0-9]+)*$ ]] || die "Ungueltige Domain."
[[ "$username" =~ ^[A-Za-z0-9._-]+$ ]] || die "Ungueltiger oder fehlender Benutzername."
command -v "$CURL_BIN" >/dev/null || die "curl fehlt."
command -v python3 >/dev/null || die "python3 fehlt."
redirect="$($CURL_BIN -sS -o /dev/null -w '%{http_code}' "http://$domain/")"
[[ "$redirect" == 301 || "$redirect" == 308 ]] || die "HTTP leitet nicht dauerhaft auf HTTPS um ($redirect)."
unauth="$($CURL_BIN -sS -o /dev/null -w '%{http_code}' "https://$domain/")"
[[ "$unauth" == 401 ]] || die "Ungeschuetzter Zugriff wurde nicht abgewiesen ($unauth)."
if [[ "$password_stdin" -eq 1 ]]; then IFS= read -r password
elif [[ -n "${EASY_AUTHOR_SMOKE_PASSWORD:-}" ]]; then password="$EASY_AUTHOR_SMOKE_PASSWORD"
elif [[ -r /dev/tty ]]; then IFS= read -r -s -p "Basic-Auth-Kennwort: " password </dev/tty; printf '\n' >/dev/tty
else die "Kennwort ueber --password-stdin bereitstellen."; fi
[[ -n "$password" ]] || die "Kennwort ist leer."
auth_config="$(mktemp)"; body="$(mktemp)"
cleanup(){ rm -f "$auth_config" "$body"; unset password EASY_AUTHOR_SMOKE_PASSWORD; }
trap cleanup EXIT INT TERM
chmod 0600 "$auth_config" "$body"
escaped="${password//\\/\\\\}"; escaped="${escaped//\"/\\\"}"
printf 'user = "%s:%s"\n' "$username" "$escaped" > "$auth_config"; unset password escaped
status="$($CURL_BIN -sS --config "$auth_config" -o /dev/null -w '%{http_code}' "https://$domain/")"; [[ "$status" == 200 ]] || die "UI liefert $status."
$CURL_BIN -fsS --config "$auth_config" "https://$domain/api/health" > "$body"
python3 -c 'import json,sys; assert json.load(open(sys.argv[1]))["status"] == "ok"' "$body" || die "Health-Antwort ist ungueltig."
$CURL_BIN -fsS --config "$auth_config" "https://$domain/api/projects" > "$body"
project_id="$(python3 -c 'import json,sys; p=json.load(open(sys.argv[1])).get("projects",[]); assert p; print(p[0]["id"])' "$body")" || die "Keine stabile Projekt-ID gefunden."
$CURL_BIN -fsS --config "$auth_config" "https://$domain/api/kanban?include_done=true" > "$body"
python3 -c 'import json,sys; d=json.load(open(sys.argv[1])); assert all(x in d["items"] and x in d["totals"] for x in ("backlog","todo","in_progress","review","done"))' "$body" || die "Kanban besitzt nicht alle fuenf Phasen."
id="${domain//./-}"; api="easy-author-$id-api"
ports="$($DOCKER_BIN inspect -f '{{json .HostConfig.PortBindings}}' "$api")"
[[ "$ports" == null || "$ports" == '{}' ]] || die "API veroeffentlicht Host-Ports: $ports"
if [[ "$skip_restart" -eq 0 ]]; then
  $DOCKER_BIN restart "$api" >/dev/null
  ready=0
  for _ in $(seq 1 30); do
    if $CURL_BIN -fsS --config "$auth_config" "https://$domain/api/health" > "$body" 2>/dev/null; then ready=1; break; fi
    sleep 2
  done
  [[ "$ready" -eq 1 ]] || die "API wurde nach Neustart nicht bereit."
  $CURL_BIN -fsS --config "$auth_config" "https://$domain/api/projects" > "$body"
  python3 -c 'import json,sys; expected=sys.argv[2]; assert expected in [p["id"] for p in json.load(open(sys.argv[1]))["projects"]]' "$body" "$project_id" || die "Projekt-ID fehlt nach Neustart."
fi
echo "PASS: Auth, API, Kanban und Persistenz fuer $domain"
