# Security Baseline --- ohne E-Mail-Verschlüsselung

## Klare Produktentscheidung

E-Mail-/Content-Verschlüsselung ist **nicht Bestandteil des MVP**.

Kein: - PGP-Workflow - S/MIME-Workflow - clientseitig verschlüsselter
Mailbody - proprietäres E2EE-Mailformat

Eine spätere Verschlüsselungsfunktion darf als optionales
Plugin/Capability ergänzt werden.

## Davon unabhängig notwendig

Das Weglassen von Mailverschlüsselung bedeutet nicht das Weglassen
technischer Basissicherheit.

Beibehalten werden: - TLS für Netzwerkverbindungen - sichere
Authentisierung - kurzlebige Sessions/Tokens - Secret Management -
Gerätewiderruf - Rate Limiting - CSRF/XSS/Injection-Schutz - sichere
Dateiupload-/Attachment-Prüfung - minimale Berechtigungen -
Backup-Zugriffsschutz - Audit sicherheitsrelevanter Aktionen

## Storage

Storage-at-rest-Schutz kann als Infrastrukturmerkmal des Hosters/Object
Stores genutzt werden, ohne ein eigenes kryptografisches Mailformat in
EasyPeasyMail einzuführen.

## Architekturregel

Core-Datenmodelle dürfen keine Verschlüsselungs-Pflicht enthalten. Ein
späteres Plugin muss sich über Capability-/Content-Adapter einklinken
können.
