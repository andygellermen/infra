// Package resolver implements the deterministic Context & Proposition layer.
package resolver

import (
	"errors"
	"sort"
	"strings"
	"unicode"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
)

const ContractVersion = "0.2"

var ErrEmptyText = errors.New("resolver text must not be empty")

type Resolver struct{}

func New() *Resolver { return &Resolver{} }

func (r *Resolver) Resolve(request domain.AnalysisRequest) (domain.ResolverResult, error) {
	text := strings.TrimSpace(request.Text)
	if text == "" {
		return domain.ResolverResult{}, ErrEmptyText
	}
	contextValue := domain.AnalysisContext(strings.ToUpper(strings.TrimSpace(string(request.Context))))
	if contextValue == "" {
		contextValue = domain.ContextUnspecified
	}
	targetType := resolveTargetType(text)
	expectationSource := resolveExpectationSource(text)
	graph := buildGraph(text, targetType, expectationSource)
	senses, ambiguities, patterns := resolveSensesAndPatterns(text, graph)
	confidence := .90
	if len(ambiguities) > 0 {
		confidence = .72
	}
	return domain.ResolverResult{
		ContractVersion: ContractVersion, Text: text, Context: contextValue,
		PropositionGraph: graph, SelectedSenses: senses, Ambiguities: ambiguities,
		TargetType: targetType, ExpectationSource: expectationSource,
		PatternCandidates: patterns, OverallConfidence: confidence,
	}, nil
}

type span struct {
	text                 string
	start, end, sentence int
	marker               string
	relation             domain.DiscourseRelationID
}

func buildGraph(text string, target domain.TargetTypeID, expectation domain.ExpectationSourceID) domain.PropositionGraph {
	spans := propositionSpans(text)
	graph := domain.PropositionGraph{Nodes: make([]domain.PropositionNode, 0, len(spans)), Edges: []domain.PropositionEdge{}}
	for index, part := range spans {
		node := propositionNode(index, part, target, expectation)
		graph.Nodes = append(graph.Nodes, node)
		if index == 0 {
			continue
		}
		marker, relation := part.marker, part.relation
		if relation == "" {
			lower := normalize(part.text)
			switch {
			case hasWord(lower, "trotzdem"):
				marker, relation = "trotzdem", domain.RelationConcession
			case hasWord(lower, "deshalb"):
				marker, relation = "deshalb", domain.RelationConsequence
			case hasWord(lower, "außerdem"):
				marker, relation = "außerdem", domain.RelationAddition
			}
		}
		if relation != "" {
			graph.Edges = append(graph.Edges, domain.PropositionEdge{
				Source: graph.Nodes[index-1].ID, Target: node.ID, Marker: marker,
				Relation: relation, Confidence: .92,
			})
		}
	}
	return graph
}

func propositionSpans(text string) []span {
	var sentences []span
	start, sentenceIndex := 0, 0
	for index := 0; index < len(text); index++ {
		if text[index] != '.' && text[index] != '!' && text[index] != '?' {
			continue
		}
		if part, ok := trimmedSpan(text, start, index+1, sentenceIndex); ok {
			sentences = append(sentences, part)
			sentenceIndex++
		}
		start = index + 1
	}
	if part, ok := trimmedSpan(text, start, len(text), sentenceIndex); ok {
		sentences = append(sentences, part)
	}

	result := make([]span, 0, len(sentences)+2)
	for _, sentence := range sentences {
		lower := strings.ToLower(sentence.text)
		markerStart, markerLength := strings.Index(lower, ", aber "), len(", aber ")
		if markerStart < 0 {
			markerStart, markerLength = strings.Index(lower, " aber "), len(" aber ")
		}
		if markerStart <= 0 {
			result = append(result, sentence)
			continue
		}
		leftStart, leftEnd := sentence.start, sentence.start+markerStart
		if left, ok := trimmedSpan(text, leftStart, leftEnd, sentence.sentence); ok {
			result = append(result, left)
		}
		rightStart := sentence.start + markerStart + markerLength
		if right, ok := trimmedSpan(text, rightStart, sentence.end, sentence.sentence); ok {
			right.marker, right.relation = "aber", domain.RelationContrast
			result = append(result, right)
		}
	}
	return result
}

func trimmedSpan(source string, start, end, sentence int) (span, bool) {
	for start < end && unicode.IsSpace(rune(source[start])) {
		start++
	}
	for end > start && unicode.IsSpace(rune(source[end-1])) {
		end--
	}
	if start >= end {
		return span{}, false
	}
	return span{text: source[start:end], start: start, end: end, sentence: sentence}, true
}

func propositionNode(index int, part span, target domain.TargetTypeID, expectation domain.ExpectationSourceID) domain.PropositionNode {
	lower := normalize(part.text)
	actor := domain.ActorUnknown
	switch {
	case hasWord(lower, "ich"), strings.Contains(lower, "für mich"):
		actor = domain.ActorSelf
	case hasWord(lower, "wir"), hasWord(lower, "uns"):
		actor = domain.ActorGroupSelf
	case hasAnyWord(lower, "du", "dir", "dich", "er", "sie", "ihm", "ihr"):
		actor = domain.ActorOtherPerson
	}
	negation := hasAnyWord(lower, "nicht", "nie", "kein", "keine", "keinen", "niemals")
	modality := resolveModality(lower)
	negationScope := domain.NegationNone
	if negation {
		switch {
		case strings.HasPrefix(lower, "nicht du "), strings.HasPrefix(lower, "nicht ich "):
			negationScope = domain.NegationActor
		case modality != domain.ModalityNone && (strings.Contains(lower, "musst") || strings.Contains(lower, "darfst")):
			negationScope = domain.NegationModality
		default:
			negationScope = domain.NegationProposition
		}
	}
	predicate := strings.Trim(lower, " .,!?;") != "ja" && len(strings.Fields(lower)) > 1
	return domain.PropositionNode{
		ID: "P" + itoa(index), SentenceIndex: part.sentence, Text: part.text,
		SourceStart: part.start, SourceEnd: part.end, Actor: actor, Predicate: predicate,
		Target: hasTarget(lower), Time: hasAnyWord(lower, "heute", "morgen", "freitag", "längst", "noch", "jederzeit", "sofort"),
		Boundary: strings.Contains(lower, "nicht infrage") || strings.Contains(lower, "grenze"),
		Decision: strings.Contains(lower, "entscheid"), Negation: negation, NegationScope: negationScope,
		Modality: modality, TargetType: target, ExpectationSource: expectation, Confidence: .90,
	}
}

func resolveModality(text string) domain.ModalityID {
	switch {
	case hasAnyWord(text, "muss", "musst", "müssen"):
		return domain.ModalityNecessity
	case hasAnyWord(text, "soll", "sollte", "solltest", "sollten"):
		return domain.ModalityExpectation
	case hasAnyWord(text, "darf", "darfst", "dürfen"):
		return domain.ModalityPermission
	case hasAnyWord(text, "möchte", "will", "wollte"):
		return domain.ModalityIntention
	case hasAnyWord(text, "dürfte"):
		return domain.ModalityProbability
	case hasAnyWord(text, "kann", "können", "könnte"):
		return domain.ModalityPossibility
	default:
		return domain.ModalityNone
	}
}

func resolveTargetType(text string) domain.TargetTypeID {
	lower := normalize(text)
	switch {
	case strings.Contains(lower, "du bist das problem"), strings.Contains(lower, "du bist ein fehler"):
		return domain.TargetPerson
	case strings.Contains(lower, "technisches problem"), strings.Contains(lower, "schnittstelle"), strings.Contains(lower, "im code"):
		return domain.TargetProcess
	case strings.Contains(lower, "fehler zeigt"), strings.Contains(lower, "nächsten versuch"):
		return domain.TargetEvent
	case strings.Contains(lower, "vereinbarung") && strings.Contains(lower, "eingehalten"):
		return domain.TargetBehavior
	default:
		return domain.TargetUnknown
	}
}

func resolveExpectationSource(text string) domain.ExpectationSourceID {
	lower := normalize(text)
	switch {
	case hasAnyWord(lower, "gesetzlich", "gesetz", "vorgeschrieben"):
		return domain.ExpectationLaw
	case strings.Contains(lower, "ich sollte") && hasAnyWord(lower, "längst", "endlich", "eigentlich"):
		return domain.ExpectationInternalized
	case strings.Contains(lower, "man sollte"):
		return domain.ExpectationCulture
	case strings.Contains(lower, "ich muss") && hasAnyWord(lower, "unbedingt", "endlich", "perfekt"):
		return domain.ExpectationInternalized
	default:
		return domain.ExpectationUnspecified
	}
}

type positionedSense struct {
	position int
	sense    domain.ResolverSense
}

func resolveSensesAndPatterns(text string, graph domain.PropositionGraph) ([]domain.ResolverSense, []domain.Ambiguity, []string) {
	lower := normalize(text)
	var positioned []positionedSense
	ambiguities := []domain.Ambiguity{}
	patterns := []string{}
	add := func(lexeme, sense string, confidence, gap float64, state domain.SenseState) {
		positioned = append(positioned, positionedSense{position: strings.Index(lower, strings.ToLower(lexeme)), sense: domain.ResolverSense{
			Lexeme: lexeme, Sense: sense, Confidence: confidence, Gap: gap, State: state,
		}})
	}

	if hasAnyWord(lower, "muss", "musst", "müssen") {
		switch {
		case strings.Contains(lower, "ich muss") && hasAnyWord(lower, "unbedingt", "endlich", "perfekt"):
			add("müssen", "INTERNAL_PRESSURE", .785, .13, domain.SenseHigh)
			patterns = appendUnique(patterns, "INTERNAL_PRESSURE")
		case strings.Contains(lower, "hindernis umfahren"):
			add("müssen", "EXTERNAL_NECESSITY", .655, .05, domain.SenseAmbiguous)
			ambiguities = append(ambiguities, domain.Ambiguity{Item: "müssen", Type: domain.AmbiguitySemantic, Top: "EXTERNAL_NECESSITY", Second: "EPISTEMIC_INFERENCE", Gap: .05})
		case hasAnyWord(lower, "gefahr", "brand", "notfall") || strings.Contains(lower, "gebäude verlassen"):
			add("müssen", "SAFETY_NECESSITY", .90, .25, domain.SenseHigh)
		default:
			add("müssen", "EXTERNAL_NECESSITY", .66, .08, domain.SenseMedium)
		}
	}
	if hasAnyWord(lower, "soll", "sollte", "solltest", "sollten") {
		switch {
		case (strings.HasPrefix(lower, "er soll ") || strings.HasPrefix(lower, "sie soll ")) && strings.Contains(lower, " sein"):
			add("sollen", "REPORTED_CLAIM", .79, .14, domain.SenseHigh)
			patterns = appendUnique(patterns, "REPORTED_CLAIM")
		case strings.Contains(lower, "ich sollte") && hasAnyWord(lower, "längst", "endlich", "eigentlich"):
			add("sollen", "INTERNALIZED_EXPECTATION", .785, .135, domain.SenseHigh)
			patterns = appendUnique(patterns, "INTERNALIZED_EXPECTATION", "SELF_PRESSURE")
		case strings.Contains(lower, "solltest du") && hasAnyWord(lower, "fragen", "hilfe", "bedarf"):
			add("sollen", "CONDITIONAL_OPENING", .805, .155, domain.SenseHigh)
			patterns = appendUnique(patterns, "CONDITIONAL_OPENING")
		default:
			add("sollen", "SOCIAL_NORM", .68, .08, domain.SenseMedium)
		}
	}
	if hasAnyWord(lower, "darf", "darfst", "dürfen") {
		if hasAnyWord(lower, "nicht", "kein", "keine") {
			add("dürfen", "PROHIBITION", .765, .075, domain.SenseMedium)
			patterns = appendUnique(patterns, "PROHIBITION")
		} else {
			add("dürfen", "PERMISSION", .69, .085, domain.SenseMedium)
			patterns = appendUnique(patterns, "CONSENT_LANGUAGE")
		}
	}
	if strings.Contains(lower, "eintritt ist frei") {
		add("frei", "FREE_OF_CHARGE", .80, .09, domain.SenseMedium)
	}
	if strings.Contains(lower, "technisches problem") {
		add("Problem", "TECHNICAL_ISSUE", .875, .135, domain.SenseHigh)
		patterns = appendUnique(patterns, "TECHNICAL_ISSUE")
	} else if strings.Contains(lower, "du bist das problem") {
		add("Problem", "PERSON_LABEL", .815, .075, domain.SenseMedium)
		patterns = appendUnique(patterns, "PERSON_DEVALUATION", "PREDICATIVE_LABELING")
	}
	if strings.Contains(lower, "fehler zeigt") {
		add("Fehler", "LEARNING_EVENT", .84, .09, domain.SenseMedium)
		patterns = appendUnique(patterns, "LEARNING_FRAME", "LEARNING_RECOVERY", "OPENING_LANGUAGE")
	}
	if strings.Contains(lower, "hindernis umfahren") {
		add("umfahren", "DRIVE_AROUND", .7825, .24, domain.SenseHigh)
	}
	if hasWord(lower, "eigentlich") {
		add("eigentlich", "ORIGINAL_INTENTION", .765, .02, domain.SenseAmbiguous)
		ambiguities = append(ambiguities, domain.Ambiguity{Item: "eigentlich", Type: domain.AmbiguitySemantic, Top: "ORIGINAL_INTENTION", Second: "HEDGE", Gap: .02})
	}

	if strings.Contains(lower, "ich verstehe") && strings.Contains(lower, "nicht infrage") {
		patterns = appendUnique(patterns, "ACKNOWLEDGEMENT", "CLEAR_BOUNDARY", "RESPECTFUL_BOUNDARY")
	}
	if strings.HasPrefix(lower, "ja") && strings.Contains(lower, "aber") {
		if hasAnyWord(lower, "nie", "immer", "sowieso") {
			patterns = appendUnique(patterns, "DISCOUNTING", "GENERALIZATION")
		} else if strings.Contains(lower, "unterscheid") {
			patterns = appendUnique(patterns, "CONSTRUCTIVE_DIFFERENTIATION")
		}
	}
	if hasAnyWord(lower, "unbedingt", "sofort") {
		patterns = appendUnique(patterns, "URGENCY")
	}

	sort.SliceStable(positioned, func(i, j int) bool { return positioned[i].position < positioned[j].position })
	senses := make([]domain.ResolverSense, len(positioned))
	for index := range positioned {
		senses[index] = positioned[index].sense
	}
	_ = graph
	return senses, ambiguities, patterns
}

func hasTarget(text string) bool {
	return hasAnyWord(text, "das", "diese", "dieser", "problem", "fehler", "hindernis", "unterlagen", "gebäude", "entscheidung", "situationen") || strings.Contains(text, "infrage")
}

func normalize(text string) string { return strings.ToLower(strings.Join(strings.Fields(text), " ")) }

func hasAnyWord(text string, words ...string) bool {
	for _, word := range words {
		if hasWord(text, word) {
			return true
		}
	}
	return false
}

func hasWord(text, word string) bool {
	for _, field := range strings.FieldsFunc(text, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) }) {
		if field == word {
			return true
		}
	}
	return false
}

func appendUnique(values []string, additions ...string) []string {
	for _, addition := range additions {
		found := false
		for _, value := range values {
			if value == addition {
				found = true
				break
			}
		}
		if !found {
			values = append(values, addition)
		}
	}
	return values
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[index:])
}
