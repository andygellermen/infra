package main

import (
	"context"
	"log"

	"github.com/andygellermann/infra/apps/easy-event-planner/internal/config"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db"
	"github.com/andygellermann/infra/apps/easy-event-planner/internal/db/migrations"
)

var buildVersion = "dev"

func main() {
	cfg, err := config.Load(buildVersion)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	sqlDB, err := db.Open(cfg.DBDriver, cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer sqlDB.Close()

	migrator := db.NewMigrator(sqlDB, migrations.Files, ".")
	result, err := migrator.Up(context.Background())
	if err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	if len(result.Applied) == 0 {
		log.Printf("migrations up to date (total=%d)", result.Total)
		return
	}

	for _, file := range result.Applied {
		log.Printf("applied migration %s", file)
	}
	log.Printf("applied %d migration(s) (total=%d)", len(result.Applied), result.Total)
}
