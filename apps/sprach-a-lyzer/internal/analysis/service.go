// Package analysis is the public boundary of the deterministic analysis module.
// Callers depend on this facade instead of its resolver and scoring internals.
package analysis

import (
	"fmt"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/engine"
)

type Request = domain.AnalysisRequest
type Result = domain.AnalysisResult
type Trace = domain.AnalysisTrace
type TraceV02 = domain.AnalysisTraceV02
type TraceProposition = domain.TraceProposition
type ContributionTraceEntryV02 = domain.ContributionTraceEntryV02
type DimensionID = domain.DimensionID
type DimensionResult = domain.DimensionResult
type Proposition = domain.Proposition
type ResolvedSense = domain.ResolvedSense
type ContributionTraceEntry = domain.ContributionTraceEntry
type AssessabilityTraceEntry = domain.AssessabilityTraceEntry
type ResonanceHint = domain.ResonanceHint
type ResolverResult = domain.ResolverResult
type PropositionGraph = domain.PropositionGraph
type PropositionNode = domain.PropositionNode
type PropositionEdge = domain.PropositionEdge
type ResolverSense = domain.ResolverSense
type Ambiguity = domain.Ambiguity
type TargetTypeID = domain.TargetTypeID
type ExpectationSourceID = domain.ExpectationSourceID
type Locale = domain.Locale
type Context = domain.AnalysisContext
type InputMode = domain.InputMode
type PresentationProfile = domain.PresentationProfile
type Mode = domain.AnalysisMode

var (
	ErrEmptyText = engine.ErrEmptyText
)

const NotAssessable = domain.NotAssessable

const (
	LocaleGerman         = domain.LocaleGerman
	ContextUnspecified   = domain.ContextUnspecified
	ContextSelfTalk      = domain.ContextSelfTalk
	ContextSafety        = domain.ContextSafety
	InputModeText        = domain.InputModeText
	ProfilePrivate       = domain.ProfilePrivate
	ProfileCorporate     = domain.ProfileCorporate
	AnalysisModeStandard = domain.AnalysisModeStandard
)

func Dimensions() []DimensionID {
	return domain.CanonicalDimensions()
}

// Core is the module-internal analysis pipeline contract. It keeps the facade
// independently testable while the deterministic engine evolves behind it.
type Core interface {
	Analyze(Request) (Result, error)
}

type ContextResolver interface {
	Resolve(Request) (ResolverResult, error)
}

type Service struct {
	core     Core
	resolver ContextResolver
}

func New(core Core) *Service {
	service := &Service{core: core}
	service.resolver, _ = core.(ContextResolver)
	return service
}

func NewDefault() *Service {
	return New(engine.NewDefault())
}

func NewWithRuleCatalogue(provider engine.CatalogueProvider) *Service {
	return New(engine.New(provider))
}

func NewWithRuntime(catalogue engine.CatalogueProvider, texts engine.TextProvider) *Service {
	return New(engine.NewWithProviders(catalogue, texts))
}

func (s *Service) Analyze(request Request) (Result, error) {
	return s.core.Analyze(request)
}

func (s *Service) Resolve(request Request) (ResolverResult, error) {
	if s.resolver == nil {
		return ResolverResult{}, fmt.Errorf("context resolver is unavailable")
	}
	return s.resolver.Resolve(request)
}
