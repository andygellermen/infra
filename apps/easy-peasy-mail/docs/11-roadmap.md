# Roadmap

## Phase 0 --- Entscheidungen

-   Domain-/Account-Modell
-   Inbound: SES Edge vs. eigener SMTP Edge
-   autoritativer Durable Store
-   Verschlüsselungs-/Key-Modell
-   Retention/Backup/Recovery
-   Conscious-Mode UX-Regeln

## MVP 0.1 --- Lesen

-   Go Core
-   React/PWA
-   lokaler Cache
-   Inbound Ingest
-   Mail-Liste + Reader
-   Body on demand
-   Attachment on demand
-   Suche
-   Conscious Mode

## MVP 0.2 --- Schreiben

-   Markdown Composer
-   SES Outbound
-   DKIM/SPF/DMARC
-   Drafts
-   Outbox
-   Scheduled Send
-   Keyboard Commands

## 0.3 --- Attention Layer

-   Newsletter Bundling
-   Labels optional
-   Spam/Quarantine
-   Reader View
-   Tracking protection
-   Quick React lokale Annotation

## 0.4 --- Resilience

-   Bootstrap snapshots
-   Failure drills
-   restore automation
-   second-region/offsite backup
-   queue replay
-   device revocation

## 0.5 --- Technology Storming

-   P2P/LAN Sync
-   WebRTC/QUIC proof of concept
-   content-addressed chunks
-   optional secondary inbound edge

## Offene Kernfragen

1.  Soll EasyPeasyMail Domains für mehrere Nutzer/Organisationen hosten?
2.  Wie lange müssen gelöschte Nachrichten recoverbar sein?
3.  Server-side encryption oder echte client-held keys für Mailbodies?
4.  Welche Funktionen müssen Web/PWA ohne laufenden eigenen Server noch
    beherrschen?
5.  Wie weit soll IMAP-Kompatibilität für Fremdclients gehen?
