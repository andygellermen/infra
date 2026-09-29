# Search Index Model

## Local First

SQLite FTS5 indexiert lokal verfügbare: - sender - recipients -
subject - preview - gecachten Body - Attachment-Dateinamen

## Server Search

Server kann Metadaten und verfügbare Bodies indexieren.
Attachment-Inhalte werden im MVP nicht automatisch
OCR-/Volltext-indexiert.

## Progressive Search

``` text
1. lokaler Treffer sofort
2. optional Server-Erweiterung
3. Resultate zusammenführen
```

## Offline

Suche bleibt auf lokal indexiertem Bestand vollständig funktionsfähig.

## Privacy / Aufwand

Kein Elasticsearch/OpenSearch im MVP. Erst reale Datenmenge und
Suchanforderungen rechtfertigen zusätzliche Infrastruktur.
