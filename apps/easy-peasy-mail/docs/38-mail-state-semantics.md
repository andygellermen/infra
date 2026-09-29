# Mail State Semantics v1

## Entscheidung

`read/unread` ist standardmäßig **globaler Account-/Mailbox-State**.

Wenn eine Mail auf dem Mac gelesen wird, erscheint sie nach
Synchronisation auch auf Smartphone und Web als gelesen.

``` text
Mac: message.read
       |
       v
Core: unread=false, version=19
       |
       +------> iPhone
       +------> Web
       +------> Tablet
```

## Warum

Der Nutzer verwaltet eine Mailbox, nicht mehrere voneinander unabhängige
Gerätemailboxen.

## Gerätespezifisch bleibt

-   Body lokal vorhanden?
-   Attachment lokal vorhanden?
-   offline gepinnt?
-   lokale Scrollposition optional
-   lokale UI-/Fensterzustände

## Optionale spätere Capability

Ein `seen_on_device` kann für Diagnose/UX existieren, darf aber `unread`
nicht ersetzen.

## Conflict

Server ist Autorität für akzeptierten Mailbox-State. Geräte
synchronisieren Mutationen mit erwarteter Version. Stale Mutationen
werden sauber neu bewertet statt still überschrieben.
