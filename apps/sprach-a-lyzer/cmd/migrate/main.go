package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/config"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db/migrations"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	if err := cfg.ValidateDatabase(); err != nil {
		fail(err)
	}
	database, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		fail(err)
	}
	defer database.Close()

	result, err := db.NewMigrator(database, migrations.Files, ".").Up(ctx)
	if err != nil {
		fail(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(result)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
