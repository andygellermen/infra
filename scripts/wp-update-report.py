#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import socket
import subprocess
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

import sys

ROOT_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT_DIR / "scripts" / "lib"))

from report_utils import load_mail_settings, read_top_level_scalar, send_report_mail


@dataclass
class WordPressSite:
    domain: str
    hostvars_path: Path
    db_name: str
    db_user: str
    db_password: str
    table_prefix: str
    version: str


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Erstellt einen WordPress-Monatsreport für verfügbare Core-/Plugin-/Theme-Updates."
    )
    parser.add_argument("--check-only", action="store_true", help="Report nur erzeugen und lokal ausgeben, ohne E-Mail-Versand.")
    parser.add_argument("--domain", help="Nur eine einzelne WordPress-Domain auswerten.")
    parser.add_argument("--wp-cli-image", default=os.environ.get("WP_CLI_IMAGE", "wordpress:cli"), help="Docker-Image für WP-CLI.")
    parser.add_argument("--report-dir", default=str(ROOT_DIR / "logs" / "reports"), help="Zielverzeichnis für den Textreport.")
    return parser.parse_args()


def discover_sites(hostvars_dir: Path, target_domain: str | None) -> list[WordPressSite]:
    sites: list[WordPressSite] = []
    for path in sorted(hostvars_dir.glob("*.yml")):
        db_name = read_top_level_scalar(path, "wp_domain_db")
        if not db_name:
            continue
        domain = path.stem
        if target_domain and domain != target_domain:
            continue
        sites.append(
            WordPressSite(
                domain=domain,
                hostvars_path=path,
                db_name=db_name,
                db_user=read_top_level_scalar(path, "wp_domain_usr"),
                db_password=read_top_level_scalar(path, "wp_domain_pwd"),
                table_prefix=read_top_level_scalar(path, "wp_table_prefix") or "wp_",
                version=read_top_level_scalar(path, "wp_version") or "latest",
            )
        )
    return sites


def run_wp_cli(site: WordPressSite, wp_cli_image: str, *wp_args: str) -> tuple[Any, str | None]:
    volume_name = f"wp_{site.domain.replace('.', '_')}_html"
    cmd = [
        "docker",
        "run",
        "--rm",
        "--network",
        "backend",
        "-e",
        "HOME=/tmp",
        "-e",
        "WORDPRESS_DB_HOST=infra-mysql",
        "-e",
        f"WORDPRESS_DB_USER={site.db_user}",
        "-e",
        f"WORDPRESS_DB_PASSWORD={site.db_password}",
        "-e",
        f"WORDPRESS_DB_NAME={site.db_name}",
        "-e",
        f"WORDPRESS_TABLE_PREFIX={site.table_prefix}",
        "-v",
        f"{volume_name}:/var/www/html",
        wp_cli_image,
        "wp",
        "--path=/var/www/html",
        "--skip-plugins",
        "--skip-themes",
        *wp_args,
        "--format=json",
    ]
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.returncode != 0:
        error = result.stderr.strip() or result.stdout.strip() or "WP-CLI command failed"
        return None, error
    stdout = result.stdout.strip()
    if not stdout:
        return [], None
    try:
        return json.loads(stdout), None
    except json.JSONDecodeError as exc:
        return None, f"Ungültige JSON-Antwort von WP-CLI: {exc}"


def build_report(sites: list[WordPressSite], wp_cli_image: str) -> tuple[str, dict[str, int]]:
    generated_at = datetime.now().astimezone().strftime("%Y-%m-%d %H:%M:%S %Z")
    hostname = socket.getfqdn() or socket.gethostname()
    lines = [
        f"WordPress Update Report",
        f"Zeit: {generated_at}",
        f"Host: {hostname}",
        f"WP-CLI Image: {wp_cli_image}",
        "",
    ]

    counts = {
        "sites_total": len(sites),
        "sites_with_updates": 0,
        "core_updates": 0,
        "plugin_updates": 0,
        "theme_updates": 0,
        "errors": 0,
    }

    for site in sites:
        core_updates, core_error = run_wp_cli(site, wp_cli_image, "core", "check-update", "--force-check")
        plugins, plugin_error = run_wp_cli(site, wp_cli_image, "plugin", "list")
        themes, theme_error = run_wp_cli(site, wp_cli_image, "theme", "list")
        site_errors = [error for error in (core_error, plugin_error, theme_error) if error]

        available_plugins = [item for item in (plugins or []) if item.get("update") == "available"]
        available_themes = [item for item in (themes or []) if item.get("update") == "available"]
        core_list = core_updates or []

        has_updates = bool(core_list or available_plugins or available_themes)
        if has_updates:
            counts["sites_with_updates"] += 1
        counts["core_updates"] += len(core_list)
        counts["plugin_updates"] += len(available_plugins)
        counts["theme_updates"] += len(available_themes)
        counts["errors"] += len(site_errors)

        lines.append(f"=== {site.domain} ===")
        lines.append(f"Container-Tag laut Hostvars: {site.version}")
        if site_errors:
            lines.append("Fehler:")
            for error in site_errors:
                lines.append(f"- {error}")
        else:
            if core_list:
                lines.append("Core-Updates:")
                for item in core_list:
                    lines.append(
                        f"- {item.get('version', 'unbekannt')} ({item.get('update_type', 'release')})"
                    )
            else:
                lines.append("Core: aktuell")

            if available_plugins:
                lines.append("Plugin-Updates:")
                for item in available_plugins:
                    lines.append(
                        f"- {item.get('name', 'unbekannt')}: {item.get('version', '?')} -> {item.get('update_version', '?')}"
                    )
            else:
                lines.append("Plugins: aktuell")

            if available_themes:
                lines.append("Theme-Updates:")
                for item in available_themes:
                    lines.append(
                        f"- {item.get('name', 'unbekannt')}: {item.get('version', '?')} -> {item.get('update_version', '?')}"
                    )
            else:
                lines.append("Themes: aktuell")
        lines.append("")

    lines.extend(
        [
            "Zusammenfassung",
            f"- Sites gesamt: {counts['sites_total']}",
            f"- Sites mit Updates: {counts['sites_with_updates']}",
            f"- Core-Updates: {counts['core_updates']}",
            f"- Plugin-Updates: {counts['plugin_updates']}",
            f"- Theme-Updates: {counts['theme_updates']}",
            f"- Auswertungsfehler: {counts['errors']}",
            "",
            "Hinweis: Premium-Themes oder Premium-Plugins tauchen nur auf, wenn ihr Lizenz-/Update-Mechanismus sie auch in WordPress/WP-CLI sichtbar macht.",
        ]
    )
    return "\n".join(lines).rstrip() + "\n", counts


def main() -> int:
    args = parse_args()
    hostvars_dir = ROOT_DIR / "ansible" / "hostvars"
    report_dir = Path(args.report_dir)
    report_dir.mkdir(parents=True, exist_ok=True)

    sites = discover_sites(hostvars_dir, args.domain)
    if not sites:
        raise SystemExit("Keine passenden WordPress-Hostvars gefunden.")

    body, counts = build_report(sites, args.wp_cli_image)
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
    report_file = report_dir / f"wp-update-report-{timestamp}.txt"
    report_file.write_text(body, encoding="utf-8")
    print(body, end="")
    print(f"Report gespeichert unter: {report_file}")

    if args.check_only:
        return 0

    mail_settings = load_mail_settings(
        ROOT_DIR,
        to_key="wp_update_report_to",
        from_key="wp_update_report_from",
        subject_prefix_key="wp_update_report_subject_prefix",
        default_subject_prefix="[wordpress]",
    )
    if mail_settings is None:
        print("⚠️  Keine Empfänger für den WordPress-Update-Report konfiguriert; E-Mail wird übersprungen.")
        return 0

    subject = (
        f"{mail_settings.subject_prefix} WordPress Update Report: "
        f"{counts['sites_with_updates']} Site(s) mit Updates, "
        f"{counts['errors']} Fehler"
    )
    send_report_mail(ROOT_DIR, mail_settings, subject=subject, body_file=report_file)
    print("✅ WordPress-Update-Report per E-Mail versendet.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
