import argparse
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch
sys.path.insert(0, str(Path(__file__).resolve().parents[3] / 'scripts/lib'))
import openproject_config as config
import openproject_cli as cli
from openproject_runtime import Site

class AddTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        env = patch.dict(os.environ, {'OP_SITE_ROOT':str(self.root/'sites'), 'OP_HOSTVARS_DIR':str(self.root/'vars')})
        env.start()
        self.addCleanup(env.stop)
        self.site = Site('op.example.org')
        self.args = argparse.Namespace(version='17.8.0', wildcard_domain='', dns_account='', skip_dns_check=True, resume=False)

    def test_existing_instance_not_overwritten(self):
        config.write_config(self.site.hostvars, config.new_config(self.site.domain))
        before = self.site.hostvars.read_bytes()
        with patch.object(self.site, 'pull') as pull:
            with self.assertRaises(FileExistsError): cli.add(self.site, self.args)
        self.assertEqual(before, self.site.hostvars.read_bytes())
        pull.assert_not_called()

    def test_resume_retains_original_secrets(self):
        config.write_config(self.site.hostvars, config.new_config(self.site.domain))
        self.site.maintenance('Neuinstallation; Details im OpenProject-Runbook')
        self.args.resume = True
        before = self.site.hostvars.read_bytes()
        with patch.object(self.site, 'pull'), patch.object(self.site, 'render'), \
             patch.object(self.site, 'start'), patch.object(self.site, 'compose', return_value=''), patch('openproject_cli.run'):
            cli.add(self.site, self.args)
        self.assertEqual(before, self.site.hostvars.read_bytes())
        self.assertFalse((self.site.path/'MAINTENANCE').exists())

    def test_resume_cannot_reseed_failed_restore(self):
        config.write_config(self.site.hostvars, config.new_config(self.site.domain))
        self.site.maintenance('/backups/original.tar.gz')
        self.args.resume = True
        with self.assertRaises(ValueError): cli.add(self.site, self.args)
