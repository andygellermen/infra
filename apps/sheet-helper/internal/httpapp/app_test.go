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
