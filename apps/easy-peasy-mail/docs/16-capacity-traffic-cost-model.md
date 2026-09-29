# Capacity, Traffic & Cost Model

## Basismodell

Für den temporären Sync-Layer nehmen wir zunächst 12 KB pro Mail für
Metadaten, Preview und Event-Overhead sowie 30 Tage Retention an.

    Mails/Tag   temporärer Footprint
  ----------- ----------------------
           10               \~3,5 MB
           50              \~17,6 MB
          100              \~35,2 MB
          500             \~175,8 MB

Anhänge dominieren dagegen sehr schnell den Traffic. Deshalb werden sie
standardmäßig nicht auf jedes Gerät repliziert.

## Traffic-Formel

``` text
metadata traffic
≈ messages × metadata_size × active_devices × sync_factor

attachment traffic
≈ opened_attachments × average_attachment_size × requesting_devices
```

`sync_factor` berücksichtigt Retries, Snapshots und Protokoll-Overhead.

## Szenarien für unsere spätere Kalkulation

-   1 / 3 / 5 Geräte
-   10 / 50 / 100 / 500 Mails pro Tag
-   5 / 20 / 50 % Attachment-Quote
-   0,5 / 2 / 10 MB durchschnittlicher Anhang
-   7 / 30 / 90 Tage HOT
-   1,1 / 1,5 / 2,0 Sync-Faktor

## Architekturziel

Metadaten dürfen großzügig repliziert werden; große Binärdaten nur
bewusst. Ein zweiter Download desselben Attachments auf demselben Gerät
sollte vermieden werden.

## Kostenregel

Kosten werden nicht nur in Euro gemessen: - Speicher - Requests -
Egress - CPU - Backup - Betriebsaufwand - Recovery-Komplexität
