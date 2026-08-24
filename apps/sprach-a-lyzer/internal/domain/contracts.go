package domain

type Locale string
type AnalysisContext string
type InputMode string
type PresentationProfile string
type AnalysisMode string

const (
	LocaleGerman Locale = "de-DE"

	ContextUnspecified AnalysisContext = "UNSPECIFIED"
	ContextSelfTalk    AnalysisContext = "SELF_TALK"
	ContextSafety      AnalysisContext = "SAFETY"

	InputModeText InputMode = "TEXT"

	ProfilePrivate   PresentationProfile = "PRIVATE"
	ProfileCorporate PresentationProfile = "CORPORATE"

	AnalysisModeStandard AnalysisMode = "STANDARD"
)

// AnalysisRequest is the version 0.1 input contract. Optional fields receive
// their documented defaults at the HTTP boundary.
type AnalysisRequest struct {
	Text                string              `json:"text"`
	Locale              Locale              `json:"locale,omitempty"`
	Context             AnalysisContext     `json:"context,omitempty"`
	InputMode           InputMode           `json:"input_mode,omitempty"`
	PresentationProfile PresentationProfile `json:"presentation_profile,omitempty"`
	AnalysisMode        AnalysisMode        `json:"analysis_mode,omitempty"`
}

// AnalysisResult is the public, deterministic version 0.1 result contract.
type AnalysisResult struct {
	Text               string                          `json:"text"`
	Context            AnalysisContext                 `json:"context"`
	InputMode          InputMode                       `json:"input_mode"`
	Propositions       []Proposition                   `json:"propositions"`
	ResolvedSenses     []ResolvedSense                 `json:"resolved_senses"`
	Patterns           []string                        `json:"patterns"`
	Dimensions         map[DimensionID]DimensionResult `json:"dimensions"`
	ContributionTrace  []ContributionTraceEntry        `json:"contribution_trace"`
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

// ContributionTraceEntry explains one rule's effective score contribution.
type ContributionTraceEntry struct {
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

// AnalysisTrace is the standalone v0.1 explainability contract. Contribution
// indexes are zero-based positions in Contributions.
type AnalysisTrace struct {
	Contributions []ContributionTraceEntry                `json:"contributions"`
	Assessability map[DimensionID]AssessabilityTraceEntry `json:"assessability"`
}

type AssessabilityTraceEntry struct {
	State               AssessabilityState `json:"state"`
	FinalAssessability  float64            `json:"final_assessability"`
	ContributionIndexes []int              `json:"contribution_indexes"`
}

// Trace derives a standalone trace without inventing unavailable intermediate
// values. More detailed factors require a future contract version.
func (result AnalysisResult) Trace() AnalysisTrace {
	trace := AnalysisTrace{
		Contributions: make([]ContributionTraceEntry, len(result.ContributionTrace)),
		Assessability: make(map[DimensionID]AssessabilityTraceEntry, len(CanonicalDimensions())),
	}
	copy(trace.Contributions, result.ContributionTrace)
	for _, id := range CanonicalDimensions() {
		dimensionResult, exists := result.Dimensions[id]
		if !exists || dimensionResult.State == "" {
			dimensionResult.State = NotAssessable
		}
		indexes := []int{}
		for index, contribution := range result.ContributionTrace {
			if contribution.Dimension == id {
				indexes = append(indexes, index)
			}
		}
		trace.Assessability[id] = AssessabilityTraceEntry{
			State: dimensionResult.State, FinalAssessability: dimensionResult.Assessability,
			ContributionIndexes: indexes,
		}
	}
	return trace
}
