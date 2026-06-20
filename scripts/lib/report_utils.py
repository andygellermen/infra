#!/usr/bin/env python3
from __future__ import annotations

import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path


TOP_LEVEL_SCALAR_RE = re.compile(r"^([A-Za-z0-9_]+):\s*(.*?)\s*$")


@dataclass
class MailSettings:
    smtp_host: str
    smtp_port: str
    smtp_user: str
    smtp_password: str
    smtp_secure: str
    smtp_require_tls: str
    mail_from: str
    mail_to: str
    subject_prefix: str


def strip_wrapping_quotes(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {"'", '"'}:
        return value[1:-1]
    return value


def read_top_level_scalar(path: Path, key: str) -> str:
    for raw_line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        if raw_line.startswith((" ", "\t", "-", "#")):
            continue
        match = TOP_LEVEL_SCALAR_RE.match(raw_line)
        if not match:
            continue
        current_key, raw_value = match.groups()
        if current_key != key:
            continue
        return strip_wrapping_quotes(raw_value)
    return ""


def resolve_secrets_file(root_dir: Path) -> Path | None:
    for candidate in (
        root_dir / "ansible" / "secrets" / "secrets.yml",
        root_dir / "ansible" / "secrets" / "secrets.yaml",
    ):
        if candidate.is_file():
            return candidate
    return None


def load_mail_settings(
    root_dir: Path,
    *,
    to_key: str,
    from_key: str,
    subject_prefix_key: str,
    default_subject_prefix: str,
) -> MailSettings | None:
    secrets_file = resolve_secrets_file(root_dir)
    if secrets_file is None:
        return None

    mail_to = read_top_level_scalar(secrets_file, to_key) or read_top_level_scalar(secrets_file, "infra_error_notify_to")
    if not mail_to:
        return None

    mail_from = (
        read_top_level_scalar(secrets_file, from_key)
        or read_top_level_scalar(secrets_file, "infra_error_notify_from")
        or read_top_level_scalar(secrets_file, "ses_from")
        or "Infra Bot <noreply@localhost>"
    )

    return MailSettings(
        smtp_host=read_top_level_scalar(secrets_file, "ses_smtp_host") or "email-smtp.eu-central-1.amazonaws.com",
        smtp_port=read_top_level_scalar(secrets_file, "ses_smtp_port") or "587",
        smtp_user=read_top_level_scalar(secrets_file, "ses_smtp_user"),
        smtp_password=read_top_level_scalar(secrets_file, "ses_smtp_password"),
        smtp_secure=read_top_level_scalar(secrets_file, "ses_smtp_secure"),
        smtp_require_tls=read_top_level_scalar(secrets_file, "ses_smtp_requireTLS") or "true",
        mail_from=mail_from,
        mail_to=mail_to,
        subject_prefix=read_top_level_scalar(secrets_file, subject_prefix_key) or default_subject_prefix,
    )


def send_report_mail(root_dir: Path, mail_settings: MailSettings, *, subject: str, body_file: Path) -> None:
    cmd = [
        sys.executable,
        str(root_dir / "scripts" / "send-error-notification.py"),
        "--smtp-host",
        mail_settings.smtp_host,
        "--smtp-port",
        str(mail_settings.smtp_port),
        "--smtp-user",
        mail_settings.smtp_user,
        "--smtp-password",
        mail_settings.smtp_password,
        "--smtp-secure",
        mail_settings.smtp_secure,
        "--smtp-require-tls",
        mail_settings.smtp_require_tls,
        "--mail-from",
        mail_settings.mail_from,
        "--mail-to",
        mail_settings.mail_to,
        "--subject",
        subject,
        "--body-file",
        str(body_file),
    ]
    subprocess.run(cmd, check=True)
