package domain

import "github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/dimension"

// DimensionID identifies one of the six canonical language dimensions.
type DimensionID = dimension.ID

const (
	DimensionAgency       = dimension.Agency
	DimensionConnection   = dimension.Connection
	DimensionAppreciation = dimension.Appreciation
	DimensionClarity      = dimension.Clarity
	DimensionVolition     = dimension.Volition
	DimensionOpenness     = dimension.Openness
)

func CanonicalDimensions() []DimensionID {
	return dimension.All()
}

// CanonicalDimension maps accepted legacy identifiers to their canonical ID.
func CanonicalDimension(id DimensionID) (DimensionID, bool) {
	canonicalID, err := dimension.Parse(string(id))
	if err != nil {
		return "", false
	}
	return canonicalID, true
}

type AssessabilityState string

const (
	NotAssessable AssessabilityState = "NOT_ASSESSABLE"
	Weak          AssessabilityState = "WEAK"
	Assessable    AssessabilityState = "ASSESSABLE"
	Strong        AssessabilityState = "STRONG"
)

type AnalysisRequest struct {
	Text                string `json:"text"`
	Locale              string `json:"locale,omitempty"`
	Context             string `json:"context,omitempty"`
	InputMode           string `json:"input_mode,omitempty"`
	PresentationProfile string `json:"presentation_profile,omitempty"`
	AnalysisMode        string `json:"analysis_mode,omitempty"`
}

type AnalysisResult struct {
	Text               string                          `json:"text"`
	Context            string                          `json:"context"`
	InputMode          string                          `json:"input_mode"`
	Propositions       []Proposition                   `json:"propositions"`
	ResolvedSenses     []ResolvedSense                 `json:"resolved_senses"`
	Patterns           []string                        `json:"patterns"`
	Dimensions         map[DimensionID]DimensionResult `json:"dimensions"`
	ContributionTrace  []Contribution                  `json:"contribution_trace"`
	ReflectionQuestion *string                         `json:"reflection_question"`
	Alternatives       []string                        `json:"alternatives"`
	ResonanceHints     []ResonanceHint                 `json:"resonance_hints"`
	Notes              []string                        `json:"notes"`
}

type Proposition struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Relation string `json:"relation,omitempty"`
}

type ResolvedSense struct {
	Lexeme     string  `json:"lexeme"`
	Sense      string  `json:"sense"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

type DimensionResult struct {
	State         AssessabilityState `json:"state"`
	Score         *float64           `json:"score"`
	Confidence    float64            `json:"confidence"`
	Assessability float64            `json:"assessability"`
}

type Contribution struct {
	RuleID    string      `json:"rule_id"`
	Evidence  string      `json:"evidence"`
	Dimension DimensionID `json:"dimension"`
	Delta     float64     `json:"delta"`
	Reason    string      `json:"reason"`
}

type ResonanceHint struct {
	Kind          string   `json:"kind"`
	Tokens        []string `json:"tokens"`
	SemanticScore bool     `json:"semantic_score"`
	Message       string   `json:"message"`
}
