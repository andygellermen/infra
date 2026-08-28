package engine

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
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
		if catalogue.Rules[ruleIndex].Key != "R-URGENCY-DETECTOR" {
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

func TestRuntimeCatalogueActivationAndDeactivationSmoke(t *testing.T) {
	t.Parallel()

	defaultEngine := NewDefault()
	catalogue, err := defaultEngine.catalogue.Active(context.Background())
	if err != nil {
		t.Fatalf("load default catalogue: %v", err)
	}
	request := domain.AnalysisRequest{Text: "Ich muss das heute unbedingt noch schaffen.", Context: domain.ContextSelfTalk}
	active, err := defaultEngine.Analyze(request)
	if err != nil {
		t.Fatalf("active Analyze() error: %v", err)
	}
	for index := range catalogue.Rules {
		if catalogue.Rules[index].Key == "R-URGENCY" {
			catalogue.Rules[index].Enabled = false
		}
	}
	disabledEngine := New(staticCatalogue{catalogue: catalogue})
	disabled, err := disabledEngine.Analyze(request)
	if err != nil {
		t.Fatalf("disabled Analyze() error: %v", err)
	}
	if len(disabled.ContributionTrace) != len(active.ContributionTrace)-2 {
		t.Fatalf("disabled contribution count = %d; active = %d", len(disabled.ContributionTrace), len(active.ContributionTrace))
	}
	for _, contribution := range disabled.ContributionTrace {
		if contribution.RuleID == "R-URGENCY" {
			t.Fatalf("disabled rule still contributed: %+v", contribution)
		}
	}
	for index := range catalogue.Rules {
		if catalogue.Rules[index].Key == "R-URGENCY" {
			catalogue.Rules[index].Enabled = true
		}
	}
	reactivated, err := New(staticCatalogue{catalogue: catalogue}).Analyze(request)
	if err != nil {
		t.Fatalf("reactivated Analyze() error: %v", err)
	}
	if len(reactivated.ContributionTrace) != len(active.ContributionTrace) ||
		*reactivated.Dimensions[domain.DimensionVolition].Score != *active.Dimensions[domain.DimensionVolition].Score {
		t.Fatalf("reactivated result did not restore parity: active=%+v reactivated=%+v", active, reactivated)
	}
}

func TestContextFactsAreAddressableWithoutResolverGuardrailBypass(t *testing.T) {
	t.Parallel()

	definitions := []rules.Definition{
		factRule("20000000-0000-4000-8000-000000000301", "R-PERSON-FACT", "PERSON_FACT", rules.Condition{Op: "AND", Children: []rules.Condition{
			{Field: "target_type", Operator: "EQUALS", Value: []byte(`"PERSON"`)},
			{Field: "selected_sense", Operator: "EQUALS", Value: []byte(`"PERSON_LABEL"`)},
			{Field: "proposition_feature", Operator: "EQUALS", Value: []byte(`"PREDICATE"`)},
		}}),
		factRule("20000000-0000-4000-8000-000000000302", "R-LAW-CONTRAST", "LAW_CONTRAST", rules.Condition{Op: "AND", Children: []rules.Condition{
			{Field: "expectation_source", Operator: "EQUALS", Value: []byte(`"LAW"`)},
			{Field: "discourse_relation", Operator: "EQUALS", Value: []byte(`"CONTRAST"`)},
			{Field: "proposition_feature", Operator: "EQUALS", Value: []byte(`"TIME"`)},
		}}),
		factRule("20000000-0000-4000-8000-000000000303", "R-CANDIDATE-BYPASS", "CANDIDATE_BYPASS", rules.Condition{
			Field: "pattern", Operator: "EQUALS", Value: []byte(`"PERSON_DEVALUATION"`),
		}),
		contributionRule("20000000-0000-4000-8000-000000000304", "R-AMBIGUOUS-SCORE", rules.Condition{
			Field: "selected_sense", Operator: "EQUALS", Value: []byte(`"PERSON_LABEL"`),
		}),
		factRule("20000000-0000-4000-8000-000000000305", "R-TRUSTED-SENSE", "TRUSTED_SENSE", rules.Condition{
			Field: "selected_sense", Operator: "EQUALS", Value: []byte(`"INTERNALIZED_EXPECTATION"`),
		}),
	}
	engine := New(staticCatalogue{catalogue: rules.Catalogue{Version: "context-test", Rules: definitions}})
	person, err := engine.Analyze(domain.AnalysisRequest{Text: "Du bist das Problem.", Context: "PRIVATE_CONVERSATION"})
	if err != nil {
		t.Fatalf("person Analyze() error: %v", err)
	}
	if containsString(person.Patterns, "PERSON_FACT") || containsString(person.Patterns, "CANDIDATE_BYPASS") || person.Dimensions[domain.DimensionAgency].Score != nil {
		t.Fatalf("ambiguous sense or resolver candidate bypassed rule guardrails: %+v", person)
	}
	trusted, err := engine.Analyze(domain.AnalysisRequest{Text: "Ich sollte längst weiter sein.", Context: "SELF_TALK"})
	if err != nil {
		t.Fatalf("trusted Analyze() error: %v", err)
	}
	if !containsString(trusted.Patterns, "TRUSTED_SENSE") {
		t.Fatalf("non-ambiguous selected sense was not addressable: %v", trusted.Patterns)
	}
	law, err := engine.Analyze(domain.AnalysisRequest{Text: "Ich bin gesetzlich bis Freitag verpflichtet, aber die Form bleibt offen.", Context: "LEGAL_ADMINISTRATIVE"})
	if err != nil {
		t.Fatalf("law Analyze() error: %v", err)
	}
	if !containsString(law.Patterns, "LAW_CONTRAST") {
		t.Fatalf("context/proposition facts did not match: %v", law.Patterns)
	}
}

func TestExpandedDiscourseRelationsAreRuleAddressable(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		id, key, relation, pattern, text string
	}{
		{"20000000-0000-4000-8000-000000000311", "R-CAUSE-FACT", "CAUSE", "CAUSE_FACT", "Ich bleibe zu Hause, weil ich krank bin."},
		{"20000000-0000-4000-8000-000000000312", "R-CONSEQUENCE-FACT", "CONSEQUENCE", "CONSEQUENCE_FACT", "Es regnet. Deshalb bleibe ich zu Hause."},
		{"20000000-0000-4000-8000-000000000313", "R-CONDITION-FACT", "CONDITION", "CONDITION_FACT", "Wenn du Zeit hast, sprechen wir morgen."},
		{"20000000-0000-4000-8000-000000000314", "R-ADDITION-FACT", "ADDITION", "ADDITION_FACT", "Ich prüfe die Unterlagen und ich melde mich morgen."},
		{"20000000-0000-4000-8000-000000000315", "R-CORRECTION-FACT", "CORRECTION", "CORRECTION_FACT", "Das ist kein Fehler, sondern ein Lernschritt."},
	}
	definitions := make([]rules.Definition, 0, len(testCases))
	for _, testCase := range testCases {
		definitions = append(definitions, factRule(testCase.id, testCase.key, testCase.pattern, rules.Condition{
			Field: "discourse_relation", Operator: "EQUALS", Value: []byte(`"` + testCase.relation + `"`),
		}))
	}
	engine := New(staticCatalogue{catalogue: rules.Catalogue{Version: "relations-test", Rules: definitions}})
	for _, testCase := range testCases {
		t.Run(testCase.relation, func(t *testing.T) {
			result, err := engine.Analyze(domain.AnalysisRequest{Text: testCase.text})
			if err != nil {
				t.Fatalf("Analyze() error: %v", err)
			}
			if !containsString(result.Patterns, testCase.pattern) {
				t.Fatalf("relation %s was not rule-addressable: %v", testCase.relation, result.Patterns)
			}
		})
	}
}

func TestContributionTraceV02LinksLocalResolverFacts(t *testing.T) {
	t.Parallel()
	definitions := []rules.Definition{
		localContributionRule("20000000-0000-4000-8000-000000000321", "R-TARGET-LOCAL", domain.DimensionAgency, rules.Condition{
			Field: "target_type", Operator: "EQUALS", Value: []byte(`"PROCESS"`),
		}),
		localContributionRule("20000000-0000-4000-8000-000000000322", "R-EXPECTATION-LOCAL", domain.DimensionClarity, rules.Condition{
			Field: "expectation_source", Operator: "EQUALS", Value: []byte(`"INTERNALIZED"`),
		}),
	}
	result, err := New(staticCatalogue{catalogue: rules.Catalogue{Version: "provenance-test", Rules: definitions}}).Analyze(domain.AnalysisRequest{
		Text: "Technisches Problem liegt in der Schnittstelle. Ich sollte längst reagieren.", Context: "WORKPLACE",
	})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	trace := result.TraceV02()
	if trace.ContractVersion != domain.AnalysisTraceV02ContractVersion || len(trace.Propositions) != 2 || len(trace.Contributions) != 2 {
		t.Fatalf("trace v0.2 envelope = %+v", trace)
	}
	wantByRule := map[string][]string{"R-TARGET-LOCAL": {"P0"}, "R-EXPECTATION-LOCAL": {"P1"}}
	for _, contribution := range trace.Contributions {
		if !slices.Equal(contribution.PropositionIDs, wantByRule[contribution.RuleID]) {
			t.Errorf("%s proposition IDs = %v; want %v", contribution.RuleID, contribution.PropositionIDs, wantByRule[contribution.RuleID])
		}
	}
	if trace.Propositions[0].TargetType != domain.TargetProcess || trace.Propositions[1].ExpectationSource != domain.ExpectationInternalized {
		t.Fatalf("trace proposition context = %+v", trace.Propositions)
	}
}

func TestSpanningPhraseContributionLinksAllSupportingPropositions(t *testing.T) {
	t.Parallel()
	result, err := NewDefault().Analyze(domain.AnalysisRequest{
		Text:    "Ich verstehe, dass dir das wichtig ist. Für mich kommt diese Lösung trotzdem nicht infrage.",
		Context: "PRIVATE_CONVERSATION",
	})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	for _, contribution := range result.TraceV02().Contributions {
		if contribution.RuleID == "R-RESPECTFUL-BOUNDARY" && !slices.Equal(contribution.PropositionIDs, []string{"P0", "P1"}) {
			t.Errorf("spanning contribution proposition IDs = %v; want P0/P1", contribution.PropositionIDs)
		}
	}
}

func localContributionRule(id, key string, dimension domain.DimensionID, condition rules.Condition) rules.Definition {
	value, confidence := 5.0, .8
	return rules.Definition{
		ContractVersion: rules.ContractVersion, ID: id, Key: key, Name: key,
		Description: "Proposition-local contribution provenance test rule.", Priority: 100, Enabled: true,
		Scope: "TEXT", Status: "TESTING", Version: 1, Condition: condition,
		Actions: []rules.Action{{
			Type: policy.AddContribution, Dimension: dimension, Value: &value, Confidence: &confidence,
			ReasonKey: "REASON_INTERNAL_PRESSURE_VOLITION", EvidenceKey: "EVIDENCE_INTERNAL_PRESSURE",
		}},
		ConfidenceModifier: 1,
	}
}

func contributionRule(id, key string, condition rules.Condition) rules.Definition {
	value, confidence := -20.0, .9
	return rules.Definition{
		ContractVersion: rules.ContractVersion, ID: id, Key: key, Name: key,
		Description: "Ambiguous resolver sense guardrail smoke rule.", Priority: 110, Enabled: true,
		Scope: "TEXT", Status: "TESTING", Version: 1, Condition: condition,
		Actions: []rules.Action{{
			Type: policy.AddContribution, Dimension: domain.DimensionAgency, Value: &value, Confidence: &confidence,
			ReasonKey: "REASON_INTERNAL_PRESSURE_VOLITION", EvidenceKey: "EVIDENCE_INTERNAL_PRESSURE",
		}},
		ConfidenceModifier: 1,
	}
}

func factRule(id, key, pattern string, condition rules.Condition) rules.Definition {
	return rules.Definition{
		ContractVersion: rules.ContractVersion, ID: id, Key: key, Name: key,
		Description: "Context/proposition fact smoke rule.", Priority: 100, Enabled: true,
		Scope: "TEXT", Status: "TESTING", Version: 1, Condition: condition,
		Actions: []rules.Action{{Type: policy.AddPattern, Key: pattern}}, ConfidenceModifier: 1,
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

func TestMissingPresentationKeyIsFailClosed(t *testing.T) {
	t.Parallel()

	defaultEngine := NewDefault()
	_, err := NewWithProviders(defaultEngine.catalogue, staticTexts{bundles: map[string]map[string]string{
		"PRIVATE/de-DE": {},
	}}).Analyze(domain.AnalysisRequest{Text: "Der Eintritt ist frei."})
	if err == nil || !containsError(err, "unpublished presentation key") {
		t.Fatalf("Analyze() error = %v; want missing presentation key", err)
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

func containsError(err error, want string) bool {
	return err != nil && strings.Contains(err.Error(), want)
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
