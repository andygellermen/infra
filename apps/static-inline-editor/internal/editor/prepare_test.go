package editor

import (
	"strings"
	"testing"
)

func TestPrepareDocumentMarksAllowedNodesInsideMainSelector(t *testing.T) {
	source := `<!doctype html><html><body><main><h1>Hallo</h1><p>Welt</p><div><p>Noch eins</p></div></main></body></html>`

	doc, err := PrepareDocument(source, "main", []string{"h1", "p"})
	if err != nil {
		t.Fatalf("PrepareDocument returned error: %v", err)
	}
	if len(doc.EditableIDs) != 3 {
		t.Fatalf("expected 3 editable ids, got %d", len(doc.EditableIDs))
	}
	if !strings.Contains(doc.HTML, `data-editor-id="node-0001"`) {
		t.Fatalf("expected first editor id in HTML")
	}
	if !strings.Contains(doc.HTML, `data-editor-tag="p"`) {
		t.Fatalf("expected editor tag marker in HTML")
	}
}

func TestPrepareDocumentSupportsClassSelector(t *testing.T) {
	source := `<!doctype html><html><body><article class="content"><p>Hallo</p></article></body></html>`

	doc, err := PrepareDocument(source, ".content", []string{"p"})
	if err != nil {
		t.Fatalf("PrepareDocument returned error: %v", err)
	}
	if len(doc.EditableIDs) != 1 {
		t.Fatalf("expected 1 editable id, got %d", len(doc.EditableIDs))
	}
}
