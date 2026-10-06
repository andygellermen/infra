#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="${EASY_AUTHOR_HOSTVARS_DIR:-$ROOT_DIR/ansible/hostvars}"
die(){ echo "FEHLER: $*" >&2; exit 1; }
require_cmd(){ command -v "$1" >/dev/null 2>&1 || die "Tool fehlt: $1"; }

domain=""; username=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --username=*) username="${1#*=}" ;;
    --help|-h) echo "Usage: $0 <domain> [--username=<name>]"; exit 0 ;;
    *) [[ -z "$domain" ]] || die "Nur eine Domain ist erlaubt."; domain="$1" ;;
  esac
  shift
done
[[ "$domain" =~ ^[a-z0-9]+([.-][a-z0-9]+)*$ ]] || die "Ungueltige oder fehlende Domain."
require_cmd htpasswd; require_cmd python3
hostvars_file="$HOSTVARS_DIR/$domain.yml"
[[ -f "$hostvars_file" ]] || die "Hostvars fehlen: $hostvars_file"
grep -Eq '^easy_author_enabled: true$' "$hostvars_file" || die "Keine aktivierten EasyAuthor-Hostvars."
if [[ -z "$username" ]]; then
  username="$(sed -nE 's/^easy_author_basic_auth_username: "([A-Za-z0-9._-]+)"$/\1/p' "$hostvars_file")"
fi
[[ "$username" =~ ^[A-Za-z0-9._-]+$ ]] || die "Ungueltiger Benutzername."
auth_line="$(htpasswd -nB -C 12 "$username")"; auth_hash="${auth_line#*:}"
[[ "$auth_hash" =~ ^\$2[aby]\$[0-9]{2}\$ ]] || die "htpasswd lieferte keinen gueltigen bcrypt-Hash."
EASY_AUTHOR_HOSTVARS_FILE="$hostvars_file" EASY_AUTHOR_AUTH_USER="$username" EASY_AUTHOR_AUTH_HASH="$auth_hash" python3 - <<'PY'
import os, re, tempfile
from pathlib import Path
p = Path(os.environ["EASY_AUTHOR_HOSTVARS_FILE"])
text = p.read_text(encoding="utf-8")
for key, value in (
    ("easy_author_basic_auth_username", os.environ["EASY_AUTHOR_AUTH_USER"]),
    ("easy_author_basic_auth_password_hash", os.environ["EASY_AUTHOR_AUTH_HASH"]),
):
    text, count = re.subn(rf"(?m)^{key}:.*$", f'{key}: "{value}"', text)
    if count != 1:
        raise SystemExit(f"{key} fehlt oder ist nicht eindeutig")
fd, name = tempfile.mkstemp(prefix=f".{p.name}.", dir=p.parent)
try:
    with os.fdopen(fd, "w", encoding="utf-8") as h:
        h.write(text); h.flush(); os.fsync(h.fileno())
    os.chmod(name, 0o600)
    os.replace(name, p)
finally:
    if os.path.exists(name): os.unlink(name)
PY
unset auth_line auth_hash
echo "OK: Basic-Auth-Zugang wurde atomar rotiert."
