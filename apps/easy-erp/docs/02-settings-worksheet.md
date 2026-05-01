# Datei: `docs/02-settings-worksheet.md`

## 1. Grundidee

Das Settings-Worksheet ist die pflegefreundliche Oberfläche für fachliche, rechtliche und technische Einstellungen. Die Go-App synchronisiert diese Daten, validiert sie und schreibt sie versioniert in SQLite.

```text
Google Settings Worksheet
  ↓
Sync Job
  ↓
Validation Layer
  ↓
SQLite Settings Tables
  ↓
ERP-Prozesse
```

## 2. Grundregeln

| Regel | Bedeutung |
|---|---|
| Google Sheet ist Pflegeebene | Menschen können Einstellungen einfach ändern |
| SQLite ist Prozesswahrheit | App nutzt nur validierte SQLite-Daten |
| Settings sind versioniert | alte Belege bleiben stabil |
| Gültigkeit ist datiert | `valid_from`, `valid_to` |
| Sync ist auditpflichtig | Änderungen werden protokolliert |
| Nummernvergabe nur in SQLite | Sheet enthält nur Muster und Konfiguration |
| fehlerhafte Settings blockieren Sync | letzte gültige Version bleibt aktiv |

## 3. Empfohlene Worksheet-Struktur

```text
ERP_Settings
├── 01_general
├── 02_company_profile
├── 03_number_ranges
├── 04_tax_rates
├── 05_payment_terms
├── 06_cancellation_policies
├── 07_email_accounts
├── 08_email_templates
├── 09_document_templates
├── 10_e_invoice_profiles
├── 11_legal_texts_agb
├── 12_drive_folders
├── 13_permissions
├── 14_feature_flags
└── 99_sync_control
```

## 4. `01_general`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| key | string | default_currency | technischer Schlüssel |
| value | string | EUR | Wert |
| value_type | enum | string | string, int, decimal, bool, json, date |
| description | text | Standardwährung | sprechende Beschreibung |
| valid_from | date | 2026-01-01 | gültig ab |
| valid_to | date/null |  | gültig bis |
| active | bool | TRUE | aktiv/inaktiv |
| version | string | 2026.1 | Version |

Beispiele:

| key | value | value_type | description | valid_from | active |
|---|---|---|---|---|---|
| default_currency | EUR | string | Standardwährung | 2026-01-01 | TRUE |
| default_locale | de-DE | string | Standardsprache | 2026-01-01 | TRUE |
| default_timezone | Europe/Berlin | string | Standard-Zeitzone | 2026-01-01 | TRUE |
| default_vat_mode | net | string | Preise netto/brutto | 2026-01-01 | TRUE |

## 5. `02_company_profile`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| profile_id | string | default | Profil-ID |
| company_name | string | Muster GmbH | Firmenname |
| street | string | Hauptstraße 1 | Straße |
| postal_code | string | 10115 | PLZ |
| city | string | Berlin | Ort |
| country_code | string | DE | Land |
| tax_number | string |  | Steuernummer |
| vat_id | string | DE123456789 | USt-ID |
| email | string | rechnung@example.de | allgemeine E-Mail |
| phone | string |  | Telefon |
| website | string |  | Website |
| iban | string |  | IBAN |
| bic | string |  | BIC |
| bank_name | string |  | Bank |
| legal_footer | text |  | Fußzeile |
| valid_from | date | 2026-01-01 | gültig ab |
| active | bool | TRUE | aktiv |

## 6. `03_number_ranges`

Wichtig: Dieses Blatt konfiguriert Nummernkreise. Es vergibt keine Nummern.

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| range_key | string | invoice_default | eindeutiger Schlüssel |
| document_type | enum | invoice | quote, order, delivery_note, invoice, correction, credit_note, payment |
| prefix | string | R | Präfix |
| year_mode | enum | calendar_year | calendar_year, fiscal_year, none |
| pattern | string | R-{YYYY}-{NNNNN} | Formatmuster |
| padding | int | 5 | Stellenzahl |
| reset_policy | enum | yearly | yearly, monthly, never |
| start_number | int | 1 | Startwert |
| current_seed | int | 1 | nur Initialwert, App vergibt atomar weiter |
| valid_from | date | 2026-01-01 | gültig ab |
| active | bool | TRUE | aktiv |

Beispiele:

| range_key | document_type | prefix | pattern | padding | reset_policy | start_number | active |
|---|---|---|---|---:|---|---:|---|
| quote_default | quote | A | A-{YYYY}-{NNNNN} | 5 | yearly | 1 | TRUE |
| order_default | order | B | B-{YYYY}-{NNNNN} | 5 | yearly | 1 | TRUE |
| invoice_default | invoice | R | R-{YYYY}-{NNNNN} | 5 | yearly | 1 | TRUE |
| correction_default | correction | K | K-{YYYY}-{NNNNN} | 5 | yearly | 1 | TRUE |
| credit_note_default | credit_note | G | G-{YYYY}-{NNNNN} | 5 | yearly | 1 | TRUE |

## 7. `04_tax_rates`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| tax_rate_key | string | vat_19 | Schlüssel |
| country_code | string | DE | Land |
| tax_name | string | Umsatzsteuer | Name |
| rate_percent | decimal | 19.00 | Steuersatz |
| category_code | string | S | Steuerkategorie |
| description | text | Regelsteuersatz | Beschreibung |
| valid_from | date | 2026-01-01 | gültig ab |
| valid_to | date/null |  | gültig bis |
| active | bool | TRUE | aktiv |

Beispiele:

| tax_rate_key | country_code | tax_name | rate_percent | category_code | description | active |
|---|---|---|---:|---|---|---|
| vat_19 | DE | Umsatzsteuer | 19.00 | S | Regelsteuersatz | TRUE |
| vat_7 | DE | Umsatzsteuer | 7.00 | AA | ermäßigter Steuersatz | TRUE |
| vat_0 | DE | Umsatzsteuer | 0.00 | Z | steuerfrei/null | TRUE |

## 8. `05_payment_terms`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| payment_term_key | string | due_14_days | Schlüssel |
| label | string | 14 Tage netto | Anzeige |
| due_days | int | 14 | Fälligkeit in Tagen |
| deposit_required | bool | TRUE | Anzahlung erforderlich |
| deposit_mode | enum | percent | percent, fixed, none |
| deposit_value | decimal | 30.00 | Prozent oder Betrag |
| final_due_days | int | 7 | Restbetrag fällig nach Rechnung |
| reminder_enabled | bool | TRUE | Mahnwesen vorbereiten |
| text_block | text | Zahlbar innerhalb... | Text für Beleg |
| active | bool | TRUE | aktiv |

Beispiele:

| payment_term_key | label | due_days | deposit_required | deposit_mode | deposit_value | final_due_days | active |
|---|---|---:|---|---|---:|---:|---|
| due_14_days | 14 Tage netto | 14 | FALSE | none | 0 | 14 | TRUE |
| deposit_30_rest_7 | 30 Prozent Anzahlung, Rest 7 Tage | 7 | TRUE | percent | 30 | 7 | TRUE |
| prepaid | Vorkasse | 0 | TRUE | percent | 100 | 0 | TRUE |

## 9. `06_cancellation_policies`

Dieses Blatt wird zusätzlich in `docs/03-storno-korrektur-policies.md` fachlich vertieft.

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| policy_key | string | default_service_cancel | Schlüssel |
| label | string | Standard-Storno | Anzeige |
| applies_to | enum | order | quote, order, invoice, service, product |
| customer_type | enum/null | business | private, business, public, any |
| time_reference | enum | service_start | order_date, service_start, delivery_date |
| threshold_unit | enum | hours | hours, days |
| threshold_from | int | 168 | Grenze von |
| threshold_to | int | 999999 | Grenze bis |
| fee_mode | enum | percent | none, fixed, percent |
| fee_value | decimal | 25.00 | Wert |
| fee_basis | enum | order_gross | order_net, order_gross, open_amount, deposit_amount |
| min_fee | decimal | 0.00 | Mindestgebühr |
| max_fee | decimal/null |  | Höchstgebühr |
| tax_rate_key | string | vat_19 | Steuer |
| active | bool | TRUE | aktiv |

## 10. `07_email_accounts`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| account_key | string | billing_ses | Schlüssel |
| sender_name | string | Muster GmbH | Absendername |
| sender_email | string | rechnung@example.de | Absenderadresse |
| reply_to | string | service@example.de | Antwortadresse |
| smtp_profile_ref | string | ses_eu | Referenz auf Secret/ENV, nicht Klartext |
| purpose | enum | billing | billing, quote, support, general |
| active | bool | TRUE | aktiv |

Wichtig: SMTP-Passwörter gehören nicht in Google Sheets. Das Sheet enthält nur Referenzen auf ENV/Secrets.

## 11. `08_email_templates`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| template_key | string | quote_send_default | Schlüssel |
| purpose | enum | quote_send | quote_send, invoice_send, payment_reminder, cancellation |
| subject_template | string | Ihr Angebot {{document_number}} | Betreff |
| body_template | text | Guten Tag... | Mailtext |
| attach_pdf | bool | TRUE | PDF anhängen |
| attach_xml | bool | FALSE | XML anhängen |
| active | bool | TRUE | aktiv |
| valid_from | date | 2026-01-01 | gültig ab |

## 12. `09_document_templates`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| template_key | string | invoice_default | Schlüssel |
| document_type | enum | invoice | quote, order, delivery_note, invoice, correction |
| google_doc_template_id | string | abc123 | Vorlage |
| output_format | enum | pdf | pdf, docx, html |
| language | string | de-DE | Sprache |
| legal_text_key | string | agb_2026_01 | AGB/Legal Text |
| active | bool | TRUE | aktiv |
| valid_from | date | 2026-01-01 | gültig ab |

## 13. `10_e_invoice_profiles`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| profile_key | string | xrechnung_default | Schlüssel |
| format | enum | xrechnung | xrechnung, zugferd, none |
| syntax | string | urn:cen.eu:en16931... | Syntaxprofil |
| seller_endpoint_id | string |  | Leitweg-ID oder Endpoint |
| seller_endpoint_scheme | string |  | Schema |
| buyer_reference_required | bool | FALSE | Käuferreferenz Pflicht |
| attach_pdf | bool | TRUE | PDF zusätzlich erzeugen |
| validate_before_send | bool | TRUE | Validierung vor Versand |
| save_validation_report | bool | TRUE | Bericht speichern |
| active | bool | TRUE | aktiv |

## 14. `11_legal_texts_agb`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| legal_text_key | string | agb_2026_01 | Schlüssel |
| title | string | AGB Januar 2026 | Titel |
| type | enum | agb | agb, privacy, cancellation_terms, payment_terms |
| text | text | ... | Rechtstext |
| version | string | 2026-01 | Version |
| valid_from | date | 2026-01-01 | gültig ab |
| valid_to | date/null |  | gültig bis |
| active | bool | TRUE | aktiv |

Wichtige Regel:

- Ein erzeugtes Angebot oder eine Rechnung referenziert die verwendete `legal_text_key` und Version.
- Änderungen an AGB dürfen alte Dokumente nicht nachträglich verändern.

## 15. `12_drive_folders`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| folder_key | string | invoices_2026 | Schlüssel |
| purpose | enum | invoices | quotes, invoices, corrections, xml, archive |
| google_drive_folder_id | string | abc123 | Zielordner |
| year | int/null | 2026 | Jahr |
| active | bool | TRUE | aktiv |

## 16. `13_permissions`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| role_key | string | accounting | Rolle |
| permission_key | string | invoice.finalize | Berechtigung |
| allowed | bool | TRUE | erlaubt |
| description | text | Rechnung finalisieren | Beschreibung |
| active | bool | TRUE | aktiv |

Beispiele für Berechtigungen:

| permission_key | Beschreibung |
|---|---|
| customer.read | Kunden lesen |
| customer.write | Kunden bearbeiten |
| quote.create | Angebot erstellen |
| quote.send | Angebot versenden |
| order.create | Bestellung erstellen |
| invoice.create | Rechnung erstellen |
| invoice.finalize | Rechnung finalisieren |
| invoice.correct | Rechnung korrigieren |
| payment.record | Zahlung erfassen |
| cancellation.create | Storno auslösen |
| settings.sync | Settings synchronisieren |

## 17. `14_feature_flags`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| flag_key | string | enable_e_invoice_xml | Feature |
| enabled | bool | FALSE | aktiv/inaktiv |
| description | text | XML-Export aktivieren | Beschreibung |
| valid_from | date | 2026-01-01 | gültig ab |

## 18. `99_sync_control`

| Spalte | Typ | Beispiel | Beschreibung |
|---|---|---|---|
| sync_group | string | settings | Gruppe |
| enabled | bool | TRUE | Sync aktiv |
| mode | enum | manual | manual, scheduled |
| interval_minutes | int | 60 | Intervall |
| last_reviewed_by | string | Andy | fachlich geprüft von |
| comment | text |  | Hinweis |

---
