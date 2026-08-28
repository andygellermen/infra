package domain

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"
)

func TestAnalysisResultTraceLinksContributionsToCanonicalDimensions(t *testing.T) {
	t.Parallel()

	result := AnalysisResult{
		Dimensions: map[DimensionID]DimensionResult{
			DimensionVolition: {State: Assessable, Assessability: 0.78},
		},
		ContributionTrace: []ContributionTraceEntry{
			{RuleID: "R-1", Evidence: "ich muss", Dimension: DimensionVolition, Delta: -10, Reason: "Druck"},
			{RuleID: "R-2", Evidence: "unbedingt", Dimension: DimensionVolition, Delta: -5.6, Reason: "Verstärker"},
		},
	}

	trace := result.Trace()
	if len(trace.Assessability) != 6 {
		t.Fatalf("trace dimensions = %d; want 6", len(trace.Assessability))
	}
	volition := trace.Assessability[DimensionVolition]
	if volition.State != Assessable || volition.FinalAssessability != 0.78 ||
		!slices.Equal(volition.ContributionIndexes, []int{0, 1}) {
		t.Fatalf("VOLITION trace = %+v", volition)
	}
	agency := trace.Assessability[DimensionAgency]
	if agency.State != NotAssessable || len(agency.ContributionIndexes) != 0 {
		t.Fatalf("AGENCY trace = %+v", agency)
	}

	trace.Contributions[0].RuleID = "CHANGED"
	if result.ContributionTrace[0].RuleID != "R-1" {
		t.Fatal("trace mutation leaked into analysis result")
	}
}

func TestAnalysisTraceV02PublishesPropositionLinksWithoutChangingResultV01(t *testing.T) {
	t.Parallel()
	result := AnalysisResult{
		Dimensions: map[DimensionID]DimensionResult{DimensionAgency: {State: Assessable, Assessability: .7}},
		ContributionTrace: []ContributionTraceEntry{{
			RuleID: "R-LOCAL", Evidence: "e", Dimension: DimensionAgency, Delta: 5, Reason: "r",
		}},
		TraceProvenance: TraceProvenance{
			Propositions: []TraceProposition{{
				ID: "P0", Text: "Text", SourceStart: 0, SourceEnd: 4,
				TargetType: TargetProcess, ExpectationSource: ExpectationUnspecified,
			}},
			ContributionPropositionIDs: [][]string{{"P0"}},
		},
	}
	encodedResult, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	if bytes.Contains(encodedResult, []byte("trace_provenance")) || bytes.Contains(encodedResult, []byte("proposition_ids")) {
		t.Fatalf("trace v0.2 provenance leaked into result v0.1: %s", encodedResult)
	}
	trace := result.TraceV02()
	if trace.ContractVersion != AnalysisTraceV02ContractVersion || len(trace.Propositions) != 1 ||
		!slices.Equal(trace.Contributions[0].PropositionIDs, []string{"P0"}) {
		t.Fatalf("trace v0.2 = %+v", trace)
	}
	trace.Contributions[0].PropositionIDs[0] = "CHANGED"
	if result.TraceProvenance.ContributionPropositionIDs[0][0] != "P0" {
		t.Fatal("trace v0.2 mutation leaked into result provenance")
	}
}

func TestAnalysisTraceJSONUsesArraysAndCanonicalIDs(t *testing.T) {
	t.Parallel()

	encoded, err := json.Marshal((AnalysisResult{Dimensions: map[DimensionID]DimensionResult{}}).Trace())
	if err != nil {
		t.Fatalf("marshal trace: %v", err)
	}
	var decoded struct {
		Contributions []ContributionTraceEntry                `json:"contributions"`
		Assessability map[DimensionID]AssessabilityTraceEntry `json:"assessability"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal trace: %v", err)
	}
	if decoded.Contributions == nil {
		t.Fatalf("empty contributions encoded as null: %s", encoded)
	}
	if _, exists := decoded.Assessability[DimensionVolition]; !exists {
		t.Fatalf("trace lacks canonical VOLITION: %s", encoded)
	}
}
