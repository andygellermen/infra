// Package analysis is the public boundary of the deterministic analysis module.
// Callers depend on this facade instead of its resolver and scoring internals.
package analysis

import (
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/domain"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/engine"
)

type Request = domain.AnalysisRequest
type Result = domain.AnalysisResult
type DimensionID = domain.DimensionID

var (
	ErrEmptyText = engine.ErrEmptyText
)

const NotAssessable = domain.NotAssessable

func Dimensions() []DimensionID {
	return domain.CanonicalDimensions()
}

// Core is the module-internal analysis pipeline contract. It keeps the facade
// independently testable while the deterministic engine evolves behind it.
type Core interface {
	Analyze(Request) (Result, error)
}

type Service struct {
	core Core
}

func New(core Core) *Service {
	return &Service{core: core}
}

func NewDefault() *Service {
	return New(engine.New())
}

func (s *Service) Analyze(request Request) (Result, error) {
	return s.core.Analyze(request)
}
