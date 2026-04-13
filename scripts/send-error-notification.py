#!/usr/bin/env python3
import argparse
import smtplib
import ssl
from email.message import EmailMessage
from email.utils import getaddresses, parseaddr
from pathlib import Path


def parse_bool(value: str | None, default: bool) -> bool:
    if value is None or value == "":
        return default
    return value.strip().lower() in {"1", "true", "yes", "on"}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--smtp-host", required=True)
    parser.add_argument("--smtp-port", type=int, required=True)
    parser.add_argument("--smtp-user", default="")
    parser.add_argument("--smtp-password", default="")
    parser.add_argument("--smtp-secure", default="")
    parser.add_argument("--smtp-require-tls", default="")
    parser.add_argument("--mail-from", required=True)
    parser.add_argument("--mail-to", required=True)
    parser.add_argument("--subject", required=True)
    parser.add_argument("--summary", required=True)
    parser.add_argument("--log-file", required=True)
    args = parser.parse_args()

    secure = parse_bool(args.smtp_secure, False)
    require_tls = parse_bool(args.smtp_require_tls, True)

    recipients = [email for _name, email in getaddresses([args.mail_to]) if email]
    if not recipients:
      raise SystemExit("no notification recipients configured")

    from_name, from_email = parseaddr(args.mail_from)
    if not from_email:
        raise SystemExit("invalid sender address")

    log_excerpt = Path(args.log_file).read_text(encoding="utf-8", errors="replace")

    msg = EmailMessage()
    msg["Subject"] = args.subject
    msg["From"] = args.mail_from
    msg["To"] = ", ".join(recipients)
    msg.set_content(
        f"{args.summary}\n\nLetzte Logzeilen:\n{'=' * 72}\n{log_excerpt}"
    )

    if secure:
        with smtplib.SMTP_SSL(args.smtp_host, args.smtp_port, context=ssl.create_default_context(), timeout=30) as smtp:
            if args.smtp_user:
                smtp.login(args.smtp_user, args.smtp_password)
            smtp.send_message(msg, from_addr=from_email, to_addrs=recipients)
        return 0

    with smtplib.SMTP(args.smtp_host, args.smtp_port, timeout=30) as smtp:
        smtp.ehlo()
        if require_tls:
            smtp.starttls(context=ssl.create_default_context())
            smtp.ehlo()
        if args.smtp_user:
            smtp.login(args.smtp_user, args.smtp_password)
        smtp.send_message(msg, from_addr=from_email, to_addrs=recipients)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
