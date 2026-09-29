# Game Day Plan

## Zweck

Resilienz wird nicht behauptet, sondern regelmäßig kaputtgetestet.

## Szenarien

### G1 Internet weg

Client offline. Lesen, Suche, Draft, Outbox prüfen.

### G2 SES Outbound weg

Versand muss queued bleiben. Wiederkehr darf exakt einmal senden.

### G3 Inbound Edge weg

SMTP-Retry/sekundären MX beobachten.

### G4 Core weg

Clients lokal nutzbar; Inbound darf nicht verloren gehen.

### G5 DB Restore

Backup wiederherstellen, Journal replayen, Counts/Hashes vergleichen.

### G6 Attachment Vault weg

Mails bleiben lesbar; Attachment UI zeigt klare Verfügbarkeit.

### G7 Gerät verloren

Device revoke; neue Syncs und Download Tokens verweigern.

### G8 Doppelzustellung

Dieselbe Mail über zwei Inbound-Pfade einspeisen; exakt ein logisches
Message-Objekt erwarten.

### G9 Queue Flood

Massive Inbound-Welle simulieren; Backpressure und Storage Limits
prüfen.

## Messwerte

-   RTO
-   RPO
-   Queue Lag
-   verlorene Messages = 0 Ziel
-   Doppelversand = 0 Ziel
-   manuelle Recovery-Schritte
