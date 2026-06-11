# easy-author – Domain Map

## 1. Übersicht

Die Domänenlandkarte beschreibt die fachlichen Kernbereiche von easy-author.

```text
Authoring
├── Projects
├── Manuscripts
├── Editor
├── Knowledge Base
├── Assets
├── Export
├── Reader Rooms
├── Comments & Reviews
├── Planning
├── Monetization
├── Cloud Sync
└── Administration
```

## 2. Domäne: Projects

Verwaltet Buch- und Schreibprojekte.

### Verantwortlichkeiten

- Projekt anlegen
- Projekt-Metadaten pflegen
- Projektstatus verwalten
- Projekt exportieren/sichern

### Objekte

- Project
- ProjectSettings
- ProjectMember

## 3. Domäne: Manuscripts

Verwaltet Buchstruktur und Inhalt.

### Verantwortlichkeiten

- Kapitel verwalten
- Reihenfolge definieren
- Fragmente und Szenen speichern
- Manuskript zusammenführen

### Objekte

- Manuscript
- Chapter
- Scene
- Fragment
- OutlineNode

## 4. Domäne: Editor

Stellt die Schreibumgebung bereit.

### Verantwortlichkeiten

- Typora-artiges Schreiben
- Markdown speichern
- Editor-State verwalten
- Autosave
- Verlinkungen erkennen
- Seitenleisten aktualisieren

### Objekte

- EditorDocument
- EditorSession
- AutosaveSnapshot

## 5. Domäne: Knowledge Base

Verwaltet Autorenwissen über das gesamte Projekt.

### Verantwortlichkeiten

- Personen, Orte, Ereignisse, Motive speichern
- Verlinkungen zu Kapiteln erkennen
- Kontext in Sidebar anzeigen
- Inkonsistenzen vorbereiten

### Objekte

- Person
- Location
- Event
- StoryThread
- Motif
- Reminder
- ResearchNote

## 6. Domäne: Assets

Verwaltet Bilder und andere Medien.

### Verantwortlichkeiten

- Upload
- Metadaten
- Lizenzdaten
- Alt-Texte
- Verwendung in Kapiteln
- Exportfreigabe

### Objekte

- Asset
- AssetLicense
- AssetUsage
- AssetVariant

## 7. Domäne: Export

Erzeugt Ausgabeformate.

### Verantwortlichkeiten

- Markdown zusammenführen
- Themes anwenden
- PDF/EPUB/DOCX erzeugen
- Export-Historie führen
- Qualität prüfen

### Objekte

- ExportProfile
- ExportJob
- ExportArtifact
- Theme

## 8. Domäne: Reader Rooms

Stellt Bücher oder Kapitel zum Lesen bereit.

### Verantwortlichkeiten

- Sichtbarkeit steuern
- Leserzugänge verwalten
- Leseproben bereitstellen
- Magic Links erzeugen

### Objekte

- ReaderRoom
- ReaderAccess
- ReadingSession

## 9. Domäne: Comments & Reviews

Ermöglicht Feedback.

### Verantwortlichkeiten

- Kommentare
- Rezensionen
- Korrekturhinweise
- Feedback-Modi
- Freigaben

### Objekte

- Comment
- Review
- FeedbackRequest
- ReviewGroup

## 10. Domäne: Planning

Unterstützt Autorenplanung.

### Verantwortlichkeiten

- Termine
- Schreibziele
- Deadlines
- Meilensteine
- Erinnerungen

### Objekte

- WritingGoal
- Milestone
- PlannerEvent

## 11. Domäne: Monetization

Bereitet bezahlte Nutzung und Buchmonetarisierung vor.

### Verantwortlichkeiten

- Spenden
- Direktverkauf
- Zugangskäufe
- Autoren-Abos

### Objekte

- DonationCampaign
- Product
- Purchase
- SubscriptionPlan

## 12. Domäne: Cloud Sync

Synchronisiert Projektdateien mit externen Speichern.

### Verantwortlichkeiten

- WebDAV/Rclone-Konfiguration
- Sync-Jobs
- Konfliktanzeige
- Backup

### Objekte

- SyncTarget
- SyncJob
- SyncConflict

## 13. Domäne: Administration

Betreibt Benutzer, Rollen, Mandanten und Systemkonfiguration.

### Verantwortlichkeiten

- User Management
- Rollen
- Mandanten
- Systemstatus
- Storage-Limits

### Objekte

- User
- Role
- Tenant
- SystemSetting

## Ergänzte Domänen aus UX-Spike 02

Neue beziehungsweise konkretisierte Domänen:

- Dashboard und Projektstart
- Werktypen und Werkstatus
- Werkstruktur mit Front Matter, Body, Back Matter und Fragmenten
- Kapiteloptionen und Kapitelstatus
- wiederverwendbare Buchseiten
- Theme-Grundmodell
- Proofing und Exportvorbereitung

