# EasyTrm – Entwicklungsplan v0.1 für Cody

> Status: fachliche Baseline / Entwicklungsgrundlage  
> Produkt: EasyTrm  
> Domain: `easytrm.de` (Arbeitsname)  
> Repository-Ziel: `apps/easy-trm/`  
> Leitidee: **EasyTrm organisiert Termine – nicht Therapien.**

## 1. Zweck

Dieses Dokument übersetzt die EasyTrm Product Requirements v0.1 in einen umsetzbaren Entwicklungsplan für Cody. Es ist die fachliche Referenz für MVP-Scope, Architekturleitplanken, Domain-Modell, Scheduling, Kalender-Synchronisation, Terminstatus, Rollen, Kommunikation, Ressourcen, Sicherheit, Tests und geplante Erweiterungen.

Scope-Test für neue Features:

> **Verbessert diese Funktion die Terminorganisation oder beginnt damit eine andere Software?**

## 2. Produktvision

EasyTrm ist eine einfache, ruhige und flexibel konfigurierbare Terminplanungsplattform für kleine nichtärztliche Praxen.

Zielgruppen:
- Physiotherapeuten
- Osteopathen
- Heilpraktiker
- andere nichtärztliche Behandler
- Einzelpraxen
- kleine Gemeinschaftspraxen

EasyTrm ist kein Patienten-Marktplatz, keine vollständige Praxissoftware, keine medizinische Dokumentation und keine Abrechnungssoftware.

Kern: **Scheduling first. Everything else later.**

Produktversprechen: **Weniger Terminorganisation. Mehr Zeit für Menschen.**

## 3. Produktprinzipien

1. **Scheduling first** – Terminfluss ist Kern.
2. **Therapist-controlled workflow** – wesentliche Terminregeln sind je Therapeut konfigurierbar.
3. **Passwordless** – Klientenzugänge ohne Passwort.
4. **Account optional** – Buchung ohne dauerhaftes Konto.
5. **Progressive Disclosure** – einfache Defaults, Spezialoptionen bei Bedarf.
6. **Constraints before optimization** – harte Regeln vor Optimierung.
7. **External calendars remain external** – externe Kalender werden berücksichtigt, nicht ersetzt.
8. **Data minimization** – keine medizinischen Behandlungsdaten im MVP.
9. **Human language before system language** – verständliche Alltagssprache.
10. **Expansion through composition** – neue Features möglichst aus vorhandenen Kernmechanismen zusammensetzen.

## 4. MVP-Scope

Muss in v0.1 enthalten sein:
- Practice / Mandant
- mehrere Therapeuten
- Rezeption / Assistenz
- mehrere Terminarten
- individuelle Arbeitszeiten
- feste und flexible Pausen
- Scheduling-Modi
- Räume / Ressourcen
- Onlinebuchung
- „nächstmöglicher Termin“
- Therapeutenauswahl
- Telefonberatung
- Rückruf als Termin oder Anfrage
- externe Kalender je Therapeut
- mehrere Kalender je Therapeut
- Privatkalender als Read-Busy-Quelle
- Betrieb ohne externen Kalender
- Kalender-Ausfallstrategie
- Slot-Hold
- Klientenbestätigung
- konfigurierbare Bestätigungsfrist
- Praxis-Termine optional sofort bestätigt
- Umbuchung
- Stornierung
- E-Mail/SMS
- Reminder
- organisatorische Hinweise
- optionaler passwortloser Klientenaccount
- Terminvorschläge
- Audit Trail
- responsive Web-App / PWA

## 5. Explicit Non-Goals

Nicht implementieren:
- Diagnosen
- Behandlungsakte
- medizinische Verlaufsdokumentation
- Rezepte
- Krankenkassenabrechnung
- Versicherungsverwaltung
- Rechnungswesen
- Zahlungsabwicklung
- Dokumentenmanagement
- Telemedizin / Video
- Marketingautomation
- Patienten-Marktplatz
- Bewertungen
- medizinische KI
- komplexes CRM
- native iOS-/Android-App

Diese Punkte sind bewusst ausgeschlossen, nicht vergessen.

## 6. Mandantenmodell

```text
Practice
├── Location(s)
├── Practitioner(s)
├── Staff / Assistant(s)
├── Service(s)
├── Resource(s)
├── Client(s)
└── Appointment(s)
```

Jede mandantenabhängige Entität besitzt eine `practice_id`.

Sicherheitsgrundsatz: **Kein fachlicher Query ohne Tenant Scope.**

## 7. Rollen

### OWNER / ADMIN
Praxis, Therapeuten, Staff, Services, Ressourcen, Rollen und SaaS-Einstellungen verwalten.

### PRACTITIONER
Eigene Termine, Arbeitszeiten, Kalender, Scheduling-Präferenzen, Vorschläge, Rückrufe und organisatorische Hinweise verwalten.

### ASSISTANT
Von v0.1 an vorgesehen. Termine sehen/anlegen/verschieben/stornieren, Klienten kontaktieren, Rückrufe und Vorschläge verwalten. Zugriff praxisweit oder auf ausgewählte Therapeuten begrenzbar. Keine geschützten Credential-/Rollen-/SaaS-Funktionen per Default.

### CLIENT
Nur eigene Termine und eigene Kommunikations-/Kontodaten.

## 8. Therapeuten-Setup

```text
Practitioner
├── Services
├── Availability Rules
├── Break Rules
├── Calendar Connections
├── Scheduling Preference
├── Resource Setup
├── Reminder Rules
├── Confirmation Rules
└── Callback Mode
```

Therapeuten derselben Praxis dürfen unterschiedlich arbeiten.

## 9. Services

Mindestens:
```text
name
duration_minutes
buffer_before_minutes
buffer_after_minutes
booking_notice_minutes
confirmation_required
confirmation_timeout_minutes
cancellation_deadline_minutes
reschedule_deadline_minutes
delivery_mode
active
```

Beispiele: Physiotherapie 30 Min., Osteopathie 60 Min., Ersttermin 75 Min., Telefonberatung 20 Min., Rückruf 10 Min.

`delivery_mode`: `ONSITE`, `PHONE`. Video nur später vorbereiten.

## 10. Telefonberatung

Normaler Service mit:
```text
call_direction:
- PRACTICE_CALLS_CLIENT
- CLIENT_CALLS_PRACTICE
```

Telefonnummer muss verifiziert sein.

## 11. Rückruf

Je Therapeut:
```text
callback_mode:
- SCHEDULED
- REQUEST
```

`SCHEDULED` nutzt Appointment Engine.  
`REQUEST` erzeugt `CallbackRequest` mit Name, Mobilnummer und optionalem Zeitfenster.

## 12. Kalenderintegration

0..n Kalender je Practitioner.

```text
mode:
- BOOKING_CALENDAR
- BLOCKING_CALENDAR

access:
- READ_BUSY
- READ_WRITE
```

Provider-Abstraktion:
```text
CalendarProvider
├── GoogleAdapter
├── CalDAVAdapter
└── future adapters
```

Scheduling kennt nur normalisierte Busy-Intervalle.

## 13. Privacy externer Kalender

Für Blockierkalender möglichst nur:
```text
busy_from
busy_until
calendar_connection_id
```

Keine privaten Titel/Orte/Beschreibungen speichern, wenn sie für Scheduling nicht nötig sind.

## 14. Calendar Sync

Keine Live-Abfrage aller Provider in der Availability-Webrequest.

```text
External Calendar
      ↓
Calendar Sync Worker
      ↓
Local Busy Cache
      ↓
Availability Engine
```

Vor finaler Buchung optional zusätzlicher Konfliktcheck.

## 15. Kalenderstatus

```text
CONNECTED
SYNCING
ERROR
AUTH_REQUIRED
DISABLED
```

Zusätzlich:
```text
last_successful_sync
last_attempt_at
failure_reason_code
```

## 16. Kalender-Ausfall

```text
failure_policy:
- USE_LAST_SYNC
- BLOCK_BOOKINGS

required_for_booking: bool
```

`USE_LAST_SYNC`: weiterbuchen mit letztem Busy-Stand.  
`BLOCK_BOOKINGS`: Onlinebuchungen pausieren.  
Betrieb ohne externen Kalender ist voll unterstützt.

## 17. Availability

Grundregel:
```text
AvailabilityRule
- weekday
- start_time
- end_time
```

Ausnahmen:
```text
AvailabilityOverride
- date
- start_time
- end_time
- type
```

Urlaub, freie Tage, Sonderzeiten, Zusatzarbeitstage.

## 18. Pausen

**Fixed Break:** definierter nicht buchbarer Zeitraum.  
**Recovery Break:** z. B. max. 180 Minuten Behandlung am Stück, danach mindestens 30 Minuten Pause.

Recovery Break ist Hard Constraint.

## 19. Scheduling-Modi

```text
CHRONOLOGICAL
COMPACT
BALANCED
```

- CHRONOLOGICAL: zeitlich sortieren
- COMPACT: vorhandene Termine sinnvoll verdichten
- BALANCED: Effizienz + gewünschte Erholung

Empfohlener Default: `BALANCED`.

## 20. Hard Constraints

Nie verletzen:
- Arbeitszeit
- Overrides
- feste Pause
- Recovery Break
- bestehende Termine
- externe Busy-Zeiten
- Dauer
- Buffer
- Booking Notice
- Resource
- Holds
- Transaktionssperren

## 21. Soft Preferences

Nur bewerten/sortieren:
- kompakte Belegung
- chronologische Darstellung
- bevorzugte Tageszeit
- geringe Lücken
- Client Availability Preferences
- Scheduling-Strategie

## 22. Availability Pipeline

```text
1. Service bestimmen
2. Practitioner bestimmen
3. Arbeitszeiten laden
4. Overrides anwenden
5. Fixed Breaks abziehen
6. externe Busy-Zeiten abziehen
7. bestehende Appointments abziehen
8. Holds abziehen
9. Resources prüfen
10. Buffer anwenden
11. Recovery Rules prüfen
12. Booking Notice prüfen
13. Kandidatenslots erzeugen
14. Hard Constraints
15. Soft Preferences
16. Sortierung
17. begrenzte Ergebnisliste
```

## 23. Ressourcen / Räume

Intern generisch `Resource`, UI zunächst „Raum“.

```text
resource_mode:
- NONE
- DEDICATED
- SHARED
- AUTO_ASSIGN
```

Ressourcen sind Hard Constraints. Auto-Assign erfolgt atomar mit Buchung.

## 24. Öffentliche Buchungs-UX

Startfrage:

> **Was können wir für dich tun?**

- Behandlung buchen
- Telefonberatung
- Rückruf
- Termin ändern

Keine Kalenderwand als Startseite.

## 25. Terminfindung

Nach Service:

> **Wie möchtest du deinen Termin finden?**

- **Nächstmöglicher Termin** – praxisweit über geeignete Therapeuten
- **Therapeut auswählen** – konkrete Person

## 26. Slot-Darstellung

Zuerst kleine geeignete Auswahl, danach „Weitere freie Zeiten“.

Später Filter: Vormittag / Nachmittag / Abend.

## 27. Client Identity

Nach Slot-Auswahl.

```text
MOBILE_ONLY
MOBILE_EMAIL
```

Immer passwortlos.

```text
ClientIdentity
- type: PHONE | EMAIL
- value_normalized
- verified_at
- is_primary
```

States: `UNVERIFIED` → `VERIFIED`.

## 28. Optionaler Client Account

Keine Accountpflicht vor Buchung.

Account/PWA:
- nächste Termine
- vergangene Termine
- Terminvorschläge
- Reminder
- Kontaktdaten

Keine medizinischen Daten.

## 29. Appointment Hold

```text
AVAILABLE → HELD
```

Kurze technische Sperre, z. B. 5 Minuten. Timeout gibt Slot frei.

Finale Exklusivität durch Datenbank/Transaktion, nicht nur Applikations-Check.

## 30. Appointment Lifecycle

```text
HELD
PENDING_CLIENT_CONFIRMATION
BOOKED
CANCELLED
RESCHEDULED
COMPLETED
NO_SHOW
EXPIRED
```

Metadaten separat:
```text
created_by
source
confirmed_at
cancelled_at
completed_at
rescheduled_from_appointment_id
```

## 31. Online-Buchung

```text
AVAILABLE
↓
HELD
↓
Kontaktdaten
↓
Identity Verification
↓
PENDING_CLIENT_CONFIRMATION
↓
Bestätigung innerhalb X
↓
BOOKED
```

Nicht rechtzeitig bestätigt:
```text
PENDING_CLIENT_CONFIRMATION → EXPIRED → Slot frei
```

## 32. Bestätigungsfrist

`confirmation_timeout_minutes` je Workflow/Service/Therapeut.

Beispiele: 30 Min., 2h, 6h, 24h, 2 Tage.

Optionaler Expiry-Warning-Reminder.

## 33. Praxis legt Termin an

Wahl:
- direkt bestätigt → `BOOKED`
- Klientenbestätigung erforderlich → `PENDING_CLIENT_CONFIRMATION`

Default konfigurierbar.

## 34. Terminverwaltung

Sicherer Link:
```text
easytrm.de/t/<secure-token>
```

Keine sequenziellen IDs.

Funktionen:
- anzeigen
- verschieben
- stornieren
- Reminder verwalten

Tokens kryptografisch sicher, hashbar, widerrufbar. Sensible Änderungen ggf. erneute OTP-Prüfung.

## 35. Umbuchung

Bestehenden Termin erst nach erfolgreichem neuen Termin freigeben.

```text
BOOKED old
↓
new slot lookup
↓
new HOLD
↓
confirm
↓
new BOOKED
↓
old RESCHEDULED
↓
old slot released
```

Abbruch lässt alten Termin unverändert.

## 36. Stornierung

Je Service:
```text
cancellation_deadline_minutes
reschedule_deadline_minutes
```

Nach Frist Self-Service blockieren und Praxis-Kontakt anbieten. Keine automatische Stornierung.

## 37. Reminder

Default: 24h vorher.

```text
ReminderRule
- offset_minutes
- channel
- enabled
- client_can_disable
```

Kanäle v0.1: EMAIL, SMS. Default `client_can_disable = true`.

## 38. Confirmation vs Reminder

Strikt trennen:
- Reservation / Confirmation Request
- Booking Confirmation
- Attendance Reminder
- optionale spätere Attendance Confirmation

Ein `BOOKED` Termin wird nicht automatisch storniert, nur weil eine spätere Attendance Confirmation ausbleibt.

## 39. Nachrichtentexte

Neutral und kurz. Beispiel:

> Praxis Sonnenseite: Ihr Termin am 1. April um 10:00 Uhr ist bestätigt. Termin ändern: easytrm.de/t/Qrg7

Keine Diagnosen/Beschwerden in SMS.

## 40. Organisatorische Hinweise

`ClientInstruction`: z. B. Handtuch, früher erscheinen, Eingang.  
`InternalOrganizationalNote`: z. B. Rückruf, telefonisch vereinbart, barrierefreier Zugang.

UI bewusst nicht „Behandlungsnotiz“.

## 41. Terminvorschläge

```text
AppointmentProposal
- practice_id
- client_id
- practitioner_id
- service_id
- expires_at
- status
```

States:
```text
CREATED
SENT
ACCEPTED
DECLINED
EXPIRED
```

Optionen blockieren Slots nicht. Bei Annahme: Availability re-check → HOLD → BOOKED.

## 42. Proposal-Modi

- Quick Proposal: System erzeugt n Kandidaten.
- Precise Proposal: Therapeut definiert Anzahl, Zeitraum, gleicher/verschiedene Tage, Tageszeit.

## 43. Follow-up Proposal

„Folgetermin anbieten“ nutzt bekannten Client, Practitioner und Service; Therapeut ergänzt z. B. „nächste Woche“ oder „in 14 Tagen“.

## 44. Dashboard

Bereiche:
- Heute
- Aufmerksamkeit
- Neu
- Terminvorschläge
- Morgen

Aufmerksamkeit z. B. unbestätigte Reservierungen, Sync-Fehler, Rückrufe.

Kein KPI-Cockpit im MVP.

## 45. Client View

Minimal:
- Name
- Mobilnummer
- E-Mail
- nächster Termin
- vergangene Termine
- organisatorische Hinweise
- Reminder-Präferenzen

Keine medizinischen Daten.

## 46. Notification Engine

Zentral, keine Direktversendung aus Fachmodulen.

Typen:
```text
BOOKING_CONFIRMATION
CLIENT_CONFIRMATION_REQUEST
RESERVATION_EXPIRY_WARNING
REMINDER
RESCHEDULE_CONFIRMATION
CANCELLATION_CONFIRMATION
APPOINTMENT_PROPOSAL
PRACTICE_NOTIFICATION
CALLBACK_NOTIFICATION
```

States:
```text
QUEUED
SENT
DELIVERED
FAILED
CANCELLED
```

## 47. Audit

Mindestens:
- Appointment erstellt
- bestätigt
- verschoben
- storniert
- durch Praxis geändert
- Proposal gesendet/angenommen
- Calendar Sync Fehler
- Rollen-/Rechteänderung
- Calendar Connection geändert

Keine unnötigen sensitiven Payloads.

## 48. Security Baseline

- TLS
- Encryption at Rest
- Tenant Isolation
- Least Privilege
- sichere Magic Links
- OTP/Verification
- Token Hashing
- Rate Limiting
- Session Rotation
- CSRF-Schutz soweit relevant
- Secret Management
- keine Klientendaten in Logs
- Lösch-/Aufbewahrungsregeln
- Audit Trail
- Backups
- Restore-Test
- MFA/2FA für Praxis-Accounts vor Produktion prüfen

## 49. Zielarchitektur

```text
Responsive Web/PWA
        │
        ▼
      Go API
        │
 ┌──────┼─────────────┐
 │      │             │
Domain  Worker      Provider Adapters
 │      │             │
 └──────┼─────────────┘
        ▼
   PostgreSQL
```

Architekturstil: **Modular Monolith**. Keine Microservices im MVP.

## 50. Fachmodule

```text
practice
identity
practitioner
client
service
availability
appointment
resource
calendar
notification
audit
```

Keine heimliche Kopplung über Fremdtabellen; klare Interfaces/Domain Services.

## 51. Zentrale Datenobjekte

```text
Practice
Location
User
PracticeMembership
AssistantPractitionerAccess
Practitioner
PractitionerSchedulingPreference
Client
ClientIdentity
ClientAccount
Service
PractitionerService
CalendarConnection
CalendarBusyInterval
CalendarSyncState
AvailabilityRule
AvailabilityOverride
BreakRule
Resource
PractitionerResource
Appointment
AppointmentHold
AppointmentPolicy
AppointmentProposal
AppointmentProposalOption
ReminderRule
AppointmentReminder
CallbackRequest
Notification
AuditEvent
```

## 52. PostgreSQL-Anforderungen

- nicht erratbare IDs
- `practice_id` konsistent
- `timestamptz`
- Praxis-/Location-Zeitzone
- Unique Constraints
- Transactional Booking
- DB-level Doppelbuchungsschutz
- keine JSON-Allzweckablage für Kernregeln
- strukturierte Migrationen
- Audit nicht überschreibbar

## 53. Zeitzonen

Intern `timestamptz` / UTC, Darstellung in Praxis-/Location-Zeitzone.

Pflichttests:
- DST vor/zurück
- nicht existente lokale Uhrzeit
- doppelt vorkommende lokale Uhrzeit

## 54. Concurrency

Test:
```text
Client A books 14:00
Client B books 14:00 simultaneously
```

Ergebnis: exakt einer wird `BOOKED`.

Erforderlich: Transaction + DB-Level Conflict Protection.

## 55. API-Schnitt grob

Public:
```text
GET  /public/{practiceSlug}/services
GET  /public/{practiceSlug}/availability
POST /public/{practiceSlug}/holds
POST /public/{practiceSlug}/bookings
POST /public/verify
GET  /t/{token}
POST /t/{token}/confirm
POST /t/{token}/reschedule
POST /t/{token}/cancel
```

Management:
```text
/api/practice
/api/practitioners
/api/services
/api/resources
/api/calendars
/api/appointments
/api/clients
/api/proposals
/api/settings
```

## 56. Frontend-Bereiche

Public:
- Was können wir für dich tun?
- Service
- Nächstmöglich / Therapeut
- Slot
- Kontaktdaten
- Verification
- Bestätigung
- Terminverwaltung

Practice:
- Heute
- Kalender
- Klienten
- Vorschläge
- Rückrufe
- Einstellungen

Setup:
- Praxis
- Therapeuten
- Kalender
- Arbeitszeiten
- Services
- Räume
- Reminder
- Buchungsregeln

## 57. UX-Leitlinien

- Mobile first
- wenige Pflichtfelder
- Kontaktdaten erst nach Slot-Auswahl
- Alltagssprache
- Progressive Disclosure
- klare Statuskommunikation
- Fehler mit Handlungsoption
- keine Sackgassen
- alte Buchung bei abgebrochener Umbuchung erhalten
- Sync-Fehler ruhig und verständlich kommunizieren

## 58. Planned Extensions

Datenmodellseitig berücksichtigen, UI später:
- Client Availability Preferences
- frühere Termine anbieten
- Warteliste
- flexible Folgetermine
- weitere Notification-Kanäle
- weitere Provider
- Video
- weitergehende Slot-Optimierung

Client Availability Preference ist Soft Preference, z. B. „Mo–Fr ab 16 Uhr“ oder „Mo 10–12, Do ab 16“.

## 59. Entwicklungsphasen

### Phase 0 – Foundation
Repo-Struktur, README, ADRs, Go, Frontend/PWA, PostgreSQL, Migrationen, Config, Logging, Health, Local Dev, CI, Testskeleton.

**Exit:** lokal startbar, Migrationen und CI grün.

### Phase 1 – Scheduling Kernel
Practice, Location, Practitioner, Service, PractitionerService, AvailabilityRule, Override, BreakRule, SchedulingPreference, CHRONOLOGICAL Slots.

**Acceptance:** Arbeitszeit, Pause, Booking Notice und Buffer korrekt.

### Phase 2 – Appointment Core
Client, Identity, Appointment, Hold, Policy, State Machine, Transaktion, Doppelbuchungsschutz, Public Booking, manuelle Praxis-Termine.

**Acceptance:** Holds verfallen, parallele Buchungen kollidieren korrekt, Pending Confirmation verfällt.

### Phase 3 – Passwordless Confirmation
OTP/Magic Link, MOBILE_ONLY, MOBILE_EMAIL, Verification, Confirmation Request, Secure Link.

### Phase 4 – Reschedule / Cancel
Sichere Umbuchung, Fristen, Audit.

### Phase 5 – Calendar Integration
Provider Interface, Google, CalDAV, Busy Cache, Worker, Sync State, Failure Policy.

### Phase 6 – Notifications
E-Mail, SMS, Reminder, Confirmations, Expiry Warning, Delivery Status.

### Phase 7 – Multi-Practitioner & Resources
„Nächstmöglich“, Resource, DEDICATED/SHARED/AUTO_ASSIGN.

### Phase 8 – Assistant / Reception
Assistant Role, Practitioner Scopes, Terminverwaltung.

### Phase 9 – Callback
SCHEDULED und REQUEST.

### Phase 10 – Appointment Proposals
Quick/Precise Proposal, Expiry, Client Acceptance.

### Phase 11 – Client Account / PWA
Meine Termine, Historie, Vorschläge, Reminder, Kontaktdaten.

### Phase 12 – Compact / Balanced Scheduling
Erst nach stabilem CHRONOLOGICAL. Scoring, Recovery Break, Slot Quality.

## 60. Vertikaler erster Slice

Früh einen echten End-to-End-Flow bauen:

```text
Practice
→ Practitioner
→ Service
→ Availability
→ Public Slot Search
→ Hold
→ Booking
→ Confirmation
```

Danach reale Kalender anbinden.

## 61. Teststrategie

Unit:
- State Machine
- Availability
- Buffer
- Break Rules
- Deadlines
- Proposal
- Reminder Scheduling

Integration:
- PostgreSQL Locking
- Migrationen
- Calendar Adapter
- Notification Provider
- Identity Verification

Concurrency:
- Doppelbuchung
- Hold vs Booking
- Resource Auto Assignment
- Proposal Acceptance vs normale Buchung

Time:
- Zeitzonen
- DST
- Fristen
- Reminder
- Hold/Confirmation Expiry

E2E:
- öffentliche Buchung
- Mobile only
- Mobile + E-Mail
- Umbuchung
- Stornierung
- Praxis-Termin
- Proposal
- Rückruf
- Kalenderausfall

## 62. Definition of Done

Ein Feature ist erst fertig, wenn:
- Domainlogik implementiert
- API validiert
- Berechtigungen geprüft
- Audit berücksichtigt
- Fehlerfall behandelt
- Happy Path getestet
- zentrale Edge Cases getestet
- keine sensitiven Daten unnötig geloggt
- Migration vorhanden, falls nötig
- Dokumentation aktualisiert
- mobile Darstellung geprüft

## 63. Definition of Product Success v0.1

Erfolgreich, wenn:
1. mehrere Therapeuten einrichtbar,
2. individuelle Arbeitsweisen möglich,
3. Buchung ohne Account,
4. „nächstmöglich“ praxisweit,
5. externe Kalender berücksichtigt,
6. Betrieb ohne externen Kalender,
7. Räume verhindern Doppelbelegung,
8. Bestätigung sicher,
9. Umbuchung riskiert alten Termin nicht,
10. Reminder zuverlässig,
11. Assistenz verwaltet Termine,
12. Vorschläge funktionieren,
13. Mobile-only funktioniert,
14. kein medizinischer Praxissoftware-Ballast.

## 64. Implementierungsregeln für Cody

Nicht ergänzen:
- Behandlungsakte
- Zahlungslogik
- Diagnosefelder
- JSON-Settings-Müllhalde für Kernregeln
- Providerlogik in Availability
- Microservices ohne zwingenden Grund
- Doppelbuchungsschutz nur im Frontend
- sequenzielle öffentliche IDs
- unnötige private Kalendertitel
- automatische Stornierung gebuchter Termine wegen fehlender später Attendance Confirmation
- Freigabe alten Termins vor erfolgreicher Umbuchung

Bei Unklarheit:
1. Product Principles prüfen,
2. Non-Goals prüfen,
3. scope-erhaltende Variante wählen,
4. als ADR/TODO dokumentieren.

## 65. Empfohlene Repo-Struktur

```text
apps/easy-trm/
├── README.md
├── docs/
│   ├── easytrm-development-plan-v0.1.md
│   ├── architecture.md
│   ├── domain-model.md
│   ├── scheduling-engine.md
│   ├── user-journeys.md
│   ├── security-privacy.md
│   ├── api.md
│   └── decisions/
│       └── ADR-xxxx-*.md
├── backend/
├── frontend/
├── migrations/
└── deploy/
```

## 66. Nächste fachliche Schritte

1. PRD gegen reale User Journeys testen.
2. Widersprüche / fehlende States identifizieren.
3. Development Plan aktualisieren.
4. Domain-Modell formalisieren.
5. State Machines formalisieren.
6. PostgreSQL-Schema entwerfen.
7. Availability Algorithmus präzisieren.
8. API Contracts definieren.
9. Wireframes ableiten.

## 67. Planned Extension Register

Parken:
- Client Availability Preferences
- frühere Termine
- Warteliste
- flexible Multi-Follow-up-Vorschläge
- weitere Notification-Kanäle
- Video
- weitere Kalenderprovider
- weitergehende Slot-Optimierung

Neue Erweiterung immer prüfen gegen:

> **EasyTrm organisiert Termine – nicht Therapien.**
