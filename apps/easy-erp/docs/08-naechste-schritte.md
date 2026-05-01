# Datei: `docs/08-naechste-schritte.md`

## 1. Sofort nächster Schritt

Vor dem Programmieren sollten diese Punkte final entschieden werden:

| Entscheidung | Empfehlung |
|---|---|
| ID-Format | ULID |
| Migration Tool | goose oder atlas |
| UI-Ansatz | serverseitige Go-Templates + HTMX oder schlichtes HTML/JS |
| PDF-Ansatz | zuerst HTML-to-PDF, später Google Docs optional |
| E-Rechnung | Datenmodell sofort, XML-Generator nach MVP-Grundfluss |
| Sync-Modus | manuell + später Scheduler |
| Auth | Magic Link + Rollen, später optional Google OAuth |

## 2. MVP-Backlog grob

### Epic 1: Projektbasis

- Go-Repo initialisieren
- Dockerfile erstellen
- SQLite einbinden
- Migrationen vorbereiten
- Healthcheck bauen
- ENV-Konfiguration definieren

### Epic 2: Auth

- User-Tabelle
- Magic-Link-Request
- Token-Hashing
- Session-Cookies
- Rollenprüfung

### Epic 3: Settings

- Settings-Worksheet definieren
- Google-Sheets-Client
- Validierung
- Sync-Run-Protokoll
- aktive Settings aus SQLite lesen

### Epic 4: Katalog

- Katalog-Worksheet definieren
- Kategorien importieren
- Hersteller importieren
- Produktgruppen importieren
- Produkte/SKUs importieren
- Select-Box-Endpunkte bauen

### Epic 5: Kunden

- Kundenformular
- Kontakte
- Adressen
- Kundentypen
- Suche

### Epic 6: Dokumente

- Angebot anlegen
- Positionen hinzufügen
- Summen berechnen
- Versionierung
- Angebot finalisieren
- PDF erstellen

### Epic 7: Bestellung/Rechnung

- Angebot in Bestellung umwandeln
- Bestellung in Rechnung umwandeln
- Rechnungsnummer atomar vergeben
- Rechnung finalisieren
- Rechnung versenden

### Epic 8: Zahlungen

- Zahlungsanforderung
- Anzahlung
- Zahlungseingang
- Payment Allocation
- Restbetrag berechnen

### Epic 9: Storno/Korrektur

- Storno anfordern
- Policy bewerten
- Gebühr berechnen
- Rückzahlung/Verrechnung vorbereiten
- Korrekturrechnung erzeugen

### Epic 10: E-Rechnung

- internes Rechnungsdatenmodell mappen
- XML-Modell vorbereiten
- Exportstatus speichern
- Validierungsbericht-Speicherung vorbereiten

## 3. Kritische fachliche Prüfstellen

- Stornogebühren steuerlich korrekt behandeln
- Anzahlung und Schlussrechnung sauber ausweisen
- Rechnungskorrektur sauber vom kaufmännischen Gutschriftfall trennen
- E-Rechnungspflicht nach Kundentyp und Vorgang bewerten
- rechtliche Texte versionieren
- alte Dokumente unveränderlich halten

## 4. Definition of Done für den MVP

Ein MVP ist fachlich belastbar, wenn:

- ein Kunde angelegt werden kann
- Produkte über 3-stufiges Select-Box-Staging gewählt werden können
- ein Angebot erstellt und finalisiert werden kann
- eine Bestellung aus dem Angebot erzeugt werden kann
- eine Rechnung aus der Bestellung erzeugt werden kann
- eine Anzahlung erfasst und verrechnet werden kann
- eine Bestellung storniert werden kann
- bei vorhandener Rechnung ein Korrekturfluss vorbereitet wird
- alle Belege nummeriert und historisiert sind
- PDF-Dateien erzeugt und archiviert werden
- E-Rechnungsdaten strukturiert vorhanden sind
- Settings und Katalogdaten aus Google Sheets synchronisiert werden
- Audit-Log kritische Aktionen abbildet

