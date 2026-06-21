#!/usr/bin/env python3
from __future__ import annotations

import argparse
import gzip
import json
import socket
import subprocess
from collections import Counter
from datetime import datetime, timedelta, timezone
from pathlib import Path

import sys

ROOT_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT_DIR / "scripts" / "lib"))

from report_utils import load_mail_settings, send_report_mail


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Erstellt einen Security-Digest aus CrowdSec-Alerts und Komponentenstatus."
    )
    parser.add_argument("--check-only", action="store_true", help="Report nur lokal erzeugen, ohne E-Mail-Versand.")
    parser.add_argument("--days", type=int, default=1, help="Rückblick in Tagen für den Digest.")
    parser.add_argument(
        "--report-dir",
        default=str(ROOT_DIR / "logs" / "reports"),
        help="Zielverzeichnis für den Textreport.",
    )
    parser.add_argument(
        "--alerts-dir",
        default=str(ROOT_DIR / "data" / "crowdsec" / "data" / "notifications"),
        help="Verzeichnis mit CrowdSec-Alert-Dateien.",
    )
    return parser.parse_args()


def parse_timestamp(raw: str) -> datetime | None:
    if not raw:
        return None
    normalized = raw.replace("Z", "+00:00")
    try:
        parsed = datetime.fromisoformat(normalized)
    except ValueError:
        return None
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def iter_alert_lines(alerts_dir: Path):
    files = sorted(alerts_dir.glob("crowdsec_alerts.ndjson*"))
    for path in files:
        if path.suffix == ".gz":
            with gzip.open(path, "rt", encoding="utf-8", errors="replace") as handle:
                for line in handle:
                    yield path, line
            continue
        with path.open("r", encoding="utf-8", errors="replace") as handle:
            for line in handle:
                yield path, line


def collect_alerts(alerts_dir: Path, days: int) -> tuple[list[dict], list[Path]]:
    since = datetime.now(timezone.utc) - timedelta(days=days)
    alerts: list[dict] = []
    files_seen: set[Path] = set()
    for path, raw_line in iter_alert_lines(alerts_dir):
        files_seen.add(path)
        stripped = raw_line.strip()
        if not stripped:
            continue
        try:
            payload = json.loads(stripped)
        except json.JSONDecodeError:
            continue
        crowdsec = payload.get("crowdsec") or {}
        stop_at = parse_timestamp(str(crowdsec.get("stop_at", "")))
        if stop_at is None or stop_at < since:
            continue
        alerts.append(
            {
                "stop_at": stop_at,
                "scenario": str(crowdsec.get("scenario", "unknown")),
                "value": str(crowdsec.get("value", "unknown")),
                "type": str(crowdsec.get("type", "unknown")),
                "duration": str(crowdsec.get("duration", "")),
            }
        )
    alerts.sort(key=lambda item: item["stop_at"], reverse=True)
    return alerts, sorted(files_seen)


def command_output(*cmd: str) -> tuple[bool, str]:
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    except Exception as exc:
        return False, str(exc)
    if result.returncode != 0:
        return False, (result.stderr or result.stdout).strip()
    return True, (result.stdout or "").strip()


def gather_component_status() -> dict[str, str]:
    statuses: dict[str, str] = {}
    for container_name in ("crowdsec", "crowdsec-bouncer-traefik"):
        ok, output = command_output("docker", "inspect", "--format", "{{.State.Status}}", container_name)
        statuses[container_name] = output if ok and output else "unavailable"
    ok, output = command_output("systemctl", "is-active", "wazuh-agent")
    statuses["wazuh-agent"] = output if ok and output else "inactive"
    return statuses


def build_report(alerts: list[dict], files_seen: list[Path], statuses: dict[str, str], days: int) -> str:
    now_local = datetime.now().astimezone().strftime("%Y-%m-%d %H:%M:%S %Z")
    hostname = socket.getfqdn() or socket.gethostname()
    scenario_counter = Counter(alert["scenario"] for alert in alerts)
    ip_counter = Counter(alert["value"] for alert in alerts)

    lines = [
        "Security Early-Warning Digest",
        f"Zeit: {now_local}",
        f"Host: {hostname}",
        f"Rückblick: letzte {days} Tage",
        "",
        "Komponentenstatus",
        f"- crowdsec: {statuses.get('crowdsec', 'unknown')}",
        f"- crowdsec-bouncer-traefik: {statuses.get('crowdsec-bouncer-traefik', 'unknown')}",
        f"- wazuh-agent: {statuses.get('wazuh-agent', 'unknown')}",
        "",
        f"Ausgewertete CrowdSec-Dateien: {len(files_seen)}",
        f"Gefundene CrowdSec-Entscheidungen im Zeitraum: {len(alerts)}",
        "",
    ]

    if scenario_counter:
        lines.append("Top-Szenarien")
        for scenario, count in scenario_counter.most_common(10):
            lines.append(f"- {scenario}: {count}")
        lines.append("")
    else:
        lines.append("Top-Szenarien")
        lines.append("- keine CrowdSec-Entscheidungen im Zeitraum")
        lines.append("")

    if ip_counter:
        lines.append("Top-Quellwerte")
        for value, count in ip_counter.most_common(10):
            lines.append(f"- {value}: {count}")
        lines.append("")

    lines.append("Letzte Ereignisse")
    if alerts:
        for alert in alerts[:10]:
            stop_at_local = alert["stop_at"].astimezone().strftime("%Y-%m-%d %H:%M:%S %Z")
            lines.append(
                f"- {stop_at_local} | {alert['scenario']} | {alert['value']} | {alert['type']} | {alert['duration']}"
            )
    else:
        lines.append("- keine CrowdSec-Entscheidungen im Zeitraum")

    lines.extend(
        [
            "",
            "Hinweis: Dieser Digest fasst CrowdSec-Frühwarnsignale und den lokalen Agent-/Container-Status zusammen.",
            "Hinweis: Wazuh-Korrelation, Mail-Policies und Dashboards bleiben weiterhin eine Aufgabe des Wazuh-Managers.",
        ]
    )
    return "\n".join(lines).rstrip() + "\n"


def main() -> int:
    args = parse_args()
    report_dir = Path(args.report_dir)
    report_dir.mkdir(parents=True, exist_ok=True)

    alerts_dir = Path(args.alerts_dir)
    alerts, files_seen = collect_alerts(alerts_dir, args.days)
    statuses = gather_component_status()
    body = build_report(alerts, files_seen, statuses, args.days)

    timestamp = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
    report_file = report_dir / f"security-digest-{timestamp}.txt"
    report_file.write_text(body, encoding="utf-8")
    print(body, end="")
    print(f"Report gespeichert unter: {report_file}")

    if args.check_only:
        return 0

    mail_settings = load_mail_settings(
        ROOT_DIR,
        to_key="security_digest_report_to",
        from_key="security_digest_report_from",
        subject_prefix_key="security_digest_report_subject_prefix",
        default_subject_prefix="[security]",
    )
    if mail_settings is None:
        print("⚠️  Keine Empfänger für den Security-Digest konfiguriert; E-Mail wird übersprungen.")
        return 0

    critical_components = [
        name
        for name in ("crowdsec", "crowdsec-bouncer-traefik", "wazuh-agent")
        if statuses.get(name) not in {"running", "active"}
    ]
    subject = (
        f"{mail_settings.subject_prefix} Security Digest: "
        f"{len(alerts)} CrowdSec-Entscheidungen"
        + (f", Komponenten prüfen: {', '.join(critical_components)}" if critical_components else "")
    )
    send_report_mail(ROOT_DIR, mail_settings, subject=subject, body_file=report_file)
    print("✅ Security-Digest per E-Mail versendet.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
