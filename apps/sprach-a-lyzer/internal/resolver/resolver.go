// Package resolver implements the deterministic Context & Proposition layer.
package resolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	assets "github.com/andygellermann/infra/apps/sprach-a-lyzer"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
)

const ContractVersion = "0.2"

var (
	ErrEmptyText         = errors.New("resolver text must not be empty")
	ErrInvalidSourceSpan = errors.New("resolver proposition span does not match source")
)

type Resolver struct {
	catalogue CatalogueProvider
}

func New() *Resolver {
	catalogue, err := DecodeCatalogue(bytes.NewReader(assets.ResolverCatalogueV01))
	if err != nil {
		panic(fmt.Sprintf("decode embedded resolver catalogue: %v", err))
	}
	return NewWithCatalogueProvider(StaticCatalogueProvider{Catalogue: catalogue})
}

func NewWithCatalogueProvider(provider CatalogueProvider) *Resolver {
	return &Resolver{catalogue: provider}
}

func (r *Resolver) Resolve(request domain.AnalysisRequest) (domain.ResolverResult, error) {
	text := strings.TrimSpace(request.Text)
	if text == "" {
		return domain.ResolverResult{}, ErrEmptyText
	}
	if r == nil || r.catalogue == nil {
		return domain.ResolverResult{}, fmt.Errorf("resolver catalogue provider is nil")
	}
	catalogue, err := r.catalogue.Active(context.Background())
	if err != nil {
		return domain.ResolverResult{}, fmt.Errorf("load active resolver catalogue: %w", err)
	}
	if err := catalogue.Validate(); err != nil {
		return domain.ResolverResult{}, fmt.Errorf("validate active resolver catalogue: %w", err)
	}
	runtime := newCatalogueRuntime(catalogue)
	contextValue := domain.AnalysisContext(strings.ToUpper(strings.TrimSpace(string(request.Context))))
	if contextValue == "" {
		contextValue = domain.ContextUnspecified
	}
	targetType := resolveTargetType(text, runtime)
	expectationSource := resolveExpectationSource(text, runtime)
	graph := buildGraph(text, targetType, expectationSource, runtime)
	senses, ambiguities, patterns := resolveSensesAndPatterns(text, contextValue, graph, runtime)
	confidence := .90
	if len(ambiguities) > 0 || hasAmbiguousSense(senses) {
		confidence = .72
	}
	result := domain.ResolverResult{
		ContractVersion: ContractVersion, Text: text, Context: contextValue,
		PropositionGraph: graph, SelectedSenses: senses, Ambiguities: ambiguities,
		TargetType: targetType, ExpectationSource: expectationSource,
		PatternCandidates: patterns, OverallConfidence: confidence,
	}
	if err := validateSourceSpans(result); err != nil {
		return domain.ResolverResult{}, err
	}
	return result, nil
}

type span struct {
	text                 string
	start, end, sentence int
	marker               string
	relation             domain.DiscourseRelationID
	confidence           float64
}

func buildGraph(text string, target domain.TargetTypeID, expectation domain.ExpectationSourceID, runtime catalogueRuntime) domain.PropositionGraph {
	spans := propositionSpans(text, runtime)
	graph := domain.PropositionGraph{Nodes: make([]domain.PropositionNode, 0, len(spans)), Edges: []domain.PropositionEdge{}}
	for index, part := range spans {
		node := propositionNode(index, part, target, expectation, runtime)
		graph.Nodes = append(graph.Nodes, node)
		if index == 0 {
			continue
		}
		marker, relation := part.marker, part.relation
		confidence := part.confidence
		if relation == "" {
			if connector, ok := runtime.relationIn(part.text); ok {
				marker, relation, confidence = connector.marker, connector.relation, connector.confidence
			}
		}
		if relation != "" {
			graph.Edges = append(graph.Edges, domain.PropositionEdge{
				Source: graph.Nodes[index-1].ID, Target: node.ID, Marker: marker,
				Relation: relation, Confidence: confidence,
			})
		}
	}
	return graph
}

func propositionSpans(text string, runtime catalogueRuntime) []span {
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
		result = append(result, splitConnectors(text, sentence, runtime, 0)...)
	}
	return result
}

func splitConnectors(source string, part span, runtime catalogueRuntime, depth int) []span {
	if depth > len(runtime.connectors) {
		return []span{part}
	}
	if comma, connector, ok := runtime.prefixClauseAt(part.text); ok {
		left, leftOK := trimmedSpan(source, part.start, part.start+comma, part.sentence)
		right, rightOK := trimmedSpan(source, part.start+comma+1, part.end, part.sentence)
		if leftOK && rightOK {
			left.marker, left.relation, left.confidence = part.marker, part.relation, part.confidence
			right.marker, right.relation, right.confidence = connector.marker, connector.relation, connector.confidence
			return append(splitConnectors(source, left, runtime, depth+1), splitConnectors(source, right, runtime, depth+1)...)
		}
	}
	markerStart, connector, ok := runtime.connectorAt(part.text, true)
	if !ok || markerStart <= 0 || !connectorSplitsPropositions(part.text, markerStart, connector) {
		return []span{part}
	}
	leftStart, leftEnd := part.start, part.start+markerStart
	for leftEnd > leftStart && (unicode.IsSpace(rune(source[leftEnd-1])) || strings.ContainsRune(",;:", rune(source[leftEnd-1]))) {
		leftEnd--
	}
	left, leftOK := trimmedSpan(source, leftStart, leftEnd, part.sentence)
	right, rightOK := trimmedSpan(source, part.start+markerStart+len(connector.marker), part.end, part.sentence)
	if !leftOK || !rightOK {
		return []span{part}
	}
	left.marker, left.relation, left.confidence = part.marker, part.relation, part.confidence
	right.marker, right.relation, right.confidence = connector.marker, connector.relation, connector.confidence
	return append(splitConnectors(source, left, runtime, depth+1), splitConnectors(source, right, runtime, depth+1)...)
}

func connectorSplitsPropositions(text string, markerStart int, connector runtimeConnector) bool {
	if connector.relation != domain.RelationAddition || connector.marker != "und" {
		return true
	}
	words := lexicalWords(normalize(text[markerStart+len(connector.marker):]))
	if len(words) == 0 {
		return false
	}
	switch words[0] {
	case "ich", "du", "wir", "ihr", "er", "sie", "es", "man":
		return true
	default:
		return false
	}
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

func propositionNode(index int, part span, target domain.TargetTypeID, expectation domain.ExpectationSourceID, runtime catalogueRuntime) domain.PropositionNode {
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
	modality := resolveModality(lower, runtime)
	negationScope := domain.NegationNone
	if negation {
		negationScope = runtime.negationScope(lower)
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

func resolveModality(text string, runtime catalogueRuntime) domain.ModalityID {
	switch {
	case runtime.matchesLexeme("MUSSEN", text):
		return domain.ModalityNecessity
	case runtime.matchesLexeme("SOLLEN", text):
		return domain.ModalityExpectation
	case runtime.matchesLexeme("DUERFEN", text):
		if hasWord(text, "dürfte") {
			return domain.ModalityProbability
		}
		return domain.ModalityPermission
	case hasAnyWord(text, "kann", "können", "könnte"):
		return domain.ModalityPossibility
	case hasAnyWord(text, "möchte", "will", "wollte"):
		return domain.ModalityIntention
	default:
		return domain.ModalityNone
	}
}

func resolveTargetType(text string, runtime catalogueRuntime) domain.TargetTypeID {
	lower := normalize(text)
	switch {
	case runtime.hasSense("PROBLEM", "PERSON_LABEL") && strings.Contains(lower, "du bist das problem"),
		runtime.hasSense("FEHLER", "IDENTITY_LABEL") && strings.Contains(lower, "du bist ein fehler"):
		return domain.TargetPerson
	case runtime.hasSense("PROBLEM", "TECHNICAL_ISSUE") && (strings.Contains(lower, "technisches problem") || strings.Contains(lower, "schnittstelle")),
		runtime.hasSense("FEHLER", "TECHNICAL_ERROR") && strings.Contains(lower, "im code"):
		return domain.TargetProcess
	case runtime.hasSense("FEHLER", "LEARNING_EVENT") && (strings.Contains(lower, "fehler zeigt") || strings.Contains(lower, "nächsten versuch")):
		return domain.TargetEvent
	case strings.Contains(lower, "vereinbarung") && strings.Contains(lower, "eingehalten"):
		return domain.TargetBehavior
	default:
		return domain.TargetUnknown
	}
}

func resolveExpectationSource(text string, runtime catalogueRuntime) domain.ExpectationSourceID {
	lower := normalize(text)
	switch {
	case hasAnyWord(lower, "gesetzlich", "gesetz", "vorgeschrieben"):
		return domain.ExpectationLaw
	case runtime.hasSense("SOLLEN", "INTERNALIZED_EXPECTATION") && strings.Contains(lower, "ich sollte") && hasAnyWord(lower, "längst", "endlich", "eigentlich"):
		return domain.ExpectationInternalized
	case strings.Contains(lower, "man sollte"):
		return domain.ExpectationCulture
	case runtime.hasSense("MUSSEN", "INTERNAL_PRESSURE") && strings.Contains(lower, "ich muss") && hasAnyWord(lower, "unbedingt", "endlich", "perfekt"):
		return domain.ExpectationInternalized
	default:
		return domain.ExpectationUnspecified
	}
}

type positionedSense struct {
	position int
	sense    domain.ResolverSense
}

func resolveSensesAndPatterns(text string, contextValue domain.AnalysisContext, graph domain.PropositionGraph, runtime catalogueRuntime) ([]domain.ResolverSense, []domain.Ambiguity, []string) {
	lower := normalize(text)
	var positioned []positionedSense
	ambiguities := []domain.Ambiguity{}
	patterns := []string{}
	add := func(key, lexeme, sense string, confidence, gap float64) bool {
		if !runtime.matchesLexeme(key, lower) || !runtime.hasSense(key, sense) {
			return false
		}
		positioned = append(positioned, positionedSense{position: runtime.lexemePosition(key, lower), sense: domain.ResolverSense{
			Lexeme: lexeme, Sense: sense, Confidence: confidence, Gap: gap, State: runtime.senseState(confidence, gap),
		}})
		return true
	}

	if runtime.matchesLexeme("MUSSEN", lower) {
		switch {
		case strings.Contains(lower, "ich muss") && hasAnyWord(lower, "unbedingt", "endlich", "perfekt"):
			if add("MUSSEN", "müssen", "INTERNAL_PRESSURE", .785, .13) {
				patterns = appendUnique(patterns, "INTERNAL_PRESSURE")
			}
		case strings.Contains(lower, "hindernis umfahren"):
			if add("MUSSEN", "müssen", "EXTERNAL_NECESSITY", .655, .05) && runtime.hasSense("MUSSEN", "EPISTEMIC_INFERENCE") {
				ambiguities = append(ambiguities, domain.Ambiguity{Item: "müssen", Type: domain.AmbiguitySemantic, Top: "EXTERNAL_NECESSITY", Second: "EPISTEMIC_INFERENCE", Gap: .05})
			}
		case contextValue == domain.ContextSafety && (hasAnyWord(lower, "gefahr", "brand", "notfall") || strings.Contains(lower, "gebäude verlassen")):
			add("MUSSEN", "müssen", "SAFETY_NECESSITY", .90, .25)
		default:
			add("MUSSEN", "müssen", "EXTERNAL_NECESSITY", .66, .08)
		}
	}
	if runtime.matchesLexeme("SOLLEN", lower) {
		switch {
		case (strings.HasPrefix(lower, "er soll ") || strings.HasPrefix(lower, "sie soll ")) && strings.Contains(lower, " sein"):
			if add("SOLLEN", "sollen", "REPORTED_CLAIM", .79, .14) {
				patterns = appendUnique(patterns, "REPORTED_CLAIM")
			}
		case strings.Contains(lower, "ich sollte") && hasAnyWord(lower, "längst", "endlich", "eigentlich"):
			if add("SOLLEN", "sollen", "INTERNALIZED_EXPECTATION", .785, .135) {
				patterns = appendUnique(patterns, "INTERNALIZED_EXPECTATION", "SELF_PRESSURE")
			}
		case strings.Contains(lower, "solltest du") && hasAnyWord(lower, "fragen", "hilfe", "bedarf"):
			if add("SOLLEN", "sollen", "CONDITIONAL_OPENING", .805, .155) {
				patterns = appendUnique(patterns, "CONDITIONAL_OPENING")
			}
		default:
			add("SOLLEN", "sollen", "SOCIAL_NORM", .68, .08)
		}
	}
	if runtime.matchesLexeme("DUERFEN", lower) {
		if hasAnyWord(lower, "nicht", "kein", "keine") {
			if add("DUERFEN", "dürfen", "PROHIBITION", .765, .075) {
				patterns = appendUnique(patterns, "PROHIBITION")
			}
		} else {
			if add("DUERFEN", "dürfen", "PERMISSION", .69, .085) {
				patterns = appendUnique(patterns, "CONSENT_LANGUAGE")
			}
		}
	}
	if strings.Contains(lower, "eintritt ist frei") {
		add("FREI", "frei", "FREE_OF_CHARGE", .80, .09)
	}
	if strings.Contains(lower, "technisches problem") {
		if add("PROBLEM", "Problem", "TECHNICAL_ISSUE", .875, .135) {
			patterns = appendUnique(patterns, "TECHNICAL_ISSUE")
		}
	} else if strings.Contains(lower, "du bist das problem") {
		if add("PROBLEM", "Problem", "PERSON_LABEL", .815, .075) {
			patterns = appendUnique(patterns, "PERSON_DEVALUATION", "PREDICATIVE_LABELING")
		}
	}
	if strings.Contains(lower, "fehler zeigt") {
		if add("FEHLER", "Fehler", "LEARNING_EVENT", .84, .09) {
			patterns = appendUnique(patterns, "LEARNING_FRAME", "LEARNING_RECOVERY", "OPENING_LANGUAGE")
		}
	}
	if strings.Contains(lower, "hindernis umfahren") {
		add("UMFAHREN", "umfahren", "DRIVE_AROUND", .7825, .24)
	}
	if runtime.matchesLexeme("EIGENTLICH", lower) {
		if add("EIGENTLICH", "eigentlich", "ORIGINAL_INTENTION", .765, .02) && runtime.hasSense("EIGENTLICH", "HEDGE") {
			ambiguities = append(ambiguities, domain.Ambiguity{Item: "eigentlich", Type: domain.AmbiguitySemantic, Top: "ORIGINAL_INTENTION", Second: "HEDGE", Gap: .02})
		}
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

func validateSourceSpans(result domain.ResolverResult) error {
	for _, node := range result.PropositionGraph.Nodes {
		if node.SourceStart < 0 || node.SourceEnd > len(result.Text) || node.SourceStart >= node.SourceEnd ||
			result.Text[node.SourceStart:node.SourceEnd] != node.Text {
			return fmt.Errorf("%w: %s at %d:%d", ErrInvalidSourceSpan, node.ID, node.SourceStart, node.SourceEnd)
		}
	}
	return nil
}

func hasAmbiguousSense(senses []domain.ResolverSense) bool {
	for _, sense := range senses {
		if sense.State == domain.SenseAmbiguous {
			return true
		}
	}
	return false
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
