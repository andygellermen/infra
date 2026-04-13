#!/usr/bin/env python3
import re
import sys
from pathlib import Path


DOMAIN_LINE_RE = re.compile(r"^domain:\s*.*$")
TRAEFIK_LINE_RE = re.compile(r"^traefik:\s*$")
TRAEFIK_DOMAIN_LINE_RE = re.compile(r"^  domain:\s*.*$")
TRAEFIK_ALIASES_LINE_RE = re.compile(r"^  aliases:\s*$")
TRAEFIK_ALIAS_ITEM_RE = re.compile(r"^    -\s*(.+?)\s*$")


def build_aliases(domain: str, configured_aliases: list[str]) -> list[str]:
    aliases: list[str] = []
    seen: set[str] = set()

    def add(alias: str) -> None:
        alias = (alias or "").strip()
        if not alias or alias == domain or alias in seen:
            return
        seen.add(alias)
        aliases.append(alias)

    add(f"www.{domain}")
    for alias in configured_aliases:
        alias = (alias or "").strip()
        if not alias:
            continue
        add(alias)
        if not alias.startswith("www."):
            add(f"www.{alias}")

    return aliases


def render_traefik_block(domain: str, aliases: list[str], extra_lines: list[str]) -> list[str]:
    block = ["traefik:\n", f"  domain: {domain}\n", "  aliases:\n"]
    block.extend(f"    - {alias}\n" for alias in aliases)

    if extra_lines and block[-1] != "\n" and extra_lines[0] != "\n":
        block.append("\n")
    block.extend(extra_lines)
    return block


def split_traefik_block(lines: list[str], start_index: int) -> tuple[int, list[str], list[str]]:
    index = start_index + 1
    alias_values: list[str] = []
    extra_lines: list[str] = []

    while index < len(lines):
        line = lines[index]
        if line and not line.startswith((" ", "\t")):
            break

        if TRAEFIK_DOMAIN_LINE_RE.match(line):
            index += 1
            continue

        if TRAEFIK_ALIASES_LINE_RE.match(line):
            index += 1
            while index < len(lines):
                alias_line = lines[index]
                match = TRAEFIK_ALIAS_ITEM_RE.match(alias_line)
                if match:
                    alias_values.append(match.group(1).strip().strip('"').strip("'"))
                    index += 1
                    continue
                if alias_line.startswith("    ") and alias_line.strip() == "":
                    index += 1
                    continue
                break
            continue

        extra_lines.append(line)
        index += 1

    return index, alias_values, extra_lines


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
    found_domain = False
    found_traefik = False
    existing_aliases: list[str] = []
    traefik_insert_index: int | None = None
    index = 0

    while index < len(lines):
        line = lines[index]

        if DOMAIN_LINE_RE.match(line):
            output.append(f"domain: {domain}\n")
            found_domain = True
            if traefik_insert_index is None:
                traefik_insert_index = len(output)
            index += 1
            continue

        if TRAEFIK_LINE_RE.match(line):
            found_traefik = True
            next_index, parsed_aliases, extra_lines = split_traefik_block(lines, index)
            existing_aliases.extend(parsed_aliases)
            output.extend(render_traefik_block(domain, build_aliases(domain, [*existing_aliases, *cli_aliases]), extra_lines))
            index = next_index
            continue

        output.append(line)
        index += 1

    if not found_domain:
        insert_at = 0
        while insert_at < len(output) and (output[insert_at].startswith("#") or output[insert_at].strip() == ""):
            insert_at += 1
        output[insert_at:insert_at] = [f"domain: {domain}\n", "\n"]
        if traefik_insert_index is None:
            traefik_insert_index = insert_at + 2

    if not found_traefik:
        traefik_block = render_traefik_block(domain, build_aliases(domain, cli_aliases), [])
        insert_at = traefik_insert_index if traefik_insert_index is not None else len(output)
        if insert_at > 0 and output[insert_at - 1].strip() != "":
            traefik_block.insert(0, "\n")
        if insert_at < len(output) and output[insert_at].strip() != "":
            traefik_block.append("\n")
        output[insert_at:insert_at] = traefik_block

    hostvars_path.write_text("".join(output), encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
