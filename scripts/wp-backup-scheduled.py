#!/usr/bin/env python3
from __future__ import annotations

import argparse
import re
import subprocess
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

import sys

ROOT_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT_DIR / "scripts" / "lib"))

from report_utils import read_top_level_scalar


VALID_SCHEDULES = {"monthly", "weekly", "daily", "off", "disabled", "manual"}
WEEKDAY_ALIASES = {
    "mon": 0,
    "monday": 0,
    "di": 1,
    "tue": 1,
    "tues": 1,
    "tuesday": 1,
    "mi": 2,
    "wed": 2,
    "wednesday": 2,
    "do": 3,
    "thu": 3,
    "thursday": 3,
    "fr": 4,
    "fri": 4,
    "friday": 4,
    "sa": 5,
    "sat": 5,
    "saturday": 5,
    "so": 6,
    "sun": 6,
    "sunday": 6,
}
DEFAULT_KEEP = {
    "weekly": 8,
    "daily": 14,
}


@dataclass
class ScheduledBackupSite:
    domain: str
    schedule: str
    keep: int
    weekday: int


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Führt hostvar-gesteuerte zusätzliche WordPress-Backups mit weekly/daily-Cadence aus."
    )
    parser.add_argument("--check-only", action="store_true", help="Nur anzeigen, welche Backups heute fällig wären.")
    parser.add_argument("--domain", help="Nur eine Domain verarbeiten.")
    parser.add_argument(
        "--output-dir",
        default=str(ROOT_DIR / "backups" / "scheduled" / "wordpress"),
        help="Zielbasisverzeichnis für die zusätzlichen Backups.",
    )
    return parser.parse_args()


def parse_schedule(path: Path) -> str:
    raw = (read_top_level_scalar(path, "wp_backup_schedule") or "monthly").strip().lower()
    return raw if raw in VALID_SCHEDULES else "monthly"


def parse_keep(path: Path, schedule: str) -> int:
    raw = read_top_level_scalar(path, "wp_backup_keep").strip()
    if raw.isdigit() and int(raw) > 0:
        return int(raw)
    return DEFAULT_KEEP.get(schedule, 3)


def parse_weekday(path: Path) -> int:
    raw = (read_top_level_scalar(path, "wp_backup_weekday") or "sunday").strip().lower()
    if raw in WEEKDAY_ALIASES:
        return WEEKDAY_ALIASES[raw]
    if raw.isdigit():
        value = int(raw)
        if 0 <= value <= 6:
            return value
        if 1 <= value <= 7:
            return (value - 1) % 7
    return WEEKDAY_ALIASES["sunday"]


def discover_due_sites(hostvars_dir: Path, target_domain: str | None, now: datetime) -> tuple[list[ScheduledBackupSite], list[str]]:
    due_sites: list[ScheduledBackupSite] = []
    notices: list[str] = []
    for path in sorted(hostvars_dir.glob("*.yml")):
        if not read_top_level_scalar(path, "wp_domain_db"):
            continue
        domain = path.stem
        if target_domain and domain != target_domain:
            continue

        schedule = parse_schedule(path)
        if schedule in {"off", "disabled", "manual"}:
            notices.append(f"{domain}: zusätzlicher Backup-Plan deaktiviert ({schedule})")
            continue
        if schedule == "monthly":
            continue

        keep = parse_keep(path, schedule)
        weekday = parse_weekday(path)

        if schedule == "daily":
            due_sites.append(ScheduledBackupSite(domain=domain, schedule=schedule, keep=keep, weekday=weekday))
            continue
        if schedule == "weekly" and now.weekday() == weekday:
            due_sites.append(ScheduledBackupSite(domain=domain, schedule=schedule, keep=keep, weekday=weekday))
            continue

    return due_sites, notices


def run_backup(site: ScheduledBackupSite, output_dir: Path, check_only: bool) -> Path:
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    target_file = output_dir / site.domain / f"wp-backup-{site.domain}-{timestamp}.tar.gz"
    cmd = [
        str(ROOT_DIR / "scripts" / "wp-backup.sh"),
        "--create",
        site.domain,
        "--output",
        str(target_file),
    ]
    if check_only:
        print(f"ℹ️  würde {site.schedule}-Backup ausführen: {' '.join(cmd)}")
        return target_file

    target_file.parent.mkdir(parents=True, exist_ok=True)
    print(f"ℹ️  starte {site.schedule}-Backup für {site.domain}")
    subprocess.run(cmd, check=True)
    return target_file


def prune_backups(site: ScheduledBackupSite, output_dir: Path, check_only: bool) -> None:
    domain_dir = output_dir / site.domain
    if not domain_dir.is_dir():
        return
    files = sorted(domain_dir.glob("wp-backup-*.tar.gz"), reverse=True)
    if len(files) <= site.keep:
        return
    for old_file in files[site.keep:]:
        if check_only:
            print(f"ℹ️  würde altes Zusatz-Backup löschen: {old_file}")
        else:
            old_file.unlink(missing_ok=True)
            print(f"ℹ️  altes Zusatz-Backup gelöscht: {old_file}")


def main() -> int:
    args = parse_args()
    hostvars_dir = ROOT_DIR / "ansible" / "hostvars"
    output_dir = Path(args.output_dir)
    now = datetime.now()

    due_sites, notices = discover_due_sites(hostvars_dir, args.domain, now)
    for notice in notices:
        print(f"ℹ️  {notice}")

    if not due_sites:
        print("✅ Keine zusätzlichen WordPress-Backups fällig.")
        return 0

    for site in due_sites:
        run_backup(site, output_dir, args.check_only)
        prune_backups(site, output_dir, args.check_only)

    if args.check_only:
        print(f"✅ Check-only abgeschlossen: {len(due_sites)} WordPress-Instanz(en) wären fällig.")
    else:
        print(f"✅ Zusätzliche WordPress-Backups abgeschlossen: {len(due_sites)} Instanz(en).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
