# Go Repository Structure

``` text
easypeasymail/
├── cmd/
│   ├── api/
│   ├── inbound/
│   ├── worker/
│   └── migrate/
├── internal/
│   ├── account/
│   ├── auth/
│   ├── mail/
│   ├── mime/
│   ├── attachment/
│   ├── spam/
│   ├── newsletter/
│   ├── sync/
│   ├── outbox/
│   ├── search/
│   ├── storage/
│   ├── crypto/
│   └── observability/
├── pkg/
│   └── protocol/
├── web/
├── migrations/
├── deploy/
│   ├── docker/
│   └── ansible/
├── docs/
└── tests/
    ├── integration/
    ├── recovery/
    └── fixtures/
```

## Leitlinie

`internal/` bleibt fachlich geschnitten. AWS, SQLite, S3 etc. werden
über Adapter angebunden und dürfen die Mail-Domäne nicht dominieren.

## Interfaces

Beispiele:

``` go
type ObjectStore interface { ... }
type EventBus interface { ... }
type MailSender interface { ... }
type InboundSource interface { ... }
```

Dadurch kann SES/S3 später ersetzt oder ergänzt werden, ohne die
Domänenlogik umzubauen.
