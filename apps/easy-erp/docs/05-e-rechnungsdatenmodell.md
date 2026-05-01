# Datei: `docs/05-e-rechnungsdatenmodell.md`

## 1. Ziel

Die Anwendung soll schon im MVP ein strukturiertes Rechnungsdatenmodell besitzen, aus dem später PDF, XML/XRechnung und optional ZUGFeRD erzeugt werden können.

## 2. Grundprinzip

```text
SQLite-Rechnungsdatenmodell
  ↓
PDF-Darstellung
  ↓
XML-E-Rechnung
  ↓
Validierungsbericht
  ↓
Archiv
```

Der strukturierte Datensatz ist führend. PDF ist nur eine Darstellung.

## 3. Rechnungstypen

| Interner Typ | Zweck | E-Rechnungs-Bedeutung |
|---|---|---|
| invoice | normale Rechnung | Commercial Invoice |
| correction_invoice | korrigierte Rechnung/Stornorechnung | Corrected Invoice |
| credit_note | kaufmännische Gutschrift | Credit Note |
| deposit_invoice | Anzahlungsrechnung | je nach Profil gesondert prüfen |
| final_invoice | Schlussrechnung | normale Rechnung mit Verrechnung |
| cancellation_fee_invoice | Rechnung über Stornogebühr | normale Rechnung |

## 4. Pflichtnahe Datenbereiche

| Bereich | Benötigte Daten |
|---|---|
| Rechnungsidentifikation | Rechnungsnummer, Typ, Datum, Währung |
| Verkäufer | Name, Adresse, Land, Steuerdaten, Kontakt |
| Käufer | Name, Adresse, Land, ggf. Leitweg-ID/Referenz |
| Leistungsdaten | Leistungsdatum, Leistungszeitraum, Beschreibung |
| Positionen | SKU, Name, Beschreibung, Menge, Einheit, Preis, Rabatt, Steuer |
| Steuer | Steuersatz, Steuerkategorie, Steuerbetrag |
| Summen | Netto, Steuer, Brutto, bereits bezahlt, offen |
| Zahlungsbedingungen | Fälligkeit, IBAN, Verwendungszweck |
| Referenzen | Angebotsnummer, Bestellnummer, Originalrechnung, Lieferschein |
| Anhänge | PDF, AGB, Leistungsnachweis, Lieferschein |
| Validierung | Prüfstatus, Fehler, Warnungen, Bericht |

## 5. Invoice Header

| Feld | Beschreibung |
|---|---|
| invoice_id | interne ID |
| invoice_number | sichtbare Rechnungsnummer |
| invoice_type | invoice, correction_invoice, credit_note etc. |
| issue_date | Rechnungsdatum |
| due_date | Fälligkeitsdatum |
| currency | Währung |
| language | Sprache |
| seller_profile_id | Verkäuferprofil |
| customer_id | Kunde |
| billing_address_snapshot | eingefrorene Rechnungsanschrift |
| delivery_address_snapshot | eingefrorene Lieferanschrift |
| buyer_reference | Käuferreferenz, falls erforderlich |
| purchase_order_reference | Bestellreferenz des Kunden |
| preceding_invoice_id | Originalrechnung bei Korrektur |
| service_date | Leistungsdatum |
| service_period_start | Leistungszeitraum Beginn |
| service_period_end | Leistungszeitraum Ende |
| payment_term_key | Zahlungsbedingung |
| legal_text_key | AGB/Legal-Text-Version |
| e_invoice_profile_key | Profil für XML-Ausgabe |

## 6. Invoice Line

| Feld | Beschreibung |
|---|---|
| line_id | Positions-ID |
| invoice_id | Rechnung |
| position_no | Positionsnummer |
| sku_snapshot | SKU zum Zeitpunkt der Erstellung |
| product_name_snapshot | Produktname |
| description_snapshot | Beschreibung |
| quantity | Menge |
| unit_code | Einheit |
| unit_price_net | Einzelpreis netto |
| discount_amount | Rabattbetrag |
| discount_percent | Rabattprozent |
| surcharge_amount | Zuschlag |
| line_net_amount | Netto-Positionsbetrag |
| tax_rate_percent | Steuersatz |
| tax_category_code | Steuerkategorie |
| tax_amount | Steuerbetrag |
| line_gross_amount | Brutto-Positionsbetrag |
| service_date | optional positionsbezogenes Leistungsdatum |

## 7. Invoice Totals

| Feld | Beschreibung |
|---|---|
| total_line_net | Summe Positionen netto |
| total_allowances | Gesamtrabatte |
| total_charges | Zuschläge |
| total_tax_basis | Steuerbemessungsgrundlage |
| total_tax_amount | Steuerbetrag |
| total_gross | Bruttosumme |
| prepaid_amount | bereits gezahlt/Anzahlung |
| payable_amount | zahlbarer Betrag |
| rounding_amount | Rundungsbetrag |

## 8. Tax Breakdown

| Feld | Beschreibung |
|---|---|
| invoice_id | Rechnung |
| tax_rate_percent | Steuersatz |
| tax_category_code | Steuerkategorie |
| taxable_amount | Bemessungsgrundlage |
| tax_amount | Steuerbetrag |
| exemption_reason | Steuerbefreiungsgrund, falls nötig |
| exemption_reason_code | Code, falls nötig |

## 9. Document References

| Referenztyp | Beschreibung |
|---|---|
| quote | zugehöriges Angebot |
| order | zugehörige Bestellung |
| delivery_note | Lieferschein |
| preceding_invoice | Originalrechnung bei Korrektur |
| customer_purchase_order | Kundenbestellung |
| contract | Vertrag |
| payment_request | Zahlungsanforderung |

## 10. E-Invoice Export Status

| Status | Bedeutung |
|---|---|
| not_required | nicht erforderlich |
| pending | vorbereitet |
| generated | XML erzeugt |
| validation_pending | Validierung ausstehend |
| valid | valide |
| warning | valide mit Warnungen |
| invalid | ungültig |
| sent | versendet |
| archived | archiviert |

## 11. Validierungsdaten

| Feld | Beschreibung |
|---|---|
| validation_id | ID |
| invoice_id | Rechnung |
| format | xrechnung, zugferd |
| validator_name | verwendetes Tool |
| validation_status | valid, warning, invalid |
| validation_report_file_id | Bericht |
| error_count | Fehleranzahl |
| warning_count | Warnungsanzahl |
| validated_at | Zeitpunkt |

## 12. Vermeidungsstrategien

- E-Rechnung nicht aus PDF rekonstruieren
- PDF und XML immer aus demselben strukturierten Datenmodell erzeugen
- Rechnungskorrektur immer mit Referenz auf Originalrechnung
- Anzahlungen strukturiert erfassen und verrechnen
- Pflichtangaben nicht nur in Anhänge auslagern
- Validierungsbericht speichern
- alte Rechnungsdaten niemals durch Produkt-/Setting-Änderungen überschreiben

---
