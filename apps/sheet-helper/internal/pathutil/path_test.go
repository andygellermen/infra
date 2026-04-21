package pathutil

import "testing"

func TestNormalize(t *testing.T) {
	tests := map[string]string{
		"":           "/",
		"   ":        "/",
		"/":          "/",
		"flyer":      "/flyer",
		"/flyer":     "/flyer",
		"flyer/":     "/flyer",
		"/flyer/":    "/flyer",
		"/flyer///":  "/flyer",
		" /flyer/  ": "/flyer",
		"//flyer///": "//flyer",
		"////":       "/",
	}

	for input, want := range tests {
		if got := Normalize(input); got != want {
			t.Fatalf("Normalize(%q) = %q, want %q", input, got, want)
		}
	}
}
