# EasyTrm – Entwicklungsplan v0.2 für Cody

> Status: konsolidierte fachliche Baseline nach User-Journey-Gegentest  
> Produkt: EasyTrm  
> Domain: \`easytrm.de\` (Arbeitsname)  
> Repository-Ziel: \`apps/easy-trm/\`  
> Leitidee: **EasyTrm organisiert Termine – nicht Therapien.**

## 1. Zweck dieses Dokuments

Dieses Dokument erweitert den Entwicklungsplan v0.1 um die verbindlichen Erkenntnisse aus dem User-Journey-Gegentest sowie die Entscheidungen zu verifizierten Kommunikationskanälen, privacy-aware Raumplanung, Attendance Reconciliation und SearchIntent.

Die vollständige v0.1-Baseline bleibt fachlich gültig, soweit sie nicht durch die folgenden v0.2-Regeln präzisiert oder ersetzt wird.

## 2. Verbindliche v0.2-Nachschärfungen

### 2.1 HELD und PENDING_CLIENT_CONFIRMATION strikt trennen

\`HELD\` ist eine kurze technische Sperre im laufenden Buchungsvorgang.

\`PENDING_CLIENT_CONFIRMATION\` ist eine fachliche Reservierung über einen konfigurierten Zeitraum und blockiert Practitioner und Resource bis Bestätigung oder Ablauf.

\`\`\`text
AVAILABLE
→ HELD
→ PENDING_CLIENT_CONFIRMATION
→ BOOKED
\`\`\`

### 2.2 Verifizierte Kommunikationskanäle only

Unverifizierte E-Mail-Adressen und Telefonnummern werden nicht für operative Kommunikation verwendet und sind in der normalen Kontaktansicht für Practitioner und Assistenz unsichtbar.

\`\`\`text
UNVERIFIED
→ nicht als regulärer Kontaktkanal nutzbar
→ nicht in normaler Kontakt-UI sichtbar

VERIFIED
→ nutzbar
→ gemäß Berechtigung sichtbar
\`\`\`

E-Mail ist niemals zwingend. Bestätigung kann vollständig Mobile-only über SMS/OTP/Magic-Link erfolgen.

### 2.3 Bestätigungsfrist absichern

\`\`\`text
confirmation_expires_at =
min(
  created_at + configured_confirmation_timeout,
  appointment_start - confirmation_safety_margin
)
\`\`\`

Eine Reservierung darf nicht bis hinter den Terminbeginn oder unvernünftig nahe an ihn heranreichen.

### 2.4 Kalenderausfall ist practitioner-lokal

Ein fehlerhafter erforderlicher Kalender entfernt nur den betroffenen Practitioner aus der Availability-Suche. Andere Practitioner bleiben buchbar.

## 3. Privacy-aware Resource Planning

### 3.1 ResourceBlock

Zusätzlich zu Terminbelegungen werden explizite Sperrungen unterstützt:

\`\`\`text
ResourceBlock
- id
- practice_id
- resource_id
- start_at
- end_at
- block_type
- reason_optional
- created_by
- created_at
\`\`\`

Beispiele: Raum defekt, Reinigung, interne Nutzung, Fremdnutzung, bewusste Reservierung.

### 3.2 Sichtbarkeit

Eine berechtigte Praxisansicht darf für eine Raumbelegung zeigen:

- Raum
- Zeit
- Practitioner
- Client/Patient gemäß Berechtigung

Die Sichtbarkeit eines medizinisch interpretierbaren Service-Namens ist davon getrennt.

> **Resource Visibility ist nicht gleich Treatment Visibility.**

### 3.3 Service und Resource Requirement entkoppeln

\`\`\`text
Service
  ↓
BookingProfile / ResourceRequirement
  ↓
ResourceCategory / Resource
\`\`\`

Die Scheduling Engine darf intern die nötige Resource bestimmen, ohne dass Assistenz den Behandlungsnamen sehen muss.

### 3.4 Neutraler BookingProfile-Layer

\`\`\`text
BookingProfile
- duration_minutes
- resource_requirement
- delivery_mode
- assistant_label
- internal_service_reference
\`\`\`

Mögliche neutrale Labels:

- Termin 30 Min.
- Termin 60 Min.
- Termin mit Behandlungsraum
- Telefontermin
- Termin ohne Raum

### 3.5 Assistenzbuchung mit Raum

Ablauf:

1. Practitioner wählen
2. Client wählen
3. BookingProfile / zulässige Terminparameter wählen
4. Engine prüft Resource
5. Raum wird vorgeschlagen oder automatisch gewählt
6. Assistenz sieht die Raumbelegung
7. Termin + Resource werden atomar gebucht

### 3.6 Atomarer Doppelbuchungsschutz

Die DB schützt gleichzeitig:

\`\`\`text
Practitioner + Zeitintervall
Resource + Zeitintervall
\`\`\`

Auto-Assignment erfolgt in derselben Transaktion.

## 4. Appointment Lock & Attendance Reconciliation

EasyTrm bleibt keine Rechnungssoftware. Neu aufgenommen wird eine operative Gegenprüfung tatsächlich stattgefundener Termine.

Ziele:

- geplante vs. tatsächlich wahrgenommene Termine unterscheiden
- geleistete Termine verlässlich zählen
- Fehltermine nachvollziehbar korrigieren
- vergangene Termine vor versehentlicher Änderung schützen

### 4.1 Zusätzlicher Status IN_PROGRESS

\`\`\`text
BOOKED
→ IN_PROGRESS
→ COMPLETED
\`\`\`

Alternativen:

\`\`\`text
BOOKED → NO_SHOW
BOOKED/COMPLETED/NO_SHOW → VOIDED
\`\`\`

\`VOIDED\` schließt den Termin aus der operativen Leistungszählung aus, ohne Historie zu löschen.

### 4.2 Kein Hard Delete vergangener Termine

\`\`\`text
started_at != null OR appointment.end_at < now
→ write_protected = true
\`\`\`

Vergangene oder gestartete Termine werden nicht normal gelöscht.

### 4.3 Admin Correction statt Hard Delete

Admin-Korrektur verlangt:

- Warnhinweis
- Korrekturgrund
- AuditEvent
- historische Nachvollziehbarkeit

UI:

\`\`\`text
Termin korrigieren / verschieben
Fehltermin ausbuchen (Admin)
Abbrechen
\`\`\`

### 4.4 Attendance Status separat halten

\`\`\`text
attendance_status:
- UNKNOWN
- PRESENT
- NO_SHOW
- CANCELLED
\`\`\`

Appointment Status beschreibt den Workflow, \`attendance_status\` die tatsächliche Teilnahme.

### 4.5 Gegenprüfungsansicht

Anzeigen:

- geplante Termine
- angetretene Termine
- No-Shows
- stornierte Termine
- korrigierte/voided Termine
- ungeklärte vergangene Termine

Keine Preise, Rechnungen oder Krankenkassenlogik.

## 5. Notification Invalidation

Bei \`RESCHEDULED\`, \`CANCELLED\`, \`EXPIRED\` oder \`VOIDED\` werden nicht mehr passende geplante Notifications auf \`CANCELLED\` gesetzt.

Neue Reminder werden aus dem aktuellen Terminstand neu berechnet.

Vor Versand prüft der Worker zusätzlich:

\`\`\`text
appointment current?
notification still applicable?
\`\`\`

## 6. SearchIntent – finale Entscheidung

\`SearchIntent\` wird verbindlicher fachlicher Input der Availability Engine:

\`\`\`text
NEXT_AVAILABLE
PRACTITIONER_SELECTED
BROWSE
PROPOSAL
\`\`\`

### NEXT_AVAILABLE

\`\`\`text
1. Hard Constraints
2. earliest start_at
3. Slot Quality als Tie-Breaker
\`\`\`

COMPACT/BALANCED darf niemals einen späteren Termin vor einen früheren stellen, wenn der Client explizit „nächstmöglich“ wählt.

### PRACTITIONER_SELECTED

\`\`\`text
1. Hard Constraints
2. Practitioner Scheduling Preference
3. start_at
\`\`\`

### BROWSE

\`\`\`text
1. Hard Constraints
2. Slot Quality / BALANCED / COMPACT
3. start_at
\`\`\`

### PROPOSAL

\`\`\`text
1. Hard Constraints
2. Proposal Constraints
3. zeitliche Streuung / gewünschte Tagesregeln
4. Slot Quality
\`\`\`

## 7. Practice-local Client Identity

Für v0.1/v0.2 bleiben \`Client\`, \`ClientAccount\` und \`ClientIdentity\` an \`practice_id\` gebunden.

Keine automatische Cross-Practice-Zusammenführung gleicher Mobilnummern oder E-Mail-Adressen.

## 8. Ergänzte zentrale Objekte

\`\`\`text
ResourceBlock
BookingProfile
AppointmentCorrection
attendance_status on Appointment
\`\`\`

## 9. Ergänzte Non-Goals

Weiterhin ausgeschlossen:

- Rechnungsstellung
- Preisberechnung
- Mahnwesen
- Zahlungsstatus
- Krankenkassenabrechnung
- Gebührenberechnung für No-Shows

Attendance Reconciliation darf nicht schleichend zu Billing werden.

## 10. Ergänzte Acceptance Criteria

### Identity

- unverifizierte Kanäle nicht regulär verwenden
- unverifizierte Kanäle in normaler Kontaktansicht verbergen
- Mobile-only vollständig ohne E-Mail

### Resources

- ResourceBlock verhindert Buchung
- Assistenz kann Raumlogistik sehen, ohne zwingend Treatment-Namen zu sehen
- Practitioner + Resource atomar blocken
- Raumbelegung ist in berechtigter Praxisansicht sichtbar

### Confirmation

- Pending Confirmation blockiert Slot bis Expiry
- Expiry nie nach Terminbeginn
- Ausfall eines Practitioner-Kalenders blockiert keine anderen Practitioner

### Notifications

- Reschedule/Cancel invalidiert alte Reminder
- Worker versendet nichts für nicht mehr passenden Appointment-State

### Attendance

- gestarteter/vergangener Termin wird schreibgeschützt
- Practitioner kann abgeschlossenen Termin nicht normal löschen
- Admin-Korrektur verlangt Bestätigung + Grund
- kein Hard Delete nötig
- Korrektur bleibt auditierbar

### SearchIntent

- NEXT_AVAILABLE priorisiert den frühesten gültigen Termin
- BROWSE darf Slot Quality höher gewichten
- PROPOSAL respektiert Vorschlagsregeln
- Ranking ist deterministisch testbar

## 11. Ergänzte Entwicklungsreihenfolge

### Appointment Core

zusätzlich:

- \`PENDING_CLIENT_CONFIRMATION\` als echte Sperre
- \`IN_PROGRESS\`
- \`attendance_status\`
- Write Protection

### Calendar Integration

zusätzlich:

- practitioner-lokales Failure Handling

### Notifications

zusätzlich:

- Invalidation / Recalculation

### Multi-Practitioner & Resources

zusätzlich:

- \`ResourceBlock\`
- \`BookingProfile\`
- privacy-aware Resource Views
- atomare Practitioner+Resource-Buchung

### Neue Phase: Attendance Reconciliation

- Attendance Status
- Write Protection
- Admin Correction / VOID
- Gegenprüfungsansicht
- Audit

Keine Billing-Funktionen.

### Scheduling Ranking

zusätzlich:

- \`SearchIntent\`
- intent-spezifische Ranking-Strategien

## 12. Vermeidungsstrategien

1. Keine Treatment-Leaks über Resource UI.
2. Kein Hard Delete historischer Termine.
3. Kein Billing-Scope-Creep.
4. Keine veralteten Notifications nach Terminänderung.
5. Kein praxisweiter Ausfall wegen eines einzelnen Kalenders.
6. Keine Slot-Optimierung gegen Client Intent.
7. Keine operative Nutzung unverifizierter Kontaktdaten.
8. Keine automatische Cross-Practice-Identitätsfusion.

## 13. Nächste Schritte

1. Domain-Modell mit den neuen Objekten formalisieren.
2. Appointment- und Attendance-State-Machines formalisieren.
3. Resource-/BookingProfile-Rechte modellieren.
4. PostgreSQL-Constraints für Practitioner- und Resource-Intervalle definieren.
5. SearchIntent-Ranking als Pseudocode spezifizieren.
6. Notification Invalidation formal definieren.
7. danach API Contracts und Wireframes ableiten.

> **EasyTrm organisiert Termine – nicht Therapien.**
