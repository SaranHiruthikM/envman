package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/SaranHiruthikM/envman/internal/config"
	"github.com/SaranHiruthikM/envman/internal/db"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	_, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	mux := chi.NewMux()

	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.App.Port), mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
