"""Per-domain lifecycle. Commands are argv arrays; secrets never enter argv/logs."""
from contextlib import contextmanager
from datetime import datetime, timezone
import fcntl
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import openproject_config as config
import openproject_archive as archive

ROOT = Path(__file__).resolve().parents[2]
WRITERS = ('web', 'worker', 'cron', 'hocuspocus')


def run(args, *, stdout=None, stdin=None):
    result = subprocess.run([str(a) for a in args], stdin=stdin,
                            stdout=stdout if stdout is not None else subprocess.PIPE,
                            stderr=subprocess.PIPE, check=False)
    if result.returncode:
        # Container/Ansible output may contain secrets. Keep it out of terminal logs.
        raise RuntimeError('Befehl fehlgeschlagen: ' + str(args[0]) + ' (Exit ' + str(result.returncode) + ')')
    return result.stdout.decode() if stdout is None else ''


def environment(data):
    c = config.validate(data)['op']
    result = {
        'OPENPROJECT_HOST__NAME': data['domain'], 'OPENPROJECT_HTTPS': 'true',
        'OPENPROJECT_HSTS': 'true', 'OPENPROJECT_EDITION': 'standard', 'SECRET_KEY_BASE': c['secret_key'],
        'DATABASE_URL': 'postgres://postgres:' + c['db_password'] + '@db/openproject?pool=20&encoding=unicode&reconnect=true',
        'OPENPROJECT_RAILS__CACHE__STORE': 'memcache', 'OPENPROJECT_CACHE__MEMCACHE__SERVER': 'cache:11211',
        'OPENPROJECT_ADDITIONAL__HOST__NAMES': config.site_id(data['domain'])[:35] + '.' + config.site_id(data['domain'])[35:] + '.internal',
        'OPENPROJECT_COLLABORATIVE__EDITING__HOCUSPOCUS__URL': 'wss://' + data['domain'] + '/hocuspocus',
        'OPENPROJECT_COLLABORATIVE__EDITING__HOCUSPOCUS__SECRET': c['collaborative_secret'],
        'OPENPROJECT_SEED__ADMIN__USER__PASSWORD': c['admin_password'],
        'OPENPROJECT_SEED__LOCALE': 'de', 'IMAP_ENABLED': 'false',
        'RAILS_MIN_THREADS': '4', 'RAILS_MAX_THREADS': '16',
    }
    if c['smtp_host']:
        result.update({
            'OPENPROJECT_EMAIL__DELIVERY__METHOD': 'smtp',
            'OPENPROJECT_SMTP__ADDRESS': c['smtp_host'], 'OPENPROJECT_SMTP__PORT': str(c['smtp_port']),
            'OPENPROJECT_SMTP__USER__NAME': c['smtp_user'], 'OPENPROJECT_SMTP__PASSWORD': c['smtp_password'],
            'OPENPROJECT_SMTP__AUTHENTICATION': c['smtp_authentication'],
            'OPENPROJECT_SMTP__ENABLE__STARTTLS__AUTO': str(c['smtp_starttls']).lower(),
            'OPENPROJECT_MAIL__FROM': c['smtp_from'],
        })
    return result


class Site:
    def __init__(self, domain):
        self.domain = config.domain(domain)
        self.project = config.site_id(self.domain)
        self.base = Path(os.environ.get('OP_SITE_ROOT', '/srv/openproject')).resolve()
        self.path = self.base / self.domain
        self.hostvars = Path(os.environ.get('OP_HOSTVARS_DIR', ROOT / 'ansible/hostvars')).resolve() / (self.domain + '.yml')
        self.backups = Path(os.environ.get('OP_BACKUP_DIR', ROOT / 'backups/openproject')).resolve() / self.domain

    @contextmanager
    def lock(self):
        locks = self.base / '.locks'
        locks.mkdir(mode=0o700, parents=True, exist_ok=True)
        with open(locks / self.domain, 'a') as lock:
            try:
                fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                raise RuntimeError('Eine andere Operation bearbeitet diese Domain') from None
            try:
                yield
            finally:
                fcntl.flock(lock, fcntl.LOCK_UN)

    def load(self, installed=False):
        data = config.read_config(self.path / 'hostvars.yml' if installed else self.hostvars)
        if data['domain'] != self.domain:
            raise ValueError('Hostvars-Domain stimmt nicht ueberein')
        return data

    def compose(self, *args, **kwargs):
        return run(['docker', 'compose', '--project-name', self.project,
                    '--project-directory', self.path, '--env-file', '/dev/null',
                    '-f', self.path / 'compose.yml', *args], **kwargs)

    def preflight(self, ansible=False):
        for binary in ['docker'] + (['ansible-playbook'] if ansible else []):
            if not shutil.which(binary):
                raise RuntimeError('Tool fehlt: ' + binary)
        run(['docker', 'info', '--format', '{{.ServerVersion}}'])
        run(['docker', 'network', 'inspect', 'traefik'])
        ver = run(['docker', 'compose', 'version', '--short']).strip().lstrip('v').split('-')[0]
        if tuple(map(int, ver.split('.')[:2])) < (2, 30):
            raise RuntimeError('Docker Compose >= 2.30 erforderlich (raw env_file)')

    def pull(self, data):
        for image in config.images(data).values():
            print('Lade Image: ' + image, flush=True)
            run(['docker', 'pull', image])

    def render(self, data):
        config.validate(data)
        self.path.mkdir(mode=0o700, parents=True, exist_ok=True)
        values = {'target_domain': self.domain, 'op_data': data, 'op_site_dir': str(self.path),
                  'op_project': self.project, 'op_web_alias': self.project[:35] + '.' + self.project[35:] + '.internal', 'op_images': config.images(data), 'op_env': environment(data)}
        with tempfile.TemporaryDirectory(prefix='op-render-') as folder:
            params = Path(folder) / 'vars.json'
            params.write_text(json.dumps(values))
            params.chmod(0o600)
            run(['ansible-playbook', '-i', ROOT / 'ansible/inventory/hosts.ini',
                 ROOT / 'ansible/playbooks/deploy-openproject.yml', '--extra-vars', '@' + str(params)])
        self.compose('--profile', 'tools', 'config', '--quiet')

    def maintenance(self, backup):
        self.path.mkdir(mode=0o700, parents=True, exist_ok=True)
        (self.path / 'MAINTENANCE').write_text('Recovery-Backup: ' + str(backup) + '\n')

    def available(self):
        if (self.path / 'MAINTENANCE').exists():
            raise RuntimeError('Instanz im Wartungszustand: MAINTENANCE lesen und Restore ausfuehren')

    def start(self, migrate=True):
        print('Starte Datenbank und Cache: ' + self.domain, flush=True)
        self.compose('up', '-d', '--wait', '--wait-timeout', '180', 'db', 'cache')
        if migrate:
            print('Initialisierung/Migration: ' + self.domain, flush=True)
            self.compose('run', '--rm', '--no-deps', '-T', 'seeder')
        print('Starte Anwendung und pruefe Bereitschaft (max. 600 s).', flush=True)
        self.compose('up', '-d', '--no-deps', '--wait', '--wait-timeout', '600', *WRITERS)
        running = self.compose('--profile', 'tools', 'ps', '--status', 'running', '--services').split()
        if not set(WRITERS).issubset(running):
            raise RuntimeError('Nicht alle OpenProject-Dienste laufen')

    def backup_path(self):
        stamp = datetime.now(timezone.utc).strftime('%Y%m%dT%H%M%S%fZ')
        return self.backups / ('op-' + stamp + '.tar.gz')

    def helper(self, script, **kwargs):
        return self.compose('run', '--rm', '--no-deps', '-T', '--user', '0',
                            '--label', 'traefik.enable=false', '--entrypoint', '/bin/sh', 'web', '-c', script, **kwargs)

    def backup(self, output, resume=True):
        data = self.load(installed=True)
        completed = False
        output = Path(output).resolve()
        if output.exists():
            raise FileExistsError('Backup-Ziel existiert bereits')
        running = self.compose('--profile', 'tools', 'ps', '--status', 'running', '--services').split()
        if 'seeder' in running:
            raise RuntimeError('Seeder laeuft; kein konsistentes Backup moeglich')
        prior = [name for name in WRITERS if name in running]
        print('Sichere Datenbank, Anhaenge und Konfiguration: ' + self.domain, flush=True)
        with tempfile.TemporaryDirectory(prefix='op-backup-') as folder:
            stage = Path(folder)
            try:
                if prior:
                    self.compose('stop', '--timeout', '60', *prior)
                with open(stage / 'database.dump', 'wb') as stream:
                    self.compose('exec', '-T', 'db', 'pg_dump', '-U', 'postgres', '-d', 'openproject', '-Fc', stdout=stream)
                with open(stage / 'assets.tar', 'wb') as stream:
                    self.helper('tar -C /var/openproject/assets -cf - .', stdout=stream)
                shutil.copyfile(self.path / 'hostvars.yml', stage / 'hostvars.yml')
                for name in ('compose.yml', 'openproject.env'):
                    shutil.copyfile(self.path / name, stage / name)
                image_ids = {key: run(['docker', 'image', 'inspect', '--format', '{{.Id}}', image]).strip()
                             for key, image in config.images(data).items()}
                archive.pack(stage, output, self.domain, image_ids)
                completed = True
            finally:
                if prior and (resume or not completed):
                    self.compose('start', *prior)
        print('Backup: ' + str(output), flush=True)
        return output

    def update(self, target):
        self.available()
        data = self.load()
        installed = self.load(installed=True)
        config.check_upgrade(installed['op']['version'], target)
        for key in ('db_password', 'postgres_version'):
            if data['op'][key] != installed['op'][key]:
                raise ValueError('Datenbank-Konfiguration braucht separaten Migrationsablauf: ' + key)
        proposed = json.loads(json.dumps(data))
        proposed['op']['version'] = target
        self.pull(proposed)
        backup = self.backup_path()
        self.backup(backup, resume=False)
        self.maintenance(backup)
        try:
            self.compose('--profile', 'tools', 'stop', '--timeout', '60', *WRITERS, 'seeder')
            config.write_config(self.hostvars, proposed, replace=True)
            self.render(proposed)
            self.start()
            (self.path / 'MAINTENANCE').unlink()
        except BaseException:
            # Deliberately do not restart an old application against migrated data.
            self.compose('--profile', 'tools', 'stop', '--timeout', '60', *WRITERS, 'seeder')
            raise RuntimeError('Update fehlgeschlagen; Wartungszustand. Recovery-Backup: ' + str(backup)) from None

    def restore(self, stage, assume_yes=False, recover=False):
        data = config.read_config(Path(stage) / 'hostvars.yml')
        if data['domain'] != self.domain:
            raise ValueError('Restore-Domain stimmt nicht')
        existing = self.hostvars.exists()
        marker = self.path / 'MAINTENANCE'
        if marker.exists() and not recover:
            raise ValueError('Wartungszustand: urspruengliches Backup mit --recover wiederherstellen')
        if recover and (not existing or not marker.is_file()):
            raise ValueError('--recover erfordert eine vorhandene Instanz im Wartungszustand')
        if recover and set(self.compose('--profile', 'tools', 'ps', '--status', 'running', '--services').split()) & (set(WRITERS) | {'seeder'}):
            raise ValueError('--recover erfordert gestoppte Anwendungsdienste')
        if existing:
            current = self.load(installed=(self.path / 'hostvars.yml').exists())
            if current['op']['postgres_version'] != data['op']['postgres_version']:
                raise ValueError('Restore erfordert dieselbe PostgreSQL-Version')
            if not assume_yes:
                if not os.isatty(0) or input('Instanz ' + self.domain + ' ueberschreiben? [ja/NEIN] ') != 'ja':
                    raise RuntimeError('Restore abgebrochen; --yes fuer nicht-interaktiven Betrieb')
        elif self.path.exists():
            raise RuntimeError('Laufzeitverzeichnis ohne Hostvars vorhanden; manuell pruefen')
        self.pull(data)
        recovery = self.backup_path() if existing else 'keine vorherige Instanz'
        if recover:
            recovery = marker.read_text().strip()
        elif existing:
            self.backup(recovery, resume=False)
        if not recover:
            self.maintenance(recovery)
        try:
            if existing:
                self.compose('--profile', 'tools', 'stop', '--timeout', '60', *WRITERS, 'seeder')
            config.write_config(self.hostvars, data, replace=existing)
            # Regenerate from validated data; never run archived Compose/YAML as code.
            self.render(data)
            self.compose('up', '-d', '--wait', '--wait-timeout', '180', 'db', 'cache')
            self.compose('exec', '-T', 'db', 'dropdb', '-U', 'postgres', '--if-exists', '--force', 'openproject')
            self.compose('exec', '-T', 'db', 'createdb', '-U', 'postgres', '-O', 'postgres', 'openproject')
            with open(Path(stage) / 'database.dump', 'rb') as stream:
                self.compose('exec', '-T', 'db', 'pg_restore', '-U', 'postgres', '-d', 'openproject',
                             '--no-owner', '--no-acl', '--exit-on-error', stdin=stream)
            with tempfile.TemporaryFile() as sql:
                sql.write(("ALTER USER postgres PASSWORD '" + data['op']['db_password'] + "';\n").encode())
                sql.seek(0)
                self.compose('exec', '-T', 'db', 'psql', '-U', 'postgres', '-d', 'postgres', '-v', 'ON_ERROR_STOP=1', stdin=sql)
            with open(Path(stage) / 'assets.tar', 'rb') as stream:
                self.helper('find /var/openproject/assets -mindepth 1 -delete && '
                            'tar -C /var/openproject/assets --no-same-owner -xf - && '
                            'chown -R app:app /var/openproject/assets', stdin=stream)
            self.start(migrate=False)
            (self.path / 'MAINTENANCE').unlink()
        except BaseException:
            self.compose('--profile', 'tools', 'stop', '--timeout', '60', *WRITERS, 'seeder')
            raise RuntimeError('Restore fehlgeschlagen; Wartungszustand. Vorsicherung: ' + str(recovery)) from None
