package ontology

import (
	"bytes"
	"slices"
	"testing"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

func TestCanonicalOntologyMatchesPolicyAndGuardrails(t *testing.T) {
	t.Parallel()
	catalogue, err := Decode(bytes.NewReader(assets.ConstructOntologyV02))
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	got := make([]policy.ConstructID, 0, len(catalogue.Constructs))
	for _, definition := range catalogue.Constructs {
		got = append(got, definition.ID)
		if definition.CoreScoring {
			t.Errorf("construct %s directly scores", definition.ID)
		}
	}
	if !slices.Equal(got, policy.Constructs()) {
		t.Fatalf("ontology IDs = %v; policy = %v", got, policy.Constructs())
	}
	if len(catalogue.Compositions) != 3 {
		t.Fatalf("compositions = %d; want 3", len(catalogue.Compositions))
	}
}

func TestOntologyRejectsDirectScoringAndIncompleteEvidence(t *testing.T) {
	t.Parallel()
	catalogue, err := Decode(bytes.NewReader(assets.ConstructOntologyV02))
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	catalogue.Constructs[0].CoreScoring = true
	if err := catalogue.Validate(); err == nil {
		t.Fatal("directly scoring construct was accepted")
	}
	catalogue, _ = Decode(bytes.NewReader(assets.ConstructOntologyV02))
	catalogue.Constructs[0].Evidence = nil
	if err := catalogue.Validate(); err == nil {
		t.Fatal("construct without evidence was accepted")
	}
}
