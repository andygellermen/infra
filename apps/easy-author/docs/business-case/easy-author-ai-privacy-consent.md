# easy-author AI Privacy and Consent

## 1. Ziel

Dieses Dokument beschreibt Datenschutz- und Einwilligungsprinzipien für KI- und Review-Funktionen in `easy-author`.

## 2. Grundsatz

Kein Inhalt wird ohne bewusste Aktion des Autors an externe ReviewProvider gesendet.

## 3. Datenschutzstufen

```text
no_ai
local_only
selection_only
chapter_with_confirmation
book_with_explicit_consent
external_provider_allowed
```

## 4. Standardverhalten

Empfohlenes Standardverhalten:

```text
KI deaktiviert
Review nur nach aktiver Auswahl
externe Provider nur nach Konfiguration
Buchreview nur nach ausdrücklicher Bestätigung
keine dauerhafte Speicherung gesendeter Volltexte
```

## 5. Consent-Dialog

Vor externem Review soll angezeigt werden:

- Provider
- Scope
- Review-Typen
- gesendete Inhalte
- gesendete Metadaten
- ob Assets enthalten sind
- ob WorkflowBoxen/Anker enthalten sind
- ob Review archiviert wird

## 6. Sensible Inhalte

Autoren sollen Kapitel oder Bücher als sensibel markieren können.

Bei sensiblen Inhalten:

- externe Reviews standardmäßig blockieren
- erneute Bestätigung erzwingen
- keine Assets mitsenden
- Volltext nicht archivieren

## 7. Bring Your Own Key

Später kann `easy-author` erlauben, dass Autoren eigene Provider-Keys hinterlegen.

Dabei wichtig:

- verschlüsselte Speicherung
- Provider pro Projekt konfigurierbar
- keine Keys in Logs
- keine Keys im Frontend ausliefern

## 8. Lokale Modelle

Für besonders sensible Projekte sollte ein lokaler ReviewProvider möglich sein.

Dieser kann schwächer sein als ein externer Anbieter, bietet aber höhere Datensouveränität.

## 9. Audit und Transparenz

ReviewSessions sollten nachvollziehbar sein:

- wann gestartet
- welcher Provider
- welcher Scope
- welche Review-Typen
- wer bestätigt hat
- wie viele ReviewItems entstanden sind

Volltexte sollten standardmäßig nicht gespeichert werden.

