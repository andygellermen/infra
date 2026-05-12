# Easy-Event-Planner – Codex Task Plan

## Repo-Struktur

```text
easy-event-planner/
  cmd/server/main.go
  cmd/migrate/main.go
  internal/config
  internal/db
  internal/tenant
  internal/auth
  internal/event
  internal/registration
  internal/payment
  internal/invitation
  internal/notification
  internal/calendar
  internal/privacy
  internal/snippet
  internal/audit
  web/admin
  web/public
  templates/emails
  deployments/docker
  deployments/ansible
  docs
```

## Regeln für Cody/Codex

Kleine Commits, Tests für kritische Logik, keine Secrets, keine Tokens im Log, Statusübergänge zentral kapseln, keine Kapazitäts-/Payment-Logik im Frontend.

## Pakete

1. Projektbootstrap: Go-Modul, Server, Config, `/healthz`, `/readyz`, `/version`, Dockerfile, Compose.
2. SQLite/Migrationen: DB-Verbindung, Migration Runner, initiales Schema, Indizes.
3. Tenant Basis: Tenant Repo, Slug Lookup, Settings, Seed Command.
4. Magic-Link Auth: Token, Hashing, Verify, Session, Logout, Rate Limit, Audit.
5. Mailer Adapter: Interface, LogMailer, SES/SMTP vorbereitet, EmailJob Worker.
6. Event Series CRUD.
7. Event CRUD mit Publish/Unpublish.
8. Public Event Pages.
9. Kostenfreie Registration mit Magic-Link-Verifizierung.
10. Waitlist mit Position und manuellem Nachrücken.
11. Admin Dashboard und Teilnehmerliste.
12. Ghost/CMS Snippet mit `include.js`, list/cards, Config und Embed-Code.
13. Morgen-E-Mail an Veranstalter.
14. Privacy Retention Jobs mit Dry-Run und Anonymisierung.
15. Event Change/Cancel/Postpone.
16. Calendar ICS für Veranstalterfeed und Teilnehmerdatei.
17. PayPal Basis mit Order und Webhook.
18. Discounts & Invitations.
19. Participant Portal.
20. Certificates.

## Erste Codex-Aufgabe

```text
Bitte implementiere Paket 1: Projektbootstrap.

Ziel:
- Go-Modul easy-event-planner initialisieren
- cmd/server/main.go anlegen
- internal/config vorbereiten
- HTTP Server mit /healthz, /readyz und /version
- Dockerfile und docker-compose.yml
- README mit Startbefehlen

Bitte klein, sauber und erweiterbar umsetzen. Keine Secrets. Tests/Smoke-Test ergänzen.
```

## Prompt-Vorlage

```text
Du arbeitest im Repository easy-event-planner.
Bitte implementiere Paket X aus docs/codex-task-plan.md.
Halte dich an die Dokumente in docs/.
Erstelle kleine, nachvollziehbare Änderungen.
Ergänze Tests für kritische Logik.
Speichere keine Secrets.
Gib am Ende Testbefehle und Zusammenfassung aus.
```
