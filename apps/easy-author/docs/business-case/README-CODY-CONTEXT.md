# easy-author – Kontextdokumente für Cody

Diese Markdown-Dateien bilden die fachliche und konzeptionelle Grundlage für die App `easy-author`.

Empfohlene Ablage im Repository:

```text
easy-author/
  docs/
    business-case/
      README-CODY-CONTEXT.md
      easy-author-product-vision.md
      easy-author-mvp-scope.md
      easy-author-domain-map.md
      easy-author-data-model.md
      easy-author-editor-ui.md
      easy-author-knowledge-base.md
      easy-author-asset-management.md
      easy-author-export-pipeline.md
      easy-author-permissions-comments-monetization.md
      easy-author-infra-deployment.md
      easy-author-workflow-instances.md
      easy-author-storage-format-strategy.md
```

Die Dokumente sind bewusst vom eigentlichen Cody-/Codex-Arbeitsauftrag getrennt.
Cody soll diese Dateien als Produkt- und Architekturkontext lesen und daraus die technische Umsetzung ableiten.

Wichtige Entscheidungen:

- Editor: Tiptap
- ProseMirror: darunterliegendes technisches Fundament
- Backend: Go
- Frontend: React
- MVP-Datenbank: SQLite-first
- PostgreSQL: später möglich, aber im ersten Spike nicht erforderlich
- IndexedDB: nicht im ersten Spike
- Speicherung: Hybridstrategie aus Markdown, Editor-JSON und relationalen Metadaten
- Fokus erster Spike: Editor, Kapitel, Workflow-Boxen, Anker, Clipboard-Box
