import sys
import unittest
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[3] / 'scripts/lib'))
try:
    import openproject_config as config
except ImportError:
    config = None

class ConfigTests(unittest.TestCase):
    def setUp(self):
        self.assertIsNotNone(config, 'configuration module missing')

    def test_domains(self):
        self.assertEqual(config.domain('OP.Example.org'), 'op.example.org')
        for value in ['../etc', 'example', 'a..org', '-a.org', 'a.org/x', 'a.org\n', 'a.org:443']:
            with self.subTest(value=value), self.assertRaises(ValueError):
                config.domain(value)

    def test_distinct_ids(self):
        self.assertNotEqual(config.site_id('a-b.example.org'), config.site_id('a.b-example.org'))

    def test_upgrade_bounds(self):
        config.check_upgrade('17.7.0', '17.8.0')
        for target in ['18.0.0', '17.6.0', 'latest', '17', '17.8.0;id']:
            with self.subTest(target=target), self.assertRaises(ValueError):
                config.check_upgrade('17.7.0', target)

    def test_config_rejects_unknown_and_secret_newline(self):
        c = config.new_config('op.example.org')
        config.validate(c)
        c['op']['smtp_password'] = '$foo"bar'
        config.validate(c)
        c['op']['smtp_password'] = 'foo\nBAR=evil'
        with self.assertRaises(ValueError):
            config.validate(c)
        c = config.new_config('op.example.org')
        c['op']['command'] = 'evil'
        with self.assertRaises(ValueError):
            config.validate(c)
