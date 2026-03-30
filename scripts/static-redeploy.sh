#!/usr/bin/env bash
set -euo pipefail

INVENTORY="./ansible/inventory/hosts.ini"
PLAYBOOK="./ansible/playbooks/deploy-static.yml"
HOSTVARS_DIR="./ansible/hostvars"

usage(){ echo "Usage: $0 <domain>|--all [--check-only]"; }
die(){ echo "❌ $*" >&2; exit 1; }
ok(){ echo "✅ $*"; }
info(){ echo "ℹ️  $*"; }
warn(){ echo "⚠️  $*"; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

manage_static_auth() {
  local hostvars_file="$1" domain="$2" auth_test_file="$3" site_dir="/srv/static/$2"
  [[ -d "$site_dir" ]] || { warn "Site-Verzeichnis fehlt noch: $site_dir. Überspringe Auth-Verwaltung für diesen Lauf."; return 0; }

  python3 - "$hostvars_file" "$domain" "$site_dir" "$auth_test_file" <<'PY'
import getpass
import json
import os
from pathlib import Path
import re
import subprocess
import sys

try:
    import yaml
except Exception as exc:
    print(f"❌ PyYAML fehlt für die Auth-Verwaltung: {exc}", file=sys.stderr)
    sys.exit(1)

hostvars_file, domain, site_dir, auth_test_file = sys.argv[1:5]
site_path = Path(site_dir)
try:
    tty_in = open("/dev/tty", "r", encoding="utf-8", errors="replace")
except OSError:
    print("❌ Kein interaktives Terminal verfügbar (/dev/tty).", file=sys.stderr)
    sys.exit(1)

with open(hostvars_file, "r", encoding="utf-8") as fh:
    data = yaml.safe_load(fh) or {}

entries = data.get("static_basic_auth_paths") or []
if not isinstance(entries, list):
    entries = []

def normalize_path(value: str) -> str:
    value = (value or "").strip()
    if not value:
        return ""
    value = "/" + value.lstrip("/")
    value = re.sub(r"/+", "/", value)
    if value != "/" and not value.endswith("/"):
        value += "/"
    return value

def sanitize_path_for_file(value: str) -> str:
    value = normalize_path(value).strip("/")
    value = re.sub(r"[^A-Za-z0-9._-]+", "-", value)
    return value or "private"

def derive_auth_file(path_value: str) -> str:
    return f"/srv/static-auth/{domain}-{sanitize_path_for_file(path_value)}.htpasswd"

def list_available_dirs():
    dirs = ["/"]
    for current_root, subdirs, _files in os.walk(site_path):
      rel = os.path.relpath(current_root, site_path)
      if rel == ".":
          continue
      dirs.append("/" + rel.strip("./").replace(os.sep, "/") + "/")
    return sorted(set(dirs), key=lambda item: (item.count("/"), item))

def show_dirs():
    dirs = list_available_dirs()
    print("Verfügbare Verzeichnisse:")
    for d in dirs:
        print(f"  - {d}")
    return dirs

def path_exists(web_path: str) -> bool:
    norm = normalize_path(web_path)
    target = site_path / norm.lstrip("/")
    return target.is_dir()

def find_probe_path(web_path: str) -> str:
    norm = normalize_path(web_path)
    target = site_path / norm.lstrip("/")
    if not target.is_dir():
        return norm
    preferred = ["index.html", "index.htm", "default.html", "default.htm"]
    for name in preferred:
        candidate = target / name
        if candidate.is_file():
            rel = os.path.relpath(candidate, site_path)
            return "/" + rel.replace(os.sep, "/")
    for current_root, _subdirs, files in os.walk(target):
        files = sorted(files)
        for name in files:
            if name.startswith("."):
                continue
            rel = os.path.relpath(Path(current_root) / name, site_path)
            return "/" + rel.replace(os.sep, "/")
    return norm

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
        prompt = f"{question}"
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

def prompt_valid_path(current: str = "", used_paths=None) -> str:
    used_paths = used_paths or set()
    while True:
        dirs = show_dirs()
        prompt = "Pfad für den Passwort-Schutz"
        if current:
            prompt += f" [{current}]"
        prompt += ": "
        print(prompt, end="", flush=True)
        raw = tty_in.readline()
        if raw == "":
            raise EOFError("EOF when reading from /dev/tty")
        raw = raw.strip()
        selected = normalize_path(raw or current)
        if not selected:
            print("Bitte einen Pfad angeben.")
            continue
        if selected in used_paths:
            print(f"Der Pfad {selected} ist bereits geschützt. Bitte einen anderen wählen.")
            continue
        if selected not in dirs or not path_exists(selected):
            print(f"Der Pfad {selected} existiert nicht.")
            continue
        if selected == "/":
            if prompt_yes_no(f"Der Root-Pfad / schützt die gesamte Website unter https://{domain}/. Wirklich fortfahren?", default=False):
                return selected
            print("Bitte einen spezifischeren Pfad wählen oder die Eingabe anpassen.")
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
used_paths = set()
removed_messages = []
auth_checks = []

for entry in entries:
    if not isinstance(entry, dict):
        continue
    current_path = normalize_path(entry.get("path", ""))
    realm = entry.get("realm", "Protected Area") or "Protected Area"
    username = (entry.get("username", "") or "").strip()
    password_hash = (entry.get("password_hash", "") or "").strip()
    auth_file = (entry.get("auth_file", "") or "").strip()

    print("")
    print(f"Vorhandener Schutz-Eintrag: {current_path or '(ohne Pfad)'}")
    if not current_path or not path_exists(current_path):
        print("Der konfigurierte Pfad ist leer oder existiert nicht mehr.")
        current_path = prompt_valid_path(used_paths=used_paths)
    elif prompt_yes_no(f"Soll der Pfad {current_path} geändert werden?", default=False):
        current_path = prompt_valid_path(current=current_path, used_paths=used_paths)

    if prompt_yes_no(f"Soll der Passwort-Schutz für https://{domain}{current_path} aufgehoben werden?", default=False):
        if auth_file and auth_file.startswith("/srv/static-auth/"):
            try:
                Path(auth_file).unlink(missing_ok=True)
            except Exception:
                pass
        removed_messages.append(f"der Password-Schutz für https://{domain}{current_path} wurde aufgehoben!")
        continue

    realm = prompt_nonempty("Realm", realm)
    username = prompt_nonempty("Benutzername", username)
    password_plain = ""

    if password_hash:
        if prompt_yes_no(f"Soll das vorhandene Kennwort für https://{domain}{current_path} geändert werden?", default=False):
            password_hash, password_plain = build_hash(username)
    else:
        print(f"Für https://{domain}{current_path} ist noch kein Kennwort gesetzt.")
        password_hash, password_plain = build_hash(username)

    auth_file = derive_auth_file(current_path)
    updated_entries.append({
        "path": current_path,
        "realm": realm,
        "username": username,
        "password_hash": password_hash,
        "auth_file": auth_file,
    })
    auth_checks.append({
        "path": current_path,
        "probe_path": find_probe_path(current_path),
        "username": username if password_plain else "",
        "password": password_plain,
    })
    used_paths.add(current_path)

while prompt_yes_no("Soll ein Verzeichnis mit Passwort-Schutz aktiviert werden?", default=False):
    print("")
    new_path = prompt_valid_path(used_paths=used_paths)
    realm = prompt_nonempty("Realm", "Protected Area")
    username = prompt_nonempty("Benutzername")
    password_hash, password_plain = build_hash(username)
    updated_entries.append({
        "path": new_path,
        "realm": realm,
        "username": username,
        "password_hash": password_hash,
        "auth_file": derive_auth_file(new_path),
    })
    auth_checks.append({
        "path": new_path,
        "probe_path": find_probe_path(new_path),
        "username": username,
        "password": password_plain,
    })
    used_paths.add(new_path)

data["static_basic_auth_paths"] = updated_entries
with open(hostvars_file, "w", encoding="utf-8") as fh:
    yaml.safe_dump(data, fh, sort_keys=False, allow_unicode=True)

if updated_entries:
    known_checks = {item["path"]: item for item in auth_checks}
    serialized_checks = []
    for item in updated_entries:
        base = {
            "path": item["path"],
            "probe_path": find_probe_path(item["path"]),
            "username": "",
            "password": "",
        }
        if item["path"] in known_checks:
            base.update(known_checks[item["path"]])
        serialized_checks.append(base)
else:
    serialized_checks = []

with open(auth_test_file, "w", encoding="utf-8") as fh:
    json.dump({"domain": domain, "entries": serialized_checks}, fh)

for msg in removed_messages:
    print(f"✅ {msg}")
PY
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

run_static_auth_checks() {
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
    path = entry.get("path", "")
    probe_path = entry.get("probe_path") or path
    username = entry.get("username", "")
    password = entry.get("password", "")
    url = f"https://{domain}{probe_path}"

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

print(f"✅ Passwort-Schutz-Selbsttest erfolgreich ({verified} Pfad{'e' if verified != 1 else ''})")
PY
}

CHECK_ONLY=0
[[ $# -ge 1 ]] || { usage; exit 1; }
TARGET="$1"; shift || true
[[ "${1:-}" == "--check-only" ]] && CHECK_ONLY=1

require_cmd curl
require_cmd dig
require_cmd python3
require_cmd ansible-playbook

if [[ "$TARGET" != "--all" ]]; then
  DOMAIN="$TARGET"
  HOSTVARS_FILE="$HOSTVARS_DIR/${DOMAIN}.yml"
  [[ -f "$HOSTVARS_FILE" ]] || die "Hostvars nicht gefunden: $HOSTVARS_FILE"
  grep -q '^static_enabled:[[:space:]]*true' "$HOSTVARS_FILE" || die "Hostvars gehören nicht zu einer statischen Site: $HOSTVARS_FILE"

  host_ip="$(curl -fsSL https://api.ipify.org || true)"
  dns_ip="$(dig +short A "$DOMAIN" | head -n1)"
  [[ -n "$host_ip" && -n "$dns_ip" ]] || die "DNS/IP Prüfung fehlgeschlagen"
  [[ "$host_ip" == "$dns_ip" ]] || die "DNS mismatch: $DOMAIN -> $dns_ip (erwartet $host_ip)"
  ok "DNS OK für $DOMAIN"
fi

if [[ "$CHECK_ONLY" -eq 1 ]]; then
  ok "Check-only abgeschlossen"
  exit 0
fi

if [[ "$TARGET" == "--all" ]]; then
  run_playbook_quiet "Static-Redeploy für alle statischen Sites abgeschlossen" \
    ansible-playbook -i "$INVENTORY" "$PLAYBOOK" \
    || die "Static-Redeploy fehlgeschlagen"
else
  AUTH_TEST_FILE="$(mktemp)"
  chmod 600 "$AUTH_TEST_FILE"
  trap 'rm -f "$AUTH_TEST_FILE"' EXIT
  manage_static_auth "$HOSTVARS_FILE" "$DOMAIN" "$AUTH_TEST_FILE"
  run_playbook_quiet "Static-Redeploy abgeschlossen" \
    ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$PLAYBOOK" \
    || die "Static-Redeploy fehlgeschlagen"
  run_static_auth_checks "$AUTH_TEST_FILE" || die "Passwort-Schutz-Selbsttest fehlgeschlagen"
fi
