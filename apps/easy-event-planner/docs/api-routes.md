# Easy-Event-Planner – API-Routenmodell

## Basis

```text
/api/v1
```

Mandantenauflösung im MVP bevorzugt über Slug:

```text
/api/v1/public/{tenantSlug}/...
/{tenantSlug}/...
```

Admin-Routen benötigen Session. Public-Routen liefern nur veröffentlichte Daten. Webhooks werden provider-spezifisch geprüft.

## System

```http
GET /healthz
GET /readyz
GET /version
```

## Auth

```http
POST /api/v1/auth/magic-link/request
POST /api/v1/auth/magic-link/verify
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

Magic-Link-Anforderung antwortet immer neutral, um Account Enumeration zu vermeiden.

## Admin Tenant

```http
GET   /api/v1/admin/tenant
PATCH /api/v1/admin/tenant
GET   /api/v1/admin/tenant/settings
PATCH /api/v1/admin/tenant/settings
```

## Admin User

```http
GET    /api/v1/admin/users
POST   /api/v1/admin/users
GET    /api/v1/admin/users/{userId}
PATCH  /api/v1/admin/users/{userId}
DELETE /api/v1/admin/users/{userId}
```

## Admin Event Series

```http
GET    /api/v1/admin/event-series
POST   /api/v1/admin/event-series
GET    /api/v1/admin/event-series/{seriesId}
PATCH  /api/v1/admin/event-series/{seriesId}
DELETE /api/v1/admin/event-series/{seriesId}
```

## Admin Events

```http
GET    /api/v1/admin/events
POST   /api/v1/admin/events
GET    /api/v1/admin/events/{eventId}
PATCH  /api/v1/admin/events/{eventId}
DELETE /api/v1/admin/events/{eventId}
POST   /api/v1/admin/events/{eventId}/publish
POST   /api/v1/admin/events/{eventId}/unpublish
POST   /api/v1/admin/events/{eventId}/cancel
POST   /api/v1/admin/events/{eventId}/postpone
POST   /api/v1/admin/events/{eventId}/mark-completed
```

## Admin Dashboard

```http
GET /api/v1/admin/dashboard
```

## Admin Registrations

```http
GET   /api/v1/admin/events/{eventId}/registrations
GET   /api/v1/admin/registrations/{registrationId}
PATCH /api/v1/admin/registrations/{registrationId}
POST  /api/v1/admin/registrations/{registrationId}/cancel
POST  /api/v1/admin/registrations/{registrationId}/mark-attended
POST  /api/v1/admin/registrations/{registrationId}/issue-certificate
GET   /api/v1/admin/registrations/{registrationId}/certificate
POST  /api/v1/admin/events/{eventId}/registrations/manual
```

## Admin Waitlist

```http
GET  /api/v1/admin/events/{eventId}/waitlist
POST /api/v1/admin/waitlist/{waitlistEntryId}/offer
POST /api/v1/admin/waitlist/{waitlistEntryId}/promote
POST /api/v1/admin/waitlist/{waitlistEntryId}/remove
```

## Admin Snippets

```http
GET    /api/v1/admin/snippets
POST   /api/v1/admin/snippets
GET    /api/v1/admin/snippets/{snippetId}
PATCH  /api/v1/admin/snippets/{snippetId}
DELETE /api/v1/admin/snippets/{snippetId}
GET    /api/v1/admin/snippets/{snippetId}/embed-code
```

Embed-Code-Beispiel:

```html
<script src="https://events.geller.men/customerxyz/include.js?config=footer-upcoming" defer></script>
```

## Admin Kalender

```http
GET  /api/v1/admin/calendar/feed
POST /api/v1/admin/calendar/feed/rotate-token
GET  /api/v1/admin/calendar/feed/embed-url
```

## Admin Privacy

```http
GET   /api/v1/admin/privacy/retention-policies
PATCH /api/v1/admin/privacy/retention-policies/{policyId}
POST  /api/v1/admin/privacy/retention-jobs/dry-run
POST  /api/v1/admin/privacy/retention-jobs/run
GET   /api/v1/admin/privacy/retention-jobs
```

## Public Events

```http
GET /api/v1/public/{tenantSlug}/events
GET /api/v1/public/{tenantSlug}/events/{eventSlug}
GET /api/v1/public/{tenantSlug}/series
GET /api/v1/public/{tenantSlug}/series/{seriesSlug}/events
```

Query-Parameter:

```text
limit
include_past
series
mode
from
to
```

## Public Registration

```http
POST /api/v1/public/{tenantSlug}/registrations/start
POST /api/v1/public/{tenantSlug}/registrations/verify
POST /api/v1/public/{tenantSlug}/registrations/{registrationId}/cancel-request
```

Start-Request:

```json
{
  "event_id": "evt_123",
  "name": "Max Mustermann",
  "email": "max@example.de",
  "phone": "+491701234567",
  "participation_type": "onsite",
  "invite_code": "FREUNDE20",
  "invite_amount_cents": 3900,
  "privacy_accepted": true,
  "turnstile_token": "..."
}
```

## Invitations & Discounts

```http
GET   /api/v1/admin/invitations
POST  /api/v1/admin/invitations
GET   /api/v1/admin/invitations/{invitationId}
PATCH /api/v1/admin/invitations/{invitationId}
POST  /api/v1/public/{tenantSlug}/invitations/resolve
```

## Participant Portal

```http
POST /api/v1/public/{tenantSlug}/participants/portal/request
POST /api/v1/public/{tenantSlug}/participants/portal/verify
GET  /api/v1/public/{tenantSlug}/participants/portal/me
GET  /api/v1/public/{tenantSlug}/participants/portal/registrations
POST /api/v1/public/{tenantSlug}/participants/portal/registrations/{registrationId}/cancel
GET  /api/v1/public/{tenantSlug}/participants/portal/certificates
GET  /api/v1/public/{tenantSlug}/participants/portal/certificates/{certificateId}
GET  /api/v1/public/{tenantSlug}/participants/portal/certificates/{certificateId}/download
POST /api/v1/public/{tenantSlug}/participants/portal/logout
```

## Public Certificate Verification

```http
GET /api/v1/public/{tenantSlug}/certificates/verify?certificate_no=...&code=...
```

## PayPal

```http
POST /api/v1/public/{tenantSlug}/payments/paypal/create-order
POST /api/v1/webhooks/paypal
```

Webhook-Pflichten:

- Signatur prüfen
- Payload speichern
- Event-ID deduplizieren
- idempotent verarbeiten
- Payment und Registration aktualisieren
- Audit-Log schreiben

## Snippet

```http
GET /{tenantSlug}/include.js
GET /{tenantSlug}/snippet.css
GET /api/v1/public/{tenantSlug}/snippet/events
```

## Kalender

```http
GET /{tenantSlug}/calendar/admin.ics?token=...
GET /api/v1/public/{tenantSlug}/registrations/{registrationId}/calendar.ics?token=...
```

## Fehlerformat

```json
{
  "error": {
    "code": "EVENT_FULL",
    "message": "Die Veranstaltung ist ausgebucht. Eine Warteliste ist verfügbar.",
    "details": {}
  }
}
```

Wichtige Fehlercodes: `UNAUTHORIZED`, `FORBIDDEN`, `EVENT_NOT_FOUND`, `EVENT_FULL`, `WAITLIST_DISABLED`, `INVALID_MAGIC_LINK`, `EXPIRED_MAGIC_LINK`, `PAYMENT_REQUIRED`, `RATE_LIMITED`, `VALIDATION_ERROR`.
