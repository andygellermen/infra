# Cross-Device Drafts v1

## Produktentscheidung

Drafts sind **geräteübergreifend bearbeitbar**.

Das ist kein Zusatzkomfort, sondern ein Kernbestandteil des
Local-First-Produktivitätsmodells.

## Zielbild

``` text
Mac
  schreibt Draft rev 12
       |
       v
Core Draft Store
       |
       +----> Smartphone rev 12
       +----> Web rev 12

Smartphone editiert -> rev 13
       |
       +----> Mac rev 13
```

## Offline

Jedes Gerät darf Drafts offline bearbeiten.

## Herausforderung: parallele Bearbeitung

Automatisches CRDT-/Google-Docs-Coediting ist für MVP unnötig.

Wir verwenden optimistische Revisionen:

``` text
draft_id
revision
base_revision
device_id
updated_at
body_markdown
```

### Normal

Client sendet Änderung auf Basis `rev 12`. Server steht auf `rev 12`.
=\> akzeptieren, `rev 13`.

### Konflikt

Mac und Smartphone ändern beide `rev 12`.

Mac gewinnt zuerst -\> `rev 13`. Smartphone sendet noch
`base_revision=12`. =\> `409 DRAFT_CONFLICT`.

Der Smartphone-Text wird **nicht verworfen**.

Core speichert/retourniert: - aktuelle Serverfassung - lokale
konkurrierende Fassung - gemeinsame Basisrevision

UI: `Beide Fassungen vergleichen` / `Meine als Kopie behalten`.

## Autosave

-   lokal sofort
-   Server debounce, z. B. 1--3 Sekunden
-   beim App-Hintergrundwechsel sofort flush
-   explizites Ctrl/Cmd+S bleibt möglich

## Attachments

Draft referenziert Vault Objects. Upload kann unabhängig vom Drafttext
fortgesetzt werden.

## Send Race

Sobald ein Draft in einen Outbox Snapshot überführt wurde, ist dieser
Snapshot unveränderlich. Der Draft kann danach archiviert/gelöscht oder
als neue Kopie weiterbearbeitet werden.

## Später

Optional echtes Collaborative Editing via CRDT, aber nur wenn reale
Nutzung es rechtfertigt.
