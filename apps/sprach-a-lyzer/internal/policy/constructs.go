package policy

type ConstructID string

const (
	ConstructGeneralization                ConstructID = "GENERALIZATION"
	ConstructModality                      ConstructID = "MODALITY"
	ConstructNegation                      ConstructID = "NEGATION"
	ConstructCertainty                     ConstructID = "CERTAINTY"
	ConstructSpecificity                   ConstructID = "SPECIFICITY"
	ConstructTemporalReference             ConstructID = "TEMPORAL_REFERENCE"
	ConstructPersonBehaviorLabeling        ConstructID = "PERSON_BEHAVIOR_LABELING"
	ConstructOptionCount                   ConstructID = "OPTION_COUNT"
	ConstructCommitmentVerb                ConstructID = "COMMITMENT_VERB"
	ConstructEmotionWord                   ConstructID = "EMOTION_WORD"
	ConstructValueWord                     ConstructID = "VALUE_WORD"
	ConstructNeedWord                      ConstructID = "NEED_WORD"
	ConstructContextualAgency              ConstructID = "CONTEXTUAL_AGENCY"
	ConstructPerceivedInfluenceScope       ConstructID = "PERCEIVED_INFLUENCE_SCOPE"
	ConstructPerceivedOptionSpace          ConstructID = "PERCEIVED_OPTION_SPACE"
	ConstructVolition                      ConstructID = "VOLITION"
	ConstructCompetenceExpectancy          ConstructID = "COMPETENCE_EXPECTANCY"
	ConstructExplicitValues                ConstructID = "EXPLICIT_VALUES"
	ConstructExplicitNeeds                 ConstructID = "EXPLICIT_NEEDS"
	ConstructAttributedMeaning             ConstructID = "ATTRIBUTED_MEANING"
	ConstructArticulatedAmbivalence        ConstructID = "ARTICULATED_AMBIVALENCE"
	ConstructPerspectiveTaking             ConstructID = "PERSPECTIVE_TAKING"
	ConstructRelationalOrientation         ConstructID = "RELATIONAL_ORIENTATION"
	ConstructBoundaryClarity               ConstructID = "BOUNDARY_CLARITY"
	ConstructArticulatedLearning           ConstructID = "ARTICULATED_LEARNING"
	ConstructOwnedCommitment               ConstructID = "OWNED_COMMITMENT"
	ConstructBeliefLikePattern             ConstructID = "BELIEF_LIKE_PATTERN"
	ConstructInternalizedExpectation       ConstructID = "INTERNALIZED_EXPECTATION"
	ConstructIdentityFusion                ConstructID = "IDENTITY_FUSION"
	ConstructHypothesizedNeed              ConstructID = "HYPOTHESIZED_NEED"
	ConstructControlPressureInterpretation ConstructID = "CONTROL_PRESSURE_INTERPRETATION"
	ConstructSystemRuleHypothesis          ConstructID = "SYSTEM_RULE_HYPOTHESIS"
	ConstructExistentialMeaning            ConstructID = "EXISTENTIAL_MEANING"
	ConstructResonance                     ConstructID = "RESONANCE"
	ConstructInnerAlignment                ConstructID = "INNER_ALIGNMENT"
	ConstructSpiritualReflection           ConstructID = "SPIRITUAL_REFLECTION"
)

func Constructs() []ConstructID {
	return []ConstructID{
		ConstructGeneralization, ConstructModality, ConstructNegation, ConstructCertainty,
		ConstructSpecificity, ConstructTemporalReference, ConstructPersonBehaviorLabeling,
		ConstructOptionCount, ConstructCommitmentVerb, ConstructEmotionWord, ConstructValueWord,
		ConstructNeedWord, ConstructContextualAgency, ConstructPerceivedInfluenceScope,
		ConstructPerceivedOptionSpace, ConstructVolition, ConstructCompetenceExpectancy,
		ConstructExplicitValues, ConstructExplicitNeeds, ConstructAttributedMeaning,
		ConstructArticulatedAmbivalence, ConstructPerspectiveTaking, ConstructRelationalOrientation,
		ConstructBoundaryClarity, ConstructArticulatedLearning, ConstructOwnedCommitment,
		ConstructBeliefLikePattern, ConstructInternalizedExpectation, ConstructIdentityFusion,
		ConstructHypothesizedNeed, ConstructControlPressureInterpretation, ConstructSystemRuleHypothesis,
		ConstructExistentialMeaning, ConstructResonance, ConstructInnerAlignment, ConstructSpiritualReflection,
	}
}
