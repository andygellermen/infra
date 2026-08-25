package presentation

import "testing"

func TestCorporateBundleNeverReturnsCanonicalKey(t *testing.T) {
	t.Parallel()

	bundle := Bundle{
		Profile:   "CORPORATE",
		Entries:   map[string]string{"RESONANCE": "Wahrnehmungswirkung"},
		Fallbacks: map[string]string{"UNKNOWN_CONCEPT": "Sprachmuster"},
	}
	if got := bundle.Resolve("RESONANCE"); got != "Wahrnehmungswirkung" {
		t.Fatalf("Resolve(RESONANCE) = %q", got)
	}
	if got := bundle.Resolve("PRIVATE_CANONICAL_KEY"); got != "Sprachmuster" {
		t.Fatalf("unknown key leaked as %q", got)
	}
}
