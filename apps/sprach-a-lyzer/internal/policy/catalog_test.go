package policy

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

type registryFixture struct {
	CanonicalIDs struct {
		AnalysisContexts []AnalysisContextID `json:"analysis_contexts"`
		ResonanceModes   []ResonanceModeID   `json:"resonance_modes"`
		RuleActionTypes  []RuleActionType    `json:"rule_action_types"`
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

	data, err := os.ReadFile("../../data/seed/sprach-a-lyzer_policy-registry_v0.3.json")
	if err != nil {
		t.Fatalf("read policy registry: %v", err)
	}
	var fixture registryFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode policy registry: %v", err)
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
	assertUnique(t, HardGuardrails())
	assertUnique(t, RuleActionTypes())
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
