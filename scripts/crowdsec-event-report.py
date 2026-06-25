#!/usr/bin/env python3
from __future__ import annotations

import argparse
import gzip
import json
import re
import socket
import subprocess
import sys
from collections import Counter
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from pathlib import Path

ROOT_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT_DIR / "scripts" / "lib"))

from report_utils import load_mail_settings, send_report_mail


TRAEFIK_ACCESSLOG_RE = re.compile(
    r'^(?P<ip>\S+) \S+ \S+ \[(?P<time>[^\]]+)\] "(?P<method>[A-Z]+) (?P<path>\S+)(?: HTTP/[^"]+)?" '
    r'(?P<status>\d{3}) (?P<size>\S+) "(?P<referer>[^"]*)" "(?P<user_agent>[^"]*)"(?P<rest>.*)$'
)
SENSITIVE_PATH_KEYWORDS = (
    "/wp-login.php",
    "/wp-admin",
    "/xmlrpc.php",
    "/wp-json",
    "/admin",
    ".php",
)
LOW_BLOCK_STATUSES = {401, 403, 404}
HIGH_SUCCESS_STATUSES = {200, 201, 202, 204, 301, 302, 307, 308}


@dataclass
class CrowdSecEvent:
    event_id: str
    machine_id: str
    stop_at: datetime
    scenario: str
    value: str
    event_type: str
    duration: str


@dataclass
class TraefikHit:
    ip: str
    timestamp: datetime | None
    method: str
    path: str
    status: int
    referer: str
    user_agent: str
    router: str
    backend: str
    raw: str


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Erstellt einen angereicherten Sofort-Report fuer neue CrowdSec-Ereignisse."
    )
    parser.add_argument("--check-only", action="store_true", help="Nur Report schreiben, keine E-Mail senden.")
    parser.add_argument(
        "--alerts-dir",
        default=str(ROOT_DIR / "data" / "crowdsec" / "data" / "notifications"),
        help="Verzeichnis mit CrowdSec-Alert-Dateien.",
    )
    parser.add_argument(
        "--report-dir",
        default=str(ROOT_DIR / "logs" / "reports"),
        help="Zielverzeichnis fuer Reports.",
    )
    parser.add_argument(
        "--state-file",
        default=str(ROOT_DIR / "data" / "crowdsec" / "state" / "event-report-state.json"),
        help="Datei fuer verarbeitete Event-IDs.",
    )
    parser.add_argument(
        "--event-lookback-minutes",
        type=int,
        default=120,
        help="Nur neue CrowdSec-Events innerhalb dieses Fensters beruecksichtigen.",
    )
    parser.add_argument(
        "--state-retention-days",
        type=int,
        default=14,
        help="So lange bleiben verarbeitete Event-IDs im State erhalten.",
    )
    parser.add_argument(
        "--traefik-container",
        default="infra-traefik",
        help="Traefik-Containername fuer Logkorrelation.",
    )
    parser.add_argument(
        "--traefik-log-lookback-minutes",
        type=int,
        default=180,
        help="Rueckblickfenster fuer Traefik-Logs.",
    )
    return parser.parse_args()


def parse_iso_timestamp(raw: str) -> datetime | None:
    normalized = raw.strip().replace("Z", "+00:00")
    if not normalized:
        return None
    try:
        parsed = datetime.fromisoformat(normalized)
    except ValueError:
        return None
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def parse_traefik_timestamp(raw: str) -> datetime | None:
    try:
        return datetime.strptime(raw, "%d/%b/%Y:%H:%M:%S %z").astimezone(timezone.utc)
    except ValueError:
        return None


def iter_alert_lines(alerts_dir: Path):
    for path in sorted(alerts_dir.glob("crowdsec_alerts.ndjson*")):
        if path.suffix == ".gz":
            with gzip.open(path, "rt", encoding="utf-8", errors="replace") as handle:
                for line in handle:
                    yield line
            continue
        with path.open("r", encoding="utf-8", errors="replace") as handle:
            for line in handle:
                yield line


def build_event_id(crowdsec: dict) -> str:
    return "|".join(
        [
            str(crowdsec.get("machine_id", "")),
            str(crowdsec.get("stop_at", "")),
            str(crowdsec.get("scenario", "")),
            str(crowdsec.get("value", "")),
            str(crowdsec.get("type", "")),
            str(crowdsec.get("duration", "")),
        ]
    )


def load_state(state_file: Path) -> dict[str, str]:
    if not state_file.is_file():
        return {}
    try:
        payload = json.loads(state_file.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}
    processed = payload.get("processed_ids")
    if not isinstance(processed, dict):
        return {}
    return {str(key): str(value) for key, value in processed.items()}


def save_state(state_file: Path, processed_ids: dict[str, str]) -> None:
    state_file.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "updated_at": datetime.now(timezone.utc).isoformat(),
        "processed_ids": processed_ids,
    }
    state_file.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def prune_state(processed_ids: dict[str, str], retention_days: int) -> dict[str, str]:
    cutoff = datetime.now(timezone.utc) - timedelta(days=retention_days)
    pruned: dict[str, str] = {}
    for event_id, raw_timestamp in processed_ids.items():
        parsed = parse_iso_timestamp(raw_timestamp)
        if parsed is None or parsed >= cutoff:
            pruned[event_id] = raw_timestamp
    return pruned


def collect_new_events(alerts_dir: Path, processed_ids: dict[str, str], lookback_minutes: int) -> list[CrowdSecEvent]:
    since = datetime.now(timezone.utc) - timedelta(minutes=lookback_minutes)
    events: list[CrowdSecEvent] = []
    seen_in_run: set[str] = set()

    for raw_line in iter_alert_lines(alerts_dir):
        stripped = raw_line.strip()
        if not stripped:
            continue
        try:
            payload = json.loads(stripped)
        except json.JSONDecodeError:
            continue
        crowdsec = payload.get("crowdsec") or {}
        stop_at = parse_iso_timestamp(str(crowdsec.get("stop_at", "")))
        if stop_at is None or stop_at < since:
            continue
        event_id = build_event_id(crowdsec)
        if event_id in processed_ids or event_id in seen_in_run:
            continue
        event = CrowdSecEvent(
            event_id=event_id,
            machine_id=str(crowdsec.get("machine_id", "unknown")),
            stop_at=stop_at,
            scenario=str(crowdsec.get("scenario", "unknown")),
            value=str(crowdsec.get("value", "unknown")),
            event_type=str(crowdsec.get("type", "unknown")),
            duration=str(crowdsec.get("duration", "")),
        )
        events.append(event)
        seen_in_run.add(event_id)

    events.sort(key=lambda item: item.stop_at)
    return events


def command_output(*cmd: str) -> tuple[bool, str]:
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
    except Exception as exc:
        return False, str(exc)
    if result.returncode != 0:
        return False, (result.stderr or result.stdout).strip()
    return True, (result.stdout or "").strip()


def parse_traefik_line(raw_line: str) -> TraefikHit | None:
    match = TRAEFIK_ACCESSLOG_RE.match(raw_line.strip())
    if not match:
        return None
    rest = match.group("rest")
    quoted_values = re.findall(r'"([^"]*)"', rest)
    router = quoted_values[0] if len(quoted_values) >= 1 else "unknown"
    backend = quoted_values[1] if len(quoted_values) >= 2 else "unknown"
    try:
        status = int(match.group("status"))
    except ValueError:
        return None
    return TraefikHit(
        ip=match.group("ip"),
        timestamp=parse_traefik_timestamp(match.group("time")),
        method=match.group("method"),
        path=match.group("path"),
        status=status,
        referer=match.group("referer"),
        user_agent=match.group("user_agent"),
        router=router,
        backend=backend,
        raw=raw_line.rstrip(),
    )


def correlate_traefik_hits(events: list[CrowdSecEvent], container_name: str, lookback_minutes: int) -> list[TraefikHit]:
    if not events:
        return []
    distinct_values = sorted({event.value for event in events if event.value and event.value != "unknown"})
    if not distinct_values:
        return []
    earliest = min(event.stop_at for event in events) - timedelta(minutes=10)
    fallback_since = datetime.now(timezone.utc) - timedelta(minutes=lookback_minutes)
    since = max(earliest, fallback_since)
    ok, output = command_output(
        "docker",
        "logs",
        "--since",
        since.isoformat(),
        container_name,
    )
    if not ok or not output:
        return []

    hits: list[TraefikHit] = []
    distinct_value_set = set(distinct_values)
    for raw_line in output.splitlines():
        parsed = parse_traefik_line(raw_line)
        if parsed is None:
            continue
        if parsed.ip not in distinct_value_set:
            continue
        hits.append(parsed)
    return hits


def gather_component_status() -> dict[str, str]:
    statuses: dict[str, str] = {}
    for container_name in ("crowdsec", "crowdsec-bouncer-traefik"):
        ok, output = command_output("docker", "inspect", "--format", "{{.State.Status}}", container_name)
        statuses[container_name] = output if ok and output else "unavailable"
    ok, output = command_output("systemctl", "is-active", "wazuh-agent")
    statuses["wazuh-agent"] = output if ok and output else "inactive"
    return statuses


def severity_score(events: list[CrowdSecEvent], hits: list[TraefikHit]) -> int:
    score = 0
    scenarios = " ".join(event.scenario.lower() for event in events)
    if any(keyword in scenarios for keyword in ("cve", "rce", "shell", "sqli", "xss", "lfi", "traversal")):
        score += 5
    elif any(keyword in scenarios for keyword in ("wordpress", "wp", "login", "xmlrpc", "http-bf", "bruteforce")):
        score += 3
    else:
        score += 1

    event_count = len(events)
    if event_count >= 20:
        score += 3
    elif event_count >= 10:
        score += 2
    elif event_count >= 3:
        score += 1

    if len({event.value for event in events}) >= 3:
        score += 1

    if hits:
        statuses = {hit.status for hit in hits}
        if any(status in HIGH_SUCCESS_STATUSES for status in statuses):
            score += 3
        elif statuses and statuses.issubset(LOW_BLOCK_STATUSES):
            score -= 1

        if any(
            any(keyword in hit.path.lower() for keyword in SENSITIVE_PATH_KEYWORDS) and hit.status in HIGH_SUCCESS_STATUSES
            for hit in hits
        ):
            score += 2

    return max(score, 0)


def severity_label(score: int) -> str:
    if score >= 8:
        return "CRITICAL"
    if score >= 5:
        return "HIGH"
    if score >= 3:
        return "MEDIUM"
    return "LOW"


def summarize_breakthrough_risk(hits: list[TraefikHit]) -> str:
    if not hits:
        return "Keine korrelierbaren Traefik-Logs zu den betroffenen Quellwerten gefunden."
    statuses = {hit.status for hit in hits}
    if statuses and statuses.issubset(LOW_BLOCK_STATUSES):
        return "Bisher kein offensichtlicher Durchbruch: korrelierte Traefik-Requests enden nur in 401/403/404."
    if any(hit.status in HIGH_SUCCESS_STATUSES for hit in hits):
        return "Achtung: Es gibt erfolgreiche oder weiterleitende Antworten (2xx/3xx) in den korrelierten Traefik-Logs."
    if any(hit.status >= 500 for hit in hits):
        return "Achtung: Es gibt 5xx-Antworten in den korrelierten Traefik-Logs; das ist kein Beweis fuer Erfolg, aber ein Risikoindikator."
    return "Gemischtes Bild in den korrelierten Traefik-Logs; bitte Details pruefen."


def recommend_actions(severity: str, events: list[CrowdSecEvent], hits: list[TraefikHit]) -> list[str]:
    scenarios = " ".join(event.scenario.lower() for event in events)
    actions: list[str] = []

    if severity in {"CRITICAL", "HIGH"}:
        actions.append("Sofort die betroffene Anwendung pruefen oder voruebergehend abschirmen, falls 2xx/3xx auf sensitive Pfade sichtbar sind.")
        actions.append("Admin-Zugaenge, API-Secrets und ggf. WordPress-Passwoerter rotieren, wenn ein erfolgreicher Login nicht ausgeschlossen werden kann.")
    else:
        actions.append("Kurz Traefik- und App-Logs gegenpruefen; aktuell wirkt der Angriff eher geblockt als erfolgreich.")

    if "xmlrpc" in scenarios:
        actions.append("Pruefen, ob `xmlrpc.php` weiterhin sauber mit 403 blockiert wird.")
    if any("wp" in event.scenario.lower() or "wordpress" in event.scenario.lower() for event in events):
        actions.append("WordPress-Admin-Schutz, MFA-Status und Update-Stand der betroffenen Instanz gegenpruefen.")
    if any(hit.status in HIGH_SUCCESS_STATUSES for hit in hits):
        actions.append("Datei- und DB-Spuren auf der Zielinstanz priorisiert pruefen, weil erfolgreiche Antworten sichtbar sind.")
    if any(hit.status >= 500 for hit in hits):
        actions.append("5xx-Spuren im Zielcontainer pruefen, um Fehlerzustand oder Exploit-Folgen frueh zu erkennen.")

    actions.append("Falls noetig: `docker exec crowdsec cscli decisions list -o human` fuer die aktive Sperre pruefen.")
    return actions


def build_report(
    events: list[CrowdSecEvent],
    hits: list[TraefikHit],
    statuses: dict[str, str],
    severity: str,
    score: int,
) -> str:
    now_local = datetime.now().astimezone().strftime("%Y-%m-%d %H:%M:%S %Z")
    hostname = socket.getfqdn() or socket.gethostname()
    scenario_counter = Counter(event.scenario for event in events)
    value_counter = Counter(event.value for event in events)
    status_counter = Counter(hit.status for hit in hits)
    path_counter = Counter(hit.path for hit in hits)
    router_counter = Counter(hit.router for hit in hits if hit.router and hit.router != "unknown")
    ua_counter = Counter(hit.user_agent for hit in hits if hit.user_agent and hit.user_agent != "-")
    breakthrough = summarize_breakthrough_risk(hits)
    actions = recommend_actions(severity, events, hits)

    lines = [
        "CrowdSec Immediate Incident Report",
        f"Zeit: {now_local}",
        f"Host: {hostname}",
        f"Schweregrad: {severity} (Score {score})",
        "",
        "Komponentenstatus",
        f"- crowdsec: {statuses.get('crowdsec', 'unknown')}",
        f"- crowdsec-bouncer-traefik: {statuses.get('crowdsec-bouncer-traefik', 'unknown')}",
        f"- wazuh-agent: {statuses.get('wazuh-agent', 'unknown')}",
        "",
        f"Neue CrowdSec-Events: {len(events)}",
        f"Betroffene Quellwerte: {len(value_counter)}",
        f"Top-Szenario: {scenario_counter.most_common(1)[0][0] if scenario_counter else 'unknown'}",
        "",
        "Einschaetzung",
        f"- {breakthrough}",
        "",
        "Empfohlene Sofortmassnahmen",
    ]
    lines.extend(f"- {action}" for action in actions)
    lines.append("")

    lines.append("CrowdSec-Szenarien")
    for scenario, count in scenario_counter.most_common(10):
        lines.append(f"- {scenario}: {count}")
    lines.append("")

    lines.append("Betroffene Quellwerte")
    for value, count in value_counter.most_common(10):
        lines.append(f"- {value}: {count}")
    lines.append("")

    lines.append("Korrelierte Traefik-Statuscodes")
    if status_counter:
        for status, count in status_counter.most_common():
            lines.append(f"- {status}: {count}")
    else:
        lines.append("- keine korrelierbaren Traefik-Requests gefunden")
    lines.append("")

    lines.append("Top-Pfade")
    if path_counter:
        for path, count in path_counter.most_common(10):
            lines.append(f"- {path}: {count}")
    else:
        lines.append("- keine korrelierbaren Pfade")
    lines.append("")

    lines.append("Top-Router")
    if router_counter:
        for router, count in router_counter.most_common(10):
            lines.append(f"- {router}: {count}")
    else:
        lines.append("- kein Router aus Logs ableitbar")
    lines.append("")

    lines.append("Top-User-Agents")
    if ua_counter:
        for user_agent, count in ua_counter.most_common(10):
            lines.append(f"- {user_agent}: {count}")
    else:
        lines.append("- keine User-Agents korreliert")
    lines.append("")

    lines.append("Letzte CrowdSec-Events")
    for event in events[-10:]:
        stop_at_local = event.stop_at.astimezone().strftime("%Y-%m-%d %H:%M:%S %Z")
        lines.append(
            f"- {stop_at_local} | {event.scenario} | {event.value} | {event.event_type} | {event.duration}"
        )
    lines.append("")

    lines.append("Letzte korrelierte Traefik-Hits")
    if hits:
        for hit in hits[-10:]:
            timestamp = hit.timestamp.astimezone().strftime("%Y-%m-%d %H:%M:%S %Z") if hit.timestamp else "unknown"
            lines.append(
                f"- {timestamp} | {hit.ip} | {hit.status} | {hit.method} {hit.path} | router={hit.router} | ua={hit.user_agent}"
            )
    else:
        lines.append("- keine korrelierbaren Traefik-Hits")

    return "\n".join(lines).rstrip() + "\n"


def build_subject(severity: str, events: list[CrowdSecEvent], hits: list[TraefikHit], subject_prefix: str) -> str:
    scenario_counter = Counter(event.scenario for event in events)
    top_scenario = scenario_counter.most_common(1)[0][0] if scenario_counter else "unknown"
    value_count = len({event.value for event in events})
    status_counter = Counter(hit.status for hit in hits)
    blocked_only = status_counter and set(status_counter).issubset(LOW_BLOCK_STATUSES)
    status_hint = "blocked" if blocked_only else "needs-review"
    return (
        f"{subject_prefix} {severity} CrowdSec Alert: "
        f"{top_scenario} | {value_count} source(s) | {status_hint}"
    )


def main() -> int:
    args = parse_args()
    report_dir = Path(args.report_dir)
    report_dir.mkdir(parents=True, exist_ok=True)

    state_file = Path(args.state_file)
    processed_ids = prune_state(load_state(state_file), args.state_retention_days)
    events = collect_new_events(Path(args.alerts_dir), processed_ids, args.event_lookback_minutes)

    if not events:
        save_state(state_file, processed_ids)
        print("Keine neuen CrowdSec-Events im betrachteten Zeitfenster.")
        return 0

    hits = correlate_traefik_hits(events, args.traefik_container, args.traefik_log_lookback_minutes)
    statuses = gather_component_status()
    score = severity_score(events, hits)
    severity = severity_label(score)
    report_body = build_report(events, hits, statuses, severity, score)

    timestamp = datetime.now(timezone.utc).strftime("%Y%m%d-%H%M%S")
    report_file = report_dir / f"crowdsec-event-report-{timestamp}.txt"
    report_file.write_text(report_body, encoding="utf-8")
    print(report_body, end="")
    print(f"Report gespeichert unter: {report_file}")

    for event in events:
        processed_ids[event.event_id] = event.stop_at.isoformat()
    save_state(state_file, processed_ids)

    if args.check_only:
        return 0

    mail_settings = load_mail_settings(
        ROOT_DIR,
        to_key="crowdsec_event_report_to",
        from_key="crowdsec_event_report_from",
        subject_prefix_key="crowdsec_event_report_subject_prefix",
        default_subject_prefix="[crowdsec-auto]",
    )
    if mail_settings is None:
        print("⚠️  Keine Empfaenger fuer CrowdSec-Sofortreports konfiguriert; E-Mail wird uebersprungen.")
        return 0

    subject = build_subject(severity, events, hits, mail_settings.subject_prefix)
    send_report_mail(ROOT_DIR, mail_settings, subject=subject, body_file=report_file)
    print("✅ CrowdSec-Sofortreport per E-Mail versendet.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
