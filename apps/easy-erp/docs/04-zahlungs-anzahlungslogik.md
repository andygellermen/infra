# Datei: `docs/04-zahlungs-anzahlungslogik.md`

## 1. Ziel

Die Zahlungslogik muss Anzahlungen, Teilzahlungen, Schlusszahlungen, Überzahlungen, Rückzahlungen und Verrechnungen sauber abbilden.

## 2. Grundprinzip

```text
Eine Zahlung ist ein Geldereignis.
Eine Zahlung wird über Allocations Forderungen, Rechnungen oder Rückzahlungen zugeordnet.
```

## 3. Zentrale Objekte

| Objekt | Zweck |
|---|---|
| PaymentRequest | Zahlungsanforderung, z. B. Anzahlung |
| Payment | tatsächlicher Zahlungseingang |
| PaymentAllocation | Zuordnung einer Zahlung zu Beleg/Forderung |
| Refund | Rückzahlung an Kunden |
| CreditBalance | Guthaben des Kunden |

## 4. Zahlungsstatus

| Status | Bedeutung |
|---|---|
| not_requested | keine Zahlung angefordert |
| deposit_requested | Anzahlung angefordert |
| deposit_overdue | Anzahlung überfällig |
| deposit_paid | Anzahlung bezahlt |
| partially_paid | teilweise bezahlt |
| paid | vollständig bezahlt |
| overpaid | überzahlt |
| refunded | erstattet |
| cancelled | Zahlungsanforderung storniert |
| written_off | ausgebucht |

## 5. Statusfluss Anzahlung

```text
order_created
  ↓
payment_terms_evaluated
  ↓
deposit_required? ── no → no_deposit_required
  ↓ yes
deposit_request_created
  ↓
deposit_requested
  ↓
payment_received? ── no → deposit_pending
  ↓ yes
payment_allocated_to_deposit
  ↓
deposit_paid
  ↓
order_ready_for_fulfillment
```

## 6. Statusfluss Rechnung und Restzahlung

```text
invoice_created
  ↓
previous_deposits_checked
  ↓
deposits_allocated_to_invoice
  ↓
remaining_amount_calculated
  ↓
invoice_sent
  ↓
payment_pending
  ↓
payment_received
  ↓
allocation_created
  ↓
invoice_partially_paid OR invoice_paid OR invoice_overpaid
```

## 7. Statusfluss Storno mit Zahlung

```text
cancellation_approved
  ↓
paid_amount_checked
  ↓
cancellation_fee_calculated
  ↓
paid_amount > fee?
  ├── yes → refund_or_credit_required
  │          ↓
  │        refund_pending OR credit_balance_created
  │          ↓
  │        refund_completed OR credit_available
  │
  ├── equal → payment_fully_consumed_by_fee
  │          ↓
  │        cancellation_completed
  │
  └── no → remaining_fee_invoice_required
             ↓
           fee_invoice_created
```

## 8. PaymentRequest-Typen

| Typ | Beschreibung |
|---|---|
| deposit | Anzahlung |
| partial | Teilzahlung |
| final | Schlusszahlung |
| cancellation_fee | Stornogebühr |
| correction_due | Forderung aus Korrektur |
| refund_due | Rückzahlung an Kunden |

## 9. Payment-Typen

| Typ | Beschreibung |
|---|---|
| bank_transfer | Überweisung |
| cash | Barzahlung |
| card | Kartenzahlung |
| paypal | PayPal |
| stripe | Stripe |
| internal_credit | internes Guthaben |
| manual_adjustment | manuelle Korrektur |

## 10. PaymentAllocation-Typen

| Typ | Beschreibung |
|---|---|
| deposit | Zahlung wird als Anzahlung zugeordnet |
| invoice | Zahlung wird einer Rechnung zugeordnet |
| partial | Teilzuordnung |
| final | Restbetrag/Schlusszahlung |
| cancellation_fee | Verrechnung mit Stornogebühr |
| refund | Rückzahlungszuordnung |
| credit_balance | Umbuchung in Kundenguthaben |

## 11. Beispiel: Anzahlung und Schlussrechnung

```text
Angebot/Bestellung: 1.000,00 € brutto
Anzahlung laut Zahlungsbedingung: 30 Prozent = 300,00 €
Zahlungseingang: 300,00 €
Rechnung später: 1.000,00 €
Allocation:
- 300,00 € auf Anzahlungsforderung
- bei Schlussrechnung 300,00 € als bereits gezahlt/verrechnet
- offener Restbetrag: 700,00 €
```

## 12. Beispiel: Überzahlung

```text
Rechnung: 700,00 € offen
Zahlungseingang: 750,00 €
Allocation:
- 700,00 € auf Rechnung
- 50,00 € als Guthaben oder Rückzahlung
Status:
- invoice_paid
- customer_credit_balance: 50,00 €
```

## 13. Beispiel: Storno mit Anzahlung

```text
Bestellung: 1.000,00 €
Anzahlung: 300,00 €
Stornogebühr: 250,00 €
Allocation:
- 250,00 € der Anzahlung auf Stornogebühr
- 50,00 € Refund oder Guthaben
Status:
- order_cancelled
- cancellation_fee_paid
- refund_pending oder credit_available
```

## 14. Vermeidungsstrategien

- kein Boolean `paid=true` als alleinige Zahlungslogik
- keine Zahlung ohne Zahlungsquelle
- keine Zahlung ohne Zuordnungslogik
- keine Verrechnung ohne Audit-Log
- keine Rückzahlung ohne Referenz auf Zahlung und Beleg
- keine Anzahlung direkt in Rechnungszeile verstecken
- keine Änderung alter Zahlungseinträge, sondern Korrekturereignisse

---
