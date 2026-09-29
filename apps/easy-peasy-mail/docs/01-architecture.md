# Architektur

## Grobbild

``` text
Internet / MX
     |
     v
+---------------------+
| Inbound Edge        |
| SMTP Receiver       |
+----------+----------+
           |
     validate / scan
           |
           v
+---------------------+       +----------------------+
| Mail Core (Go)      |------>| Durable Mail Store   |
| parse/dedupe/state  |       | encrypted            |
+----+-----------+----+       +----------------------+
     |           |
     | events    | attachment refs
     v           v
+----------+   +------------------+
| AWS Sync |   | Attachment Store |
| temp bus |   | Infra/Object     |
+----+-----+   +------------------+
     |
 +---+-------------------+
 |           |           |
 v           v           v
Web/PWA   Desktop     Mobile
cache     SQLite      SQLite

Outbound:
Client -> Mail Core -> Amazon SES -> Internet
```

## Wesentliche Änderung gegenüber der ersten Skizze

Wenn EasyPeasyMail eine echte Alternative zu Gmail/Yahoo sein soll,
benötigen wir für **Inbound** einen autoritativen Empfangspfad. Amazon
SES kann optional auch als Inbound Edge dienen, ist aber nicht mit einem
vollständigen IMAP-Mailbox-Server gleichzusetzen. Der EasyPeasyMail Core
übernimmt deshalb Mailbox-Logik, Zustände und Verteilung.

## Services im bestehenden Infra-Stack

-   `easypeasy-api` --- Go API, Auth, Mailbox State, Sync.
-   `easypeasy-inbound` --- SMTP/Inbound Adapter.
-   `easypeasy-worker` --- Parsing, Spam, Newsletter Detection,
    Scheduling.
-   `easypeasy-web` --- React/PWA.
-   `easypeasy-store` --- durable encrypted mail/attachment storage.
-   `easypeasy-db` --- serverseitige Metadaten.
-   Traefik --- HTTPS/API-Routing; **nicht** SMTP-Proxy.
-   Ansible/Docker/Portainer --- Provisionierung und Betrieb.

## Architekturregel

Der Server speichert die autoritative Zustandsinformation; Clients
halten robuste lokale Replikate. AWS verteilt Events/Snapshots und kann
Inbound/Outbound-Funktionen übernehmen, ohne alleinige Wahrheit des
Systems zu werden.
