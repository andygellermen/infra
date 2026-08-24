package golden_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/engine"
)

type suite struct {
	Version               string               `json:"version"`
	CanonicalDimensionIDs []domain.DimensionID `json:"canonical_dimension_ids"`
	Cases                 []goldenCase         `json:"cases"`
}

type goldenCase struct {
	ID       string                 `json:"id"`
	Request  domain.AnalysisRequest `json:"request"`
	Expected expected               `json:"expected"`
}

type expected struct {
	Senses                 []string                       `json:"senses"`
	Patterns               []string                       `json:"patterns"`
	Scores                 map[domain.DimensionID]float64 `json:"scores"`
	Unassessable           []domain.DimensionID           `json:"unassessable"`
	MinimumTraceEntries    int                            `json:"minimum_trace_entries"`
	ReflectionQuestion     bool                           `json:"reflection_question"`
	MinimumAlternatives    int                            `json:"minimum_alternatives"`
	ResonanceHint          string                         `json:"resonance_hint"`
	SemanticScoreUnchanged bool                           `json:"semantic_score_unchanged"`
}

func TestVerticalSlice(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../../data/golden/sprach-a-lyzer_vertical-slice_v0.1.json")
	if err != nil {
		t.Fatalf("read golden suite: %v", err)
	}

	var testSuite suite
	if err := json.Unmarshal(raw, &testSuite); err != nil {
		t.Fatalf("decode golden suite: %v", err)
	}
	if !slices.Equal(testSuite.CanonicalDimensionIDs, domain.Dimensions) {
		t.Fatalf("golden dimension IDs %v differ from core IDs %v", testSuite.CanonicalDimensionIDs, domain.Dimensions)
	}
	if len(testSuite.Cases) != 6 {
		t.Fatalf("golden suite contains %d cases; want 6", len(testSuite.Cases))
	}

	analyzer := engine.New()
	for _, testCase := range testSuite.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			result, err := analyzer.Analyze(testCase.Request)
			if err != nil {
				t.Fatalf("Analyze() error: %v", err)
			}

			for _, want := range testCase.Expected.Senses {
				if !hasSense(result, want) {
					t.Errorf("resolved senses %v do not contain %q", result.ResolvedSenses, want)
				}
			}
			for _, want := range testCase.Expected.Patterns {
				if !slices.Contains(result.Patterns, want) {
					t.Errorf("patterns %v do not contain %q", result.Patterns, want)
				}
			}
			for dimension, want := range testCase.Expected.Scores {
				got := result.Dimensions[dimension].Score
				if got == nil || *got != want {
					t.Errorf("%s score = %v; want %.1f", dimension, got, want)
				}
			}
			for _, dimension := range testCase.Expected.Unassessable {
				got := result.Dimensions[dimension]
				if got.State != domain.NotAssessable || got.Score != nil {
					t.Errorf("%s = %+v; want NOT_ASSESSABLE with nil score", dimension, got)
				}
			}
			if len(result.ContributionTrace) < testCase.Expected.MinimumTraceEntries {
				t.Errorf("trace contains %d entries; want at least %d", len(result.ContributionTrace), testCase.Expected.MinimumTraceEntries)
			}
			if (result.ReflectionQuestion != nil) != testCase.Expected.ReflectionQuestion {
				t.Errorf("reflection question presence = %v; want %v", result.ReflectionQuestion != nil, testCase.Expected.ReflectionQuestion)
			}
			if len(result.Alternatives) < testCase.Expected.MinimumAlternatives {
				t.Errorf("alternatives contains %d entries; want at least %d", len(result.Alternatives), testCase.Expected.MinimumAlternatives)
			}
			if testCase.Expected.ResonanceHint != "" {
				if len(result.ResonanceHints) != 1 || result.ResonanceHints[0].Kind != testCase.Expected.ResonanceHint {
					t.Errorf("resonance hints = %v; want %q", result.ResonanceHints, testCase.Expected.ResonanceHint)
				}
				if testCase.Expected.SemanticScoreUnchanged && result.ResonanceHints[0].SemanticScore {
					t.Error("resonance hint unexpectedly changes semantic score")
				}
			}
		})
	}
}

func hasSense(result domain.AnalysisResult, want string) bool {
	for _, sense := range result.ResolvedSenses {
		if sense.Sense == want {
			return true
		}
	}
	return false
}
