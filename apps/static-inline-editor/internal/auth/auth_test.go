package auth

import "testing"

func TestEmailAllowed(t *testing.T) {
	allowed := []string{"andy@example.org", "redaktion@example.org"}
	if !EmailAllowed(allowed, " Andy@Example.org ") {
		t.Fatalf("expected email to be allowed")
	}
	if EmailAllowed(allowed, "nobody@example.org") {
		t.Fatalf("expected email to be rejected")
	}
}
