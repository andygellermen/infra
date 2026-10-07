# EasyTrm – PostgreSQL Model & Database Invariants v0.1

> Technische Baseline für Cody
> Bezug: Development Plan v0.2, State Machines v0.1, Domain Model v0.1, Scheduling Engine v0.1

## 1. Grundentscheidungen

- PostgreSQL
- UUIDs via `gen_random_uuid()`
- `timestamptz` für konkrete Ereigniszeiten
- IANA-Zeitzonen wie `Europe/Berlin`
- `practice_id` auf allen mandantenabhängigen Tabellen
- `btree_gist` für Exclusion Constraints

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS btree_gist;
```

## 2. Zentrale Tabellen

```text
practices
locations
users
practice_memberships
practitioners
assistant_practitioner_access
practitioner_scheduling_preferences
clients
client_identities
client_accounts
services
practitioner_services
booking_profiles
appointment_policies
resources
resource_requirements
practitioner_resources
resource_blocks
availability_rules
availability_overrides
break_rules
calendar_connections
calendar_busy_intervals
appointments
appointment_corrections
appointment_proposals
appointment_proposal_options
reminder_rules
notifications
callback_requests
audit_events
```

## 3. Tenant-Invariante

Beziehungen zwischen mandantenabhängigen Tabellen werden bevorzugt als Composite FK modelliert:

```sql
FOREIGN KEY (practice_id, practitioner_id)
REFERENCES practitioners(practice_id, id)
```

Referenzierte Tabellen erhalten dafür `UNIQUE (practice_id, id)`.

**INV-001:** Keine fachliche Beziehung darf über Practices hinweg zeigen.

## 4. Client Identity

`Client` und `ClientIdentity` bleiben practice-local. Keine automatische Cross-Practice-Zusammenführung.

```text
ClientIdentity:
- PHONE | EMAIL
- UNVERIFIED | VERIFIED | REVOKED
- value_normalized
- verified_at
- is_primary
```

**INV-002:** Nur VERIFIED Identities dürfen regulär für E-Mail/SMS und Terminverwaltung genutzt werden.

## 5. BookingProfile

`BookingProfile` trennt sichtbare Terminlogistik vom fachlich interpretierbaren Service.

Beispiel:

```text
Service: Osteopathie Ersttermin
BookingProfile: 75-Minuten-Termin mit Behandlungsraum
```

Damit kann Assistenz Räume planen, ohne zwingend die Behandlung zu sehen.

## 6. Appointments

Zentrale Felder:

```text
id
practice_id
location_id
practitioner_id
client_id nullable nur bei HELD
service_id optional
booking_profile_id optional
resource_id optional
source
status
attendance_status
starts_at
ends_at
buffer_before_minutes
buffer_after_minutes
hold_expires_at
confirmation_expires_at
confirmed_at
started_at
completed_at
cancelled_at
rescheduled_from_appointment_id
rescheduled_to_appointment_id
write_protected
version
```

Blockierende States:

```text
HELD
PENDING_CLIENT_CONFIRMATION
BOOKED
IN_PROGRESS
```

Nicht blockierend:

```text
COMPLETED
NO_SHOW
CANCELLED
RESCHEDULED
EXPIRED
VOIDED
```

## 7. Effective Occupancy Range

Buffer gehören zur Belegung:

```sql
ALTER TABLE appointments
ADD COLUMN occupancy_range tstzrange
GENERATED ALWAYS AS (
  tstzrange(
    starts_at - make_interval(mins => buffer_before_minutes),
    ends_at   + make_interval(mins => buffer_after_minutes),
    '[)'
  )
) STORED;
```

Halboffene Intervalle erlauben direkt angrenzende Termine ohne künstliche Überschneidung.

## 8. Practitioner-Exclusion

**INV-003:** Ein Practitioner darf nicht zwei aktive, überlappende Belegungen besitzen.

```sql
ALTER TABLE appointments
ADD CONSTRAINT appointments_no_practitioner_overlap
EXCLUDE USING gist (
  practice_id WITH =,
  practitioner_id WITH =,
  occupancy_range WITH &&
)
WHERE (status IN ('HELD','PENDING_CLIENT_CONFIRMATION','BOOKED','IN_PROGRESS'));
```

Das schützt auch bei parallelen Requests.

## 9. Resource-Exclusion

**INV-004:** Eine Resource darf nicht doppelt aktiv belegt sein.

```sql
ALTER TABLE appointments
ADD CONSTRAINT appointments_no_resource_overlap
EXCLUDE USING gist (
  practice_id WITH =,
  resource_id WITH =,
  occupancy_range WITH &&
)
WHERE (
  resource_id IS NOT NULL
  AND status IN ('HELD','PENDING_CLIENT_CONFIRMATION','BOOKED','IN_PROGRESS')
);
```

## 10. Holds

Finale v0.1-Entscheidung:

> `HELD` lebt in `appointments`, nicht in einer separaten Occupancy-Tabelle.

Vorteil: derselbe Exclusion Constraint schützt Hold, Reservation und Booking.

Regeln:

```text
HELD benötigt hold_expires_at
client_id darf nur bei HELD NULL sein
Expiry Worker setzt HELD -> EXPIRED
```

## 11. Pending Confirmation

`PENDING_CLIENT_CONFIRMATION` ist eine echte fachliche Reservierung und blockiert vollständig.

DB-Regel:

```text
confirmation_expires_at < starts_at
```

Eine strengere Safety Margin wird im Domain/Application Layer gesetzt.

## 12. Appointment / Attendance Consistency

Erlaubte Kernkombinationen:

```text
BOOKED + UNKNOWN
IN_PROGRESS + UNKNOWN
COMPLETED + PRESENT
NO_SHOW + NO_SHOW
CANCELLED + CANCELLED
VOIDED + historischer Attendance-Wert
```

**INV-005:** Widersprüchliche Kombinationen werden per CHECK verhindert.

## 13. Reschedule Integrity

**INV-006:** Ein Appointment darf nicht auf sich selbst als Reschedule-Ziel/-Quelle zeigen.

Composite Self-FKs halten Reschedules innerhalb derselben Practice.

Transaktion:

```text
neuen Slot sichern
new -> BOOKED
old -> RESCHEDULED
Referenzen setzen
Commit
```

Wenn der neue Slot scheitert, bleibt der alte Termin `BOOKED`.

## 14. Write Protection

Ab `IN_PROGRESS`, spätestens nach Terminende, wird der Termin strukturell schreibgeschützt.

Geschützt sind mindestens:

- Practitioner
- Client
- Start/Ende
- Resource
- Service
- BookingProfile

DB-Trigger verhindert direkte Änderungen.

Historische Korrektur erfolgt über `AppointmentCorrection`, nicht über stilles Überschreiben.

## 15. VOID statt Hard Delete

**INV-007:** Historische Termine werden nicht physisch gelöscht.

Fehltermin:

```text
Appointment -> VOIDED
+ AppointmentCorrection
+ AuditEvent
```

`appointment_corrections` und `audit_events` sind append-only.

## 16. ResourceBlock

`ResourceBlock` besitzt ein `tstzrange` und ist Hard Constraint.

Da PostgreSQL keinen Exclusion Constraint über zwei verschiedene Tabellen bilden kann, gilt v0.1:

1. Resource/Advisory Lock innerhalb der Booking-Transaction
2. aktive ResourceBlocks prüfen
3. Appointment-Konflikt prüfen
4. Appointment schreiben
5. Appointment-Exclusion schützt parallele Terminbelegungen

Eine gemeinsame `resource_occupancies`-Tabelle bleibt spätere Option.

## 17. Optimistic Versioning

`appointments.version bigint NOT NULL DEFAULT 1`.

Updates verwenden `WHERE id = ? AND version = ?` und erhöhen die Version.

0 aktualisierte Rows => `CONCURRENT_MODIFICATION`.

## 18. Notifications

`Notification.destination_identity_id` muss zur selben Practice gehören und VERIFIED sein.

Application Guard ist verpflichtend; Constraint Trigger als zusätzliche DB-Sicherung empfohlen.

Bei `RESCHEDULED`, `CANCELLED`, `EXPIRED`, `VOIDED` werden obsolete `QUEUED` Notifications auf `CANCELLED` gesetzt.

## 19. Proposals

`AppointmentProposalOption` blockiert keinen Slot.

Erst Auswahl:

```text
Availability re-check
-> HELD
-> Reservation/Booking
```

## 20. Wichtige Partial Indexes

```sql
CREATE INDEX appointments_pending_confirmation_idx
ON appointments (confirmation_expires_at)
WHERE status = 'PENDING_CLIENT_CONFIRMATION';

CREATE INDEX appointments_hold_expiry_idx
ON appointments (hold_expires_at)
WHERE status = 'HELD';

CREATE INDEX appointments_client_history_idx
ON appointments (practice_id, client_id, starts_at DESC);

CREATE INDEX notifications_due_idx
ON notifications (scheduled_at)
WHERE state = 'QUEUED';
```

## 21. Row Level Security

Nicht MVP-kritisch. Application Scoping + Composite FKs zuerst.

RLS später als Defense-in-Depth prüfen.

## 22. Primäre Datenbank-Invarianten

```text
INV-001 Tenant Isolation
INV-002 Verified communication only
INV-003 Practitioner exclusivity
INV-004 Resource exclusivity
INV-005 Appointment/Attendance consistency
INV-006 Reschedule integrity
INV-007 No historical hard delete
INV-008 HELD blocks until expiry
INV-009 PENDING blocks until confirmation/expiry
INV-010 terminal inactive states release occupancy
INV-011 valid interval: starts_at < ends_at
INV-012 confirmation expiry before appointment
INV-013 proposals do not reserve
INV-014 Audit/Corrections append-only
INV-015 write protection after start/history
```

## 23. Pflicht-Concurrency-Tests

1. Zwei Clients buchen denselben Practitioner-Zeitraum.
2. Zwei Practitioner buchen denselben Shared Room.
3. Confirmation-Expiry und Client-Bestätigung laufen gleichzeitig.
4. Reschedule konkurriert mit normaler Buchung des Zielslots.
5. Proposal Acceptance konkurriert mit Onlinebuchung.

Erwartung: immer genau ein konsistenter Endzustand.

## 24. SQL Error Mapping

```text
appointments_no_practitioner_overlap -> PRACTITIONER_UNAVAILABLE
appointments_no_resource_overlap     -> RESOURCE_UNAVAILABLE
client_identity_unique               -> IDENTITY_ALREADY_EXISTS
```

Rohes SQL darf nie an Frontend/Client durchgereicht werden.

## 25. Migrationsreihenfolge

```text
0001_extensions.sql
0002_enums.sql
0003_practice_identity.sql
0004_practitioners_clients.sql
0005_services_resources.sql
0006_availability.sql
0007_calendars.sql
0008_appointments.sql
0009_constraints.sql
0010_notifications.sql
0011_audit.sql
```

Migrationen sind forward-only und werden in CI auf leerer DB und Upgrade-Pfad getestet.

## 26. Finale v0.1-Technikentscheidungen

- HELD in `appointments`: **ja**
- RLS als erste Abhängigkeit: **nein**
- gemeinsame Resource-Occupancy-Tabelle: **noch nein**
- vollständige State Machine in Triggern: **nein**
- harte Invarianten zusätzlich in DB: **ja**

## 27. Definition of Done – Database Layer

- Composite Tenant FKs vorhanden
- Practitioner Exclusion Constraint grün
- Resource Exclusion Constraint grün
- HELD/PENDING blockieren korrekt
- terminale States geben Belegung frei
- Attendance-Check greift
- Confirmation-Expiry-Check greift
- Write Protection greift
- VERIFIED Destination abgesichert
- Audit/Corrections append-only
- Concurrency-Tests grün
- Migrationen reproduzierbar

## 28. Leitgedanke

```text
UI validation
↓
Application validation
↓
Domain invariants
↓
Database invariants
```

PostgreSQL ist in EasyTrm nicht nur Speicher, sondern die letzte Sicherheitslinie gegen Doppelbuchungen, Cross-Tenant-Verknüpfungen und inkonsistente historische Termine.