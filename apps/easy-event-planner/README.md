# Easy-Event-Planner

Event-Service fuer kleine und mittelgrosse Veranstaltungen mit mandantenfaehiger Veroeffentlichung, Registrierung, Warteliste und spaeter Payment/Einladungen.

## Bootstrap-Status

Die technischen Grundlagen aus Paket 1 bis Paket 20 aus `docs/codex-task-plan.md` sind umgesetzt:

- Go-Modul und Server-Einstiegspunkt
- Config-Layer mit `EEP_*`-Umgebungsvariablen
- System-Endpunkte: `/healthz`, `/readyz`, `/version`
- SQLite-Verbindung mit PRAGMAs (`foreign_keys`, `WAL`, `busy_timeout`)
- Migration Runner (`cmd/migrate`) mit eingebetteten SQL-Migrationen
- Initiales Schema + Indizes aus `docs/data-model.md`
- Tenant-Basis (`internal/tenant`) mit Repo, Slug-Lookup und Settings-Upsert
- Seed-Command (`cmd/seed`) fuer den ersten Tenant-Test
- Magic-Link-Auth (`internal/auth`) mit Token-Hashing, Verify, Session, Logout, Rate-Limit und Audit-Log
- Mailer-Adapter (`internal/notification`) mit `log`/`smtp`/`ses`-Provider-Setup
- EmailJob Repository + Worker (`cmd/worker`) fuer `email_jobs` Queue-Verarbeitung
- Event-Series CRUD (`internal/event` + Admin-API `/api/v1/admin/event-series`)
- Event-CRUD mit Publish/Unpublish (`internal/event` + Admin-API `/api/v1/admin/events`)
- Public Event Pages (`internal/event` + Public-API `/api/v1/public/{tenantSlug}/...`)
- Public Registration mit Magic-Link-Verifizierung (`internal/registration` + `/api/v1/public/{tenantSlug}/registrations/start|verify`)
- Waitlist-Verwaltung mit Offer/Promote (`internal/registration` + Admin-API `/api/v1/admin/events/{eventId}/waitlist` und `/api/v1/admin/waitlist/{waitlistEntryId}/offer|promote`)
- Admin Dashboard und Teilnehmerliste (`internal/registration` + Admin-API `/api/v1/admin/dashboard`, `/api/v1/admin/events/{eventId}/registrations`, `/api/v1/admin/registrations/{registrationId}`)
- Ghost/CMS-Snippet mit Config-CRUD, Embed-Code, `include.js`, `snippet.css` und Public-Snippet-Events (`internal/snippet`, `/api/v1/admin/snippets`, `/{tenantSlug}/include.js`, `/{tenantSlug}/snippet.css`, `/api/v1/public/{tenantSlug}/snippet/events`)
- Tages-/Morgenuebersicht fuer Veranstalter als automatischer Email-Job (`internal/notification`, Template `organizer_morning_summary`)
- Privacy-Retention mit Tenant-Policies, Dry-Run/Run, Teilnehmeranonymisierung, Magic-Link-/Session-/Email-Job-Cleanup, Audit-Log und Admin-API (`internal/privacy`, `/api/v1/admin/privacy/...`)
- Event-Pflegeaktionen mit Statusuebergaengen (`changed`, `postponed`, `cancelled`, `completed`) und Admin-API-Aktionen (`/publish`, `/unpublish`, `/cancel`, `/postpone`, `/mark-completed`)
- PayPal-Basisfluss mit optionaler Real-API-Integration (`OAuth2`, `create-order`) und Webhook-Signaturpruefung inklusive deduplizierter Event-Verarbeitung (`/api/v1/public/{tenantSlug}/payments/paypal/create-order`, `/api/v1/webhooks/paypal`)
- Discounts & Invitations mit Admin-CRUD, Public-Resolve, Nutzungs-/Zeitfenster-/Scope-Validierung und Registrierungseinloesung inkl. Redemptions (`/api/v1/admin/invitations`, `/api/v1/public/{tenantSlug}/invitations/resolve`)
- Teilnehmer-Portal mit Teilnehmer-Magic-Link, eigener Teilnehmer-Session, Registrierungsliste und Selbst-Storno (`/api/v1/public/{tenantSlug}/participants/portal/request|verify|me|registrations|logout`, `/api/v1/public/{tenantSlug}/participants/portal/registrations/{registrationId}/cancel`)
- Teilnahmezertifikate inkl. Admin-Ausstellung, Teilnehmer-Download und oeffentlicher Verifikation (`/api/v1/admin/registrations/{registrationId}/mark-attended|issue-certificate|certificate`, `/api/v1/public/{tenantSlug}/participants/portal/certificates`, `/api/v1/public/{tenantSlug}/certificates/verify`)
- Dockerfile und `docker-compose.yml`
- Unit-Tests und Smoke-Test-Skript

## Schnellstart lokal

```bash
cd apps/easy-event-planner
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go run ./cmd/migrate
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go run ./cmd/seed
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go run ./cmd/server
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go run ./cmd/worker
```

Standardadresse: `http://localhost:8080`

Systemchecks:

```bash
curl -fsS http://localhost:8080/healthz
curl -fsS http://localhost:8080/readyz
curl -fsS http://localhost:8080/version
```

Auth-Checks (Beispiel):

```bash
curl -fsS -X POST http://localhost:8080/api/v1/auth/magic-link/request \
  -H 'Content-Type: application/json' \
  -d '{"tenant_slug":"demo","email":"owner@example.com","purpose":"organizer_login"}'
```

## Tests

```bash
cd apps/easy-event-planner
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go test ./...
```

## Migrationen

```bash
cd apps/easy-event-planner
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go run ./cmd/migrate
```

## Tenant-Seed

```bash
cd apps/easy-event-planner
GOCACHE=$(pwd)/.gocache GOMODCACHE=$(pwd)/.gomodcache go run ./cmd/seed
```

Wichtige Seed-Variablen:

- `EEP_SEED_TENANT_SLUG` (Default: `demo`)
- `EEP_SEED_TENANT_NAME` (Default: `Demo Tenant`)
- `EEP_SEED_TENANT_PUBLIC_BASE_URL` (Default: `${EEP_BASE_URL}/${EEP_SEED_TENANT_SLUG}`)
- `EEP_SEED_TENANT_TIMEZONE` (Default: `Europe/Berlin`)
- `EEP_SEED_TENANT_LOCALE` (Default: `de-DE`)
- `EEP_SEED_RETENTION_DAYS` (Default: `30`)
- `EEP_SEED_SENDER_EMAIL`
- `EEP_SEED_SENDER_NAME`
- `EEP_SEED_PAYPAL_MODE` (`disabled`, `sandbox`, `live`)
- `EEP_SEED_SETTINGS_JSON`

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

Migration im Container:

```bash
cd apps/easy-event-planner
docker compose run --rm easy-event-planner easy-event-planner-migrate
docker compose run --rm easy-event-planner easy-event-planner-seed
docker compose run --rm easy-event-planner easy-event-planner-worker
```

Wichtige Mail/Worker-Variablen:

- `EEP_MAIL_PROVIDER` (`log`, `smtp`, `ses`)
- `EEP_MAIL_FROM` (z. B. `noreply@events.geller.men`)
- `EEP_MAIL_FROM_NAME`
- `EEP_SES_REGION` (nur vorbereitend, z. B. `eu-north-1`)
- `EEP_SES_SMTP_HOST`
- `EEP_SES_SMTP_PORT` (Default `587`)
- `EEP_SES_SMTP_USER`
- `EEP_SES_SMTP_PASS`
- `EEP_EMAIL_WORKER_POLL_INTERVAL` (Default `3s`)
- `EEP_EMAIL_WORKER_BATCH_SIZE` (Default `10`)

Wichtige PayPal-Haertung-Variablen:

- `EEP_PAYPAL_USE_REAL_API` (`true` aktiviert echte PayPal-API-Aufrufe)
- `EEP_PAYPAL_CLIENT_ID`
- `EEP_PAYPAL_CLIENT_SECRET`
- `EEP_PAYPAL_WEBHOOK_ID`
- `EEP_PAYPAL_SANDBOX_API_BASE_URL` (Default `https://api-m.sandbox.paypal.com`)
- `EEP_PAYPAL_LIVE_API_BASE_URL` (Default `https://api-m.paypal.com`)
- `EEP_PAYPAL_HTTP_TIMEOUT` (Default `15s`)

Wichtige Zertifikats-Variablen:

- `EEP_CERTIFICATE_STORAGE_DIR` (Default `${PWD}/certificates`)
- `EEP_CERTIFICATE_ACCESS_TTL` (Default `30m`)

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

Praktischer End-to-End-Testlauf (Admin -> Public -> Teilnehmerportal -> Zertifikatsverifikation) mit produktionsnahen Tenant-Settings.
