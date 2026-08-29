package ontology

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/policy"
)

type CatalogueProvider interface {
	Active(context.Context) (Catalogue, error)
}

type StaticProvider struct{ Catalogue Catalogue }

func (s StaticProvider) Active(context.Context) (Catalogue, error) { return s.Catalogue, nil }

type Evidence struct {
	ConstructID    policy.ConstructID
	InferenceClass string
	ClaimMode      string
	PropositionIDs []string
}

type CompositionMatch struct {
	Pattern        string
	ConstructIDs   []policy.ConstructID
	PropositionIDs []string
}

type Result struct {
	Evidence     []Evidence
	Compositions []CompositionMatch
}

type Runtime struct{ provider CatalogueProvider }

func NewRuntime(provider CatalogueProvider) *Runtime { return &Runtime{provider: provider} }

func (r *Runtime) Resolve(ctx context.Context, resolution domain.ResolverResult) (Result, error) {
	if r == nil || r.provider == nil {
		return Result{}, fmt.Errorf("construct ontology provider is nil")
	}
	catalogue, err := r.provider.Active(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("load active construct ontology: %w", err)
	}
	if err := catalogue.Validate(); err != nil {
		return Result{}, fmt.Errorf("validate active construct ontology: %w", err)
	}
	result := Result{Evidence: []Evidence{}, Compositions: []CompositionMatch{}}
	for _, definition := range catalogue.Constructs {
		for _, node := range resolution.PropositionGraph.Nodes {
			if matchesDefinition(definition, node, resolution.SelectedSenses) {
				result.Evidence = append(result.Evidence, Evidence{
					ConstructID: definition.ID, InferenceClass: definition.InferenceClass,
					ClaimMode: definition.ClaimMode, PropositionIDs: []string{node.ID},
				})
			}
		}
	}
	for _, composition := range catalogue.Compositions {
		if match, ok := matchComposition(composition, result.Evidence, resolution.PropositionGraph); ok {
			result.Compositions = append(result.Compositions, match)
		}
	}
	return result, nil
}

func matchesDefinition(definition Definition, node domain.PropositionNode, senses []domain.ResolverSense) bool {
	for _, signal := range definition.RuntimeSignals {
		if matchesSignal(signal, node, senses) {
			return true
		}
	}
	return false
}

func matchesSignal(signal RuntimeSignal, node domain.PropositionNode, senses []domain.ResolverSense) bool {
	text := strings.ToLower(strings.Join(strings.Fields(node.Text), " "))
	for _, phrase := range signal.PhrasesAll {
		if !strings.Contains(text, strings.ToLower(phrase)) {
			return false
		}
	}
	if len(signal.PhrasesAny) > 0 && !containsAnyPhrase(text, signal.PhrasesAny) {
		return false
	}
	if len(signal.Actors) > 0 && !slices.Contains(signal.Actors, policy.ActorID(node.Actor)) {
		return false
	}
	if len(signal.Modalities) > 0 && !slices.Contains(signal.Modalities, policy.ModalityID(node.Modality)) {
		return false
	}
	if len(signal.TargetTypes) > 0 && !slices.Contains(signal.TargetTypes, policy.TargetTypeID(node.TargetType)) {
		return false
	}
	if len(signal.ExpectationSources) > 0 && !slices.Contains(signal.ExpectationSources, policy.ExpectationSourceID(node.ExpectationSource)) {
		return false
	}
	for _, feature := range signal.PropositionFeatures {
		if !nodeFeature(node, feature) {
			return false
		}
	}
	if len(signal.SelectedSenses) > 0 {
		matched := false
		for _, sense := range senses {
			if sense.PropositionID == node.ID && sense.State != domain.SenseAmbiguous && slices.Contains(signal.SelectedSenses, sense.Sense) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return !signal.empty()
}

func containsAnyPhrase(text string, phrases []string) bool {
	for _, phrase := range phrases {
		if strings.Contains(text, strings.ToLower(phrase)) {
			return true
		}
	}
	return false
}

func nodeFeature(node domain.PropositionNode, feature string) bool {
	switch feature {
	case "PREDICATE":
		return node.Predicate
	case "TARGET":
		return node.Target
	case "TIME":
		return node.Time
	case "BOUNDARY":
		return node.Boundary
	case "DECISION":
		return node.Decision
	case "NEGATION":
		return node.Negation
	default:
		return strings.HasPrefix(feature, "MODALITY_") && strings.TrimPrefix(feature, "MODALITY_") == string(node.Modality)
	}
}

func matchComposition(composition Composition, evidence []Evidence, graph domain.PropositionGraph) (CompositionMatch, bool) {
	positions := make(map[string]int, len(graph.Nodes))
	for index, node := range graph.Nodes {
		positions[node.ID] = index
	}
	choices := make([][]Evidence, len(composition.RequiredConstructs))
	for index, required := range composition.RequiredConstructs {
		for _, item := range evidence {
			if item.ConstructID == required {
				choices[index] = append(choices[index], item)
			}
		}
		if len(choices[index]) == 0 {
			return CompositionMatch{}, false
		}
	}
	selected := make([]Evidence, len(choices))
	var search func(int) bool
	search = func(index int) bool {
		if index == len(choices) {
			return validEvidenceSequence(selected, composition, positions, graph)
		}
		for _, candidate := range choices[index] {
			selected[index] = candidate
			if search(index + 1) {
				return true
			}
		}
		return false
	}
	if !search(0) {
		return CompositionMatch{}, false
	}
	ids := []string{}
	for _, item := range selected {
		for _, id := range item.PropositionIDs {
			if !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
	}
	sort.SliceStable(ids, func(i, j int) bool { return positions[ids[i]] < positions[ids[j]] })
	return CompositionMatch{Pattern: composition.OutputPattern, ConstructIDs: slices.Clone(composition.RequiredConstructs), PropositionIDs: ids}, true
}

func validEvidenceSequence(selected []Evidence, composition Composition, positions map[string]int, graph domain.PropositionGraph) bool {
	indexes := make([]int, len(selected))
	for index, item := range selected {
		if len(item.PropositionIDs) != 1 {
			return false
		}
		indexes[index] = positions[item.PropositionIDs[0]]
		if index > 0 && composition.Ordered && indexes[index] <= indexes[index-1] {
			return false
		}
	}
	if indexes[len(indexes)-1]-indexes[0] > composition.MaximumPropositionGap {
		return false
	}
	if len(composition.RelationsAny) == 0 {
		return true
	}
	for index := 1; index < len(selected); index++ {
		if !hasRelation(graph.Edges, selected[index-1].PropositionIDs[0], selected[index].PropositionIDs[0], composition.RelationsAny) {
			return false
		}
	}
	return true
}

func hasRelation(edges []domain.PropositionEdge, source, target string, allowed []policy.DiscourseRelationID) bool {
	for _, edge := range edges {
		if edge.Source == source && edge.Target == target && slices.Contains(allowed, policy.DiscourseRelationID(edge.Relation)) {
			return true
		}
	}
	return false
}
