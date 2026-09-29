"""Versioned backup archives; never execute saved configuration or extract links."""
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import shutil
import tarfile
import tempfile
from datetime import datetime, timezone
from openproject_config import read_config

FILES = {'database.dump', 'assets.tar', 'hostvars.yml', 'compose.yml', 'openproject.env'}
MAX_BYTES = 100 * 1024 ** 3
MAX_MEMBERS = 1000000


def digest(path):
    result = hashlib.sha256()
    with open(path, 'rb') as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b''):
            result.update(chunk)
    return result.hexdigest()


def members(tf, outer=False):
    seen, size = set(), 0
    for count, member in enumerate(tf, 1):
        path = PurePosixPath(member.name)
        name = str(path)
        if (path.is_absolute() or '..' in path.parts or '\\' in member.name
                or any(ord(c) < 32 for c in member.name)
                or not (member.isfile() or member.isdir()) or member.issparse() or member.mode & 0o7000
                or name in seen or count > MAX_MEMBERS):
            raise ValueError('Unsicherer oder doppelter Archiveintrag')
        if outer and (name not in FILES | {'manifest.json'} or not member.isfile()):
            raise ValueError('Unerwarteter Archiveintrag')
        seen.add(name)
        size += member.size
        if member.size < 0 or size > MAX_BYTES:
            raise ValueError('Archivgroessenlimit (100 GiB) ueberschritten')
        yield member
    if outer and seen != FILES | {'manifest.json'}:
        raise ValueError('Backup unvollstaendig')


def check_stage(stage, expected_domain):
    stage = Path(stage)
    manifest = json.loads((stage / 'manifest.json').read_text())
    if manifest.get('format') != 1 or manifest.get('domain') != expected_domain:
        raise ValueError('Backup-Format oder Domain stimmt nicht ueberein')
    if set(manifest.get('sha256', {})) != FILES:
        raise ValueError('Unvollstaendige Pruefsummen')
    for name in FILES:
        if digest(stage / name) != manifest['sha256'][name]:
            raise ValueError('Backup-Pruefsumme stimmt nicht: ' + name)
    data = read_config(stage / 'hostvars.yml')
    if data['domain'] != expected_domain or data['op']['version'] != manifest.get('version'):
        raise ValueError('Hostvars passen nicht zum Backup')
    with open(stage / 'database.dump', 'rb') as stream:
        if stream.read(5) != b'PGDMP':
            raise ValueError('Kein PostgreSQL Custom-Dump')
    with tarfile.open(stage / 'assets.tar', 'r:') as tf:
        for _ in members(tf):
            pass
    return manifest


def extract(source, destination, expected_domain):
    destination = Path(destination)
    destination.mkdir(mode=0o700, parents=True, exist_ok=False)
    try:
        with tarfile.open(source, 'r:gz') as tf:
            for member in members(tf, outer=True):
                # No tarfile.extractall: exactly five payload names and manifest.
                if member.name in ('manifest.json', 'hostvars.yml', 'compose.yml', 'openproject.env') and member.size > 1024 ** 2:
                    raise ValueError('Konfigurationsdatei zu gross')
                with tf.extractfile(member) as src, open(destination / member.name, 'xb') as dst:
                    shutil.copyfileobj(src, dst)
                os.chmod(destination / member.name, 0o600)
        return check_stage(destination, expected_domain)
    except BaseException:
        shutil.rmtree(destination)
        raise


def pack(stage, target, expected_domain, image_ids):
    stage, target = Path(stage), Path(target)
    if target.exists():
        raise FileExistsError('Backup-Ziel existiert bereits')
    data = read_config(stage / 'hostvars.yml')
    manifest = dict(format=1, domain=expected_domain, version=data['op']['version'],
                    created_at=datetime.now(timezone.utc).isoformat(), image_ids=image_ids,
                    sha256={name: digest(stage / name) for name in FILES})
    (stage / 'manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
    check_stage(stage, expected_domain)
    target.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
    fd, temporary = tempfile.mkstemp(prefix='.op-backup-', dir=target.parent)
    os.close(fd)
    try:
        with tarfile.open(temporary, 'w:gz') as tf:
            for name in sorted(FILES | {'manifest.json'}):
                tf.add(stage / name, arcname=name, recursive=False)
        # Read back the actual compressed archive before publication.
        with tempfile.TemporaryDirectory(dir=target.parent) as verify:
            extract(temporary, Path(verify) / 'checked', expected_domain)
        os.link(temporary, target)
    finally:
        os.unlink(temporary)
