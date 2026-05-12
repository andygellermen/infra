# Easy-Event-Planner – MVP Scope

## Ziel des ersten MVP

Der erste MVP liefert einen stabilen Kern: Veranstalter loggen sich per Magic Link ein, legen Events an, veröffentlichen diese auf einer autarken Eventseite und binden sie per Snippet in Ghost ein. Teilnehmer melden sich kostenfrei an, bestätigen per Magic Link und landen bei ausgebuchten Veranstaltungen auf der Warteliste.

## MVP 1 – Event-Kern

Enthalten:

- Mandantenfähigkeit
- Magic-Link-Login für Veranstalter
- einfache Rollen: `owner`, `admin`, `event_manager`, `readonly`
- Eventreihen und Einzelveranstaltungen
- öffentliche Eventübersicht je Mandant
- öffentliche Eventdetailseite
- kostenfreie Anmeldung
- Magic-Link-Verifizierung für Teilnehmer
- Kapazitätsprüfung
- Warteliste
- Admin-Dashboard
- Teilnehmerliste
- Ghost-/CMS-Snippet für kommende Events
- E-Mail-Benachrichtigungen
- Morgen-E-Mail an Veranstalter
- Basis-Audit-Log
- Basis-Rate-Limiting
- einfache Datenschutz-/Anonymisierungsjobs

Nicht enthalten:

- PayPal
- Rabatt-/Gutschein-/Sponsoring-Links
- Spendenbasis
- Teilnehmerportal
- Zertifikate
- umfangreiche Kalenderintegration
- WordPress-Plugin
- Rechnungsstellung

## MVP 2 – Änderungen und Kalender

Enthalten:

- Eventstatus `changed`, `postponed`, `cancelled`
- Änderungsnotiz und Änderungsverlauf
- automatische Teilnehmerbenachrichtigung
- Veranstalter-Kalenderfeed als `.ics`-URL
- Teilnehmer-Kalenderdatei mit Anmeldebestätigung
- Kalenderupdates bei Änderung, Verschiebung oder Absage
- Admin-Direktlink im Veranstalter-Kalendereintrag

## MVP 3 – Payment

Enthalten:

- PayPal-Konfiguration je Mandant
- Ticketpreise
- PayPal Order Creation
- PayPal Webhook Handling
- Reservierungsablauf
- Zahlungsstatus
- Zahlungsbestätigung
- manuelle Rückerstattungsmarkierung

## MVP 4 – Einladungen und Rabatte

Enthalten:

- Rabattcodes
- Gutscheinlinks
- vollständig gesponserte Teilnahme
- teilbare Rabattlinks
- Nutzungslimits und Ablaufdaten
- Spendenmodell mit optionaler oder Mindestspende

## MVP 5 – Teilnehmerportal und Zertifikate

Enthalten:

- Teilnehmer-Magic-Link
- aktuelle und vergangene Veranstaltungen
- Stornierungsoption
- Teilnahmebescheinigungen
- Zertifikate als PDF
- Zertifikatsnummer und Verifikationslink

## MVP-1 Akzeptanzkriterien

```text
[x] Tenant kann angelegt werden
[x] Veranstalter kann Magic Link anfordern
[x] Login-Link ist einmalig und kurzlebig
[x] Eventreihe kann angelegt werden
[x] Event kann veröffentlicht werden
[x] Public Event Page zeigt kommende Events
[x] Teilnehmer kann Anmeldung starten
[x] Verifizierung bestätigt Anmeldung
[x] Kapazität wird nicht überschritten
[x] Warteliste greift bei ausgebuchtem Event
[x] Ghost-Snippet zeigt Events
[x] Veranstalter sieht Teilnehmerliste
[x] Morgen-E-Mail wird erzeugt
[x] Retention Dry-Run und Anonymisierung funktionieren
```

## MVP-Vermeidungsstrategie

Die wichtigste Gefahr ist Funktionsüberladung. PayPal, Rabatte, Zertifikate und Kalender sind wichtig, dürfen aber den stabilen Event-/Anmeldekern nicht blockieren. Erst wenn Anmeldung, Warteliste und Snippet zuverlässig laufen, werden Payment und Komfortfunktionen aufgesetzt.
