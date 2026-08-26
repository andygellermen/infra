// Package policy owns canonical product IDs and engine-locked safeguards.
//
// Adjustable fachliche values belong to versioned rule and parameter sets.
// Values in this package are contracts: an admin import or AI adapter must not
// override them.
package policy

type PrivacyDefaults struct {
	RawTextRetention        string
	AnalysisStorage         bool
	RawAudioStorage         bool
	PersonalHistory         string
	ManagerIndividualAccess bool
	EmployeeRanking         bool
	HRSelectionUse          bool
}

const (
	RetentionDiscardAfterAnalysis = "DISCARD_AFTER_ANALYSIS"
	HistoryExplicitOptIn          = "EXPLICIT_OPT_IN"
)

func DefaultPrivacy() PrivacyDefaults {
	return PrivacyDefaults{
		RawTextRetention: RetentionDiscardAfterAnalysis,
		PersonalHistory:  HistoryExplicitOptIn,
	}
}

type FeatureFlagID string

const (
	AIExplanation      FeatureFlagID = "AI_EXPLANATION"
	AIRephrasing       FeatureFlagID = "AI_REPHRASING"
	AIAdaptiveQuestion FeatureFlagID = "AI_ADAPTIVE_QUESTION"
	LocalLLM           FeatureFlagID = "LOCAL_LLM"
	CloudLLM           FeatureFlagID = "CLOUD_LLM"
	LocalASR           FeatureFlagID = "LOCAL_ASR"
	ProsodyAnalysis    FeatureFlagID = "PROSODY_ANALYSIS"
	StoreRawText       FeatureFlagID = "STORE_RAW_TEXT"
	StoreAudio         FeatureFlagID = "STORE_AUDIO"
)

func FeatureFlags() []FeatureFlagID {
	return []FeatureFlagID{
		AIExplanation, AIRephrasing, AIAdaptiveQuestion, LocalLLM, CloudLLM,
		LocalASR, ProsodyAnalysis, StoreRawText, StoreAudio,
	}
}

func DefaultFeatureFlags() map[FeatureFlagID]bool {
	result := make(map[FeatureFlagID]bool, len(FeatureFlags()))
	for _, id := range FeatureFlags() {
		result[id] = false
	}
	return result
}

type AnalysisContextID string

const (
	ContextUnspecified         AnalysisContextID = "UNSPECIFIED"
	ContextSelfTalk            AnalysisContextID = "SELF_TALK"
	ContextSafety              AnalysisContextID = "SAFETY"
	ContextLegalAdministrative AnalysisContextID = "LEGAL_ADMINISTRATIVE"
	ContextWorkplace           AnalysisContextID = "WORKPLACE"
	ContextPrivateConversation AnalysisContextID = "PRIVATE_CONVERSATION"
	ContextFamily              AnalysisContextID = "FAMILY"
	ContextCoaching            AnalysisContextID = "COACHING"
	ContextModeration          AnalysisContextID = "MODERATION"
	ContextWebsite             AnalysisContextID = "WEBSITE"
	ContextPublicInformation   AnalysisContextID = "PUBLIC_INFORMATION"
)

func AnalysisContexts() []AnalysisContextID {
	return []AnalysisContextID{
		ContextUnspecified, ContextSelfTalk, ContextSafety,
		ContextLegalAdministrative, ContextWorkplace, ContextPrivateConversation,
		ContextFamily, ContextCoaching, ContextModeration, ContextWebsite,
		ContextPublicInformation,
	}
}

type TargetTypeID string

const (
	TargetPerson      TargetTypeID = "PERSON"
	TargetSelf        TargetTypeID = "SELF"
	TargetBehavior    TargetTypeID = "BEHAVIOR"
	TargetEvent       TargetTypeID = "EVENT"
	TargetObject      TargetTypeID = "OBJECT"
	TargetProcess     TargetTypeID = "PROCESS"
	TargetIdea        TargetTypeID = "IDEA"
	TargetGroup       TargetTypeID = "GROUP"
	TargetInstitution TargetTypeID = "INSTITUTION"
	TargetUnknown     TargetTypeID = "UNKNOWN"
)

func TargetTypes() []TargetTypeID {
	return []TargetTypeID{TargetPerson, TargetSelf, TargetBehavior, TargetProcess, TargetEvent, TargetObject, TargetIdea, TargetGroup, TargetInstitution, TargetUnknown}
}

type ExpectationSourceID string

const (
	ExpectationSelf         ExpectationSourceID = "SELF"
	ExpectationOtherPerson  ExpectationSourceID = "OTHER_PERSON"
	ExpectationGroup        ExpectationSourceID = "GROUP"
	ExpectationInstitution  ExpectationSourceID = "INSTITUTION"
	ExpectationLaw          ExpectationSourceID = "LAW"
	ExpectationCulture      ExpectationSourceID = "CULTURE"
	ExpectationUnspecified  ExpectationSourceID = "UNSPECIFIED"
	ExpectationInternalized ExpectationSourceID = "INTERNALIZED"
)

func ExpectationSources() []ExpectationSourceID {
	return []ExpectationSourceID{ExpectationSelf, ExpectationOtherPerson, ExpectationGroup, ExpectationInstitution, ExpectationLaw, ExpectationCulture, ExpectationUnspecified, ExpectationInternalized}
}

type DiscourseRelationID string

const (
	RelationContrast    DiscourseRelationID = "CONTRAST"
	RelationConcession  DiscourseRelationID = "CONCESSION"
	RelationCause       DiscourseRelationID = "CAUSE"
	RelationConsequence DiscourseRelationID = "CONSEQUENCE"
	RelationAddition    DiscourseRelationID = "ADDITION"
	RelationCondition   DiscourseRelationID = "CONDITION"
	RelationCorrection  DiscourseRelationID = "CORRECTION"
	RelationDiscounting DiscourseRelationID = "DISCOUNTING"
)

func DiscourseRelations() []DiscourseRelationID {
	return []DiscourseRelationID{RelationContrast, RelationConcession, RelationCause, RelationConsequence, RelationAddition, RelationCondition, RelationCorrection, RelationDiscounting}
}

type ResonanceModeID string

const (
	ResonanceOff          ResonanceModeID = "OFF"
	ResonanceHintOnly     ResonanceModeID = "HINT_ONLY"
	ResonanceModerate     ResonanceModeID = "MODERATE"
	ResonanceFull         ResonanceModeID = "FULL"
	ResonancePersonalized ResonanceModeID = "PERSONALIZED"
)

func ResonanceModes() []ResonanceModeID {
	return []ResonanceModeID{
		ResonanceOff, ResonanceHintOnly, ResonanceModerate,
		ResonanceFull, ResonancePersonalized,
	}
}

type GuardrailID string

const (
	MissingEvidenceIsNull           GuardrailID = "MISSING_EVIDENCE_IS_NULL"
	ScoreIsBounded                  GuardrailID = "SCORE_IS_BOUNDED"
	ConfidenceIsBounded             GuardrailID = "CONFIDENCE_IS_BOUNDED"
	CalculationsAreFinite           GuardrailID = "CALCULATIONS_ARE_FINITE"
	NoCircularRuleChains            GuardrailID = "NO_CIRCULAR_RULE_CHAINS"
	NoInfiniteModifierChains        GuardrailID = "NO_INFINITE_MODIFIER_CHAINS"
	NoSemanticHomophoneInheritance  GuardrailID = "NO_SEMANTIC_HOMOPHONE_INHERITANCE"
	NoPersonDiagnosis               GuardrailID = "NO_PERSON_DIAGNOSIS"
	NoTraitClaims                   GuardrailID = "NO_TRAIT_CLAIMS"
	NoEmployeeRanking               GuardrailID = "NO_EMPLOYEE_RANKING"
	QuestionScoreBiasIsZero         GuardrailID = "QUESTION_SCORE_BIAS_IS_ZERO"
	QuestionAloneIsNotAssessable    GuardrailID = "QUESTION_ALONE_IS_NOT_ASSESSABLE"
	ResonanceDoesNotScoreCore       GuardrailID = "RESONANCE_DOES_NOT_SCORE_CORE"
	CorporateHasNoCanonicalFallback GuardrailID = "CORPORATE_HAS_NO_CANONICAL_FALLBACK"
	LLMCannotScore                  GuardrailID = "LLM_CANNOT_SCORE"
	WingScoreEvaluatesText          GuardrailID = "WING_SCORE_EVALUATES_TEXT"
	RawTextStorageRequiresOptIn     GuardrailID = "RAW_TEXT_STORAGE_REQUIRES_OPT_IN"
	RawAudioStorageRequiresOptIn    GuardrailID = "RAW_AUDIO_STORAGE_REQUIRES_OPT_IN"
)

func HardGuardrails() []GuardrailID {
	return []GuardrailID{
		MissingEvidenceIsNull, ScoreIsBounded, ConfidenceIsBounded,
		CalculationsAreFinite, NoCircularRuleChains, NoInfiniteModifierChains,
		NoSemanticHomophoneInheritance, NoPersonDiagnosis, NoTraitClaims,
		NoEmployeeRanking, QuestionScoreBiasIsZero, QuestionAloneIsNotAssessable,
		ResonanceDoesNotScoreCore, CorporateHasNoCanonicalFallback, LLMCannotScore,
		WingScoreEvaluatesText, RawTextStorageRequiresOptIn,
		RawAudioStorageRequiresOptIn,
	}
}

type RuleActionType string

const (
	AddContribution      RuleActionType = "ADD_CONTRIBUTION"
	MultiplyContribution RuleActionType = "MULTIPLY_CONTRIBUTION"
	CapMin               RuleActionType = "CAP_MIN"
	CapMax               RuleActionType = "CAP_MAX"
	SetValue             RuleActionType = "SET"
	Invert               RuleActionType = "INVERT"
	Suppress             RuleActionType = "SUPPRESS"
	AddPattern           RuleActionType = "ADD_PATTERN"
	AddExplanation       RuleActionType = "ADD_EXPLANATION"
	AddReflectionPrompt  RuleActionType = "ADD_REFLECTION_PROMPT"
	AddAlternative       RuleActionType = "ADD_ALTERNATIVE"
	MarkNonAssessable    RuleActionType = "MARK_NON_ASSESSABLE"
	SelectSense          RuleActionType = "SELECT_SENSE"
	AddResonanceHint     RuleActionType = "ADD_RESONANCE_HINT"
	StopRuleChain        RuleActionType = "STOP_RULE_CHAIN"
)

func RuleActionTypes() []RuleActionType {
	return []RuleActionType{
		AddContribution, MultiplyContribution, CapMin, CapMax, SetValue, Invert,
		Suppress, AddPattern, AddExplanation, AddReflectionPrompt, AddAlternative,
		MarkNonAssessable, SelectSense, AddResonanceHint, StopRuleChain,
	}
}
