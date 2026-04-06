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

func TestPrepareDocumentSupportsSelectorFallbackList(t *testing.T) {
	source := `<!doctype html><html><body><article class="content"><p>Hallo</p></article></body></html>`

	doc, err := PrepareDocument(source, "main, .content, body", []string{"p"})
	if err != nil {
		t.Fatalf("PrepareDocument returned error: %v", err)
	}
	if len(doc.EditableIDs) != 1 {
		t.Fatalf("expected 1 editable id, got %d", len(doc.EditableIDs))
	}
}

func TestPrepareDocumentRemovesScriptTagsFromEditView(t *testing.T) {
	source := `<!doctype html><html><head><script src="/app.js"></script></head><body><main><p>Hallo</p><script>alert(1)</script></main></body></html>`

	doc, err := PrepareDocument(source, "main", []string{"p"})
	if err != nil {
		t.Fatalf("PrepareDocument returned error: %v", err)
	}
	if strings.Contains(strings.ToLower(doc.HTML), "<script") {
		t.Fatalf("expected scripts to be removed from edit view html")
	}
}
