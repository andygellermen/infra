# Trust & Operations Overview

## Zweck

Trust & Operations ist keine Zusatzschicht nach dem Produkt. Sie
entscheidet, ob EasyPeasyMail als Maildienst dauerhaft vertrauenswürdig
betrieben werden kann.

## Säulen

1.  Identity & Account Security
2.  Abuse Prevention
3.  Sender Reputation
4.  Delivery Observability
5.  Reliability / SRE
6.  Backup & Recovery
7.  Auditability
8.  Incident Response
9.  Capacity / Quotas
10. Privacy / Data Lifecycle

## Grundsatz

Jeder kritische Nutzerzustand braucht: - eine eindeutige technische
Wahrheit, - einen beobachtbaren Zustand, - einen Recovery-Pfad, - einen
verantwortlichen Worker/Service, - eine definierte Eskalation.

## Kein Silent Failure

EasyPeasyMail darf kritische Fehler nicht als Erfolg darstellen.
Insbesondere: - Nachricht nicht committed -\> nicht „empfangen" -
Providerzustand unbekannt -\> nicht „gesendet" - Attachment nicht
verfügbar -\> nicht „lokal" - Draft-Konflikt -\> nicht still
überschreiben
