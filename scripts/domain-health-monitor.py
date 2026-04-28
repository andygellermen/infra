#!/usr/bin/env python3
from __future__ import annotations

import argparse
import datetime as dt
import http.client
import json
import math
import socket
import ssl
import subprocess
import sys
import tempfile
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

USER_AGENT = "infra-domain-health-monitor/1.0"
DEFAULT_TIMEOUT = 15.0
DEFAULT_WARN_CERT_DAYS = 14
SITE_OK_STATUSES = tuple(range(200, 400)) + (401, 403)
REDIRECT_PERMANENT_STATUSES = (301, 308)
REDIRECT_TEMPORARY_STATUSES = (302, 307)


@dataclass
class Probe:
    key: str
    kind: str
    host: str
    path: str
    url: str
    expected_statuses: list[int]
    source: str
    expected_location_prefix: str | None = None


@dataclass
class ProbeResult:
    key: str
    status: str
    summary: str
    host: str
    url: str
    kind: str
    checked_at: str
    http_status: int | None = None
    location: str | None = None
    cert_days_left: int | None = None
    source: str | None = None


def parse_args() -> argparse.Namespace:
    root_dir = Path(__file__).resolve().parents[1]
    parser = argparse.ArgumentParser(
        description=(
            "Prueft automatisch alle konfigurierten Infra-Domains auf HTTPS/TLS-, "
            "Statuscode- und Redirect-Probleme und verschickt Aenderungsalarme per SES."
        )
    )
    parser.add_argument(
        "--root-dir",
        default=str(root_dir),
        help="Infra-Root-Verzeichnis. Standard ist das Repo oberhalb von scripts/.",
    )
    parser.add_argument(
        "--state-file",
        default="",
        help="Optionaler Pfad fuer den Statusspeicher. Standard: data/domain-health-monitor/state.json",
    )
    parser.add_argument(
        "--timeout",
        type=float,
        default=DEFAULT_TIMEOUT,
        help=f"Timeout pro DNS/TLS/HTTP-Pruefung in Sekunden. Standard: {DEFAULT_TIMEOUT}",
    )
    parser.add_argument(
        "--warn-cert-days",
        type=int,
        default=DEFAULT_WARN_CERT_DAYS,
        help=(
            "Warnschwelle fuer bald ablaufende Zertifikate in Tagen. "
            f"Standard: {DEFAULT_WARN_CERT_DAYS}"
        ),
    )
    parser.add_argument(
        "--no-mail",
        action="store_true",
        help="Keine Mail versenden, nur pruefen und stdout schreiben.",
    )
    parser.add_argument(
        "--no-recovery-mail",
        action="store_true",
        help="Keine Recovery-Mail versenden, wenn Fehler verschwinden.",
    )
    return parser.parse_args()


def parse_bool(value: Any, default: bool = False) -> bool:
    if value is None:
        return default
    if isinstance(value, bool):
        return value
    text = str(value).strip().lower()
    if text in {"1", "true", "yes", "on"}:
        return True
    if text in {"0", "false", "no", "off"}:
        return False
    return default


def strip_inline_comment(value: str) -> str:
    in_single = False
    in_double = False
    for index, char in enumerate(value):
        if char == "'" and not in_double:
            in_single = not in_single
        elif char == '"' and not in_single:
            in_double = not in_double
        elif char == "#" and not in_single and not in_double:
            return value[:index].rstrip()
    return value.rstrip()


def parse_scalar(value: str) -> str:
    text = strip_inline_comment(value.strip())
    if len(text) >= 2 and text[0] == text[-1] and text[0] in {"'", '"'}:
        return text[1:-1]
    return text


def parse_top_level_scalars(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    if not path.exists():
        return values

    with path.open("r", encoding="utf-8") as handle:
        for raw_line in handle:
            if not raw_line or raw_line[0].isspace():
                continue
            line = raw_line.strip()
            if not line or line.startswith("#") or line == "---":
                continue
            if ":" not in line:
                continue
            key, value = line.split(":", 1)
            values[key.strip()] = parse_scalar(value)
    return values


def parse_hostvars_file(path: Path) -> dict[str, Any]:
    payload: dict[str, Any] = {"traefik": {"aliases": []}}
    in_traefik = False
    in_aliases = False

    with path.open("r", encoding="utf-8") as handle:
        for raw_line in handle:
            line = raw_line.rstrip("\n")
            stripped = line.strip()
            if not stripped or stripped.startswith("#") or stripped == "---":
                continue

            if not line.startswith(" "):
                in_traefik = False
                in_aliases = False

                if stripped == "traefik:":
                    in_traefik = True
                    continue

                if ":" not in stripped:
                    continue
                key, value = stripped.split(":", 1)
                key = key.strip()
                if key in {
                    "domain",
                    "ghost_domain_db",
                    "wp_domain_db",
                    "static_enabled",
                    "sheet_helper_enabled",
                    "static_editor_enabled",
                }:
                    payload[key] = parse_scalar(value)
                continue

            if not in_traefik:
                continue

            if line.startswith("  aliases:") or stripped == "aliases:":
                in_aliases = True
                continue

            if in_aliases and line.startswith("    - "):
                payload["traefik"]["aliases"].append(parse_scalar(line[6:]))
                continue

            if line.startswith("  ") and not line.startswith("    "):
                in_aliases = False

    return payload


def parse_redirects_file(path: Path) -> list[dict[str, Any]]:
    if not path.exists():
        return []

    entries: list[dict[str, Any]] = []
    current: dict[str, Any] | None = None
    in_aliases = False

    with path.open("r", encoding="utf-8") as handle:
        for raw_line in handle:
            line = raw_line.rstrip("\n")
            stripped = line.strip()
            if not stripped or stripped.startswith("#") or stripped == "---":
                continue
            if stripped == "redirects:":
                continue

            if line.startswith("  - "):
                if current:
                    entries.append(current)
                current = {"aliases": []}
                in_aliases = False
                remainder = line[4:].strip()
                if remainder and ":" in remainder:
                    key, value = remainder.split(":", 1)
                    current[key.strip()] = parse_scalar(value)
                continue

            if current is None:
                continue

            if stripped == "aliases:" or line.startswith("    aliases:"):
                in_aliases = True
                continue

            if in_aliases and line.startswith("      - "):
                current["aliases"].append(parse_scalar(line[8:]))
                continue

            if in_aliases and not line.startswith("      - "):
                in_aliases = False

            if line.startswith("    ") and ":" in stripped:
                key, value = stripped.split(":", 1)
                current[key.strip()] = parse_scalar(value)

    if current:
        entries.append(current)
    return entries


def normalize_host(host: str) -> str:
    return host.strip().rstrip(".").encode("idna").decode("ascii")


def iter_unique_hosts(main_domain: str, aliases: list[str]) -> list[str]:
    seen: set[str] = set()
    hosts: list[str] = []
    for candidate in [main_domain, *aliases]:
        if not candidate:
            continue
        host = normalize_host(candidate)
        if host in seen:
            continue
        seen.add(host)
        hosts.append(host)
    return hosts


def hostvars_has_public_service(payload: dict[str, Any]) -> bool:
    if payload.get("ghost_domain_db"):
        return True
    if payload.get("wp_domain_db"):
        return True
    if parse_bool(payload.get("static_enabled")):
        return True
    if parse_bool(payload.get("sheet_helper_enabled")):
        return True
    if parse_bool(payload.get("static_editor_enabled")):
        return True
    return False


def discover_site_probes(root_dir: Path) -> dict[str, Probe]:
    hostvars_dir = root_dir / "ansible" / "hostvars"
    probes: dict[str, Probe] = {}
    if not hostvars_dir.exists():
        return probes

    for path in sorted(hostvars_dir.glob("*.yml")):
        payload = parse_hostvars_file(path)
        if not hostvars_has_public_service(payload):
            continue

        main_domain = str(payload.get("domain") or path.stem).strip()
        traefik_cfg = payload.get("traefik") or {}
        aliases = traefik_cfg.get("aliases") or []
        if not isinstance(aliases, list):
            aliases = []

        for host in iter_unique_hosts(main_domain, [str(item) for item in aliases]):
            url = f"https://{host}/"
            probes[url] = Probe(
                key=url,
                kind="site",
                host=host,
                path="/",
                url=url,
                expected_statuses=list(SITE_OK_STATUSES),
                source=str(path.relative_to(root_dir)),
            )
    return probes


def expected_redirect_prefix(scheme: str, target: str) -> str:
    normalized_target = target.strip()
    if normalized_target.startswith("http://") or normalized_target.startswith("https://"):
        return normalized_target
    return f"{scheme}://{normalized_target}"


def discover_redirect_probes(root_dir: Path, probes: dict[str, Probe]) -> None:
    redirects_file = root_dir / "ansible" / "redirects" / "redirects.yml"
    entries = parse_redirects_file(redirects_file)
    for item in entries:
        source = str(item.get("source") or "").strip()
        target = str(item.get("target") or "").strip()
        aliases = item.get("aliases") or []
        if not source or not target or not isinstance(aliases, list):
            continue

        scheme = str(item.get("target_scheme") or "https").strip() or "https"
        permanent = parse_bool(item.get("permanent"), True)
        statuses = (
            REDIRECT_PERMANENT_STATUSES if permanent else REDIRECT_TEMPORARY_STATUSES
        )
        prefix = expected_redirect_prefix(scheme, target)

        for host in iter_unique_hosts(source, [str(alias) for alias in aliases]):
            url = f"https://{host}/"
            probes[url] = Probe(
                key=url,
                kind="redirect",
                host=host,
                path="/",
                url=url,
                expected_statuses=list(statuses),
                expected_location_prefix=prefix,
                source=str(redirects_file.relative_to(root_dir)),
            )


def discover_probes(root_dir: Path) -> list[Probe]:
    probes = discover_site_probes(root_dir)
    discover_redirect_probes(root_dir, probes)
    return [probes[key] for key in sorted(probes)]


def load_state(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {}
    try:
        with path.open("r", encoding="utf-8") as handle:
            data = json.load(handle)
    except Exception:
        return {}
    if isinstance(data, dict):
        return data
    return {}


def save_state(path: Path, results: list[ProbeResult]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "probes": {result.key: asdict(result) for result in results},
    }
    with path.open("w", encoding="utf-8") as handle:
        json.dump(payload, handle, indent=2, sort_keys=True)
        handle.write("\n")


def parse_not_after(value: str) -> dt.datetime:
    return dt.datetime.strptime(value, "%b %d %H:%M:%S %Y %Z").replace(
        tzinfo=dt.timezone.utc
    )


def check_tls(host: str, timeout: float) -> int | None:
    context = ssl.create_default_context()
    with socket.create_connection((host, 443), timeout=timeout) as raw_sock:
        with context.wrap_socket(raw_sock, server_hostname=host) as tls_sock:
            cert = tls_sock.getpeercert()
            not_after = cert.get("notAfter")
            if not not_after:
                return None
            expires_at = parse_not_after(str(not_after))
    remaining = expires_at - dt.datetime.now(dt.timezone.utc)
    return math.ceil(remaining.total_seconds() / 86400)


def check_http(
    host: str, path: str, timeout: float
) -> tuple[int, str | None]:
    context = ssl.create_default_context()
    conn = http.client.HTTPSConnection(
        host=host,
        port=443,
        timeout=timeout,
        context=context,
    )
    try:
        conn.request(
            "GET",
            path,
            headers={
                "Host": host,
                "User-Agent": USER_AGENT,
                "Accept": "*/*",
                "Connection": "close",
            },
        )
        response = conn.getresponse()
        location = response.getheader("Location")
        response.read(1024)
        return response.status, location
    finally:
        conn.close()


def run_probe(probe: Probe, timeout: float, warn_cert_days: int) -> ProbeResult:
    checked_at = dt.datetime.now(dt.timezone.utc).isoformat()

    try:
        cert_days_left = check_tls(probe.host, timeout)
        status_code, location = check_http(probe.host, probe.path, timeout)
    except ssl.SSLCertVerificationError as exc:
        return ProbeResult(
            key=probe.key,
            status="fail",
            summary=f"TLS verification failed: {exc.verify_message}",
            host=probe.host,
            url=probe.url,
            kind=probe.kind,
            checked_at=checked_at,
            source=probe.source,
        )
    except ssl.SSLError as exc:
        return ProbeResult(
            key=probe.key,
            status="fail",
            summary=f"TLS handshake failed: {exc}",
            host=probe.host,
            url=probe.url,
            kind=probe.kind,
            checked_at=checked_at,
            source=probe.source,
        )
    except socket.gaierror as exc:
        return ProbeResult(
            key=probe.key,
            status="fail",
            summary=f"DNS lookup failed: {exc}",
            host=probe.host,
            url=probe.url,
            kind=probe.kind,
            checked_at=checked_at,
            source=probe.source,
        )
    except TimeoutError:
        return ProbeResult(
            key=probe.key,
            status="fail",
            summary="Connection timed out",
            host=probe.host,
            url=probe.url,
            kind=probe.kind,
            checked_at=checked_at,
            source=probe.source,
        )
    except OSError as exc:
        return ProbeResult(
            key=probe.key,
            status="fail",
            summary=f"Connection failed: {exc}",
            host=probe.host,
            url=probe.url,
            kind=probe.kind,
            checked_at=checked_at,
            source=probe.source,
        )

    if status_code not in probe.expected_statuses:
        return ProbeResult(
            key=probe.key,
            status="fail",
            summary=(
                f"Unexpected HTTP status {status_code}; "
                f"expected one of {sorted(probe.expected_statuses)}"
            ),
            host=probe.host,
            url=probe.url,
            kind=probe.kind,
            checked_at=checked_at,
            http_status=status_code,
            location=location,
            cert_days_left=cert_days_left,
            source=probe.source,
        )

    if probe.expected_location_prefix:
        if not location:
            return ProbeResult(
                key=probe.key,
                status="fail",
                summary=(
                    f"Redirect expected, but Location header is missing; "
                    f"expected prefix {probe.expected_location_prefix}"
                ),
                host=probe.host,
                url=probe.url,
                kind=probe.kind,
                checked_at=checked_at,
                http_status=status_code,
                cert_days_left=cert_days_left,
                source=probe.source,
            )
        expected_prefixes = {
            probe.expected_location_prefix,
            probe.expected_location_prefix.rstrip("/") + "/",
        }
        if not any(location.startswith(prefix) for prefix in expected_prefixes):
            return ProbeResult(
                key=probe.key,
                status="fail",
                summary=(
                    f"Unexpected redirect target {location}; expected prefix "
                    f"{probe.expected_location_prefix}"
                ),
                host=probe.host,
                url=probe.url,
                kind=probe.kind,
                checked_at=checked_at,
                http_status=status_code,
                location=location,
                cert_days_left=cert_days_left,
                source=probe.source,
            )

    if cert_days_left is not None and cert_days_left <= warn_cert_days:
        return ProbeResult(
            key=probe.key,
            status="warn",
            summary=(
                f"Certificate expires soon ({cert_days_left} day(s) left); "
                f"HTTP {status_code}"
            ),
            host=probe.host,
            url=probe.url,
            kind=probe.kind,
            checked_at=checked_at,
            http_status=status_code,
            location=location,
            cert_days_left=cert_days_left,
            source=probe.source,
        )

    summary = f"HTTP {status_code}"
    if probe.kind == "redirect" and location:
        summary = f"HTTP {status_code} -> {location}"
    return ProbeResult(
        key=probe.key,
        status="ok",
        summary=summary,
        host=probe.host,
        url=probe.url,
        kind=probe.kind,
        checked_at=checked_at,
        http_status=status_code,
        location=location,
        cert_days_left=cert_days_left,
        source=probe.source,
    )


def classify_changes(
    previous_state: dict[str, Any], current_results: list[ProbeResult]
) -> tuple[list[ProbeResult], list[ProbeResult], list[ProbeResult]]:
    previous_probes = previous_state.get("probes") or {}
    if not isinstance(previous_probes, dict):
        previous_probes = {}

    new_or_changed_issues: list[ProbeResult] = []
    recoveries: list[ProbeResult] = []
    active_issues: list[ProbeResult] = []

    for result in current_results:
        if result.status != "ok":
            active_issues.append(result)

        previous = previous_probes.get(result.key) or {}
        previous_status = previous.get("status")
        previous_summary = previous.get("summary")

        if result.status == "ok":
            if previous_status in {"warn", "fail"}:
                recoveries.append(result)
            continue

        if previous_status != result.status or previous_summary != result.summary:
            new_or_changed_issues.append(result)

    return new_or_changed_issues, recoveries, active_issues


def load_secrets(root_dir: Path) -> dict[str, Any]:
    secrets_file = root_dir / "ansible" / "secrets" / "secrets.yml"
    return parse_top_level_scalars(secrets_file)


def mail_config_from_secrets(secrets: dict[str, Any]) -> dict[str, Any] | None:
    notify_to = str(secrets.get("infra_error_notify_to") or "").strip()
    if not notify_to:
        return None

    notify_from = str(secrets.get("infra_error_notify_from") or "").strip()
    if not notify_from:
        notify_from = str(secrets.get("ses_from") or "Infra <noreply@localhost>").strip()

    return {
        "mail_to": notify_to,
        "mail_from": notify_from,
        "subject_prefix": str(secrets.get("infra_error_notify_subject_prefix") or "[infra]").strip(),
        "smtp_host": str(secrets.get("ses_smtp_host") or "email-smtp.eu-central-1.amazonaws.com").strip(),
        "smtp_port": int(secrets.get("ses_smtp_port") or 587),
        "smtp_secure": str(secrets.get("ses_smtp_secure") or "").strip(),
        "smtp_require_tls": str(secrets.get("ses_smtp_requireTLS") or "true").strip(),
        "smtp_user": str(secrets.get("ses_smtp_user") or "").strip(),
        "smtp_password": str(secrets.get("ses_smtp_password") or "").strip(),
    }


def build_subject(
    prefix: str,
    new_or_changed: list[ProbeResult],
    recoveries: list[ProbeResult],
    active_issues: list[ProbeResult],
) -> str:
    fail_count = sum(1 for item in active_issues if item.status == "fail")
    warn_count = sum(1 for item in active_issues if item.status == "warn")

    if new_or_changed:
        return (
            f"{prefix} Domain monitor alert: "
            f"{fail_count} failure(s), {warn_count} warning(s) active"
        )
    if active_issues:
        return (
            f"{prefix} Domain monitor recovery/update: "
            f"{fail_count} failure(s), {warn_count} warning(s) still active"
        )
    return f"{prefix} Domain monitor recovery: all tracked changes resolved"


def format_result_line(result: ProbeResult) -> str:
    bits = [f"- [{result.status.upper()}] {result.url}", f"{result.summary}"]
    if result.source:
        bits.append(f"source={result.source}")
    return " | ".join(bits)


def build_report_text(
    root_dir: Path,
    state_file: Path,
    probes: list[Probe],
    new_or_changed: list[ProbeResult],
    recoveries: list[ProbeResult],
    active_issues: list[ProbeResult],
) -> tuple[str, str]:
    now = dt.datetime.now(dt.timezone.utc).strftime("%Y-%m-%d %H:%M:%S %Z")
    lines = [
        f"Zeit: {now}",
        f"Host: {socket.getfqdn() or socket.gethostname()}",
        f"Arbeitsverzeichnis: {root_dir}",
        f"State-Datei: {state_file}",
        f"Gepruefte Probes: {len(probes)}",
        f"Aktive Probleme: {len(active_issues)}",
        "",
    ]

    if new_or_changed:
        lines.append("Neue oder geaenderte Probleme:")
        for item in new_or_changed:
            lines.append(format_result_line(item))
        lines.append("")

    if recoveries:
        lines.append("Behobene Probleme:")
        for item in recoveries:
            lines.append(format_result_line(item))
        lines.append("")

    if active_issues:
        lines.append("Aktueller Problemstand:")
        for item in active_issues:
            lines.append(format_result_line(item))
    else:
        lines.append("Aktueller Problemstand: keine aktiven Probleme")

    summary = (
        f"Domain monitor Lauf abgeschlossen. "
        f"Neue/geaenderte Probleme: {len(new_or_changed)}, "
        f"Recoveries: {len(recoveries)}, "
        f"aktive Probleme: {len(active_issues)}."
    )
    return summary, "\n".join(lines) + "\n"


def send_notification(
    root_dir: Path,
    mail_cfg: dict[str, Any],
    subject: str,
    summary: str,
    report_text: str,
) -> None:
    script = root_dir / "scripts" / "send-error-notification.py"
    if not script.exists():
        raise SystemExit(f"Mail-Script fehlt: {script}")

    with tempfile.NamedTemporaryFile(
        "w", encoding="utf-8", delete=False, prefix="domain-health-", suffix=".log"
    ) as handle:
        handle.write(report_text)
        temp_path = Path(handle.name)

    try:
        subprocess.run(
            [
                sys.executable,
                str(script),
                "--smtp-host",
                mail_cfg["smtp_host"],
                "--smtp-port",
                str(mail_cfg["smtp_port"]),
                "--smtp-user",
                mail_cfg["smtp_user"],
                "--smtp-password",
                mail_cfg["smtp_password"],
                "--smtp-secure",
                mail_cfg["smtp_secure"],
                "--smtp-require-tls",
                mail_cfg["smtp_require_tls"],
                "--mail-from",
                mail_cfg["mail_from"],
                "--mail-to",
                mail_cfg["mail_to"],
                "--subject",
                subject,
                "--summary",
                summary,
                "--log-file",
                str(temp_path),
            ],
            check=True,
        )
    finally:
        temp_path.unlink(missing_ok=True)


def print_console_summary(
    probes: list[Probe],
    results: list[ProbeResult],
    new_or_changed: list[ProbeResult],
    recoveries: list[ProbeResult],
) -> None:
    fail_count = sum(1 for item in results if item.status == "fail")
    warn_count = sum(1 for item in results if item.status == "warn")
    ok_count = sum(1 for item in results if item.status == "ok")

    print(
        f"Domain monitor: {len(probes)} probe(s), "
        f"{ok_count} ok, {warn_count} warning(s), {fail_count} failure(s)"
    )
    if new_or_changed:
        print("Neue/geaenderte Probleme:")
        for item in new_or_changed:
            print(format_result_line(item))
    if recoveries:
        print("Behobene Probleme:")
        for item in recoveries:
            print(format_result_line(item))


def main() -> int:
    args = parse_args()
    root_dir = Path(args.root_dir).resolve()
    state_file = (
        Path(args.state_file).resolve()
        if args.state_file
        else root_dir / "data" / "domain-health-monitor" / "state.json"
    )

    probes = discover_probes(root_dir)
    if not probes:
        print("Domain monitor: keine Probes entdeckt. Hostvars/Redirects pruefen.")
        save_state(state_file, [])
        return 0

    current_results = [
        run_probe(probe, timeout=args.timeout, warn_cert_days=args.warn_cert_days)
        for probe in probes
    ]
    previous_state = load_state(state_file)
    new_or_changed, recoveries, active_issues = classify_changes(
        previous_state, current_results
    )

    print_console_summary(probes, current_results, new_or_changed, recoveries)

    summary, report_text = build_report_text(
        root_dir=root_dir,
        state_file=state_file,
        probes=probes,
        new_or_changed=new_or_changed,
        recoveries=recoveries,
        active_issues=active_issues,
    )

    should_mail = bool(new_or_changed)
    if recoveries and not args.no_recovery_mail:
        should_mail = True

    if should_mail and not args.no_mail:
        secrets = load_secrets(root_dir)
        mail_cfg = mail_config_from_secrets(secrets)
        if mail_cfg is None:
            print(
                "Domain monitor: Mailversand uebersprungen, "
                "infra_error_notify_to fehlt in ansible/secrets/secrets.yml"
            )
        else:
            subject = build_subject(
                prefix=mail_cfg["subject_prefix"],
                new_or_changed=new_or_changed,
                recoveries=recoveries,
                active_issues=active_issues,
            )
            send_notification(root_dir, mail_cfg, subject, summary, report_text)
            print("Domain monitor: Benachrichtigung versendet")

    save_state(state_file, current_results)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
