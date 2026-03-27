# Troubleshooting Runbook

## Ziel

Diese Notiz dokumentiert einen typischen Störfall im Infra-Stack:

- alter und neuer Infra-Stack laufen parallel
- ein alter Container blockiert Ports, Volumes oder Datenpfade
- nach Stoppen des alten `traefik` brechen Routing und abhängige Web-Container sichtbar weg

Das Ziel ist eine sichere Wiederherstellung ohne unnötiges Löschen produktiver Daten.


## Typische Symptome

- `infra-traefik` läuft, aber Websites sind nicht erreichbar
- Ghost-Container hängen in `Restarting (2)`
- `infra-mysql` ist nur `Created`
- `infra-portainer` hängt in `Restarting`
- Traefik loggt Meldungen wie:
  - `unable to find the IP address for the container`
  - `middleware "crowdsec-admin@docker" does not exist`
  - `middleware "crowdsec-api@docker" does not exist`
- gleichzeitig existieren noch alte Container wie:
  - `traefik`
  - `portainer`
  - `ghost-mysql`


## Ursache in der Praxis

In unserem konkreten Vorfall waren alte und neue Infra-Komponenten gleichzeitig vorhanden:

- neuer Stack:
  - `infra-traefik`
  - `infra-mysql`
  - `infra-portainer`
- alter Stack:
  - `traefik`
  - `ghost-mysql`
  - `portainer`

Dadurch entstanden zwei typische Kollisionen:

1. `ghost-mysql` belegte `127.0.0.1:3306`, dadurch konnte `infra-mysql` nicht starten.
2. `portainer` hielt die Portainer-DB/den Datenpfad offen, dadurch lief `infra-portainer` in Timeouts.

Folge:

- Ghost versuchte `infra-mysql` aufzulösen und bekam `ENOTFOUND`
- Traefik fand Dienste oder Middlewares nicht stabil
- der alte Traefik war gestoppt, daher fiel die bisher noch funktionierende Routing-Schicht weg


## Erste Diagnose

Diese Befehle zuerst auf dem Server ausführen:

```bash
docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Image}}\t{{.Ports}}'
docker network ls
docker logs --tail 100 infra-traefik
docker inspect infra-traefik --format '{{json .State}}'
```

Wichtige Fragen:

- Gibt es alte und neue Infra-Container parallel?
- Läuft `infra-mysql` wirklich?
- Läuft `crowdsec-bouncer-traefik`?
- Ist `infra-portainer` wirklich oben oder nur in einem Restart-Loop?


## MySQL-Störung erkennen

Prüfen:

```bash
docker logs --tail 100 infra-mysql
docker inspect infra-mysql --format '{{json .State}}'
```

Typischer Fehler:

```text
Bind for 127.0.0.1:3306 failed: port is already allocated
```

Dann blockiert meist ein Altcontainer wie `ghost-mysql` den Port.

Sichere Reihenfolge:

```bash
docker stop ghost-mysql
docker start infra-mysql
docker logs --tail 50 infra-mysql
docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
```

Wichtig:

- alten DB-Container zuerst nur stoppen, nicht sofort löschen
- erst löschen, wenn der neue Stack stabil läuft


## Ghost-Störung erkennen

Prüfen:

```bash
docker logs --tail 100 ghost-andy-geller-men
```

Typischer Fehler:

```text
Error: getaddrinfo ENOTFOUND infra-mysql
```

Das bedeutet in der Regel:

- `infra-mysql` läuft nicht
- oder der Container ist nicht korrekt im `backend`-Netz erreichbar

Nach erfolgreichem MySQL-Fix gegebenenfalls alle Web-Container sauber neu deployen:

```bash
./scripts/redeploy-all-web.sh
```


## CrowdSec/Bouncer-Störung erkennen

Prüfen:

```bash
docker logs --tail 100 infra-traefik
docker ps -a --format 'table {{.Names}}\t{{.Status}}' | grep crowdsec
```

Typische Traefik-Fehler:

- `middleware "crowdsec-admin@docker" does not exist`
- `middleware "crowdsec-api@docker" does not exist`

Dann CrowdSec/Bouncer neu deployen:

```bash
ansible-playbook -i ./ansible/inventory/hosts.ini ./ansible/playbooks/deploy-crowdsec.yml
```

Hinweis:

- In einem Vorfall war die Bouncer-Key-Erzeugung in der Rolle fehlerhaft.
- Der Fix liegt in `ansible/playbooks/roles/crowdsec/tasks/main.yml`.


## Portainer-Störung erkennen

Prüfen:

```bash
docker logs --tail 100 infra-portainer
```

Typischer Fehler:

```text
failed opening store | error=timeout
```

Das spricht oft dafür, dass noch ein alter `portainer` dieselbe Datenbankdatei bzw. dasselbe Volume blockiert.

Sichere Reihenfolge:

```bash
docker stop portainer
docker start infra-portainer
docker logs --tail 50 infra-portainer
```

Auch hier:

- alten Portainer zuerst nur stoppen
- erst bei stabilem neuen Betrieb löschen


## Empfohlene Recovery-Reihenfolge

Wenn alter und neuer Stack parallel vorhanden sind, hat sich diese Reihenfolge bewährt:

1. Zustand erfassen mit `docker ps -a`, `docker logs`, `docker inspect`
2. `infra-mysql` reparieren und hochbringen
3. `infra-traefik` prüfen
4. CrowdSec + Traefik-Bouncer deployen
5. `infra-portainer` stabilisieren
6. Web-Container redeployen
7. erst danach Altcontainer entfernen

Typischer Ablauf:

```bash
./scripts/infra-update-all.sh --portainer-domain=<fqdn>
./scripts/redeploy-all-web.sh
docker ps --format 'table {{.Names}}\t{{.Status}}'
```


## Altcontainer erst am Ende entfernen

Erst wenn alles stabil läuft:

```bash
docker rm traefik
docker rm portainer
docker rm ghost-mysql
```

Vorher nicht löschen, weil:

- der alte Container im Fehlerfall noch als Referenz dient
- sich Volumes, Ports oder Konfigurationspfade sonst schwerer nachvollziehen lassen


## Schnellcheck nach der Wiederherstellung

```bash
docker ps --format 'table {{.Names}}\t{{.Status}}'
docker logs --tail 50 infra-traefik
docker logs --tail 50 infra-portainer
```

Erwartung:

- `infra-mysql` ist `Up`
- `infra-traefik` ist `Up`
- `crowdsec` ist `Up`
- `crowdsec-bouncer-traefik` ist `Up`
- `infra-portainer` ist `Up`
- alle `ghost-*` und `wp-*` Container sind `Up`
- alte Altcontainer sind gestoppt oder entfernt


## Merksätze

- Nicht sofort löschen, erst stabilisieren.
- MySQL zuerst, dann Bouncer/Traefik, dann Portainer, dann Web-Container.
- Alte und neue Infra-Container parallel sind ein Warnsignal.
- `Created` bei MySQL ist fast nie harmlos.
- `ENOTFOUND infra-mysql` bei Ghost zeigt fast immer auf ein DB-/Netzproblem.
