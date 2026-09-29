"""Opt-in real Docker roundtrip on a NEW disposable domain; leaves evidence intact."""
import argparse
from pathlib import Path
import subprocess
import sys
import tempfile
sys.path.insert(0, str(Path(__file__).resolve().parents[3] / 'scripts/lib'))
from openproject_runtime import ROOT, Site

p = argparse.ArgumentParser(description=__doc__)
p.add_argument('--domain', required=True)
p.add_argument('--confirm-disposable', action='store_true', required=True)
p.add_argument('--update-version', help='Optional newer same-major release to test')
a = p.parse_args()
site = Site(a.domain)
if site.hostvars.exists() or site.path.exists():
    p.error('Domain is not empty; refusing to touch an existing instance')

def cli(action, *args):
    subprocess.run([sys.executable, str(ROOT/'scripts/lib/openproject_cli.py'), action,
                    '--domain='+site.domain, *args], check=True)

def sql(statement):
    with tempfile.TemporaryFile() as stream:
        stream.write(statement.encode())
        stream.seek(0)
        return site.compose('exec', '-T', 'db', 'psql', '-U', 'postgres', '-d', 'openproject',
                            '-v', 'ON_ERROR_STOP=1', '-At', stdin=stream).strip()

cli('add', '--skip-dns-check')
# Probe row and attachment bytes cover both persistence stores without API tokens.
# The README also requires a real work-package/upload and public HTTPS smoke test.
with site.lock():
    sql("CREATE TABLE infra_restore_probe (value text); INSERT INTO infra_restore_probe VALUES ('before');")
    site.helper("printf before > /var/openproject/assets/infra-restore-probe.txt")
backup = site.backup_path()
cli('backup', '--output='+str(backup))
with site.lock():
    sql("UPDATE infra_restore_probe SET value='after';")
    site.helper("printf after > /var/openproject/assets/infra-restore-probe.txt")
cli('restore', '--backup='+str(backup), '--dry-run')
cli('restore', '--backup='+str(backup), '--yes')
with site.lock():
    assert sql('SELECT value FROM infra_restore_probe;') == 'before'
    assert site.helper('cat /var/openproject/assets/infra-restore-probe.txt') == 'before'
if a.update_version:
    cli('update', '--version='+a.update_version)
    with site.lock():
        assert sql('SELECT value FROM infra_restore_probe;') == 'before'
print('ROUNDTRIP PASSED. Test instance and backups retained:', site.domain)
