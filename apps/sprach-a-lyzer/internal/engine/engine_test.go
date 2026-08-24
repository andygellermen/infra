package engine

import (
	"errors"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
)

func TestAnalyzeRejectsEmptyText(t *testing.T) {
	t.Parallel()

	_, err := New().Analyze(domain.AnalysisRequest{Text: "  "})
	if !errors.Is(err, ErrEmptyText) {
		t.Fatalf("Analyze() error = %v; want ErrEmptyText", err)
	}
}

func TestMissingEvidenceIsNotNeutralScore(t *testing.T) {
	t.Parallel()

	result, err := New().Analyze(domain.AnalysisRequest{Text: "Ein sachlicher Satz."})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	for dimension, value := range result.Dimensions {
		if value.Score != nil {
			t.Errorf("%s score = %v; want nil without evidence", dimension, *value.Score)
		}
	}
}

func TestSafetyContextOverridesInternalPressure(t *testing.T) {
	t.Parallel()

	result, err := New().Analyze(domain.AnalysisRequest{
		Text: "Du musst sofort das Gebäude verlassen!", Context: "SAFETY",
	})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	if len(result.ResolvedSenses) != 1 || result.ResolvedSenses[0].Sense != "SAFETY_NECESSITY" {
		t.Fatalf("resolved senses = %v; want SAFETY_NECESSITY", result.ResolvedSenses)
	}
	if result.Dimensions[domain.DimensionVolition].Score != nil {
		t.Fatal("safety directive must not receive a coercion score")
	}
}
