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

type relationsScopeSuite struct {
	SuiteVersion      string               `json:"suite_version"`
	ResolverContract  string               `json:"resolver_contract"`
	ResolverCatalogue string               `json:"resolver_catalogue"`
	Cases             []relationsScopeCase `json:"cases"`
}

type relationsScopeCase struct {
	ID       string                 `json:"id"`
	Request  domain.AnalysisRequest `json:"request"`
	Expected struct {
		Nodes []relationsScopeNode     `json:"nodes"`
		Edges []domain.PropositionEdge `json:"edges"`
	} `json:"expected"`
}

type relationsScopeNode struct {
	Text          string                 `json:"text"`
	Actor         domain.ActorID         `json:"actor"`
	Negation      bool                   `json:"negation"`
	NegationScope domain.NegationScopeID `json:"negation_scope"`
	Modality      domain.ModalityID      `json:"modality"`
}

func TestRelationsAndScopeGoldenV03(t *testing.T) {
	t.Parallel()
	suite := loadRelationsScopeSuite(t)
	if suite.SuiteVersion != "0.3" || suite.ResolverContract != resolver.ContractVersion ||
		suite.ResolverCatalogue != resolver.CatalogueVersion || len(suite.Cases) != 9 {
		t.Fatalf("unexpected relations/scope suite header: %+v", suite)
	}
	engine := resolver.New()
	for _, testCase := range suite.Cases {
		t.Run(testCase.ID, func(t *testing.T) {
			result, err := engine.Resolve(testCase.Request)
			if err != nil {
				t.Fatalf("Resolve() error: %v", err)
			}
			if len(result.PropositionGraph.Nodes) != len(testCase.Expected.Nodes) {
				t.Fatalf("nodes = %d; want %d", len(result.PropositionGraph.Nodes), len(testCase.Expected.Nodes))
			}
			for index, node := range result.PropositionGraph.Nodes {
				got := relationsScopeNode{Text: node.Text, Actor: node.Actor, Negation: node.Negation, NegationScope: node.NegationScope, Modality: node.Modality}
				if !reflect.DeepEqual(got, testCase.Expected.Nodes[index]) {
					t.Fatalf("node %d = %+v; want %+v", index, got, testCase.Expected.Nodes[index])
				}
			}
			assertEqual(t, "edges", testCase.Expected.Edges, result.PropositionGraph.Edges)
			assertGraphIntegrity(t, result)
			for _, node := range result.PropositionGraph.Nodes {
				if result.Text[node.SourceStart:node.SourceEnd] != node.Text {
					t.Fatalf("node %s does not match source span", node.ID)
				}
			}
		})
	}
}

func loadRelationsScopeSuite(t *testing.T) relationsScopeSuite {
	t.Helper()
	file, err := os.Open("../../data/golden/sprach-a-lyzer_relations-scope_v0.3.json")
	if err != nil {
		t.Fatalf("open relations/scope suite: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var suite relationsScopeSuite
	if err := decoder.Decode(&suite); err != nil {
		t.Fatalf("decode relations/scope suite: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("decode trailing relations/scope data: %v", err)
	}
	return suite
}
