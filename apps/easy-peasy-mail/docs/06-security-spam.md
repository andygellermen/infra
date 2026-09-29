# Security, Spam & Trust Pipeline

## Mehrstufige Inbound-Pipeline

``` text
SMTP connection
 -> connection reputation / rate limits
 -> SPF/DKIM/DMARC result
 -> malware scan
 -> block/allow policy
 -> spam scoring
 -> newsletter/list detection
 -> mailbox
```

## Spam

Nicht nur `Spam/kein Spam`, sondern erklärbare Signale: -
SPF/DKIM/DMARC - SES verdicts, falls SES Inbound verwendet wird -
RBL/DNSBL-Signale wie Spamhaus nur entsprechend deren
Nutzungsbedingungen - lokale/domainweite Allow-/Blocklisten -
Nutzerentscheidung `Spam` als eigenes Signal - Rate/Anomalie-Signale -
Newsletter-/Mailinglisten-Erkennung separat von Spam

## Harter Filter

`REJECT` sollte nur bei Signalen verwendet werden, bei denen wir das
Risiko legitimer Verluste bewusst akzeptieren. Für unsichere Fälle:
`QUARANTINE/SPAM` statt Vernichtung.

## Serverweite Listen

``` text
ALLOW
BLOCK
QUARANTINE
NEWSLETTER
TRUSTED-SENDER
```

Lokale Nutzerregeln dürfen die serverweite Policy ergänzen, aber
sicherheitskritische Malware-Entscheidungen nicht unbemerkt umgehen.

## SES

SES kann beim Empfang SPF/DKIM/DMARC-Ergebnisse sowie
Spam-/Malware-Verdicts liefern. Diese Ergebnisse sind wertvolle
Eingangssignale; die eigentliche EasyPeasyMail-Policy bleibt unsere
Entscheidung.
