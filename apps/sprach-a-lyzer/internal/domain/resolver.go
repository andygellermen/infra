package domain

import "github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"

type ActorID = policy.ActorID
type TargetTypeID = policy.TargetTypeID
type ExpectationSourceID = policy.ExpectationSourceID
type DiscourseRelationID = policy.DiscourseRelationID
type ModalityID = policy.ModalityID
type NegationScopeID = policy.NegationScopeID
type SenseState = policy.SenseStateID
type AmbiguityType = policy.AmbiguityTypeID

const (
	ActorSelf        = policy.ActorSelf
	ActorOtherPerson = policy.ActorOtherPerson
	ActorGroupSelf   = policy.ActorGroupSelf
	ActorUnknown     = policy.ActorUnknown

	TargetPerson      = policy.TargetPerson
	TargetSelf        = policy.TargetSelf
	TargetBehavior    = policy.TargetBehavior
	TargetEvent       = policy.TargetEvent
	TargetObject      = policy.TargetObject
	TargetProcess     = policy.TargetProcess
	TargetIdea        = policy.TargetIdea
	TargetGroup       = policy.TargetGroup
	TargetInstitution = policy.TargetInstitution
	TargetUnknown     = policy.TargetUnknown

	ExpectationSelf         = policy.ExpectationSelf
	ExpectationOtherPerson  = policy.ExpectationOtherPerson
	ExpectationGroup        = policy.ExpectationGroup
	ExpectationInstitution  = policy.ExpectationInstitution
	ExpectationLaw          = policy.ExpectationLaw
	ExpectationCulture      = policy.ExpectationCulture
	ExpectationUnspecified  = policy.ExpectationUnspecified
	ExpectationInternalized = policy.ExpectationInternalized

	RelationContrast    = policy.RelationContrast
	RelationConcession  = policy.RelationConcession
	RelationCause       = policy.RelationCause
	RelationConsequence = policy.RelationConsequence
	RelationAddition    = policy.RelationAddition
	RelationCondition   = policy.RelationCondition
	RelationCorrection  = policy.RelationCorrection
	RelationDiscounting = policy.RelationDiscounting

	ModalityNone        = policy.ModalityNone
	ModalityNecessity   = policy.ModalityNecessity
	ModalityPossibility = policy.ModalityPossibility
	ModalityPermission  = policy.ModalityPermission
	ModalityExpectation = policy.ModalityExpectation
	ModalityIntention   = policy.ModalityIntention
	ModalityProbability = policy.ModalityProbability

	NegationNone        = policy.NegationNone
	NegationProposition = policy.NegationProposition
	NegationModality    = policy.NegationModality
	NegationActor       = policy.NegationActor
	NegationAmbiguous   = policy.NegationAmbiguous

	SenseHigh      = policy.SenseHigh
	SenseMedium    = policy.SenseMedium
	SenseAmbiguous = policy.SenseAmbiguous

	AmbiguitySemantic = policy.AmbiguitySemantic
)

type ResolverResult struct {
	ContractVersion   string              `json:"contract_version"`
	Text              string              `json:"text"`
	Context           AnalysisContext     `json:"context"`
	PropositionGraph  PropositionGraph    `json:"proposition_graph"`
	SelectedSenses    []ResolverSense     `json:"selected_senses"`
	Ambiguities       []Ambiguity         `json:"ambiguities"`
	TargetType        TargetTypeID        `json:"target_type"`
	ExpectationSource ExpectationSourceID `json:"expectation_source"`
	PatternCandidates []string            `json:"pattern_candidates"`
	OverallConfidence float64             `json:"overall_confidence"`
}

type PropositionGraph struct {
	Nodes []PropositionNode `json:"nodes"`
	Edges []PropositionEdge `json:"edges"`
}

type PropositionNode struct {
	ID                string              `json:"id"`
	SentenceIndex     int                 `json:"sentence_index"`
	Text              string              `json:"text"`
	SourceStart       int                 `json:"source_start"`
	SourceEnd         int                 `json:"source_end"`
	Actor             ActorID             `json:"actor"`
	Predicate         bool                `json:"predicate"`
	Target            bool                `json:"target"`
	Time              bool                `json:"time"`
	Boundary          bool                `json:"boundary"`
	Decision          bool                `json:"decision"`
	Negation          bool                `json:"negation"`
	NegationScope     NegationScopeID     `json:"negation_scope"`
	Modality          ModalityID          `json:"modality"`
	TargetType        TargetTypeID        `json:"target_type"`
	ExpectationSource ExpectationSourceID `json:"expectation_source"`
	Confidence        float64             `json:"confidence"`
}

type PropositionEdge struct {
	Source     string              `json:"source"`
	Target     string              `json:"target"`
	Marker     string              `json:"marker"`
	Relation   DiscourseRelationID `json:"relation"`
	Confidence float64             `json:"confidence"`
}

type ResolverSense struct {
	Lexeme     string     `json:"lexeme"`
	Sense      string     `json:"sense"`
	Confidence float64    `json:"confidence"`
	Gap        float64    `json:"gap"`
	State      SenseState `json:"state"`
	// PropositionID is runtime provenance. Resolver Result v0.2 remains
	// byte-compatible; Trace v0.2 publishes this link through contributions.
	PropositionID string `json:"-"`
}

type Ambiguity struct {
	Item   string        `json:"item"`
	Type   AmbiguityType `json:"type"`
	Top    string        `json:"top"`
	Second string        `json:"second"`
	Gap    float64       `json:"gap"`
}
