package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/andygellermann/infra/apps/sheet-helper/internal/config"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/httpapp"
	"github.com/andygellermann/infra/apps/sheet-helper/internal/storage"
	sheetsync "github.com/andygellermann/infra/apps/sheet-helper/internal/sync"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	store, err := storage.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer store.Close()

	if err := store.InitSchema(context.Background()); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	if cfg.SeedFile != "" {
		if err := sheetsync.SeedFromFile(context.Background(), store, cfg.SeedFile); err != nil {
			log.Fatalf("seed data: %v", err)
		}
	}

	app := httpapp.New(cfg, store)
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("sheet-helper listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
