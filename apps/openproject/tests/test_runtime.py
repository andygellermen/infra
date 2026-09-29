import argparse
import fcntl
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch
sys.path.insert(0, str(Path(__file__).resolve().parents[3] / 'scripts/lib'))
import openproject_config as config
try:
    import openproject_runtime as runtime
except ImportError:
    runtime = None

ROOT = Path(__file__).resolve().parents[3]
class RuntimeTests(unittest.TestCase):
    def setUp(self):
        self.assertIsNotNone(runtime, 'runtime module missing')
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        env = patch.dict(os.environ, {'OP_SITE_ROOT':str(self.root/'sites'),
              'OP_HOSTVARS_DIR':str(self.root/'vars'), 'OP_BACKUP_DIR':str(self.root/'backups')})
        env.start()
        self.addCleanup(env.stop)
        self.site = runtime.Site('op.example.org')

    def install(self, data):
        config.write_config(self.site.hostvars, data)
        self.site.path.mkdir(parents=True, exist_ok=True)
        config.write_config(self.site.path/'hostvars.yml', data)

    def test_lock_contention(self):
        with self.site.lock():
            with self.assertRaises(RuntimeError):
                with runtime.Site('op.example.org').lock():
                    pass

    def test_raw_environment_secret(self):
        data = config.new_config('op.example.org')
        data['op']['smtp_host'] = 'smtp.example.org'
        data['op']['smtp_password'] = '$cash"quote'
        env = runtime.environment(data)
        self.assertEqual(env['OPENPROJECT_SMTP__PASSWORD'], '$cash"quote')
        self.assertEqual(env['OPENPROJECT_HOST__NAME'], 'op.example.org')

    def test_cli_help_and_rejections(self):
        for name in ('add', 'update', 'backup', 'restore'):
            script = ROOT / ('scripts/op-'+name+'.sh')
            result = subprocess.run(['bash', str(script), '--help'], capture_output=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            for args in [[], ['--domain=../etc'], ['--domain=op.example.org', '--nonsense']]:
                result = subprocess.run(['bash', str(script)] + args, input=b'', capture_output=True)
                self.assertNotEqual(result.returncode, 0)
                self.assertNotIn(b'Traceback', result.stderr)

    def test_backup_failure_restarts_only_prior_services(self):
        data = config.new_config(self.site.domain)
        self.install(data)
        self.site.path.mkdir(parents=True, exist_ok=True)
        events = []
        def compose(*args, **kwargs):
            events.append(args)
            if 'ps' in args: return 'web\ncron\n'
            if 'pg_dump' in args: raise RuntimeError('dump failed')
            return ''
        with patch.object(self.site, 'compose', side_effect=compose):
            with self.assertRaises(RuntimeError):
                self.site.backup(self.root/'failed.tar.gz')
        self.assertIn(('start', 'web', 'cron'), events)
        self.assertFalse((self.root/'failed.tar.gz').exists())

    def test_update_backup_failure_prevents_migration(self):
        self.install(config.new_config(self.site.domain, '17.7.0'))
        events = []
        with patch.object(self.site, 'pull', side_effect=lambda d: events.append('pull')), \
             patch.object(self.site, 'backup', side_effect=RuntimeError('backup failed')), \
             patch.object(self.site, 'compose', side_effect=lambda *a, **k: events.append(a)), \
             patch.object(self.site, 'render', side_effect=lambda d: events.append('render')):
            with self.assertRaises(RuntimeError): self.site.update('17.8.0')
        self.assertEqual(events, ['pull'])
        self.assertEqual(config.read_config(self.site.hostvars)['op']['version'], '17.7.0')

    def test_failed_migration_stays_stopped(self):
        self.install(config.new_config(self.site.domain, '17.7.0'))
        self.site.path.mkdir(parents=True, exist_ok=True)
        events = []
        def compose(*args, **kwargs):
            events.append(args)
            if args[:1] == ('run',) and 'seeder' in args: raise RuntimeError('migration failed')
            return ''
        with patch.object(self.site, 'pull'), patch.object(self.site, 'backup'), \
             patch.object(self.site, 'render'), patch.object(self.site, 'compose', side_effect=compose):
            with self.assertRaises(RuntimeError): self.site.update('17.8.0')
        self.assertTrue((self.site.path/'MAINTENANCE').exists())
        self.assertFalse(any(a[:1] == ('start',) for a in events))
        self.assertFalse(any(a[:1] == ('up',) and 'web' in a for a in events))

    def test_missing_image_keeps_configuration(self):
        self.install(config.new_config(self.site.domain, '17.7.0'))
        with patch.object(self.site, 'pull', side_effect=RuntimeError('missing image')), \
             patch.object(self.site, 'backup') as backup:
            with self.assertRaises(RuntimeError): self.site.update('17.8.0')
        backup.assert_not_called()
        self.assertEqual(config.read_config(self.site.hostvars)['op']['version'], '17.7.0')

    def test_successful_backup_preserves_stopped_services(self):
        import io
        import tarfile
        import openproject_archive as archive
        self.install(config.new_config(self.site.domain))
        self.site.path.mkdir(parents=True, exist_ok=True)
        for name in ('compose.yml', 'openproject.env'):
            (self.site.path/name).write_text('fixture')
        payload = io.BytesIO()
        with tarfile.open(fileobj=payload, mode='w') as tf:
            m = tarfile.TarInfo('file.txt')
            m.size = 5
            tf.addfile(m, io.BytesIO(b'hello'))
        events = []
        def compose(*args, **kwargs):
            events.append(args)
            if 'ps' in args: return 'web\n'
            if 'pg_dump' in args: kwargs['stdout'].write(b'PGDMPtest')
            return ''
        def helper(script, **kwargs): kwargs['stdout'].write(payload.getvalue())
        target = self.root/'ok.tar.gz'
        desired = config.read_config(self.site.hostvars)
        desired['op']['smtp_host'] = 'next.example.org'
        config.write_config(self.site.hostvars, desired, replace=True)
        with patch.object(self.site, 'compose', side_effect=compose), patch.object(self.site, 'helper', side_effect=helper), \
             patch('openproject_runtime.run', return_value='sha256:'+'a'*64):
            self.site.backup(target)
        archive.extract(target, self.root/'validated', self.site.domain)
        self.assertEqual(config.read_config(self.root/'validated/hostvars.yml')['op']['smtp_host'], '')
        self.assertIn(('start', 'web'), events)
        self.assertNotIn(('start', 'worker'), events)
        with patch.object(self.site, 'compose') as docker:
            with self.assertRaises(FileExistsError): self.site.backup(target)
        docker.assert_not_called()

    def test_restart_failure_reports_error(self):
        self.install(config.new_config(self.site.domain))
        def compose(*args, **kwargs):
            if 'ps' in args: return 'web\n'
            if args[:1] == ('start',): raise RuntimeError('restart failed')
            if 'pg_dump' in args: raise RuntimeError('dump failed')
            return ''
        with patch.object(self.site, 'compose', side_effect=compose):
            with self.assertRaisesRegex(RuntimeError, 'restart failed'):
                self.site.backup(self.root/'failed.tar.gz')

    def test_restore_backup_failure_aborts_before_changes(self):
        self.install(config.new_config(self.site.domain))
        stage = self.root/'stage'
        stage.mkdir()
        config.write_config(stage/'hostvars.yml', config.new_config(self.site.domain))
        before = self.site.hostvars.read_bytes()
        with patch.object(self.site, 'pull'), patch.object(self.site, 'backup', side_effect=RuntimeError('failed')), \
             patch.object(self.site, 'compose') as docker:
            with self.assertRaises(RuntimeError): self.site.restore(stage, assume_yes=True)
        docker.assert_not_called()
        self.assertEqual(self.site.hostvars.read_bytes(), before)

    def test_restore_import_failure_does_not_start_app(self):
        stage = self.root/'stage'
        stage.mkdir()
        config.write_config(stage/'hostvars.yml', config.new_config(self.site.domain))
        (stage/'database.dump').write_bytes(b'PGDMPtest')
        events = []
        def compose(*args, **kwargs):
            events.append(args)
            if 'pg_restore' in args: raise RuntimeError('import failed')
            return ''
        with patch.object(self.site, 'pull'), patch.object(self.site, 'render'), \
             patch.object(self.site, 'compose', side_effect=compose):
            with self.assertRaises(RuntimeError): self.site.restore(stage)
        self.assertTrue((self.site.path/'MAINTENANCE').exists())
        self.assertFalse(any(a[:1] == ('up',) and 'web' in a for a in events))

    def test_seeder_failure_never_starts_app(self):
        events = []
        def compose(*args, **kwargs):
            events.append(args)
            if args[:1] == ('run',) and 'seeder' in args: raise RuntimeError('seed failed')
            return ''
        with patch.object(self.site, 'compose', side_effect=compose):
            with self.assertRaises(RuntimeError): self.site.start()
        self.assertFalse(any(a[:1] == ('up',) and 'web' in a for a in events))

    def test_update_keeps_writers_stopped_after_snapshot(self):
        self.install(config.new_config(self.site.domain, '17.7.0'))
        with patch.object(self.site, 'pull'), patch.object(self.site, 'backup') as backup, \
             patch.object(self.site, 'render'), patch.object(self.site, 'start'), patch.object(self.site, 'compose'):
            self.site.update('17.8.0')
        self.assertFalse(backup.call_args.kwargs.get('resume', True))

    def test_explicit_maintenance_recovery_skips_broken_database_dump(self):
        self.install(config.new_config(self.site.domain))
        self.site.maintenance('/safe/original-backup.tar.gz')
        stage = self.root/'stage'
        stage.mkdir()
        config.write_config(stage/'hostvars.yml', config.new_config(self.site.domain))
        (stage/'database.dump').write_bytes(b'PGDMPtest')
        (stage/'assets.tar').write_bytes(b'fixture')
        with patch.object(self.site, 'pull'), patch.object(self.site, 'backup') as backup, \
             patch.object(self.site, 'render'), patch.object(self.site, 'compose', return_value=''), \
             patch.object(self.site, 'helper'), patch.object(self.site, 'start'):
            self.site.restore(stage, assume_yes=True, recover=True)
        backup.assert_not_called()
        self.assertFalse((self.site.path/'MAINTENANCE').exists())

    def test_recovery_flag_rejected_for_healthy_instance(self):
        self.install(config.new_config(self.site.domain))
        stage = self.root/'stage'
        stage.mkdir()
        config.write_config(stage/'hostvars.yml', config.new_config(self.site.domain))
        with self.assertRaises(ValueError):
            self.site.restore(stage, assume_yes=True, recover=True)

    def test_recovery_rejects_running_seeder(self):
        self.install(config.new_config(self.site.domain))
        self.site.maintenance('/safe/original.tar.gz')
        stage = self.root/'stage'
        stage.mkdir()
        config.write_config(stage/'hostvars.yml', config.new_config(self.site.domain))
        with patch.object(self.site, 'compose', return_value='seeder\n'):
            with self.assertRaises(ValueError): self.site.restore(stage, assume_yes=True, recover=True)
