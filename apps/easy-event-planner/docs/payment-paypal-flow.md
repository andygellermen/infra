# Easy-Event-Planner – PayPal Payment Flow

## Grundsatz

PayPal ist nachgelagert zum stabilen Event-/Anmeldekern. Die Zahlungswahrheit entsteht serverseitig über PayPal-Webhooks oder serverseitige Order-Prüfung, nicht über Browser-Rückleitung.

## Standardflow

```text
Registration starten -> Preis berechnen -> Kapazität prüfen -> Registration=reserved -> reserved_until=now+15min -> PayPal Order erzeugen -> payment_pending -> Teilnehmer zahlt -> Webhook -> Payment=paid -> Registration=confirmed
```

## Status

Payment: `not_required`, `created`, `payment_pending`, `paid`, `failed`, `cancelled`, `refunded`, `partially_refunded`, `expired`.

Registration bei Payment: `reserved`, `payment_pending`, `confirmed`, `expired`, `cancelled`, `waitlist`.

## Reservierung

Reservierungen laufen automatisch ab. Abgelaufene Reservierungen geben Plätze frei. Ein Job `expire-stale-reservations` prüft `reserved_until`.

## Preisberechnung

```text
base_price - discount + optional_donation = final_amount
```

Endbetrag darf nicht negativ sein. Sponsoring kann Endbetrag auf 0 setzen. Spendenbetrag wird getrennt gespeichert.

## Webhook-Verarbeitung

Pflichten: Payload speichern, Signatur prüfen, Event-ID deduplizieren, idempotent verarbeiten, Payment/Registration aktualisieren, E-Mail-Jobs erzeugen, Audit-Log schreiben.

## Fehlerfälle

- Zahlung abgebrochen: Payment=`cancelled`, Reservierung freigeben.
- Zahlung fehlgeschlagen: Payment=`failed`, erneuter Versuch bis Ablauf.
- Webhook doppelt: 200 OK, keine doppelte Verarbeitung.
- Zahlung nach Ablauf: Kapazität erneut prüfen, sonst manuelle Klärung/Rückerstattung.

## Akzeptanzkriterien

```text
[x] Order serverseitig erstellt
[x] Reservierung schützt Kapazität
[x] Webhook dedupliziert
[x] Zahlung bestätigt Anmeldung
[x] Abbruch gibt Platz frei
```
