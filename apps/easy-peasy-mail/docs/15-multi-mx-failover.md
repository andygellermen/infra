# Multi-MX & Inbound Failover

## SMTP-Eigenschaft nutzen

Mailzustellung ist bereits store-and-forward. Wenn der primäre MX
temporär nicht erreichbar ist, versuchen sendende Mailserver
typischerweise erneut; ein sekundärer MX kann zusätzliche Resilienz
schaffen.

## Stufe 1

``` text
MX 10 -> SES Inbound Edge
          |
          v
      EasyPeasy Core
```

## Stufe 2 --- unabhängiger zweiter Pfad

``` text
MX 10 -> Edge A / SES --------+
                              +-> dedupe -> Core
MX 20 -> Edge B / own SMTP ---+
```

## Kritischer Punkt

Ein sekundärer MX darf **kein schwächer geschütztes Hintertor** sein.
Beide Pfade benötigen vergleichbare Anti-Abuse-/Auth-/Rate-Limit-Regeln.

## Deduplizierung

Nicht allein `Message-ID` vertrauen. Kombination aus: -
transport/inbound id - canonical content hash - recipient/account -
Zeitfenster - MIME hash

## Failure Drills

-   Edge A 6 Stunden weg.
-   Core 24 Stunden weg.
-   DNS/MX-Änderung fehlerhaft.
-   Queue wächst stark.
-   Nachricht doppelt über beide Edges.
-   S3/Event-Zustellung verzögert.
