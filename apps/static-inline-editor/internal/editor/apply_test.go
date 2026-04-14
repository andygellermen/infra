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

func TestApplyRegionsHTMLPreservesInlineClassesAndStructure(t *testing.T) {
	source := `<!doctype html><html><body><main><h1>Hallo</h1><div><a class="btn btn-primary" href="/kontakt">Kontakt</a></div></main></body></html>`
	regions := map[string]string{
		"node-0001": `Hallo <span class="underline">Welt</span>`,
		"node-0002": `<a class="btn btn-primary" href="/kontakt">Kontakt</a><button class="cta" type="button">Mehr</button>`,
	}

	out, err := ApplyRegionsHTML(source, "main", []string{"h1", "div"}, []string{"strong", "em", "a", "span", "button", "br"}, regions)
	if err != nil {
		t.Fatalf("ApplyRegionsHTML returned error: %v", err)
	}
	if !strings.Contains(out, `<span class="underline">Welt</span>`) {
		t.Fatalf("expected span class to survive, got %q", out)
	}
	if !strings.Contains(out, `<a class="btn btn-primary" href="/kontakt">Kontakt</a>`) {
		t.Fatalf("expected anchor classes to survive, got %q", out)
	}
	if !strings.Contains(out, `<button class="cta" type="button">Mehr</button>`) {
		t.Fatalf("expected button structure to survive, got %q", out)
	}
}

func TestApplyRegionsHTMLConvertsEditedBlockWrappersToLineBreaks(t *testing.T) {
	source := `<!doctype html><html><body><main><p>Alt</p></main></body></html>`
	regions := map[string]string{
		"node-0001": `<div>Erste Zeile</div><div>Zweite Zeile</div>`,
	}

	out, err := ApplyRegionsHTML(source, "main", []string{"p"}, []string{"strong", "em", "a", "span", "button", "br"}, regions)
	if err != nil {
		t.Fatalf("ApplyRegionsHTML returned error: %v", err)
	}
	if !strings.Contains(out, `<p>Erste Zeile<br/>Zweite Zeile<br/></p>`) && !strings.Contains(out, `<p>Erste Zeile<br>Zweite Zeile<br></p>`) {
		t.Fatalf("expected edited block wrappers to survive as line breaks, got %q", out)
	}
}
