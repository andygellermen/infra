package httpapp

import "testing"

func TestGravatarURL(t *testing.T) {
	got := gravatarURL(" Andy@Gellermann.Berlin ")
	want := "https://gravatar.com/avatar/397206df1d3142c5e4c6c1b7bbf9738ce2c0c232a2f3bb1c592be38d94024a79"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
