# Storage & Lifecycle

## Drei Temperaturklassen

### HOT

-   z.  B. letzte 30 Tage
-   Header + Body lokal
-   kleine relevante Anhänge optional

### WARM

-   Header + Preview + Suchindex lokal
-   Body on demand
-   Attachment on demand

### COLD

-   Metadaten/Suchtreffer lokal
-   Body/Attachment aus durable store oder rekonstruierbarer Quelle

## AWS Temporary Layer

Geeignet für: - Sync Events - Bootstrap Snapshots - kurzlebige
verschlüsselte Mailobjekte - Zustell-/Verarbeitungsereignisse

Beispiel-TTLs: - Event delta: 14 Tage - Bootstrap snapshot: 7 Tage -
temporäres MIME: 1--7 Tage - Preview cache: 30 Tage

## Anhänge

Default: nicht auf jedes Gerät replizieren. - Metadaten sofort. - Datei
erst bei Öffnung. - optional `pin offline`. - SHA-256 für Integrität und
mögliche Deduplizierung.

## Größenmodell

Die entscheidenden Variablen für die spätere Simulation: - Mails/Tag -
durchschnittliche Header-/Preview-Größe - Body-Größe -
Attachment-Quote - Attachment-Mittelwert - Gerätezahl - HOT-Zeitraum -
Wiederholungs-/Retry-Faktor
