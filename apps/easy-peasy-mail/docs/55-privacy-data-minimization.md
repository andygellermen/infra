# Privacy & Data Minimization

## Prinzip

Beobachtbarkeit ohne unnötige Inhaltsprotokollierung.

## Logs

Nicht standardmäßig loggen: - Body - vollständige Attachments - Auth
Tokens - Passwörter - Capability Tokens

## Metriken

Wo möglich aggregieren: - Counts - Latenzen - Queue Age - Fehlerklassen

## Audit

Sicherheitsrelevante Aktionen protokollieren, z. B.: - Login - Device
add/revoke - Recovery change - forwarding rule - export - domain/account
administration

## Retention

Logs, Audit, Mailcontent, Vault Objects und Backups haben getrennte
Retention Policies.
