# Capacity & Quota Model

## Warum

Unbegrenztheit ist kein Feature, sondern ein späterer Ausfall.

## Quotas

Pro Account/Tenant definierbar: - Mailbox logical size - Vault size -
Outbound/day - recipients/message - max inbound message size - max
classic MIME attachment - EasyPeasyDrop object size - Draft attachment
staging - API rate

## Soft vs Hard

Soft Limit: Warnung / Upgrade / Aufräumhinweis.

Hard Limit: nur dort, wo Betrieb oder Abuse-Schutz es verlangt.

## Backpressure

Worker und Ingress müssen Queue-Wachstum kontrollieren. Keine
unbeschränkte Parallelität bei Mailbombing.
