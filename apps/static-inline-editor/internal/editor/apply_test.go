package editor

import (
	"strings"
	"testing"
)

func TestApplyRegionHTMLReplacesMainContentWithSanitizedMarkup(t *testing.T) {
	source := `<!doctype html><html><body><main><h1>Alt</h1><p>Text</p></main></body></html>`
	regionHTML := `<h1>Neu</h1><p>Hallo <strong>Welt</strong> <script>alert(1)</script></p><div><p>Noch eins</p></div>`

	out, err := ApplyRegionHTML(source, "main", []string{"h1", "p"}, []string{"strong", "em", "a", "br"}, regionHTML)
	if err != nil {
		t.Fatalf("ApplyRegionHTML returned error: %v", err)
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
	if !strings.Contains(out, "<p>Noch eins</p>") {
		t.Fatalf("expected nested allowed block to be preserved")
	}
}
