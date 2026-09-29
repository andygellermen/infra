"""Strict, data-only configuration for the OpenProject lifecycle scripts."""
import hashlib
import json
import os
from pathlib import Path
import re
import secrets
import tempfile
import yaml

DEFAULT_VERSION = '17.8.0'
DEFAULTS = dict(version=DEFAULT_VERSION, postgres_version='17.11', cache_version='1.6.45-alpine',
                hocuspocus_version='17.8.0', wildcard_domain='', dns_account='', middleware='',
                smtp_host='', smtp_port=587, smtp_user='', smtp_password='',
                smtp_from='', smtp_authentication='plain', smtp_starttls=True)


def domain(value):
    if not isinstance(value, str) or value != value.strip():
        raise ValueError('Ungueltige Domain')
    try:
        value = value.encode('idna').decode('ascii').lower()
    except UnicodeError:
        raise ValueError('Ungueltige Domain') from None
    labels = value.split('.')
    if len(value) > 253 or len(labels) < 2 or any(
        not re.fullmatch(r'[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?', p) for p in labels
    ) or labels[-1].isdigit():
        raise ValueError('Ungueltige Domain')
    return value


def site_id(value):
    return 'op-' + hashlib.sha256(domain(value).encode()).hexdigest()


def version(value):
    if not isinstance(value, str) or not re.fullmatch(r'[0-9]+\.[0-9]+\.[0-9]+', value):
        raise ValueError('Version muss X.Y.Z sein')
    return tuple(map(int, value.split('.')))


def check_upgrade(current, target):
    old, new = version(current), version(target)
    if old[0] != new[0] or new < old:
        raise ValueError('Major-Wechsel und Downgrades sind nicht unterstuetzt')


def new_config(name, release=DEFAULT_VERSION):
    return {'domain': domain(name), 'op_enabled': True, 'op': dict(
        DEFAULTS, version=release, db_password=secrets.token_hex(32),
        secret_key=secrets.token_hex(64), collaborative_secret=secrets.token_hex(32),
        admin_password=secrets.token_hex(24))}


def validate(data):
    if not isinstance(data, dict) or set(data) - {'domain', 'op_enabled', 'op', 'tls_mode', 'tls_wildcard_domain', 'tls_dns_account'} or not {'domain', 'op_enabled', 'op'}.issubset(data):
        raise ValueError('Ungueltige OpenProject-Hostvars-Struktur')
    if data['op_enabled'] is not True or data['domain'] != domain(data['domain']):
        raise ValueError('OpenProject-Domain nicht aktiviert/normalisiert')
    cfg = data['op']
    keys = set(DEFAULTS) | {'db_password', 'secret_key', 'collaborative_secret', 'admin_password'}
    if not isinstance(cfg, dict) or set(cfg) != keys:
        raise ValueError('Unbekannte oder fehlende OpenProject-Konfigurationsfelder')
    for key, value in cfg.items():
        if key in ('smtp_port', 'smtp_starttls'):
            continue
        if not isinstance(value, str) or any(ord(c) < 32 or ord(c) == 127 for c in value) or '{{' in value or '{%' in value or '{#' in value:
            raise ValueError('Ungueltiger Konfigurationswert: ' + key)
    if version(cfg['version'])[0] != 17:
        raise ValueError('Diese Integration unterstuetzt OpenProject 17')
    if version(cfg['hocuspocus_version'])[0] != 17:
        raise ValueError('Hocuspocus muss Version 17 sein')
    if not re.fullmatch(r'17\.[0-9]+', cfg['postgres_version']):
        raise ValueError('PostgreSQL 17.x erforderlich')
    if not re.fullmatch(r'1\.6\.[0-9]+-alpine', cfg['cache_version']):
        raise ValueError('Memcached 1.6.x-alpine erforderlich')
    for key in ('db_password', 'secret_key', 'collaborative_secret', 'admin_password'):
        if not re.fullmatch(r'[a-f0-9]{32,128}', cfg[key]):
            raise ValueError('Ungueltiges generiertes Secret: ' + key)
    if type(cfg['smtp_port']) is not int or not 1 <= cfg['smtp_port'] <= 65535:
        raise ValueError('Ungueltiger SMTP-Port')
    if type(cfg['smtp_starttls']) is not bool:
        raise ValueError('smtp_starttls muss boolesch sein')
    if cfg['smtp_authentication'] not in ('plain', 'login', 'cram_md5'):
        raise ValueError('Ungueltige SMTP-Authentifizierung')
    if cfg['wildcard_domain']:
        apex = domain(cfg['wildcard_domain'])
        if data['domain'] != apex and data['domain'].split('.', 1)[1] != apex:
            raise ValueError('Domain wird vom Wildcard-Zertifikat nicht abgedeckt')
    if cfg['middleware'] and not re.fullmatch(r'[A-Za-z0-9_@,.-]+', cfg['middleware']):
        raise ValueError('Ungueltige Middleware')
    if cfg['dns_account'] and not re.fullmatch(r'[A-Za-z0-9_-]+', cfg['dns_account']):
        raise ValueError('Ungueltiger DNS-Account')
    metadata = dict(tls_mode='wildcard' if cfg['wildcard_domain'] else 'standard',
                    tls_wildcard_domain=cfg['wildcard_domain'], tls_dns_account=cfg['dns_account'])
    if any(key in data and data[key] != value for key, value in metadata.items()):
        raise ValueError('Inkonsistente TLS-Metadaten')
    return data


def read_config(path):
    with open(path, encoding='utf-8') as stream:
        return validate(yaml.safe_load(stream))


def write_config(path, data, replace=False):
    data = dict(data, tls_mode='wildcard' if data['op']['wildcard_domain'] else 'standard',
                tls_wildcard_domain=data['op']['wildcard_domain'], tls_dns_account=data['op']['dns_account'])
    validate(data)
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    fd, temp = tempfile.mkstemp(dir=path.parent)
    try:
        with os.fdopen(fd, 'w') as stream:
            json.dump(data, stream, indent=2)  # JSON is a safe YAML subset.
            stream.write('\n')
        if replace:
            os.replace(temp, path)
        else:
            os.link(temp, path)  # atomic, refuses to overwrite existing files
    finally:
        if os.path.exists(temp):
            os.unlink(temp)


def images(data):
    c = validate(data)['op']
    return {'app': 'openproject/openproject:' + c['version'] + '-slim',
            'db': 'postgres:' + c['postgres_version'],
            'cache': 'memcached:' + c['cache_version'],
            'hocuspocus': 'openproject/hocuspocus:' + c['hocuspocus_version']}
