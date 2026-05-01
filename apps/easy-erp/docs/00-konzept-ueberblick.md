# Datei: `docs/00-konzept-ueberblick.md`

## 1. Zielbild

Die Anwendung soll als Go-Applikation in Andys bestehendem Infra-Stack betrieben werden:

```text
Traefik
  ↓
Go-App als Docker-Container
  ↓
SQLite als operative Datenbank
  ↓
Google Sheets als Pflege-/Sync-/Reporting-Schicht
  ↓
Google Docs/Drive als Dokumentenvorlagen- und Ablageschicht
  ↓
SMTP/Amazon SES für Versand
```

Die Anwendung ist kein klassisches Voll-ERP, sondern ein **fokussiertes Mini-ERP** für:

- Kundenverwaltung
- strukturierte Produkt-/Katalogauswahl
- Angebotserstellung
- Angebotsversionierung
- Bestellungserzeugung
- Rechnungserstellung
- Anzahlungen und Teilzahlungen
- Zahlungseingänge
- Storno und Rechnungskorrektur
- PDF-Ausgabe
- E-Rechnungsvorbereitung im XML-Datenmodell
- Google-Sheets-Synchronisierung
- Google-Docs-Vorlagen mit Platzhaltern
- Magic-Link-Login und Rollenmodell
- Audit-Log und Prozesshistorie

## 2. Architekturprinzipien

| Prinzip | Entscheidung |
|---|---|
| Operative Wahrheit | SQLite |
| Pflegeoberfläche für Stammdaten | Google Sheets |
| Dokumentenvorlagen | Google Docs |
| PDF-Ausgabe | Google Drive Export oder Go/HTML-to-PDF |
| E-Rechnung | strukturierter XML-Export aus Rechnungsdatenmodell |
| Rechnungsnummern | ausschließlich atomar in SQLite vergeben |
| Select-Box-Daten | aus SQLite-Cache, nicht live aus Google Sheets |
| Statuswechsel | ereignisbasiert mit Historie |
| Storno | nie löschen, immer Gegenprozess/Korrektur |
| Anzahlungen | eigene Zahlungslogik mit Zuordnung |
| Settings | Google Worksheet als Pflegeebene, SQLite als validierte Wahrheit |

## 3. Grundmodule

| Modul | Zweck | MVP-Relevanz |
|---|---|---|
| Auth | Magic Link, Sessions, Rollen | hoch |
| Settings | Nummernkreise, Fristen, AGB, E-Mail, E-Rechnung | hoch |
| Customers | Kunden, Kontakte, Adressen | hoch |
| Catalog | Kategorien, Hersteller, Produktgruppen, SKUs | hoch |
| Documents | Angebot, Bestellung, Lieferschein, Rechnung | hoch |
| Document Versions | Angebots- und Dokumentenversionierung | hoch |
| Payments | Anzahlungen, Teilzahlungen, Restzahlungen | hoch |
| Cancellation | Storno, Fristen, Gebühren | hoch |
| Corrections | Rechnungskorrektur, Gutschrift, Stornorechnung | hoch |
| Templates | Google-Docs-Vorlagen und Platzhalter | mittel/hoch |
| Files | PDF, XML, Ablage, Versandhistorie | hoch |
| E-Invoice | strukturiertes Rechnungsdatenmodell, XML | vorbereiten hoch, Generator mittel |
| Audit | Ereignis- und Änderungslog | hoch |
| Sync | Google-Sheets-Import und Export | hoch |

## 4. MVP-Schnitt

Der MVP sollte nicht zu klein gedacht werden, weil Storno, Anzahlungen und Rechnungsnummern sonst später teuer nachgerüstet werden müssten.

### MVP unbedingt enthalten

- Magic-Link-Login
- Rollen: Admin, Bearbeiter, Buchhaltung, Leser
- Google-Sheets-Sync für Settings und Katalogdaten
- SQLite als führende Datenbank
- Kundenverwaltung
- Produktkatalog mit 3-stufigem Select-Box-Staging
- Angebote mit Versionierung
- Umwandlung Angebot → Bestellung
- Rechnungserstellung aus Bestellung
- Anzahlungsanforderung und Zahlungserfassung
- Storno-Basislogik
- Rechnungskorrektur/Stornorechnung als Datenmodell
- PDF-Erzeugung
- E-Rechnungsdatenmodell vorbereitet
- Audit-Log

### MVP-nah vorbereiten

- XML-Export XRechnung/ZUGFeRD
- Validierungsbericht für E-Rechnung
- DATEV-/Buchhaltungs-Export
- Mahnwesen
- Rückzahlungen
- mehrstufige Freigaben

---
