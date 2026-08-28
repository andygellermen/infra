package policy

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

type registryFixture struct {
	RegistryVersion string `json:"registry_version"`
	VersionContract struct {
		AnalysisTrace     string `json:"analysis_trace"`
		ResolverResult    string `json:"resolver_result"`
		ResolverCatalogue string `json:"resolver_catalogue"`
	} `json:"version_contract"`
	CanonicalIDs struct {
		AnalysisContexts   []AnalysisContextID   `json:"analysis_contexts"`
		ResonanceModes     []ResonanceModeID     `json:"resonance_modes"`
		RuleActionTypes    []RuleActionType      `json:"rule_action_types"`
		TargetTypes        []TargetTypeID        `json:"target_types"`
		ExpectationSources []ExpectationSourceID `json:"expectation_sources"`
		DiscourseRelations []DiscourseRelationID `json:"discourse_relations"`
		Actors             []ActorID             `json:"actors"`
		Modalities         []ModalityID          `json:"modalities"`
		NegationScopes     []NegationScopeID     `json:"negation_scopes"`
		SenseStates        []SenseStateID        `json:"sense_states"`
		AmbiguityTypes     []AmbiguityTypeID     `json:"ambiguity_types"`
	} `json:"canonical_ids"`
	PrivacyDefaults struct {
		RawTextRetention        string `json:"raw_text_retention"`
		AnalysisStorage         bool   `json:"analysis_storage"`
		RawAudioStorage         bool   `json:"raw_audio_storage"`
		PersonalHistory         string `json:"personal_history"`
		ManagerIndividualAccess bool   `json:"manager_individual_access"`
		EmployeeRanking         bool   `json:"employee_ranking"`
		HRSelectionUse          bool   `json:"hr_selection_use"`
	} `json:"privacy_defaults"`
	FeatureFlags []struct {
		ID      FeatureFlagID `json:"id"`
		Default bool          `json:"default"`
	} `json:"feature_flags"`
	HardGuardrails []struct {
		ID       GuardrailID `json:"id"`
		Editable bool        `json:"editable"`
	} `json:"hard_guardrails"`
}

func TestCanonicalRegistryMatchesCodeContracts(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_policy-registry_v0.5.json")
	if err != nil {
		t.Fatalf("read policy registry: %v", err)
	}
	var fixture registryFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode policy registry: %v", err)
	}
	if fixture.RegistryVersion != RegistryVersion {
		t.Fatalf("registry version = %q; want %q", fixture.RegistryVersion, RegistryVersion)
	}
	if fixture.VersionContract.ResolverResult != "0.2" || fixture.VersionContract.ResolverCatalogue != "0.1" {
		t.Fatalf("resolver version contract = %+v; want result 0.2 and catalogue 0.1", fixture.VersionContract)
	}
	if fixture.VersionContract.AnalysisTrace != "0.2" {
		t.Fatalf("analysis trace version = %q; want 0.2", fixture.VersionContract.AnalysisTrace)
	}
	if !slices.Equal(fixture.CanonicalIDs.RuleActionTypes, RuleActionTypes()) {
		t.Errorf("rule action IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.RuleActionTypes, RuleActionTypes())
	}
	if !slices.Equal(fixture.CanonicalIDs.AnalysisContexts, AnalysisContexts()) {
		t.Errorf("analysis context IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.AnalysisContexts, AnalysisContexts())
	}
	if !slices.Equal(fixture.CanonicalIDs.ResonanceModes, ResonanceModes()) {
		t.Errorf("resonance mode IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.ResonanceModes, ResonanceModes())
	}
	if !slices.Equal(fixture.CanonicalIDs.TargetTypes, TargetTypes()) {
		t.Errorf("target type IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.TargetTypes, TargetTypes())
	}
	if !slices.Equal(fixture.CanonicalIDs.ExpectationSources, ExpectationSources()) {
		t.Errorf("expectation source IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.ExpectationSources, ExpectationSources())
	}
	if !slices.Equal(fixture.CanonicalIDs.DiscourseRelations, DiscourseRelations()) {
		t.Errorf("discourse relation IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.DiscourseRelations, DiscourseRelations())
	}
	if !slices.Equal(fixture.CanonicalIDs.Actors, Actors()) {
		t.Errorf("actor IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.Actors, Actors())
	}
	if !slices.Equal(fixture.CanonicalIDs.Modalities, Modalities()) {
		t.Errorf("modality IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.Modalities, Modalities())
	}
	if !slices.Equal(fixture.CanonicalIDs.NegationScopes, NegationScopes()) {
		t.Errorf("negation scope IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.NegationScopes, NegationScopes())
	}
	if !slices.Equal(fixture.CanonicalIDs.SenseStates, SenseStates()) {
		t.Errorf("sense state IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.SenseStates, SenseStates())
	}
	if !slices.Equal(fixture.CanonicalIDs.AmbiguityTypes, AmbiguityTypes()) {
		t.Errorf("ambiguity type IDs drifted: registry=%v code=%v", fixture.CanonicalIDs.AmbiguityTypes, AmbiguityTypes())
	}

	flags := DefaultFeatureFlags()
	if len(fixture.FeatureFlags) != len(flags) {
		t.Fatalf("feature flag count = %d; want %d", len(fixture.FeatureFlags), len(flags))
	}
	for _, flag := range fixture.FeatureFlags {
		if _, exists := flags[flag.ID]; !exists {
			t.Errorf("registry contains unknown feature flag %q", flag.ID)
		}
		if flag.Default {
			t.Errorf("feature flag %q must default to false", flag.ID)
		}
	}

	wantGuardrails := HardGuardrails()
	gotGuardrails := make([]GuardrailID, 0, len(fixture.HardGuardrails))
	for _, guardrail := range fixture.HardGuardrails {
		gotGuardrails = append(gotGuardrails, guardrail.ID)
		if guardrail.Editable {
			t.Errorf("hard guardrail %q must not be editable", guardrail.ID)
		}
	}
	if !slices.Equal(gotGuardrails, wantGuardrails) {
		t.Errorf("hard guardrails drifted: registry=%v code=%v", gotGuardrails, wantGuardrails)
	}

	wantPrivacy := DefaultPrivacy()
	if fixture.PrivacyDefaults.RawTextRetention != wantPrivacy.RawTextRetention ||
		fixture.PrivacyDefaults.AnalysisStorage != wantPrivacy.AnalysisStorage ||
		fixture.PrivacyDefaults.RawAudioStorage != wantPrivacy.RawAudioStorage ||
		fixture.PrivacyDefaults.PersonalHistory != wantPrivacy.PersonalHistory ||
		fixture.PrivacyDefaults.ManagerIndividualAccess != wantPrivacy.ManagerIndividualAccess ||
		fixture.PrivacyDefaults.EmployeeRanking != wantPrivacy.EmployeeRanking ||
		fixture.PrivacyDefaults.HRSelectionUse != wantPrivacy.HRSelectionUse {
		t.Errorf("privacy defaults drifted: registry=%+v code=%+v", fixture.PrivacyDefaults, wantPrivacy)
	}
}

func TestCanonicalIDsAreUnique(t *testing.T) {
	t.Parallel()
	assertUnique(t, FeatureFlags())
	assertUnique(t, AnalysisContexts())
	assertUnique(t, ResonanceModes())
	assertUnique(t, TargetTypes())
	assertUnique(t, ExpectationSources())
	assertUnique(t, DiscourseRelations())
	assertUnique(t, Actors())
	assertUnique(t, Modalities())
	assertUnique(t, NegationScopes())
	assertUnique(t, SenseStates())
	assertUnique(t, AmbiguityTypes())
	assertUnique(t, HardGuardrails())
	assertUnique(t, RuleActionTypes())
}

func TestPolicyRegistryV03RemainsHistorical(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_policy-registry_v0.3.json")
	if err != nil {
		t.Fatalf("read policy registry v0.3: %v", err)
	}
	var historical struct {
		RegistryVersion string `json:"registry_version"`
		CanonicalIDs    struct {
			Actors []ActorID `json:"actors"`
		} `json:"canonical_ids"`
		HardGuardrails []struct {
			ID GuardrailID `json:"id"`
		} `json:"hard_guardrails"`
	}
	if err := json.Unmarshal(data, &historical); err != nil {
		t.Fatalf("decode policy registry v0.3: %v", err)
	}
	if historical.RegistryVersion != "0.3" || len(historical.CanonicalIDs.Actors) != 0 || len(historical.HardGuardrails) != 18 {
		t.Fatalf("historical registry v0.3 changed: version=%q actors=%d guardrails=%d", historical.RegistryVersion, len(historical.CanonicalIDs.Actors), len(historical.HardGuardrails))
	}
}

func TestPolicyRegistryV04RemainsHistorical(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_policy-registry_v0.4.json")
	if err != nil {
		t.Fatalf("read policy registry v0.4: %v", err)
	}
	var historical struct {
		RegistryVersion string `json:"registry_version"`
		VersionContract struct {
			AnalysisTrace string `json:"analysis_trace"`
		} `json:"version_contract"`
		HardGuardrails []struct {
			ID GuardrailID `json:"id"`
		} `json:"hard_guardrails"`
	}
	if err := json.Unmarshal(data, &historical); err != nil {
		t.Fatalf("decode policy registry v0.4: %v", err)
	}
	if historical.RegistryVersion != "0.4" || historical.VersionContract.AnalysisTrace != "0.1" || len(historical.HardGuardrails) != 21 {
		t.Fatalf("historical registry v0.4 changed: version=%q trace=%q guardrails=%d", historical.RegistryVersion, historical.VersionContract.AnalysisTrace, len(historical.HardGuardrails))
	}
}

func assertUnique[T comparable](t *testing.T, values []T) {
	t.Helper()
	seen := make(map[T]bool, len(values))
	for _, value := range values {
		if seen[value] {
			t.Errorf("duplicate canonical ID %v", value)
		}
		seen[value] = true
	}
}
