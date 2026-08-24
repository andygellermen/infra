package dimension

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestCanonicalCatalogue(t *testing.T) {
	t.Parallel()

	want := []ID{Agency, Connection, Appreciation, Clarity, Volition, Openness}
	if got := All(); !slices.Equal(got, want) {
		t.Fatalf("All() = %v; want %v", got, want)
	}
	copy := All()
	copy[0] = Volition
	if All()[0] != Agency {
		t.Fatal("caller mutated canonical dimension catalogue")
	}
}

func TestCanonicalizeLegacyFreeWill(t *testing.T) {
	t.Parallel()

	got, mapping, err := Canonicalize("FREE_WILL")
	if err != nil {
		t.Fatalf("Canonicalize() error: %v", err)
	}
	if got != Volition || !mapping.Legacy {
		t.Fatalf("Canonicalize() = %q, %+v; want VOLITION legacy mapping", got, mapping)
	}
}

func TestJSONAlwaysEmitsCanonicalID(t *testing.T) {
	t.Parallel()

	var id ID
	if err := json.Unmarshal([]byte(`"FREE_WILL"`), &id); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if id != Volition {
		t.Fatalf("legacy JSON decoded as %q; want VOLITION", id)
	}
	encoded, err := json.Marshal(ID(LegacyFreeWill))
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}
	if string(encoded) != `"VOLITION"` {
		t.Fatalf("Marshal() = %s; want VOLITION", encoded)
	}
}

func TestParseRejectsUnknownID(t *testing.T) {
	t.Parallel()

	if _, err := Parse("FREEDOM"); err == nil {
		t.Fatal("unknown dimension unexpectedly accepted")
	}
}
