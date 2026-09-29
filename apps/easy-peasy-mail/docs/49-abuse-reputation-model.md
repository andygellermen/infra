# Abuse & Reputation Model

## Zwei Richtungen

### Inbound Abuse

Spam, Phishing, Malware, Mailbombing.

### Outbound Abuse

kompromittierte Accounts, automatisierter Spam, Phishing, ungewöhnliche
Versandspitzen.

## Outbound Controls

-   Account-/Tenant-Quotas
-   Rate Limits
-   Burst Limits
-   Recipient Limits
-   neue Accounts konservativer
-   Bounce-/Complaint-Signale
-   ungewöhnliche Versandmuster
-   temporäres Send Hold bei starken Anomalien

## Reputation

Provider-/Domain-Reputation ist Betriebsvermögen.

Metriken: - bounce rate - complaint rate - reject rate - blocklist
observations - DKIM/SPF/DMARC health - outbound volume trend

## Policy

Automatisierung darf begrenzen/quarantänisieren. Dauerhafte Sperren mit
erheblicher Nutzerwirkung brauchen nachvollziehbaren Grund und
Review-Pfad.
