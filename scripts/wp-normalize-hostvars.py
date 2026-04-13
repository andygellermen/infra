#!/usr/bin/env python3
import re
import sys
from pathlib import Path


DOMAIN_LINE_RE = re.compile(r"^domain:\s*.*$")
TRAEFIK_LINE_RE = re.compile(r"^traefik:\s*$")
ALIASES_LINE_RE = re.compile(r"^\s*aliases:\s*$")
ALIAS_ITEM_RE = re.compile(r"^\s*-\s*(.+?)\s*$")


def build_aliases(domain: str, configured_aliases: list[str]) -> list[str]:
    aliases: list[str] = []
    seen: set[str] = set()

    def add(alias: str) -> None:
        alias = (alias or "").strip().strip('"').strip("'")
        if not alias or alias == domain or alias in seen:
            return
        seen.add(alias)
        aliases.append(alias)

    add(f"www.{domain}")
    for alias in configured_aliases:
        add(alias)
        if alias and not alias.startswith("www."):
            add(f"www.{alias}")

    return aliases


def render_traefik_block(domain: str, aliases: list[str]) -> list[str]:
    lines = ["traefik:\n", f"  domain: {domain}\n", "  aliases:\n"]
    lines.extend(f"    - {alias}\n" for alias in aliases)
    return lines


def collect_aliases_from_traefik_block(lines: list[str], start_index: int) -> tuple[int, list[str]]:
    aliases: list[str] = []
    index = start_index + 1
    in_aliases = False

    while index < len(lines):
        line = lines[index]
        if line and not line.startswith((" ", "\t")):
            break

        if ALIASES_LINE_RE.match(line):
            in_aliases = True
            index += 1
            continue

        if in_aliases:
            match = ALIAS_ITEM_RE.match(line)
            if match:
                aliases.append(match.group(1).strip())
                index += 1
                continue
            if line.strip() == "" or line.lstrip().startswith("#"):
                index += 1
                continue
            in_aliases = False

        index += 1

    return index, aliases


def main() -> int:
    if len(sys.argv) < 3:
        print("Usage: wp-normalize-hostvars.py <hostvars-file> <domain> [alias1 alias2 ...]", file=sys.stderr)
        return 1

    hostvars_path = Path(sys.argv[1])
    domain = sys.argv[2].strip()
    cli_aliases = [alias.strip() for alias in sys.argv[3:] if alias.strip()]

    if not hostvars_path.exists():
        print(f"❌ Hostvars-Datei fehlt: {hostvars_path}", file=sys.stderr)
        return 1

    lines = hostvars_path.read_text(encoding="utf-8").splitlines(keepends=True)
    output: list[str] = []
    existing_aliases: list[str] = []
    found_domain = False
    insert_after_domain: int | None = None
    index = 0

    while index < len(lines):
        line = lines[index]

        if DOMAIN_LINE_RE.match(line):
            output.append(f"domain: {domain}\n")
            found_domain = True
            insert_after_domain = len(output)
            index += 1
            continue

        if TRAEFIK_LINE_RE.match(line):
            next_index, aliases = collect_aliases_from_traefik_block(lines, index)
            existing_aliases.extend(aliases)
            index = next_index
            continue

        output.append(line)
        index += 1

    if not found_domain:
        insert_at = 0
        while insert_at < len(output) and (output[insert_at].startswith("#") or output[insert_at].strip() == ""):
            insert_at += 1
        output[insert_at:insert_at] = [f"domain: {domain}\n", "\n"]
        insert_after_domain = insert_at + 1

    aliases = build_aliases(domain, [*existing_aliases, *cli_aliases])
    traefik_block = render_traefik_block(domain, aliases)

    insert_at = insert_after_domain if insert_after_domain is not None else len(output)
    if insert_at < len(output) and output[insert_at].strip() != "":
        traefik_block.append("\n")
    output[insert_at:insert_at] = traefik_block

    hostvars_path.write_text("".join(output), encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
