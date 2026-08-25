package resolver_test

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/resolver"
)

type goldenSuite struct {
	SuiteVersion     string       `json:"suite_version"`
	ResolverContract string       `json:"resolver_contract"`
	Cases            []goldenCase `json:"cases"`
}

type goldenCase struct {
	ID       string                 `json:"id"`
	Request  domain.AnalysisRequest `json:"request"`
	Expected expectedResult         `json:"expected"`
}

type expectedResult struct {
	Nodes             []expectedNode             `json:"nodes"`
	Edges             []domain.PropositionEdge   `json:"edges"`
	SelectedSenses    []domain.ResolverSense     `json:"selected_senses"`
	Ambiguities       []domain.Ambiguity         `json:"ambiguities"`
	TargetType        domain.TargetTypeID        `json:"target_type"`
	ExpectationSource domain.ExpectationSourceID `json:"expectation_source"`
	PatternCandidates []string                   `json:"pattern_candidates"`
	OverallConfidence float64                    `json:"overall_confidence"`
}

type expectedNode struct {
	Text          string                 `json:"text"`
	Actor         domain.ActorID         `json:"actor"`
	Predicate     bool                   `json:"predicate"`
	Target        bool                   `json:"target"`
	Time          bool                   `json:"time"`
	Boundary      bool                   `json:"boundary"`
	Decision      bool                   `json:"decision"`
	Negation      bool                   `json:"negation"`
	NegationScope domain.NegationScopeID `json:"negation_scope"`
	Modality      domain.ModalityID      `json:"modality"`
}

func TestContextPropositionGolden(t *testing.T) {
	t.Parallel()
	suite := loadSuite(t)
	if suite.SuiteVersion != "0.1" || suite.ResolverContract != resolver.ContractVersion || len(suite.Cases) != 7 {
		t.Fatalf("unexpected suite header: %+v", suite)
	}
	engine := resolver.New()
	for _, testCase := range suite.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			got, err := engine.Resolve(testCase.Request)
			if err != nil {
				t.Fatalf("Resolve() error: %v", err)
			}
			if got.ContractVersion != suite.ResolverContract || got.Text != testCase.Request.Text || got.Context != testCase.Request.Context {
				t.Fatalf("resolver envelope = %+v", got)
			}
			assertNodes(t, got.Text, testCase.Expected.Nodes, got.PropositionGraph.Nodes)
			assertEqual(t, "edges", testCase.Expected.Edges, got.PropositionGraph.Edges)
			assertEqual(t, "selected senses", testCase.Expected.SelectedSenses, got.SelectedSenses)
			assertEqual(t, "ambiguities", testCase.Expected.Ambiguities, got.Ambiguities)
			assertEqual(t, "pattern candidates", testCase.Expected.PatternCandidates, got.PatternCandidates)
			if got.TargetType != testCase.Expected.TargetType || got.ExpectationSource != testCase.Expected.ExpectationSource || got.OverallConfidence != testCase.Expected.OverallConfidence {
				t.Fatalf("context resolution = %s/%s/%v; want %s/%s/%v", got.TargetType, got.ExpectationSource, got.OverallConfidence, testCase.Expected.TargetType, testCase.Expected.ExpectationSource, testCase.Expected.OverallConfidence)
			}
			assertGraphIntegrity(t, got)
		})
	}
}

func assertNodes(t *testing.T, source string, want []expectedNode, got []domain.PropositionNode) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("nodes = %d; want %d", len(got), len(want))
	}
	for index, node := range got {
		actual := expectedNode{Text: node.Text, Actor: node.Actor, Predicate: node.Predicate, Target: node.Target, Time: node.Time, Boundary: node.Boundary, Decision: node.Decision, Negation: node.Negation, NegationScope: node.NegationScope, Modality: node.Modality}
		if !reflect.DeepEqual(actual, want[index]) {
			t.Fatalf("node %d = %+v; want %+v", index, actual, want[index])
		}
		if node.SourceStart < 0 || node.SourceEnd > len(source) || node.SourceStart >= node.SourceEnd || source[node.SourceStart:node.SourceEnd] != node.Text {
			t.Fatalf("node %s has invalid source span %d:%d for %q", node.ID, node.SourceStart, node.SourceEnd, source)
		}
	}
}

func assertGraphIntegrity(t *testing.T, result domain.ResolverResult) {
	t.Helper()
	nodes := make(map[string]bool, len(result.PropositionGraph.Nodes))
	for _, node := range result.PropositionGraph.Nodes {
		if nodes[node.ID] {
			t.Fatalf("duplicate node ID %s", node.ID)
		}
		nodes[node.ID] = true
	}
	for _, edge := range result.PropositionGraph.Edges {
		if !nodes[edge.Source] || !nodes[edge.Target] || edge.Source == edge.Target {
			t.Fatalf("invalid graph edge %+v", edge)
		}
	}
}

func assertEqual(t *testing.T, label string, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("%s = %+v; want %+v", label, got, want)
	}
}

func loadSuite(t *testing.T) goldenSuite {
	t.Helper()
	file, err := os.Open("../../data/golden/sprach-a-lyzer_context-proposition_v0.1.json")
	if err != nil {
		t.Fatalf("open golden suite: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var suite goldenSuite
	if err := decoder.Decode(&suite); err != nil {
		t.Fatalf("decode golden suite: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("decode trailing golden data: %v", err)
	}
	return suite
}
