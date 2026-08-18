#!/usr/bin/env python3
"""Monitor Easy Event Planner containers, health endpoints, and fresh logs."""

from __future__ import annotations

import argparse
import datetime as dt
import fcntl
import hashlib
import json
import os
import re
import socket
import subprocess
import sys
import tempfile
from dataclasses import dataclass
from pathlib import Path
from typing import Any


SCRIPT_DIR = Path(__file__).resolve().parent
ROOT_DIR_DEFAULT = SCRIPT_DIR.parent
sys.path.insert(0, str(SCRIPT_DIR))

from lib.report_utils import load_mail_settings, read_top_level_scalar, send_report_mail


ERROR_RE = re.compile(
    r"(?i)(?:\blevel\s*[=:]\s*(?:error|warn(?:ing)?)\b|"
    r"\b(?:error|fatal|panic|exception|failed|failure|warning)\b|"
    r"database is locked|deadline exceeded|connection refused|"
    r"out of memory|oomkilled|tls handshake|smtp[^\n]*(?:reject|fail))"
)
ERROR_ONLY_RE = re.compile(
    r"(?i)(?:\blevel\s*[=:]\s*error\b|"
    r"\b(?:error|fatal|panic|exception|failed|failure)\b|"
    r"database is locked|deadline exceeded|connection refused|"
    r"out of memory|oomkilled|tls handshake|smtp[^\n]*(?:reject|fail))"
)
EMAIL_RE = re.compile(r"(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b")
SECRET_RE = re.compile(
    r"(?i)(\b(?:authorization|cookie|password|passwd|secret|session|token|code)\b"
    r"\s*[=:]\s*)(?:bearer\s+)?([^\s&;,]+)"
)
QUERY_SECRET_RE = re.compile(
    r"(?i)([?&](?:authorization|code|password|secret|session|token)=)[^&#\s]+"
)
BENIGN_ZERO_COUNTER_RE = re.compile(
    r"(?i)\b(?:errors?|failed|failures?|warnings?)\s*[=:]\s*0\b"
)


@dataclass(frozen=True)
class Site:
    domain: str
    base_url: str
    app_container: str
    worker_container: str
    smoke_insecure: bool


@dataclass(frozen=True)
class ProbeResult:
    key: str
    status: str
    summary: str


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Prueft EEP-Container, Systemendpunkte und neue ERROR/WARN-Logzeilen "
            "und versendet Zustandsaenderungen per SES/SMTP."
        )
    )
    target = parser.add_mutually_exclusive_group(required=True)
    target.add_argument("--domain", help="Nur diese EEP-Domain pruefen")
    target.add_argument("--all", action="store_true", help="Alle eep_enabled-Hostvars pruefen")
    parser.add_argument("--root-dir", default=str(ROOT_DIR_DEFAULT))
    parser.add_argument("--state-dir", help="Default: <root>/data/eep-log-monitor")
    parser.add_argument("--initial-lookback-minutes", type=int, default=65)
    parser.add_argument("--timeout", type=int, default=10)
    parser.add_argument("--max-log-lines", type=int, default=5000)
    parser.add_argument("--max-report-lines", type=int, default=120)
    parser.add_argument("--no-warnings", action="store_true", help="WARN-Muster ignorieren")
    parser.add_argument("--no-mail", action="store_true", help="Nur pruefen und ausgeben")
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Ohne Mail und ohne Aenderung des Monitor-Zustands pruefen",
    )
    parser.add_argument("--no-recovery-mail", action="store_true")
    parser.add_argument("--test-mail", action="store_true", help="Testmail senden und beenden")
    args = parser.parse_args()

    for name in ("initial_lookback_minutes", "timeout", "max_log_lines", "max_report_lines"):
        if getattr(args, name) < 1:
            parser.error(f"--{name.replace('_', '-')} muss groesser als 0 sein")
    if args.test_mail and (args.no_mail or args.dry_run):
        parser.error("--test-mail kann nicht mit --no-mail oder --dry-run kombiniert werden")
    return args


def parse_bool(value: str, default: bool = False) -> bool:
    if not value:
        return default
    return value.strip().lower() in {"1", "true", "yes", "on"}


def discover_sites(root_dir: Path, domain: str | None) -> list[Site]:
    hostvars_dir = root_dir / "ansible" / "hostvars"
    if not hostvars_dir.is_dir():
        raise RuntimeError(f"Hostvars-Verzeichnis fehlt: {hostvars_dir}")

    files = [hostvars_dir / f"{domain}.yml"] if domain else sorted(hostvars_dir.glob("*.yml"))
    sites: list[Site] = []
    for path in files:
        if not path.is_file():
            if domain:
                raise RuntimeError(f"Hostvars fehlen: {path}")
            continue
        if not parse_bool(read_top_level_scalar(path, "eep_enabled")):
            if domain:
                raise RuntimeError(f"eep_enabled ist nicht aktiv in {path}")
            continue
        site_domain = read_top_level_scalar(path, "domain") or path.stem
        site_id = site_domain.replace(".", "-")
        sites.append(
            Site(
                domain=site_domain,
                base_url=(
                    read_top_level_scalar(path, "eep_smoke_base_url")
                    or read_top_level_scalar(path, "eep_base_url")
                    or f"https://{site_domain}"
                ).rstrip("/"),
                app_container=(
                    read_top_level_scalar(path, "eep_container_name")
                    or f"easy-event-planner-{site_id}"
                ),
                worker_container=(
                    read_top_level_scalar(path, "eep_worker_container_name")
                    or f"easy-event-planner-worker-{site_id}"
                ),
                smoke_insecure=parse_bool(read_top_level_scalar(path, "eep_smoke_insecure")),
            )
        )
    if not sites:
        raise RuntimeError("Keine aktive EEP-Instanz gefunden")
    return sites


def run_command(command: list[str], timeout: int = 30) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        command,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        timeout=timeout,
        check=False,
    )


def inspect_container(name: str) -> tuple[ProbeResult, dict[str, Any] | None]:
    result = run_command(["docker", "inspect", name])
    if result.returncode != 0:
        detail = sanitize_line(result.stderr.strip() or "Container nicht gefunden")
        return ProbeResult(f"container:{name}", "fail", detail), None
    try:
        payload = json.loads(result.stdout)[0]
        state = payload.get("State") or {}
        restart_count = int(payload.get("RestartCount") or 0)
    except (ValueError, KeyError, IndexError, TypeError) as exc:
        return ProbeResult(f"container:{name}", "fail", f"docker inspect unlesbar: {exc}"), None

    status = str(state.get("Status") or "unknown")
    running = bool(state.get("Running"))
    oom_killed = bool(state.get("OOMKilled"))
    exit_code = int(state.get("ExitCode") or 0)
    state_error = sanitize_line(str(state.get("Error") or "").strip())
    details = {"restart_count": restart_count, "oom_killed": oom_killed}

    if not running:
        summary = f"nicht aktiv (status={status}, exit={exit_code}, oom_killed={str(oom_killed).lower()})"
        if state_error:
            summary += f", error={state_error}"
        return ProbeResult(f"container:{name}", "fail", summary), details
    if oom_killed:
        return ProbeResult(f"container:{name}", "fail", "aktiv, aber OOMKilled ist gesetzt"), details
    return ProbeResult(f"container:{name}", "ok", "aktiv"), details


def probe_endpoint(
    site: Site, path: str, expected_status: str, expected_body: str, timeout: int
) -> ProbeResult:
    body_handle = tempfile.NamedTemporaryFile(delete=False)
    body_path = Path(body_handle.name)
    body_handle.close()
    command = [
        "curl",
        "--silent",
        "--show-error",
        "--location",
        "--connect-timeout",
        str(timeout),
        "--max-time",
        str(timeout),
        "--output",
        str(body_path),
        "--write-out",
        "%{http_code}",
    ]
    if site.smoke_insecure:
        command.append("--insecure")
    url = f"{site.base_url}{path}"
    command.append(url)
    try:
        result = run_command(command, timeout=timeout + 5)
        status_code = result.stdout.strip() if result.returncode == 0 else "000"
        body = body_path.read_text(encoding="utf-8", errors="replace")
        key = f"endpoint:{site.domain}:{path}"
        body_matches = not expected_body or expected_body.lower() in body.lower()
        if status_code == expected_status and body_matches:
            return ProbeResult(key, "ok", f"{url} -> HTTP {status_code}")
        detail = sanitize_line(result.stderr.strip())
        expectation = f"HTTP {expected_status}"
        if expected_body:
            expectation += f" mit '{expected_body}' im Body"
        summary = f"{url} -> HTTP {status_code}, erwartet {expectation}"
        if detail:
            summary += f" ({detail})"
        return ProbeResult(key, "fail", summary)
    finally:
        body_path.unlink(missing_ok=True)


def sanitize_line(line: str) -> str:
    line = QUERY_SECRET_RE.sub(r"\1<redacted>", line)
    line = SECRET_RE.sub(r"\1<redacted>", line)
    line = EMAIL_RE.sub("<redacted-email>", line)
    return line[:2000]


def collect_log_matches(
    container: str,
    since: str,
    until: str,
    *,
    timeout: int,
    max_log_lines: int,
    max_report_lines: int,
    include_warnings: bool,
) -> tuple[list[str], str | None]:
    result = run_command(
        [
            "docker",
            "logs",
            "--timestamps",
            "--since",
            since,
            "--until",
            until,
            "--tail",
            str(max_log_lines),
            container,
        ],
        timeout=max(timeout, 30),
    )
    if result.returncode != 0:
        return [], sanitize_line(result.stderr.strip() or "docker logs fehlgeschlagen")

    pattern = ERROR_RE if include_warnings else ERROR_ONLY_RE
    matches: list[str] = []
    for raw_line in (result.stdout + result.stderr).splitlines():
        searchable_line = BENIGN_ZERO_COUNTER_RE.sub("", raw_line)
        if pattern.search(searchable_line):
            matches.append(sanitize_line(raw_line))
            if len(matches) >= max_report_lines:
                break
    return matches, None


def load_state(path: Path) -> dict[str, Any]:
    if not path.is_file():
        return {"version": 1, "sites": {}}
    try:
        payload = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {"version": 1, "sites": {}}
    return payload if isinstance(payload, dict) else {"version": 1, "sites": {}}


def save_state(path: Path, state: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True, mode=0o700)
    fd, temp_name = tempfile.mkstemp(prefix="state-", suffix=".json", dir=path.parent)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            json.dump(state, handle, indent=2, sort_keys=True)
            handle.write("\n")
        os.chmod(temp_name, 0o600)
        os.replace(temp_name, path)
    finally:
        if os.path.exists(temp_name):
            os.unlink(temp_name)


def fingerprint(result: ProbeResult) -> str:
    return hashlib.sha256(f"{result.status}\0{result.summary}".encode()).hexdigest()


def classify_probe_changes(
    current: list[ProbeResult], previous: dict[str, Any]
) -> tuple[list[ProbeResult], list[ProbeResult], dict[str, Any]]:
    changed: list[ProbeResult] = []
    recoveries: list[ProbeResult] = []
    current_state: dict[str, Any] = {}
    for result in current:
        old = previous.get(result.key) if isinstance(previous, dict) else None
        old_status = old.get("status") if isinstance(old, dict) else None
        if result.status != "ok" and (old_status == "ok" or not old or old.get("fingerprint") != fingerprint(result)):
            changed.append(result)
        elif result.status == "ok" and old_status in {"warn", "fail"}:
            recoveries.append(result)
        current_state[result.key] = {
            "status": result.status,
            "summary": result.summary,
            "fingerprint": fingerprint(result),
        }
    return changed, recoveries, current_state


def iso_utc(moment: dt.datetime) -> str:
    return moment.astimezone(dt.timezone.utc).isoformat(timespec="microseconds").replace("+00:00", "Z")


def send_test_mail(root_dir: Path) -> None:
    settings = load_mail_settings(
        root_dir,
        to_key="eep_log_monitor_to",
        from_key="eep_log_monitor_from",
        subject_prefix_key="eep_log_monitor_subject_prefix",
        default_subject_prefix="[EEP]",
    )
    if settings is None:
        raise RuntimeError(
            "Mail-Empfaenger fehlt: eep_log_monitor_to oder infra_error_notify_to konfigurieren"
        )
    with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as handle:
        handle.write("Der EEP-Logmonitor kann ueber die konfigurierte SES/SMTP-Verbindung senden.\n")
        body_path = Path(handle.name)
    try:
        send_report_mail(
            root_dir,
            settings,
            subject=f"{settings.subject_prefix} Logmonitor Test",
            body_file=body_path,
        )
    finally:
        body_path.unlink(missing_ok=True)


def main() -> int:
    args = parse_args()
    root_dir = Path(args.root_dir).resolve()
    state_dir = Path(args.state_dir).resolve() if args.state_dir else root_dir / "data" / "eep-log-monitor"
    state_dir.mkdir(parents=True, exist_ok=True, mode=0o700)
    os.chmod(state_dir, 0o700)

    lock_path = state_dir / "monitor.lock"
    lock_handle = lock_path.open("a+", encoding="utf-8")
    os.chmod(lock_path, 0o600)
    try:
        fcntl.flock(lock_handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError:
        print("EEP-Logmonitor: Ein anderer Lauf ist noch aktiv; dieser Lauf wird uebersprungen.")
        return 0

    if args.test_mail:
        send_test_mail(root_dir)
        print("EEP-Logmonitor: Testmail versendet.")
        return 0

    mail_settings = None
    if not args.no_mail and not args.dry_run:
        mail_settings = load_mail_settings(
            root_dir,
            to_key="eep_log_monitor_to",
            from_key="eep_log_monitor_from",
            subject_prefix_key="eep_log_monitor_subject_prefix",
            default_subject_prefix="[EEP]",
        )
        if mail_settings is None:
            raise RuntimeError(
                "Mail-Empfaenger fehlt: eep_log_monitor_to oder infra_error_notify_to konfigurieren"
            )

    sites = discover_sites(root_dir, args.domain)
    state_file = state_dir / "state.json"
    previous_state = {"version": 1, "sites": {}} if args.dry_run else load_state(state_file)
    previous_sites = previous_state.get("sites") or {}
    run_started = dt.datetime.now(dt.timezone.utc)
    initial_since = iso_utc(run_started - dt.timedelta(minutes=args.initial_lookback_minutes))

    report_sections: list[str] = []
    changed_total: list[ProbeResult] = []
    recovery_total: list[ProbeResult] = []
    log_event_count = 0
    next_sites: dict[str, Any] = {}

    for site in sites:
        old_site = previous_sites.get(site.domain) or {}
        probes: list[ProbeResult] = []
        log_sections: list[str] = []
        cursors = dict(old_site.get("log_cursors") or {})
        old_container_meta = old_site.get("containers") or {}
        next_container_meta: dict[str, Any] = {}

        for container in (site.app_container, site.worker_container):
            probe, metadata = inspect_container(container)
            probes.append(probe)
            if metadata is not None:
                next_container_meta[container] = metadata
                old_meta = old_container_meta.get(container) or {}
                old_restarts = int(old_meta.get("restart_count") or 0)
                new_restarts = int(metadata.get("restart_count") or 0)
                if new_restarts > old_restarts and old_container_meta:
                    probes.append(
                        ProbeResult(
                            f"restart:{container}",
                            "warn",
                            f"RestartCount von {old_restarts} auf {new_restarts} gestiegen",
                        )
                    )
                else:
                    probes.append(ProbeResult(f"restart:{container}", "ok", f"RestartCount={new_restarts}"))

            since = cursors.get(container) or initial_since
            matches, log_error = collect_log_matches(
                container,
                since,
                iso_utc(run_started),
                timeout=args.timeout,
                max_log_lines=args.max_log_lines,
                max_report_lines=args.max_report_lines,
                include_warnings=not args.no_warnings,
            )
            if log_error:
                probes.append(ProbeResult(f"logs:{container}", "fail", log_error))
            else:
                cursors[container] = iso_utc(run_started)
                probes.append(ProbeResult(f"logs:{container}", "ok", "Logs lesbar"))
                if matches:
                    log_event_count += len(matches)
                    log_sections.extend([f"[{container}] {line}" for line in matches])

        probes.extend(
            [
                probe_endpoint(site, "/healthz", "200", "ok", args.timeout),
                probe_endpoint(site, "/readyz", "200", "ready", args.timeout),
                probe_endpoint(site, "/version", "200", "", args.timeout),
            ]
        )

        changed, recoveries, probe_state = classify_probe_changes(probes, old_site.get("probes") or {})
        changed_total.extend(changed)
        recovery_total.extend(recoveries)
        next_sites[site.domain] = {
            "probes": probe_state,
            "log_cursors": cursors,
            "containers": next_container_meta,
        }

        if changed or recoveries or log_sections:
            lines = [f"Instanz: {site.domain}"]
            if changed:
                lines.append("Neue oder geaenderte Probleme:")
                lines.extend(
                    f"- [{item.status.upper()}] {item.key}: {item.summary}" for item in changed
                )
            if recoveries:
                lines.append("Entwarnungen:")
                lines.extend(f"- {item.key}: {item.summary}" for item in recoveries)
            if log_sections:
                lines.append(f"Neue auffaellige Logzeilen (seit {old_site.get('checked_at') or initial_since}):")
                lines.extend(f"- {line}" for line in log_sections)
            report_sections.append("\n".join(lines))

        next_sites[site.domain]["checked_at"] = iso_utc(run_started)

    next_state = {"version": 1, "checked_at": iso_utc(run_started), "sites": next_sites}
    should_mail = bool(changed_total or log_event_count or (recovery_total and not args.no_recovery_mail))

    print(
        f"EEP-Logmonitor: {len(sites)} Instanz(en), {len(changed_total)} neue/geaenderte "
        f"Probleme, {log_event_count} auffaellige Logzeile(n), {len(recovery_total)} Entwarnung(en)."
    )
    if report_sections:
        print("\n\n".join(report_sections))

    if should_mail and not args.no_mail and not args.dry_run and mail_settings is not None:
        host = socket.getfqdn() or socket.gethostname()
        active_failures = sum(
            1
            for site_state in next_sites.values()
            for probe in site_state["probes"].values()
            if probe["status"] != "ok"
        )
        if changed_total or log_event_count:
            subject = (
                f"{mail_settings.subject_prefix} Alarm: {len(changed_total)} Zustand, "
                f"{log_event_count} Logtreffer"
            )
        else:
            subject = f"{mail_settings.subject_prefix} Entwarnung: EEP stabil"
        body = "\n".join(
            [
                f"Zeit: {iso_utc(run_started)}",
                f"Host: {host}",
                f"Gepruefte Instanzen: {len(sites)}",
                f"Aktive Zustandsprobleme: {active_failures}",
                f"Neue Logtreffer: {log_event_count}",
                "",
                "\n\n".join(report_sections),
                "",
                "Hinweis: Zugangsdaten, Tokens und E-Mail-Adressen wurden im Logauszug redigiert.",
            ]
        )
        with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as handle:
            handle.write(body)
            body_path = Path(handle.name)
        try:
            send_report_mail(root_dir, mail_settings, subject=subject, body_file=body_path)
            print("EEP-Logmonitor: Benachrichtigung versendet.")
        finally:
            body_path.unlink(missing_ok=True)

    if not args.dry_run:
        save_state(state_file, next_state)
    else:
        print("EEP-Logmonitor: Dry-Run; Zustand wurde nicht gespeichert und keine Mail versendet.")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (
        RuntimeError,
        FileNotFoundError,
        subprocess.CalledProcessError,
        subprocess.TimeoutExpired,
    ) as exc:
        print(f"EEP-Logmonitor FEHLER: {exc}", file=sys.stderr)
        raise SystemExit(2)
