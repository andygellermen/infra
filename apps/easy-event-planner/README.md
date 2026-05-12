# Easy-Event-Planner

Event-Service fuer kleine und mittelgrosse Veranstaltungen mit mandantenfaehiger Veroeffentlichung, Registrierung, Warteliste und spaeter Payment/Einladungen.

## Bootstrap-Status

Der technische Projektbootstrap (Paket 1 aus `docs/codex-task-plan.md`) ist umgesetzt:

- Go-Modul und Server-Einstiegspunkt
- Config-Layer mit `EEP_*`-Umgebungsvariablen
- System-Endpunkte: `/healthz`, `/readyz`, `/version`
- Dockerfile und `docker-compose.yml`
- Unit-Tests und Smoke-Test-Skript

## Schnellstart lokal

```bash
cd apps/easy-event-planner
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go run ./cmd/server
```

Standardadresse: `http://localhost:8080`

Systemchecks:

```bash
curl -fsS http://localhost:8080/healthz
curl -fsS http://localhost:8080/readyz
curl -fsS http://localhost:8080/version
```

## Tests

```bash
cd apps/easy-event-planner
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go test ./...
```

## Smoke-Test

```bash
cd apps/easy-event-planner
./smoke/http-smoke.sh
```

## Docker

```bash
cd apps/easy-event-planner
docker compose up --build
```

## Dokumente

- [domain-map.md](docs/domain-map.md)
- [mvp-scope.md](docs/mvp-scope.md)
- [data-model.md](docs/data-model.md)
- [api-routes.md](docs/api-routes.md)
- [auth-magic-link.md](docs/auth-magic-link.md)
- [registration-flow.md](docs/registration-flow.md)
- [payment-paypal-flow.md](docs/payment-paypal-flow.md)
- [discount-invitation-model.md](docs/discount-invitation-model.md)
- [privacy-retention.md](docs/privacy-retention.md)
- [ghost-snippet-model.md](docs/ghost-snippet-model.md)
- [admin-ui-structure.md](docs/admin-ui-structure.md)
- [infra-deployment.md](docs/infra-deployment.md)
- [codex-task-plan.md](docs/codex-task-plan.md)

## Naechster technischer Schritt

Paket 2 aus `docs/codex-task-plan.md`: SQLite-Anbindung, Migration Runner und initiales Schema.
