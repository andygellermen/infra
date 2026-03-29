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

manage_wp_auth() {
  local hostvars_file="$1" domain="$2" auth_test_file="$3"

  python3 - "$hostvars_file" "$domain" "$auth_test_file" <<'PY'
import getpass
import json
import subprocess
import sys

try:
    import yaml
except Exception as exc:
    print(f"❌ PyYAML fehlt für die Auth-Verwaltung: {exc}", file=sys.stderr)
    sys.exit(1)

hostvars_file, domain, auth_test_file = sys.argv[1:4]
try:
    tty_in = open("/dev/tty", "r", encoding="utf-8", errors="replace")
except OSError:
    print("❌ Kein interaktives Terminal verfügbar (/dev/tty).", file=sys.stderr)
    sys.exit(1)

with open(hostvars_file, "r", encoding="utf-8") as fh:
    data = yaml.safe_load(fh) or {}

entries = data.get("wp_basic_auth_scopes") or []
if not isinstance(entries, list):
    entries = []

scope_meta = {
    "frontend": {
        "label": "gesamte Website (/)",
        "url": f"https://{domain}/",
        "realm": "Protected Website",
    },
    "admin": {
        "label": "WordPress-Admin (/wp-admin, /wp-login.php)",
        "url": f"https://{domain}/wp-login.php",
        "realm": "Protected Admin",
    },
    "api": {
        "label": "WordPress-API (/wp-json)",
        "url": f"https://{domain}/wp-json",
        "realm": "Protected API",
    },
}
scope_order = ["frontend", "admin", "api"]

def prompt_yes_no(question: str, default: bool = False) -> bool:
    suffix = "[Y/n]" if default else "[y/N]"
    while True:
        print(f"{question} {suffix} ", end="", flush=True)
        answer = tty_in.readline()
        if answer == "":
            raise EOFError("EOF when reading from /dev/tty")
        answer = answer.strip().lower()
        if not answer:
            return default
        if answer in {"y", "yes", "j", "ja"}:
            return True
        if answer in {"n", "no", "nein"}:
            return False
        print("Bitte mit yes oder no antworten.")

def prompt_nonempty(question: str, default: str = "") -> str:
    while True:
        prompt = question
        if default:
            prompt += f" [{default}]"
        prompt += ": "
        print(prompt, end="", flush=True)
        value = tty_in.readline()
        if value == "":
            raise EOFError("EOF when reading from /dev/tty")
        value = value.strip()
        if not value and default:
            return default
        if value:
            return value
        print("Dieser Wert darf nicht leer sein.")

def scope_url(scope: str) -> str:
    return scope_meta[scope]["url"]

def scope_label(scope: str) -> str:
    return scope_meta[scope]["label"]

def prompt_scope(current: str = "", used_scopes=None) -> str:
    used_scopes = used_scopes or set()
    while True:
        print("Verfügbare WordPress-Bereiche:")
        for scope in scope_order:
            if scope in used_scopes and scope != current:
                continue
            print(f"  - {scope}: {scope_label(scope)}")
        prompt = "Bereich für den Passwort-Schutz"
        if current:
            prompt += f" [{current}]"
        prompt += ": "
        print(prompt, end="", flush=True)
        raw = tty_in.readline()
        if raw == "":
            raise EOFError("EOF when reading from /dev/tty")
        selected = (raw.strip() or current).lower()
        if not selected:
            print("Bitte einen Bereich angeben.")
            continue
        if selected not in scope_meta:
            print(f"Unbekannter Bereich: {selected}. Erlaubt sind frontend, admin oder api.")
            continue
        if selected in used_scopes and selected != current:
            print(f"Der Bereich {selected} ist bereits geschützt. Bitte einen anderen wählen.")
            continue
        if selected == "frontend":
            if prompt_yes_no(f"Der Bereich frontend schützt die gesamte Website unter {scope_url('frontend')}. Wirklich fortfahren?", default=False):
                return selected
            print("Bitte einen spezifischeren Bereich wählen oder die Eingabe anpassen.")
            continue
        return selected

def build_hash(username: str):
    while True:
        pw1 = getpass.getpass(f"Passwort für {username}: ")
        pw2 = getpass.getpass(f"Passwort für {username} wiederholen: ")
        if not pw1:
            print("Das Passwort darf nicht leer sein.")
            continue
        if pw1 != pw2:
            print("Die Passwörter stimmen nicht überein. Bitte erneut eingeben.")
            continue
        try:
            result = subprocess.run(
                ["htpasswd", "-nbB", username, pw1],
                check=True,
                capture_output=True,
                text=True,
            )
        except FileNotFoundError:
            print("❌ Tool fehlt: htpasswd. Bitte installieren mit: sudo apt install apache2-utils", file=sys.stderr)
            sys.exit(1)
        return result.stdout.strip().split(":", 1)[1], pw1

updated_entries = []
used_scopes = set()
removed_messages = []
auth_checks = []

for entry in entries:
    if not isinstance(entry, dict):
        continue
    scope = (entry.get("scope", "") or "").strip().lower()
    realm = entry.get("realm", "") or ""
    username = (entry.get("username", "") or "").strip()
    password_hash = (entry.get("password_hash", "") or "").strip()

    print("")
    print(f"Vorhandener Schutz-Eintrag: {scope or '(ohne Bereich)'}")
    if scope not in scope_meta:
        print("Der konfigurierte Bereich ist leer oder ungültig.")
        scope = prompt_scope(used_scopes=used_scopes)
    elif prompt_yes_no(f"Soll der Bereich {scope} geändert werden?", default=False):
        scope = prompt_scope(current=scope, used_scopes=used_scopes)

    if prompt_yes_no(f"Soll der Passwort-Schutz für {scope_url(scope)} aufgehoben werden?", default=False):
        removed_messages.append(f"der Password-Schutz für {scope_url(scope)} wurde aufgehoben!")
        continue

    realm = prompt_nonempty("Realm", realm or scope_meta[scope]["realm"])
    username = prompt_nonempty("Benutzername", username)
    password_plain = ""

    if password_hash:
        if prompt_yes_no(f"Soll das vorhandene Kennwort für {scope_url(scope)} geändert werden?", default=False):
            password_hash, password_plain = build_hash(username)
    else:
        print(f"Für {scope_url(scope)} ist noch kein Kennwort gesetzt.")
        password_hash, password_plain = build_hash(username)

    updated_entries.append({
        "scope": scope,
        "realm": realm,
        "username": username,
        "password_hash": password_hash,
    })
    auth_checks.append({
        "scope": scope,
        "username": username if password_plain else "",
        "password": password_plain,
    })
    used_scopes.add(scope)

while prompt_yes_no("Soll ein WordPress-Bereich mit Passwort-Schutz aktiviert werden?", default=False):
    print("")
    scope = prompt_scope(used_scopes=used_scopes)
    realm = prompt_nonempty("Realm", scope_meta[scope]["realm"])
    username = prompt_nonempty("Benutzername")
    password_hash, password_plain = build_hash(username)
    updated_entries.append({
        "scope": scope,
        "realm": realm,
        "username": username,
        "password_hash": password_hash,
    })
    auth_checks.append({
        "scope": scope,
        "username": username,
        "password": password_plain,
    })
    used_scopes.add(scope)

data["wp_basic_auth_scopes"] = updated_entries
with open(hostvars_file, "w", encoding="utf-8") as fh:
    yaml.safe_dump(data, fh, sort_keys=False, allow_unicode=True)

known_checks = {item["scope"]: item for item in auth_checks}
serialized_checks = []
for item in updated_entries:
    base = {
        "scope": item["scope"],
        "username": "",
        "password": "",
    }
    if item["scope"] in known_checks:
        base.update(known_checks[item["scope"]])
    serialized_checks.append(base)

with open(auth_test_file, "w", encoding="utf-8") as fh:
    json.dump({"domain": domain, "entries": serialized_checks}, fh)

for msg in removed_messages:
    print(f"✅ {msg}")
PY
}

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
  local container="$1" domain="$2" frontend_auth_expected="${3:-0}"
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
      401)
        if [[ "$frontend_auth_expected" == "1" ]]; then
          ok "Öffentlicher HTTPS-Check bestätigt aktiven Frontend-Passwortschutz (Finalstatus 401, Redirects ${public_redirects})"
        else
          warn "Öffentlicher HTTPS-Check meldet Finalstatus ${public_status} (Redirects ${public_redirects})"
        fi
        ;;
      *)
        warn "Öffentlicher HTTPS-Check meldet Finalstatus ${public_status} (Redirects ${public_redirects})"
        ;;
    esac
  else
    warn "Öffentlicher HTTPS-Check konnte nicht ausgeführt werden"
  fi
}

run_wp_auth_checks() {
  local auth_test_file="$1"
  [[ -s "$auth_test_file" ]] || return 0

  python3 - "$auth_test_file" <<'PY'
import json
import subprocess
import sys

auth_test_file = sys.argv[1]
with open(auth_test_file, "r", encoding="utf-8") as fh:
    payload = json.load(fh)

domain = payload.get("domain", "")
entries = payload.get("entries") or []
if not domain or not entries:
    sys.exit(0)

scope_paths = {
    "frontend": "/",
    "admin": "/wp-login.php",
    "api": "/wp-json",
}

def curl_status(url, username="", password=""):
    cmd = ["curl", "-k", "-sS", "-o", "/dev/null", "-w", "%{http_code}"]
    if username or password:
        cmd.extend(["-u", f"{username}:{password}"])
    cmd.append(url)
    result = subprocess.run(cmd, capture_output=True, text=True)
    return result.stdout.strip() if result.returncode == 0 else "000"

failures = []
verified = 0
for entry in entries:
    scope = entry.get("scope", "")
    username = entry.get("username", "")
    password = entry.get("password", "")
    if scope not in scope_paths:
        continue
    url = f"https://{domain}{scope_paths[scope]}"

    public_status = curl_status(url)
    if public_status != "401":
        failures.append(f"Passwort-Schutz greift nicht wie erwartet für {url} (ohne Auth: {public_status}, erwartet 401)")
        continue
    verified += 1

    if username and password:
        auth_status = curl_status(url, username=username, password=password)
        if auth_status not in {"200", "204", "301", "302", "303", "307", "308", "403", "404"}:
            failures.append(f"Authentifizierter Check fehlgeschlagen für {url} ({auth_status})")

if failures:
    for message in failures:
        print(f"❌ {message}", file=sys.stderr)
    sys.exit(1)

print(f"✅ Passwort-Schutz-Selbsttest erfolgreich ({verified} Bereich{'e' if verified != 1 else ''})")
PY
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

AUTH_TEST_FILE="$(mktemp)"
chmod 600 "$AUTH_TEST_FILE"
trap 'rm -f "$AUTH_TEST_FILE"' EXIT
manage_wp_auth "$HOSTVARS_FILE" "$DOMAIN" "$AUTH_TEST_FILE"
FRONTEND_AUTH_EXPECTED="$(python3 - "$AUTH_TEST_FILE" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as fh:
    payload = json.load(fh)
entries = payload.get("entries") or []
print("1" if any(item.get("scope") == "frontend" for item in entries) else "0")
PY
)"

run_playbook_quiet "WordPress-Redeploy abgeschlossen" \
  ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$PLAYBOOK" \
  || die "WordPress-Redeploy fehlgeschlagen"
run_wp_auth_checks "$AUTH_TEST_FILE" || die "Passwort-Schutz-Selbsttest fehlgeschlagen"
run_post_redeploy_checks "$CONTAINER" "$DOMAIN" "$FRONTEND_AUTH_EXPECTED"
