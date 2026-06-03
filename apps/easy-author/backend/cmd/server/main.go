package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/andygellermann/infra/apps/easy-author/backend/internal/config"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/db"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/httpapi"
	"github.com/andygellermann/infra/apps/easy-author/backend/internal/store"
)

func main() {
	cfg := config.Load()

	database, err := db.OpenSQLite(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("easy-author backend: open database: %v", err)
	}
	defer database.Close()

	appStore := store.New(database, cfg.LibraryDir)
	if err := appStore.Init(context.Background()); err != nil {
		log.Fatalf("easy-author backend: init store: %v", err)
	}

	app := httpapi.New(cfg, appStore)
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("easy-author backend listening on http://%s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("easy-author backend: %v", err)
	}
}
