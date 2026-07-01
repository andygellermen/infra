#!/usr/bin/env python3
import argparse
import smtplib
import ssl


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
    args = parser.parse_args()

    secure = parse_bool(args.smtp_secure, False)
    require_tls = parse_bool(args.smtp_require_tls, True)

    if secure:
        with smtplib.SMTP_SSL(args.smtp_host, args.smtp_port, context=ssl.create_default_context(), timeout=30) as smtp:
            if args.smtp_user:
                smtp.login(args.smtp_user, args.smtp_password)
            print(f"SMTP login ok via SMTPS {args.smtp_host}:{args.smtp_port}")
        return 0

    with smtplib.SMTP(args.smtp_host, args.smtp_port, timeout=30) as smtp:
        smtp.ehlo()
        if require_tls:
            smtp.starttls(context=ssl.create_default_context())
            smtp.ehlo()
        if args.smtp_user:
            smtp.login(args.smtp_user, args.smtp_password)
        print(f"SMTP login ok via SMTP {args.smtp_host}:{args.smtp_port} (starttls={'on' if require_tls else 'off'})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
