# Easy-Event-Planner – Technische Domänenlandkarte

## Zielbild

Der Easy-Event-Planner ist ein eigenständiger, mandantenfähiger Event-Service für Ghost-CMS und perspektivisch WordPress sowie weitere CMS-Systeme. Die CMS-Integration erfolgt über kopierbare Snippets und Links aus dem Admin-Interface. Ghost bleibt Veröffentlichungs- und Content-Ort; Anmeldung, Zahlung, Warteliste, Datenschutz, Kalender und Veranstalterprozesse laufen serverseitig im Easy-Event-Planner.

```text
CMS-Seite / Ghost-Content
  -> JavaScript-Snippet oder Event-Link
  -> Public Event UI / Snippet Renderer
  -> Easy-Event-Planner API
  -> Domain Services
  -> SQLite im MVP, später optional PostgreSQL
  -> SES, PayPal, ICS, PDF, Jobs
```

## Bounded Contexts

### 1. Tenant Management
Verwaltet Mandanten, Branding, Domains, Standardzeitzone, E-Mail-Absender, PayPal-Konfiguration, Datenschutzfristen und Snippet-Grundeinstellungen.

Entitäten: `Tenant`, `TenantSettings`, `TenantUser`, `TenantBranding`.

### 2. Auth & Magic Links
Passwortloser Login für Veranstalter und später Teilnehmer. Magic Links sind kurzlebig, einmalig verwendbar und werden nur gehasht gespeichert.

Entitäten: `MagicLink`, `Session`, `TenantUser`, `ParticipantIdentity`.

### 3. Event Management
Verwaltet Eventreihen und einzelne Termine inklusive Status, Sichtbarkeit, Teilnahmeart, Ort, Online-Link, Kapazität und Änderungs-/Absage-Informationen.

Entitäten: `EventSeries`, `Event`, `EventTicket`.

Eventstatus: `draft`, `scheduled`, `changed`, `postponed`, `cancelled`, `completed`, `archived`.

### 4. Registration & Waitlist
Steuert Anmeldung, Verifizierung, Platzprüfung, Warteliste, Stornierung und Teilnahme-Status.

Entitäten: `Participant`, `Registration`, `WaitlistEntry`.

Registrierungsstatus: `verification_pending`, `reserved`, `payment_pending`, `confirmed`, `waitlist`, `cancelled`, `expired`, `attended`, `no_show`.

### 5. Payment
Kapselt PayPal-Checkout, Payment-Status, Webhook-Verarbeitung, Reservierungsablauf und spätere Rückerstattungen. Zahlungswahrheit kommt serverseitig aus PayPal-Webhooks oder serverseitiger Prüfung, niemals nur aus dem Browser.

Entitäten: `Payment`, `PayPalWebhookEvent`, `Refund`.

### 6. Invitation, Discount, Sponsoring & Donation
Ermöglicht Einladungslinks, Rabattcodes, Gutscheinlinks, vollständig gesponserte Teilnahme, teilbare Rabattlinks und spendenbasierte Teilnahme.

Entitäten: `InvitationLink`, `DiscountRedemption`, `DonationConfig`.

### 7. Notification
Versendet Magic Links, Anmeldebestätigungen, Wartelistenmails, Zahlungsbestätigungen, Änderungs-/Absagemails und Morgen-E-Mails an Veranstalter.

Entitäten: `EmailTemplate`, `EmailJob`, `EmailLog`.

### 8. Calendar
Stellt Veranstalter-Kalenderfeeds per `.ics`-URL bereit und erzeugt Teilnehmer-Kalenderdateien mit Bestätigungsmails. Veranstalter-Einträge enthalten Admin-Direktlinks, Teilnehmer-Einträge niemals.

Entitäten: `CalendarFeed`, `CalendarToken`.

### 9. Snippet & CMS Embed
Erzeugt kopierbare Ghost-/CMS-Snippets und rendert öffentliche Eventlisten, Cards, Tabellen oder Button-Widgets.

Entitäten: `SnippetConfig`, `PublicRenderConfig`.

### 10. Privacy & Retention
Konfiguriert Lösch-/Anonymisierungsfristen je Mandant und Datenkategorie. Unterstützt Dry-Run, Audit-Log und datensparsame Defaults.

Entitäten: `RetentionPolicy`, `DataRetentionJob`.

### 11. Participant Portal & Certificates
Spätere Ausbaustufe: Teilnehmer sehen aktuelle und vergangene Veranstaltungen, laden Bescheinigungen/Zertifikate herunter und verwalten Stornierungen per Magic Link.

Entitäten: `Certificate`, `CertificateTemplate`, `ParticipantSession`.

### 12. Audit & Security
Protokolliert sicherheits- und fachlich relevante Aktionen: Login, Eventänderungen, Anmeldungen, Zahlungen, Kalenderfeed-Rotation, Löschjobs.

Entitäten: `AuditLog`, `RateLimit`, `SecurityEvent`.

## Zentrale Regeln

1. Kein CMS verändert Daten direkt.
2. Alle Schreiboperationen laufen über die API.
3. Snippets liefern keine sensiblen Daten aus.
4. Kapazitätsentscheidungen erfolgen transaktional serverseitig.
5. Magic Links werden nie im Klartext gespeichert.
6. PayPal-Webhooks werden idempotent verarbeitet.
7. Veranstaltungsänderungen erzeugen Benachrichtigung und optional Kalenderupdate.
8. Personenbezogene Daten werden nach Frist anonymisiert oder gelöscht.
9. Jede kritische Aktion wird auditiert.

## Empfohlene Reihenfolge

```text
1 Tenant + Magic-Link Auth
2 EventSeries + Events
3 Public Event Pages
4 Registration + Waitlist
5 Ghost/CMS Snippet
6 Organizer Morning Summary
7 Privacy Retention Jobs
8 Event Change/Cancel/Postpone
9 Calendar ICS
10 PayPal
11 Discounts/Invitations/Sponsoring
12 Participant Portal
13 Certificates
```
