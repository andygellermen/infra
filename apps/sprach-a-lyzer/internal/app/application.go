// Package app is the composition root of the modular monolith.
package app

import (
	"database/sql"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/analysis"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/knowledge"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/presentation"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/questions"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/rules"
)

type Application struct {
	Analysis     *analysis.Service
	Knowledge    *knowledge.Service
	Rules        *rules.Service
	Presentation *presentation.Service
	Questions    *questions.Service
	Readiness    db.SchemaPinger
}

func New(database *sql.DB) *Application {
	ruleRepository := rules.NewPostgresRepository(database)
	presentationService := presentation.New(presentation.NewPostgresRepository(database))
	analysisService := analysis.NewWithRuntime(ruleRepository, presentationService)
	return &Application{
		Analysis:     analysisService,
		Knowledge:    knowledge.New(knowledge.NewPostgresRepository(database)),
		Rules:        rules.New(ruleRepository),
		Presentation: presentationService,
		Questions:    questions.New(analysisService),
		Readiness:    db.NewSchemaPinger(database, db.RequiredSchemaVersion),
	}
}
