package ontology

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/resolver"
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

func TestRuntimeResolvesThreePropositionCompositions(t *testing.T) {
	t.Parallel()
	catalogue, err := Decode(bytes.NewReader(assets.ConstructOntologyV02))
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	runtime := NewRuntime(StaticProvider{Catalogue: catalogue})
	testCases := []struct {
		name, text, pattern string
		constructs          []policy.ConstructID
	}{
		{"respectful", "Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.", "RESPECTFUL_BOUNDARY", []policy.ConstructID{policy.ConstructPerspectiveTaking, policy.ConstructBoundaryClarity}},
		{"agency", "Ich muss unbedingt alles allein lösen, aber ich kann um Hilfe bitten.", "AGENCY_RECOVERY", []policy.ConstructID{policy.ConstructContextualAgency, policy.ConstructControlPressureInterpretation}},
		{"learning", "Ich bin ein Fehler, aber der Fehler zeigt mir den nächsten Versuch.", "LEARNING_RECOVERY", []policy.ConstructID{policy.ConstructPersonBehaviorLabeling, policy.ConstructArticulatedLearning}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resolution, err := resolver.New().Resolve(domain.AnalysisRequest{Text: testCase.text, Context: domain.ContextSelfTalk})
			if err != nil {
				t.Fatalf("resolver Resolve() error: %v", err)
			}
			result, err := runtime.Resolve(context.Background(), resolution)
			if err != nil {
				t.Fatalf("ontology Resolve() error: %v", err)
			}
			if len(result.Compositions) != 1 || result.Compositions[0].Pattern != testCase.pattern || !slices.Equal(result.Compositions[0].PropositionIDs, []string{"P0", "P1"}) {
				t.Fatalf("composition = %+v", result.Compositions)
			}
			got := []policy.ConstructID{}
			for _, evidence := range result.Evidence {
				got = append(got, evidence.ConstructID)
			}
			for _, want := range testCase.constructs {
				if !slices.Contains(got, want) {
					t.Errorf("evidence %v lacks %s", got, want)
				}
			}
		})
	}
}

func TestRuntimeFailsClosedWhenOntologyIsUnavailableOrInvalid(t *testing.T) {
	t.Parallel()
	want := errors.New("ontology unavailable")
	_, err := NewRuntime(failingProvider{err: want}).Resolve(context.Background(), domain.ResolverResult{})
	if !errors.Is(err, want) {
		t.Fatalf("Resolve() error = %v; want wrapped provider error", err)
	}
	catalogue, err := Decode(bytes.NewReader(assets.ConstructOntologyV02))
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	catalogue.HardGuardrails = nil
	if _, err := NewRuntime(StaticProvider{Catalogue: catalogue}).Resolve(context.Background(), domain.ResolverResult{}); err == nil {
		t.Fatal("invalid ontology was accepted")
	}
}

func TestRuntimeRejectsWrongOrderAndRelations(t *testing.T) {
	t.Parallel()
	catalogue, err := Decode(bytes.NewReader(assets.ConstructOntologyV02))
	if err != nil {
		t.Fatalf("Decode() error: %v", err)
	}
	runtime := NewRuntime(StaticProvider{Catalogue: catalogue})
	texts := []string{
		"Für mich kommt diese Lösung nicht infrage. Ich verstehe, dass dir das wichtig ist.",
		"Ich muss unbedingt alles allein lösen und ich kann um Hilfe bitten.",
		"Ich bin ein Fehler, weil der Fehler mir den nächsten Versuch zeigt.",
	}
	for _, text := range texts {
		resolution, err := resolver.New().Resolve(domain.AnalysisRequest{Text: text, Context: domain.ContextSelfTalk})
		if err != nil {
			t.Fatalf("resolver Resolve() error: %v", err)
		}
		result, err := runtime.Resolve(context.Background(), resolution)
		if err != nil {
			t.Fatalf("ontology Resolve() error: %v", err)
		}
		if len(result.Compositions) != 0 {
			t.Errorf("%q produced unsupported composition %+v", text, result.Compositions)
		}
	}
}

type failingProvider struct{ err error }

func (f failingProvider) Active(context.Context) (Catalogue, error) { return Catalogue{}, f.err }

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
