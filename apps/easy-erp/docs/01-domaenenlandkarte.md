# Datei: `docs/01-domaenenlandkarte.md`

## 1. Domänenlandkarte

```text
Easy ERP @ Google-Steroids

├── Identity & Access
│   ├── User
│   ├── Role
│   ├── Permission
│   ├── Magic Link Token
│   └── Session
│
├── Settings & Policies
│   ├── Company Profile
│   ├── Number Range
│   ├── Tax Rate
│   ├── Payment Terms
│   ├── Cancellation Policy
│   ├── E-Mail Account
│   ├── E-Mail Template
│   ├── Document Template
│   ├── E-Invoice Profile
│   ├── Legal Text / AGB
│   └── Feature Flag
│
├── Customers
│   ├── Customer
│   ├── Contact
│   ├── Address
│   ├── Customer Type
│   └── Customer Notes
│
├── Catalog
│   ├── Catalog Category
│   ├── Manufacturer
│   ├── Product Group
│   ├── Product
│   ├── SKU
│   ├── Price Rule
│   ├── Stock Rule
│   └── Product Option
│
├── Documents
│   ├── Quote
│   ├── Order
│   ├── Delivery Note
│   ├── Invoice
│   ├── Correction Invoice
│   ├── Credit Note
│   ├── Document Item
│   ├── Document Version
│   ├── Document Reference
│   └── Document Status Event
│
├── Payments
│   ├── Payment Request
│   ├── Payment
│   ├── Payment Allocation
│   ├── Refund
│   └── Payment Status Event
│
├── Cancellation & Correction
│   ├── Cancellation Event
│   ├── Cancellation Fee
│   ├── Cancellation Decision
│   ├── Correction Document
│   ├── Refund Decision
│   └── Cancellation Audit
│
├── Output & Communication
│   ├── Generated File
│   ├── PDF Output
│   ├── XML Output
│   ├── Google Doc Copy
│   ├── Drive Folder
│   ├── Mail Dispatch
│   └── Dispatch Event
│
├── Sync
│   ├── Sheet Source
│   ├── Sheet Sync Run
│   ├── Sheet Sync Error
│   ├── Sheet Row Mapping
│   └── Catalog Cache
│
└── Audit
    ├── Audit Log
    ├── Business Event
    ├── Value Change
    └── Process Trace
```

## 2. Kernentitäten

| Entität | Beschreibung | Quelle/Führung |
|---|---|---|
| User | App-Benutzer | SQLite |
| Role | Rollenmodell | SQLite/Settings-Sync |
| Setting | validierte Systemeinstellung | Google Sheet → SQLite |
| NumberRange | Nummernkreise für Belege | Settings → SQLite, Vergabe nur SQLite |
| Customer | Kunde | SQLite, optional Sheet-Export |
| Contact | Ansprechpartner | SQLite |
| Address | Rechnungs-/Lieferanschrift | SQLite |
| CatalogCategory | Hauptkategorie | Google Sheet → SQLite Cache |
| Manufacturer | Hersteller | Google Sheet → SQLite Cache |
| ProductGroup | Produktgruppe | Google Sheet → SQLite Cache |
| Product | Produkt/SKU | Google Sheet → SQLite Cache |
| Document | Angebot/Bestellung/Rechnung etc. | SQLite |
| DocumentItem | Belegposition mit Snapshot | SQLite |
| DocumentVersion | Version eines Angebots/Dokuments | SQLite |
| PaymentRequest | Zahlungsanforderung | SQLite |
| Payment | Zahlungseingang | SQLite |
| PaymentAllocation | Zuordnung Zahlung zu Beleg/Forderung | SQLite |
| CancellationPolicy | Storno-Regelwerk | Google Sheet → SQLite |
| CancellationEvent | konkretes Storno-Ereignis | SQLite |
| CorrectionDocument | Korrektur-/Stornorechnung | SQLite |
| GeneratedFile | PDF/XML/Google-Doc-Datei | SQLite + Drive |
| AuditLog | Änderungs- und Ereignishistorie | SQLite |

## 3. Bounded Contexts

### 3.1 Identity & Access

Zweck:

- Magic-Link-Login
- Sitzungsverwaltung
- Rollen und Berechtigungen
- Schutz sensibler Aktionen

Wichtige Regeln:

- Magic Links sind einmalig verwendbar.
- Tokens werden nur gehasht gespeichert.
- sensible Aktionen benötigen passende Berechtigung.
- Admin-Einstellungen sind auditpflichtig.

### 3.2 Settings & Policies

Zweck:

- zentrale Pflege von steuerlichen, buchhalterischen und rechtlichen Einstellungen
- Synchronisation aus einem separaten Settings-Worksheet
- Versionierung von rechtlich relevanten Einstellungen

Wichtige Regeln:

- Settings aus Google Sheets werden validiert.
- Operative Prozesse lesen nur aus SQLite.
- alte Dokumente behalten die zum Erstellzeitpunkt gültige Setting-Version.
- Rechnungsnummern werden nie in Google Sheets vergeben.

### 3.3 Customers

Zweck:

- Kunden und deren Adressen/Kontakte verwalten
- Unterscheidung von Privatkunde, Geschäftskunde, öffentlicher Stelle, Partner

Wichtige Regeln:

- Rechnungsanschrift wird bei Belegerstellung eingefroren.
- Lieferanschrift kann abweichen.
- E-Rechnungspflicht hängt u. a. vom Kundentyp und Land ab.

### 3.4 Catalog

Zweck:

- mehrstufige Produkt-/SKU-Auswahl
- Katalogpflege über Google Sheets
- schnelle Select-Boxen aus SQLite-Cache

Wichtige Regeln:

- Produktdaten werden vor Nutzung synchronisiert.
- Auswahl erfolgt nie live gegen Google Sheets.
- Belegpositionen speichern Snapshots von Name, Preis, Steuer, Beschreibung.

### 3.5 Documents

Zweck:

- Angebot, Bestellung, Lieferschein, Rechnung, Korrekturbelege
- Dokumentenstatus und Statushistorie
- Dokumentenreferenzen

Wichtige Regeln:

- Dokumente werden nicht gelöscht, sondern storniert/korrigiert.
- Umwandlungen erzeugen neue Instanzen.
- historische Positionen bleiben unverändert.

### 3.6 Payments

Zweck:

- Anzahlungen
- Teilzahlungen
- Restzahlungen
- Rückzahlungen
- Zahlungszuordnung

Wichtige Regeln:

- Zahlungseingang ist eigenes Objekt.
- Zahlung kann mehreren Forderungen zugeordnet werden.
- Anzahlungen werden bei Schlussrechnung verrechnet.

### 3.7 Cancellation & Correction

Zweck:

- Stornoanfragen
- Stornofristen
- Stornogebühren
- Korrektur-/Gutschriftfluss

Wichtige Regeln:

- Storno ist ein Ereignis, kein Löschen.
- Rechnungskorrektur referenziert Originalrechnung.
- Zahlungen werden verrechnet, zurückgezahlt oder als Guthaben geführt.

### 3.8 Output & Communication

Zweck:

- PDF und XML erzeugen
- Google Docs kopieren
- Dokumente per E-Mail versenden
- Versandhistorie speichern

Wichtige Regeln:

- versendete Dokumente werden archiviert.
- Empfänger, Zeitpunkt und Datei-Hash werden gespeichert.
- XML und PDF entstehen aus demselben strukturierten Datensatz.

### 3.9 Sync

Zweck:

- Google Sheets importieren
- Settings validieren
- Katalogdaten cachen
- Fehler protokollieren

Wichtige Regeln:

- Sync ist nachvollziehbar.
- Fehlerhafte Zeilen werden nicht stillschweigend übernommen.
- letzte gültige Konfiguration bleibt aktiv, wenn neuer Sync fehlschlägt.

### 3.10 Audit

Zweck:

- Nachvollziehbarkeit aller relevanten Ereignisse
- Änderungsprotokoll
- Prozesshistorie

Wichtige Regeln:

- Rechnungen, Zahlungen, Stornos, Settings und Versand sind auditpflichtig.
- alte und neue Werte werden bei kritischen Änderungen gespeichert.

---
