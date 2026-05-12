# Easy-Event-Planner – Privacy & Retention Konzept

## Ziel

Teilnehmerdaten sollen nur so lange gespeichert werden, wie sie organisatorisch benötigt werden. Danach werden sie gelöscht oder anonymisiert. Zahlungs- und Nachweisdaten werden getrennt betrachtet.

## Datenkategorien

`participant_contact_data`, `registration_data`, `payment_reference_data`, `certificate_data`, `technical_logs`, `audit_logs`, `anonymous_statistics`.

## Empfohlene Defaults

```text
Teilnehmer-Kontaktdaten: 30 Tage nach Eventende
Wartelisten-Kontaktdaten: 30 Tage nach Eventende
Magic Links: 7 Tage nach Ablauf
Sessions: 30 Tage nach Ablauf
E-Mail-Jobs: 90 Tage
Audit-Logs: 180 Tage
anonymisierte Statistiken: unbegrenzt
```

## Retention Policy

Pro Tenant und Datenkategorie: `action = anonymize | delete`, `retention_days`, `enabled`.

## Anonymisierung

Name, E-Mail und Telefon werden entfernt, Registrierungsstatistik bleibt erhalten. So bleiben Eventzahlen, Status und Teilnahmeart auswertbar, ohne Personenbezug zu behalten.

## Dry-Run

Jeder Löschjob muss Dry-Run unterstützen. Ergebnis: Anzahl betroffener Datensätze pro Kategorie, ohne Änderung.

## Jobflow

```text
Policies laden -> betroffene Daten zählen -> bei Dry-Run stoppen -> anonymisieren/löschen -> Audit-Log -> Jobstatus speichern
```

## Vermeidungsstrategien

- Gegen zu frühe Löschung: konfigurierbare Fristen, Dry-Run, Admin-Hinweis.
- Gegen zu späte Löschung: aktive Defaults, automatischer Job, Monitoring.
- Gegen unvollständige Anonymisierung: Datenkategorien vollständig pflegen und keine Vollpayloads langfristig speichern.

## Akzeptanzkriterien

```text
[x] Policy pro Tenant
[x] Dry-Run
[x] Teilnehmeranonymisierung
[x] Magic-Link Cleanup
[x] Audit-Log
```
