# EasyTrm – Scheduling & Availability Engine v0.1

> Status: fachlich-technische Spezifikation für Cody  
> Bezug: `easytrm-development-plan-v0.2.md`, `easytrm-state-machines-v0.1.md`, `easytrm-domain-model-v0.1.md`  
> Ziel: deterministische Slot-Ermittlung vor Persistenz-/SQL-Detaildesign

## 1. Zweck

Die Engine beantwortet ausschließlich:

> **Welche Termine sind für diese Anfrage fachlich zulässig und in welcher Reihenfolge sollen sie angezeigt werden?**

Sie erstellt selbst keine Buchung.

```text
Availability Query
      ↓
Slot Candidates
      ↓
Auswahl
      ↓
Appointment Hold
      ↓
Booking Transaction
```

Damit bleiben **Suchen** und **Buchen** getrennte Verantwortungen.

## 2. Kernprinzipien

1. Hard Constraints dürfen niemals verletzt werden.
2. Soft Preferences beeinflussen nur Ranking.
3. `SearchIntent` beeinflusst Ranking, nicht Hard Constraints.
4. Die Engine ist kalenderprovider-unabhängig.
5. Externe Kalender liefern normalisierte Busy-Intervalle.
6. Räume/Resources sind gleichwertige Belegungs-Constraints.
7. `HELD`, `PENDING_CLIENT_CONFIRMATION`, `BOOKED`, `IN_PROGRESS` blockieren.
8. `CANCELLED`, `EXPIRED`, `VOIDED`, `RESCHEDULED` blockieren nicht.
9. Slot-Berechnung ist deterministisch.
10. Beim Booking wird der Slot transaktional erneut geprüft.

## 3. Input: AvailabilityQuery

```text
AvailabilityQuery
- practice_id
- location_id optional
- service_id optional
- booking_profile_id optional
- search_intent
- practitioner_id optional
- range_start
- range_end
- client_id optional
- client_preferences optional
- result_limit
```

`SearchIntent`:

```text
NEXT_AVAILABLE
PRACTITIONER_SELECTED
BROWSE
PROPOSAL
```

## 4. Output: SlotCandidate

```text
SlotCandidate
- practitioner_id
- resource_id optional
- start_at
- end_at
- effective_start_at
- effective_end_at
- score
- score_reasons
```

`SlotCandidate` ist flüchtig und wird nicht als Termin persistiert.

## 5. Effective Interval

Konfliktprüfung arbeitet mit dem durch Buffer erweiterten Intervall:

```text
effective_start = appointment_start - buffer_before
effective_end   = appointment_end + buffer_after
```

Intervalle werden halboffen behandelt:

```text
[effective_start, effective_end)
```

Dadurch kollidieren direkt angrenzende Termine ohne Buffer nicht.

## 6. Blocking Appointment States

Blockieren Availability:

```text
HELD
PENDING_CLIENT_CONFIRMATION
BOOKED
IN_PROGRESS
```

Blockieren nicht:

```text
COMPLETED
NO_SHOW
CANCELLED
RESCHEDULED
EXPIRED
VOIDED
```

## 7. Practitioner-Auswahl

### NEXT_AVAILABLE
Alle aktiven und für Service/BookingProfile zulässigen Practitioner. Ein Practitioner mit `BLOCK_BOOKINGS` wegen Kalenderausfall wird nur aus dieser Kandidatenmenge entfernt; andere bleiben buchbar.

### PRACTITIONER_SELECTED
Nur der explizit gewählte Practitioner.

### BROWSE
Alle geeigneten Practitioner oder eine vorgefilterte Menge.

### PROPOSAL
Alle gemäß Proposal-Regeln zulässigen Practitioner.

## 8. SchedulingSpec

Vor der Slot-Erzeugung werden alle fachlichen Overrides einmalig in eine normalisierte Spezifikation aufgelöst:

```text
SchedulingSpec
- duration_minutes
- buffer_before_minutes
- buffer_after_minutes
- booking_notice_minutes
- confirmation_policy
- resource_requirement
- delivery_mode
- eligible_practitioners
```

Quellen:

```text
Practice Defaults
→ Practitioner Overrides
→ Service Overrides
→ PractitionerService Overrides
```

Spezifischer gewinnt.

## 9. Working Availability

Pro Practitioner:

1. Wochenverfügbarkeit laden.
2. Für Query-Zeitraum in konkrete Intervalle expandieren.
3. Praxis-/Location-Zeitzone anwenden.
4. Availability Overrides anwenden.
5. Fixed Breaks abziehen.

Ergebnis:

```text
WorkingInterval[]
```

## 10. Availability Overrides

Empfohlene Typen:

```text
CLOSED
OPEN_EXTRA
REPLACE_HOURS
```

- `CLOSED` entfernt Verfügbarkeit.
- `OPEN_EXTRA` ergänzt Zeit.
- `REPLACE_HOURS` ersetzt die reguläre Tageszeit.

Widersprüchliche Overrides sollen schon beim Pflegen verhindert werden.

## 11. Fixed Breaks

Beispiel:

```text
Work: 09:00–18:00
Break: 12:30–13:00
```

ergibt:

```text
09:00–12:30
13:00–18:00
```

Fixed Break ist Hard Constraint.

## 12. Externe Busy Intervals

Die Engine liest nur `CalendarBusyInterval[]`.

### USE_LAST_SYNC
Letzten bekannten Busy-Stand verwenden.

### BLOCK_BOOKINGS
Practitioner aus neuer Online-Suche entfernen, wenn die Verbindung laut Sync-/Staleness-Regel nicht vertrauenswürdig ist.

### Optionaler Kalender
Fehler beeinflusst Buchbarkeit nicht.

Keine Live-Provider-Abfrage innerhalb normaler Availability-Requests.

## 13. Interne Busy Intervals

Aus aktiven Appointments/Holds:

```text
InternalBusyInterval
- practitioner_id
- resource_id optional
- start
- end
- source
```

Aktive States:

```text
HELD
PENDING_CLIENT_CONFIRMATION
BOOKED
IN_PROGRESS
```

## 14. Resource Blocks

`ResourceBlock` ist Hard Constraint.

Keine Zuweisung auf blockierte Räume/Resources.

## 15. Candidate Generation

Aus freien Working Intervals werden Startzeitpunkte im Slot-Raster erzeugt.

Beispiel:

```text
slot_interval = 15 min
duration = 60 min
```

Mögliche Starts in 09:00–12:00:

```text
09:00
09:15
09:30
09:45
10:00
10:15
10:30
10:45
11:00
```

## 16. Candidate Generation Pseudocode

```text
function generateCandidateStarts(interval, slotInterval, duration, buffers):
    start = ceilToGrid(interval.start, slotInterval)

    while start + duration <= interval.end:
        effectiveStart = start - buffers.before
        effectiveEnd   = start + duration + buffers.after

        if effectiveStart >= interval.start
           and effectiveEnd <= interval.end:
            yield start

        start += slotInterval
```

## 17. Hard Constraint Evaluation

```text
function isHardValid(candidate, context):
    if candidate.start < now + bookingNotice:
        return false

    if not insideWorkingAvailability(candidate.effectiveInterval):
        return false

    if overlapsFixedBreak(candidate.effectiveInterval):
        return false

    if overlapsExternalBusy(candidate.effectiveInterval):
        return false

    if overlapsInternalPractitionerBusy(candidate.effectiveInterval):
        return false

    if violatesRecoveryBreak(candidate, context):
        return false

    resource = resolveResource(candidate, context)

    if context.requiresResource and resource == NONE:
        return false

    if resource != NONE
       and resourceUnavailable(resource, candidate.effectiveInterval):
        return false

    return true
```

## 18. Resource Resolution

```text
NONE
DEDICATED
SHARED
AUTO_ASSIGN
```

Bei `AUTO_ASSIGN`:

1. alle zulässigen Resources bestimmen,
2. blockierte/belegte entfernen,
3. stabil sortieren,
4. eine Resource für den Candidate vorschlagen.

Die endgültige Resource-Zuweisung erfolgt erst beim transaktionalen Hold/Booking.

## 19. Privacy-aware Resource Resolution

Availability kann über `booking_profile_id` statt `service_id` aufgerufen werden.

Damit kann die Assistenz neutrale Profile verwenden wie:

```text
60-Minuten-Termin mit Behandlungsraum
```

ohne zwingend die medizinisch interpretierbare Behandlungsart zu sehen.

## 20. Recovery Break

Recovery Break wird dynamisch aus bestehenden Terminen + Candidate berechnet und nicht als statisches Zeitfenster gespeichert.

Beispiel:

```text
max_continuous_work = 180 min
minimum_break = 30 min
```

Ein Candidate ist unzulässig, wenn durch ihn ein unzureichend unterbrochener Arbeitsblock die Maximaldauer überschreitet.

Pseudocode:

```text
function violatesRecoveryBreak(candidate, existingAppointments, rule):
    intervals = effectiveIntervals(existingAppointments)
    intervals.add(candidate.effectiveInterval)
    intervals.sortByStart()

    blocks = mergeIntervalsWhereGapIsLessThan(
        intervals,
        rule.minimumBreak
    )

    for block in blocks:
        continuousWork = occupiedMinutes(block)

        if continuousWork > rule.maxContinuousWork:
            return true

    return false
```

## 21. Booking Notice

```text
candidate.start >= now + booking_notice
```

Vergleich in absoluter Zeit; Darstellung lokal.

## 22. Query Range

Öffentliche Suche rechnet nur begrenzte Zeiträume.

Empfehlung:

```text
Initial: 7–14 Tage
Weitere Termine: Pagination / next range
```

Keine unbegrenzten Kalenderabfragen.

## 23. Soft Scoring

Erst nach bestandenen Hard Constraints.

Mögliche Faktoren:

```text
adjacent_to_previous
adjacent_to_next
gap_fragmentation
preferred_daypart
client_preference
practitioner_compactness
balanced_break_quality
```

Keine zufälligen Scores.

## 24. CHRONOLOGICAL

Keine Optimierungslogik nötig.

```text
sort start_at ASC
```

## 25. COMPACT

Ziel: unnötige Lücken vermeiden.

Relative Regeln:

```text
hoher Bonus: direkter Anschluss
mittlerer Bonus: kleine Lücke
Malus: neue isolierte Restlücke
```

Konkrete Punktwerte bleiben Implementierungsparameter.

## 26. BALANCED

Ziel: Kompaktheit + Erholung.

Bewertet:

- gute Anschlussfähigkeit,
- sinnvolle Tagesblöcke,
- gewünschte Pausen,
- geringe Fragmentierung.

Hard Recovery Break bleibt separat und darf nie durch Score überstimmt werden.

## 27. Client Availability Preferences

Planned Extension, nur Soft Preference.

Beispiel:

```text
Do 16:00–20:00
```

innerhalb: Bonus; außerhalb: weiterhin buchbar.

## 28. SearchIntent Ranking – verbindlich

### NEXT_AVAILABLE

```text
1. start_at ASC
2. score DESC
3. practitioner_id stable
4. resource_id stable
```

Der früheste gültige Slot gewinnt immer. Slot Quality ist nur Tie-Breaker.

### PRACTITIONER_SELECTED

Bei `CHRONOLOGICAL`:

```text
1. start_at ASC
2. stable tie-breakers
```

Bei `COMPACT` / `BALANCED`:

```text
1. score DESC
2. start_at ASC
3. resource stable
```

### BROWSE

```text
1. score DESC
2. start_at ASC
3. stable tie-breakers
```

### PROPOSAL

```text
1. Proposal Constraints
2. gewünschte zeitliche Streuung
3. score DESC
4. start_at ASC
```

## 29. Proposal Diversification

```text
function selectProposalCandidates(candidates, count, diversityRule):
    if diversityRule == DIFFERENT_DAYS:
        bestPerDay = bestCandidatePerLocalDay(candidates)
        return top(bestPerDay, count)

    if diversityRule == SAME_DAY:
        groups = groupByLocalDay(candidates)
        day = chooseBestDayWithEnoughCandidates(groups, count)
        return top(groups[day], count)

    return top(candidates, count)
```

## 30. NEXT_AVAILABLE über mehrere Practitioner

```text
function nextAvailable(query):
    allCandidates = []

    for practitioner in eligiblePractitioners(query):
        if practitionerBlockedByRequiredCalendarFailure(practitioner):
            continue

        allCandidates += computeCandidates(practitioner, query)

    sort allCandidates by:
        start_at ASC,
        score DESC,
        practitioner_id stable,
        resource_id stable

    return first query.result_limit
```

## 31. Reason Codes

Leere oder abgelehnte Ergebnisse liefern maschinenlesbare Gründe:

```text
NO_WORKING_HOURS
CALENDAR_BLOCKED
CALENDAR_SYNC_BLOCKED
NO_RESOURCE
BOOKING_NOTICE_NOT_MET
NO_MATCHING_SLOTS
SLOT_UNAVAILABLE
RESOURCE_UNAVAILABLE
HOLD_EXPIRED
IDENTITY_NOT_VERIFIED
```

UI übersetzt diese in verständliche Sprache.

## 32. Availability Result ist keine Garantie

Response kann enthalten:

```text
generated_at
valid_for_seconds
```

Bei Auswahl wird erneut geprüft.

## 33. Hold Creation

```text
function createHold(request):
    begin transaction

    resolved = revalidateExactSlot(request)

    if not resolved.valid:
        rollback
        return SLOT_UNAVAILABLE

    lock practitioner interval
    lock/allocate resource interval

    create AppointmentHold(
        expires_at = now + holdTTL
    )

    commit
```

## 34. Hold -> Reservation / Booking

```text
function convertHold(hold, client, policy):
    begin transaction

    lock hold

    if hold.expired:
        rollback
        return HOLD_EXPIRED

    verify identity/policy

    if policy.confirmation_required:
        create Appointment(
            status = PENDING_CLIENT_CONFIRMATION
        )
    else:
        create Appointment(
            status = BOOKED
        )

    hold.state = CONVERTED

    commit
    publish domain event
```

## 35. Confirmation Expiry Worker

```text
find Appointments
where status = PENDING_CLIENT_CONFIRMATION
and confirmation_expires_at <= now
```

Dann transaktional:

```text
lock appointment
if still pending:
    status = EXPIRED
    cancel stale notifications
    write audit event
```

Slot/Resource werden dadurch wieder verfügbar.

## 36. Rescheduling

Der alte Termin bleibt bestehen, bis der neue Slot sicher ist.

```text
function reschedule(oldAppointment, selectedSlot):
    begin transaction

    lock oldAppointment
    assert oldAppointment.status == BOOKED
    assert deadline permits reschedule

    newAllocation = lockAndValidate(selectedSlot)

    create new Appointment(status = BOOKED)

    oldAppointment.status = RESCHEDULED
    oldAppointment.rescheduled_to = new.id
    new.rescheduled_from = old.id

    invalidate old notifications
    schedule new notifications

    commit
```

Abbruch vor Commit lässt alten Termin unverändert.

## 37. Cancellation

```text
function cancel(appointment, actor):
    lock appointment
    verify rights
    verify deadline for client self-service

    status = CANCELLED
    attendance_status = CANCELLED
    invalidate notifications
    audit
```

## 38. Calendar Staleness

Mindestens:

```text
last_successful_sync
```

Optional:

```text
max_stale_duration
```

Auch `USE_LAST_SYNC` darf nach sehr langer Stale-Zeit warnen oder gemäß Policy blockieren.

## 39. Caching

Availability darf kurz gecacht werden, aber nie als Buchungsgarantie gelten.

Möglicher Cache Key:

```text
practice
service/profile
practitioner
search_intent
date range
policy version
availability version
```

MVP: zunächst keine komplexe Cache-Schicht, solange Performance reicht.

## 40. Idempotenz

Schreibende Endpunkte sollten Idempotency unterstützen:

```text
POST /holds
POST /bookings
POST /confirm
POST /reschedule
POST /cancel
```

Retry darf keine Doppeltermine erzeugen.

## 41. Determinismus

Bei identischem Datenstand und Query ist die Reihenfolge identisch.

Stabile Tie-Breaker:

```text
start_at
score
practitioner_id
resource_id
```

## 42. Performance-Prinzip

Zuerst:

1. korrekte Intervalllogik,
2. deterministische Tests,
3. klare DB-Abfragen,
4. dann erst Cache/Precomputation.

## 43. Unit-Test-Pflichtfälle

### Basic
- freier Tag
- voller Tag
- Booking Notice
- Buffer before/after
- Fixed Break

### Appointment States
- HELD blockiert
- PENDING blockiert
- BOOKED blockiert
- IN_PROGRESS blockiert
- CANCELLED/EXPIRED/VOIDED blockieren nicht

### Calendar
- Busy blockiert
- USE_LAST_SYNC
- BLOCK_BOOKINGS
- Practitioner-Ausfall blockiert andere nicht

### Resource
- Dedicated
- Shared
- Auto Assign
- ResourceBlock
- paralleler Raumkonflikt

### Recovery
- unter Grenze
- exakt an Grenze
- über Grenze
- ausreichende Pause trennt Block

### SearchIntent
- NEXT_AVAILABLE nimmt früheren Slot
- CHRONOLOGICAL sortiert zeitlich
- COMPACT bevorzugt Anschluss
- BALANCED berücksichtigt Erholung
- PROPOSAL diversifiziert

### Time
- DST Beginn
- DST Ende
- Tagesgrenzen
- Location-Zeitzone

## 44. Integration-/Concurrency-Pflichtfälle

1. Zwei Clients halten denselben Practitioner-Slot gleichzeitig.
2. Zwei Practitioner benötigen denselben Shared Room.
3. Auto Assign unter Konkurrenz.
4. Proposal Acceptance konkurriert mit normaler Buchung.
5. Confirmation Expiry trifft gleichzeitig mit Bestätigung ein.
6. Reschedule konkurriert mit Buchung des Zielslots.
7. Calendar Sync ändert Busy Cache während Hold-Erstellung.

Ergebnis muss immer transaktional konsistent sein.

## 45. Logging

Loggen:

```text
query_id
practice_id
practitioner_count
candidate_count
duration_ms
reason_code_counts
```

Nicht loggen:

- Client-Name
- Telefonnummer
- E-Mail
- private Kalendertexte
- medizinisch interpretierbare Hinweise

## 46. Metrics

Später:

```text
availability_query_duration
candidate_count
empty_result_rate
hold_conflict_rate
booking_conflict_rate
resource_conflict_rate
calendar_block_rate
```

Keine personenbezogenen Labels.

## 47. Definition of Done – Scheduling Engine

MVP-reif, wenn:

- CHRONOLOGICAL vollständig korrekt,
- NEXT_AVAILABLE praxisweit korrekt,
- Hard Constraints testabgedeckt,
- Practitioner + Resource Konflikte korrekt,
- Failure Policy korrekt,
- Hold/Booking-Revalidation korrekt,
- DST-Tests grün,
- deterministische Sortierung garantiert.

`COMPACT` und `BALANCED` werden danach schrittweise aktiviert.

## 48. Implementierungsreihenfolge für Cody

### Slice 1
```text
Working Availability
+ Service duration
+ Booking Notice
+ CHRONOLOGICAL
```

### Slice 2
```text
Appointments
+ Holds
+ Buffers
```

### Slice 3
```text
Resources
+ ResourceBlocks
```

### Slice 4
```text
External Busy Intervals
+ Failure Policy
```

### Slice 5
```text
NEXT_AVAILABLE across practitioners
```

### Slice 6
```text
Recovery Break
```

### Slice 7
```text
COMPACT
+ BALANCED
```

### Slice 8
```text
PROPOSAL diversification
+ Client Preferences later
```

## 49. Architectural Boundary

Availability darf:

- lesen,
- normalisieren,
- Kandidaten erzeugen,
- Hard Constraints prüfen,
- Scores berechnen,
- sortieren.

Availability darf nicht:

- Appointment final persistieren,
- Notifications senden,
- Calendar Provider live aufrufen,
- Client Identity verifizieren,
- Audit Events für reine Suchvorgänge erzeugen.

## 50. Nächster Schritt

Aus dieser Spezifikation wird das PostgreSQL-Schema abgeleitet.

Besonders zu lösen:

- `tstzrange`
- Exclusion Constraints
- partielle Konfliktregeln je Appointment State
- Practitioner-Konflikte
- Resource-Konflikte
- Holds mit Expiry
- tenant-scoped Unique Constraints
- Reschedule-Referenzen
- verified ClientIdentity
- append-only Audit/Corrections
