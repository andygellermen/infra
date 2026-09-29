# Attachment Vault Protocol v1

## Upload Draft Attachment

1.  Client meldet Metadaten/Hash.
2.  Core prüft vorhandenes account-lokales Objekt.
3.  Falls vorhanden: Referenz erzeugen.
4.  Falls nicht: kurzlebige Upload Capability ausstellen.
5.  Client lädt direkt in Vault/Object Store.
6.  Core verifiziert Größe/Hash.
7.  Scan.
8.  Objekt `ACTIVE` oder `QUARANTINED`.

## Download

Client fragt Download Capability für Attachment-ID an.

Core prüft: - Account ownership - Device/session - Attachment state -
Message access

Capability ist kurzlebig und gilt nur für das konkrete Objekt.

## Inline Content

CID-Inhalte sind Attachment Objects mit `disposition=INLINE`. Reader
lädt sie nur gemäß Privacy-/Remote-Content-Regeln.

## EasyPeasyDrop Outbound

Große Datei: - Vault Object - Share Grant - Ablaufzeit - optional
max_downloads - revoke_at? - recipient binding optional

Öffentliche URL enthält keine interne Object-ID als alleinige
Autorisierung.

## Integrity

Download kann Hash prüfen. Client speichert lokale Cache-Metadaten
getrennt vom kanonischen Attachment.
