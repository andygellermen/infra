# Easy-Event-Planner – Discount, Invitation, Voucher & Sponsoring Model

## Ziel

Veranstalter können Einladungslinks, Rabattcodes, Gutscheinlinks, gesponserte Teilnahmen, teilbare Rabattlinks und Spendenmodelle verwenden.

## Linktypen

`plain_invitation`, `discount_percent`, `discount_fixed`, `voucher_fixed`, `voucher_full`, `sponsorship_full`, `donation_enabled`, `early_bird`, `shareable_referral`.

## Zentrale Felder

`tenant_id`, `event_id`, `series_id`, `code`, `label`, `invite_type`, `discount_type`, `discount_value`, `max_uses`, `used_count`, `max_uses_per_email`, `starts_at`, `expires_at`, `is_shareable`, `status`.

## Rabattarten

- Prozentual: 20 % Rabatt.
- Fester Betrag: 10 EUR Rabatt.
- Vollständig: Preis wird auf 0 gesetzt.
- Sponsoring: Preis 0, intern als gesponsert markiert.

## Teilbare Links

Teilbare Links benötigen Ablaufdatum, Nutzungslimit, Pro-E-Mail-Limit, Rate Limiting und Monitoring.

## Spendenbasis

Varianten: kostenlos mit optionaler Spende, empfohlene Spende, Mindestspende, Ticketpreis plus freiwillige Spende.

## Validierung

Ein Code ist nur gültig, wenn Tenant passt, Status aktiv ist, Zeitraum passt, Event/Reihe passt, Nutzungslimits nicht erschöpft sind und Pro-E-Mail-Limit nicht verletzt wird.

## Akzeptanzkriterien

```text
[x] Code reduziert Preis
[x] Sponsoring setzt Preis auf 0
[x] Nutzungslimit wird eingehalten
[x] Ablaufdatum wird geprüft
[x] Redemption wird protokolliert
[x] Link kann pausiert werden
```
