# MVP Technical Breakdown

## Epic A --- Identity & Account

-   Domain/account model
-   authentication
-   device registration
-   session/token lifecycle

## Epic B --- Inbound

-   SES/domain/MX setup
-   temporary S3 ingest
-   event receiver
-   MIME parser
-   dedupe
-   durable commit
-   spam/auth metadata

## Epic C --- Mail Core

-   messages/threads/state
-   bodies
-   attachment references
-   event journal
-   API

## Epic D --- Client

-   React/PWA shell
-   local cache
-   inbox
-   reader
-   offline state
-   command palette

## Epic E --- Composer/Outbound

-   Markdown editor
-   drafts
-   outbox
-   SES send
-   scheduled send
-   bounce/complaint handling

## Epic F --- Conscious Layer

-   newsletter bundling
-   reader mode
-   tracking protection
-   Moodivador
-   print personalities

## Definition of MVP

Eine eigene Domain kann Mail empfangen; ein EasyPeasyMail Client kann
diese offline lesen; der Nutzer kann eine Markdown-Mail verfassen und
über SES authentifiziert versenden; Zustände synchronisieren sich nach
Wiederverbindung zuverlässig.
