package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/rules"
)

func TestAnalyzeRejectsEmptyText(t *testing.T) {
	t.Parallel()

	_, err := NewDefault().Analyze(domain.AnalysisRequest{Text: "  "})
	if !errors.Is(err, ErrEmptyText) {
		t.Fatalf("Analyze() error = %v; want ErrEmptyText", err)
	}
}

func TestRuntimeCatalogueControlsRuleActivation(t *testing.T) {
	t.Parallel()

	defaultEngine := NewDefault()
	catalogue, err := defaultEngine.catalogue.Active(context.Background())
	if err != nil {
		t.Fatalf("load default catalogue: %v", err)
	}
	for index := range catalogue.Rules {
		if catalogue.Rules[index].Key == "R-INTERNAL-PRESSURE" {
			catalogue.Rules[index].Enabled = false
		}
	}
	result, err := New(staticCatalogue{catalogue: catalogue}).Analyze(domain.AnalysisRequest{
		Text: "Ich muss das heute unbedingt noch schaffen.", Context: domain.ContextSelfTalk,
	})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	if containsString(result.Patterns, "INTERNAL_PRESSURE") {
		t.Fatalf("disabled catalogue rule still emitted INTERNAL_PRESSURE: %+v", result)
	}
	for _, contribution := range result.ContributionTrace {
		if contribution.RuleID == "R-INTERNAL-PRESSURE" {
			t.Fatalf("disabled catalogue rule still contributed: %+v", contribution)
		}
	}
}

func TestRuntimeCatalogueExecutesActionPayload(t *testing.T) {
	t.Parallel()

	defaultEngine := NewDefault()
	catalogue, err := defaultEngine.catalogue.Active(context.Background())
	if err != nil {
		t.Fatalf("load default catalogue: %v", err)
	}
	for ruleIndex := range catalogue.Rules {
		if catalogue.Rules[ruleIndex].Key != "R-URGENCY" {
			continue
		}
		for actionIndex := range catalogue.Rules[ruleIndex].Actions {
			catalogue.Rules[ruleIndex].Actions[actionIndex].Key = "CATALOGUE_URGENCY"
		}
	}
	result, err := New(staticCatalogue{catalogue: catalogue}).Analyze(domain.AnalysisRequest{Text: "Sofort handeln."})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	if !containsString(result.Patterns, "CATALOGUE_URGENCY") || containsString(result.Patterns, "URGENCY") {
		t.Fatalf("runtime action payload was not executed: patterns=%v", result.Patterns)
	}
}

func TestRuntimeCatalogueFailureIsFailClosed(t *testing.T) {
	t.Parallel()

	want := errors.New("catalogue unavailable")
	_, err := New(failingCatalogue{err: want}).Analyze(domain.AnalysisRequest{Text: "Ein Satz."})
	if !errors.Is(err, want) {
		t.Fatalf("Analyze() error = %v; want wrapped catalogue error", err)
	}
}

type failingCatalogue struct{ err error }

func (f failingCatalogue) Active(context.Context) (rules.Catalogue, error) {
	return rules.Catalogue{}, f.err
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestMissingEvidenceIsNotNeutralScore(t *testing.T) {
	t.Parallel()

	result, err := NewDefault().Analyze(domain.AnalysisRequest{Text: "Ein sachlicher Satz."})
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

	result, err := NewDefault().Analyze(domain.AnalysisRequest{
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
