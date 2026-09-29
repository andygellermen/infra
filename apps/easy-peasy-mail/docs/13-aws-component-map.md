# AWS Component Map

## Ziel

AWS soll EasyPeasyMail hochverfügbar unterstützen, ohne zur unnötigen
dauerhaften Mailbox zu werden.

## Empfohlene Rollen

  -----------------------------------------------------------------------------
  Aufgabe                 AWS-Baustein            Rolle
  ----------------------- ----------------------- -----------------------------
  Outbound                Amazon SES              Versand, DKIM,
                                                  Bounce/Complaint Events

  Inbound Edge            Amazon SES Receiving    MX-Empfang,
                                                  Auth-/Spam-/Malware-Signale

  temporäres MIME         S3                      kurzlebiger Rohmail-Puffer

  Event-Verteilung        SQS/SNS bzw.            Entkopplung zwischen Ingest
                          EventBridge             und Core

  Verschlüsselung         KMS                     Schlüssel für AWS-seitige
                                                  Objekte/Secrets

  Secrets                 Secrets                 Credentials/Keys
                          Manager/Parameter Store 

  Observability           CloudWatch              technische Metriken/Alarme
  -----------------------------------------------------------------------------

## Favorisierter Inbound-Fluss

``` text
MX
 -> SES Receiving
 -> S3 TEMP MIME
 -> Event
 -> EasyPeasy Ingest Worker
 -> validate idempotency
 -> durable EasyPeasy store
 -> ACK/checkpoint
 -> S3 lifecycle expiration
```

SES kann Rohmail nach S3 zustellen; die Anwendung übernimmt daraus
Parsing und Mailbox-Logik. Der temporäre Bucket erhält Lifecycle-Regeln.
AWS bleibt damit Edge/Relay, nicht Produktwahrheit.

## Alternative

Ein eigener SMTP-Edge kann später parallel als zweiter Pfad erprobt
werden. Er erhöht Unabhängigkeit, aber auch Betriebs- und
Security-Verantwortung.

## Noch zu entscheiden

-   SQS vs. EventBridge für interne Events.
-   Region(en) und zulässige Inbound-Regionen.
-   Retention je Objektklasse.
-   KMS-Key-Modell.
