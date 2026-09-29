# Amazon SES Mail Flow

## Outbound

``` text
EasyPeasyMail Client
 -> Core/Outbox
 -> SES API or SMTP/TLS
 -> DKIM signing
 -> recipient MX
```

Empfehlung: - verifizierte Domain Identity - Easy DKIM - Custom MAIL
FROM - SPF - DMARC - Bounce/Complaint Events zurück in EasyPeasyMail -
idempotente Outbox

## Wichtige begriffliche Korrektur

SES erzeugt **kein automatisches Whitelisting bei fremden Empfängern**.
SMTP über SES und sauber konfigurierte SPF/DKIM/DMARC-Authentifizierung
stärken Authentizität und Zustellbarkeit, ersetzen aber keine Reputation
und garantieren keinen Posteingang.

## Inbound Option A --- SES Edge

``` text
MX -> SES Receiving
   -> auth/spam/malware verdict
   -> temporary S3
   -> event
   -> EasyPeasyMail ingest
   -> durable own store
```

Vorteil: Der öffentlich exponierte SMTP-Empfang und ein Teil der
Vorprüfung liegen bei AWS.

## Inbound Option B --- eigener SMTP Edge

``` text
MX -> EasyPeasyMail SMTP
   -> RBL/auth/scan
   -> durable queue
   -> Mail Core
```

Vorteil: maximale Unabhängigkeit. Nachteil: deutlich mehr Betriebs-,
Reputation-, Abuse- und Security-Verantwortung.

## Favorit fürs erste belastbare System

**Hybrid:** SES als möglicher Inbound Edge + eigener EasyPeasyMail Core
als Mailbox-/Sync-Wahrheit + SES Outbound. Danach kann ein eigener
zweiter MX-/Inbound-Pfad als Resilienz-Experiment hinzukommen.
