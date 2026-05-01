# Datei: `docs/07-go-repo-struktur.md`

## 1. Ziel

Die Go-Repo-Struktur soll übersichtlich, modular und Docker-/Ansible-fähig sein. Sie soll die fachlichen Domänen sichtbar machen, ohne sofort zu schwergewichtig zu werden.

## 2. Vorgeschlagene Struktur

```text
easy-erp/
├── cmd/
│   └── easy-erp/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── server.go
│   │   ├── routes.go
│   │   ├── middleware.go
│   │   └── dependencies.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── env.go
│   │
│   ├── auth/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── magic_link.go
│   │   └── sessions.go
│   │
│   ├── settings/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── validator.go
│   │   └── sync.go
│   │
│   ├── customers/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── models.go
│   │
│   ├── catalog/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── sync.go
│   │   ├── selectbox.go
│   │   └── models.go
│   │
│   ├── documents/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── numbering.go
│   │   ├── transitions.go
│   │   ├── totals.go
│   │   ├── snapshots.go
│   │   └── models.go
│   │
│   ├── payments/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── allocation.go
│   │   └── models.go
│   │
│   ├── cancellation/
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── policy_engine.go
│   │   ├── fee_calculator.go
│   │   └── models.go
│   │
│   ├── corrections/
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── corrected_invoice.go
│   │   └── models.go
│   │
│   ├── einvoice/
│   │   ├── service.go
│   │   ├── model.go
│   │   ├── mapper.go
│   │   ├── xrechnung.go
│   │   ├── zugferd.go
│   │   └── validator.go
│   │
│   ├── templates/
│   │   ├── service.go
│   │   ├── google_docs.go
│   │   ├── renderer.go
│   │   └── placeholders.go
│   │
│   ├── files/
│   │   ├── service.go
│   │   ├── repository.go
│   │   ├── pdf.go
│   │   ├── hash.go
│   │   └── storage.go
│   │
│   ├── mail/
│   │   ├── service.go
│   │   ├── sender.go
│   │   ├── templates.go
│   │   └── repository.go
│   │
│   ├── google/
│   │   ├── sheets_client.go
│   │   ├── docs_client.go
│   │   ├── drive_client.go
│   │   └── auth.go
│   │
│   ├── sync/
│   │   ├── scheduler.go
│   │   ├── runner.go
│   │   ├── errors.go
│   │   └── repository.go
│   │
│   ├── audit/
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── models.go
│   │
│   └── platform/
│       ├── db/
│       │   ├── db.go
│       │   ├── tx.go
│       │   └── migrations.go
│       ├── http/
│       │   ├── errors.go
│       │   └── responses.go
│       ├── security/
│       │   ├── csrf.go
│       │   ├── cookies.go
│       │   └── tokens.go
│       └── time/
│           └── clock.go
│
├── web/
│   ├── templates/
│   │   ├── layout.html
│   │   ├── login.html
│   │   ├── customers/
│   │   ├── catalog/
│   │   ├── documents/
│   │   ├── payments/
│   │   └── settings/
│   └── static/
│       ├── app.css
│       └── app.js
│
├── migrations/
│   ├── 0001_init.sql
│   ├── 0002_settings.sql
│   ├── 0003_catalog.sql
│   ├── 0004_documents.sql
│   ├── 0005_payments.sql
│   ├── 0006_cancellation.sql
│   └── 0007_einvoice.sql
│
├── docs/
│   ├── 00-konzept-ueberblick.md
│   ├── 01-domaenenlandkarte.md
│   ├── 02-settings-worksheet.md
│   ├── 03-storno-korrektur-policies.md
│   ├── 04-zahlungs-anzahlungslogik.md
│   ├── 05-e-rechnungsdatenmodell.md
│   ├── 06-sqlite-ddl.md
│   └── 07-go-repo-struktur.md
│
├── deploy/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── traefik.labels.example.yml
│   └── ansible/
│       ├── tasks.yml
│       ├── templates/
│       └── vars.example.yml
│
├── scripts/
│   ├── dev.sh
│   ├── migrate.sh
│   ├── sync-settings.sh
│   ├── sync-catalog.sh
│   └── backup-sqlite.sh
│
├── testdata/
│   ├── settings_sample.csv
│   ├── catalog_sample.csv
│   └── invoice_sample.json
│
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
└── README.md
```

## 3. Paketverantwortlichkeiten

| Package | Verantwortung |
|---|---|
| `auth` | Magic Link, Sessions, Rollenprüfung |
| `settings` | Settings-Sync, Validierung, Zugriff auf aktive Settings |
| `customers` | Kunden, Kontakte, Adressen |
| `catalog` | Kategorien, Hersteller, Produktgruppen, SKUs, Select-Boxen |
| `documents` | Angebot, Bestellung, Rechnung, Status, Nummern, Summen |
| `payments` | Zahlungsanforderungen, Zahlungseingänge, Allocations |
| `cancellation` | Storno-Policies, Fee-Berechnung, Stornoentscheidungen |
| `corrections` | Korrektur-/Stornorechnungen, Gutschriften |
| `einvoice` | Datenmodell, Mapping, XML-Export, Validierung |
| `templates` | Google Docs, Platzhalter, Rendering |
| `files` | PDF/XML-Dateien, Hashing, Ablage |
| `mail` | E-Mail-Versand und Versandhistorie |
| `google` | API-Clients für Sheets, Docs, Drive |
| `sync` | geplanter/manueller Sync, Sync-Logs |
| `audit` | Audit-Log für kritische Vorgänge |

## 4. Service-Schichten

Die Anwendung sollte bewusst serviceorientiert bleiben:

```text
Handler
  ↓
Service
  ↓
Repository
  ↓
SQLite
```

Beispiel: Rechnung finalisieren

```text
POST /documents/{id}/finalize
  ↓
documents.Handler.Finalize
  ↓
documents.Service.FinalizeInvoice
  ↓
- Permission prüfen
- Status prüfen
- Nummer atomar vergeben
- Summen final berechnen
- Snapshots fixieren
- E-Rechnungspflicht bewerten
- Audit schreiben
  ↓
Repository Transaction
```

## 5. Wichtige API-Routen für MVP

```text
GET  /login
POST /login/request-link
GET  /auth/magic
POST /logout

GET  /customers
GET  /customers/new
POST /customers
GET  /customers/{id}
POST /customers/{id}

GET  /catalog/select/categories
GET  /catalog/select/manufacturers?category=...
GET  /catalog/select/groups?category=...&manufacturer=...
GET  /catalog/select/products?group=...

GET  /documents
GET  /documents/new?type=quote
POST /documents
GET  /documents/{id}
POST /documents/{id}/items
POST /documents/{id}/finalize
POST /documents/{id}/send
POST /documents/{id}/convert-to-order
POST /documents/{id}/convert-to-invoice

POST /payments/requests
POST /payments
POST /payments/{id}/allocate

POST /cancellations
POST /cancellations/{id}/approve
POST /cancellations/{id}/complete

POST /settings/sync
POST /catalog/sync

POST /einvoice/{document_id}/generate
POST /einvoice/{document_id}/validate
```

## 6. Deployment-Anforderungen

### ENV-Variablen

```text
APP_ENV=production
APP_BASE_URL=https://erp.example.de
APP_COOKIE_SECRET=...
APP_DB_PATH=/data/easy-erp.sqlite

GOOGLE_CREDENTIALS_FILE=/run/secrets/google_credentials.json
GOOGLE_SETTINGS_SPREADSHEET_ID=...
GOOGLE_CATALOG_SPREADSHEET_ID=...

SMTP_PROFILE=ses_eu
SMTP_HOST=email-smtp.eu-north-1.amazonaws.com
SMTP_PORT=587
SMTP_USER=...
SMTP_PASS=...
SMTP_FROM=rechnung@example.de

SESSION_TTL_HOURS=12
MAGIC_LINK_TTL_MINUTES=15
```

### Docker Volumes

```text
/data              SQLite, lokale Dateien, temporäre Exporte
/run/secrets       Google Credentials, SMTP Secrets
```

### Healthcheck

```text
GET /healthz
GET /readyz
```

`/healthz` prüft App-Prozess.  
`/readyz` prüft DB-Zugriff und optional Google-API-Konfiguration.

## 7. Entwicklungsreihenfolge

| Phase | Inhalt |
|---|---|
| 1 | Repo, Docker, SQLite, Migrationen, Healthcheck |
| 2 | Auth/Magic Link, Rollen, Sessions |
| 3 | Settings-Sync und Validierung |
| 4 | Katalog-Sync und Select-Box-API |
| 5 | Kundenverwaltung |
| 6 | Angebote und Dokumentpositionen |
| 7 | Nummernkreise, Finalisierung, PDF |
| 8 | Angebot → Bestellung → Rechnung |
| 9 | Anzahlungen und Zahlungseingänge |
| 10 | Storno-/Korrekturfluss |
| 11 | E-Rechnungsdatenmodell und XML-Prototyp |
| 12 | Audit, Backups, Admin-Übersichten |

## 8. Vermeidungsstrategien

- Fachlogik nicht in Handlern verstecken
- Google-Clients nicht direkt aus Domänenservices heraus wild nutzen; besser über Sync-/Integration-Pakete
- Nummernvergabe immer in Transaktion
- Geldbeträge nie als Float
- alte Dokumente nie überschreiben
- Select-Boxen aus SQLite bedienen
- Sync-Fehler sichtbar machen
- Secrets nie in Google Sheets speichern
- E-Rechnung früh als Datenmodell vorbereiten

---
