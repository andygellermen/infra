# Easy-Event-Planner – Registration Flow

## Grundregeln

- Das Frontend trifft keine finale Platzentscheidung.
- Kapazitäten werden serverseitig geprüft.
- Confirmed-Anmeldungen überschreiten die Kapazität nicht.
- Kostenfreie Events nutzen Magic-Link-Verifizierung.
- Kostenpflichtige Events nutzen Reservierung plus PayPal.
- Warteliste greift, wenn keine Plätze verfügbar sind und Warteliste aktiv ist.

## Kostenfreier Flow

```text
Event öffnen -> Formular absenden -> Participant anlegen/finden -> Registration=verification_pending -> Magic Link senden -> Link prüfen -> Kapazität prüfen -> confirmed oder waitlist -> Outcome-E-Mail senden (inkl. Bestätigung bzw. Wartelisten-Hinweis und Abmeldehinweis)
```

## Kostenpflichtiger Flow

```text
Formular absenden -> Kapazität prüfen -> Registration=reserved -> reserved_until setzen -> PayPal Order erzeugen -> payment_pending -> Webhook bestätigt Zahlung -> confirmed -> Bestätigungs-E-Mail senden
```

## Warteliste

```text
Event voll -> Registration=waitlist -> Position vergeben -> Bestätigung senden -> bei freiem Platz Angebot senden -> Angebot läuft ab oder wird angenommen
```

## Platzbindung

Platzbindend: `confirmed`, `reserved` mit gültigem `reserved_until`, `payment_pending` mit gültigem `reserved_until`.

Nicht platzbindend: `verification_pending`, `waitlist`, `cancelled`, `expired`, `rejected`.

## Pflichtfelder MVP

`event_id`, `name`, `email`, `privacy_accepted`.

Optional: `phone`, `participation_type`, `message`, `invite_code`.

## Stornierung

Teilnehmer storniert im Teilnehmer-Portal innerhalb der tenantweit konfigurierten Abmeldefrist, Registration wird `cancelled`, Platz wird frei, Warteliste kann nachrücken.

## Missbrauchsschutz

Turnstile oder vergleichbare Prüfung, Honeypot, Rate Limit pro IP/E-Mail/Event, Double-Opt-In, Audit-Log.

## Akzeptanzkriterien

```text
[x] Anmeldung startet
[x] Magic-Link-Verifizierung funktioniert
[x] Kapazität wird eingehalten
[x] Warteliste funktioniert
[x] Veranstalter wird informiert
[x] Teilnehmer erhält Bestätigung
```
