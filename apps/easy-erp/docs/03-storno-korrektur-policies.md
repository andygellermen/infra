# Datei: `docs/03-storno-korrektur-policies.md`

## 1. Ziel

Das Storno-/Korrekturmodul bildet fachliche Gegenprozesse ab, ohne operative oder buchhalterische Vorgänge zu löschen. Der Ablauf unterscheidet zwischen:

- Angebot verwerfen
- Bestellung stornieren
- Rechnung korrigieren/stornieren
- Anzahlung verrechnen
- Rückzahlung auslösen
- Stornogebühr abrechnen
- Gutschrift/Korrekturbeleg erzeugen

## 2. Grundprinzip

```text
Storno ist kein Löschen.
Storno ist ein neues Ereignis mit eigener Dokumentation.
```

## 3. Fachliche Fälle

| Fall | Ausgangslage | Ergebnis |
|---|---|---|
| Angebot abgelehnt | keine Bestellung, keine Rechnung | Status `quote_declined` |
| Bestellung storniert ohne Rechnung | Bestellung aktiv, keine Rechnung | `cancellation_event`, optional Stornogebühr-Rechnung |
| Bestellung storniert mit Anzahlung | Zahlung vorhanden | Verrechnung, Rückzahlung oder Guthaben |
| Rechnung bereits erstellt | Rechnung final | Korrektur-/Stornorechnung mit Referenz |
| Rechnung teilweise bezahlt | Zahlung vorhanden | Korrektur + Payment Allocation + ggf. Refund |
| Rechnung vollständig bezahlt | Zahlung vollständig | Korrektur + Rückzahlung/Gutschrift |
| Teilstorno | nur einzelne Positionen betroffen | positionsbezogene Korrektur |

## 4. Statusmodell für Storno

```text
cancellation_requested
  ↓
cancellation_policy_evaluated
  ↓
cancellation_fee_calculated
  ↓
requires_manual_review? ── yes → cancellation_review_required
  ↓ no
cancellation_approved
  ↓
correction_required? ── yes → correction_document_created
  ↓ no
refund_required? ── yes → refund_pending → refund_completed
  ↓ no
cancellation_completed
```

## 5. Policy-Tabelle

| Feld | Typ | Bedeutung |
|---|---|---|
| policy_key | string | eindeutiger Schlüssel |
| label | string | sprechender Name |
| applies_to | enum | order, invoice, service, product |
| customer_type | enum | private, business, public, any |
| product_category_key | string/null | optional auf Kategorie beschränkt |
| time_reference | enum | order_date, service_start, delivery_date, invoice_date |
| threshold_from_hours | int | Beginn des Zeitfensters |
| threshold_to_hours | int | Ende des Zeitfensters |
| fee_mode | enum | none, fixed, percent |
| fee_value | decimal | Betrag oder Prozentsatz |
| fee_basis | enum | order_net, order_gross, open_amount, deposit_amount, item_amount |
| min_fee | decimal | Mindestgebühr |
| max_fee | decimal/null | Maximalgebühr |
| tax_rate_key | string | Steuersatz für Gebühr |
| requires_manual_review | bool | manuelle Prüfung erforderlich |
| active | bool | aktiv |
| valid_from | date | gültig ab |
| valid_to | date/null | gültig bis |

## 6. Beispiel-Policies

| policy_key | label | threshold_from_hours | threshold_to_hours | fee_mode | fee_value | fee_basis | requires_manual_review |
|---|---|---:|---:|---|---:|---|---|
| cancel_free_14d | kostenlos bis 14 Tage vorher | 336 | 999999 | none | 0 | order_gross | FALSE |
| cancel_25_7d | 25 Prozent bis 7 Tage vorher | 168 | 335 | percent | 25 | order_gross | FALSE |
| cancel_50_48h | 50 Prozent bis 48 Stunden vorher | 48 | 167 | percent | 50 | order_gross | FALSE |
| cancel_100_late | 100 Prozent unter 48 Stunden | 0 | 47 | percent | 100 | order_gross | TRUE |

## 7. Korrekturtypen

| Typ | Beschreibung | Verwendung |
|---|---|---|
| corrected_invoice | Rechnungskorrektur/Stornorechnung | Korrektur einer bereits erstellten Rechnung |
| credit_note | kaufmännische Gutschrift | eigenständige Gutschrift, nicht zwingend Rechnungskorrektur |
| refund | Rückzahlung | Geldfluss zurück an Kunden |
| cancellation_fee_invoice | Rechnung über Stornogebühr | wenn keine Rechnung vorhanden, aber Gebühr anfällt |
| partial_correction | Teilkorrektur | einzelne Positionen werden korrigiert |

## 8. Storno-Entscheidungslogik

```text
Input:
- target_document_id
- cancellation_reason
- cancellation_date
- service_start/delivery_date
- customer_type
- payment_status
- invoice_status

Ablauf:
1. Zielbeleg laden.
2. passende Storno-Policy ermitteln.
3. Zeitdifferenz zur Referenz berechnen.
4. Gebühr berechnen.
5. prüfen, ob Rechnung existiert.
6. prüfen, ob Zahlung/Anzahlung existiert.
7. Entscheidung erzeugen:
   - nur Statuswechsel
   - Stornogebühr-Rechnung
   - Korrekturrechnung
   - Rückzahlung
   - Verrechnung
   - manuelle Prüfung
8. Audit-Log schreiben.
```

## 9. Berechnungsbeispiele

### Beispiel A: Bestellung ohne Rechnung, keine Zahlung

```text
Bestellung: 1.000,00 € brutto
Storno: 10 Tage vor Leistung
Policy: 25 Prozent
Ergebnis:
- Bestellung wird storniert
- Stornogebühr: 250,00 € brutto
- neue Stornogebühr-Rechnung wird erstellt
```

### Beispiel B: Bestellung mit 300,00 € Anzahlung

```text
Bestellung: 1.000,00 € brutto
Anzahlung: 300,00 €
Storno: 10 Tage vor Leistung
Stornogebühr: 250,00 €
Ergebnis:
- 250,00 € werden mit Anzahlung verrechnet
- 50,00 € werden erstattet oder als Guthaben geführt
```

### Beispiel C: Rechnung bereits erstellt und bezahlt

```text
Rechnung: 1.000,00 € brutto
Zahlung: 1.000,00 €
Storno: vollständig
Stornogebühr: 250,00 €
Ergebnis:
- Rechnungskorrektur über -1.000,00 € oder positionsbezogene Korrektur
- neue Rechnung/Stornogebühr über 250,00 €, falls getrennt geführt
- Rückzahlung/Guthaben: 750,00 €
```

## 10. Vermeidungsstrategien

- keine Storno-Logik über Löschen
- keine Storno-Logik als Freitextfeld
- keine Rechnungskorrektur ohne Referenz auf Originalrechnung
- keine Rückzahlung ohne Payment Allocation
- keine Stornogebühr ohne steuerliche Einordnung
- keine rückwirkende Veränderung des Originaldokuments
- keine manuelle Korrektur von Rechnungsnummern

---
