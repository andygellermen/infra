# Sprach-A-Lyzer – geschütztes MVP-Deployment v0.1

**Status:** APPROVED
**Ziel:** interner Smoke-Test des v0.6 MVP Candidate
**Referenzdomain:** `sal.geller.men`

## Sicherheitsmodell

- Traefik veröffentlicht ausschließlich HTTPS und leitet HTTP auf HTTPS um.
- Basic Auth schützt die vollständige Domain einschließlich UI, API, Admin und
  Health-Endpunkten.
- Traefik entfernt den `Authorization`-Header vor der Weitergabe.
- Der API-Container veröffentlicht keinen Host-Port.
- PostgreSQL besitzt keinen Host-Port und liegt in einem internen Docker-Netz.
- Das Klartext-Zugriffskennwort wird weder committed noch im Deployment
  gespeichert.
- Die ignorierten Hostvars enthalten einen bcrypt-Hash und ein zufälliges
  256-Bit-Datenbankkennwort; die gerenderte Runtime-Umgebung hat Modus `0600`.

## Einmalige Vorbereitung auf dem Infra-Server

Die Domain muss bereits per A-Record auf den Server zeigen. Danach:

```bash
cd /home/andy/infra
git pull --ff-only

sudo apt-get update
sudo apt-get install -y apache2-utils

./scripts/sal-add.sh sal.geller.men --username=andy
```

`sal-add.sh` prüft DNS, fragt das Zugriffskennwort verdeckt zweimal ab, erzeugt
das PostgreSQL-Kennwort und schreibt ausschließlich die durch Git ignorierte
Datei `ansible/hostvars/sal.geller.men.yml` mit Modus `0600`.

## Deployment

```bash
./scripts/sal-redeploy.sh sal.geller.men
```

Der Ablauf ist Build → PostgreSQL → Migration → idempotenter Foundation Seed →
API → unauthentifizierter 401-Test → authentifizierter Produkt-/Readiness-Test.
Für den letzten Test wird das Kennwort verdeckt abgefragt und nicht gespeichert.

Danach stehen bereit:

- `https://sal.geller.men/`
- `https://sal.geller.men/admin`
- `https://sal.geller.men/health/ready`

## Betrieb

```bash
docker ps --filter 'name=sprach-a-lyzer-sal-geller-men'
docker logs --tail 100 sprach-a-lyzer-sal-geller-men
docker logs --tail 100 sprach-a-lyzer-sal-geller-men-postgres

./scripts/sal-smoke-check.sh sal.geller.men --username=andy
./scripts/sal-redeploy.sh sal.geller.men --check-only --skip-auth-smoke
```

Ein gewöhnlicher Redeploy bewahrt das benannte PostgreSQL-Volume. Der MVP-Pfad
speichert keine Analyse- oder Sitzungstexte; der Datenbankinhalt besteht im
Smoke-Test aus reproduzierbaren Migrationen und Seed-Daten. Eine konsistente
`pg_dump`-/Restore-Integration wird vor dem ersten redaktionellen
Managed-Knowledge-Einsatz ergänzt und ist nicht Teil dieses Smoke-Deployments.

## Kennwortwechsel

```bash
./scripts/sal-rotate-secrets.sh sal.geller.men
./scripts/sal-redeploy.sh sal.geller.men
```

Das Rotationsskript fragt das neue Zugriffskennwort verdeckt zweimal ab,
wechselt zusätzlich das PostgreSQL-Rollenkennwort und aktualisiert die
ignorierten Hostvars atomar. Falls deren Aktualisierung scheitert, setzt es das
bisherige Datenbankkennwort zurück. Es gibt keinen Geheimwert auf der Konsole
aus. Der unmittelbar folgende Redeploy synchronisiert die Container mit den
neuen Werten. Das Klartext-Zugriffskennwort gehört niemals in Git oder die
Hostvars.

Wurde ein Geheimwert versehentlich in einem Konsolenprotokoll offengelegt, ist
dieselbe vollständige Rotation vor dem nächsten Redeploy verpflichtend.
