package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

const catalogueFixturePath = "../../data/seed/sprach-a-lyzer_resolver-catalogue_v0.1.json"

func TestDecodeCanonicalResolverCatalogue(t *testing.T) {
	t.Parallel()
	file, err := os.Open(catalogueFixturePath)
	if err != nil {
		t.Fatalf("open resolver catalogue: %v", err)
	}
	defer file.Close()
	catalogue, err := DecodeCatalogue(file)
	if err != nil {
		t.Fatalf("DecodeCatalogue() error: %v", err)
	}
	if catalogue.Version != CatalogueVersion || len(catalogue.Lexemes) != 8 || len(catalogue.Connectors) != 8 || len(catalogue.ScopeRules) != 3 {
		t.Fatalf("unexpected resolver catalogue: version=%s lexemes=%d connectors=%d scopes=%d", catalogue.Version, len(catalogue.Lexemes), len(catalogue.Connectors), len(catalogue.ScopeRules))
	}
	if catalogue.SenseThresholds.High.MinimumConfidence != 0.75 || catalogue.SenseThresholds.High.MinimumGap != 0.20 ||
		catalogue.SenseThresholds.Medium.MinimumConfidence != 0.60 || catalogue.SenseThresholds.Medium.MinimumGap != 0.10 ||
		catalogue.SenseThresholds.Fallback != policy.SenseAmbiguous {
		t.Fatalf("canonical sense thresholds drifted: %+v", catalogue.SenseThresholds)
	}
	for _, guardrail := range catalogue.HardGuardrails {
		if !slices.Contains(policy.HardGuardrails(), guardrail) {
			t.Errorf("resolver catalogue uses unregistered guardrail %q", guardrail)
		}
	}
}

func TestDecodeResolverCatalogueRejectsUnknownField(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(catalogueFixturePath)
	if err != nil {
		t.Fatalf("read resolver catalogue: %v", err)
	}
	invalid := strings.Replace(string(data), `"locale": "de-DE"`, `"locale": "de-DE", "calibration_magic": true`, 1)
	_, err = DecodeCatalogue(strings.NewReader(invalid))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("DecodeCatalogue() error = %v; want unknown field", err)
	}
}

func TestResolverCatalogueRequiresLockedGuardrails(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(catalogueFixturePath)
	if err != nil {
		t.Fatalf("read resolver catalogue: %v", err)
	}
	invalid := strings.Replace(string(data), `    "RESOLVER_CANDIDATE_CANNOT_BYPASS_RULES"`, `    "NO_TRAIT_CLAIMS"`, 1)
	_, err = DecodeCatalogue(strings.NewReader(invalid))
	if err == nil || !strings.Contains(err.Error(), "hard guardrails") {
		t.Fatalf("DecodeCatalogue() error = %v; want locked guardrail rejection", err)
	}
}

func TestResolverCatalogueSenseActivationAndReactivation(t *testing.T) {
	t.Parallel()
	catalogue := loadCanonicalCatalogue(t)
	request := domainRequest("Du bist das Problem.", policy.ContextPrivateConversation)
	active, err := NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: catalogue}).Resolve(request)
	if err != nil {
		t.Fatalf("active Resolve() error: %v", err)
	}
	if len(active.SelectedSenses) != 1 || active.SelectedSenses[0].Sense != "PERSON_LABEL" {
		t.Fatalf("active senses = %v; want PERSON_LABEL", active.SelectedSenses)
	}

	disabled := cloneCatalogue(t, catalogue)
	for lexemeIndex := range disabled.Lexemes {
		if disabled.Lexemes[lexemeIndex].Key != "PROBLEM" {
			continue
		}
		senses := disabled.Lexemes[lexemeIndex].Senses
		disabled.Lexemes[lexemeIndex].Senses = slices.DeleteFunc(senses, func(sense SenseDefinition) bool {
			return sense.ID == "PERSON_LABEL"
		})
	}
	inactive, err := NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: disabled}).Resolve(request)
	if err != nil {
		t.Fatalf("disabled Resolve() error: %v", err)
	}
	if len(inactive.SelectedSenses) != 0 || len(inactive.PatternCandidates) != 0 || inactive.TargetType != policy.TargetUnknown {
		t.Fatalf("disabled catalogue leaked PERSON_LABEL facts: %+v", inactive)
	}

	reactivated, err := NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: catalogue}).Resolve(request)
	if err != nil {
		t.Fatalf("reactivated Resolve() error: %v", err)
	}
	if !reflect.DeepEqual(reactivated, active) {
		t.Fatalf("reactivated result differs from baseline")
	}
}

func TestResolverCatalogueControlsThresholdConnectorAndScope(t *testing.T) {
	t.Parallel()
	catalogue := loadCanonicalCatalogue(t)

	thresholdVariant := cloneCatalogue(t, catalogue)
	thresholdVariant.SenseThresholds.Medium.MinimumGap = .07
	thresholdResult, err := NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: thresholdVariant}).Resolve(
		domainRequest("Du bist das Problem.", policy.ContextPrivateConversation),
	)
	if err != nil {
		t.Fatalf("threshold Resolve() error: %v", err)
	}
	if thresholdResult.SelectedSenses[0].State != policy.SenseMedium || thresholdResult.OverallConfidence != .90 {
		t.Fatalf("catalogue threshold was not applied: %+v", thresholdResult.SelectedSenses)
	}

	connectorVariant := cloneCatalogue(t, catalogue)
	for index := range connectorVariant.Connectors {
		if connectorVariant.Connectors[index].Key == "CONNECTOR_TROTZDEM" {
			connectorVariant.Connectors[index].Markers = []string{"gleichwohl"}
		}
	}
	connectorResult, err := NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: connectorVariant}).Resolve(
		domainRequest("Ich verstehe dich. Trotzdem bleibt meine Grenze.", policy.ContextPrivateConversation),
	)
	if err != nil {
		t.Fatalf("connector Resolve() error: %v", err)
	}
	if len(connectorResult.PropositionGraph.Edges) != 0 {
		t.Fatalf("deactivated connector still emitted edge: %v", connectorResult.PropositionGraph.Edges)
	}

	scopeVariant := cloneCatalogue(t, catalogue)
	for index := range scopeVariant.ScopeRules {
		if scopeVariant.ScopeRules[index].Key == "NEGATION_MODAL" {
			scopeVariant.ScopeRules[index].Cues = []string{"musst nicht", "solltest nicht"}
		}
	}
	scopeResult, err := NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: scopeVariant}).Resolve(
		domainRequest("Du darfst das nicht.", policy.ContextUnspecified),
	)
	if err != nil {
		t.Fatalf("scope Resolve() error: %v", err)
	}
	if scopeResult.PropositionGraph.Nodes[0].NegationScope != policy.NegationProposition {
		t.Fatalf("catalogue scope rule was not applied: %+v", scopeResult.PropositionGraph.Nodes[0])
	}

	actorVariant := cloneCatalogue(t, catalogue)
	for index := range actorVariant.ScopeRules {
		if actorVariant.ScopeRules[index].Key == "NEGATION_ACTOR_PREFIX" {
			actorVariant.ScopeRules[index].Cues = []string{"nicht wir"}
		}
	}
	actorResult, err := NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: actorVariant}).Resolve(
		domainRequest("Nicht du musst das entscheiden.", policy.ContextUnspecified),
	)
	if err != nil {
		t.Fatalf("actor scope Resolve() error: %v", err)
	}
	if actorResult.PropositionGraph.Nodes[0].NegationScope != policy.NegationProposition {
		t.Fatalf("deactivated actor cue still emitted ACTOR scope: %+v", actorResult.PropositionGraph.Nodes[0])
	}
}

func TestResolverHandlesMultipleCatalogueConnectorsInOneSentence(t *testing.T) {
	t.Parallel()
	result, err := New().Resolve(domainRequest("Wenn du Zeit hast, sprechen wir und wir planen morgen.", policy.ContextUnspecified))
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	wantRelations := []domain.DiscourseRelationID{domain.RelationCondition, domain.RelationAddition}
	if len(result.PropositionGraph.Nodes) != 3 || len(result.PropositionGraph.Edges) != len(wantRelations) {
		t.Fatalf("graph = %+v; want three nodes and two edges", result.PropositionGraph)
	}
	for index, relation := range wantRelations {
		if result.PropositionGraph.Edges[index].Relation != relation {
			t.Fatalf("edge %d relation = %s; want %s", index, result.PropositionGraph.Edges[index].Relation, relation)
		}
	}
}

func TestAdditionConnectorDoesNotSplitNominalCoordination(t *testing.T) {
	t.Parallel()
	result, err := New().Resolve(domainRequest("Ich kaufe Brot und Butter.", policy.ContextUnspecified))
	if err != nil {
		t.Fatalf("Resolve() error: %v", err)
	}
	if len(result.PropositionGraph.Nodes) != 1 || len(result.PropositionGraph.Edges) != 0 {
		t.Fatalf("nominal coordination became a proposition relation: %+v", result.PropositionGraph)
	}
}

func TestResolverCatalogueFailuresAreFailClosed(t *testing.T) {
	t.Parallel()
	want := errors.New("resolver catalogue unavailable")
	_, err := NewWithCatalogueProvider(catalogueProviderFunc(func(context.Context) (Catalogue, error) {
		return Catalogue{}, want
	})).Resolve(domainRequest("Ein Satz.", policy.ContextUnspecified))
	if !errors.Is(err, want) {
		t.Fatalf("provider error = %v; want wrapped error", err)
	}
	invalid := loadCanonicalCatalogue(t)
	invalid.HardGuardrails = invalid.HardGuardrails[:2]
	_, err = NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: invalid}).Resolve(domainRequest("Ein Satz.", policy.ContextUnspecified))
	if err == nil || !strings.Contains(err.Error(), "hard guardrails") {
		t.Fatalf("invalid catalogue error = %v; want guardrail rejection", err)
	}
}

func TestPropositionSpanGuardrailRejectsMismatch(t *testing.T) {
	t.Parallel()
	result := domain.ResolverResult{Text: "Quelle", PropositionGraph: domain.PropositionGraph{Nodes: []domain.PropositionNode{{
		ID: "P0", Text: "falsch", SourceStart: 0, SourceEnd: 6,
	}}}}
	if err := validateSourceSpans(result); !errors.Is(err, ErrInvalidSourceSpan) {
		t.Fatalf("validateSourceSpans() error = %v; want ErrInvalidSourceSpan", err)
	}
}

type catalogueProviderFunc func(context.Context) (Catalogue, error)

func (f catalogueProviderFunc) Active(ctx context.Context) (Catalogue, error) { return f(ctx) }

func loadCanonicalCatalogue(t *testing.T) Catalogue {
	t.Helper()
	file, err := os.Open(catalogueFixturePath)
	if err != nil {
		t.Fatalf("open resolver catalogue: %v", err)
	}
	defer file.Close()
	catalogue, err := DecodeCatalogue(file)
	if err != nil {
		t.Fatalf("decode resolver catalogue: %v", err)
	}
	return catalogue
}

func cloneCatalogue(t *testing.T, source Catalogue) Catalogue {
	t.Helper()
	data, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("marshal resolver catalogue: %v", err)
	}
	var result Catalogue
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("clone resolver catalogue: %v", err)
	}
	return result
}

func domainRequest(text string, contextID policy.AnalysisContextID) domain.AnalysisRequest {
	return domain.AnalysisRequest{Text: text, Context: domain.AnalysisContext(contextID)}
}
