# Attachment Vault --- EasyPeasyDrop

## Entscheidungsidee

Anhänge werden grundsätzlich aus der aktiven Mail-Replikation
ausgelagert. Die Mail enthält danach nur noch eine stabile Attachment
Reference.

``` text
Mail
 ├─ Header
 ├─ Body
 └─ AttachmentRef[]
       |
       v
  EasyPeasyDrop / Attachment Vault
```

## Warum

Attachments sind der größte Speicher- und Traffic-Treiber. Ein eigener
Vault erlaubt: - Lazy Download - lokale Cache-Regeln - Deduplizierung -
Malware-Quarantäne - unabhängige Retention - Streaming statt kompletter
Mail-Replikation - sichere Freigabelinks für große Outbound-Dateien

## Content-addressed Storage

Jede Datei erhält einen Hash:

``` text
sha256(file) -> object_id
```

Die Mail speichert:

``` text
attachment_id
object_id
filename
mime_type
size
sha256
disposition
content_id?
availability
```

Mehrfach vorkommende identische Dateien können innerhalb eines
Accounts/Tenants nur einmal physisch gespeichert werden. Globale
nutzerübergreifende Deduplizierung vermeiden wir zunächst aus
Privacy-Gründen.

## Inbound

MIME-Mail wird geparst. Anhänge werden extrahiert, gescannt und in den
Vault geschrieben. Der Body behält Referenzen für normale Attachments
und Inline/CID-Inhalte.

## Outbound --- zwei Modi

### Normal Attachment

Kleine Datei wird beim Versand wieder als MIME-Attachment eingebettet.
Der Empfänger braucht EasyPeasyMail nicht.

### EasyPeasyDrop

Große Datei bleibt im Vault. Die Nachricht enthält einen zeitlich
begrenzten Download-Link.

Das ist funktional Mail-Drop-artig, aber vollständig unter unserer
Kontrolle.

## Link Security

-   kryptografisch zufälliger Capability Token
-   Ablaufdatum
-   optional Download-Limit
-   optional Empfängerbindung
-   optional Passwort als zusätzliche Schicht
-   Widerruf
-   Audit ohne unnötiges Tracking
-   keine erratbaren Object IDs als öffentliche URLs

## Wichtige Grenze

Anhänge **immer aus der Client-Replikation auslagern**: ja. Anhänge
**immer aus der eigentlichen MIME-Mail entfernen**: nein.

Bei empfangenen Nachrichten muss das Original rekonstruierbar bleiben.
Bei ausgehenden Nachrichten entscheidet eine Größen-/Policy-Grenze
zwischen normalem MIME-Anhang und EasyPeasyDrop.
