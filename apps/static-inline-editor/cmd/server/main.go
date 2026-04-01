package main

import (
	"log"
	"net/http"

	"github.com/andygellermann/infra/apps/static-inline-editor/internal/config"
	"github.com/andygellermann/infra/apps/static-inline-editor/internal/httpapp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	app := httpapp.New(cfg)

	log.Printf("static-inline-editor listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, app.Handler()); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
