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
  local hostvars_file="$1" domain="$2" site_dir="/srv/static/$2"
  [[ -d "$site_dir" ]] || { warn "Site-Verzeichnis fehlt noch: $site_dir. Überspringe Auth-Verwaltung für diesen Lauf."; return 0; }

  python3 - "$hostvars_file" "$domain" "$site_dir" <<'PY'
import getpass
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

hostvars_file, domain, site_dir = sys.argv[1:4]
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
        return selected

def build_hash(username: str) -> str:
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
        return result.stdout.strip().split(":", 1)[1]

updated_entries = []
used_paths = set()
removed_messages = []

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

    if password_hash:
        if prompt_yes_no(f"Soll das vorhandene Kennwort für https://{domain}{current_path} geändert werden?", default=False):
            password_hash = build_hash(username)
    else:
        print(f"Für https://{domain}{current_path} ist noch kein Kennwort gesetzt.")
        password_hash = build_hash(username)

    auth_file = derive_auth_file(current_path)
    updated_entries.append({
        "path": current_path,
        "realm": realm,
        "username": username,
        "password_hash": password_hash,
        "auth_file": auth_file,
    })
    used_paths.add(current_path)

while prompt_yes_no("Soll ein weiteres Verzeichnis geschützt werden?", default=False):
    print("")
    new_path = prompt_valid_path(used_paths=used_paths)
    realm = prompt_nonempty("Realm", "Protected Area")
    username = prompt_nonempty("Benutzername")
    password_hash = build_hash(username)
    updated_entries.append({
        "path": new_path,
        "realm": realm,
        "username": username,
        "password_hash": password_hash,
        "auth_file": derive_auth_file(new_path),
    })
    used_paths.add(new_path)

data["static_basic_auth_paths"] = updated_entries
with open(hostvars_file, "w", encoding="utf-8") as fh:
    yaml.safe_dump(data, fh, sort_keys=False, allow_unicode=True)

for msg in removed_messages:
    print(f"✅ {msg}")
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
  ansible-playbook -i "$INVENTORY" "$PLAYBOOK"
  ok "Static-Redeploy für alle statischen Sites abgeschlossen"
else
  manage_static_auth "$HOSTVARS_FILE" "$DOMAIN"
  ansible-playbook -i "$INVENTORY" -e "target_domain=${DOMAIN}" "$PLAYBOOK"
  ok "Static-Redeploy abgeschlossen"
fi
