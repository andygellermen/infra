package httpapp

import "testing"

func TestSyncTokenFromPath(t *testing.T) {
	valid := "/s0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	token, ok := syncTokenFromPath(valid)
	if !ok {
		t.Fatalf("expected valid sync token path")
	}
	if token != valid[1:] {
		t.Fatalf("unexpected token %q", token)
	}
	token, ok = syncTokenFromPath(valid + "/")
	if !ok {
		t.Fatalf("expected trailing slash to be ignored for sync token path")
	}
	if token != valid[1:] {
		t.Fatalf("unexpected token %q after trailing slash normalization", token)
	}

	invalid := []string{
		"/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"/s0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcde",
		"/s0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdeg",
		"/s0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef/extra",
	}
	for _, candidate := range invalid {
		if _, ok := syncTokenFromPath(candidate); ok {
			t.Fatalf("expected invalid sync token path for %q", candidate)
		}
	}
}

func TestNormalizedPathRemovesTrailingSlash(t *testing.T) {
	if got := normalizedPath("/flyer/"); got != "/flyer" {
		t.Fatalf("expected /flyer, got %q", got)
	}
	if got := normalizedPath("/"); got != "/" {
		t.Fatalf("expected root path to stay /, got %q", got)
	}
}
