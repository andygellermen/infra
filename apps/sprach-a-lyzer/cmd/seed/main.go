package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/config"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/db"
	"github.com/andygellermann/infra/apps/sprach-a-lyzer/internal/seed"
)

func main() {
	foundationPath := flag.String("foundation", "data/seed/sprach-a-lyzer_foundation_v0.4.json", "foundation seed JSON")
	goldenPath := flag.String("golden", "data/golden/sprach-a-lyzer_vertical-slice_v0.2.json", "vertical-slice golden JSON")
	flag.Parse()

	foundationFile, err := os.Open(*foundationPath)
	if err != nil {
		fail(fmt.Errorf("open foundation seed: %w", err))
	}
	defer foundationFile.Close()
	goldenFile, err := os.Open(*goldenPath)
	if err != nil {
		fail(fmt.Errorf("open golden seed: %w", err))
	}
	defer goldenFile.Close()

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

	result, err := seed.Apply(ctx, database, foundationFile, goldenFile)
	if err != nil {
		fail(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(result)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
