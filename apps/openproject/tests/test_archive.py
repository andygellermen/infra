import io
import json
from pathlib import Path
import sys
import tarfile
import tempfile
import unittest
sys.path.insert(0, str(Path(__file__).resolve().parents[3] / 'scripts/lib'))
import openproject_config as config
try:
    import openproject_archive as archive
except ImportError:
    archive = None

class ArchiveTests(unittest.TestCase):
    def setUp(self):
        self.assertIsNotNone(archive, 'archive module missing')
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.stage = self.root / 'stage'
        self.stage.mkdir()
        config.write_config(self.stage / 'hostvars.yml', config.new_config('op.example.org'))
        (self.stage / 'database.dump').write_bytes(b'PGDMPfixture')
        (self.stage / 'compose.yml').write_text('services: {}')
        (self.stage / 'openproject.env').write_text('SECRET=private')
        with tarfile.open(self.stage / 'assets.tar', 'w') as tf:
            member = tarfile.TarInfo('files/test.txt')
            member.size = 5
            tf.addfile(member, io.BytesIO(b'hello'))
        self.target = self.root / 'backup.tar.gz'

    def pack(self):
        archive.pack(self.stage, self.target, 'op.example.org', {})

    def test_roundtrip_and_no_overwrite(self):
        self.pack()
        self.assertEqual(self.target.stat().st_mode & 0o777, 0o600)
        with self.assertRaises(FileExistsError):
            self.pack()
        dest = self.root / 'extract'
        archive.extract(self.target, dest, 'op.example.org')
        self.assertEqual((dest / 'database.dump').read_bytes(), b'PGDMPfixture')
        with self.assertRaises(ValueError):
            archive.extract(self.target, self.root / 'wrong', 'other.example.org')

    def test_bad_nested_paths_and_links(self):
        for name, kind in [('../escaped', tarfile.REGTYPE), ('/etc/escaped', tarfile.REGTYPE),
                           ('link', tarfile.SYMTYPE), ('hard', tarfile.LNKTYPE)]:
            with self.subTest(name=name):
                with tarfile.open(self.stage / 'assets.tar', 'w') as tf:
                    m = tarfile.TarInfo(name)
                    m.type = kind
                    m.linkname = '/etc/passwd'
                    tf.addfile(m)
                with self.assertRaises(ValueError):
                    self.pack()
                self.assertFalse(self.target.exists())

    def test_checksum_and_duplicates_rejected(self):
        self.pack()
        bad = self.root / 'bad.tar.gz'
        with tarfile.open(self.target) as src, tarfile.open(bad, 'w:gz') as dst:
            for m in src:
                payload = src.extractfile(m).read()
                if m.name == 'database.dump':
                    payload = b'PGDMPtampered'
                m.size = len(payload)
                dst.addfile(m, io.BytesIO(payload))
        with self.assertRaises(ValueError):
            archive.extract(bad, self.root / 'badout', 'op.example.org')
        with tarfile.open(self.target) as src, tarfile.open(bad, 'w:gz') as dst:
            for m in src:
                payload = src.extractfile(m).read()
                dst.addfile(m, io.BytesIO(payload))
                if m.name == 'database.dump':
                    dst.addfile(m, io.BytesIO(payload))
        with self.assertRaises(ValueError):
            archive.extract(bad, self.root / 'dupout', 'op.example.org')

    def test_setuid_payload_rejected(self):
        with tarfile.open(self.stage / 'assets.tar', 'w') as tf:
            member = tarfile.TarInfo('unexpected-executable')
            member.mode = 0o4755
            tf.addfile(member)
        with self.assertRaises(ValueError): self.pack()
