"""CLI shared by the four Bash entrypoints."""
import argparse
import ipaddress
import os
from pathlib import Path
import signal
import socket
import sys
import tempfile
import urllib.request
import yaml
import openproject_archive as archive
import openproject_config as config
from openproject_runtime import ROOT, Site, run, WRITERS


def parser(action):
    p = argparse.ArgumentParser(prog='op-' + action + '.sh', description='OpenProject: ' + action)
    p.add_argument('--domain', help='Domain; ohne Angabe interaktive Abfrage')
    if action in ('add', 'update'):
        p.add_argument('--version', default=config.DEFAULT_VERSION if action == 'add' else None, help='Feste Version X.Y.Z')
    if action == 'add':
        p.add_argument('--wildcard-domain', default='')
        p.add_argument('--dns-account', default='')
        p.add_argument('--skip-dns-check', action='store_true')
        p.add_argument('--resume', action='store_true', help='Fehlgeschlagene Neuinstallation mit vorhandenen Secrets fortsetzen')
    if action == 'backup':
        p.add_argument('--output', type=Path)
    if action == 'restore':
        p.add_argument('--backup', type=Path, required=True)
        p.add_argument('--dry-run', action='store_true')
        p.add_argument('--yes', action='store_true')
        p.add_argument('--recover', action='store_true', help='Nur Wartungszustand: ohne neue Sicherung der defekten Datenbank wiederherstellen')
    return p


def dns_check(name):
    with urllib.request.urlopen('https://api.ipify.org', timeout=15) as response:
        address = str(ipaddress.IPv4Address(response.read(64).decode().strip()))
    resolved = {item[4][0] for item in socket.getaddrinfo(name, None, socket.AF_INET)}
    if address not in resolved:
        raise ValueError('DNS zeigt nicht auf diesen Host; A-Record korrigieren')


def shared_smtp(data):
    for name in ('secrets.yml', 'secrets.yaml'):
        file = ROOT / 'ansible/secrets' / name
        if file.exists():
            values = yaml.safe_load(file.read_text())
            if not isinstance(values, dict):
                raise ValueError('SES-Secrets muessen eine YAML-Map sein; Vault vorher entschluesseln')
            mapping = {'ses_smtp_host': 'smtp_host', 'ses_smtp_port': 'smtp_port',
                       'ses_smtp_user': 'smtp_user', 'ses_smtp_password': 'smtp_password', 'ses_from': 'smtp_from'}
            for source, target in mapping.items():
                if values.get(source) is not None:
                    data['op'][target] = values[source]
            break
    if not data['op']['smtp_host']:
        print('Hinweis: SMTP ist nicht konfiguriert; Mailversand nach Einrichtung pruefen.', flush=True)
    config.validate(data)


def add(site, args):
    if args.resume:
        marker = site.path / 'MAINTENANCE'
        if not marker.is_file() or not marker.read_text().startswith('Recovery-Backup: Neuinstallation;'):
            raise ValueError('--resume ist nur fuer eine fehlgeschlagene Neuinstallation erlaubt')
        # DB/cache may survive a failed install; only writers block resume.
        running = site.compose('--profile', 'tools', 'ps', '--status', 'running', '--services').split() if (site.path / 'compose.yml').exists() else []
        if set(running) & (set(WRITERS) | {'seeder'}):
            raise ValueError('Vor --resume muessen Anwendung und Seeder gestoppt sein')
        data = site.load()
    else:
        if site.hostvars.exists() or site.path.exists():
            raise FileExistsError('Instanz/Hostvars existiert bereits; nichts ueberschrieben')
        data = config.new_config(site.domain, args.version)
        data['op']['wildcard_domain'] = config.domain(args.wildcard_domain) if args.wildcard_domain else ''
        data['op']['dns_account'] = args.dns_account
        shared_smtp(data)
    if not args.skip_dns_check:
        dns_check(site.domain)
    # Existing stack owns Traefik, certificates and provider credentials.
    run(['docker', 'network', 'inspect', 'traefik'])
    site.pull(data)
    if not args.resume:
        config.write_config(site.hostvars, data)
    site.maintenance('Neuinstallation; Details im OpenProject-Runbook')
    try:
        site.render(data)
        site.start()
        (site.path/'MAINTENANCE').unlink()
    except BaseException:
        if (site.path/'compose.yml').exists():
            site.compose('--profile', 'tools', 'stop', '--timeout', '60', *WRITERS, 'seeder')
        raise RuntimeError('Installation fehlgeschlagen. Hostvars bleiben erhalten; Recovery im Runbook.') from None
    print('Container bereit: https://' + site.domain + ' (oeffentlichen HTTPS-Zugriff pruefen)')
    print('Login: admin. Initialpasswort privat in ' + str(site.hostvars) + ' unter op.admin_password.')


def main(action):
    os.umask(0o077)
    args = parser(action).parse_args()
    if not args.domain:
        if not sys.stdin.isatty():
            raise ValueError('--domain erforderlich ohne Terminal')
        args.domain = input('OpenProject-Domain: ')
    site = Site(args.domain)
    if action == 'update' and not args.version:
        if not sys.stdin.isatty():
            raise ValueError('--version erforderlich ohne Terminal')
        args.version = input('Zielversion (X.Y.Z): ')
    if action in ('add', 'update'):
        config.version(args.version)
    if action == 'restore':
        # Snapshot the archive into private staging once; no validate/reopen race.
        with tempfile.TemporaryDirectory(prefix='op-restore-') as folder:
            stage = Path(folder) / 'checked'
            archive.extract(args.backup, stage, site.domain)
            if args.dry_run:
                print('Archiv vollstaendig geprueft; keine Instanz geaendert.')
                return
            if os.geteuid() != 0:
                raise RuntimeError('Bitte mit sudo auf dem Infra-Host ausfuehren')
            site.preflight(ansible=True)
            with site.lock():
                site.restore(stage, args.yes, args.recover)
    else:
        if os.geteuid() != 0:
            raise RuntimeError('Bitte mit sudo auf dem Infra-Host ausfuehren')
        site.preflight(ansible=action in ('add', 'update'))
        with site.lock():
            if action == 'add':
                add(site, args)
            elif action == 'update':
                site.update(args.version)
            else:
                site.available()
                site.backup(args.output or site.backup_path())
    print('OpenProject ' + action + ' erfolgreich: ' + site.domain)


if __name__ == '__main__':
    def interrupted(signum, frame):
        raise KeyboardInterrupt
    signal.signal(signal.SIGTERM, interrupted)
    try:
        main(sys.argv.pop(1))
    except (Exception, KeyboardInterrupt) as error:
        # Avoid parser exception strings that might echo secret YAML lines.
        safe = str(error) if isinstance(error, (ValueError, RuntimeError, FileExistsError)) else type(error).__name__
        print('Fehler: ' + safe, file=sys.stderr)
        sys.exit(1)
