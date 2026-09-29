# Pre-MVP Engineering Gates

## Gate 1 --- Data Truth

-   DDL reviewed
-   invariants reviewed
-   delete/retention semantics fixed
-   draft semantics fixed

## Gate 2 --- Delivery Safety

-   inbound commit boundary test
-   send idempotency test
-   unknown provider result handled
-   poison MIME quarantine

## Gate 3 --- Recovery

-   DB restore proven
-   object restore proven
-   `.eml` reconstruction proven
-   rebootstrap proven

## Gate 4 --- Identity

-   device revoke
-   MFA/passkey path
-   recovery path
-   session expiry

## Gate 5 --- Abuse

-   outbound quota
-   rate limiting
-   bounce/complaint ingestion
-   spam/quarantine baseline

## Gate 6 --- Observability

-   message correlation
-   queue metrics
-   delivery timeline
-   actionable alerts

## Gate 7 --- Client

-   offline read/search
-   offline draft
-   cross-device draft
-   outbox
-   Moodivador state separation

## MVP Go

Erst wenn Gates 1--6 in einer Integrationsumgebung belastbar sind, darf
ein externer Pilot Mail als primäre Adresse verwenden.
