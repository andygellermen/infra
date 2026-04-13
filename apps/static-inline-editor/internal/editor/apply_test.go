package editor

import (
	"strings"
	"testing"
)

func TestApplyRegionsHTMLReplacesEditableBlocksAndStripsEditorAttrs(t *testing.T) {
	source := `<!doctype html><html><body><main><h1>Alt</h1><p>Text</p></main></body></html>`
	regions := map[string]string{
		"node-0001": `Neu`,
		"node-0002": `Hallo <strong>Welt</strong> <script>alert(1)</script>`,
	}

	out, err := ApplyRegionsHTML(source, "main", []string{"h1", "p"}, []string{"strong", "em", "a", "br"}, regions)
	if err != nil {
		t.Fatalf("ApplyRegionsHTML returned error: %v", err)
	}
	if !strings.Contains(out, "<h1>Neu</h1>") {
		t.Fatalf("expected updated heading in output")
	}
	if strings.Contains(strings.ToLower(out), "<script") {
		t.Fatalf("expected script tags to be stripped")
	}
	if !strings.Contains(out, "<strong>Welt</strong>") {
		t.Fatalf("expected allowed inline markup to remain")
	}
	if strings.Contains(out, "data-editor-id") || strings.Contains(out, "data-editable") {
		t.Fatalf("expected editor attributes to be removed from saved output")
	}
}

func TestApplyRegionsHTMLSupportsSelectorFallbackList(t *testing.T) {
	source := `<!doctype html><html><body><article class="content"><p>Alt</p></article></body></html>`
	regions := map[string]string{"node-0001": `Neu`}

	out, err := ApplyRegionsHTML(source, "main, .content, body", []string{"p"}, []string{"strong", "em", "a", "br"}, regions)
	if err != nil {
		t.Fatalf("ApplyRegionsHTML returned error: %v", err)
	}
	if !strings.Contains(out, "<p>Neu</p>") {
		t.Fatalf("expected updated paragraph in output")
	}
}
