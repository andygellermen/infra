# Decisions & Ideas Log

## Fest

-   EasyPeasyMail Client + Server + SES Outbound.
-   Conscious Mode ist Kernidentität.
-   Local First.
-   AWS temporär für Distribution/Bootstrap.
-   Body/Attachments on demand.
-   Newsletter hart bündeln, aber nicht als Spam behandeln.
-   Keyboard-first.
-   Markdown Composer.
-   Keine Absenderbilder/Logos im Conscious Mode.

## Neu aufgenommen

-   drei UI-Modi: Conscious, Development, funktionsreicher Classic/Power
    („Burnout" als interner Arbeitstitel).
-   serverweite und lokale Spam-/Trust-Regeln.
-   Spamhaus/RBL als mögliches Signal.
-   Scheduled Send mit Quick Picker.
-   Offline-/Download-Status visuell minimal anzeigen.
-   Quick React als lokale Annotation mit explizitem
    Weitergabe-Workflow.
-   P2P erneut als Technology-Storming, nicht als MVP-Abhängigkeit.

## Zu prüfen

-   SES Inbound als Edge.
-   zweiter MX/zweite Inbound-Region.
-   JMAP als modernes Client-Protokoll.
-   WebRTC vs. QUIC für direkte Gerätesynchronisation.
-   Suchindex: SQLite FTS lokal, serverseitige Ergänzung optional.
-   Attachment Dedupe.
-   Privacy-/Tracking-Proxy für Remote Images.

## Produktregel

Jede neue Funktion muss beantworten: 1. Verbessert sie Kommunikation,
Resilienz oder bewusste Nutzung? 2. Kann sie im Conscious Mode
unsichtbar bleiben? 3. Welche neue Abhängigkeit erzeugt sie? 4. Was
passiert bei Ausfall? 5. Wie kann der Nutzer seine Daten wieder
herausbekommen?

## Moodivador / Print
- Moodivador bleibt Arbeitsname für Conscious, Development und BurnOut.
- BurnOut bleibt bewusst als kontraststarker Modus erhalten.
- Bildschirmmodus und Druckprofil werden entkoppelt; jeder Modus erhält eine Print Personality.

## Attachment Vault / EasyPeasyDrop
- Attachments grundsätzlich aus der aktiven Geräte-Replikation auslagern und über Referenzen laden.
- Content-addressed Attachment Vault als zentrale Idee.
- Große Outbound-Dateien optional als zeitlich begrenzter EasyPeasyDrop-Link statt MIME-Anhang.
- Original-Mail muss rekonstruierbar bleiben; keine irreversible Attachment-Extraktion ohne Retention-/Exportpfad.

## Verschlüsselung
- EasyPeasyMail-eigene Mail-/Content-Verschlüsselung ist aus Core und MVP gestrichen.
- Später ausschließlich als optionales Plugin/Capability prüfen.
- TLS, Auth, Secrets und Infrastruktur-Zugriffsschutz bleiben Security-Baseline.

## Mail-State und Drafts
- Read/Unread ist globaler Mailbox-State: auf einem Gerät gelesen -> nach Sync überall gelesen.
- Cache-/Offline-Zustände bleiben gerätespezifisch.
- Drafts sind geräteübergreifend bearbeitbar und offline editierbar.
- MVP nutzt optimistische Draft-Revisionen statt Echtzeit-CRDT; konkurrierender Text wird niemals still verworfen.
- Kausale Invarianten inspirieren Moodivador-Filterviews; Conscious zeigt nur Handlungsrelevantes, Development technische Zustände, BurnOut die Vollsicht.

## Trust & Operations
- Trust & Operations ist eine eigene Architekturschicht.
- Kein Silent Failure bei Empfang, Versand, Draft-Konflikten oder Attachment-Verfügbarkeit.
- Identity/Device Security, Abuse, Reputation, Delivery Observability, SLOs, Restore und Incident Response sind Pre-MVP-Themen.
- 100-Szenarien-Stresstest wird als Engineering-Prüfkatalog geführt.
