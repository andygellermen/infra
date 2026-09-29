import os
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch
import yaml
sys.path.insert(0, str(Path(__file__).resolve().parents[3] / 'scripts/lib'))
import openproject_config as config
from openproject_runtime import Site

class DeploymentTests(unittest.TestCase):
    def test_render_and_compose_secret_roundtrip(self):
        with tempfile.TemporaryDirectory() as folder, patch.dict(os.environ, {'OP_SITE_ROOT':folder}):
            site = Site('op.example.org')
            data = config.new_config(site.domain)
            data['op'].update(smtp_host='smtp.example.org', smtp_password='$cash"quoted# x')
            site.render(data)
            rendered = yaml.safe_load((site.path/'compose.yml').read_text())
            self.assertNotIn('ports', rendered['services']['db'])
            self.assertNotIn('ports', rendered['services']['cache'])
            self.assertNotIn('traefik', rendered['services']['db']['networks'])
            parsed = __import__('json').loads(site.compose('--profile', 'tools', 'config', '--format', 'json'))
            # Compose config escapes dollars for re-usable YAML/JSON output.
            self.assertEqual(parsed['services']['web']['environment']['OPENPROJECT_SMTP__PASSWORD'], '$$cash"quoted# x')
            self.assertIn('OPENPROJECT_SMTP__PASSWORD=$cash"quoted# x\n', (site.path/'openproject.env').read_text())
            self.assertEqual(parsed['services']['web']['environment']['OPENPROJECT_HOST__NAME'], 'op.example.org')
            self.assertEqual((site.path/'compose.yml').stat().st_mode & 0o777, 0o600)
            self.assertIn('Host(`op.example.org`)', str(parsed['services']['web']['labels']))
            self.assertEqual(parsed['services']['seeder']['restart'], 'no')

    def test_wildcard_metadata(self):
        with tempfile.TemporaryDirectory() as folder:
            data = config.new_config('op.example.org')
            data['op'].update(wildcard_domain='example.org', dns_account='primary')
            file = Path(folder)/'hostvars.yml'
            config.write_config(file, data)
            stored = yaml.safe_load(file.read_text())
            self.assertEqual(stored['tls_mode'], 'wildcard')
            self.assertEqual(stored['tls_dns_account'], 'primary')

    def test_template_expressions_rejected(self):
        data = config.new_config('op.example.org')
        data['op']['smtp_password'] = '{{ lookup("pipe", "id") }}'
        with self.assertRaises(ValueError):
            config.validate(data)
