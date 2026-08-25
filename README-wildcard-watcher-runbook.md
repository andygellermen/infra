# Wildcard Watcher Runbook

Kompaktes Runbook fuer den Neuaufbau nach Server-Migration oder Disaster-Recovery.
Ziel: Wildcard-Zertifikate per SSH auf Zielserver verteilen und den `systemd.path`-Watcher stabil aktivieren.

## Scope
Dieses Runbook gilt fuer den Quell-/ACME-Server und die Verteilung ueber:
- `./scripts/wildcard-distribute.sh`
- `./scripts/wildcard-distribute-on-change.sh`
- `ansible/wildcards/systemd/wildcard-distribute-on-change.service`
- `ansible/wildcards/systemd/wildcard-distribute-on-change.path`

## Voraussetzungen
1. Repo liegt auf dem Quellserver unter `/home/andy/infra` (oder Pfade entsprechend anpassen).
2. `ansible/secrets/secrets.yml` und `data/traefik/acme/acme.json` sind vorhanden.
3. `ansible/wildcards/export.yml` ist gepflegt (Domains, Ziele, `post_deploy_command`).
4. SSH-Key fuer Zielzugriffe liegt auf dem Quellserver unter `/root/.ssh/id_wildcard_stage`.
5. Alle folgenden Befehle laufen als `root`.

## SSH Setup
1. Key erzeugen (falls noch nicht vorhanden):
```bash
ssh-keygen -t ed25519 -f /root/.ssh/id_wildcard_stage -C "wildcard-distribution@$(hostname -f)"
chmod 600 /root/.ssh/id_wildcard_stage
chmod 644 /root/.ssh/id_wildcard_stage.pub
```
2. Public Key auf Zielserver kopieren:
```bash
ssh-copy-id -i /root/.ssh/id_wildcard_stage.pub root@<ziel-ip>
```
3. Known Hosts pflegen:
```bash
ssh-keyscan -H <ziel-ip-1> <ziel-ip-2> <ziel-ip-3> >> /root/.ssh/known_hosts
chmod 600 /root/.ssh/known_hosts
```
Die Fingerprints muessen vor dem Eintragen ueber einen vertrauenswuerdigen zweiten Kanal
(zum Beispiel die Server-Konsole des Hosters) geprueft werden.
4. Passwortlosen Login hart pruefen:
```bash
ssh -o BatchMode=yes -o PreferredAuthentications=publickey -o PasswordAuthentication=no -i /root/.ssh/id_wildcard_stage root@<ziel-ip> true
```

## Verteilungs-Smoketest
1. Dry-Run:
```bash
cd /home/andy/infra
./scripts/wildcard-distribute.sh --all --config ./ansible/wildcards/export.yml --dry-run
```
2. Initiale echte Verteilung erzwingen:
```bash
./scripts/wildcard-distribute-on-change.sh --all --config ./ansible/wildcards/export.yml --state-dir ./data/wildcard-distribution-state --force
```
3. Direkt danach normaler Lauf ohne `--force`:
```bash
./scripts/wildcard-distribute-on-change.sh --all --config ./ansible/wildcards/export.yml --state-dir ./data/wildcard-distribution-state
```
Erwartung: unveraenderte Zertifikate werden uebersprungen.

## Systemd Watcher Aktivierung
1. Units installieren:
```bash
cp /home/andy/infra/ansible/wildcards/systemd/wildcard-distribute-on-change.service /etc/systemd/system/
cp /home/andy/infra/ansible/wildcards/systemd/wildcard-distribute-on-change.path /etc/systemd/system/
systemctl daemon-reload
```
2. Eventuelle Altfehler zuruecksetzen:
```bash
systemctl reset-failed wildcard-distribute-on-change.service
```
3. Watcher aktivieren:
```bash
systemctl enable --now wildcard-distribute-on-change.path
systemctl status wildcard-distribute-on-change.path --no-pager
```

## Trigger Test
1. Datei-Event ausloesen:
```bash
touch /home/andy/infra/data/traefik/acme/acme.json
sleep 2
```
2. Nur aktuelle Logs ansehen:
```bash
journalctl -u wildcard-distribute-on-change.service --since "2 minutes ago" --no-pager
```
Erwartung: genau ein erfolgreicher Service-Lauf pro Testevent.

## Betriebskontrolle
1. Watcher muss `active (waiting)` sein:
```bash
systemctl status wildcard-distribute-on-change.path --no-pager
```
2. Letzter Service-Exit darf nicht `failed` sein:
```bash
systemctl status wildcard-distribute-on-change.service --no-pager
```
3. State-Dateien vorhanden:
```bash
ls -la /home/andy/infra/data/wildcard-distribution-state
```

## Typische Fehlerbilder
1. `Identity file /root/.ssh/... not accessible: Permission denied`
Kontext: Skript wurde nicht als `root` gestartet.
Fix: als `root` ausfuehren oder Pfade auf User-Home umstellen.

2. `Host key verification failed` oder `REMOTE HOST IDENTIFICATION HAS CHANGED`
Kontext: `known_hosts` fehlt, ist unvollstaendig oder enthaelt nach einem legitimen
Hardware-/Servertausch noch den alten Host-Key. Ein geaenderter Key darf nicht
ungeprueft ersetzt werden, da dieselbe Meldung auch auf einen MITM-Angriff hinweisen kann.

Fix nach Abgleich des neuen Fingerprints ueber die Hoster-Konsole (Befehle als derselbe
Benutzer wie der Watcher, hier `root`, ausfuehren):
```bash
ssh-keygen -f /root/.ssh/known_hosts -R '<ziel-ip>'
new_host_key="$(mktemp)"
ssh-keyscan -t ed25519 -H '<ziel-ip>' > "$new_host_key"
ssh-keygen -lf "$new_host_key"
# Nur nach erfolgreichem Fingerprint-Abgleich fortfahren:
cat "$new_host_key" >> /root/.ssh/known_hosts
rm -f "$new_host_key"
chmod 600 /root/.ssh/known_hosts
```
Alternativ kann direkt nach dem Entfernen des bestaetigten alten Keys genau ein Lauf
neue, bisher unbekannte Keys nicht-interaktiv aufnehmen. Bereits bekannte, geaenderte
Keys werden dabei weiterhin abgelehnt:
```bash
./scripts/wildcard-distribute-on-change.sh --all \
  --config ./ansible/wildcards/export.yml \
  --state-dir ./data/wildcard-distribution-state \
  --force --accept-new-host-keys
```
`--accept-new-host-keys` setzt OpenSSH `StrictHostKeyChecking=accept-new`; die Option
umgeht daher keinen Konflikt mit einem noch vorhandenen alten Host-Key.

3. `start-limit-hit` bei der Service-Unit
Kontext: alte Trigger-Schleife oder Event-Burst.
Fix: aktuelle `.path`-Unit deployen, `systemctl daemon-reload`, `systemctl reset-failed`, Watcher neu starten.

4. `mesg: ttyname failed: Inappropriate ioctl for device`
Kontext: nicht-interaktive SSH-Shell.
Bewertung: kosmetisch, kein Verteilungsfehler.

## Notfall-Rueckfall
1. Watcher stoppen:
```bash
systemctl disable --now wildcard-distribute-on-change.path
```
2. Manuell verteilen:
```bash
cd /home/andy/infra
./scripts/wildcard-distribute-on-change.sh --all --config ./ansible/wildcards/export.yml --state-dir ./data/wildcard-distribution-state --force
```
