# Easy-Event-Planner – Magic-Link-Authentifizierung

## Ziel

Magic Links sind der primäre Login- und Verifizierungsmechanismus. Sie werden für Veranstalter-Login, Teilnehmerbestätigung, Teilnehmerportal, Wartelistenangebot, Stornierung und Zertifikatszugriff genutzt.

## Flow

```text
E-Mail eingeben -> Token erzeugen -> Hash speichern -> Link senden -> Token prüfen -> Session oder Aktion ausführen -> Link verbrauchen
```

## Sicherheitsregeln

- Token kryptografisch zufällig erzeugen.
- Token nie im Klartext speichern.
- Hash mit serverseitigem Pepper speichern.
- Token zweckgebunden speichern: `organizer_login`, `registration_verify`, `participant_login`, `waitlist_offer`, `registration_cancel`, `certificate_access`.
- Token kurzlebig und einmalig verwendbar machen.
- Nach erfolgreicher Nutzung `used_at` setzen.
- Token aus URL entfernen: erfolgreicher Verify führt auf saubere Redirect-URL.
- Anforderungs-Endpunkt antwortet neutral, auch wenn E-Mail unbekannt ist.

## Empfohlene Gültigkeiten

```text
organizer_login: 15 Minuten
registration_verify: 30 Minuten
participant_login: 15 Minuten
waitlist_offer: 12–24 Stunden
registration_cancel: 30 Minuten
certificate_access: 30 Minuten
```

## Session

Cookie-Empfehlung: `HttpOnly`, `Secure`, `SameSite=Lax`. Cookie enthält nur zufälligen Session-Wert; Datenbank speichert Hash. Logout setzt `revoked_at`.

## Rate Limits

Mindestens nach IP, E-Mail, Tenant und Purpose limitieren. Beispiel: 5 Login-Links pro 15 Minuten je E-Mail/IP.

## Audit Events

`magic_link_requested`, `magic_link_sent`, `magic_link_verified`, `magic_link_rejected`, `session_created`, `session_revoked`.

## Akzeptanzkriterien

```text
[x] Token ist gehasht gespeichert
[x] Link ist einmalig verwendbar
[x] Link läuft ab
[x] neutrale Antwort bei Anforderung
[x] Session-Cookie ist sicher konfiguriert
[x] Rate Limiting aktiv
[x] Audit-Log aktiv
```
